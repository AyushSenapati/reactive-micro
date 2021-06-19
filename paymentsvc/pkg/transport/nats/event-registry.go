package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/event"
	pe "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/lib/policy-enforcer"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandlerFuncs struct {
	EventAccountCreatedHandler  nats.Handler
	EventPolicyUpdatedHandler   nats.Handler
	EventProductReservedHandler nats.Handler
}

func getTargetSub(reqChan, svcName string) string {
	return reqChan + "." + svcName
}

func initEventHandlerFuncs(svc service.IPaymentService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventAccountCreatedHandler:  makeEventAccountCreatedHandler(svc),
		EventPolicyUpdatedHandler:   makeEventPolicyUpdatedHandler(svc),
		EventProductReservedHandler: makeEventProductReservedHandler(svc),
	}
}

func (ehf *EventHandlerFuncs) GetSubscription(nc *nats.EncodedConn) (subscriptions []*nats.Subscription, err error) {
	var s *nats.Subscription
	var t svcevent.EventInfo
	targetSvc := "paymentsvc"

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

	// subscribe to EventProductReserved
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventProductReserved)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventProductReservedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	return
}

func makeEventAccountCreatedHandler(svc service.IPaymentService) nats.Handler {
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

func makeEventPolicyUpdatedHandler(svc service.IPaymentService) nats.Handler {
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

func makeEventProductReservedHandler(svc service.IPaymentService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventProductReservedPayload
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
		err = svc.HandleProductReservedEvent(ctx, p.OrderID, p.AccntID, float32(p.Payble))
		if err != nil {
			fmt.Println("event handler [EventProductReserved] err:", err)
		} else {
			m.Ack()
		}
	}
}
