package flags

import (
	"errors"
	"testing"
)

func TestErrorTypeStringAndError(t *testing.T) {
	tests := []struct {
		tp   ErrorType
		want string
	}{
		{ErrUnknown, "unknown"},
		{ErrExpectedArgument, "expected argument"},
		{ErrUnknownFlag, "unknown flag"},
		{ErrUnknownGroup, "unknown group"},
		{ErrMarshal, "marshal"},
		{ErrHelp, "help"},
		{ErrVersion, "version"},
		{ErrNoArgumentForBool, "no argument for bool"},
		{ErrRequired, "required"},
		{ErrShortNameTooLong, "short name too long"},
		{ErrDuplicatedFlag, "duplicated flag"},
		{ErrTag, "tag"},
		{ErrCommandRequired, "command required"},
		{ErrUnknownCommand, "unknown command"},
		{ErrInvalidChoice, "invalid choice"},
		{ErrInvalidTag, "invalid tag"},
		{ErrOptionConflict, "option conflict"},
		{ErrOptionRequirement, "option requirement"},
		{ErrorType(999), "unrecognized error type"},
	}

	for _, test := range tests {
		if got := test.tp.String(); got != test.want {
			t.Fatalf("expected %q, got %q", test.want, got)
		}

		if got := test.tp.Error(); got != test.want {
			t.Fatalf("expected Error()=%q, got %q", test.want, got)
		}
	}
}

func TestWrapErrorAndIniErrorFormatting(t *testing.T) {
	source := newError(ErrUnknownFlag, "boom")
	if got := wrapError(source); got != source {
		t.Fatalf("expected wrapped error pointer to be preserved")
	}

	raw := errors.New("raw")
	wrapped := wrapError(raw)

	if wrapped.Type != ErrUnknown {
		t.Fatalf("expected ErrUnknown, got %v", wrapped.Type)
	}

	if wrapped.Message != "raw" {
		t.Fatalf("expected message %q, got %q", "raw", wrapped.Message)
	}

	iniErr := (&IniError{File: "a.ini", LineNumber: 7, Message: "bad value"}).Error()
	if iniErr != "a.ini:7: bad value" {
		t.Fatalf("unexpected ini error formatting: %q", iniErr)
	}
}
