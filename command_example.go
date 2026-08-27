// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type commandExamplePartKind uint8

const (
	commandExampleArg commandExamplePartKind = iota
	commandExampleOption
	commandExampleShortOption
	commandExampleRaw
)

var _ = renderCommandExample

type commandExamplePart struct {
	kind   commandExamplePartKind
	value  string
	target any
}

// CommandExample describes one command invocation for help and generated documentation.
// Program and command names are resolved when the example is rendered;
// option targets are resolved from their bound struct fields.
type CommandExample struct {
	description string
	parts       []commandExamplePart
}

// Example creates an empty command example builder.
func Example() *CommandExample {
	return &CommandExample{}
}

// Describe sets the optional text displayed before the example.
func (e *CommandExample) Describe(text string) *CommandExample {
	if e == nil {
		return nil
	}
	e.description = text
	return e
}

// Arg appends a positional argument to the example.
func (e *CommandExample) Arg(value string) *CommandExample {
	return e.appendPart(commandExamplePart{kind: commandExampleArg, value: value})
}

// Option appends an option using its canonical long form when rendered.
// The target is resolved against the current command metadata at render time.
func (e *CommandExample) Option(target any, value ...string) *CommandExample {
	return e.appendPart(commandExamplePart{
		kind:   commandExampleOption,
		target: target,
		value:  firstExampleValue(value),
	})
}

// ShortOption appends an option using its short form when rendered.
func (e *CommandExample) ShortOption(target any, value ...string) *CommandExample {
	return e.appendPart(commandExamplePart{
		kind:   commandExampleShortOption,
		target: target,
		value:  firstExampleValue(value),
	})
}

// Raw appends a shell fragment verbatim to the example.
func (e *CommandExample) Raw(value string) *CommandExample {
	return e.appendPart(commandExamplePart{kind: commandExampleRaw, value: value})
}

func (e *CommandExample) appendPart(part commandExamplePart) *CommandExample {
	if e == nil {
		return nil
	}

	e.parts = append(e.parts, part)
	return e
}

func firstExampleValue(values []string) string {
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func cloneCommandExample(example *CommandExample) *CommandExample {
	if example == nil {
		return nil
	}

	clone := &CommandExample{description: example.description}
	clone.parts = append([]commandExamplePart(nil), example.parts...)
	return clone
}

func cloneCommandExamples(examples []*CommandExample) []*CommandExample {
	if len(examples) == 0 {
		return nil
	}

	clones := make([]*CommandExample, 0, len(examples))
	for _, example := range examples {
		clones = append(clones, cloneCommandExample(example))
	}

	return clones
}

func (c *Command) resolveExampleOption(target any) (*Option, error) {
	if target == nil {
		return nil, errors.New("example option target must not be nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return nil, errors.New("example option target must be a non-nil pointer")
	}

	var found *Option
	for _, command := range c.optionScopeCommands() {
		command.eachGroup(func(group *Group) {
			for _, option := range group.options {
				if found != nil || !option.value.CanAddr() {
					continue
				}
				field := option.value.Addr()
				if field.Type() == targetValue.Type() && field.Pointer() == targetValue.Pointer() {
					found = option
				}
			}
		})
	}

	if found == nil {
		return nil, fmt.Errorf("example option target %T is not registered in command scope", target)
	}

	return found, nil
}

func (c *Command) validateCommandExample(example *CommandExample) error {
	if example == nil {
		return nil
	}

	for _, part := range example.parts {
		if part.kind != commandExampleOption && part.kind != commandExampleShortOption {
			continue
		}

		if _, err := c.resolveExampleOption(part.target); err != nil {
			return err
		}
	}

	return nil
}

// ExampleShell identifies the shell syntax used to render command examples.
type ExampleShell string

const (
	// ExampleShellBash renders examples for Bash-compatible shells.
	ExampleShellBash ExampleShell = "bash"
	// ExampleShellPwsh renders examples for PowerShell.
	ExampleShellPwsh ExampleShell = "pwsh"
)

type renderedCommandExample struct {
	Description string
	Command     string
}

func renderCommandExample(command *Command, programName string, shell ExampleShell, example *CommandExample) (renderedCommandExample, error) {
	if command == nil || example == nil {
		return renderedCommandExample{}, errors.New("command and example must not be nil")
	}
	if shell != ExampleShellBash && shell != ExampleShellPwsh {
		return renderedCommandExample{}, fmt.Errorf("unsupported example shell %q", shell)
	}

	parts := []string{quoteExampleToken(programName, shell)}
	for current := command; current != nil; {
		parent, ok := current.parent.(*Command)
		if !ok {
			break
		}
		parts = append(parts, current.Name)
		current = parent
	}

	// Command names were collected leaf-first.
	for left, right := 1, len(parts)-1; left < right; left, right = left+1, right-1 {
		parts[left], parts[right] = parts[right], parts[left]
	}

	for _, part := range example.parts {
		switch part.kind {
		case commandExampleArg:
			parts = append(parts, quoteExampleToken(part.value, shell))

		case commandExampleOption, commandExampleShortOption:
			option, err := command.resolveExampleOption(part.target)
			if err != nil {
				return renderedCommandExample{}, err
			}

			name := ""
			switch {
			case part.kind == commandExampleOption && option.LongName != "":
				name = "--" + option.LongNameWithNamespace()
			case option.ShortName != 0:
				name = "-" + string(option.ShortName)
			default:
				return renderedCommandExample{}, fmt.Errorf("option target %T has no usable name", part.target)
			}

			parts = append(parts, name)
			if part.value != "" {
				parts = append(parts, quoteExampleToken(part.value, shell))
			}

		case commandExampleRaw:
			if part.value != "" {
				parts = append(parts, part.value)
			}
		}
	}

	return renderedCommandExample{Description: example.description, Command: strings.Join(parts, " ")}, nil
}

func quoteExampleToken(value string, shell ExampleShell) string {
	if value != "" && isSafeExampleToken(value) {
		return value
	}
	if shell == ExampleShellPwsh {
		return pwshSingleQuote(value)
	}

	return bashSingleQuote(value)
}

func isSafeExampleToken(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_./:-+=,@%", r) {
			continue
		}
		return false
	}
	return true
}
