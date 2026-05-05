// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build windows

package flags

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
)

func colorOutputWriter(w io.Writer) io.Writer {
	file, ok := w.(*os.File)
	if !ok {
		return w
	}

	return colorable.NewColorable(file)
}
