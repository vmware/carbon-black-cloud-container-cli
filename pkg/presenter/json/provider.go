/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package json

// Provider implement the methods needed for creating json.
type Provider interface {
	Title() string
	Footer() string
}
