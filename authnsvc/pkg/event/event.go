package event

import (
	"context"
	"fmt"
	"time"

	svcconf "github.com/AyushSenapati/reactive-micro/authnsvc/conf"
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
	reqID := ctx.Value("X-Request-ID")
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
		fmt.Println("payload checker is not set for event:", name)
	}

	e := &Event{Meta: getEventMeta(ctx, string(name)), Payload: payload}

	// return &Event{Meta: meta, Payload: payload}, err
	return e, nil
}

// EventPublisher makes it easy to add events and send them all at once
type EventPublisher struct {
	events []IEvent
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

func (ep *EventPublisher) AddEvent(e IEvent, err error) {
	if e == nil || err != nil {
		fmt.Printf("error adding event to the publisher [err: %v]\n", err)
		return
	}
	ep.events = append(ep.events, e)
}

func (ep *EventPublisher) Publish(nc *nats.EncodedConn) {
	for _, e := range ep.events {
		err := e.Publish(nc)
		if err != nil {
			fmt.Printf("error publishing event: %s [%v]\n", e.Name(), err)
		} else {
			fmt.Printf("published: %s payload: %v\n", e.Name(), e.GetPayload())
		}
	}
}
