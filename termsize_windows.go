// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build windows

package flags

import (
	"flag"
	"syscall"
	"unsafe"
)

type (
	short int16
	word  uint16

	smallRect struct {
		Left   short
		Top    short
		Right  short
		Bottom short
	}

	coord struct {
		X short
		Y short
	}

	consoleScreenBufferInfo struct {
		Size              coord
		CursorPosition    coord
		Attributes        word
		Window            smallRect
		MaximumWindowSize coord
	}
)

var kernel32DLL = syscall.NewLazyDLL("kernel32.dll")
var getConsoleScreenBufferInfoProc = kernel32DLL.NewProc("GetConsoleScreenBufferInfo")

func getError(r1, _ uintptr, lastErr error) error {
	// If the function fails, the return value is zero.
	if r1 == 0 {
		if lastErr != nil {
			return lastErr
		}
		return syscall.EINVAL
	}
	return nil
}

func getStdHandle(stdhandle int) (uintptr, error) {
	handle, err := syscall.GetStdHandle(stdhandle)
	if err != nil {
		return 0, err
	}
	return uintptr(handle), nil
}

// GetConsoleScreenBufferInfo retrieves information about the specified console screen buffer.
// http://msdn.microsoft.com/en-us/library/windows/desktop/ms683171(v=vs.85).aspx
func getConsoleScreenBufferInfo(handle uintptr) (*consoleScreenBufferInfo, error) {
	var info consoleScreenBufferInfo
	if err := getError(getConsoleScreenBufferInfoProc.Call(handle, uintptr(unsafe.Pointer(&info)), 0)); err != nil {
		return nil, err
	}
	return &info, nil
}

func getTerminalColumns() int {
	if flag.Lookup("test.v") != nil {
		return defaultTermSize
	}

	stdoutHandle, err := getStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return defaultTermSize
	}

	info, err := getConsoleScreenBufferInfo(stdoutHandle)
	if err != nil {
		return defaultTermSize
	}

	if info.MaximumWindowSize.X > 0 {
		return int(info.MaximumWindowSize.X)
	}

	return defaultTermSize
}
