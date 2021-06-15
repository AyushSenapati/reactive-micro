package event

import (
	"errors"
	"fmt"
)

// Errors of event package
var (
	ErrNilNATSConnObj   = errors.New("nil nats conn obj received")
	ErrInvalidPayload   = errors.New("invalid payload")
	ErrUnsupportedEvent = errors.New("unsupported event")
)

type ErrUnregisteredEvent struct {
	Name EventName
}

func (e *ErrUnregisteredEvent) Error() string {
	return fmt.Sprintf("event store entry not found for event: %s", e.Name)
}

type ErrEventReqChNotSet struct {
	Name EventName
}

func (e *ErrEventReqChNotSet) Error() string {
	return fmt.Sprintf("request channel is not set for event: %s", e.Name)
}

type ErrNewEvent struct {
	Name EventName
}

func (e *ErrNewEvent) Error() string {
	return fmt.Sprintf("error creating event: %s", e.Name)
}

type ErrNilVerifyFunc struct {
	Name EventName
}

func (e *ErrNilVerifyFunc) Error() string {
	return fmt.Sprintf("verify func is not provided for event: %s", e.Name)
}
