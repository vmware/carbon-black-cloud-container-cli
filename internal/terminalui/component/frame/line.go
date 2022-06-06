package frame

import (
	"fmt"
	"strings"
)

// Line represents one line on the terminal frame.
type Line struct {
	frame      *Frame
	lineNumber int
}

// NewLine will initialize a new line.
func NewLine(frame *Frame, num int) *Line {
	return &Line{
		frame:      frame,
		lineNumber: num,
	}
}

// Render will render msg on the current line, if msg contains multiple lines, print in the next lines.
func (l *Line) Render(msg string) error {
	for i, m := range strings.Split(msg, "\n") {
		m = fmt.Sprintf("\033[K\033[G%v", m)

		if err := l.frame.Render(l.lineNumber+i, m); err != nil {
			return err
		}
	}

	return nil
}

// Clear will clear the whole line.
func (l *Line) Clear() error {
	return l.frame.Render(l.lineNumber, "\033[G\033[K")
}
