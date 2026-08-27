package flags

import (
	"strings"
	"testing"
)

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

	if err := command.SetExamples(example); err != nil {
		t.Fatalf("unexpected SetExamples error: %v", err)
	}
	example.parts[0].value = "changed"
	stored := command.Examples()
	stored[0].parts[0].value = "mutated"

	got := command.Examples()
	if got[0].parts[0].value != "value" {
		t.Fatalf("stored example was mutated through caller data: %q", got[0].parts[0].value)
	}

	if err := command.SetExamples(); err != nil {
		t.Fatalf("unexpected clear examples error: %v", err)
	}
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
	if err := command.SetExamples(Example().Arg("value")); err != nil {
		t.Fatalf("unexpected SetExamples error: %v", err)
	}

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

func TestCommandExampleOptionResolutionUsesCommandScope(t *testing.T) {
	var opts struct {
		Global  string `long:"global"`
		Command struct {
			Local string `long:"local" short:"l"`
		} `command:"command"`
		Sibling struct {
			Other string `long:"other"`
		} `command:"sibling"`
	}
	p := NewParser(&opts, None)
	command := p.Command.Find("command")

	global, err := command.resolveExampleOption(&opts.Global)
	if err != nil || global.LongName != "global" {
		t.Fatalf("global option resolution = %v, %v", global, err)
	}
	local, err := command.resolveExampleOption(&opts.Command.Local)
	if err != nil || local.LongName != "local" {
		t.Fatalf("local option resolution = %v, %v", local, err)
	}
	if _, err := command.resolveExampleOption(&opts.Sibling.Other); err == nil {
		t.Fatal("sibling option was accepted")
	}
	if _, err := command.resolveExampleOption(nil); err == nil {
		t.Fatal("nil option target was accepted")
	}

	if err := local.SetLongName("renamed"); err != nil {
		t.Fatalf("unexpected long name update error: %v", err)
	}
	if err := local.SetShortName('r'); err != nil {
		t.Fatalf("unexpected short name update error: %v", err)
	}
	resolved, err := command.resolveExampleOption(&opts.Command.Local)
	if err != nil || resolved.LongName != "renamed" || resolved.ShortName != 'r' {
		t.Fatalf("renamed option resolution = %v, %v", resolved, err)
	}
}

func TestRenderCommandExample(t *testing.T) {
	var opts struct {
		Global  string `long:"global"`
		Command struct {
			Local string `long:"local-name" short:"l"`
		} `command:"command"`
	}
	p := NewParser(&opts, None)
	command := p.Command.Find("command")
	example := Example().
		Describe("Run command").
		Arg("path with spaces").
		Option(&opts.Global, "$value").
		ShortOption(&opts.Command.Local, "local").
		Raw("| jq '.')")

	bash, err := renderCommandExample(command, "my-app", ExampleShellBash, example)
	if err != nil {
		t.Fatalf("unexpected Bash render error: %v", err)
	}
	if bash.Command != "my-app command 'path with spaces' --global '$value' -l local | jq '.')" {
		t.Fatalf("unexpected Bash example: %q", bash.Command)
	}

	pwsh, err := renderCommandExample(command, "my app", ExampleShellPwsh, Example().Arg("it's `$value"))
	if err != nil {
		t.Fatalf("unexpected PowerShell render error: %v", err)
	}
	if pwsh.Command != "'my app' command 'it''s `$value'" {
		t.Fatalf("unexpected PowerShell example: %q", pwsh.Command)
	}
}

func TestWriteHelpIncludesCommandExamples(t *testing.T) {
	var opts struct {
		Command struct{} `command:"command"`
	}
	p := NewParser(&opts, None)
	command := p.Command.Find("command")
	if err := command.SetExamples(Example().Describe("Run it").Arg("value")); err != nil {
		t.Fatalf("unexpected SetExamples error: %v", err)
	}
	if _, err := p.ParseArgs([]string{"command"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	var out strings.Builder
	p.WriteHelp(&out)
	got := out.String()
	if !strings.Contains(got, "Examples:") || !strings.Contains(got, "Run it:") ||
		!strings.Contains(got, "command value") {
		t.Fatalf("help does not contain command example:\n%s", got)
	}
}

func TestMarkdownDocumentationIncludesCommandExamples(t *testing.T) {
	var opts struct {
		Namespace string `long:"namespace"`
		Command   struct {
			Value string `long:"value"`
		} `command:"command"`
	}
	p := NewParser(&opts, None)
	command := p.Command.Find("command")
	if err := command.SetExamples(Example().Describe("Run command").Option(&opts.Namespace, "prod").Option(&opts.Command.Value, "value")); err != nil {
		t.Fatalf("unexpected SetExamples error: %v", err)
	}

	var out strings.Builder
	if err := p.WriteDoc(&out, DocFormatMarkdown, WithProgramName("app")); err != nil {
		t.Fatalf("unexpected documentation error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "#### Examples") || !strings.Contains(got, "app command --namespace prod --value value") {
		t.Fatalf("markdown documentation does not contain command example:\n%s", got)
	}
}

func TestHTMLAndManDocumentationIncludeCommandExamples(t *testing.T) {
	var opts struct {
		Command struct{} `command:"command"`
	}
	p := NewParser(&opts, None)
	command := p.Command.Find("command")
	if err := command.SetExamples(Example().Describe("Run it").Arg("value")); err != nil {
		t.Fatalf("unexpected SetExamples error: %v", err)
	}

	for _, format := range []DocFormat{DocFormatHTML, DocFormatMan} {
		var out strings.Builder
		if err := p.WriteDoc(&out, format, WithProgramName("app")); err != nil {
			t.Fatalf("unexpected %s documentation error: %v", format, err)
		}
		got := out.String()
		if !strings.Contains(got, "Examples") || !strings.Contains(got, "app command value") {
			t.Fatalf("%s documentation does not contain command example:\n%s", format, got)
		}
	}
}
