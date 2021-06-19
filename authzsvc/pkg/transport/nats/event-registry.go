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

func getTargetSub(reqChan, svcName string) string {
	return reqChan + "." + svcName
}

func initEventHandlerFuncs(svc service.IAuthzService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventUpsertPolicyHandler:   makeEventUpsertPolicyHandler(svc),
		EventRemovePolicyHandler:   makeEventRemovePolicyHandler(svc),
		EventAccountDeletedHandler: makeEventAccountDeletedHandler(svc),
	}
}

func (ehf *EventHandlerFuncs) GetSubscription(nc *nats.EncodedConn) (subscriptions []*nats.Subscription, err error) {
	var s *nats.Subscription
	var t svcevent.EventInfo
	targetSvc := "authzsvc"

	// subscribe to EventUpsertPolicy
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventUpsertPolicy)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventUpsertPolicyHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventRemovePolicy
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventRemovePolicy)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventRemovePolicyHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	// subscribe to EventAccountDeleted
	t, err = svcevent.Registry.GetEventInfo(svcevent.EventAccountDeleted)
	if err != nil {
		return
	}
	s, err = nc.Subscribe(getTargetSub(t.ReqChan, targetSvc), ehf.EventAccountDeletedHandler)
	if err != nil {
		return
	}
	subscriptions = append(subscriptions, s)

	return
}

func makeEventUpsertPolicyHandler(svc service.IAuthzService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventUpsertPolicyPayload
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

		fmt.Println("payload:", string(m.Data))
		ctx := context.WithValue(context.Background(), "X-Request-ID", e.Meta.RequestID)
		err = svc.UpsertPolicy(ctx, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		if err != nil {
			fmt.Printf("event handler: error in upsert operation [%v]\n", err)
		} else {
			m.Ack()
		}
	}
}

func makeEventRemovePolicyHandler(svc service.IAuthzService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventRemovePolicyPayload
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
		err = svc.RemovePolicy(ctx, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		if err != nil {
			fmt.Printf("event handler: error removing policy [%v]\n", err)
		} else {
			m.Ack()
		}
	}
}

func makeEventAccountDeletedHandler(svc service.IAuthzService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventAccountDeletedPayload
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
		err = svc.RemovePolicyBySub(ctx, fmt.Sprint(p.AccntID))
		if err != nil {
			fmt.Printf("event handler: error removing policy for account: %v [%v]\n", p.AccntID, err)
		} else {
			m.Ack()
		}
	}
}
