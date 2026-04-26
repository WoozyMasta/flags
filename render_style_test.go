package flags

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHelpRenderStyleSetters(t *testing.T) {
	var opts struct {
		Value string `long:"value" env:"APP_VALUE" description:"Value option"`
	}

	p := NewNamedParser("render-style", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	p.SetHelpFlagRenderStyle(RenderStyleWindows)
	p.SetHelpEnvRenderStyle(RenderStyleWindows)

	var windowsOut bytes.Buffer
	p.WriteHelp(&windowsOut)
	gotWindows := windowsOut.String()

	if !strings.Contains(gotWindows, "/value") {
		t.Fatalf("expected windows flag style in help, got:\n%s", gotWindows)
	}

	if !strings.Contains(gotWindows, "%APP_VALUE%") {
		t.Fatalf("expected windows env style in help, got:\n%s", gotWindows)
	}

	p.SetHelpFlagRenderStyle(RenderStylePOSIX)
	p.SetHelpEnvRenderStyle(RenderStylePOSIX)

	var posixOut bytes.Buffer
	p.WriteHelp(&posixOut)
	gotPOSIX := posixOut.String()

	if !strings.Contains(gotPOSIX, "--value") {
		t.Fatalf("expected posix flag style in help, got:\n%s", gotPOSIX)
	}

	if !strings.Contains(gotPOSIX, "$APP_VALUE") {
		t.Fatalf("expected posix env style in help, got:\n%s", gotPOSIX)
	}
}

func TestHelpRenderStyleDetectFromShellOption(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		Value string `long:"value" env:"APP_VALUE" description:"Value option"`
	}

	p := NewNamedParser("render-style", DetectShellFlagStyle|DetectShellEnvStyle)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_ = os.Setenv("GO_FLAGS_SHELL", "pwsh")
	var pwshOut bytes.Buffer
	p.WriteHelp(&pwshOut)

	gotPwsh := pwshOut.String()
	if !strings.Contains(gotPwsh, "/value") || !strings.Contains(gotPwsh, "%APP_VALUE%") {
		t.Fatalf("expected windows-style output for pwsh shell, got:\n%s", gotPwsh)
	}

	_ = os.Setenv("GO_FLAGS_SHELL", "bash")
	var bashOut bytes.Buffer
	p.WriteHelp(&bashOut)

	gotBash := bashOut.String()
	if !strings.Contains(gotBash, "--value") || !strings.Contains(gotBash, "$APP_VALUE") {
		t.Fatalf("expected posix-style output for bash shell, got:\n%s", gotBash)
	}
}

func TestHelpRenderStyleDetectWindowsPrefersWindowsOverShellVar(t *testing.T) {
	if !isWindowsRuntime() {
		t.Skip("windows-only behavior")
	}

	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		Value string `long:"value" env:"APP_VALUE" description:"Value option"`
	}

	p := NewNamedParser("render-style", DetectShellFlagStyle|DetectShellEnvStyle)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_ = os.Setenv("SHELL", "/usr/bin/bash")
	_ = os.Unsetenv("GO_FLAGS_SHELL")
	_ = os.Unsetenv("MSYSTEM")
	_ = os.Unsetenv("OSTYPE")

	if shellKind(detectParentShellName()) == RenderStylePOSIX {
		t.Skip("session has POSIX-like parent shell on Windows; windows-preference assertion not applicable")
	}

	var out bytes.Buffer
	p.WriteHelp(&out)
	got := out.String()

	if !strings.Contains(got, "/value") || !strings.Contains(got, "%APP_VALUE%") {
		t.Fatalf("expected windows-style output in powershell/cmd-like windows environment, got:\n%s", got)
	}
}

func TestHelpFlagAliasesFollowDetectedPOSIXStyleOnWindows(t *testing.T) {
	if !isWindowsRuntime() {
		t.Skip("windows-only behavior")
	}

	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	_ = os.Setenv("GO_FLAGS_SHELL", "bash")

	p := NewNamedParser("render-style", HelpFlag|DetectShellFlagStyle)
	_, err := p.ParseArgs([]string{"--help"})
	if err == nil {
		t.Fatal("expected ErrHelp")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}

	helpText := flagsErr.Message
	if strings.Contains(helpText, "-?") {
		t.Fatalf("expected POSIX help aliases only, got:\n%s", helpText)
	}
	if !strings.Contains(helpText, "-h, --help") {
		t.Fatalf("expected POSIX help alias, got:\n%s", helpText)
	}
}

func TestDocRenderStyleSetters(t *testing.T) {
	var opts struct {
		Value string `long:"value" env:"APP_VALUE" description:"Value option"`
	}

	p := NewNamedParser("render-doc-style", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	p.SetHelpFlagRenderStyle(RenderStyleWindows)
	p.SetHelpEnvRenderStyle(RenderStyleWindows)

	var winDoc bytes.Buffer
	if err := p.WriteDoc(&winDoc, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownList)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	gotWinDoc := winDoc.String()
	if !strings.Contains(gotWinDoc, "/value") || !strings.Contains(gotWinDoc, "%APP_VALUE%") {
		t.Fatalf("expected windows-style markers in rendered doc, got:\n%s", gotWinDoc)
	}

	p.SetHelpFlagRenderStyle(RenderStylePOSIX)
	p.SetHelpEnvRenderStyle(RenderStylePOSIX)

	var posixDoc bytes.Buffer
	if err := p.WriteDoc(&posixDoc, DocFormatMarkdown, WithBuiltinTemplate(DocTemplateMarkdownList)); err != nil {
		t.Fatalf("unexpected write doc error: %v", err)
	}

	gotPOSIXDoc := posixDoc.String()
	if !strings.Contains(gotPOSIXDoc, "--value") || !strings.Contains(gotPOSIXDoc, "$APP_VALUE") {
		t.Fatalf("expected posix-style markers in rendered doc, got:\n%s", gotPOSIXDoc)
	}
}
