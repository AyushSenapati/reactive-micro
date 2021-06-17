package policyenforcer

// matcherFunc defines how to match request policy with the stored policy
type matcherFunc func(r, p Policy) bool

// defaultMatcher function will be used when no custom
// matcher function in provided to the policy enforcer
func defaultMatcher(r, p Policy) bool {
	if r.sub == p.sub && r.rtype == p.rtype && (r.act == p.act || p.act == "*") && (r.rid == p.rid || p.rid == "*") {
		return true
	}
	return false
}

type PolicyEnforcer interface {
	Enforce(string, matcherFunc) bool
	GetResourceIDs(sub, rtype, act string) []string
}

func NewPolicyEnforcer(ps PolicyStorage) (PolicyEnforcer, error) {
	if ps == nil {
		return nil, ErrNilPolicyStorage
	}
	return &policyEnforcer{storage: ps}, nil
}

type policyEnforcer struct {
	storage PolicyStorage
}

func (pe *policyEnforcer) Enforce(s string, matcher matcherFunc) bool {
	r, err := getPolicyFromString(s)
	if err != nil {
		return false
	}

	if matcher == nil {
		matcher = defaultMatcher
	}

	for _, p := range pe.storage.GetPolicyForSub(r.sub) {
		if matcher(r, p) {
			return true
		}
	}
	return false
}

func (pe *policyEnforcer) GetResourceIDs(sub, rtype, act string) (rids []string) {
	fPolicies := pe.storage.GetPolicyForSub(sub)
	for _, fp := range fPolicies {
		if rtype == fp.rtype && act == fp.act {
			rids = append(rids, fp.rid)
		}
	}
	return
}
