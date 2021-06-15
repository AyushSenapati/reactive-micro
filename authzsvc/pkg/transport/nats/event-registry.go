package nats

import (
	"context"
	"encoding/json"
	"fmt"

	// "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/endpoint"
	svcevent "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandlerFuncs struct {
	EventUpsertPolicyHandler   nats.Handler
	EventRemovePolicyHandler   nats.Handler
	EventAccountDeletedHandler nats.Handler
}

func initEventHandlerFuncs(svc service.IAuthzService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventUpsertPolicyHandler:   makeEventUpsertPolicyHandler(svc),
		EventRemovePolicyHandler:   makeEventRemovePolicyHandler(svc),
		EventAccountDeletedHandler: makeEventAccountDeletedHandler(svc),
	}
}

func (er *EventHandlerFuncs) GetSubscription(nc *nats.EncodedConn) (subscriptions []*nats.Subscription, err error) {
	var s *nats.Subscription
	var t svcevent.EventInfo

	t, err = svcevent.Registry.GetEventInfo(svcevent.EventUpsertPolicy)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventUpsertPolicyHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	t, err = svcevent.Registry.GetEventInfo(svcevent.EventRemovePolicy)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventRemovePolicyHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	t, err = svcevent.Registry.GetEventInfo(svcevent.EventAccountDeleted)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(t.ReqChan, er.EventAccountDeletedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	return
}

func makeEventUpsertPolicyHandler(svc service.IAuthzService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventUpsertPolicyPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err := svc.UpsertPolicy(ctx, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		if err != nil {
			fmt.Printf("event handler: error in upsert operation [%v]\n", err)
		}
	}
}

func makeEventRemovePolicyHandler(svc service.IAuthzService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var p svcevent.EventRemovePolicyPayload
		json.Unmarshal(encodedEvent, &p)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err := svc.RemovePolicy(ctx, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		if err != nil {
			fmt.Printf("event handler: error removing policy [%v]\n", err)
		}
	}
}

func makeEventAccountDeletedHandler(svc service.IAuthzService) nats.Handler {
	return func(e *svcevent.Event) {
		encodedEvent, _ := json.Marshal(e.Payload)
		fmt.Println(string(encodedEvent))

		var accnt svcevent.EventAccountDeletedPayload
		json.Unmarshal(encodedEvent, &accnt)

		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err := svc.RemovePolicyBySub(ctx, fmt.Sprint(accnt.AccntID))
		if err != nil {
			fmt.Printf("event handler: error removing policy for account: %v [%v]\n", accnt.AccntID, err)
		}
	}
}
