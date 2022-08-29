package bus

import (
	"fmt"

	"gitlab.bit9.local/octarine/cbctl/internal"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
)

// EventType is the type of an event message.
type EventType string

// All the event types for the bus.
const (
	NewVersionAvailable            EventType = "new-version-event"
	NewMessageDetected             EventType = "new-message-event"
	NewErrorDetected               EventType = "new-error-event"
	PullDockerImage                EventType = "pull-docker-image-event"
	CopyImage                      EventType = "copy-image-event"
	FetchImage                     EventType = "fetch-image-event"
	ReadImage                      EventType = "read-image-event"
	ReadLayer                      EventType = "read-layer-event"
	CatalogerStarted               EventType = "syft-cataloger-started-event"
	CatalogerFinished              EventType = "syft-cataloger-finished-event"
	ScanStarted                    EventType = "image-scanning-started-event"
	ScanFinished                   EventType = "image-scanning-finished-event"
	PrintSBOM                      EventType = "print-sbom-event"
	ValidateFinishedWithViolations EventType = "validate-finished-with-violations"
	ValidateFinishedSuccessfully   EventType = "validate-finished-successfully"
)

// Event is the interface for the message in the bus.
type Event interface {
	Type() EventType
	Value() interface{}
	IsEnd() bool
}

// NewEvent returns a base event.
func NewEvent(eventType EventType, value interface{}, isEnd bool) Event {
	return newBaseEvent(eventType, value, isEnd)
}

// NewVersionEvent returns a NewVersionAvailable type event.
func NewVersionEvent(version string) Event {
	msg := fmt.Sprintf("New version of %s is available: %s", internal.ApplicationName, version)
	return newBaseEvent(NewVersionAvailable, msg, false)
}

// NewMessageEvent returns a NewMessageDetected type event.
func NewMessageEvent(msg string, isEnd bool) Event {
	return newBaseEvent(NewMessageDetected, msg, isEnd)
}

// baseEvent wraps the type and value that define an Event.
type baseEvent struct {
	eventType EventType
	value     interface{}
	isEnd     bool
}

func newBaseEvent(eventType EventType, value interface{}, isEnd bool) *baseEvent {
	return &baseEvent{
		eventType: eventType,
		value:     value,
		isEnd:     isEnd,
	}
}

// Type returns the type of the event.
func (b baseEvent) Type() EventType {
	return b.eventType
}

// Value returns the value of the event.
func (b baseEvent) Value() interface{} {
	return b.value
}

// IsEnd denotes if this event is the last event.
func (b baseEvent) IsEnd() bool {
	return b.isEnd
}

// ErrorEvent is the NewErrorDetected type event.
type ErrorEvent struct {
	Event
	err error
}

// NewErrorEvent returns a NewErrorDetected type event.
func NewErrorEvent(err error) *ErrorEvent {
	msg := fmt.Sprintf("Shutting down the CLI tool:\n%s\n", cberr.ErrorMessage(err))

	return &ErrorEvent{
		Event: newBaseEvent(NewErrorDetected, msg, true),
		err:   err,
	}
}

// ExitCode returns the exit code of the error of the event.
func (e ErrorEvent) ExitCode() int {
	return cberr.ErrorExitCode(e.err)
}
