// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
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
		names = append(names, "`"+k.String()+"`")
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

func (p *parseState) checkOptionRelations(parser *Parser) error {
	for c := parser.Command; c != nil; c = c.Active {
		if err := p.checkCommandOptionRelations(parser, c); err != nil {
			return err
		}
	}

	return nil
}

func (p *parseState) checkCommandOptionRelations(parser *Parser, command *Command) error {
	xorGroups := make(map[string][]*Option)
	andGroups := make(map[string][]*Option)

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
		}
	})

	if err := p.checkXorOptionRelations(parser, xorGroups); err != nil {
		return err
	}

	return p.checkAndOptionRelations(parser, andGroups)
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

func relationRequired(options []*Option) bool {
	for _, option := range options {
		if option.Required {
			return true
		}
	}

	return false
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
		names = append(names, option.String())
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
					map[string]string{"flag": option.String()},
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
					map[string]string{"flag": option.String()},
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
						map[string]string{"flag": option.String()},
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
				map[string]string{"flag": option.String()},
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
				return nil, err
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

	return newError(
		ErrMarshal,
		p.i18nTextf(
			"err.marshal.option",
			"invalid argument for flag `{flag}`{expected}: {error}",
			map[string]string{
				"flag":     option.String(),
				"expected": expected,
				"error":    err.Error(),
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

	return newError(
		ErrUnknownFlag,
		p.i18nTextf(
			"err.unknown_flag",
			"unknown flag `{flag}`",
			map[string]string{"flag": name},
		),
	)
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
	if len(s.command.commands) > 0 && len(s.retargs) == 0 {
		if cmd := s.lookup.commands[s.arg]; cmd != nil {
			if len(s.positional) > 0 {
				if _, ok := cmd.data.(builtinCommand); !ok {
					return s.addArgs(s.arg)
				}
			}

			s.command.Active = cmd
			cmd.fillParseState(s)

			return nil
		} else if !s.command.SubcommandsOptional {
			if len(s.positional) > 0 {
				return s.addArgs(s.arg)
			}

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

	if len(s.positional) > 0 {
		return s.addArgs(s.arg)
	}

	return s.addArgs(s.arg)
}
