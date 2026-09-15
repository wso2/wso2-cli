package prompt

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

type PromptConfirmOpts struct {
	Title           string
	AffirmativeText string
	NegativeText    string
	Default         bool
}

type promptConfirmMessage struct {
	Opts  PromptConfirmOpts
	Value *bool
}

func NewPromptConfirmMessage(opts PromptConfirmOpts, value *bool) *promptConfirmMessage {
	if opts.AffirmativeText == "" {
		opts.AffirmativeText = "Yes"
	}

	if opts.NegativeText == "" {
		opts.NegativeText = "No"
	}

	*value = opts.Default

	return &promptConfirmMessage{
		Opts:  opts,
		Value: value,
	}
}

func (m *promptConfirmMessage) Prompt() error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(m.Opts.Title).
				Value(m.Value).
				Affirmative(m.Opts.AffirmativeText).
				Negative(m.Opts.NegativeText),
		),
	).WithTheme(defaultHuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	text := m.Opts.NegativeText

	if *m.Value {
		text = m.Opts.AffirmativeText
	}

	fmt.Fprintln(
		utils.IO.Out,
		fmt.Sprintf(
			"%s %s %s",
			utils.CS.Bold(utils.CS.Green("?")),
			utils.CS.Bold(m.Opts.Title),
			utils.CS.Green(text),
		),
	)

	return nil
}
