package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	svcpe "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/lib/policy-enforcer"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/repo"
	"github.com/nats-io/nats.go"
)

// Middleware represents service middleware type
type Middleware func(IAuthNService) IAuthNService

type IAuthNService interface {
	// Handlers of the events
	HandlePolicyUpdatedEvent(ctx context.Context, method, sub, rtype, rid, act string) error

	// auth service methods
	GenToken(ctx context.Context, accnt dto.LoginRequest) dto.LoginResponse

	// account service methods
	CreateAccount(ctx context.Context, accnt dto.CreateAccountRequest) dto.CreateAccountResponse
	ListAccount(ctx context.Context, aids []uint, qp *dto.BasicQueryParam) dto.ListAccountResponse
	DeleteAccount(ctx context.Context, aid uint) (err error)
}

type basicAuthNService struct {
	accntrepo repo.UserRepository
	authnrepo repo.AuthNRepository
	nc        *nats.EncodedConn
	ps        svcpe.PolicyStorage
}

// NewBasicAuthNService returns a naive, stateless implementation of AuthNService.
func NewBasicAuthNService() *basicAuthNService {
	return &basicAuthNService{}
}

type SvcConf func(*basicAuthNService) error

func WithRepo(r repo.UserRepository) SvcConf {
	return func(svc *basicAuthNService) error {
		svc.accntrepo = r
		return nil
	}
}

func WithNATSEncodedConn(nc *nats.EncodedConn) SvcConf {
	return func(svc *basicAuthNService) error {
		if nc == nil {
			return errors.New("nats encoded client not provided")
		}
		svc.nc = nc
		return nil
	}
}

func WithPolicyStorage(ps svcpe.PolicyStorage) SvcConf {
	return func(svc *basicAuthNService) error {
		if ps == nil {
			return errors.New("policy storage obj can't be empty")
		}
		svc.ps = ps
		return nil
	}
}

// New returns a Authn service implementation with all of the expected middlewares wired in.
func New(mws []Middleware, SvcConfs ...SvcConf) IAuthNService {
	svc := NewBasicAuthNService()
	for _, configure := range SvcConfs {
		if configure != nil {
			if err := configure(svc); err != nil {
				fmt.Println("svc error:", err)
				return nil
			}
		}
	}

	var s IAuthNService
	s = svc
	var counter int = 0
	for _, m := range mws {
		s = m(s)
		counter++
	}
	fmt.Println("service middlewares configured:", counter)
	return s
}
