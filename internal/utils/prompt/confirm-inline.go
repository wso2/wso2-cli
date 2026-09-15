package prompt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

type YesNoPromptOpts struct {
	Title   string
	Default bool
}

type yesNoPrompt struct {
	opts  YesNoPromptOpts
	value *bool
}

func (p *yesNoPrompt) Prompt() error {
	inputText := ""

	placeholder := ""
	if *p.value {
		placeholder = "(Y/n)"
	} else {
		placeholder = "(y/N)"
	}

	input := huh.NewInput().
		Title(fmt.Sprintf("%s %s", p.opts.Title, placeholder)).
		Value(&inputText).
		Validate(validateInput).
		Prompt(" ").
		Inline(true)

	form := huh.NewForm(
		huh.NewGroup(
			input,
		),
	).WithTheme(defaultHuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	output := ""

	if resolveInput(inputText, p.opts.Default) {
		*p.value = true
		output = "Yes"
	} else {
		*p.value = false
		output = "No"
	}

	fmt.Fprintln(
		utils.IO.Out,
		fmt.Sprintf(
			"%s %s %s",
			utils.CS.Bold(utils.CS.Green("?")),
			utils.CS.Bold(p.opts.Title),
			utils.CS.Green(output),
		),
	)

	return nil
}

func validateInput(input string) error {
	allowedList := []string{"y", "n", "yes", "no", ""}

	for _, v := range allowedList {
		if strings.ToLower(input) == v {
			return nil
		}
	}

	return fmt.Errorf("Invalid input. Please enter 'y' or 'n'")
}

func resolveInput(input string, defaultValue bool) bool {
	if input == "" {
		return defaultValue
	}

	if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
		return true
	}

	return false
}

func NewYesNoPrompt(opts YesNoPromptOpts, value *bool) *yesNoPrompt {
	*value = opts.Default
	return &yesNoPrompt{opts, value}
}
