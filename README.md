# flags

Reflection-based command-line parser for Go.

It supports flags, positional arguments, groups, subcommands, INI files,
shell completion, and man/help generation.

## Installation

```bash
go get github.com/woozymasta/flags
```

## Quick Start

```go
package main

import (
  "errors"
  "fmt"
  "os"

  "github.com/woozymasta/flags"
)

type Options struct {
  Verbose []bool `short:"v" long:"verbose" description:"Enable verbose output"`
  Name    string `short:"n" long:"name" required:"true" description:"User name"`
}

func main() {
  var opts Options
  parser := flags.NewParser(&opts, flags.Default)

  rest, err := parser.Parse()
  if err != nil {
    var ferr *flags.Error
    if ok := errors.As(err, &ferr); ok && ferr.Type == flags.ErrHelp {
      os.Exit(0)
    }
    os.Exit(1)
  }

  fmt.Printf("verbose level: %d\n", len(opts.Verbose))
  fmt.Printf("name: %s\n", opts.Name)
  fmt.Printf("rest args: %v\n", rest)
}
```

## Core Features

* Short and long flags (`-v`, `--verbose`)
* Optional and required arguments
* Slices and maps as option values
* Positional arguments
* Nested option groups
* Commands and subcommands
* Defaults from tags and environment variables
* INI parse/write support
* Bash/Zsh completion
* Help and man page generation

## Option Values

Common value types:

* primitive scalars (`string`, `int*`, `uint*`, `float*`, `bool`)
* slices (option can be repeated)
* maps (default key/value delimiter is `:`)
* custom types via:
  * `flags.Unmarshaler` / `flags.Marshaler`
  * `encoding.TextUnmarshaler` / `encoding.TextMarshaler`
* terminated option lists via `terminator:"..."` (`find -exec` style)

Example map and slice options:

```go
type Options struct {
  Include []string       `short:"I" description:"Include path"`
  Labels  map[string]int `long:"label" description:"label:value pairs"`
}
```

Terminated list options:

```go
type Options struct {
  Exec []string `long:"exec" terminator:";"`
}
```

Then `--exec echo hello ; --other-flag` stores `["echo", "hello"]` in `Exec`.

## Positional Arguments

Use `positional-args:"yes"` on a struct field:

```go
type Options struct {
  Args struct {
    Input  string
    Output string
  } `positional-args:"yes" required:"yes"`
}
```

## Groups

Group options for help readability and logical structure:

```go
type Options struct {
  Global bool `long:"global"`

  Database struct {
    Host string `long:"host"`
    Port int    `long:"port"`
  } `group:"Database"`
}
```

You can also add groups programmatically with `parser.AddGroup(...)`.

## Commands

Two ways:

1. Struct tag: `command:"name"`
1. Programmatic: `parser.AddCommand(...)`

Example:

```go
type Options struct {
  Add struct {
    Force bool `short:"f" long:"force"`
  } `command:"add" description:"Add an item"`
}
```

If command type implements `Execute(args []string) error`, it will be called.

## Defaults

Use `default:"..."` to define fallback values:

```go
type Options struct {
  Port    int      `long:"port" default:"8080"`
  Servers []string `long:"server" default:"a.example" default:"b.example"`
}
```

For map values, key/value delimiter is `:` by default (can be changed with
`key-value-delimiter:"="`).

If you need to keep pre-populated values and apply defaults only to empty
fields, enable parser option `flags.DefaultsIfEmpty`.
This is useful for non-empty/non-nil prefilled structs in integration code.

## Environment Variables

Use `env:"..."` to override defaults from environment:

```go
type Options struct {
  Port  int      `long:"port" default:"8080" env:"APP_PORT"`
  Hosts []string `long:"host" env:"APP_HOSTS" env-delim:","`
}
```

Use `env-namespace` on groups to prefix env names:

```go
type Options struct {
  DB struct {
    Host string `long:"host" env:"HOST"`
    Port int    `long:"port" env:"PORT"`
  } `group:"Database" env-namespace:"DB"`
}
```

With defaults, that resolves to `DB_HOST` and `DB_PORT`.

Use `SetEnvPrefix(...)` for a global application prefix:

```go
parser := flags.NewParser(&opts, flags.Default)
parser.SetEnvPrefix("MY_APP")
```

Then `PORT` resolves to `MY_APP_PORT`, and grouped values resolve like
`MY_APP_DB_HOST`.

## Tag Customization

If your structs already use tags for other libraries, you can remap flags
tag names without changing parser constructors.

Use a common prefix for all tags:

```go
type Cfg struct {
  Path string `flag-short:"p" flag-long:"path" flag-description:"path to config"`
}

parser := flags.NewParser(&cfg, flags.Default)
_ = parser.SetTagPrefix("flag-")
```

Or override specific tag names:

```go
type Cfg struct {
  Path string `my-short:"p" long:"path"`
}

parser := flags.NewParser(&cfg, flags.Default)
tags := flags.NewFlagTags()
tags.Short = "my-short"
_ = parser.SetFlagTags(tags)
```

## INI

INI support is available via `NewIniParser(...)`:

```go
parser := flags.NewParser(&opts, flags.Default)
ini := flags.NewIniParser(parser)
_ = ini.ParseFile("app.ini")
```

Useful INI tags:

* `ini-name:"..."` to override key name in INI
* `no-ini:"true"` to exclude a field from INI processing

## Help

Help:

```go
parser := flags.NewParser(&opts, flags.Default)
parser.WriteHelp(os.Stdout)
```

If you need to preserve leading whitespace in multi-line descriptions
(for lists or code snippets), enable:

```go
parser := flags.NewParser(&opts, flags.Default|flags.KeepDescriptionWhitespace)
```

## Man Pages

Generate a man page:

```go
parser := flags.NewParser(&opts, flags.Default)
parser.WriteManPage(os.Stdout)
```

## Shell Completion

Generate shell script output from your app:

```go
if opts.Completion != "" {
  _ = parser.WriteNamedCompletion(
    os.Stdout,
    flags.CompletionShell(opts.Completion),
    "myapp",
  )
  return
}
```

Use it:

```bash
./myapp --completion bash > ./myapp.bash
source <(./myapp --completion bash)
```

Raw completion mode:

```bash
GO_FLAGS_COMPLETION=1 ./myapp --some-arg prefix
```

Templates:
[`examples/bash-completion`](examples/bash-completion),
[`examples/zsh-completion`](examples/zsh-completion).

## Documentation and Examples

* API docs: <https://pkg.go.dev/github.com/woozymasta/flags>
* Example app: [`examples/main.go`](examples/main.go)
