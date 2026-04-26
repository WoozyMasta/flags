// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Command represents an application command. Commands can be added to the
// parser (which itself is a command) and are selected/executed when its name
// is specified on the command line. The Command type embeds a Group and
// therefore also carries a set of command specific options.
type Command struct {
	lookupCache lookup

	// Embedded, see Group for more information
	*Group

	// The active sub command (set by parsing) or nil
	Active *Command

	// The name by which the command can be invoked
	Name string

	// Group name used to organize commands in help/docs.
	CommandGroup string

	// Optional i18n key for CommandGroup.
	CommandGroupI18nKey string

	// Aliases for the command
	Aliases []string

	// All direct subcommands of this command.
	commands []*Command

	// Positional arguments declared for this command.
	args []*Arg

	// Display sort index used in help/docs command ordering.
	// Positive values are shown first, then zero, then negative.
	Order int

	lookupCacheGeneration uint64

	// Whether subcommands are optional
	SubcommandsOptional bool

	// Whether positional arguments are required
	ArgsRequired bool

	// Whether to pass all arguments after the first non option as remaining
	// command line arguments. This is equivalent to strict POSIX processing.
	// This is command-local version of PassAfterNonOption Parser flag. It
	// cannot be turned off when PassAfterNonOption Parser flag is set.
	PassAfterNonOption bool

	// Whether the built-in help group has already been attached.
	hasBuiltinHelpGroup bool

	lookupCacheValid bool
}

// Commander is an interface which can be implemented by any command added in
// the options. When implemented, the Execute method will be called for the last
// specified (sub)command providing the remaining command line arguments.
type Commander interface {
	// Execute will be called for the last active (sub)command. The
	// args argument contains the remaining command line arguments. The
	// error that Execute returns will be eventually passed out of the
	// Parse method of the Parser.
	Execute(args []string) error
}

// Usage is an interface which can be implemented to show a custom usage string
// in the help message shown for a command.
type Usage interface {
	// Usage is called for commands to allow customized printing of command
	// usage in the generated help message.
	Usage() string
}

type lookup struct {
	shortNames map[string]*Option
	longNames  map[string]*Option

	commands map[string]*Command
}

// AddCommand adds a new command to the parser with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the command. The provided data can implement the Command and
// Usage interfaces.
func (c *Command) AddCommand(command string, shortDescription string, longDescription string, data any) (*Command, error) {
	cmd := newCommand(command, shortDescription, longDescription, data)

	cmd.parent = c

	if err := cmd.scan(); err != nil {
		return nil, err
	}

	c.commands = append(c.commands, cmd)

	if p := c.parser(); p != nil {
		p.invalidateLookupCache()
	}

	return cmd, nil
}

// SetName updates command name used for lookup and help output.
func (c *Command) SetName(name string) {
	c.Name = name
	c.touchLookupCache()
}

// SetAliases replaces command aliases used for lookup and help output.
func (c *Command) SetAliases(aliases ...string) {
	c.Aliases = append(c.Aliases[:0], aliases...)
	c.touchLookupCache()
}

// SetCommandGroup updates help/docs group used for this command.
func (c *Command) SetCommandGroup(group string) {
	c.CommandGroup = group
	c.CommandGroupI18nKey = ""
}

// SetOrder updates display sort index used for help/docs command ordering.
func (c *Command) SetOrder(order int) {
	c.Order = order
}

// AddAlias appends one command alias.
func (c *Command) AddAlias(alias string) {
	c.Aliases = append(c.Aliases, alias)
	c.touchLookupCache()
}

// SetShortDescription updates command short description.
func (c *Command) SetShortDescription(description string) {
	c.ShortDescription = description
}

// SetShortDescriptionI18nKey sets i18n key for command short description.
func (c *Command) SetShortDescriptionI18nKey(key string) {
	c.Group.SetShortDescriptionI18nKey(key)
}

// SetLongDescription updates command long description.
func (c *Command) SetLongDescription(description string) {
	c.LongDescription = description
}

// SetLongDescriptionI18nKey sets i18n key for command long description.
func (c *Command) SetLongDescriptionI18nKey(key string) {
	c.Group.SetLongDescriptionI18nKey(key)
}

// SetIniName updates stable INI section token used for this command block.
func (c *Command) SetIniName(name string) {
	c.Group.SetIniName(name)
}

// SetHidden controls command visibility in help/completion/docs.
func (c *Command) SetHidden(hidden bool) {
	c.Hidden = hidden
}

// SetSubcommandsOptional configures whether subcommand selection is optional.
func (c *Command) SetSubcommandsOptional(optional bool) {
	c.SubcommandsOptional = optional
}

// SetPassAfterNonOption configures command-local strict POSIX pass-through behavior.
func (c *Command) SetPassAfterNonOption(enabled bool) {
	c.PassAfterNonOption = enabled
}

// SetArgsRequired configures whether positional args are required by default.
func (c *Command) SetArgsRequired(required bool) {
	c.ArgsRequired = required
}

// AddGroup adds a new group to the command with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the group.
func (c *Command) AddGroup(shortDescription string, longDescription string, data any) (*Group, error) {
	group := newGroup(shortDescription, longDescription, data)

	group.parent = c

	if err := group.scanType(c.scanSubcommandHandler(group)); err != nil {
		return nil, err
	}

	c.groups = append(c.groups, group)

	if p := c.parser(); p != nil {
		p.invalidateLookupCache()
	}

	return group, nil
}

// Commands returns a list of subcommands of this command.
func (c *Command) Commands() []*Command {
	return c.commands
}

// Find locates the subcommand with the given name and returns it. If no such
// command can be found Find will return nil.
func (c *Command) Find(name string) *Command {
	for _, cc := range c.commands {
		if cc.match(name) {
			return cc
		}
	}

	return nil
}

// FindOptionByLongName finds an option that is part of the command, or any of
// its parent commands, by matching its long name (including the option
// namespace).
func (c *Command) FindOptionByLongName(longName string) (option *Option) {
	for option == nil && c != nil {
		option = c.Group.FindOptionByLongName(longName)

		c, _ = c.parent.(*Command)
	}

	return option
}

// FindOptionByShortName finds an option that is part of the command, or any of
// its parent commands, by matching its long name (including the option
// namespace).
func (c *Command) FindOptionByShortName(shortName rune) (option *Option) {
	for option == nil && c != nil {
		option = c.Group.FindOptionByShortName(shortName)

		c, _ = c.parent.(*Command)
	}

	return option
}

// Args returns a list of positional arguments associated with this command.
func (c *Command) Args() []*Arg {
	ret := make([]*Arg, len(c.args))
	copy(ret, c.args)

	return ret
}

func (c *Command) touchLookupCache() {
	if p := c.parser(); p != nil {
		p.invalidateLookupCache()
	}
}

func newCommand(name string, shortDescription string, longDescription string, data any) *Command {
	return &Command{
		Group: newGroup(shortDescription, longDescription, data),
		Name:  name,
	}
}

func (c *Command) scanSubcommandHandler(parentg *Group) scanHandler {
	f := func(realval reflect.Value, sfield *reflect.StructField) (bool, error) {
		mtag := newMultiTag(string(sfield.Tag))

		if err := mtag.Parse(); err != nil {
			return true, err
		}
		if p := c.parser(); p != nil {
			p.normalizeStructTag(&mtag)
		}

		positional := mtag.Get(FlagTagPositionalArgs)

		if len(positional) != 0 {
			stype := realval.Type()

			for i := 0; i < stype.NumField(); i++ {
				field := stype.Field(i)

				m := newMultiTag((string(field.Tag)))

				if err := m.Parse(); err != nil {
					return true, err
				}
				if p := c.parser(); p != nil {
					p.normalizeStructTag(&m)
				}

				name := m.Get(FlagTagPositionalArgName)

				if len(name) == 0 {
					name = field.Name
				}
				nameI18n := m.Get(FlagTagArgNameI18n)
				descriptionI18n := m.Get(FlagTagArgDescriptionI18n)
				if descriptionI18n == "" {
					descriptionI18n = m.Get(FlagTagDescriptionI18n)
				}
				completionHint, err := parseCompletionHint(m.Get(FlagTagCompletion), field.Name)
				if err != nil {
					return true, err
				}

				required := -1
				requiredMaximum := -1
				delimiter := parserTagListDelimiter(c.parser())
				def, err := collectTagValues(m, FlagTagDefault, FlagTagDefaults, field.Name, delimiter)
				if err != nil {
					return true, err
				}

				sreq := m.Get(FlagTagRequired)

				if sreq != "" {
					required = 1

					rng := strings.SplitN(sreq, "-", 2)

					if len(rng) > 1 {
						if preq, err := strconv.ParseInt(rng[0], 10, 32); err == nil {
							required = int(preq)
						}

						if preq, err := strconv.ParseInt(rng[1], 10, 32); err == nil {
							requiredMaximum = int(preq)
						}
					} else {
						if preq, err := strconv.ParseInt(sreq, 10, 32); err == nil {
							required = int(preq)
						}
					}
				}

				arg := &Arg{
					Name:               name,
					NameI18nKey:        nameI18n,
					Description:        m.Get(FlagTagDescription),
					DescriptionI18nKey: descriptionI18n,
					Default:            def,
					completionHint:     completionHint,
					Required:           required,
					RequiredMaximum:    requiredMaximum,

					value: realval.Field(i),
					tag:   m,
					cmd:   c,
				}

				c.args = append(c.args, arg)

				argsRequired, _, err := parseStructBoolTag(mtag, FlagTagRequired, sfield.Name)
				if err != nil {
					return true, err
				}
				if argsRequired {
					c.ArgsRequired = true
				}
			}

			return true, nil
		}

		subcommand := mtag.Get(FlagTagCommand)

		if len(subcommand) != 0 {
			var ptrval reflect.Value

			if realval.Kind() == reflect.Ptr {
				ptrval = realval

				if ptrval.IsNil() {
					ptrval.Set(reflect.New(ptrval.Type().Elem()))
				}
			} else {
				ptrval = realval.Addr()
			}

			shortDescription := mtag.Get(FlagTagDescription)
			longDescription := mtag.Get(FlagTagLongDescription)
			shortDescriptionI18n := mtag.Get(FlagTagDescriptionI18n)
			if shortDescriptionI18n == "" {
				shortDescriptionI18n = mtag.Get(FlagTagCommandI18n)
			}
			longDescriptionI18n := mtag.Get(FlagTagLongDescriptionI18n)
			iniGroup := mtag.Get(FlagTagIniGroup)
			if iniGroup == "" {
				iniGroup = mtag.Get(FlagTagIniName)
			}

			if (shortDescriptionI18n != "" || longDescriptionI18n != "") && iniGroup == "" {
				return true, newErrorf(
					ErrInvalidTag,
					"command `%s' uses localized description tags and must define `%s' for a stable INI section name",
					subcommand,
					FlagTagIniGroup,
				)
			}

			delimiter := parserTagListDelimiter(c.parser())
			aliases, err := collectTagValues(mtag, FlagTagAlias, FlagTagAliases, sfield.Name, delimiter)
			if err != nil {
				return true, err
			}

			subcommandsOptional, _, err := parseStructBoolTag(mtag, FlagTagSubCommandsOptional, sfield.Name)
			if err != nil {
				return true, err
			}

			passAfterNonOption, _, err := parseStructBoolTag(mtag, FlagTagPassAfterNonOption, sfield.Name)
			if err != nil {
				return true, err
			}
			order := 0
			if rawOrder := mtag.Get(FlagTagOrder); rawOrder != "" {
				parsedOrder, convErr := strconv.Atoi(rawOrder)
				if convErr != nil {
					return true, newErrorf(
						ErrInvalidTag,
						"invalid integer value `%s' for tag `%s' on field `%s'",
						rawOrder,
						FlagTagOrder,
						sfield.Name,
					)
				}
				order = parsedOrder
			}

			hidden, _, err := parseStructBoolTag(mtag, FlagTagHidden, sfield.Name)
			if err != nil {
				return true, err
			}
			immediate, _, err := parseStructBoolTag(mtag, FlagTagImmediate, sfield.Name)
			if err != nil {
				return true, err
			}

			subc, err := c.AddCommand(subcommand, shortDescription, longDescription, ptrval.Interface())

			if err != nil {
				return true, err
			}

			subc.Hidden = hidden
			subc.Immediate = immediate
			subc.IniName = iniGroup
			subc.CommandGroup = mtag.Get(FlagTagCommandGroup)
			subc.ShortDescriptionI18nKey = shortDescriptionI18n
			subc.LongDescriptionI18nKey = longDescriptionI18n
			subc.Order = order

			if subcommandsOptional {
				subc.SubcommandsOptional = true
			}

			if len(aliases) > 0 {
				subc.Aliases = aliases
			}

			if passAfterNonOption {
				subc.PassAfterNonOption = true
			}

			if p := c.parser(); p != nil {
				p.invalidateLookupCache()
			}

			return true, nil
		}

		return parentg.scanSubGroupHandler(realval, sfield)
	}

	return f
}

func (c *Command) scan() error {
	return c.scanType(c.scanSubcommandHandler(c.Group))
}

func (c *Command) checkForDuplicateFlagsInScope() *Error {
	shortNames := make(map[rune]*Option)
	longNames := make(map[string]*Option)

	for _, cmd := range c.optionScopeCommands() {
		if err := addDuplicateFlagScope(cmd.Group, shortNames, longNames); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) optionScopeCommands() []*Command {
	var reversed []*Command

	for cmd := c; cmd != nil; {
		reversed = append(reversed, cmd)

		parent, ok := cmd.parent.(*Command)
		if !ok {
			break
		}
		cmd = parent
	}

	out := make([]*Command, 0, len(reversed))
	for idx := len(reversed) - 1; idx >= 0; idx-- {
		out = append(out, reversed[idx])
	}

	return out
}

func (c *Command) eachOption(f func(*Command, *Group, *Option)) {
	c.eachCommand(func(c *Command) {
		c.eachGroup(func(g *Group) {
			for _, option := range g.options {
				f(c, g, option)
			}
		})
	})
}

func (c *Command) sortedOptionsForGroup(g *Group) []*Option {
	return g.sortedOptionsForDisplay()
}

func (c *Command) eachCommand(f func(*Command)) {
	f(c)

	for _, cc := range c.commands {
		cc.eachCommand(f)
	}
}

func (c *Command) eachActiveGroup(f func(cc *Command, g *Group)) {
	c.eachGroup(func(g *Group) {
		f(c, g)
	})

	if c.Active != nil {
		c.Active.eachActiveGroup(f)
	}
}

func (c *Command) addHelpGroups(showHelp func() error, showVersion func() error) {
	if !c.hasBuiltinHelpGroup {
		c.addHelpGroup(showHelp, showVersion)
		c.hasBuiltinHelpGroup = true
	}

	for _, cc := range c.commands {
		cc.addHelpGroups(showHelp, showVersion)
	}
}

func (c *Command) needsHelpGroups() bool {
	if !c.hasBuiltinHelpGroup {
		return true
	}

	for _, cc := range c.commands {
		if cc.needsHelpGroups() {
			return true
		}
	}

	return false
}

func (c *Command) makeLookup() lookup {
	if p := c.parser(); p != nil && c.lookupCacheValid && c.lookupCacheGeneration == p.lookupGeneration {
		return c.lookupCache
	}

	ret := lookup{
		shortNames: make(map[string]*Option),
		longNames:  make(map[string]*Option),
		commands:   make(map[string]*Command),
	}

	parent := c.parent

	var parents []*Command

	for parent != nil {
		if cmd, ok := parent.(*Command); ok {
			parents = append(parents, cmd)
			parent = cmd.parent
		} else {
			parent = nil
		}
	}

	for i := len(parents) - 1; i >= 0; i-- {
		parents[i].fillLookup(&ret, true)
	}

	c.fillLookup(&ret, false)

	if p := c.parser(); p != nil {
		c.lookupCache = ret
		c.lookupCacheGeneration = p.lookupGeneration
		c.lookupCacheValid = true
	}

	return ret
}

func (c *Command) parser() *Parser {
	var parent = c.parent

	for parent != nil {
		switch v := parent.(type) {
		case *Parser:
			return v
		case *Command:
			parent = v.parent
		case *Group:
			parent = v.parent
		default:
			return nil
		}
	}

	return nil
}

func (c *Command) fillLookup(ret *lookup, onlyOptions bool) {
	c.eachGroup(func(g *Group) {
		for _, option := range g.options {
			if option.ShortName != 0 {
				ret.shortNames[string(option.ShortName)] = option
			}
			for _, shortAlias := range option.ShortAliases {
				ret.shortNames[string(shortAlias)] = option
			}

			if len(option.LongName) > 0 {
				ret.longNames[option.LongNameWithNamespace()] = option
			}
			for _, longAlias := range option.LongAliasesWithNamespace() {
				ret.longNames[longAlias] = option
			}
		}
	})

	if onlyOptions {
		return
	}

	for _, subcommand := range c.commands {
		ret.commands[subcommand.Name] = subcommand

		for _, a := range subcommand.Aliases {
			ret.commands[a] = subcommand
		}
	}
}

func (c *Command) groupByName(name string) *Group {
	if grp := c.Group.groupByName(name); grp != nil {
		return grp
	}

	for _, subc := range c.commands {
		for _, commandName := range subc.iniLookupNames() {
			prefix := commandName + "."

			if strings.HasPrefix(name, prefix) {
				if grp := subc.groupByName(name[len(prefix):]); grp != nil {
					return grp
				}
			} else if name == commandName {
				return subc.Group
			}
		}
	}

	return nil
}

func (c *Command) iniSectionName() string {
	if c.Group != nil && c.IniName != "" {
		return c.IniName
	}

	return c.Name
}

func (c *Command) iniLookupNames() []string {
	ret := []string{c.Name}

	if name := c.iniSectionName(); name != "" && name != c.Name {
		ret = append(ret, name)
	}

	return ret
}

func (c *Command) sortedVisibleCommands() []*Command {
	ret := c.visibleCommands()
	p := c.parser()
	if p == nil {
		sort.Slice(ret, func(i, j int) bool {
			return ret[i].Name < ret[j].Name
		})
		return ret
	}
	if !p.shouldSortCommandsForDisplay(ret) {
		return ret
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return p.compareCommands(ret[i], ret[j]) < 0
	})

	return ret
}

func (p *Parser) shouldSortCommandsForDisplay(commands []*Command) bool {
	if p.commandSort != CommandSortByDeclaration {
		return true
	}

	for _, cmd := range commands {
		if cmd.Order != 0 {
			return true
		}
	}

	return false
}

func (p *Parser) compareCommands(a *Command, b *Command) int {
	ab := orderBucket(a.Order)
	bb := orderBucket(b.Order)
	if ab != bb {
		return compareInt(ab, bb)
	}

	if a.Order != b.Order {
		// Higher order first within same bucket.
		return compareInt(b.Order, a.Order)
	}

	switch p.commandSort {
	case CommandSortByNameDesc:
		return compareString(b.Name, a.Name)
	case CommandSortByNameAsc:
		return compareString(a.Name, b.Name)
	default:
		return 0
	}
}

func (c *Command) visibleCommands() []*Command {
	ret := make([]*Command, 0, len(c.commands))

	for _, cmd := range c.commands {
		if !cmd.Hidden {
			ret = append(ret, cmd)
		}
	}

	return ret
}

func (c *Command) match(name string) bool {
	if c.Name == name {
		return true
	}

	return slices.Contains(c.Aliases, name)
}

func (c *Command) hasHelpOptions() bool {
	ret := false

	c.eachGroup(func(g *Group) {
		if g.isBuiltinHelp {
			return
		}

		for _, opt := range g.options {
			if opt.showInHelp() {
				ret = true
			}
		}
	})

	return ret
}

func (c *Command) fillParseState(s *parseState) {
	s.positional = make([]*Arg, len(c.args))
	copy(s.positional, c.args)

	s.lookup = c.makeLookup()
	s.command = c
}
