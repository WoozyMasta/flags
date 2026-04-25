// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build !plan9 && !appengine && !wasm

package flags

import (
	"os"

	"golang.org/x/term"
)

func getTerminalColumns() int {
	for _, file := range []*os.File{os.Stdout, os.Stderr, os.Stdin} {
		if file == nil {
			continue
		}

		fd, ok := terminalFD(file)
		if !ok {
			continue
		}

		width, _, err := term.GetSize(fd)
		if err == nil && width > 0 {
			return width
		}
	}

	return defaultTermSize
}

func terminalFD(file *os.File) (int, bool) {
	fd := file.Fd()
	maxInt := int(^uint(0) >> 1)
	if fd > uintptr(maxInt) {
		return 0, false
	}

	return int(fd), true //nolint:gosec // fd is checked against max int above.
}
