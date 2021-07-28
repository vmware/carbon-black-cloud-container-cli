/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package version manages the all commands related to version.
package version

import (
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/printtool"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/version"
)

// Cmd will return the version command.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the cli tool version and build info",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			go fetchVersion()
			terminalui.NewDisplay().DisplayEvents()
		},
	}
}

func fetchVersion() {
	currentVersion := version.GetCurrentVersion()

	versionMsg := `       
        __         __  __
  _____/ /_  _____/ /_/ /
 / ___/ __ \/ ___/ __/ / 
/ /__/ /_/ / /__/ /_/ /  
\___/_.___/\___/\__/_/   

`

	versionMsg += printtool.Tprintf(`Application: {{.appName}}
Version:     {{.version}}
BuildDate:   {{.buildDate}}
Platform:    {{.platform}}
GoVersion:   {{.goVersion}}
Compiler:    {{.compiler}} 
`, map[string]interface{}{
		"appName":   internal.ApplicationName,
		"version":   currentVersion.Version,
		"buildDate": currentVersion.BuildDate,
		"platform":  currentVersion.Platform,
		"goVersion": currentVersion.GoVersion,
		"compiler":  currentVersion.Compiler,
	})

	bus.Publish(bus.NewMessageEvent(versionMsg, true))
}
