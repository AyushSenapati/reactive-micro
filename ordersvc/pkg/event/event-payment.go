package event

import "github.com/google/uuid"

const EventPayment EventName = "EventPayment"

// register the event to the registry
func init() {
	Registry.register(EventPayment, EventInfo{
		ReqChan: "paymentsvc.EventPayment",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventPaymentPayload)
			return ok
		},
	})
}

type EventPaymentPayload struct {
	OrderID uuid.UUID `json:"order_id"`
	AccntID uint      `json:"account_id"`
	Status  string    `json:"status"` // can be on of [payment_successful / payment_failed]
}
