package event

import "github.com/google/uuid"

const EventOrderApproved EventName = "EventOrderApproved"

// register the event to the registry
func init() {
	Registry.register(EventOrderApproved, EventInfo{
		ReqChan: "ordersvc.EventOrderApproved",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventOrderApprovedPayload)
			return ok
		},
	})
}

type EventOrderApprovedPayload struct {
	OID     uuid.UUID `json:"order_id"`
	AccntID uint      `json:"account_id"`
}
