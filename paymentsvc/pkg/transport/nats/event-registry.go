package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandlerFuncs struct {
	EventAccountCreatedHandler  nats.Handler
	EventPolicyUpdatedHandler   nats.Handler
	EventProductReservedHandler nats.Handler
}

func initEventHandlerFuncs(svc service.IPaymentService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventAccountCreatedHandler:  makeEventAccountCreatedHandler(svc),
		EventPolicyUpdatedHandler:   makeEventPolicyUpdatedHandler(svc),
		EventProductReservedHandler: makeEventProductReservedHandler(svc),
	}
}

func (er *EventHandlerFuncs) GetSubscription(nc *nats.EncodedConn) (subscriptions []*nats.Subscription, err error) {
	var s *nats.Subscription
	var t svcevent.EventInfo

	// subscribe to EventAccountCreated
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventAccountCreated)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventAccountCreatedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventPolicyUpdated
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventPolicyUpdated)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventPolicyUpdatedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventProductReserved
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventProductReserved)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventProductReservedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	return
}

func makeEventAccountCreatedHandler(svc service.IPaymentService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventAccountCreatedPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventAccountCreated] err:",
			svc.HandleAccountCreatedEvent(ctx, p.AccntID, p.Role))
	}
}

func makeEventPolicyUpdatedHandler(svc service.IPaymentService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventPolicyUpdatedPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventPolicyUpdated] err:",
			svc.HandlePolicyUpdatedEvent(ctx, p.Method, p.Sub, p.ResourceType, p.ResourceID, p.Action))
	}
}

func makeEventProductReservedHandler(svc service.IPaymentService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventProductReservedPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventProductReserved] err:",
			svc.HandleProductReservedEvent(ctx, p.OrderID, p.AccntID, float32(p.Payble)))
	}
}
