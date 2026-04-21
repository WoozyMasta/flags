// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

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
