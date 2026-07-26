// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"reflect"
)

// Validate re-runs parser-level metadata checks after programmatic mutations.
// It applies Configurer hooks and then validates duplicate flag names.
func (p *Parser) Validate() error {
	if p.internalError != nil {
		return p.internalError
	}

	if err := p.applyConfigurators(); err != nil {
		return err
	}

	p.EnsureBuiltinOptions()
	if err := p.EnsureBuiltinCommands(); err != nil {
		return err
	}
	if err := p.validateDuplicateCommands(); err != nil {
		return err
	}

	if err := p.validateDuplicateFlags(); err != nil {
		return err
	}

	if err := p.validateRequiresProvides(); err != nil {
		return err
	}

	p.validationDirty = false
	return nil
}

// Rebuild rescans groups and commands using current tag mapping options.
func (p *Parser) Rebuild() error {
	return p.rebuildTree()
}

func (p *Parser) applyConfigurators() error {
	if !p.configDirty || p.configuring {
		return nil
	}

	p.configuring = true
	defer func() {
		p.configuring = false
	}()

	seen := make(map[uintptr]struct{})

	run := func(data any) error {
		if data == nil {
			return nil
		}

		v := reflect.ValueOf(data)
		if v.IsValid() && v.Kind() == reflect.Pointer && !v.IsNil() {
			ptr := v.Pointer()
			if _, ok := seen[ptr]; ok {
				return nil
			}
			seen[ptr] = struct{}{}
		}

		cfg, ok := data.(Configurer)
		if !ok {
			return nil
		}

		if err := cfg.ConfigureFlags(p); err != nil {
			return fmt.Errorf("configure flags for %T: %w", data, err)
		}

		return nil
	}

	var cfgErr error
	p.eachCommand(func(c *Command) {
		if cfgErr != nil {
			return
		}

		if err := run(c.data); err != nil {
			cfgErr = err
			return
		}

		c.eachGroup(func(g *Group) {
			if cfgErr != nil {
				return
			}

			if err := run(g.data); err != nil {
				cfgErr = err
			}
		})
	})

	if cfgErr != nil {
		return cfgErr
	}

	p.configDirty = false
	return nil
}

// validateRequiresProvides checks, for every command in the tree,
// that each `requires`/`group-requires` token used anywhere in that command's own group scope
// has at least one matching `provides`/`group-provides` token in the same scope.
// A dangling token is a structural configuration mistake
// (not something a caller can fix by passing different CLI args),
// so it is reported as ErrInvalidTag here rather than deferred to the per-parse relation checks
// in checkRequiresOptionRelations/checkRequiresGroupRelations,
// which only report whether a live requirer's dependency is currently satisfied.
func (p *Parser) validateRequiresProvides() error {
	var relErr error

	p.eachCommand(func(c *Command) {
		if relErr != nil {
			return
		}

		relErr = c.checkRequiresProvidesScope()
	})

	return relErr
}

func (c *Command) checkRequiresProvidesScope() error {
	requiresTokens := make(map[string]bool)
	providesTokens := make(map[string]bool)
	groupRequiresTokens := make(map[string]bool)
	groupProvidesTokens := make(map[string]bool)

	c.eachGroup(func(g *Group) {
		for _, option := range g.options {
			for _, token := range option.Requires {
				if token != "" {
					requiresTokens[token] = true
				}
			}
			for _, token := range option.Provides {
				if token != "" {
					providesTokens[token] = true
				}
			}
		}
		for _, token := range g.Requires {
			if token != "" {
				groupRequiresTokens[token] = true
			}
		}
		for _, token := range g.Provides {
			if token != "" {
				groupProvidesTokens[token] = true
			}
		}
	})

	for _, token := range sortedKeys(requiresTokens) {
		if !providesTokens[token] {
			return newErrorf(
				ErrInvalidTag,
				"requires token `%s` has no matching `%s` option in command `%s`",
				token,
				FlagTagProvides,
				c.Name,
			)
		}
	}

	for _, token := range sortedKeys(groupRequiresTokens) {
		if !groupProvidesTokens[token] {
			return newErrorf(
				ErrInvalidTag,
				"group-requires token `%s` has no matching `%s` group in command `%s`",
				token,
				FlagTagGroupProvides,
				c.Name,
			)
		}
	}

	return nil
}

func (p *Parser) validateDuplicateFlags() error {
	var dupErr error

	p.eachCommand(func(c *Command) {
		if dupErr != nil {
			return
		}

		if err := c.checkForDuplicateFlagsInScope(); err != nil {
			dupErr = err
		}
	})

	return dupErr
}

func (p *Parser) validateDuplicateCommands() error {
	var dupErr error

	p.eachCommand(func(c *Command) {
		if dupErr != nil {
			return
		}

		seen := make(map[string]*Command)
		for _, cmd := range c.commands {
			names := append([]string{cmd.Name}, cmd.Aliases...)
			for _, name := range names {
				if name == "" {
					continue
				}
				if other, ok := seen[name]; ok {
					if other == cmd {
						continue
					}
					dupErr = newErrorf(
						ErrDuplicatedFlag,
						"command `%s` uses the same name or alias `%s` as command `%s`",
						cmd.Name,
						name,
						other.Name,
					)
					return
				}
				seen[name] = cmd
			}
		}
	})

	return dupErr
}

func (p *Parser) validateDuplicateEnvKeys() error {
	envKeys := make(map[string]*Option)
	var dupErr error

	p.eachOption(func(_ *Command, _ *Group, option *Option) {
		if dupErr != nil {
			return
		}

		key := option.EnvKeyWithNamespace()
		if key == "" {
			return
		}

		if other, ok := envKeys[key]; ok {
			dupErr = newErrorf(
				ErrDuplicatedFlag,
				"option `%s` uses the same env key `%s` as option `%s`",
				option,
				key,
				other,
			)
			return
		}

		envKeys[key] = option
	})

	return dupErr
}
