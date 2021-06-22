package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/dto"
	cl "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/logger"
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/repo"
	"github.com/nats-io/nats.go"
)

type IAuthzService interface {
	UpsertPolicy(ctx context.Context, sub, resourceType, resourceID, action string) error
	ListPolicy(ctx context.Context, reqObj dto.ListPolicyRequest) dto.ListPolicyResponse
	RemovePolicy(ctx context.Context, sub, resourceType, resourceID, action string) error
	RemovePolicyBySub(ctx context.Context, sub string) error
}

type basicAuthzService struct {
	cl   *cl.CustomLogger
	repo repo.AuthzRepo
	nc   *nats.EncodedConn
}

// NewBasicAuthzService returns a naive, stateless implementation of AuthzService
func NewBasicAuthzService() *basicAuthzService {
	return &basicAuthzService{}
}

type SvcConf func(*basicAuthzService) error

func WithRepo(r repo.AuthzRepo) SvcConf {
	return func(svc *basicAuthzService) error {
		svc.repo = r
		return nil
	}
}

func WithNATSEncodedConn(nc *nats.EncodedConn) SvcConf {
	return func(svc *basicAuthzService) error {
		if nc == nil {
			return errors.New("nats encoded client not provided")
		}
		svc.nc = nc
		return nil
	}
}

// New returns a AuthzService implementation with
// all of the expected config/middleware wired in.
func New(logger *cl.CustomLogger, svcconfs ...SvcConf) IAuthzService {
	svc := NewBasicAuthzService()
	svc.cl = logger
	for _, configure := range svcconfs {
		if configure != nil {
			if err := configure(svc); err != nil {
				logger.Error(context.TODO(), fmt.Sprintf("svc err: %v", err))
				return nil
			}
		}
	}

	return svc
}
