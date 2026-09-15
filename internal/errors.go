package internal

import "errors"

var (
	ErrUserInterrupted = errors.New("user interrupted")
	ErrNonInteractive  = errors.New("non-interactive mode: missing required parameter")
)
