package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"

	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/endpoint"
	ce "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/error"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// makeCreateOrderHandler creates the handler logic
func makeCreateOrderHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("POST").Path("/orders").Handler(
		kithttp.NewServer(
			endpoints.CreateOrderEndpoint,
			decodeCreateOrderRequest,
			encodeHTTPGenericResponse,
			options...,
		))
}

// decodeCreateOrderRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeCreateOrderRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	req := dto.CreateOrderRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		err = ce.ErrInvalidReqBody
	}
	return req, err
}

// makeListOrderHandler creates the handler logic
func makeListOrderHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("GET").Path("/orders").Handler(
		kithttp.NewServer(
			endpoints.ListOrderEndpoint,
			decodeListOrderRequest,
			encodeHTTPGenericResponse,
			options...,
		))
}

// decodeListOrderRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeListOrderRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	qp := processBasicQP(r)
	return qp, nil
}
