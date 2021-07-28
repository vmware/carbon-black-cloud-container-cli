/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/zalando/go-keyring"
)

var appConfig *AppConfig

const authenticatedByKeyringMaskValue = "[authenticated by keyring]"

// AppConfig is the config of the cli app.
type AppConfig struct {
	ActiveUserProfile string
	AccessToKeyring   bool
	Properties        map[string]*Property
	CliOpt            *CliOption
}

// Property is the property of a single user.
type Property struct {
	SaasURL          string
	OrgKey           string
	AuthByKeyring    bool
	CBApiID          string
	CBApiKey         string
	DefaultBuildStep string
}

// CliOption contains all the cli flag options.
type CliOption struct {
	ConfigFile  string `mapstructure:"config"`
	UserProfile string `mapstructure:"user-profile"`
	PlainMode   bool   `mapstructure:"plain-mode"`
}

func init() {
	appConfig = &AppConfig{
		ActiveUserProfile: "",
		AccessToKeyring:   false,
		Properties:        make(map[string]*Property),
		CliOpt:            &CliOption{},
	}
}

// Config will return the appConfig singleton instance.
func Config() *AppConfig {
	return appConfig
}

// LoadAppConfig will initialize application config from viper.
func LoadAppConfig() {
	if err := viper.Unmarshal(appConfig.CliOpt); err != nil {
		logrus.Println("Cannot bind cli flag options: ", err)
	}

	// Fetch env variable with application name prefix
	viper.SetEnvPrefix(internal.ApplicationName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if appConfig.CliOpt.ConfigFile != "" {
		// use config file from the flag.
		viper.SetConfigFile(appConfig.CliOpt.ConfigFile)
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		logrus.Println("Using config file: ", viper.ConfigFileUsed())
	} else {
		logrus.Println("No config file detected")
	}

	appConfig.ActiveUserProfile = ConvertToValidProfileName(appConfig.CliOpt.UserProfile)
	initUserProperty(appConfig.ActiveUserProfile)

	// check access to keyring
	if _, err := keyring.Get(appConfig.ActiveUserProfile, CBApiID.String()); err == nil || err == keyring.ErrNotFound {
		appConfig.AccessToKeyring = true
	}

	setUserProfiles()
}

// GetConfig will get the config by a given option from the active profile.
func GetConfig(o Option) string {
	switch o {
	case ActiveUserProfile:
		return appConfig.ActiveUserProfile
	case SaasURL:
		return appConfig.Properties[appConfig.ActiveUserProfile].SaasURL
	case OrgKey:
		return appConfig.Properties[appConfig.ActiveUserProfile].OrgKey
	case CBApiID:
		return appConfig.Properties[appConfig.ActiveUserProfile].CBApiID
	case CBApiKey:
		return appConfig.Properties[appConfig.ActiveUserProfile].CBApiKey
	case DefaultBuildStep:
		return appConfig.Properties[appConfig.ActiveUserProfile].DefaultBuildStep
	case cntOfOptions:
		fallthrough
	default:
		panic(fmt.Sprintf("Invalid config option provided: %v", o))
	}
}

// SetConfigByOption will set the config by a given option-value to the active profile.
func SetConfigByOption(user string, o Option, value string) {
	if user == "" {
		user = appConfig.ActiveUserProfile
	}

	initUserProperty(user)

	switch o {
	case SaasURL:
		appConfig.Properties[user].SaasURL = value
	case OrgKey:
		appConfig.Properties[user].OrgKey = value
	case CBApiKey:
		appConfig.Properties[user].CBApiKey = value
	case CBApiID:
		appConfig.Properties[user].CBApiID = value
	case DefaultBuildStep:
		appConfig.Properties[user].DefaultBuildStep = value
	case ActiveUserProfile, cntOfOptions:
		fallthrough
	default:
		break
	}
}

// PersistConfig will persist all the configs.
func PersistConfig() error {
	writeViper := viper.New()

	writeViper.Set(ActiveUserProfile.String(), appConfig.ActiveUserProfile)

	for user, profile := range appConfig.Properties {
		writeViper.Set(SaasURL.StringWithPrefix(user), profile.SaasURL)
		writeViper.Set(OrgKey.StringWithPrefix(user), profile.OrgKey)
		writeViper.Set(DefaultBuildStep.StringWithPrefix(user), profile.DefaultBuildStep)

		// overwrite with mask value for those values saved in keyring
		if profile.AuthByKeyring {
			writeViper.Set(CBApiKey.StringWithPrefix(user), authenticatedByKeyringMaskValue)
			writeViper.Set(CBApiID.StringWithPrefix(user), authenticatedByKeyringMaskValue)
		} else {
			writeViper.Set(CBApiKey.StringWithPrefix(user), profile.CBApiKey)
			writeViper.Set(CBApiID.StringWithPrefix(user), profile.CBApiID)

			// remove from keyring
			_ = keyring.Delete(user, CBApiKey.String())
			_ = keyring.Delete(user, CBApiID.String())
		}
	}

	if err := writeViper.WriteConfigAs(appConfig.CliOpt.ConfigFile); err != nil {
		errMsg := fmt.Sprintf("Failed to persist config to %s", appConfig.CliOpt.ConfigFile)
		e := cberr.NewError(cberr.ConfigErr, errMsg, err)
		logrus.Errorln(e)

		return e
	}

	return nil
}

// ConvertToValidProfileName will convert input to valid profile name.
func ConvertToValidProfileName(input string) string {
	profile := strings.ReplaceAll(input, ".", "_")
	if profile == "" {
		if activeProfile := viper.GetString(ActiveUserProfile.String()); activeProfile != "" {
			if !strings.HasPrefix(activeProfile, internal.ApplicationName) {
				profile = fmt.Sprintf("%s_%s", internal.ApplicationName, activeProfile)
			} else {
				profile = activeProfile
			}
		} else {
			profile = fmt.Sprintf("%s_default", internal.ApplicationName)
		}
	}

	if !strings.HasPrefix(profile, internal.ApplicationName) {
		profile = fmt.Sprintf("%s_%s", internal.ApplicationName, profile)
	}

	return profile
}

// setUserProfiles will set all the user profiles from the viper.
func setUserProfiles() {
	for _, key := range viper.AllKeys() {
		// user is the current user of this key & option in the option name of this key
		user, optionName := appConfig.ActiveUserProfile, key

		// if a key contains dot that means it already has a prefix, we need to update user and option
		if splitKeys := strings.SplitN(key, ".", 2); len(splitKeys) == 2 {
			user, optionName = splitKeys[0], splitKeys[1]
		}

		value := viper.GetString(key)
		// if the value is authenticated means we are fetching result from keyring,
		// do not overwrite it with the mask value
		if value == authenticatedByKeyringMaskValue {
			initUserProperty(user)

			if v, err := keyring.Get(user, optionName); err == nil {
				value = v
				appConfig.Properties[user].AuthByKeyring = true
			}
		}

		if _, option := Contains(optionName); option >= 0 {
			SetConfigByOption(user, option, value)
		}
	}
}

func initUserProperty(name string) {
	if _, ok := appConfig.Properties[name]; !ok {
		appConfig.Properties[name] = &Property{}
	}
}
