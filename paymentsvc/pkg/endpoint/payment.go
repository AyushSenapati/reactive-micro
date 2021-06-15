package endpoint

import (
	"context"

	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/service"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

func makeRechargeWalletEndpoint(s service.IPaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reqObj, ok := request.(dto.RechargeWalletRequest)
		if !ok {
			return dto.RechargeWalletResponse{Err: err}, nil
		}
		claim := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)

		txid, err := s.RechargeWallet(ctx, claim.AccntID, reqObj.Amount)
		return dto.RechargeWalletResponse{
			TXID: txid,
			Err:  err,
		}, nil
	}
}

func makeListTransactionsEndpoint(s service.IPaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		qp, _ := request.(*dto.BasicQueryParam)
		return s.ListTransactions(ctx, []uuid.UUID{}, qp), nil
	}
}
