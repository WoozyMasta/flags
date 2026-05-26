// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"reflect"
	"sort"
)

// ExportMode controls which options are included in ExportArgs output.
type ExportMode int

const (
	// ExportModeResolved includes all options with a resolved (non-zero) value.
	// Empty slices and maps are skipped unless WithIncludeEmptyCollections is
	// set. Booleans are always included regardless of their value.
	ExportModeResolved ExportMode = iota

	// ExportModeExplicit includes only options that were explicitly provided on
	// the command line and were not applied from a default tag or environment
	// variable.
	ExportModeExplicit

	// ExportModeNonZero includes only options whose current value is non-zero.
	ExportModeNonZero
)

// ExportSecretPolicy controls how secret options are handled in ExportArgs.
type ExportSecretPolicy int

const (
	// ExportSecretsExclude omits secret options from the output. Default.
	ExportSecretsExclude ExportSecretPolicy = iota

	// ExportSecretsInclude exports secret options with their current values.
	ExportSecretsInclude
)

type exportArgsConfig struct {
	commandPath             []string
	mode                    ExportMode
	secrets                 ExportSecretPolicy
	includeHidden           bool
	includeDeprecated       bool
	includeEmptyCollections bool
}

// ExportArgsOption is a functional option for ExportArgs.
type ExportArgsOption func(*exportArgsConfig)

// WithMode sets the selection mode for ExportArgs.
// Default is ExportModeResolved.
func WithMode(mode ExportMode) ExportArgsOption {
	return func(cfg *exportArgsConfig) {
		cfg.mode = mode
	}
}

// WithCommandPath targets a specific subcommand by path. Each element is a
// command name navigated from the root. An error is returned if any element is
// not found. Without this option ExportArgs follows the active command chain
// resolved by the most recent ParseArgs call.
func WithCommandPath(path ...string) ExportArgsOption {
	return func(cfg *exportArgsConfig) {
		cfg.commandPath = append(cfg.commandPath[:0], path...)
	}
}

// WithSecrets controls secret option handling.
// Default is ExportSecretsExclude.
func WithSecrets(policy ExportSecretPolicy) ExportArgsOption {
	return func(cfg *exportArgsConfig) {
		cfg.secrets = policy
	}
}

// WithHidden includes hidden options in the output.
// Default is false.
func WithHidden(v bool) ExportArgsOption {
	return func(cfg *exportArgsConfig) {
		cfg.includeHidden = v
	}
}

// WithIncludeDeprecated controls whether deprecated options are exported.
// Default is true.
func WithIncludeDeprecated(v bool) ExportArgsOption {
	return func(cfg *exportArgsConfig) {
		cfg.includeDeprecated = v
	}
}

// WithIncludeEmptyCollections includes slice and map options even when empty.
// Default is false.
func WithIncludeEmptyCollections(v bool) ExportArgsOption {
	return func(cfg *exportArgsConfig) {
		cfg.includeEmptyCollections = v
	}
}

// ExportArgs serializes the current parser state into CLI argv tokens suitable
// for passing back to ParseArgs. The returned slice is deterministic for the
// same parser state. Subcommand names are emitted as tokens before their
// respective option flags.
//
// Defaults:
//   - Mode:              ExportModeResolved
//   - Secrets:           ExportSecretsExclude
//   - Hidden:            excluded
//   - Deprecated:        included
//   - Empty collections: excluded
func (p *Parser) ExportArgs(opts ...ExportArgsOption) ([]string, error) {
	cfg := exportArgsConfig{
		mode:              ExportModeResolved,
		secrets:           ExportSecretsExclude,
		includeHidden:     false,
		includeDeprecated: true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	chain, err := p.resolveExportCommandChain(&cfg)
	if err != nil {
		return nil, err
	}

	var args []string

	for i, cmd := range chain {
		if i > 0 {
			args = append(args, cmd.Name)
		}

		var groupErr error

		cmd.eachGroup(func(g *Group) {
			if groupErr != nil {
				return
			}

			for _, option := range g.options {
				if !exportInclude(option, &cfg) {
					continue
				}

				tokens, tokenErr := exportOptionTokens(option)
				if tokenErr != nil {
					groupErr = tokenErr
					return
				}

				args = append(args, tokens...)
			}
		})

		if groupErr != nil {
			return nil, groupErr
		}
	}

	return args, nil
}

// resolveExportCommandChain returns the chain of commands to export.
func (p *Parser) resolveExportCommandChain(cfg *exportArgsConfig) ([]*Command, error) {
	if len(cfg.commandPath) == 0 {
		return exportActiveChain(p.Command), nil
	}

	chain := []*Command{p.Command}
	current := p.Command

	for _, name := range cfg.commandPath {
		var next *Command

		for _, sub := range current.commands {
			if sub.match(name) {
				next = sub
				break
			}
		}

		if next == nil {
			return nil, newErrorf(ErrUnknownCommand, "unknown command `%s`", name)
		}

		chain = append(chain, next)
		current = next
	}

	return chain, nil
}

// exportActiveChain builds the command chain by following .Active pointers.
func exportActiveChain(root *Command) []*Command {
	chain := []*Command{root}
	cur := root

	for cur.Active != nil {
		cur = cur.Active
		chain = append(chain, cur)
	}

	return chain
}

// exportInclude reports whether the option should appear in ExportArgs output.
func exportInclude(option *Option, cfg *exportArgsConfig) bool {
	if option.isFunc() {
		return false
	}

	if cfg.secrets == ExportSecretsExclude && option.Secret {
		return false
	}

	if !cfg.includeHidden && option.Hidden {
		return false
	}

	if !cfg.includeDeprecated && option.Deprecated != "" {
		return false
	}

	switch cfg.mode {
	case ExportModeExplicit:
		return option.isSet && !option.isSetDefault

	case ExportModeNonZero:
		return !option.isEmpty()

	default: // ExportModeResolved
		if option.isBool() {
			return true
		}

		kind := option.value.Type().Kind()

		if kind == reflect.Slice || kind == reflect.Map {
			if option.value.Len() == 0 {
				return cfg.includeEmptyCollections
			}

			return true
		}

		return !option.isEmpty()
	}
}

// exportOptionTokens converts one option to its argv token(s).
func exportOptionTokens(option *Option) ([]string, error) {
	flagName := exportFlagName(option)
	sep := string(defaultNameArgDelimiter)
	val := option.value
	kind := val.Type().Kind()

	switch kind {
	case reflect.Slice:
		if val.Len() == 0 {
			return nil, nil
		}

		tokens := make([]string, 0, val.Len())

		for i := range val.Len() {
			s, err := convertToString(val.Index(i), option.tag)
			if err != nil {
				return nil, fmt.Errorf("export flag %s index %d: %w", flagName, i, err)
			}

			tokens = append(tokens, flagName+sep+s)
		}

		return tokens, nil

	case reflect.Map:
		if val.Len() == 0 {
			return nil, nil
		}

		delim := option.tag.Get(FlagTagKeyValueDelimiter)
		if delim == "" {
			delim = ":"
		}

		mkeys := val.MapKeys()
		strKeys := make([]string, len(mkeys))
		keyMap := make(map[string]reflect.Value, len(mkeys))

		for i, k := range mkeys {
			ks, err := convertToString(k, option.tag)
			if err != nil {
				return nil, fmt.Errorf("export flag %s map key: %w", flagName, err)
			}

			strKeys[i] = ks
			keyMap[ks] = k
		}

		sort.Strings(strKeys)

		tokens := make([]string, 0, len(strKeys))

		for _, ks := range strKeys {
			vs, err := convertToString(val.MapIndex(keyMap[ks]), option.tag)
			if err != nil {
				return nil, fmt.Errorf("export flag %s map value for key %q: %w", flagName, ks, err)
			}

			tokens = append(tokens, flagName+sep+ks+delim+vs)
		}

		return tokens, nil

	default:
		s, err := convertToString(val, option.tag)
		if err != nil {
			return nil, fmt.Errorf("export flag %s: %w", flagName, err)
		}

		return []string{flagName + sep + s}, nil
	}
}

// exportFlagName returns the canonical flag prefix+name token.
func exportFlagName(option *Option) string {
	if long := option.LongNameWithNamespace(); long != "" {
		return defaultLongOptDelimiter + long
	}

	return string(defaultShortOptDelimiter) + string(option.ShortName)
}
