// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build !windows && !plan9 && !appengine && !wasm && !aix

package flags

import (
	"flag"

	"golang.org/x/sys/unix"
)

func getTerminalColumns() int {
	if flag.Lookup("test.v") != nil {
		return defaultTermSize
	}

	ws, err := unix.IoctlGetWinsize(0, unix.TIOCGWINSZ)
	if err != nil {
		return defaultTermSize
	}
	return int(ws.Col)
}
