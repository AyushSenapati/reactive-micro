package event

import "github.com/google/uuid"

const EventOrderCreated EventName = "EventOrderCreated"

// register the event to the registry
func init() {
	Registry.register(EventOrderCreated, EventInfo{
		ReqChan: "ordersvc.EventOrderCreated",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventOrderCreatedPayload)
			return ok
		},
	})
}

type EventOrderCreatedPayload struct {
	OrderID     uuid.UUID `json:"order_id"`
	OrderStatus string    `json:"order_status"`
	AccntID     uint      `json:"account_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Qty         int       `json:"quantity"`
}
