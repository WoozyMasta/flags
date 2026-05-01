# Testing

Parser tests should focus on behavior first and presentation second.
Use exact output snapshots only where formatting is the behavior being tested.

## Use ParseArgs

Use `ParseArgs` in tests. It avoids depending on process `os.Args`.

```go
parser := flags.NewParser(&opts, flags.Default&^flags.PrintErrors)
_, err := parser.ParseArgs([]string{"--name", "alice"})
```

Disable `PrintErrors` so tests do not write parser output unless
the test is explicitly checking output routing.

## Assert Error Types

Parser errors are `*flags.Error`. Assert `Type` before checking text.

```go
var ferr *flags.Error
if !errors.As(err, &ferr) {
  t.Fatalf("expected *flags.Error, got %T", err)
}
if ferr.Type != flags.ErrValidation {
  t.Fatalf("expected validation error, got %s", ferr.Type)
}
```

Exact message tests are appropriate for parser diagnostics.
Application tests should usually prefer type checks or substrings.

## Help and Version Control Flow

`ErrHelp` and `ErrVersion` are successful control flow for applications.
Test them as such.

```go
_, err := parser.ParseArgs([]string{"--help"})
var ferr *flags.Error
if !errors.As(err, &ferr) || ferr.Type != flags.ErrHelp {
  t.Fatalf("expected help, got %v", err)
}
```

If `PrintErrors` is disabled,
help text is returned in the error message but not printed automatically.

## Environment Tests

Restore environment variables after each test.
Use `t.Setenv` when available.

```go
t.Setenv("APP_PORT", "9000")
```

Test explicit `env` behavior separately from `EnvProvisioning` behavior.
Auto-derived names are part of the public config contract.

## INI Round Trips

For INI behavior, use `strings.NewReader` and `bytes.Buffer` where possible.
Only use temp files when testing file APIs.

```go
ini := flags.NewIniParser(parser)
err := ini.Parse(strings.NewReader("[Application Options]\nport = 8080\n"))
```

When testing generated examples, assert important fragments
instead of every byte unless formatting stability is the actual requirement.

## Completion Tests

Completion can be tested through raw completion mode or a custom
`CompletionHandler`.

Use `t.Setenv("GO_FLAGS_COMPLETION", "1")` for parser-level completion paths.
Use a custom handler when you want
to capture candidates without process exit behavior.

Avoid tests that depend on real filesystem layout unless
the test creates a temporary directory and controls the contents.

## Golden Help and Docs

Golden tests are useful for:

* built-in help layout;
* generated markdown;
* generated HTML;
* man output;
* custom templates.

Keep golden inputs small.
Large snapshots are hard to review and easy to update blindly.

Use explicit render style and help width in golden tests.
Do not let terminal size, OS, or shell detection change expected output.

```go
parser.SetHelpWidth(80)
parser.SetHelpFlagRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpEnvRenderStyle(flags.RenderStylePOSIX)
```

## Validation Tests

For validators, test both valid and invalid values.
Also test invalid tag usage when adding a new validator.

A good validator test set includes:

* one success case;
* one parse failure case;
* one invalid-tag case;
* one slice case if slices are supported.

## Command Execution Tests

Keep parse tests separate from command execution tests where possible.

For `Execute`, use small command structs that record calls.
For `CommandChain`, assert parent-to-leaf order and stop-on-error behavior.

When using `CommandHandler`, test whether the handler calls or suppresses
`Execute` intentionally.

## Race and Side Effects

Parser construction and parsing should stay cheap and predictable.
Avoid tests that depend on global mutable state unless the state is restored.

Completion, help, version, and docs paths should not perform application work.
Test this for commands with immediate options or built-in helper commands.

## Test Selection Rules

Prefer small table tests for parser rules.  
Prefer explicit named tests for complex command trees.

When a regression is about user-visible text,
assert the exact text near the feature.  
When a regression is about behavior,
assert the structured result.
