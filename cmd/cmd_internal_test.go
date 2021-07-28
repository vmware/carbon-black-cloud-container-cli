/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"bytes"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestCommandLineHelpText(t *testing.T) {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)

	rootCmd.SetArgs([]string{"--help"})

	if err := Execute(); err != nil {
		t.Errorf("Error while displaying help text: %v", err)
	}
}

func TestCommandLineInvalidSubcommand(t *testing.T) {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)

	rootCmd.SetArgs([]string{"foo"})

	if err := Execute(); err == nil {
		t.Error("Missing expected 'unknown command' error")
	}
}

func TestAuthCommand(t *testing.T) {
	keyring.MockInit()

	testID, testKey := "test_id", "test_key"
	rootCmd.SetArgs([]string{"auth", "set", testID, testKey})

	if err := Execute(); err != nil {
		t.Error(err)
	}

	defaultProfileName := "cbctl_default"

	// check if the api access is saved in the credential store
	savedID, errID := keyring.Get(defaultProfileName, "cb_api_id")
	savedKey, errKey := keyring.Get(defaultProfileName, "cb_api_key")

	if errKey == nil && errID == nil {
		if savedID != testID || savedKey != testKey {
			t.Errorf("Api id/key are not correctly saved in the credential store:\n"+
				"Saved id: %s; Original id: %s\n"+
				"Saved key: %s; Original key: %s\n",
				savedID, testID, savedKey, testKey)
		}
	}
}
