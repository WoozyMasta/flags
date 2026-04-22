package flags

import (
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
