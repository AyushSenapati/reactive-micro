package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/dto"
	svcpe "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/lib/policy-enforcer"
	cl "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/logger"
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/repo"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Middleware represents service middleware type
type Middleware func(IOrderService) IOrderService

type IOrderService interface {
	// Handlers of the events
	HandleAccountCreatedEvent(ctx context.Context, accntID uint, role string) error
	HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error
	HandleErrReservingProductEvent(ctx context.Context, oid uuid.UUID) error
	HandleProductReservedEvent(ctx context.Context, oid uuid.UUID) error
	HandlePaymentEvent(ctx context.Context, oid uuid.UUID, aid uint, status string) error

	ListOrder(ctx context.Context, oids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListOrderResponse
	CreateOrder(ctx context.Context, pid uuid.UUID, qty int) (uuid.UUID, error)
}

type basicOrderService struct {
	cl   *cl.CustomLogger
	repo repo.OrderRepository
	nc   *nats.EncodedConn
	ps   svcpe.PolicyStorage
}

// NewBasicOrderService returns a naive, stateless implementation of OrderService
func NewBasicOrderService() *basicOrderService {
	return &basicOrderService{}
}

type SvcConf func(*basicOrderService) error

func WithRepo(r repo.OrderRepository) SvcConf {
	return func(svc *basicOrderService) error {
		svc.repo = r
		return nil
	}
}

func WithNATSEncodedConn(nc *nats.EncodedConn) SvcConf {
	return func(svc *basicOrderService) error {
		if nc == nil {
			return errors.New("nats encoded client not provided")
		}
		svc.nc = nc
		return nil
	}
}

func WithPolicyStorage(ps svcpe.PolicyStorage) SvcConf {
	return func(svc *basicOrderService) error {
		if ps == nil {
			return errors.New("policy storage obj can't be empty")
		}
		svc.ps = ps
		return nil
	}
}

// New returns a OrderService implementation with
// all of the expected config/middleware wired in.
func New(logger *cl.CustomLogger, mws []Middleware, svcconfs ...SvcConf) IOrderService {
	svc := NewBasicOrderService()
	svc.cl = logger
	for _, configure := range svcconfs {
		if configure != nil {
			if err := configure(svc); err != nil {
				logger.Error(context.TODO(), fmt.Sprintf("svc err: %v", err))
				return nil
			}
		}
	}

	var s IOrderService
	s = svc

	var counter int = 0
	for _, m := range mws {
		s = m(s)
		counter++
	}

	return s
}
