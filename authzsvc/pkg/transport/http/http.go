package http

import (
	"net/http"

	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPHandler returns a handler that makes a
// set of endpoints available on predefined paths.
func NewHTTPHandler(endpoints endpoint.Endpoints, options map[string][]kithttp.ServerOption) http.Handler {
	m := mux.NewRouter()
	m = m.PathPrefix("/v1/authzsvc").Subrouter()

	makeUpsertPolicyHandler(m, endpoints, options["UpsertPolicy"])
	makeListPolicyHandler(m, endpoints, options["ListPolicy"])
	makeRemovePolicyHandler(m, endpoints, options["RemovePolicy"])
	makeRemovePolicyBySubHandler(m, endpoints, options["RemovePolicyBySub"])

	return m
}
