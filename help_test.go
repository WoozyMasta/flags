package flags

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type helpOptions struct {
	Verbose          []bool       `short:"v" long:"verbose" description:"Show verbose debug information" ini-name:"verbose"`
	Call             func(string) `short:"c" description:"Call phone number" ini-name:"call"`
	PtrSlice         []*string    `long:"ptrslice" description:"A slice of pointers to string"`
	EmptyDescription bool         `long:"empty-description"`

	Default           string            `long:"default" default:"Some\nvalue" description:"Test default value"`
	DefaultArray      []string          `long:"default-array" default:"Some value" default:"Other\tvalue" description:"Test default array value"`
	DefaultMap        map[string]string `long:"default-map" default:"some:value" default:"another:value" description:"Testdefault map value"`
	EnvDefault1       string            `long:"env-default1" default:"Some value" env:"ENV_DEFAULT" description:"Test env-default1 value"`
	EnvDefault2       string            `long:"env-default2" env:"ENV_DEFAULT" description:"Test env-default2 value"`
	OptionWithArgName string            `long:"opt-with-arg-name" value-name:"something" description:"Option with named argument"`
	OptionWithChoices string            `long:"opt-with-choices" value-name:"choice" choice:"dog" choice:"cat" description:"Option with choices"`
	Hidden            string            `long:"hidden" description:"Hidden option" hidden:"yes"`

	HiddenOptionWithVeryLongName bool `long:"hidden-option-very-long-name" hidden:"yes"`

	OnlyIni string `ini-name:"only-ini" description:"Option only available in ini"`

	Other struct {
		StringSlice []string       `short:"s" default:"some" default:"value" description:"A slice of strings"`
		IntMap      map[string]int `long:"intmap" default:"a:1" description:"A map from string to int" ini-name:"int-map"`
	} `group:"Other Options"`

	HiddenGroup struct {
		InsideHiddenGroup string `long:"inside-hidden-group" description:"Inside hidden group"`
		Padder            bool   `long:"hidden-group-option-long-name"`
	} `group:"Hidden group" hidden:"yes"`

	GroupWithOnlyHiddenOptions struct {
		SecretFlag bool `long:"secret" description:"Hidden flag in a non-hidden group" hidden:"yes"`
	} `group:"Non-hidden group with only hidden options"`

	Group struct {
		Opt                  string `long:"opt" description:"This is a subgroup option"`
		HiddenInsideGroup    string `long:"hidden-inside-group" description:"Hidden inside group" hidden:"yes"`
		NotHiddenInsideGroup string `long:"not-hidden-inside-group" description:"Not hidden inside group" hidden:"false"`

		Group struct {
			Opt string `long:"opt" description:"This is a subsubgroup option"`
		} `group:"Subsubgroup" namespace:"sap"`
	} `group:"Subgroup" namespace:"sip"`

	Bommand struct {
		Hidden bool `long:"hidden" description:"A hidden option" hidden:"yes"`
	} `command:"bommand" description:"A command with only hidden options"`

	Command struct {
		ExtraVerbose []bool `long:"extra-verbose" description:"Use for extra verbosity"`
	} `command:"command" alias:"cm" alias:"cmd" description:"A command"`

	HiddenCommand struct {
		ExtraVerbose []bool `long:"extra-verbose" description:"Use for extra verbosity"`
	} `command:"hidden-command" description:"A hidden command" hidden:"yes"`

	ParentCommand struct {
		Opt        string `long:"opt" description:"This is a parent command option"`
		SubCommand struct {
			Opt string `long:"opt" description:"This is a sub command option"`
		} `command:"sub" description:"A sub command"`
	} `command:"parent" description:"A parent command"`

	Args struct {
		Filename     string  `positional-arg-name:"filename" description:"A filename with a long description to trigger line wrapping"`
		Number       int     `positional-arg-name:"num" description:"A number"`
		HiddenInHelp float32 `positional-arg-name:"hidden-in-help" required:"yes"`
	} `positional-args:"yes"`
}

func TestHelp(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	os.Setenv("ENV_DEFAULT", "env-def")

	var opts helpOptions
	p := NewNamedParser("TestHelp", HelpFlag)
	p.AddGroup("Application Options", "The application options", &opts)

	_, err := p.ParseArgs([]string{"--help"})

	if err == nil {
		t.Fatalf("Expected help error")
	}

	if e, ok := err.(*Error); !ok {
		t.Fatalf("Expected flags.Error, but got %T", err)
	} else {
		if e.Type != ErrHelp {
			t.Errorf("Expected flags.ErrHelp type, but got %s", e.Type)
		}

		needles := []string{
			"Usage:",
			"Application Options:",
			"Other Options:",
			"Subgroup:",
			"Subsubgroup:",
			"Help Options:",
			"Arguments:",
			"Available commands:",
			"A command with only hidden options",
			"A command (aliases: cm, cmd)",
			"A parent command",
			"Test default value",
			"(default: \"Some\\nvalue\")",
			"Test env-default2 value",
			"Option with choices",
			"A filename with a long description",
		}

		if runtime.GOOS == "windows" {
			needles = append(needles, "/v, /verbose", "[%ENV_DEFAULT%]")
		} else {
			needles = append(needles, "-v, --verbose", "[$ENV_DEFAULT]")
		}

		for _, needle := range needles {
			if !strings.Contains(e.Message, needle) {
				t.Fatalf("expected help message to contain %q, got:\n%s", needle, e.Message)
			}
		}
	}
}

func TestHelpShowsRepeatableHints(t *testing.T) {
	var opts struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose mode"`
		Name    string `short:"n" long:"name" description:"User name"`
	}

	p := NewNamedParser("TestHelpHints", HelpFlag)
	p.Options |= ShowRepeatableInHelp

	_, err := p.AddGroup("Application Options", "", &opts)
	if err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err = p.ParseArgs([]string{"--help"})
	if err == nil {
		t.Fatalf("expected help error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}

	if !strings.Contains(flagsErr.Message, "repeatable") {
		t.Fatalf("expected help to contain repeatable marker, got:\n%s", flagsErr.Message)
	}
}

func TestHelpHideEnvInHelp(t *testing.T) {
	var opts struct {
		Config string `long:"config" env:"APP_CONFIG" description:"Path to config"`
	}

	p := NewNamedParser("TestHelpHideEnv", HelpFlag|HideEnvInHelp)
	_, err := p.AddGroup("Application Options", "", &opts)
	if err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err = p.ParseArgs([]string{"--help"})
	if err == nil {
		t.Fatalf("expected help error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}

	if strings.Contains(flagsErr.Message, "APP_CONFIG") {
		t.Fatalf("expected help without env placeholder, got:\n%s", flagsErr.Message)
	}
}

func TestHelpShowsCommandAliasesWithoutDescription(t *testing.T) {
	var opts struct {
		Cmd struct{} `command:"run" alias:"r"`
	}

	p := NewNamedParser("TestCommandAlias", HelpFlag)
	p.Options |= ShowCommandAliases

	_, err := p.AddGroup("Application Options", "", &opts)
	if err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err = p.ParseArgs([]string{"--help"})
	if err == nil {
		t.Fatalf("expected help error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}

	if !strings.Contains(flagsErr.Message, "run (aliases: r)") {
		t.Fatalf("expected command alias in help, got:\n%s", flagsErr.Message)
	}
}

func TestHelpColorEnabled(t *testing.T) {
	var opts struct {
		Name string   `long:"name" description:"Name value"`
		Tags []string `long:"tag" description:"Tag value"`
		Mode string   `long:"mode" choice:"fast" choice:"safe" description:"Mode"`
		Run  struct{} `command:"run" alias:"r" description:"Run task"`
		Args struct {
			Input string `positional-arg-name:"input" description:"Input resource"`
		} `positional-args:"yes"`
	}

	p := NewNamedParser("ColorHelp", ColorHelp|ShowRepeatableInHelp)
	p.LongDescription = "Color help parser long description"
	p.SetHelpColorScheme(HelpColorScheme{
		BaseText:                HelpTextStyle{},
		LongDescription:         HelpTextStyle{UseFG: true, FG: ColorBrightWhite},
		SubcommandOptionsHeader: HelpTextStyle{UseFG: true, FG: ColorBrightMagenta},
		OptionLong:              HelpTextStyle{UseFG: true, FG: ColorRed, Bold: true},
		OptionChoices:           HelpTextStyle{UseFG: true, FG: ColorGreen, Bold: true},
		OptionDefault:           HelpTextStyle{UseFG: true, FG: ColorMagenta},
		OptionEnv:               HelpTextStyle{UseFG: true, FG: ColorBlue},
		GroupHeader:             HelpTextStyle{UseFG: true, FG: ColorYellow, Bold: true},
		CommandsHeader:          HelpTextStyle{UseFG: true, FG: ColorCyan, Bold: true},
		CommandName:             HelpTextStyle{UseFG: true, FG: ColorCyan, Bold: true},
		CommandAliases:          HelpTextStyle{UseFG: true, FG: ColorBrightMagenta},
		UsageHeader:             HelpTextStyle{UseFG: true, FG: ColorBrightYellow, Bold: true},
		UsageText:               HelpTextStyle{UseFG: true, FG: ColorBrightBlue, Bold: true},
		OptionDesc:              HelpTextStyle{},
		ArgumentsHeader:         HelpTextStyle{UseFG: true, FG: ColorYellow, Bold: true},
		ArgumentName:            HelpTextStyle{UseFG: true, FG: ColorYellow, Bold: true},
		ArgumentDesc:            HelpTextStyle{UseFG: true, FG: ColorBlue},
	})

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	p.WriteHelp(&out)
	got := out.String()

	wantLong := applyHelpTextStyle(defaultLongOptDelimiter+"name", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionLong))
	if !strings.Contains(got, wantLong) {
		t.Fatalf("expected colored long flag token %q, got:\n%s", wantLong, got)
	}

	wantChoices := applyHelpTextStyle("[fast|safe]", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionChoices))
	if !strings.Contains(got, wantChoices) {
		t.Fatalf("expected colored choices token %q, got:\n%s", wantChoices, got)
	}

	wantRepeatable := applyHelpTextStyle("repeatable", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.OptionChoices))
	if !strings.Contains(got, wantRepeatable) {
		t.Fatalf("expected colored repeatable marker %q, got:\n%s", wantRepeatable, got)
	}

	wantArgKey := applyHelpTextStyle("  input:", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.ArgumentName))
	if !strings.Contains(got, wantArgKey) {
		t.Fatalf("expected colored argument key %q, got:\n%s", wantArgKey, got)
	}

	wantArgDesc := applyHelpTextStyle("Input resource", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.ArgumentDesc))
	if !strings.Contains(got, wantArgDesc) {
		t.Fatalf("expected colored argument description %q, got:\n%s", wantArgDesc, got)
	}

	wantUsageString := applyHelpTextStyle("ColorHelp", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.UsageText))
	if !strings.Contains(got, wantUsageString) {
		t.Fatalf("expected colored usage string %q, got:\n%s", wantUsageString, got)
	}

	wantLongDesc := applyHelpTextStyle("Color help parser long description", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.LongDescription))
	if !strings.Contains(got, wantLongDesc) {
		t.Fatalf("expected colored long description %q, got:\n%s", wantLongDesc, got)
	}

	wantAliases := applyHelpTextStyle(" (aliases: r)", mergeHelpStyle(p.helpColorScheme.BaseText, p.helpColorScheme.CommandAliases))
	if !strings.Contains(got, wantAliases) {
		t.Fatalf("expected colored command aliases %q, got:\n%s", wantAliases, got)
	}
}

func TestHelpColorDisabled(t *testing.T) {
	var opts struct {
		Name string `long:"name" description:"Name value"`
	}

	p := NewNamedParser("ColorHelp", None)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	var out bytes.Buffer
	p.WriteHelp(&out)

	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("did not expect ANSI color sequences when ColorHelp is disabled:\n%s", out.String())
	}
}

func TestMan(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	os.Setenv("ENV_DEFAULT", "env-def")

	var opts helpOptions
	p := NewNamedParser("TestMan", HelpFlag)
	p.ShortDescription = "Test manpage generation"
	p.LongDescription = "This is a somewhat `longer' description of what this does.\nWith multiple lines."
	p.AddGroup("Application Options", "The application options", &opts)

	for _, cmd := range p.Commands() {
		cmd.LongDescription = fmt.Sprintf("Longer `%s' description", cmd.Name)
	}

	var buf bytes.Buffer
	p.WriteManPage(&buf)

	got := buf.String()

	tt := time.Now()
	source_date_epoch := os.Getenv("SOURCE_DATE_EPOCH")
	if source_date_epoch != "" {
		sde, err := strconv.ParseInt(source_date_epoch, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid SOURCE_DATE_EPOCH: %s", err))
		}
		tt = time.Unix(sde, 0)
	}

	var envDefaultName string

	if runtime.GOOS == "windows" {
		envDefaultName = "%ENV_DEFAULT%"
	} else {
		envDefaultName = "$ENV_DEFAULT"
	}

	expectedHeader := fmt.Sprintf(`.TH TestMan 1 "%s"`, tt.Format("2 January 2006"))
	for _, needle := range []string{
		expectedHeader,
		`.SH NAME`,
		`TestMan \- Test manpage generation`,
		`.SH SYNOPSIS`,
		`\fBTestMan\fP [OPTIONS]`,
		`.SH DESCRIPTION`,
		`This is a somewhat \fBlonger\fP description of what this does.`,
		`With multiple lines.`,
		`.SH OPTIONS`,
		`.SS Application Options`,
		`The application options`,
		`\fB\fB\-\-env-default2\fR <default: \fI` + envDefaultName + `\fR>\fP`,
		`.SH COMMANDS`,
		`.SS command`,
		`\fBAliases\fP: cm, cmd`,
		`\fBUsage\fP: TestMan [OPTIONS] command [command-OPTIONS]`,
		`.SS parent sub`,
		`This is a sub command option`,
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected %q in man output, got:\n%s", needle, got)
		}
	}
}

type helpCommandNoOptions struct {
	Command struct {
	} `command:"command" description:"A command"`
}

func TestHelpCommand(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	os.Setenv("ENV_DEFAULT", "env-def")

	var opts helpCommandNoOptions
	p := NewNamedParser("TestHelpCommand", HelpFlag)
	p.AddGroup("Application Options", "The application options", &opts)

	_, err := p.ParseArgs([]string{"command", "--help"})

	if err == nil {
		t.Fatalf("Expected help error")
	}

	if e, ok := err.(*Error); !ok {
		t.Fatalf("Expected flags.Error, but got %T", err)
	} else {
		if e.Type != ErrHelp {
			t.Errorf("Expected flags.ErrHelp type, but got %s", e.Type)
		}

		var expected string

		if runtime.GOOS == "windows" {
			expected = `Usage:
  TestHelpCommand [OPTIONS] command

Help Options:
  /?              Show this help message
  /h, /help       Show this help message
`
		} else {
			expected = `Usage:
  TestHelpCommand [OPTIONS] command

Help Options:
  -h, --help      Show this help message
`
		}

		assertDiff(t, e.Message, expected, "help message")
	}
}

func TestHiddenCommandNoBuiltinHelp(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	os.Setenv("ENV_DEFAULT", "env-def")

	// no auto added help group
	p := NewNamedParser("TestHelpCommand", 0)
	// and no usage information either
	p.Usage = ""

	// add custom help group which is not listed in --help output
	var help struct {
		ShowHelp func() error `short:"h" long:"help"`
	}
	help.ShowHelp = func() error {
		return &Error{Type: ErrHelp}
	}
	hlpgrp, err := p.AddGroup("Help Options", "", &help)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	hlpgrp.Hidden = true
	hlp := p.FindOptionByLongName("help")
	hlp.Description = "Show this help message"
	// make sure the --help option is hidden
	hlp.Hidden = true

	// add a hidden command
	var hiddenCmdOpts struct {
		Foo        bool `short:"f" long:"very-long-foo-option" description:"Very long foo description"`
		Bar        bool `short:"b" description:"Option bar"`
		Positional struct {
			PositionalFoo string `positional-arg-name:"<positional-foo>" description:"positional foo"`
		} `positional-args:"yes"`
	}
	cmdHidden, err := p.Command.AddCommand("hidden", "Hidden command description", "Long hidden command description", &hiddenCmdOpts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// make it hidden
	cmdHidden.Hidden = true
	if len(cmdHidden.Options()) != 2 {
		t.Fatalf("unexpected options count")
	}
	// which help we ask for explicitly
	_, err = p.ParseArgs([]string{"hidden", "--help"})

	if err == nil {
		t.Fatalf("Expected help error")
	}
	if e, ok := err.(*Error); !ok {
		t.Fatalf("Expected flags.Error, but got %T", err)
	} else {
		if e.Type != ErrHelp {
			t.Errorf("Expected flags.ErrHelp type, but got %s", e.Type)
		}

		var expected string

		if runtime.GOOS == "windows" {
			expected = `Usage:
  TestHelpCommand hidden [hidden-OPTIONS] [<positional-foo>]

Long hidden command description

[hidden command arguments]
  <positional-foo>:         positional foo
`
		} else {
			expected = `Usage:
  TestHelpCommand hidden [hidden-OPTIONS] [<positional-foo>]

Long hidden command description

[hidden command arguments]
  <positional-foo>:         positional foo
`
		}
		h := &bytes.Buffer{}
		p.WriteHelp(h)

		assertDiff(t, h.String(), expected, "help message")
	}
}

func TestHelpDefaults(t *testing.T) {
	var expected string

	if runtime.GOOS == "windows" {
		expected = `Usage:
  TestHelpDefaults [OPTIONS]

Application Options:
      /with-default:               With default (default: default-value)
      /without-default:            Without default
      /with-programmatic-default:  With programmatic default
                                   (default: default-value)

Help Options:
  /?                               Show this help message
  /h, /help                        Show this help message
`
	} else {
		expected = `Usage:
  TestHelpDefaults [OPTIONS]

Application Options:
      --with-default=              With default (default: default-value)
      --without-default=           Without default
      --with-programmatic-default= With programmatic default
                                   (default: default-value)

Help Options:
  -h, --help                       Show this help message
`
	}

	tests := []struct {
		Args   []string
		Output string
	}{
		{
			Args:   []string{"-h"},
			Output: expected,
		},
		{
			Args:   []string{"--with-default", "other-value", "--with-programmatic-default", "other-value", "-h"},
			Output: expected,
		},
	}

	for _, test := range tests {
		var opts struct {
			WithDefault             string `long:"with-default" default:"default-value" description:"With default"`
			WithoutDefault          string `long:"without-default" description:"Without default"`
			WithProgrammaticDefault string `long:"with-programmatic-default" description:"With programmatic default"`
		}

		opts.WithProgrammaticDefault = "default-value"

		p := NewNamedParser("TestHelpDefaults", HelpFlag)
		p.AddGroup("Application Options", "The application options", &opts)

		_, err := p.ParseArgs(test.Args)

		if err == nil {
			t.Fatalf("Expected help error")
		}

		if e, ok := err.(*Error); !ok {
			t.Fatalf("Expected flags.Error, but got %T", err)
		} else {
			if e.Type != ErrHelp {
				t.Errorf("Expected flags.ErrHelp type, but got %s", e.Type)
			}

			assertDiff(t, e.Message, test.Output, "help message")
		}
	}
}

func TestHelpRestArgs(t *testing.T) {
	opts := struct {
		Verbose bool `short:"v"`
	}{}

	p := NewNamedParser("TestHelpDefaults", HelpFlag)
	p.AddGroup("Application Options", "The application options", &opts)

	retargs, err := p.ParseArgs([]string{"-h", "-v", "rest"})

	if err == nil {
		t.Fatalf("Expected help error")
	}

	assertStringArray(t, retargs, []string{"-v", "rest"})
}

func TestWrapText(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod " +
		"tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim " +
		"veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea " +
		"commodo consequat. Duis aute irure dolor in reprehenderit in voluptate " +
		"velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint " +
		"occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum."

	got := wrapText(s, 60, "      ", true)
	expected := `Lorem ipsum dolor sit amet, consectetur adipisicing elit,
      sed do eiusmod tempor incididunt ut labore et dolore magna
      aliqua. Ut enim ad minim veniam, quis nostrud exercitation
      ullamco laboris nisi ut aliquip ex ea commodo consequat.
      Duis aute irure dolor in reprehenderit in voluptate velit
      esse cillum dolore eu fugiat nulla pariatur. Excepteur sint
      occaecat cupidatat non proident, sunt in culpa qui officia
      deserunt mollit anim id est laborum.`

	assertDiff(t, got, expected, "wrapped text")
}

func TestWrapParagraph(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n"
	s += "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\n\n"
	s += "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.\n\n"
	s += "Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.\n"

	got := wrapText(s, 60, "      ", true)
	expected := `Lorem ipsum dolor sit amet, consectetur adipisicing elit,
      sed do eiusmod tempor incididunt ut labore et dolore magna
      aliqua.

      Ut enim ad minim veniam, quis nostrud exercitation ullamco
      laboris nisi ut aliquip ex ea commodo consequat.

      Duis aute irure dolor in reprehenderit in voluptate velit
      esse cillum dolore eu fugiat nulla pariatur.

      Excepteur sint occaecat cupidatat non proident, sunt in
      culpa qui officia deserunt mollit anim id est laborum.
`

	assertDiff(t, got, expected, "wrapped paragraph")
}

func TestWrapTextKeepWhitespace(t *testing.T) {
	s := "List:\n  - alpha\n  - beta"

	gotTrimmed := wrapText(s, 80, "", true)
	if strings.Contains(gotTrimmed, "\n  - alpha") {
		t.Fatalf("expected trimmed output to remove leading spaces, got %q", gotTrimmed)
	}

	gotPreserved := wrapText(s, 80, "", false)
	if !strings.Contains(gotPreserved, "\n  - alpha") {
		t.Fatalf("expected preserved output to keep leading spaces, got %q", gotPreserved)
	}
}

func TestHelpAdaptiveLayoutKeepsDescriptionWidth(t *testing.T) {
	var opts struct {
		Format string `long:"output-format-negotiation-policy-for-generated-artifacts" value-name:"OUTPUT_FORMAT_NEGOTIATION_POLICY_IDENTIFIER" choice:"prefer-human-readable-markdown-with-inline-metadata" choice:"prefer-machine-readable-json-with-stable-field-order" choice:"prefer-manpage-compatible-plain-text-with-unicode-disabled" description:"Description marker for adaptive layout"`
	}

	p := NewNamedParser("AdaptiveHelp", None)
	if err := p.SetMaxLongNameLength(256); err != nil {
		t.Fatalf("unexpected set max long name length error: %v", err)
	}
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	opt := p.FindOptionByLongName("output-format-negotiation-policy-for-generated-artifacts")
	if opt == nil {
		t.Fatalf("expected option to be registered")
	}

	info := p.getAlignmentInfo()
	info.terminalColumns = 100

	var out bytes.Buffer
	w := bufio.NewWriter(&out)
	p.writeHelpOption(w, opt, info, true, p.optionRenderFormat())
	_ = w.Flush()

	got := out.String()
	if !strings.Contains(got, "Description marker for adaptive layout") {
		t.Fatalf("expected description marker in output, got:\n%s", got)
	}

	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapped multi-line output, got:\n%s", got)
	}

	descIdx := -1
	for _, line := range lines {
		if i := strings.Index(line, "Description marker for adaptive layout"); i >= 0 {
			descIdx = i
			break
		}
	}

	if descIdx == -1 {
		t.Fatalf("expected description line, got:\n%s", got)
	}

	if descIdx > 55 {
		t.Fatalf("expected description to start within left 55 columns, got index=%d\n%s", descIdx, got)
	}
}

func TestHelpAdaptiveLayoutBreaksAfterNameDelimiter(t *testing.T) {
	var opts struct {
		Verbose bool   `short:"v" long:"verbose" description:"Verbose"`
		Format  string `long:"output-format-negotiation-policy-for-generated-artifacts" value-name:"OUTPUT_FORMAT_NEGOTIATION_POLICY_IDENTIFIER" choice:"prefer-human-readable-markdown-with-inline-metadata" choice:"prefer-machine-readable-json-with-stable-field-order" choice:"prefer-manpage-compatible-plain-text-with-unicode-disabled" description:"Delimiter break test"`
	}

	p := NewNamedParser("AdaptiveHelpBreak", None)
	if err := p.SetMaxLongNameLength(256); err != nil {
		t.Fatalf("unexpected set max long name length error: %v", err)
	}
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	opt := p.FindOptionByLongName("output-format-negotiation-policy-for-generated-artifacts")
	if opt == nil {
		t.Fatalf("expected option to be registered")
	}

	info := p.getAlignmentInfo()
	info.terminalColumns = 100

	var out bytes.Buffer
	w := bufio.NewWriter(&out)
	p.writeHelpOption(w, opt, info, true, p.optionRenderFormat())
	_ = w.Flush()

	got := out.String()
	delimiter := string(p.optionRenderFormat().nameDelimiter)

	if !strings.Contains(got, delimiter+"\n") {
		t.Fatalf("expected break after name delimiter, got:\n%s", got)
	}

	if strings.Contains(got, "IDENT-\n") {
		t.Fatalf("unexpected hyphen split inside identifier, got:\n%s", got)
	}

	if !strings.Contains(got, "valid values:") {
		t.Fatalf("expected auto choice-list rendering marker, got:\n%s", got)
	}
}

func TestHelpShowChoiceListInHelpForcesList(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" value-name:"MODE" choice:"fast" choice:"safe" description:"Mode option"`
	}

	p := NewNamedParser("ChoiceList", None)
	p.Options |= ShowChoiceListInHelp
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	opt := p.FindOptionByLongName("mode")
	if opt == nil {
		t.Fatalf("expected option to be registered")
	}

	info := p.getAlignmentInfo()
	info.terminalColumns = 80

	var out bytes.Buffer
	w := bufio.NewWriter(&out)
	p.writeHelpOption(w, opt, info, true, p.optionRenderFormat())
	_ = w.Flush()

	got := out.String()
	if !strings.Contains(got, "valid values:") {
		t.Fatalf("expected forced choice-list marker, got:\n%s", got)
	}
	if !strings.Contains(got, "> fast") || !strings.Contains(got, "> safe") {
		t.Fatalf("expected forced list items for choices, got:\n%s", got)
	}
}

func TestHelpKeepDescriptionWhitespace(t *testing.T) {
	var opts struct {
		Cmd struct{} `command:"cmd" description:"Run command" long-description:"Usage:\n  cmd --flag value\n\n  - item 1\n  - item 2"`
	}

	trimmed := NewNamedParser("TrimDesc", HelpFlag)
	if _, err := trimmed.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err := trimmed.ParseArgs([]string{"cmd", "--help"})
	if err == nil {
		t.Fatalf("expected help error")
	}

	e, ok := err.(*Error)
	if !ok || e.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}

	if strings.Contains(e.Message, "\n  cmd --flag value") {
		t.Fatalf("expected default help to trim leading spaces, got:\n%s", e.Message)
	}

	preserved := NewNamedParser("KeepDesc", HelpFlag|KeepDescriptionWhitespace)
	if _, err := preserved.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err = preserved.ParseArgs([]string{"cmd", "--help"})
	if err == nil {
		t.Fatalf("expected help error")
	}

	e, ok = err.(*Error)
	if !ok || e.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}

	if !strings.Contains(e.Message, "\n  cmd --flag value") || !strings.Contains(e.Message, "\n  - item 1") {
		t.Fatalf("expected help to keep leading spaces, got:\n%s", e.Message)
	}
}

func TestHelpDefaultMask(t *testing.T) {
	var tests = []struct {
		opts    interface{}
		present string
	}{
		{
			opts: &struct {
				Value string `short:"v" default:"123" description:"V"`
			}{},
			present: "V (default: 123)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default:"123" default-mask:"abc" description:"V"`
			}{},
			present: "V (default: abc)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default:"123" default-mask:"-" description:"V"`
			}{},
			present: "V\n",
		},
		{
			opts: &struct {
				Value string `short:"v" description:"V"`
			}{Value: "123"},
			present: "V (default: 123)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default-mask:"abc" description:"V"`
			}{Value: "123"},
			present: "V (default: abc)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default-mask:"-" description:"V"`
			}{Value: "123"},
			present: "V\n",
		},
	}

	for _, test := range tests {
		p := NewParser(test.opts, HelpFlag)
		_, err := p.ParseArgs([]string{"-h"})
		if flagsErr, ok := err.(*Error); ok && flagsErr.Type == ErrHelp {
			err = nil
		}
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		h := &bytes.Buffer{}
		w := bufio.NewWriter(h)
		p.writeHelpOption(w, p.FindOptionByShortName('v'), p.getAlignmentInfo(), true, p.optionRenderFormat())
		w.Flush()
		if strings.Index(h.String(), test.present) < 0 {
			t.Errorf("Not present %q\n%s", test.present, h.String())
		}
	}
}

func TestHelpPositionalDefault(t *testing.T) {
	var opts struct {
		Args struct {
			Output string `positional-arg-name:"output" description:"Output file path" default:"foo.txt"`
		} `positional-args:"yes"`
	}

	p := NewNamedParser("PositionalDefaultHelp", HelpFlag)
	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	_, err := p.ParseArgs([]string{"--help"})
	if err == nil {
		t.Fatalf("expected help error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}

	if !strings.Contains(flagsErr.Message, "Output file path (default: foo.txt)") {
		t.Fatalf("expected positional default in help output, got:\n%s", flagsErr.Message)
	}
}

func TestWroteHelp(t *testing.T) {
	type testInfo struct {
		value  error
		isHelp bool
	}
	tests := map[string]testInfo{
		"No error":    {value: nil, isHelp: false},
		"Plain error": {value: errors.New("an error"), isHelp: false},
		"ErrUnknown":  {value: newError(ErrUnknown, "an error"), isHelp: false},
		"ErrHelp":     {value: newError(ErrHelp, "an error"), isHelp: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res := WroteHelp(test.value)
			if test.isHelp != res {
				t.Errorf("Expected %t, got %t", test.isHelp, res)
			}
		})
	}
}
