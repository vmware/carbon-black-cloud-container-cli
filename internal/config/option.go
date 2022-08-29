package config

import (
	"fmt"
	"strings"
)

// Option is the behind number of a config, used for enum.
type Option int

const (
	// ActiveUserProfile is the current active profile.
	ActiveUserProfile Option = iota
	// SaasURL is the cloud SaaS url.
	SaasURL
	// OrgKey is the org_id.
	OrgKey
	// DefaultBuildStep in the default build step.
	DefaultBuildStep
	cntOfOptions

	// CBApiID is the carbon black api id;
	// put under cntOfOptions since the following two options are for auth.
	CBApiID
	// CBApiKey is the carbon black api key.
	CBApiKey
)

// String will return the option string.
func (o Option) String() string {
	switch o {
	case ActiveUserProfile:
		return "active_user_profile"
	case SaasURL:
		return "saas_url"
	case OrgKey:
		return "org_key"
	case CBApiID:
		return "cb_api_id"
	case CBApiKey:
		return "cb_api_key"
	case DefaultBuildStep:
		return "default_build_step"
	case cntOfOptions:
		fallthrough
	default:
		panic(fmt.Sprintf("Invalid config option provided: %d", o))
	}

	return ""
}

// StringWithPrefix will return the option string with a prefix.
func (o Option) StringWithPrefix(prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, o.String())
}

// Contains will check if an input config is a valid one or not.
func Contains(typedName string) (bool, Option) {
	for i := 0; i < int(cntOfOptions); i++ {
		if typedName == Option(i).String() {
			return true, Option(i)
		}
	}

	// check api id & key, since these two are not valid config options, will return false
	if typedName == CBApiID.String() {
		return false, CBApiID
	}

	if typedName == CBApiKey.String() {
		return false, CBApiKey
	}

	return false, -1
}

// SuggestionsFor provides suggestions for the typedName.
func SuggestionsFor(typedName string) []string {
	var suggestions []string

	// suggestionsMinimumDistance defines minimum levenshtein distance to display suggestions.
	// Must be > 0.
	suggestionsMinimumDistance := 3

	for i := 0; i < int(cntOfOptions); i++ {
		ld := levenshteinDistance(typedName, Option(i).String())
		suggestByLevenshtein := ld <= suggestionsMinimumDistance
		suggestByPrefix := strings.HasPrefix(Option(i).String(), typedName)

		if suggestByLevenshtein || suggestByPrefix {
			suggestions = append(suggestions, Option(i).String())
		}
	}

	return suggestions
}

// levenshteinDistance compares two strings and returns the levenshtein distance between them.
func levenshteinDistance(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}

	for i := range d {
		d[i][0] = i
	}

	for j := range d[0] {
		d[0][j] = j
	}

	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}
	}

	return d[len(s)][len(t)]
}
