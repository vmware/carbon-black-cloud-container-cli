/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package progressformatter is an implementation based on github.com/wagoodman/go-progress/simple
package progressformatter

import (
	"strings"
	"sync"

	"github.com/gookit/color"
	"github.com/wagoodman/go-progress"
)

// Theme is the theme of the progress bar.
type Theme struct {
	Filled         string
	Empty          string
	StartDelimiter string
	EndDelimiter   string
}

// DefaultTheme is the default theme for progress formatter.
var DefaultTheme = Theme{
	Filled:         color.HEX("#fcba03").Sprintf(">"),
	Empty:          color.HEX("#777777").Sprintf(">"),
	StartDelimiter: color.HEX("#777777").Sprintf("["),
	EndDelimiter:   color.HEX("#777777").Sprintf("]"),
}

// Formatter is the formatter of progress bar.
type Formatter struct {
	theme Theme
	width int
	lock  sync.Mutex
}

// NewFormatter will initialize a formatter with the default theme.
func NewFormatter(width int) *Formatter {
	return NewFormatterWithTheme(width, DefaultTheme)
}

// NewFormatterWithTheme will initialize a formatter with a specific theme.
func NewFormatterWithTheme(width int, theme Theme) *Formatter {
	return &Formatter{
		width: width,
		theme: theme,
	}
}

// Format will format the bar with progress.
func (f *Formatter) Format(p progress.Progress) string {
	f.lock.Lock()
	defer f.lock.Unlock()

	completedRatio := p.Ratio()
	if completedRatio < 0 {
		completedRatio = 0
	}

	completedCount := int(completedRatio * float64(f.width))

	todoCount := f.width - completedCount
	if todoCount < 0 {
		todoCount = 0
	}

	completedSection := strings.Repeat(f.theme.Filled, completedCount)
	todoSection := strings.Repeat(f.theme.Empty, todoCount)

	return f.theme.StartDelimiter + completedSection + todoSection + f.theme.EndDelimiter
}
