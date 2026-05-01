# Localization

Localization is opt-in.
When enabled, parser help, parser errors, generated docs,
INI example metadata, and selected user text tags can resolve through catalogs.

The source text in tags remains the fallback.
This keeps the CLI usable even when a catalog key is missing.

## Parser I18n

Configure localization with `SetI18n`:

```go
//go:embed i18n/*.json
var i18nFS embed.FS

catalog, err := flags.NewJSONCatalogDirFS(i18nFS, "i18n")
if err != nil {
  return err
}

parser := flags.NewParser(&opts, flags.Default)
parser.SetI18n(flags.I18nConfig{
  Locale:      "ru",
  UserCatalog: catalog,
})
parser.SetI18nFallbackLocales("en")
```

Use `DisableI18n` to turn parser localization off again.
Use `LocaleChain` to inspect the resolved lookup order in tests
or diagnostics.

`UserCatalog` overrides built-in keys.
It can also contain application-level keys.

`ModuleCatalog` is normally left empty so the built-in module catalog is used.
Set it only when embedding a custom module catalog intentionally.

## User Text Tags

Use i18n tags when generated output should be translated:

* `description-i18n`
* `long-description-i18n`
* `group-i18n`
* `command-i18n`
* `value-name-i18n`
* `arg-name-i18n`
* `arg-description-i18n`

Example:

```go
type Options struct {
  Output string `long:"output" description:"Output file" description-i18n:"opt.output"`
}
```

The literal `description` remains useful as fallback source text.
Do not leave source text empty only because a catalog exists.

## Application Messages

Use `NewLocalizer` when application code needs the same catalog and locale
behavior without depending on a parser instance.

```go
localizer := flags.NewLocalizer(flags.I18nConfig{
  Locale:      "ru",
  UserCatalog: catalog,
})

msg := localizer.Localize("app.greeting", "Hello, {name}", map[string]string{
  "name": "Alice",
})
```

Placeholders use `{name}` syntax.
They are simple string replacements, not a full message-format engine.

## Locale Chain

Locale lookup order is:

1. explicit `I18nConfig.Locale`;
1. detected environment locale;
1. OS fallback locale where supported;
1. configured fallback locales;
1. English built-in fallback;
1. source text.

Detected environment variables include
`LC_ALL`, `LC_MESSAGES`, `LANG` and `LANGUAGE`.
Locale tokens are normalized before lookup.

## Catalog Loading

JSON catalogs can be loaded from `fs.FS`.

Use `NewJSONCatalog` when catalog bytes are already loaded.
Use `NewJSONCatalogFS` when all files are already selected.  
Use `NewJSONCatalogDirFS` when the catalog lives in one embedded directory.

The loader accepts simple locale/key maps and common go-i18n-style JSON
message objects.
The package does not require `github.com/nicksnyder/go-i18n`.

Supported JSON shapes include:

```json
{"locale":"en","messages":{"app.hello":"Hello"}}
```

```json
{"en":{"app.hello":"Hello"},"ru":{"app.hello":"Привет"}}
```

Flat message maps are also supported when the locale can be inferred from the
file name, for example `messages.en.json`.

## Stable Identifiers

Localized display names should not become stable config identifiers.

If localized group or command names participate in INI output,
set `ini-group` explicitly.
This keeps INI section names stable across locales.

For the same reason,
prefer stable i18n keys such as `command.deploy.description`
instead of keys that contain user-facing wording.

## Coverage Checks

Use catalog coverage checks in tests.
They catch missing keys and placeholder mismatches before release.

Typical checks are `CheckCatalogCoverage` and `CheckMergedCatalogCoverage`.
Use merged coverage when user catalogs override or extend module catalogs.

`CheckCatalogCoverage` works on one JSON catalog.
`CheckMergedCatalogCoverage` checks the effective result after applying
`UserCatalog` over `ModuleCatalog`.

`I18nCoverageConfig.CheckPlaceholders` also compares placeholder names.
Reports contain one `I18nCoverageReport` per locale.
Each report lists missing keys and `I18nPlaceholderMismatch` entries,
including the expected and actual placeholder sets.

## Localization Rules

Translate behavior-facing text, not internal identifiers.

Keep placeholders consistent across locales.
A translated string that drops `{name}` can produce confusing runtime messages.

Prefer concise translated help.
Long localized prose belongs in generated docs more than in terminal help.
