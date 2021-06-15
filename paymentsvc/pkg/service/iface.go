package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/dto"
	svcpe "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/lib/policy-enforcer"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/repo"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Middleware represents service middleware type
type Middleware func(IPaymentService) IPaymentService

type IPaymentService interface {
	// Handlers of the events
	HandleAccountCreatedEvent(ctx context.Context, accntID uint, role string) error
	HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error
	HandleProductReservedEvent(ctx context.Context, oid uuid.UUID, aid uint, payble float32) error

	RechargeWallet(ctx context.Context, aid uint, amount float32) (uuid.UUID, error)
	ListTransactions(ctx context.Context, txids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListTransactionsResponse
}

type basicPaymentService struct {
	repo repo.PaymentRepository
	nc   *nats.EncodedConn
	ps   svcpe.PolicyStorage
}

// NewBasicPaymentService returns a naive, stateless implementation of IPaymentService
func NewBasicPaymentService() *basicPaymentService {
	return &basicPaymentService{}
}

type SvcConf func(*basicPaymentService) error

func WithRepo(r repo.PaymentRepository) SvcConf {
	return func(svc *basicPaymentService) error {
		svc.repo = r
		return nil
	}
}

func WithNATSEncodedConn(nc *nats.EncodedConn) SvcConf {
	return func(svc *basicPaymentService) error {
		if nc == nil {
			return errors.New("nats encoded client not provided")
		}
		svc.nc = nc
		return nil
	}
}

func WithPolicyStorage(ps svcpe.PolicyStorage) SvcConf {
	return func(svc *basicPaymentService) error {
		if ps == nil {
			return errors.New("policy storage obj can't be empty")
		}
		svc.ps = ps
		return nil
	}
}

// New returns a InventoryService implementation with
// all of the expected config/middleware wired in.
func New(mws []Middleware, svcconfs ...SvcConf) IPaymentService {
	svc := NewBasicPaymentService()
	for _, configure := range svcconfs {
		if configure != nil {
			if err := configure(svc); err != nil {
				fmt.Println("svc error:", err)
				return nil
			}
		}
	}

	var s IPaymentService
	s = svc

	var counter int = 0
	for _, m := range mws {
		s = m(s)
		counter++
	}
	fmt.Println("service middlewares configured:", counter)
	return s
}
