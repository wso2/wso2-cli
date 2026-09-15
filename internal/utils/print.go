package utils

import "fmt"

// Create func printInfo which wraps fmt.Fprintf(IO.Out, format, a...)
func PrintInfo(format string, a ...interface{}) {
	fmt.Fprintf(IO.Out, format, a...)
}

// Create func printError which wraps fmt.Fprintf(IO.ErrOut, format, a...)
func PrintError(format string, a ...interface{}) {
	fmt.Fprintf(IO.ErrOut, format, a...)
}
