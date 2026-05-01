# Parsing Rules

This page describes the command-line syntax accepted by the parser.
It focuses on how tokens are consumed before values reach conversion,
validation, commands, or application code.

## Long Options

Long options use `--name`.

```bash
app --verbose
app --output file.txt
app --output=file.txt
```

A long bool option does not require a value.
A non-bool long option needs a value unless `optional` is set.

Prefer `--name=value` when the value might look like another option.
It removes ambiguity for humans and parsers.

## Short Options

Short options use `-x`.

```bash
app -v
app -o file.txt
app -ofile.txt
app -o=file.txt
```

A short bool can be combined with other short bools.

```bash
app -abc
```

This is parsed as `-a -b -c` when each option can be used without an argument.

## Counters and Repeated Short Flags

A counter short option supports compact repetition.

```bash
app -vvv
app -v -v -v
app -v=3
```

For a counter, these forms add to the numeric value.  
For a slice, repeated options append values.

Use a counter when only the final level matters.  
Use a slice when every occurrence matters.

## Option Values That Look Like Options

A value that starts with `-` can be ambiguous.
Signed numeric options accept negative values when the target type allows them.
Other option values may be interpreted as a new option.

Use an explicit separator to avoid ambiguity:

```bash
app --pattern=-x
app -p=-x
```

Custom `ValueValidator` can also allow a value before normal conversion when a
type owns special syntax.

## Bool Values

By default, bool options are switches.

```bash
app --verbose
```

Passing a value to a bool option fails unless `AllowBoolValues` is enabled.

```go
parser := flags.NewParser(&opts, flags.Default|flags.AllowBoolValues)
```

Then values such as `--verbose=true` are accepted.
Use this only when explicit bool values are part of the CLI style.

## Optional Option Values

`optional:"true"` lets an option appear without a value.
`optional-value` defines what is stored when the value is omitted.

```bash
app --color
app --color=always
```

Prefer `--color=value` for the valued form.
Optional values can make `--color next-token` ambiguous.

## Double Dash

With `PassDoubleDash`, `--` stops option parsing.
The remaining tokens are returned as rest args.

```bash
app --verbose -- --not-a-flag value
```

This is useful for wrapper commands, subprocess arguments,
and values that must not be parsed by the outer CLI.

`Default` includes `PassDoubleDash`.

## Unknown Options

By default, unknown options fail with `ErrUnknownFlag`.

`IgnoreUnknown` returns unknown options as remaining args instead.
Use it for partial parsers and wrappers.
Avoid it for strict CLIs where typos should be errors.

`UnknownOptionHandler` can implement custom handling.
It receives the unknown option, possible split value, and remaining args.
It can return a modified arg list or an error.

See [Handlers and Integration Points][] when unknown options need rewriting
instead of simple pass-through.

## Commands and Option Scope

Global options belong to the root parser scope.
Command-local options belong to the selected command scope.

A global option is valid before or after the command token.
A command-local option is valid after that command is selected.

```bash
app --verbose run
app run --verbose
```

Both are valid only if `--verbose` is global
or belongs to `run` after `run` is selected.

Sibling commands can reuse option names. Those scopes do not conflict.

## Pass After Non-Option

`PassAfterNonOption` stops option parsing after the first non-option token.
This is strict POSIX-style behavior.

The command tag `pass-after-non-option`
applies the behavior to one command scope.
It is better for wrapper commands than enabling the parser-wide option.

## Terminator Options

`terminator` makes an option consume values until a terminator token.

```go
type Options struct {
  Exec []string `long:"exec" terminator:";"`
}
```

```bash
app --exec echo hello ';' --verbose
```

The terminator is not stored in the value.
After the terminator, normal parsing resumes.

## Windows Option Style

On Windows,
the parser supports slash-prefixed options such as `/v` and `/verbose`.
Windows-style value delimiters may also use `:`.

The `forceposix` build tag disables Windows-style parsing behavior for builds
that need POSIX behavior everywhere.

Render-style options affect help/docs presentation.
They do not change parser semantics.

## Token Design Rules

Use long options in scripts.  
Use short options for frequent interactive flags.

Use `--` for subprocess or wrapper arguments.  
Use explicit `--name=value` when values can be confused with options.

Do not rely on ambiguous token order.
A CLI that is easy to parse is usually easier for users too.

[Handlers and Integration Points]: handlers.md
