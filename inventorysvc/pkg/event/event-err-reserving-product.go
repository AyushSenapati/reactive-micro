package event

import "github.com/google/uuid"

const EventErrReservingProduct EventName = "EventErrReservingProduct"

// register the event to the registry
func init() {
	Registry.register(EventErrReservingProduct, EventInfo{
		ReqChan: "inventorysvc.EventErrReservingProduct",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventErrReservingProductPayload)
			return ok
		},
	})
}

type EventErrReservingProductPayload struct {
	OrderID uuid.UUID `json:"order_id"`
}
