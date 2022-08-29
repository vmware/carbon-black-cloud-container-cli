package user

import (
	"fmt"
	"reflect"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/config"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
)

// ListCmd will return the user list sub command.
func ListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Manage Carbon Black user profile",
		Long:  `Show all the user profiles and select active user profile.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			go selectUser()
			terminalui.NewDisplay().DisplayEvents()
		},
	}
}

func selectUser() {
	var msg string

	label := fmt.Sprintf("Current active user [%s]", config.Config().ActiveUserProfile)
	prompt := promptui.Select{
		Label: label,
		Items: reflect.ValueOf(config.Config().Properties).MapKeys(),
	}

	_, profile, err := prompt.Run()
	if err != nil {
		msg = "Failed to select user profile"
		e := cberr.NewError(cberr.ConfigErr, msg, err)
		bus.Publish(bus.NewErrorEvent(e))
		logrus.Errorln(e)

		return
	}

	config.Config().ActiveUserProfile = profile
	msg = fmt.Sprintf("Active user profile selected: %s", profile)
	bus.Publish(bus.NewMessageEvent(msg, true))
}
