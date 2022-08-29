package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vmware/carbon-black-cloud-container-cli/cmd/auth"
	configcmd "github.com/vmware/carbon-black-cloud-container-cli/cmd/config"
	"github.com/vmware/carbon-black-cloud-container-cli/cmd/image"
	"github.com/vmware/carbon-black-cloud-container-cli/cmd/k8sobject"
	"github.com/vmware/carbon-black-cloud-container-cli/cmd/user"
	"github.com/vmware/carbon-black-cloud-container-cli/cmd/version"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
)

var defaultConfigHome string

const (
	permModeReadWrite        = 0666
	permModeReadWriteExecute = 0700
)

func init() {
	setGlobalCliOptions()
	addSubCommands()

	cobra.OnInitialize(
		initEventChan,
		initLog,
		initConfig,
	)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func addSubCommands() {
	rootCmd.AddCommand(version.Cmd())
	rootCmd.AddCommand(configcmd.Cmd())
	rootCmd.AddCommand(auth.Cmd())
	rootCmd.AddCommand(user.Cmd())
	rootCmd.AddCommand(image.Cmd())
	rootCmd.AddCommand(k8sobject.Cmd())
}

// initConfig will initialize the debug log, if set by user.
func initLog() {
	flag := rootCmd.Flag("debug")
	if !flag.Changed {
		logrus.SetOutput(ioutil.Discard)
		return
	}

	if file, err := os.OpenFile(flag.Value.String(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, permModeReadWrite); err == nil {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetOutput(file)

		msg := fmt.Sprintf("[You are in DEBUG mode, the debug log will be saved in %s]", flag.Value.String())
		bus.Publish(bus.NewMessageEvent(msg, false))

		return
	}

	// failed to open user's assigned log, use the default one
	// create folder in the default config home if not exists
	createFolder(defaultConfigHome)

	if file, err := os.OpenFile(flag.NoOptDefVal, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, permModeReadWrite); err == nil {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetOutput(file)

		msg := fmt.Sprintf(
			"[You are in DEBUG mode, cannot save log in %s, the debug log will be saved in %s]",
			flag.Value.String(), flag.NoOptDefVal)
		bus.Publish(bus.NewMessageEvent(msg, false))
	}
}

// initEventChan will initialize the event channel and set it to bus singleton.
func initEventChan() {
	bufferSize := 10
	eventChan := make(chan bus.Event, bufferSize)
	bus.SetEventChan(eventChan)
}

// initConfig will initialize the app config via viper.
func initConfig() {
	// if no config file set, create the default folder if not exist
	flag := rootCmd.Flag("config")
	if !flag.Changed {
		createFolder(defaultConfigHome)
	}

	config.LoadAppConfig()
}

func setGlobalCliOptions() {
	// get config file home
	defaultConfigHome = findHomeDir()

	cfgDefaultName := fmt.Sprintf("%s/.%s.yaml", defaultConfigHome, internal.ApplicationName)
	flag := "config"
	rootCmd.PersistentFlags().StringP(flag, "c", cfgDefaultName, "config file")
	_ = viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag))

	flag = "user-profile"
	rootCmd.PersistentFlags().StringP(flag, "u", "", "user profile")
	_ = viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag))

	flag = "plain-mode"
	rootCmd.PersistentFlags().Bool(flag, false, "display ui on plain mode")
	_ = viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag))

	flag = "debug"
	logDefaultName := fmt.Sprintf("%s/debug.log", defaultConfigHome)
	rootCmd.PersistentFlags().String(flag, logDefaultName, "enable debug log")
	rootCmd.Flag(flag).NoOptDefVal = logDefaultName

	for flag, usage := range config.ConfigFileOverrides {
		rootCmd.PersistentFlags().String(flag, "", usage)
		_ = viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag))
	}
}

// findHomeDir will find the home directory follow XDG standards.
func findHomeDir() string {
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		if stat, err := os.Stat(xdgConfigHome); err != nil {
			if os.IsNotExist(err) {
				msg := fmt.Sprintf("The folder \"%s\" is not exist, creating the folder", xdgConfigHome)
				bus.Publish(bus.NewMessageEvent(msg, false))
			} else {
				msg := fmt.Sprintf("Failed to locate folder %s", xdgConfigHome)
				e := cberr.NewError(cberr.ConfigErr, msg, err)
				bus.Publish(bus.NewErrorEvent(e))
			}
		} else if !stat.IsDir() {
			msg := fmt.Sprintf("\"%s\" is not a folder", xdgConfigHome)
			e := cberr.NewError(cberr.ConfigErr, msg, err)
			bus.Publish(bus.NewErrorEvent(e))
		}

		return fmt.Sprintf("%s/.%s", xdgConfigHome, internal.ApplicationName)
	}

	// failed to detect XDG config home, use the default home directory
	defaultHome, err := homedir.Dir()
	if err != nil {
		msg := "Failed to get the home directory"
		e := cberr.NewError(cberr.ConfigErr, msg, err)
		bus.Publish(bus.NewErrorEvent(e))
	}

	return fmt.Sprintf("%s/.%s", defaultHome, internal.ApplicationName)
}

// createFolder if the folder is not exist.
func createFolder(dir string) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, permModeReadWriteExecute); err != nil {
				msg := fmt.Sprintf("Failed to create directory \"%s\", please create it manually", dir)
				e := cberr.NewError(cberr.ConfigErr, msg, err)
				bus.Publish(bus.NewErrorEvent(e))
			}
		}
	}
}
