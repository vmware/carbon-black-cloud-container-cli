// Package version manages the all commands related to version.
package version

import (
	"github.com/spf13/cobra"
	"gitlab.bit9.local/octarine/cbctl/internal"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui"
	"gitlab.bit9.local/octarine/cbctl/internal/util/printtool"
	"gitlab.bit9.local/octarine/cbctl/internal/version"
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
