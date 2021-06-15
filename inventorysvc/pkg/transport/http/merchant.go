package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/endpoint"
	ce "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/error"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// makeCreateMerchantHandler creates the handler logic
func makeCreateMerchantHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("POST").Path("/merchants").Handler(
		kithttp.NewServer(
			endpoints.CreateMerchantEndpoint,
			decodeCreateMerchantRequest,
			encodeHTTPGenericResponse,
			options...,
		))
}

// decodeCreateMerchantRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeCreateMerchantRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	req := dto.CreateMerchantRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		err = ce.ErrInvalidReqBody
	}
	return req, err
}

// makeListMerchantHandler creates the handler logic
func makeListMerchantHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("GET").Path("/merchants").Handler(
		kithttp.NewServer(
			endpoints.ListMerchantEndpoint,
			decodeListMerchantRequest,
			encodeHTTPGenericResponse,
			options...,
		))
}

// decodeListMerchantRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeListMerchantRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	return r, nil
}
