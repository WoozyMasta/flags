package flags

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestWriteVersionUsesOverrides(t *testing.T) {
	p := NewNamedParser("myapp", None)
	p.SetVersion("v9.9.9")
	p.SetVersionCommit("cafebabe")
	p.SetVersionTime(time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC))
	p.SetVersionURL("https://github.com/example/repo")

	var b strings.Builder
	p.WriteVersion(&b, VersionFieldsCore)
	out := b.String()

	if !strings.Contains(out, "version:  v9.9.9") {
		t.Fatalf("expected overridden version in output, got:\n%s", out)
	}
	if !strings.Contains(out, "commit:   cafebabe") {
		t.Fatalf("expected overridden commit in output, got:\n%s", out)
	}
	if !strings.Contains(out, "url:      https://github.com/example/repo") {
		t.Fatalf("expected overridden url in output, got:\n%s", out)
	}
}

func TestWriteVersionFieldMask(t *testing.T) {
	p := NewNamedParser("myapp", None)
	p.SetVersion("v1.0.0")
	p.SetVersionCommit("abc123")
	p.SetVersionURL("https://example.test/repo")

	var b strings.Builder
	p.WriteVersion(&b, VersionFieldVersion|VersionFieldCommit)
	out := b.String()

	if strings.Contains(out, "file:") {
		t.Fatalf("did not expect file field in masked output:\n%s", out)
	}
	if !strings.Contains(out, "version:  v1.0.0") {
		t.Fatalf("expected version field, got:\n%s", out)
	}
	if !strings.Contains(out, "commit:   abc123") {
		t.Fatalf("expected commit field, got:\n%s", out)
	}
}

func TestWriteVersionTargetField(t *testing.T) {
	p := NewNamedParser("myapp", None)
	p.SetVersionTarget("linux", "amd64")

	var b strings.Builder
	p.WriteVersion(&b, VersionFieldTarget)
	out := b.String()

	if !strings.Contains(out, "target:   linux/amd64") {
		t.Fatalf("expected target field in output, got:\n%s", out)
	}
}

func TestWriteVersionTargetFieldFallback(t *testing.T) {
	p := NewNamedParser("myapp", None)
	p.SetVersionTarget("", "")

	var b strings.Builder
	p.WriteVersion(&b, VersionFieldTarget)
	out := b.String()

	want := "target:   " + runtime.GOOS + "/" + runtime.GOARCH
	if !strings.Contains(out, want) {
		t.Fatalf("expected runtime target fallback %q, got:\n%s", want, out)
	}
}

func TestWriteVersionEndsWithANSIResetWhenColorEnabled(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	_ = os.Setenv("FORCE_COLOR", "1")

	p := NewNamedParser("myapp", ColorHelp)
	p.SetVersion("v1.2.3")
	p.SetHelpColorScheme(DefaultHelpColorScheme())

	var b strings.Builder
	p.WriteVersion(&b, VersionFieldVersion)
	out := b.String()

	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI colored output, got:\n%s", out)
	}
	if !strings.HasSuffix(out, ansiReset) {
		t.Fatalf("expected version output to end with ANSI reset, got:\n%s", out)
	}
}

func TestWriteVersionPadsLinesWithBaseBackground(t *testing.T) {
	p := NewNamedParser("myapp", ColorHelp)
	p.SetVersion("v1.2.3")
	p.SetHelpColorScheme(HelpColorScheme{
		BaseText:     HelpTextStyle{UseBG: true, BG: ColorBrightWhite},
		VersionLabel: HelpTextStyle{UseFG: true, FG: ColorBlue, Bold: true},
		VersionValue: HelpTextStyle{UseFG: true, FG: ColorBlack},
	})
	if err := p.SetHelpWidth(24); err != nil {
		t.Fatalf("unexpected set help width error: %v", err)
	}

	var b strings.Builder
	p.WriteVersion(&b, VersionFieldVersion)
	out := b.String()
	lines := strings.Split(strings.TrimSuffix(out, ansiReset), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected at least one version line, got empty output")
	}
	if got := textWidth(lines[0]); got != 24 {
		t.Fatalf("expected padded line width 24, got %d line=%q", got, lines[0])
	}
}
