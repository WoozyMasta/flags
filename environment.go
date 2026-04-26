// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// TTYInfo describes tty availability for standard streams.
type TTYInfo struct {
	Stdin  bool
	Stdout bool
	Stderr bool
}

// EnvironmentInfo describes detected runtime environment hints.
type EnvironmentInfo struct {
	OS              string
	Shell           string
	ShellStyle      RenderStyle
	CompletionShell CompletionShell
	TerminalColumns int
	TerminalRows    int
	Locale          string
	TTY             TTYInfo
}

// DetectEnvironment returns a snapshot of detected runtime environment hints.
func DetectEnvironment() EnvironmentInfo {
	columns, rows := DetectTerminalSize()

	return EnvironmentInfo{
		OS:              RuntimeOS(),
		Shell:           DetectShell(),
		ShellStyle:      DetectShellStyle(),
		CompletionShell: DetectCompletionShell(),
		TerminalColumns: columns,
		TerminalRows:    rows,
		Locale:          DetectLocale(),
		TTY:             DetectTTY(),
	}
}

// DetectTTY reports tty state for stdin/stdout/stderr in one call.
func DetectTTY() TTYInfo {
	return TTYInfo{
		Stdin:  DetectFileTTY(os.Stdin),
		Stdout: DetectFileTTY(os.Stdout),
		Stderr: DetectFileTTY(os.Stderr),
	}
}

// DetectWriterTTY reports whether writer is a tty-capable output.
// Non-file writers are treated as tty to preserve historical behavior.
func DetectWriterTTY(writer io.Writer) bool {
	if writer == nil {
		return false
	}

	file, ok := writer.(*os.File)
	if !ok {
		return true
	}

	return DetectFileTTY(file)
}

// DetectColorSupport reports whether ANSI colors should be used for a writer.
// Colors are enabled for tty writers, disabled when NO_COLOR is set, and can
// be forced with FORCE_COLOR.
func DetectColorSupport(writer io.Writer) bool {
	if detectNoColor() {
		return false
	}
	if detectForceColor() {
		return true
	}

	return DetectWriterTTY(writer)
}

// DetectFileTTY reports whether an open file descriptor is a tty.
func DetectFileTTY(file *os.File) bool {
	if file == nil {
		return false
	}

	return DetectFDTTY(file.Fd())
}

// DetectFDTTY reports whether file descriptor points to a tty.
func DetectFDTTY(fd uintptr) bool {
	maxInt := int(^uint(0) >> 1)
	if fd > uintptr(maxInt) {
		return false
	}

	return term.IsTerminal(int(fd))
}

func detectNoColor() bool {
	_, exists := os.LookupEnv("NO_COLOR")
	return exists
}

func detectForceColor() bool {
	value, exists := os.LookupEnv("FORCE_COLOR")
	if !exists {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
