// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"errors"
	"fmt"
)

var (
	// ErrNotPointerToStruct indicates that a provided data container is not
	// a pointer to a struct. Only pointers to structs are valid data containers
	// for options.
	ErrNotPointerToStruct = errors.New("provided data is not a pointer to struct")

	// ErrEmptyCommandName indicates that completion generation was requested
	// without a command name.
	ErrEmptyCommandName = errors.New("command name must not be empty")

	// ErrNilWriter indicates that a nil output writer was passed.
	ErrNilWriter = errors.New("nil writer")

	// ErrNegativeMaxLongNameLength indicates an invalid negative limit.
	ErrNegativeMaxLongNameLength = errors.New("max long name length cannot be negative")

	// ErrNegativeCommandOptionIndent indicates an invalid negative indent.
	ErrNegativeCommandOptionIndent = errors.New("command option indent cannot be negative")

	// ErrNegativeHelpWidth indicates an invalid negative help width.
	ErrNegativeHelpWidth = errors.New("help width cannot be negative")

	// ErrNULTagListDelimiter indicates that zero rune was used as delimiter.
	ErrNULTagListDelimiter = errors.New("tag list delimiter cannot be NUL")

	// ErrSetConsoleTitleFailed indicates an unknown SetConsoleTitleW failure.
	ErrSetConsoleTitleFailed = errors.New("SetConsoleTitleW failed")

	// ErrCounterNonNegative indicates that counter increment cannot be negative.
	ErrCounterNonNegative = errors.New("counter increment must be non-negative")

	// ErrCounterOverflow indicates counter arithmetic overflow.
	ErrCounterOverflow = errors.New("counter overflow")

	// ErrCounterIncrementTooLarge indicates counter increment is too large.
	ErrCounterIncrementTooLarge = errors.New("counter increment is too large")

	// ErrCounterInvalidType indicates invalid type for counter option.
	ErrCounterInvalidType = errors.New("counter option must use integer type")
)

// ErrorType represents the type of error.
type ErrorType uint

const (
	// ErrUnknown indicates a generic error.
	ErrUnknown ErrorType = iota

	// ErrExpectedArgument indicates that an argument was expected.
	ErrExpectedArgument

	// ErrUnknownFlag indicates an unknown flag.
	ErrUnknownFlag

	// ErrUnknownGroup indicates an unknown group.
	ErrUnknownGroup

	// ErrMarshal indicates a marshalling error while converting values.
	ErrMarshal

	// ErrHelp indicates that the built-in help was shown (the error
	// contains the help message).
	ErrHelp

	// ErrVersion indicates that the built-in version output was shown.
	ErrVersion

	// ErrNoArgumentForBool indicates that an argument was given for a
	// boolean flag (which don't not take any arguments).
	ErrNoArgumentForBool

	// ErrRequired indicates that a required flag was not provided.
	ErrRequired

	// ErrShortNameTooLong indicates that a short flag name was specified,
	// longer than one character.
	ErrShortNameTooLong

	// ErrDuplicatedFlag indicates that a short or long flag has been
	// defined more than once
	ErrDuplicatedFlag

	// ErrTag indicates an error while parsing flag tags.
	ErrTag

	// ErrCommandRequired indicates that a command was required but not
	// specified
	ErrCommandRequired

	// ErrUnknownCommand indicates that an unknown command was specified.
	ErrUnknownCommand

	// ErrInvalidChoice indicates an invalid option value which only allows
	// a certain number of choices.
	ErrInvalidChoice

	// ErrInvalidTag indicates an invalid tag or invalid use of an existing tag
	ErrInvalidTag

	// ErrOptionConflict indicates that mutually exclusive options were used together.
	ErrOptionConflict

	// ErrOptionRequirement indicates that an option relation requirement failed.
	ErrOptionRequirement

	// ErrValidation indicates that parsed value validation failed.
	ErrValidation
)

func (e ErrorType) String() string {
	switch e {
	case ErrUnknown:
		return "unknown"
	case ErrExpectedArgument:
		return "expected argument"
	case ErrUnknownFlag:
		return "unknown flag"
	case ErrUnknownGroup:
		return "unknown group"
	case ErrMarshal:
		return "marshal"
	case ErrHelp:
		return "help"
	case ErrVersion:
		return "version"
	case ErrNoArgumentForBool:
		return "no argument for bool"
	case ErrRequired:
		return "required"
	case ErrShortNameTooLong:
		return "short name too long"
	case ErrDuplicatedFlag:
		return "duplicated flag"
	case ErrTag:
		return "tag"
	case ErrCommandRequired:
		return "command required"
	case ErrUnknownCommand:
		return "unknown command"
	case ErrInvalidChoice:
		return "invalid choice"
	case ErrInvalidTag:
		return "invalid tag"
	case ErrOptionConflict:
		return "option conflict"
	case ErrOptionRequirement:
		return "option requirement"
	case ErrValidation:
		return "validation"
	}

	return "unrecognized error type"
}

func (e ErrorType) Error() string {
	return e.String()
}

// IsWarning reports whether the error type should be treated as warning-level.
func (e ErrorType) IsWarning() bool {
	switch e {
	case ErrRequired, ErrCommandRequired:
		return true
	default:
		return false
	}
}

// Error represents a parser error. The error returned from Parse is of this
// type. The error contains both a Type and Message.
type Error struct {
	// The error message
	Message string

	// The type of error
	Type ErrorType
}

// Error returns the error's message
func (e *Error) Error() string {
	return e.Message
}

func newError(tp ErrorType, message string) *Error {
	return &Error{
		Type:    tp,
		Message: message,
	}
}

func newErrorf(tp ErrorType, format string, args ...any) *Error {
	return newError(tp, fmt.Sprintf(format, args...))
}

func wrapError(err error) *Error {
	ret, ok := err.(*Error)

	if !ok {
		return newError(ErrUnknown, err.Error())
	}

	return ret
}
