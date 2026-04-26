// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build windows

package flags

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type processInfo struct {
	name     string
	parentID uint32
}

func detectParentShellName() string {
	procs, ok := snapshotProcesses()
	if !ok {
		return ""
	}

	pid := windows.GetCurrentProcessId()

	// Walk a few levels up to find a known shell process.
	for range 12 {
		current, ok := procs[pid]
		if !ok || current.parentID == 0 || current.parentID == pid {
			return ""
		}

		parent, ok := procs[current.parentID]
		if !ok {
			return ""
		}

		if name := normalizeShellName(parent.name); name != "" {
			return name
		}

		pid = current.parentID
	}

	return ""
}

func snapshotProcesses() (map[uint32]processInfo, bool) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, false
	}
	defer func() {
		_ = windows.CloseHandle(snapshot)
	}()

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, false
	}

	procs := make(map[uint32]processInfo, 256)
	for {
		name := windows.UTF16ToString(entry.ExeFile[:])
		procs[entry.ProcessID] = processInfo{
			parentID: entry.ParentProcessID,
			name:     strings.TrimSpace(name),
		}

		if err := windows.Process32Next(snapshot, &entry); err != nil {
			break
		}
	}

	return procs, true
}
