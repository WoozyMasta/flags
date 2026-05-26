// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"reflect"
	"slices"
	"testing"
)

// flagArg builds a platform-correct "flag=value" / "flag:value" token.
func flagArg(name, value string) string {
	return defaultLongOptDelimiter + name + string(defaultNameArgDelimiter) + value
}

// flagPrefix returns just the flag prefix used for prefix-matching assertions.
func flagPrefix(name string) string {
	return defaultLongOptDelimiter + name + string(defaultNameArgDelimiter)
}

// mustExportArgs calls ExportArgs and fails the test on error.
func mustExportArgs(t *testing.T, p *Parser, opts ...ExportArgsOption) []string {
	t.Helper()

	args, err := p.ExportArgs(opts...)
	if err != nil {
		t.Fatalf("ExportArgs returned unexpected error: %v", err)
	}

	return args
}

func mustParseArgs(t *testing.T, p *Parser, args ...string) {
	t.Helper()

	if _, err := p.ParseArgs(args); err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
}

func assertContainsPrefix(t *testing.T, args []string, prefix string) {
	t.Helper()

	for _, a := range args {
		if len(a) >= len(prefix) && a[:len(prefix)] == prefix {
			return
		}
	}

	t.Errorf("expected a token with prefix %q in args %v", prefix, args)
}

func assertNotContainsPrefix(t *testing.T, args []string, prefix string) {
	t.Helper()

	for _, a := range args {
		if len(a) >= len(prefix) && a[:len(prefix)] == prefix {
			t.Errorf("did not expect a token with prefix %q in args %v", prefix, args)
			return
		}
	}
}

func TestExportArgsBoolResolved(t *testing.T) {
	var opts struct {
		Verbose bool `long:"verbose"`
		Quiet   bool `long:"quiet"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"verbose")

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("verbose", "true"))
	// quiet defaults to false - Resolved always includes bools
	assertContains(t, args, flagArg("quiet", "false"))
}

func TestExportArgsBoolNonZero(t *testing.T) {
	var opts struct {
		Verbose bool `long:"verbose"`
		Quiet   bool `long:"quiet"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"verbose")

	args := mustExportArgs(t, p, WithMode(ExportModeNonZero))

	assertContains(t, args, flagArg("verbose", "true"))
	// quiet=false is zero - must be absent
	assertNotContainsPrefix(t, args, flagPrefix("quiet"))
}

func TestExportArgsBoolExplicit(t *testing.T) {
	var opts struct {
		Verbose bool   `long:"verbose"`
		Name    string `long:"name" default:"default-name"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"verbose")

	args := mustExportArgs(t, p, WithMode(ExportModeExplicit))

	// verbose was explicitly set
	assertContains(t, args, flagArg("verbose", "true"))
	// name was set via default tag - must be absent in Explicit mode
	assertNotContainsPrefix(t, args, flagPrefix("name"))
}

func TestExportArgsScalarString(t *testing.T) {
	var opts struct {
		Name string `long:"name"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"name", "hello")

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("name", "hello"))
}

func TestExportArgsScalarInt(t *testing.T) {
	var opts struct {
		Count int `long:"count"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"count", "42")

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("count", "42"))
}

func TestExportArgsSliceInOrder(t *testing.T) {
	var opts struct {
		Tags []string `long:"tag"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"tag", "c",
		defaultLongOptDelimiter+"tag", "a",
		defaultLongOptDelimiter+"tag", "b",
	)

	args := mustExportArgs(t, p)

	tagC := flagArg("tag", "c")
	tagA := flagArg("tag", "a")
	tagB := flagArg("tag", "b")

	assertContains(t, args, tagC)
	assertContains(t, args, tagA)
	assertContains(t, args, tagB)

	// verify input order is preserved
	ci, ai, bi := slices.Index(args, tagC), slices.Index(args, tagA), slices.Index(args, tagB)
	if ci >= ai || ai >= bi {
		t.Errorf("slice order not preserved: c=%d a=%d b=%d in %v", ci, ai, bi, args)
	}
}

func TestExportArgsEmptySliceSkipped(t *testing.T) {
	var opts struct {
		Tags []string `long:"tag"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p)

	args := mustExportArgs(t, p)

	assertNotContainsPrefix(t, args, flagPrefix("tag"))
}

func TestExportArgsMapSortedKeys(t *testing.T) {
	var opts struct {
		Env map[string]string `long:"env"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"env", "Z:last",
		defaultLongOptDelimiter+"env", "A:first",
		defaultLongOptDelimiter+"env", "M:mid",
	)

	args := mustExportArgs(t, p)

	// collect env= tokens
	pfx := flagPrefix("env")
	var envArgs []string

	for _, a := range args {
		if len(a) > len(pfx) && a[:len(pfx)] == pfx {
			envArgs = append(envArgs, a)
		}
	}

	if len(envArgs) != 3 {
		t.Fatalf("expected 3 env tokens, got %v", envArgs)
	}

	if envArgs[0] != flagArg("env", "A:first") ||
		envArgs[1] != flagArg("env", "M:mid") ||
		envArgs[2] != flagArg("env", "Z:last") {
		t.Errorf("map keys not sorted: %v", envArgs)
	}
}

func TestExportArgsMapCustomDelimiter(t *testing.T) {
	var opts struct {
		Env map[string]string `long:"env" key-value-delimiter:"="`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"env", "key=val")

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("env", "key=val"))
}

func TestExportArgsShortOnlyFallback(t *testing.T) {
	if defaultShortOptDelimiter != '-' {
		t.Skip("short-only fallback test requires '-' short delimiter")
	}

	var opts struct {
		Level int `short:"l"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, "-l", "3")

	args := mustExportArgs(t, p)

	assertContains(t, args, "-l=3")
}

func TestExportArgsSecretExcludedByDefault(t *testing.T) {
	var opts struct {
		Password string `long:"password" secret:"true"`
		Name     string `long:"name"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"password", "s3cret",
		defaultLongOptDelimiter+"name", "alice",
	)

	args := mustExportArgs(t, p)

	assertNotContainsPrefix(t, args, flagPrefix("password"))
	assertContains(t, args, flagArg("name", "alice"))
}

func TestExportArgsSecretIncluded(t *testing.T) {
	var opts struct {
		Password string `long:"password" secret:"true"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"password", "s3cret")

	args := mustExportArgs(t, p, WithSecrets(ExportSecretsInclude))

	assertContains(t, args, flagArg("password", "s3cret"))
}

func TestExportArgsHiddenExcludedByDefault(t *testing.T) {
	var opts struct {
		Hidden string `long:"hidden" hidden:"true"`
		Name   string `long:"name"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"hidden", "secret",
		defaultLongOptDelimiter+"name", "bob",
	)

	args := mustExportArgs(t, p)

	assertNotContainsPrefix(t, args, flagPrefix("hidden"))
	assertContains(t, args, flagArg("name", "bob"))
}

func TestExportArgsHiddenIncluded(t *testing.T) {
	var opts struct {
		Hidden string `long:"hidden" hidden:"true"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"hidden", "val")

	args := mustExportArgs(t, p, WithHidden(true))

	assertContains(t, args, flagArg("hidden", "val"))
}

func TestExportArgsDeprecatedIncludedByDefault(t *testing.T) {
	var opts struct {
		Old string `long:"old" deprecated:"use --new instead"`
		New string `long:"new"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"old", "value")

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("old", "value"))
}

func TestExportArgsDeprecatedExcluded(t *testing.T) {
	var opts struct {
		Old string `long:"old" deprecated:"use --new instead"`
		New string `long:"new"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"old", "value")

	args := mustExportArgs(t, p, WithIncludeDeprecated(false))

	assertNotContainsPrefix(t, args, flagPrefix("old"))
}

func TestExportArgsNamespace(t *testing.T) {
	var opts struct {
		Network struct {
			Host string `long:"host"`
			Port int    `long:"port"`
		} `group:"network" namespace:"net" env-namespace:"NET"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	// namespace delimiter defaults to "." so option names are "net.host", "net.port"
	nd := p.NamespaceDelimiter

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"net"+nd+"host", "localhost",
		defaultLongOptDelimiter+"net"+nd+"port", "8080",
	)

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("net"+nd+"host", "localhost"))
	assertContains(t, args, flagArg("net"+nd+"port", "8080"))
}

func TestExportArgsSubcommandActiveChain(t *testing.T) {
	var rootOpts struct {
		Verbose bool `long:"verbose"`
	}

	var serverOpts struct {
		Port int `long:"port"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("root", "", &rootOpts) //nolint

	_, err := p.AddCommand("server", "server command", "", &serverOpts)
	if err != nil {
		t.Fatal(err)
	}

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"verbose",
		"server",
		defaultLongOptDelimiter+"port", "9090",
	)

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("verbose", "true"))
	assertContains(t, args, "server")
	assertContains(t, args, flagArg("port", "9090"))

	// "server" token must appear before its options
	serverIdx := slices.Index(args, "server")
	portIdx := slices.Index(args, flagArg("port", "9090"))

	if serverIdx < 0 || portIdx < 0 || serverIdx >= portIdx {
		t.Errorf("expected 'server' before port flag: args=%v", args)
	}
}

func TestExportArgsWithCommandPath(t *testing.T) {
	var rootOpts struct {
		Config string `long:"config"`
	}

	var deployOpts struct {
		Force bool `long:"force"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("root", "", &rootOpts) //nolint

	_, err := p.AddCommand("deploy", "deploy command", "", &deployOpts)
	if err != nil {
		t.Fatal(err)
	}

	// set deploy option directly (no parsing needed for WithCommandPath target)
	deployOpts.Force = true

	args := mustExportArgs(t, p, WithCommandPath("deploy"))

	assertContains(t, args, "deploy")
	assertContains(t, args, flagArg("force", "true"))
}

func TestExportArgsWithCommandPathUnknownReturnsError(t *testing.T) {
	p := NewNamedParser("test", None)

	var opts struct{}
	p.AddGroup("opts", "", &opts) //nolint

	_, err := p.ExportArgs(WithCommandPath("nonexistent"))
	if err == nil {
		t.Fatal("expected error for unknown command path, got nil")
	}

	flagErr, ok := err.(*Error)
	if !ok || flagErr.Type != ErrUnknownCommand {
		t.Errorf("expected ErrUnknownCommand, got %v", err)
	}
}

func TestExportArgsNestedSubcommands(t *testing.T) {
	var rootOpts struct {
		Verbose bool `long:"verbose"`
	}

	var serverOpts struct {
		Host string `long:"host"`
	}

	var runOpts struct {
		Workers int `long:"workers"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("root", "", &rootOpts) //nolint

	serverCmd, err := p.AddCommand("server", "server", "", &serverOpts)
	if err != nil {
		t.Fatal(err)
	}

	_, err = serverCmd.AddCommand("run", "run", "", &runOpts)
	if err != nil {
		t.Fatal(err)
	}

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"verbose",
		"server",
		defaultLongOptDelimiter+"host", "127.0.0.1",
		"run",
		defaultLongOptDelimiter+"workers", "4",
	)

	args := mustExportArgs(t, p)

	assertContains(t, args, flagArg("verbose", "true"))
	assertContains(t, args, "server")
	assertContains(t, args, flagArg("host", "127.0.0.1"))
	assertContains(t, args, "run")
	assertContains(t, args, flagArg("workers", "4"))

	serverIdx := slices.Index(args, "server")
	runIdx := slices.Index(args, "run")

	if serverIdx < 0 || runIdx < 0 || serverIdx >= runIdx {
		t.Errorf("expected 'server' before 'run': args=%v", args)
	}
}

func TestExportArgsDeterministic(t *testing.T) {
	var opts struct {
		Env  map[string]string `long:"env"`
		Tags []string          `long:"tag"`
		Name string            `long:"name"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"name", "app",
		defaultLongOptDelimiter+"tag", "b",
		defaultLongOptDelimiter+"tag", "a",
		defaultLongOptDelimiter+"env", "Z:last",
		defaultLongOptDelimiter+"env", "A:first",
	)

	first := mustExportArgs(t, p)
	second := mustExportArgs(t, p)
	third := mustExportArgs(t, p)

	if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(second, third) {
		t.Errorf("ExportArgs is not deterministic:\nfirst:  %v\nsecond: %v\nthird:  %v", first, second, third)
	}
}

func TestExportArgsRoundTripScalars(t *testing.T) {
	var opts struct {
		Name  string `long:"name"`
		Count int    `long:"count"`
		Debug bool   `long:"debug"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"name", "hello",
		defaultLongOptDelimiter+"count", "7",
		defaultLongOptDelimiter+"debug",
	)

	exported := mustExportArgs(t, p)

	var opts2 struct {
		Name  string `long:"name"`
		Count int    `long:"count"`
		Debug bool   `long:"debug"`
	}

	p2 := NewNamedParser("test", AllowBoolValues)
	p2.AddGroup("opts", "", &opts2) //nolint

	if _, err := p2.ParseArgs(exported); err != nil {
		t.Fatalf("re-parse failed: %v (args=%v)", err, exported)
	}

	if opts2.Name != "hello" || opts2.Count != 7 || !opts2.Debug {
		t.Errorf("round-trip mismatch: %+v", opts2)
	}
}

func TestExportArgsRoundTripSlice(t *testing.T) {
	var opts struct {
		Tags []string `long:"tag"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"tag", "x",
		defaultLongOptDelimiter+"tag", "y",
		defaultLongOptDelimiter+"tag", "z",
	)

	exported := mustExportArgs(t, p)

	var opts2 struct {
		Tags []string `long:"tag"`
	}

	p2 := NewNamedParser("test", None)
	p2.AddGroup("opts", "", &opts2) //nolint

	if _, err := p2.ParseArgs(exported); err != nil {
		t.Fatalf("re-parse failed: %v (args=%v)", err, exported)
	}

	if !reflect.DeepEqual(opts2.Tags, []string{"x", "y", "z"}) {
		t.Errorf("round-trip slice mismatch: %v", opts2.Tags)
	}
}

func TestExportArgsRoundTripMap(t *testing.T) {
	var opts struct {
		Env map[string]string `long:"env"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"env", "KEY1:val1",
		defaultLongOptDelimiter+"env", "KEY2:val2",
	)

	exported := mustExportArgs(t, p)

	var opts2 struct {
		Env map[string]string `long:"env"`
	}

	p2 := NewNamedParser("test", None)
	p2.AddGroup("opts", "", &opts2) //nolint

	if _, err := p2.ParseArgs(exported); err != nil {
		t.Fatalf("re-parse failed: %v (args=%v)", err, exported)
	}

	want := map[string]string{"KEY1": "val1", "KEY2": "val2"}
	if !reflect.DeepEqual(opts2.Env, want) {
		t.Errorf("round-trip map mismatch: got %v, want %v", opts2.Env, want)
	}
}

func TestExportArgsRoundTripSubcommand(t *testing.T) {
	var rootOpts struct {
		Verbose bool `long:"verbose"`
	}

	var deployOpts struct {
		Env   string `long:"env"`
		Force bool   `long:"force"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("root", "", &rootOpts)           //nolint
	p.AddCommand("deploy", "", "", &deployOpts) //nolint

	mustParseArgs(t, p,
		defaultLongOptDelimiter+"verbose",
		"deploy",
		defaultLongOptDelimiter+"env", "prod",
		defaultLongOptDelimiter+"force",
	)

	exported := mustExportArgs(t, p)

	var rootOpts2 struct {
		Verbose bool `long:"verbose"`
	}
	var deployOpts2 struct {
		Env   string `long:"env"`
		Force bool   `long:"force"`
	}

	p2 := NewNamedParser("test", AllowBoolValues)
	p2.AddGroup("root", "", &rootOpts2)           //nolint
	p2.AddCommand("deploy", "", "", &deployOpts2) //nolint

	if _, err := p2.ParseArgs(exported); err != nil {
		t.Fatalf("round-trip re-parse failed: %v (args=%v)", err, exported)
	}

	if !rootOpts2.Verbose {
		t.Errorf("round-trip: expected verbose=true")
	}

	if deployOpts2.Env != "prod" || !deployOpts2.Force {
		t.Errorf("round-trip deploy opts mismatch: %+v", deployOpts2)
	}
}

func TestExportArgsExplicitSkipsDefault(t *testing.T) {
	var opts struct {
		Name string `long:"name" default:"world"`
		Set  string `long:"set"`
	}

	p := NewNamedParser("test", None)
	p.AddGroup("opts", "", &opts) //nolint

	mustParseArgs(t, p, defaultLongOptDelimiter+"set", "explicit")

	args := mustExportArgs(t, p, WithMode(ExportModeExplicit))

	assertContains(t, args, flagArg("set", "explicit"))
	assertNotContainsPrefix(t, args, flagPrefix("name"))
}
