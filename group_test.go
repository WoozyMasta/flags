package flags

import (
	"errors"
	"strings"
	"testing"
)

type dynamicDefaultValue string

func (v *dynamicDefaultValue) Default() ([]string, error) {
	return []string{"from-provider"}, nil
}

type failingDynamicDefaultValue string

func (v *failingDynamicDefaultValue) Default() ([]string, error) {
	return nil, errors.New("dynamic default failed")
}

func TestGroupInline(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Group struct {
			G bool `short:"g"`
		} `group:"Grouped Options"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "-g")

	assertStringArray(t, ret, []string{})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Group.G {
		t.Errorf("Expected Group.G to be true")
	}

	if p.Command.Group.Find("Grouped Options") == nil {
		t.Errorf("Expected to find group `Grouped Options`")
	}
}

func TestGroupAdd(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	var grp = struct {
		G bool `short:"g"`
	}{}

	p := NewParser(&opts, Default)
	g, err := p.AddGroup("Grouped Options", "", &grp)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	ret, err := p.ParseArgs([]string{"-v", "-g", "rest"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !grp.G {
		t.Errorf("Expected Group.G to be true")
	}

	if p.Command.Group.Find("Grouped Options") != g {
		t.Errorf("Expected to find group `Grouped Options`")
	}

	if p.Groups()[1] != g {
		t.Errorf("Expected group %#v, but got %#v", g, p.Groups()[0])
	}

	if g.Options()[0].ShortName != 'g' {
		t.Errorf("Expected short name `g` but got %v", g.Options()[0].ShortName)
	}

	if g.Data() != &grp {
		t.Errorf("Expected data pointer %#v, but got %#v", &grp, g.Data())
	}
}

func TestGroupNestedInline(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Group struct {
			G bool `short:"g"`

			Nested struct {
				N string `long:"n"`
			} `group:"Nested Options"`
		} `group:"Grouped Options"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "-g", "--n", "n", "rest")

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Group.G {
		t.Errorf("Expected Group.G to be true")
	}

	assertString(t, opts.Group.Nested.N, "n")

	if p.Command.Group.Find("Grouped Options") == nil {
		t.Errorf("Expected to find group `Grouped Options`")
	}

	if p.Command.Group.Find("Nested Options") == nil {
		t.Errorf("Expected to find group `Nested Options`")
	}
}

func TestGroupNestedInlineNamespace(t *testing.T) {
	var opts = struct {
		Opt string `long:"opt"`

		Group struct {
			Opt   string `long:"opt"`
			Group struct {
				Opt string `long:"opt"`
			} `group:"Subsubgroup" namespace:"sap"`
		} `group:"Subgroup" namespace:"sip"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "--opt", "a", "--sip.opt", "b", "--sip.sap.opt", "c", "rest")

	assertStringArray(t, ret, []string{"rest"})

	assertString(t, opts.Opt, "a")
	assertString(t, opts.Group.Opt, "b")
	assertString(t, opts.Group.Group.Opt, "c")

	for _, name := range []string{"Subgroup", "Subsubgroup"} {
		if p.Command.Group.Find(name) == nil {
			t.Errorf("Expected to find group '%s'", name)
		}
	}
}

func TestGroupHiddenTagNo(t *testing.T) {
	var opts = struct {
		Group struct {
			Opt string `long:"opt"`
		} `group:"Visible Group" hidden:"no"`
	}{}

	p := NewParser(&opts, None)
	g := p.Command.Group.Find("Visible Group")
	if g == nil {
		t.Fatalf("expected group to be found")
	}

	if g.Hidden {
		t.Fatalf("expected group hidden to be false")
	}
}

func TestGroupHiddenTagInvalidValue(t *testing.T) {
	var opts = struct {
		Group struct {
			Opt string `long:"opt"`
		} `group:"Broken Group" hidden:"maybe"`
	}{}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestGroupI18nTagRequiresIniGroup(t *testing.T) {
	var opts = struct {
		Group struct {
			Opt string `long:"opt"`
		} `group:"Localized Group" group-i18n:"group.localized"`
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

func TestDuplicateShortFlags(t *testing.T) {
	var opts struct {
		Verbose   []bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
		Variables []string `short:"v" long:"variable" description:"Set a variable value."`
	}

	args := []string{
		"--verbose",
		"-v", "123",
		"-v", "456",
	}

	_, err := ParseArgs(&opts, args)

	if err == nil {
		t.Errorf("Expected an error with type ErrDuplicatedFlag")
	} else {
		err2 := err.(*Error)
		if err2.Type != ErrDuplicatedFlag {
			t.Errorf("Expected an error with type ErrDuplicatedFlag")
		}
	}
}

func TestDuplicateLongFlags(t *testing.T) {
	var opts struct {
		Test1 []bool   `short:"a" long:"testing" description:"Test 1"`
		Test2 []string `short:"b" long:"testing" description:"Test 2."`
	}

	args := []string{
		"--testing",
	}

	_, err := ParseArgs(&opts, args)

	if err == nil {
		t.Errorf("Expected an error with type ErrDuplicatedFlag")
	} else {
		err2 := err.(*Error)
		if err2.Type != ErrDuplicatedFlag {
			t.Errorf("Expected an error with type ErrDuplicatedFlag")
		}
	}
}

func TestFindOptionByLongFlag(t *testing.T) {
	var opts struct {
		Testing bool `long:"testing" description:"Testing"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("testing")

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	assertString(t, opt.LongName, "testing")
}

func TestFindOptionByShortFlag(t *testing.T) {
	var opts struct {
		Testing bool `short:"t" description:"Testing"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByShortName('t')

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	if opt.ShortName != 't' {
		t.Errorf("Expected 't', but got %v", opt.ShortName)
	}
}

func TestFindOptionByLongFlagInSubGroup(t *testing.T) {
	var opts struct {
		Group struct {
			Testing bool `long:"testing" description:"Testing"`
		} `group:"sub-group"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("testing")

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	assertString(t, opt.LongName, "testing")
}

func TestFindOptionByShortFlagInSubGroup(t *testing.T) {
	var opts struct {
		Group struct {
			Testing bool `short:"t" description:"Testing"`
		} `group:"sub-group"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByShortName('t')

	if opt == nil {
		t.Errorf("Expected option, but found none")
	}

	if opt.ShortName != 't' {
		t.Errorf("Expected 't', but got %v", opt.ShortName)
	}
}

func TestAddOptionNonOptional(t *testing.T) {
	var opts struct {
		Test bool
	}
	p := NewParser(&opts, Default)
	p.AddOption(&Option{
		LongName: "test",
	}, &opts.Test)
	_, err := p.ParseArgs([]string{"--test"})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !opts.Test {
		t.Errorf("option not set")
	}
}

func TestDynamicDefaultProvider(t *testing.T) {
	var opts struct {
		Value dynamicDefaultValue `long:"value"`
	}

	_, err := ParseArgs(&opts, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, string(opts.Value), "from-provider")
}

func TestDynamicDefaultProviderError(t *testing.T) {
	var opts struct {
		Value failingDynamicDefaultValue `long:"value"`
	}

	_, err := ParseArgs(&opts, nil)

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}

	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, but got %T", err)
	}

	if flagsErr.Type != ErrMarshal {
		t.Fatalf("Expected ErrMarshal, but got %v", flagsErr.Type)
	}
}
