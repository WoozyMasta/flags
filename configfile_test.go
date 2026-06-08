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
}

// ---- configPreScanArgs ------------------------------------------------------

func TestConfigPreScanArgs(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		longName string
		want     string
	}{
		{"long equals", []string{"--config=app.ini"}, "config", "app.ini"},
		{"long space", []string{"--config", "app.ini"}, "config", "app.ini"},
		{"short equals", []string{"-c=app.ini"}, "config", "app.ini"},
		{"short space", []string{"-c", "app.ini"}, "config", "app.ini"},
		{"absent", []string{"--verbose"}, "config", ""},
		{"long next is flag", []string{"--config", "--verbose"}, "config", ""},
		{"custom long name", []string{"--cfg=x.json"}, "cfg", "x.json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := configPreScanArgs(tc.args, tc.longName)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
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
