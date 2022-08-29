package auth

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	keyring "github.com/zalando/go-keyring"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/config"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui"
	"gitlab.bit9.local/octarine/cbctl/internal/util/printtool"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
)

const numArgs = 2

// SetCbAPIAccessCmd will return the command for setting carbon api access.
func SetCbAPIAccessCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <api_id> <api_secret_key>",
		Short: "Set API ID and API Secret Key",
		Long:  `Save Carbon Black API ID and API Secret Key in the credential store`,
		Args:  cobra.MaximumNArgs(numArgs),
		Run: func(cmd *cobra.Command, args []string) {
			go func() {
				if len(args) == 0 {
					handleAuthAccessInInteractiveMode()
					return
				}

				handleAuthAccessFromArgs(args)
			}()

			terminalui.NewDisplay().DisplayEvents()
		},
	}
}

func handleAuthAccessFromArgs(args []string) {
	var msg string

	user := config.GetConfig(config.ActiveUserProfile)
	apiID, apiKey := args[0], args[1]

	config.SetConfigByOption(user, config.CBApiID, apiID)
	config.SetConfigByOption(user, config.CBApiKey, apiKey)

	if errs := SaveCbAPIAccess(config.GetConfig(config.ActiveUserProfile), apiID, apiKey); errs != nil {
		logrus.Errorf("Cannot save access in keyring: %v", errs)

		msg = fmt.Sprintf("No keyring found; storing credentials in %s instead", viper.ConfigFileUsed())
		bus.Publish(bus.NewMessageEvent(msg, true))

		return
	}

	msg = "Saving the Carbon Black API Access in keyring"
	bus.Publish(bus.NewMessageEvent(msg, true))
}

func handleAuthAccessInInteractiveMode() {
	var msg string

	user := config.Config().ActiveUserProfile

	if config.Config().AccessToKeyring {
		prompt := promptui.Prompt{
			Label:     "Valid keyring detected, do you want to save access into keyring",
			IsConfirm: true,
		}

		if input, _ := prompt.Run(); input == "y" {
			config.Config().Properties[user].AuthByKeyring = true
		} else {
			config.Config().Properties[user].AuthByKeyring = false
		}
	}

	authOptions := []config.Option{config.CBApiID, config.CBApiKey}
	for _, option := range authOptions {
		label := fmt.Sprintf("%v:", option)
		prompt := promptui.Prompt{
			Label:     label,
			Templates: printtool.Templates(),
			Mask:      '#',
		}

		input, err := prompt.Run()
		if err != nil {
			msg = fmt.Sprintf("Failed to set [%v] for user [%v]", option, user)
			e := cberr.NewError(cberr.ConfigErr, msg, err)
			bus.Publish(bus.NewErrorEvent(e))
			logrus.Errorln(e)

			return
		}

		config.SetConfigByOption(user, option, input)
	}

	if config.Config().Properties[user].AuthByKeyring {
		_ = SaveCbAPIAccess(user, config.GetConfig(config.CBApiID), config.GetConfig(config.CBApiKey))
	}

	msg = "Saving the Carbon Black API Access"
	bus.Publish(bus.NewMessageEvent(msg, true))
}

// SaveCbAPIAccess will save the cb access to credential store via keyring.
func SaveCbAPIAccess(profile, apiID, apiKey string) []error {
	errKey := keyring.Set(profile, config.CBApiKey.String(), apiKey)
	errID := keyring.Set(profile, config.CBApiID.String(), apiID)

	if errID == nil && errKey == nil {
		config.Config().Properties[profile].AuthByKeyring = true
		return nil
	}

	return []error{errKey, errID}
}
