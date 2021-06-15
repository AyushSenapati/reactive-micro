package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/error"
	svcpe "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/lib/policy-enforcer"
	kitjwt "github.com/go-kit/kit/auth/jwt"
)

type authzMW struct {
	pe   svcpe.PolicyEnforcer
	next IAuthNService
}

func NewAuthzMW(pe svcpe.PolicyEnforcer) Middleware {
	return func(ia IAuthNService) IAuthNService {
		return &authzMW{pe: pe, next: ia}
	}
}

func (m *authzMW) HandlePolicyUpdatedEvent(ctx context.Context, t, sub, rtype, rid, act string) error {
	return m.next.HandlePolicyUpdatedEvent(ctx, t, sub, rtype, rid, act)
}

func (m *authzMW) DeleteAccount(ctx context.Context, aid uint) (err error) {
	claim := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	reqPolicy := fmt.Sprintf("%v:%s:%s:%v", claim.AccntID, "accounts", "delete", aid)
	fmt.Println("request policy:", reqPolicy)
	if !m.pe.Enforce(reqPolicy, nil) {
		return ce.ErrInsufficientPerm
	}
	return m.next.DeleteAccount(ctx, aid)
}

func (m *authzMW) GenToken(ctx context.Context, accnt dto.LoginRequest) dto.LoginResponse {
	return m.next.GenToken(ctx, accnt)
}

func (m *authzMW) CreateAccount(ctx context.Context, accnt dto.CreateAccountRequest) dto.CreateAccountResponse {
	return m.next.CreateAccount(ctx, accnt)
}

func (m *authzMW) ListAccount(ctx context.Context, aids []uint, qp *dto.BasicQueryParam) dto.ListAccountResponse {
	claim, ok := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	if !ok {
		return dto.ListAccountResponse{Err: kitjwt.ErrTokenContextMissing}
	}
	rids := m.pe.GetResourceIDs(fmt.Sprint(claim.AccntID), "accounts", "*")
	rids = append(rids, m.pe.GetResourceIDs(fmt.Sprint(claim.AccntID), "accounts", "get")...)
	if len(rids) <= 0 {
		return dto.ListAccountResponse{Err: ce.ErrInsufficientPerm}
	}
	for _, rid := range rids {
		if rid == "*" {
			return m.next.ListAccount(ctx, []uint{}, qp)
		}
		id, err := strconv.Atoi(rid)
		if err != nil {
			continue
		}
		aids = append(aids, uint(id))
	}
	return m.next.ListAccount(ctx, aids, qp)
}
