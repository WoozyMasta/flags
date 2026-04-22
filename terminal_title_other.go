// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

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
