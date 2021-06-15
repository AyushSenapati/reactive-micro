package endpoint

import (
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a order service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	CreateMerchantEndpoint endpoint.Endpoint
	ListMerchantEndpoint   endpoint.Endpoint
	CreateProductEndpoint  endpoint.Endpoint
	ListProductEndpoint    endpoint.Endpoint
}

// New returns a Endpoints struct that wraps the provided service, and wires in all of the
// expected endpoint middlewares
func New(s service.IInventoryService, mdw map[string][]endpoint.Middleware) Endpoints {
	eps := Endpoints{
		CreateMerchantEndpoint: makeCreateMerchantEndpoint(s),
		ListMerchantEndpoint:   makeListMerchantEndpoint(s),
		CreateProductEndpoint:  makeCreateProductEndpoint(s),
		ListProductEndpoint:    makeListProductEndpoint(s),
	}

	// apply transport middlewares
	for _, m := range mdw["CreateMerchant"] {
		eps.CreateMerchantEndpoint = m(eps.CreateMerchantEndpoint)
	}

	for _, m := range mdw["ListMerchant"] {
		eps.ListMerchantEndpoint = m(eps.ListMerchantEndpoint)
	}

	for _, m := range mdw["CreateProduct"] {
		eps.CreateProductEndpoint = m(eps.CreateProductEndpoint)
	}

	for _, m := range mdw["ListProduct"] {
		eps.ListProductEndpoint = m(eps.ListProductEndpoint)
	}

	return eps
}
