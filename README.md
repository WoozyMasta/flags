# flags

Build Go CLIs faster with reflection-powered parsing,
rich help, completion, and docs out of the box.

## Core Features

* POSIX-inspired command-line parsing
* getopt/getopt_long-style flags with practical Go extensions
* Short and long flags with optional/required values
* Repeatable slice/map options and typed defaults
* Structured CLI design: commands, subcommands, groups, and positional args
* Flexible input sources: flags, environment variables, and INI
* Rich help, completion, and documentation output (man/markdown/html)
* Opt-in localization for help, parser errors, docs, and user-facing metadata
* Custom value parsing with standard Go interfaces

## Content

* [Installation](#installation)
* [Quick Start](#quick-start)
* [Upgrade from go-flags](#upgrade-from-jessevdkgo-flags)
* [Error Handling](#error-handling)
* [Option Values](#option-values)
* [Struct Tags Reference](#struct-tags-reference)
* [Tag Customization](#tag-customization)
* [Positional Arguments](#positional-arguments)
* [Groups](#groups)
* [Commands](#commands)
* [Defaults](#defaults)
* [Localization](#localization)
* [Environment Variables](#environment-variables)
* [INI Config](#ini-config)
* [Programmatic Configuration](#programmatic-configuration)
* [Option Sorting](#option-sorting)
* [Help](#help)
* [Version](#version)
* [Shell Completion](#shell-completion)
* [Color Schemes](#color-schemes)
* [Hidden and Secret Options](#hidden-and-secret-options)
* [Documentation Rendering](#documentation-rendering)
* [Templating](#templating)

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
  // flags.Default enables HelpFlag, PrintErrors and PassDoubleDash.
  parser := flags.NewParser(&opts, flags.Default)

  rest, err := parser.Parse()
  if err != nil {
    var ferr *flags.Error
    // ErrHelp is a benign control-flow error: help was shown.
    if ok := errors.As(err, &ferr); ok && ferr.Type == flags.ErrHelp {
      os.Exit(0)
    }
    // With flags.Default, parse errors are already printed once by the parser.
    os.Exit(1)
  }

  fmt.Printf("verbose level: %d\n", len(opts.Verbose))
  fmt.Printf("name: %s\n", opts.Name)
  fmt.Printf("rest args: %v\n", rest)
}
```

## Error Handling

Recommended pattern with `flags.Default`:

```go
rest, err := parser.Parse()
if err != nil {
  var ferr *flags.Error
  if errors.As(err, &ferr) && ferr.Type == flags.ErrHelp {
    os.Exit(0)
  }
  // Do not print err again: flags.Default includes PrintErrors.
  os.Exit(1)
}
_ = rest
```

If you want full control over error output (single custom logger/format),
disable built-in printing:

```go
parser := flags.NewParser(&opts, flags.Default &^ flags.PrintErrors)
_, err := parser.Parse()
if err != nil {
  var ferr *flags.Error
  if errors.As(err, &ferr) && ferr.Type == flags.ErrHelp {
    fmt.Fprintln(os.Stdout, ferr)
    os.Exit(0)
  }
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
}
```

## Upgrade from jessevdk/go-flags

This repository is a fork of
[`github.com/jessevdk/go-flags`](https://github.com/jessevdk/go-flags)
and is intended to be a drop-in replacement for most projects.

> [!IMPORTANT]
> One explicit behavior change: default maximum `long` flag length is now `32`.
> Longer names require `Parser.SetMaxLongNameLength(...)`.

Upgrade checklist:

* Update import path to `github.com/woozymasta/flags`.
* Ensure build/runtime Go version is `1.25+`.
* If your CLI uses `long` names longer than `32`,
 configure parser limit explicitly.

Other changes should not affect normal parsing behavior in existing apps.
The main visible runtime difference can be help/man rendering output:
long options with large choice lists now wrap more predictably,
and generated man output is not guaranteed to match previous
byte-for-byte snapshots.

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

## Struct Tags Reference

All struct tags are configurable:

* Use `parser.SetTagPrefix("flag-")` to apply a common prefix for all tags.
* Use `parser.SetFlagTags(...)` for full custom tag-name mapping.
* List tags (`defaults`, `choices`, `aliases`) use `;` by default.
* Use `parser.SetTagListDelimiter(',')` to change list-tag delimiter.
* Prefer plural tags (`defaults`, `choices`, `aliases`,
  `short-aliases`, `long-aliases`) for new code.
* Singular legacy tags are still supported for compatibility, but do not mix
  singular and plural variants on the same field.
* Boolean tags (`required`, `optional`, `hidden`, `immediate`,
  `no-flag`, `no-ini`, `subcommands-optional`, `pass-after-non-option`,
  `unquote`, `auto-env`) accept:
  `true/false`, `yes/no`, `y/n`, `1/0`, `on/off`.

### Option tags

* `short`: one-letter short option name used as `-v`.
* `long`: canonical long option name used as `--verbose`.
* `description`: primary text shown in help/docs for the option.
* `description-i18n`: i18n key for option description text.
* `long-description`: extended text for man/doc templates.
* `required`: fails parse if option is missing after defaults/env are applied.
* `optional`: allows option with or without explicit value.
* `optional-value`: value used when optional argument is omitted.
* `value-name`: placeholder name in help (for example `--port=PORT`).
* `value-name-i18n`: i18n key for value placeholder text.
* `default` / `defaults`: default value(s) for missing option input.
* `choice` / `choices`: allowed value whitelist.
* `short-alias` / `short-aliases`: additional short names.
* `long-alias` / `long-aliases`: additional long names.
* `default-mask`: hides real default in help/docs (for secrets/tokens).
* `env`: explicit environment key used as fallback source.
* `auto-env`: derive env key from `long` for this option only.
* `env-delim`: split env value for slices/maps (for example `a,b,c`).
* `base`: integer radix for parse and defaults (for example `16`).
* `key-value-delimiter`: key/value separator for map values (default `:`).
* `no-flag`: disables CLI parsing for this field, keeps it in struct only.
* `hidden`: keeps option parseable, but removes it from help/completion/docs.
* `immediate`: marks option as preemptive flow trigger (skips required checks
  and command execution when present).
* `no-ini`: excludes option from INI read/write flow.
* `ini-name`: custom key name for INI read/write instead of flag name.
* `unquote`: controls string unquoting for quoted CLI values.
* `order`: explicit render priority in help/man/completion sorting.
* `terminator`: consume args until token, `find -exec` style (`[]T`, `[][]T`).

### Group tags

* `group`: marks nested struct as a named option group.
* `description`: group heading shown in help/docs.
* `description-i18n`: i18n key for group long/extended description.
* `long-description`: extended prose for group-focused docs/man output.
* `long-description-i18n`: i18n key for group long description text.
* `group-i18n`: i18n key for group heading text.
* `namespace`: prefixes child long flags (for example `db.host`).
* `env-namespace`: prefixes child env keys before global env prefix.
* `ini-group`: stable INI section token for this group.
* `hidden`: hides the group from help/completion/docs, keeps parsing active.
* `immediate`: marks all options in the group subtree as immediate.

### Command tags

* `command`: marks field as subcommand and command scope root.
* `description`: one-line command summary in help/docs.
* `description-i18n`: i18n key for command summary text.
* `long-description`: full command description for docs/man output.
* `long-description-i18n`: i18n key for command long description text.
* `command-i18n`: i18n key for command summary text.
* `command-group`: display group for command help/docs; it does not affect
  parsing or INI section names.
* `alias` / `aliases`: command aliases.
* `ini-group`: stable INI section token for this command root.
* `subcommands-optional`: command can run without child subcommand selection.
* `pass-after-non-option`: enables command-local POSIX pass-through mode.
* `hidden`: hides command from help/completion/docs, keeps it executable.
* `immediate`: marks command scope as immediate for required/execution bypass.

### Positional-argument tags

* `positional-args`: marks nested struct as positional argument container.
* `required`: for positional args you can use `yes/no`, `1` (required), `N`
  (at least `N` values for `[]T`) or `N-M` (from `N` to `M` values for `[]T`).
* `positional-arg-name`: custom display name for usage/help placeholders.
* `arg-group`: display group for positional argument help/docs.
* `arg-name-i18n`: i18n key for positional display name text.
* `description`: help/docs description for the positional argument.
* `arg-description-i18n`: i18n key for positional description text.
* `default` / `defaults`: fallback values for positional argument.

### Tag conflicts

Do not combine singular and plural forms on one field:
`default` vs `defaults`, `choice` vs `choices`, `alias` vs `aliases`,
`short-alias` vs `short-aliases`, `long-alias` vs `long-aliases`.

## Tag Customization

If your structs already use tags for other libraries, you can remap flags
tag names without changing parser constructors.

### Prefix Remapping

Use a common prefix for all tags:

```go
type Cfg struct {
  Path string `flag-short:"p" flag-long:"path" flag-description:"path to config"`
}

parser := flags.NewParser(&cfg, flags.Default)
_ = parser.SetTagPrefix("flag-")
```

### Explicit Tag Mapping

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

## Positional Arguments

Use `positional-args:"yes"` on a struct field:

```go
type Options struct {
  Args struct {
    Input  string `arg-group:"Input"`
    Output string `arg-group:"Output"`
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
  } `command:"add" command-group:"Content" description:"Add an item"`
}
```

If command type implements `Execute(args []string) error`, it will be called.
Use `command-group` only for display grouping in help/docs; keep `ini-group`
for stable INI section names.

### Built-in help commands

Built-in command entry points are opt-in:

```go
parser := flags.NewParser(&opts, flags.Default|flags.HelpCommands)
```

`HelpCommands` enables `help`, `version`, `completion`, `docs`, and `config`.
You can enable them individually with `HelpCommand`, `VersionCommand`,
`CompletionCommand`, `DocsCommand`, and `ConfigCommand`.

By default these commands are shown under `Help Commands` in help/docs.
Use `parser.SetBuiltinCommandGroup("Reference")` to rename the display group,
or `parser.SetBuiltinCommandGroup("")` to render them without a group.

Common command forms:

```bash
app help
app version
app completion --shell bash ./completion.bash
app docs md --template table ./docs.md
app docs html --template styled ./docs.html
app docs man ./app.1
app config --comment-width 88 ./config.ini
```

## Defaults

Prefer `defaults:"..."` for new code.
Keep `default:"..."` only for legacy compatibility.

### Basic Defaults

Use `defaults:"..."` to define fallback values:

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

### Config-First Flows

For config-first flows (for example JSON/YAML -> flags), use:

* `flags.RequiredFromValues` to treat non-empty prefilled values as satisfying
  `required`.
* `flags.ConfiguredValues` as a convenience alias for
  `flags.DefaultsIfEmpty | flags.RequiredFromValues`.

Minimal mixed configuration pattern:

```go
cfg := Config{}
// json.Unmarshal(...) or yaml.Unmarshal(...) into cfg

parser := flags.NewParser(&cfg, flags.Default|flags.ConfiguredValues)
_, err := parser.Parse()
```

Note: for scalar zero values (`0`, `false`, `""`) Go cannot distinguish
"not provided" from "explicitly set to zero" without pointer fields.

## Localization

Localization is opt-in. Configure parser i18n with `SetI18n(...)`.
Built-in help, parser errors, generated docs, and INI example metadata use the
configured catalog. User-defined descriptions and placeholders can opt in with
`*-i18n` tags while keeping the original text as source fallback.

```go
//go:embed i18n/*.json
var i18nFS embed.FS

catalog, _ := flags.NewJSONCatalogDirFS(i18nFS, "i18n")

parser := flags.NewParser(&opts, flags.Default)
parser.SetI18n(flags.I18nConfig{
  Locale:      "ru",
  UserCatalog: catalog,
})
parser.SetI18nFallbackLocales("en")
```

Use `NewLocalizer(...)` when application code needs the same catalogs and
locale chain without depending on `Parser`:

```go
localizer := flags.NewLocalizer(flags.I18nConfig{
  Locale:      "ru",
  UserCatalog: catalog,
})

message := localizer.Localize("app.greeting", "Hello, {name}", map[string]string{
  "name": "Alice",
})
```

Common user text tags:

* `description-i18n`, `long-description-i18n`
* `group-i18n`, `command-i18n`
* `value-name-i18n`, `arg-name-i18n`, `arg-description-i18n`

JSON catalogs can be loaded from `fs.FS` with `NewJSONCatalogFS(...)` or
`NewJSONCatalogDirFS(...)`. User catalogs override built-in keys and can also
hold application-level keys, so one catalog workflow can cover both CLI text
and app text. Common go-i18n-style JSON message objects are accepted, but this
package does not depend on `github.com/nicksnyder/go-i18n`.

Locale resolution order:

1. Explicit `I18nConfig.Locale` from `SetI18n(...)`
1. Environment (`LC_ALL`, `LC_MESSAGES`, `LANG`, `LANGUAGE`)
1. Windows OS locale fallback when env is empty
1. Catalog fallback chain: exact locale -> base locale -> configured fallbacks
   (default includes `en`) -> source text

If localized group or command names are used with INI, set `ini-group:"..."`
so INI section names stay stable across locales.

Use `CheckCatalogCoverage(...)` or `CheckMergedCatalogCoverage(...)` in tests
to catch missing catalog keys and placeholder mismatches.

End-to-end example with embedded catalogs:
[`examples/i18n/main.go`](examples/i18n/main.go).

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

### Global Prefix

Use `SetEnvPrefix(...)` for a global application prefix:

```go
parser := flags.NewParser(&opts, flags.Default)
parser.SetEnvPrefix("MY_APP")
```

Then `PORT` resolves to `MY_APP_PORT`, and grouped values resolve like
`MY_APP_DB_HOST`.

### Auto-Derived Env Keys

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

## INI Config

INI support is available via `NewIniParser(...)`:

```go
parser := flags.NewParser(&opts, flags.Default)
ini := flags.NewIniParser(parser)
_ = ini.ParseFile("app.ini")
```

Render current parser values back to INI:

```go
ini := flags.NewIniParser(parser)
ini.Write(os.Stdout, flags.IniIncludeComments|flags.IniIncludeDefaults|flags.IniCommentDefaults)
```

### Write Example INI

Render a starter/example INI with structured comments:

```go
ini := flags.NewIniParser(parser)
ini.WriteExample(os.Stdout)
// Optional comment wrapping width:
ini.WriteExampleWithOptions(os.Stdout, flags.IniExampleOptions{CommentWidth: 88})
```

`WriteExample` behavior:

* `required` options are always rendered as active keys.
* Non-required options are commented when they look unset/default.
* Comment block includes description, `choices` (if set), and details
  like `repeatable` / `key-value-delimiter` where applicable.
* Section names and INI keys are stable identifiers and are not localized.
* If group/command localization tags are used, set `ini-group` explicitly
  so INI section names stay stable across locales.

Quick demo:
[`examples/advanced/main.go`](examples/advanced/main.go) supports
`--demo-ini` to render example INI from current CLI values.

### INI Tags

Useful INI tags:

* `ini-name:"..."` to override key name in INI
* `ini-group:"..."` to override section token for group/command INI blocks
* `no-ini:"true"` to exclude a field from INI processing

## Programmatic Configuration

If tags are not enough, implement `flags.Configurer`
on your options/group/command struct:

```go
type Options struct {
  Verbose bool `long:"verbose"`
  Run struct{} `command:"run" description:"Run workload"`
  Args struct {
    Target string
  } `positional-args:"yes"`
}

func (o *Options) ConfigureFlags(p *flags.Parser) error {
  if opt := p.FindOptionByLongName("verbose"); opt != nil {
    opt.AddLongAlias("debug")
    _ = opt.SetEnv("APP_VERBOSE", "")
  }

  if cmd := p.Find("run"); cmd != nil {
    cmd.AddAlias("execute")
    cmd.SetShortDescription("Execute workload")
  }

  args := p.Command.Args()
  if len(args) > 0 {
    args[0].SetName("target")
    args[0].SetDefault("local")
  }

  return nil
}
```

`ConfigureFlags` runs before parse when parser topology changes.
Use `parser.Validate()` to run configurators and duplicate-name checks manually.
Use `parser.Rebuild()` after changing tag mapping settings programmatically.

### Runtime Setters

Runtime tuning APIs are intentionally exposed as `Set*` methods:
see [`Option`](option.go), [`Command`](command.go), [`Group`](group.go),
and [`Arg`](arg.go) for the current full list.

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

### Type Rank Customization

For `OptionSortByType`, type rank can be customized:

```go
_ = parser.SetOptionTypeOrder([]flags.OptionTypeClass{
  flags.OptionTypeString,
  flags.OptionTypeBool,
})
```

## Help

Use built-in help rendering when you want a fast and predictable CLI help page.
`WriteHelp` is optimized for runtime use, while the template API (`WriteDoc`)
is intended for generating documentation files.

```go
parser := flags.NewParser(&opts, flags.Default)
parser.WriteHelp(os.Stdout)
```

### Help Behavior Flags

Common help behavior flags:

* `PrintHelpOnStderr`: print auto-help (`ErrHelp`) to `stderr`.
* `PrintErrorsOnStdout`: print non-help parse errors to `stdout`.
* `PrintHelpOnInputErrors`: print built-in help before common user-input parse
  errors (`ErrRequired`, `ErrCommandRequired`, unknown flags/commands, etc.).
* `VersionFlag`: add built-in `-v/--version` output (`ErrVersion`).
* `ShowCommandAliases`: force showing command aliases in the `Available commands`
  section even when a command has no short description (without this flag,
  aliases are shown only for commands with short descriptions).
* `ShowRepeatableInHelp`: append a `repeatable` marker for slice/map options.
* `SetTerminalTitle`: set terminal title during parsing (uses parser `Name`
  or `parser.TerminalTitle`).
* `HideEnvInHelp`: hide env placeholders (`$ENV`/`%ENV%`) in built-in help
  (see also [Hidden and Secret Options](#hidden-and-secret-options)).

### Description Formatting

Preserve indentation in multi-line descriptions (for lists/code blocks):

```go
parser := flags.NewParser(&opts, flags.Default|flags.KeepDescriptionWhitespace)
```

### Render Style Detection

Shell-aware render style for flags/env placeholders in help and docs:

* `DetectShellFlagStyle`: detect rendered flag style from shell context.
* `DetectShellEnvStyle`: detect rendered env placeholder style from shell
  context.
* Explicit overrides:
  * `SetHelpFlagRenderStyle(...)`
  * `SetHelpEnvRenderStyle(...)`

Available styles:

* `RenderStyleAuto`: platform fallback.
* `RenderStylePOSIX`: `-v`, `--verbose`, `$ENV`.
* `RenderStyleWindows`: `/v`, `/verbose`, `%ENV%`.
* `RenderStyleShell`: shell/process-based detection.

`GO_FLAGS_SHELL` can be used as explicit runtime override for detection
(`bash`, `zsh`, `fish`, `pwsh`, `powershell`, `cmd`).

```go
parser := flags.NewParser(&opts, flags.Default|
  flags.DetectShellFlagStyle|
  flags.DetectShellEnvStyle)

// Optional explicit overrides:
parser.SetHelpFlagRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpEnvRenderStyle(flags.RenderStyleWindows)
```

`forceposix` build tag is still available and affects parser defaults
(delimiters/parsing behavior on Windows).
It is useful for deterministic tests/builds.
Runtime render-style APIs only change presentation in help/docs.

## Version

Built-in version output is enabled with `flags.VersionFlag`.
It adds `-v/--version` to `Help Options` and returns `ErrVersion`.

```go
parser := flags.NewParser(&opts, flags.Default|flags.VersionFlag)
```

By default, built-in `-v/--version` uses a compact field set:
`flags.VersionFieldsCore`.

### Version Fields

You can switch to all fields or a custom mask:

```go
parser.SetVersionFields(flags.VersionFieldsAll)
// or:
parser.SetVersionFields(flags.VersionFieldVersion | flags.VersionFieldCommit)
```

Default metadata source is `runtime/debug.ReadBuildInfo()`.
For best auto-discovery results, build with `-buildvcs=auto`.

If both help and version are provided, help has priority:
`-h/--help` wins over `-v/--version`.

### Build-Time Overrides

For reproducible release metadata, set explicit values via setters.
Typical pattern:

Use this pattern when you want deterministic version output
in release builds instead of relying only on auto-detected metadata.

```go
package buildvars

var (
  Version = "dev"
  Commit  = "unknown"
  URL     = "https://github.com/example/project"
)
```

```go
package main

func main() {
  parser := flags.NewParser(&opts, flags.Default|flags.VersionFlag)
  parser.SetVersion(buildvars.Version)
  parser.SetVersionCommit(buildvars.Commit)
  parser.SetVersionURL(buildvars.URL)
}
```

Manual output can also be rendered with explicit field selection:

```go
parser.WriteVersion(os.Stdout, flags.VersionFieldsCore)
```

Example bash variables and build command (ldflags -> your buildvars):

```bash
MODULE="$(GOWORK=off go list -m -f '{{.Path}}')"
URL="https://$MODULE"
COMMIT="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
VERSION="$(
  git describe --match 'v[0-9]*' --dirty='.m' --always --tags 2>/dev/null || \
  echo v0.0.0
)"

go build -buildvcs=auto \
  -ldflags "-X ${MODULE}/internal/buildvars.Version=${VERSION} \
  -X ${MODULE}/internal/buildvars.Commit=${COMMIT} \
  -X ${MODULE}/internal/buildvars.URL=${URL}" \
  ./...
```

### Change Option

You can retune built-in version option names and description:

```go
parser := flags.NewParser(&opts, flags.Default|flags.VersionFlag)

if versionOpt := parser.BuiltinVersionOption(); versionOpt != nil {
  _ = versionOpt.SetShortName('B')
  _ = versionOpt.SetLongName("build-info")
  versionOpt.SetDescription("Show build information")
}
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

## Color Schemes

Colorized output can be enabled for built-in help and parser errors.

* Enable with parser flag `ColorHelp`.
* Enable parser error coloring with `ColorErrors`.
* Built-in schemes: `DefaultHelpColorScheme()` and
  `HighContrastHelpColorScheme()`.
* Built-in error schemes: `DefaultErrorColorScheme()` and
  `HighContrastErrorColorScheme()`.
* Use `SetHelpColorScheme(...)` for custom role colors.
* Use `SetErrorColorScheme(...)` for custom parser error colors.
  Warnings (`ErrRequired`, `ErrCommandRequired`) and critical errors use
  separate roles.
* Colors are auto-disabled when output target is not a TTY
  (for example pipes/log files) or when `NO_COLOR` is set.

```go
parser := flags.NewParser(&opts, flags.Default|flags.ColorHelp|flags.ColorErrors)
parser.SetHelpColorScheme(flags.DefaultHelpColorScheme())
parser.SetErrorColorScheme(flags.DefaultErrorColorScheme())
// For stronger contrast:
// parser.SetHelpColorScheme(flags.HighContrastHelpColorScheme())
// parser.SetErrorColorScheme(flags.HighContrastErrorColorScheme())
```

For non-colored logs/CI output, do not enable `ColorHelp` / `ColorErrors`.

### Custom Schemes

Use `SetHelpColorScheme(...)` / `SetErrorColorScheme(...)` to apply your own
styles for role-based help and parser error rendering.

## Hidden and Secret Options

Use hidden and masking controls when CLI internals or sensitive defaults
must stay out of public help/docs.

> [!CAUTION]
> `hidden:"true"` only hides entities from help/completion/docs output.
> It does not disable parsing and is not a security boundary.

* `hidden:"true"`: option/group/command stays parseable, but is removed from
  built-in help, completion, and generated docs.
* `default-mask:"***"`: replaces displayed default value in help/docs.
* `HideEnvInHelp`: suppresses rendered env placeholders (`$ENV` / `%ENV%`)
  in built-in help.
* `WithIncludeHidden(true)`: include hidden entities in `WriteDoc`.
* `WithMarkHidden(true)`: explicitly mark hidden entities in rendered docs.

```go
type Options struct {
  Token string `long:"token" env:"APP_TOKEN" default-mask:"***" hidden:"true"`
}

parser := flags.NewParser(&opts, flags.Default|flags.HideEnvInHelp)
_ = parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithBuiltinTemplate(flags.DocTemplateMarkdownList),
  flags.WithIncludeHidden(true),
  flags.WithMarkHidden(true),
)
```

## Documentation Rendering

Use `WriteDoc` to generate parser documentation in markdown/man/html formats:

```go
if err := parser.WriteDoc(os.Stdout, flags.DocFormatMarkdown); err != nil {
  panic(err)
}
```

### Built-In Templates

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

### Hidden Entities in Output

Include hidden options/groups/commands when needed:

```go
_ = parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithBuiltinTemplate(flags.DocTemplateMarkdownList),
  flags.WithIncludeHidden(true),
)
```

`WriteManPage` is kept for backward compatibility and internally uses the
same doc templating pipeline (`man/default`).

`WriteHelp` stays unchanged for the core fast path.

### Inspect Templates

```go
for _, name := range flags.ListBuiltinTemplates() {
  fmt.Println(name)
}

_ = flags.WriteBuiltinTemplate(os.Stdout, flags.DocTemplateMarkdownList)
```

Complete example with groups, commands, env and doc rendering modes:
[`examples/advanced/main.go`](examples/advanced/main.go).

### Rendered Examples

Rendered template snapshots used by tests (also useful as docs/examples):

[markdown-list.unix.md](examples/doc-rendered/markdown-list.unix.md),
[html-default.unix.html](examples/doc-rendered/html-default.unix.html),
[man-default.posix.1](examples/doc-rendered/man-default.posix.1).
See `examples/doc-rendered` for additional variants.

## Templating

Use custom templates when built-ins are not enough:

```go
tpl := "{{ .Doc.Name }} - {{ .Doc.ShortDescription }}\n"
_ = parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithTemplateString(tpl),
)
```

`WithTemplateString`/`WithTemplateBytes` also work with
`flags.DocFormatMan` and `flags.DocFormatHTML`.

```go
_ = parser.WriteDoc(os.Stdout, flags.DocFormatMan)
```

### Template Helpers

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
