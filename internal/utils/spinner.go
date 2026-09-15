package utils

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/cli/cli/v2/pkg/iostreams"
)

func CreateSpinner(msg string, finalMsg string) *spinner.Spinner {
	// if msg does not start with a space, add one
	if msg[0] != ' ' {
		msg = " " + msg
	}
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithSuffix(msg), spinner.WithFinalMSG(finalMsg))
	s.Writer = iostreams.System().ErrOut
	return s
}
