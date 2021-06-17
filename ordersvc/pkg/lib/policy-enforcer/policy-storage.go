package policyenforcer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/patrickmn/go-cache"
)

type Policy struct {
	sub, rtype, rid, act string
}

func sliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// ePolicy - Efficient Policy
// in the cache every sub should be mapped to an ePolicy object
type ePolicy map[string]map[string][]string

func (ep *ePolicy) buildFromStrPolicies(rawPolicies []string) {
	for _, rp := range rawPolicies {
		fp, err := getPolicyFromString(rp)
		if err != nil {
			continue
		}
		ep.upsert(fp)
	}
}

func (ep *ePolicy) buildFromFormattedPolicies(fps []Policy) {
	for _, fp := range fps {
		ep.upsert(fp)
	}
}

func (ep *ePolicy) upsert(fp Policy) error {
	fep := make(ePolicy)
	fep[fp.rtype] = map[string][]string{
		fp.act: {fp.rid},
	}
	err := mergo.MapWithOverwrite(ep, fep, mergo.WithAppendSlice)
	if err != nil {
		msg := fmt.Sprintf("merging fep error. [%v]", err)
		fmt.Println(msg)
		return errors.New(msg)
	}
	return nil
}

func (ep *ePolicy) remove(fp Policy) {
	a, ok := (*ep)[fp.rtype]
	if !ok { // means resource type/policy does not exist
		return
	}
	resourceIDs, ok := a[fp.act]
	if !ok { // means action/policy does not exist
		return
	}

	idx := sliceIndex(len(resourceIDs), func(i int) bool { return resourceIDs[i] == fp.rid })
	if idx < 0 {
		return
	}

	resourceIDs[idx], resourceIDs[len(resourceIDs)-1] = resourceIDs[len(resourceIDs)-1], resourceIDs[idx]
	resourceIDs = resourceIDs[:len(resourceIDs)-1]
	a[fp.act] = resourceIDs
}

func (ep *ePolicy) fpolicies(sub string) (policies []Policy) {
	if len(sub) <= 0 {
		return
	}
	for resourceType, actions := range *ep {
		for action, resourceIDs := range actions {
			for _, rid := range resourceIDs {
				policies = append(
					policies, Policy{sub: sub, rtype: resourceType, act: action, rid: rid})
			}
		}
	}
	return
}

// gets you formatted policy from its string version
// ex: 1:accounts:get:2 means 1 can read account with id 2
func getPolicyFromString(s string) (Policy, error) {
	p := Policy{}
	c := strings.Split(s, ":")
	if len(c) != 4 {
		return p, ErrInvalidPolicy
	}
	p.sub, p.rtype, p.act, p.rid = c[0], c[1], c[2], c[3]
	return p, nil
}

type policiesResponse struct {
	Policies []string `json:"policies"`
}

type getPoliciesRequest struct {
	Sub          string `json:"sub"`
	ResourceType string `json:"resource_type"`
}

type PolicyStorageMW func(PolicyStorage) PolicyStorage

type PolicyStorage interface {
	GetPolicyForSub(sub string) []Policy
	UpdatePolicy(method, sub, rtype, rid, act string) error
}

type cachedPolicyStorage struct {
	url    string
	rtypes []string // resource types supported by this service
	cache  *cache.Cache
}

func NewCachedPolicyStorageMW(url string, rtype []string, c *cache.Cache) (PolicyStorage, error) {
	if len(url) == 0 || len(rtype) == 0 {
		return nil, errors.New("url and resource types must be provided")
	}
	if c == nil {
		c = cache.New(5*time.Minute, 10*time.Minute) // sets default cache
	}
	return &cachedPolicyStorage{
		url:    url,
		rtypes: rtype,
		cache:  c,
	}, nil
}

func (cps *cachedPolicyStorage) FetchPolicyForSub(sub string) *ePolicy {
	sPolicies := []string{}
	client := &http.Client{}

	// get policies for the subject and all the resource types supported by this service
	for _, rtype := range cps.rtypes {
		body := &getPoliciesRequest{Sub: sub, ResourceType: rtype}
		jsonbody, _ := json.Marshal(body)
		req, _ := http.NewRequest("GET", cps.url, bytes.NewBuffer(jsonbody))
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if resp.StatusCode != 200 {
			fmt.Printf(
				"failed fetching policies for sub: %s [got status code: %d]\n",
				sub, resp.StatusCode)
			continue
		}
		defer resp.Body.Close()
		respBody := policiesResponse{}
		json.NewDecoder(resp.Body).Decode(&respBody)
		sPolicies = append(sPolicies, respBody.Policies...)
	}

	ep := new(ePolicy)
	ep.buildFromStrPolicies(sPolicies)
	return ep
}

func (cps *cachedPolicyStorage) GetPolicyForSub(sub string) (fPolicies []Policy) {
	cachedEPolicy, found := cps.cache.Get(sub) // gets ePolicy obj

	// if entry not found for the subject get policies from authz svc and cache it
	if !found {
		fmt.Println("cache miss for sub:", sub)
		ep := cps.FetchPolicyForSub(sub)
		fPolicies := ep.fpolicies(sub)
		if len(fPolicies) > 0 {
			fmt.Println("caching policies for sub:", sub)
			// adds formatted policies to the cache with default expiry
			cps.cache.Set(sub, ep, 0)
		}
		return fPolicies
	}

	// if an entry is found for the subject update the expiry of the cache
	fmt.Println("cache hit for sub:", sub)
	ep := cachedEPolicy.(*ePolicy)
	cps.cache.Set(sub, ep, 0) // updates expiration by default expiry
	return ep.fpolicies(sub)
}

func (cps *cachedPolicyStorage) UpdatePolicy(method, sub, rtype, rid, act string) error {
	var found bool
	var err error

	for _, r := range cps.rtypes {
		if rtype == r {
			found = true
			break
		}
	}
	if !found {
		msg := fmt.Sprintf("cps: rtype-%s not supported by this svc. skip re-caching", rtype)
		fmt.Println(msg)
		return ErrUnsupportedRtype
	}

	// if policy for the sub is not found in the cache, skip updating cache
	cachedEPolicy, found := cps.cache.Get(sub) // gets formatted policies
	if !found {
		msg := fmt.Sprintf("cps: policy for sub-%s not found. skip fetching", sub)
		fmt.Println(msg)
		return ErrSubNotCached
	}

	ep := cachedEPolicy.(*ePolicy)
	fp := Policy{sub: sub, rtype: rtype, rid: rid, act: act}
	if method == "put" {
		err = ep.upsert(fp)
	} else if method == "delete" {
		ep.remove(fp)
	} else {
		return nil
	}

	if err != nil {
		return err
	}

	fmt.Printf("cps: updated[%s] cache for sub-%s\n", method, sub)
	cps.cache.Set(sub, ep, 0) // adds ePolicy object to the cache with default expiry

	return nil
}
