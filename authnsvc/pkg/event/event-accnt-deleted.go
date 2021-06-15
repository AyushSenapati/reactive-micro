package event

const EventAccountDeleted EventName = "EventAccountDeleted"

// register the event to the registry
func init() {
	Registry.register(EventAccountDeleted, EventInfo{
		ReqChan: "accntsvc.EventAccountDeleted",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventAccountDeletedPayload)
			return ok
		},
	})
}

type EventAccountDeletedPayload struct {
	AccntID uint `json:"accnt_id"`
}
