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
