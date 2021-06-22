package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	svcevent "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/model"
	"github.com/google/uuid"
)

func (svc *basicInventoryService) CreateMerchant(ctx context.Context, aid uint, name string) dto.CreateMerchantResponse {
	mid, err := svc.repo.CreateMerchant(ctx, name, aid)
	if err != nil {
		return dto.CreateMerchantResponse{Err: err}
	}

	// if merchant was registered successfully assign it required permissions
	eventPublisher := svcevent.NewEventPublisher()
	eventErr := eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(aid),
			ResourceType: "merchants",
			ResourceID:   mid.String(),
			Action:       "*",
		},
	))
	svc.cl.LogIfError(ctx, eventErr)

	eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(aid),
			ResourceType: "products",
			ResourceID:   "*",
			Action:       "post",
		},
	))
	svc.cl.LogIfError(ctx, eventErr)

	eventErr = eventPublisher.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf(
			"published events: %v", eventPublisher.GetEventNames()))
	}

	return dto.CreateMerchantResponse{ID: mid, Err: err}
}

func (svc *basicInventoryService) ListMerchant(ctx context.Context, mids []uuid.UUID) dto.ListMerchantResponse {
	var merchantObjs []model.Merchant
	var err error

	if len(mids) > 0 {
		merchantObjs, err = svc.repo.ListMerchantByIDs(ctx, mids)
	} else {
		merchantObjs, err = svc.repo.ListMerchant(ctx)
	}

	if err != nil {
		return dto.ListMerchantResponse{Err: err}
	}

	var merchants []dto.MerchantDetailResponse
	for _, mo := range merchantObjs {
		merchants = append(merchants, dto.MerchantDetailResponse{
			ID:   mo.ID,
			Name: mo.Name,
		})
	}

	return dto.ListMerchantResponse{Merchants: merchants}
}
