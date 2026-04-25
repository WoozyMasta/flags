package flags

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

type benchCLIOptions struct {
	Verbose []bool            `short:"v" long:"verbose" description:"Enable verbose output"`
	Config  string            `long:"config" description:"Path to config file"`
	Timeout time.Duration     `long:"timeout" description:"Request timeout"`
	Retries int               `long:"retries" description:"Retry count"`
	Labels  map[string]string `long:"label" description:"Key/value labels"`
	Tags    []string          `long:"tag" description:"Free-form tags"`

	Add struct {
		Force bool `short:"f" long:"force" description:"Force add"`
		Pos   struct {
			Path string `positional-arg-name:"path" description:"Input path" required:"yes"`
		} `positional-args:"yes"`
	} `command:"add" description:"Add input paths"`
}

func benchmarkArgs() []string {
	return []string{
		"-vv",
		"--config=config.yaml",
		"--timeout=150ms",
		"--retries=3",
		"--label", "env:prod",
		"--label", "zone:us",
		"--tag", "alpha",
		"--tag", "beta",
		"add",
		"-f",
		"a.txt",
	}
}

func benchmarkParser() *Parser {
	var opts benchCLIOptions
	return NewParser(&opts, Default&^PrintErrors)
}

func BenchmarkParseArgsReusedParser(b *testing.B) {
	p := benchmarkParser()
	args := benchmarkArgs()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := p.ParseArgs(args); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

func BenchmarkParseArgsNewParser(b *testing.B) {
	args := benchmarkArgs()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p := benchmarkParser()
		if _, err := p.ParseArgs(args); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

func BenchmarkWriteHelp(b *testing.B) {
	p := benchmarkParser()
	if _, err := p.ParseArgs(benchmarkArgs()); err != nil {
		b.Fatalf("parse failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p.WriteHelp(io.Discard)
	}
}

func BenchmarkWriteManPage(b *testing.B) {
	p := benchmarkParser()
	if _, err := p.ParseArgs(benchmarkArgs()); err != nil {
		b.Fatalf("parse failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p.WriteManPage(io.Discard)
	}
}

const benchINI = `[Application Options]
verbose = true
verbose = true
config = config.yaml
timeout = 150ms
retries = 3
label = env:prod
label = zone:us
tag = alpha
tag = beta
`

const benchI18nCatalogJSON = `{
  "en": {
    "app.greeting": "Hello, {target}",
    "app.status": "Processed {count} items"
  },
  "ru": {
    "app.greeting": "Привет, {target}",
    "app.status": "Обработано элементов: {count}"
  },
  "eo": {
    "app.greeting": "Saluton, {target}",
    "app.status": "Traktis {count} erojn"
  }
}`

func BenchmarkIniParse(b *testing.B) {
	p := benchmarkParser()
	inip := NewIniParser(p)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := inip.Parse(strings.NewReader(benchINI)); err != nil {
			b.Fatalf("ini parse failed: %v", err)
		}
	}
}

func BenchmarkIniWrite(b *testing.B) {
	p := benchmarkParser()
	inip := NewIniParser(p)
	if err := inip.Parse(strings.NewReader(benchINI)); err != nil {
		b.Fatalf("ini parse failed: %v", err)
	}

	var out bytes.Buffer

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		out.Reset()
		inip.Write(&out, IniDefault|IniIncludeDefaults)
	}
}

func BenchmarkIniWriteExample(b *testing.B) {
	p := benchmarkParser()
	inip := NewIniParser(p)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		inip.WriteExampleWithOptions(io.Discard, IniExampleOptions{CommentWidth: 88})
	}
}

func BenchmarkWriteCompletionBash(b *testing.B) {
	p := benchmarkParser()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.WriteNamedCompletion(io.Discard, CompletionShellBash, "bench"); err != nil {
			b.Fatalf("completion render failed: %v", err)
		}
	}
}

func BenchmarkWriteVersion(b *testing.B) {
	p := NewNamedParser("bench", VersionFlag)
	p.SetVersion("v1.2.3")
	p.SetVersionCommit("abcdef0")
	p.SetVersionURL("https://example.test/bench")

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p.WriteVersion(io.Discard, VersionFieldsCore)
	}
}

func BenchmarkLocalizerLocalize(b *testing.B) {
	catalog, err := NewJSONCatalog([]byte(benchI18nCatalogJSON))
	if err != nil {
		b.Fatalf("catalog load failed: %v", err)
	}

	localizer := NewLocalizer(I18nConfig{
		Locale:          "ru-RU",
		FallbackLocales: []string{"en"},
		UserCatalog:     catalog,
	})
	data := map[string]string{"target": "мир"}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = localizer.Localize("app.greeting", "Hello, {target}", data)
	}
}

func BenchmarkNewJSONCatalog(b *testing.B) {
	data := []byte(benchI18nCatalogJSON)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := NewJSONCatalog(data); err != nil {
			b.Fatalf("catalog load failed: %v", err)
		}
	}
}

func BenchmarkWriteHelpI18n(b *testing.B) {
	p := benchmarkParser()
	p.SetI18n(I18nConfig{Locale: "ru"})

	if _, err := p.ParseArgs(benchmarkArgs()); err != nil {
		b.Fatalf("parse failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p.WriteHelp(io.Discard)
	}
}

func BenchmarkWriteDocMarkdownI18n(b *testing.B) {
	p := benchmarkParser()
	p.SetI18n(I18nConfig{Locale: "ru"})

	if _, err := p.ParseArgs(benchmarkArgs()); err != nil {
		b.Fatalf("parse failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.WriteDoc(io.Discard, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownList)); err != nil {
			b.Fatalf("doc render failed: %v", err)
		}
	}
}

func BenchmarkCheckMergedCatalogCoverageBuiltin(b *testing.B) {
	cfg := I18nCoverageConfig{
		BaseLocale:        "en",
		CheckPlaceholders: true,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := CheckMergedCatalogCoverage(cfg); err != nil {
			b.Fatalf("coverage check failed: %v", err)
		}
	}
}
