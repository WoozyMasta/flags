package flags

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDeprecatedOptionTagParsed(t *testing.T) {
	var opts struct {
		Old string `long:"old" deprecated:"use --new instead" description:"Old flag"`
		New string `long:"new" description:"New flag"`
	}

	p := NewNamedParser("app", None)
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	opt := p.FindOptionByLongName("old")
	if opt == nil {
		t.Fatal("expected to find --old")
	}
	if opt.Deprecated != "use --new instead" {
		t.Fatalf("expected deprecated reason, got %q", opt.Deprecated)
	}
	if !opt.IsDeprecated() {
		t.Fatal("expected IsDeprecated() to return true")
	}

	newOpt := p.FindOptionByLongName("new")
	if newOpt.IsDeprecated() {
		t.Fatal("expected IsDeprecated() to return false for non-deprecated option")
	}
}

func TestDeprecatedOptionSetterAPI(t *testing.T) {
	var opts struct {
		Flag string `long:"flag" description:"A flag"`
	}

	p := NewNamedParser("app", None)
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	opt := p.FindOptionByLongName("flag")
	if opt == nil {
		t.Fatal("expected to find --flag")
	}
	if opt.IsDeprecated() {
		t.Fatal("expected option to not be deprecated initially")
	}

	opt.SetDeprecated("use --other instead")
	if !opt.IsDeprecated() {
		t.Fatal("expected IsDeprecated() after SetDeprecated")
	}
	if opt.Deprecated != "use --other instead" {
		t.Fatalf("expected reason %q, got %q", "use --other instead", opt.Deprecated)
	}

	opt.SetDeprecated("")
	if opt.IsDeprecated() {
		t.Fatal("expected IsDeprecated() to return false after clearing")
	}
}

func TestDeprecatedCommandTagParsed(t *testing.T) {
	var opts struct {
		Old struct{} `command:"old" deprecated:"use new instead" description:"Old command"`
		New struct{} `command:"new" description:"New command"`
	}

	p := NewParser(&opts, None)

	oldCmd := p.Find("old")
	if oldCmd == nil {
		t.Fatal("expected to find command old")
	}
	if oldCmd.Deprecated != "use new instead" {
		t.Fatalf("expected deprecated reason, got %q", oldCmd.Deprecated)
	}

	newCmd := p.Find("new")
	if newCmd == nil {
		t.Fatal("expected to find command new")
	}
	if newCmd.Deprecated != "" {
		t.Fatalf("expected non-deprecated command to have empty Deprecated, got %q", newCmd.Deprecated)
	}
}

func TestDeprecatedOptionShowsInHelp(t *testing.T) {
	var opts struct {
		Old string `long:"old" deprecated:"use --new instead" description:"Old flag"`
	}

	p := NewNamedParser("app", None)
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var help bytes.Buffer
	p.WriteHelp(&help)

	if !strings.Contains(help.String(), "deprecated") {
		t.Fatalf("expected 'deprecated' in help output, got:\n%s", help.String())
	}
	if !strings.Contains(help.String(), "use --new instead") {
		t.Fatalf("expected deprecation reason in help output, got:\n%s", help.String())
	}
}

func TestDeprecatedCommandShowsInHelp(t *testing.T) {
	var opts struct {
		Old struct{} `command:"old" deprecated:"use new instead" description:"Old command"`
	}

	p := NewParser(&opts, None)

	var help bytes.Buffer
	p.WriteHelp(&help)

	if !strings.Contains(help.String(), "deprecated") {
		t.Fatalf("expected 'deprecated' in help output, got:\n%s", help.String())
	}
	if !strings.Contains(help.String(), "use new instead") {
		t.Fatalf("expected deprecation reason in help output, got:\n%s", help.String())
	}
}

func TestDeprecatedOptionEmitsWarningOnParse(t *testing.T) {
	var opts struct {
		Old string `long:"old" deprecated:"use --new instead"`
		New string `long:"new"`
	}

	p := NewNamedParser("app", None)
	p.Options |= PrintErrorsOnStdout
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out bytes.Buffer
	origStdout := captureParserOutput(p, &out)


	if _, err := p.ParseArgs([]string{"--old", "val"}); err != nil {
		origStdout()
		t.Fatalf("unexpected parse error: %v", err)
	}
	origStdout()

	warning := out.String()
	if !strings.Contains(warning, "old") {
		t.Fatalf("expected flag name in warning output, got: %q", warning)
	}
	if !strings.Contains(warning, "use --new instead") {
		t.Fatalf("expected deprecation reason in warning output, got: %q", warning)
	}
}

func TestDeprecatedOptionWarningOnlyOnFirstUse(t *testing.T) {
	var opts struct {
		Multi []string `long:"multi" deprecated:"use --other instead"`
	}

	p := NewNamedParser("app", None)
	p.Options |= PrintErrorsOnStdout
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out bytes.Buffer
	restore := captureParserOutput(p, &out)

	if _, err := p.ParseArgs([]string{"--multi", "a", "--multi", "b"}); err != nil {
		restore()
		t.Fatalf("unexpected parse error: %v", err)
	}
	restore()

	count := strings.Count(out.String(), "use --other instead")
	if count != 1 {
		t.Fatalf("expected warning exactly once, got %d occurrences in: %q", count, out.String())
	}
}

func TestDeprecatedOptionSilenced(t *testing.T) {
	var opts struct {
		Old string `long:"old" deprecated:"use --new instead"`
	}

	p := NewNamedParser("app", None)
	p.Options |= SilenceDeprecationWarnings | PrintErrorsOnStdout
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out bytes.Buffer
	restore := captureParserOutput(p, &out)
	defer restore()

	if _, err := p.ParseArgs([]string{"--old", "val"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if out.Len() > 0 {
		t.Fatalf("expected no output with SilenceDeprecationWarnings, got: %q", out.String())
	}
}

func TestDeprecatedCommandEmitsWarningOnParse(t *testing.T) {
	var opts struct {
		Old struct{} `command:"old" deprecated:"use new instead"`
	}

	p := NewParser(&opts, None)
	p.Options |= PrintErrorsOnStdout

	var out bytes.Buffer
	restore := captureParserOutput(p, &out)

	if _, err := p.ParseArgs([]string{"old"}); err != nil {
		restore()
		t.Fatalf("unexpected parse error: %v", err)
	}
	restore()

	warning := out.String()
	if !strings.Contains(warning, "old") {
		t.Fatalf("expected command name in warning output, got: %q", warning)
	}
	if !strings.Contains(warning, "use new instead") {
		t.Fatalf("expected deprecation reason in warning output, got: %q", warning)
	}
}

func TestDeprecatedCommandSilenced(t *testing.T) {
	var opts struct {
		Old struct{} `command:"old" deprecated:"use new instead"`
	}

	p := NewParser(&opts, None)
	p.Options |= SilenceDeprecationWarnings | PrintErrorsOnStdout

	var out bytes.Buffer
	restore := captureParserOutput(p, &out)
	defer restore()

	if _, err := p.ParseArgs([]string{"old"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if out.Len() > 0 {
		t.Fatalf("expected no output with SilenceDeprecationWarnings, got: %q", out.String())
	}
}

func TestDeprecatedOptionShowsInDocTemplates(t *testing.T) {
	var opts struct {
		Old string `long:"old" deprecated:"use --new instead" description:"Old flag"`
	}

	p := NewNamedParser("app", None)
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, tt := range []struct {
		name   string
		format DocFormat
		opts   []DocOption
	}{
		{name: "markdown-list", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownList)}},
		{name: "markdown-table", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownTable)}},
		{name: "markdown-code", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownCode)}},
		{name: "html-default", format: DocFormatHTML, opts: []DocOption{WithBuiltinTemplate(DocTemplateHTMLDefault)}},
		{name: "html-styled", format: DocFormatHTML, opts: []DocOption{WithBuiltinTemplate(DocTemplateHTMLStyled)}},
		{name: "man", format: DocFormatMan, opts: []DocOption{WithBuiltinTemplate(DocTemplateManDefault)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := p.WriteDoc(&out, tt.format, tt.opts...); err != nil {
				t.Fatalf("unexpected write doc error: %v", err)
			}
			if !strings.Contains(out.String(), "use --new instead") {
				t.Fatalf("expected deprecation reason in %s output, got:\n%s", tt.name, out.String())
			}
		})
	}
}

func TestDeprecatedCommandShowsInDocTemplates(t *testing.T) {
	var opts struct {
		Old struct {
			Flag string `long:"flag"`
		} `command:"old" deprecated:"use new instead" description:"Old command"`
	}

	p := NewParser(&opts, None)

	for _, tt := range []struct {
		name   string
		format DocFormat
		opts   []DocOption
	}{
		{name: "markdown-list", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownList)}},
		{name: "markdown-table", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownTable)}},
		{name: "markdown-code", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownCode)}},
		{name: "html-default", format: DocFormatHTML, opts: []DocOption{WithBuiltinTemplate(DocTemplateHTMLDefault)}},
		{name: "html-styled", format: DocFormatHTML, opts: []DocOption{WithBuiltinTemplate(DocTemplateHTMLStyled)}},
		{name: "man", format: DocFormatMan, opts: []DocOption{WithBuiltinTemplate(DocTemplateManDefault)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := p.WriteDoc(&out, tt.format, tt.opts...); err != nil {
				t.Fatalf("unexpected write doc error: %v", err)
			}
			if !strings.Contains(out.String(), "use new instead") {
				t.Fatalf("expected deprecation reason in %s output, got:\n%s", tt.name, out.String())
			}
		})
	}
}

func TestDeprecatedOptionUsesCustomTagName(t *testing.T) {
	var opts struct {
		Old string `long:"old" superseded-by:"--new" description:"Old flag"`
	}

	p := NewNamedParser("app", None)
	tags := NewFlagTags()
	tags.Deprecated = "superseded-by"
	if err := p.SetFlagTags(tags); err != nil {
		t.Fatalf("unexpected set flag tags error: %v", err)
	}
	if _, err := p.AddGroup("Options", "", &opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	opt := p.FindOptionByLongName("old")
	if opt == nil {
		t.Fatal("expected to find --old")
	}
	if opt.Deprecated != "--new" {
		t.Fatalf("expected deprecated reason %q, got %q", "--new", opt.Deprecated)
	}
}

// captureParserOutput redirects the parser's PrintErrorsOnStdout output to buf.
// Returns a restore function that should be deferred.
func captureParserOutput(p *Parser, buf *bytes.Buffer) func() {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w

	done := make(chan struct{})
	go func() {
		defer close(done)
		b := make([]byte, 4096)
		for {
			n, err := r.Read(b)
			if n > 0 {
				buf.Write(b[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	return func() {
		w.Close()
		<-done
		r.Close()
		os.Stdout = orig
	}
}
