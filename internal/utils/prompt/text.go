package prompt

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

type PromptTextInputOpts struct {
	Message  string
	Prompt   string
	Validate func(string) error
}

type promptTextInputMessage struct {
	Opts  PromptTextInputOpts
	Value *string
}

func NewPromptTextInputMessage(opts PromptTextInputOpts, value *string) *promptTextInputMessage {
	return &promptTextInputMessage{
		Opts:  opts,
		Value: value,
	}
}

func (m *promptTextInputMessage) Prompt() error {

	textInput := huh.NewText().
		Title(m.Opts.Message).
		Value(m.Value)

	if m.Opts.Validate != nil {
		textInput.Validate(m.Opts.Validate)
	}

	form := huh.NewForm(
		huh.NewGroup(
			textInput,
		),
	).WithTheme(defaultHuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	fmt.Fprintln(
		utils.IO.Out,
		fmt.Sprintf(
			"%s %s %s",
			utils.CS.Bold(utils.CS.Green("?")),
			utils.CS.Bold(m.Opts.Message),
			utils.CS.Green(fmt.Sprintf("%v", *m.Value)),
		),
	)

	return nil
}
