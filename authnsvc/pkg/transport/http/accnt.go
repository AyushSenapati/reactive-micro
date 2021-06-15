package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"strconv"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/endpoint"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// makeCreateAccountHandler creates the handler logic
func makeCreateAccountHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("POST").Path("/accounts").Handler(
		kithttp.NewServer(
			endpoints.CreateAccountEndpoint,
			decodeCreateAccountRequest,
			encodeHTTPGenericResponse,
			options...,
		))
}

// decodeCreateAccountRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeCreateAccountRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	req := dto.CreateAccountRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	return req, err
}

// makeDeleteAccountHandler creates the handler logic
func makeDeleteAccountHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("DELETE").Path("/accounts/{aid:[0-9]+}").Handler(
		kithttp.NewServer(
			endpoints.DeleteAccountEndpoint,
			decodeDeleteUserRequest,
			encodeHTTPGenericResponse,
			options...))
}

// decodeDeleteAccountRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeDeleteUserRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	// Get account ID from the URI
	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["aid"])
	if err != nil {
		return nil, err
	}

	return uint(aid), err
}

func makeListAccountHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("GET").Path("/accounts").Handler(
		kithttp.NewServer(
			endpoints.ListAccountEndpoint,
			decodeListResouceRequest,
			encodeHTTPGenericResponse,
			options...))
}

func decodeListResouceRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	qp := processBasicQP(r)
	return qp, nil
}
