// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import "testing"

// TestAddHelpGroupsRetriesAndSurfacesErrorOnFailure ensures a failed
// built-in help/version group AddGroup call does not permanently and
// silently leave help/version flags absent: the guard stays false so a
// later attempt can retry, and the failure is recorded so it surfaces on
// the next ParseArgs call instead of being discarded.
func TestAddHelpGroupsRetriesAndSurfacesErrorOnFailure(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewParser(&opts, HelpFlag)
	p.MaxLongNameLength = 3 // "help"/"version" exceed this, so AddGroup must fail

	p.EnsureBuiltinOptions()

	if p.hasBuiltinHelpGroup {
		t.Fatal("expected hasBuiltinHelpGroup to remain false after a failed AddGroup")
	}
	if p.internalError == nil {
		t.Fatal("expected the AddGroup failure to be recorded as an internal error")
	}

	if _, err := p.ParseArgs([]string{"--value=x"}); err == nil {
		t.Fatal("expected ParseArgs to surface the recorded internal error")
	}

	// Retry: once the underlying problem is fixed, a fresh attempt succeeds
	// and the guard is finally set.
	p2 := NewParser(&opts, HelpFlag)
	p2.MaxLongNameLength = 3
	p2.EnsureBuiltinOptions()
	if p2.hasBuiltinHelpGroup {
		t.Fatal("expected first attempt to still fail")
	}

	p2.MaxLongNameLength = DefaultMaxLongNameLength
	p2.EnsureBuiltinOptions()
	if !p2.hasBuiltinHelpGroup {
		t.Fatal("expected retry to succeed once MaxLongNameLength no longer blocks it")
	}
}

// TestEnsureBuiltinDotEnvOptionsReturnsAddGroupError mirrors the same
// guard-consistency and error-visibility guarantee for the dotenv builtin
// group, following EnsureBuiltinConfigOptions' existing error-returning
// pattern.
func TestEnsureBuiltinDotEnvOptionsReturnsAddGroupError(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewParser(&opts, DotEnvFlags)
	p.MaxLongNameLength = 3 // "env-file"/"no-env"/"env-override" exceed this

	if err := p.EnsureBuiltinDotEnvOptions(); err == nil {
		t.Fatal("expected EnsureBuiltinDotEnvOptions to return the AddGroup error")
	}
	if p.hasDotEnvGroup {
		t.Fatal("expected hasDotEnvGroup to remain false after a failed AddGroup")
	}

	p.MaxLongNameLength = DefaultMaxLongNameLength
	if err := p.EnsureBuiltinDotEnvOptions(); err != nil {
		t.Fatalf("expected retry to succeed once MaxLongNameLength no longer blocks it, got: %v", err)
	}
	if !p.hasDotEnvGroup {
		t.Fatal("expected hasDotEnvGroup to be set after a successful retry")
	}
}
