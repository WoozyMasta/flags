// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build !windows

package flags

func detectParentShellName() string {
	return ""
}
