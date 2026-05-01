# Getting Started

`flags` builds command-line parsers from Go structs.
Struct fields become options, commands, groups,
or positional arguments based on tags.

The core workflow is small:

1. Define a struct that models the CLI.
1. Create a parser with `NewParser` or `NewNamedParser`.
1. Call `Parse` or `ParseArgs`.
1. Read populated struct fields.
1. Handle the returned error.

## Installation

```bash
go get github.com/woozymasta/flags
```

## Minimal Parser

```go
package main

import (
  "errors"
  "fmt"
  "os"

  "github.com/woozymasta/flags"
)

type Options struct {
  Verbose bool   `short:"v" long:"verbose" description:"Show verbose output"`
  Region  string `long:"region" default:"eu-west-1" description:"Cloud region"`
}

func main() {
  var opts Options
  parser := flags.NewParser(&opts, flags.Default)

  _, err := parser.Parse()
  if err != nil {
    var ferr *flags.Error
    if errors.As(err, &ferr) && ferr.Type == flags.ErrHelp {
      os.Exit(0)
    }
    os.Exit(1)
  }

  fmt.Println(opts.Region, opts.Verbose)
}
```

`Default` enables common parser behavior:

* help flag registration;
* parser error printing;
* `--` pass-through handling.

Use `Default&^PrintErrors` in tests or libraries when callers should own
error output.

## Parse and ParseArgs

`Parse` reads from `os.Args[1:]`. Use it in final applications.

`ParseArgs` accepts a slice.
Use it in tests, wrappers, embedded parsers, and examples.

Both return remaining arguments and an error:

```go
rest, err := parser.ParseArgs([]string{"--verbose", "extra"})
```

Remaining arguments are values that were intentionally not consumed.
They may come from `--`, unknown-option handling, pass-through modes,
or command-specific behavior.

## Error Handling

Parser errors use `*flags.Error`.
The error type tells callers whether the failure is about missing input,
unknown flags, invalid choices, validation, help output, or another category.

```go
var ferr *flags.Error
if errors.As(err, &ferr) {
  switch ferr.Type {
  case flags.ErrHelp:
    return nil
  case flags.ErrValidation:
    return fmt.Errorf("invalid command-line value: %w", err)
  }
}
```

Do not string-match error text in application code.
Use `errors.As` and `Error.Type`.

Tests may still assert exact messages when checking presentation,
but exact message tests are intentionally more brittle.

## Naming Rules

An option field needs at least one of `short` or `long`.
If no name tag is provided, a mapped field is still a potential option,
but it cannot be addressed from the command line.
In practice, new option fields should define `long`.

```go
type Options struct {
  Output string `short:"o" long:"output"`
}
```

Short names are one rune.
Long names are strings and may be limited by `MaxLongNameLength`.
The project default is intentionally conservative to keep help readable.

## Struct Field Kinds

Common field kinds work directly:

* `bool` for switches;
* strings for text;
* integer and float types for numbers;
* `time.Duration` for durations;
* slices for repeated values;
* maps for key/value input;
* structs for groups and commands;
* pointers when nil vs zero must be observable.

Custom conversion is supported through standard Go interfaces
and package-specific mapper hooks.

## When to Use Parser Options

Parser options are bit flags passed to `NewParser`:

```go
parser := flags.NewParser(&opts, flags.Default|flags.HelpCommands)
```

Use parser options for parser-wide behavior.
Use struct tags for field-local behavior.
Use runtime setters when metadata is not known until code runs.

This separation keeps command-line contracts visible
in the struct while still allowing dynamic integration code.
