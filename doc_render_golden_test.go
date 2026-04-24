package flags

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteDocBuiltinTemplatesGolden(t *testing.T) {
	if runtime.GOOS == "windows" && defaultLongOptDelimiter != "--" {
		t.Skip("run with -tags forceposix to validate unix doc snapshots on windows")
	}
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	os.Setenv("SOURCE_DATE_EPOCH", "1700000000")

	var opts struct {
		Verbose bool              `short:"v" long:"verbose" description:"Enable verbose output"`
		Config  string            `long:"config" description:"Path to config" default:"config.yaml" env:"APP_CONFIG" required:"true"`
		Mode    string            `long:"mode" description:"Execution mode" choice:"fast" choice:"safe" default:"fast"`
		Tags    []string          `long:"tag" description:"Tag filter" default:"api"`
		Headers map[string]string `long:"header" description:"HTTP headers" default:"x-env:dev" key-value-delimiter:":"`
		Secret  string            `long:"secret" description:"Secret value" hidden:"true" default-mask:"***"`
		Level   string            `long:"level" description:"Log level" optional:"yes" optional-value:"info"`
		Args    struct {
			Input string `positional-arg-name:"input" description:"Input resource"`
		} `positional-args:"yes"`

		DB struct {
			Host string `long:"host" description:"Database host" default:"127.0.0.1" env:"HOST"`
			Port int    `long:"port" description:"Database port" default:"5432" env:"PORT"`
		} `group:"Database Options" namespace:"db" env-namespace:"DB"`
		HiddenGroup struct {
			Debug string `long:"debug" description:"Hidden debug selector"`
		} `group:"Internal" hidden:"true"`

		Run struct {
			Force bool `long:"force" description:"Force execution"`
			Plan  bool `long:"plan" description:"Show execution plan only"`
			Args  struct {
				Target string `positional-arg-name:"target" description:"Deployment target"`
			} `positional-args:"yes"`
		} `command:"run" description:"Run command" long-description:"Execute deployment workflow."`
		Status struct {
			JSON bool `long:"json" description:"JSON output"`
		} `command:"status" description:"Show status" long-description:"Read and print current status."`
		HiddenCmd struct {
			Noisy bool `long:"noisy" description:"Noise"`
		} `command:"internal" description:"Internal command" hidden:"true"`
	}

	p := NewNamedParser("golden-doc", None)
	p.ShortDescription = "Golden doc parser"
	p.LongDescription = "Long description for golden tests.\nIncludes options, groups and commands."
	p.SetEnvPrefix("MY_APP")

	if _, err := p.AddGroup("Application Options", "Main options", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	outputs := map[string]string{}
	for _, tc := range []struct {
		name   string
		format DocFormat
		opts   []DocOption
	}{
		{name: "markdown-list", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownList)}},
		{
			name:   "markdown-list-hidden",
			format: DocFormatMarkdown,
			opts: []DocOption{
				WithBuiltinTemplate(DocTemplateMarkdownList),
				WithIncludeHidden(true),
				WithMarkHidden(true),
			},
		},
		{name: "markdown-table", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownTable)}},
		{name: "markdown-code", format: DocFormatMarkdown, opts: []DocOption{WithBuiltinTemplate(DocTemplateMarkdownCode)}},
		{name: "html-default", format: DocFormatHTML, opts: []DocOption{WithBuiltinTemplate(DocTemplateHTMLDefault)}},
		{name: "html-styled", format: DocFormatHTML, opts: []DocOption{WithBuiltinTemplate(DocTemplateHTMLStyled)}},
		{name: "man-default", format: DocFormatMan, opts: []DocOption{WithBuiltinTemplate(DocTemplateManDefault)}},
	} {
		var out bytes.Buffer
		if err := p.WriteDoc(&out, tc.format, tc.opts...); err != nil {
			t.Fatalf("unexpected write doc error (%s): %v", tc.name, err)
		}

		rendered := out.String()
		if tc.format == DocFormatMarkdown {
			rendered = normalizeMarkdown(rendered)
		} else {
			rendered = normalizeNewlines(rendered)
		}
		outputs[tc.name] = rendered
	}

	suffix := "unix"

	update := os.Getenv("UPDATE_DOC_EXAMPLES") == "1"
	for name, got := range outputs {
		ext := ".md"
		if strings.HasPrefix(name, "html-") {
			ext = ".html"
		} else if strings.HasPrefix(name, "man-") {
			ext = ".1"
		}
		fileSuffix := suffix
		if strings.HasPrefix(name, "man-") {
			fileSuffix = "posix"
		}
		path := filepath.Join("examples", "doc-rendered", name+"."+fileSuffix+ext)
		if update {
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatalf("failed to update golden %s: %v", path, err)
			}
		}

		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read golden %s: %v (set UPDATE_DOC_EXAMPLES=1 to generate)", path, err)
		}

		wantContent := string(want)
		if strings.HasSuffix(ext, ".md") {
			wantContent = normalizeMarkdown(wantContent)
		} else {
			wantContent = normalizeNewlines(wantContent)
		}

		if got != wantContent {
			assertDiff(t, got, wantContent, "doc golden "+name)
		}
	}
}

func normalizeNewlines(in string) string {
	in = strings.ReplaceAll(in, "\r\n", "\n")
	in = strings.TrimRight(in, "\n")
	return in + "\n"
}

func normalizeMarkdown(in string) string {
	in = strings.ReplaceAll(in, "\r\n", "\n")
	lines := strings.Split(in, "\n")
	out := make([]string, 0, len(lines))
	prevBlank := false

	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		blank := strings.TrimSpace(line) == ""
		if blank {
			line = ""
		}
		if blank && prevBlank {
			continue
		}
		out = append(out, line)
		prevBlank = blank
	}

	result := strings.Join(out, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}
