// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"reflect"
)

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
		if v.IsValid() && v.Kind() == reflect.Ptr && !v.IsNil() {
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
						"command `%s' uses the same name or alias `%s' as command `%s'",
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
				"option `%s' uses the same env key `%s' as option `%s'",
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

	return p.validateDuplicateFlags()
}

// Rebuild rescans groups and commands using current tag mapping options.
func (p *Parser) Rebuild() error {
	return p.rebuildTree()
}
