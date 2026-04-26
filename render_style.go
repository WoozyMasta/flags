// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"os"
	"path/filepath"
	"runtime"
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
	longDelimiter  string
	envPrefix      string
	envSuffix      string
	shortDelimiter rune
	nameDelimiter  rune
}

var shellStyles = map[string]RenderStyle{
	"sh":         RenderStylePOSIX,
	"ash":        RenderStylePOSIX,
	"dash":       RenderStylePOSIX,
	"bash":       RenderStylePOSIX,
	"zsh":        RenderStylePOSIX,
	"fish":       RenderStylePOSIX,
	"ksh":        RenderStylePOSIX,
	"mksh":       RenderStylePOSIX,
	"cmd":        RenderStyleWindows,
	"powershell": RenderStyleWindows,
	"pwsh":       RenderStyleWindows,
}

// DetectShellStyle returns detected shell rendering style.
func DetectShellStyle() RenderStyle {
	if explicit := shellKind(strings.TrimSpace(os.Getenv("GO_FLAGS_SHELL"))); explicit != RenderStyleAuto {
		return explicit
	}

	if name := DetectShell(); name != "" {
		if style := shellKind(name); style != RenderStyleAuto {
			return style
		}
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

// DetectShell returns detected shell name (for example: bash, zsh, pwsh).
// Returns empty string when shell cannot be detected.
func DetectShell() string {
	if explicit := normalizeShellName(strings.TrimSpace(os.Getenv("GO_FLAGS_SHELL"))); explicit != "" {
		return explicit
	}

	if isWindowsRuntime() {
		if parentName := detectParentShellName(); parentName != "" {
			return parentName
		}
	}

	if msystem := strings.TrimSpace(os.Getenv("MSYSTEM")); msystem != "" && isLikelyMSYSSession() {
		return "bash"
	}
	if ostype := strings.ToLower(strings.TrimSpace(os.Getenv("OSTYPE"))); (strings.HasPrefix(ostype, "msys") || strings.HasPrefix(ostype, "cygwin")) && isLikelyMSYSSession() {
		return "bash"
	}

	for _, name := range candidateShellNames() {
		if normalized := normalizeShellName(name); normalized != "" {
			return normalized
		}
	}

	return ""
}

// DetectCompletionShell returns detected completion shell format.
// Unsupported/unknown shells fallback to bash.
func DetectCompletionShell() CompletionShell {
	switch DetectShell() {
	case string(CompletionShellZsh):
		return CompletionShellZsh
	case string(CompletionShellPwsh), "powershell":
		return CompletionShellPwsh
	default:
		// Unsupported/unknown shells fallback to bash completion script.
		return CompletionShellBash
	}
}

// RuntimeOS returns runtime OS identifier (for example: windows, linux, darwin).
func RuntimeOS() string {
	return runtime.GOOS
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

func normalizeShellName(raw string) string {
	name := strings.ToLower(strings.TrimSpace(raw))
	name = strings.TrimSuffix(filepath.Base(name), ".exe")

	if _, ok := shellStyles[name]; ok {
		return name
	}

	return ""
}

func shellKind(raw string) RenderStyle {
	name := normalizeShellName(raw)
	if name == "" {
		return RenderStyleAuto
	}

	return shellStyles[name]
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
		return DetectShellStyle()
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
