package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandlerFuncs struct {
	EventAccountCreatedHandler nats.Handler
	EventPolicyUpdatedHandler  nats.Handler
	EventOrderCreatedHandler   nats.Handler
	EventOrderApprovedHandler  nats.Handler
	EventOrderCanceledHandler  nats.Handler
}

func initEventHandlerFuncs(svc service.IInventoryService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventAccountCreatedHandler: makeEventAccountCreatedHandler(svc),
		EventPolicyUpdatedHandler:  makeEventPolicyUpdatedHandler(svc),
		EventOrderCreatedHandler:   makeEventOrderCreatedHandler(svc),
		EventOrderApprovedHandler:  makeEventOrderApprovedHandler(svc),
		EventOrderCanceledHandler:  makeEventOrderCanceledHandler(svc),
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

	// subscribe to EventOrderCreated
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventOrderCreated)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventOrderCreatedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventOrderApproved
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventOrderApproved)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventOrderApprovedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventOrderCanceled
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventOrderCanceled)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventOrderCanceledHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	return
}

func makeEventAccountCreatedHandler(svc service.IInventoryService) nats.Handler {
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

func makeEventPolicyUpdatedHandler(svc service.IInventoryService) nats.Handler {
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

func makeEventOrderCreatedHandler(svc service.IInventoryService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventOrderCreatedPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventOrderCreated] err:",
			svc.HandleOrderCreatedEvent(ctx, p.OrderID, p.ProductID, p.OrderStatus, p.Qty, p.AccntID))
	}
}

func makeEventOrderApprovedHandler(svc service.IInventoryService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventOrderApprovedPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventPolicyUpdated] err:",
			svc.HandleOrderApprovedEvent(ctx, p.OID))
	}
}

func makeEventOrderCanceledHandler(svc service.IInventoryService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventOrderCanceledPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		fmt.Println(
			"event handler [EventPolicyUpdated] err:",
			svc.HandleOrderCanceledEvent(ctx, p.OID))
	}
}
