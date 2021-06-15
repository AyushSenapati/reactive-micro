package event

const EventAccountAuthenticated EventName = "EventAccountAuthenticated"

// register the event to the registry
func init() {
	Registry.register(EventAccountAuthenticated, EventInfo{
		ReqChan: "accntsvc.EventAccountAuthenticated",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventAccountAuthenticatedPayload)
			return ok
		},
	})
}

type EventAccountAuthenticatedPayload struct {
	AccntID uint `json:"accnt_id"`
}
