package http

import (
	"net/http"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPHandler returns a handler that makes a
// set of endpoints available on predefined paths.
func NewHTTPHandler(endpoints endpoint.Endpoints, options map[string][]kithttp.ServerOption) http.Handler {
	m := mux.NewRouter()
	m = m.PathPrefix("/v1/inventorysvc").Subrouter()

	makeCreateMerchantHandler(m, endpoints, options["CreateMerchant"])
	makeListMerchantHandler(m, endpoints, options["ListMerchant"])
	makeCreateProductHandler(m, endpoints, options["CreateProduct"])
	makeListProductHandler(m, endpoints, options["ListProduct"])

	return m
}
