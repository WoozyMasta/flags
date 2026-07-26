// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (p *Parser) shouldSkipRequiredValidation() bool {
	if p.Command != nil {
		for cmd := p.Active; cmd != nil; cmd = cmd.Active {
			if _, ok := cmd.data.(builtinCommand); ok {
				return true
			}
		}
	}

	return p.immediateRequested
}

func (p *Parser) shouldSkipCommandExecution() bool {
	return p.immediateRequested
}

func (p *parseState) eof() bool {
	return len(p.args) == 0
}

func (p *parseState) pop() string {
	if p.eof() {
		return ""
	}

	p.arg = p.args[0]
	p.args = p.args[1:]

	return p.arg
}

func (p *parseState) checkRequired(parser *Parser) error {
	c := parser.Command

	var required []*Option

	for c != nil {
		c.eachGroup(func(g *Group) {
			for _, option := range g.options {
				missingRequired := !option.isSet
				if missingRequired &&
					(parser.Options&RequiredFromValues) != None &&
					!option.isEmpty() {
					missingRequired = false
				}

				if missingRequired && option.Required && !option.hasRelationGroups() {
					required = append(required, option)
					continue
				}

				if !option.hasRelationGroups() && option.requiredValueRange {
					count := option.requiredValueCount()
					if count < option.requiredValueMin ||
						(option.requiredValueMax != -1 && count > option.requiredValueMax) {
						required = append(required, option)
					}
				}
			}
		})

		c = c.Active
	}

	if len(required) == 0 {
		if len(p.positional) > 0 {
			var reqnames []string

			for _, arg := range p.positional {
				argRequired := (!arg.isRemaining() && p.command.ArgsRequired) || arg.Required != -1 || arg.RequiredMaximum != -1

				if !argRequired {
					continue
				}

				if arg.isRemaining() {
					if arg.value.Len() < arg.Required {
						var arguments string

						if arg.Required > 1 {
							arguments = "arguments, but got only " + strconv.Itoa(arg.value.Len())
						} else {
							arguments = "argument"
						}

						reqnames = append(reqnames, "`"+arg.localizedName()+" (at least "+strconv.Itoa(arg.Required)+" "+arguments+")`")
					} else if arg.RequiredMaximum != -1 && arg.value.Len() > arg.RequiredMaximum {
						if arg.RequiredMaximum == 0 {
							reqnames = append(reqnames, "`"+arg.localizedName()+" (zero arguments)`")
						} else {
							var arguments string

							if arg.RequiredMaximum > 1 {
								arguments = "arguments, but got " + strconv.Itoa(arg.value.Len())
							} else {
								arguments = "argument"
							}

							reqnames = append(reqnames, "`"+arg.localizedName()+" (at most "+strconv.Itoa(arg.RequiredMaximum)+" "+arguments+")`")
						}
					}
				} else {
					reqnames = append(reqnames, "`"+arg.localizedName()+"`")
				}
			}

			if len(reqnames) == 0 {
				return nil
			}

			var msg string

			if len(reqnames) == 1 {
				msg = parser.i18nTextf(
					"err.required.argument.single",
					"the required argument {arg} was not provided",
					map[string]string{"arg": reqnames[0]},
				)
			} else {
				msg = parser.i18nTextf(
					"err.required.argument.multi",
					"the required arguments {args} and {last} were not provided",
					map[string]string{
						"args": strings.Join(reqnames[:len(reqnames)-1], ", "),
						"last": reqnames[len(reqnames)-1],
					},
				)
			}

			p.err = newError(ErrRequired, msg)
			return p.err
		}

		return nil
	}

	names := make([]string, 0, len(required))

	for _, k := range required {
		names = append(names, "`"+optionRequiredName(k)+"`")
	}

	sort.Strings(names)

	var msg string

	if len(names) == 1 {
		msg = parser.i18nTextf(
			"err.required.flag.single",
			"the required flag {flag} was not specified",
			map[string]string{"flag": names[0]},
		)
	} else {
		msg = parser.i18nTextf(
			"err.required.flag.multi",
			"the required flags {flags} and {last} were not specified",
			map[string]string{
				"flags": strings.Join(names[:len(names)-1], ", "),
				"last":  names[len(names)-1],
			},
		)
	}

	p.err = newError(ErrRequired, msg)
	return p.err
}

func optionRequiredName(option *Option) string {
	if !option.requiredValueRange {
		return option.renderString()
	}

	count := option.requiredValueCount()
	name := option.renderString()

	if count < option.requiredValueMin {
		values := "value"
		got := ""

		if option.requiredValueMin > 1 {
			values = "values"
			got = ", but got only " + strconv.Itoa(count)
		}

		return name + " (at least " + strconv.Itoa(option.requiredValueMin) + " " + values + got + ")"
	}

	if option.requiredValueMax == 0 {
		return name + " (zero values)"
	}

	values := "value"
	got := ""
	if option.requiredValueMax > 1 {
		values = "values"
		got = ", but got " + strconv.Itoa(count)
	}

	return name + " (at most " + strconv.Itoa(option.requiredValueMax) + " " + values + got + ")"
}

func (p *parseState) checkOptionRelations(parser *Parser) error {
	for c := parser.Command; c != nil; c = c.Active {
		if err := p.checkCommandOptionRelations(parser, c); err != nil {
			return err
		}
		if err := p.checkCommandGroupRelations(parser, c); err != nil {
			return err
		}
	}

	return nil
}

func (p *parseState) checkCommandOptionRelations(parser *Parser, command *Command) error {
	xorGroups := make(map[string][]*Option)
	andGroups := make(map[string][]*Option)
	orGroups := make(map[string][]*Option)
	nandGroups := make(map[string][]*Option)
	requiresGroups := make(map[string][]*Option)
	providesGroups := make(map[string][]*Option)

	command.eachGroup(func(g *Group) {
		for _, option := range g.options {
			for _, group := range option.XorGroups {
				if group != "" {
					xorGroups[group] = append(xorGroups[group], option)
				}
			}
			for _, group := range option.AndGroups {
				if group != "" {
					andGroups[group] = append(andGroups[group], option)
				}
			}
			for _, group := range option.OrGroups {
				if group != "" {
					orGroups[group] = append(orGroups[group], option)
				}
			}
			for _, group := range option.NandGroups {
				if group != "" {
					nandGroups[group] = append(nandGroups[group], option)
				}
			}
			for _, token := range option.Requires {
				if token != "" {
					requiresGroups[token] = append(requiresGroups[token], option)
				}
			}
			for _, token := range option.Provides {
				if token != "" {
					providesGroups[token] = append(providesGroups[token], option)
				}
			}
		}
	})

	if err := p.checkXorOptionRelations(parser, xorGroups); err != nil {
		return err
	}
	if err := p.checkAndOptionRelations(parser, andGroups); err != nil {
		return err
	}
	if err := p.checkOrOptionRelations(parser, orGroups); err != nil {
		return err
	}
	if err := p.checkNandOptionRelations(parser, nandGroups); err != nil {
		return err
	}

	return p.checkRequiresOptionRelations(parser, requiresGroups, providesGroups)
}

func (p *parseState) checkXorOptionRelations(parser *Parser, groups map[string][]*Option) error {
	names := sortedKeys(groups)

	for _, name := range names {
		options := groups[name]
		set := setOptions(parser, options)

		switch {
		case len(set) > 1:
			msg := parser.i18nTextf(
				"err.option_conflict.xor",
				"flags {flags} are mutually exclusive",
				map[string]string{"flags": optionList(parser, set)},
			)
			p.err = newError(ErrOptionConflict, msg)
			return p.err
		case len(set) == 0 && relationRequired(options):
			msg := parser.i18nTextf(
				"err.option_requirement.xor",
				"one of flags {flags} must be specified",
				map[string]string{"flags": optionDisjunction(parser, options)},
			)
			p.err = newError(ErrOptionRequirement, msg)
			return p.err
		}
	}

	return nil
}

func (p *parseState) checkAndOptionRelations(parser *Parser, groups map[string][]*Option) error {
	names := sortedKeys(groups)

	for _, name := range names {
		options := groups[name]
		set := setOptions(parser, options)

		if len(set) == len(options) {
			continue
		}

		if len(set) == 0 && !relationRequired(options) {
			continue
		}

		msg := parser.i18nTextf(
			"err.option_requirement.and",
			"flags {flags} must be specified together",
			map[string]string{"flags": optionList(parser, options)},
		)
		p.err = newError(ErrOptionRequirement, msg)
		return p.err
	}

	return nil
}

// checkOrOptionRelations enforces that at least one option in each `or` relation group is set.
// Any number of members may be set together, unlike xor.
// The violation message is identical to xor's empty-required case,
// so the same i18n key is reused.
func (p *parseState) checkOrOptionRelations(parser *Parser, groups map[string][]*Option) error {
	names := sortedKeys(groups)

	for _, name := range names {
		options := groups[name]
		set := setOptions(parser, options)

		if len(set) == 0 {
			msg := parser.i18nTextf(
				"err.option_requirement.xor",
				"one of flags {flags} must be specified",
				map[string]string{"flags": optionDisjunction(parser, options)},
			)
			p.err = newError(ErrOptionRequirement, msg)
			return p.err
		}
	}

	return nil
}

// checkNandOptionRelations enforces that not all options
// in each `nand` relation group are set together.
// Any smaller subset, including none, is allowed.
// A single-member group can never violate this
// and is a permanent no-op, the same way a single-member xor group is.
func (p *parseState) checkNandOptionRelations(parser *Parser, groups map[string][]*Option) error {
	names := sortedKeys(groups)

	for _, name := range names {
		options := groups[name]
		if len(options) < 2 {
			continue
		}

		set := setOptions(parser, options)
		if len(set) == len(options) {
			msg := parser.i18nTextf(
				"err.option_conflict.nand",
				"flags {flags} cannot all be specified together",
				map[string]string{"flags": optionList(parser, options)},
			)
			p.err = newError(ErrOptionConflict, msg)
			return p.err
		}
	}

	return nil
}

// checkRequiresOptionRelations enforces that whenever an option tagged with a `requires` token is set,
// at least one option tagged with the matching `provides` token is also set.
// Satisfaction is many-to-many: any live provider satisfies any live requirer sharing the same token.
// A token with requirers but no providers anywhere in the command is a structural mistake
// caught earlier by Parser.Validate's dangling-requires check, not here.
func (p *parseState) checkRequiresOptionRelations(
	parser *Parser,
	requiresGroups map[string][]*Option,
	providesGroups map[string][]*Option,
) error {
	names := sortedKeys(requiresGroups)

	for _, token := range names {
		requirers := setOptions(parser, requiresGroups[token])
		if len(requirers) == 0 {
			continue
		}

		providers := setOptions(parser, providesGroups[token])
		if len(providers) > 0 {
			continue
		}

		msg := parser.i18nTextf(
			"err.option_requirement.requires",
			"flag {flag} requires one of flags {targets}",
			map[string]string{
				"flag":    optionList(parser, requirers),
				"targets": optionDisjunction(parser, providesGroups[token]),
			},
		)
		p.err = newError(ErrOptionRequirement, msg)
		return p.err
	}

	return nil
}

func relationRequired(options []*Option) bool {
	for _, option := range options {
		if option.Required {
			return true
		}
	}

	return false
}

// checkCommandGroupRelations mirrors checkCommandOptionRelations but operates on whole option groups
// (declared via the `group:"..."` tag) instead of individual options.
// Group-relation tokens (group-xor, group-and, ...) live in a namespace entirely separate
// from option-relation tokens (xor, and, ...):
// the two are collected into different maps below and never compared against each other,
// so a coincidental string match between an option's `xor:"db"`
// and an unrelated group's `group-xor:"db"` cannot interact.
func (p *parseState) checkCommandGroupRelations(parser *Parser, command *Command) error {
	xorGroups := make(map[string][]*Group)
	andGroups := make(map[string][]*Group)
	orGroups := make(map[string][]*Group)
	nandGroups := make(map[string][]*Group)
	requiresGroups := make(map[string][]*Group)
	providesGroups := make(map[string][]*Group)

	command.eachGroup(func(g *Group) {
		for _, token := range g.XorGroups {
			if token != "" {
				xorGroups[token] = append(xorGroups[token], g)
			}
		}
		for _, token := range g.AndGroups {
			if token != "" {
				andGroups[token] = append(andGroups[token], g)
			}
		}
		for _, token := range g.OrGroups {
			if token != "" {
				orGroups[token] = append(orGroups[token], g)
			}
		}
		for _, token := range g.NandGroups {
			if token != "" {
				nandGroups[token] = append(nandGroups[token], g)
			}
		}
		for _, token := range g.Requires {
			if token != "" {
				requiresGroups[token] = append(requiresGroups[token], g)
			}
		}
		for _, token := range g.Provides {
			if token != "" {
				providesGroups[token] = append(providesGroups[token], g)
			}
		}
	})

	if err := p.checkXorGroupRelations(parser, xorGroups); err != nil {
		return err
	}
	if err := p.checkAndGroupRelations(parser, andGroups); err != nil {
		return err
	}
	if err := p.checkOrGroupRelations(parser, orGroups); err != nil {
		return err
	}
	if err := p.checkNandGroupRelations(parser, nandGroups); err != nil {
		return err
	}

	return p.checkRequiresGroupRelations(parser, requiresGroups, providesGroups)
}

func (p *parseState) checkXorGroupRelations(parser *Parser, groups map[string][]*Group) error {
	names := sortedKeys(groups)

	for _, name := range names {
		candidates := groups[name]
		active := setGroups(parser, candidates)

		switch {
		case len(active) > 1:
			msg := parser.i18nTextf(
				"err.option_conflict.group_xor",
				"option groups {groups} are mutually exclusive",
				map[string]string{"groups": groupList(parser, active)},
			)
			p.err = newError(ErrOptionConflict, msg)
			return p.err
		case len(active) == 0 && relationRequiredGroups(candidates):
			msg := parser.i18nTextf(
				"err.option_requirement.group_xor",
				"one of option groups {groups} must be specified",
				map[string]string{"groups": groupDisjunction(parser, candidates)},
			)
			p.err = newError(ErrOptionRequirement, msg)
			return p.err
		}
	}

	return nil
}

func (p *parseState) checkAndGroupRelations(parser *Parser, groups map[string][]*Group) error {
	names := sortedKeys(groups)

	for _, name := range names {
		candidates := groups[name]
		active := setGroups(parser, candidates)

		if len(active) == len(candidates) {
			continue
		}

		if len(active) == 0 && !relationRequiredGroups(candidates) {
			continue
		}

		msg := parser.i18nTextf(
			"err.option_requirement.group_and",
			"option groups {groups} must be specified together",
			map[string]string{"groups": groupList(parser, candidates)},
		)
		p.err = newError(ErrOptionRequirement, msg)
		return p.err
	}

	return nil
}

// checkOrGroupRelations enforces that at least one group in each `group-or` relation is active.
// The violation message reuses the group-xor empty-required key,
// exactly like the option-level or/xor reuse.
func (p *parseState) checkOrGroupRelations(parser *Parser, groups map[string][]*Group) error {
	names := sortedKeys(groups)

	for _, name := range names {
		candidates := groups[name]
		active := setGroups(parser, candidates)

		if len(active) == 0 {
			msg := parser.i18nTextf(
				"err.option_requirement.group_xor",
				"one of option groups {groups} must be specified",
				map[string]string{"groups": groupDisjunction(parser, candidates)},
			)
			p.err = newError(ErrOptionRequirement, msg)
			return p.err
		}
	}

	return nil
}

// checkNandGroupRelations enforces that not all groups in each `group-nand` relation are active together.
// A single-member relation can never violate this and is a permanent no-op,
// the same way a single-member group-xor is.
func (p *parseState) checkNandGroupRelations(parser *Parser, groups map[string][]*Group) error {
	names := sortedKeys(groups)

	for _, name := range names {
		candidates := groups[name]
		if len(candidates) < 2 {
			continue
		}

		active := setGroups(parser, candidates)
		if len(active) == len(candidates) {
			msg := parser.i18nTextf(
				"err.option_conflict.group_nand",
				"option groups {groups} cannot all be specified together",
				map[string]string{"groups": groupList(parser, candidates)},
			)
			p.err = newError(ErrOptionConflict, msg)
			return p.err
		}
	}

	return nil
}

// checkRequiresGroupRelations mirrors checkRequiresOptionRelations for groups:
// whenever an active group tagged with a `group-requires` token exists,
// at least one group tagged with the matching `group-provides` token must also be active.
// A token with requirers but no providers anywhere in the command
// is caught earlier by Parser.Validate's dangling-requires check.
func (p *parseState) checkRequiresGroupRelations(
	parser *Parser,
	requiresGroups map[string][]*Group,
	providesGroups map[string][]*Group,
) error {
	names := sortedKeys(requiresGroups)

	for _, token := range names {
		requirers := setGroups(parser, requiresGroups[token])
		if len(requirers) == 0 {
			continue
		}

		providers := setGroups(parser, providesGroups[token])
		if len(providers) > 0 {
			continue
		}

		msg := parser.i18nTextf(
			"err.option_requirement.group_requires",
			"option group {group} requires one of option groups {groups}",
			map[string]string{
				"group":  groupList(parser, requirers),
				"groups": groupDisjunction(parser, providesGroups[token]),
			},
		)
		p.err = newError(ErrOptionRequirement, msg)
		return p.err
	}

	return nil
}

func relationRequiredGroups(groups []*Group) bool {
	for _, g := range groups {
		if g.Required {
			return true
		}
	}

	return false
}

func setGroups(parser *Parser, groups []*Group) []*Group {
	out := make([]*Group, 0, len(groups))

	for _, g := range groups {
		if groupSatisfied(parser, g) {
			out = append(out, g)
		}
	}

	return out
}

// groupSatisfied reports whether a group is active:
// any option belonging to it or to any of its nested subgroups (recursive) is satisfied.
// Applying the same group-relation token to a group and its own ancestor/descendant
// is nonsensical (the descendant's state leaks into the ancestor's)
// and is not guarded against here; see the Group Tags documentation.
func groupSatisfied(parser *Parser, g *Group) bool {
	satisfied := false

	g.eachGroup(func(gg *Group) {
		if satisfied {
			return
		}
		for _, option := range gg.options {
			if optionSatisfied(parser, option) {
				satisfied = true
				return
			}
		}
	})

	return satisfied
}

func groupList(parser *Parser, groups []*Group) string {
	return joinedGroupList(parser, groups, "err.list.conjunction", "{items} and {last}")
}

func groupDisjunction(parser *Parser, groups []*Group) string {
	return joinedGroupList(parser, groups, "err.list.disjunction", "{items} or {last}")
}

func joinedGroupList(parser *Parser, groups []*Group, key string, fallback string) string {
	names := make([]string, 0, len(groups))

	for _, g := range groups {
		names = append(names, g.ShortDescription)
	}

	sort.Strings(names)

	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}

	return parser.i18nTextf(
		key,
		fallback,
		map[string]string{
			"items": strings.Join(names[:len(names)-1], ", "),
			"last":  names[len(names)-1],
		},
	)
}

func setOptions(parser *Parser, options []*Option) []*Option {
	out := make([]*Option, 0, len(options))

	for _, option := range options {
		if optionSatisfied(parser, option) {
			out = append(out, option)
		}
	}

	return out
}

func optionSatisfied(parser *Parser, option *Option) bool {
	if option.isSet {
		return true
	}

	return (parser.Options&RequiredFromValues) != None && !option.isEmpty()
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func optionList(parser *Parser, options []*Option) string {
	return joinedOptionList(parser, options, "err.list.conjunction", "{items} and {last}")
}

func optionDisjunction(parser *Parser, options []*Option) string {
	return joinedOptionList(parser, options, "err.list.disjunction", "{items} or {last}")
}

func joinedOptionList(parser *Parser, options []*Option, key string, fallback string) string {
	names := make([]string, 0, len(options))

	for _, option := range options {
		names = append(names, option.renderString())
	}

	sort.Strings(names)

	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}

	return parser.i18nTextf(
		key,
		fallback,
		map[string]string{
			"items": strings.Join(names[:len(names)-1], ", "),
			"last":  names[len(names)-1],
		},
	)
}

func (p *parseState) estimateCommand() error {
	commands := p.command.sortedVisibleCommands()
	cmdnames := make([]string, len(commands))
	parser := p.command.parser()
	i18nTextf := func(key, fallback string, data map[string]string) string {
		if parser == nil {
			for k, v := range data {
				fallback = strings.ReplaceAll(fallback, "{"+k+"}", v)
			}
			return fallback
		}

		return parser.i18nTextf(key, fallback, data)
	}

	for i, v := range commands {
		cmdnames[i] = v.Name
	}

	var msg string
	var errtype ErrorType

	if len(p.retargs) != 0 {
		c, l := closestChoice(p.retargs[0], cmdnames)
		msg = i18nTextf(
			"err.command.unknown",
			"Unknown command `{command}`",
			map[string]string{"command": p.retargs[0]},
		)
		errtype = ErrUnknownCommand

		switch {
		case float32(l)/float32(len(c)) < 0.5:
			msg = i18nTextf(
				"err.command.did_you_mean",
				"{base}, did you mean `{choice}`?",
				map[string]string{
					"base":   msg,
					"choice": c,
				},
			)
		case len(cmdnames) == 1:
			msg = i18nTextf(
				"err.command.should_use",
				"{base}. You should use the {command} command",
				map[string]string{
					"base":    msg,
					"command": cmdnames[0],
				},
			)
		case len(cmdnames) > 1:
			msg = i18nTextf(
				"err.command.specify_one",
				"{base}. Please specify one command of: {commands} or {last}",
				map[string]string{
					"base":     msg,
					"commands": strings.Join(cmdnames[:len(cmdnames)-1], ", "),
					"last":     cmdnames[len(cmdnames)-1],
				},
			)
		}
	} else {
		errtype = ErrCommandRequired

		switch {
		case len(cmdnames) == 1:
			msg = i18nTextf(
				"err.command.required.single",
				"Please specify the {command} command",
				map[string]string{"command": cmdnames[0]},
			)
		case len(cmdnames) > 1:
			msg = i18nTextf(
				"err.command.required.multi",
				"Please specify one command of: {commands} or {last}",
				map[string]string{
					"commands": strings.Join(cmdnames[:len(cmdnames)-1], ", "),
					"last":     cmdnames[len(cmdnames)-1],
				},
			)
		}
	}

	return newError(errtype, msg)
}

func (p *Parser) parseOption(s *parseState, _ string, option *Option, canarg bool, argument string, hasArgument bool) (err error) {
	if option.Deprecated != "" && !option.isSet && (p.Options&SilenceDeprecationWarnings) == None {
		p.printDeprecationWarning("flag", option.renderString(), option.Deprecated)
	}

	if option.Counter {
		if !hasArgument && canarg && !s.eof() {
			next := s.args[0]
			if option.isValidValue(next) == nil {
				if p.Options&PassDoubleDash == 0 || next != "--" {
					argument = s.pop()
					hasArgument = true
				}
			}
		}

		delta := uint64(1)
		if hasArgument {
			delta, err = parseCounterDelta(argument, option)
			if err != nil {
				return p.marshalError(option, err)
			}
		}

		if err = option.applyCounterDelta(delta); err != nil {
			return p.marshalError(option, err)
		}

		if option.IsImmediate() {
			p.immediateRequested = true
		}

		return nil
	}

	switch {
	case !option.canArgument():
		if hasArgument && (p.Options&AllowBoolValues) == None {
			return newError(
				ErrNoArgumentForBool,
				p.i18nTextf(
					"err.bool.no_argument",
					"bool flag `{flag}` cannot have an argument",
					map[string]string{"flag": option.renderString()},
				),
			)
		}
		var value *string
		if hasArgument {
			value = &argument
		}
		err = option.Set(value)
	case option.isTerminated():
		if hasArgument {
			return newError(
				ErrExpectedArgument,
				p.i18nTextf(
					"err.terminated.inline_argument",
					"terminated option flag `{flag}` cannot use inline argument syntax",
					map[string]string{"flag": option.renderString()},
				),
			)
		}

		args, collectErr := p.collectTerminatedArgs(s, option)
		if collectErr != nil {
			return collectErr
		}
		err = option.SetTerminated(args)
	case hasArgument || (canarg && !s.eof()):
		var arg string

		if hasArgument {
			arg = argument
		} else {
			arg = s.pop()

			if validationErr := option.isValidValue(arg); validationErr != nil {
				return newErrorf(ErrExpectedArgument, "%s", validationErr)
			} else if p.Options&PassDoubleDash != 0 && arg == "--" {
				return newError(
					ErrExpectedArgument,
					p.i18nTextf(
						"err.expected_argument.double_dash",
						"expected argument for flag `{flag}`, but got double dash `--`",
						map[string]string{"flag": option.renderString()},
					),
				)
			}
		}

		if option.tag.Get(FlagTagUnquote) != "false" {
			arg, err = unquoteIfPossible(arg)
		}

		if err == nil {
			err = option.Set(&arg)
		}
	case option.OptionalArgument:
		option.empty()

		for _, v := range option.OptionalValue {
			err = option.Set(&v)

			if err != nil {
				break
			}
		}
	default:
		err = newError(
			ErrExpectedArgument,
			p.i18nTextf(
				"err.expected_argument.flag",
				"expected argument for flag `{flag}`",
				map[string]string{"flag": option.renderString()},
			),
		)
	}

	if err != nil {
		if _, ok := err.(*Error); !ok {
			err = p.marshalError(option, err)
		}
	} else if option.IsImmediate() {
		p.immediateRequested = true
	}

	return err
}

func parseCounterDelta(raw string, option *Option) (uint64, error) {
	switch option.value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, err
		}
		if parsed < 0 {
			return 0, ErrCounterNonNegative
		}
		return uint64(parsed), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil

	default:
		return 0, ErrCounterInvalidType
	}
}

func (p *Parser) collectTerminatedArgs(s *parseState, option *Option) ([]string, error) {
	args := make([]string, 0, 4)

	for !s.eof() {
		arg := s.pop()

		if arg == option.Terminator {
			break
		}

		if option.tag.Get(FlagTagUnquote) != "false" {
			unquoted, err := unquoteIfPossible(arg)
			if err != nil {
				return nil, p.marshalError(option, err)
			}
			arg = unquoted
		}

		args = append(args, arg)
	}

	return args, nil
}

func (p *Parser) marshalError(option *Option, err error) *Error {
	expected := ""

	if expectedType := p.expectedType(option); expectedType != "" {
		expected = p.i18nTextf(
			"err.marshal.expected",
			" (expected {type})",
			map[string]string{"type": expectedType},
		)
	}

	errorText := err.Error()
	if option.Secret {
		errorText = secretValueMask
	}

	return newError(
		ErrMarshal,
		p.i18nTextf(
			"err.marshal.option",
			"invalid argument for flag `{flag}`{expected}: {error}",
			map[string]string{
				"flag":     option.renderString(),
				"expected": expected,
				"error":    errorText,
			},
		),
	)
}

func (p *Parser) expectedType(option *Option) string {
	valueType := option.value.Type()

	if valueType.Kind() == reflect.Func {
		return ""
	}

	return valueType.String()
}

func (p *Parser) parseLong(s *parseState, name string, argument string, hasArgument bool) error {
	if option := s.lookup.longNames[name]; option != nil {
		// Only long options that are required can consume an argument
		// from the argument list
		canarg := !option.OptionalArgument

		return p.parseOption(s, name, option, canarg, argument, hasArgument)
	}

	msg := p.i18nTextf(
		"err.unknown_flag",
		"unknown flag `{flag}`",
		map[string]string{"flag": name},
	)

	if len(s.lookup.longNames) > 0 {
		names := make([]string, 0, len(s.lookup.longNames))
		for n := range s.lookup.longNames {
			names = append(names, n)
		}
		c, l := closestChoice(name, names)
		if len(c) > 0 && l > 0 && l < utf8.RuneCountInString(name) && float32(l)/float32(utf8.RuneCountInString(c)) < 0.5 {
			rf := p.optionRenderFormat()
			msg = p.i18nTextf(
				"err.flag.did_you_mean",
				"{base}, did you mean `{choice}`?",
				map[string]string{
					"base":   msg,
					"choice": rf.longDelimiter + c,
				},
			)
		}
	}

	return newError(ErrUnknownFlag, msg)
}

func (p *Parser) splitShortConcatArg(s *parseState, optname string) (string, *string) {
	c, n := utf8.DecodeRuneInString(optname)

	if n == len(optname) {
		return optname, nil
	}

	first := string(c)

	if option := s.lookup.shortNames[first]; option != nil && option.canArgument() {
		if option.Counter && isCounterShortCluster(optname, c) {
			return optname, nil
		}
		arg := optname[n:]
		return first, &arg
	}

	return optname, nil
}

func isCounterShortCluster(optname string, first rune) bool {
	if utf8.RuneCountInString(optname) <= 1 {
		return false
	}

	for _, c := range optname {
		if c != first {
			return false
		}
	}

	return true
}

func (p *Parser) parseShort(s *parseState, optname string, argument string, hasArgument bool) error {
	if !hasArgument {
		var ptr *string
		optname, ptr = p.splitShortConcatArg(s, optname)
		if ptr != nil {
			argument = *ptr
			hasArgument = true
		}
	}

	for i, c := range optname {
		shortname := string(c)

		if option := s.lookup.shortNames[shortname]; option != nil {
			// Only the last short argument can consume an argument from
			// the arguments list, and only if it's non optional
			canarg := (i+utf8.RuneLen(c) == len(optname)) && !option.OptionalArgument

			if err := p.parseOption(s, shortname, option, canarg, argument, hasArgument); err != nil {
				return err
			}
		} else {
			return newError(
				ErrUnknownFlag,
				p.i18nTextf(
					"err.unknown_flag",
					"unknown flag `{flag}`",
					map[string]string{"flag": shortname},
				),
			)
		}

		// Only the first option can have a concatted argument, so just
		// clear argument here
		argument = ""
		hasArgument = false
	}

	return nil
}

func (p *parseState) addArgs(args ...string) error {
	for len(p.positional) > 0 && len(args) > 0 {
		arg := p.positional[0]
		raw := args[0]
		if arg.io.role != "" {
			normalized, normErr := arg.normalizeIOValue(raw)
			if normErr != nil {
				p.err = newErrorf(ErrMarshal, "invalid positional argument `%s`: %v", arg.localizedName(), normErr)
				return p.err
			}
			raw = normalized
		}

		if err := convert(raw, arg.value, arg.tag); err != nil {
			p.err = err
			return err
		}

		if !arg.isRemaining() {
			p.positional = p.positional[1:]
		}

		args = args[1:]
	}

	p.retargs = append(p.retargs, args...)
	return nil
}

func (p *parseState) applyPositionalDefaults(parser *Parser, defaultsIfEmpty bool) error {
	for len(p.positional) > 0 {
		arg := p.positional[0]

		applied := false
		if len(arg.Default) > 0 {
			if err := arg.applyDefault(defaultsIfEmpty); err != nil {
				p.err = newError(
					ErrMarshal,
					parser.i18nTextf(
						"err.marshal.argument_default",
						"invalid default for argument `{arg}`: {error}",
						map[string]string{
							"arg":   arg.localizedName(),
							"error": err.Error(),
						},
					),
				)
				return p.err
			}
			applied = true
		} else {
			ok, err := arg.applyIOFallback()
			if err != nil {
				p.err = newError(
					ErrMarshal,
					parser.i18nTextf(
						"err.marshal.argument_default",
						"invalid default for argument `{arg}`: {error}",
						map[string]string{
							"arg":   arg.localizedName(),
							"error": err.Error(),
						},
					),
				)
				return p.err
			}
			applied = ok
		}
		if !applied {
			break
		}

		p.positional = p.positional[1:]

		if arg.isRemaining() {
			break
		}
	}

	return nil
}

func (p *Parser) parseNonOption(s *parseState) error {
	isStrictCmds := (p.Options&StrictCommands) != 0 || s.command.StrictSubcommands
	isStrictPos := (p.Options&StrictPositionalArgs) != 0 || s.command.StrictArgs

	if len(s.command.commands) > 0 && len(s.retargs) == 0 {
		if cmd := s.lookup.commands[s.arg]; cmd != nil {
			if len(s.positional) > 0 {
				if _, ok := cmd.data.(builtinCommand); !ok {
					return s.addArgs(s.arg)
				}
			}

			s.command.Active = cmd
			cmd.fillParseState(s)

			if cmd.Deprecated != "" && (p.Options&SilenceDeprecationWarnings) == None {
				p.printDeprecationWarning("command", cmd.Name, cmd.Deprecated)
			}

			return nil
		} else if s.command.defaultCommand != "" {
			cmd := s.lookup.commands[s.command.defaultCommand]
			if cmd == nil {
				return newErrorf(
					ErrUnknownCommand,
					"default command %q not found",
					s.command.defaultCommand,
				)
			}

			s.command.Active = cmd
			cmd.fillParseState(s)

			if cmd.Deprecated != "" && (p.Options&SilenceDeprecationWarnings) == None {
				p.printDeprecationWarning("command", cmd.Name, cmd.Deprecated)
			}

			return p.parseNonOption(s)
		} else if !s.command.SubcommandsOptional || isStrictCmds {
			if len(s.positional) > 0 && !isStrictCmds {
				// Lenient mode: treat unknown command token as positional arg.
				return s.addArgs(s.arg)
			}

			if p.UnknownCommandHandler != nil {
				if err := p.UnknownCommandHandler(s.arg, s.args); err != nil {
					s.err = err
					return s.err
				}
				// Handler accepted the token - add to retargs,
				// mark as handled so estimateCommand() is suppressed at the end of the parse loop.
				s.retargs = append(s.retargs, s.arg)
				s.unknownCmdHandled = true
				return nil
			}

			if isStrictCmds {
				// Strict mode: add directly to retargs (bypass positional slots)
				// so estimateCommand() produces the right "Unknown command" message.
				// Set s.err directly to preserve the error even when SubcommandsOptional is true
				// (where estimateCommand wouldn't run after the loop).
				s.retargs = append(s.retargs, s.arg)
				s.err = s.estimateCommand()
				return s.err
			}

			// Original non-strict path (SubcommandsOptional=false, no positional):
			// add to retargs and return a break signal; estimateCommand() runs after the parse loop.
			if err := s.addArgs(s.arg); err != nil {
				return err
			}
			return newError(
				ErrUnknownCommand,
				p.i18nTextf(
					"err.command.unknown",
					"Unknown command `{command}`",
					map[string]string{"command": s.arg},
				),
			)
		}
	}

	if isStrictPos && len(s.positional) == 0 {
		s.err = newError(
			ErrUnexpectedArgument,
			p.i18nTextf(
				"err.arg.unexpected",
				"Unexpected argument `{arg}`",
				map[string]string{"arg": s.arg},
			),
		)
		return s.err
	}

	return s.addArgs(s.arg)
}

func (p *Parser) printDeprecationWarning(kind, name, msg string) {
	var text string

	if kind == "command" {
		text = p.i18nTextf(
			"warn.deprecated.command",
			"warning: command `{command}` is deprecated: {reason}",
			map[string]string{"command": name, "reason": msg},
		)
	} else {
		text = p.i18nTextf(
			"warn.deprecated.flag",
			"warning: flag `{flag}` is deprecated: {reason}",
			map[string]string{"flag": name, "reason": msg},
		)
	}

	w := io.Writer(os.Stderr)
	if (p.Options & PrintErrorsOnStdout) != None {
		w = os.Stdout
	}

	_, _ = fmt.Fprintln(w, text)
}
