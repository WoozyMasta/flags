# Environment

Environment support has two separate parts: value sources for options,
and runtime environment detection for presentation decisions.

Do not mix them conceptually. `env` tags configure option defaults.
Detection helpers describe the process environment,
shell, terminal, locale, and color support.

## Option Values from Environment

Use `env` to provide a fallback value for an option.

```go
type Options struct {
  Port int `long:"port" default:"8080" env:"APP_PORT"`
}
```

Environment values are applied before required checks.
A required option can therefore be satisfied by environment.

For slices and maps,
use `env-delim` to split one environment variable into several values.

```go
type Options struct {
  Hosts []string `long:"host" env:"APP_HOSTS" env-delim:","`
}
```

## Environment Namespaces

Use `env-namespace` on groups to prefix child option environment keys.

```go
type Options struct {
  DB struct {
    Host string `long:"host" env:"HOST"`
  } `group:"Database" env-namespace:"DB"`
}
```

Use `SetEnvPrefix` for an application-wide prefix.

```go
parser.SetEnvPrefix("MY_APP")
```

With both examples combined,
`HOST` resolves as `MY_APP_DB_HOST`.

## Auto Provisioning

`EnvProvisioning` derives environment keys from long option names.

```go
parser := flags.NewParser(&opts, flags.Default|flags.EnvProvisioning)
```

`long:"cache-dir"` becomes `CACHE_DIR`.

Use `auto-env:"true"` to enable derivation on one option.
Use `auto-env:"false"` to opt out when global provisioning is enabled.

Auto provisioning is convenient,
but it expands your public configuration surface. Use it deliberately.

## Runtime Environment Snapshot

`DetectEnvironment` returns a single snapshot:

```go
env := flags.DetectEnvironment()
fmt.Println(env.OS, env.Shell, env.Locale)
```

The snapshot includes:

* runtime OS;
* detected shell;
* detected completion shell;
* detected locale;
* terminal width and height;
* tty state for stdin, stdout, and stderr;
* detected render style.

Use this for application decisions that should match parser behavior.

`EnvironmentInfo` contains these fields:

* `OS`: runtime OS identifier from `RuntimeOS`.
* `Shell`: detected shell name from `DetectShell`.
* `CompletionShell`: completion script target from `DetectCompletionShell`.
* `Locale`: detected locale from `DetectLocale`.
* `TerminalColumns` and `TerminalRows`: detected terminal size.
* `TTY`: standard stream tty state from `DetectTTY`.
* `ShellStyle`: render style from `DetectShellStyle`.

Use the snapshot when several decisions should be based on one read of the
process environment.
Use the focused helpers when only one signal is needed.

## Shell and Render Style

`DetectShell` returns a shell name such
as `bash`, `zsh`, `pwsh`, or `cmd` when detection succeeds.

`DetectShellStyle` returns a render style useful for help and docs.
It chooses POSIX-style or Windows-style presentation.

`DetectCompletionShell` returns `bash`, `zsh`, or `pwsh`.
Unsupported shells fall back to bash completion format.

`GO_FLAGS_SHELL` can override shell detection for presentation.
This is useful in tests and unusual terminal setups.

`RuntimeOS` returns the Go runtime OS identifier.
It is useful when application behavior needs the same OS value used by parser
render-style detection.

## TTY and Color

`DetectTTY` reports tty status for standard streams.
It returns `TTYInfo` with `Stdin`, `Stdout`, and `Stderr` booleans.

`DetectFileTTY`, `DetectFDTTY`, and `DetectWriterTTY` are lower-level helpers.
Use them when output is not standard output.

`DetectTerminalSize` returns terminal columns and rows.
When size detection fails, it returns the parser fallback size.

`DetectColorSupport` checks color policy for a writer.
It disables color when `NO_COLOR` is set.
It enables color when `FORCE_COLOR` is set to a truthy value.
Otherwise it depends on tty detection.

Non-file writers are treated as color-capable by `DetectWriterTTY`
to preserve historical behavior.
Use explicit parser color options when output policy must be deterministic.

## Locale Detection

`DetectLocale` checks locale-related environment variables
and OS fallback where available.
The i18n system normalizes locale tokens and builds a fallback chain.

Use explicit `I18nConfig.Locale` for deterministic tests.
Rely on detection for final interactive tools.

## Environment Rules

Use `env` tags for stable configuration contracts.
Use detection helpers for presentation and runtime adaptation.

Avoid making parse behavior depend heavily on detected shell or terminal.
Scripts should see stable semantics across environments.
