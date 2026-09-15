package prompt

import (
	"errors"
	"strconv"

	"github.com/charmbracelet/huh"
)

var defaultHuhTheme = huh.ThemeBase16()

func ValidateNotEmpty(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	return nil
}

func ValidateNotEmptyInt(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	if _, err := strconv.Atoi(s); err != nil {
		return errors.New("not a number")
	}

	return nil
}
