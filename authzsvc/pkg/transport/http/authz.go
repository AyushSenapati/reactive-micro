package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func makeListPolicyHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("GET").Path("/policies").Handler(
		kithttp.NewServer(
			endpoints.ListPolicyEndpoint,
			decodeListPolicyRequest,
			encodeHTTPGenericResponse,
			options...,
		),
	)
}

func decodeListPolicyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := dto.ListPolicyRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	body, _ := json.Marshal(req)
	fmt.Println(string(body))
	return req, err
}

func makeUpsertPolicyHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("POST", "PUT").Path("/policies").Handler(
		kithttp.NewServer(
			endpoints.UpsertPolicyEndpoint,
			decodeUpsertPolicyRequest,
			encodeHTTPGenericResponse,
			options...,
		),
	)
}

func decodeUpsertPolicyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := dto.UpsertPolicyRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	body, _ := json.Marshal(req)
	fmt.Println(string(body))
	return req, err
}

func makeRemovePolicyHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("DELETE").Path("/policies").Handler(
		kithttp.NewServer(
			endpoints.RemovePolicyEndpoint,
			decodeRemovePolicyRequest,
			encodeHTTPGenericResponse,
			options...,
		),
	)
}

func decodeRemovePolicyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := dto.RemovePolicyRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	body, _ := json.Marshal(req)
	fmt.Println(string(body))
	return req, err
}

func makeRemovePolicyBySubHandler(m *mux.Router, endpoints endpoint.Endpoints, options []kithttp.ServerOption) {
	m.Methods("DELETE").Path("/policies/{sub}").Handler(
		kithttp.NewServer(
			endpoints.RemovePolicyBySubEndpoint,
			decodeRemovePolicyBySubRequest,
			encodeHTTPGenericResponse,
			options...,
		),
	)
}

func decodeRemovePolicyBySubRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	sub := vars["sub"]
	return sub, nil
}
