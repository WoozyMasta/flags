// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"errors"
	"os"
	"strings"
)

// SetDotEnvFile sets the .env file path used when the DotEnv, DotEnvOverride,
// or DotEnvFlags option is active. The default is ".env".
func (p *Parser) SetDotEnvFile(filename string) {
	p.dotEnvFile = filename
}

// SetDotEnvNoExpand disables variable expansion in .env values when disabled
// is true. By default expansion of $VAR, ${VAR}, ${VAR:=default} and $${VAR}
// is enabled.
func (p *Parser) SetDotEnvNoExpand(disabled bool) {
	p.dotEnvNoExpand = disabled
}

// LoadDotEnv reads one or more .env files and sets environment variables for
// the current process. Existing environment variables are NOT overridden.
//
// When called with no arguments the file configured by SetDotEnvFile (default
// ".env") is used. A missing default file is silently ignored; a missing
// explicitly named file returns an error.
func (p *Parser) LoadDotEnv(filenames ...string) error {
	return p.loadDotEnvFiles(false, filenames...)
}

// OverloadDotEnv is like LoadDotEnv but overrides existing environment
// variables with values from the .env files.
func (p *Parser) OverloadDotEnv(filenames ...string) error {
	return p.loadDotEnvFiles(true, filenames...)
}

// loadDotEnvFiles is the shared implementation for LoadDotEnv and OverloadDotEnv.
func (p *Parser) loadDotEnvFiles(override bool, filenames ...string) error {
	explicit := len(filenames) > 0

	if !explicit {
		name := p.dotEnvFile
		if name == "" {
			name = ".env"
		}

		filenames = []string{name}
	}

	for _, name := range filenames {
		if err := p.loadOneDotEnvFile(name, override, !explicit); err != nil {
			return err
		}
	}

	return nil
}

// loadOneDotEnvFile reads a single .env file and applies its values.
// When silentMissing is true a missing file is not an error.
func (p *Parser) loadOneDotEnvFile(filename string, override, silentMissing bool) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		if silentMissing && errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return p.i18nTextfErr(
			"err.dotenv.load",
			"failed to load .env file `{file}`: {error}",
			map[string]string{"file": filename, "error": err.Error()},
		)
	}

	vars, err := parseDotEnvBytes(src, p.dotEnvNoExpand)
	if err != nil {
		return p.i18nTextfErr(
			"err.dotenv.parse",
			"failed to parse .env file `{file}`: {error}",
			map[string]string{"file": filename, "error": err.Error()},
		)
	}

	for k, v := range vars {
		if override {
			if err := os.Setenv(k, v); err != nil {
				return err
			}
		} else if _, exists := os.LookupEnv(k); !exists {
			if err := os.Setenv(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// i18nTextfErr is a convenience helper that creates a formatted error using
// the i18n system.
func (p *Parser) i18nTextfErr(key, fallback string, data map[string]string) error {
	return errors.New(p.i18nTextf(key, fallback, data))
}

// EnsureBuiltinDotEnvOptions materialises the "Env Options" group with the
// three pre-configured flags when DotEnvFlags is set and the group has not
// yet been added. It is called automatically from ParseArgs and can also be
// called by application code to access the group before parsing.
func (p *Parser) EnsureBuiltinDotEnvOptions() {
	if (p.Options&DotEnvFlags) == None || p.hasDotEnvGroup {
		return
	}

	p.addDotEnvGroup()
}

// addDotEnvGroup adds the "Env Options" group containing the three dotenv
// flags and records their long names for the pre-scan step.
func (p *Parser) addDotEnvGroup() {
	fileFlagLong := "env-file"
	noEnvFlagLong := "no-env"
	overrideFlagLong := "env-override"

	type envOpts struct {
		File     string `long:"env-file" value-name:"FILE" description:"Path to .env file" description-i18n:"help.builtin.dotenv.file.desc" value-name-i18n:"help.builtin.dotenv.file.name" auto-env:"false"`
		NoEnv    bool   `long:"no-env" description:"Disable .env file loading" description-i18n:"help.builtin.dotenv.no_env.desc" auto-env:"false"`
		Override bool   `long:"env-override" description:"Override existing environment variables from .env file" description-i18n:"help.builtin.dotenv.override.desc" auto-env:"false"`
	}

	data := &envOpts{}
	grp, err := p.AddGroup("Env Options", "", data)
	if err != nil {
		return
	}

	grp.SetShortDescriptionI18nKey("help.group.env_options")

	p.dotEnvFileFlagName = fileFlagLong
	p.dotEnvDisableFlagName = noEnvFlagLong
	p.dotEnvOverrideFlagName = overrideFlagLong
	p.hasDotEnvGroup = true
}

// BuiltinDotEnvFileOption returns the --env-file option when DotEnvFlags is
// enabled. It materialises the group lazily and returns nil when unavailable.
func (p *Parser) BuiltinDotEnvFileOption() *Option {
	p.EnsureBuiltinDotEnvOptions()
	return p.FindOptionByLongName(p.dotEnvFileFlagName)
}

// BuiltinDotEnvNoEnvOption returns the --no-env option when DotEnvFlags is
// enabled. It materialises the group lazily and returns nil when unavailable.
func (p *Parser) BuiltinDotEnvNoEnvOption() *Option {
	p.EnsureBuiltinDotEnvOptions()
	return p.FindOptionByLongName(p.dotEnvDisableFlagName)
}

// BuiltinDotEnvOverrideOption returns the --env-override option when
// DotEnvFlags is enabled. Returns nil when unavailable.
func (p *Parser) BuiltinDotEnvOverrideOption() *Option {
	p.EnsureBuiltinDotEnvOptions()
	return p.FindOptionByLongName(p.dotEnvOverrideFlagName)
}

// applyDotEnv is called from ParseArgs when any of the DotEnv-family options
// are active. It pre-scans args for the registered dotenv flags and then
// loads the .env file.
func (p *Parser) applyDotEnv(args []string) error {
	override := (p.Options & DotEnvOverride) != None
	disabled := false

	// Pre-scan args for the three dotenv flags (if the group was added).
	if p.hasDotEnvGroup {
		file, noEnv, envOverride := dotEnvPreScanArgs(
			args,
			p.dotEnvFileFlagName,
			p.dotEnvDisableFlagName,
			p.dotEnvOverrideFlagName,
		)

		if file != "" {
			p.dotEnvFile = file
		}

		if noEnv {
			disabled = true
		}

		if envOverride {
			override = true
		}
	}

	if disabled {
		return nil
	}

	return p.loadDotEnvFiles(override)
}

// dotEnvPreScanArgs performs a minimal scan of args to extract dotenv flag
// values before the main parse runs. Only long-name forms are recognised:
//
//	--file=VALUE  or  --file VALUE
//	--bool-flag
func dotEnvPreScanArgs(args []string, fileLong, noEnvLong, overrideLong string) (file string, noEnv, override bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "--") {
			continue
		}

		arg = arg[2:]

		// --name=VALUE form
		if before, after, ok := strings.Cut(arg, "="); ok {
			name := before
			val := after

			if name == fileLong {
				file = val
			}

			continue
		}

		// --name VALUE or bare --name forms
		switch arg {
		case fileLong:
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				file = args[i+1]
				i++
			}
		case noEnvLong:
			noEnv = true
		case overrideLong:
			override = true
		}
	}

	return file, noEnv, override
}
