/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package table

// Provider implement the methods needed for creating table.
type Provider interface {
	Title() string
	Footer() string
	Header() []string
	Rows() [][]string
}
