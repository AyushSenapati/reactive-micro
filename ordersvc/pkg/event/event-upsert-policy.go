package event

const EventUpsertPolicy EventName = "EventUpsertPolicy"

// register the event to the registry
func init() {
	Registry.register(EventUpsertPolicy, EventInfo{
		ReqChan: "authzsvc.EventUpsertPolicy",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventUpsertPolicyPayload)
			return ok
		},
	})
}

type EventUpsertPolicyPayload struct {
	Sub          string `json:"subject"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Action       string `json:"action"`
}
