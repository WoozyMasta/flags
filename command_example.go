// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"errors"
	"fmt"
	"reflect"
)

type commandExamplePartKind uint8

const (
	commandExampleArg commandExamplePartKind = iota
	commandExampleOption
	commandExampleShortOption
	commandExampleRaw
)

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
