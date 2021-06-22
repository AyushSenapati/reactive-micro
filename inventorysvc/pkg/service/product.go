package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	svcevent "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/model"
	"github.com/google/uuid"
)

func (svc *basicInventoryService) CreateProduct(
	ctx context.Context, aid uint, mid uuid.UUID, name, desc string, qty int, price float32) dto.CreateProductResponse {

	pid, err := svc.repo.CreateProduct(ctx, name, desc, mid, qty, price)
	if err != nil {
		return dto.CreateProductResponse{Err: err}
	}

	eventPublisher := svcevent.NewEventPublisher()
	eventErr := eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(aid),
			ResourceType: "products",
			ResourceID:   pid.String(),
			Action:       "*",
		},
	))
	svc.cl.LogIfError(ctx, eventErr)

	eventErr = eventPublisher.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf("published events: %s", svcevent.EventUpsertPolicy))
	}

	return dto.CreateProductResponse{ID: pid, Err: err}
}

func (svc *basicInventoryService) ListProduct(
	ctx context.Context, pids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListProductResponse {

	var prodObjs []model.Product
	var err error

	if len(pids) > 0 {
		prodObjs, err = svc.repo.ListProductByIDs(ctx, pids, qp)
	} else {
		prodObjs, err = svc.repo.ListProduct(ctx, qp)
	}

	if err != nil {
		return dto.ListProductResponse{Err: err}
	}

	var products []dto.ProductDetailsResponse
	for _, po := range prodObjs {
		tmpProd := dto.ProductDetailsResponse{
			ID:       po.ID,
			Name:     po.Name,
			Desc:     po.Desc,
			Price:    po.Price,
			LowStock: new(bool),
		}
		if po.Qty <= 5 {
			*tmpProd.LowStock = true
		} else {
			*tmpProd.LowStock = false
		}
		products = append(products, tmpProd)
	}

	return dto.ListProductResponse{Products: products}
}
