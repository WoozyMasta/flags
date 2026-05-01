# Value Validators

Value validators are post-parse checks declared through `validate-*` tags.
They apply to option fields and positional argument fields.

Validators run after values are assigned from command-line input,
default tags, environment variables, and other configured default sources.

## Supported Targets

String validators support `string` and `[]string` fields.
For slices, every element is validated.

Numeric validators support signed integers, unsigned integers,
and floats. Numeric slices are also supported.

A validator used on the wrong target type is an invalid tag error.
This catches configuration mistakes at parser construction or parse time.

## String Validators

`validate-non-empty:"true"` rejects strings that
are empty after trimming spaces.
Use it when an empty string is never meaningful.

`validate-regex:"PATTERN"` requires the whole string
to match the regular expression.
The match is full-value, not substring-based.

```go
type Options struct {
  Name string `long:"name" validate-non-empty:"true" validate-regex:"[a-z]+"`
}
```

`validate-min-len:"N"` requires at least `N` runes.

`validate-max-len:"N"` requires at most `N` runes.

Length validators count runes, not bytes.

## Path Validators

`validate-existing-file:"true"` requires an existing regular file.
Directories fail this check.

`validate-existing-dir:"true"` requires an existing directory.
Files fail this check.

`validate-readable:"true"` requires the path to be readable by the current
process.

`validate-writable:"true"` checks that an existing path is writable,
or that the parent directory of a missing path is writable.
This supports output-file use cases where the file may not exist yet.

`validate-path-abs:"true"` requires an absolute path according to the current
platform.

```go
type Options struct {
  Input  string `long:"input" validate-existing-file:"true" validate-readable:"true"`
  Output string `long:"output" validate-writable:"true" validate-path-abs:"true"`
}
```

Path validators work on strings.
They do not expand shells, open final output files,
or protect against every filesystem race.
Application code still owns final file operations.

## Numeric Validators

`validate-min:"N"` requires a numeric value greater than or equal to `N`.

`validate-max:"N"` requires a numeric value less than or equal to `N`.

The tag value is parsed according to the target type.
For unsigned fields, negative bounds are invalid tags.
For floats, floating-point bounds are accepted.

```go
type Options struct {
  Retries int     `long:"retries" validate-min:"0" validate-max:"10"`
  Ratio   float64 `long:"ratio" validate-min:"0" validate-max:"1"`
}
```

## Defaults Are Validated

Defaults are not exempt.
If a default violates a validator,
parsing fails even when the user did not pass the option.

```go
type Options struct {
  Name string `long:"name" default:"x" validate-min-len:"2"`
}
```

This fails because the configured default does not satisfy the declared
contract.

## Optional Missing Values

An unset option with no default and no environment value
is skipped when its zero value is still empty.

Once a value source sets the option, validators run.
This avoids rejecting optional fields merely because their zero value
would not satisfy a rule intended for provided values.

## Error Messages

Validator failures use `ErrValidation`.
Messages are localizable through the parser catalog.

Regex failures include the required pattern.
This matters for CLI usability:
the user should see what shape the value must match.

## Combining Validators

Validators can be combined when each rule describes the same contract.

```go
type Options struct {
  Slug string `long:"slug" validate-non-empty:"true" validate-regex:"[a-z0-9-]+" validate-max-len:"32"`
}
```

Keep combinations readable. If a field needs many validators,
consider a domain-specific type with custom parsing or validation.

## When Not to Use Tag Validators

Do not use tag validators for rules that depend on remote services,
current user permissions beyond a simple filesystem check,
or relationships between multiple fields.

Use application-level validation for cross-field and domain rules.
Tag validators should stay local to one value.
