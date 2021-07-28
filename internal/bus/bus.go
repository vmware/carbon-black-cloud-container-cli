/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bus

var (
	eventChan chan Event
	active    bool
)

// EventChan will return the event channel.
func EventChan() chan Event {
	return eventChan
}

// SetEventChan sets the singleton event channel.
func SetEventChan(c chan Event) {
	eventChan = c

	if c != nil {
		active = true
	}
}

// Publish an event onto the channel. If there is no channel set by the calling application, this does nothing.
func Publish(event Event) {
	if active {
		eventChan <- event
	}
}
