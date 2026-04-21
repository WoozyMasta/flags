package flags

import (
	"strings"
	"testing"
)

func TestPassDoubleDash(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, PassDoubleDash)
	ret, err := p.ParseArgs([]string{"-v", "--", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{"-v", "-g"})
}

func TestPassAfterNonOption(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositional(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{})
	assertStringArray(t, opts.Positional.Rest, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "1", "2", "3"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{})
	for i, rest := range opts.Positional.Rest {
		if rest != i+1 {
			assertErrorf(t, "Expected %v got %v", i+1, rest)
		}
	}
}

func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	tests := []struct {
		opts        []string
		errContains string
		ret         []string
	}{
		{
			[]string{"-v", "notint1", "notint2", "notint3"},
			"notint1",
			[]string{"notint1", "notint2", "notint3"},
		},
		{
			[]string{"-v", "1", "notint2", "notint3"},
			"notint2",
			[]string{"1", "notint2", "notint3"},
		},
	}

	for _, test := range tests {
		p := NewParser(&opts, PassAfterNonOption)
		ret, err := p.ParseArgs(test.opts)

		if err == nil {
			assertErrorf(t, "Expected error")
			return
		}

		if !strings.Contains(err.Error(), test.errContains) {
			assertErrorf(t, "Expected the first illegal argument in the error")
		}

		assertStringArray(t, ret, test.ret)
	}
}

func TestOptionPublicAccessorsAndNames(t *testing.T) {
	var opts struct {
		Alpha   int  `short:"a" long:"alpha" default:"5"`
		Beta    bool `long:"beta"`
		Charlie bool `short:"c"`
	}

	p := NewParser(&opts, None)
	optAlpha := p.FindOptionByLongName("alpha")
	if optAlpha == nil {
		t.Fatalf("option alpha not found")
	}

	if optAlpha.Field().Name != "Alpha" {
		t.Fatalf("unexpected field: %s", optAlpha.Field().Name)
	}

	if optAlpha.Value().(int) != 0 {
		t.Fatalf("expected zero value before parse")
	}

	if optAlpha.IsSet() || optAlpha.IsSetDefault() {
		t.Fatalf("expected unset option state before parse")
	}

	if got := optAlpha.shortAndLongName(); got != string(defaultShortOptDelimiter)+"a/"+optAlpha.LongName {
		t.Fatalf("unexpected short/long name: %q", got)
	}

	if _, err := p.ParseArgs([]string{}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !optAlpha.IsSet() || !optAlpha.IsSetDefault() {
		t.Fatalf("expected default to mark option as set via default")
	}

	if optAlpha.Value().(int) != 5 {
		t.Fatalf("expected parsed default value 5, got %v", optAlpha.Value())
	}

	optBeta := p.FindOptionByLongName("beta")
	if optBeta == nil {
		t.Fatalf("option beta not found")
	}

	if got := optBeta.shortAndLongName(); got != optBeta.LongName {
		t.Fatalf("unexpected long-only name: %q", got)
	}

	optCharlie := p.FindOptionByShortName('c')
	if optCharlie == nil {
		t.Fatalf("option charlie not found")
	}

	if got := optCharlie.shortAndLongName(); got != string(defaultShortOptDelimiter)+"c" {
		t.Fatalf("unexpected short-only name: %q", got)
	}
}
