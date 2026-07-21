// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONKeyLong(t *testing.T) {
	cases := [][2]string{
		{"my-flag", "my-flag"},
		{"verbose", "verbose"},
		{"no-color", "no-color"},
	}
	for _, c := range cases {
		if got := JSONKeyLong(c[0]); got != c[1] {
			t.Errorf("JSONKeyLong(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestJSONKeyCamel(t *testing.T) {
	cases := [][2]string{
		{"my-long-flag", "myLongFlag"},
		{"verbose", "verbose"},
		{"no-color", "noColor"},
		{"a-b-c", "aBC"},
	}
	for _, c := range cases {
		if got := JSONKeyCamel(c[0]); got != c[1] {
			t.Errorf("JSONKeyCamel(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestJSONKeySnake(t *testing.T) {
	cases := [][2]string{
		{"my-long-flag", "my_long_flag"},
		{"verbose", "verbose"},
		{"no-color", "no_color"},
	}
	for _, c := range cases {
		if got := JSONKeySnake(c[0]); got != c[1] {
			t.Errorf("JSONKeySnake(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestJSONKeyPascal(t *testing.T) {
	cases := [][2]string{
		{"my-long-flag", "MyLongFlag"},
		{"verbose", "Verbose"},
		{"no-color", "NoColor"},
	}
	for _, c := range cases {
		if got := JSONKeyPascal(c[0]); got != c[1] {
			t.Errorf("JSONKeyPascal(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestJsonParseBasic(t *testing.T) {
	var opts struct {
		Verbose bool   `long:"verbose"`
		Output  string `long:"output"`
		Count   int    `long:"count"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	err := jp.Parse(strings.NewReader(`{"verbose":true,"output":"file.txt","count":5}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !opts.Verbose {
		t.Error("verbose should be true")
	}
	if opts.Output != "file.txt" {
		t.Errorf("output = %q, want %q", opts.Output, "file.txt")
	}
	if opts.Count != 5 {
		t.Errorf("count = %d, want 5", opts.Count)
	}
}

func TestJsonParseJsonTagOverridesLong(t *testing.T) {
	var opts struct {
		Name string `long:"name" json:"myName"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	err := jp.Parse(strings.NewReader(`{"myName":"hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Name != "hello" {
		t.Errorf("Name = %q, want %q", opts.Name, "hello")
	}
}

func TestJsonParseJsonTagExcludes(t *testing.T) {
	var opts struct {
		Secret string `long:"secret" json:"-"`
		Public string `long:"public"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	err := jp.Parse(strings.NewReader(`{"secret":"should-not-set","public":"ok"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Secret != "" {
		t.Errorf("Secret should remain empty, got %q", opts.Secret)
	}
	if opts.Public != "ok" {
		t.Errorf("Public = %q, want %q", opts.Public, "ok")
	}
}

func TestJsonParseJsonTagEmptyNameUsesLong(t *testing.T) {
	var opts struct {
		Name string `long:"my-name" json:",omitempty"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	err := jp.Parse(strings.NewReader(`{"my-name":"value"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Name != "value" {
		t.Errorf("Name = %q, want %q", opts.Name, "value")
	}
}

func TestJsonParseNestedGroup(t *testing.T) {
	type dbOpts struct {
		Host string `long:"host"`
		Port int    `long:"port"`
	}

	var opts struct {
		Verbose bool   `long:"verbose"`
		DB      dbOpts `group:"Database" namespace:"db"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	input := `{"verbose":true,"db":{"host":"localhost","port":5432}}`
	if err := jp.Parse(strings.NewReader(input)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !opts.Verbose {
		t.Error("verbose should be true")
	}
	if opts.DB.Host != "localhost" {
		t.Errorf("host = %q, want %q", opts.DB.Host, "localhost")
	}
	if opts.DB.Port != 5432 {
		t.Errorf("port = %d, want 5432", opts.DB.Port)
	}
}

func TestJsonParseAnonymousGroup(t *testing.T) {
	// Group with no namespace: options surface at parent level
	type inner struct {
		Flag string `long:"inner-flag"`
	}
	var opts struct {
		Outer string `long:"outer"`
		Inner inner  `group:"Inner"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	if err := jp.Parse(strings.NewReader(`{"outer":"a","inner-flag":"b"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Outer != "a" {
		t.Errorf("outer = %q, want %q", opts.Outer, "a")
	}
	if opts.Inner.Flag != "b" {
		t.Errorf("inner-flag = %q, want %q", opts.Inner.Flag, "b")
	}
}

func TestJsonParseSlice(t *testing.T) {
	var opts struct {
		Tags []string `long:"tag"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	if err := jp.Parse(strings.NewReader(`{"tag":["a","b","c"]}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts.Tags) != 3 || opts.Tags[0] != "a" || opts.Tags[1] != "b" || opts.Tags[2] != "c" {
		t.Errorf("tags = %v, want [a b c]", opts.Tags)
	}
}

func TestJsonParseMap(t *testing.T) {
	var opts struct {
		Labels map[string]string `long:"label"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	if err := jp.Parse(strings.NewReader(`{"label":{"env":"prod","team":"platform"}}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Labels["env"] != "prod" {
		t.Errorf("label[env] = %q, want %q", opts.Labels["env"], "prod")
	}
	if opts.Labels["team"] != "platform" {
		t.Errorf("label[team] = %q, want %q", opts.Labels["team"], "platform")
	}
}

func TestJsonParseAsDefaults(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)
	jp.ParseAsDefaults = true

	// Apply JSON as defaults first
	if err := jp.Parse(strings.NewReader(`{"value":"from-json"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then parse CLI args - CLI should win
	if _, err := p.ParseArgs([]string{"--value=from-cli"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Value != "from-cli" {
		t.Errorf("value = %q, want %q (CLI should override JSON default)", opts.Value, "from-cli")
	}
}

func TestJsonParseAsDefaultsNoCLI(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)
	jp.ParseAsDefaults = true

	if err := jp.Parse(strings.NewReader(`{"value":"from-json"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := p.ParseArgs([]string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Value != "from-json" {
		t.Errorf("value = %q, want %q (JSON default should apply when no CLI)", opts.Value, "from-json")
	}
}

func TestJsonParseMap_ParseMapMethod(t *testing.T) {
	var opts struct {
		Host string `long:"host"`
		Port int    `long:"port"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	// Simulate what a YAML/TOML decoder would produce
	data := map[string]interface{}{
		"host": "db.example.com",
		"port": float64(3306), // JSON numbers are float64
	}

	if err := jp.ParseMap(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Host != "db.example.com" {
		t.Errorf("host = %q, want %q", opts.Host, "db.example.com")
	}
	if opts.Port != 3306 {
		t.Errorf("port = %d, want 3306", opts.Port)
	}
}

// TestJsonParseMapRollsBackOnLaterFieldError ensures
// a config file that fails partway through does not leave earlier,
// successfully-converted fields applied to the bound struct:
// the whole ParseMap call is transactional.
func TestJsonParseMapRollsBackOnLaterFieldError(t *testing.T) {
	var opts struct {
		First  string `long:"first"`
		Second int    `long:"second"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	err := jp.ParseMap(map[string]any{
		"first":  "applied",
		"second": "not-a-number",
	})
	if err == nil {
		t.Fatal("expected error converting the second field")
	}

	if opts.First != "" {
		t.Errorf("expected First to roll back to its zero value, got %q", opts.First)
	}
	if opts.Second != 0 {
		t.Errorf("expected Second to remain at its zero value, got %d", opts.Second)
	}
}

// TestJsonParseMapRollsBackSliceOnLaterFieldError covers
// a slice-typed option that received some elements before a later, unrelated field fails.
func TestJsonParseMapRollsBackSliceOnLaterFieldError(t *testing.T) {
	var opts struct {
		Tags   []string `long:"tags"`
		Second int      `long:"second"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	err := jp.ParseMap(map[string]any{
		"tags":   []any{"a", "b", "c"},
		"second": "not-a-number",
	})
	if err == nil {
		t.Fatal("expected error converting the second field")
	}

	if len(opts.Tags) != 0 {
		t.Errorf("expected Tags to roll back to empty, got %v", opts.Tags)
	}
}

func TestJsonParseCustomKeyName(t *testing.T) {
	var opts struct {
		MyFlag string `long:"my-flag"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)
	jp.KeyName = JSONKeyCamel

	if err := jp.Parse(strings.NewReader(`{"myFlag":"hello"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.MyFlag != "hello" {
		t.Errorf("MyFlag = %q, want %q", opts.MyFlag, "hello")
	}
}

func TestJsonParseUnknownKeysIgnored(t *testing.T) {
	var opts struct {
		Name string `long:"name"`
	}

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	// Extra keys in JSON should be silently ignored
	err := jp.Parse(strings.NewReader(`{"name":"test","unknown-key":"ignored","another":42}`))
	if err != nil {
		t.Fatalf("unexpected error for unknown JSON keys: %v", err)
	}

	if opts.Name != "test" {
		t.Errorf("name = %q, want %q", opts.Name, "test")
	}
}

func TestJsonWriteBasic(t *testing.T) {
	var opts struct {
		Verbose bool   `long:"verbose"`
		Output  string `long:"output"`
	}
	opts.Verbose = true
	opts.Output = "out.txt"

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	var buf bytes.Buffer
	if err := jp.Write(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}

	if got["verbose"] != true {
		t.Errorf("verbose = %v, want true", got["verbose"])
	}
	if got["output"] != "out.txt" {
		t.Errorf("output = %v, want %q", got["output"], "out.txt")
	}
}

func TestJsonWriteFileRoundTrip(t *testing.T) {
	var opts struct {
		Verbose bool   `long:"verbose"`
		Output  string `long:"output"`
	}
	opts.Verbose = true
	opts.Output = "out.txt"

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := jp.WriteFile(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, raw)
	}

	if got["verbose"] != true {
		t.Errorf("verbose = %v, want true", got["verbose"])
	}
	if got["output"] != "out.txt" {
		t.Errorf("output = %v, want %q", got["output"], "out.txt")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("unexpected readdir error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one file in %s after WriteFile, got %d: %v", dir, len(entries), entries)
	}
}

func TestJsonWriteNestedGroup(t *testing.T) {
	type serverOpts struct {
		Host string `long:"host"`
		Port int    `long:"port"`
	}
	var opts struct {
		Server serverOpts `group:"Server" namespace:"server"`
	}
	opts.Server.Host = "localhost"
	opts.Server.Port = 8080

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	var buf bytes.Buffer
	if err := jp.Write(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	srv, ok := got["server"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected server to be an object, got %T", got["server"])
	}
	if srv["host"] != "localhost" {
		t.Errorf("server.host = %v, want %q", srv["host"], "localhost")
	}
}

func TestJsonWriteOmitempty(t *testing.T) {
	var opts struct {
		Name  string `long:"name" json:",omitempty"`
		Value string `long:"value"`
	}
	// Name is zero value, Value is set
	opts.Value = "set"

	p := NewParser(&opts, None)
	jp := NewJSONParser(p)

	var buf bytes.Buffer
	if err := jp.Write(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if _, exists := got["name"]; exists {
		t.Error("name with omitempty and zero value should be omitted")
	}
	if got["value"] != "set" {
		t.Errorf("value = %v, want %q", got["value"], "set")
	}
}

func TestJsonRoundTrip(t *testing.T) {
	type dbOpts struct {
		Host string `long:"host"`
		Port int    `long:"port"`
	}
	var orig struct {
		Verbose bool   `long:"verbose"`
		Output  string `long:"output"`
		DB      dbOpts `group:"DB" namespace:"db"`
	}
	orig.Verbose = true
	orig.Output = "result.txt"
	orig.DB.Host = "pg.local"
	orig.DB.Port = 5432

	p1 := NewParser(&orig, None)
	jp1 := NewJSONParser(p1)

	var buf bytes.Buffer
	if err := jp1.Write(&buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var restored struct {
		Verbose bool   `long:"verbose"`
		Output  string `long:"output"`
		DB      dbOpts `group:"DB" namespace:"db"`
	}

	p2 := NewParser(&restored, None)
	jp2 := NewJSONParser(p2)

	if err := jp2.Parse(strings.NewReader(buf.String())); err != nil {
		t.Fatalf("parse error: %v\nJSON:\n%s", err, buf.String())
	}

	if restored.Verbose != orig.Verbose {
		t.Errorf("Verbose: got %v, want %v", restored.Verbose, orig.Verbose)
	}
	if restored.Output != orig.Output {
		t.Errorf("Output: got %q, want %q", restored.Output, orig.Output)
	}
	if restored.DB.Host != orig.DB.Host {
		t.Errorf("DB.Host: got %q, want %q", restored.DB.Host, orig.DB.Host)
	}
	if restored.DB.Port != orig.DB.Port {
		t.Errorf("DB.Port: got %d, want %d", restored.DB.Port, orig.DB.Port)
	}
}

func TestJsonParseFileNotFound(t *testing.T) {
	var opts struct{ Value string }
	p := NewParser(&opts, None)
	err := NewJSONParser(p).ParseFile("this-file-should-not-exist-json.json")
	if err == nil {
		t.Fatal("expected error for missing JSON file")
	}
}
