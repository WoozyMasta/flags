// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"io"
	"os"
	"path/filepath"
)

// writeFileAtomically writes the content produced by write to a temporary file
// in the same directory as filename, then renames it into place,
// so a reader never observes a partially-written file and a failed write never clobbers an existing file.
// The temporary file is created with mode 0600 (config files may hold secret option values);
// it is removed if anything fails before the rename.
func writeFileAtomically(filename string, write func(io.Writer) error) error {
	dir := filepath.Dir(filename)

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(filename)+".*.tmp")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()
	renamed := false

	defer func() {
		if !renamed {
			_ = os.Remove(tmpName)
		}
	}()

	if err := write(tmp); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, filename); err != nil {
		return err
	}

	renamed = true
	return nil
}
