// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bufio"
	"io"
)

// LaunchedFromExplorer reports whether the current process
// was likely started by double-clicking it in Windows Explorer
// (its immediate parent process is explorer.exe), as opposed to being run from a terminal.
// It always returns false on non-Windows platforms.
func LaunchedFromExplorer() bool {
	return detectLaunchedFromExplorer()
}

// WaitForEnter writes prompt to w (if non-empty) and blocks until a line is read from r,
// so a console window does not close before the user reads preceding output.
// It is typically called with os.Stdout/os.Stdin after checking [LaunchedFromExplorer].
// Returns nil on a clean read or EOF; any other read error is returned.
func WaitForEnter(w io.Writer, r io.Reader, prompt string) error {
	if prompt != "" {
		if _, err := io.WriteString(w, prompt); err != nil {
			return err
		}
	}

	_, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}
