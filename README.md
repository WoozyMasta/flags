# flags

<!-- markdownlint-disable-next-line MD033 -->
<img src="docs/assets/logo.svg" alt="flags" align="right" width="250">

`flags` is a reflection-powered command-line parser for Go.
It lets you describe a CLI with ordinary Go structs and tags,
then handles parsing, help, completion, configuration, validation,
localization, and generated documentation around that model.

It keeps the familiar struct-tag workflow, while adding stricter validation,
better generated output, modern completion/docs tooling,
and practical integration APIs.

<!-- markdownlint-disable MD033 -->
<br clear="right"/>
<p align="center">
<a href="https://woozymasta.github.io/flags/">Documentation Site</a> ·
<a href="https://pkg.go.dev/github.com/woozymasta/flags">Go Reference</a>
</p>
<!-- markdownlint-enable MD033 -->

## Why Use It

Use `flags` when you want a CLI that is defined close to the Go data it fills.
The struct remains the public contract, and the parser can derive help,
completion, config examples, and documentation from the same source.

Useful strengths:

* short and long options with POSIX-style parsing;
* commands, subcommands, option groups, and positional arguments;
* typed values, defaults, slices, maps, counters, and custom parsers;
* environment variables and INI configuration;
* value validators for common string, path, and numeric constraints;
* shell completion for bash, zsh, and PowerShell;
* generated help, markdown, HTML, and manpage output;
* opt-in localization for parser text and user-facing metadata;
* runtime setters for generated or application-owned metadata.

## Installation

```bash
go get github.com/woozymasta/flags
```

## Minimal Example

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
  Name    string `short:"n" long:"name" required:"true" description:"User name"`
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

  if opts.Verbose {
    fmt.Println("verbose mode enabled")
  }
  fmt.Printf("hello, %s\n", opts.Name)
}
```

Run it:

```bash
app --name Alice
app -v --name Alice
app --help
```

`flags.Default` enables the built-in help flag,
parser error printing, and `--` pass-through handling.
For libraries and tests, use `flags.Default &^ flags.PrintErrors`
so callers control output.

## Commands

Commands are struct fields tagged with `command`.
A command can own options, positional arguments, subcommands,
and an `Execute(args []string) error` method.

```go
type AddCommand struct {
  Name string `long:"name" required:"true"`
}

func (c *AddCommand) Execute(args []string) error {
  fmt.Println("add", c.Name)
  return nil
}

type Options struct {
  Add AddCommand `command:"add" description:"Add item"`
}
```

This gives a command shape like:

```bash
app add --name task
```

## Configuration

Small tools often use tag defaults and environment variables:

```go
type Options struct {
  Port  int    `long:"port" default:"8080" env:"APP_PORT"`
  Token string `long:"token" env:"APP_TOKEN" required:"true" default-mask:"***"`
}
```

Applications that load a config file before parsing CLI arguments can use
`flags.ConfiguredValues` so prefilled struct values are kept and can satisfy
`required` checks.

INI support is available when a simple generated or user-editable config file
is useful:

```go
ini := flags.NewIniParser(parser)
_ = ini.ParseFile("app.ini")
```

## Help, Completion

The same parser model can render user-facing output:

```go
parser.WriteHelp(os.Stdout)
_ = parser.WriteNamedCompletion(os.Stdout, flags.CompletionShellBash, "app")
_ = parser.WriteDoc(os.Stdout, flags.DocFormatMarkdown)
```

## Documentation

Use the source documentation when working in the repository:
[docs/][Source Documentation].  
Use the published site for rendered guides and navigation:
[Documentation Site][].  
Use Go Reference for API documentation generated from Go symbols:
[Go Reference][].

<!-- links -->

[Documentation Site]: https://woozymasta.github.io/flags/
[Go Reference]: https://pkg.go.dev/github.com/woozymasta/flags
[Source Documentation]: docs/
