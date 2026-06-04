# Documentation Templates

`WriteDoc` renders parser documentation through templates.
It is intended for generated files, manpages, HTML reference pages,
and internal audit output.

Use `WriteHelp` for runtime terminal help.
Use `WriteDoc` when the output is documentation.

## Formats

Supported output formats are:

* `DocFormatMarkdown`
* `DocFormatHTML`
* `DocFormatMan`
* `DocFormatJSON`

```go
err := parser.WriteDoc(os.Stdout, flags.DocFormatMarkdown)
```

The format controls escaping and default template selection.
A custom template can still be supplied for any supported format.
`DocFormatJSON` does not use templates - it serializes the doc model directly.

## Built-in Templates

Built-in templates cover common output styles:

* `DocTemplateMarkdownList` renders `markdown/list`.
  This is the default markdown template.
* `DocTemplateMarkdownTable` renders `markdown/table`.
* `DocTemplateMarkdownCode` renders `markdown/code`.
* `DocTemplateHTMLDefault` renders `html/default`.
  This is the default HTML template.
* `DocTemplateHTMLStyled` renders `html/styled`.
* `DocTemplateManDefault` renders `man/default`.
  This is the default manpage template.

Use exported template constants when selecting a template from code.

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithBuiltinTemplate(flags.DocTemplateMarkdownList),
)
```

List available templates:

```go
for _, name := range flags.ListBuiltinTemplates() {
  fmt.Println(name)
}
```

Print a built-in template source:

```go
err := flags.WriteBuiltinTemplate(os.Stdout, flags.DocTemplateMarkdownList)
```

`ListBuiltinTemplates` returns template names sorted by name.
`WriteBuiltinTemplate` returns an error for an unknown template name.

## Doc Options

`DocOption` values customize one `WriteDoc` call.
They do not mutate the parser.

Use `WithBuiltinTemplate` to select one of the embedded templates:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatHTML,
  flags.WithBuiltinTemplate(flags.DocTemplateHTMLStyled),
)
```

Use `WithTemplateString` or `WithTemplateBytes` when the template is owned by
the application.
Custom template text replaces the selected built-in template for that call.

Use `WithTemplateData` for values that are not part of the parser model:
repository URLs, build IDs, generated-file notices, or site metadata.

Use `WithIncludeHidden` and `WithMarkHidden` for internal documentation.
Hidden entities are excluded unless `WithIncludeHidden(true)` is set.

Use `WithBuiltinCommands` to control which built-in commands appear in
the generated output.
Pass `nil` to include all (the default when calling `WriteDoc` directly).
Pass a non-nil slice to restrict to the named commands only:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithBuiltinCommands([]string{"version"}),
)
```

When using the built-in `docs` command, the `--builtins` flag does the same.
Its choices are limited to actually-enabled commands;
the default is `help`, `version`, `completion`.

```sh
myapp docs html --builtins version --builtins completion output.html
```

Use `WithDocWrapWidth` to control markdown text wrapping for one render call.
A width of zero disables wrapping:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithDocWrapWidth(100),
)
```

## Custom Template Text

Use a template string:

```go
tpl := "{{ .Doc.Name }} - {{ .Doc.ShortDescription }}\n"
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithTemplateString(tpl),
)
```

Use bytes when the template is loaded from an embedded file:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatHTML,
  flags.WithTemplateBytes(templateBytes),
)
```

## Additional Template Data

`WithTemplateData` injects extra values into the template context.

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithTemplateData(map[string]any{
    "GeneratedBy": "release-tool",
  }),
)
```

Use this for build metadata, project links, or site-specific values.
Do not use it to hide parser metadata from the parser model.

## Hidden Entities

Hidden options, groups, and commands are excluded by default.

Include them explicitly for internal docs:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithIncludeHidden(true),
  flags.WithMarkHidden(true),
)
```

`WithMarkHidden` only marks hidden entities that are already included.
It does not include them by itself.

## Template Model

The template receives a parser documentation model.
The root object has:

* `.Doc`, the parser documentation model.
* `.Data`, the map passed through `WithTemplateData`.
* `.MarkHidden`, the `WithMarkHidden` value.

`.Doc` contains parser metadata, generated time, usage, positional arguments,
option groups, command groups, commands, subcommands, defaults, env metadata,
INI metadata, visibility, choices, aliases, raw tag metadata, and render forms.

Prefer reading the model through exported rendered examples and tests before
writing complex custom templates.
The model is meant for documentation output, not for application logic.

## Template Helpers

Built-in helpers cover common rendering needs:

* `i18n` resolves an i18n key with an optional fallback.
* `hiddenMark` reports whether a hidden marker should be rendered.
* `optionForms` renders short and long option forms for one option.
* `codeJoin`, `code`, `codeFenceOpen`, and `codeFenceClose`
  render markdown-friendly code fragments.
* `join`, `wrap`, `markdownWrap`, `markdownWrapIndent`, and `indent`
  format text blocks.
* `quoteMarkdown`, `quoteMan`, `manInline`, and `quoteHTML`
  escape text for the target output format.
* `defaultValue` formats a displayed default value.
* `isRequired`, `hasDefault`, `hasEnv`, `isBool`, and `isCollection`
  test option state.
* `hasOptionDefaults`, `hasOptionEnv`, and `hasNamedCommandGroups`
  help templates skip empty columns or headings.

Use helpers instead of duplicating escaping logic in templates.

## Render Style

Doc rendering uses the same render-style settings as help.

Use explicit setters when generated docs must be stable:

```go
parser.SetHelpFlagRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpEnvRenderStyle(flags.RenderStylePOSIX)
```

For a single render call, pass `WithDocRenderStyle`:

```go
err := parser.WriteDoc(
  w,
  flags.DocFormatMarkdown,
  flags.WithBuiltinTemplate(flags.DocTemplateMarkdownList),
  flags.WithDocRenderStyle(flags.RenderStylePOSIX),
)
```

Shell detection is useful for runtime help,
but generated repository files should usually use explicit styles.

## Generated Documentation Language

`WriteDoc` and the built-in `docs` command
use the same locale detection as runtime help.
The locale is read from `LC_ALL`, `LC_MESSAGES`, `LANG`,
and `LANGUAGE` environment variables in that order.
On Windows, the system UI language is used
as a fallback when none of the variables are set.

To generate documentation in a specific language, set `LANG` before the call:

```sh
# Unix
LANG=ru myapp docs markdown docs/reference.md
LANG=de myapp docs html docs/reference.html
```

```powershell
# Windows PowerShell
$env:LANG = "fr"
myapp docs markdown docs\reference.md
```

This works for all output formats - Markdown, HTML, man, and JSON.
The locale affects option descriptions, section headings,
and any i18n-tagged strings registered in the application catalog.

For reproducible generated files in a repository, fix the locale explicitly
so the output does not depend on the machine running the build:

```sh
LANG=en myapp docs markdown docs/reference.md
```

## JSON Manifest

`DocFormatJSON` serializes the doc model as a machine-readable JSON object.
It does not use templates.

```go
err := parser.WriteDoc(os.Stdout, flags.DocFormatJSON)
```

The manifest contains the full command tree with options, positional arguments,
env keys, INI keys, choices, defaults, aliases, deprecation markers,
and visibility flags.
Secret option values and choices are masked
the same way as in all other formats.

`DocFormatJSON` accepts the same `DocOption` values as other formats:

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatJSON,
  flags.WithIncludeHidden(true),
  flags.WithProgramName("myapp"),
)
```

The built-in `docs json` command exposes the same options at the CLI level,
including `--include-hidden`, `--program-name`, `--style`, and `--compact`
for single-line output without indentation:

```bash
myapp docs json
myapp docs json --compact
myapp docs json --include-hidden manifest.json
```

Use `DocFormatJSON` for documentation sites, GUI wrappers, editor integrations,
and snapshot tests that verify CLI contracts without parsing rendered text.

## Template CLI Subcommands

The built-in `docs template` command provides two subcommands
for working with templates from the command line without writing Go code.

### Export a Built-in Template

`docs template export` writes a built-in template to stdout or a file.
Use `--name` to select which template to export.

```sh
myapp docs template export --name markdown/list
myapp docs template export --name html/default > my-template.html
myapp docs template export --name man/default man-template.roff
```

Choices for `--name` come from `ListBuiltinTemplates()`
and are limited to templates that are actually embedded.

### Render with a Custom Template

`docs template render` renders the parser doc model through a template read
from a file path or from stdin (pass `-` or omit the argument to read stdin).
Output goes to stdout by default; a second positional argument writes to a file.

```sh
myapp docs template render my-template.html output.html
myapp docs template render - < my-template.html
```

Use `--format` to control which doc format is used for escaping
and the defaulttemplate selection context (default: `markdown`):

```sh
myapp docs template render --format html my-template.html output.html
```

`docs template render` accepts the same options as the other `docs` subcommands

### Edit and Re-render a Built-in Template

Export, edit, and re-render in one pipeline:

```sh
myapp docs template export --name html/default | myapp docs template render --format html
```

Or capture an intermediate copy to edit:

```sh
myapp docs template export --name markdown/list > my-template.md.tmpl
# edit my-template.md.tmpl
myapp docs template render my-template.md.tmpl docs/reference.md
```

## Project Metadata

Project metadata is set on the parser directly rather than through
`WithTemplateData`.
All built-in templates (man, HTML, Markdown) render the metadata
sections when the corresponding fields are set.

Man-page-specific fields:

```go
parser.SetManSection(8)
parser.SetSeeAlso("grep(1)", "sed(1)")
```

Fields shared with version output (set once, used by both):

```go
parser.SetVersionAuthor("Name Surname <email@example.com>")
parser.SetVersionURL("https://example.com/myapp")
parser.SetVersionBugsURL("https://github.com/example/myapp/issues")
parser.SetVersionLicense("MIT")
```

* `SetManSection` sets the section number in the `.TH` header line.
  Valid values are 1-9. The default is 1 (user commands).
  This field is man-page-only and has no effect in other formats.
* `SetSeeAlso` adds cross-reference entries to the `SEE ALSO` section
  in all formats. In man pages entries follow the `grep(1)` convention.
* `SetVersionAuthor` adds an `AUTHOR` section.
* `SetVersionURL` adds the project URL to the `SEE ALSO` section.
* `SetVersionBugsURL` adds a `BUGS` section.
* `SetVersionLicense` adds a `LICENSE` section.

Section order: `AUTHOR` → `BUGS` → `SEE ALSO` → `LICENSE`.

The metadata is also included in `DocFormatJSON` output under the `meta`
key, and is omitted from JSON when none of the fields are set explicitly.

Use `WithTemplateData` for values that are not part of the parser model.
Use these setters when the value is a standard project or version field.

## Manpage Compatibility

`WriteManPage` is kept for compatibility.
It internally uses the manpage documentation pipeline.

Prefer `WriteDoc(..., DocFormatMan, ...)` for new code because it exposes the
same template options as other formats.

## Template Rules

Keep public docs templates boring and stable.
Use custom templates to fit a documentation site,
not to change parser behavior.

Add golden tests for custom templates.
Generated docs are presentation output,
and snapshots are the fastest way to catch accidental regressions.
