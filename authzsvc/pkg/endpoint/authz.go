package endpoint

import (
	"context"

	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/dto"
	svcevent "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
)

func MakeUpsertPolicyEndpoint(s service.IAuthzService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reqObj, ok := request.(dto.UpsertPolicyRequest)
		if !ok {
			return nil, svcevent.ErrInvalidPayload
		}
		err = s.UpsertPolicy(ctx, reqObj.Sub, reqObj.ResourceType, reqObj.ResourceID, reqObj.Action)
		return nil, err
	}
}

func MakeListPolicyEndpoint(s service.IAuthzService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reqObj := request.(dto.ListPolicyRequest)
		response = s.ListPolicy(ctx, reqObj)
		return
	}
}

func MakeRemovePolicyEndpoint(s service.IAuthzService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reqObj, ok := request.(dto.RemovePolicyRequest)
		if !ok {
			return nil, svcevent.ErrInvalidPayload
		}

		return nil, s.RemovePolicy(ctx, reqObj.Sub, reqObj.ResourceType, reqObj.ResourceID, reqObj.Action)
	}
}

func MakeRemovePolicyBySubEndpoint(s service.IAuthzService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		sub, ok := request.(string)
		if !ok {
			return nil, svcevent.ErrInvalidPayload
		}
		return nil, s.RemovePolicyBySub(ctx, sub)
	}
}
