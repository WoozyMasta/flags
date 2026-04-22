package flags

import (
	"os"
	"testing"
)

func TestOptionSettersLookupAndParse(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("value")
	if opt == nil {
		t.Fatalf("expected option")
	}

	if err := opt.SetLongName("primary"); err != nil {
		t.Fatalf("unexpected SetLongName error: %v", err)
	}
	if err := opt.SetLongAliases("legacy", "compat"); err != nil {
		t.Fatalf("unexpected SetLongAliases error: %v", err)
	}
	opt.SetShortName('p')
	opt.SetShortAliases('P')

	if _, err := p.ParseArgs([]string{"--legacy", "one"}); err != nil {
		t.Fatalf("unexpected parse error for long alias: %v", err)
	}
	if opts.Value != "one" {
		t.Fatalf("expected value from long alias, got %q", opts.Value)
	}

	if _, err := p.ParseArgs([]string{"-P", "two"}); err != nil {
		t.Fatalf("unexpected parse error for short alias: %v", err)
	}
	if opts.Value != "two" {
		t.Fatalf("expected value from short alias, got %q", opts.Value)
	}
}

func TestOptionSetterLongNameLengthValidation(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewParser(&opts, None)
	if err := p.SetMaxLongNameLength(5); err != nil {
		t.Fatalf("unexpected SetMaxLongNameLength error: %v", err)
	}

	opt := p.FindOptionByLongName("value")
	if opt == nil {
		t.Fatalf("expected option")
	}

	if err := opt.SetLongName("too-long-name"); err == nil {
		t.Fatalf("expected length validation error")
	}
}

func TestOptionSettersDefaultsChoicesEnv(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		Mode string `long:"mode"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("mode")
	if opt == nil {
		t.Fatalf("expected option")
	}

	opt.SetChoices("fast", "safe")
	opt.SetDefault("fast")
	opt.SetEnv("APP_MODE", "")
	opt.SetRequired(true)

	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error with default: %v", err)
	}
	if opts.Mode != "fast" {
		t.Fatalf("expected default mode fast, got %q", opts.Mode)
	}

	_ = os.Setenv("APP_MODE", "broken")
	_, err := p.ParseArgs(nil)
	if err == nil {
		t.Fatalf("expected invalid choice error from env")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrInvalidChoice {
		t.Fatalf("expected ErrInvalidChoice, got %v", err)
	}
}
