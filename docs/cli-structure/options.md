# Options

Options are mapped struct fields that can be set from command-line flags,
defaults, environment variables, INI configuration, or runtime code.

An option is not a command, group, or positional argument container.
It is usually a scalar, slice, map, function, or custom value type.

## Basic Options

```go
type Options struct {
  Verbose bool   `short:"v" long:"verbose" description:"Show verbose output"`
  Output  string `short:"o" long:"output" value-name:"FILE"`
}
```

Boolean options do not require an argument.
Passing `--verbose` sets the field to `true`.

Non-boolean options require a value unless `optional` is set.
Values may be passed as `--output=file`, `--output file`, `-ofile`,
`-o=file`, or `-o file` depending on option style and platform rules.

## Optional Arguments

`optional` lets a non-bool option appear without a value.
Use `optional-value` to define what value is assigned in that case.

```go
type Options struct {
  Color string `long:"color" optional:"true" optional-value:"auto"`
}
```

Then `--color` stores `auto`, while `--color=always` stores `always`.

Optional values are convenient, but they can make command lines harder to read.
Prefer required values unless the bare option has a clear meaning.

## Repeated Values

Slices collect repeated option values.

```go
type Options struct {
  Include []string `short:"I" long:"include"`
}
```

`-I a -I b` stores `[]string{"a", "b"}`.

Maps collect key/value pairs. The default key/value delimiter is `:`.

```go
type Options struct {
  Label map[string]string `long:"label"`
}
```

`--label env:prod --label tier:api` stores two map entries.
Use `key-value-delimiter` when another separator is clearer.

Repeatable options can require a value count.
Use `required:"N"` to require at least `N` values,
or `required:"N-M"` to require a bounded range.

```go
type Options struct {
  Include []string `short:"I" long:"include" required:"1-"`
  Label   map[string]string `long:"label" required:"2-4"`
}
```

Scalar options still use boolean `required` values such as `true` or `yes`.
Numeric `required` values on scalar options are rejected,
except `1` and `0` keep their normal boolean meaning.

## Choices

Use `choices` when only a small closed set is valid.

```go
type Options struct {
  Format string `long:"format" choices:"json;yaml;text" default:"json"`
}
```

Choices improve three things at once:
parse validation, help output, and shell completion.

For dynamic sets, prefer a custom completer and application-level validation.
A tag should not pretend that a runtime-dependent list is static.

## Counters

A counter is an integer option that increments instead of replacing the value.

```go
type Options struct {
  Verbose int `short:"v" long:"verbose" counter:"true"`
}
```

The following forms are equivalent for a signed integer counter:

```bash
app -vvv
app -v -v -v
app --verbose=3
app -v=3
```

Counter values must be non-negative.
Counter fields must be integer or unsigned integer types.  
Use a slice of bools only when each occurrence needs to remain observable.
Use `counter` when only the final level matters.

## Option Relations

Use `xor` for mutually exclusive options.

```go
type Options struct {
  JSON bool `long:"json" xor:"format"`
  YAML bool `long:"yaml" xor:"format"`
}
```

`--json --yaml` fails because both options belong to the same `xor` group.

Use `and` when options only make sense together.

```go
type Options struct {
  User string `long:"user" and:"auth"`
  Pass string `long:"pass" and:"auth"`
}
```

`--user root` fails unless `--pass` is also set.
Passing neither is allowed unless one member is also `required`.

Relation groups are local to the command where options are defined.
They are not global application-level locks.

## Hidden and Immediate Options

`hidden` keeps an option usable but removes it from help, completion, and docs.
Use it for compatibility aliases and internal switches.

`immediate` marks an option that should stop normal validation and execution.
This is for help, version, completion, generated docs, and similar flows.

Do not use `immediate` for ordinary business logic.
It changes parser control flow and should stay rare.

## Pointers

Pointers preserve the difference between omitted input and explicit zero values.

```go
type Options struct {
  Limit *int `long:"limit"`
}
```

If `--limit` is omitted, `Limit` stays nil.
If `--limit=0` is passed, `Limit` points to zero.

Use pointers sparingly. They add nil checks throughout application code.
For most options, a default value is simpler.

## Custom Values

Custom value types can implement package conversion interfaces,
or standard encoding interfaces such as `encoding.TextUnmarshaler`.

Use custom conversion when the CLI value has a domain-specific syntax.
Use validators when the value is still a normal string or number,
but needs a simple constraint.

## Defaults and Required Checks

Defaults and environment values are applied before required checks.
This means a required option can be satisfied without appearing on the command
line if a configured default source provides it.

This is intentional. `required` means the application must receive a value,
not necessarily that the user must type the flag in every invocation.

## Option Design Rules

Keep command-line option names stable. Breaking names breaks scripts.

Prefer long names for all public options.
Short names are valuable for frequent interactive use,
but they are a limited namespace.

Keep option tags close to the field.
If a field needs many tags, consider whether the option is doing too much.
