package flags

import (
	"bytes"
	"strings"
	"testing"
)

func TestGrayHelpColorScheme(t *testing.T) {
	s := GrayHelpColorScheme()
	base := DefaultHelpColorScheme()

	for _, role := range []HelpTextStyle{
		s.LongDescription,
		s.OptionShort,
		s.OptionLong,
		s.OptionValueName,
		s.OptionPunctuation,
		s.OptionEnv,
		s.OptionDefault,
		s.OptionChoices,
		s.ArgumentName,
		s.CommandName,
		s.CommandAliases,
		s.VersionLabel,
	} {
		if !role.UseFG || role.FG != ColorBrightBlack {
			t.Fatalf("expected gray role style, got %+v", role)
		}
	}

	for _, tc := range []struct {
		name string
		got  HelpTextStyle
		want HelpTextStyle
	}{
		{name: "BaseText", got: s.BaseText, want: base.BaseText},
		{name: "SubcommandOptionsHeader", got: s.SubcommandOptionsHeader, want: HelpTextStyle{}},
		{name: "OptionDesc", got: s.OptionDesc, want: base.OptionDesc},
		{name: "UsageHeader", got: s.UsageHeader, want: HelpTextStyle{}},
		{name: "UsageText", got: s.UsageText, want: HelpTextStyle{}},
		{name: "CommandSectionHeader", got: s.CommandSectionHeader, want: HelpTextStyle{}},
		{name: "CommandGroupHeader", got: s.CommandGroupHeader, want: HelpTextStyle{}},
		{name: "CommandDesc", got: s.CommandDesc, want: HelpTextStyle{}},
		{name: "ArgumentsHeader", got: s.ArgumentsHeader, want: HelpTextStyle{}},
		{name: "ArgumentDesc", got: s.ArgumentDesc, want: base.ArgumentDesc},
		{name: "VersionValue", got: s.VersionValue, want: HelpTextStyle{}},
		{name: "GroupHeader", got: s.GroupHeader, want: HelpTextStyle{}},
	} {
		if tc.got != tc.want {
			t.Fatalf("expected %s to match default style, got %+v want %+v", tc.name, tc.got, tc.want)
		}
	}
}

func TestGrayErrorColorSchemeAlias(t *testing.T) {
	got := GrayErrorColorScheme()
	want := DefaultErrorColorScheme()

	if got != want {
		t.Fatalf("expected GrayErrorColorScheme to alias DefaultErrorColorScheme, got %+v want %+v", got, want)
	}
}

func TestGrayHelpColorSchemeOutputTargetsSelectedRoles(t *testing.T) {
	var opts struct {
		Name string   `short:"n" long:"name" value-name:"NAME" env:"GRAY_HELP_NAME" default:"qa" description:"Name value"`
		Tags []string `long:"tag" description:"Tag value"`
		Run  struct{} `command:"run" alias:"r" command-group:"Runner" description:"Run task"`
	}

	p := NewNamedParser("gray-help", ColorHelp|ShowRepeatableInHelp)
	p.LongDescription = "Gray help parser long description"
	p.SetHelpColorScheme(GrayHelpColorScheme())

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	p.WriteHelp(&out)
	got := out.String()
	format := p.optionRenderFormat()

	if !strings.Contains(got, "\x1b[90m") {
		t.Fatalf("expected gray ANSI sequences in output, got:\n%s", got)
	}

	wantLong := applyHelpTextStyle(defaultLongOptDelimiter+"name", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionLong))
	if !strings.Contains(got, wantLong) {
		t.Fatalf("expected gray long option token %q, got:\n%s", wantLong, got)
	}
	wantComma := applyHelpTextStyle(",", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionPunctuation))
	if !strings.Contains(got, wantComma) {
		t.Fatalf("expected gray comma punctuation %q, got:\n%s", wantComma, got)
	}
	wantNameDelimiter := applyHelpTextStyle(string(format.nameDelimiter), mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionPunctuation))
	if !strings.Contains(got, wantNameDelimiter) {
		t.Fatalf("expected gray delimiter punctuation %q, got:\n%s", wantNameDelimiter, got)
	}
	wantValueName := applyHelpTextStyle("NAME", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionValueName))
	if !strings.Contains(got, wantValueName) {
		t.Fatalf("expected gray value-name token %q, got:\n%s", wantValueName, got)
	}

	wantCommand := applyHelpTextStyle("run", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.CommandName))
	if !strings.Contains(got, wantCommand) {
		t.Fatalf("expected gray command name %q, got:\n%s", wantCommand, got)
	}
	if !strings.Contains(got, "Available commands:") {
		t.Fatalf("expected uncolored command section header, got:\n%s", got)
	}
	if !strings.Contains(got, "  Runner:") {
		t.Fatalf("expected uncolored command group header, got:\n%s", got)
	}

	wantLongDesc := applyHelpTextStyle("Gray help parser long description", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.LongDescription))
	if !strings.Contains(got, wantLongDesc) {
		t.Fatalf("expected gray long description %q, got:\n%s", wantLongDesc, got)
	}

	wantDefault := applyHelpTextStyle("default: qa", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionDefault))
	if !strings.Contains(got, wantDefault) {
		t.Fatalf("expected gray default helper label %q, got:\n%s", wantDefault, got)
	}

	envFrag := format.envPrefix + "GRAY_HELP_NAME" + format.envSuffix
	wantEnv := applyHelpTextStyle(envFrag, mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionEnv))
	if !strings.Contains(got, wantEnv) {
		t.Fatalf("expected gray env helper label %q, got:\n%s", wantEnv, got)
	}

	wantRepeatable := applyHelpTextStyle("repeatable", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionChoices))
	if !strings.Contains(got, wantRepeatable) {
		t.Fatalf("expected gray repeatable helper label %q, got:\n%s", wantRepeatable, got)
	}

	if !strings.HasSuffix(got, ansiReset) {
		t.Fatalf("expected help output to end with ANSI reset, got:\n%s", got)
	}
}

func TestHighContrastHelpColorSchemeUnderlineOnlyHeaders(t *testing.T) {
	s := HighContrastHelpColorScheme()

	headerRoles := []HelpTextStyle{
		s.UsageHeader,
		s.SubcommandOptionsHeader,
		s.CommandSectionHeader,
		s.CommandGroupHeader,
		s.ArgumentsHeader,
		s.GroupHeader,
	}
	for _, role := range headerRoles {
		if !role.Underline {
			t.Fatalf("expected header role to be underlined, got %+v", role)
		}
	}

	nonHeaderRoles := []HelpTextStyle{
		s.BaseText,
		s.LongDescription,
		s.VersionLabel,
		s.VersionValue,
		s.OptionShort,
		s.OptionLong,
		s.OptionValueName,
		s.OptionPunctuation,
		s.OptionDesc,
		s.OptionEnv,
		s.OptionDefault,
		s.OptionChoices,
		s.UsageText,
		s.CommandName,
		s.CommandDesc,
		s.CommandAliases,
		s.ArgumentName,
		s.ArgumentDesc,
	}
	for _, role := range nonHeaderRoles {
		if role.Underline {
			t.Fatalf("did not expect underline on non-header role, got %+v", role)
		}
	}
}
