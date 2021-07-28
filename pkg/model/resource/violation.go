/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource

// PolicyViolation represent a violation of a policy.
type PolicyViolation struct {
	Rule      string                 `json:"rule"`
	Risk      string                 `json:"risk"`
	Violation map[string]interface{} `json:"violation"`
}

// ValidatedResourceResponse represent a resource validation response from the backend.
type ValidatedResourceResponse struct {
	Scope            `json:",inline"`
	Policy           string            `json:"policy"`
	PolicyViolations []PolicyViolation `json:"policy_violations"`
}
