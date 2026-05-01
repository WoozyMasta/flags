# Migration Notes

This project is a fork of `github.com/jessevdk/go-flags`.
It is intended to remain familiar, but it adds stricter metadata checks,
new tags, localization, documentation rendering, I/O templates, validators,
and parser presentation controls.

Use this page as a practical checklist when moving an existing CLI.

## Import Path

Change imports to the new module path:

```go
import "github.com/woozymasta/flags"
```

Run tests after the import change before adopting new features.
This separates migration risk from feature changes.

## Long Name Limit

The default maximum long option name length is `32` runes.
This keeps help output readable and catches accidental huge names.

If an existing CLI intentionally uses longer names,
configure the parser explicitly:

```go
parser.SetMaxLongNameLength(0)  // disable limit
```

or set a project-specific limit:

```go
parser.SetMaxLongNameLength(64)
```

## Error Handling

Use `errors.As` with `*flags.Error`.
Do not string-match parser messages in application code.

```go
var ferr *flags.Error
if errors.As(err, &ferr) && ferr.Type == flags.ErrHelp {
  return nil
}
```

Message text is localizable and may change as diagnostics improve.
Error types are the stable application-facing contract.

## Defaults and Prefilled Config

If your old flow prefilled the options struct before parsing,
review default behavior.

Use `ConfiguredValues` only when the struct already contains intentional
configuration values before CLI parsing starts.
For example, this applies when JSON, YAML, INI,
or application bootstrap code has already filled the same struct.

```go
parser := flags.NewParser(&cfg, flags.Default|flags.ConfiguredValues)
```

This keeps non-empty prefilled values and lets them satisfy required checks.
If the struct starts empty, do not add this option during migration.

## Tags

Existing common tags such as `short`, `long`, `description`, `required`,
`default`, `env`, `choice`, `group`, and `command` remain recognizable.

Newer tags add behavior around:

* option aliases;
* option relations with `xor` and `and`;
* counter options;
* I/O templates;
* value validators;
* localization keys;
* INI stability;
* documentation rendering.

Adopt them gradually.
Do not rewrite all tags at once unless tests already cover the CLI surface.

## Commands

Existing `Execute(args []string) error` command behavior
remains leaf-oriented by default.

Enable `CommandChain` only if parent commands
should also execute from parent to leaf.
It is not a same-level command pipeline.

If an application uses custom command execution wiring,
consider `Parser.CommandHandler` instead of putting dependency construction
inside command structs.

## Help and Docs Snapshots

Help and manpage output may differ from older snapshots.
This fork has additional wrapping, sorting, render-style, color, localization,
and docs-template features.

If tests assert exact output, refresh snapshots intentionally.
Prefer targeted assertions for behavior and
a smaller number of golden tests for presentation.

## INI and Localization

INI section names and keys should be stable identifiers.
If group or command display names are localized,
set `ini-group` to prevent locale-dependent config sections.

User-facing text can move to catalogs through `*-i18n` tags
while keeping the literal tag text as fallback.

## Recommended Migration Order

1. Change import path and run tests.
1. Fix long option name limit issues if any.
1. Review error handling and remove string matching.
1. Decide whether the app is CLI-first or config-first.
1. Enable `ConfiguredValues` only where prefilled config is intentional.
1. Refresh help/docs snapshots.
1. Add new validators and I/O templates only after baseline behavior is stable.

## Migration Rules

Keep migration commits separate from feature commits.
A clean migration makes regressions easier to identify.

When adopting new parser features, add negative tests for invalid tags,
invalid values, and help/version control-flow cases.
