// Package version contains all metadata for cli.
package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	hashiVersion "github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"gitlab.bit9.local/octarine/cbctl/internal/config"
	"gitlab.bit9.local/octarine/cbctl/internal/util/httptool"
)

// all variables here should be assigned via go-liner during build process;
// set by `go build -ldflags "-X ..."`.
var (
	// This version is a placeholder, it will be replaced during the build.
	version     = "v0.0.0"
	syftVersion = "v0.15.1"
	buildDate   = ""
	platform    = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Version defines the application version details (generally from build information).
type Version struct {
	Version     string `json:"version"`     // application semantic version
	SyftVersion string `json:"syftVersion"` // version of syft
	BuildDate   string `json:"buildDate"`   // date of the build
	GoVersion   string `json:"goVersion"`   // go runtime version at build-time
	Compiler    string `json:"compiler"`    // compiler used at build-time
	Platform    string `json:"platform"`    // GOOS and GOARCH at build-time
}

// GetCurrentVersion provide metadata of current version.
func GetCurrentVersion() Version {
	return Version{
		Version:     version,
		SyftVersion: syftVersion,
		BuildDate:   buildDate,
		GoVersion:   runtime.Version(),
		Compiler:    runtime.Compiler,
		Platform:    platform,
	}
}

// IsUpdateAvailable indicates if there is a newer application version available, and if so, what the new version is.
func IsUpdateAvailable() (bool, string, error) {
	currentVersionStr := GetCurrentVersion().Version
	if currentVersionStr == "" {
		return false, "", nil
	}

	currentVersion, err := hashiVersion.NewVersion(currentVersionStr)
	if err != nil {
		return false, "", fmt.Errorf("failed to parse current version: %w", err)
	}

	latestVersion, err := fetchLatestVersion()
	if err != nil {
		return false, "", err
	}

	if latestVersion.GreaterThan(currentVersion) {
		return true, latestVersion.String(), nil
	}

	return false, "", nil
}

func fetchLatestVersion() (*hashiVersion.Version, error) {
	saasTmpl := strings.Trim(config.GetConfig(config.SaasURL), "/")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/orgs")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1")
	saasTmpl = strings.TrimSuffix(saasTmpl, "/v1beta")
	basePath := fmt.Sprintf("%s/v1/orgs/%s", saasTmpl, config.GetConfig(config.OrgKey))
	latestVersionFetchPath := fmt.Sprintf("%s/management/cli_instances/latest_version", basePath)

	session := httptool.NewRequestSession(config.GetConfig(config.CBApiID), config.GetConfig(config.CBApiKey))

	_, resp, err := session.RequestData(http.MethodGet, latestVersionFetchPath, nil)
	if err != nil {
		logrus.Errorf("Failed to fetch the latest version: %v", err)
		return nil, err
	}

	var versionInfo struct {
		Version   string `json:"version"`
		BuildTime string `json:"build_time"`
	}

	if err := json.Unmarshal(resp, &versionInfo); err != nil {
		logrus.Errorf("Failed to unmarshal the version info: %v", err)
		return nil, err
	}

	// once distribution platform is decided, implement this
	return hashiVersion.NewVersion(versionInfo.Version)
}
