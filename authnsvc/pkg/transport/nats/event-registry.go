package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcevent "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/event"
	pe "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/lib/policy-enforcer"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	"github.com/nats-io/nats.go"
)

type EventHandlerFuncs struct {
	EventPolicyUpdatedHandler nats.Handler
}

func getTargetSub(reqChan, svcName string) string {
	return reqChan + "." + svcName
}

func initEventHandlerFuncs(svc service.IAuthNService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventPolicyUpdatedHandler: makeEventPolicyUpdatedHandler(svc),
	}
}

func (ehf *EventHandlerFuncs) GetSubscription(nc *nats.EncodedConn) (subscriptions []*nats.Subscription, err error) {
	var s *nats.Subscription
	var t svcevent.EventInfo
	targetSvc := "authnsvc"

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

	return
}

func makeEventPolicyUpdatedHandler(svc service.IAuthNService) nats.Handler {
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
