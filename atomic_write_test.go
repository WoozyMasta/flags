// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteFileAtomicallySuccess(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	err := writeFileAtomically(target, func(w io.Writer) error {
		_, err := w.Write([]byte("hello"))
		return err
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("unexpected readdir error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one file in %s after a successful write, got %d: %v", dir, len(entries), entries)
	}
}

func TestWriteFileAtomicallyPreservesExistingContentOnFailure(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	if err := os.WriteFile(target, []byte("original"), 0o600); err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	sentinel := errors.New("write failed")

	err := writeFileAtomically(target, func(w io.Writer) error {
		_, _ = w.Write([]byte("partial"))
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(got) != "original" {
		t.Fatalf("expected original content to survive a failed write, got %q", got)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("unexpected readdir error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected the failed write's temp file to be cleaned up, got %d entries: %v", len(entries), entries)
	}
}

func TestWriteFileAtomicallyUsesRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file mode bits are not meaningful on Windows")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	err := writeFileAtomically(target, func(w io.Writer) error {
		_, err := w.Write([]byte("secret"))
		return err
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("unexpected stat error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected file mode 0600, got %o", perm)
	}
}
