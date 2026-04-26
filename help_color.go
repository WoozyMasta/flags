// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"io"
	"strconv"
	"strings"
)

// ANSIColor is an ANSI 16-color palette entry.
type ANSIColor uint8

const (
	// ColorBlack is ANSI black.
	ColorBlack ANSIColor = iota
	// ColorRed is ANSI red.
	ColorRed
	// ColorGreen is ANSI green.
	ColorGreen
	// ColorYellow is ANSI yellow.
	ColorYellow
	// ColorBlue is ANSI blue.
	ColorBlue
	// ColorMagenta is ANSI magenta.
	ColorMagenta
	// ColorCyan is ANSI cyan.
	ColorCyan
	// ColorWhite is ANSI white.
	ColorWhite
	// ColorBrightBlack is ANSI bright black.
	ColorBrightBlack
	// ColorBrightRed is ANSI bright red.
	ColorBrightRed
	// ColorBrightGreen is ANSI bright green.
	ColorBrightGreen
	// ColorBrightYellow is ANSI bright yellow.
	ColorBrightYellow
	// ColorBrightBlue is ANSI bright blue.
	ColorBrightBlue
	// ColorBrightMagenta is ANSI bright magenta.
	ColorBrightMagenta
	// ColorBrightCyan is ANSI bright cyan.
	ColorBrightCyan
	// ColorBrightWhite is ANSI bright white.
	ColorBrightWhite
)

// HelpTextStyle describes color and font style attributes for help output.
type HelpTextStyle struct {
	FG        ANSIColor
	BG        ANSIColor
	UseFG     bool
	UseBG     bool
	Bold      bool
	Italic    bool
	Underline bool
}

// HelpColorScheme configures help color roles.
type HelpColorScheme struct {
	BaseText                HelpTextStyle
	LongDescription         HelpTextStyle
	SubcommandOptionsHeader HelpTextStyle
	OptionShort             HelpTextStyle
	OptionLong              HelpTextStyle
	OptionDesc              HelpTextStyle
	OptionEnv               HelpTextStyle
	OptionDefault           HelpTextStyle
	OptionChoices           HelpTextStyle
	UsageHeader             HelpTextStyle
	UsageText               HelpTextStyle
	CommandsHeader          HelpTextStyle
	CommandName             HelpTextStyle
	CommandDesc             HelpTextStyle
	CommandAliases          HelpTextStyle
	ArgumentsHeader         HelpTextStyle
	ArgumentName            HelpTextStyle
	ArgumentDesc            HelpTextStyle
	GroupHeader             HelpTextStyle
}

// ErrorColorScheme configures parser error color roles.
type ErrorColorScheme struct {
	Warning  HelpTextStyle
	Critical HelpTextStyle
}

// DefaultHelpColorScheme returns the default built-in color scheme.
func DefaultHelpColorScheme() HelpColorScheme {
	return HelpColorScheme{
		BaseText:                HelpTextStyle{},
		LongDescription:         HelpTextStyle{},
		SubcommandOptionsHeader: HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		OptionShort:             HelpTextStyle{UseFG: true, FG: ColorBrightCyan, Bold: true},
		OptionLong:              HelpTextStyle{UseFG: true, FG: ColorCyan, Bold: true},
		OptionDesc:              HelpTextStyle{},
		OptionEnv:               HelpTextStyle{UseFG: true, FG: ColorBrightBlue},
		OptionDefault:           HelpTextStyle{UseFG: true, FG: ColorBrightMagenta},
		OptionChoices:           HelpTextStyle{UseFG: true, FG: ColorBrightGreen},
		UsageHeader:             HelpTextStyle{UseFG: true, FG: ColorBrightYellow, Bold: true},
		UsageText:               HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		CommandsHeader:          HelpTextStyle{UseFG: true, FG: ColorBrightYellow},
		CommandName:             HelpTextStyle{UseFG: true, FG: ColorBrightYellow},
		CommandDesc:             HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		CommandAliases:          HelpTextStyle{UseFG: true, FG: ColorBrightBlack},
		ArgumentsHeader:         HelpTextStyle{UseFG: true, FG: ColorBrightYellow},
		ArgumentName:            HelpTextStyle{UseFG: true, FG: ColorBrightCyan, Bold: true},
		ArgumentDesc:            HelpTextStyle{},
		GroupHeader:             HelpTextStyle{UseFG: true, FG: ColorBrightWhite, Bold: true},
	}
}

// HighContrastHelpColorScheme returns a high-contrast built-in color scheme.
func HighContrastHelpColorScheme() HelpColorScheme {
	return HelpColorScheme{
		BaseText:                HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		LongDescription:         HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		SubcommandOptionsHeader: HelpTextStyle{UseFG: true, FG: ColorBrightBlue},
		OptionShort:             HelpTextStyle{UseFG: true, FG: ColorBrightYellow, Bold: true},
		OptionLong:              HelpTextStyle{UseFG: true, FG: ColorBrightYellow, Bold: true},
		OptionDesc:              HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		OptionEnv:               HelpTextStyle{UseFG: true, FG: ColorBrightCyan, Bold: true},
		OptionDefault:           HelpTextStyle{UseFG: true, FG: ColorBrightMagenta, Bold: true},
		OptionChoices:           HelpTextStyle{UseFG: true, FG: ColorBrightGreen, Bold: true},
		UsageHeader:             HelpTextStyle{UseFG: true, FG: ColorBrightBlue, Bold: true},
		UsageText:               HelpTextStyle{UseFG: true, FG: ColorBrightWhite, Bold: true},
		CommandsHeader:          HelpTextStyle{UseFG: true, FG: ColorBrightBlue},
		CommandName:             HelpTextStyle{UseFG: true, FG: ColorBrightBlue, Bold: true},
		CommandDesc:             HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		CommandAliases:          HelpTextStyle{UseFG: true, FG: ColorBrightCyan},
		ArgumentsHeader:         HelpTextStyle{UseFG: true, FG: ColorBrightBlue},
		ArgumentName:            HelpTextStyle{UseFG: true, FG: ColorBrightYellow, Bold: true},
		ArgumentDesc:            HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		GroupHeader:             HelpTextStyle{UseFG: true, FG: ColorBrightWhite, Bold: true, Underline: true},
	}
}

// DefaultErrorColorScheme returns default parser error color roles.
func DefaultErrorColorScheme() ErrorColorScheme {
	return ErrorColorScheme{
		Warning:  HelpTextStyle{UseFG: true, FG: ColorBrightYellow, Bold: true},
		Critical: HelpTextStyle{UseFG: true, FG: ColorBrightRed, Bold: true},
	}
}

// HighContrastErrorColorScheme returns a high-contrast parser error scheme.
func HighContrastErrorColorScheme() ErrorColorScheme {
	return ErrorColorScheme{
		Warning:  HelpTextStyle{UseFG: true, FG: ColorBlack, UseBG: true, BG: ColorBrightYellow, Bold: true},
		Critical: HelpTextStyle{UseFG: true, FG: ColorWhite, UseBG: true, BG: ColorRed, Bold: true},
	}
}

func ansiCode(color ANSIColor, bg bool) int {
	if color <= ColorWhite {
		if bg {
			return int(color) + 40
		}
		return int(color) + 30
	}

	if bg {
		return int(color-ColorBrightBlack) + 100
	}
	return int(color-ColorBrightBlack) + 90
}

func mergeHelpStyle(base HelpTextStyle, overlay HelpTextStyle) HelpTextStyle {
	out := base
	if overlay.UseFG {
		out.FG = overlay.FG
		out.UseFG = true
	}
	if overlay.UseBG {
		out.BG = overlay.BG
		out.UseBG = true
	}
	out.Bold = out.Bold || overlay.Bold
	out.Italic = out.Italic || overlay.Italic
	out.Underline = out.Underline || overlay.Underline
	return out
}

func applyHelpTextStyle(text string, style HelpTextStyle) string {
	if text == "" {
		return ""
	}

	prefix := helpStylePrefix(style)
	if prefix == "" {
		return text
	}

	return prefix + text + "\x1b[0m"
}

func helpStylePrefix(style HelpTextStyle) string {
	var codes []string

	if style.Bold {
		codes = append(codes, "1")
	}
	if style.Italic {
		codes = append(codes, "3")
	}
	if style.Underline {
		codes = append(codes, "4")
	}
	if style.UseFG {
		codes = append(codes, strconv.Itoa(ansiCode(style.FG, false)))
	}
	if style.UseBG {
		codes = append(codes, strconv.Itoa(ansiCode(style.BG, true)))
	}
	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

func (p *Parser) colorizeHelp(text string, role HelpTextStyle) string {
	if text == "" || (p.Options&ColorHelp) == None || !p.helpColorEnabled {
		return text
	}
	style := mergeHelpStyle(p.helpColorScheme.BaseText, role)
	styled := applyHelpTextStyle(text, style)

	basePrefix := helpStylePrefix(p.helpColorScheme.BaseText)
	if basePrefix != "" && strings.HasSuffix(styled, "\x1b[0m") {
		styled = strings.TrimSuffix(styled, "\x1b[0m") + basePrefix
	}

	return styled
}

func (p *Parser) colorizeError(err error, text string, writer io.Writer) string {
	if text == "" || (p.Options&ColorErrors) == None {
		return text
	}
	if !DetectColorSupport(writer) {
		return text
	}

	flagsErr, ok := err.(*Error)
	if !ok {
		return applyHelpTextStyle(text, p.errorColorScheme.Critical)
	}

	if flagsErr.Type == ErrHelp || flagsErr.Type == ErrVersion {
		return text
	}

	if flagsErr.Type.IsWarning() {
		return applyHelpTextStyle(text, p.errorColorScheme.Warning)
	}

	return applyHelpTextStyle(text, p.errorColorScheme.Critical)
}
