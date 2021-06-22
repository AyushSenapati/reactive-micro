package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcconf "github.com/AyushSenapati/reactive-micro/paymentsvc/conf"
	svcevent "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/event"
	pe "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/lib/policy-enforcer"
	cl "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/logger"
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

func initEventHandlerFuncs(logger *cl.CustomLogger, svc service.IPaymentService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventAccountCreatedHandler:  makeEventAccountCreatedHandler(logger, svc),
		EventPolicyUpdatedHandler:   makeEventPolicyUpdatedHandler(logger, svc),
		EventProductReservedHandler: makeEventProductReservedHandler(logger, svc),
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

func makeEventAccountCreatedHandler(logger *cl.CustomLogger, svc service.IPaymentService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventAccountCreatedPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventAccountCreated] err: %v", err))
			return
		}

		err = json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventAccountCreated] err: %v", err))
			return
		}

		err = svc.HandleAccountCreatedEvent(ctx, p.AccntID, p.Role)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventAccountCreated] err: %v", err))
			return
		}
		m.Ack()
	}
}

func makeEventPolicyUpdatedHandler(logger *cl.CustomLogger, svc service.IPaymentService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventPolicyUpdatedPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventPolicyUpdated] err: %v", err))
			return
		}

		err = json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventPolicyUpdated] err: %v", err))
			return
		}

		err = svc.HandlePolicyUpdatedEvent(ctx, p.Method, p.Sub, p.ResourceType, p.ResourceID, p.Action)
		if (err == nil) || (err == pe.ErrUnsupportedRtype) || (err == pe.ErrSubNotCached) {
			m.Ack() // if no error occurred processing event ack it
			return
		}
		logger.Error(ctx, fmt.Sprintf("event handler [EventPolicyUpdated] err: %v", err))
	}
}

func makeEventProductReservedHandler(logger *cl.CustomLogger, svc service.IPaymentService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventProductReservedPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventProductReserved] err: %v", err))
			return
		}

		err = json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventProductReserved] err: %v", err))
			return
		}

		err = svc.HandleProductReservedEvent(ctx, p.OrderID, p.AccntID, float32(p.Payble))
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventProductReserved] err: %v", err))
			return
		}
		m.Ack()
	}
}
