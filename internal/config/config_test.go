package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
	"gitlab.bit9.local/octarine/cbctl/internal"
	"gitlab.bit9.local/octarine/cbctl/internal/config"
)

func TestContains(t *testing.T) {
	commonOptions := []string{"org_key", "saas_url", "active_user_profile"}
	authOptions := []string{"cb_api_id", "cb_api_key"}
	invalidOptions := []string{"org_id", "fake_key"}

	for _, s := range commonOptions {
		if ok, o := config.Contains(s); !ok || o == -1 {
			t.Errorf("Common option didn't return true")
		}
	}

	for _, s := range authOptions {
		if ok, o := config.Contains(s); !ok {
			if o == -1 {
				t.Errorf("Auth option didn't return right option")
			}
		}
	}

	for _, s := range invalidOptions {
		if ok, o := config.Contains(s); ok || o > -1 {
			t.Errorf("Contains return true for invalid input")
		}
	}
}

func TestSuggestionsFor(t *testing.T) {
	testWords := []string{"org_key", "org_", "rrg_key", "org_ke", "trg_kyy", "orgKey"}
	expectedSuggestions := []string{"org_key"}

	for _, w := range testWords {
		if s := config.SuggestionsFor(w); !reflect.DeepEqual(s, expectedSuggestions) {
			t.Errorf("Unexpected suggestion get: %v", s)
		}
	}
}

func TestLoadConfigFromEnvVariables(t *testing.T) {
	var (
		apiIDFromEnvVar  = "api-id-from-env-var"
		apiKeyFromEnvVar = "api-key-from-env-var"
		orgKeyFromEnvVar = "org-key-from-env-var"

		saasURLFromConfigFile = "https://test-1.com"
	)

	mockValues := map[string]string{
		"user-profile":     "test-1",
		"test-2.org_key":   "test-2-org",
		"test-2.cb_api_id": "123",
		// these value conflict with the ones from the env variable
		// these should be overridden by the env variables
		"cb_api_id":  "ABC123",
		"cb_api_key": "ABC1234",
		"org_key":    "test-1-org",
		// this value is not overridden by env variables, so it should match
		"saas_url": saasURLFromConfigFile,
	}
	for k, v := range mockValues {
		viper.Set(k, v)
	}

	_ = os.Setenv("CBCTL_CB_API_ID", apiIDFromEnvVar)
	_ = os.Setenv("CBCTL_CB_API_KEY", apiKeyFromEnvVar)
	_ = os.Setenv("CBCTL_ORG_KEY", orgKeyFromEnvVar)

	t.Cleanup(func() {
		viper.Reset()

		_ = os.Setenv("CBCTL_CB_API_ID", "")
		_ = os.Setenv("CBCTL_CB_API_KEY", "")
		_ = os.Setenv("CBCTL_ORG_KEY", "")
	})

	config.LoadAppConfig()

	require.Equal(t, apiKeyFromEnvVar, config.GetConfig(config.CBApiKey))
	require.Equal(t, apiIDFromEnvVar, config.GetConfig(config.CBApiID))
	require.Equal(t, orgKeyFromEnvVar, config.GetConfig(config.OrgKey))
	require.Equal(t, saasURLFromConfigFile, config.GetConfig(config.SaasURL))
}

func TestLoadConfigCLIValuesOverride(t *testing.T) {
	var (
		apiIDFromCLIFlags   = "api-id-from-cli-flags"
		apiKeyFromCLIFlags  = "api-key-from-cli-flags"
		orgKeyFromCLIFlags  = "org-key-from-cli-flags"
		saasURLFromCLIFlags = "saas-url-from-cli-flags"
	)

	// initialize mock values
	mockValues := map[string]string{
		"user-profile":     "test-1",
		"org_key":          "test-1-org",
		"cb_api_id":        "ABC123",
		"saas_url":         "https://test-1.com",
		"test-2.org_key":   "test-2-org",
		"test-2.cb_api_id": "123",

		// these are options passed as CLI args and should override the ones coming from config file
		"cb-api-id":  apiIDFromCLIFlags,
		"cb-api-key": apiKeyFromCLIFlags,
		"org-key":    orgKeyFromCLIFlags,
		"saas-url":   saasURLFromCLIFlags,
	}
	for k, v := range mockValues {
		viper.Set(k, v)
	}

	defer viper.Reset()

	config.LoadAppConfig()

	require.Equal(t, apiIDFromCLIFlags, config.GetConfig(config.CBApiID))
	require.Equal(t, apiKeyFromCLIFlags, config.GetConfig(config.CBApiKey))
	require.Equal(t, orgKeyFromCLIFlags, config.GetConfig(config.OrgKey))
	require.Equal(t, saasURLFromCLIFlags, config.GetConfig(config.SaasURL))
}

func TestLoadAndPersistAppConfig(t *testing.T) {
	// initialize mock values
	mockValues := map[string]string{
		"user-profile":     "test-1",
		"org_key":          "test-1-org",
		"cb_api_id":        "[authenticated by keyring]",
		"saas_url":         "https://test-1.com",
		"test-2.org_key":   "test-2-org",
		"test-2.cb_api_id": "123",
	}
	for k, v := range mockValues {
		viper.Set(k, v)
	}

	defer viper.Reset()

	keyring.MockInit()
	_ = keyring.Set("cbctl_test-1", config.CBApiID.String(), "123")
	config.LoadAppConfig()

	// judge if configs are set properly
	if config.GetConfig(config.ActiveUserProfile) !=
		fmt.Sprintf("%s_%s", internal.ApplicationName, mockValues["user-profile"]) {
		t.Errorf("User Profile is not set properly, expected: %s, actual: %s",
			mockValues["user-profile"], config.GetConfig(config.ActiveUserProfile))
	}

	if config.GetConfig(config.OrgKey) != mockValues["org_key"] {
		t.Errorf("Org Key is not set properly, expected: %s, actual: %s",
			mockValues["org_key"], config.GetConfig(config.OrgKey))
	}

	if config.GetConfig(config.SaasURL) != mockValues["saas_url"] {
		t.Errorf("SAAS URL is not set properly, expected: %s, actual: %s",
			mockValues["saas_url"], config.GetConfig(config.SaasURL))
	}

	if config.Config().Properties["cbctl_test-1"].AuthByKeyring != true {
		t.Errorf("AuthByKeyring should be set to true")
	}

	if config.Config().Properties["test-2"].OrgKey != mockValues["test-2.org_key"] {
		t.Errorf("Org Key for '%s' is not set properly, expected: %s, actual: %s",
			"test-2", mockValues["saas_url"], config.GetConfig(config.SaasURL))
	}

	// test SetConfigByOption
	config.SetConfigByOption("test-2", config.SaasURL, "https://test-2.com")

	if config.Config().Properties["test-2"].SaasURL != "https://test-2.com" {
		t.Errorf("Org Key for '%s' is not set properly, expected: %s, actual: %s",
			"test-2", mockValues["saas_url"], config.GetConfig(config.SaasURL))
	}

	// test PersistConfig
	// fetch the test file position and split out the base dir
	_, testPath, _, _ := runtime.Caller(0) // nolint:dogsled
	baseDir := strings.Split(testPath, "internal")[0]
	targetConfigFile := filepath.Join(baseDir, "test/configfile/writeconfig.yaml")

	defer func() {
		if err := os.Remove(targetConfigFile); err != nil {
			t.Logf("Failed to remove test config file (%s): %v", targetConfigFile, err)
		}
	}()

	config.Config().CliOpt.ConfigFile = targetConfigFile

	if err := config.PersistConfig(); err != nil {
		t.Errorf("failed to persist config: %v", err)
	}
}

func TestConvertToValidProfileName(t *testing.T) {
	// input without cli prefix
	testInput := "test-input"
	convertedProfile := config.ConvertToValidProfileName(testInput)

	if convertedProfile != fmt.Sprintf("%s_%s", internal.ApplicationName, testInput) {
		t.Errorf("Profile name converted wrongly: %s", convertedProfile)
	}

	// input with cli prefix
	testInput = internal.ApplicationName + "_input"
	convertedProfile = config.ConvertToValidProfileName(testInput)

	if convertedProfile != testInput {
		t.Errorf("Profile name converted wrongly: %s", convertedProfile)
	}

	// empty input and no profile set
	convertedProfile = config.ConvertToValidProfileName("")

	if convertedProfile != fmt.Sprintf("%s_default", internal.ApplicationName) {
		t.Errorf("Profile name converted wrongly: %s", convertedProfile)
	}

	// empty input and with profile set (without prefix)
	testInput = "test-input"
	viper.Set(config.ActiveUserProfile.String(), testInput)
	convertedProfile = config.ConvertToValidProfileName("")

	if convertedProfile != fmt.Sprintf("%s_%s", internal.ApplicationName, testInput) {
		t.Errorf("Profile name converted wrongly: %s", convertedProfile)
	}

	// empty input and with profile set (with prefix)
	testInput = internal.ApplicationName + "_input"
	viper.Set(config.ActiveUserProfile.String(), testInput)
	convertedProfile = config.ConvertToValidProfileName("")

	if convertedProfile != testInput {
		t.Errorf("Profile name converted wrongly: %s", convertedProfile)
	}
}
