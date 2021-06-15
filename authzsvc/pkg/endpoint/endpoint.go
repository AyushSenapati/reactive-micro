package endpoint

import (
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	UpsertPolicyEndpoint      endpoint.Endpoint
	ListPolicyEndpoint        endpoint.Endpoint
	RemovePolicyEndpoint      endpoint.Endpoint
	RemovePolicyBySubEndpoint endpoint.Endpoint
}

// New returns a Endpoints struct that wraps the provided service, and wires in all of the
// expected endpoint middlewares
func New(s service.IAuthzService, mdw map[string][]endpoint.Middleware) Endpoints {
	eps := Endpoints{
		UpsertPolicyEndpoint:      MakeUpsertPolicyEndpoint(s),
		ListPolicyEndpoint:        MakeListPolicyEndpoint(s),
		RemovePolicyEndpoint:      MakeRemovePolicyEndpoint(s),
		RemovePolicyBySubEndpoint: MakeRemovePolicyBySubEndpoint(s),
	}
	for _, m := range mdw["CreateAccount"] {
		eps.ListPolicyEndpoint = m(eps.ListPolicyEndpoint)
	}
	return eps
}
