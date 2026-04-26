// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build plan9 || appengine || wasm
// +build plan9 appengine wasm

package flags

// DetectTerminalSize returns detected terminal size in columns and rows.
func DetectTerminalSize() (int, int) {
	return defaultTermSize, defaultTermRows
}
