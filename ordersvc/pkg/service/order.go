package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/dto"
	svcevent "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/model"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/google/uuid"
)

func (svc *basicOrderService) CreateOrder(ctx context.Context, pid uuid.UUID, qty int) (uuid.UUID, error) {
	claim := ctx.Value(kitjwt.JWTClaimsContextKey).(*dto.CustomClaim)
	oid, err := svc.repo.CreateOrder(ctx, claim.AccntID, qty, pid, model.OrderStatusPending)
	if err != nil {
		return oid, err
	}

	// on order create fire order created and upsert policy events
	eventPublisher := svcevent.NewEventPublisher()
	eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventOrderCreated,
		svcevent.EventOrderCreatedPayload{
			OrderID:     oid,
			OrderStatus: string(model.OrderStatusPending),
			AccntID:     claim.AccntID,
			ProductID:   pid,
			Qty:         qty,
		}))

	// account must have read permission on newly created order
	eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(claim.AccntID),
			ResourceType: "orders",
			ResourceID:   oid.String(),
			Action:       "get",
		},
	))

	// account must have update permission on newly created order
	eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(claim.AccntID),
			ResourceType: "orders",
			ResourceID:   oid.String(),
			Action:       "put",
		},
	))

	eventPublisher.Publish(svc.nc)

	// dont send the event error to the client as
	// client does not need to know about the event details
	return oid, err
}

func (svc *basicOrderService) ListOrder(ctx context.Context, oids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListOrderResponse {
	var orderObjs []model.Order
	var err error
	if len(oids) > 0 {
		orderObjs, err = svc.repo.ListOrderByIDs(ctx, oids, qp)
	} else {
		orderObjs, err = svc.repo.ListOrder(ctx, qp)
	}

	if err != nil {
		svc.cl.Error(ctx, fmt.Sprintf("err getting orders [%v]", err))
		return dto.ListOrderResponse{Err: err}
	}

	var orders []dto.GetOrderResponse
	for _, o := range orderObjs {
		orders = append(orders, dto.GetOrderResponse{
			OID:      o.ID.String(),
			Status:   o.Status,
			Qty:      o.Qty,
			ProdName: "", // call inventory svc to get product details
		})
	}
	return dto.ListOrderResponse{Orders: orders}
}
