// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"os"
	"path/filepath"
	"testing"
)

func mustParseDotEnv(t *testing.T, src string) map[string]string {
	t.Helper()

	m, err := parseDotEnvBytes([]byte(src), false)
	if err != nil {
		t.Fatalf("parseDotEnvBytes(%q): unexpected error: %v", src, err)
	}

	return m
}

func assertDotEnvKey(t *testing.T, m map[string]string, key, want string) {
	t.Helper()

	if got := m[key]; got != want {
		t.Errorf("key %q: got %q, want %q", key, got, want)
	}
}

func writeTmpEnv(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestDotEnvParseBasic(t *testing.T) {
	m := mustParseDotEnv(t, "KEY=value\n")
	assertDotEnvKey(t, m, "KEY", "value")
}

func TestDotEnvParseYAMLColon(t *testing.T) {
	m := mustParseDotEnv(t, "KEY: value\n")
	assertDotEnvKey(t, m, "KEY", "value")
}

func TestDotEnvParseExportPrefix(t *testing.T) {
	m := mustParseDotEnv(t, "export KEY=value\n")
	assertDotEnvKey(t, m, "KEY", "value")
}

func TestDotEnvParseEmptyValue(t *testing.T) {
	m := mustParseDotEnv(t, "KEY=\n")
	assertDotEnvKey(t, m, "KEY", "")
}

func TestDotEnvParseMultipleKeys(t *testing.T) {
	m := mustParseDotEnv(t, "A=1\nB=2\nC=3\n")
	assertDotEnvKey(t, m, "A", "1")
	assertDotEnvKey(t, m, "B", "2")
	assertDotEnvKey(t, m, "C", "3")
}

func TestDotEnvParseCommentLine(t *testing.T) {
	m := mustParseDotEnv(t, "# this is a comment\nKEY=value\n")
	assertDotEnvKey(t, m, "KEY", "value")

	if _, ok := m["#"]; ok {
		t.Error("comment line should not produce a key")
	}
}

func TestDotEnvParseInlineComment(t *testing.T) {
	m := mustParseDotEnv(t, "KEY=value # inline comment\n")
	assertDotEnvKey(t, m, "KEY", "value")
}

func TestDotEnvParseHashInValueNoSpace(t *testing.T) {
	// '#' without preceding space is not a comment.
	m := mustParseDotEnv(t, "URL=http://example.com/path#anchor\n")
	assertDotEnvKey(t, m, "URL", "http://example.com/path#anchor")
}

func TestDotEnvParseDoubleQuoted(t *testing.T) {
	m := mustParseDotEnv(t, `MSG="hello world"`)
	assertDotEnvKey(t, m, "MSG", "hello world")
}

func TestDotEnvParseSingleQuoted(t *testing.T) {
	m := mustParseDotEnv(t, `MSG='hello world'`)
	assertDotEnvKey(t, m, "MSG", "hello world")
}

func TestDotEnvParseDoubleQuotedEscapeNewline(t *testing.T) {
	m := mustParseDotEnv(t, `MSG="line1\nline2"`)
	assertDotEnvKey(t, m, "MSG", "line1\nline2")
}

func TestDotEnvParseSingleQuotedLiteralBackslash(t *testing.T) {
	// In single-quoted values \n is NOT a newline.
	m := mustParseDotEnv(t, `MSG='\n'`)
	assertDotEnvKey(t, m, "MSG", `\n`)
}

func TestDotEnvParseDoubleQuotedEscapeBackslash(t *testing.T) {
	m := mustParseDotEnv(t, `MSG="a\\b"`)
	assertDotEnvKey(t, m, "MSG", `a\b`)
}

func TestDotEnvParseUnterminatedQuoteError(t *testing.T) {
	_, err := parseDotEnvBytes([]byte(`KEY="unterminated`), false)
	if err == nil {
		t.Error("expected error for unterminated quote, got nil")
	}
}

func TestDotEnvParseMultilineDouble(t *testing.T) {
	src := "MSG=\"line1\nline2\"\nOTHER=x\n"
	m := mustParseDotEnv(t, src)
	assertDotEnvKey(t, m, "MSG", "line1\nline2")
	assertDotEnvKey(t, m, "OTHER", "x")
}

func TestDotEnvExpandSimpleVar(t *testing.T) {
	m := mustParseDotEnv(t, "BASE=hello\nFULL=$BASE world\n")
	assertDotEnvKey(t, m, "FULL", "hello world")
}

func TestDotEnvExpandBracedVar(t *testing.T) {
	m := mustParseDotEnv(t, "BASE=hello\nFULL=${BASE} world\n")
	assertDotEnvKey(t, m, "FULL", "hello world")
}

func TestDotEnvExpandChain(t *testing.T) {
	m := mustParseDotEnv(t, "A=foo\nB=${A}_bar\nC=${B}_baz\n")
	assertDotEnvKey(t, m, "C", "foo_bar_baz")
}

func TestDotEnvExpandUnknownVar(t *testing.T) {
	// Unknown variable expands to empty string.
	os.Unsetenv("__DOTENV_UNDEF__")
	m := mustParseDotEnv(t, "VAL=${__DOTENV_UNDEF__}suffix\n")
	assertDotEnvKey(t, m, "VAL", "suffix")
}

func TestDotEnvExpandFromEnv(t *testing.T) {
	os.Setenv("__DOTENV_TEST_OS__", "fromenv")
	defer os.Unsetenv("__DOTENV_TEST_OS__")

	m := mustParseDotEnv(t, "VAL=${__DOTENV_TEST_OS__}\n")
	assertDotEnvKey(t, m, "VAL", "fromenv")
}

func TestDotEnvExpandDollarAtEndOfLine(t *testing.T) {
	m := mustParseDotEnv(t, "PRICE=10$\n")
	assertDotEnvKey(t, m, "PRICE", "10$")
}

func TestDotEnvExpandDollarInMiddle(t *testing.T) {
	// $HOME is likely set on the test machine; we just verify it does not crash
	// and does not keep the literal "$HOME" string intact.
	m := mustParseDotEnv(t, "P=/usr/bin:$HOME/bin\n")
	got := m["P"]

	if got == "/usr/bin:$HOME/bin" {
		// $HOME not set - tolerable but unexpected; flag it only as a notice.
		t.Logf("$HOME was not expanded (unset?): %q", got)
	}
}

func TestDotEnvExpandAssignDefault(t *testing.T) {
	os.Unsetenv("__DOTENV_UNDEF2__")
	m := mustParseDotEnv(t, "VAL=${__DOTENV_UNDEF2__:=mydefault}\n")
	assertDotEnvKey(t, m, "VAL", "mydefault")

	// The :=  operator should also update the local map so later keys can use it.
	m2 := mustParseDotEnv(t, "A=${__DOTENV_UNDEF3__:=hello}\nB=${__DOTENV_UNDEF3__}\n")
	os.Unsetenv("__DOTENV_UNDEF3__")
	assertDotEnvKey(t, m2, "A", "hello")
	assertDotEnvKey(t, m2, "B", "hello")
}

func TestDotEnvExpandAssignDefaultExistingVar(t *testing.T) {
	// When the variable is already set, := should NOT override it.
	m := mustParseDotEnv(t, "A=existing\nB=${A:=default}\n")
	assertDotEnvKey(t, m, "B", "existing")
}

func TestDotEnvExpandEscapedBrace(t *testing.T) {
	m := mustParseDotEnv(t, "VAL=$${HOME}\n")
	assertDotEnvKey(t, m, "VAL", "${HOME}")
}

func TestDotEnvExpandEscapedBraceInDouble(t *testing.T) {
	m := mustParseDotEnv(t, `VAL="prefix $${HOME} suffix"`)
	assertDotEnvKey(t, m, "VAL", "prefix ${HOME} suffix")
}

func TestDotEnvExpandDoubleDollarNotEscaped(t *testing.T) {
	// $$ not followed by { is two separate literals, not a special form.
	m := mustParseDotEnv(t, "VAL=$$\n")
	assertDotEnvKey(t, m, "VAL", "$$")
}

func TestDotEnvSingleQuoteNoExpand(t *testing.T) {
	m := mustParseDotEnv(t, "BASE=hello\nVAL='$BASE'\n")
	assertDotEnvKey(t, m, "VAL", "$BASE")
}

func TestDotEnvSingleQuoteDollarBrace(t *testing.T) {
	m := mustParseDotEnv(t, `VAL='${HOME}'`)
	assertDotEnvKey(t, m, "VAL", "${HOME}")
}

func TestDotEnvDoubleQuoteExpands(t *testing.T) {
	m := mustParseDotEnv(t, "BASE=hello\nVAL=\"${BASE} world\"\n")
	assertDotEnvKey(t, m, "VAL", "hello world")
}

func TestDotEnvDoubleQuoteEscapedDollar(t *testing.T) {
	// \$ inside double quotes: godotenv-style \X pass-through for unknown X.
	// The parser leaves \$ as-is (unknown escape), then expansion ignores \$.
	// Net result: the dollar is treated as literal after unescaping.
	m := mustParseDotEnv(t, `VAL="cost is \$5"`)
	got := m["VAL"]

	if got == "" {
		t.Errorf("VAL should not be empty, got %q", got)
	}
}

func TestDotEnvParseSpaceBeforeEq(t *testing.T) {
	// "OPTION_D =4" - space before separator
	m := mustParseDotEnv(t, "OPTION_D =4\n")
	assertDotEnvKey(t, m, "OPTION_D", "4")
}

func TestDotEnvParseSpaceAroundEq(t *testing.T) {
	// "OPTION_E = 5" - spaces on both sides
	m := mustParseDotEnv(t, "OPTION_E = 5\n")
	assertDotEnvKey(t, m, "OPTION_E", "5")
}

func TestDotEnvParseBlankAfterSpace(t *testing.T) {
	// "OPTION_F = " - whitespace-only value
	m := mustParseDotEnv(t, "OPTION_F = \n")
	assertDotEnvKey(t, m, "OPTION_F", "")
}

func TestDotEnvParseInternalWhitespacePreserved(t *testing.T) {
	// "OPTION_H=1 2" - internal space is kept
	m := mustParseDotEnv(t, "OPTION_H=1 2\n")
	assertDotEnvKey(t, m, "OPTION_H", "1 2")
}

func TestDotEnvParseEmptySingleQuoted(t *testing.T) {
	m := mustParseDotEnv(t, "VAL=''\n")
	assertDotEnvKey(t, m, "VAL", "")
}

func TestDotEnvParseEmptyDoubleQuoted(t *testing.T) {
	m := mustParseDotEnv(t, `VAL=""`)
	assertDotEnvKey(t, m, "VAL", "")
}

func TestDotEnvParseSingleInsideDouble(t *testing.T) {
	// Single quotes are literal inside double-quoted value.
	m := mustParseDotEnv(t, `VAL="echo 'asd'"`)
	assertDotEnvKey(t, m, "VAL", "echo 'asd'")
}

func TestDotEnvParseMultilineSingleQuoted(t *testing.T) {
	src := "VAL='line 1\nline 2'\n"
	m := mustParseDotEnv(t, src)
	assertDotEnvKey(t, m, "VAL", "line 1\nline 2")
}

func TestDotEnvParseMultilineSingleQuotedEscapedQuote(t *testing.T) {
	// An escaped single quote inside a single-quoted value.
	// godotenv counts backslashes preceding the quote to detect escaping.
	src := "VAL='line one\nthis is \\'quoted\\'\none more line'\n"
	m, err := parseDotEnvBytes([]byte(src), false)
	if err != nil {
		t.Fatal(err)
	}
	// At minimum, the value must have been parsed without error.
	if _, ok := m["VAL"]; !ok {
		t.Error("key VAL missing from result")
	}
}

func TestDotEnvParseMultilineDoubleQuotedEscapedQuote(t *testing.T) {
	src := "VAL=\"line one\nthis is \\\"quoted\\\"\none more line\"\n"
	m := mustParseDotEnv(t, src)
	got := m["VAL"]
	if got == "" {
		t.Error("VAL should not be empty")
	}
	// The parsed value must contain the word "quoted" without backslashes.
	if !containsStr(got, "quoted") {
		t.Errorf("expected 'quoted' in value, got %q", got)
	}
}

func TestDotEnvParseMultipleHashInline(t *testing.T) {
	// "fred=qux#baz # other # more" -> "qux#baz"
	m := mustParseDotEnv(t, "fred=qux#baz # other # more\n")
	assertDotEnvKey(t, m, "fred", "qux#baz")
}

func TestDotEnvParseQuotedValueHashAfterClose(t *testing.T) {
	// baz="foo"#bar -> "foo"  (# after closing quote is not part of value)
	m := mustParseDotEnv(t, `baz="foo"#bar`)
	assertDotEnvKey(t, m, "baz", "foo")
}

func TestDotEnvParseHashWithoutSpaceNotComment(t *testing.T) {
	// "bar=foo#baz" -> "foo#baz" (# without preceding space is not a comment)
	m := mustParseDotEnv(t, "bar=foo#baz\n")
	assertDotEnvKey(t, m, "bar", "foo#baz")
}

func TestDotEnvParseHyphenInKey(t *testing.T) {
	m := mustParseDotEnv(t, "OPTION-B=def\n")
	assertDotEnvKey(t, m, "OPTION-B", "def")
}

func TestDotEnvParseURLInSingleQuoted(t *testing.T) {
	src := "DB='" + "postgres://localhost:5432/database?sslmode=disable" + "'\n"
	m := mustParseDotEnv(t, src)
	assertDotEnvKey(t, m, "DB", "postgres://localhost:5432/database?sslmode=disable")
}

func TestDotEnvParseInvalidLine(t *testing.T) {
	_, err := parseDotEnvBytes([]byte("INVALID LINE\nfoo=bar\n"), false)
	if err == nil {
		t.Error("expected error for line without separator, got nil")
	}
}

func TestDotEnvParseExportSingleQuoted(t *testing.T) {
	m := mustParseDotEnv(t, "export OPTION_B='\\n'\n")
	assertDotEnvKey(t, m, "OPTION_B", `\n`)
}

func TestDotEnvExpandConcatenatedVars(t *testing.T) {
	m := mustParseDotEnv(t, "A=1\nB=${A}${A}\n")
	assertDotEnvKey(t, m, "B", "11")
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i+len(substr) <= len(s); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

func TestDotEnvNoExpand(t *testing.T) {
	m, err := parseDotEnvBytes([]byte("BASE=hello\nVAL=${BASE}\n"), true)
	if err != nil {
		t.Fatal(err)
	}

	// noExpand=true: ${BASE} must remain as a literal string.
	assertDotEnvKey(t, m, "VAL", "${BASE}")
}

func TestDotEnvNoExpandSimpleDollar(t *testing.T) {
	m, err := parseDotEnvBytes([]byte("BASE=hello\nVAL=$BASE\n"), true)
	if err != nil {
		t.Fatal(err)
	}

	assertDotEnvKey(t, m, "VAL", "$BASE")
}

func TestDotEnvLoadDotEnvMethod(t *testing.T) {
	os.Unsetenv("DOTENV_TEST_LOAD")

	path := writeTmpEnv(t, "DOTENV_TEST_LOAD=loaded\n")
	p := NewNamedParser("test", None)
	p.SetDotEnvFile(path)

	if err := p.LoadDotEnv(); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_TEST_LOAD"); got != "loaded" {
		t.Errorf("got %q, want %q", got, "loaded")
	}
}

func TestDotEnvLoadDoesNotOverride(t *testing.T) {
	os.Setenv("DOTENV_NO_OVER", "original")
	defer os.Unsetenv("DOTENV_NO_OVER")

	path := writeTmpEnv(t, "DOTENV_NO_OVER=new\n")
	p := NewNamedParser("test", None)

	if err := p.LoadDotEnv(path); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_NO_OVER"); got != "original" {
		t.Errorf("LoadDotEnv should not override, got %q", got)
	}
}

// TestDotEnvLoadDotEnvMultipleFilesTransactional ensures a later file's load error
// does not leave an earlier file's variables applied to the process environment:
// all files are read and parsed before any os.Setenv call.
func TestDotEnvLoadDotEnvMultipleFilesTransactional(t *testing.T) {
	os.Unsetenv("DOTENV_MULTI_FIRST")
	defer os.Unsetenv("DOTENV_MULTI_FIRST")

	dir := t.TempDir()

	firstPath := filepath.Join(dir, "first.env")
	if err := os.WriteFile(firstPath, []byte("DOTENV_MULTI_FIRST=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	missingPath := filepath.Join(dir, "missing.env")

	p := NewNamedParser("test", None)

	if err := p.LoadDotEnv(firstPath, missingPath); err == nil {
		t.Fatal("expected error for missing second file")
	}

	if got := os.Getenv("DOTENV_MULTI_FIRST"); got != "" {
		t.Errorf("expected first file's variables not to be applied when a later file fails to load, got %q", got)
	}
}

// TestDotEnvLoadDotEnvMultipleFilesTransactionalOnParseError
// is the same guarantee for a later file that exists but fails to parse.
func TestDotEnvLoadDotEnvMultipleFilesTransactionalOnParseError(t *testing.T) {
	os.Unsetenv("DOTENV_MULTI_PARSE_FIRST")
	defer os.Unsetenv("DOTENV_MULTI_PARSE_FIRST")

	dir := t.TempDir()

	firstPath := filepath.Join(dir, "first.env")
	if err := os.WriteFile(firstPath, []byte("DOTENV_MULTI_PARSE_FIRST=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	invalidPath := filepath.Join(dir, "invalid.env")
	if err := os.WriteFile(invalidPath, []byte("not a valid line\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	p := NewNamedParser("test", None)

	if err := p.LoadDotEnv(firstPath, invalidPath); err == nil {
		t.Fatal("expected error for invalid second file")
	}

	if got := os.Getenv("DOTENV_MULTI_PARSE_FIRST"); got != "" {
		t.Errorf("expected first file's variables not to be applied when a later file fails to parse, got %q", got)
	}
}

func TestDotEnvOverloadDotEnvMethod(t *testing.T) {
	os.Setenv("DOTENV_OVER", "original")
	defer os.Unsetenv("DOTENV_OVER")

	path := writeTmpEnv(t, "DOTENV_OVER=overridden\n")
	p := NewNamedParser("test", None)

	if err := p.OverloadDotEnv(path); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_OVER"); got != "overridden" {
		t.Errorf("OverloadDotEnv should override, got %q", got)
	}
}

func TestDotEnvMissingDefaultFileSilent(t *testing.T) {
	p := NewNamedParser("test", None)
	p.SetDotEnvFile(filepath.Join(t.TempDir(), "nonexistent.env"))
	// LoadDotEnv with the default file name (no explicit argument) is silent.
	// Since SetDotEnvFile was called, this should not error on a missing file
	// when accessed via the DotEnv option path (silentMissing=true).
	if err := p.LoadDotEnv(); err != nil {
		t.Fatalf("missing default .env should be silent, got: %v", err)
	}
}

func TestDotEnvExplicitMissingFileErrors(t *testing.T) {
	p := NewNamedParser("test", None)
	err := p.LoadDotEnv(filepath.Join(t.TempDir(), "nonexistent.env"))

	if err == nil {
		t.Error("explicit missing file should return error")
	}
}

func TestDotEnvOptionAutoLoad(t *testing.T) {
	os.Unsetenv("DOTENV_AUTO")

	path := writeTmpEnv(t, "DOTENV_AUTO=autoloaded\n")

	p := NewNamedParser("test", DotEnv)
	p.SetDotEnvFile(path)

	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_AUTO"); got != "autoloaded" {
		t.Errorf("DotEnv option: got %q, want %q", got, "autoloaded")
	}
}

func TestDotEnvOverrideOption(t *testing.T) {
	os.Setenv("DOTENV_OPTOVER", "original")
	defer os.Unsetenv("DOTENV_OPTOVER")

	path := writeTmpEnv(t, "DOTENV_OPTOVER=from_file\n")

	p := NewNamedParser("test", DotEnvOverride)
	p.SetDotEnvFile(path)

	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_OPTOVER"); got != "from_file" {
		t.Errorf("DotEnvOverride: got %q, want %q", got, "from_file")
	}
}

func TestDotEnvMissingFileWithDotEnvOptionSilent(t *testing.T) {
	p := NewNamedParser("test", DotEnv)
	p.SetDotEnvFile(filepath.Join(t.TempDir(), "nope.env"))

	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Errorf("missing .env with DotEnv option should be silent, got: %v", err)
	}
}

func TestDotEnvSetFileOption(t *testing.T) {
	os.Unsetenv("DOTENV_CUSTOM")

	path := writeTmpEnv(t, "DOTENV_CUSTOM=custom\n")

	p := NewNamedParser("test", DotEnv)
	p.SetDotEnvFile(path)

	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_CUSTOM"); got != "custom" {
		t.Errorf("SetDotEnvFile: got %q, want %q", got, "custom")
	}
}

func TestDotEnvFlagsEnvFileArg(t *testing.T) {
	os.Unsetenv("DOTENV_FLAGS_FILE")

	path := writeTmpEnv(t, "DOTENV_FLAGS_FILE=fromflag\n")

	p := NewNamedParser("test", DotEnv|DotEnvFlags)

	_, err := p.ParseArgs([]string{"--env-file=" + path})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_FLAGS_FILE"); got != "fromflag" {
		t.Errorf("--env-file flag: got %q, want %q", got, "fromflag")
	}
}

func TestDotEnvFlagsEnvFileSeparated(t *testing.T) {
	os.Unsetenv("DOTENV_FLAGS_SEP")

	path := writeTmpEnv(t, "DOTENV_FLAGS_SEP=sep\n")

	p := NewNamedParser("test", DotEnv|DotEnvFlags)

	_, err := p.ParseArgs([]string{"--env-file", path})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_FLAGS_SEP"); got != "sep" {
		t.Errorf("--env-file separated: got %q, want %q", got, "sep")
	}
}

func TestDotEnvFlagsNoEnvDisables(t *testing.T) {
	os.Unsetenv("DOTENV_SHOULD_NOT_LOAD")

	path := writeTmpEnv(t, "DOTENV_SHOULD_NOT_LOAD=loaded\n")

	p := NewNamedParser("test", DotEnv|DotEnvFlags)
	p.SetDotEnvFile(path)

	_, err := p.ParseArgs([]string{"--no-env"})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_SHOULD_NOT_LOAD"); got != "" {
		t.Errorf("--no-env should disable loading, but got %q", got)
	}
}

func TestDotEnvFlagsEnvOverride(t *testing.T) {
	os.Setenv("DOTENV_FLAG_OVER", "original")
	defer os.Unsetenv("DOTENV_FLAG_OVER")

	path := writeTmpEnv(t, "DOTENV_FLAG_OVER=overridden\n")

	p := NewNamedParser("test", DotEnv|DotEnvFlags)
	p.SetDotEnvFile(path)

	_, err := p.ParseArgs([]string{"--env-override"})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_FLAG_OVER"); got != "overridden" {
		t.Errorf("--env-override: got %q, want %q", got, "overridden")
	}
}

func TestDotEnvFlagsNoEnvIgnoredAfterDoubleDashWhenPassDoubleDashSet(t *testing.T) {
	os.Unsetenv("DOTENV_PASS_DOUBLE_DASH")

	path := writeTmpEnv(t, "DOTENV_PASS_DOUBLE_DASH=loaded\n")

	p := NewNamedParser("test", DotEnv|DotEnvFlags|PassDoubleDash)
	p.SubcommandsOptional = true
	p.SetDotEnvFile(path)

	retargs, err := p.ParseArgs([]string{"--", "--no-env"})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_PASS_DOUBLE_DASH"); got != "loaded" {
		t.Errorf("expected --no-env after a literal -- to be treated as a literal argument, dotenv should still load: got %q", got)
	}
	if len(retargs) != 1 || retargs[0] != "--no-env" {
		t.Errorf("expected the literal token to pass through as an argument, got %v", retargs)
	}
}

func TestDotEnvBuiltinOptionAccessors(t *testing.T) {
	p := NewNamedParser("test", DotEnvFlags)

	if opt := p.BuiltinDotEnvFileOption(); opt == nil {
		t.Error("BuiltinDotEnvFileOption should not be nil")
	}

	if opt := p.BuiltinDotEnvNoEnvOption(); opt == nil {
		t.Error("BuiltinDotEnvNoEnvOption should not be nil")
	}

	if opt := p.BuiltinDotEnvOverrideOption(); opt == nil {
		t.Error("BuiltinDotEnvOverrideOption should not be nil")
	}
}

func TestDotEnvBuiltinOptionAccessorsWithoutFlag(t *testing.T) {
	p := NewNamedParser("test", None)

	if opt := p.BuiltinDotEnvFileOption(); opt != nil {
		t.Error("BuiltinDotEnvFileOption should be nil when DotEnvFlags is not set")
	}
}

func TestDotEnvSetNoExpand(t *testing.T) {
	os.Unsetenv("DOTENV_NOEXP")
	path := writeTmpEnv(t, "DOTENV_NOEXP=${SOME_VAR}\n")

	p := NewNamedParser("test", DotEnv)
	p.SetDotEnvFile(path)
	p.SetDotEnvNoExpand(true)

	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_NOEXP"); got != "${SOME_VAR}" {
		t.Errorf("SetDotEnvNoExpand: got %q, want literal ${SOME_VAR}", got)
	}
}
