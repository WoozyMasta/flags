# Examples

The repository includes examples for common parser setups.
Use them to verify end-to-end behavior before copying isolated snippets.

## Basic Example

Path: `examples/basic`

This is the smallest multi-command application example.
It demonstrates:

* parser construction;
* root options;
* command structs;
* `Execute(args []string)`;
* command-local options;
* simple defaults.

Run it from the repository root:

```bash
go run ./examples/basic --help
```

Use this example when learning the normal application shape.

## Advanced Example

Path: `examples/advanced`

This example exercises most user-facing features in one app.
It demonstrates:

* built-in helper commands;
* help and docs rendering;
* command and option sorting;
* environment and INI flows;
* dynamic defaults;
* custom value parsing;
* colors and render styles.

Run it:

```bash
go run ./examples/advanced --help
```

Use this example when checking how features interact.

## Custom Tag Names

Path: `examples/custom-flag-tags`

This example demonstrates tag remapping through:

* `SetTagPrefix`;
* `SetFlagTags`;
* prefixed struct tags;
* mixed custom and default tag names.

Run it:

```bash
go run ./examples/custom-flag-tags --help
```

Use this example when a struct is shared with another tag-based package.

## I18n Example

Path: `examples/i18n`

This example demonstrates:

* embedded JSON catalogs;
* parser localization;
* localized help;
* localized parser errors;
* `Localizer` for application messages.

Run it:

```bash
go run ./examples/i18n --help
```

Use this example when adding translations to a CLI.

## Completion Templates

Completion registration templates are kept under `examples/completion`.
They are shell-side examples for installing generated completion scripts.

The parser can generate scripts with:

```go
parser.WriteNamedCompletion(os.Stdout, flags.CompletionShellBash, "myapp")
```

Use the templates as integration examples,
not as parser behavior tests.

## Rendered Documentation Snapshots

Rendered docs snapshots are kept under `examples/doc-rendered`
when generated outputs are present.
They are useful as reference output for markdown, HTML, and man templates.

Generated docs are presentation artifacts.
If a template changes intentionally,
review the rendered diff before accepting it.

## Choosing an Example

* Use `examples/basic` when starting a new CLI.
* Use `examples/advanced` when integrating multiple features.
* Use `examples/i18n` when localizing.
* Use `examples/custom-flag-tags` when sharing structs with other libraries.
* Use completion templates when wiring shell integration.

## Keeping Examples Useful

Examples should stay small enough to read.
If an example needs too much explanation,
prefer a dedicated docs page plus a focused example.

When adding a new feature, add either:

* a targeted test for behavior;
* a small example if users need to see integration;
* a cookbook recipe if the feature combines existing APIs.
