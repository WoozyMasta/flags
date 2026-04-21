// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

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

Define options as struct fields and annotate them with tags:

	type Options struct {
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose output"`
	}

If `-v` or `--verbose` appears, `true` is appended to `Verbose`.
For example, `-vvv` yields `[]bool{true, true, true}`.

Slices accept repeated values naturally:

	type Options struct {
		Include []string `short:"I" description:"Include directory"`
	}

Maps are parsed as `key:value`:

	type Options struct {
		AuthorInfo map[string]string `short:"a"`
	}

Example input:

	-a name:Jesse -a "surname:van den Kieboom"

For custom value conversion, implement [Marshaler] and [Unmarshaler].

# Struct Tags

An option field must define at least one of `short` or `long`.

General option tags:

  - `short`: single-character short option name
  - `long`: long option name
  - `required`: marks option as required; parser returns [ErrRequired] when missing
  - `description`: short help text
  - `long-description`: extended text (currently used in generated man pages)
  - `no-flag`: ignore field as command-line option
  - `hidden`: hide from help and man pages

Value and default tags:

  - `optional`: marks option argument as optional; must be passed as `--opt=value`
  - `optional-value`: value used when optional option appears without explicit argument
  - `default`: default value (repeat for slice/map entries)
  - `default-mask`: display replacement for default in help; `-` hides default entirely
  - `env`: environment variable that overrides default value
  - `env-delim`: split `env` value by delimiter for slice/map fields
  - `value-name`: placeholder name shown in help
  - `choice`: allowed value constraint (repeatable), for example
    `long:"animal" choice:"cat" choice:"dog"`
  - `base`: radix for integer parsing, default `10`
  - `key-value-delimiter`: delimiter used when parsing map values, default `:`
  - `unquote`: when set to `false`, disables automatic unquoting of argument values

INI tags:

  - `ini-name`: explicit INI key name
  - `no-ini`: ignore field for INI parsing/writing

Group and command tags:

  - `group`: treat struct field as a named option group
  - `namespace`: prefix long option names inside group hierarchy
  - `env-namespace`: prefix environment variable names inside group hierarchy
  - `command`: treat struct field as a command
  - `subcommands-optional`: make subcommands under this command optional
  - `pass-after-non-option`: for this command, stop option parsing after the
    first non-option argument
  - `alias`: extra command name (repeatable)
  - `positional-args`: map trailing positional arguments into struct fields
  - `positional-arg-name`: placeholder label for positional help

For `positional-args`, arguments are optional by default.
Use `required` either on the positional struct field or on individual fields.
For a trailing slice field, `required:"N"` means at least `N` values.

# Option Groups

Groups organize related options in help output and in parser structure.
You can define groups in three ways:

 1. Create a parser with [NewNamedParser].
 2. Add groups programmatically with [Parser.AddGroup].
 3. Add nested struct fields tagged with `group:"name"`.

# Commands

Commands split CLI behavior into explicit actions (similar to `git add`,
`git commit`, and so on). You can define commands in two ways:

 1. Add commands programmatically with [Parser.AddCommand].
 2. Add struct fields tagged with `command:"name"`.

If the selected command implements [Commander], its `Execute` method runs
after parsing with the remaining arguments.

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
*/
package flags
