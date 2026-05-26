# Exporting Arguments

`ExportArgs` serializes the current parser state into CLI argv (`[]string`)
suitable for passing back to `ParseArgs`.
Use it to capture a resolved configuration snapshot, re-launch a process
with the same flags, or build audit logs without manual field mapping.

Subcommand names are emitted as tokens before their option flags.
Output is deterministic for the same parser state; map keys are sorted.

## Basic Usage

```go
parser := flags.NewParser(&opts, flags.Default)
_, err := parser.Parse()
if err != nil { ... }

args, err := parser.ExportArgs()
// args is something like: ["--verbose=true", "--host=localhost", "--port=8080"]
```

Re-parse the exported args into a fresh struct:

```go
var opts2 MyOptions
p2 := flags.NewParser(&opts2, flags.Default|flags.AllowBoolValues)
_, err = p2.ParseArgs(args)
```

`AllowBoolValues` is required when re-parsing because `ExportArgs` always
emits explicit `--flag=true` / `--flag=false` tokens for boolean options.

## Selection Modes

`WithMode` controls which options appear in the output.

The default mode is `ExportModeResolved`.
It includes all options with a resolved (non-zero) value.  
Booleans are always included regardless of their value.  
Empty slices and maps are excluded unless `WithIncludeEmptyCollections` is set.

`ExportModeExplicit` includes only options that were explicitly provided
on the command line and were not applied from a default tag
or environment variable.

`ExportModeNonZero` includes only options whose current value is non-zero.

```go
// Only flags the user explicitly passed
args, err := parser.ExportArgs(flags.WithMode(flags.ExportModeExplicit))

// All non-zero values
args, err := parser.ExportArgs(flags.WithMode(flags.ExportModeNonZero))
```

## Targeting a Subcommand

Without `WithCommandPath`, `ExportArgs` follows the active command chain
resolved by the most recent `ParseArgs` call.

`WithCommandPath` targets a specific subcommand by path regardless of
which command was last active:

```go
args, err := parser.ExportArgs(flags.WithCommandPath("server", "start"))
```

An error is returned if any path element is not found.

## Secret Options

Secret options (tagged `secret:"true"`) are excluded by default.
Use `WithSecrets` to include them with their current values:

```go
args, err := parser.ExportArgs(flags.WithSecrets(flags.ExportSecretsInclude))
```

## Hidden and Deprecated Options

Hidden options are excluded by default.
Deprecated options are included by default.

```go
args, err := parser.ExportArgs(flags.WithHidden(true))
args, err := parser.ExportArgs(flags.WithIncludeDeprecated(false))
```

## Empty Collections

Slice and map options with zero elements are excluded by default.

```go
args, err := parser.ExportArgs(flags.WithIncludeEmptyCollections(true))
```

## Serialization Format

* Scalar values emit a single `--name=value` token.
* Boolean values always emit an explicit `--name=true` or `--name=false` token.
* Slice values emit one token per element in index order.
* Map values emit one token per entry in sorted key order,
  using the `key-value-delimiter` tag value (default `:`).
* Options without a long name fall back to short form: `-x=value`.

On Windows the platform-native delimiter is used automatically,
so `/name:value` is produced instead of `--name=value`.

## Combining Options

All functional options compose:

```go
args, err := parser.ExportArgs(
    flags.WithMode(flags.ExportModeExplicit),
    flags.WithSecrets(flags.ExportSecretsInclude),
    flags.WithHidden(true),
    flags.WithCommandPath("server"),
)
```
