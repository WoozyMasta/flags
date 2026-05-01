# Version Output

Version output can be exposed as a built-in flag,
a built-in command, or a direct call to `WriteVersion`.

Use it when release metadata should be discoverable
without starting normal application work.

## Version Flag

Enable the built-in version flag with `VersionFlag`:

```go
parser := flags.NewParser(&opts, flags.Default|flags.VersionFlag)
```

The parser adds `-v` and `--version` to the help options group.
When requested,
it returns `ErrVersion` and prints version output if `PrintErrors` is enabled.

If help and version are both requested, help has priority.

## Version Command

Enable the built-in version command with `VersionCommand`,
or enable all built-in helper commands with `HelpCommands`.

```go
parser := flags.NewParser(&opts, flags.Default|flags.HelpCommands)
```

The command form is useful when `-v` is already used by the application,
or when helper commands are part of the public CLI design.

## Version Fields

`VersionFieldsCore` is the compact default.
It includes the main fields users usually need.

`VersionFieldsAll` enables all known fields.

You can set an explicit mask:

```go
parser.SetVersionFields(flags.VersionFieldVersion | flags.VersionFieldCommit)
```

Available field bits include file, version, commit, build time, repository URL,
package path, module path, modified marker, Go version, and target platform.

Exact field bits are:

* `VersionFieldFile`
* `VersionFieldVersion`
* `VersionFieldCommit`
* `VersionFieldBuilt`
* `VersionFieldURL`
* `VersionFieldPath`
* `VersionFieldModule`
* `VersionFieldModified`
* `VersionFieldGoVersion`
* `VersionFieldTarget`

## Metadata Source

The default metadata source is `runtime/debug.ReadBuildInfo()`.
For VCS metadata, build with Go's VCS build info enabled.
The usual release build mode is `-buildvcs=auto`.

Auto-detected metadata is convenient,
but release pipelines often need explicit values for reproducibility.

Use `ReadVersionInfo` when application code needs the detected build metadata
without creating a parser:

```go
info := flags.ReadVersionInfo()
```

Use `parser.VersionInfo()` when parser-level overrides should be merged
with detected build metadata.

`VersionInfo` contains executable, module, VCS, Go toolchain,
repository URL, build target, and dirty-tree metadata.

## Build-Time Overrides

Set explicit version values through parser setters:

```go
parser.SetVersion(buildvars.Version)
parser.SetVersionCommit(buildvars.Commit)
parser.SetVersionTime(buildvars.BuiltAt)
parser.SetVersionURL(buildvars.URL)
parser.SetVersionTarget(buildvars.GOOS, buildvars.GOARCH)
```

Use `SetVersionInfo` when the release pipeline already has a complete
`VersionInfo` value:

```go
parser.SetVersionInfo(flags.VersionInfo{
  Version:  buildvars.Version,
  Revision: buildvars.Commit,
  URL:      buildvars.URL,
})
```

A common pattern is to keep build variables in a small package:

```go
package buildvars

var (
  Version = "dev"
  Commit  = "unknown"
  URL     = "https://github.com/example/project"
)
```

Then override them with `-ldflags` in release builds.

## Direct Rendering

Use `WriteVersion` when version output is controlled by your own command or
option:

```go
parser.WriteVersion(os.Stdout, flags.VersionFieldsCore)
```

This avoids adding built-in parser behavior while still using the same format.

## Retuning the Built-in Option

The built-in version option can be materialized and changed:

```go
parser := flags.NewParser(&opts, flags.Default|flags.VersionFlag)

if versionOpt := parser.BuiltinVersionOption(); versionOpt != nil {
  _ = versionOpt.SetShortName('B')
  _ = versionOpt.SetLongName("build-info")
  versionOpt.SetDescription("Show build information")
}
```

Use this only when the default `-v/--version` conflicts
with existing CLI behavior.
For new applications, the conventional default is easier for users.

## Version Output Rules

Version output should be side-effect-free.
It should not require config files, network access, or command execution.

Set explicit build values in releases.
Rely on auto-detection for development builds and local tools.
