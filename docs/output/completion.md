# Completion

Shell completion is generated from the parser model.
It knows about commands, command aliases, option names, option aliases,
choices, completion hints, custom completers, and hidden metadata.

Completion runs by executing your program.
Startup code should avoid side effects before completion has a chance to exit.

## Supported Shells

Built-in script generation supports:

* `CompletionShellBash`, script value `bash`;
* `CompletionShellZsh`, script value `zsh`;
* `CompletionShellPwsh`, script value `pwsh`.

Generate a script for a known shell:

```go
err := parser.WriteNamedCompletion(
  os.Stdout,
  flags.CompletionShellBash,
  "myapp",
)
```

Use parser name as the command name:

```go
err := parser.WriteCompletion(os.Stdout, flags.CompletionShellBash)
```

Detect the shell format:

```go
err := parser.WriteAutoCompletion(os.Stdout)
```

Unknown shell detection falls back to bash script format.
Passing an unsupported shell value to `WriteNamedCompletion` returns an error.
Passing an empty command name to `WriteNamedCompletion` returns
`ErrEmptyCommandName`.

## Raw Completion Mode

Completion scripts call the application with `GO_FLAGS_COMPLETION=1`.
The parser then treats the current command-line tokens
as completion input and prints candidates.

Verbose mode can include descriptions:

```bash
GO_FLAGS_COMPLETION=verbose ./myapp --format j
```

The default printer emits descriptions only in verbose mode
and only when there is more than one candidate.
Without verbose mode, only candidate items are printed.

Do not perform irreversible startup work before parsing
when completion mode is active.
The shell may call completion often.

## Completion Sources

Value completion priority is:

1. custom `Completer` implementation;
1. `choice` or `choices` tags;
1. `completion` hint;
1. built-in bool values when `AllowBoolValues` is enabled.

This order lets type-specific completion override generic metadata.
Slices use the element type when checking for `Completer`.

`Completion` contains:

* `Item`, the value inserted by the shell.
* `Description`, optional text shown by verbose completion output.

## Choices

Static choices feed completion automatically.

```go
type Options struct {
  Format string `long:"format" choices:"json;yaml;text"`
}
```

Use choices for small fixed sets. For dynamic sets, implement `Completer`.

## Completion Hints

* `completion:"file"` completes files.
* `completion:"dir"` completes directories.
* `completion:"none"` disables value completion.
* absent or empty `completion` uses automatic behavior.

```go
type Options struct {
  Config string `long:"config" completion:"file"`
  Root   string `long:"root" completion:"dir"`
}
```

I/O templates with `io-kind:"file"` or `io-kind:"auto"`
imply file completion when no explicit completion hint is set.

## Custom Completers

Implement `Completer` on the target type.

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

Candidates can include descriptions:

```go
flags.Completion{Item: "json", Description: "Machine-readable JSON"}
```

Keep completers fast.
Avoid network calls unless they are cached and clearly expected by users.

## Filename Completer

`flags.Filename` is a built-in string alias with file completion.

```go
type Options struct {
  Input flags.Filename `long:"input"`
}
```

Use `Filename` when the type itself should express file completion.
Use `completion:"file"` when the field should remain a plain string.

If a single completed item is a directory,
`Filename` appends `/` to make the next path segment convenient.

## Option Name Completion

Completion includes short and long option names valid
in the active command scope.
Hidden options are omitted. Aliases are included where applicable.

For inline values such as `--opt=value`,
completion avoids inserting an unwanted space
when the shell supports that behavior.

## Command Completion

Completion includes commands valid in the current command scope.
Hidden commands are omitted.
Command aliases can be completed.

Command-local options become available after
the command token is selected, matching normal parsing rules.

## Completion Handler

`Parser.CompletionHandler` can override how completion candidates are handled.
By default, candidates are printed and the application exits.

Use a custom handler in tests, embedded tools,
or applications that need to collect completions without exiting.
When a custom handler is set,
the parser returns `nil, nil` after the handler runs.

See [Handlers and Integration Points][] for the full handler lifecycle.

## Built-in Completion Command

`CompletionCommand` adds a built-in `completion` command.
`HelpCommands` enables it together with the other helper commands.

The built-in command can write shell scripts
and auto-detect shell format when the shell option is omitted.
The command writes to stdout when the output argument is omitted.

## Completion Rules

Completion should reflect real parse behavior.
If help says a value accepts `json` or `yaml`,
completion should not suggest unrelated values.

Avoid slow startup paths before parsing.
Completion quality drops quickly when every tab press takes noticeable time.

[Handlers and Integration Points]: ../parser-model/handlers.md
