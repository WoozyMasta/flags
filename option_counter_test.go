package flags

import "testing"

func TestOptionCounterIncrementsForRepeatedShortFlags(t *testing.T) {
	if defaultShortOptDelimiter != '-' {
		t.Skip("short flag clustering is only supported with '-' short delimiter")
	}

	var opts struct {
		Verbose int `short:"v" long:"verbose" counter:"true"`
	}

	assertParseSuccess(
		t,
		&opts,
		string(defaultShortOptDelimiter)+"vvv",
	)

	if opts.Verbose != 3 {
		t.Fatalf("expected verbose counter to be 3, got %d", opts.Verbose)
	}
}

func TestOptionCounterIncrementsForSeparateShortFlags(t *testing.T) {
	var opts struct {
		Verbose int `short:"v" long:"verbose" counter:"true"`
	}

	assertParseSuccess(
		t,
		&opts,
		string(defaultShortOptDelimiter)+"v",
		string(defaultShortOptDelimiter)+"v",
		string(defaultShortOptDelimiter)+"v",
	)

	if opts.Verbose != 3 {
		t.Fatalf("expected verbose counter to be 3, got %d", opts.Verbose)
	}
}

func TestOptionCounterAcceptsExplicitValues(t *testing.T) {
	var opts struct {
		Verbose int `short:"v" long:"verbose" counter:"true"`
	}

	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"verbose", "3",
		string(defaultShortOptDelimiter)+"v", "2",
	)

	if opts.Verbose != 5 {
		t.Fatalf("expected verbose counter to be 5, got %d", opts.Verbose)
	}
}

func TestOptionCounterAcceptsConcatenatedShortValue(t *testing.T) {
	if defaultShortOptDelimiter != '-' {
		t.Skip("short concatenated value form is only supported with '-' short delimiter")
	}

	var opts struct {
		Verbose int `short:"v" long:"verbose" counter:"true"`
	}

	assertParseSuccess(
		t,
		&opts,
		string(defaultShortOptDelimiter)+"v3",
	)

	if opts.Verbose != 3 {
		t.Fatalf("expected verbose counter to be 3, got %d", opts.Verbose)
	}
}

func TestOptionCounterAcceptsEqualsValue(t *testing.T) {
	if defaultShortOptDelimiter != '-' || defaultLongOptDelimiter != "--" {
		t.Skip("equals-value form is only supported with GNU-style delimiters")
	}

	var opts struct {
		Verbose int `short:"v" long:"verbose" counter:"true"`
	}

	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"verbose=3",
		string(defaultShortOptDelimiter)+"v=2",
	)

	if opts.Verbose != 5 {
		t.Fatalf("expected verbose counter to be 5, got %d", opts.Verbose)
	}
}

func TestOptionCounterRejectsNegativeValues(t *testing.T) {
	var opts struct {
		Verbose int `short:"v" long:"verbose" counter:"true"`
	}

	assertParseFail(
		t,
		ErrMarshal,
		"invalid argument for flag `"+string(defaultShortOptDelimiter)+"v, "+defaultLongOptDelimiter+"verbose' (expected int): counter increment must be non-negative",
		&opts,
		defaultLongOptDelimiter+"verbose", "-1",
	)
}

func TestOptionCounterRejectsNonIntegerTypes(t *testing.T) {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" counter:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{string(defaultShortOptDelimiter) + "v"})
	assertError(
		t,
		err,
		ErrInvalidTag,
		"counter tag `counter' requires integer option type on field `Verbose'",
	)
}
