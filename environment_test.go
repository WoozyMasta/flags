// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"os"
	"testing"
)

func TestEnvironmentAPIDetectShellFromOverride(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "/bin/zsh")

	if got := DetectShell(); got != "zsh" {
		t.Fatalf("expected zsh shell, got %q", got)
	}
	if got := DetectShellStyle(); got != RenderStylePOSIX {
		t.Fatalf("expected POSIX shell style, got %v", got)
	}
	if got := DetectCompletionShell(); got != CompletionShellZsh {
		t.Fatalf("expected zsh completion shell, got %q", got)
	}
}

func TestEnvironmentAPIDetectEnvironmentSnapshot(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "pwsh")

	env := DetectEnvironment()
	if env.OS == "" {
		t.Fatal("expected non-empty runtime OS")
	}
	if env.Shell != "pwsh" {
		t.Fatalf("expected pwsh shell, got %q", env.Shell)
	}
	if env.CompletionShell != CompletionShellPwsh {
		t.Fatalf("expected pwsh completion shell, got %q", env.CompletionShell)
	}
	if env.TerminalColumns <= 0 {
		t.Fatalf("expected positive terminal width, got %d", env.TerminalColumns)
	}
	if env.TerminalRows <= 0 {
		t.Fatalf("expected positive terminal height, got %d", env.TerminalRows)
	}
	wantTTY := DetectTTY()
	if env.TTY != wantTTY {
		t.Fatalf("unexpected tty snapshot: got=%+v want=%+v", env.TTY, wantTTY)
	}
}

func TestEnvironmentAPIDetectColorSupport(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	_ = os.Unsetenv("NO_COLOR")
	_ = os.Unsetenv("FORCE_COLOR")

	var buf bytes.Buffer
	if !DetectColorSupport(&buf) {
		t.Fatal("expected non-file writer to preserve tty behavior")
	}

	_ = os.Setenv("FORCE_COLOR", "1")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	defer r.Close()
	defer w.Close()
	if !DetectColorSupport(w) {
		t.Fatal("expected FORCE_COLOR to enable colors on non-tty writer")
	}

	_ = os.Setenv("NO_COLOR", "1")
	if DetectColorSupport(&buf) {
		t.Fatal("expected NO_COLOR to disable colors")
	}
}

func TestEnvironmentAPIDetectWriterTTY(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	defer r.Close()
	defer w.Close()

	if DetectWriterTTY(w) {
		t.Fatal("expected pipe writer to be non-tty")
	}

	var buf bytes.Buffer
	if !DetectWriterTTY(&buf) {
		t.Fatal("expected non-file writer to be treated as tty")
	}
}

func TestEnvironmentAPIDetectTTY(t *testing.T) {
	tty := DetectTTY()
	if tty != (TTYInfo{
		Stdin:  DetectFileTTY(os.Stdin),
		Stdout: DetectFileTTY(os.Stdout),
		Stderr: DetectFileTTY(os.Stderr),
	}) {
		t.Fatalf("unexpected tty state: %+v", tty)
	}
}

func TestEnvironmentAPIDetectFileTTYNil(t *testing.T) {
	if DetectFileTTY(nil) {
		t.Fatal("expected nil file to be non-tty")
	}
}

func TestEnvironmentAPIDetectFDTTY(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	defer r.Close()
	defer w.Close()

	if DetectFDTTY(r.Fd()) {
		t.Fatal("expected pipe read descriptor to be non-tty")
	}
	if DetectFDTTY(w.Fd()) {
		t.Fatal("expected pipe write descriptor to be non-tty")
	}
	if DetectFDTTY(^uintptr(0)) {
		t.Fatal("expected invalid descriptor to be non-tty")
	}
}
