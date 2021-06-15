package event

import "github.com/google/uuid"

const EventOrderCanceled EventName = "EventOrderCanceled"

// register the event to the registry
func init() {
	Registry.register(EventOrderCanceled, EventInfo{
		ReqChan: "ordersvc.EventOrderCanceled",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventOrderCanceledPayload)
			return ok
		},
	})
}

type EventOrderCanceledPayload struct {
	OID     uuid.UUID `json:"order_id"`
	AccntID uint      `json:"account_id"`
}
