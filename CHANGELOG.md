<!-- markdownlint-disable MD024 -->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## [0.3.0] - 2026-05-02

### Added

* Parser-level batch setters for command descriptions and their i18n keys:
  `SetCommandShortDescriptions`, `SetCommandLongDescriptions`,
  `SetCommandDescriptions`, `SetCommandShortDescriptionI18nKeys`,
  `SetCommandLongDescriptionI18nKeys` and `SetCommandDescriptionI18nKeys`.
* Built-in `docs man|html|md` commands now support `--program-name`
  to override binary/program name used in generated templates.
* Built-in `docs html|md` commands now support `--toc` to include
  table-of-contents blocks with section and command anchors.
* Built-in `docs man|html|md` commands now support `--trim-descriptions`
  to force description whitespace trimming in rendered documentation.

### Changed

* Built-in doc templates were refined:
  HTML now renders full nested command details (matching markdown),
  preserves multiline descriptions, and improves styled TOC/spacing;
  markdown uses heading-based TOC anchors and keeps a single trailing newline;
  env keys are no longer rendered as synthetic default values.

[0.3.0]: https://github.com/WoozyMasta/flags/compare/v0.2.0...v0.3.0

## [0.2.0] - 2026-05-01

### Added

* `xor` and `and` options for mutually exclusive and dependent flags.
* `counter` option for integer occurrence counting,
  including repeated short flags and explicit numeric increments.
* `io-*` tags (`io`, `io-kind`, `io-stream`, `io-open`)
  for string option/positional input/output templates with validation;
  positional `auto`/`stream` modes support `stdin`/`stdout` fallback.
* `CommandChain` parser option for executing active commands
  from parent to leaf while keeping leaf-only command execution as the default.
* `validate-*` tags for string, path, and numeric post-parse validation
  on options and positional arguments.
* Numeric `required` ranges for repeatable options,
  allowing slice and map flags to require value counts.
* Documentation site structure with focused topic pages
  and broader public API coverage.

### Changed

* `ParseArgs` now re-runs duplicate flag/command validation
  only when parser metadata is dirty (after mutations/rebuild/config changes),
  reducing overhead for reused parser instances.
* User-facing error messages now use symmetric backtick quoting
  instead of legacy `` `value' `` quoting.

[0.2.0]: https://github.com/WoozyMasta/flags/compare/v0.1.1...v0.2.0

## [0.1.1] - 2026-04-27

### Added

* `AutoShowChoiceListInHelp` for width-based automatic rendering of
  vertical `valid values` lists in built-in help.

### Changed

* Built-in help no longer auto-renders vertical `valid values` lists by default.

[0.1.1]: https://github.com/WoozyMasta/flags/compare/v0.1.0...v0.1.1

## [0.1.0] - 2026-04-26

### Added

* Shell completion script generation via `Parser.WriteCompletion(...)`,
  `Parser.WriteNamedCompletion(...)`, and `Parser.WriteAutoCompletion(...)`
  for bash, zsh, and pwsh.
  Completion output includes command/option aliases, option `choices`,
  `completion` tag hints (`file`, `dir`, `none`) for options/positionals,
  bool value candidates when
  `AllowBoolValues` is enabled, and no-space handling for inline option values.
  Shell auto-detection (`zsh` / `pwsh`) falls back to `bash` when unknown.
* Template-based parser documentation rendering via `Parser.WriteDoc(...)`,
  `DocFormat`, `DocOption`, built-in markdown/html/man templates,
  custom template sources/data, hidden-entity controls, and template registry
  helpers `ListBuiltinTemplates(...)` / `WriteBuiltinTemplate(...)`.
* Opt-in i18n for built-in help, errors, version output, generated docs,
  INI examples, completion descriptions, and user-facing metadata through
  `*-i18n` tags, JSON catalog loaders, built-in locale catalogs,
  locale fallback, `Localizer`, and catalog coverage validation helpers.
* Built-in version output via `VersionFlag`, `VersionInfo`,
  `ReadVersionInfo(...)`, `Parser.WriteVersion(...)`, version override setters,
  and `VersionFields` field masks.
* Configurable help/error presentation: `ColorHelp`, `ColorErrors`,
  help/error color schemes, `ShowCommandAliases`, `ShowRepeatableInHelp`,
  `HideEnvInHelp`, `KeepDescriptionWhitespace`, shell-aware render styles,
  help width control, command option indentation control,
  and terminal-title updates.
* Parser error/output routing controls:
  `PrintHelpOnStderr`, `PrintErrorsOnStdout` and `PrintHelpOnInputErrors`.
* Display grouping for commands in CLI help and generated documentation via
  `command-group`, `Command.SetCommandGroup(...)`, and `.Doc.CommandGroups`.
* Opt-in built-in `help`, `version`, `completion`, `docs`, and `config`
  commands via `HelpCommand`, `VersionCommand`, `CompletionCommand`,
  `DocsCommand`, `ConfigCommand`, and the `HelpCommands` convenience mask.
* Extended struct-tag support: parser tag remapping,
  plural list tags (`defaults`, `choices`, `aliases`),
  configurable list delimiters, option aliases,
  `terminator`, `order`, `auto-env`, `immediate`, command aliases,
  command-local parsing controls, i18n tags, stable INI names,
  and positional argument defaults.
* Runtime configuration APIs for parser, command, group, option
  and positional metadata, including `Configurer`, `Parser.Validate()`,
  `Parser.Rebuild()`, built-in option accessors,
  and setter methods for aliases, visibility, defaults, choices,
  env/INI metadata, parsing behavior, and display metadata.
* Configuration-first parsing helpers:
  `DefaultsIfEmpty`, `RequiredFromValues`, and `ConfiguredValues`.
* Environment provisioning helpers:
  `Parser.SetEnvPrefix(...)`, `EnvProvisioning`,
  and per-option `auto-env` opt-in/opt-out behavior.
* Environment detection API for runtime hints:
  `DetectEnvironment()`, `DetectTTY()`, `DetectFileTTY(...)`,
  `DetectWriterTTY(...)`, and `DetectColorSupport(...)`.
* Dynamic defaults via `DefaultProvider`,
  plus support for `encoding.TextMarshaler` / `encoding.TextUnmarshaler`
  with existing `flags.Marshaler` / `flags.Unmarshaler` precedence.
* Configurable option ordering through `Parser.SetOptionSort(...)`,
  `Parser.SetOptionTypeOrder(...)`, and the `order` tag.
* INI example rendering via `IniParser.WriteExample(...)`
  and `IniParser.WriteExampleWithOptions(...)`.
* Advanced examples, i18n examples, custom tag examples,
  rendered documentation snapshots, zsh completion template,
  and benchmark coverage for core parse/help/INI/doc flows.

### Changed

* Module path moved to `github.com/woozymasta/flags`.
* Dependency switched from `github.com/sergi/go-diff`
  to `github.com/google/go-cmp`.
* ⚠️ **breaking**: minimum supported Go version increased
  from `1.20` to `1.25`.
* Dependencies were modernized, including `golang.org/x/sys`,
  `golang.org/x/text`, and `golang.org/x/term`.
* Built-in help rendering was overhauled:
  wrapping/alignment use display width,
  terminal width is detected through `x/term` with stdio fallback,
  choice lists and value placeholders adapt to available width,
  command option descriptions share the global description column,
  and width can be disabled with `Parser.SetHelpWidth(0)`.
* Built-in documentation/man rendering now goes through the shared template
  renderer; the legacy standalone man writer was replaced by `man/default`.
* Parser validation now reports duplicate short/long/env names
  and alias collisions, including built-in help/version option conflicts;
  metadata setter methods validate updates before applying them.
* Boolean struct-tag parsing is unified across option/group/command tags
  and rejects invalid boolean values consistently.
* Built-in help/version options are materialized lazily
  and can be customized before parsing.
* Package documentation, README content, examples,
  and struct-tag reference were rewritten around the forked module path
  and current feature set.
* CI/checking workflow was replaced with lint, formatting, security,
  cross-platform, race, benchmark, generated-output, and release checks.
* Source headers, repository metadata, line-ending normalization,
  markdown linting and license text were normalized.
* ⚠️ **breaking**: default maximum `long` flag length is now `32`.
  For longer names, configure parser limit explicitly via
  `Parser.SetMaxLongNameLength(...)`.

### Fixed

* `examples/basic/main.go` now checks help errors via
  `*flags.Error` + `errors.As`.
* Positional help rendering no longer panics
  for positional-only parsers or unicode positional names.
* Generated man-page timestamps honor `SOURCE_DATE_EPOCH` in UTC.
* Duplicate `default` tags and other invalid example metadata were removed.
* Invalid env-provided `choices` values now report validation errors.

### Removed

* Legacy GitHub workflow and cross-compile shell script
  were replaced by the new CI/release workflow and Makefile targets.
* Legacy standalone man-page implementation was removed
  in favor of the template-backed documentation renderer.

[0.1.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.6.1...v0.1.0

---

> [!NOTE]  
> The sections **below** are a reconstructed change history
> from before this fork was created,
> mirrored as `legacy/*` tags in this repository.  
> Entries **above** this note follow the fork's own versioning line.
> This project does not inherit old upstream release tags
> and starts its own versioning sequence.

---

## [legacy/v1.6.1] - 2024-06-15

commit: `c02e333e441eb1187c25e6d689d769d499ec2a0b`

### Changed

* Reverted a minor cleanup change related to an unused parameter.

[legacy/v1.6.1]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.6.0...legacy%2Fv1.6.1

## [legacy/v1.6.0] - 2024-06-15

commit: `1898d831bc780f0fcce3ea97d73a9df1b1e27ed4`

### Added

* `AllowBoolValues` option for explicit bool flag values.
* Per-command `pass-after-non-option` behavior.
* `key-value-delimiter` tag for map parsing.
* `SOURCE_DATE_EPOCH` support in man-page related tests.
* AIX and Solaris related portability/build updates.

### Changed

* Go toolchain/workflow modernization (`go 1.20`, updated `x/sys`, GitHub Actions).
* Numeric parsing improvements (underscore support).

### Fixed

* Help behavior for required positional arguments.
* Panic when rendering help for hidden command/group combinations.
* INI zero-value write behavior.

[legacy/v1.6.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.5.0...legacy%2Fv1.6.0

## [legacy/v1.5.0] - 2021-03-21

commit: `1878de27329cba29066dc088d84b3ce743885f82`

### Added

* Programmatic option addition to groups.
* `Option.Field()` and `Option.IsSetDefault()`.
* Better handling of unknown INI sections with `IgnoreUnknown`.

### Changed

* Map/slice reference values are cleared on first explicit set.
* INI defaults can override built-in defaults.
* Completion excludes hidden commands.
* Terminal width detection moved to `golang.org/x/sys/unix`.
* Project switched to Go modules.

### Fixed

* Subcommand INI section and man-page usage text issues.
* Error reporting for invalid env defaults includes flag context.

[legacy/v1.5.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.4.0...legacy%2Fv1.5.0

## [legacy/v1.4.0] - 2018-03-31

commit: `c6ca198ec95c841fdb89fc0de7496fed11ab854e`

### Added

* macOS support updates and CI/runtime compatibility improvements.

### Changed

* Environment lookup moved to `os.LookupEnv`.
* Build/test tooling refresh for newer Go versions.

### Fixed

* Removed/limited unsafe syscall usage in constrained environments.

[legacy/v1.4.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.3.0...legacy%2Fv1.4.0

## [legacy/v1.3.0] - 2017-07-20

commit: `96dc06278ce32a0e9d957d590bb987c81ee66407`

### Added

* Better completion behavior for short/multi-flag patterns.

### Fixed

* Empty-subcommand crash.
* `default-mask:"-"` behavior.
* Several completion and parsing edge cases.

[legacy/v1.3.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.2.0...legacy%2Fv1.3.0

## [legacy/v1.2.0] - 2017-02-12

commit: `48cf8722c3375517aba351d1f7577c40663a4407`

### Added

* `Option.IsSetDefault()` exposure.

### Fixed

* Pointer initialization for custom marshalers.
* Non-tagged struct fields are no longer modified during parsing.

[legacy/v1.2.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.1.0...legacy%2Fv1.2.0

## [legacy/v1.1.0] - 2017-02-12

commit: `8bc97d602c3bfeb5fc6fc9b5a9c898f245495637`

### Changed

* Historical release marker retained for compatibility.

[legacy/v1.1.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.1...legacy%2Fv1.1.0

## [legacy/v1.1] - 2016-11-04

commit: `8bc97d602c3bfeb5fc6fc9b5a9c898f245495637`

### Added

* Force POSIX-style flags on Windows via build tag.
* Signed negative number handling improvements.
* Better INI and man-page behavior.

### Fixed

* Help output stream behavior (`--help` to stdout).
* Windows-related test/doc issues.

[legacy/v1.1]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1...legacy%2Fv1.1

## [legacy/v1] - 2013-11-22

commit: `37c8226983775d404b6edfebd44be1078bd0fe95`

### Added

* Windows-style option support.
* `Marshaler`/`Unmarshaler` interfaces.
* `default-mask` support.
* `Usage` interface.

[legacy/v1]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv0.1...legacy%2Fv1

## [legacy/v0.1] - 2013-08-26

commit: `1c98f1f5b27ef97fb039f258dce6aa14bd80ce41`

### Added

* First tagged release.

[legacy/v0.1]: https://github.com/WoozyMasta/flags/tree/legacy/v0.1

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
