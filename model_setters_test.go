package flags

import (
	"os"
	"testing"
)

func TestCommandSettersLookup(t *testing.T) {
	var opts struct {
		Run struct{} `command:"run" description:"Run command"`
	}

	p := NewParser(&opts, None)
	cmd := p.Find("run")
	if cmd == nil {
		t.Fatalf("expected command run")
	}

	cmd.SetName("execute")
	cmd.SetAliases("ex")
	cmd.SetShortDescription("Execute command")
	cmd.SetLongDescription("Execute long description")

	if p.Find("run") != nil {
		t.Fatalf("expected old command name to be removed from lookup")
	}
	if p.Find("execute") == nil || p.Find("ex") == nil {
		t.Fatalf("expected renamed command and alias in lookup")
	}

	if _, err := p.ParseArgs([]string{"execute"}); err != nil {
		t.Fatalf("unexpected parse error for renamed command: %v", err)
	}
}

func TestGroupSetNamespaceAffectsLookup(t *testing.T) {
	var opts struct {
		DB struct {
			Host string `long:"host" env:"HOST"`
		} `group:"Database"`
	}

	p := NewParser(&opts, None)
	groups := p.Groups()
	if len(groups) == 0 {
		t.Fatalf("expected at least one group")
	}

	dbGroup := groups[0]
	dbGroup.SetNamespace("db")
	dbGroup.SetEnvNamespace("DB")

	if _, err := p.ParseArgs([]string{"--db.host", "localhost"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if opts.DB.Host != "localhost" {
		t.Fatalf("expected host from namespaced long option, got %q", opts.DB.Host)
	}

	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	oldEnv.Restore()
	_ = os.Setenv("DB_HOST", "db-from-env")

	var opts2 struct {
		DB struct {
			Host string `long:"host" env:"HOST"`
		} `group:"Database"`
	}

	p2 := NewParser(&opts2, None)
	p2.Groups()[0].SetEnvNamespace("DB")
	if _, err := p2.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error with env namespace: %v", err)
	}
	if opts2.DB.Host != "db-from-env" {
		t.Fatalf("expected host from env namespace, got %q", opts2.DB.Host)
	}
}

func TestArgSettersDefaultAndRequiredRange(t *testing.T) {
	var opts struct {
		Positional struct {
			Target string `description:"Target"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	args := p.Command.Args()
	if len(args) != 1 {
		t.Fatalf("expected one positional arg")
	}

	args[0].SetName("target")
	args[0].SetDescription("Deployment target")
	args[0].SetDefault("local")

	if _, err := p.ParseArgs(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if opts.Positional.Target != "local" {
		t.Fatalf("expected positional default local, got %q", opts.Positional.Target)
	}

	var opts2 struct {
		Positional struct {
			Values []string
		} `positional-args:"yes"`
	}

	p2 := NewParser(&opts2, None)
	a := p2.Command.Args()[0]
	if err := a.SetRequiredRange(2, 3); err != nil {
		t.Fatalf("unexpected SetRequiredRange error: %v", err)
	}

	_, err := p2.ParseArgs([]string{"one"})
	if err == nil {
		t.Fatalf("expected required positional range error")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrRequired {
		t.Fatalf("expected ErrRequired, got %v", err)
	}

	if err := a.SetRequiredRange(3, 2); err == nil {
		t.Fatalf("expected invalid required range error")
	}
}
