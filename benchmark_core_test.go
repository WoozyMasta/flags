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
