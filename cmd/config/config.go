// Package config manages the all commands related to config.
package config

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/printtool"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
)

const numArgs = 2

// Cmd will return the config command.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config <option> <value>",
		Short: "Manage Carbon Black configuration",
		Long: printtool.Tprintf(`Get or set an octarine config option.
To get an option use '{{.appName}} config <option>'
To set an option use '{{.appName}} config <option> <value>'
To set configs in interactive mode use '{{.appName}} config'

Available options:
  active_user_profile - Current user profile
  org_key             - Org key
  saas_url            - Cloud SaaS url
`, map[string]interface{}{
			"appName": internal.ApplicationName,
		}),
		Args: cobra.MaximumNArgs(numArgs),
		Run: func(cmd *cobra.Command, args []string) {
			go func() {
				if len(args) == 0 {
					handleConfigInInteractiveMode()
					return
				}

				handleConfigFromArgs(args)
			}()

			terminalui.NewDisplay().DisplayEvents()
		},
	}
}

func handleConfigFromArgs(args []string) {
	var msg string

	user := config.GetConfig(config.ActiveUserProfile)
	setValue := len(args) == numArgs
	key := args[0]

	valid, option := config.Contains(key)
	if !valid {
		suggestionsString := ""

		if suggestions := config.SuggestionsFor(key); len(suggestions) > 0 {
			suggestionsString += "\n\nDid you mean this?\n"
			for _, s := range suggestions {
				suggestionsString += fmt.Sprintf("\t%v\t", s)
			}
		}

		msg = fmt.Sprintf("%s is an invalid config option%s", key, suggestionsString)
		e := cberr.NewError(cberr.ConfigErr, msg, nil)
		bus.Publish(bus.NewErrorEvent(e))
		logrus.Errorln(e)

		return
	}

	if setValue {
		value := args[1]
		config.SetConfigByOption(user, option, value)

		msg = fmt.Sprintf("Saving config [%s]: [%s] for user <%s>", option, value, user)
		bus.Publish(bus.NewMessageEvent(msg, true))

		return
	}

	if value := config.GetConfig(option); value != "" {
		msg = fmt.Sprintf("Config [%s] for user <%s> is: %s", option, user, value)
		bus.Publish(bus.NewMessageEvent(msg, true))

		return
	}

	msg = fmt.Sprintf("No value for option [%s] for user <%s>", option, user)
	bus.Publish(bus.NewMessageEvent(msg, true))
}

func handleConfigInInteractiveMode() {
	var msg string

	user := config.GetConfig(config.ActiveUserProfile)

	configOptions := []config.Option{config.SaasURL, config.OrgKey, config.DefaultBuildStep}
	for _, option := range configOptions {
		label := fmt.Sprintf("%v:", option)
		prompt := promptui.Prompt{
			Label:     label,
			Templates: printtool.Templates(),
			Default:   config.GetConfig(option),
		}

		input, err := prompt.Run()
		if err != nil {
			msg = fmt.Sprintf("Failed to set option [%v] for user [%v]", option, user)
			e := cberr.NewError(cberr.ConfigErr, msg, nil)
			bus.Publish(bus.NewErrorEvent(e))
			logrus.Errorln(e)

			return
		}

		input = strings.TrimSpace(input)
		if input != "" {
			config.SetConfigByOption(user, option, input)
		}
	}

	msg = fmt.Sprintf("Saving configs for user [%v]", user)
	bus.Publish(bus.NewMessageEvent(msg, true))
}
