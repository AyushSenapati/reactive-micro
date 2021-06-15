package dto

type ListPolicyRequest struct {
	Sub          string `json:"sub"`
	ResourceType string `json:"resource_type"`
}

type ListPolicyResponse struct {
	Policies []string `json:"policies"`
}

type Policy struct {
	Sub          string `json:"subject"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Action       string `json:"action"`
}

type UpsertPolicyRequest struct {
	*Policy
}

type RemovePolicyRequest struct {
	*Policy
}
