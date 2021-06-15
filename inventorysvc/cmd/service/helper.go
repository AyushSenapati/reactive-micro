package service

import (
	httptransport "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/transport/http"
	kithttp "github.com/go-kit/kit/transport/http"
)

func defaultHttpOptions() map[string][]kithttp.ServerOption {
	options := map[string][]kithttp.ServerOption{}
	addSrvOptToALlMethods(options, kithttp.ServerErrorEncoder(httptransport.ErrorEncoder))
	return options
}

func addSrvOptToALlMethods(options map[string][]kithttp.ServerOption, o kithttp.ServerOption) {
	for _, method := range allMethods {
		options[method] = append(options[method], o)
	}
}

// func addEndpointMWToAllMethods(mw map[string][]kitep.Middleware, m kitep.Middleware) {
// 	for _, method := range allMethods {
// 		mw[method] = append(mw[method], m)
// 	}
// }