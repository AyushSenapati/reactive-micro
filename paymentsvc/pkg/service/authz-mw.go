package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/error"
	svcpe "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/lib/policy-enforcer"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/google/uuid"
)

type authzMW struct {
	pe   svcpe.PolicyEnforcer
	next IPaymentService
}

func NewAuthzMW(pe svcpe.PolicyEnforcer) Middleware {
	return func(ia IPaymentService) IPaymentService {
		return &authzMW{pe: pe, next: ia}
	}
}

func (m *authzMW) HandleAccountCreatedEvent(ctx context.Context, accntID uint, role string) error {
	return m.next.HandleAccountCreatedEvent(ctx, accntID, role)
}

func (m *authzMW) HandlePolicyUpdatedEvent(ctx context.Context, t, sub, rtype, rid, act string) error {
	return m.next.HandlePolicyUpdatedEvent(ctx, t, sub, rtype, rid, act)
}

func (m *authzMW) HandleProductReservedEvent(ctx context.Context, oid uuid.UUID, aid uint, payble float32) error {
	return m.next.HandleProductReservedEvent(ctx, oid, aid, payble)
}

func (m *authzMW) RechargeWallet(ctx context.Context, aid uint, amount float32) (uuid.UUID, error) {
	reqPolicy := fmt.Sprintf("%v:%s:%s:%v", aid, "transactions", "post", "*")
	if !m.pe.Enforce(ctx, reqPolicy, nil) {
		return uuid.Nil, ce.ErrInsufficientPerm
	}
	return m.next.RechargeWallet(ctx, aid, amount)
}

func (m *authzMW) ListTransactions(ctx context.Context, txids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListTransactionsResponse {
	claim, ok := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	if !ok {
		return dto.ListTransactionsResponse{Err: kitjwt.ErrTokenContextMissing}
	}

	// get possible transaction IDs that can be retrived by the authenticated subject
	rids := m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "transactions", "*")
	rids = append(rids, m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "transactions", "get")...)

	if len(rids) <= 0 {
		return dto.ListTransactionsResponse{Err: ce.ErrInsufficientPerm}
	}

	for _, rid := range rids {
		if rid == "*" {
			return m.next.ListTransactions(ctx, []uuid.UUID{}, qp)
		}
		id, err := uuid.Parse(rid)
		if err != nil {
			continue
		}
		txids = append(txids, id)
	}

	return m.next.ListTransactions(ctx, txids, qp)
}
