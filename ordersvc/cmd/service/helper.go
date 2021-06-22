package service

import (
	"context"
	"net/http"

	svcconf "github.com/AyushSenapati/reactive-micro/ordersvc/conf"
	httptransport "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/transport/http"
	kithttp "github.com/go-kit/kit/transport/http"
)

func defaultHttpOptions() map[string][]kithttp.ServerOption {
	options := map[string][]kithttp.ServerOption{}
	addSrvOptToALlMethods(options, kithttp.ServerErrorEncoder(httptransport.ErrorEncoder))
	addSrvOptToALlMethods(options, kithttp.ServerBefore(moveReqIDToCtx()))

	return options
}

func addSrvOptToALlMethods(options map[string][]kithttp.ServerOption, o kithttp.ServerOption) {
	for _, method := range allMethods {
		options[method] = append(options[method], o)
	}
}

// helper function to copy RequestID from HTTP header to the context
func moveReqIDToCtx() kithttp.RequestFunc {
	return func(c context.Context, r *http.Request) context.Context {
		reqID := r.Header.Get(svcconf.C.ReqIDKey)
		return context.WithValue(c, svcconf.C.ReqIDKey, reqID)
	}
}
