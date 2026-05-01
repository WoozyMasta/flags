# Commands

Commands split one executable into named actions.
They are the CLI equivalent of verbs such as
`add`, `remove`, `deploy`, or `config set`.

A command can own options, positional arguments, child commands,
and execution behavior.

## Struct Commands

A struct field tagged with `command` becomes a command.

```go
type AddCommand struct {
  Name string `long:"name" required:"true"`
}

type Options struct {
  Add AddCommand `command:"add" description:"Add an item"`
}
```

Command-local options become valid after the command token is parsed.
If `--name` belongs to `add`, then `app add --name x` is valid,
but `app --name x add` is not.

Global options belong outside command structs.
They remain valid before or after the selected command token.

## Programmatic Commands

Commands can also be registered with `Parser.AddCommand`.
Use this for plugin systems, late-bound command metadata,
or commands that are not represented by the main options struct.

Prefer struct commands for normal applications.
They keep the CLI shape visible and testable.

## Command Execution

A command can implement `Commander`:

```go
type AddCommand struct {
  Name string `long:"name" required:"true"`
}

func (c *AddCommand) Execute(args []string) error {
  fmt.Println("add", c.Name, args)
  return nil
}
```

After parsing succeeds, the parser calls `Execute` for the selected command.
The `args` slice contains remaining arguments that were not consumed.

By default only the selected leaf command executes.
Parent commands are used for structure and shared options,
but their `Execute` methods are not called.

## Command Chain

Enable `CommandChain` when every active command from parent to leaf should run.

```go
parser := flags.NewParser(&opts, flags.Default|flags.CommandChain)
```

The chain runs parent to leaf.
If a command returns an error, the chain stops and parse returns that error.

`CommandChain` is not a sibling command pipeline.
It does not allow `app cmd1 cmd2 cmd3` as three same-level actions.
It only affects the already selected command path,
for example `app project deploy run`.

Use it for setup/teardown style command trees,
not for arbitrary command batching.

## Command Handler

`Parser.CommandHandler` intercepts command execution.
It receives the selected command and remaining args.

Use it when execution needs dependency injection, centralized logging,
transaction boundaries, or test instrumentation.

If a custom handler is set,
it is responsible for calling `Execute` when that is desired.

See [Handlers and Integration Points][] for the full handler lifecycle.

## Aliases

`alias` adds one alternative command name.
`aliases` adds several alternatives through the tag-list delimiter.

Aliases are useful for compatibility and common abbreviations.
Do not overuse them. Every alias is another public entry point to support.

## Optional Subcommands

`subcommands-optional` lets a command with child commands be selected without
requiring one of its children.

Use it when the parent command has useful behavior on its own.
Avoid it when the parent is only a namespace,
because ambiguous command trees are harder to explain in help output.

## Command Groups

`command-group` changes how commands are grouped in help and generated docs.
It does not change parsing.

Use command groups when one command list contains mixed categories,
such as user commands and maintenance commands.

## Built-in Commands

Built-in commands are opt-in parser options.
They can expose help, version, completion, documentation, and config output.

`HelpCommands` enables the full built-in command set.
Individual bits can be used when only part of that set should be public.

Built-in commands are normal command entry points from a user perspective.
They are usually marked immediate internally,
so they can run without unrelated required options blocking them.

## Pass-Through Commands

`pass-after-non-option` changes command-local parsing so option parsing stops
after the first non-option argument.

This is useful for wrapper commands:

```go
type Options struct {
  Exec struct {
    Program string `positional-arg-name:"program" required:"true"`
    Args    []string
  } `command:"exec" pass-after-non-option:"true" positional-args:"true"`
}
```

The wrapped program can then receive arguments that would otherwise look like
options for the outer CLI.

## Command Design Rules

Keep command names verb-like. Keep command structs small.
Move shared settings into parent command structs or option groups.

If a command requires complex runtime dependencies,
put parsing in the command struct and execution wiring in `CommandHandler`.
That keeps command-line parsing independent from application construction.

[Handlers and Integration Points]: ../parser-model/handlers.md
