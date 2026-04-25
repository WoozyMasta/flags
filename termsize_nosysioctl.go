// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build plan9 || appengine || wasm
// +build plan9 appengine wasm

package flags

func getTerminalColumns() int {
	return defaultTermSize
}
