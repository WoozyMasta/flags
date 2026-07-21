// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SetConfigFile sets the default config file path used when ConfigFlags is
// active and --config is not passed on the command line. When unset and
// --config is absent, no config file is loaded automatically.
func (p *Parser) SetConfigFile(filename string) {
	p.configFile = filename
}

// EnsureBuiltinConfigOptions materialises the "Config Options" group when
// ConfigFlags is set. It is called automatically from ParseArgs and can also
// be called by application code to access the group before parsing.
func (p *Parser) EnsureBuiltinConfigOptions() error {
	if (p.Options&ConfigFlags) == None || p.hasConfigGroup {
		return nil
	}

	return p.addConfigGroup((p.Options & DotEnvFlags) != None)
}

// addConfigGroup creates the unified "Config Options" group containing the
// --config flag (when ConfigFlags set) and env flags (when DotEnvFlags set).
func (p *Parser) addConfigGroup(withEnv bool) error {
	type configGroupAllOpts struct {
		ConfigFile string `auto-env:"false" long:"config"       description-i18n:"help.builtin.config.file.desc"     description:"Path to configuration file" value-name:"FILE" value-name-i18n:"help.builtin.command.output.name" short:"c"`
		EnvFile    string `auto-env:"false" long:"env-file"     description-i18n:"help.builtin.dotenv.file.desc"     description:"Path to .env file" value-name:"FILE" value-name-i18n:"help.builtin.dotenv.file.name"`
		NoEnv      bool   `auto-env:"false" long:"no-env"       description-i18n:"help.builtin.dotenv.no_env.desc"   description:"Disable .env file loading"`
		Override   bool   `auto-env:"false" long:"env-override" description-i18n:"help.builtin.dotenv.override.desc" description:"Override existing environment variables"`
	}

	grp, err := p.AddGroup("Config Options", "", &configGroupAllOpts{})
	if err != nil {
		return err
	}

	grp.SetShortDescriptionI18nKey("help.group.config_options")

	dotEnvDefault := p.dotEnvFile
	if dotEnvDefault == "" {
		dotEnvDefault = ".env"
	}

	for _, o := range grp.options {
		switch o.LongName {
		case "env-file":
			o.Hidden = !withEnv
			if withEnv {
				o.Default = []string{dotEnvDefault}
			}
		case "no-env", "env-override":
			o.Hidden = !withEnv
		}
	}

	p.configFileFlagName = "config"
	p.hasConfigGroup = true

	if withEnv {
		p.dotEnvFileFlagName = "env-file"
		p.dotEnvDisableFlagName = "no-env"
		p.dotEnvOverrideFlagName = "env-override"
		p.hasDotEnvGroup = true
	}

	return nil
}

// applyConfigFile pre-scans args for -c/--config (and any aliases registered on that option)
// using the same tokenizing rules as the main parser, and loads the file.
// Falls back to p.configFile when --config is absent.
// Nothing is loaded when neither source provides a path.
func (p *Parser) applyConfigFile(args []string) error {
	path := p.configFile

	if opt := p.FindOptionByLongName(p.configFileFlagName); opt != nil {
		var value string
		p.bootstrapScanArgs(args, map[*Option]*string{opt: &value}, nil)

		if value != "" {
			path = value
		}
	}

	if path == "" {
		return nil
	}

	return p.loadConfigFile(path)
}

// loadConfigFile detects the format and applies the file as defaults.
func (p *Parser) loadConfigFile(path string) error {
	iniEnabled := (p.Options & ConfigIni) != None
	jsonEnabled := (p.Options & ConfigJSON) != None

	format, err := detectConfigFileFormat(path, iniEnabled, jsonEnabled)
	if err != nil {
		return err
	}

	switch format {
	case "ini":
		ip := NewIniParser(p)
		ip.ParseAsDefaults = true

		return ip.ParseFile(path)

	case "json":
		jp := NewJSONParser(p)
		jp.ParseAsDefaults = true

		if p.jsonKeyName != nil {
			jp.KeyName = p.jsonKeyName
		}

		return jp.ParseFile(path)
	}

	return nil
}

// detectConfigFileFormat returns "ini" or "json" based on file extension
// or first byte, respecting which formats are enabled.
func detectConfigFileFormat(path string, iniEnabled, jsonEnabled bool) (string, error) {
	lower := strings.ToLower(path)

	// Extension-based detection
	switch {
	case strings.HasSuffix(lower, ".json"):
		if !jsonEnabled {
			return "", fmt.Errorf("config file %q is JSON but JSON config is not enabled (add ConfigJSON to parser options)", path)
		}

		return "json", nil

	case strings.HasSuffix(lower, ".ini"), strings.HasSuffix(lower, ".conf"):
		if !iniEnabled {
			return "", fmt.Errorf("config file %q is INI but INI config is not enabled (add ConfigIni to parser options)", path)
		}

		return "ini", nil
	}

	// First-byte detection when extension is absent or unknown
	f, err := os.Open(path) //nolint:gosec // path comes from the user-supplied --config flag.
	if err != nil {
		return "", err
	}

	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)

	b, err := r.ReadByte()
	if err != nil {
		return "", fmt.Errorf("config file %q: %w", path, err)
	}

	if b == '{' {
		if !jsonEnabled {
			return "", fmt.Errorf("config file %q looks like JSON but JSON config is not enabled", path)
		}

		return "json", nil
	}

	if iniEnabled {
		return "ini", nil
	}

	if jsonEnabled {
		return "json", nil
	}

	return "", fmt.Errorf("config file %q: no config format (ConfigIni or ConfigJSON) is enabled", path)
}
