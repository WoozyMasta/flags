// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestLaunchedFromExplorerFalseUnderTestRunner(t *testing.T) {
	// go test is never launched by double-clicking it in Explorer
	// (its parent is the go tool, a shell, or a test runner),
	// so this should reliably be false on every platform, including Windows.
	if LaunchedFromExplorer() {
		t.Fatalf("expected LaunchedFromExplorer to be false when run via `go test`")
	}
}

func TestWaitForEnterWritesPromptAndBlocksUntilNewline(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("\n")

	if err := WaitForEnter(&out, in, "Press Enter to exit..."); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := out.String(); got != "Press Enter to exit..." {
		t.Fatalf("unexpected prompt output: %q", got)
	}
}

func TestWaitForEnterEmptyPromptWritesNothing(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("first\nsecond\n")

	if err := WaitForEnter(&out, in, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("expected no output for empty prompt, got %q", out.String())
	}
}

func TestWaitForEnterTreatsEOFAsSuccess(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("")

	if err := WaitForEnter(&out, in, "prompt"); err != nil {
		t.Fatalf("expected EOF to be treated as success, got: %v", err)
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWaitForEnterReturnsWriteError(t *testing.T) {
	err := WaitForEnter(errWriter{}, strings.NewReader("\n"), "prompt")
	if err == nil || err.Error() != "write failed" {
		t.Fatalf("expected write error to propagate, got: %v", err)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestWaitForEnterReturnsReadError(t *testing.T) {
	var out bytes.Buffer

	err := WaitForEnter(&out, errReader{}, "")
	if err == nil || err.Error() != "read failed" {
		t.Fatalf("expected read error to propagate, got: %v", err)
	}
}
