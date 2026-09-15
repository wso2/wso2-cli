package prompt

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

type PromptSelectOpts[T comparable] struct {
	Message     string
	Description string
	Default     T
	Values      []T
	Options     []huh.Option[T] `validate:"required_without=Values,excluded_with=Values"`
}

type promptSelectMessage[T comparable] struct {
	Opts  PromptSelectOpts[T]
	Value *T
}

func (m *promptSelectMessage[T]) Prompt() error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	opts := make([]huh.Option[T], 0)

	if m.Opts.Values != nil {
		opts = huh.NewOptions(m.Opts.Values...)
	} else {
		opts = m.Opts.Options
	}

	selector := huh.NewSelect[T]().
		Title(m.Opts.Message).
		Options(opts...).
		Description(m.Opts.Description).
		Value(m.Value)

	form := huh.NewForm(
		huh.NewGroup(
			selector,
		),
	).WithTheme(defaultHuhTheme)

	err := form.Run()

	if err != nil {
		return err
	}

	selectedKey := ""

	if m.Opts.Values != nil {
		selectedKey = fmt.Sprintf("%v", *m.Value)
	} else {
		for _, val := range opts {
			if val.Value == *m.Value {
				selectedKey = val.Key
				break
			}
		}
	}

	fmt.Fprintln(
		utils.IO.Out,
		fmt.Sprintf(
			"%s %s %s",
			utils.CS.Bold(utils.CS.Green("?")),
			utils.CS.Bold(m.Opts.Message),
			utils.CS.Green(selectedKey),
		))
	return nil
}

func NewPromptSelectMessage[T comparable](opts PromptSelectOpts[T], value *T) *promptSelectMessage[T] {
	*value = opts.Default
	return &promptSelectMessage[T]{
		Opts:  opts,
		Value: value,
	}
}
