// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build windows

package flags

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func detectLaunchedFromExplorer() bool {
	procs, ok := snapshotProcesses()
	if !ok {
		return false
	}

	current, ok := procs[windows.GetCurrentProcessId()]
	if !ok {
		return false
	}

	parent, ok := procs[current.parentID]
	if !ok {
		return false
	}

	name := strings.ToLower(strings.TrimSuffix(filepath.Base(parent.name), ".exe"))
	return name == "explorer"
}
