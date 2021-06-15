package nats

import (
	"errors"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandler struct {
	nc           *nats.EncodedConn
	handlers     *EventHandlerFuncs
	subcriptions []*nats.Subscription
	cancel       chan struct{}
}

func NewEventHandler(nc *nats.EncodedConn, svc service.IAuthNService) *EventHandler {
	return &EventHandler{
		nc:           nc,
		handlers:     initEventHandlerFuncs(svc),
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

	fmt.Println("event handler: initialised")
	<-eh.cancel
	fmt.Println("event handler: closed")
	return nil
}

func (eh *EventHandler) Interrupt(err error) {
	fmt.Println("event handler: cleanup started")
	eh.nc.Close()
	close(eh.cancel)
	for _, s := range eh.subcriptions {
		s.Unsubscribe()
	}
	fmt.Println("event handler: cleanup completed")
}
