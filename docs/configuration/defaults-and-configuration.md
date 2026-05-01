# Defaults and Configuration

Most CLIs need more than one value source.
A user can pass a flag explicitly, a config file can provide project defaults,
an environment variable can override deployment-specific values,
and struct tags can provide safe fallback values.

The parser supports all of these sources,
but the application should still have one clear precedence model.
If nobody can explain which source wins, the CLI will surprise users.

## Two Common Application Shapes

A CLI-first application treats command-line flags
as the main configuration surface.
Defaults and environment variables fill missing values.
This is the simplest model and fits many small tools.

A config-first application loads a config file or another in-memory config
object before parsing command-line arguments.
The command line then overrides selected values.
This fits services, deployment tools,
and CLIs where most settings live in a project file.

`ConfiguredValues` exists for the second shape.
It tells the parser that non-empty values already present
in the struct are intentional configuration values,
not accidental leftovers that should be reset by tag defaults.

## CLI-First Flow

In a CLI-first flow, construct the parser directly from an empty struct.

```go
type Config struct {
  Port int    `long:"port" default:"8080" env:"APP_PORT"`
  Host string `long:"host" default:"127.0.0.1"`
}

var cfg Config
parser := flags.NewParser(&cfg, flags.Default)
_, err := parser.Parse()
```

Precedence is straightforward:

1. command-line value;
1. environment value when `env` is configured;
1. tag default;
1. Go zero value.

Use this model unless there is a real config file or preloaded config object.

## Config-First Flow

In a config-first flow,
load the config struct before constructing or parsing with `flags`.

```go
type Config struct {
  ConfigPath string `long:"config" default:"app.yaml"`
  Port       int    `long:"port" default:"8080"`
  Token      string `long:"token" required:"true" env:"APP_TOKEN"`
}

cfg := Config{}

// Example only: load YAML, JSON, TOML, or another source into cfg here.
// The loaded file might set Port and Token before CLI parsing starts.
if err := loadProjectConfig("app.yaml", &cfg); err != nil {
  return err
}

parser := flags.NewParser(&cfg, flags.Default|flags.ConfiguredValues)
_, err := parser.Parse()
```

With `ConfiguredValues`, the parser keeps non-empty values
already loaded into `cfg`.
A required option is also considered satisfied when
the struct already has a non-empty value.

That matters for `Token` in the example.
If the config file already provided the token,
the user does not also need to type `--token`.
If neither config nor environment nor CLI provides it,
required validation still fails.

## What ConfiguredValues Expands To

`ConfiguredValues` is only a convenience option:

```go
flags.DefaultsIfEmpty | flags.RequiredFromValues
```

`DefaultsIfEmpty` says:
apply tag and environment defaults only when the current field is empty.

`RequiredFromValues` says:
when checking `required`, count a non-empty prefilled field as present.

Use `ConfiguredValues` when both statements are true for your application.
Use the individual options only when you intentionally want
half of that behavior.

## DefaultsIfEmpty Alone

Use `DefaultsIfEmpty` when prefilled values should survive,
but required fields must still be explicitly satisfied
by parser-managed sources.

This is rare.
Most config-first applications want prefilled values to satisfy `required` too.

One reasonable use is migration code where old defaults
are loaded into the struct for compatibility,
but new required settings must be acknowledged explicitly by users.

## RequiredFromValues Alone

Use `RequiredFromValues` when prefilled values should satisfy `required`,
but normal tag defaults should still reset empty parser-managed fields.

This is also rare. It is useful only when the application
has a separate reason to keep the normal default-reset behavior.

If that sounds unclear, do not use the option alone.
Use `ConfiguredValues` or the default CLI-first behavior.

## Tag Defaults

`default` provides one fallback value.

```go
Port int `long:"port" default:"8080"`
```

`defaults` provides multiple fallback values for repeated fields.

```go
Server []string `long:"server" defaults:"a.example;b.example"`
```

A default is not the same as a config value.
It is the parser's fallback when no stronger source provides a value.

For maps, the default key/value delimiter is `:`.
Use `key-value-delimiter` when another separator is part of the public syntax.

## Environment Values

`env` declares an environment variable fallback.

```go
Token string `long:"token" env:"APP_TOKEN" required:"true"`
```

Environment values are useful for deployment-specific values and secrets.
They are applied before required checks.

For slices and maps,
`env-delim` splits one environment variable into several values.

```go
Hosts []string `long:"host" env:"APP_HOSTS" env-delim:","`
```

Use `SetEnvPrefix` when a whole application should share a prefix.

```go
parser.SetEnvPrefix("MY_APP")
```

## Auto-Derived Environment Keys

`EnvProvisioning` derives environment keys from long option names when `env` is
not set.

```go
parser := flags.NewParser(&cfg, flags.Default|flags.EnvProvisioning)
```

For example, `long:"cache-dir"` becomes `CACHE_DIR`.

This is convenient for large CLIs,
but it also creates a large public environment-variable surface.
Use it only when that is intentional.

## INI as a Config Source

INI support uses parser metadata to read and write simple config files.

```go
parser := flags.NewParser(&cfg, flags.Default|flags.ConfiguredValues)
ini := flags.NewIniParser(parser)
if err := ini.ParseFile("app.ini"); err != nil {
  return err
}
_, err := parser.Parse()
```

If CLI values should override INI values,
parse INI before parsing command-line args.

If INI output is public, keep `ini-name` and `ini-group` stable.
Do not let localized display names become config-file identifiers.

## Source Precedence

For a config-first app,
a practical precedence model is:

1. application prefill from config file or code;
1. environment values declared with `env` or derived by env provisioning;
1. command-line values;
1. domain validation in application code.

Command-line values are still the user's explicit final override.
The parser options only decide how prefilled values and defaults participate in
that process.

## Zero Values and Pointers

Go scalar zero values are ambiguous.
For an `int`, zero can mean "not configured" or "configured as zero".
For a `bool`, false has the same problem.

Use pointer fields when that difference matters:

```go
type Config struct {
  Limit *int `long:"limit"`
}
```

Nil means missing.
A pointer to zero means explicitly configured as zero.

## Choosing a Flow

Choose the application shape first. Then choose parser options.

If the struct starts empty, use `flags.Default`.

If the struct is loaded from a config file before CLI parsing,
use `flags.Default|flags.ConfiguredValues` and document the precedence model
for users.

Avoid mixing several hidden default mechanisms.
Users should be able to answer: "where did this value come from?"
