package endpoint

import (
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	LoginEndpoint endpoint.Endpoint
	// RenewAccessTokenEndpoint endpoint.Endpoint
	// VerifyTokenEndpoint      endpoint.Endpoint
	// RevokeTokenEndpoint      endpoint.Endpoint
	CreateAccountEndpoint endpoint.Endpoint
	ListAccountEndpoint   endpoint.Endpoint
	// GetUserEndpoint          endpoint.Endpoint
	// UpdateUserEndpoint       endpoint.Endpoint
	DeleteAccountEndpoint endpoint.Endpoint
	// CreateRoleEndpoint       endpoint.Endpoint
	// ListRoleEndpoint         endpoint.Endpoint
	// DeleteRoleEndpoint       endpoint.Endpoint
}

// New returns a Endpoints struct that wraps the provided service, and wires in all of the
// expected endpoint middlewares
func New(s service.IAuthNService, mdw map[string][]endpoint.Middleware) Endpoints {
	eps := Endpoints{
		// CreateRoleEndpoint:       MakeCreateRoleEndpoint(s),
		CreateAccountEndpoint: MakeCreateAccountEndpoint(s),
		// DeleteRoleEndpoint:       MakeDeleteRoleEndpoint(s),
		DeleteAccountEndpoint: MakeDeleteAccountEndpoint(s),
		// GetUserEndpoint:          MakeGetUserEndpoint(s),
		// ListRoleEndpoint:         MakeListRoleEndpoint(s),
		ListAccountEndpoint: MakeListAccountEndpoint(s),
		LoginEndpoint:       MakeLoginEndpoint(s),
		// RenewAccessTokenEndpoint: MakeRenewAccessTokenEndpoint(s),
		// RevokeTokenEndpoint:      MakeRevokeTokenEndpoint(s),
		// UpdateUserEndpoint:       MakeUpdateUserEndpoint(s),
		// VerifyTokenEndpoint:      MakeVerifyTokenEndpoint(s),
	}
	for _, m := range mdw["Login"] {
		eps.LoginEndpoint = m(eps.LoginEndpoint)
	}
	// for _, m := range mdw["RenewAccessToken"] {
	// 	eps.RenewAccessTokenEndpoint = m(eps.RenewAccessTokenEndpoint)
	// }
	// for _, m := range mdw["VerifyToken"] {
	// 	eps.VerifyTokenEndpoint = m(eps.VerifyTokenEndpoint)
	// }
	// for _, m := range mdw["RevokeToken"] {
	// 	eps.RevokeTokenEndpoint = m(eps.RevokeTokenEndpoint)
	// }
	for _, m := range mdw["CreateAccount"] {
		eps.CreateAccountEndpoint = m(eps.CreateAccountEndpoint)
	}
	for _, m := range mdw["ListAccount"] {
		eps.ListAccountEndpoint = m(eps.ListAccountEndpoint)
	}
	// for _, m := range mdw["GetUser"] {
	// 	eps.GetUserEndpoint = m(eps.GetUserEndpoint)
	// }
	// for _, m := range mdw["UpdateUser"] {
	// 	eps.UpdateUserEndpoint = m(eps.UpdateUserEndpoint)
	// }
	for _, m := range mdw["DeleteAccount"] {
		eps.DeleteAccountEndpoint = m(eps.DeleteAccountEndpoint)
	}
	// for _, m := range mdw["CreateRole"] {
	// 	eps.CreateRoleEndpoint = m(eps.CreateRoleEndpoint)
	// }
	// for _, m := range mdw["ListRole"] {
	// 	eps.ListRoleEndpoint = m(eps.ListRoleEndpoint)
	// }
	// for _, m := range mdw["DeleteRole"] {
	// 	eps.DeleteRoleEndpoint = m(eps.DeleteRoleEndpoint)
	// }
	return eps
}
