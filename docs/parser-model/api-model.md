# API Model

The public parser model is built around five main objects:
`Parser`, `Command`, `Group`, `Option` and `Arg`.

Struct tags create these objects during parser scanning.
Runtime APIs can inspect and tune them after scanning.

## Parser

`Parser` is the root object.
It owns parser options, top-level command state, help rendering settings,
i18n settings, tag mapping, completion handling, and command execution handling.

Create one with:

```go
parser := flags.NewParser(&opts, flags.Default)
```

or:

```go
parser := flags.NewNamedParser("myapp", flags.Default)
```

Use `NewNamedParser` when the executable name should be stable in tests,
generated docs, or embedded use.

## Command

`Command` represents a command scope. The root parser embeds a root command.
Subcommands are child commands.

Commands can own groups, options, positional arguments, and child commands.

A command struct can implement `Commander`:

```go
func (c *RunCommand) Execute(args []string) error {
  return nil
}
```

By default, only the selected leaf command executes.
`CommandChain` executes active commands from parent to leaf.

## Group

`Group` represents a group of options.
Groups are used for help organization, metadata, namespaces,
environment namespaces, and INI sections.

Groups can be created by struct tags or `Parser.AddGroup`.

Use groups for user-facing organization,
not only internal code organization.

## Option

`Option` represents one flag-like value.
It stores names, aliases, defaults, environment metadata, choices,
validation config, I/O config, render metadata, and the reflect value target.

Find options through parser and group lookup helpers, for example:

```go
opt := parser.FindOptionByLongName("verbose")
```

Use setters to adjust metadata programmatically.
Check errors from setters that can violate parser constraints.

## Arg

`Arg` represents a positional argument.
It stores name, description, defaults, required range, completion hint,
I/O metadata, validation config, and the reflect target.

Access command args with `Command.Args()`.
The returned slice is a copy of the internal slice,
so callers cannot mutate command structure accidentally by editing the slice.
Use setters on the `Arg` objects themselves for metadata changes.

## Scanning and Rebuild

The parser scans struct tags when groups and commands are added.
Some runtime changes require rescanning.

`SetTagPrefix`, `SetFlagTags`, `SetTagListDelimiter`,
and `SetMaxLongNameLength` rebuild attached groups and commands.

Use `Parser.Rebuild()` after programmatic changes that require a full rebuild.
Use `Parser.Validate()` to run configurators and duplicate metadata checks
without parsing user arguments.

## Configurer

`Configurer` lets a struct tune parser metadata after scan.

```go
func (o *Options) ConfigureFlags(p *flags.Parser) error {
  if opt := p.FindOptionByLongName("verbose"); opt != nil {
    opt.AddLongAlias("debug")
  }
  return nil
}
```

Use it for small runtime adjustments.
Do not use it to hide the entire CLI contract from struct tags.

## Lookup Rules

Option name conflicts are checked in valid scopes.
Sibling commands can reuse option names.
Parent and active child scopes must avoid ambiguous names.

Aliases participate in duplicate checks.
Built-in help and version flags also participate when enabled.

Command names and aliases are checked among commands in the same command scope.

## Handlers

`UnknownOptionHandler` customizes unknown option behavior.

`CompletionHandler` customizes completion output handling.

`CommandHandler` customizes command execution. If it is set,
it is responsible for calling command `Execute` when that is desired.

Use handlers for integration points,
not for ordinary field parsing.

## API Model Rules

Keep the struct as the source of truth for static CLI shape.
Use the API model to inspect, tune,
and integrate that shape with application runtime.

If code needs many runtime mutations, consider whether the CLI
is generated or whether the struct tags are simply underused.
