# Handlers and Integration Points

Handlers are escape hatches around the parser lifecycle.
Use them when parsing is still owned by `flags`,
but integration with the application needs custom behavior.

Prefer tags and parser options for normal CLI structure.
Use handlers for boundaries: unknown option policy, completion capture,
command execution wiring, and metadata customization.

## Unknown Option Handler

`Parser.UnknownOptionHandler` runs when parsing sees an unknown option
and `IgnoreUnknown` is not enabled.

The handler receives:

* the unknown option name without the option prefix;
* a `SplitArgument` for inline values such as `--opt=value`;
* the remaining command-line args.

It returns the replacement remaining arg list,
or an error to stop parsing.

```go
parser.UnknownOptionHandler = func(
  option string,
  arg flags.SplitArgument,
  args []string,
) ([]string, error) {
  if value, ok := arg.Value(); ok {
    return append([]string{"--" + option + "=" + value}, args...), nil
  }

  return append([]string{"--" + option}, args...), nil
}
```

Use this when unknown options need domain-specific rewriting.
Use `IgnoreUnknown` when unknown options should simply become remaining args.

## Completion Handler

`Parser.CompletionHandler` receives completion candidates in raw completion
mode.
Without a custom handler,
the parser prints candidates and exits the process.

```go
var got []flags.Completion
parser.CompletionHandler = func(items []flags.Completion) {
  got = append(got, items...)
}
```

Use this in tests,
embedded tools,
or applications that need to collect candidates without process exit.
After a custom completion handler runs,
`ParseArgs` returns `nil, nil`.

## Command Handler

`Parser.CommandHandler` intercepts command execution.
It receives the selected `Commander` and the remaining args.

```go
parser.CommandHandler = func(command flags.Commander, args []string) error {
  if command == nil {
    return nil
  }

  log.Printf("run command with %d remaining args", len(args))
  return command.Execute(args)
}
```

If a custom handler is set,
it owns execution.
The parser does not call `Execute` automatically after the handler returns.

The command can be `nil` when parsing finishes without an executable command.
This lets one handler own both command and no-command cases.

With `CommandChain`,
the handler is called for each active command that implements `Commander`,
in parent-to-leaf order.

## Commander

`Commander` is implemented by command data structs:

```go
type Commander interface {
  Execute(args []string) error
}
```

`Execute` runs after parsing, defaults, validation,
required checks, and option relation checks.
The `args` slice contains remaining tokens that were not consumed by parsing.

Return ordinary Go errors from application code.
Parser errors should normally be produced by the parser,
not by command implementations.

## Usage

`Usage` lets command data customize the usage fragment shown in help
and generated docs:

```go
type Usage interface {
  Usage() string
}
```

Use it when a command has a compact domain-specific usage shape
that tags cannot express clearly.
Keep the returned text stable,
because users may copy it into scripts or documentation.

## Configurer

`Configurer` lets data structs adjust parser metadata after tag scanning:

```go
type Configurer interface {
  ConfigureFlags(parser *flags.Parser) error
}
```

Use it for programmatic metadata that cannot be expressed cleanly in tags:
dynamic descriptions, generated choices, custom grouping,
or shared conventions across several commands.

`ConfigureFlags` is called when parser topology changes,
for example after adding groups or commands,
or after changing the active tag mapping.

Keep `ConfigureFlags` deterministic.
It should configure parser metadata,
not perform application startup work.

## Integration Rules

Prefer normal parser features first.
Handlers are powerful,
but they make control flow less visible from struct tags.

Keep handlers small and test them directly.
When a handler changes parse behavior,
add tests for both accepted and rejected command lines.

Avoid side effects before completion mode can return.
Completion invokes the application often during shell interaction.
