// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"strings"
	"testing"
)

func assertFlagDidYouMean(t *testing.T, input string, wantSuggestion string, opts interface{}) {
	t.Helper()
	p := NewParser(opts, None)
	p.SetHelpFlagRenderStyle(RenderStylePOSIX)
	_, err := p.ParseArgs([]string{input})
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	e, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T: %v", err, err)
	}
	if e.Type != ErrUnknownFlag {
		t.Fatalf("expected ErrUnknownFlag, got %v", e.Type)
	}
	if wantSuggestion == "" {
		if strings.Contains(e.Message, "did you mean") {
			t.Fatalf("expected no suggestion, got: %q", e.Message)
		}
	} else {
		if !strings.Contains(e.Message, wantSuggestion) {
			t.Fatalf("expected suggestion %q in: %q", wantSuggestion, e.Message)
		}
	}
}

func TestFlagDidYouMeanTypo(t *testing.T) {
	var opts struct {
		Output  string `long:"output"`
		Verbose bool   `long:"verbose" short:"v"`
	}
	assertFlagDidYouMean(t, "--ouput", "--output", &opts)
}

func TestFlagDidYouMeanMissingChar(t *testing.T) {
	var opts struct {
		Verbose bool `long:"verbose"`
	}
	assertFlagDidYouMean(t, "--verbos", "--verbose", &opts)
}

func TestFlagDidYouMeanNoSuggestionForUnrelated(t *testing.T) {
	var opts struct {
		Output  string `long:"output"`
		Verbose bool   `long:"verbose"`
	}
	assertFlagDidYouMean(t, "--xyz", "", &opts)
}

func TestFlagDidYouMeanInCommandScope(t *testing.T) {
	var opts struct {
		Deploy struct {
			Region string `long:"region"`
		} `command:"deploy"`
	}
	// Command scope: use assertParseFail but match platform-agnostic flag name
	p := NewParser(&opts, None)
	p.SetHelpFlagRenderStyle(RenderStylePOSIX)
	_, err := p.ParseArgs([]string{"deploy", "--regio"})
	if err == nil {
		t.Fatal("expected parse error")
	}
	e, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if e.Type != ErrUnknownFlag {
		t.Fatalf("expected ErrUnknownFlag, got %v", e.Type)
	}
	if !strings.Contains(e.Message, "--region") {
		t.Fatalf("expected '--region' suggestion in: %q", e.Message)
	}
}

func TestFlagDidYouMeanShortFlagNoSuggestion(t *testing.T) {
	var opts struct {
		Verbose bool `long:"verbose" short:"v"`
	}
	// short flags (-x) do not get did-you-mean suggestions
	assertParseFail(t, ErrUnknownFlag, "unknown flag `x`", &opts, "-x")
}
