package flags

import (
	"bytes"
	"strings"
	"testing"
)

func TestSecretOptionRedactsHelpDocsAndINI(t *testing.T) {
	var opts struct {
		Token  string `long:"token" env:"APP_TOKEN" default:"default-secret" optional:"true" optional-value:"optional-secret" default-mask:"mask-secret" secret:"true" description:"API token"`
		Choice string `long:"choice" choice:"choice-secret" secret:"true" description:"Secret choice"`
	}

	p := NewNamedParser("secret-app", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var help bytes.Buffer
	p.WriteHelp(&help)
	assertSecretOutputRedacted(t, "help", help.String(), []string{
		"default-secret",
		"optional-secret",
		"choice-secret",
		"mask-secret",
	})
	if !strings.Contains(help.String(), secretValueMask) {
		t.Fatalf("expected secret mask in help output, got:\n%s", help.String())
	}
	if !strings.Contains(help.String(), "APP_TOKEN") {
		t.Fatalf("expected env key to remain visible in help output, got:\n%s", help.String())
	}

	for _, tt := range []struct {
		name   string
		format DocFormat
		opts   []DocOption
	}{
		{name: "markdown", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownList)}},
		{name: "html", format: DocFormatHTML, opts: []DocOption{WithBuiltinTemplate(DocTemplateHTMLDefault)}},
		{name: "man", format: DocFormatMan, opts: []DocOption{WithBuiltinTemplate(DocTemplateManDefault)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := p.WriteDoc(&out, tt.format, tt.opts...); err != nil {
				t.Fatalf("unexpected write doc error: %v", err)
			}
			assertSecretOutputRedacted(t, tt.name, out.String(), []string{
				"default-secret",
				"optional-secret",
				"choice-secret",
				"mask-secret",
			})
			if !strings.Contains(out.String(), secretValueMask) {
				t.Fatalf("expected secret mask in %s output, got:\n%s", tt.name, out.String())
			}
		})
	}

	var custom bytes.Buffer
	err := p.WriteDoc(
		&custom,
		DocFormatMarkdown,
		WithTemplateString(`{{ $o := index (index .Doc.Groups 0).Options 0 }}{{ index $o.Tags "default" }}|{{ index $o.Tags "default-mask" }}|{{ index $o.Tags "optional-value" }}`),
	)
	if err != nil {
		t.Fatalf("unexpected custom doc error: %v", err)
	}
	assertSecretOutputRedacted(t, "custom doc", custom.String(), []string{
		"default-secret",
		"optional-secret",
		"mask-secret",
	})

	var iniOpts struct {
		Token string `long:"token" ini-name:"Token" secret:"true"`
	}
	iniParser := NewNamedParser("secret-ini", None)
	if _, err := iniParser.AddGroup("Application Options", "", &iniOpts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}
	if _, err := iniParser.ParseArgs([]string{"--token", "runtime-secret"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	var ini bytes.Buffer
	NewIniParser(iniParser).Write(&ini, IniIncludeDefaults)
	assertSecretOutputRedacted(t, "ini", ini.String(), []string{"runtime-secret"})
	if !strings.Contains(ini.String(), secretValueMask) {
		t.Fatalf("expected secret mask in ini output, got:\n%s", ini.String())
	}

	var example bytes.Buffer
	NewIniParser(p).WriteExample(&example)
	assertSecretOutputRedacted(t, "ini example", example.String(), []string{
		"default-secret",
		"choice-secret",
	})
	if !strings.Contains(example.String(), secretValueMask) {
		t.Fatalf("expected secret mask in ini example output, got:\n%s", example.String())
	}
}

func TestSecretOptionRedactsParserErrors(t *testing.T) {
	var opts struct {
		Choice string `long:"choice" choice:"allowed-secret" secret:"true"`
		Number int    `long:"number" secret:"true"`
		Regex  string `long:"regex" validate-regex:"^allowed$" secret:"true"`
	}

	p := NewNamedParser("secret-errors", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	tests := []struct {
		name   string
		args   []string
		leaked []string
	}{
		{
			name:   "choice",
			args:   []string{"--choice", "provided-secret"},
			leaked: []string{"provided-secret", "allowed-secret"},
		},
		{
			name:   "marshal",
			args:   []string{"--number", "number-secret"},
			leaked: []string{"number-secret"},
		},
		{
			name:   "validation",
			args:   []string{"--regex", "regex-secret"},
			leaked: []string{"regex-secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.ParseArgs(tt.args)
			if err == nil {
				t.Fatalf("expected parse error")
			}
			msg := err.Error()
			assertSecretOutputRedacted(t, tt.name, msg, tt.leaked)
			if !strings.Contains(msg, secretValueMask) {
				t.Fatalf("expected secret mask in error, got: %s", msg)
			}
		})
	}
}

func TestSecretOptionDoesNotChangeStoredValue(t *testing.T) {
	var opts struct {
		Token string `long:"token" secret:"true"`
	}

	p := NewNamedParser("secret-storage", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	if _, err := p.ParseArgs([]string{"--token", "runtime-secret"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Token != "runtime-secret" {
		t.Fatalf("expected stored value to remain unchanged, got %q", opts.Token)
	}
}

func TestSecretOptionCanBeSetProgrammatically(t *testing.T) {
	var opts struct {
		Token string `long:"token" default:"programmatic-secret" description:"API token"`
	}

	p := NewNamedParser("secret-api", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	opt := p.FindOptionByLongName("token")
	if opt == nil {
		t.Fatalf("expected token option")
	}
	opt.SetSecret(true)
	if !opt.IsSecret() {
		t.Fatalf("expected option to report secret")
	}

	var help bytes.Buffer
	p.WriteHelp(&help)
	assertSecretOutputRedacted(t, "help", help.String(), []string{"programmatic-secret"})
	if !strings.Contains(help.String(), secretValueMask) {
		t.Fatalf("expected secret mask in help output, got:\n%s", help.String())
	}
}

func TestSecretOptionUsesCustomTagMapping(t *testing.T) {
	var opts struct {
		Token string `long:"token" default:"custom-secret" sensitive:"true" description:"API token"`
	}

	p := NewNamedParser("secret-custom-tag", None)
	tags := NewFlagTags()
	tags.Secret = "sensitive"
	if err := p.SetFlagTags(tags); err != nil {
		t.Fatalf("unexpected set flag tags error: %v", err)
	}
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var help bytes.Buffer
	p.WriteHelp(&help)
	assertSecretOutputRedacted(t, "help", help.String(), []string{"custom-secret"})
	if !strings.Contains(help.String(), secretValueMask) {
		t.Fatalf("expected secret mask in help output, got:\n%s", help.String())
	}
}

func assertSecretOutputRedacted(t *testing.T, context string, text string, leaked []string) {
	t.Helper()

	for _, secret := range leaked {
		if strings.Contains(text, secret) {
			t.Fatalf("did not expect %q in %s output:\n%s", secret, context, text)
		}
	}
}
