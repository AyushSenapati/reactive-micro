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

func makeCreateProductEndpoint(s service.IInventoryService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		claim, ok := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
		if !ok {
			return dto.CreateMerchantResponse{Err: kitjwt.ErrTokenContextMissing}, nil
		}

		reqObj, ok := request.(dto.CreateProductRequest)
		if !ok {
			return dto.CreateProductResponse{Err: ce.ErrInvalidReqBody}, nil
		}

		return s.CreateProduct(ctx, claim.AccntID, reqObj.MID, reqObj.Name, reqObj.Desc, reqObj.Qty, reqObj.Price), nil
	}

}

func makeListProductEndpoint(s service.IInventoryService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		qp, _ := request.(*dto.BasicQueryParam)
		return s.ListProduct(ctx, []uuid.UUID{}, qp), nil
	}
}
