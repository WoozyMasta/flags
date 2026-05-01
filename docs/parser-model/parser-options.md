# Parser Options

Parser options are bit flags passed to `NewParser` or `NewNamedParser`.
They control parser-wide behavior.

Use tags for field-local behavior.
Use parser options for behavior that affects the whole command line,
error flow, help output, completion, or command execution.

## Combining Options

Combine options with `|`:

```go
parser := flags.NewParser(&opts, flags.Default|flags.HelpCommands)
```

Remove an option from a preset with `&^`:

```go
parser := flags.NewParser(&opts, flags.Default&^flags.PrintErrors)
```

`Default` is:

```go
flags.HelpFlag | flags.PrintErrors | flags.PassDoubleDash
```

It is a good application default.
Libraries and tests often disable `PrintErrors` so the caller owns output.

## Help and Version Options

* `HelpFlag` adds `-h` and `--help`.
  When help is requested, parse returns `ErrHelp`.
  With `PrintErrors`, help is printed automatically.
* `VersionFlag` adds built-in `-v` and `--version`.
  When version is requested, parse returns `ErrVersion`.
  With `PrintErrors`, version output is printed automatically.

If both help and version are requested, help has priority.

Use `BuiltinHelpOption` and `BuiltinVersionOption`
when code needs to retune built-in option names or descriptions.

## Error Output Options

`PrintErrors` prints parser errors automatically.
It prints help to standard output by default,
and other parse errors to standard error by default.

`PrintHelpOnStderr` routes help output to standard error.
Use it when help is shown because of an error path rather than
an explicit user request.

`PrintErrorsOnStdout` routes non-help parse errors to standard output.
This is uncommon,
but useful for CLIs whose output contract expects all parser text on stdout.

`PrintHelpOnInputErrors` prints help before common user-input errors.
This is useful for interactive tools, but can be noisy in scripts.

## Argument Passing Options

`PassDoubleDash` stops parsing after `--`.
Everything after `--` is returned as remaining args.

`IgnoreUnknown` returns unknown options as remaining args instead of failing.
Use it for wrappers and partial parsers.
Avoid it for strict public CLIs where typos should fail fast.

`PassAfterNonOption` stops option parsing after the first non-option argument.
This is strict POSIX-style behavior.
For command-local behavior, prefer the `pass-after-non-option` command tag.

`AllowBoolValues` allows explicit bool values such as `--flag=true`.
Without it, passing an argument to a bool option is an error.

## Defaults and Config Options

`DefaultsIfEmpty` applies tag and environment defaults only to empty fields.
Use it when structs may be prefilled before parsing and those values should not
be overwritten by tag defaults.

`RequiredFromValues` lets non-empty prefilled values satisfy `required`.
Use it when prefilled config values should count as real input.

`ConfiguredValues` combines both:

```go
flags.DefaultsIfEmpty | flags.RequiredFromValues
```

This is the common mode when JSON, YAML, INI,
or application code fills the struct before CLI parsing.
The detailed precedence model is documented in
[Defaults and Configuration][].

`EnvProvisioning` auto-generates environment keys from long option names when
an option does not define `env`.
Use it for CLIs that intentionally expose environment overrides broadly.

## Help Presentation Options

* `KeepDescriptionWhitespace` preserves leading and trailing whitespace
  in help descriptions.  
  Use it for descriptions that contain lists or code-like indentation.
* `ShowCommandAliases` forces aliases to render in command lists even when a
  command has no short description.
* `ShowRepeatableInHelp` appends a repeatable marker for slice,
  map, and repeatable positional values.
* `ShowChoiceListInHelp` forces choices to render as a vertical list.
* `AutoShowChoiceListInHelp` renders choices vertically
  only when the available width is tight.
* `HideEnvInHelp` suppresses environment placeholders in built-in help output.  
  This is useful when env names are internal or noisy.
* `DetectShellFlagStyle` detects POSIX vs Windows-style flag rendering
  for help and generated docs.  
  It changes presentation only. It does not change parsing.
* `DetectShellEnvStyle` detects environment placeholder rendering,
  for example `$NAME` vs `%NAME%`.
* `SetTerminalTitle` updates the terminal title during parsing
  using parser name or `parser.TerminalTitle`.  
  Use it only for interactive terminal applications.

## Color Options

* `ColorHelp` enables ANSI-colored built-in help when color support is detected.
* `ColorErrors` enables colored parser errors.
  Warning-like errors and critical errors use different roles.

Color output respects `NO_COLOR` and `FORCE_COLOR` through the environment
helpers.

Use `SetHelpColorScheme` and `SetErrorColorScheme`
when the built-in colors do not match the application.
Built-in help schemes are `DefaultHelpColorScheme`,
`HighContrastHelpColorScheme`, and `GrayHelpColorScheme`.
Built-in error schemes are `DefaultErrorColorScheme`,
`HighContrastErrorColorScheme`, and `GrayErrorColorScheme`.

Custom schemes are built from `HelpTextStyle` and `ANSIColor`.

## Built-in Command Options

* `HelpCommand` adds a `help` command.
* `VersionCommand` adds a `version` command.
* `CompletionCommand` adds a `completion` command.
* `DocsCommand` adds `docs` subcommands for generated documentation formats.
* `ConfigCommand` adds a `config` command that writes an example INI.
* `HelpCommands` enables all built-in helper commands.

Built-in commands are opt-in so applications
do not expose extra public commands accidentally.

## Command Execution Option

`CommandChain` changes command execution from leaf-only to parent-to-leaf.

Without it, only the selected leaf command implementing `Commander` runs.
With it, every active command implementing `Commander`
runs in command path order.

This is not a same-level command pipeline.
It affects only the selected command tree path.

## Choosing Options by Situation

For a normal command-line application, start with `Default`.

```go
flags.Default
```

It gives users `-h` / `--help`, prints parser errors once,
and lets `--` pass the rest of the command line through.
This is the expected behavior for most final binaries.

For library code, tests, and embedded parsers, disable automatic printing.

```go
flags.Default &^ flags.PrintErrors
```

The caller can then decide where errors are logged,
whether help is returned as text, and what exit code to use.
This avoids duplicated output in tests and applications with their own logger.

For applications with a real config file loaded before parsing CLI args,
add `ConfiguredValues`.

```go
flags.Default | flags.ConfiguredValues
```

Use this only when the struct already contains intentional values
before parsing starts.
It keeps those values from being overwritten by tag defaults
and lets them satisfy `required`.
If the struct starts empty, do not add it.

For tools that should expose built-in helper commands,
add `HelpCommands`.

```go
flags.Default | flags.HelpCommands
```

This publishes commands such as help, version, completion, docs and config.
Do not enable it automatically in libraries or minimal tools;
every built-in command becomes part of the public CLI surface.

For interactive terminal tools, add presentation options deliberately.

```go
flags.Default |
  flags.HelpCommands |
  flags.AutoShowChoiceListInHelp |
  flags.ColorHelp |
  flags.ColorErrors
```

This improves terminal help and diagnostics,
but it is not always appropriate for CI logs, machine-readable tools,
or CLIs where stable plain output matters more than presentation.

For generated docs and golden tests,
prefer explicit setters over detection options:

```go
parser.SetHelpWidth(80)
parser.SetHelpFlagRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpEnvRenderStyle(flags.RenderStylePOSIX)
```

That keeps snapshots stable across shells,
operating systems, and terminal widths.

Sorting and layout setters are covered in [Runtime Configuration][].

[Defaults and Configuration]: ../configuration/defaults-and-configuration.md
[Runtime Configuration]: runtime-configuration.md
