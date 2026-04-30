package flags

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateStringOptionTags(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.txt")
	writeTestFile(t, file, "data")

	var opts struct {
		Input  string `long:"input" validate-existing-file:"true" validate-readable:"true"`
		Output string `long:"output" validate-writable:"true" validate-path-abs:"true"`
		Name   string `long:"name" validate-non-empty:"true" validate-regex:"[a-z]+" validate-min-len:"3" validate-max-len:"8"`
	}

	assertParseSuccess(t, &opts,
		"--input", file,
		"--output", filepath.Join(dir, "out.txt"),
		"--name", "worker",
	)
}

func TestValidateStringOptionFailure(t *testing.T) {
	var opts struct {
		Name string `long:"name" validate-non-empty:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--name", " "})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "must not be empty")
}

func TestValidateRegexRequiresFullMatch(t *testing.T) {
	var opts struct {
		Name string `long:"name" validate-regex:"[a-z]+"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--name", "abc123"})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "does not match pattern `[a-z]+`")
}

func TestValidateStringSlice(t *testing.T) {
	var opts struct {
		Names []string `long:"name" validate-non-empty:"true" validate-min-len:"2"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--name", "ok", "--name", "x"})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "shorter than required length")
}

func TestValidateNumericOptionTags(t *testing.T) {
	var opts struct {
		Ints   []int   `long:"int" validate-min:"2" validate-max:"5"`
		Uint   uint    `long:"uint" validate-min:"1" validate-max:"9"`
		Float  float64 `long:"float" validate-min:"1.5" validate-max:"2.5"`
		Unset  int     `long:"unset" validate-min:"10"`
		Preset int     `long:"preset" validate-min:"10"`
	}
	opts.Preset = 11

	assertParseSuccess(t, &opts, "--int", "2", "--int", "5", "--uint", "3", "--float", "2")
}

func TestValidateNumericOptionFailure(t *testing.T) {
	var opts struct {
		Count int `long:"count" validate-min:"2" validate-max:"5"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--count", "6"})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "greater than allowed maximum")
}

func TestValidateDefaultValue(t *testing.T) {
	var opts struct {
		Name string `long:"name" default:"x" validate-min-len:"2"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs(nil)

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "shorter than required length")
}

func TestValidatePositionalArgument(t *testing.T) {
	var opts struct {
		Args struct {
			Name string `validate-non-empty:"true" validate-min-len:"3"`
		} `positional-args:"true" optional:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"ab"})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "shorter than required length")
}

func TestValidateExistingDirAndWritableFailure(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-dir")
	writeTestFile(t, file, "data")

	var opts struct {
		Dir string `long:"dir" validate-existing-dir:"true"`
		Out string `long:"out" validate-writable:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--dir", file, "--out", filepath.Join(file, "out.txt")})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "must be an existing directory")
}

func TestValidateExistingFileFailure(t *testing.T) {
	var opts struct {
		Input string `long:"input" validate-existing-file:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--input", t.TempDir()})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "must be an existing file")
}

func TestValidateWritableFailure(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-dir")
	writeTestFile(t, file, "data")

	var opts struct {
		Out string `long:"out" validate-writable:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--out", filepath.Join(file, "out.txt")})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "must be writable")
}

func TestValidateReadableFailure(t *testing.T) {
	var opts struct {
		Input string `long:"input" validate-readable:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--input", filepath.Join(t.TempDir(), "missing.txt")})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "must be readable")
}

func TestValidatePathAbsFailure(t *testing.T) {
	var opts struct {
		Path string `long:"path" validate-path-abs:"true"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"--path", "relative.txt"})

	assertErrorTypeAndMessageContains(t, err, ErrValidation, "must be absolute")
}

func TestValidateInvalidTags(t *testing.T) {
	tests := []struct {
		name string
		opts any
	}{
		{
			name: "invalid bool",
			opts: &struct {
				Value string `long:"value" validate-non-empty:"maybe"`
			}{},
		},
		{
			name: "invalid regex",
			opts: &struct {
				Value string `long:"value" validate-regex:"["`
			}{},
		},
		{
			name: "invalid min len",
			opts: &struct {
				Value string `long:"value" validate-min-len:"-1"`
			}{},
		},
		{
			name: "string min",
			opts: &struct {
				Value string `long:"value" validate-min:"1"`
			}{},
		},
		{
			name: "int non empty",
			opts: &struct {
				Value int `long:"value" validate-non-empty:"true"`
			}{},
		},
		{
			name: "negative uint min",
			opts: &struct {
				Value uint `long:"value" validate-min:"-1"`
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.opts, Default&^PrintErrors)
			_, err := parser.ParseArgs(nil)
			assertErrorType(t, err, ErrInvalidTag)
		})
	}
}

func assertErrorType(t *testing.T, err error, typ ErrorType) {
	t.Helper()

	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T: %v", err, err)
	}
	if flagsErr.Type != typ {
		t.Fatalf("expected error type %s, got %s: %s", typ, flagsErr.Type, flagsErr.Message)
	}
}

func assertErrorTypeAndMessageContains(t *testing.T, err error, typ ErrorType, contains string) {
	t.Helper()

	assertErrorType(t, err, typ)
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error %q to contain %q", err.Error(), contains)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
