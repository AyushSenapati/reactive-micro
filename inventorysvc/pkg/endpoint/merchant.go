package endpoint

import (
	"context"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/error"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/service"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

func makeCreateMerchantEndpoint(s service.IInventoryService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		claim, ok := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
		if !ok {
			return dto.CreateMerchantResponse{Err: kitjwt.ErrTokenContextMissing}, nil
		}
		reqObj, ok := request.(dto.CreateMerchantRequest)
		if !ok {
			return dto.CreateMerchantResponse{Err: ce.ErrInvalidReqBody}, nil
		}

		return s.CreateMerchant(ctx, claim.AccntID, reqObj.Name), nil
	}
}

func makeListMerchantEndpoint(s service.IInventoryService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return s.ListMerchant(ctx, []uuid.UUID{}), nil
	}
}
