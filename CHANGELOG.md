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

## [Unreleased][] - 2026-04-21

### Added

* Completion script generation API:
  `Parser.WriteCompletion(...)` and `Parser.WriteNamedCompletion(...)`
  with `bash` and `zsh` support.
* Parser tag customization API:
  `Parser.SetTagPrefix(...)` and `Parser.SetFlagTags(...)`.
* Global environment variable prefix support via `Parser.SetEnvPrefix(...)`.
* Dynamic defaults via `DefaultProvider`.
* Parser option `DefaultsIfEmpty` to apply defaults only to empty values.
* `Group.Data()` accessor.
* Support for `encoding.TextMarshaler` / `encoding.TextUnmarshaler`
  (with `flags.Marshaler` / `flags.Unmarshaler` precedence).
* `examples/zsh-completion` template.
* Benchmarks for core parsing/help/INI flows.
* `terminator` option tag for find-style terminated argument lists
  (supports both `[]T` and `[][]T` option targets).
* Parser option `KeepDescriptionWhitespace` to keep leading indentation in
  multi-line help descriptions.
* Parser option `EnvProvisioning` to auto-derive `env` keys from
  `long` tags when `env` is not explicitly set.
* Option tag `auto-env:"true"` for per-flag env key derivation from `long`
  without enabling the global parser option.
* `auto-env:"false"` per-option opt-out when global `EnvProvisioning` is enabled.
* Unified boolean tag parsing for option/group/command boolean tags
  (`true/false`, `yes/no`, `y/n`, `1/0`, `on/off`) with validation errors
  for invalid values.

### Changed

* Module path moved to `github.com/woozymasta/flags`.
* Dependency switched from `github.com/sergi/go-diff` to
  `github.com/google/go-cmp`.
* Package/module docs and README were reworked and expanded.
* SPDX headers were introduced across source files.
* CI/checking pipeline was modernized (linting, alignment, cross-platform jobs).

### Fixed

* `examples/main.go` now checks help errors via `*flags.Error` + `errors.As`.
* Duplicate `default` struct tags in examples were removed.
* Windows `gofmt` instability fixed via `.gitattributes` (LF normalization).

[Unreleased]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.6.1...HEAD

---

> [!NOTE]  
> The sections **below** are a reconstructed change history
> from before this fork was created,
> mirrored as `legacy/*` tags in this repository.  
> Entries **above** this note follow the fork's own versioning line.
> This project does not inherit old upstream release tags
> and starts its own versioning sequence.

---

## [legacy/v1.6.1][] - 2024-06-15

commit: `c02e333e441eb1187c25e6d689d769d499ec2a0b`

### Changed

* Reverted a minor cleanup change related to an unused parameter.

[legacy/v1.6.1]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.6.0...legacy%2Fv1.6.1

## [legacy/v1.6.0][] - 2024-06-15

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

## [legacy/v1.5.0][] - 2021-03-21

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

## [legacy/v1.4.0][] - 2018-03-31

commit: `c6ca198ec95c841fdb89fc0de7496fed11ab854e`

### Added

* macOS support updates and CI/runtime compatibility improvements.

### Changed

* Environment lookup moved to `os.LookupEnv`.
* Build/test tooling refresh for newer Go versions.

### Fixed

* Removed/limited unsafe syscall usage in constrained environments.

[legacy/v1.4.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.3.0...legacy%2Fv1.4.0

## [legacy/v1.3.0][] - 2017-07-20

commit: `96dc06278ce32a0e9d957d590bb987c81ee66407`

### Added

* Better completion behavior for short/multi-flag patterns.

### Fixed

* Empty-subcommand crash.
* `default-mask:"-"` behavior.
* Several completion and parsing edge cases.

[legacy/v1.3.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.2.0...legacy%2Fv1.3.0

## [legacy/v1.2.0][] - 2017-02-12

commit: `48cf8722c3375517aba351d1f7577c40663a4407`

### Added

* `Option.IsSetDefault()` exposure.

### Fixed

* Pointer initialization for custom marshalers.
* Non-tagged struct fields are no longer modified during parsing.

[legacy/v1.2.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.1.0...legacy%2Fv1.2.0

## [legacy/v1.1.0][] - 2017-02-12

commit: `8bc97d602c3bfeb5fc6fc9b5a9c898f245495637`

### Changed

* Historical release marker retained for compatibility.

[legacy/v1.1.0]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1.1...legacy%2Fv1.1.0

## [legacy/v1.1][] - 2016-11-04

commit: `8bc97d602c3bfeb5fc6fc9b5a9c898f245495637`

### Added

* Force POSIX-style flags on Windows via build tag.
* Signed negative number handling improvements.
* Better INI and man-page behavior.

### Fixed

* Help output stream behavior (`--help` to stdout).
* Windows-related test/doc issues.

[legacy/v1.1]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv1...legacy%2Fv1.1

## [legacy/v1][] - 2013-11-22

commit: `37c8226983775d404b6edfebd44be1078bd0fe95`

### Added

* Windows-style option support.
* `Marshaler`/`Unmarshaler` interfaces.
* `default-mask` support.
* `Usage` interface.

[legacy/v1]: https://github.com/WoozyMasta/flags/compare/legacy%2Fv0.1...legacy%2Fv1

## [legacy/v0.1][] - 2013-08-26

commit: `1c98f1f5b27ef97fb039f258dce6aa14bd80ce41`

### Added

* First tagged release.

[legacy/v0.1]: https://github.com/WoozyMasta/flags/tree/legacy/v0.1

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
