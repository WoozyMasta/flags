// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"errors"
	"os"
)

// SetDotEnvFile sets the .env file path used when the DotEnv, DotEnvOverride, or DotEnvFlags option is active.
// The default is ".env".
func (p *Parser) SetDotEnvFile(filename string) {
	p.dotEnvFile = filename
}

// SetDotEnvNoExpand disables variable expansion in .env values when disabled is true.
// By default expansion of $VAR, ${VAR}, ${VAR:=default} and $${VAR} is enabled.
func (p *Parser) SetDotEnvNoExpand(disabled bool) {
	p.dotEnvNoExpand = disabled
}

// LoadDotEnv reads one or more .env files and sets environment variables for the current process.
// Existing environment variables are NOT overridden.
//
// When called with no arguments the file configured by SetDotEnvFile (default ".env") is used.
// A missing default file is silently ignored; a missing explicitly named file returns an error.
//
// This mutates the process environment (os.Setenv),
// which is global state shared with any other code running in the same process
// see the package doc's "Concurrency" section.
func (p *Parser) LoadDotEnv(filenames ...string) error {
	return p.loadDotEnvFiles(false, filenames...)
}

// OverloadDotEnv is like LoadDotEnv but overrides existing environment variables with values from the .env files.
// See LoadDotEnv's doc comment for the process-environment concurrency caveat.
func (p *Parser) OverloadDotEnv(filenames ...string) error {
	return p.loadDotEnvFiles(true, filenames...)
}

// loadDotEnvFiles is the shared implementation for LoadDotEnv and OverloadDotEnv.
// All files are read and parsed first; process environment variables
// are only set once every file has parsed successfully,
// so a later file's error never leaves an earlier file's variables applied.
func (p *Parser) loadDotEnvFiles(override bool, filenames ...string) error {
	explicit := len(filenames) > 0

	if !explicit {
		name := p.dotEnvFile
		if name == "" {
			name = ".env"
		}

		filenames = []string{name}
	}

	merged := make(map[string]string)

	for _, name := range filenames {
		vars, err := p.readOneDotEnvFile(name, !explicit)
		if err != nil {
			return err
		}

		for k, v := range vars {
			if override {
				merged[k] = v // Later files win when overriding.
			} else if _, exists := merged[k]; !exists {
				merged[k] = v // Earlier files win otherwise.
			}
		}
	}

	for k, v := range merged {
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

// readOneDotEnvFile reads and parses a single .env file without applying any values to the process environment.
// When silentMissing is true a missing file returns (nil, nil) instead of an error.
func (p *Parser) readOneDotEnvFile(filename string, silentMissing bool) (map[string]string, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		if silentMissing && errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, p.i18nTextfErr(
			"err.dotenv.load",
			"failed to load .env file `{file}`: {error}",
			map[string]string{"file": filename, "error": err.Error()},
		)
	}

	vars, err := parseDotEnvBytes(src, p.dotEnvNoExpand)
	if err != nil {
		return nil, p.i18nTextfErr(
			"err.dotenv.parse",
			"failed to parse .env file `{file}`: {error}",
			map[string]string{"file": filename, "error": err.Error()},
		)
	}

	return vars, nil
}

// i18nTextfErr is a convenience helper that creates a formatted error using the i18n system.
func (p *Parser) i18nTextfErr(key, fallback string, data map[string]string) error {
	return errors.New(p.i18nTextf(key, fallback, data))
}

// EnsureBuiltinDotEnvOptions materialises the "Env Options" group with the three pre-configured flags
// when DotEnvFlags is set and the group has not yet been added.
// It is called automatically from ParseArgs
// and can also be called by application code to access the group before parsing.
func (p *Parser) EnsureBuiltinDotEnvOptions() error {
	if (p.Options&DotEnvFlags) == None || p.hasDotEnvGroup {
		return nil
	}

	return p.addDotEnvGroup()
}

// addDotEnvGroup adds the "Env Options" group containing the three dotenv flags
// and records their long names for the pre-scan step.
func (p *Parser) addDotEnvGroup() error {
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
		return err
	}

	grp.SetShortDescriptionI18nKey("help.group.env_options")

	dotEnvDefault := p.dotEnvFile
	if dotEnvDefault == "" {
		dotEnvDefault = ".env"
	}

	for _, o := range grp.options {
		if o.LongName == fileFlagLong {
			o.Default = []string{dotEnvDefault}

			break
		}
	}

	p.dotEnvFileFlagName = fileFlagLong
	p.dotEnvDisableFlagName = noEnvFlagLong
	p.dotEnvOverrideFlagName = overrideFlagLong
	p.hasDotEnvGroup = true

	return nil
}

// BuiltinDotEnvFileOption returns the --env-file option when DotEnvFlags is enabled.
// It materialises the group lazily and returns nil when unavailable
// (including when materialising the group fails; call EnsureBuiltinDotEnvOptions
// directly to observe that error).
func (p *Parser) BuiltinDotEnvFileOption() *Option {
	_ = p.EnsureBuiltinDotEnvOptions()
	return p.FindOptionByLongName(p.dotEnvFileFlagName)
}

// BuiltinDotEnvNoEnvOption returns the --no-env option when DotEnvFlags is enabled.
// It materialises the group lazily and returns nil when unavailable
// (including when materialising the group fails; call EnsureBuiltinDotEnvOptions
// directly to observe that error).
func (p *Parser) BuiltinDotEnvNoEnvOption() *Option {
	_ = p.EnsureBuiltinDotEnvOptions()
	return p.FindOptionByLongName(p.dotEnvDisableFlagName)
}

// BuiltinDotEnvOverrideOption returns the --env-override option when
// DotEnvFlags is enabled. Returns nil when unavailable (including when
// materialising the group fails; call EnsureBuiltinDotEnvOptions directly
// to observe that error).
func (p *Parser) BuiltinDotEnvOverrideOption() *Option {
	_ = p.EnsureBuiltinDotEnvOptions()
	return p.FindOptionByLongName(p.dotEnvOverrideFlagName)
}

// applyDotEnv is called from ParseArgs when any of the DotEnv-family options are active.
// It pre-scans args for the registered dotenv flags (and any aliases registered on them),
// using the same tokenizing rules as the main parser, and then loads the .env file.
func (p *Parser) applyDotEnv(args []string) error {
	override := (p.Options & DotEnvOverride) != None
	disabled := false

	if p.hasDotEnvGroup {
		var file string
		var noEnv, envOverride bool

		valueTargets := map[*Option]*string{}
		if opt := p.FindOptionByLongName(p.dotEnvFileFlagName); opt != nil {
			valueTargets[opt] = &file
		}

		boolTargets := map[*Option]*bool{}
		if opt := p.FindOptionByLongName(p.dotEnvDisableFlagName); opt != nil {
			boolTargets[opt] = &noEnv
		}
		if opt := p.FindOptionByLongName(p.dotEnvOverrideFlagName); opt != nil {
			boolTargets[opt] = &envOverride
		}

		p.bootstrapScanArgs(args, valueTargets, boolTargets)

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
