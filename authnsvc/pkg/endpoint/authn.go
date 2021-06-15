package endpoint

import (
	"context"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

func MakeLoginEndpoint(s service.IAuthNService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		loginReq := request.(dto.LoginRequest)
		resp := s.GenToken(ctx, loginReq)
		return resp, nil
	}
}
