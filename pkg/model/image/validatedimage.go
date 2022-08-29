package image

import (
	"fmt"

	"gitlab.bit9.local/octarine/cbctl/internal/util/colorizer"
	"gitlab.bit9.local/octarine/cbctl/internal/util/tabletool"
)

const (
	riskHeader       = "Risk"
	ruleHeader       = "Rule"
	policyHeader     = "Policy"
	violationsHeader = "Violations"
)

// ValidatedImage response model from guardrails validator service.
type ValidatedImage struct {
	Identifier       `json:",inline"`
	PolicyViolations []PolicyViolation `json:"policy_violations"`
}

// ValidatedImageOption is the option for showing validated image result.
type ValidatedImageOption struct {
	tabletool.Option
}

var validatedImageOpts ValidatedImageOption

// NewValidatedImage will initialize a validated image.
func NewValidatedImage(identifier Identifier, violations []PolicyViolation, opts ValidatedImageOption) *ValidatedImage {
	validatedImageOpts = opts

	return &ValidatedImage{
		Identifier:       identifier,
		PolicyViolations: violations,
	}
}

// Title is the title of the ValidatedImage result.
func (v *ValidatedImage) Title() string {
	return fmt.Sprintf("Validate result for %s (%s):", v.FullTag, v.ManifestDigest)
}

// Header is the header columns of the ValidatedImage result.
func (v *ValidatedImage) Header() []string {
	return []string{
		riskHeader,
		ruleHeader,
		policyHeader,
		violationsHeader,
	}
}

// Rows returns all the violations of the ValidatedImage result as list of rows.
func (v *ValidatedImage) Rows() [][]string {
	result := make([][]string, 0)

	for _, violation := range v.PolicyViolations {
		result = append(result, []string{
			colorizer.ColorizeRisk(violation.GetRisk()),
			violation.GetRuleName(),
			violation.GetPolicyName(),
			violation.GetViolation(),
		})
	}

	return result
}
