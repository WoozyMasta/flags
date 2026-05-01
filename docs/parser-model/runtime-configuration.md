# Runtime Configuration

Tags describe stable CLI contracts.
Runtime configuration changes parser metadata from Go code.
Use it when metadata is generated, loaded from another system,
or shared across several parser shapes.

## Construction and Parse Helpers

Use package-level helpers for small programs:

```go
_, err := flags.Parse(&opts)
_, err = flags.ParseArgs(&opts, []string{"--verbose"})
```

Both helpers create a parser with `flags.Default`.
Use `NewParser` or `NewNamedParser` when code needs parser options,
runtime metadata changes, output generation, or tests with a stable name.

```go
parser := flags.NewNamedParser("myapp", flags.Default)
_, err := parser.ParseArgs(args)
```

`NewParser` uses `os.Args[0]` as the application name.
`NewNamedParser` uses the provided name.

`IniParse` is the matching convenience helper for reading an INI file into
a struct with default parser behavior.

## Configurer

Implement `flags.Configurer` on an options, group, or command struct:

```go
type Options struct {
  Verbose bool `long:"verbose"`
  Run     struct{} `command:"run" description:"Run workload"`
}

func (o *Options) ConfigureFlags(p *flags.Parser) error {
  if opt := p.FindOptionByLongName("verbose"); opt != nil {
    opt.AddLongAlias("debug")
    _ = opt.SetEnv("APP_VERBOSE", "")
  }

  if cmd := p.Find("run"); cmd != nil {
    cmd.AddAlias("execute")
    cmd.SetShortDescription("Execute workload")
  }

  return nil
}
```

`ConfigureFlags` runs before parsing when parser topology
or tag mapping has changed.
It is the right place for small post-scan adjustments.

Keep it focused.
Large hidden configuration blocks make the CLI harder to understand than tags.

## Manual Validation and Rebuild

Use `parser.Validate()` when code needs to force configurators
and duplicate metadata checks before parsing.

Use `parser.Rebuild()` after programmatic changes that require rescanning.
Tag remapping methods already rebuild attached groups and commands.

Most metadata setters do not require `Rebuild`.
They update the existing parser model directly.
Use `Rebuild` after changes that affect scanned struct tags or topology.

## Lookup APIs

`Parser` embeds the root `Command`,
so command lookup APIs are available on the parser itself.

Find options by visible names:

```go
opt := parser.FindOptionByLongName("verbose")
short := parser.FindOptionByShortName('v')
```

Long-name lookup includes group namespaces and long aliases.
Short-name lookup includes short aliases.
Command lookup uses command names and aliases:

```go
cmd := parser.Find("deploy")
```

`Command.FindOptionByLongName` and `Command.FindOptionByShortName`
search the command and its parent commands.
`Group.FindOptionByLongName` and `Group.FindOptionByShortName`
search the group and its subgroups.

Use `Command.Commands()` for direct child commands.
Use `Command.Args()` for command positional arguments.
Use `Group.Groups()`, `Group.Options()`, and `Group.Data()`
to inspect group contents and the backing struct pointer.

## Programmatic Creation

`Parser.AddGroup` adds an option group to the root parser.
`Command.AddGroup` adds command-local options.
`Group.AddGroup` adds nested option groups.

```go
var global struct {
  Verbose bool `long:"verbose"`
}

_, err := parser.AddGroup("Global Options", "", &global)
```

`Parser.AddCommand` and `Command.AddCommand` add commands from data structs:

```go
var deploy struct {
  DryRun bool `long:"dry-run"`
}

_, err := parser.AddCommand("deploy", "Deploy release", "", &deploy)
```

`Group.AddOption` attaches an already-built `Option` to a group.
Prefer struct tags or `AddGroup` for normal code.
Use `AddOption` only for low-level integrations that already own an `Option`.

## Tag Remapping

`SetTagPrefix` applies a prefix to all parser tags.

```go
type Config struct {
  Path string `flag-short:"p" flag-long:"path"`
}

parser := flags.NewParser(&cfg, flags.Default)
_ = parser.SetTagPrefix("flag-")
```

`SetFlagTags` customizes individual tag names.
Use it when only one or two tags conflict with another library.

```go
tags := flags.NewFlagTags()
tags.Short = "cli-short"
_ = parser.SetFlagTags(tags)
```

`SetTagListDelimiter` changes list tag splitting for tags such as
`defaults`, `choices` and `aliases`.

## Parser Setters

Parser setters control global parser behavior.
Common examples:

* `SetTagPrefix`, `SetFlagTags`, and `SetTagListDelimiter`
  for struct-tag mapping.
* `SetEnvPrefix` for global environment prefixes.
* `SetHelpWidth` for help wrapping width.
* `SetMaxLongNameLength` for long option name limits.
* `SetOptionSort` for option render order.
* `SetCommandSort` for command render order.
* command metadata batch setters:
  `SetCommandShortDescriptions`, `SetCommandLongDescriptions`,
  `SetCommandDescriptions`, `SetCommandShortDescriptionI18nKeys`,
  `SetCommandLongDescriptionI18nKeys` and `SetCommandDescriptionI18nKeys`.
* `SetOptionTypeOrder` for type-based option sorting.
* `SetBuiltinCommandGroup` for built-in command display grouping.
* `SetCommandOptionIndent` for command option indentation in help.
* `SetHelpFlagRenderStyle` for flag token rendering.
* `SetHelpEnvRenderStyle` for environment placeholder rendering.
* `SetHelpColorScheme` and `SetErrorColorScheme` for ANSI color roles.
* `SetI18n` and `SetI18nFallbackLocales` for localization.
* version setters such as `SetVersionInfo`, `SetVersion`,
  `SetVersionCommit`, `SetVersionTime`, `SetVersionURL`,
  `SetVersionTarget`, and `SetVersionFields`.

Use parser setters for behavior that belongs to the whole CLI, not one field.

## Option Setters

Option setters change one option's metadata.
They can rename an option, add aliases, set defaults, set env keys,
set choices, set required/hidden state, or adjust ordering.

Find options by canonical name:

```go
if opt := parser.FindOptionByLongName("verbose"); opt != nil {
  opt.SetDescription("Show verbose output")
}
```

When a setter can fail, check the error.
Errors usually mean invalid metadata, duplicate names,
or a value that violates parser constraints.

Useful runtime accessors:

* `Value()` returns the current Go value.
* `Field()` returns the reflected struct field.
* `IsSet()` reports whether parsing explicitly set the option.
* `IsSetDefault()` reports whether the current value came from defaults.
* `LongNameWithNamespace()`, `LongAliasesWithNamespace()`,
  and `EnvKeyWithNamespace()` return rendered names with namespaces applied.

`Set(value)` converts and assigns one option value programmatically.
It marks the option as set and prevents later parser defaults from replacing it.
Use it when code wants the same conversion path as command-line parsing.

## Command, Group, and Arg Setters

Commands expose setters for names, aliases, descriptions, command groups,
visibility, argument requirements, and other command-local metadata.
They also expose `SetOrder`, `SetIniName`, `SetSubcommandsOptional`,
and `SetPassAfterNonOption`.

Groups expose setters for descriptions, namespaces, INI section names,
visibility, and immediate behavior.

Arguments expose setters for display name, description, defaults,
completion hints, and validation-related metadata.
`Arg.SetRequiredRange` uses the same min/max model as repeatable required
option ranges.

Use these setters when metadata is not naturally part of the struct tag.
For stable public names, prefer tags.

## Sorting

Sorting affects help, completion, and generated documentation.
It does not change parse behavior.

Options default to declaration order.
Command lists default to ascending command name sorting.

Use `order` tags for item-level priority.
Positive order moves an item toward the top.
Negative order moves it toward the bottom.
Zero keeps normal sorting behavior.
Within the same priority bucket,
higher positive values are shown before lower positive values.

`SetOptionSort` accepts:

* `OptionSortByDeclaration`, preserving struct declaration order;
* `OptionSortByNameAsc`, sorting by rendered option name ascending;
* `OptionSortByNameDesc`, sorting by rendered option name descending;
* `OptionSortByType`, sorting by option type class,
  then by rendered option name.

`OptionSortByType` uses `OptionTypeClass` ranks:

* `OptionTypeBool`
* `OptionTypeNumber`
* `OptionTypeString`
* `OptionTypeDuration`
* `OptionTypeCollection`
* `OptionTypeCustom`

Use `SetOptionTypeOrder` to override that rank.
The provided list may be partial.
Missing classes keep their default relative order after the provided classes.
Duplicate or unknown classes return an error.

```go
err := parser.SetOptionTypeOrder([]flags.OptionTypeClass{
  flags.OptionTypeString,
  flags.OptionTypeBool,
})
```

`SetCommandSort` accepts:

* `CommandSortByDeclaration`, preserving struct declaration order;
* `CommandSortByNameAsc`, sorting by command name ascending;
* `CommandSortByNameDesc`, sorting by command name descending.

Use parser sorting modes for broad policies.
Do not mix many explicit orders with a sort policy unless
the output has a clear reason to be curated.

## Help Layout

`SetHelpWidth` controls wrapping for built-in help.
When unset,
the parser uses detected terminal width with a fallback of 80 columns.
Width `0` disables wrapping.
Negative widths return `ErrNegativeHelpWidth`.

`SetCommandOptionIndent` adds extra spaces before command option rows
in built-in help output.
Use it when command-local options should be visually nested under commands.
Negative indentation returns `ErrNegativeCommandOptionIndent`.

Use explicit help width and render styles in generated docs and golden tests:

```go
_ = parser.SetHelpWidth(80)
parser.SetHelpFlagRenderStyle(flags.RenderStylePOSIX)
parser.SetHelpEnvRenderStyle(flags.RenderStylePOSIX)
```

## Runtime Mutation Rules

Use runtime configuration to integrate with application code.
Do not use it to hide a static CLI contract.

A reader should be able to understand the basic command-line interface
from the struct alone.
Runtime configuration should fill in dynamic details, not redefine everything.
