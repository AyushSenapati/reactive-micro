package event

const EventPolicyUpdated EventName = "EventPolicyUpdated"

// register the event to the registry
func init() {
	Registry.register(EventPolicyUpdated, EventInfo{
		ReqChan: "authzsvc.EventPolicyUpdated",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventPolicyUpdatedPayload)
			return ok
		},
	})
}

type EventPolicyUpdatedPayload struct {
	Method       string `json:"method"` // could be put/delete
	Sub          string `json:"subject"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Action       string `json:"action"`
}
