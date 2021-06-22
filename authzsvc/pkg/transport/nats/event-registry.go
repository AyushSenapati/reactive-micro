package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcconf "github.com/AyushSenapati/reactive-micro/authzsvc/conf"
	svcevent "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/event"
	cl "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/logger"
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

func initEventHandlerFuncs(logger *cl.CustomLogger, svc service.IAuthzService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventUpsertPolicyHandler:   makeEventUpsertPolicyHandler(logger, svc),
		EventRemovePolicyHandler:   makeEventRemovePolicyHandler(logger, svc),
		EventAccountDeletedHandler: makeEventAccountDeletedHandler(logger, svc),
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

func makeEventUpsertPolicyHandler(logger *cl.CustomLogger, svc service.IAuthzService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventUpsertPolicyPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventUpsertPolicy] err: %v", err))
			return
		}

		json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventUpsertPolicy] err: %v", err))
			return
		}

		err = svc.UpsertPolicy(ctx, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventUpsertPolicy] err: %v", err))
			return
		}
		m.Ack()
	}
}

func makeEventRemovePolicyHandler(logger *cl.CustomLogger, svc service.IAuthzService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventRemovePolicyPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventRemovePolicy] err: %v", err))
			return
		}

		json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventRemovePolicy] err: %v", err))
			return
		}

		err = svc.RemovePolicy(ctx, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventRemovePolicy] err: %v", err))
			return
		}
		m.Ack()
	}
}

func makeEventAccountDeletedHandler(logger *cl.CustomLogger, svc service.IAuthzService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventAccountDeletedPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventAccountDeleted] err: %v", err))
			return
		}

		json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventAccountDeleted] err: %v", err))
			return
		}

		err = svc.RemovePolicyBySub(ctx, fmt.Sprint(p.AccntID))
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventAccountDeleted] err: %v", err))
			return
		}
		m.Ack()
	}
}
