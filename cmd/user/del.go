package user

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui"
	"github.com/zalando/go-keyring"
)

// DelCmd will return the user delete sub command.
func DelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "del <user>",
		Short: "Delete an existing user",
		Long:  `Delete an existing user.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			go deleteUser(args[0])
			terminalui.NewDisplay().DisplayEvents()
		},
	}
}

func deleteUser(user string) {
	var msg string

	user = config.ConvertToValidProfileName(user)

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure to delete user %s", user),
		IsConfirm: true,
	}

	if input, _ := prompt.Run(); input == "y" {
		if _, ok := config.Config().Properties[user]; ok {
			delete(config.Config().Properties, user)
			_ = keyring.Delete(user, config.CBApiID.String())
			_ = keyring.Delete(user, config.CBApiKey.String())

			if user == config.GetConfig(config.ActiveUserProfile) {
				config.Config().ActiveUserProfile = config.ConvertToValidProfileName("default")
			}

			msg = fmt.Sprintf("User %s has been deleted", user)
			bus.Publish(bus.NewMessageEvent(msg, true))

			return
		}

		msg = fmt.Sprintf("User %s is not a valid user", user)
		bus.Publish(bus.NewMessageEvent(msg, false))
	}

	msg = fmt.Sprintf("User %s deletion has been canceled", user)
	bus.Publish(bus.NewMessageEvent(msg, true))
}
