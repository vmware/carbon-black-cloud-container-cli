package user

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui"
)

// AddCmd will return the user add sub command.
func AddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <user>",
		Short: "Add a new user",
		Long:  `Add a new user.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			go addUser(args[0])
			terminalui.NewDisplay().DisplayEvents()
		},
	}
}

func addUser(user string) {
	var msg string

	user = config.ConvertToValidProfileName(user)
	users := reflect.ValueOf(config.Config().Properties).MapKeys()

	for _, u := range users {
		if user == u.String() {
			msg = fmt.Sprintf("User %s already exists", user)
			bus.Publish(bus.NewMessageEvent(msg, true))

			return
		}
	}

	config.Config().Properties[user] = &config.Property{}
	msg = fmt.Sprintf("User %s added", user)
	bus.Publish(bus.NewMessageEvent(msg, true))
}
