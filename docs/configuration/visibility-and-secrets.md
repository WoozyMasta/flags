# Visibility and Secrets

The parser has several controls for hiding CLI metadata from generated output.
These controls improve help readability and reduce accidental disclosure,
but they are not security boundaries.

Secrets still live in process memory, command histories, environment variables,
config files, or logs depending on how the application handles them.

## Hidden Entities

`hidden:"true"` can be used on options, groups, and commands.

Hidden entities stay parseable and executable.
They are omitted from built-in help, completion, and generated docs by default.

Use hidden entities for:

* compatibility flags kept for old scripts;
* internal debugging switches;
* commands intended only for automation;
* migration windows where public docs should not advertise a feature.

Do not use hidden entities as authorization.
If a command should not be available, do not register it.

## Default Masks

`default-mask` changes how defaults are displayed.

```go
type Options struct {
  Token string `long:"token" env:"APP_TOKEN" default-mask:"***"`
}
```

Use a visible mask such as `***` when users should know a default exists.
Use `default-mask:"-"` when the default should not be rendered at all.

Masking affects display.
It does not remove the actual value from memory or from application logs.

## Environment Placeholders

`HideEnvInHelp` suppresses environment variable placeholders in built-in help.

```go
parser := flags.NewParser(&opts, flags.Default|flags.HideEnvInHelp)
```

Use it when env names are noisy, internal, or not helpful for most users.

Generated docs can still expose env metadata depending on template options.
Review generated output before publishing.

## INI Exclusion

`no-ini:"true"` excludes an option from INI read/write flows.

Use it for secrets that should not be written into example config files,
or for runtime-only values that do not belong in persisted configuration.

If an option can be supplied from env and should never be persisted,
combine `env` with `no-ini`.

## Generated Docs and Hidden Items

`WriteDoc` excludes hidden entities by default.
For internal documentation, include them explicitly:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithIncludeHidden(true),
  flags.WithMarkHidden(true),
)
```

`WithMarkHidden(true)` only marks hidden entities already included
in the rendered model. It does not include them by itself.

## Help vs Audit Output

Public help should be concise and user-focused.
Internal audit output may need hidden flags, env names, INI names, and defaults.

Use separate render commands or templates for those audiences.
Do not make normal `--help` carry every internal detail.

## Secret Handling Rules

Prefer environment variables or external secret stores for secrets.
Avoid command-line secret flags when possible,
because shell history and process listings can expose them.

If a secret must be a CLI option, mask defaults,
hide env placeholders if needed, and be careful with logs and error messages.
