/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/*
Package bus provides access to a singleton instance of an event bus (provided by the calling application). These events
can provide static information, but also have an object as a payload for which the consumer can poll for updates.

Note that the singleton instance is only allowed to publish events and not subscribe to them --this is intentional.
Internal library interactions should continue to use traditional in-execution-path approaches for data sharing
(e.g. function returns and channels) and not depend on bus subscriptions for critical interactions (e.g. one part of the
lib publishes an event and another part of the lib subscribes and reacts to that event). The bus is provided only as a
means for consumers to observe events emitted from the library (such as to provide a rich UI) and not to allow
consumers to augment or otherwise change execution.
*/
package bus
