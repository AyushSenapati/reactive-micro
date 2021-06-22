package service

import (
	"context"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/repo"
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
	var eventErr error

	if err != nil {
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventPayment,
			svcevent.EventPaymentPayload{
				OrderID: oid,
				AccntID: aid,
				Status:  "payment_failed",
			},
		))
		svc.cl.LogIfError(ctx, eventErr)
	} else {
		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventUpsertPolicy,
			svcevent.EventUpsertPolicyPayload{
				Sub:          fmt.Sprint(aid),
				ResourceType: "transactions",
				ResourceID:   txid.String(),
				Action:       "get",
			},
		))
		svc.cl.LogIfError(ctx, eventErr)

		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventPayment,
			svcevent.EventPaymentPayload{
				OrderID: oid,
				AccntID: aid,
				Status:  "payment_successful",
			},
		))
		svc.cl.LogIfError(ctx, eventErr)
	}

	eventErr = eventPublisher.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf("published events: %v", eventPublisher.GetEventNames()))
	}

	// set err to nil, so that event handler would not consider this err
	// as application error which would lead event handler to retry EventProductReserved
	if err == repo.ErrInsufficientBalance {
		err = nil
	}

	return err
}
