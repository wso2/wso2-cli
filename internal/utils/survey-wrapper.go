package utils

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/config"
)

func PromptSelection[T any](selectPrompt *survey.Select, selection *T, opts ...survey.AskOpt) error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	err := survey.AskOne(selectPrompt, selection, opts...)

	if err != nil {
		if err.Error() == "interrupt" {
			return internal.ErrUserInterrupted
		}

		return err
	}

	return nil
}

func PromptConfirm[T any](confirmPrompt *survey.Confirm, selection *T, opts ...survey.AskOpt) error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	err := survey.AskOne(confirmPrompt, selection, opts...)

	if err != nil {
		if err.Error() == "interrupt" {
			return internal.ErrUserInterrupted
		}

		return err
	}

	return nil
}

func PromptInput[T any](inputPrompt *survey.Input, inputVar *T, opts ...survey.AskOpt) error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	err := survey.AskOne(inputPrompt, inputVar, opts...)

	if err != nil {
		if err.Error() == "interrupt" {
			return internal.ErrUserInterrupted
		}

		return err
	}

	return nil
}

func PromptMultiLineInput[T any](inputPrompt *survey.Multiline, inputVar *T, opts ...survey.AskOpt) error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	err := survey.AskOne(inputPrompt, inputVar, opts...)

	if err != nil {
		if err.Error() == "interrupt" {
			return internal.ErrUserInterrupted
		}

		return err
	}

	return nil
}

func PromptPassword[T any](inputPrompt *survey.Password, inputVar *T, opts ...survey.AskOpt) error {
	if config.NonInteractive {
		return internal.ErrNonInteractive
	}

	err := survey.AskOne(inputPrompt, inputVar, opts...)

	if err != nil {
		if err.Error() == "interrupt" {
			return internal.ErrUserInterrupted
		}

		return err
	}

	return nil
}
