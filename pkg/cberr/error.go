package cberr

import (
	"bytes"
	"fmt"
)

// Error is the customized error type for cli, implement the builtin error interface.
type Error struct {
	// machine-readable error code
	code Code
	// human-readable message
	message string
	// original error
	err error
}

// NewError initializes a new cli-error.
func NewError(code Code, message string, err error) *Error {
	return &Error{
		code:    code,
		message: message,
		err:     err,
	}
}

// Error returns the string representation of the error message.
func (e *Error) Error() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "<%d> ", e.code)
	buf.WriteString(e.message)

	// if wrapping an error, print its Error() message
	if e.err != nil {
		buf.WriteString(" - " + e.err.Error())
	}

	return buf.String()
}

// ExitCode returns the code for this error.
func (e *Error) ExitCode() int {
	return e.code.exitCode()
}

// ErrorCode returns the code of the root error, if available.
func ErrorCode(err error) Code {
	if err == nil {
		return -1
	} else if e, ok := err.(*Error); ok && e.code >= 0 {
		return e.code
	} else if ok && e.err != nil {
		return ErrorCode(e.err)
	}

	return UnclassifiedErr
}

// ErrorExitCode returns the exit code of the root error, if available.
func ErrorExitCode(err error) int {
	if err == nil {
		return 0
	} else if e, ok := err.(*Error); ok && e.code >= 0 {
		return e.ExitCode()
	} else if ok && e.err != nil {
		return ErrorExitCode(e.err)
	}

	return 1
}

// ErrorMessage returns the human-readable message of the error, if available.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.message != "" {
		return e.message
	} else if ok && e.err != nil {
		return ErrorMessage(e.err)
	}

	return err.Error()
}
