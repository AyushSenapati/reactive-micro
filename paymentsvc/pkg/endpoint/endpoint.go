package endpoint

import (
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a order service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	RechargeWalletEndpoint   endpoint.Endpoint
	ListTransactionsEndpoint endpoint.Endpoint
}

// New returns a Endpoints struct that wraps the provided service, and wires in all of the
// expected endpoint middlewares
func New(s service.IPaymentService, mdw map[string][]endpoint.Middleware) Endpoints {
	eps := Endpoints{
		RechargeWalletEndpoint:   makeRechargeWalletEndpoint(s),
		ListTransactionsEndpoint: makeListTransactionsEndpoint(s),
	}

	// apply transport middlewares
	for _, m := range mdw["RechargeWallet"] {
		eps.RechargeWalletEndpoint = m(eps.RechargeWalletEndpoint)
	}

	for _, m := range mdw["ListTransactions"] {
		eps.ListTransactionsEndpoint = m(eps.ListTransactionsEndpoint)
	}

	return eps
}
