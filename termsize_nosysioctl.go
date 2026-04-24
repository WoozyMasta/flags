// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build plan9 || appengine || wasm || aix
// +build plan9 appengine wasm aix

package flags

func getTerminalColumns() int {
	return defaultTermSize
}
