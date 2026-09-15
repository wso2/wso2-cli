package prompt

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

type PromptInputOpts struct {
	Message     string
	Description string
	Prompt      string
	Default     string
	IsSecure    bool
	Validate    func(string) error
}

type promptInputMessage struct {
	Opts  PromptInputOpts
	Value *string
}

func NewPromptInputMessage(opts PromptInputOpts, value *string) *promptInputMessage {
	if opts.Prompt == "" {
		opts.Prompt = "> "
	}

	if opts.Default != "" {
		*value = opts.Default
	}

	return &promptInputMessage{
		Opts:  opts,
		Value: value,
	}
}

func (m *promptInputMessage) Prompt() error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	input := huh.NewInput().
		Title(m.Opts.Message).
		Prompt(m.Opts.Prompt).
		Password(m.Opts.IsSecure).
		Description(m.Opts.Description).
		Value(m.Value).
		Inline(true)

	if m.Opts.Validate != nil {
		input.Validate(m.Opts.Validate)
	}

	form := huh.NewForm(
		huh.NewGroup(input),
	).WithTheme(defaultHuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	if m.Opts.IsSecure {
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(
				"%s %s %s",
				utils.CS.Bold(utils.CS.Green("?")),
				utils.CS.Bold(m.Opts.Message),
				utils.CS.Green("*******"),
			),
		)
	} else {
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(
				"%s %s %s",
				utils.CS.Bold(utils.CS.Green("?")),
				utils.CS.Bold(m.Opts.Message),
				utils.CS.Green(fmt.Sprintf("%v", *m.Value)),
			),
		)
	}

	return nil
}
