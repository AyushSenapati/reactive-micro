package endpoint

import (
	"context"

	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/error"
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/service"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

// func MakeHandleAccountCreatedEventEndpoint(s service.IOrderService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
// 		reqObj, ok := request.(svcevent.EventAccountCreatedPayload)
// 		if !ok {
// 			return nil, svcevent.ErrInvalidPayload
// 		}
// 		err = s.HandleAccountCreatedEvent(ctx, reqObj.AccntID, reqObj.Role)
// 		return nil, err
// 	}
// }

func MakeCreatedOrderEndpoint(s service.IOrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reqObj, ok := request.(dto.CreateOrderRequest)
		if !ok {
			return dto.CreateOrderResponse{Err: ce.ErrInvalidReqBody}, nil
		}

		oid, err := s.CreateOrder(ctx, reqObj.PID, reqObj.Qty)
		return dto.CreateOrderResponse{OID: oid, Err: err}, nil
	}
}

func MakeListOrderEndpoint(s service.IOrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		qp, _ := request.(*dto.BasicQueryParam)
		return s.ListOrder(ctx, []uuid.UUID{}, qp), nil
	}
}
