package endpoint

import (
	"context"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

func MakeCreateAccountEndpoint(s service.IAuthNService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqObj := request.(dto.CreateAccountRequest)
		respObj := s.CreateAccount(ctx, reqObj)
		return respObj, nil
	}
}

func MakeDeleteAccountEndpoint(s service.IAuthNService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		aid := request.(uint)
		return s.DeleteAccount(ctx, aid), nil
	}
}

func MakeListAccountEndpoint(s service.IAuthNService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		qp, _ := request.(*dto.BasicQueryParam)
		return s.ListAccount(ctx, []uint{}, qp), nil
	}
}
