package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	svcpe "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/lib/policy-enforcer"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/repo"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Middleware represents service middleware type
type Middleware func(IInventoryService) IInventoryService

type IInventoryService interface {
	// Handlers of the events
	HandleAccountCreatedEvent(ctx context.Context, aid uint, role string) error
	HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error
	HandleOrderCreatedEvent(ctx context.Context, oid, pid uuid.UUID, status string, qty int, aid uint) error
	HandleOrderApprovedEvent(ctx context.Context, oid uuid.UUID) error
	HandleOrderCanceledEvent(ctx context.Context, oid uuid.UUID) error

	CreateMerchant(ctx context.Context, aid uint, name string) dto.CreateMerchantResponse
	ListMerchant(ctx context.Context, mids []uuid.UUID) dto.ListMerchantResponse

	CreateProduct(ctx context.Context, aid uint, mid uuid.UUID, name, desc string, qty int, price float32) dto.CreateProductResponse
	ListProduct(ctx context.Context, pids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListProductResponse
}

type basicInventoryService struct {
	repo repo.InventoryRepository
	nc   *nats.EncodedConn
	ps   svcpe.PolicyStorage
}

// NewBasicInventoryService returns a naive, stateless implementation of IInventoryService
func NewBasicInventoryService() *basicInventoryService {
	return &basicInventoryService{}
}

type SvcConf func(*basicInventoryService) error

func WithRepo(r repo.InventoryRepository) SvcConf {
	return func(svc *basicInventoryService) error {
		svc.repo = r
		return nil
	}
}

func WithNATSEncodedConn(nc *nats.EncodedConn) SvcConf {
	return func(svc *basicInventoryService) error {
		if nc == nil {
			return errors.New("nats encoded client not provided")
		}
		svc.nc = nc
		return nil
	}
}

func WithPolicyStorage(ps svcpe.PolicyStorage) SvcConf {
	return func(svc *basicInventoryService) error {
		if ps == nil {
			return errors.New("policy storage obj can't be empty")
		}
		svc.ps = ps
		return nil
	}
}

// New returns a InventoryService implementation with
// all of the expected config/middleware wired in.
func New(mws []Middleware, svcconfs ...SvcConf) IInventoryService {
	svc := NewBasicInventoryService()
	for _, configure := range svcconfs {
		if configure != nil {
			if err := configure(svc); err != nil {
				fmt.Println("svc error:", err)
				return nil
			}
		}
	}

	var s IInventoryService
	s = svc

	var counter int = 0
	for _, m := range mws {
		s = m(s)
		counter++
	}
	fmt.Println("service middlewares configured:", counter)
	return s
}
