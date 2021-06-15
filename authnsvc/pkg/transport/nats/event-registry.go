package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandlerFuncs struct {
	EventPolicyUpdatedHandler nats.Handler
}

func initEventHandlerFuncs(svc service.IAuthNService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventPolicyUpdatedHandler: makeEventPolicyUpdatedHandler(svc),
	}
}

func (er *EventHandlerFuncs) GetSubscription(nc *nats.EncodedConn) (subscriptions []*nats.Subscription, err error) {
	var s *nats.Subscription
	var t svcevent.EventInfo

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

	return
}

func makeEventPolicyUpdatedHandler(svc service.IAuthNService) nats.Handler {
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
