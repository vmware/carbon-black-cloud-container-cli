package bus

import (
	"fmt"
	stereoevent "github.com/anchore/stereoscope/pkg/event"
	syftevent "github.com/anchore/syft/syft/event"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
)

// EventType is the type of event message.
type EventType string

// All the event types for the bus.
const (
	// Messages below are internal to cbctl and must not clash with libraries that post to the event bus

	StartScanTryFetchImageID       EventType = "start-scan-fetch-image-id"
	NewVersionAvailable            EventType = "new-version-event"
	NewMessageDetected             EventType = "new-message-event"
	NewErrorDetected               EventType = "new-error-event"
	NewCollectLayers               EventType = "new-collect-layers"
	ScanStarted                    EventType = "image-scanning-started-event"
	ScanFinished                   EventType = "image-scanning-finished-event"
	PrintSBOM                      EventType = "print-sbom-event"
	PrintPayload                   EventType = "print-payload-event"
	ValidateFinishedWithViolations EventType = "validate-finished-with-violations"
	ValidateFinishedSuccessfully   EventType = "validate-finished-successfully"

	// Events below this are not meant to be produced directly by cbctl.
	// Instead, they are messages coming from other libraries over the service bus

	// Messages by stereoscope

	PullDockerImage = EventType(stereoevent.PullDockerImage)
	FetchImage      = EventType(stereoevent.FetchImage)
	ReadImage       = EventType(stereoevent.ReadImage)
	ReadLayer       = EventType(stereoevent.ReadLayer)

	// Messages by syft

	CatalogerStarted = EventType(syftevent.PackageCatalogerStarted)
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
