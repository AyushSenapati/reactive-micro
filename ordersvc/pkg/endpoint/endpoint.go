package endpoint

import (
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a order service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	CreateOrderEndpoint endpoint.Endpoint
	ListOrderEndpoint   endpoint.Endpoint
}

// New returns a Endpoints struct that wraps the provided service, and wires in all of the
// expected endpoint middlewares
func New(s service.IOrderService, mdw map[string][]endpoint.Middleware) Endpoints {
	eps := Endpoints{
		CreateOrderEndpoint: MakeCreatedOrderEndpoint(s),
		ListOrderEndpoint:   MakeListOrderEndpoint(s),
	}

	// apply transport middlewares
	for _, m := range mdw["CreateOrder"] {
		eps.CreateOrderEndpoint = m(eps.CreateOrderEndpoint)
	}
	for _, m := range mdw["ListOrder"] {
		eps.ListOrderEndpoint = m(eps.ListOrderEndpoint)
	}

	return eps
}
