package service

import (
	"context"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/event"
	"github.com/google/uuid"
)

func (svc *basicPaymentService) HandleAccountCreatedEvent(ctx context.Context, accntID uint, role string) error {
	err := svc.repo.EnableWallet(ctx, accntID, 100.0)
	if err != nil {
		return err
	}

	eventPublisher := svcevent.NewEventPublisher()
	if role != "customer" {
		return nil
	}

	// customer can do transactions
	eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(accntID),
			ResourceType: "transactions",
			ResourceID:   "*",
			Action:       "post",
		},
	))

	eventPublisher.Publish(svc.nc)

	return nil
}

func (svc *basicPaymentService) HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error {
	return svc.ps.UpdatePolicy(method, sub, rtype, rid, act)
}

func (svc *basicPaymentService) HandleProductReservedEvent(ctx context.Context, oid uuid.UUID, aid uint, payble float32) error {
	txid, err := svc.repo.ExecuteTX(ctx, aid, payble, false)

	// time.Sleep(10 * time.Second)
	eventPublisher := svcevent.NewEventPublisher()
	eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(aid),
			ResourceType: "transactions",
			ResourceID:   txid.String(),
			Action:       "get",
		},
	))

	if err != nil {
		eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventPayment,
			svcevent.EventPaymentPayload{
				OrderID: oid,
				AccntID: aid,
				Status:  "payment_failed",
			},
		))
	} else {
		eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventPayment,
			svcevent.EventPaymentPayload{
				OrderID: oid,
				AccntID: aid,
				Status:  "payment_successful",
			},
		))
	}

	eventPublisher.Publish(svc.nc)

	return err
}