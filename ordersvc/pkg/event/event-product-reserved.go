package event

import "github.com/google/uuid"

const EventProductReserved EventName = "EventProductReserved"

// register the event to the registry
func init() {
	Registry.register(EventProductReserved, EventInfo{
		ReqChan: "inventorysvc.EventProductReserved",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventProductReservedPayload)
			return ok
		},
	})
}

type EventProductReservedPayload struct {
	OrderID uuid.UUID `json:"order_id"`
	AccntID uint      `json:"account_id"`
	Payble  float32   `json:"payble"`
}
