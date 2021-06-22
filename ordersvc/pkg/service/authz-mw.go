package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/error"
	svcpe "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/lib/policy-enforcer"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/google/uuid"
)

type authzMW struct {
	pe   svcpe.PolicyEnforcer
	next IOrderService
}

func NewAuthzMW(pe svcpe.PolicyEnforcer) Middleware {
	return func(ia IOrderService) IOrderService {
		return &authzMW{pe: pe, next: ia}
	}
}

func (m *authzMW) HandleAccountCreatedEvent(ctx context.Context, accntID uint, role string) error {
	return m.next.HandleAccountCreatedEvent(ctx, accntID, role)
}

func (m *authzMW) HandlePolicyUpdatedEvent(ctx context.Context, t, sub, rtype, rid, act string) error {
	return m.next.HandlePolicyUpdatedEvent(ctx, t, sub, rtype, rid, act)
}

func (m *authzMW) HandleErrReservingProductEvent(ctx context.Context, oid uuid.UUID) error {
	return m.next.HandleErrReservingProductEvent(ctx, oid)
}

func (m *authzMW) HandleProductReservedEvent(ctx context.Context, oid uuid.UUID) error {
	return m.next.HandleProductReservedEvent(ctx, oid)
}

func (m *authzMW) HandlePaymentEvent(ctx context.Context, oid uuid.UUID, aid uint, status string) error {
	return m.next.HandlePaymentEvent(ctx, oid, aid, status)
}

func (m *authzMW) CreateOrder(ctx context.Context, pid uuid.UUID, qty int) (uuid.UUID, error) {
	claim := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	reqPolicy := fmt.Sprintf("%v:%s:%s:%v", claim.AccntID, "orders", "post", "*")
	if !m.pe.Enforce(ctx, reqPolicy, nil) {
		return uuid.Nil, ce.ErrInsufficientPerm
	}
	return m.next.CreateOrder(ctx, pid, qty)
}

func (m *authzMW) ListOrder(ctx context.Context, oids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListOrderResponse {
	claim, ok := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	if !ok {
		return dto.ListOrderResponse{Err: kitjwt.ErrTokenContextMissing}
	}
	rids := m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "orders", "*")
	rids = append(rids, m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "orders", "get")...)
	if len(rids) <= 0 {
		return dto.ListOrderResponse{Err: ce.ErrInsufficientPerm}
	}
	for _, rid := range rids {
		if rid == "*" {
			return m.next.ListOrder(ctx, []uuid.UUID{}, qp)
		}
		id, err := uuid.Parse(rid)
		if err != nil {
			continue
		}
		oids = append(oids, id)
	}

	return m.next.ListOrder(ctx, oids, qp)
}
