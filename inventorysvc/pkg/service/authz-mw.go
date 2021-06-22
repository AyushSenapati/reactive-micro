package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/error"
	svcpe "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/lib/policy-enforcer"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/google/uuid"
)

type authzMW struct {
	pe   svcpe.PolicyEnforcer
	next IInventoryService
}

func NewAuthzMW(pe svcpe.PolicyEnforcer) Middleware {
	return func(ia IInventoryService) IInventoryService {
		return &authzMW{pe: pe, next: ia}
	}
}

func (m *authzMW) HandleAccountCreatedEvent(ctx context.Context, accntID uint, role string) error {
	return m.next.HandleAccountCreatedEvent(ctx, accntID, role)
}

func (m *authzMW) HandlePolicyUpdatedEvent(ctx context.Context, t, sub, rtype, rid, act string) error {
	return m.next.HandlePolicyUpdatedEvent(ctx, t, sub, rtype, rid, act)
}

func (m *authzMW) HandleOrderCreatedEvent(ctx context.Context, oid, pid uuid.UUID, status string, qty int, aid uint) error {
	return m.next.HandleOrderCreatedEvent(ctx, oid, pid, status, qty, aid)
}

func (m *authzMW) HandleOrderApprovedEvent(ctx context.Context, rpid uuid.UUID) error {
	return m.next.HandleOrderApprovedEvent(ctx, rpid)
}

func (m *authzMW) HandleOrderCanceledEvent(ctx context.Context, rpid uuid.UUID) error {
	return m.next.HandleOrderCanceledEvent(ctx, rpid)
}

func (m *authzMW) CreateMerchant(ctx context.Context, aid uint, name string) dto.CreateMerchantResponse {
	reqPolicy := fmt.Sprintf("%v:%s:%s:%v", aid, "merchants", "post", "*")
	if !m.pe.Enforce(ctx, reqPolicy, nil) {
		fmt.Println(ce.ErrInsufficientPerm)
		return dto.CreateMerchantResponse{Err: ce.ErrInsufficientPerm}
	}
	return m.next.CreateMerchant(ctx, aid, name)
}

func (m *authzMW) ListMerchant(ctx context.Context, mids []uuid.UUID) dto.ListMerchantResponse {
	claim, ok := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	if !ok {
		return dto.ListMerchantResponse{Err: kitjwt.ErrTokenContextMissing}
	}
	rids := m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "merchants", "*")
	rids = append(rids, m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "merchants", "get")...)
	if len(rids) <= 0 {
		return dto.ListMerchantResponse{Err: ce.ErrInsufficientPerm}
	}
	for _, rid := range rids {
		if rid == "*" {
			return m.next.ListMerchant(ctx, []uuid.UUID{})
		}
		id, err := uuid.Parse(rid)
		if err != nil {
			continue
		}
		mids = append(mids, id)
	}

	return m.next.ListMerchant(ctx, mids)
}

func (m *authzMW) CreateProduct(ctx context.Context, aid uint, mid uuid.UUID, name, desc string, qty int, price float32) dto.CreateProductResponse {
	reqPolicy := fmt.Sprintf("%v:%s:%s:%v", aid, "products", "post", "*")
	if !m.pe.Enforce(ctx, reqPolicy, nil) {
		return dto.CreateProductResponse{Err: ce.ErrInsufficientPerm}
	}
	return m.next.CreateProduct(ctx, aid, mid, name, desc, qty, price)
}

func (m *authzMW) ListProduct(ctx context.Context, pids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListProductResponse {
	claim, ok := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	if !ok {
		return dto.ListProductResponse{Err: kitjwt.ErrTokenContextMissing}
	}
	// customers are allowed to view all the products
	if claim.Role == "customer" {
		return m.next.ListProduct(ctx, []uuid.UUID{}, qp)
	}
	// sellers are allowed to view which they have created
	rids := m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "products", "*")
	rids = append(rids, m.pe.GetResourceIDs(ctx, fmt.Sprint(claim.AccntID), "products", "get")...)
	if len(rids) <= 0 {
		return dto.ListProductResponse{Err: ce.ErrInsufficientPerm}
	}
	for _, rid := range rids {
		if rid == "*" {
			return m.next.ListProduct(ctx, []uuid.UUID{}, qp)
		}
		id, err := uuid.Parse(rid)
		if err != nil {
			continue
		}
		pids = append(pids, id)
	}

	return m.next.ListProduct(ctx, pids, qp)
}
