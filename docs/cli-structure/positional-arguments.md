# Positional Arguments

Positional arguments are values identified by order instead of option name.
They are useful when the command grammar is naturally short,
for example `copy SRC DST` or `grep PATTERN FILE`.

In `flags`, positional arguments are usually declared through
a nested struct marked with `positional-args`.

## Basic Form

```go
type Options struct {
  Args struct {
    Input  string `positional-arg-name:"input" description:"Input file"`
    Output string `positional-arg-name:"output" description:"Output file"`
  } `positional-args:"yes" required:"yes"`
}
```

Fields inside the positional container are consumed in declaration order.
The container belongs to the command scope where it is declared.

## Required Rules

For scalar option fields,
`required` is a boolean tag.
For positional argument fields and repeatable option fields,
the parser also accepts numeric forms.

`required:"yes"` means the positional value is required.

For a trailing slice, `required:"N"` means at least `N` values.

For a trailing slice,
`required:"N-M"` means from `N` to `M` values.
Use `required:"N-"` for at least `N` values without an upper bound.

```go
type Options struct {
  Args struct {
    Pattern string   `required:"yes"`
    Files   []string `required:"1-"`
  } `positional-args:"yes"`
}
```

Do not add `required:"0-"` to an optional rest slice.
A trailing slice without `required` may already be empty.

Use a slice only at the end of the positional list.
A slice consumes remaining positional values.

## Defaults

`default` and `defaults` can provide fallback values for positional fields.
Defaults are applied before validation.

This is useful for commands where the common target is obvious,
but callers can still override it explicitly.

Do not use defaults to hide surprising behavior.
A positional default should be visible in help or documented near the command.

## I/O Positionals

I/O templates are especially useful for positional arguments.
They can model common CLI conventions:

```go
type Options struct {
  IO struct {
    Input  string `io:"in" io-kind:"auto"`
    Output string `io:"out" io-kind:"auto"`
  } `positional-args:"yes"`
}
```

For this form:

* omitted input becomes `stdin`;
* omitted output becomes `stdout`;
* `-` is normalized to the configured stream;
* file paths stay as paths in `auto` mode.

The parser normalizes values.
It does not open files.
Application code decides whether `stdin`, `stdout`, or a path becomes an
actual `io.Reader` or `io.Writer`.

## Validation

Validation tags can be placed on positional fields.

```go
type Options struct {
  Args struct {
    Input string `validate-existing-file:"true" validate-readable:"true"`
  } `positional-args:"yes" required:"yes"`
}
```

Validators run after positional values are assigned.
For slices, each element is checked.

Use validators for stable input rules.  
Use application code for rules that depend
on external state beyond basic file or value checks.

## Completion

`completion` can be used on positional string fields.

```go
type Options struct {
  Args struct {
    Config string `completion:"file"`
  } `positional-args:"yes"`
}
```

Supported hints are `file`, `dir`, and `none`.

I/O positional fields with `io-kind:"file"` or `io-kind:"auto"` imply file
completion when no explicit completion hint is set.

## Command-Local Positionals

Each command can have its own positional container.

```go
type Options struct {
  Add struct {
    Args struct {
      Name string
    } `positional-args:"yes" required:"yes"`
  } `command:"add"`
}
```

`app add item` fills `Add.Args.Name`.
A sibling command does not see that positional definition.

## Positionals vs Options

Use positional arguments for values that users naturally remember by order.
Use named options when the meaning is not obvious.

Two or three positional values are usually readable.
Long positional grammars become fragile,
especially when some values are optional.

If a command mixes optional positionals and many options,
prefer explicit options for the optional values.
