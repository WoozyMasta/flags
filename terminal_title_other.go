// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build !windows

package flags

import (
	"fmt"
	"os"
)

func setTerminalTitle(title string) error {
	_, err := fmt.Fprintf(os.Stderr, "\x1b]0;%s\x07", title)
	return err
}
