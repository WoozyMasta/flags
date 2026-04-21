// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

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

	if !strings.Contains(out.String(), "## OPTIONS") {
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
		"|Option|Description|Default|Env|Required|",
		"|---|---|---|---|---|",
		defaultLongOptDelimiter + "verbose",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in markdown output, got:\n%s", needle, got)
		}
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

	for _, needle := range []string{"/hidden", "### internal"} {
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
