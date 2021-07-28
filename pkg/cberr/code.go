/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package cberr

// Code is the type of machine-readable error code.
type Code int

// ErrCode types.
const (
	UnclassifiedErr Code = iota
	ConfigErr
	HTTPConnectionErr
	HTTPUnsuccessfulResponseErr
	HTTPNotFoundErr
	HTTPNotAllowedErr
	SBOMGenerationErr
	ScanFailedErr
	ValidateFailedErr
	TimeoutErr
	DisplayErr
	PolicyViolationErr
)

//nolint:gomnd
func (c Code) exitCode() int {
	switch c {
	case UnclassifiedErr:
		return 0
	case ConfigErr:
		return 0
	case HTTPConnectionErr:
		return 1
	case HTTPUnsuccessfulResponseErr:
		return 1
	case HTTPNotFoundErr:
		return 1
	case HTTPNotAllowedErr:
		return 1
	case SBOMGenerationErr:
		return 1
	case ScanFailedErr:
		return 1
	case ValidateFailedErr:
		return 1
	case TimeoutErr:
		return 1
	case PolicyViolationErr:
		return 127
	case DisplayErr:
		return 1
	default:
		return 0
	}
}
