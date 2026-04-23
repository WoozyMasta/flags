// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build windows

package flags

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func setTerminalTitle(title string) error {
	ptr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}

	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTitleW")

	ret, _, callErr := proc.Call(uintptr(unsafe.Pointer(ptr)))
	if ret == 0 {
		if callErr != nil && callErr != windows.ERROR_SUCCESS {
			return callErr
		}
		return ErrSetConsoleTitleFailed
	}

	return nil
}
