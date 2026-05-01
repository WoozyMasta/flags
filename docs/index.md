# flags

<!-- markdownlint-disable-next-line MD033 -->
<img src="assets/logo.svg" alt="flags" align="right" width="300">

`flags` is a reflection-powered command-line parser for Go.
It builds CLIs from structs and adds practical tooling around parsing:
commands, help, completion, configuration, validation, localization,
and generated documentation.

<!-- markdownlint-disable-next-line MD033 -->
<br clear="right"/>

## Quick Start

Install the module:

```bash
go get github.com/woozymasta/flags
```

Minimal parser:

```go
type Options struct {
  Verbose bool   `short:"v" long:"verbose" description:"Show verbose output"`
  Name    string `long:"name" required:"true" description:"User name"`
}

var opts Options
parser := flags.NewParser(&opts, flags.Default)
_, err := parser.Parse()
```

`flags.Default` enables the built-in help flag, error printing,
and `--` pass-through handling.

## Where to Start

New project: read [Getting Started][], then [Options][],
[Commands][], and [Positional Arguments][].

Migrating from `jessevdk/go-flags`:
read [Migration Notes][], [Compatibility][], and [Error Handling][].

Building a polished CLI: read [Help, Completion, and Docs][],
[Completion][], [Version Metadata][], and [Localization][].

Adding config support:
read [Defaults and Configuration][], [Environment][], and [INI Configuration][].

## Core Concepts

Struct tags describe the public CLI contract.
Parser options control parser-wide behavior.
Runtime setters are available when metadata is generated or discovered by code.

Commands model actions. Groups organize related options.
Positional arguments model ordered values.
Custom value interfaces let domain types own parsing,
formatting, completion, and dynamic defaults.

## Practical Rules

* Keep public option and command names stable.
* Use `ErrorType` instead of matching error text.
* Use explicit render styles for generated docs and golden tests.
* Use validators for local value rules,
  and application code for cross-field or domain validation.

[Commands]: cli-structure/commands.md
[Compatibility]: maintenance/compatibility.md
[Completion]: output/completion.md
[Defaults and Configuration]: configuration/defaults-and-configuration.md
[Environment]: configuration/environment.md
[Error Handling]: parser-model/error-handling.md
[Getting Started]: start/getting-started.md
[Help, Completion, and Docs]: output/help-completion-docs.md
[INI Configuration]: configuration/ini.md
[Localization]: output/localization.md
[Migration Notes]: maintenance/migration.md
[Options]: cli-structure/options.md
[Positional Arguments]: cli-structure/positional-arguments.md
[Version Metadata]: output/version.md
