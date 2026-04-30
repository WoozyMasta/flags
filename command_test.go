package flags

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestCommandInline(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "cmd", "-g")

	assertStringArray(t, ret, []string{})

	if p.Active == nil {
		t.Errorf("Expected active command")
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.G to be true")
	}

	if p.Command.Find("cmd") != p.Active {
		t.Errorf("Expected to find command `cmd` to be active")
	}
}

func TestCommandInlineMulti(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		C1 struct {
		} `command:"c1"`

		C2 struct {
			G bool `short:"g"`
		} `command:"c2"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "c2", "-g")

	assertStringArray(t, ret, []string{})

	if p.Active == nil {
		t.Errorf("Expected active command")
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.C2.G {
		t.Errorf("Expected C2.G to be true")
	}

	if p.Command.Find("c1") == nil {
		t.Errorf("Expected to find command `c1`")
	}

	if c2 := p.Command.Find("c2"); c2 == nil {
		t.Errorf("Expected to find command `c2`")
	} else if c2 != p.Active {
		t.Errorf("Expected to find command `c2` to be active")
	}
}

func TestCommandI18nTagRequiresIniGroup(t *testing.T) {
	var opts = struct {
		Command struct {
			Opt bool `long:"opt"`
		} `command:"run" command-i18n:"command.run"`
	}{}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}

	if !strings.Contains(flagsErr.Message, FlagTagIniGroup) {
		t.Fatalf("expected %q in error, got: %s", FlagTagIniGroup, flagsErr.Message)
	}
}

func TestCommandFlagOrder1(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, ErrUnknownFlag, "unknown flag `g`", &opts, "-v", "-g", "cmd")
}

func TestCommandFlagOrder2(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd", "-v", "-g")

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.G to be true")
	}
}

func TestCommandFlagOrderSub(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`

			SubCommand struct {
				B bool `short:"b"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd", "sub", "-v", "-g", "-b")

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.G to be true")
	}

	if !opts.Command.SubCommand.B {
		t.Errorf("Expected Command.SubCommand.B to be true")
	}
}

func TestCommandFlagOverride1(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}{}

	assertCommandFlagOverrideRejected(t, &opts, "-v", "cmd")
}

func TestCommandFlagOverride2(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}{}

	assertCommandFlagOverrideRejected(t, &opts, "cmd", "-v")
}

func TestCommandFlagOverrideSub(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`

			SubCommand struct {
				Value bool `short:"v"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertCommandFlagOverrideRejected(t, &opts, "cmd", "sub", "-v")
}

func TestCommandFlagOverrideSub2(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`

			SubCommand struct {
				G bool `short:"g"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertCommandFlagOverrideRejected(t, &opts, "cmd", "sub", "-v")
}

func assertCommandFlagOverrideRejected(t *testing.T, data interface{}, args ...string) {
	t.Helper()

	_, err := NewParser(data, Default&^PrintErrors).ParseArgs(args)
	if err == nil {
		t.Fatalf("expected duplicate flag error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestCommandEstimate(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Cmd1 struct {
		} `command:"remove"`

		Cmd2 struct {
		} `command:"add"`
	}{}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{})

	assertError(t, err, ErrCommandRequired, "Please specify one command of: add or remove")
}

func TestCommandEstimate2(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Cmd1 struct {
		} `command:"remove"`

		Cmd2 struct {
		} `command:"add"`
	}{}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"rmive"})

	assertError(t, err, ErrUnknownCommand, "Unknown command `rmive`, did you mean `remove`?")
}

type testCommand struct {
	G        bool `short:"g"`
	Executed bool
	EArgs    []string
}

func (c *testCommand) Execute(args []string) error {
	c.Executed = true
	c.EArgs = args

	return nil
}

func TestCommandExecute(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command testCommand `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "-v", "cmd", "-g", "a", "b")

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.Executed {
		t.Errorf("Did not execute command")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.C to be true")
	}

	assertStringArray(t, opts.Command.EArgs, []string{"a", "b"})
}

type chainCommand struct {
	calls *[]string
	err   error
	name  string
}

func (c *chainCommand) Execute(args []string) error {
	if c.calls != nil {
		*c.calls = append(*c.calls, c.name)
	}

	return c.err
}

func TestCommandExecuteDefaultOnlyRunsLeaf(t *testing.T) {
	var calls []string

	root := &chainCommand{name: "root", calls: &calls}
	parent := &chainCommand{name: "parent", calls: &calls}
	leaf := &chainCommand{name: "leaf", calls: &calls}

	parser := NewNamedParser("app", Default&^PrintErrors)
	parser.Command.data = root

	parentCommand, err := parser.AddCommand("parent", "", "", parent)
	if err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parentCommand.AddCommand("leaf", "", "", leaf); err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parser.ParseArgs([]string{"parent", "leaf"}); err != nil {
		t.Fatalf("Unexpected parse error: %v", err)
	}

	assertStringArray(t, calls, []string{"leaf"})
}

func TestCommandChainExecuteRunsParentToLeaf(t *testing.T) {
	var calls []string

	root := &chainCommand{name: "root", calls: &calls}
	parent := &chainCommand{name: "parent", calls: &calls}
	leaf := &chainCommand{name: "leaf", calls: &calls}

	parser := NewNamedParser("app", Default&^PrintErrors|CommandChain)
	parser.Command.data = root

	parentCommand, err := parser.AddCommand("parent", "", "", parent)
	if err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parentCommand.AddCommand("leaf", "", "", leaf); err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parser.ParseArgs([]string{"parent", "leaf", "arg"}); err != nil {
		t.Fatalf("Unexpected parse error: %v", err)
	}

	assertStringArray(t, calls, []string{"root", "parent", "leaf"})
}

func TestCommandChainExecuteLeafOnly(t *testing.T) {
	var calls []string

	leaf := &chainCommand{name: "leaf", calls: &calls}

	parser := NewNamedParser("app", Default&^PrintErrors|CommandChain)

	if _, err := parser.AddCommand("leaf", "", "", leaf); err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parser.ParseArgs([]string{"leaf"}); err != nil {
		t.Fatalf("Unexpected parse error: %v", err)
	}

	assertStringArray(t, calls, []string{"leaf"})
}

func TestCommandChainExecuteStopsOnError(t *testing.T) {
	var calls []string
	expected := errors.New("parent failed")

	root := &chainCommand{name: "root", calls: &calls}
	parent := &chainCommand{name: "parent", calls: &calls, err: expected}
	leaf := &chainCommand{name: "leaf", calls: &calls}

	parser := NewNamedParser("app", Default&^PrintErrors|CommandChain)
	parser.Command.data = root

	parentCommand, err := parser.AddCommand("parent", "", "", parent)
	if err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parentCommand.AddCommand("leaf", "", "", leaf); err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	_, err = parser.ParseArgs([]string{"parent", "leaf"})
	if !errors.Is(err, expected) {
		t.Fatalf("Expected %v, got %v", expected, err)
	}

	assertStringArray(t, calls, []string{"root", "parent"})
}

func TestCommandChainHandlerRunsParentToLeaf(t *testing.T) {
	var calls []string

	root := &chainCommand{name: "root"}
	parent := &chainCommand{name: "parent"}
	leaf := &chainCommand{name: "leaf"}
	names := map[Commander]string{
		root:   "root",
		parent: "parent",
		leaf:   "leaf",
	}

	parser := NewNamedParser("app", Default&^PrintErrors|CommandChain)
	parser.Command.data = root
	parser.CommandHandler = func(command Commander, args []string) error {
		calls = append(calls, names[command])
		assertStringArray(t, args, []string{"arg"})
		return nil
	}

	parentCommand, err := parser.AddCommand("parent", "", "", parent)
	if err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parentCommand.AddCommand("leaf", "", "", leaf); err != nil {
		t.Fatalf("Unexpected command error: %v", err)
	}

	if _, err := parser.ParseArgs([]string{"parent", "leaf", "arg"}); err != nil {
		t.Fatalf("Unexpected parse error: %v", err)
	}

	assertStringArray(t, calls, []string{"root", "parent", "leaf"})
}

func TestCommandClosest(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Cmd1 struct {
		} `command:"remove"`

		Cmd2 struct {
		} `command:"add"`
	}{}

	args := assertParseFail(t, ErrUnknownCommand, "Unknown command `addd`, did you mean `add`?", &opts, "-v", "addd")

	assertStringArray(t, args, []string{"addd"})
}

func TestCommandAdd(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	var cmd = struct {
		G bool `short:"g"`
	}{}

	p := NewParser(&opts, Default)
	c, err := p.AddCommand("cmd", "", "", &cmd)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	ret, err := p.ParseArgs([]string{"-v", "cmd", "-g", "rest"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !cmd.G {
		t.Errorf("Expected Command.G to be true")
	}

	if p.Command.Find("cmd") != c {
		t.Errorf("Expected to find command `cmd`")
	}

	if p.Commands()[0] != c {
		t.Errorf("Expected command %#v, but got %#v", c, p.Commands()[0])
	}

	if c.Options()[0].ShortName != 'g' {
		t.Errorf("Expected short name `g` but got %v", c.Options()[0].ShortName)
	}
}

func TestCommandNestedInline(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`

			Nested struct {
				N string `long:"n"`
			} `command:"nested"`
		} `command:"cmd"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "cmd", "-g", "nested", "--n", "n", "rest")

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.G to be true")
	}

	assertString(t, opts.Command.Nested.N, "n")

	if c := p.Command.Find("cmd"); c == nil {
		t.Errorf("Expected to find command `cmd`")
	} else {
		if c != p.Active {
			t.Errorf("Expected `cmd` to be the active parser command")
		}

		if nested := c.Find("nested"); nested == nil {
			t.Errorf("Expected to find command `nested`")
		} else if nested != c.Active {
			t.Errorf("Expected to find command `nested` to be the active `cmd` command")
		}
	}
}

func TestRequiredOnCommand(t *testing.T) {
	var opts = struct {
		Value bool `short:"v" required:"true"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, ErrRequired, fmt.Sprintf("the required flag `%cv` was not specified", defaultShortOptDelimiter), &opts, "cmd")
}

func TestRequiredAllOnCommand(t *testing.T) {
	var opts = struct {
		Value   bool `short:"v" required:"true"`
		Missing bool `long:"missing" required:"true"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(
		t,
		ErrRequired,
		fmt.Sprintf(
			"the required flags `%smissing` and `%cv` were not specified",
			defaultLongOptDelimiter,
			defaultShortOptDelimiter,
		),
		&opts,
		"cmd",
	)
}

func TestDefaultOnCommand(t *testing.T) {
	var opts = struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd")

	if opts.Command.G != "value" {
		t.Errorf("Expected G to be \"value\"")
	}
}

func TestAfterNonCommand(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Cmd1 struct {
		} `command:"remove"`

		Cmd2 struct {
		} `command:"add"`
	}{}

	assertParseFail(t, ErrUnknownCommand, "Unknown command `nocmd`. Please specify one command of: add or remove", &opts, "nocmd", "remove")
}

func TestSubcommandsOptional(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Cmd1 struct {
		} `command:"remove"`

		Cmd2 struct {
		} `command:"add"`
	}{}

	p := NewParser(&opts, None)
	p.SubcommandsOptional = true

	_, err := p.ParseArgs([]string{"-v"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}
}

func TestSubcommandsOptionalAfterNonCommand(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Cmd1 struct {
		} `command:"remove"`

		Cmd2 struct {
		} `command:"add"`
	}{}

	p := NewParser(&opts, None)
	p.SubcommandsOptional = true

	retargs, err := p.ParseArgs([]string{"nocmd", "remove"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	assertStringArray(t, retargs, []string{"nocmd", "remove"})
}

func TestCommandAlias(t *testing.T) {
	var opts = struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd" alias:"cm"`
	}{}

	assertParseSuccess(t, &opts, "cm")

	if opts.Command.G != "value" {
		t.Errorf("Expected G to be \"value\"")
	}
}

func TestCommandAliasesListTag(t *testing.T) {
	var opts = struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd" aliases:"cm;c"`
	}{}

	assertParseSuccess(t, &opts, "c")

	if opts.Command.G != "value" {
		t.Errorf("Expected G to be \"value\"")
	}
}

func TestCommandAliasAndAliasesTagsConflict(t *testing.T) {
	var opts = struct {
		Command struct{} `command:"cmd" alias:"cm" aliases:"c"`
	}{}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestCommandAliasesListTagCustomDelimiter(t *testing.T) {
	var opts = struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd" aliases:"cm,c"`
	}{}

	p := NewParser(&opts, None)

	if err := p.SetTagListDelimiter(','); err != nil {
		t.Fatalf("unexpected set delimiter error: %v", err)
	}

	_, err := p.ParseArgs([]string{"cm"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Command.G != "value" {
		t.Errorf("Expected G to be \"value\"")
	}
}

func TestSubCommandFindOptionByLongFlag(t *testing.T) {
	var opts struct {
		Testing bool `long:"testing" description:"Testing"`
	}

	var cmd struct {
		Other bool `long:"other" description:"Other"`
	}

	p := NewParser(&opts, Default)
	c, _ := p.AddCommand("command", "Short", "Long", &cmd)

	opt := c.FindOptionByLongName("other")

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	assertString(t, opt.LongName, "other")

	opt = c.FindOptionByLongName("testing")

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	assertString(t, opt.LongName, "testing")
}

func TestSubCommandFindOptionByShortFlag(t *testing.T) {
	var opts struct {
		Testing bool `short:"t" description:"Testing"`
	}

	var cmd struct {
		Other bool `short:"o" description:"Other"`
	}

	p := NewParser(&opts, Default)
	c, _ := p.AddCommand("command", "Short", "Long", &cmd)

	opt := c.FindOptionByShortName('o')

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	if opt.ShortName != 'o' {
		t.Errorf("Expected 'o', but got %v", opt.ShortName)
	}

	opt = c.FindOptionByShortName('t')

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	if opt.ShortName != 't' {
		t.Errorf("Expected 'o', but got %v", opt.ShortName)
	}
}

type fooCmd struct {
	Flag bool `short:"f"`
	args []string
}

func (foo *fooCmd) Execute(s []string) error {
	foo.args = s
	return nil
}

func TestCommandPassAfterNonOption(t *testing.T) {
	var opts = struct {
		Value bool   `short:"v"`
		Foo   fooCmd `command:"foo"`
	}{}
	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "foo", "-f", "bar", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Foo.Flag {
		t.Errorf("Expected Foo.Flag to be true")
	}

	assertStringArray(t, ret, []string{"bar", "-v", "-g"})
	assertStringArray(t, opts.Foo.args, []string{"bar", "-v", "-g"})
}

type barCmd struct {
	fooCmd
	Positional struct {
		Args []string
	} `positional-args:"yes"`
}

func TestCommandPassAfterNonOptionWithPositional(t *testing.T) {
	var opts = struct {
		Value bool   `short:"v"`
		Bar   barCmd `command:"bar"`
	}{}
	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "bar", "-f", "baz", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Bar.Flag {
		t.Errorf("Expected Bar.Flag to be true")
	}

	assertStringArray(t, ret, []string{})
	assertStringArray(t, opts.Bar.args, []string{})
	assertStringArray(t, opts.Bar.Positional.Args, []string{"baz", "-v", "-g"})
}

type cmdLocalPassAfterNonOptionMix struct {
	FlagA bool `short:"a"`
	Cmd1  struct {
		FlagB      bool `short:"b"`
		Positional struct {
			Args []string
		} `positional-args:"yes"`
	} `command:"cmd1" pass-after-non-option:"yes"`
	Cmd2 struct {
		FlagB      bool `short:"b"`
		Positional struct {
			Args []string
		} `positional-args:"yes"`
	} `command:"cmd2"`
}

func TestCommandLocalPassAfterNonOptionMixCmd1(t *testing.T) {
	var opts cmdLocalPassAfterNonOptionMix

	assertParseSuccess(t, &opts, "cmd1", "-b", "arg1", "-a", "arg2", "-x")

	if opts.FlagA {
		t.Errorf("Expected FlagA to be false")
	}

	if !opts.Cmd1.FlagB {
		t.Errorf("Expected Cmd1.FlagB to be true")
	}

	assertStringArray(t, opts.Cmd1.Positional.Args, []string{"arg1", "-a", "arg2", "-x"})
}

func TestCommandLocalPassAfterNonOptionMixCmd2(t *testing.T) {
	var opts cmdLocalPassAfterNonOptionMix

	assertParseSuccess(t, &opts, "cmd2", "-b", "arg1", "-a", "arg2")

	if !opts.FlagA {
		t.Errorf("Expected FlagA to be true")
	}

	if !opts.Cmd2.FlagB {
		t.Errorf("Expected Cmd2.FlagB to be true")
	}

	assertStringArray(t, opts.Cmd2.Positional.Args, []string{"arg1", "arg2"})
}

func TestCommandLocalPassAfterNonOptionMixCmd2UnkownFlag(t *testing.T) {
	var opts cmdLocalPassAfterNonOptionMix

	assertParseFail(t, ErrUnknownFlag, "unknown flag `x`", &opts, "cmd2", "-b", "arg1", "-a", "arg2", "-x")
}

type cmdLocalPassAfterNonOptionNest struct {
	FlagA bool `short:"a"`
	Cmd1  struct {
		FlagB bool `short:"b"`
		Cmd2  struct {
			FlagC bool `short:"c"`
			Cmd3  struct {
				FlagD bool `short:"d"`
			} `command:"cmd3"`
		} `command:"cmd2" subcommands-optional:"yes" pass-after-non-option:"yes"`
	} `command:"cmd1"`
}

func TestCommandLocalPassAfterNonOptionNest1(t *testing.T) {
	var opts cmdLocalPassAfterNonOptionNest

	ret := assertParseSuccess(t, &opts, "cmd1", "cmd2", "-a", "x", "-b", "cmd3", "-c", "-d")

	if !opts.FlagA {
		t.Errorf("Expected FlagA to be true")
	}

	if opts.Cmd1.FlagB {
		t.Errorf("Expected Cmd1.FlagB to be false")
	}

	if opts.Cmd1.Cmd2.FlagC {
		t.Errorf("Expected Cmd1.Cmd2.FlagC to be false")
	}

	if opts.Cmd1.Cmd2.Cmd3.FlagD {
		t.Errorf("Expected Cmd1.Cmd2.Cmd3.FlagD to be false")
	}

	assertStringArray(t, ret, []string{"x", "-b", "cmd3", "-c", "-d"})
}

func TestCommandLocalPassAfterNonOptionNest2(t *testing.T) {
	var opts cmdLocalPassAfterNonOptionNest

	ret := assertParseSuccess(t, &opts, "cmd1", "cmd2", "cmd3", "-a", "x", "-b", "-c", "-d")

	if !opts.FlagA {
		t.Errorf("Expected FlagA to be true")
	}

	if !opts.Cmd1.FlagB {
		t.Errorf("Expected Cmd1.FlagB to be true")
	}

	if !opts.Cmd1.Cmd2.FlagC {
		t.Errorf("Expected Cmd1.Cmd2.FlagC to be true")
	}

	if !opts.Cmd1.Cmd2.Cmd3.FlagD {
		t.Errorf("Expected Cmd1.Cmd2.Cmd3.FlagD to be true")
	}

	assertStringArray(t, ret, []string{"x"})
}

func TestCommandBooleanTagsNoValues(t *testing.T) {
	var opts = struct {
		Cmd struct {
			Positional struct {
				Args []string
			} `positional-args:"yes"`
		} `command:"cmd" pass-after-non-option:"no" subcommands-optional:"no"`
	}{}

	assertParseFail(t, ErrUnknownFlag, "unknown flag `x`", &opts, "cmd", "arg1", "-x")
}

func TestCommandBooleanTagsInvalidValue(t *testing.T) {
	var opts struct {
		Cmd struct{} `command:"cmd" pass-after-non-option:"maybe"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestCommandPositionalRequiredTagNo(t *testing.T) {
	var opts struct {
		Cmd struct {
			Positional struct {
				Name string
			} `positional-args:"yes" required:"no"`
		} `command:"cmd"`
	}

	p := NewParser(&opts, None)
	cmd := p.Command.Find("cmd")
	if cmd == nil {
		t.Fatalf("command cmd not found")
	}

	if cmd.ArgsRequired {
		t.Fatalf("expected ArgsRequired to be false")
	}
}

func TestCommandPositionalRequiredTagInvalidValue(t *testing.T) {
	var opts struct {
		Cmd struct {
			Positional struct {
				Name string
			} `positional-args:"yes" required:"maybe"`
		} `command:"cmd"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestCommandArgsReturnsCopyOfSlice(t *testing.T) {
	var opts struct {
		Cmd struct {
			Positional struct {
				Name string
			} `positional-args:"yes"`
		} `command:"cmd"`
	}

	p := NewParser(&opts, None)
	cmd := p.Command.Find("cmd")
	if cmd == nil {
		t.Fatalf("command cmd not found")
	}

	got := cmd.Args()
	originalLen := len(got)
	got = append(got, &Arg{Name: "extra"})

	if len(cmd.Args()) != originalLen {
		t.Fatalf("expected original args length to stay %d", originalLen)
	}
}
