/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package spinner is a simple implementation of spinner component
package spinner

import (
	"strings"
	"sync"
)

// DefaultDotSet is the default dot set for spinner.
var DefaultDotSet = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"

// Spinner is the spinner indicator for terminal.
type Spinner struct {
	index   int
	charset []string
	lock    sync.Mutex
}

// NewSpinner will create a new spinner with default charset.
func NewSpinner() *Spinner {
	return NewSpinnerWithCharset(DefaultDotSet)
}

// NewSpinnerWithCharset will create a new spinner based on the charset.
func NewSpinnerWithCharset(charset string) *Spinner {
	return &Spinner{
		charset: strings.Split(charset, ""),
	}
}

// Current will show the current character of the spinner.
func (s *Spinner) Current() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.charset[s.index]
}

// Next will show the next character of the spinner.
func (s *Spinner) Next() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	c := s.charset[s.index]

	s.index++
	if s.index >= len(s.charset) {
		s.index = 0
	}

	return c
}
