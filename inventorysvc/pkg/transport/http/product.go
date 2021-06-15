package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/endpoint"
	ce "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/error"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// makeCreateProductHandler creates the handler logic
func makeCreateProductHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("POST").Path("/merchants/{mid}/products").Handler(
		kithttp.NewServer(
			endpoints.CreateProductEndpoint,
			decodeCreateProductRequest,
			encodeHTTPGenericResponse,
			options...,
		))
}

// decodeCreateProductRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeCreateProductRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	req := dto.CreateProductRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		return nil, ce.ErrInvalidReqBody
	}
	vars := mux.Vars(r)
	mid := vars["mid"]
	req.MID, err = uuid.Parse(mid)
	if err != nil {
		return nil, &ce.ResourceNotFoundErr{Type: "merchant", ID: mid}
	}
	return req, err
}

// makeListProductHandler creates the handler logic
func makeListProductHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("GET").Path("/products").Handler(
		kithttp.NewServer(
			endpoints.ListProductEndpoint,
			decodeListProductRequest,
			encodeHTTPGenericResponse,
			options...,
		))
}

// decodeListProductRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeListProductRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	qp := processBasicQP(r)
	return qp, nil
}
