# INI Configuration

INI support lets a parser read and write option values using the same metadata
that drives CLI parsing and help output.

Use it for simple local configuration files,
starter config generation, and examples.
For complex application configuration, JSON, YAML, TOML,
or domain-specific formats may still be a better fit.

## Reading INI

Create an INI parser from an existing parser:

```go
parser := flags.NewParser(&opts, flags.Default)
ini := flags.NewIniParser(parser)
err := ini.ParseFile("app.ini")
```

`ParseFile` reads from a file path. `Parse` reads from an `io.Reader`.

For the simplest default parser shape,
`IniParse` is available as a convenience helper:

```go
err := flags.IniParse("app.ini", &opts)
```

Use `NewIniParser` when the application needs parser options,
custom tag names, groups, commands, or runtime metadata tuning.

INI parse errors may be `*flags.Error` or `*flags.IniError`.
`IniError` includes file and line information when available.

Unknown INI sections and keys fail by default.
When the parser has `IgnoreUnknown`, unknown sections and keys are skipped.

## Writing INI

Render current values:

```go
ini := flags.NewIniParser(parser)
ini.Write(os.Stdout, flags.IniDefault|flags.IniIncludeDefaults)
```

Write to a file:

```go
err := ini.WriteFile("app.ini", flags.IniDefault|flags.IniIncludeDefaults)
```

`IniDefault` includes comments.
`IniIncludeDefaults` includes default values.
`IniCommentDefaults` comments default values when defaults are included.
`IniIncludeComments` includes option description comments.

Use `IniNone` when no write options should be enabled.

## Example INI

`WriteExample` renders a starter configuration file:

```go
ini := flags.NewIniParser(parser)
ini.WriteExample(os.Stdout)
```

Use `WriteExampleWithOptions` to configure comment wrapping:

```go
ini.WriteExampleWithOptions(os.Stdout, flags.IniExampleOptions{
  CommentWidth: 88,
})
```

Example output is intended for humans.
It includes descriptions, choices, repeatable markers,
and key/value delimiter hints where useful.

## Section Names

Groups and commands create INI sections.
Section names should be stable identifiers, not localized display text.

Use `ini-group` on groups and commands when display names may change:

```go
type Options struct {
  DB struct {
    Host string `long:"host" ini-name:"host"`
  } `group:"Database" ini-group:"database"`
}
```

If a group or command uses localized names, set `ini-group` explicitly.
The parser rejects some localized metadata combinations without
a stable INI section token.

Nested groups and commands use dot notation in section names.
For example, a command `remote` with group `Auth` can be addressed as
`[remote.Auth]`.
Section matching is case-insensitive.

## Key Names

By default, INI keys follow option names.
Use `ini-name` when the config file key should differ from the CLI flag.

```go
type Options struct {
  DisplayName string `long:"display-name" ini-name:"name"`
}
```

Use `no-ini` to exclude an option from INI parsing and writing.

```go
type Options struct {
  Token string `long:"token" no-ini:"true"`
}
```

When reading INI, option keys are matched in this order:

1. `ini-name` tag, when present.
1. Go struct field name.
1. Long option name.
1. Short option name.

This lets old config keys remain stable even when CLI names are retuned.

## Parsing as Defaults

`IniParser.ParseAsDefaults` makes parsed INI values behave like parser defaults
instead of explicit user input.

```go
ini := flags.NewIniParser(parser)
ini.ParseAsDefaults = true
err := ini.ParseFile("app.ini")
```

Use this when an INI file provides fallback values,
but command-line input should still be treated as stronger user intent.

Without `ParseAsDefaults`, parsing INI sets option values directly.
This is the usual config-file flow when the INI file is an intentional source
before command-line override.

## Required Values

INI values can satisfy required options.
This supports config-first tools where users store required settings in a file
and override them on the command line only when needed.

When combining INI with struct prefill,
use the config-first model from Defaults and Configuration.
In practice that usually means parsing INI before command-line args
and using `ConfiguredValues` when the same struct was already prefilled
by application config code.

## CLI Overrides

A common flow is:

1. Build parser.
1. Parse INI file.
1. Parse command-line args.
1. Use the final struct.

This lets command-line input override config-file values.
Keep this order consistent across the application.

## Quoting and Values

INI parsing uses the same option conversion pipeline as CLI values.
String quoting is handled by the INI parser.
Map and slice values use option metadata such as `key-value-delimiter`.

Do not rely on INI for arbitrary nested structures.
If the config needs nested objects, INI may be the wrong format.

## INI Stability Rules

Use INI output as a user-facing contract.
Changing `ini-name` or `ini-group` can break existing config files.

When adding localization, review INI section names separately from help text.
Localized help is good. Locale-dependent config keys are not.
