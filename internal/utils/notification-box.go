package utils

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mgutz/ansi"
	"golang.org/x/term"
)

type NotifyBorderColor int

const (
	Black NotifyBorderColor = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Default
)

var NotifyBorderColorMap = map[NotifyBorderColor]string{
	Black:   "black",
	Red:     "red",
	Green:   "green",
	Yellow:  "yellow",
	Blue:    "blue",
	Magenta: "magenta",
	Cyan:    "cyan",
	White:   "white",
	Default: "default",
}

func getTerminalWidth() int {
	w, _, err := term.GetSize(0)

	if err != nil {
		return 80

	}

	if w > 80 {
		return 80
	}

	return w

}

func CreateNotificationBox(borderCol NotifyBorderColor) lipgloss.Style {
	var color string
	if _, ok := NotifyBorderColorMap[borderCol]; ok {
		color = NotifyBorderColorMap[borderCol]
	} else {
		color = NotifyBorderColorMap[Default]
	}

	return lipgloss.NewStyle().
		Padding(1).
		Margin(2).
		Width(getTerminalWidth() - 6).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.ANSIColor(ansi.Colors[color])).
		Align(lipgloss.Left)
}
