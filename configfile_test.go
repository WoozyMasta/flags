// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"os"
	"path/filepath"
	"testing"
)

// ---- detectConfigFileFormat -------------------------------------------------

func TestDetectConfigFileFormatByExtension(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name       string
		filename   string
		content    string
		iniEnabled bool
		jsonEnable bool
		want       string
	}{
		{"json ext", "cfg.json", `{"a":1}`, false, true, "json"},
		{"ini ext", "cfg.ini", "[section]\nkey=val", true, false, "ini"},
		{"conf ext", "cfg.conf", "key=val", true, false, "ini"},
		{"json wins by ext", "cfg.json", `{"a":1}`, true, true, "json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.filename)
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatal(err)
			}

			got, err := detectConfigFileFormat(path, tc.iniEnabled, tc.jsonEnable)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectConfigFileFormatByMagicByte(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name       string
		content    string
		iniEnabled bool
		jsonEnable bool
		want       string
	}{
		{"json magic byte, json enabled", `{"a":1}`, false, true, "json"},
		{"ini fallback, ini enabled", "key=val", true, false, "ini"},
		{"json magic byte, both enabled", `{"a":1}`, true, true, "json"},
		{"non-json content, both enabled", "key=val", true, true, "ini"},
		{"json with leading whitespace", "  \n\t{\"a\":1}", true, true, "json"},
		{"json with utf-8 BOM", "\xEF\xBB\xBF{\"a\":1}", true, true, "json"},
		{"json with utf-8 BOM and whitespace", "\xEF\xBB\xBF \n{\"a\":1}", true, true, "json"},
		{"ini with leading whitespace", "  \nkey=val", true, true, "ini"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name+".cfg")
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatal(err)
			}

			got, err := detectConfigFileFormat(path, tc.iniEnabled, tc.jsonEnable)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectConfigFileFormatErrors(t *testing.T) {
	dir := t.TempDir()

	t.Run("json ext but json disabled", func(t *testing.T) {
		path := filepath.Join(dir, "cfg.json")
		if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := detectConfigFileFormat(path, true, false)
		if err == nil {
			t.Fatal("expected error for JSON extension with JSON disabled")
		}
	})

	t.Run("ini ext but ini disabled", func(t *testing.T) {
		path := filepath.Join(dir, "cfg.ini")
		if err := os.WriteFile(path, []byte("key=val"), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := detectConfigFileFormat(path, false, true)
		if err == nil {
			t.Fatal("expected error for INI extension with INI disabled")
		}
	})

	t.Run("json magic byte but json disabled", func(t *testing.T) {
		path := filepath.Join(dir, "cfg.cfg")
		if err := os.WriteFile(path, []byte(`{"a":1}`), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := detectConfigFileFormat(path, true, false)
		if err == nil {
			t.Fatal("expected error for JSON content with JSON disabled")
		}
	})

	t.Run("no format enabled", func(t *testing.T) {
		path := filepath.Join(dir, "cfg2.cfg")
		if err := os.WriteFile(path, []byte("key=val"), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := detectConfigFileFormat(path, false, false)
		if err == nil {
			t.Fatal("expected error when no format is enabled")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(dir, "empty.cfg")
		if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := detectConfigFileFormat(path, true, true)
		if err == nil {
			t.Fatal("expected error for an empty config file")
		}
	})

	t.Run("whitespace-only file", func(t *testing.T) {
		path := filepath.Join(dir, "whitespace.cfg")
		if err := os.WriteFile(path, []byte("  \n\t\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := detectConfigFileFormat(path, true, true)
		if err == nil {
			t.Fatal("expected error for a whitespace-only config file")
		}
	})
}

// ---- bootstrapScanArgs (config) ---------------------------------------------

func TestBootstrapScanArgsConfigValue(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"long equals", []string{"--config=app.ini"}, "app.ini"},
		{"long space", []string{"--config", "app.ini"}, "app.ini"},
		{"short equals", []string{"-c=app.ini"}, "app.ini"},
		{"short space", []string{"-c", "app.ini"}, "app.ini"},
		{"absent", []string{"--verbose"}, ""},
		{"long next is flag", []string{"--config", "--verbose"}, ""},
		// A bare dash-prefixed value looks like an option to the real
		// tokenizer too (argumentIsOption), so it is rejected the same way
		// the main parser would reject it for any other value-carrying
		// option; the `=` form is the documented way around that.
		{"dash-prefixed value needs equals form", []string{"--config", "-prod.json"}, ""},
		{"dash-prefixed value via equals form", []string{"--config=-prod.json"}, "-prod.json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var data struct {
				Config string `long:"config" short:"c"`
			}
			p := NewParser(&data, None)
			opt := p.FindOptionByLongName("config")

			var got string
			p.bootstrapScanArgs(tc.args, map[*Option]*string{opt: &got}, nil)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBootstrapScanArgsStopsAtDoubleDashWhenEnabled(t *testing.T) {
	var data struct {
		Config string `long:"config"`
	}
	p := NewParser(&data, PassDoubleDash)
	opt := p.FindOptionByLongName("config")

	var got string
	p.bootstrapScanArgs([]string{"--", "--config=evil.json"}, map[*Option]*string{opt: &got}, nil)
	if got != "" {
		t.Errorf("expected scan to stop at a literal --, got %q", got)
	}
}

func TestBootstrapScanArgsDoesNotStopAtDoubleDashWhenDisabled(t *testing.T) {
	var data struct {
		Config string `long:"config"`
	}
	p := NewParser(&data, None)
	opt := p.FindOptionByLongName("config")

	var got string
	p.bootstrapScanArgs([]string{"--", "--config=app.ini"}, map[*Option]*string{opt: &got}, nil)
	if got != "app.ini" {
		t.Errorf("expected scan to keep matching past -- when PassDoubleDash is not set, got %q", got)
	}
}

func TestBootstrapScanArgsRecognizesLongAliases(t *testing.T) {
	var data struct {
		Config string `long:"config"`
	}
	p := NewParser(&data, None)
	opt := p.FindOptionByLongName("config")
	if err := opt.SetLongAliases("cfg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got string
	p.bootstrapScanArgs([]string{"--cfg=app.ini"}, map[*Option]*string{opt: &got}, nil)
	if got != "app.ini" {
		t.Errorf("expected long alias to be recognized, got %q", got)
	}
}

// ---- SetConfigFile + ConfigFlags integration --------------------------------

func TestSetConfigFileAppliedAsDefault(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "app.ini")

	if err := os.WriteFile(iniPath, []byte("[Application Options]\nverbose = true\n"), 0o600); err != nil {
		t.Fatal(iniPath)
	}

	var opts struct {
		Verbose bool `long:"verbose"`
	}

	p := NewParser(&opts, ConfigIni|ConfigFlags)
	p.SubcommandsOptional = true
	p.SetConfigFile(iniPath)

	if _, err := p.ParseArgs([]string{}); err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}

	if !opts.Verbose {
		t.Error("expected verbose=true from default config file")
	}
}

func TestConfigFlagExplicitPathOverridesDefault(t *testing.T) {
	dir := t.TempDir()

	defaultIni := filepath.Join(dir, "default.ini")
	if err := os.WriteFile(defaultIni, []byte("[Application Options]\nverbose = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	explicitIni := filepath.Join(dir, "explicit.ini")
	if err := os.WriteFile(explicitIni, []byte("[Application Options]\nverbose = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var opts struct {
		Verbose bool `long:"verbose"`
	}

	p := NewParser(&opts, ConfigIni|ConfigFlags)
	p.SubcommandsOptional = true
	p.SetConfigFile(defaultIni)

	if _, err := p.ParseArgs([]string{"--config", explicitIni}); err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}

	if !opts.Verbose {
		t.Error("expected verbose=true from explicit --config path")
	}
}

func TestConfigFlagIgnoredAfterDoubleDashWhenPassDoubleDashSet(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "evil.ini")
	if err := os.WriteFile(iniPath, []byte("[Application Options]\nverbose = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var opts struct {
		Verbose bool `long:"verbose"`
	}

	p := NewParser(&opts, ConfigIni|ConfigFlags|PassDoubleDash)
	p.SubcommandsOptional = true

	retargs, err := p.ParseArgs([]string{"--", "--config=" + iniPath})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}

	if opts.Verbose {
		t.Error("expected --config after a literal -- to be treated as a literal argument, not loaded")
	}
	if len(retargs) != 1 || retargs[0] != "--config="+iniPath {
		t.Errorf("expected the literal token to pass through as an argument, got %v", retargs)
	}
}

func TestConfigFlagWindowsStyleSlash(t *testing.T) {
	if defaultLongOptDelimiter != "/" {
		t.Skip("Windows-style option prefix is not active in this build")
	}

	dir := t.TempDir()
	iniPath := filepath.Join(dir, "app.ini")
	if err := os.WriteFile(iniPath, []byte("[Application Options]\nverbose = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var opts struct {
		Verbose bool `long:"verbose"`
	}

	p := NewParser(&opts, ConfigIni|ConfigFlags)
	p.SubcommandsOptional = true

	if _, err := p.ParseArgs([]string{"/config:" + iniPath}); err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}

	if !opts.Verbose {
		t.Error("expected verbose=true from Windows-style /config:path")
	}
}

func TestConfigFlagWithJSONFile(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "app.json")

	if err := os.WriteFile(jsonPath, []byte(`{"verbose":true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var opts struct {
		Verbose bool `long:"verbose"`
	}

	p := NewParser(&opts, ConfigJSON|ConfigFlags)
	p.SubcommandsOptional = true

	if _, err := p.ParseArgs([]string{"--config", jsonPath}); err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}

	if !opts.Verbose {
		t.Error("expected verbose=true from JSON config")
	}
}

func TestConfigFlagNoFileNoError(t *testing.T) {
	var opts struct {
		Verbose bool `long:"verbose"`
	}

	p := NewParser(&opts, ConfigIni|ConfigFlags)
	p.SubcommandsOptional = true

	if _, err := p.ParseArgs([]string{}); err != nil {
		t.Fatalf("ParseArgs should succeed when no config provided: %v", err)
	}
}
