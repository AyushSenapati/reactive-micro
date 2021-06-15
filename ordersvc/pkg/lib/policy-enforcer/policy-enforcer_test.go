package policyenforcer

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestBuildFromStrPolicies(t *testing.T) {
	policies := new(ePolicy)
	policies.buildFromStrPolicies([]string{"1:orders:post:*", "1:orders:get:100"})
	encodedFeps, _ := json.Marshal(policies)
	fmt.Println(string(encodedFeps))
}

func TestBuildFromFormattedPolicies(t *testing.T) {
	policies := new(ePolicy)
	var fps []Policy
	for _, rp := range []string{"1:orders:post:*", "1:orders:get:100"} {
		fp, _ := getPolicyFromString(rp)
		fps = append(fps, fp)
	}
	policies.buildFromFormattedPolicies(fps)
	encodedFeps, _ := json.Marshal(policies)
	fmt.Println(string(encodedFeps))
}

func TestRemovePolicyFromEPolicy(t *testing.T) {
	ep := new(ePolicy)
	var fps []Policy
	for _, rp := range []string{"1:orders:post:*", "1:orders:get:100", "1:orders:get:101"} {
		fp, _ := getPolicyFromString(rp)
		fps = append(fps, fp)
	}
	ep.buildFromFormattedPolicies(fps)

	fp := Policy{sub: "1", rtype: "orders", rid: "100", act: "get"}
	ep.remove(fp)
	encodedFeps, _ := json.Marshal(ep)
	fmt.Println(string(encodedFeps))

	fmt.Println(ep.fpolicies("1"))
}
