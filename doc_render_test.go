package flags

import (
	"bytes"
	"strings"
	"testing"
)

func TestListBuiltinTemplates(t *testing.T) {
	got := ListBuiltinTemplates()
	want := []string{
		DocTemplateHTMLDefault,
		DocTemplateHTMLStyled,
		DocTemplateManDefault,
		DocTemplateMarkdownCode,
		DocTemplateMarkdownList,
		DocTemplateMarkdownTable,
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected template count: got %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected template at %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestWriteBuiltinTemplate(t *testing.T) {
	var out bytes.Buffer

	if err := WriteBuiltinTemplate(&out, DocTemplateMarkdownList); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), `doc.tmpl.markdown.section.options`) {
		t.Fatalf("expected markdown template content, got: %q", out.String())
	}

	if err := WriteBuiltinTemplate(&out, "unknown/template"); err == nil {
		t.Fatalf("expected error for unknown template")
	}
}

func TestWriteDocMarkdownBuiltin(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" description:"Enable verbose output"`
		Run     struct {
			Force bool `long:"force" description:"Force execution"`
		} `command:"run" description:"Run command"`
	}

	p := NewNamedParser("doc-app", None)
	p.ShortDescription = "Doc app"
	p.LongDescription = "Long description"

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"# doc-app",
		"## OPTIONS",
		defaultLongOptDelimiter + "verbose",
		"## COMMANDS",
		"### run",
		defaultLongOptDelimiter + "force",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in markdown output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocMarkdownBuiltinLocalized(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" required:"true" description:"Enable verbose output"`
	}

	p := NewNamedParser("doc-i18n", None)
	p.ShortDescription = "Doc i18n"
	p.LongDescription = "Long description"

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	p.SetI18n(I18nConfig{Locale: "ru"})

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownList)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"## ОПЦИИ",
		"Обязательно: `да`",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in localized markdown output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocMarkdownListOptionDescriptionsUseContinuationLine(t *testing.T) {
	var opts struct {
		Locale string `short:"l" long:"locale" value-name:"LOCALE" description:"Override language for help, errors, and application text"`
		Help   bool   `short:"h" long:"help" description:"Show help"`
	}

	p := NewNamedParser("doc-list", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownList)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	shortLocale := string(defaultShortOptDelimiter) + "l LOCALE"
	longLocale := string(defaultLongOptDelimiter) + "locale LOCALE"
	shortHelp := string(defaultShortOptDelimiter) + "h"
	longHelp := string(defaultLongOptDelimiter) + "help"
	for _, needle := range []string{
		"* `" + shortLocale + "`, `" + longLocale + "` -\n  Override language for help, errors, and application text",
		"* `" + shortHelp + "`, `" + longHelp + "` -\n  Show help",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in markdown list output, got:\n%s", needle, got)
		}
	}

	if strings.Contains(got, "application text\n\n* `"+shortHelp+"`") {
		t.Fatalf("did not expect blank line between option items, got:\n%s", got)
	}
}

func TestWriteDocMarkdownUsesLocalizedRootDescription(t *testing.T) {
	p := NewNamedParser("doc-i18n-root", None)
	p.Command.SetLongDescriptionI18nKey("app.description")
	p.SetI18n(I18nConfig{
		Locale: "ru",
		UserCatalog: mapCatalog{
			"ru": {"app.description": "Локализованное описание приложения"},
		},
	})

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownList)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "## ОПИСАНИЕ") {
		t.Fatalf("expected localized description section, got:\n%s", got)
	}
	if !strings.Contains(got, "Локализованное описание приложения") {
		t.Fatalf("expected localized root description, got:\n%s", got)
	}
}

func TestWriteDocBuiltinTemplatesSkipEmptyRootSections(t *testing.T) {
	formats := []struct {
		name     string
		format   DocFormat
		template string
		empty    []string
	}{
		{
			name:     "markdown-list",
			format:   DocFormatMarkdown,
			template: DocTemplateMarkdownList,
			empty:    []string{"## DESCRIPTION", "## OPTIONS"},
		},
		{
			name:     "markdown-table",
			format:   DocFormatMarkdown,
			template: DocTemplateMarkdownTable,
			empty:    []string{"## DESCRIPTION", "## OPTIONS"},
		},
		{
			name:     "markdown-code",
			format:   DocFormatMarkdown,
			template: DocTemplateMarkdownCode,
			empty:    []string{"## DESCRIPTION", "## OPTIONS"},
		},
		{
			name:     "man",
			format:   DocFormatMan,
			template: DocTemplateManDefault,
			empty:    []string{".SH DESCRIPTION", ".SH OPTIONS"},
		},
		{
			name:     "html",
			format:   DocFormatHTML,
			template: DocTemplateHTMLDefault,
			empty:    []string{"<h2>Description</h2>", "<h2>Options</h2>"},
		},
		{
			name:     "html-styled",
			format:   DocFormatHTML,
			template: DocTemplateHTMLStyled,
			empty:    []string{"<h2>Description</h2>", "<h2>Options</h2>"},
		},
	}

	for _, tt := range formats {
		t.Run(tt.name, func(t *testing.T) {
			p := NewNamedParser("empty-doc", None)

			var out bytes.Buffer
			if err := p.WriteDoc(&out, tt.format, WithBuiltinTemplate(tt.template)); err != nil {
				t.Fatalf("unexpected write doc error: %v", err)
			}

			got := out.String()
			for _, needle := range tt.empty {
				if strings.Contains(got, needle) {
					t.Fatalf("did not expect empty section %q, got:\n%s", needle, got)
				}
			}
		})
	}
}

func TestWriteDocMarkdownBuiltinTable(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" description:"Enable verbose output"`
	}

	p := NewNamedParser("doc-table", None)
	p.ShortDescription = "Doc table"
	p.LongDescription = "Long description"

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownTable)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"|Option|Description|Required|",
		"|---|---|---|",
		defaultLongOptDelimiter + "verbose",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in markdown output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocMarkdownBuiltinTableSkipsEmptyColumns(t *testing.T) {
	var opts struct {
		Plain   bool   `long:"plain" description:"Plain option"`
		EnvOnly string `long:"env-only" env:"APP_ENV" description:"Env option"`
		Nested  struct {
			Help bool `long:"help" description:"Help option"`
		} `group:"Nested"`
	}

	p := NewNamedParser("doc-table-columns", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownTable)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "|Option|Description|Default|Environment|Required|") {
		t.Fatalf("expected default/env columns for group with env-derived default, got:\n%s", got)
	}
	if !strings.Contains(got, "|Option|Description|Required|") {
		t.Fatalf("expected compact table for group with empty default/env columns, got:\n%s", got)
	}
}

func TestWriteDocMarkdownBuiltinCode(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" description:"Enable verbose output"`
	}

	p := NewNamedParser("doc-code", None)
	p.ShortDescription = "Doc code"
	p.LongDescription = "Long description"

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownCode)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"```text",
		defaultLongOptDelimiter + "verbose",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in markdown output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocCustomTemplate(t *testing.T) {
	var opts struct {
		Value string `long:"value" description:"Value"`
	}

	p := NewNamedParser("custom-doc", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	err := p.WriteDoc(
		&out,
		DocFormatMarkdown,
		WithTemplateString("{{ .Doc.Name }}|{{ index .Data \"mode\" }}"),
		WithTemplateData(map[string]any{"mode": "custom"}),
	)
	if err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	if got := out.String(); got != "custom-doc|custom" {
		t.Fatalf("unexpected custom template output: %q", got)
	}
}

func TestWriteDocTemplateHelpers(t *testing.T) {
	var opts struct {
		Flag  bool     `long:"flag" required:"true" description:"A long line that should be wrapped for helper verification"`
		List  []string `long:"list" env:"APP_LIST"`
		Value string   `long:"value"`
	}

	p := NewNamedParser("helpers-doc", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	tpl := strings.Join([]string{
		"{{ $g := index .Doc.Groups 0 }}",
		"{{ $flag := index $g.Options 0 }}",
		"{{ $list := index $g.Options 1 }}",
		"{{ if isRequired $flag }}req{{ end }}",
		"|{{ if isBool $flag }}bool{{ end }}",
		"|{{ if isCollection $list }}collection{{ end }}",
		"|{{ if hasEnv $list }}env={{ $list.Env }}{{ end }}",
		"|{{ if hasDefault $list }}default{{ else }}no-default{{ end }}",
		"|{{ wrap \"a b c d e\" 3 }}",
		"|{{ indent \"x\\ny\" 2 }}",
		"|{{ markdownWrap \"a b c d e f g h i j\" 8 }}",
	}, "")

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithTemplateString(tpl)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"req|bool|collection|env=APP_LIST|default|",
		"|a b c d e|",
		"  x\n  y",
		"|a b c d e\nf g h i j",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in helper output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocManDefault(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" description:"Enable verbose output"`
	}

	p := NewNamedParser("doc-man", None)
	p.ShortDescription = "Doc man"
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMan); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	if got := out.String(); !strings.Contains(got, ".TH doc-man 1") {
		t.Fatalf("expected man output header, got:\n%s", got)
	}
}

func TestWriteDocManCustomTemplate(t *testing.T) {
	var opts struct {
		Value string `long:"value" description:"Value"`
	}

	p := NewNamedParser("custom-man", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	err := p.WriteDoc(
		&out,
		DocFormatMan,
		WithTemplateString("{{ .Doc.Name }}|{{ quoteMan \"a\\\\b\" }}"),
	)
	if err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	if got := out.String(); got != "custom-man|a\\\\b" {
		t.Fatalf("unexpected custom man template output: %q", got)
	}
}

func TestWriteDocHTMLDefault(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" description:"Enable <verbose> output"`
	}

	p := NewNamedParser("doc-html", None)
	p.ShortDescription = "Doc html"
	p.LongDescription = "Long <description>"
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatHTML); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"<!doctype html>",
		"<h1>doc-html</h1>",
		"&lt;description&gt;",
		defaultLongOptDelimiter + "verbose",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in html output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocHTMLStyled(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" description:"Enable <verbose> output"`
	}

	p := NewNamedParser("doc-html-styled", None)
	p.ShortDescription = "Doc html styled"
	p.LongDescription = "Long <description>"
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatHTML, WithBuiltinTemplate(DocTemplateHTMLStyled)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"<!doctype html>",
		"prefers-color-scheme: dark",
		"<h1>doc-html-styled</h1>",
		"&lt;description&gt;",
		defaultLongOptDelimiter + "verbose",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in styled html output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocHTMLCustomTemplate(t *testing.T) {
	var opts struct {
		Value string `long:"value" description:"Value"`
	}

	p := NewNamedParser("custom-html", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	err := p.WriteDoc(
		&out,
		DocFormatHTML,
		WithTemplateString("{{ .Doc.Name }}|{{ quoteHTML \"a<b\" }}"),
	)
	if err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	if got := out.String(); got != "custom-html|a&lt;b" {
		t.Fatalf("unexpected custom html template output: %q", got)
	}
}

func TestWriteDocErrors(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewNamedParser("doc-errors", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	if err := p.WriteDoc(nil, DocFormatMarkdown); err == nil {
		t.Fatalf("expected nil writer error")
	}

	var out bytes.Buffer

	if err := p.WriteDoc(&out, DocFormatMan, WithBuiltinTemplate("unknown/template")); err == nil {
		t.Fatalf("expected unknown man builtin template error")
	}

	if err := p.WriteDoc(&out, DocFormatHTML, WithBuiltinTemplate("unknown/template")); err == nil {
		t.Fatalf("expected unknown html builtin template error")
	}
}

func TestWriteDocIncludeHidden(t *testing.T) {
	var opts struct {
		Visible string `long:"visible" description:"Visible option"`
		Hidden  string `long:"hidden" description:"Hidden option" hidden:"true"`
		HCmd    struct {
			Flag string `long:"flag" description:"Hidden command flag"`
		} `command:"internal" description:"Hidden cmd" hidden:"true"`
	}

	p := NewNamedParser("hidden-doc", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var withoutHidden bytes.Buffer
	if err := p.WriteDoc(&withoutHidden, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownList)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}
	if strings.Contains(withoutHidden.String(), "--hidden") || strings.Contains(withoutHidden.String(), "internal") {
		t.Fatalf("did not expect hidden entities in default output:\n%s", withoutHidden.String())
	}

	var withHidden bytes.Buffer
	if err := p.WriteDoc(
		&withHidden,
		DocFormatMarkdown,
		WithBuiltinTemplate(DocTemplateMarkdownList),
		WithIncludeHidden(true),
	); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	hiddenFlag := string(defaultLongOptDelimiter) + "hidden"
	for _, needle := range []string{hiddenFlag, "### internal"} {
		if !strings.Contains(withHidden.String(), needle) {
			t.Fatalf("expected %q in hidden-enabled output, got:\n%s", needle, withHidden.String())
		}
	}

	if strings.Contains(withHidden.String(), "*(hidden)") {
		t.Fatalf("did not expect hidden markers by default, got:\n%s", withHidden.String())
	}

	var withHiddenMarked bytes.Buffer
	if err := p.WriteDoc(
		&withHiddenMarked,
		DocFormatMarkdown,
		WithBuiltinTemplate(DocTemplateMarkdownList),
		WithIncludeHidden(true),
		WithMarkHidden(true),
	); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	if !strings.Contains(withHiddenMarked.String(), "*(hidden)") {
		t.Fatalf("expected hidden markers when enabled, got:\n%s", withHiddenMarked.String())
	}
}

func TestWriteDocTemplateHasFullTagMetadata(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" description:"Mode" choice:"fast" choice:"safe" no-ini:"true" key-value-delimiter:":"`
		Cmd  struct {
			Flag bool `long:"flag" description:"Flag"`
		} `command:"run" description:"Run command" pass-after-non-option:"true" subcommands-optional:"true"`
	}

	p := NewNamedParser("meta-doc", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	tpl := strings.Join([]string{
		"{{ $g := index .Doc.Groups 0 }}",
		"{{ $o := index $g.Options 0 }}",
		"choice={{ index $o.Tags \"choice\" 0 }},",
		"noini={{ index $o.Tags \"no-ini\" 0 }},",
		"kvd={{ index $o.Tags \"key-value-delimiter\" 0 }},",
		"grpns={{ $g.Namespace }},",
		"cmd={{ (index .Doc.Commands 0).Name }},",
		"pass={{ (index .Doc.Commands 0).PassAfterNonOption }},",
		"subopt={{ (index .Doc.Commands 0).SubcommandsOptional }}",
	}, "")

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithTemplateString(tpl)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"choice=fast",
		"noini=true",
		"kvd=:",
		"cmd=run",
		"pass=true",
		"subopt=true",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in metadata output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocTemplateIncludesCommandAndArgGroups(t *testing.T) {
	var opts struct {
		Positional struct {
			Input  string `positional-arg-name:"input" arg-group:"Input" description:"Input file"`
			Output string `positional-arg-name:"output" arg-group:"Output" description:"Output file"`
		} `positional-args:"yes"`

		Add struct{} `command:"add" command-group:"Content" description:"Add item"`

		Config struct{} `command:"config" command-group:"Administration" description:"Configure app"`
	}

	p := NewNamedParser("group-doc", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	tpl := strings.Join([]string{
		"arg-group={{ (index .Doc.ArgGroups 0).Name }}:",
		"{{ (index (index .Doc.ArgGroups 0).Args 0).Name }},",
		"cmd-group={{ (index .Doc.CommandGroups 0).Name }}:",
		"{{ (index (index .Doc.CommandGroups 0).Commands 0).Name }}",
	}, "")

	var out bytes.Buffer
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithTemplateString(tpl)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	got := out.String()
	for _, needle := range []string{
		"arg-group=Input:input",
		"cmd-group=Content:add",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in grouped metadata output, got:\n%s", needle, got)
		}
	}
}

func TestWriteDocBuiltinTemplatesRenderCommandAndArgGroups(t *testing.T) {
	var opts struct {
		Positional struct {
			Input string `positional-arg-name:"input" arg-group:"Input" description:"Input file"`
		} `positional-args:"yes"`

		Add struct{} `command:"add" command-group:"Content" description:"Add item"`
	}

	p := NewNamedParser("group-doc", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	tests := []struct {
		name     string
		format   DocFormat
		template string
		want     []string
	}{
		{
			name:     "markdown list",
			format:   DocFormatMarkdown,
			template: DocTemplateMarkdownList,
			want:     []string{"**Content**", "### Input"},
		},
		{
			name:     "markdown table",
			format:   DocFormatMarkdown,
			template: DocTemplateMarkdownTable,
			want:     []string{"**Content**", "### Input"},
		},
		{
			name:     "markdown code",
			format:   DocFormatMarkdown,
			template: DocTemplateMarkdownCode,
			want:     []string{"**Content**", "[Input]"},
		},
		{
			name:     "html default",
			format:   DocFormatHTML,
			template: DocTemplateHTMLDefault,
			want:     []string{"<h3>Content</h3>"},
		},
		{
			name:     "html styled",
			format:   DocFormatHTML,
			template: DocTemplateHTMLStyled,
			want:     []string{"<h3>Content</h3>"},
		},
		{
			name:     "man default",
			format:   DocFormatMan,
			template: DocTemplateManDefault,
			want:     []string{"\\fBContent\\fP"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := p.WriteDoc(&out, test.format, WithBuiltinTemplate(test.template)); err != nil {
				t.Fatalf("unexpected write doc error: %v", err)
			}
			got := out.String()
			for _, want := range test.want {
				if !strings.Contains(got, want) {
					t.Fatalf("expected %q in grouped built-in doc output, got:\n%s", want, got)
				}
			}
		})
	}
}
