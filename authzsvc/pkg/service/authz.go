package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/dto"
	svcevent "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/event"
)

func (svc *basicAuthzService) UpsertPolicy(ctx context.Context, sub, resourceType, resourceID, action string) error {
	err := svc.repo.UpsertPolicy(ctx, sub, resourceType, resourceID, action)
	svc.cl.LogIfError(ctx, fmt.Errorf("upsert policy err [%v]", err))

	// on successful upsert policy operation fire policy updated event
	// to let other services aware of the changes and update their cache
	if err == nil {
		eventPublisher := svcevent.NewEventPublisher()
		eventErr := eventPublisher.AddEvent(
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
		svc.cl.LogIfError(ctx, eventErr)

		eventErr = eventPublisher.Publish(svc.nc)
		svc.cl.LogIfError(ctx, eventErr)
		if eventErr == nil {
			svc.cl.Debug(ctx, fmt.Sprintf("published events: %s", svcevent.EventPolicyUpdated))
		}
	}

	return err
}

func (svc *basicAuthzService) ListPolicy(ctx context.Context, reqObj dto.ListPolicyRequest) (resp dto.ListPolicyResponse) {
	resp.Policies = svc.repo.ListPolicy(ctx, reqObj.Sub, reqObj.ResourceType)
	return
}

func (svc *basicAuthzService) RemovePolicy(ctx context.Context, sub, resourceType, resourceID, action string) error {
	err := svc.repo.RemovePolicy(ctx, sub, resourceType, resourceID, action)
	svc.cl.LogIfError(ctx, fmt.Errorf("error removing policy [%v]", err))

	// on successful removal of a policy fire policy updated event
	// to let other services aware of the changes and update their cache
	if err == nil {
		eventPublisher := svcevent.NewEventPublisher()
		eventErr := eventPublisher.AddEvent(
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
		svc.cl.LogIfError(ctx, eventErr)
		eventErr = eventPublisher.Publish(svc.nc)
		svc.cl.LogIfError(ctx, eventErr)
		if eventErr == nil {
			svc.cl.Debug(ctx, fmt.Sprintf("published events: %s", svcevent.EventPolicyUpdated))
		}
	}

	return err
}

func (svc *basicAuthzService) RemovePolicyBySub(ctx context.Context, sub string) error {
	return svc.repo.RemovePolicyBySub(ctx, sub)
}
