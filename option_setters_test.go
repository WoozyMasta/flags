package flags

import (
	"bytes"
	"os"
	"strings"
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
	if err := opt.SetShortName('p'); err != nil {
		t.Fatalf("unexpected SetShortName error: %v", err)
	}
	if err := opt.SetShortAliases('P'); err != nil {
		t.Fatalf("unexpected SetShortAliases error: %v", err)
	}

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
	if err := opt.SetEnv("APP_MODE", ""); err != nil {
		t.Fatalf("unexpected SetEnv error: %v", err)
	}
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

func TestOptionSettersBaseMapDelimiterAndUnquote(t *testing.T) {
	var opts struct {
		Hex    int               `long:"hex"`
		Labels map[string]string `long:"label"`
		Name   string            `long:"name"`
	}

	p := NewParser(&opts, None)

	hexOpt := p.FindOptionByLongName("hex")
	if hexOpt == nil {
		t.Fatalf("expected hex option")
	}
	if err := hexOpt.SetBase(16); err != nil {
		t.Fatalf("unexpected SetBase error: %v", err)
	}

	labelsOpt := p.FindOptionByLongName("label")
	if labelsOpt == nil {
		t.Fatalf("expected labels option")
	}
	labelsOpt.SetKeyValueDelimiter("=")

	nameOpt := p.FindOptionByLongName("name")
	if nameOpt == nil {
		t.Fatalf("expected name option")
	}
	nameOpt.SetUnquote(false)

	_, err := p.ParseArgs([]string{"--hex", "ff", "--label", "a=1", "--name", "\"bob\""})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Hex != 255 {
		t.Fatalf("expected hex value 255, got %d", opts.Hex)
	}
	if got := opts.Labels["a"]; got != "1" {
		t.Fatalf("expected labels[a]=1, got %q", got)
	}
	if opts.Name != "\"bob\"" {
		t.Fatalf("expected quoted value to be preserved, got %q", opts.Name)
	}
}

func TestOptionSetterSetBaseValidation(t *testing.T) {
	var opts struct {
		Value int `long:"value"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("value")
	if opt == nil {
		t.Fatalf("expected option")
	}

	if err := opt.SetBase(1); err == nil {
		t.Fatalf("expected invalid base validation error")
	}
}

func TestOptionSettersIniAndNoIni(t *testing.T) {
	var opts struct {
		Name string `long:"name"`
		Skip string `long:"skip"`
	}

	p := NewParser(&opts, None)

	nameOpt := p.FindOptionByLongName("name")
	if nameOpt == nil {
		t.Fatalf("expected name option")
	}
	nameOpt.SetIniName("display_name")

	skipOpt := p.FindOptionByLongName("skip")
	if skipOpt == nil {
		t.Fatalf("expected skip option")
	}
	skipOpt.SetNoIni(true)

	ini := NewIniParser(p)
	input := strings.NewReader("[Application Options]\ndisplay_name = alice\n")
	if err := ini.Parse(input); err != nil {
		t.Fatalf("unexpected ini parse error: %v", err)
	}

	if opts.Name != "alice" {
		t.Fatalf("expected name from custom ini key, got %q", opts.Name)
	}
	if opts.Skip != "" {
		t.Fatalf("expected no-ini option to be ignored on parse, got %q", opts.Skip)
	}

	var out bytes.Buffer
	ini.Write(&out, IniNone)
	rendered := out.String()
	if strings.Contains(rendered, "skip =") {
		t.Fatalf("expected no-ini option to be excluded from ini write, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "display_name = alice") {
		t.Fatalf("expected custom ini name in output, got:\n%s", rendered)
	}
}

func TestOptionSetterAutoEnv(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		SomeFunction string `long:"some-function"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("some-function")
	if opt == nil {
		t.Fatalf("expected option")
	}
	if err := opt.SetAutoEnv(true); err != nil {
		t.Fatalf("unexpected SetAutoEnv error: %v", err)
	}

	_ = os.Setenv("SOME_FUNCTION", "from-env")
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if opts.SomeFunction != "from-env" {
		t.Fatalf("expected env-derived value, got %q", opts.SomeFunction)
	}
}

func TestOptionSetterRejectsDuplicateShortName(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose"`
		Value   bool `short:"x" long:"value"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("value")
	if opt == nil {
		t.Fatalf("expected option")
	}

	err := opt.SetShortName('v')
	if err == nil {
		t.Fatalf("expected duplicate short-name error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestOptionSetterRejectsDuplicateLongName(t *testing.T) {
	var opts struct {
		One string `long:"one"`
		Two string `long:"two"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("two")
	if opt == nil {
		t.Fatalf("expected option")
	}

	err := opt.SetLongName("one")
	if err == nil {
		t.Fatalf("expected duplicate long-name error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestOptionSetterRejectsDuplicateEnvKey(t *testing.T) {
	var opts struct {
		First  string `long:"first" env:"APP_SHARED"`
		Second string `long:"second"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByLongName("second")
	if opt == nil {
		t.Fatalf("expected option")
	}

	err := opt.SetEnv("APP_SHARED", "")
	if err == nil {
		t.Fatalf("expected duplicate env-key error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestOptionSetterAutoEnvRequiresLongName(t *testing.T) {
	var opts struct {
		Value string `short:"v"`
	}

	p := NewParser(&opts, None)
	opt := p.FindOptionByShortName('v')
	if opt == nil {
		t.Fatalf("expected option")
	}

	if err := opt.SetAutoEnv(true); err == nil {
		t.Fatalf("expected SetAutoEnv error for option without long name")
	}
}
