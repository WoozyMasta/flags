package flags

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type defaultOptions struct {
	Int           int `long:"i"`
	IntDefault    int `long:"id" default:"1"`
	IntUnderscore int `long:"idu" default:"1_0"`

	Float64           float64 `long:"f"`
	Float64Default    float64 `long:"fd" default:"-3.14"`
	Float64Underscore float64 `long:"fdu" default:"-3_3.14"`

	NumericFlag bool `short:"3"`

	String            string `long:"str"`
	StringDefault     string `long:"strd" default:"abc"`
	StringNotUnquoted string `long:"strnot" unquote:"false"`

	Time        time.Duration `long:"t"`
	TimeDefault time.Duration `long:"td" default:"1m"`

	Map        map[string]int `long:"m"`
	MapDefault map[string]int `long:"md" default:"a:1"`

	Slice        []int `long:"s"`
	SliceDefault []int `long:"sd" default:"1" default:"2"`
}

func TestDefaults(t *testing.T) {
	var tests = []struct {
		msg         string
		args        []string
		expected    defaultOptions
		expectedErr string
	}{
		{
			msg:  "no arguments, expecting default values",
			args: []string{},
			expected: defaultOptions{
				Int:           0,
				IntDefault:    1,
				IntUnderscore: 10,

				Float64:           0.0,
				Float64Default:    -3.14,
				Float64Underscore: -33.14,

				NumericFlag: false,

				String:        "",
				StringDefault: "abc",

				Time:        0,
				TimeDefault: time.Minute,

				Map:        map[string]int{},
				MapDefault: map[string]int{"a": 1},

				Slice:        []int{},
				SliceDefault: []int{1, 2},
			},
		},
		{
			msg: "non-zero value arguments, expecting overwritten arguments",
			args: []string{
				"--i=3", "--id=3", "--idu=3_3", "--f=-2.71", "--fd=2.71",
				"--fdu=2_2.71", "-3", "--str=def", "--strd=def", "--t=3ms",
				"--td=3ms", "--m=c:3", "--md=c:3", "--s=3", "--sd=3",
			},
			expected: defaultOptions{
				Int:           3,
				IntDefault:    3,
				IntUnderscore: 33,

				Float64:           -2.71,
				Float64Default:    2.71,
				Float64Underscore: 22.71,

				NumericFlag: true,

				String:        "def",
				StringDefault: "def",

				Time:        3 * time.Millisecond,
				TimeDefault: 3 * time.Millisecond,

				Map:        map[string]int{"c": 3},
				MapDefault: map[string]int{"c": 3},

				Slice:        []int{3},
				SliceDefault: []int{3},
			},
		},
		{
			msg:         "non-zero value arguments, expecting overwritten arguments",
			args:        []string{"-3=true"},
			expectedErr: "bool flag `" + makeShortName("3") + "' cannot have an argument",
		},
		{
			msg: "zero value arguments, expecting overwritten arguments",
			args: []string{
				"--i=0", "--id=0", "--idu=0", "--f=0", "--fd=0", "--fdu=0",
				"--str", "", "--strd=\"\"", "--t=0ms", "--td=0s", "--m=:0",
				"--md=:0", "--s=0", "--sd=0",
			},
			expected: defaultOptions{
				Int:           0,
				IntDefault:    0,
				IntUnderscore: 0,

				Float64:           0,
				Float64Default:    0,
				Float64Underscore: 0,

				String:        "",
				StringDefault: "",

				Time:        0,
				TimeDefault: 0,

				Map:        map[string]int{"": 0},
				MapDefault: map[string]int{"": 0},

				Slice:        []int{0},
				SliceDefault: []int{0},
			},
		},
	}

	for _, test := range tests {
		var opts defaultOptions

		_, err := ParseArgs(&opts, test.args)
		if test.expectedErr != "" {
			if err == nil {
				t.Errorf("%s:\nExpected error containing substring %q", test.msg, test.expectedErr)
			} else if !strings.Contains(err.Error(), test.expectedErr) {
				t.Errorf("%s:\nExpected error %q to contain substring %q", test.msg, err, test.expectedErr)
			}
		} else {
			if err != nil {
				t.Fatalf("%s:\nUnexpected error: %v", test.msg, err)
			}

			if opts.Slice == nil {
				opts.Slice = []int{}
			}

			if diff := cmp.Diff(test.expected, opts); diff != "" {
				t.Errorf("%s:\nUnexpected options with arguments %+v (-expected +actual):\n%s", test.msg, test.args, diff)
			}
		}
	}
}

func TestNoDefaultsForBools(t *testing.T) {
	var opts struct {
		DefaultBool bool `short:"d" default:"true"`
	}

	if runtime.GOOS == "windows" {
		assertParseFail(
			t,
			ErrInvalidTag,
			"boolean flag `/d' may not have default values, they always default "+
				"to `false' and can only be turned on",
			&opts,
		)
	} else {
		assertParseFail(
			t,
			ErrInvalidTag,
			"boolean flag `-d' may not have default values, they always default "+
				"to `false' and can only be turned on",
			&opts,
		)
	}
}

func TestUnquoting(t *testing.T) {
	var tests = []struct {
		arg   string
		err   error
		value string
	}{
		{
			arg:   "\"abc",
			err:   strconv.ErrSyntax,
			value: "",
		},
		{
			arg:   "\"\"abc\"",
			err:   strconv.ErrSyntax,
			value: "",
		},
		{
			arg:   "\"abc\"",
			err:   nil,
			value: "abc",
		},
		{
			arg:   "\"\\\"abc\\\"\"",
			err:   nil,
			value: "\"abc\"",
		},
		{
			arg:   "\"\\\"abc\"",
			err:   nil,
			value: "\"abc",
		},
	}

	for _, test := range tests {
		var opts defaultOptions

		for _, delimiter := range []bool{false, true} {
			p := NewParser(&opts, None)

			var err error
			if delimiter {
				_, err = p.ParseArgs([]string{"--str=" + test.arg, "--strnot=" + test.arg})
			} else {
				_, err = p.ParseArgs([]string{"--str", test.arg, "--strnot", test.arg})
			}

			if test.err == nil {
				if err != nil {
					t.Fatalf("Expected no error but got: %v", err)
				}

				if test.value != opts.String {
					t.Fatalf("Expected String to be %q but got %q", test.value, opts.String)
				}
				if q := strconv.Quote(test.value); q != opts.StringNotUnquoted {
					t.Fatalf("Expected StringDefault to be %q but got %q", q, opts.StringNotUnquoted)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error")
				} else if e, ok := err.(*Error); ok {
					if strings.HasPrefix(e.Message, test.err.Error()) {
						t.Fatalf("Expected error message to end with %q but got %v", test.err.Error(), e.Message)
					}
				}
			}
		}
	}
}

// EnvRestorer keeps a copy of a set of env variables and can restore the env from them
type EnvRestorer struct {
	env map[string]string
}

func (r *EnvRestorer) Restore() {
	os.Clearenv()

	for k, v := range r.env {
		os.Setenv(k, v)
	}
}

// EnvSnapshot returns a snapshot of the currently set env variables
func EnvSnapshot() *EnvRestorer {
	r := EnvRestorer{make(map[string]string)}

	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)

		if len(parts) != 2 {
			panic("got a weird env variable: " + kv)
		}

		r.env[parts[0]] = parts[1]
	}

	return &r
}

type envNestedOptions struct {
	Foo string `long:"foo" default:"z" env:"FOO"`
}

type envDefaultOptions struct {
	Int    int              `long:"i" default:"1" env:"TEST_I"`
	Time   time.Duration    `long:"t" default:"1m" env:"TEST_T"`
	Map    map[string]int   `long:"m" default:"a:1" env:"TEST_M" env-delim:";"`
	Slice  []int            `long:"s" default:"1" default:"2" env:"TEST_S" env-delim:","`
	Nested envNestedOptions `group:"nested" namespace:"nested" env-namespace:"NESTED"`
}

func TestEnvDefaults(t *testing.T) {
	var tests = []struct {
		msg         string
		args        []string
		expected    envDefaultOptions
		expectedErr string
		env         map[string]string
	}{
		{
			msg:  "no arguments, no env, expecting default values",
			args: []string{},
			expected: envDefaultOptions{
				Int:   1,
				Time:  time.Minute,
				Map:   map[string]int{"a": 1},
				Slice: []int{1, 2},
				Nested: envNestedOptions{
					Foo: "z",
				},
			},
		},
		{
			msg:  "no arguments, env defaults, expecting env default values",
			args: []string{},
			expected: envDefaultOptions{
				Int:   2,
				Time:  2 * time.Minute,
				Map:   map[string]int{"a": 2, "b": 3},
				Slice: []int{4, 5, 6},
				Nested: envNestedOptions{
					Foo: "a",
				},
			},
			env: map[string]string{
				"TEST_I":     "2",
				"TEST_T":     "2m",
				"TEST_M":     "a:2;b:3",
				"TEST_S":     "4,5,6",
				"NESTED_FOO": "a",
			},
		},
		{
			msg:         "no arguments, malformed env defaults, expecting parse error",
			args:        []string{},
			expectedErr: `parsing "two": invalid syntax`,
			env: map[string]string{
				"TEST_I": "two",
			},
		},
		{
			msg:  "non-zero value arguments, expecting overwritten arguments",
			args: []string{"--i=3", "--t=3ms", "--m=c:3", "--s=3", "--nested.foo=\"p\""},
			expected: envDefaultOptions{
				Int:   3,
				Time:  3 * time.Millisecond,
				Map:   map[string]int{"c": 3},
				Slice: []int{3},
				Nested: envNestedOptions{
					Foo: "p",
				},
			},
			env: map[string]string{
				"TEST_I":     "2",
				"TEST_T":     "2m",
				"TEST_M":     "a:2;b:3",
				"TEST_S":     "4,5,6",
				"NESTED_FOO": "a",
			},
		},
		{
			msg:  "zero value arguments, expecting overwritten arguments",
			args: []string{"--i=0", "--t=0ms", "--m=:0", "--s=0", "--nested.foo=\"\""},
			expected: envDefaultOptions{
				Int:   0,
				Time:  0,
				Map:   map[string]int{"": 0},
				Slice: []int{0},
				Nested: envNestedOptions{
					Foo: "",
				},
			},
			env: map[string]string{
				"TEST_I":     "2",
				"TEST_T":     "2m",
				"TEST_M":     "a:2;b:3",
				"TEST_S":     "4,5,6",
				"NESTED_FOO": "a",
			},
		},
	}

	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	for _, test := range tests {
		var opts envDefaultOptions
		oldEnv.Restore()
		for envKey, envValue := range test.env {
			os.Setenv(envKey, envValue)
		}
		_, err := NewParser(&opts, None).ParseArgs(test.args)
		if test.expectedErr != "" {
			if err == nil {
				t.Errorf("%s:\nExpected error containing substring %q", test.msg, test.expectedErr)
			} else if !strings.Contains(err.Error(), test.expectedErr) {
				t.Errorf("%s:\nExpected error %q to contain substring %q", test.msg, err, test.expectedErr)
			}
		} else {
			if err != nil {
				t.Fatalf("%s:\nUnexpected error: %v", test.msg, err)
			}

			if opts.Slice == nil {
				opts.Slice = []int{}
			}

			if diff := cmp.Diff(test.expected, opts); diff != "" {
				t.Errorf("%s:\nUnexpected options with arguments %+v (-expected +actual):\n%s", test.msg, test.args, diff)
			}
		}
	}
}

func TestEnvInvalidChoiceReturnsError(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		Mode string `long:"mode" env:"APP_MODE" choice:"fast" choice:"safe" required:"yes"`
	}

	_ = os.Setenv("APP_MODE", "broken")

	p := NewParser(&opts, None)
	_, err := p.ParseArgs(nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if flagsErr.Type != ErrInvalidChoice {
		t.Fatalf("expected ErrInvalidChoice, got %v (%s)", flagsErr.Type, flagsErr.Message)
	}
}

func TestEnvPrefix(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		Port int `long:"port" env:"PORT"`
		DB   struct {
			Host string `long:"host" env:"HOST"`
		} `group:"db" env-namespace:"DB"`
	}

	_ = os.Setenv("MY_APP_PORT", "8081")
	_ = os.Setenv("MY_APP_DB_HOST", "db.local")

	p := NewParser(&opts, None)
	p.SetEnvPrefix("MY_APP")

	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Port != 8081 {
		t.Fatalf("expected port 8081, got %d", opts.Port)
	}

	if opts.DB.Host != "db.local" {
		t.Fatalf("expected db host db.local, got %q", opts.DB.Host)
	}
}

func TestEnvProvisioning(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		SomeFunction string `long:"some-function" default:"fallback"`
	}

	_ = os.Setenv("SOME_FUNCTION", "from-env")

	p := NewParser(&opts, EnvProvisioning)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.SomeFunction != "from-env" {
		t.Fatalf("expected env default from long name, got %q", opts.SomeFunction)
	}
}

func TestEnvProvisioningPunctuationToUnderscore(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		SomeFunction string `long:"some.function" default:"fallback"`
	}

	_ = os.Setenv("SOME_FUNCTION", "from-env")

	p := NewParser(&opts, EnvProvisioning)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.SomeFunction != "from-env" {
		t.Fatalf("expected punctuation to map to underscore, got %q", opts.SomeFunction)
	}
}

func TestEnvProvisioningDisabledByDefault(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		SomeFunction string `long:"some-function" default:"fallback"`
	}

	_ = os.Setenv("SOME_FUNCTION", "from-env")

	p := NewParser(&opts, None)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.SomeFunction != "fallback" {
		t.Fatalf("expected tag default when auto env is disabled, got %q", opts.SomeFunction)
	}
}

func TestEnvProvisioningExplicitEnvWins(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		SomeFunction string `long:"some-function" env:"EXPLICIT_ENV" default:"fallback"`
	}

	_ = os.Setenv("SOME_FUNCTION", "auto-env")
	_ = os.Setenv("EXPLICIT_ENV", "explicit-env")

	p := NewParser(&opts, EnvProvisioning)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.SomeFunction != "explicit-env" {
		t.Fatalf("expected explicit env to win, got %q", opts.SomeFunction)
	}
}

func TestEnvProvisioningWithNamespacesAndPrefix(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		DB struct {
			SomeFunction string `long:"some-function"`
		} `group:"db" env-namespace:"DB"`
	}

	_ = os.Setenv("MY_APP_DB_SOME_FUNCTION", "from-env")

	p := NewParser(&opts, EnvProvisioning)
	p.SetEnvPrefix("MY_APP")
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.DB.SomeFunction != "from-env" {
		t.Fatalf("expected env with namespace+prefix, got %q", opts.DB.SomeFunction)
	}
}

func TestAutoEnvTagWithoutGlobalOption(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		SomeFunction string `long:"some-function" auto-env:"true"`
	}

	_ = os.Setenv("SOME_FUNCTION", "from-env")

	p := NewParser(&opts, None)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.SomeFunction != "from-env" {
		t.Fatalf("expected auto-env tag to derive env key, got %q", opts.SomeFunction)
	}
}

func TestAutoEnvTagRequiresLongName(t *testing.T) {
	var opts struct {
		ShortOnly string `short:"s" auto-env:"true"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestEnvProvisioningRequiresLongNameWhenEnvMissing(t *testing.T) {
	var opts struct {
		ShortOnly string `short:"s"`
	}

	p := NewParser(&opts, EnvProvisioning)
	_, err := p.ParseArgs(nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestEnvProvisioningNoLongButExplicitEnvIsValid(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		ShortOnly string `short:"s" env:"SHORT_ONLY_ENV"`
	}

	_ = os.Setenv("SHORT_ONLY_ENV", "from-env")

	p := NewParser(&opts, EnvProvisioning)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.ShortOnly != "from-env" {
		t.Fatalf("expected explicit env to work without long name, got %q", opts.ShortOnly)
	}
}

func TestEnvProvisioningAutoEnvFalseOptOut(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()

	var opts struct {
		SomeFunction string `long:"some-function" default:"fallback" auto-env:"false"`
	}

	_ = os.Setenv("SOME_FUNCTION", "from-env")

	p := NewParser(&opts, EnvProvisioning)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.SomeFunction != "fallback" {
		t.Fatalf("expected auto-env opt-out to keep default, got %q", opts.SomeFunction)
	}
}

func TestEnvProvisioningAutoEnvFalseSkipsLongValidation(t *testing.T) {
	var opts struct {
		ShortOnly string `short:"s" auto-env:"false"`
	}

	p := NewParser(&opts, EnvProvisioning)
	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("expected no error when auto-env is explicitly disabled, got %v", err)
	}
}

func TestAutoEnvTagInvalidValue(t *testing.T) {
	var opts struct {
		Value string `long:"value" auto-env:"maybe"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestHelpWithEnvProvisioningDoesNotPanic(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewParser(&opts, HelpFlag|EnvProvisioning)
	_, err := p.ParseArgs([]string{"--help"})
	if err == nil {
		t.Fatalf("expected help error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrHelp {
		t.Fatalf("expected ErrHelp, got %v", err)
	}
}

func TestDefaultsIfEmptyPrefilledAndCLI(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	oldEnv.Restore()
	_ = os.Setenv("TEST_DEFAULTS_IF_EMPTY", "from-env")

	var opts struct {
		Value string `long:"value" default:"from-tag" env:"TEST_DEFAULTS_IF_EMPTY"`
	}

	opts.Value = "prefilled"

	parser := NewParser(&opts, DefaultsIfEmpty)
	_, err := parser.ParseArgs(nil)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Value != "prefilled" {
		t.Fatalf("expected prefilled value to be preserved, got %q", opts.Value)
	}

	_, err = parser.ParseArgs([]string{"--value=from-cli"})
	if err != nil {
		t.Fatalf("unexpected parse error with CLI value: %v", err)
	}

	if opts.Value != "from-cli" {
		t.Fatalf("expected CLI value to win, got %q", opts.Value)
	}
}

func TestDefaultsIfEmptyCollections(t *testing.T) {
	var opts struct {
		Map   map[string]int `long:"map" default:"a:1"`
		Slice []int          `long:"slice" default:"2"`
	}

	opts.Map = map[string]int{}
	opts.Slice = []int{}

	parser := NewParser(&opts, DefaultsIfEmpty)
	_, err := parser.ParseArgs(nil)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if diff := cmp.Diff(map[string]int{"a": 1}, opts.Map); diff != "" {
		t.Fatalf("unexpected map defaults (-expected +actual):\n%s", diff)
	}

	if diff := cmp.Diff([]int{2}, opts.Slice); diff != "" {
		t.Fatalf("unexpected slice defaults (-expected +actual):\n%s", diff)
	}

	opts.Map = map[string]int{"x": 9}
	opts.Slice = []int{9}

	_, err = parser.ParseArgs(nil)
	if err != nil {
		t.Fatalf("unexpected parse error on second pass: %v", err)
	}

	if diff := cmp.Diff(map[string]int{"x": 9}, opts.Map); diff != "" {
		t.Fatalf("non-empty map should be preserved (-expected +actual):\n%s", diff)
	}

	if diff := cmp.Diff([]int{9}, opts.Slice); diff != "" {
		t.Fatalf("non-empty slice should be preserved (-expected +actual):\n%s", diff)
	}
}

func TestRequiredFromValuesPrefilledSatisfiesRequired(t *testing.T) {
	var opts struct {
		Value string `json:"value,omitempty" long:"value" required:"yes"`
	}

	opts.Value = "from-config"

	parser := NewParser(&opts, RequiredFromValues)
	if _, err := parser.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Value != "from-config" {
		t.Fatalf("expected prefilled value to be preserved, got %q", opts.Value)
	}
}

func TestRequiredFromValuesEmptyStillFailsRequired(t *testing.T) {
	var opts struct {
		Value string `json:"value,omitempty" long:"value" required:"yes"`
	}

	parser := NewParser(&opts, RequiredFromValues)
	_, err := parser.ParseArgs(nil)
	if err == nil {
		t.Fatalf("expected required parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrRequired {
		t.Fatalf("expected ErrRequired, got %v", err)
	}
}

func TestConfiguredValuesAliasKeepsPrefilledAndSatisfiesRequired(t *testing.T) {
	var opts struct {
		Value string `json:"value,omitempty" long:"value" required:"yes" default:"from-tag"`
	}

	opts.Value = "from-config"

	parser := NewParser(&opts, ConfiguredValues)
	if _, err := parser.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Value != "from-config" {
		t.Fatalf("expected prefilled value to survive defaults, got %q", opts.Value)
	}
}

type CustomFlag struct {
	Value string
}

func (c *CustomFlag) UnmarshalFlag(s string) error {
	c.Value = s
	return nil
}

func (c *CustomFlag) IsValidValue(s string) error {
	if !(s == "-1" || s == "-foo") {
		return errors.New("invalid flag value")
	}
	return nil
}

func TestOptionAsArgument(t *testing.T) {
	var tests = []struct {
		args        []string
		expectError bool
		errType     ErrorType
		errMsg      string
		rest        []string
	}{
		{
			// short option must not be accepted as argument
			args:        []string{"--string-slice", "foobar", "--string-slice", "-o"},
			expectError: true,
			errType:     ErrExpectedArgument,
			errMsg:      "expected argument for flag `" + defaultLongOptDelimiter + "string-slice', but got option `-o'",
		},
		{
			// long option must not be accepted as argument
			args:        []string{"--string-slice", "foobar", "--string-slice", "--other-option"},
			expectError: true,
			errType:     ErrExpectedArgument,
			errMsg:      "expected argument for flag `" + defaultLongOptDelimiter + "string-slice', but got option `--other-option'",
		},
		{
			// long option must not be accepted as argument
			args:        []string{"--string-slice", "--"},
			expectError: true,
			errType:     ErrExpectedArgument,
			errMsg:      "expected argument for flag `" + defaultLongOptDelimiter + "string-slice', but got double dash `--'",
		},
		{
			// quoted and appended option should be accepted as argument (even if it looks like an option)
			args: []string{"--string-slice", "foobar", "--string-slice=\"--other-option\""},
		},
		{
			// Accept any single character arguments including '-'
			args: []string{"--string-slice", "-"},
		},
		{
			// Do not accept arguments which start with '-' even if the next character is a digit
			args:        []string{"--string-slice", "-3.14"},
			expectError: true,
			errType:     ErrExpectedArgument,
			errMsg:      "expected argument for flag `" + defaultLongOptDelimiter + "string-slice', but got option `-3.14'",
		},
		{
			// Do not accept arguments which start with '-' if the next character is not a digit
			args:        []string{"--string-slice", "-character"},
			expectError: true,
			errType:     ErrExpectedArgument,
			errMsg:      "expected argument for flag `" + defaultLongOptDelimiter + "string-slice', but got option `-character'",
		},
		{
			args: []string{"-o", "-", "-"},
			rest: []string{"-", "-"},
		},
		{
			// Accept arguments which start with '-' if the next character is a digit
			args: []string{"--int-slice", "-3"},
		},
		{
			// Accept arguments which start with '-' if the next character is a digit
			args: []string{"--int16", "-3"},
		},
		{
			// Accept arguments which start with '-' if the next character is a digit
			args: []string{"--float32", "-3.2"},
		},
		{
			// Accept arguments which start with '-' if the next character is a digit
			args: []string{"--float32ptr", "-3.2"},
		},
		{
			// Accept arguments for values that pass the IsValidValue fuction for value validators
			args: []string{"--custom-flag", "-foo"},
		},
		{
			// Accept arguments for values that pass the IsValidValue fuction for value validators
			args: []string{"--custom-flag", "-1"},
		},
		{
			// Rejects arguments for values that fail the IsValidValue fuction for value validators
			args:        []string{"--custom-flag", "-2"},
			expectError: true,
			errType:     ErrExpectedArgument,
			errMsg:      "invalid flag value",
		},
	}

	var opts struct {
		StringSlice []string   `long:"string-slice"`
		IntSlice    []int      `long:"int-slice"`
		Int16       int16      `long:"int16"`
		Float32     float32    `long:"float32"`
		Float32Ptr  *float32   `long:"float32ptr"`
		OtherOption bool       `long:"other-option" short:"o"`
		Custom      CustomFlag `long:"custom-flag" short:"c"`
	}

	for _, test := range tests {
		if test.expectError {
			assertParseFail(t, test.errType, test.errMsg, &opts, test.args...)
		} else {
			args := assertParseSuccess(t, &opts, test.args...)

			assertStringArray(t, args, test.rest)
		}
	}
}

func TestTerminatedOptions(t *testing.T) {
	type terminatedOpts struct {
		Slice         []int      `short:"s" long:"slice" terminator:"END"`
		MultipleSlice [][]string `short:"m" long:"multiple" terminator:";"`
		Bool          bool       `short:"v"`
	}

	tests := []struct {
		name                  string
		parserOpts            Options
		args                  []string
		wantSlice             []int
		wantMultipleSlice     [][]string
		wantBool              bool
		wantRest              []string
		wantErrContains       string
		wantErrContainsSecond string
	}{
		{
			name: "terminators usage",
			args: []string{
				"-s", "1", "2", "3", "END",
				"-m", "bin", "-xyz", "--foo", "bar", "-v", "foo bar", ";",
				"-v",
				"-m", "-xyz", "--foo",
			},
			wantSlice: []int{1, 2, 3},
			wantMultipleSlice: [][]string{
				{"bin", "-xyz", "--foo", "bar", "-v", "foo bar"},
				{"-xyz", "--foo"},
			},
			wantBool: true,
		},
		{
			name: "slice overwritten",
			args: []string{
				"-s", "1", "2", "END",
				"-s", "3", "4",
			},
			wantSlice: []int{3, 4},
		},
		{
			name: "terminator omitted for last option",
			args: []string{
				"-s", "1", "2", "3",
			},
			wantSlice: []int{1, 2, 3},
		},
		{
			name: "short names jumbled",
			args: []string{
				"-vm", "--foo", "-v", "bar", ";",
				"-s", "1", "2",
			},
			wantSlice:         []int{1, 2},
			wantMultipleSlice: [][]string{{"--foo", "-v", "bar"}},
			wantBool:          true,
		},
		{
			name: "terminator must be a token",
			args: []string{
				"-m", "--foo", "-v;", "-v",
			},
			wantMultipleSlice: [][]string{{"--foo", "-v;", "-v"}},
		},
		{
			name:       "double dash preserved inside terminated option",
			parserOpts: PassDoubleDash,
			args: []string{
				"-m", "--foo", "--", "bar", ";",
				"-v",
				"--", "--foo", "bar",
			},
			wantMultipleSlice: [][]string{{"--foo", "--", "bar"}},
			wantBool:          true,
			wantRest:          []string{"--foo", "bar"},
		},
		{
			name:                  "inline argument syntax rejected",
			args:                  []string{"-m=foo", "bar"},
			wantErrContains:       "terminated option flag",
			wantErrContainsSecond: "cannot use inline argument syntax",
		},
		{
			name:                  "inline argument syntax rejected for empty value",
			args:                  []string{"-m=", "foo"},
			wantErrContains:       "terminated option flag",
			wantErrContainsSecond: "cannot use inline argument syntax",
		},
		{
			name:              "no args",
			args:              []string{"-m", ";", "-s", "END"},
			wantMultipleSlice: [][]string{{}},
		},
		{
			name:              "no args without terminator",
			args:              []string{"-m"},
			wantMultipleSlice: [][]string{{}},
		},
		{
			name: "missing args in the middle",
			args: []string{
				"-m", "a", ";",
				"-m", ";",
				"-m", "b",
			},
			wantMultipleSlice: [][]string{{"a"}, {}, {"b"}},
		},
		{
			name:              "empty string argument",
			args:              []string{"-m", ""},
			wantMultipleSlice: [][]string{{""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := terminatedOpts{}
			parser := NewParser(&opts, tt.parserOpts)

			rest, err := parser.ParseArgs(tt.args)
			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatalf("expected parse error")
				}
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.wantErrContains, err.Error())
				}
				if tt.wantErrContainsSecond != "" && !strings.Contains(err.Error(), tt.wantErrContainsSecond) {
					t.Fatalf("expected error to contain %q, got %q", tt.wantErrContainsSecond, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if opts.Bool != tt.wantBool {
				t.Fatalf("expected Bool=%v, got %v", tt.wantBool, opts.Bool)
			}

			if diff := cmp.Diff(tt.wantSlice, opts.Slice, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("unexpected Slice (-expected +actual):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantMultipleSlice, opts.MultipleSlice, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("unexpected MultipleSlice (-expected +actual):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantRest, rest, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("unexpected rest args (-expected +actual):\n%s", diff)
			}
		})
	}
}

func TestTerminatedOptionInvalidTag(t *testing.T) {
	var opts struct {
		Invalid int `short:"t" terminator:"END"`
	}

	parser := NewParser(&opts, None)
	_, err := parser.ParseArgs(nil)

	if err == nil {
		t.Fatalf("expected parse error")
	}
	if !strings.Contains(err.Error(), "must be a slice or slice of slices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnknownFlagHandler(t *testing.T) {

	var opts struct {
		Flag1 string `long:"flag1"`
		Flag2 string `long:"flag2"`
	}

	p := NewParser(&opts, None)

	var unknownFlag1 string
	var unknownFlag2 bool
	var unknownFlag3 string

	// Set up a callback to intercept unknown options during parsing
	p.UnknownOptionHandler = func(option string, arg SplitArgument, args []string) ([]string, error) {
		if option == "unknownFlag1" {
			if argValue, ok := arg.Value(); ok {
				unknownFlag1 = argValue
				return args, nil
			}
			// consume a value from remaining args list
			unknownFlag1 = args[0]
			return args[1:], nil
		} else if option == "unknownFlag2" {
			// treat this one as a bool switch, don't consume any args
			unknownFlag2 = true
			return args, nil
		} else if option == "unknownFlag3" {
			if argValue, ok := arg.Value(); ok {
				unknownFlag3 = argValue
				return args, nil
			}
			// consume a value from remaining args list
			unknownFlag3 = args[0]
			return args[1:], nil
		}

		return args, fmt.Errorf("Unknown flag: %v", option)
	}

	// Parse args containing some unknown flags, verify that
	// our callback can handle all of them
	_, err := p.ParseArgs([]string{"--flag1=stuff", "--unknownFlag1", "blah", "--unknownFlag2", "--unknownFlag3=baz", "--flag2=foo"})

	if err != nil {
		assertErrorf(t, "Parser returned unexpected error %v", err)
	}

	assertString(t, opts.Flag1, "stuff")
	assertString(t, opts.Flag2, "foo")
	assertString(t, unknownFlag1, "blah")
	assertString(t, unknownFlag3, "baz")

	if !unknownFlag2 {
		assertErrorf(t, "Flag should have been set by unknown handler, but had value: %v", unknownFlag2)
	}

	// Parse args with unknown flags that callback doesn't handle, verify it returns error
	_, err = p.ParseArgs([]string{"--flag1=stuff", "--unknownFlagX", "blah", "--flag2=foo"})

	if err == nil {
		assertErrorf(t, "Parser should have returned error, but returned nil")
	}
}

func TestChoices(t *testing.T) {
	var opts struct {
		Choice string `long:"choose" choice:"v1" choice:"v2"`
	}

	assertParseFail(
		t,
		ErrInvalidChoice,
		"Invalid value `invalid' for option `"+
			defaultLongOptDelimiter+"choose'. Allowed values are: v1 or v2",
		&opts,
		"--choose",
		"invalid",
	)
	assertParseSuccess(t, &opts, "--choose", "v2")
	assertString(t, opts.Choice, "v2")
}

func TestEmbedded(t *testing.T) {
	type embedded struct {
		V bool `short:"v"`
	}
	var opts struct {
		embedded
	}

	assertParseSuccess(t, &opts, "-v")

	if !opts.V {
		t.Errorf("Expected V to be true")
	}
}

type command struct {
}

func (c *command) Execute(args []string) error {
	return nil
}

func TestCommandHandlerNoCommand(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	parser := NewParser(&opts, Default&^PrintErrors)

	var executedCommand Commander
	var executedArgs []string

	executed := false

	parser.CommandHandler = func(command Commander, args []string) error {
		executed = true

		executedCommand = command
		executedArgs = args

		return nil
	}

	_, err := parser.ParseArgs([]string{"arg1", "arg2"})

	if err != nil {
		t.Fatalf("Unexpected parse error: %s", err)
	}

	if !executed {
		t.Errorf("Expected command handler to be executed")
	}

	if executedCommand != nil {
		t.Errorf("Did not exect an executed command")
	}

	assertStringArray(t, executedArgs, []string{"arg1", "arg2"})
}

func TestCommandHandler(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Command command `command:"cmd"`
	}{}

	parser := NewParser(&opts, Default&^PrintErrors)

	var executedCommand Commander
	var executedArgs []string

	executed := false

	parser.CommandHandler = func(command Commander, args []string) error {
		executed = true

		executedCommand = command
		executedArgs = args

		return nil
	}

	_, err := parser.ParseArgs([]string{"cmd", "arg1", "arg2"})

	if err != nil {
		t.Fatalf("Unexpected parse error: %s", err)
	}

	if !executed {
		t.Errorf("Expected command handler to be executed")
	}

	if executedCommand == nil {
		t.Errorf("Expected command handler to be executed")
	}

	assertStringArray(t, executedArgs, []string{"arg1", "arg2"})
}

func TestAllowBoolValues(t *testing.T) {
	var tests = []struct {
		msg                string
		args               []string
		expectedErr        string
		expected           bool
		expectedNonOptArgs []string
	}{
		{
			msg:      "no value",
			args:     []string{"-v"},
			expected: true,
		},
		{
			msg:      "true value",
			args:     []string{"-v=true"},
			expected: true,
		},
		{
			msg:      "false value",
			args:     []string{"-v=false"},
			expected: false,
		},
		{
			msg:         "bad value",
			args:        []string{"-v=badvalue"},
			expectedErr: `parsing "badvalue": invalid syntax`,
		},
		{
			// this test is to ensure flag values can only be specified as --flag=value and not "--flag value".
			// if "--flag value" was supported it's not clear if value should be a non-optional argument
			// or the value for the flag.
			msg:                "validate flags can only be set with a value immediately following an assignment operator (=)",
			args:               []string{"-v", "false"},
			expected:           true,
			expectedNonOptArgs: []string{"false"},
		},
	}

	for _, test := range tests {
		var opts = struct {
			Value bool `short:"v"`
		}{}
		parser := NewParser(&opts, AllowBoolValues)
		nonOptArgs, err := parser.ParseArgs(test.args)

		if test.expectedErr == "" {
			if err != nil {
				t.Fatalf("%s:\nUnexpected parse error: %s", test.msg, err)
			}
			if opts.Value != test.expected {
				t.Errorf("%s:\nExpected %v; got %v", test.msg, test.expected, opts.Value)
			}
			if diff := cmp.Diff(test.expectedNonOptArgs, nonOptArgs, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("%s:\nUnexpected non-argument options (-expected +actual):\n%s", test.msg, diff)
			}
		} else {
			if err == nil {
				t.Errorf("%s:\nExpected error containing substring %q", test.msg, test.expectedErr)
			} else if !strings.Contains(err.Error(), test.expectedErr) {
				t.Errorf("%s:\nExpected error %q to contain substring %q", test.msg, err, test.expectedErr)
			}
		}
	}
}

func captureStdIO(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed creating stdout pipe: %v", err)
	}

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed creating stderr pipe: %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout, err := io.ReadAll(stdoutR)
	if err != nil {
		t.Fatalf("failed reading stdout: %v", err)
	}

	stderr, err := io.ReadAll(stderrR)
	if err != nil {
		t.Fatalf("failed reading stderr: %v", err)
	}

	return string(stdout), string(stderr)
}

func TestParseConvenienceFunctionsUseOSArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	os.Args = []string{"app", "--value=from-args"}

	var opts1 struct {
		Value string `long:"value"`
	}

	rest, err := Parse(&opts1)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(rest) != 0 {
		t.Fatalf("expected no remaining args, got %v", rest)
	}

	if opts1.Value != "from-args" {
		t.Fatalf("unexpected parsed value: %q", opts1.Value)
	}

	var opts2 struct {
		Value string `long:"value"`
	}
	parser := NewParser(&opts2, None)

	rest, err = parser.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(rest) != 0 {
		t.Fatalf("expected no remaining args, got %v", rest)
	}

	if opts2.Value != "from-args" {
		t.Fatalf("unexpected parsed value: %q", opts2.Value)
	}
}

func TestPrintErrorsDestination(t *testing.T) {
	var helpOpts struct {
		Value bool `long:"value"`
	}

	helpParser := NewParser(&helpOpts, Default)
	stdout, stderr := captureStdIO(t, func() {
		_, _ = helpParser.ParseArgs([]string{"--help"})
	})

	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected help on stdout, got %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("expected empty stderr for help, got %q", stderr)
	}

	var failOpts struct {
		Value bool `long:"value"`
	}

	failParser := NewParser(&failOpts, Default)
	stdout, stderr = captureStdIO(t, func() {
		_, _ = failParser.ParseArgs([]string{"--unknown"})
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout for parse error, got %q", stdout)
	}

	if !strings.Contains(stderr, "unknown flag") {
		t.Fatalf("expected parse error on stderr, got %q", stderr)
	}
}

func TestVersionFlagPrintsVersionAndOverridesOtherFlags(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	parser := NewNamedParser("version-test", VersionFlag|PrintErrors)
	_, _ = parser.AddGroup("Application Options", "", &opts)
	parser.SetVersion("v1.2.3")
	parser.SetVersionCommit("deadbeef")
	parser.SetVersionURL("https://example.test/repo")

	stdout, stderr := captureStdIO(t, func() {
		_, _ = parser.ParseArgs([]string{"--version", "--unknown"})
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr for version output, got %q", stderr)
	}
	if !strings.Contains(stdout, "version:  v1.2.3") {
		t.Fatalf("expected version output, got %q", stdout)
	}
	if strings.Contains(stdout, "unknown flag") {
		t.Fatalf("expected version to override other flags, got %q", stdout)
	}

	stdout, stderr = captureStdIO(t, func() {
		_, _ = parser.ParseArgs([]string{"--unknown", "--version"})
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr for version output, got %q", stderr)
	}
	if !strings.Contains(stdout, "version:  v1.2.3") {
		t.Fatalf("expected version output when version is after invalid flags, got %q", stdout)
	}
}

func TestVersionFlagSkipsRequiredChecks(t *testing.T) {
	var opts struct {
		ReleaseID string `long:"release-id" required:"yes"`
	}

	parser := NewNamedParser("version-test-required", VersionFlag|PrintErrors)
	_, _ = parser.AddGroup("Application Options", "", &opts)
	parser.SetVersion("v1.2.3")

	stdout, stderr := captureStdIO(t, func() {
		_, _ = parser.ParseArgs([]string{"--version"})
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr for version output, got %q", stderr)
	}
	if !strings.Contains(stdout, "version:  v1.2.3") {
		t.Fatalf("expected version output, got %q", stdout)
	}
	if strings.Contains(stdout, "required flag") {
		t.Fatalf("expected required checks to be skipped for version output, got %q", stdout)
	}
}

func TestHelpTakesPriorityOverVersion(t *testing.T) {
	parser := NewNamedParser("priority-test", HelpFlag|VersionFlag|PrintErrors)

	stdout, stderr := captureStdIO(t, func() {
		_, _ = parser.ParseArgs([]string{"--version", "--help"})
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected help output, got %q", stdout)
	}
	if strings.Contains(stdout, "version:") {
		t.Fatalf("expected help to win over version, got %q", stdout)
	}
}

func TestPrintErrorsHelpToStderr(t *testing.T) {
	var opts struct {
		Value bool `long:"value"`
	}

	parser := NewParser(&opts, Default|PrintHelpOnStderr)
	stdout, stderr := captureStdIO(t, func() {
		_, _ = parser.ParseArgs([]string{"--help"})
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout for help when PrintHelpOnStderr is set, got %q", stdout)
	}

	if !strings.Contains(stderr, "Usage:") {
		t.Fatalf("expected help on stderr, got %q", stderr)
	}
}

func TestPrintErrorsToStdout(t *testing.T) {
	var opts struct {
		Value bool `long:"value"`
	}

	parser := NewParser(&opts, Default|PrintErrorsOnStdout)
	stdout, stderr := captureStdIO(t, func() {
		_, _ = parser.ParseArgs([]string{"--unknown"})
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr when PrintErrorsOnStdout is set, got %q", stderr)
	}

	if !strings.Contains(stdout, "unknown flag") {
		t.Fatalf("expected parse error on stdout, got %q", stdout)
	}
}

func TestPrintErrorsInternalParserErrorDestination(t *testing.T) {
	type invalidOptions struct {
		One bool `long:"dup"`
		Two bool `long:"dup"`
	}

	stdout, stderr := captureStdIO(t, func() {
		p := NewParser(&invalidOptions{}, Default)
		_, _ = p.ParseArgs(nil)
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout for internal parser error with default options, got %q", stdout)
	}

	if !strings.Contains(stderr, "same long name") {
		t.Fatalf("expected internal parser error on stderr, got %q", stderr)
	}
}

func TestNoPrintErrorsInternalParserErrorSilent(t *testing.T) {
	type invalidOptions struct {
		One bool `long:"dup"`
		Two bool `long:"dup"`
	}

	stdout, stderr := captureStdIO(t, func() {
		p := NewParser(&invalidOptions{}, None)
		_, _ = p.ParseArgs(nil)
	})

	if stdout != "" || stderr != "" {
		t.Fatalf("expected no output when PrintErrors is disabled, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestSetTerminalTitleUsesParserName(t *testing.T) {
	p := NewNamedParser("title-from-name", SetTerminalTitle)

	called := false
	got := ""
	originalSetter := terminalTitleSetter
	terminalTitleSetter = func(title string) error {
		called = true
		got = title
		return nil
	}
	defer func() {
		terminalTitleSetter = originalSetter
	}()

	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !called {
		t.Fatalf("expected terminal title setter to be called")
	}

	if got != "title-from-name" {
		t.Fatalf("expected title-from-name, got %q", got)
	}
}

func TestSetTerminalTitleOverride(t *testing.T) {
	p := NewNamedParser("title-from-name", SetTerminalTitle)
	p.TerminalTitle = "custom title"

	called := false
	got := ""
	originalSetter := terminalTitleSetter
	terminalTitleSetter = func(title string) error {
		called = true
		got = title
		return nil
	}
	defer func() {
		terminalTitleSetter = originalSetter
	}()

	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !called {
		t.Fatalf("expected terminal title setter to be called")
	}

	if got != "custom title" {
		t.Fatalf("expected custom title, got %q", got)
	}
}

func TestSetTerminalTitleSkippedInCompletionMode(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	os.Setenv("GO_FLAGS_COMPLETION", "1")

	p := NewNamedParser("title-from-name", SetTerminalTitle)
	p.CompletionHandler = func(_ []Completion) {}

	called := false
	originalSetter := terminalTitleSetter
	terminalTitleSetter = func(_ string) error {
		called = true
		return nil
	}
	defer func() {
		terminalTitleSetter = originalSetter
	}()

	ret, err := p.ParseArgs(nil)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if ret != nil {
		t.Fatalf("expected nil result in completion mode, got %#v", ret)
	}

	if called {
		t.Fatalf("did not expect terminal title setter in completion mode")
	}
}

func TestExpectedTypeForFuncAndNonFunc(t *testing.T) {
	var opts struct {
		Do func() `long:"do"`
		S  string `long:"s"`
	}

	p := NewParser(&opts, None)

	doOpt := p.FindOptionByLongName("do")
	if doOpt == nil {
		t.Fatalf("option do not found")
	}

	if got := p.expectedType(doOpt); got != "" {
		t.Fatalf("expected empty type for func option, got %q", got)
	}

	sOpt := p.FindOptionByLongName("s")
	if sOpt == nil {
		t.Fatalf("option s not found")
	}

	if got := p.expectedType(sOpt); got != "string" {
		t.Fatalf("expected string type, got %q", got)
	}
}

func TestSetTagPrefix(t *testing.T) {
	var opts struct {
		Path string `flag-short:"p" flag-long:"path"`
	}

	p := NewParser(&opts, None)

	if err := p.SetTagPrefix("flag-"); err != nil {
		t.Fatalf("unexpected set prefix error: %v", err)
	}

	_, err := p.ParseArgs([]string{"-p", "tmp-path"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Path != "tmp-path" {
		t.Fatalf("expected tmp-path, got %q", opts.Path)
	}
}

func TestSetTagPrefixTerminator(t *testing.T) {
	var opts struct {
		Exec []string `flag-short:"e" flag-terminator:";"`
	}

	p := NewParser(&opts, None)

	if err := p.SetTagPrefix("flag-"); err != nil {
		t.Fatalf("unexpected set prefix error: %v", err)
	}

	if _, err := p.ParseArgs([]string{"-e", "echo", "hello", ";"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if diff := cmp.Diff([]string{"echo", "hello"}, opts.Exec, cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("unexpected terminated values (-expected +actual):\n%s", diff)
	}
}

func TestSetFlagTags(t *testing.T) {
	var opts struct {
		Path string `my-short:"p" long:"path"`
	}

	p := NewParser(&opts, None)
	tags := NewFlagTags()
	tags.Short = "my-short"

	if err := p.SetFlagTags(tags); err != nil {
		t.Fatalf("unexpected set tags error: %v", err)
	}

	_, err := p.ParseArgs([]string{"-p", "var-path"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Path != "var-path" {
		t.Fatalf("expected var-path, got %q", opts.Path)
	}
}

func TestDefaultsListTag(t *testing.T) {
	var opts struct {
		Slice []string `long:"slice" defaults:"one;two;three"`
	}

	_, err := ParseArgs(&opts, nil)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if diff := cmp.Diff([]string{"one", "two", "three"}, opts.Slice, cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("unexpected defaults from list tag (-expected +actual):\n%s", diff)
	}
}

func TestChoicesListTag(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" choices:"fast;safe"`
	}

	p := NewParser(&opts, None)

	if _, err := p.ParseArgs([]string{"--mode=safe"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Mode != "safe" {
		t.Fatalf("expected mode to be safe, got %q", opts.Mode)
	}

	_, err := p.ParseArgs([]string{"--mode=broken"})
	if err == nil {
		t.Fatalf("expected parse error for invalid choice")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidChoice {
		t.Fatalf("expected ErrInvalidChoice, got %v", err)
	}
}

func TestSetTagListDelimiter(t *testing.T) {
	var opts struct {
		Slice []int  `long:"slice" defaults:"1,2,3"`
		Mode  string `long:"mode" choices:"fast,safe"`
	}

	p := NewParser(&opts, None)

	if err := p.SetTagListDelimiter(','); err != nil {
		t.Fatalf("unexpected set delimiter error: %v", err)
	}

	if _, err := p.ParseArgs([]string{"--mode=safe"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if diff := cmp.Diff([]int{1, 2, 3}, opts.Slice, cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("unexpected defaults from custom delimiter (-expected +actual):\n%s", diff)
	}

	if opts.Mode != "safe" {
		t.Fatalf("expected mode to be safe, got %q", opts.Mode)
	}
}

func TestSetTagListDelimiterRejectsNUL(t *testing.T) {
	p := NewParser(nil, None)

	err := p.SetTagListDelimiter(0)
	if err == nil {
		t.Fatalf("expected error for NUL delimiter")
	}
}

func TestListTagNotRepeatable(t *testing.T) {
	var opts struct {
		Value string `long:"value" long-aliases:"one;two" long-aliases:"three;four"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestDefaultMaxLongNameLengthRejectsLongTag(t *testing.T) {
	var opts struct {
		Value string `long:"this-long-option-name-is-definitely-over-thirty-two-chars"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}

	if !strings.Contains(flagsErr.Message, "exceeds max length") {
		t.Fatalf("expected long-name length validation error, got: %s", flagsErr.Message)
	}
}

func TestSetMaxLongNameLengthAllowsLongTag(t *testing.T) {
	var opts struct {
		Value string `long:"this-long-option-name-is-definitely-over-thirty-two-chars"`
	}

	p := NewNamedParser("longname", None)
	if err := p.SetMaxLongNameLength(128); err != nil {
		t.Fatalf("unexpected set max long name length error: %v", err)
	}

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	if _, err := p.ParseArgs([]string{"--this-long-option-name-is-definitely-over-thirty-two-chars=value"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Value != "value" {
		t.Fatalf("expected parsed long option value, got %q", opts.Value)
	}
}

func TestSetMaxLongNameLengthZeroDisablesLimit(t *testing.T) {
	var opts struct {
		Value string `long:"this-long-option-name-is-definitely-over-thirty-two-chars"`
	}

	p := NewNamedParser("longname", None)
	if err := p.SetMaxLongNameLength(0); err != nil {
		t.Fatalf("unexpected set max long name length error: %v", err)
	}

	if _, err := p.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("unexpected add group error: %v", err)
	}

	if _, err := p.ParseArgs([]string{"--this-long-option-name-is-definitely-over-thirty-two-chars=value"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Value != "value" {
		t.Fatalf("expected parsed long option value, got %q", opts.Value)
	}
}

func TestDefaultAndDefaultsTagsConflict(t *testing.T) {
	var opts struct {
		Value string `long:"value" default:"one" defaults:"two;three"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestChoiceAndChoicesTagsConflict(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" choice:"fast" choices:"safe;strict"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestOptionLongAlias(t *testing.T) {
	var opts struct {
		NoCache bool `long:"nocache" long-alias:"no-cache"`
	}

	_, err := ParseArgs(&opts, []string{"--no-cache"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !opts.NoCache {
		t.Fatalf("expected NoCache to be true")
	}
}

func TestOptionLongAliasesList(t *testing.T) {
	var opts struct {
		NoCache bool `long:"nocache" long-aliases:"no-cache;no_cache"`
	}

	_, err := ParseArgs(&opts, []string{"--no_cache"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if !opts.NoCache {
		t.Fatalf("expected NoCache to be true")
	}
}

func TestOptionShortAlias(t *testing.T) {
	var opts struct {
		Count int `short:"c" short-alias:"C"`
	}

	_, err := ParseArgs(&opts, []string{"-C", "7"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Count != 7 {
		t.Fatalf("expected Count=7, got %d", opts.Count)
	}
}

func TestOptionShortAliasesList(t *testing.T) {
	var opts struct {
		Count int `short:"c" short-aliases:"x;X"`
	}

	_, err := ParseArgs(&opts, []string{"-X", "9"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Count != 9 {
		t.Fatalf("expected Count=9, got %d", opts.Count)
	}
}

func TestOptionAliasCollisions(t *testing.T) {
	var opts struct {
		One bool `long:"one" long-alias:"two"`
		Two bool `long:"two"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestOptionShortAliasTooLong(t *testing.T) {
	var opts struct {
		Value bool `short:"v" short-alias:"vv"`
	}

	_, err := ParseArgs(&opts, nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if flagsErr, ok := err.(*Error); !ok || flagsErr.Type != ErrShortNameTooLong {
		t.Fatalf("expected ErrShortNameTooLong, got %v", err)
	}
}

func TestSetTagPrefixOptionAliases(t *testing.T) {
	var opts struct {
		Path string `flag-long:"path" flag-long-alias:"p-ath" flag-short:"p" flag-short-alias:"P"`
	}

	p := NewParser(&opts, None)
	if err := p.SetTagPrefix("flag-"); err != nil {
		t.Fatalf("unexpected set prefix error: %v", err)
	}

	if _, err := p.ParseArgs([]string{"--p-ath", "ok"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Path != "ok" {
		t.Fatalf("expected parsed alias value, got %q", opts.Path)
	}

	if _, err := p.ParseArgs([]string{"-P", "ok2"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Path != "ok2" {
		t.Fatalf("expected parsed short alias value, got %q", opts.Path)
	}
}

type configureDefaultsOptions struct {
	Port int `long:"port"`
}

func (o *configureDefaultsOptions) ConfigureFlags(parser *Parser) error {
	opt := parser.FindOptionByLongName("port")
	if opt == nil {
		return errors.New("port option not found")
	}

	opt.Default = []string{"8080"}
	return nil
}

func TestConfigurerAppliedBeforeParse(t *testing.T) {
	var opts configureDefaultsOptions
	p := NewParser(&opts, None)

	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if opts.Port != 8080 {
		t.Fatalf("expected configured default 8080, got %d", opts.Port)
	}
}

type configureErrorOptions struct {
	Value string `long:"value"`
}

func (o *configureErrorOptions) ConfigureFlags(_ *Parser) error {
	return errors.New("configure failed")
}

func TestConfigurerErrorFailsParse(t *testing.T) {
	var opts configureErrorOptions
	p := NewParser(&opts, None)

	_, err := p.ParseArgs(nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	if !strings.Contains(err.Error(), "configure failed") {
		t.Fatalf("expected configure error message, got %v", err)
	}
}

type configureDuplicateOptions struct {
	One string `long:"one"`
	Two string `long:"two"`
}

func (o *configureDuplicateOptions) ConfigureFlags(parser *Parser) error {
	opt := parser.FindOptionByLongName("two")
	if opt == nil {
		return errors.New("two option not found")
	}

	opt.LongName = "one"
	return nil
}

func TestConfigurerDuplicateFlagsValidation(t *testing.T) {
	var opts configureDuplicateOptions
	p := NewParser(&opts, None)

	_, err := p.ParseArgs(nil)
	if err == nil {
		t.Fatalf("expected parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrDuplicatedFlag {
		t.Fatalf("expected ErrDuplicatedFlag, got %v", err)
	}
}

func TestParserValidate(t *testing.T) {
	var opts configureDefaultsOptions
	p := NewParser(&opts, None)

	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}
