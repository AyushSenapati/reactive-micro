package http

import (
	"net/http"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPHandler returns a handler that makes a set of endpoints available on
// predefined paths.
func NewHTTPHandler(endpoints endpoint.Endpoints, options map[string][]kithttp.ServerOption) http.Handler {
	m := mux.NewRouter()
	m = m.PathPrefix("/v1/authnsvc").Subrouter()
	makeLoginHandler(m, endpoints, options["Login"])
	// makeRenewAccessTokenHandler(m, endpoints, options["RenewAccessToken"])
	// makeVerifyTokenHandler(m, endpoints, options["VerifyToken"])
	// makeRevokeTokenHandler(m, endpoints, options["RevokeToken"])
	makeCreateAccountHandler(m, endpoints, options["CreateAccount"])
	makeListAccountHandler(m, endpoints, options["ListAccount"])
	// makeGetUserHandler(m, endpoints, options["GetUser"])
	// makeUpdateUserHandler(m, endpoints, options["UpdateUser"])
	makeDeleteAccountHandler(m, endpoints, options["DeleteAccount"])
	// makeCreateRoleHandler(m, endpoints, options["CreateRole"])
	// makeListRoleHandler(m, endpoints, options["ListRole"])
	// makeDeleteRoleHandler(m, endpoints, options["DeleteRole"])
	return m
}
