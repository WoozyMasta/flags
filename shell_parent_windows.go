// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

//go:build windows

package flags

import (
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type processInfo struct {
	parentID uint32
	name     string
}

func detectParentShellStyle() RenderStyle {
	procs, ok := snapshotProcesses()
	if !ok {
		return RenderStyleAuto
	}

	pid := uint32(os.Getpid())

	// Walk a few levels up to find a known shell process.
	for range 12 {
		current, ok := procs[pid]
		if !ok || current.parentID == 0 || current.parentID == pid {
			return RenderStyleAuto
		}

		parent, ok := procs[current.parentID]
		if !ok {
			return RenderStyleAuto
		}

		if style := shellKind(parent.name); style != RenderStyleAuto {
			return style
		}

		pid = current.parentID
	}

	return RenderStyleAuto
}

func snapshotProcesses() (map[uint32]processInfo, bool) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, false
	}
	defer windows.CloseHandle(snapshot)

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
