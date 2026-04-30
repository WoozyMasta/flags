// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

/*
Package flags provides a reflection-based command-line parser.

It is similar to Go's standard `flag` package, but adds richer modeling:
long and short options, nested groups, subcommands, map and slice values,
environment-variable defaults, INI integration, and shell completion.

# Features

Core features:

  - Short options (for example, `-v`)
  - Long options (for example, `--verbose`)
  - Options with required, optional, or no arguments
  - Optional argument values (`--flag=value`) with fallback `optional-value`
  - Defaults from tags and environment variables (including slices and maps)
  - Grouped options and nested namespaces
  - Subcommands and command aliases
  - Help and man-page generation
  - `--` passthrough support (`PassDoubleDash`)
  - Optional unknown-flag ignoring
  - Combined short flags (for example, `-aux`)
  - Option argument forms such as `-I/usr/include`, `-I=/usr/include`, `-I /usr/include`
  - Repeated options (store in slices or use counter semantics)
  - Primitive scalar types, maps, and callback/function options

Windows-specific behavior:

  - Slash-prefixed options (`/v`, `/verbose`)
  - `:` delimiter for option arguments in Windows mode
  - Windows-style help rendering
  - `forceposix` build tag to disable Windows-style parsing

# Quick Start

Minimal parse flow:

	type Options struct {
		Verbose bool   `short:"v" long:"verbose" description:"Show verbose output"`
		Region  string `long:"region" default:"eu-west-1" description:"Cloud region"`
	}

	var opts Options
	parser := NewParser(&opts, Default|HelpCommands)

	_, err := parser.Parse()
	if err != nil {
		var ferr *Error
		if errors.As(err, &ferr) && ferr.Type == ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

`HelpCommands` opt-in enables built-in commands:
`help`, `version`, `completion`, `docs`, `config`.

For custom value conversion, implement [Marshaler] and [Unmarshaler].

# Struct Tags

An option field must define at least one of `short` or `long`.

General option tags:

  - `short`: single-character short option name
  - `long`: long option name
  - `required`: marks option as required; parser returns [ErrRequired] when missing
  - `xor`: delimiter-separated exclusive option relation groups
  - `and`: delimiter-separated all-or-none option relation groups
  - `counter`: integer counter mode; each occurrence increments by 1
  - `description`: short help text
  - `long-description`: extended text (currently used in generated man pages)
  - `no-flag`: ignore field as command-line option
  - `hidden`: hide from help and man pages

Value and default tags:

  - `optional`: marks option argument as optional; must be passed as `--opt=value`
  - `optional-value`: value used when optional option appears without explicit argument
  - `order`: display/completion priority in group block sorting
  - `default`: default value (repeat for slice/map entries)
  - `defaults`: delimiter-separated default list (non-repeatable)
  - `default-mask`: display replacement for default in help; `-` hides default entirely
  - `env`: environment variable that overrides default value
  - `auto-env`: derive environment variable from `long` name when enabled
  - `env-delim`: split `env` value by delimiter for slice/map fields
  - `value-name`: placeholder name shown in help
  - `choice`: allowed value constraint (repeatable), for example
    `long:"animal" choices:"cat;dog"`
  - `choices`: delimiter-separated allowed values (non-repeatable)
  - `completion`: completion hint (`file`, `dir`, `none`) used when no
    custom completer or choices are defined
  - `base`: radix for integer parsing, default `10`
  - `key-value-delimiter`: delimiter used when parsing map values, default `:`
  - `unquote`: when set to `false`, disables automatic unquoting of argument values
  - `short-alias`: extra short option name (repeatable)
  - `short-aliases`: delimiter-separated short aliases (non-repeatable)
  - `long-alias`: extra long option name (repeatable)
  - `long-aliases`: delimiter-separated long aliases (non-repeatable)

INI tags:

  - `ini-name`: explicit INI key name
  - `no-ini`: ignore field for INI parsing/writing

Group and command tags:

  - `group`: treat struct field as a named option group
  - `namespace`: prefix long option names inside group hierarchy
  - `env-namespace`: prefix environment variable names inside group hierarchy
  - `command`: treat struct field as a command
  - `command-group`: display group for command help/docs
  - `subcommands-optional`: make subcommands under this command optional
  - `pass-after-non-option`: for this command, stop option parsing after the
    first non-option argument
  - `alias`: extra command name (repeatable)
  - `aliases`: delimiter-separated command aliases (non-repeatable)
  - `positional-args`: map trailing positional arguments into struct fields
  - `io`: string I/O role (`in`, `out`) for options/positionals
  - `io-kind`: string I/O kind (`auto`, `stream`, `file`, `string`)
  - `io-stream`: stream token (`stdin`, `stdout`, `stderr`)
  - `io-open`: output file mode metadata (`truncate`, `append`)
  - `positional-arg-name`: placeholder label for positional help
  - `completion`: positional completion hint (`file`, `dir`, `none`)

For `positional-args`, arguments are optional by default.
Use `required` either on the positional struct field or on individual fields.
For a trailing slice field, `required:"N"` means at least `N` values.
For positional args, `io:"in"`/`io:"out"` with `io-kind:"auto"` or
`io-kind:"stream"` defaults omitted values to `stdin`/`stdout`.
For options, no implicit fallback is applied when the flag is omitted.
When `completion` is not set, `io-kind:"file"` and `io-kind:"auto"` imply
file completion hints.

# Error Handling

With [Default] parser options, [PrintErrors] is enabled and parse errors are
printed by the parser. A typical pattern is:

	_, err := parser.Parse()
	if err != nil {
		var ferr *Error
		if errors.As(err, &ferr) && ferr.Type == ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

If you need full control over output formatting/routing, disable [PrintErrors]
and print returned errors yourself.

# Option Groups

Groups organize related options in help output and in parser structure.
You can define groups in three ways:

 1. Create a parser with [NewNamedParser].
 2. Add groups programmatically with [Parser.AddGroup].
 3. Add nested struct fields tagged with `group:"name"`.

For post-scan programmatic adjustments, implement [Configurer] on your
options/group/command data type and mutate option metadata via parser APIs.
Common runtime setters are available on [Option] (for example names, aliases,
defaults, env bindings, choices, required/hidden flags, and order).
Command/group/arg metadata can also be tuned at runtime via setters on
[Command], [Group], and [Arg].

# Commands

Commands split CLI behavior into explicit actions (similar to `git add`,
`git commit`, and so on). You can define commands in two ways:

 1. Add commands programmatically with [Parser.AddCommand].
 2. Add struct fields tagged with `command:"name"`.

If the selected command implements [Commander], its `Execute` method runs
after parsing with the remaining arguments.

By default only the selected leaf command executes. Enable [CommandChain] to
execute every active command implementing [Commander] from parent to leaf. If a
command returns an error, the chain stops and [Parser.Parse] returns that error.

Built-in command entry points are opt-in through parser option bits:
[HelpCommand], [VersionCommand], [CompletionCommand], [DocsCommand],
and [ConfigCommand]. [HelpCommands] enables the full set. Built-in commands
are grouped as `Help Commands` in help/docs by default; use
[Parser.SetBuiltinCommandGroup] to rename that display group or set it to an
empty string. Built-in `completion` auto-detects shell format when `--shell`
is omitted (`zsh`/`pwsh`) and falls back to `bash`.

Command-local options become valid after the command token is parsed.
With a global `-v` option and an `add` command, these are equivalent:

	./app -v add
	./app add -v

If `-v` exists only on `add`, then `./app -v add` fails, while `./app add -v`
works.

# Completion

Completion mode is enabled by setting `GO_FLAGS_COMPLETION`:

	GO_FLAGS_COMPLETION=1 ./completion-example arg1 arg2 arg3

The last argument (`arg3`) is treated as the value to complete.
When `GO_FLAGS_COMPLETION=verbose`, completion descriptions are emitted
when multiple candidates exist.

Because completion executes your program, avoid side effects during startup
(for example, in `init` routines).

Bash integration example:

	_completion_example() {
		args=("${COMP_WORDS[@]:1:$COMP_CWORD}")
		local IFS=$'\n'
		COMPREPLY=($(GO_FLAGS_COMPLETION=1 ${COMP_WORDS[0]} "${args[@]}"))
		return 0
	}
	complete -F _completion_example completion-example

Completion requires [PassDoubleDash], and the parser enforces it automatically
when `GO_FLAGS_COMPLETION` is set.

For custom value completion, implement [Completer].
[Filename] is a built-in example. Slices and arrays are also completable when
their element type implements [Completer].
When no custom completer is available, completion source priority is:
`choices` -> `completion` hint (`file`/`dir`/`none`) -> built-in bool values.
*/
package flags
