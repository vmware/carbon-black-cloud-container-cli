/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package scan

import (
	"fmt"
	"testing"
)

func TestParseAuth(t *testing.T) {
	username, password := "username", "password"
	credential := fmt.Sprintf("%s:%s", username, password)
	testOpt := Option{Credential: credential}

	parsedUsername, parsedPassword := testOpt.parseAuth()
	if username != parsedUsername || password != parsedPassword {
		t.Error("Parsed username and/or password not match")
	}
}
