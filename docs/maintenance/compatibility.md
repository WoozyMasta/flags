# Compatibility

Compatibility is about what users and scripts can rely on.
For a CLI parser, compatibility includes command names, option names,
config keys, error types, help shape, and generated docs.

Not every visible string has the same stability level.

## Stable CLI Surface

Treat these as public once released:

* command names;
* command aliases;
* long option names;
* short option names;
* option aliases;
* positional argument order;
* accepted values and choices;
* environment variable names;
* INI section names and keys;
* exit behavior for help and version.

Changing any of these can break scripts.
If a change is necessary, keep aliases or compatibility paths when possible.

## Error Types vs Messages

`ErrorType` is the application-facing compatibility surface.
Message text is user-facing and may change for clarity,
localization, or better placeholders.

Applications should branch on `Error.Type`.
Tests for parser diagnostics may assert exact messages,
but application integration tests should usually avoid that.

## Help Output

Help output is user-visible, but it is also presentation.
Small layout changes may happen when width logic, render style,
localization, or sorting changes.

If your application tests exact help text,
set deterministic help width and render style.
Keep golden fixtures focused.

## Generated Docs

Generated docs are more stable when produced with explicit templates,
explicit render style, and fixed parser metadata.

If docs are committed to a repository,
regenerate them intentionally as part of release work.
Do not let shell detection or terminal width affect committed docs.

## Tag Compatibility

Singular and plural tag forms exist for compatibility in several places,
for example `default` vs `defaults` and `choice` vs `choices`.

Do not mix singular and plural forms for the same concept on one field.
The parser rejects conflicts so behavior stays predictable.

Prefer plural list tags for new code when several values are known at once.

## Built-in Option Conflicts

Built-in help and version options reserve names when enabled.
If an application already uses `-v`, enabling `VersionFlag` can conflict.

Retune the built-in option or use `VersionCommand` instead.

```go
if opt := parser.BuiltinVersionOption(); opt != nil {
  _ = opt.SetShortName('B')
}
```

## INI Stability

INI identifiers should not be localized.
Use `ini-group` and `ini-name` when display names are unstable or translated.

Changing config keys is a breaking change for users with existing config files.
Consider accepting old keys during a migration period when practical.

## Hidden Compatibility Flags

`hidden:"true"` is useful for compatibility flags that should still parse but
should not be advertised.

Hidden flags are not security.
They are only removed from normal help, completion, and docs output.

## Migration Windows

For renamed options, prefer this pattern:

1. add the new name;
1. keep the old name as an alias or hidden option;
1. document the new name;
1. remove the old path only in a breaking release.

When aliases are not enough because semantics changed,
return a clear application-level error for the old form.

## Versioning Guidance

Patch releases should avoid visible CLI behavior changes except fixes.
Minor releases can add commands, options, validators, completion and docs.
Major releases can remove or rename public CLI surface.

For pre-v1 releases,
still document compatibility-sensitive changes in the changelog.
Users run scripts against `v0` tools too.

## Release Checklist

Before releasing, review:

* changed command names;
* changed option names;
* changed env keys;
* changed INI keys;
* changed error types;
* changed generated docs templates;
* changed help snapshots.

If the change affects automation, call it out clearly in release notes.
