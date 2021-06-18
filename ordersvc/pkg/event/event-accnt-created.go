package event

const EventAccountCreated EventName = "EventAccountCreated"

// register the event to the registry
func init() {
	Registry.register(EventAccountCreated, EventInfo{
		ReqChan: "authnsvc.EventAccountCreated",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventAccountCreatedPayload)
			return ok
		},
	})
}

type EventAccountCreatedPayload struct {
	AccntID uint   `json:"accnt_id"`
	Role    string `json:"role"`
}
