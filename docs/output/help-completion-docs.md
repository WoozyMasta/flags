# Help, Completion, and Generated Docs

Help, completion, and generated documentation all come from the same parser
model.
That is the main reason to describe the CLI with structs and tags:
one contract can drive runtime help, shell integration, reference docs,
and internal audit output.

This page is the overview.
Use the detailed pages for implementation details:

* [Completion][] covers shell scripts, raw completion mode,
  completion sources, and custom completers.
* [Documentation Templates][] covers `WriteDoc`, built-in templates,
  custom templates, hidden entities, and template helpers.
* [Version Metadata][] covers built-in version output and build metadata.
* [Windows and POSIX][] covers render styles and platform-specific output.

## Runtime Help

Use `WriteHelp` for terminal help:

```go
parser := flags.NewParser(&opts, flags.Default)
parser.WriteHelp(os.Stdout)
```

With `HelpFlag`, the parser registers `-h` and `--help`.
With `PrintErrors`, help is printed automatically when parsing returns
`ErrHelp`.

Treat `ErrHelp` as successful control flow.
Do not print the returned help error again when `PrintErrors` is enabled.

## Completion

Completion scripts are generated from the same command tree,
option metadata, aliases, choices, and completion hints used by help.

```go
err := parser.WriteNamedCompletion(
  os.Stdout,
  flags.CompletionShellBash,
  "myapp",
)
```

Completion executes the application in completion mode.
Keep startup code side-effect-light before parsing,
because shells may invoke completion on every tab press.

## Generated Documentation

Use `WriteDoc` when output is meant to be committed,
published, or shipped as a reference document.

```go
err := parser.WriteDoc(os.Stdout, flags.DocFormatMarkdown)
```

Generated documentation supports markdown, HTML, and manpage formats.
Use built-in templates for normal output,
or custom templates when a site or release process needs a specific shape.

`WriteHelp` should stay optimized for runtime terminal output.
`WriteDoc` is the richer documentation path.

## Presentation Settings

Help, completion descriptions, and docs share presentation metadata:
descriptions, `long-description`, aliases, choices, defaults,
environment names, hidden state, order, and render style.

Use explicit render styles for generated files and golden tests.
Shell or OS detection is useful for runtime output,
but generated repository files should be stable across machines.

Color is also runtime presentation.
Enable it for interactive terminals,
not for logs or committed snapshots.

## Color Schemes

`ColorHelp` enables ANSI color roles for built-in help and version output.
`ColorErrors` enables ANSI color roles for parser errors.

Use built-in schemes first:

* `DefaultHelpColorScheme`
* `HighContrastHelpColorScheme`
* `GrayHelpColorScheme`
* `DefaultErrorColorScheme`
* `HighContrastErrorColorScheme`
* `GrayErrorColorScheme`

Apply custom schemes with parser setters:

```go
parser.SetHelpColorScheme(flags.HighContrastHelpColorScheme())
parser.SetErrorColorScheme(flags.HighContrastErrorColorScheme())
```

Custom schemes use `HelpColorScheme`, `ErrorColorScheme`,
`HelpTextStyle`, and `ANSIColor`.
Each `HelpColorScheme` field maps one rendered help role,
such as option names, usage text, command names, env hints,
defaults, choices, or positional arguments.

Colors are still gated by runtime color detection.
`NO_COLOR` disables color.
`FORCE_COLOR` enables color for supported writers.
When neither is set, tty detection decides.

## Hidden Metadata

`hidden` removes options, groups, and commands from normal help,
completion, and generated docs.
The parser still accepts hidden entities.

Generated docs can include hidden entities for internal audit output:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithIncludeHidden(true),
  flags.WithMarkHidden(true),
)
```

Hidden metadata is not a security boundary.
Use it to keep public help readable,
not to protect secrets.

[Completion]: completion.md
[Documentation Templates]: doc-templates.md
[Version Metadata]: version.md
[Windows and POSIX]: windows-and-posix.md
