package utils

import "github.com/cli/cli/v2/pkg/iostreams"

var (
	IO = iostreams.System()
	CS = IO.ColorScheme()
)
