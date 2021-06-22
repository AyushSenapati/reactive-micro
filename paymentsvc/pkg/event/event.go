package event

import (
	"context"
	"fmt"
	"time"

	svcconf "github.com/AyushSenapati/reactive-micro/paymentsvc/conf"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type EventName string

// Registry is the service EventRegistry where all the
// active events which are to be fired or handled must register themselves
var Registry = &EventRegistry{
	Registry: make(map[EventName]EventInfo),
}

// Registry holds all the event name and their transport details mapping
// It helps in verifying the event name and getting their request/response channel
type EventRegistry struct {
	Registry map[EventName]EventInfo
}

type EventInfo struct {
	ReqChan        string
	RespChan       string
	isValidPayload func(interface{}) bool
}

func (er *EventRegistry) register(name EventName, et EventInfo) {
	er.Registry[name] = et
}

func (er *EventRegistry) GetEventInfo(name EventName) (EventInfo, error) {
	t, ok := er.Registry[name]
	if !ok {
		return t, &ErrUnregisteredEvent{Name: name}
	}
	return t, nil
}

type IEvent interface {
	Name() string
	GetPayload() interface{}
	Publish(*nats.EncodedConn) error
}

type Event struct {
	Meta    EventMeta   `json:"meta"`
	Payload interface{} `json:"payload"`
}

func (e *Event) Name() string {
	return e.Meta.Name
}

func (e *Event) GetPayload() interface{} {
	return e.Payload
}

func (e *Event) Publish(nc *nats.EncodedConn) error {
	if nc == nil {
		return ErrNilNATSConnObj
	}
	t, err := Registry.GetEventInfo(EventName(e.Meta.Name))
	if err != nil {
		return err
	}
	if t.ReqChan == "" {
		return &ErrEventReqChNotSet{EventName(e.Meta.Name)}
	}
	return nc.Publish(t.ReqChan, e)
}

type EventMeta struct {
	Version   string    `json:"version"`
	Source    string    `json:"source"`
	Time      time.Time `json:"time"`
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	RequestID string    `json:"req_id"`
}

func getEventMeta(ctx context.Context, name string) EventMeta {
	reqID := ctx.Value(svcconf.C.ReqIDKey)
	if reqID == nil {
		reqID = ""
	}
	return EventMeta{
		Version:   "1.0",
		Source:    svcconf.C.SVCName,
		Time:      time.Now(),
		Name:      name,
		ID:        uuid.New().String(),
		RequestID: reqID.(string),
	}
}

// NewEvent is the factory to generate all the event
func NewEvent(ctx context.Context, name EventName, payload interface{}) (IEvent, error) {
	// check if the event is registered in the registry
	t, err := Registry.GetEventInfo(name)
	if err != nil {
		return nil, err
	}

	if t.isValidPayload != nil {
		ok := t.isValidPayload(payload)
		if !ok {
			return nil, ErrInvalidPayload
		}
	} else {
		return nil, fmt.Errorf("payload checker is not set for event: %s", name)
	}

	e := &Event{Meta: getEventMeta(ctx, string(name)), Payload: payload}

	return e, nil
}

// EventPublisher makes it easy to add events and send them all at once
type EventPublisher struct {
	events []IEvent
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

func (ep *EventPublisher) AddEvent(e IEvent, err error) error {
	if e == nil || err != nil {
		err := fmt.Errorf("event publisher: err adding event [err: %v]", err)
		return err
	}
	ep.events = append(ep.events, e)
	return nil
}

// Publish publishes the added events to registered NATS subjects
// if error occurs while publishing any event, the publisher returns the error
// immediately instead of try publishing other events
func (ep *EventPublisher) Publish(nc *nats.EncodedConn) error {
	for _, e := range ep.events {
		err := e.Publish(nc)
		if err != nil {
			return fmt.Errorf("event publisher: error publishing event: %s [%v]", e.Name(), err)
		}
	}
	return nil
}

func (ep *EventPublisher) GetEventNames() (names []string) {
	for _, e := range ep.events {
		names = append(names, e.Name())
	}
	return
}
