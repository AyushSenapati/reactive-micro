package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/ordersvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandlerFuncs struct {
	EventAccountCreatedHandler      nats.Handler
	EventPolicyUpdatedHandler       nats.Handler
	EventErrReservingProductHandler nats.Handler
	EventProductReservedHandler     nats.Handler
	EventPaymentHandler             nats.Handler
}

func initEventHandlerFuncs(svc service.IOrderService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventAccountCreatedHandler:      makeEventAccountCreatedHandler(svc),
		EventPolicyUpdatedHandler:       makeEventPolicyUpdatedHandler(svc),
		EventErrReservingProductHandler: makeEventErrReservingProductHandler(svc),
		EventProductReservedHandler:     makeEventProductReservedHandler(svc),
		EventPaymentHandler:             makeEventPaymentHandler(svc),
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

	// subscribe to EventErrReservingProduct
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventErrReservingProduct)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventErrReservingProductHandler)
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

	// subscribe to EventPayment
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventPayment)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventPaymentHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	return
}

func makeEventAccountCreatedHandler(svc service.IOrderService) nats.Handler {
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

func makeEventPolicyUpdatedHandler(svc service.IOrderService) nats.Handler {
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

func makeEventErrReservingProductHandler(svc service.IOrderService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventErrReservingProductPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventErrReservingProduct] err:",
			svc.HandleErrReservingProductEvent(ctx, p.OrderID))
	}
}

func makeEventProductReservedHandler(svc service.IOrderService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventProductReservedPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventProductReserved] err:",
			svc.HandleProductReservedEvent(ctx, p.OrderID))
	}
}

func makeEventPaymentHandler(svc service.IOrderService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventPaymentPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventPayment] err:",
			svc.HandlePaymentEvent(ctx, p.OrderID, p.AccntID, p.Status))
	}
}
