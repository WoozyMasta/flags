# Cookbook

This page collects practical CLI recipes.
Each recipe shows a common shape and explains when to use it.
For full behavior details, follow the `Related` links under each recipe.

## Minimal Application Options

Use this when the program has a few global settings and no commands.

```go
type Options struct {
  Verbose bool   `short:"v" long:"verbose" description:"Show verbose output"`
  Output  string `short:"o" long:"output" value-name:"FILE" description:"Output file"`
}

var opts Options
parser := flags.NewParser(&opts, flags.Default)
_, err := parser.Parse()
```

This is the simplest shape.
Start here before adding commands, config files, or custom parsing.

Related: [Getting Started][] and [Options][].

## Required Value with a Default

Use this when the application must always receive a value,
but a safe fallback exists.

```go
type Options struct {
  Region string `long:"region" required:"true" default:"eu-west-1"`
}
```

The required check passes because the default supplies a value.
Use this for values that must exist in application code,
not necessarily values that the user must type every time.

Related: [Defaults and Configuration][] and [Options][].

## Environment Override

Use this when deployment-specific values should come from environment.

```go
type Options struct {
  Token string `long:"token" env:"APP_TOKEN" required:"true" default-mask:"***"`
}
```

The user can pass `--token`, or the process can provide `APP_TOKEN`.
The displayed default is masked.

Use this for secrets and deployment configuration.
Avoid logging the parsed value.

Related: [Environment][] and [Visibility and Secrets][].

## Repeated Include Paths

Use a slice when an option can appear multiple times.

```go
type Options struct {
  Include []string `short:"I" long:"include" value-name:"DIR"`
}
```

Example:

```bash
app -I ./include -I ./vendor/include
```

The field receives both values in order.

Related: [Options][].

## Key/Value Labels

Use a map for repeated key/value input.

```go
type Options struct {
  Label map[string]string `long:"label" key-value-delimiter:"="`
}
```

Example:

```bash
app --label env=prod --label tier=api
```

Keep map syntaxes simple.
If users need nested config, use a config file instead.

Related: [Options][] and [Defaults and Configuration][].

## Verbose Counter

Use a counter when repeated flags should become a level.

```go
type Options struct {
  Verbose int `short:"v" long:"verbose" counter:"true"`
}
```

Accepted forms:

```bash
app -vvv
app -v -v -v
app --verbose=3
```

Use `Verbose >= 1`, `Verbose >= 2`, and so on in logging setup.
Do not use `[]bool` unless every occurrence has meaning.

Related: [Options][] and [Parsing Rules][].

## Simple Command

Use commands when the executable has distinct actions.

```go
type AddCommand struct {
  Name string `long:"name" required:"true"`
}

func (c *AddCommand) Execute(args []string) error {
  fmt.Println("add", c.Name)
  return nil
}

type Options struct {
  Add AddCommand `command:"add" description:"Add item"`
}
```

Command-local options become valid after the command token:

```bash
app add --name item
```

Related: [Commands][].

## Command with Positional Arguments

Use positionals when values are naturally identified by order.

```go
type CopyCommand struct {
  Args struct {
    Source string `positional-arg-name:"source" required:"true"`
    Target string `positional-arg-name:"target" required:"true"`
  } `positional-args:"yes"`
}
```

Example:

```bash
app copy input.txt output.txt
```

Use named options instead when the meaning is not obvious from order.

Related: [Commands][] and [Positional Arguments][].

## Git-Style Command Tree

Use nested commands for command families.

```go
type Options struct {
  Verbose int `short:"v" counter:"true"`

  Remote struct {
    Add struct {
      Args struct {
        Name string `required:"true"`
        URL  string `required:"true"`
      } `positional-args:"yes"`
    } `command:"add" description:"Add remote"`
  } `command:"remote" description:"Manage remotes"`
}
```

Global options stay at the root.
Command-specific arguments stay inside command structs.

Related: [Commands][] and [Groups][].

## Wrapper Command

Use pass-through parsing when a command forwards arguments to another program.

```go
type Options struct {
  Exec struct {
    Args struct {
      Program string   `required:"true"`
      Rest    []string
    } `positional-args:"yes"`
  } `command:"exec" pass-after-non-option:"true"`
}
```

Example:

```bash
app exec grep --color=always pattern file.txt
```

After the program name, arguments can look like options for the wrapped command.
`Rest` is optional because a trailing slice without `required` may be empty.
Use `--` when users need an explicit boundary.

Related: [Commands][], [Positional Arguments][], and [Parsing Rules][].

## Mutually Exclusive Formats

Use `xor` when options cannot be used together.

```go
type Options struct {
  JSON bool `long:"json" xor:"format"`
  YAML bool `long:"yaml" xor:"format"`
  Text bool `long:"text" xor:"format"`
}
```

This allows zero or one format option.
Add `required:"true"` to one member when exactly one format is required.

Related: [Options][] and [Struct Tags][].

## Paired Credentials

Use `and` when options only make sense together.

```go
type Options struct {
  User string `long:"user" and:"auth"`
  Pass string `long:"pass" and:"auth"`
}
```

If either option is set, both must be set.
If neither is set, the group is allowed.

Add `required:"true"` when both must always be present.

Related: [Options][] and [Struct Tags][].

## Config-First CLI

Use this pattern only when another source fills
the config struct before command-line parsing.
For example, a project file may define defaults for the whole repository,
while CLI flags override values for one invocation.

```go
cfg := Config{}
// Load config file into cfg before parsing.

parser := flags.NewParser(&cfg, flags.Default|flags.ConfiguredValues)
_, err := parser.Parse()
```

If the struct starts empty, do not use this pattern.
Use `flags.Default` without `flags.ConfiguredValues` instead.

Related: [Defaults and Configuration][] and [Parser Options][].

## Environment Prefix

Use a prefix when environment keys should be namespaced by application.

```go
type Options struct {
  Port int `long:"port" env:"PORT" default:"8080"`
}

parser := flags.NewParser(&opts, flags.Default)
parser.SetEnvPrefix("APP")
```

The final key is `APP_PORT`.

Related: [Environment][] and [Groups][].

## Required Existing File

Use validators when the value must satisfy a local filesystem rule before
application logic runs.

```go
type Options struct {
  Input string `long:"input" required:"true" validate-existing-file:"true" validate-readable:"true"`
}
```

This is appropriate when the option must be a real file path.
Do not use it for fields that also accept `stdin`.

Related: [Value Validators][] and [I/O Templates][].

## Output Path That May Not Exist

Use `validate-writable` for output paths that may be created later.

```go
type Options struct {
  Output string `long:"output" validate-writable:"true"`
}
```

The validator accepts a missing file when the parent directory is writable.
Application code still owns the final file creation.

Related: [Value Validators][].

## Input to Output Filter

Use positional I/O templates for filter-style tools.

```go
type Options struct {
  IO struct {
    Input  string `io:"in" io-kind:"auto"`
    Output string `io:"out" io-kind:"auto"`
  } `positional-args:"yes"`
}
```

User behavior:

```bash
app
app input.txt
app input.txt output.txt
app - output.txt
```

Parser behavior:

* omitted input becomes `stdin`;
* omitted output becomes `stdout`;
* `-` maps to the role stream;
* paths stay as strings.

Application code still opens files or uses standard streams.

Related: [I/O Templates][] and [Positional Arguments][].

## Static Completion Values

Use `choices` when possible values are known at compile time.

```go
type Options struct {
  Format string `long:"format" choices:"json;yaml;text"`
}
```

Choices validate input, render in help, and feed shell completion.

Related: [Options][] and [Completion][].

## File and Directory Completion

Use completion hints when the value is a path.

```go
type Options struct {
  Config string `long:"config" completion:"file"`
  Root   string `long:"root" completion:"dir"`
}
```

Use `completion:"none"` when completion would be misleading.

Related: [Completion][] and [I/O Templates][].

## Custom Completion

Use `Completer` when values are dynamic or domain-specific.

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

Keep completion fast. Shells may call it on every tab press.

Related: [Completion][] and [Custom Values][].

## Localized Help

Use i18n tags when help and docs should be translated.

```go
catalog, err := flags.NewJSONCatalogDirFS(i18nFS, "i18n")
if err != nil {
  return err
}

parser := flags.NewParser(&opts, flags.Default)
parser.SetI18n(flags.I18nConfig{
  Locale:      "ru",
  UserCatalog: catalog,
})
```

Keep literal descriptions as fallback source text.
Use stable catalog keys.

Related: [Localization][].

## Retuned Version Flag

Use this only when the default `-v` conflicts with an existing public option.

```go
parser := flags.NewParser(&opts, flags.Default|flags.VersionFlag)

if versionOpt := parser.BuiltinVersionOption(); versionOpt != nil {
  _ = versionOpt.SetShortName('B')
  _ = versionOpt.SetLongName("build-info")
}
```

For new tools, prefer conventional `-v` and `--version` when available.

Related: [Version Metadata][] and [Parser Options][].

## Internal Docs Including Hidden Options

Use this for audit docs, not for normal user help.

```go
err := parser.WriteDoc(
  os.Stdout,
  flags.DocFormatMarkdown,
  flags.WithIncludeHidden(true),
  flags.WithMarkHidden(true),
)
```

Hidden options stay parseable. They are not a security boundary.

Related: [Documentation Templates][] and [Visibility and Secrets][].

[Commands]: ../cli-structure/commands.md
[Completion]: ../output/completion.md
[Custom Values]: ../cli-structure/custom-values.md
[Defaults and Configuration]: ../configuration/defaults-and-configuration.md
[Documentation Templates]: ../output/doc-templates.md
[Environment]: ../configuration/environment.md
[Getting Started]: getting-started.md
[Groups]: ../cli-structure/groups.md
[I/O Templates]: ../cli-structure/io-templates.md
[Localization]: ../output/localization.md
[Options]: ../cli-structure/options.md
[Parser Options]: ../parser-model/parser-options.md
[Parsing Rules]: ../parser-model/parsing-rules.md
[Positional Arguments]: ../cli-structure/positional-arguments.md
[Struct Tags]: ../cli-structure/struct-tags.md
[Value Validators]: ../cli-structure/validators.md
[Version Metadata]: ../output/version.md
[Visibility and Secrets]: ../configuration/visibility-and-secrets.md
