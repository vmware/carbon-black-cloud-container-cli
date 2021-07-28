/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/colorizer"
)

const (
	namespaceHeader = "Namespace"
	kindHeader      = "Kind"
	nameHeader      = "Name"
	ruleHeader      = "Rule"
	riskHeader      = "Risk"
	filePathHeader  = "File"

	k8sPoliciesTemplate = "kubernetes/policy/policies"
)

// ViolatedResources represents a group of violating resources.
type ViolatedResources struct {
	Policy    string              `json:"-"`
	Resources []ValidatedResource `json:"objects"`
}

// PolicyViolationsCount get the count of violated policies.
func (v ViolatedResources) PolicyViolationsCount() int {
	count := 0

	for _, resource := range v.Resources {
		count += len(resource.PolicyViolations)
	}

	return count
}

// Title is the title of the ViolatedResources result.
func (v ViolatedResources) Title() string {
	return fmt.Sprintf("Found %d violations for policy \"%s\":", v.PolicyViolationsCount(), v.Policy)
}

// Footer is the footer of the ViolatedResources result.
func (v ViolatedResources) Footer() string {
	reportLink, err := url.Parse(config.GetConfig(config.SaasURL))
	if err != nil {
		logrus.Fatal(fmt.Errorf("failed to parse SaaS URL: %w", err))
	}

	reportLink.Path = k8sPoliciesTemplate

	footer := fmt.Sprintf("Detailed report can be found at\n%s", reportLink)

	return footer
}

// Header is the header columns of the ViolatedResources result.
func (v ViolatedResources) Header() []string {
	return []string{
		namespaceHeader,
		kindHeader,
		nameHeader,
		ruleHeader,
		riskHeader,
		filePathHeader,
	}
}

// Rows returns all the violations of the ViolatedResources result as list of rows.
func (v ViolatedResources) Rows() [][]string {
	result := make([][]string, 0)

	for _, resource := range v.Resources {
		for _, violation := range resource.PolicyViolations {
			result = append(result, []string{
				resource.Namespace,
				resource.Kind,
				resource.Name,
				violation.Rule,
				colorizer.ColorizeRisk(violation.Risk),
				resource.FilePath,
			})
		}
	}

	return result
}

// ValidatedResourcesByPolicy represents a result of resources validation aggregated by policy.
type ValidatedResourcesByPolicy map[string]*ViolatedResources

// ValidatedResource represents a result of one resource validation.
type ValidatedResource struct {
	Scope            `json:",inline"`
	FilePath         string            `json:"file_path"`
	Policy           string            `json:"policy"`
	PolicyViolations []PolicyViolation `json:"policy_violations"`
}

// ValidatedResources response model for the validate resource command.
type ValidatedResources struct {
	Errors            []string
	ViolatedResources []ValidatedResource
}

// GetErrors return a multi-error constructed from all the errors.
func (v ValidatedResources) GetErrors() error {
	var err error
	for _, e := range v.Errors {
		err = multierror.Append(err, fmt.Errorf(e))
	}

	return err
}

// ToValidatedResourcesByPolicy converts the ValidatedResources to ValidatedResourcesByPolicy.
func (v ValidatedResources) ToValidatedResourcesByPolicy() ValidatedResourcesByPolicy {
	result := make(ValidatedResourcesByPolicy)

	for _, violatedResource := range v.ViolatedResources {
		if len(violatedResource.PolicyViolations) == 0 {
			continue
		}

		policy := violatedResource.Policy
		if _, ok := result[policy]; !ok {
			result[policy] = &ViolatedResources{
				Policy:    policy,
				Resources: make([]ValidatedResource, 0),
			}
		}

		result[policy].Resources = append(result[policy].Resources, violatedResource)
	}

	return result
}

// Title is the title of result.
func (v ValidatedResourcesByPolicy) Title() string {
	return "Validation results"
}

// PolicyViolationsCount get the count of violated policies.
func (v ValidatedResourcesByPolicy) PolicyViolationsCount() int {
	count := 0

	for _, value := range v {
		count += value.PolicyViolationsCount()
	}

	return count
}
