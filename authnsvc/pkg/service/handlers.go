package service

import (
	"context"
)

func (svc *basicAuthNService) HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error {
	return svc.ps.UpdatePolicy(method, sub, rtype, rid, act)
}
