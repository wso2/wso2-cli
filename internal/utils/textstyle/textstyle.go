package textstyle

import (
	"fmt"
	"strings"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/mgutz/ansi"
)

type ForegroundAttr int

const (
	BLINK ForegroundAttr = iota
	BOLD
	DIM
	BRIGHT
	INVERT
	STRIKE
	UNDERLINE
)

type BackgroundAttr int

const (
	BRIGHT_BG BackgroundAttr = iota
)

type Color int

const (
	BLACK Color = iota
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
)

type textStyle struct {
	enabled    bool
	foreground string
	background string
	foreAttrs  string
	backAttrs  string
}

func (ts *textStyle) genStyleFunc(text string) string {
	foregroundArr := make([]string, 0)
	backgroundArr := make([]string, 0)

	foregroundStr := ""
	backgroundStr := ""

	if ts.foreground != "" {
		foregroundArr = append(foregroundArr, ts.foreground)
	}
	if ts.background != "" {
		backgroundArr = append(backgroundArr, ts.background)
	}
	if ts.foreAttrs != "" {
		foregroundArr = append(foregroundArr, ts.foreAttrs)
	}
	if ts.backAttrs != "" {
		backgroundArr = append(backgroundArr, ts.backAttrs)
	}

	if len(foregroundArr) > 0 {
		foregroundStr = strings.Join(foregroundArr, "+")
	}

	if len(backgroundArr) > 0 {
		backgroundStr = strings.Join(backgroundArr, "+")
	}

	finalString := ""

	if foregroundStr != "" {
		finalString = foregroundStr
	}
	if backgroundStr != "" {
		if finalString != "" {
			finalString = finalString + ":"
		}
		finalString = finalString + backgroundStr
	}

	return ansi.ColorFunc(finalString)(text)
}

func (ts *textStyle) Text(text string) string {
	if !ts.enabled {
		return text
	}
	return ts.genStyleFunc(text)
}

func (ts *textStyle) Textf(format string, a ...interface{}) string {
	return ts.Text(fmt.Sprintf(format, a...))
}

func CreateTextStyle() *textStyle {
	return &textStyle{
		enabled:    iostreams.System().ColorEnabled(),
		foreground: "default",
	}
}

func (ts *textStyle) SetTextColor(color Color) *textStyle {
	ts.foreground = colorResolve(color)
	return ts
}

func (ts *textStyle) AddTextStyle(attr ForegroundAttr) *textStyle {
	ts.foreAttrs += foregroundAttrResolve(attr)
	return ts
}

func colorResolve(color Color) string {
	switch color {
	case BLACK:
		return "black"
	case RED:
		return "red"
	case GREEN:
		return "green"
	case YELLOW:
		return "yellow"
	case BLUE:
		return "blue"
	case MAGENTA:
		return "magenta"
	case CYAN:
		return "cyan"
	case WHITE:
		return "white"
	default:
		return "white"
	}
}

func foregroundAttrResolve(attr ForegroundAttr) string {
	switch attr {
	case BLINK:
		return "B"
	case BOLD:
		return "b"
	case DIM:
		return "d"
	case BRIGHT:
		return "h"
	case INVERT:
		return "i"
	case STRIKE:
		return "s"
	case UNDERLINE:
		return "u"
	default:
		return ""
	}
}

func backgroundAttrResolve(attr BackgroundAttr) string {
	switch attr {
	case BRIGHT_BG:
		return "bright"
	default:
		return ""
	}
}
