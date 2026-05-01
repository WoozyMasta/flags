# Custom Values

Custom value APIs let application types own parsing,
formatting, completion, validation, and dynamic defaults.

Use them when a value has domain syntax that is more specific than a simple
string, number, or enum.

## Built-in Conversion

The parser directly supports common scalar types:
strings, bools, signed and unsigned integers, floats, and `time.Duration`.

Slices append values. Maps parse key/value entries.
Pointers are allocated when a value is provided.

Integer parsing honors `base`. Map parsing honors `key-value-delimiter`.

## Unmarshaler

Implement `flags.Unmarshaler` when a type should parse a CLI string itself.

```go
type Level int

func (l *Level) UnmarshalFlag(value string) error {
  switch value {
  case "debug":
    *l = 10
  case "info":
    *l = 20
  default:
    return fmt.Errorf("unknown level %q", value)
  }
  return nil
}
```

`UnmarshalFlag` receives the raw string after option parsing
has selected the value. It should return a clear error for invalid input.
The parser wraps conversion failures as `ErrMarshal`.

## TextUnmarshaler

Types implementing `encoding.TextUnmarshaler` are also supported.

Use the standard interface when the type is useful outside this package.
Use `flags.Unmarshaler` when the syntax is specifically a CLI flag syntax.

The parser checks flags-specific interfaces before standard text interfaces.

## Marshaler and TextMarshaler

`flags.Marshaler` and `encoding.TextMarshaler` control conversion back to text.
They are used by help/default rendering, INI writing,
and other places where current values become strings.

```go
func (l Level) MarshalFlag() (string, error) {
  switch l {
  case 10:
    return "debug", nil
  case 20:
    return "info", nil
  default:
    return "unknown", nil
  }
}
```

Keep marshal output parseable by the matching unmarshal logic when possible.
That makes INI and generated examples safer.

## ValueValidator

`ValueValidator` checks a raw option argument before conversion.

```go
type ExistingUser string

func (u *ExistingUser) IsValidValue(value string) error {
  if strings.HasPrefix(value, "-") {
    return fmt.Errorf("user name must not look like an option")
  }
  return nil
}
```

This hook is useful when an option value might look like another option,
or when a raw string check should happen before normal conversion.

For ordinary string, path, and numeric constraints, prefer `validate-*` tags.
They are more visible in the CLI contract and localize parser diagnostics.

## DefaultProvider

`DefaultProvider` supplies dynamic default strings during parsing.

```go
type RuntimeDefault struct{}

func (RuntimeDefault) Default() ([]string, error) {
  if value := os.Getenv("APP_DEFAULT_REGION"); value != "" {
    return []string{value}, nil
  }
  return []string{"eu-west-1"}, nil
}
```

Use dynamic defaults for values that are cheap,
local, and stable during parser setup.
Do not use them for slow network calls.

A `DefaultProvider` returns strings because defaults
are applied through the same conversion pipeline as tag defaults.

## Completer

Implement `Completer` for custom shell completion.

```go
type Region string

func (r *Region) Complete(match string) []flags.Completion {
  values := []string{"eu-west-1", "us-east-1", "ap-south-1"}
  out := make([]flags.Completion, 0, len(values))
  for _, value := range values {
    if strings.HasPrefix(value, match) {
      out = append(out, flags.Completion{Item: value})
    }
  }
  return out
}
```

Completion candidates can include descriptions.
Completion runs by executing the program,
so keep completion code side-effect-light and fast.

The built-in `Filename` type is a small example of custom completion.

## Function Options

Function fields can be used as options.
A zero-argument function behaves like a bool-style trigger.
A one-argument function receives a converted value.
If the function returns an error, that error becomes the parse error.

Use function options sparingly. They can hide side effects inside parsing,
which makes tests and help/version flows harder to reason about.

## Slices, Maps, and Custom Types

For slices, conversion is applied to each provided value and appended.
If the element type implements custom conversion,
that element conversion is used.

For maps,
the key and value are each converted through the normal conversion pipeline.
This allows typed keys and typed values.

Keep custom map syntaxes simple.
Complex configuration belongs in a config file, not a single CLI option.

## Choosing Extension Points

Prefer standard Go interfaces when the type is generally reusable.
Prefer `flags` interfaces when behavior is CLI-specific.

Use custom values for domain syntax.  
Use `choices` for small static sets.  
Use validators for simple constraints.  
Use application-level validation for cross-field rules.
