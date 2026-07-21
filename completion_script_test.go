package flags

import (
	"errors"
	"strings"
	"testing"
)

func TestWriteNamedCompletionBash(t *testing.T) {
	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteNamedCompletion(&out, CompletionShellBash, "my-app"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "complete -F _my_app 'my-app'") {
		t.Fatalf("unexpected bash script:\n%s", got)
	}
	if !strings.Contains(got, "compopt -o nospace") {
		t.Fatalf("expected nospace handling in bash script:\n%s", got)
	}
}

func TestWriteNamedCompletionZsh(t *testing.T) {
	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteNamedCompletion(&out, CompletionShellZsh, "my-app"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "#compdef my-app") {
		t.Fatalf("missing zsh compdef header:\n%s", got)
	}

	if !strings.Contains(got, "compdef _my_app 'my-app'") {
		t.Fatalf("missing zsh compdef command:\n%s", got)
	}
	if !strings.Contains(got, "compadd -S '' --") {
		t.Fatalf("missing zsh nospace completion branch:\n%s", got)
	}
}

func TestWriteNamedCompletionPwsh(t *testing.T) {
	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteNamedCompletion(&out, CompletionShellPwsh, "my-app"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Register-ArgumentCompleter") {
		t.Fatalf("missing pwsh completer registration:\n%s", got)
	}
	if !strings.Contains(got, "GO_FLAGS_COMPLETION") {
		t.Fatalf("missing GO_FLAGS_COMPLETION integration in pwsh script:\n%s", got)
	}
}

func TestWriteCompletionUsesParserName(t *testing.T) {
	p := NewNamedParser("tool-name", None)
	var out strings.Builder

	if err := p.WriteCompletion(&out, CompletionShellBash); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "complete -F _tool_name 'tool-name'") {
		t.Fatalf("unexpected completion output:\n%s", out.String())
	}
}

func TestWriteNamedCompletionErrors(t *testing.T) {
	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteNamedCompletion(&out, CompletionShellBash, ""); err == nil {
		t.Fatal("expected error for empty command name")
	}

	if err := p.WriteNamedCompletion(&out, CompletionShell("fish"), "app"); err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}

// TestWriteNamedCompletionRejectsDangerousCommandNames is a corpus of command names
// carrying shell metacharacters that must never reach a generated completion script unescaped.
// WriteNamedCompletion must reject every one of them with ErrInvalidCommandName
// rather than emit a script that could execute the embedded payload if sourced.
// None of these contain a path separator;
// see TestWriteNamedCompletionStripsDirectoryComponents for how a name that does is handled
// (filepath.Base runs first, so a malicious prefix ahead of a "/" is discarded, not smuggled through).
func TestWriteNamedCompletionRejectsDangerousCommandNames(t *testing.T) {
	corpus := []string{
		"app;touch",
		"app`touch`",
		"app$(touch)",
		"app\ntouch",
		"app'; touch; '",
		`app"; touch; "`,
		"app&&touch",
		"app|touch",
		"app\x00pwn",
		"app pwn",
		"$env:PWNED",
		"app${IFS}pwn",
	}

	for _, name := range corpus {
		t.Run(name, func(t *testing.T) {
			p := NewNamedParser("app", None)

			for _, shell := range []CompletionShell{CompletionShellBash, CompletionShellZsh, CompletionShellPwsh} {
				var out strings.Builder
				err := p.WriteNamedCompletion(&out, shell, name)
				if err == nil {
					t.Fatalf("shell %s: expected error for dangerous command name %q, got script:\n%s", shell, name, out.String())
				}
				if !errors.Is(err, ErrInvalidCommandName) {
					t.Fatalf("shell %s: expected ErrInvalidCommandName for %q, got: %v", shell, name, err)
				}
				if out.Len() != 0 {
					t.Fatalf("shell %s: expected no output written for rejected command name %q, got:\n%s", shell, name, out.String())
				}
			}
		})
	}
}

// TestWriteNamedCompletionStripsDirectoryComponents mirrors a realistic exploitation path:
// an attacker able to control argv[0] (e.g. by placing a maliciously-named executable)
// can't smuggle shell metacharacters through via a directory component either,
// since filepath.Base discards everything up to and including the last path separator before the allowlist check.
func TestWriteNamedCompletionStripsDirectoryComponents(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"/usr/local/bin/my-app", "my-app"},
		{"../../etc/passwd", "passwd"},
		{"app;touch /tmp/pwn", "pwn"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewNamedParser("app", None)
			var out strings.Builder

			if err := p.WriteNamedCompletion(&out, CompletionShellBash, tc.name); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := "complete -F _" + strings.ReplaceAll(tc.want, "-", "_") + " '" + tc.want + "'"
			if !strings.Contains(out.String(), want) {
				t.Fatalf("expected only the trailing path component %q to survive, got:\n%s", tc.want, out.String())
			}
		})
	}
}

func TestWriteAutoCompletionDetectsZsh(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "zsh")

	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteAutoCompletion(&out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "#compdef app") {
		t.Fatalf("expected zsh completion output:\n%s", out.String())
	}
}

func TestWriteAutoCompletionDetectsPwsh(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "pwsh")

	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteAutoCompletion(&out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Register-ArgumentCompleter") {
		t.Fatalf("expected pwsh completion output:\n%s", out.String())
	}
}

func TestWriteAutoCompletionFallbacksToBash(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "fish")

	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteAutoCompletion(&out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "complete -F _app 'app'") {
		t.Fatalf("expected bash completion fallback output:\n%s", out.String())
	}
}
