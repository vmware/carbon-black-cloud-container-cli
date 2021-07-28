/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package scan

import (
	"strings"
)

// Option is the option used for image related cmd.
type Option struct {
	// ForceScan is the option whether to force scan an image no matter it is scanned or not.
	ForceScan bool
	// UseDockerDaemon is whether to use docker daemon to pull the image
	UseDockerDaemon bool
	// Credential is the auth string used for login to registry, format: USERNAME[:PASSWORD]
	Credential string
	// ShouldCleanup is whether to delete the docker image pulled by docker
	ShouldCleanup bool
	// FullTag is the tag set to override in the image
	FullTag string
	// Timeout is the duration (second) for the scan process
	Timeout int
}

func (o Option) parseAuth() (username string, password string) {
	up := strings.SplitN(o.Credential, ":", 2)

	username = up[0]

	if len(up) == 2 {
		password = up[1]
	}

	return username, password
}
