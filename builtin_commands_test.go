// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltinCommandsDisabledByDefault(t *testing.T) {
	p := NewNamedParser("builtin-default", Default)

	if err := p.EnsureBuiltinCommands(); err != nil {
		t.Fatalf("unexpected ensure built-in commands error: %v", err)
	}

	for _, name := range []string{"help", "version", "completion", "docs", "config"} {
		if cmd := p.Find(name); cmd != nil {
			t.Fatalf("did not expect built-in command %q by default", name)
		}
	}
}

func TestBuiltinCommandsCanBeEnabledAfterInitialEnsure(t *testing.T) {
	p := NewNamedParser("builtin-late", None)

	if err := p.EnsureBuiltinCommands(); err != nil {
		t.Fatalf("unexpected initial ensure built-in commands error: %v", err)
	}

	p.Options |= HelpCommand
	if err := p.EnsureBuiltinCommands(); err != nil {
		t.Fatalf("unexpected second ensure built-in commands error: %v", err)
	}

	if cmd := p.Find("help"); cmd == nil {
		t.Fatalf("expected late-enabled help command")
	}
}

func TestBuiltinCommandsUseHelpCommandsGroup(t *testing.T) {
	p := NewNamedParser("builtin-group", HelpCommands)

	if err := p.EnsureBuiltinCommands(); err != nil {
		t.Fatalf("unexpected ensure built-in commands error: %v", err)
	}

	for _, name := range []string{"help", "version", "completion", "docs", "config"} {
		cmd := p.Find(name)
		if cmd == nil {
			t.Fatalf("expected built-in command %q", name)
		}
		if cmd.CommandGroup != "Help Commands" {
			t.Fatalf("unexpected group for %q: %q", name, cmd.CommandGroup)
		}
	}
}

func TestBuiltinCommandGroupIsLocalizedByDefault(t *testing.T) {
	p := NewNamedParser("builtin-group-i18n", HelpCommand)
	p.SetI18n(I18nConfig{Locale: "ru"})

	var out strings.Builder
	p.WriteHelp(&out)

	if !strings.Contains(out.String(), "Команды справки:") {
		t.Fatalf("expected localized built-in command group, got:\n%s", out.String())
	}
}

func TestBuiltinCommandHelpTextIsLocalized(t *testing.T) {
	p := NewNamedParser("builtin-command-i18n", Default|HelpCommands)
	p.SetI18n(I18nConfig{Locale: "ru"})

	var out strings.Builder
	p.WriteHelp(&out)

	for _, want := range []string{
		"Команды справки:",
		"Показать справку",
		"Показать информацию о версии",
		"Сгенерировать shell completion",
		"Сгенерировать документацию",
		"Сгенерировать пример INI-конфигурации",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected root help to contain %q, got:\n%s", want, out.String())
		}
	}

	p = NewNamedParser("builtin-command-i18n", Default|HelpCommands)
	p.SetI18n(I18nConfig{Locale: "ru"})
	stdout, stderr := captureStdIO(t, func() {
		_, _ = p.ParseArgs([]string{"docs", "md", "--help"})
	})
	if stderr != "" {
		t.Fatalf("expected empty stderr for localized built-in command help, got %q", stderr)
	}
	for _, want := range []string{
		"ШАБЛОН",
		"Шаблон Markdown-документации",
		"Включать скрытые опции, группы и команды",
		"Помечать скрытые сущности в документации",
		"Путь к выходному файлу",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected built-in command help to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestBuiltinCommandGroupCanBeCleared(t *testing.T) {
	p := NewNamedParser("builtin-group", HelpCommands)
	p.SetBuiltinCommandGroup("")

	if err := p.EnsureBuiltinCommands(); err != nil {
		t.Fatalf("unexpected ensure built-in commands error: %v", err)
	}

	if cmd := p.Find("help"); cmd == nil || cmd.CommandGroup != "" {
		t.Fatalf("expected empty built-in command group, got %#v", cmd)
	}
}

func TestBuiltinCommandNameConflict(t *testing.T) {
	var opts struct {
		Help struct{} `command:"help"`
	}

	p := NewNamedParser("builtin-conflict", HelpCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err := p.ParseArgs([]string{"help"})
	if err == nil {
		t.Fatalf("expected built-in command conflict")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestDisabledBuiltinCommandNameDoesNotConflict(t *testing.T) {
	var opts struct {
		Completion struct{} `command:"completion"`
	}

	p := NewNamedParser("builtin-disabled-conflict", HelpCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "completion.bash")

	if _, err := p.ParseArgs([]string{"completion", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestEnabledBuiltinCommandAliasConflict(t *testing.T) {
	var opts struct {
		UserHelp struct{} `command:"user-help" alias:"help"`
	}

	p := NewNamedParser("builtin-alias-conflict", HelpCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err := p.ParseArgs([]string{"help"})
	if err == nil {
		t.Fatalf("expected built-in command alias conflict")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestBuiltinCompletionCommandDoesNotReserveShortShellFlag(t *testing.T) {
	var opts struct {
		Short bool `short:"s" description:"Application short flag"`
	}

	p := NewNamedParser("builtin-short", CompletionCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "completion.bash")

	if _, err := p.ParseArgs([]string{"completion", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestEnabledBuiltinCommandLongFlagConflictIsValidated(t *testing.T) {
	var opts struct {
		Shell string `long:"shell" description:"Application shell option"`
	}

	p := NewNamedParser("builtin-long-conflict", CompletionCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err := p.ParseArgs([]string{"completion"})
	if err == nil {
		t.Fatalf("expected built-in command option conflict")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestBuiltinCompletionCommandWritesFile(t *testing.T) {
	p := NewNamedParser("builtin-completion", CompletionCommand)
	out := filepath.Join(t.TempDir(), "completion.bash")

	if _, err := p.ParseArgs([]string{"completion", "--shell", "bash", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(got), "builtin-completion") {
		t.Fatalf("expected completion script for parser name, got:\n%s", string(got))
	}
}

func TestBuiltinCompletionCommandAutoDetectsShell(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "zsh")

	p := NewNamedParser("builtin-completion-auto-zsh", CompletionCommand)
	out := filepath.Join(t.TempDir(), "completion.zsh")

	if _, err := p.ParseArgs([]string{"completion", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(got), "#compdef") {
		t.Fatalf("expected zsh completion script, got:\n%s", string(got))
	}
}

func TestBuiltinCompletionCommandAutoDetectsPwsh(t *testing.T) {
	t.Setenv("GO_FLAGS_SHELL", "pwsh")

	p := NewNamedParser("builtin-completion-auto-pwsh", CompletionCommand)
	out := filepath.Join(t.TempDir(), "completion.ps1")

	if _, err := p.ParseArgs([]string{"completion", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(got), "Register-ArgumentCompleter") {
		t.Fatalf("expected pwsh completion script, got:\n%s", string(got))
	}
}

func TestBuiltinCommandSkipsApplicationRequiredValidation(t *testing.T) {
	var opts struct {
		Value string `long:"value" required:"true"`
	}

	p := NewNamedParser("builtin-required", CompletionCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "completion.bash")

	if _, err := p.ParseArgs([]string{"completion", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestBuiltinHelpCommandShowsRootHelpByDefault(t *testing.T) {
	var opts struct {
		Deploy struct{} `command:"deploy" description:"Deploy selected targets"`
	}

	p := NewNamedParser("builtin-help-root", HelpCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	stdout, stderr := captureStdIO(t, func() {
		_, _ = p.ParseArgs([]string{"help"})
	})
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	for _, want := range []string{
		"Usage:",
		"builtin-help-root",
		"Available commands:",
		"deploy",
		"help",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected root help to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestBuiltinHelpCommandShowsCommandHelpForTarget(t *testing.T) {
	var opts struct {
		Deploy struct {
			Force bool `long:"force" description:"Force deploy"`
		} `command:"deploy" description:"Deploy selected targets"`
	}

	p := NewNamedParser("builtin-help-target", HelpCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	stdout, stderr := captureStdIO(t, func() {
		_, _ = p.ParseArgs([]string{"help", "deploy"})
	})
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	for _, want := range []string{
		"Usage:",
		"builtin-help-target",
		"deploy",
		"[deploy command options]",
		defaultLongOptDelimiter + "force",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected command help to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestBuiltinHelpCommandTakesPrecedenceOverRootPositionalArgs(t *testing.T) {
	var opts struct {
		Positional struct {
			Target string `positional-arg-name:"target" required:"true"`
		} `positional-args:"yes"`

		Value string `long:"value" required:"true"`
	}

	p := NewNamedParser("builtin-precedence", HelpCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	stdout, stderr := captureStdIO(t, func() {
		_, _ = p.ParseArgs([]string{"help"})
	})
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected help output, got:\n%s", stdout)
	}
}

func TestBuiltinDocsCommandWritesFile(t *testing.T) {
	p := NewNamedParser("builtin-docs", DocsCommand)
	out := filepath.Join(t.TempDir(), "docs.md")

	if _, err := p.ParseArgs([]string{"docs", "md", "--template", "table", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(got), "# builtin-docs") {
		t.Fatalf("expected markdown documentation, got:\n%s", string(got))
	}
}

func TestBuiltinDocsCommandProgramNameOverrideMarkdown(t *testing.T) {
	p := NewNamedParser("app.exe", DocsCommand)
	out := filepath.Join(t.TempDir(), "docs.md")

	if _, err := p.ParseArgs([]string{"docs", "md", "--program-name", "app", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	text := string(got)
	if !strings.Contains(text, "# app") {
		t.Fatalf("expected overridden program name in markdown docs, got:\n%s", text)
	}
	if strings.Contains(text, "app.exe") {
		t.Fatalf("did not expect original binary name in markdown docs, got:\n%s", text)
	}
}

func TestBuiltinDocsCommandProgramNameOverrideMan(t *testing.T) {
	p := NewNamedParser("app.exe", DocsCommand)
	out := filepath.Join(t.TempDir(), "docs.1")

	if _, err := p.ParseArgs([]string{"docs", "man", "--program-name", "app", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	text := string(got)
	if !strings.Contains(text, ".TH app 1 ") {
		t.Fatalf("expected overridden program name in man docs, got:\n%s", text)
	}
	if strings.Contains(text, "app.exe") {
		t.Fatalf("did not expect original binary name in man docs, got:\n%s", text)
	}
}

func TestBuiltinDocsCommandTOCMarkdown(t *testing.T) {
	p := NewNamedParser("app.exe", HelpCommands)
	out := filepath.Join(t.TempDir(), "docs.md")

	if _, err := p.ParseArgs([]string{"docs", "md", "--toc", "--program-name", "app", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	text := string(got)
	for _, want := range []string{
		"## Table of Contents",
		"[COMMANDS](#commands)",
		"[help](#help)",
		"[version](#version)",
		"[completion](#completion)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected markdown toc entry %q, got:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{"(#name)", "(#synopsis)"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("did not expect %q in toc, got:\n%s", unwanted, text)
		}
	}
}

func TestBuiltinDocsCommandTOCHTML(t *testing.T) {
	p := NewNamedParser("app.exe", HelpCommands)
	out := filepath.Join(t.TempDir(), "docs.html")

	if _, err := p.ParseArgs([]string{"docs", "html", "--toc", "--program-name", "app", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	text := string(got)
	for _, want := range []string{
		"<h2>Table of Contents</h2>",
		"href=\"#commands\">COMMANDS</a>",
		"href=\"#command-help\">help</a>",
		"href=\"#command-version\">version</a>",
		"href=\"#command-completion\">completion</a>",
		"id=\"command-docs-html\">docs html</h3>",
		"template TEMPLATE",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected html toc entry %q, got:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{"href=\"#name\"", "href=\"#synopsis\""} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("did not expect %q in toc, got:\n%s", unwanted, text)
		}
	}
}

func TestBuiltinConfigCommandWritesFile(t *testing.T) {
	var opts struct {
		Value string `long:"value" description:"Config value"`
	}

	p := NewNamedParser("builtin-config", ConfigCommand)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "config.ini")

	if _, err := p.ParseArgs([]string{"config", "--comment-width", "40", out}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(got), "value") {
		t.Fatalf("expected config example, got:\n%s", string(got))
	}
}
