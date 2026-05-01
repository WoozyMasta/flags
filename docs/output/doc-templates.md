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

```go
err := parser.WriteDoc(os.Stdout, flags.DocFormatMarkdown)
```

The format controls escaping and default template selection.
A custom template can still be supplied for any supported format.

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
* `join`, `wrap`, `markdownWrap`, and `indent` format text blocks.
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

Shell detection is useful for runtime help,
but generated repository files should usually use explicit styles.

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
