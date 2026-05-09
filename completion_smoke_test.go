// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildCompletionSmokeParser returns a parser with commands, options, and
// choice-constrained flags suitable for smoke-testing completion output.
func buildCompletionSmokeParser(t *testing.T) *Parser {
	t.Helper()

	var opts struct {
		Verbose bool     `short:"v" long:"verbose" description:"Verbose output"`
		Output  string   `short:"o" long:"output" description:"Output file" default:"-"`
		Format  string   `long:"format" choices:"json;yaml;text" default:"json" description:"Output format"`
		Tags    []string `long:"tag" description:"Filter tags"`

		Run struct {
			Force bool `long:"force" description:"Force execution"`
			Count int  `short:"n" long:"count" description:"Repeat count" default:"1"`
		} `command:"run" description:"Execute the task"`

		Status struct {
			Watch bool `short:"w" long:"watch" description:"Watch for changes"`
			JSON  bool `long:"json" description:"JSON output"`
		} `command:"status" description:"Show current status"`

		Init struct {
			Dir string `short:"d" long:"dir" description:"Target directory" default:"."`
		} `command:"init" description:"Initialize workspace"`
	}

	p := NewNamedParser("smoke-app", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("add group: %v", err)
	}

	return p
}

// assertCompletionNonEmpty generates a completion script for shell and
// commandName and verifies the output is non-empty and contains every token.
func assertCompletionNonEmpty(t *testing.T, p *Parser, shell CompletionShell, commandName string, tokens ...string) {
	t.Helper()

	var buf bytes.Buffer
	if err := p.WriteNamedCompletion(&buf, shell, commandName); err != nil {
		t.Fatalf("WriteNamedCompletion(%s): %v", shell, err)
	}

	got := buf.String()
	if strings.TrimSpace(got) == "" {
		t.Fatalf("completion script for shell %q is empty", shell)
	}

	for _, tok := range tokens {
		if !strings.Contains(got, tok) {
			t.Fatalf("completion script for shell %q missing expected token %q:\n%s", shell, tok, got)
		}
	}
}

// assertCompletionSnapshot generates a completion script for shell and
// commandName and compares it to the file at path. When
// UPDATE_COMPLETION_SNAPSHOTS=1 is set the snapshot is written instead of
// compared.
func assertCompletionSnapshot(t *testing.T, p *Parser, shell CompletionShell, commandName, path string) {
	t.Helper()

	var buf bytes.Buffer
	if err := p.WriteNamedCompletion(&buf, shell, commandName); err != nil {
		t.Fatalf("WriteNamedCompletion(%s): %v", shell, err)
	}

	got := normalizeNewlines(buf.String())

	if os.Getenv("UPDATE_COMPLETION_SNAPSHOTS") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create snapshot directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("failed to write snapshot %s: %v", path, err)
		}
		return
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read snapshot %s: %v (set UPDATE_COMPLETION_SNAPSHOTS=1 to generate)", path, err)
	}

	assertDiff(t, got, normalizeNewlines(string(raw)), "completion snapshot "+string(shell))
}

// assertCompletionCandidates runs the parser's internal completion engine
// against args and verifies each expected item appears in the candidates.
func assertCompletionCandidates(t *testing.T, p *Parser, args []string, expected ...string) {
	t.Helper()

	c := &completion{parser: p}
	results := c.complete(args)

	items := make([]string, len(results))
	for i, r := range results {
		items[i] = r.Item
	}
	got := strings.Join(items, "\n")

	for _, want := range expected {
		found := false
		for _, item := range items {
			if item == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("completion for args %v: expected candidate %q, got:\n%s", args, want, got)
		}
	}
}

// Script structure smoke tests.

func TestCompletionSmokeScriptBash(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionNonEmpty(t, p, CompletionShellBash, "smoke-app",
		"GO_FLAGS_COMPLETION",
		"complete -F _smoke_app smoke-app",
		"compopt -o nospace",
		"COMP_WORDS",
		"COMPREPLY",
	)
}

func TestCompletionSmokeScriptZsh(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionNonEmpty(t, p, CompletionShellZsh, "smoke-app",
		"GO_FLAGS_COMPLETION",
		"#compdef smoke-app",
		"compdef _smoke_app smoke-app",
		"compadd -S '' --",
	)
}

func TestCompletionSmokeScriptPwsh(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionNonEmpty(t, p, CompletionShellPwsh, "smoke-app",
		"GO_FLAGS_COMPLETION",
		"Register-ArgumentCompleter",
		"smoke-app",
		"$commandAst",
	)
}

// Snapshot regression tests.

func TestCompletionSmokeSnapshotBash(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionSnapshot(t, p, CompletionShellBash, "smoke-app",
		filepath.Join("examples", "completion", "smoke-app.bash"))
}

func TestCompletionSmokeSnapshotZsh(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionSnapshot(t, p, CompletionShellZsh, "smoke-app",
		filepath.Join("examples", "completion", "smoke-app.zsh"))
}

func TestCompletionSmokeSnapshotPwsh(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionSnapshot(t, p, CompletionShellPwsh, "smoke-app",
		filepath.Join("examples", "completion", "smoke-app.ps1"))
}

// Runtime candidate smoke tests.

func TestCompletionSmokeCandidatesCommands(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionCandidates(t, p, []string{""},
		"run", "status", "init",
	)
}

func TestCompletionSmokeCandidatesOptions(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	// Single '-' prefix: options with a long name are returned as --long.
	// Short-only options would appear as -x. Options with both names show
	// only the long form because the short is marked as a repeat.
	assertCompletionCandidates(t, p, []string{"-"},
		"--verbose", "--output", "--format", "--tag",
	)
}

func TestCompletionSmokeCandidatesLongOptions(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionCandidates(t, p, []string{"--"},
		"--verbose", "--output", "--format", "--tag",
	)
}

func TestCompletionSmokeCandidatesChoices(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	// Choices are returned with the full option=value prefix.
	assertCompletionCandidates(t, p, []string{"--format="},
		"--format=json", "--format=yaml", "--format=text",
	)
}

func TestCompletionSmokeCandidatesSubcommandOptions(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionCandidates(t, p, []string{"run", "--"},
		"--force", "--count",
	)
}

func TestCompletionSmokeCandidatesPartialCommand(t *testing.T) {
	p := buildCompletionSmokeParser(t)
	assertCompletionCandidates(t, p, []string{"st"},
		"status",
	)
}
