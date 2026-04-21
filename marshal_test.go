package flags

import (
	"fmt"
	"testing"
)

type marshalled string

func (m *marshalled) UnmarshalFlag(value string) error {
	if value == "yes" {
		*m = "true"
	} else if value == "no" {
		*m = "false"
	} else {
		return fmt.Errorf("`%s' is not a valid value, please specify `yes' or `no'", value)
	}

	return nil
}

func (m marshalled) MarshalFlag() (string, error) {
	if m == "true" {
		return "yes", nil
	}

	return "no", nil
}

type marshalledError bool

func (m marshalledError) MarshalFlag() (string, error) {
	return "", newErrorf(ErrMarshal, "Failed to marshal")
}

type textMarshalled string

func (m *textMarshalled) UnmarshalText(value []byte) error {
	switch string(value) {
	case "yes":
		*m = "true"
	case "no":
		*m = "false"
	default:
		return fmt.Errorf("`%s' is not a valid value, please specify `yes' or `no'", string(value))
	}

	return nil
}

func (m textMarshalled) MarshalText() ([]byte, error) {
	if m == "true" {
		return []byte("yes"), nil
	}

	return []byte("no"), nil
}

type dualMarshalled string

func (m *dualMarshalled) UnmarshalFlag(value string) error {
	*m = dualMarshalled("flag:" + value)
	return nil
}

func (m dualMarshalled) MarshalFlag() (string, error) {
	return "from-flag-marshaler", nil
}

func (m *dualMarshalled) UnmarshalText(value []byte) error {
	*m = dualMarshalled("text:" + string(value))
	return nil
}

func (m dualMarshalled) MarshalText() ([]byte, error) {
	return []byte("from-text-marshaler"), nil
}

func TestUnmarshal(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=yes")

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestUnmarshalDefault(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v" default:"yes"`
	}{}

	ret := assertParseSuccess(t, &opts)

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestUnmarshalOptional(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v" optional:"yes" optional-value:"yes"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v")

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestUnmarshalError(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v"`
	}{}

	assertParseFail(t, ErrMarshal, fmt.Sprintf("invalid argument for flag `%cv' (expected flags.marshalled): `invalid' is not a valid value, please specify `yes' or `no'", defaultShortOptDelimiter), &opts, "-vinvalid")
}

func TestUnmarshalPositionalError(t *testing.T) {
	var opts = struct {
		Args struct {
			Value marshalled
		} `positional-args:"yes"`
	}{}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"invalid"})

	msg := "`invalid' is not a valid value, please specify `yes' or `no'"

	if err == nil {
		assertFatalf(t, "Expected error: %s", msg)
		return
	}

	if err.Error() != msg {
		assertErrorf(t, "Expected error message %#v, but got %#v", msg, err.Error())
	}
}

func TestMarshalError(t *testing.T) {
	var opts = struct {
		Value marshalledError `short:"v"`
	}{}

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]

	_, err := convertToString(o.value, o.tag)

	assertError(t, err, ErrMarshal, "Failed to marshal")
}

func TestTextUnmarshalRoundTrip(t *testing.T) {
	var opts = struct {
		Value textMarshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=yes")
	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]
	s, err := convertToString(o.value, o.tag)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, s, "yes")
}

func TestMarshalInterfacePriority(t *testing.T) {
	var opts = struct {
		Value dualMarshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=abc")
	assertStringArray(t, ret, []string{})

	assertString(t, string(opts.Value), "flag:abc")

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]
	s, err := convertToString(o.value, o.tag)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, s, "from-flag-marshaler")
}
