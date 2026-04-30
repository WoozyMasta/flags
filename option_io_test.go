package flags

import (
	"strings"
	"testing"
)

func TestOptionIOInfersFileCompletionWhenUnset(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-kind:"auto"`
	}

	parser := NewParser(&opts, None)
	opt := parser.FindOptionByLongName("input")
	if opt == nil {
		t.Fatal("expected input option to be found")
	}
	if opt.completionHint != completionHintFile {
		t.Fatalf("expected completion hint file, got %v", opt.completionHint)
	}
}

func TestOptionIOKeepsExplicitCompletion(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-kind:"auto" completion:"none"`
	}

	parser := NewParser(&opts, None)
	opt := parser.FindOptionByLongName("input")
	if opt == nil {
		t.Fatal("expected input option to be found")
	}
	if opt.completionHint != completionHintNone {
		t.Fatalf("expected completion hint none, got %v", opt.completionHint)
	}
}

func TestOptionIOStreamInAcceptsOnlyStdinOrDash(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-kind:"stream"`
	}

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"input", "stdin")
	assertString(t, opts.Input, "stdin")

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"input", "-")
	assertString(t, opts.Input, "stdin")

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{defaultLongOptDelimiter + "input", "stderr"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if flagsErr.Type != ErrMarshal {
		t.Fatalf("expected ErrMarshal, got %v", flagsErr.Type)
	}
	if !strings.Contains(flagsErr.Message, "io `in` accepts only `stdin` or `-`") {
		t.Fatalf("unexpected message: %s", flagsErr.Message)
	}
}

func TestOptionIOStreamOutAcceptsStdoutStderrOrDash(t *testing.T) {
	var opts struct {
		Output string `long:"output" io:"out" io-kind:"stream"`
	}

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"output", "stdout")
	assertString(t, opts.Output, "stdout")

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"output", "stderr")
	assertString(t, opts.Output, "stderr")

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"output", "-")
	assertString(t, opts.Output, "stdout")

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{defaultLongOptDelimiter + "output", "stdin"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if flagsErr.Type != ErrMarshal {
		t.Fatalf("expected ErrMarshal, got %v", flagsErr.Type)
	}
	if !strings.Contains(flagsErr.Message, "io `out` accepts only `stdout`, `stderr`, or `-`") {
		t.Fatalf("unexpected message: %s", flagsErr.Message)
	}
}

func TestOptionIOStringKindKeepsRawValue(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-kind:"string"`
	}

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"input", "stdin")
	assertString(t, opts.Input, "stdin")
}

func TestOptionIOStreamKindDoesNotInferFileCompletion(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-kind:"stream"`
	}

	parser := NewParser(&opts, None)
	opt := parser.FindOptionByLongName("input")
	if opt == nil {
		t.Fatal("expected input option to be found")
	}
	if opt.completionHint != completionHintAuto {
		t.Fatalf("expected completion hint auto, got %v", opt.completionHint)
	}
}

func TestOptionIOAutoNormalizesDash(t *testing.T) {
	var opts struct {
		Output string `long:"output" io:"out" io-kind:"auto" io-stream:"stderr"`
	}

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"output", "-")
	assertString(t, opts.Output, "stderr")
}

func TestOptionIOFileRejectsStreamToken(t *testing.T) {
	var opts struct {
		Output string `long:"output" io:"out" io-kind:"file"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{defaultLongOptDelimiter + "output", "-"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if flagsErr.Type != ErrMarshal {
		t.Fatalf("expected ErrMarshal, got %v", flagsErr.Type)
	}
	if !strings.Contains(flagsErr.Message, "io-kind `file` does not allow stream token `-`") {
		t.Fatalf("unexpected message: %s", flagsErr.Message)
	}
}

func TestOptionIONoFallbackWhenFlagIsMissing(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-kind:"auto"`
	}

	assertParseSuccess(t, &opts)
	assertString(t, opts.Input, "")
}

func TestOptionIOTagRequiresStringType(t *testing.T) {
	var opts struct {
		Input int `long:"input" io:"in"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs(nil)
	assertError(t, err, ErrInvalidTag, "field `Input` with tag `io` must be a string option")
}

func TestOptionIORejectsOpenTagForInputRole(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-open:"append"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs(nil)
	assertError(
		t,
		err,
		ErrInvalidTag,
		"tag `io-open` on field `Input` requires `io:\"out\"`",
	)
}

func TestOptionIORejectsInvalidKindValue(t *testing.T) {
	var opts struct {
		Input string `long:"input" io:"in" io-kind:"wat"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs(nil)
	assertError(
		t,
		err,
		ErrInvalidTag,
		"invalid value `wat` for tag `io-kind` on field `Input` (expected auto, stream, file, or string)",
	)
}

func TestOptionIORejectsInvalidStreamValue(t *testing.T) {
	var opts struct {
		Output string `long:"output" io:"out" io-stream:"wat"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs(nil)
	assertError(
		t,
		err,
		ErrInvalidTag,
		"invalid value `wat` for tag `io-stream` on field `Output` (expected stdin, stdout, or stderr)",
	)
}
