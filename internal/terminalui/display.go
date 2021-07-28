/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package terminalui provides interface for display handlers
package terminalui

import (
	"os"
	"runtime"

	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui/dynamicui"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui/plainui"
	"golang.org/x/term"
)

// Display is the interface with the function for displaying events.
type Display interface {
	DisplayEvents()
}

// NewDisplay will select a display handler based on the environment.
func NewDisplay() Display {
	isStdoutATty := term.IsTerminal(int(os.Stdout.Fd()))
	isStderrATty := term.IsTerminal(int(os.Stderr.Fd()))
	notATerminal := !isStderrATty && !isStdoutATty

	switch {
	case notATerminal || runtime.GOOS == "windows" || config.Config().CliOpt.PlainMode:
		return plainui.NewDisplay()
	default:
		return dynamicui.NewDisplay()
	}
}
