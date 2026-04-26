package flags

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestComplete struct {
}

func (t *TestComplete) Complete(match string) []Completion {
	options := []string{
		"hello world",
		"hello universe",
		"hello multiverse",
	}

	ret := make([]Completion, 0, len(options))

	for _, o := range options {
		if strings.HasPrefix(o, match) {
			ret = append(ret, Completion{
				Item: o,
			})
		}
	}

	return ret
}

var completionTestOptions struct {
	Verbose  bool   `short:"v" long:"verbose" long-alias:"chatty" short-alias:"V" description:"Verbose messages"`
	Debug    bool   `short:"d" long:"debug" description:"Enable debug"`
	Info     bool   `short:"i" description:"Display info"`
	Version  bool   `long:"version" description:"Show version"`
	Mode     string `long:"mode" choices:"fast;safe;full" description:"Run mode"`
	Required bool   `long:"required" required:"true" description:"This is required"`
	Hidden   bool   `long:"hidden" hidden:"true" description:"This is hidden"`

	AddCommand struct {
		Positional struct {
			Filename Filename
		} `positional-args:"yes"`
	} `command:"add" description:"add an item"`

	AddMultiCommand struct {
		Positional struct {
			Filename []Filename
		} `positional-args:"yes"`
		Extra []Filename `short:"f"`
	} `command:"add-multi" description:"add multiple items"`

	AddMultiCommandFlag struct {
		Files []Filename `short:"f"`
	} `command:"add-multi-flag" description:"add multiple items via flags"`

	RemoveCommand struct {
		Other bool     `short:"o"`
		File  Filename `short:"f" long:"filename"`
	} `command:"rm" aliases:"remove;delete" description:"remove an item"`

	RenameCommand struct {
		Completed TestComplete `short:"c" long:"completed"`
	} `command:"rename" description:"rename an item"`

	HiddenCommand struct {
	} `command:"hidden" description:"hidden command" hidden:"true"`
}

type completionTest struct {
	Args             []string
	Completed        []string
	ShowDescriptions bool
}

var completionTests []completionTest

func makeLongName(option string) string {
	return defaultLongOptDelimiter + option
}

func makeShortName(option string) string {
	return string(defaultShortOptDelimiter) + option
}

func init() {
	_, sourcefile, _, _ := runtime.Caller(0)
	completionTestSourcedir := filepath.Join(filepath.SplitList(path.Dir(sourcefile))...)

	completionTestFilename, _ := filepath.Glob(filepath.Join(completionTestSourcedir, "completion*"))
	completionTestFilenameShort := make([]string, 0, len(completionTestFilename))
	completionTestFilenameShortEq := make([]string, 0, len(completionTestFilename))
	completionTestFilenameLongEq := make([]string, 0, len(completionTestFilename))

	for _, v := range completionTestFilename {
		completionTestFilenameShort = append(completionTestFilenameShort, "-f"+v)
		completionTestFilenameShortEq = append(completionTestFilenameShortEq, "-f="+v)
		completionTestFilenameLongEq = append(completionTestFilenameLongEq, "--filename="+v)
	}

	completionTestSubdir := []string{
		filepath.Join(completionTestSourcedir, "examples/advanced"),
		filepath.Join(completionTestSourcedir, "examples/basic"),
		filepath.Join(completionTestSourcedir, "examples/completion"),
		filepath.Join(completionTestSourcedir, "examples/custom-flag-tags"),
		filepath.Join(completionTestSourcedir, "examples/doc-rendered"),
		filepath.Join(completionTestSourcedir, "examples/i18n"),
	}

	completionTests = []completionTest{
		{
			// Short names
			[]string{makeShortName("")},
			[]string{makeLongName("chatty"), makeLongName("debug"), makeLongName("mode"), makeLongName("required"), makeLongName("verbose"), makeLongName("version"), makeShortName("i")},
			false,
		},

		{
			// Short names full
			[]string{makeShortName("i")},
			[]string{makeShortName("i")},
			false,
		},

		{
			// Short names concatenated
			[]string{"-dv"},
			[]string{"-dv"},
			false,
		},

		{
			// Long names
			[]string{"--"},
			[]string{"--chatty", "--debug", "--mode", "--required", "--verbose", "--version"},
			false,
		},

		{
			// Long names with descriptions
			[]string{"--"},
			[]string{
				"--chatty    # Verbose messages",
				"--debug     # Enable debug",
				"--mode      # Run mode",
				"--required  # This is required",
				"--verbose   # Verbose messages",
				"--version   # Show version",
			},
			true,
		},

		{
			// Long names partial
			[]string{makeLongName("ver")},
			[]string{makeLongName("verbose"), makeLongName("version")},
			false,
		},

		{
			// Commands
			[]string{""},
			[]string{"add", "add-multi", "add-multi-flag", "delete", "remove", "rename", "rm"},
			false,
		},

		{
			// Commands with descriptions
			[]string{""},
			[]string{
				"add             # add an item",
				"add-multi       # add multiple items",
				"add-multi-flag  # add multiple items via flags",
				"delete          # remove an item",
				"remove          # remove an item",
				"rename          # rename an item",
				"rm              # remove an item",
			},
			true,
		},

		{
			// Commands partial
			[]string{"r"},
			[]string{"remove", "rename", "rm"},
			false,
		},

		{
			// Positional filename
			[]string{"add", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
			false,
		},

		{
			// Multiple positional filename (1 arg)
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
			false,
		},
		{
			// Multiple positional filename (2 args)
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
			false,
		},
		{
			// Multiple positional filename (3 args)
			[]string{
				"add-multi",
				filepath.Join(completionTestSourcedir, "completion.go"),
				filepath.Join(completionTestSourcedir, "completion.go"),
				filepath.Join(completionTestSourcedir, "completion"),
			},
			completionTestFilename,
			false,
		},

		{
			// Flag filename
			[]string{"rm", makeShortName("f"), filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
			false,
		},

		{
			// Flag short concat last filename
			[]string{"rm", "-of", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
			false,
		},

		{
			// Flag concat filename
			[]string{"rm", "-f" + filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilenameShort,
			false,
		},

		{
			// Flag equal concat filename
			[]string{"rm", "-f=" + filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilenameShortEq,
			false,
		},

		{
			// Flag concat long filename
			[]string{"rm", "--filename=" + filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilenameLongEq,
			false,
		},

		{
			// Flag long filename
			[]string{"rm", "--filename", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
			false,
		},

		{
			// To subdir
			[]string{"rm", "--filename", filepath.Join(completionTestSourcedir, "examples/completion/ba")},
			[]string{filepath.Join(completionTestSourcedir, "examples/completion/bash")},
			false,
		},

		{
			// Subdirectory
			[]string{"rm", "--filename", filepath.Join(completionTestSourcedir, "examples") + "/"},
			completionTestSubdir,
			false,
		},

		{
			// Custom completed
			[]string{"rename", makeShortName("c"), "hello un"},
			[]string{"hello universe"},
			false,
		},
		{
			// Choice completion
			[]string{"--mode", "f"},
			[]string{"fast", "full"},
			false,
		},
		{
			// Choice completion with equals form
			[]string{"--mode=f"},
			[]string{"--mode=fast", "--mode=full"},
			false,
		},
		{
			// Bool completion for inline assignment when AllowBoolValues is enabled.
			[]string{"--debug=t"},
			[]string{"--debug=true"},
			false,
		},
		{
			// Bool completion should suggest both values on empty inline argument.
			[]string{"--debug="},
			[]string{"--debug=false", "--debug=true"},
			false,
		},
		{
			// Multiple flag filename
			[]string{"add-multi-flag", makeShortName("f"), filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
			false,
		},
	}
}

func TestCompletion(t *testing.T) {
	p := NewParser(&completionTestOptions, Default|AllowBoolValues)
	c := &completion{parser: p}

	for _, test := range completionTests {
		if test.ShowDescriptions {
			continue
		}

		ret := c.complete(test.Args)
		items := make([]string, len(ret))

		for i, v := range ret {
			items[i] = v.Item
		}

		sort.Strings(items)
		sort.Strings(test.Completed)

		if diff := cmp.Diff(test.Completed, items); diff != "" {
			t.Errorf("Args: %#v, showDescriptions=%v, mismatch (-expected +actual):\n%s", test.Args, test.ShowDescriptions, diff)
		}
	}
}

func TestParserCompletion(t *testing.T) {
	for _, test := range completionTests {
		if test.ShowDescriptions {
			os.Setenv("GO_FLAGS_COMPLETION", "verbose")
		} else {
			os.Setenv("GO_FLAGS_COMPLETION", "1")
		}

		tmp := os.Stdout

		r, w, _ := os.Pipe()
		os.Stdout = w

		out := make(chan string)

		go func() {
			var buf bytes.Buffer

			io.Copy(&buf, r)

			out <- buf.String()
		}()

		p := NewParser(&completionTestOptions, None|AllowBoolValues)

		p.CompletionHandler = func(items []Completion) {
			comp := &completion{parser: p}
			comp.print(items, test.ShowDescriptions)
		}

		_, err := p.ParseArgs(test.Args)

		w.Close()

		os.Stdout = tmp

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		got := strings.Split(strings.Trim(<-out, "\n"), "\n")

		if diff := cmp.Diff(test.Completed, got); diff != "" {
			t.Errorf("Completion output mismatch (-expected +actual):\n%s", diff)
		}
	}

	os.Setenv("GO_FLAGS_COMPLETION", "")
}

func TestCompletionSkipPositional(t *testing.T) {
	c := &completion{}
	s := &parseState{
		positional: []*Arg{{}, {}},
	}

	c.skipPositional(s, 1)
	if len(s.positional) != 1 {
		t.Fatalf("expected 1 positional item left, got %d", len(s.positional))
	}

	c.skipPositional(s, 5)
	if s.positional != nil {
		t.Fatalf("expected positional args to be cleared")
	}
}

func completionItemsForArgs(t *testing.T, parser *Parser, args []string) []string {
	t.Helper()

	c := &completion{parser: parser}
	completions := c.complete(args)
	items := make([]string, len(completions))
	for i, item := range completions {
		items[i] = item.Item
	}
	sort.Strings(items)

	return items
}

func TestCompletionHintFile(t *testing.T) {
	tempDir := t.TempDir()
	first := filepath.Join(tempDir, "cfg-a.yaml")
	second := filepath.Join(tempDir, "cfg-b.yaml")
	if err := os.WriteFile(first, []byte("a"), 0o600); err != nil {
		t.Fatalf("write file error: %v", err)
	}
	if err := os.WriteFile(second, []byte("b"), 0o600); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	var opts struct {
		Config string `long:"config" completion:"file"`
	}
	p := NewParser(&opts, None)
	items := completionItemsForArgs(t, p, []string{"--config", filepath.Join(tempDir, "cfg-")})

	want := []string{first, second}
	sort.Strings(want)
	if diff := cmp.Diff(want, items); diff != "" {
		t.Fatalf("completion mismatch (-want +got):\n%s", diff)
	}
}

func TestCompletionHintDir(t *testing.T) {
	tempDir := t.TempDir()
	dirA := filepath.Join(tempDir, "work-a")
	dirB := filepath.Join(tempDir, "work-b")
	if err := os.Mkdir(dirA, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.Mkdir(dirB, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "work-file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	var opts struct {
		Dir string `long:"dir" completion:"dir"`
	}
	p := NewParser(&opts, None)
	items := completionItemsForArgs(t, p, []string{"--dir", filepath.Join(tempDir, "work-")})

	want := []string{dirA + "/", dirB + "/"}
	sort.Strings(want)
	if diff := cmp.Diff(want, items); diff != "" {
		t.Fatalf("completion mismatch (-want +got):\n%s", diff)
	}
}

func TestCompletionHintNoneDisablesBoolCompletion(t *testing.T) {
	var opts struct {
		Debug bool `long:"debug" completion:"none"`
	}
	p := NewParser(&opts, None|AllowBoolValues)
	items := completionItemsForArgs(t, p, []string{"--debug="})
	if len(items) != 0 {
		t.Fatalf("expected no completion items, got %#v", items)
	}
}

func TestCompletionHintChoicesOverrideNone(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" choices:"fast;safe" completion:"none"`
	}
	p := NewParser(&opts, None)
	items := completionItemsForArgs(t, p, []string{"--mode=f"})

	want := []string{"--mode=fast"}
	if diff := cmp.Diff(want, items); diff != "" {
		t.Fatalf("completion mismatch (-want +got):\n%s", diff)
	}
}

func TestCompletionHintPositionalDir(t *testing.T) {
	tempDir := t.TempDir()
	dir := filepath.Join(tempDir, "target-dir")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "target-file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	var opts struct {
		Run struct {
			Positional struct {
				Path string `completion:"dir"`
			} `positional-args:"yes"`
		} `command:"run"`
	}
	p := NewParser(&opts, None)
	items := completionItemsForArgs(t, p, []string{"run", filepath.Join(tempDir, "target-")})

	want := []string{dir + "/"}
	if diff := cmp.Diff(want, items); diff != "" {
		t.Fatalf("completion mismatch (-want +got):\n%s", diff)
	}
}
