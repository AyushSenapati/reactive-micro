package nats

import (
	"context"
	"encoding/json"
	"fmt"

	svcconf "github.com/AyushSenapati/reactive-micro/inventorysvc/conf"
	svcevent "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/event"
	pe "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/lib/policy-enforcer"
	cl "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/logger"
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

func initEventHandlerFuncs(logger *cl.CustomLogger, svc service.IInventoryService) *EventHandlerFuncs {
	return &EventHandlerFuncs{
		EventAccountCreatedHandler: makeEventAccountCreatedHandler(logger, svc),
		EventPolicyUpdatedHandler:  makeEventPolicyUpdatedHandler(logger, svc),
		EventOrderCreatedHandler:   makeEventOrderCreatedHandler(logger, svc),
		EventOrderApprovedHandler:  makeEventOrderApprovedHandler(logger, svc),
		EventOrderCanceledHandler:  makeEventOrderCanceledHandler(logger, svc),
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

func makeEventAccountCreatedHandler(logger *cl.CustomLogger, svc service.IInventoryService) nats.Handler {
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

func makeEventPolicyUpdatedHandler(logger *cl.CustomLogger, svc service.IInventoryService) nats.Handler {
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

func makeEventOrderCreatedHandler(logger *cl.CustomLogger, svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventOrderCreatedPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderCreated] err: %v", err))
			return
		}

		err = json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderCreated] err: %v", err))
			return
		}

		err = svc.HandleOrderCreatedEvent(ctx, p.OrderID, p.ProductID, p.OrderStatus, p.Qty, p.AccntID)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderCreated] err: %v", err))
			return
		}
		m.Ack()
	}
}

func makeEventOrderApprovedHandler(logger *cl.CustomLogger, svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventOrderApprovedPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderApproved] err: %v", err))
			return
		}

		err = json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderApproved] err: %v", err))
			return
		}

		err = svc.HandleOrderApprovedEvent(ctx, p.OID)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderApproved] err: %v", err))
			return
		}
		m.Ack()
	}
}

func makeEventOrderCanceledHandler(logger *cl.CustomLogger, svc service.IInventoryService) nats.Handler {
	return func(m *nats.Msg) {
		var e svcevent.Event
		var p svcevent.EventOrderCanceledPayload

		json.Unmarshal(m.Data, &e)
		ctx := context.WithValue(context.Background(), svcconf.C.ReqIDKey, e.Meta.RequestID)
		logger.Debug(ctx, fmt.Sprintf("event info: %s", string(m.Data)))

		encodedPayload, err := json.Marshal(e.Payload)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderCanceled] err: %v", err))
			return
		}

		err = json.Unmarshal(encodedPayload, &p)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderCanceled] err: %v", err))
			return
		}

		err = svc.HandleOrderCanceledEvent(ctx, p.OID)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("event handler [EventOrderCanceled] err: %v", err))
			return
		}
		m.Ack()
	}
}
