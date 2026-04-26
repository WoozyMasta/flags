package flags

import (
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
	if !strings.Contains(got, "complete -F _my_app my-app") {
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

	if !strings.Contains(got, "compdef _my_app my-app") {
		t.Fatalf("missing zsh compdef command:\n%s", got)
	}
	if !strings.Contains(got, "compadd -S '' --") {
		t.Fatalf("missing zsh nospace completion branch:\n%s", got)
	}
}

func TestWriteCompletionUsesParserName(t *testing.T) {
	p := NewNamedParser("tool-name", None)
	var out strings.Builder

	if err := p.WriteCompletion(&out, CompletionShellBash); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "complete -F _tool_name tool-name") {
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

func TestWriteAutoCompletionFallbacksToBash(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "pwsh")

	p := NewNamedParser("app", None)
	var out strings.Builder

	if err := p.WriteAutoCompletion(&out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "complete -F _app app") {
		t.Fatalf("expected bash completion output:\n%s", out.String())
	}
}
