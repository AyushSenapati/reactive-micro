package service

import (
	"context"

	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/dto"
	svcevent "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/event"
)

func (svc *basicAuthzService) UpsertPolicy(ctx context.Context, sub, resourceType, resourceID, action string) error {
	err := svc.repo.UpsertPolicy(ctx, sub, resourceType, resourceID, action)

	// on successful upsert policy operation fire policy updated event
	// to let other services aware of the changes and update their cache
	if err == nil {
		eventPublisher := svcevent.NewEventPublisher()
		eventPublisher.AddEvent(
			svcevent.NewEvent(
				ctx, svcevent.EventPolicyUpdated,
				svcevent.EventPolicyUpdatedPayload{
					Method:       "put",
					Sub:          sub,
					ResourceType: resourceType,
					ResourceID:   resourceID,
					Action:       action,
				},
			))
		eventPublisher.Publish(svc.nc)
	}

	return err
}

func (svc *basicAuthzService) ListPolicy(ctx context.Context, reqObj dto.ListPolicyRequest) (resp dto.ListPolicyResponse) {
	resp.Policies = svc.repo.ListPolicy(ctx, reqObj.Sub, reqObj.ResourceType)
	return
}

func (svc *basicAuthzService) RemovePolicy(ctx context.Context, sub, resourceType, resourceID, action string) error {
	err := svc.repo.RemovePolicy(ctx, sub, resourceType, resourceID, action)

	// on successful removal of a policy fire policy updated event
	// to let other services aware of the changes and update their cache
	if err == nil {
		eventPublisher := svcevent.NewEventPublisher()
		eventPublisher.AddEvent(
			svcevent.NewEvent(
				ctx, svcevent.EventPolicyUpdated,
				svcevent.EventPolicyUpdatedPayload{
					Method:       "delete",
					Sub:          sub,
					ResourceType: resourceType,
					ResourceID:   resourceID,
					Action:       action,
				},
			))
		eventPublisher.Publish(svc.nc)
	}

	return err
}

func (svc *basicAuthzService) RemovePolicyBySub(ctx context.Context, sub string) error {
	return svc.repo.RemovePolicyBySub(ctx, sub)
}
