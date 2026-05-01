# Error Handling

Parser errors are returned as `*flags.Error`.
The important stable fields are `Type` and `Message`.
Application code should branch on `Type`, not on message text.

Messages are user-facing, localizable, and allowed to improve over time.

Some public APIs return ordinary sentinel errors instead of `*flags.Error`
when the failure is not a parse diagnostic.
Examples include `ErrNilWriter`, `ErrEmptyCommandName`,
and invalid presentation limits.
Use `errors.Is` for those errors and `errors.As` for `*flags.Error`.

## Basic Pattern

With `flags.Default`, parser errors are printed automatically.
Do not print the same error again unless you disabled `PrintErrors`.

```go
_, err := parser.Parse()
if err != nil {
  if flags.WroteHelp(err) {
    os.Exit(0)
  }

  var ferr *flags.Error
  if errors.As(err, &ferr) && ferr.Type == flags.ErrVersion {
    os.Exit(0)
  }
  os.Exit(1)
}
```

For libraries and tests, disable automatic output:

```go
parser := flags.NewParser(&opts, flags.Default&^flags.PrintErrors)
_, err := parser.ParseArgs(args)
```

`WroteHelp` is a convenience helper for the common `ErrHelp` case.
It is safe to call with `nil` or non-parser errors.
It only reports help output,
not version output.

## Error Types

* `ErrUnknown` is a generic wrapped error.
  It usually means a lower-level error was not already a parser error.
* `ErrExpectedArgument` means an option needed a value but no value was present.
* `ErrUnknownFlag` means the command line used an unknown option.
* `ErrUnknownGroup` means a group lookup failed.
  This is mostly relevant to programmatic APIs.
* `ErrMarshal` means value conversion failed.
  This includes invalid numbers, custom unmarshaler errors,
  I/O template normalization errors, and similar conversion problems.
* `ErrHelp` means built-in help was requested.
  Treat it as successful control flow.
* `ErrVersion` means built-in version output was requested.
  Treat it as successful control flow.
* `ErrNoArgumentForBool` means a bool option received an argument while
  `AllowBoolValues` was not enabled.
* `ErrRequired` means a required option or argument was not satisfied after
  defaults and environment values were applied.
* `ErrShortNameTooLong` means a short option tag contained more than one rune.
* `ErrDuplicatedFlag` means two options registered the same short
  or long name in the same valid scope.
* `ErrTag` means a generic tag parse failure.
* `ErrCommandRequired` means a command or subcommand was required but omitted.
* `ErrUnknownCommand` means a command token did not match a known command.
* `ErrInvalidChoice` means a parsed value was outside `choice` or `choices`.
* `ErrInvalidTag` means a tag exists but is invalid for that field,
  value, or combination.
* `ErrOptionConflict` means an `xor` relation was violated.
* `ErrOptionRequirement` means an `and` or required relation was violated.
* `ErrValidation` means a `validate-*` post-parse validator failed.

`ErrorType.String()` returns stable short names such as `unknown flag`,
`required`, and `validation`.
`ErrorType.Error()` returns the same string,
so an `ErrorType` can be used directly as an `error` value when needed.

## Warnings vs Critical Errors

`ErrorType.IsWarning()` currently marks required-option
and required-command failures as warning-level.
Colorized error output can use this distinction.

Do not use warning status to decide exit codes automatically.
A missing required option is still a parse failure.

## Output Routing

* `PrintErrors` prints parser errors.
* `PrintHelpOnStderr` routes help output to stderr.
* `PrintErrorsOnStdout` routes non-help errors to stdout.
* `PrintHelpOnInputErrors` prints help before common input errors.

`PrintHelpOnInputErrors` applies to common user-input failures:
missing required values, missing required commands, unknown flags,
unknown commands, missing option arguments, invalid choices,
and bool values passed while `AllowBoolValues` is disabled.
It does not print help for version output or application errors.

Use these options for final applications.
Libraries should normally return errors and let callers route output.

## Wrapping Application Errors

Command `Execute` methods and `CommandHandler` may return ordinary Go errors.
The parser returns those errors to the caller.

Wrap application errors with context in normal Go style:

```go
return fmt.Errorf("open config: %w", err)
```

Do not convert application errors into `*flags.Error`
unless the error is truly part of parser behavior.

See [Handlers and Integration Points][] for command execution hooks.

## Testing Errors

Prefer asserting `Error.Type` in tests.

Exact message tests are useful for parser diagnostics,
but they are intentionally brittle.
Keep them close to the feature that owns the message.

For application tests, message substring checks are usually enough.

[Handlers and Integration Points]: handlers.md
