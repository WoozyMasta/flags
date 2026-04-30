package flags

import "testing"

func TestPositionalIOFallbackDefaults(t *testing.T) {
	var opts struct {
		IO struct {
			Input  string `io:"in"`
			Output string `io:"out"`
		} `positional-args:"yes"`
	}

	assertParseSuccess(t, &opts)
	assertString(t, opts.IO.Input, "stdin")
	assertString(t, opts.IO.Output, "stdout")
}

func TestPositionalIONormalizesDashToken(t *testing.T) {
	var opts struct {
		IO struct {
			Input  string `io:"in" io-kind:"stream"`
			Output string `io:"out" io-kind:"stream" io-stream:"stderr"`
		} `positional-args:"yes"`
	}

	assertParseSuccess(t, &opts, "-", "-")
	assertString(t, opts.IO.Input, "stdin")
	assertString(t, opts.IO.Output, "stderr")
}

func TestPositionalIOAutoKeepsFilePath(t *testing.T) {
	var opts struct {
		IO struct {
			Input string `io:"in" io-kind:"auto"`
		} `positional-args:"yes"`
	}

	assertParseSuccess(t, &opts, "input.txt")
	assertString(t, opts.IO.Input, "input.txt")
}

func TestPositionalIORejectsInvalidStreamForRole(t *testing.T) {
	var opts struct {
		IO struct {
			Input string `io:"in" io-kind:"stream"`
		} `positional-args:"yes"`
	}

	assertParseFail(
		t,
		ErrMarshal,
		"invalid positional argument `Input`: io `in` accepts only `stdin` or `-`",
		&opts,
		"stderr",
	)
}

func TestPositionalIORejectsStreamTokenInFileMode(t *testing.T) {
	var opts struct {
		IO struct {
			Output string `io:"out" io-kind:"file"`
		} `positional-args:"yes"`
	}

	assertParseFail(
		t,
		ErrMarshal,
		"invalid positional argument `Output`: io-kind `file` does not allow stream token `-`",
		&opts,
		"-",
	)
}

func TestPositionalIOTagRequiresStringType(t *testing.T) {
	var opts struct {
		IO struct {
			Input int `io:"in"`
		} `positional-args:"yes"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs(nil)
	assertError(t, err, ErrInvalidTag, "field `Input` with tag `io` must be a string positional argument")
}

func TestPositionalIOInfersFileCompletionWhenUnset(t *testing.T) {
	var opts struct {
		IO struct {
			Input string `io:"in" io-kind:"auto"`
		} `positional-args:"yes"`
	}

	parser := NewParser(&opts, None)
	args := parser.Command.Args()
	if len(args) != 1 {
		t.Fatalf("expected 1 positional arg, got %d", len(args))
	}
	if args[0].completionHint != completionHintFile {
		t.Fatalf("expected completion hint file, got %v", args[0].completionHint)
	}
}

func TestPositionalIOKeepsExplicitCompletion(t *testing.T) {
	var opts struct {
		IO struct {
			Input string `io:"in" io-kind:"auto" completion:"none"`
		} `positional-args:"yes"`
	}

	parser := NewParser(&opts, None)
	args := parser.Command.Args()
	if len(args) != 1 {
		t.Fatalf("expected 1 positional arg, got %d", len(args))
	}
	if args[0].completionHint != completionHintNone {
		t.Fatalf("expected completion hint none, got %v", args[0].completionHint)
	}
}
