package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"

	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/endpoint"
	ce "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/error"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func makeRechargeWalletHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("POST").Path("/recharge-wallet").Handler(
		kithttp.NewServer(
			endpoints.RechargeWalletEndpoint,
			decodeRechargeWalletRequest,
			encodeHTTPGenericResponse,
			options...,
		),
	)
}

// decodeRechargeWalletRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeRechargeWalletRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	req := dto.RechargeWalletRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		err = ce.ErrInvalidReqBody
	}
	return req, err
}

func makeListTransactionsHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("GET").Path("/transactions").Handler(
		kithttp.NewServer(
			endpoints.ListTransactionsEndpoint,
			decodeListTransactionsRequest,
			encodeHTTPGenericResponse,
			options...,
		),
	)
}

// decodeListTransactionsRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeListTransactionsRequest(_ context.Context, r *stdhttp.Request) (interface{}, error) {
	qp := processBasicQP(r)
	return qp, nil
}
