package flags

import (
	"errors"
	"testing"
)

// Without StrictCommands and with SubcommandsOptional=true, unknown tokens
// become retargs without error (the lenient baseline).
func TestStrictCommandsSubcommandsOptionalLenient(t *testing.T) {
	var opts = struct {
		Cmd struct {
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, None)
	p.SubcommandsOptional = true

	ret, err := p.ParseArgs([]string{"nocmd"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertStringArray(t, ret, []string{"nocmd"})
}

// StrictCommands errors on an unknown token even when SubcommandsOptional is true.
func TestStrictCommandsSubcommandsOptional(t *testing.T) {
	var opts = struct {
		Cmd struct {
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, StrictCommands)
	p.SubcommandsOptional = true

	_, err := p.ParseArgs([]string{"nocmd"})

	// estimateCommand: 1 cmd → "should use" suffix
	assertError(t, err, ErrUnknownCommand, "Unknown command `nocmd`. You should use the cmd command")
}

// StrictCommands errors on an unknown subcommand even when positional args are
// defined on the active command.
func TestStrictCommandsUnknownWithPositional(t *testing.T) {
	var opts = struct {
		Cmd struct {
			Sub struct {
			} `command:"sub"`
			Positional struct {
				Name string
			} `positional-args:"yes"`
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, StrictCommands)
	_, err := p.ParseArgs([]string{"cmd", "notasub"})

	// estimateCommand on cmd: 1 subcommand "sub" → "should use" suffix
	assertError(t, err, ErrUnknownCommand, "Unknown command `notasub`. You should use the sub command")
}

// Known commands still succeed with StrictCommands enabled.
func TestStrictCommandsKnownCommandSucceeds(t *testing.T) {
	var opts = struct {
		Cmd struct {
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, StrictCommands)
	_, err := p.ParseArgs([]string{"cmd"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// Command.StrictSubcommands provides per-command strictness without the global flag.
func TestStrictSubcommandsPerCommand(t *testing.T) {
	var opts = struct {
		Cmd struct {
			Sub struct {
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, None)
	p.Find("cmd").StrictSubcommands = true

	_, err := p.ParseArgs([]string{"cmd", "unknown"})

	// estimateCommand on cmd: 1 subcommand "sub" → "should use" suffix
	assertError(t, err, ErrUnknownCommand, "Unknown command `unknown`. You should use the sub command")
}

// StrictSubcommands on a child does not affect the root or sibling commands.
func TestStrictSubcommandsDoesNotAffectParent(t *testing.T) {
	var opts = struct {
		Cmd struct {
			Sub struct {
			} `command:"sub"`
		} `command:"cmd"`
		Other struct {
		} `command:"other"`
	}{}

	p := NewParser(&opts, None)
	p.Find("cmd").StrictSubcommands = true

	// "other" is known at root — should succeed
	_, err := p.ParseArgs([]string{"other"})
	if err != nil {
		t.Fatalf("Unexpected error for known command: %v", err)
	}

	// "cmd sub" — known subcommand — should succeed
	_, err = p.ParseArgs([]string{"cmd", "sub"})
	if err != nil {
		t.Fatalf("Unexpected error for cmd sub: %v", err)
	}

	// "cmd unknown" — cmd is strict — should error
	_, err = p.ParseArgs([]string{"cmd", "unknown"})
	assertError(t, err, ErrUnknownCommand, "Unknown command `unknown`. You should use the sub command")
}

// UnknownCommandHandler returning nil adds the token to retargs and continues.
func TestUnknownCommandHandlerNilReturn(t *testing.T) {
	var opts = struct {
		Cmd struct {
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, None)
	var called string
	p.UnknownCommandHandler = func(command string, args []string) error {
		called = command
		return nil
	}

	ret, err := p.ParseArgs([]string{"nocmd"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertString(t, called, "nocmd")
	assertStringArray(t, ret, []string{"nocmd"})
}

// UnknownCommandHandler returning an error propagates that error to the caller.
func TestUnknownCommandHandlerErrorReturn(t *testing.T) {
	var opts = struct {
		Cmd struct {
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, None)
	sentinel := errors.New("custom handler error")
	p.UnknownCommandHandler = func(command string, args []string) error {
		return sentinel
	}

	_, err := p.ParseArgs([]string{"nocmd"})

	if !errors.Is(err, sentinel) {
		t.Errorf("Expected sentinel error, got: %v", err)
	}
}

// UnknownCommandHandler receives the remaining CLI args after the unknown token.
func TestUnknownCommandHandlerReceivesRemainingArgs(t *testing.T) {
	var opts = struct {
		Cmd struct {
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, None)
	var gotArgs []string
	p.UnknownCommandHandler = func(command string, args []string) error {
		gotArgs = args
		return nil
	}

	_, err := p.ParseArgs([]string{"nocmd", "extra1", "extra2"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertStringArray(t, gotArgs, []string{"extra1", "extra2"})
}

// UnknownCommandHandler is called even when StrictCommands is active; returning nil
// adds the token to retargs.
func TestUnknownCommandHandlerWithStrictCommands(t *testing.T) {
	var opts = struct {
		Cmd struct {
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, StrictCommands)
	p.SubcommandsOptional = true

	var called string
	p.UnknownCommandHandler = func(command string, args []string) error {
		called = command
		return nil
	}

	ret, err := p.ParseArgs([]string{"nocmd"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertString(t, called, "nocmd")
	assertStringArray(t, ret, []string{"nocmd"})
}

// StrictPositionalArgs errors when an extra arg is passed beyond declared count.
func TestStrictPositionalArgsExtra(t *testing.T) {
	var opts = struct {
		Positional struct {
			Name string
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, StrictPositionalArgs)
	_, err := p.ParseArgs([]string{"known", "extra"})

	assertError(t, err, ErrUnexpectedArgument, "Unexpected argument `extra`")
}

// StrictPositionalArgs errors when no positional args are declared but one is passed.
func TestStrictPositionalArgsNoDeclared(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, StrictPositionalArgs)
	_, err := p.ParseArgs([]string{"-v", "unexpected"})

	assertError(t, err, ErrUnexpectedArgument, "Unexpected argument `unexpected`")
}

// StrictPositionalArgs does not restrict a slice (remaining) positional field —
// it acts as an unbounded collector.
func TestStrictPositionalArgsSliceUnlimited(t *testing.T) {
	var opts = struct {
		Positional struct {
			Rest []string
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, StrictPositionalArgs)
	_, err := p.ParseArgs([]string{"a", "b", "c"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertStringArray(t, opts.Positional.Rest, []string{"a", "b", "c"})
}

// Command.StrictArgs provides per-command positional arg strictness.
func TestStrictArgsPerCommand(t *testing.T) {
	var opts = struct {
		Cmd struct {
			Positional struct {
				Name string
			} `positional-args:"yes"`
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, None)
	p.Find("cmd").StrictArgs = true

	_, err := p.ParseArgs([]string{"cmd", "known", "extra"})

	assertError(t, err, ErrUnexpectedArgument, "Unexpected argument `extra`")
}

// Command.StrictArgs only restricts positional args on that specific command.
func TestStrictArgsDoesNotAffectRoot(t *testing.T) {
	var opts = struct {
		Cmd struct {
			Positional struct {
				Name string
			} `positional-args:"yes"`
		} `command:"cmd"`
	}{}

	p := NewParser(&opts, None)
	p.Find("cmd").StrictArgs = true
	p.SubcommandsOptional = true

	// Extra arg at root level — no strict at root — goes to retargs
	ret, err := p.ParseArgs([]string{"extra"})
	if err != nil {
		t.Fatalf("Unexpected error at root level: %v", err)
	}
	assertStringArray(t, ret, []string{"extra"})
}

// Args after -- bypass strict positional checking and go to retargs.
func TestStrictPositionalArgsPassDoubleDash(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, StrictPositionalArgs|PassDoubleDash)
	ret, err := p.ParseArgs([]string{"-v", "--", "a", "b"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertStringArray(t, ret, []string{"a", "b"})
}
