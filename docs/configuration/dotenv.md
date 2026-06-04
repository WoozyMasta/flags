# .env File Support

`flags` can load a `.env` file
and populate environment variables before parsing flags.
This lets you keep local configuration out of source control
and still benefit from the `env` tag-based option defaults.

The feature is opt-in. Default parser behaviour
is unchanged when no `DotEnv`-family option is set.

## Quick Start

```go
parser := flags.NewParser(&opts, flags.Default|flags.DotEnv)
_, err := parser.Parse()
```

With `DotEnv` enabled, a `.env` file in the current working directory
is read automatically before arguments are parsed.
A missing `.env` is silently ignored.

## .env File Format

```sh
# Comments start with #
HOST=localhost
PORT=5432

# Export prefix is optional
export APP_MODE=production

# Quoted values preserve whitespace
GREETING="Hello, World!"
NOTE='literal \n - no escape here'

# Inline comments (must be preceded by whitespace)
TIMEOUT=30 # seconds
```

Key rules:

* Keys must match `[A-Za-z0-9_.-]`.
* The `export` prefix is stripped.
* Both `KEY=VALUE` and `KEY: VALUE` (YAML-style) are accepted.
* Single-quoted values: no variable expansion, no escape sequences.
* Double-quoted values: expansion + `\n`, `\r`, `\\` escape sequences.
* Unquoted values: expansion + trailing whitespace and inline comments stripped.

## Variable Expansion

Values may reference other variables using the following syntax.

| Syntax            | Behaviour                                         |
| ----------------- | ------------------------------------------------- |
| `$VAR`            | Substitute value of `VAR`                         |
| `${VAR}`          | Substitute value of `VAR` (braced form)           |
| `${VAR:=default}` | Use `VAR`; if unset or empty, set it to `default` |
| `$${VAR}`         | Literal `${VAR}` - no substitution                |

Variable lookup checks values already defined in the same `.env` file first,
then falls back to `os.Getenv`.
An undefined variable expands to an empty string.

```sh
BASE=/opt/app
DATA_DIR=${BASE}/data         # → /opt/app/data
CACHE_DIR=${TMP_DIR:=/tmp}/cache # sets TMP_DIR if unset
TEMPLATE=$${UNESCAPED}        # → literal ${UNESCAPED}
```

Single-quoted values are never expanded regardless of the global setting.

## Parser Options

* `DotEnv` - load `.env` before parsing; skip silently if missing
* `DotEnvOverride` - like `DotEnv` but overrides existing env vars
* `DotEnvFlags` - add "Env Options" group with three pre-configured flags

```go
// Load .env, keep existing env vars intact
parser := flags.NewParser(&opts, flags.Default|flags.DotEnv)

// Load .env, overwrite existing env vars
parser := flags.NewParser(&opts, flags.Default|flags.DotEnvOverride)

// Add --env-file / --no-env / --env-override CLI flags
parser := flags.NewParser(&opts, flags.Default|flags.DotEnv|flags.DotEnvFlags)
```

## Env Options Group (DotEnvFlags)

When `DotEnvFlags` is set,
a built-in "Env Options" group is added with three flags
that are evaluated **before** the `.env` file is loaded:

```txt
Usage:
  myapp [OPTIONS]

Env Options:
  --env-file=FILE   Path to .env file
  --no-env          Disable .env file loading
  --env-override    Override existing environment variables from .env file
```

These flags are pre-scanned before the main parse,
so specifying `--env-file custom.env` on the command line
takes effect even for options whose defaults come from environment variables.

## Programmatic API

Use the programmatic API for full control without relying on a CLI flag.

```go
parser := flags.NewParser(&opts, flags.Default)

// Load once manually (no override)
if err := parser.LoadDotEnv(".env", ".env.local"); err != nil {
    log.Fatal(err)
}

// Load with override
if err := parser.OverloadDotEnv("production.env"); err != nil {
    log.Fatal(err)
}

// Change the default file used by DotEnv option
parser.SetDotEnvFile("config/.env")

// Disable variable expansion
parser.SetDotEnvNoExpand(true)

_, err = parser.Parse()
```

`LoadDotEnv` and `OverloadDotEnv` accept zero or more filenames.
When called with no arguments the parser
uses file configured by `SetDotEnvFile` (default `".env"`).  
A missing default file is silently ignored;
a missing explicitly named file returns an error.

## Disabling Expansion

Call `SetDotEnvNoExpand(true)` to treat all `$` sequences as literal characters.
This is useful when values contain dollar signs that should not be interpreted,
for example shell scripts or regular expressions.

```go
parser.SetDotEnvNoExpand(true)
// REGEX=^[a-z]{3}$  →  stored as-is, no expansion attempted
```

## Accessing Built-in Options

When `DotEnvFlags` is set you can retrieve the built-in options
to further customise them (rename, hide, change description, etc.):

```go
parser := flags.NewParser(&opts, flags.Default|flags.DotEnv|flags.DotEnvFlags)

// Materialise the group early (optional - ParseArgs does this automatically)
parser.EnsureBuiltinDotEnvOptions()

fileOpt := parser.BuiltinDotEnvFileOption()
fileOpt.LongName = "config" // rename --env-file to --config
```
