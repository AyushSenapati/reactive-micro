package nats

import (
	"context"
	"errors"

	cl "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/logger"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandler struct {
	cl           *cl.CustomLogger
	nc           *nats.EncodedConn
	handlers     *EventHandlerFuncs
	subcriptions []*nats.Subscription
	cancel       chan struct{}
}

func NewEventHandler(logger *cl.CustomLogger, nc *nats.EncodedConn, svc service.IAuthNService) *EventHandler {
	return &EventHandler{
		cl:           logger,
		nc:           nc,
		handlers:     initEventHandlerFuncs(logger, svc),
		subcriptions: []*nats.Subscription{},
		cancel:       make(chan struct{}),
	}
}

func (eh *EventHandler) Execute() error {
	if eh.nc == nil {
		return errors.New("transport [nats]: no connection obj")
	}
	s, err := eh.handlers.GetSubscription(eh.nc)
	if err != nil {
		return err
	}
	eh.subcriptions = s

	eh.cl.Info(context.TODO(), "event handler: initialised")
	<-eh.cancel
	eh.cl.Info(context.TODO(), "event handler: closed")
	return nil
}

func (eh *EventHandler) Interrupt(err error) {
	eh.cl.Info(context.TODO(), "event handler: cleanup started")
	close(eh.cancel)
	for _, s := range eh.subcriptions {
		s.Unsubscribe()
	}
	eh.nc.Close()
	eh.cl.Info(context.TODO(), "event handler: cleanup completed")
}
