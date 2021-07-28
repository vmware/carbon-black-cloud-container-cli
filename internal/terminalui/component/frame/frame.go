/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package frame

import (
	"fmt"
	"io"
	"sync"
)

// Frame is the frame of a terminal output, it will control the position of cursor on terminal.
// Based on ANSI-CSI escape codes.
type Frame struct {
	writer      io.Writer
	currentLine int
	size        int
	lock        sync.Mutex
}

// NewFrame will initialize a new frame.
func NewFrame(writer io.Writer) *Frame {
	return &Frame{
		writer:      writer,
		currentLine: 0,
		size:        0,
	}
}

// Append will append a new line in the frame.
func (f *Frame) Append() *Line {
	f.lock.Lock()
	defer f.lock.Unlock()

	_ = f.moveTo(f.size)
	f.size++
	line := NewLine(f, f.size)

	if f.size != 1 {
		_ = f.writeString("\n")
	}

	f.currentLine = f.size

	return line
}

// Render will write message on the specific line.
func (f *Frame) Render(line int, msg string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if err := f.moveTo(line); err != nil {
		return err
	}

	return f.writeString(msg)
}

func (f *Frame) moveUp(line int) error {
	return f.writeString(fmt.Sprintf("\033[%dF", line))
}

func (f *Frame) moveDown(line int) error {
	return f.writeString(fmt.Sprintf("\033[%dE", line))
}

func (f *Frame) moveTo(line int) error {
	// cannot use append method since they shared a lock
	if line > f.size {
		for i := f.size + 1; i <= line; i++ {
			if err := f.writeString("\n"); err != nil {
				return err
			}
		}

		f.currentLine = line
		f.size = line
	}

	if err := f.moveUp(f.currentLine); err != nil {
		return err
	}

	if err := f.moveDown(line); err != nil {
		return err
	}

	f.currentLine = line

	return nil
}

// ShowCursor will show cursor.
func (f *Frame) ShowCursor() error {
	return f.writeString("\033[?25h")
}

// HideCursor will hide cursor.
func (f *Frame) HideCursor() error {
	return f.writeString("\033[?25l")
}

func (f *Frame) writeString(msg string) error {
	if _, err := io.WriteString(f.writer, msg); err != nil {
		return err
	}

	return nil
}
