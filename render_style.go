// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"os"
	"path/filepath"
	"strings"
)

// RenderStyle controls how flags and environment variables are rendered in
// help and documentation output.
type RenderStyle uint8

const (
	// RenderStyleAuto uses the current default behavior.
	RenderStyleAuto RenderStyle = iota
	// RenderStylePOSIX renders POSIX-style output (--flag, $ENV).
	RenderStylePOSIX
	// RenderStyleWindows renders Windows-style output (/flag, %ENV%).
	RenderStyleWindows
	// RenderStyleShell detects shell style from environment variables.
	RenderStyleShell
)

type optionRenderFormat struct {
	shortDelimiter rune
	longDelimiter  string
	nameDelimiter  rune
	envPrefix      string
	envSuffix      string
}

func (p *Parser) resolveFlagRenderStyle() RenderStyle {
	style := p.helpFlagStyle
	if style == RenderStyleAuto && (p.Options&DetectShellFlagStyle) != None {
		style = RenderStyleShell
	}
	return p.resolveRenderStyle(style, true)
}

func (p *Parser) resolveEnvRenderStyle() RenderStyle {
	style := p.helpEnvStyle
	if style == RenderStyleAuto && (p.Options&DetectShellEnvStyle) != None {
		style = RenderStyleShell
	}
	return p.resolveRenderStyle(style, false)
}

func (p *Parser) resolveRenderStyle(style RenderStyle, forFlags bool) RenderStyle {
	switch style {
	case RenderStylePOSIX, RenderStyleWindows:
		return style
	case RenderStyleShell:
		return detectShellRenderStyle()
	default:
		if forFlags {
			if defaultLongOptDelimiter == "--" {
				return RenderStylePOSIX
			}
			return RenderStyleWindows
		}
		if isWindowsRuntime() {
			return RenderStyleWindows
		}
		return RenderStylePOSIX
	}
}

func (p *Parser) optionRenderFormat() optionRenderFormat {
	flagsStyle := p.resolveFlagRenderStyle()
	envStyle := p.resolveEnvRenderStyle()

	format := optionRenderFormat{
		shortDelimiter: defaultShortOptDelimiter,
		longDelimiter:  defaultLongOptDelimiter,
		nameDelimiter:  defaultNameArgDelimiter,
		envPrefix:      "$",
	}

	if flagsStyle == RenderStyleWindows {
		format.shortDelimiter = '/'
		format.longDelimiter = "/"
		format.nameDelimiter = ':'
	} else {
		format.shortDelimiter = '-'
		format.longDelimiter = "--"
		format.nameDelimiter = '='
	}

	if envStyle == RenderStyleWindows {
		format.envPrefix = "%"
		format.envSuffix = "%"
	}

	return format
}

func detectShellRenderStyle() RenderStyle {
	if explicit := shellKind(strings.TrimSpace(os.Getenv("GO_FLAGS_SHELL"))); explicit != RenderStyleAuto {
		return explicit
	}

	if isWindowsRuntime() {
		if parentStyle := detectParentShellStyle(); parentStyle != RenderStyleAuto {
			return parentStyle
		}
	}

	if msystem := strings.TrimSpace(os.Getenv("MSYSTEM")); msystem != "" && isLikelyMSYSSession() {
		return RenderStylePOSIX
	}
	if ostype := strings.ToLower(strings.TrimSpace(os.Getenv("OSTYPE"))); (strings.HasPrefix(ostype, "msys") || strings.HasPrefix(ostype, "cygwin")) && isLikelyMSYSSession() {
		return RenderStylePOSIX
	}

	if isWindowsRuntime() {
		return RenderStyleWindows
	}

	for _, name := range candidateShellNames() {
		if style := shellKind(name); style != RenderStyleAuto {
			return style
		}
	}

	if isWindowsRuntime() {
		return RenderStyleWindows
	}

	return RenderStylePOSIX
}

func candidateShellNames() []string {
	candidates := make([]string, 0, 4)

	// pwsh on Unix can expose these even when SHELL is inherited from POSIX.
	if os.Getenv("POWERSHELL_DISTRIBUTION_CHANNEL") != "" || os.Getenv("PSModulePath") != "" {
		candidates = append(candidates, "pwsh")
	}

	if v := strings.TrimSpace(os.Getenv("SHELL")); v != "" {
		candidates = append(candidates, v)
	}

	if v := strings.TrimSpace(os.Getenv("ComSpec")); v != "" {
		candidates = append(candidates, v)
	}

	if v := strings.TrimSpace(os.Getenv("TERM_PROGRAM")); v != "" {
		candidates = append(candidates, v)
	}

	return candidates
}

func shellKind(raw string) RenderStyle {
	name := strings.ToLower(strings.TrimSpace(raw))
	name = strings.TrimSuffix(filepath.Base(name), ".exe")

	switch name {
	case "sh", "ash", "dash", "bash", "zsh", "fish", "ksh", "mksh":
		return RenderStylePOSIX
	case "cmd", "powershell", "pwsh":
		return RenderStyleWindows
	default:
		return RenderStyleAuto
	}
}

func isWindowsRuntime() bool {
	return os.PathSeparator == '\\'
}

func isLikelyMSYSSession() bool {
	pwd := strings.TrimSpace(os.Getenv("PWD"))
	if strings.HasPrefix(pwd, "/") {
		return true
	}

	path := strings.ToLower(os.Getenv("PATH"))
	if strings.Contains(path, "/usr/bin") || strings.Contains(path, "\\git\\usr\\bin") {
		return true
	}

	return false
}
