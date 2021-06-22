package service

import (
	"context"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/model"
	"github.com/google/uuid"
)

func (svc *basicOrderService) HandleAccountCreatedEvent(ctx context.Context, accntID uint, role string) error {
	// only customers should have create order permission
	if role != "customer" {
		return nil
	}

	authzEvent, err := svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(accntID),
			ResourceType: "orders",
			ResourceID:   "*",
			Action:       "post",
		})
	if err != nil {
		err = &svcevent.ErrNewEvent{Name: svcevent.EventUpsertPolicy}
		svc.cl.Error(ctx, err)
		return err
	}

	err = authzEvent.Publish(svc.nc)
	svc.cl.LogIfError(ctx, err)
	if err == nil {
		svc.cl.Debug(ctx, fmt.Sprintf("published events: %s", svcevent.EventUpsertPolicy))
	}

	return err
}

func (svc *basicOrderService) HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error {
	return svc.ps.UpdatePolicy(method, sub, rtype, rid, act)
}

func (svc *basicOrderService) HandleErrReservingProductEvent(ctx context.Context, oid uuid.UUID) error {
	return svc.repo.UpdateOrderStatus(ctx, oid, model.OrderStatusProductOutOfStock)
}

func (svc *basicOrderService) HandleProductReservedEvent(ctx context.Context, oid uuid.UUID) error {
	return svc.repo.UpdateOrderStatus(ctx, oid, model.OrderStatusPaymentPending)
}

func (svc *basicOrderService) HandlePaymentEvent(ctx context.Context, oid uuid.UUID, aid uint, status string) error {
	eventPublisher := svcevent.NewEventPublisher()
	var eventErr error

	if status == "payment_successful" {
		err := svc.repo.UpdateOrderStatus(ctx, oid, model.OrderStatusPaid)
		if err != nil {
			svc.cl.Error(ctx, err)
			return err
		}
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventOrderApproved,
			svcevent.EventOrderApprovedPayload{
				OID:     oid,
				AccntID: aid,
			},
		))
	} else {
		err := svc.repo.UpdateOrderStatus(ctx, oid, model.OrderStatusFailed)
		if err != nil {
			svc.cl.Error(ctx, err)
			return err
		}
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventOrderCanceled,
			svcevent.EventOrderCanceledPayload{
				OID:     oid,
				AccntID: aid,
			},
		))
	}

	svc.cl.LogIfError(ctx, eventErr)
	eventErr = eventPublisher.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf("published events: %v", eventPublisher.GetEventNames()))
	}

	return nil
}
