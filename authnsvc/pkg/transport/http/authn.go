package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// makeLoginHandler creates the handler logic
func makeLoginHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("POST").Path("/login").Handler(
		kithttp.NewServer(endpoints.LoginEndpoint, decodeLoginRequest, encodeHTTPGenericResponse, options...))
}

// decodeLoginRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := dto.LoginRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
