# JSON Configuration

JSON support lets a parser read and write option values using
the same metadata that drives CLI parsing and help output.

Use it when you prefer a structured, hierarchical config format.
JSON is natively supported by Go without additional dependencies.

## Reading JSON

Create a JSON parser from an existing parser:

```go
parser := flags.NewParser(&opts, flags.Default)
jp := flags.NewJSONParser(parser)
err := jp.ParseFile("app.json")
```

`ParseFile` reads from a file path. `Parse` reads from an `io.Reader`.
Use `NewJSONParser` when the application needs parser options,
groups, commands or other metadata.

Unknown JSON keys are silently ignored.

## Writing JSON

Render current values as indented JSON:

```go
jp := flags.NewJSONParser(parser)
jp.Write(os.Stdout)
```

Write to a file:

```go
err := jp.WriteFile("app.json")
```

Output is a JSON object where each key corresponds to a flag's long name
(or the value of its `json` struct tag).

## Key Names

By default, JSON keys match the long flag name exactly.
To use a different key, set the standard Go `json` struct tag:

```go
type Options struct {
  MaxRetries int `long:"max-retries" json:"maxRetries"`
}
```

The JSON config would then use `"maxRetries"` as the key.

Use `json:"-"` to exclude a field from JSON parsing and writing:

```go
type Options struct {
  Token string `long:"token" json:"-"`
}
```

When the `json` name component is empty, the key falls back to the `long` tag:

```go
type Options struct {
  Name string `long:"my-name" json:",omitempty"`
}
// JSON key is "my-name"; field is omitted from WriteFile output when empty.
```

Key resolution priority:

1. `json` struct tag name, when non-empty and not `"-"`.
1. Long flag name, optionally transformed by `KeyName`.
1. Struct field name as last resort.

## Key Name Transformers

`JSONParser.KeyName` controls how the long flag name is converted
to a JSON key when no `json` tag is present.
Built-in transformers:

* `flags.JSONKeyLong` - `"my-long-flag"` (default, identity)
* `flags.JSONKeyCamel` - `"myLongFlag"`
* `flags.JSONKeySnake` - `"my_long_flag"`
* `flags.JSONKeyPascal` - `"MyLongFlag"`

Set a transformer before parsing:

```go
jp := flags.NewJSONParser(parser)
jp.KeyName = flags.JSONKeyCamel
err := jp.ParseFile("app.json")
```

Any `func(string) string` is accepted when the built-ins are not enough.

## Nested Groups and Commands

Groups with a `namespace` tag become nested JSON objects.
The namespace value is the JSON object key.

```go
type Options struct {
  DB struct {
    Host string `long:"host"`
    Port int    `long:"port"`
  } `group:"Database" namespace:"db"`
}
```

Corresponding JSON:

```json
{
  "db": {
    "host": "localhost",
    "port": 5432
  }
}
```

Groups without a namespace are anonymous:
their options appear at the enclosing object level.
This is the default for the top-level struct passed to `NewParser`.

The CLI namespace delimiter (`.` or `-`) has no effect on JSON keys.
Nesting is always determined by the `namespace` value itself.

**Commands** are represented as nested objects keyed by the command name.
This mirrors the INI section approach and supports arbitrary nesting depth:

```json
{
  "global-flag": "value",
  "deploy": {
    "release-id": "v1.0.0",
    "force": false
  },
  "parent": {
    "sub": {
      "deep-flag": "value"
    }
  }
}
```

This lets a config file supply defaults for command-specific options.
CLI-provided commands and their flags still override config values
when `ParseAsDefaults` is set.

## Parsing as Defaults

`JSONParser.ParseAsDefaults` makes parsed JSON values behave like defaults
instead of explicit user input:

```go
jp := flags.NewJSONParser(parser)
jp.ParseAsDefaults = true
err := jp.ParseFile("app.json")
// then:
_, err = parser.Parse()
```

With `ParseAsDefaults`,
a command-line flag for the same option overrides the JSON value.
Without it, JSON values are treated as if the user set them explicitly,
and a later `--flag` still wins because CLI parsing runs after.

The typical flow without `ParseAsDefaults`:

1. Build parser.
1. Parse JSON file.
1. Parse command-line args.
1. Use the final struct.

## omitempty / omitzero

When writing JSON, a field tagged `json:",omitempty"` or `json:",omitzero"`
is omitted from output if its value is the zero value for its type:

```go
type Options struct {
  Name    string `long:"name" json:",omitempty"`   // omitted when ""
  Timeout int    `long:"timeout" json:",omitzero"`  // omitted when 0
}
```

These tags have no effect during reading.

## Built-in Config File Flag

Add `ConfigJSON` (and optionally `ConfigIni`) to parser options to enable
the built-in `-c / --config FILE` flag and the `config` command:

```go
parser := flags.NewParser(&opts,
    flags.Default|flags.ConfigJSON|flags.ConfigFlags)
parser.SetJSONKeyName(flags.JSONKeyPascal) // key naming for builtin config command
```

`ConfigFlags` adds `-c / --config FILE` that is pre-scanned before argument
parsing. The file is applied as defaults, so CLI flags always win.
The flag must be passed explicitly - no config file is loaded automatically
when the flag is absent.

`SetJSONKeyName` controls the key naming convention used by
`config --format json`. It does not affect `NewJSONParser` usage.

`ConfigCommand` (which equals `ConfigIni | ConfigJSON`) adds a `config`
subcommand that renders a config template. When both formats are enabled,
the command accepts `--format ini|json`.

## YAML and TOML

`ParseMap` accepts a `map[string]interface{}` directly,
enabling other configuration formats without adding dependencies
to the core module:

```go
import "go.yaml.in/yaml/v3"

var raw map[string]interface{}
if err := yaml.Unmarshal(fileBytes, &raw); err != nil { ... }

jp := flags.NewJSONParser(parser)
if err := jp.ParseMap(raw); err != nil { ... }
```

The same approach works with any TOML or other decoder
that produces a map of the same type.
