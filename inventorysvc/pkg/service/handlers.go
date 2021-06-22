package service

import (
	"context"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/event"
	"github.com/google/uuid"
)

func (svc *basicInventoryService) HandleAccountCreatedEvent(ctx context.Context, aid uint, role string) error {
	eventPublisher := svcevent.NewEventPublisher()
	var eventErr error

	if role == "customer" {
		// customer can list all the products
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventUpsertPolicy,
			svcevent.EventUpsertPolicyPayload{
				Sub:          fmt.Sprint(aid),
				ResourceType: "products",
				ResourceID:   "*",
				Action:       "get",
			},
		))
	} else if role == "seller" {
		// seller can register a merchant
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventUpsertPolicy,
			svcevent.EventUpsertPolicyPayload{
				Sub:          fmt.Sprint(aid),
				ResourceType: "merchants",
				ResourceID:   "*",
				Action:       "post",
			},
		))
	}

	svc.cl.LogIfError(ctx, eventErr)
	eventErr = eventPublisher.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf(
			"published events: %v", eventPublisher.GetEventNames()))
	}

	return nil
}

func (svc *basicInventoryService) HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error {
	return svc.ps.UpdatePolicy(method, sub, rtype, rid, act)
}

func (svc *basicInventoryService) HandleOrderCreatedEvent(ctx context.Context, oid, pid uuid.UUID, status string, qty int, aid uint) error {
	price, err := svc.repo.ReserveProduct(ctx, oid, pid, qty)

	eventPublisher := svcevent.NewEventPublisher()
	var eventErr error

	if err != nil {
		// if there was error in reserving specified product quantity,
		// fire EventErrReservingProduct event
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventErrReservingProduct,
			svcevent.EventErrReservingProductPayload{
				OrderID: oid,
			},
		))
	} else {
		// if products were reserved successfully, fire EventProductReserved event
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventProductReserved,
			svcevent.EventProductReservedPayload{
				OrderID: oid,
				AccntID: aid,
				Payble:  price,
			},
		))
	}

	svc.cl.LogIfError(ctx, eventErr)
	eventErr = eventPublisher.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf(
			"published events: %v", eventPublisher.GetEventNames()))
	}

	return err
}

func (svc *basicInventoryService) HandleOrderApprovedEvent(ctx context.Context, oid uuid.UUID) error {
	// if order is approved then just remove the reserved product
	return svc.repo.RemoveReservedProduct(ctx, oid)
}

func (svc *basicInventoryService) HandleOrderCanceledEvent(ctx context.Context, oid uuid.UUID) error {
	// if for any reason order is canceled/failed undo the reserve product
	// operation by adding the reserved product qty back to the products
	return svc.repo.UndoReserveProduct(ctx, oid)
}
