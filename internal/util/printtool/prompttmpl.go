/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package printtool

import "github.com/manifoldco/promptui"

// Templates returns the template for interactive prompt ui.
func Templates() *promptui.PromptTemplates {
	return &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . }} ",
		Invalid: "{{ . }} ",
		Success: "{{ . | bold }} ",
	}
}
