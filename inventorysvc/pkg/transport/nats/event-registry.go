package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/event"
	pe "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/lib/policy-enforcer"
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

func getTargetSub(reqChan, svcName string) string {
	return reqChan + "." + svcName
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

func (ehf *EventHandlerFuncs) GetSubscription(nc *nats.EncodedConn) (subscriptions []*nats.Subscription, err error) {
	var s *nats.Subscription
	var t svcevent.EventInfo
	targetSvc := "inventorysvc"

	// subscribe to EventAccountCreated
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventAccountCreated)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventAccountCreatedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventPolicyUpdated
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventPolicyUpdated)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventPolicyUpdatedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventOrderCreated
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventOrderCreated)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventOrderCreatedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventOrderApproved
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventOrderApproved)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventOrderApprovedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventOrderCanceled
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventOrderCanceled)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventOrderCanceledHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	return
}

func makeEventAccountCreatedHandler(svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventAccountCreatedPayload
		json.Unmarshal(m.Data, &e)

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			fmt.Println("marshalling event payload err:", err)
			return
		}

		json.Unmarshal(encodedPayload, &p)
		if err != nil {
			fmt.Println("unmarshalling err:", err)
			return
		}

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err = svc.HandleAccountCreatedEvent(ctx, p.AccntID, p.Role)
		if err != nil {
			fmt.Println("event handler [EventAccountCreated] err:", err)
		} else {
			m.Ack()
		}
	}
}

func makeEventPolicyUpdatedHandler(svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventPolicyUpdatedPayload
		json.Unmarshal(m.Data, &e)

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			fmt.Println("marshalling event payload err:", err)
			return
		}

		err = json.Unmarshal(encodedPayload, &p)
		if err != nil {
			fmt.Println("unmarshalling err:", err)
			return
		}

		fmt.Println("payload:", string(m.Data))
		ctx := context.WithValue(context.Background(), "X-Request-ID", "")
		err = svc.HandlePolicyUpdatedEvent(ctx, p.Method, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		fmt.Println("event handler [EventPolicyUpdated] err:", err)
		if (err == nil) || (err == pe.ErrUnsupportedRtype) || (err == pe.ErrSubNotCached) {
			m.Ack() // if no error occurred processing event ack it
		}
	}
}

func makeEventOrderCreatedHandler(svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventOrderCreatedPayload
		json.Unmarshal(m.Data, &e)

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			fmt.Println("marshalling event payload err:", err)
			return
		}

		json.Unmarshal(encodedPayload, &p)
		if err != nil {
			fmt.Println("unmarshalling err:", err)
			return
		}

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err = svc.HandleOrderCreatedEvent(ctx, p.OrderID, p.ProductID, p.OrderStatus, p.Qty, p.AccntID)
		if err != nil {
			fmt.Println("event handler [EventOrderCreated] err:", err)
		} else {
			m.Ack()
		}
	}
}

func makeEventOrderApprovedHandler(svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventOrderApprovedPayload
		json.Unmarshal(m.Data, &e)

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			fmt.Println("marshalling event payload err:", err)
			return
		}

		json.Unmarshal(encodedPayload, &p)
		if err != nil {
			fmt.Println("unmarshalling err:", err)
			return
		}

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err = svc.HandleOrderApprovedEvent(ctx, p.OID)
		if err != nil {
			fmt.Println("event handler [EventPolicyUpdated] err:", err)
		} else {
			m.Ack()
		}
	}
}

func makeEventOrderCanceledHandler(svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventOrderCanceledPayload
		json.Unmarshal(m.Data, &e)

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			fmt.Println("marshalling event payload err:", err)
			return
		}

		json.Unmarshal(encodedPayload, &p)
		if err != nil {
			fmt.Println("unmarshalling err:", err)
			return
		}

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err = svc.HandleOrderCanceledEvent(ctx, p.OID)
		if err != nil {
			fmt.Println("event handler [EventPolicyUpdated] err:", err)
		} else {
			m.Ack()
		}
	}
}
