# Windows and POSIX

The parser supports POSIX-style CLI conventions
and Windows-specific option style.
Presentation can also adapt to shell and platform conventions.

Separate parsing behavior from rendering behavior.
Changing help render style does not change how command-line tokens are parsed.

## POSIX Style

POSIX-style options use `-` and `--`.

```bash
app -v --output=file.txt -- input
```

`--` stops option parsing when `PassDoubleDash` is enabled.
`Default` includes `PassDoubleDash`.

Short options can be combined when they do not need arguments.

```bash
app -abc
```

## Windows Style

On Windows, slash-prefixed options are supported.

```powershell
app /v /output:file.txt
```

Windows-style help rendering can show slash-prefixed forms
and Windows env placeholder style.

Use this behavior when the CLI is intended to feel native in Windows shells.

## forceposix Build Tag

The `forceposix` build tag disables Windows-style parsing defaults.

Use it when a project needs the same POSIX parse behavior
on all platforms, including Windows.

This is most useful for developer tools, CI tools,
and CLIs documented exclusively with POSIX syntax.

## Render Style

Render style controls help and generated docs. It does not change parsing.

Available styles include:

* `RenderStyleAuto`
* `RenderStylePOSIX`
* `RenderStyleWindows`
* `RenderStyleShell`

Use explicit setters for generated docs:

```go
parser.SetHelpFlagRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpEnvRenderStyle(flags.RenderStylePOSIX)
```

Use detection for interactive runtime help:

```go
parser := flags.NewParser(&opts, flags.Default|
  flags.DetectShellFlagStyle|
  flags.DetectShellEnvStyle)
```

## Shell Detection

`DetectShell` detects common shells such as bash, zsh, pwsh,
PowerShell, and cmd.

`GO_FLAGS_SHELL` can override shell detection.
Use it in tests or unusual shell setups.

`DetectShellStyle` maps shell and OS information to a render style.

## Completion Shells

Completion script generation supports bash, zsh, and pwsh.

`DetectCompletionShell` maps the runtime shell to one of those formats.
Unsupported shells fall back to bash.

Shell completion support is about generated shell scripts.
It does not imply that the parser accepts different command-line syntax.

## Environment Placeholder Style

POSIX rendering uses `$NAME`. Windows rendering uses `%NAME%`.

This affects help and docs only.
Actual environment variable lookup uses the configured key string
and the OS environment.

## Paths and Validation

Path validation uses platform semantics. For example,
`validate-path-abs` checks absolute paths according to the current OS.

Do not hardcode POSIX path assumptions in tests that run on Windows.
Use `filepath` helpers and temporary directories.

## Golden Tests

Golden help and docs should set render style explicitly.
Otherwise output can vary by shell, OS, or environment variables.

Recommended setup for stable POSIX snapshots:

```go
parser.SetHelpFlagRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpEnvRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpWidth(80)
```

## Cross-Platform Rules

Document one primary style for users.
Supporting both POSIX and Windows parsing is useful,
but mixed examples can confuse readers.

For cross-platform tools, prefer long options
and avoid relying on shell-specific quoting in examples.
