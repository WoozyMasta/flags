# flags

Reflection-based command-line parser for Go.

## Core Features

* Short and long flags (`-v`, `--verbose`)
* Optional and required arguments
* Slices and maps as option values
* Positional arguments
* Nested option groups
* Commands and subcommands
* Defaults from tags and environment variables
* INI parse/write support
* Bash/Zsh completion script generation
* Help output and template-based documentation rendering (man/markdown/html)
* Configurable parse output routing for help/errors (`stdout` or `stderr`)

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

## Struct Tags Reference

All struct tags are configurable:

* Use `parser.SetTagPrefix("flag-")` to apply a common prefix for all tags.
* Use `parser.SetFlagTags(...)` for full custom tag-name mapping.
* List tags (`defaults`, `choices`, `aliases`) use `;` by default.
* Use `parser.SetTagListDelimiter(',')` to change list-tag delimiter.
* Boolean tags like `required`, `optional`, `hidden`, `no-flag`, `no-ini`,
  `subcommands-optional`, `pass-after-non-option`, `unquote`, `auto-env`:
  * positive: `true`, `yes`, `y`, `1`, `on`
  * negative: `false`, `no`, `n`, `0`, `off`

### Option tags

* `short`: one-letter short option name used as `-v`.
* `long`: canonical long option name used as `--verbose`.
* `description`: primary text shown in help/docs for the option.
* `long-description`: extended text for man/doc templates.
* `required`: fails parse if option is missing after defaults/env are applied.
* `optional`: allows option with or without explicit value.
* `optional-value`: value used when optional argument is omitted.
* `value-name`: placeholder name in help (for example `--port=PORT`).
* `default`: legacy repeatable default tag, mainly for backward compatibility.
* `defaults`: preferred multi-value default tag using list delimiter.
* `choice`: legacy repeatable whitelist tag for accepted values.
* `choices`: preferred whitelist tag with delimiter-separated values.
* `default-mask`: hides real default in help/docs (for secrets/tokens).
* `env`: explicit environment key used as fallback source.
* `auto-env`: derive env key from `long` for this option only.
* `env-delim`: split env value for slices/maps (for example `a,b,c`).
* `base`: integer radix for parse and defaults (for example `16`).
* `key-value-delimiter`: key/value separator for map values (default `:`).
* `no-flag`: disables CLI parsing for this field, keeps it in struct only.
* `hidden`: keeps option parseable, but removes it from help/completion/docs.
* `no-ini`: excludes option from INI read/write flow.
* `ini-name`: custom key name for INI read/write instead of flag name.
* `unquote`: controls string unquoting for quoted CLI values.
* `order`: explicit render priority in help/man/completion sorting.
* `terminator`: consume args until token, `find -exec` style (`[]T`, `[][]T`).

### Group tags

* `group`: marks nested struct as a named option group.
* `description`: group heading shown in help/docs.
* `long-description`: extended prose for group-focused docs/man output.
* `namespace`: prefixes child long flags (for example `db.host`).
* `env-namespace`: prefixes child env keys before global env prefix.
* `hidden`: hides the group from help/completion/docs, keeps parsing active.

### Command tags

* `command`: marks field as subcommand and command scope root.
* `description`: one-line command summary in help/docs.
* `long-description`: full command description for docs/man output.
* `alias`: legacy repeatable alias tag for backward compatibility.
* `aliases`: preferred delimiter-separated alias list.
* `subcommands-optional`: command can run without child subcommand selection.
* `pass-after-non-option`: enables command-local POSIX pass-through mode.
* `hidden`: hides command from help/completion/docs, keeps it executable.

### Positional-argument tags

* `positional-args`: marks nested struct as positional argument container.
* `required`: requires positional values to be provided by user.
* `positional-arg-name`: custom display name for usage/help placeholders.

### Tag conflicts

* `default` conflicts with `defaults`.
* `choice` conflicts with `choices`.
* `alias` conflicts with `aliases`.
* Use only one tag style from each pair on the same field.

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
  Servers []string `long:"server" defaults:"a.example;b.example"`
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

You can also auto-derive env names from `long` tags when `env` is not set:

```go
parser := flags.NewParser(&opts, flags.Default|flags.EnvProvisioning)
```

Example: `long:"some-function"` becomes `SOME_FUNCTION`.

For per-option behavior (without global parser flag), use:

```go
type Options struct {
  SomeFunction string `long:"some-function" auto-env:"true"`
}
```

When global `EnvProvisioning` is enabled, use `auto-env:"false"` to opt out
for a specific option.
Boolean tag values accept `true/false`, `yes/no`, `y/n`, `1/0`, and `on/off`.

## Option Sorting

By default option order is unchanged (declaration order). You can configure
sorting policy per parser:

```go
parser := flags.NewParser(&opts, flags.Default)
parser.SetOptionSort(flags.OptionSortByNameAsc)
```

Supported modes:

* `flags.OptionSortByDeclaration`
* `flags.OptionSortByNameAsc`
* `flags.OptionSortByNameDesc`
* `flags.OptionSortByType`

Use `order:"N"` on option fields for priority within a group block:

* `order > 0` moves option toward top
* `order < 0` moves option toward bottom
* `order == 0` uses configured sort mode

For `OptionSortByType`, type rank can be customized:

```go
_ = parser.SetOptionTypeOrder([]flags.OptionTypeClass{
  flags.OptionTypeString,
  flags.OptionTypeBool,
})
```

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

## Templates

Use `WriteDoc` to render parser documentation:

```go
if err := parser.WriteDoc(os.Stdout, flags.DocFormatMarkdown); err != nil {
  panic(err)
}
```

Current built-in templates:

* `markdown/list`: readable default markdown with option metadata.
* `markdown/table`: compact markdown table view.
* `markdown/code`: CLI-like code block output.
* `html/default`: simple standalone HTML documentation page.
* `html/styled`: styled HTML page with built-in CSS variables
  and automatic light/dark theme via `prefers-color-scheme`.
* `man/default`: classic groff/manpage output.

Use exported constants (`flags.DocTemplate*`) instead of hardcoded strings
when selecting built-ins programmatically.

Include hidden options/groups/commands when needed:

```go
_ = parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithBuiltinTemplate(flags.DocTemplateMarkdownList),
  flags.WithIncludeHidden(true),
)
```

Use custom templates:

```go
tpl := "{{ .Doc.Name }} - {{ .Doc.ShortDescription }}\n"
_ = parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithTemplateString(tpl),
)
```

The same `WithTemplateString`/`WithTemplateBytes` flow also works with
`flags.DocFormatMan` and `flags.DocFormatHTML`.

```go
_ = parser.WriteDoc(os.Stdout, flags.DocFormatMan)
```

`WriteManPage` is kept for backward compatibility and internally uses the
same doc templating pipeline (`man/default`).

`WriteHelp` stays unchanged for the core fast path.

Inspect/export built-ins:

```go
for _, name := range flags.ListBuiltinTemplates() {
  fmt.Println(name)
}

_ = flags.WriteBuiltinTemplate(os.Stdout, flags.DocTemplateMarkdownList)
```

Complete example with groups, commands, env and doc rendering modes:
[`examples/advanced/main.go`](examples/advanced/main.go).

Rendered template snapshots used by tests (also useful as docs/examples):
[markdown-list.unix.md](examples/doc-rendered/markdown-list.unix.md),
[html-default.unix.html](examples/doc-rendered/html-default.unix.html),
[man-default.posix.1](examples/doc-rendered/man-default.posix.1).
See `examples/doc-rendered` for additional variants.

Detailed roadmap: [`docs-plan.md`](docs-plan.md)

## Template Helpers

Built-in helper functions available in templates:

* `hiddenMark`: returns `true` only when hidden markers are enabled
  (`WithMarkHidden(true)`) and the current entity is hidden.
* `optionForms`: returns split option forms (for example `-v`, `--verbose`,
  including value/optional suffixes when applicable).
* `codeJoin`: wraps each item with backticks and joins using a comma.
* `join`: joins string slices with a separator (for example aliases list).
* `wrap`: wraps plain text to a target width.
* `markdownWrap`: wraps markdown text to target width (default `76`).
* `indent`: adds fixed left indentation to each line.
* `defaultValue`: formats a default suffix like `(default: VALUE)` or empty.
* `code`: wraps text with markdown backticks.
* `codeFenceOpen`: returns the opening fenced code marker for `text` blocks.
* `codeFenceClose`: returns the closing fenced code marker.
* `quoteMan`: escapes text for man/groff-style output.
* `manInline`: applies inline man formatting for backtick-quoted fragments.
* `quoteMarkdown`: escapes markdown-sensitive backslashes.
* `quoteHTML`: escapes HTML entities (`<`, `>`, `&`, quotes).
* `isRequired`: returns `true` when option is required.
* `hasDefault`: returns `true` when option has default value
  (or env-derived fallback text).
* `hasEnv`: returns `true` when option has resolved environment key.
* `isBool`: returns `true` for boolean options.
* `isCollection`: returns `true` for slice/array/map options.

## Documentation and Examples

* API docs: <https://pkg.go.dev/github.com/woozymasta/flags>
* Example app: [`examples/main.go`](examples/main.go)
