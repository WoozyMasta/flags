package flags

import "testing"

func TestCommandExampleBuilderPreservesDeclarationOrder(t *testing.T) {
	var opts struct {
		Name string `long:"name" short:"n"`
	}

	example := Example().
		Describe("Create a resource").
		Arg("./output").
		Option(&opts.Name, "value").
		ShortOption(&opts.Name).
		Raw("> result.txt")

	if example.description != "Create a resource" {
		t.Fatalf("unexpected description: %q", example.description)
	}
	if len(example.parts) != 4 {
		t.Fatalf("unexpected part count: %d", len(example.parts))
	}

	wantKinds := []commandExamplePartKind{
		commandExampleArg,
		commandExampleOption,
		commandExampleShortOption,
		commandExampleRaw,
	}
	for idx, want := range wantKinds {
		if example.parts[idx].kind != want {
			t.Fatalf("part %d kind = %d, want %d", idx, example.parts[idx].kind, want)
		}
	}
	if example.parts[0].value != "./output" || example.parts[1].value != "value" ||
		example.parts[3].value != "> result.txt" {
		t.Fatalf("unexpected part values: %#v", example.parts)
	}
	if example.parts[1].target != &opts.Name || example.parts[2].target != &opts.Name {
		t.Fatal("option targets were not preserved")
	}
}

func TestCommandExampleBuilderCopiesOptionValue(t *testing.T) {
	values := []string{"initial", "ignored"}
	example := Example().Option(nil, values...)
	values[0] = "changed"

	if example.parts[0].value != "initial" {
		t.Fatalf("option value changed after builder call: %q", example.parts[0].value)
	}
}

func TestCommandExamplesAreCopiedOnSetAndGet(t *testing.T) {
	var opts struct {
		Command struct{} `command:"command"`
	}
	p := NewParser(&opts, None)
	command := p.Command.Find("command")
	example := Example().Describe("description").Arg("value")

	command.SetExamples(example)
	example.parts[0].value = "changed"
	stored := command.Examples()
	stored[0].parts[0].value = "mutated"

	got := command.Examples()
	if got[0].parts[0].value != "value" {
		t.Fatalf("stored example was mutated through caller data: %q", got[0].parts[0].value)
	}

	command.SetExamples()
	if len(command.Examples()) != 0 {
		t.Fatal("empty SetExamples did not clear examples")
	}
}

func TestCommandExamplesSurviveParserRebuild(t *testing.T) {
	var opts struct {
		Command struct{} `command:"command"`
	}
	p := NewParser(&opts, None)
	command := p.Command.Find("command")
	command.SetExamples(Example().Arg("value"))

	if err := p.SetTagListDelimiter(';'); err != nil {
		t.Fatalf("unexpected rebuild error: %v", err)
	}

	rebuilt, err := p.CommandFor(&opts.Command)
	if err != nil {
		t.Fatalf("unexpected command lookup error: %v", err)
	}
	if examples := rebuilt.Examples(); len(examples) != 1 || examples[0].parts[0].value != "value" {
		t.Fatalf("examples did not survive rebuild: %#v", examples)
	}
}

func TestSetCommandExamplesSupportsNestedPathsAtomically(t *testing.T) {
	var opts struct {
		Profile struct {
			Export struct{} `command:"export"`
		} `command:"profile"`
	}
	p := NewParser(&opts, None)

	example := Example().Arg("profile.yaml")
	if err := p.SetCommandExamples(map[string][]*CommandExample{
		"profile export": {example},
	}); err != nil {
		t.Fatalf("unexpected nested command error: %v", err)
	}

	command, err := p.CommandFor(&opts.Profile.Export)
	if err != nil {
		t.Fatalf("unexpected command lookup error: %v", err)
	}
	if len(command.Examples()) != 1 {
		t.Fatal("nested command examples were not stored")
	}

	if err := p.SetCommandExamples(map[string][]*CommandExample{
		"profile export": {Example().Arg("new.yaml")},
		"missing":        {Example().Arg("ignored.yaml")},
	}); err == nil {
		t.Fatal("expected unknown command path error")
	}
	if got := command.Examples()[0].parts[0].value; got != "profile.yaml" {
		t.Fatalf("invalid registration partially changed examples: %q", got)
	}
}
