// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// A Parser provides command line option parsing. It can contain several
// option groups each with their own set of options.
type Parser struct {
	// Internal parser scan/setup error returned by Parse/ParseArgs.
	internalError error

	// Embedded, see Command for more information
	*Command

	// UnknownOptionsHandler is a function which gets called when the parser
	// encounters an unknown option. The function receives the unknown option
	// name, a SplitArgument which specifies its value if set with an argument
	// separator, and the remaining command line arguments.
	// It should return a new list of remaining arguments to continue parsing,
	// or an error to indicate a parse failure.
	UnknownOptionHandler func(option string, arg SplitArgument, args []string) ([]string, error)

	// CompletionHandler is a function gets called to handle the completion of
	// items. By default, the items are printed and the application is exited.
	// You can override this default behavior by specifying a custom CompletionHandler.
	CompletionHandler func(items []Completion)

	// CommandHandler is a function that gets called to handle execution of a
	// command. By default, the command will simply be executed. This can be
	// overridden to perform certain actions (such as applying global flags)
	// just before the command is executed. Note that if you override the
	// handler it is your responsibility to call the command.Execute function.
	//
	// The command passed into CommandHandler may be nil in case there is no
	// command to be executed when parsing has finished.
	CommandHandler func(command Commander, args []string) error

	// Type rank order used by OptionSortByType.
	optionTypeRank map[OptionTypeClass]int

	// Active struct-tag mapping used while scanning option metadata.
	flagTags FlagTags

	// A usage string to be displayed in the help message.
	Usage string

	// NamespaceDelimiter separates group namespaces and option long names
	NamespaceDelimiter string

	// EnvNamespaceDelimiter separates group env namespaces and env keys
	EnvNamespaceDelimiter string

	// EnvPrefix prepends all resolved environment variable keys.
	EnvPrefix string

	// TerminalTitle overrides terminal title text when SetTerminalTitle is enabled.
	// If empty, parser Name is used.
	TerminalTitle string

	// Monotonic generation used to invalidate cached lookup maps.
	lookupGeneration uint64

	// Option flags changing the behavior of the parser.
	Options Options

	// TagListDelimiter splits values for list-based struct tags such as
	// defaults/choices/aliases.
	TagListDelimiter rune

	// Active option sorting mode for grouped option presentation.
	optionSort OptionSortMode
}

// SplitArgument represents the argument value of an option that was passed using
// an argument separator.
type SplitArgument interface {
	// String returns the option's value as a string, and a boolean indicating
	// if the option was present.
	Value() (string, bool)
}

type strArgument struct {
	value   string
	present bool
}

func (s strArgument) Value() (string, bool) {
	if !s.present {
		return "", false
	}

	return s.value, true
}

// Options provides parser options that change the behavior of the option
// parser.
type Options uint

// OptionSortMode configures how options are ordered within each group block.
type OptionSortMode uint8

const (
	// OptionSortByDeclaration keeps original declaration order.
	OptionSortByDeclaration OptionSortMode = iota
	// OptionSortByNameAsc sorts by option name ascending.
	OptionSortByNameAsc
	// OptionSortByNameDesc sorts by option name descending.
	OptionSortByNameDesc
	// OptionSortByType sorts by configured type rank, then by name.
	OptionSortByType
)

// OptionTypeClass groups option value types for type-based sorting.
type OptionTypeClass uint8

const (
	// OptionTypeBool classifies boolean option values.
	OptionTypeBool OptionTypeClass = iota
	// OptionTypeNumber classifies integer/float option values.
	OptionTypeNumber
	// OptionTypeString classifies string option values.
	OptionTypeString
	// OptionTypeDuration classifies time.Duration option values.
	OptionTypeDuration
	// OptionTypeCollection classifies slice/map/array option values.
	OptionTypeCollection
	// OptionTypeCustom classifies all remaining option value types.
	OptionTypeCustom
)

const (
	// None indicates no options.
	None Options = 0

	// HelpFlag adds a default Help Options group to the parser containing
	// -h and --help options. When either -h or --help is specified on the
	// command line, the parser will return the special error of type
	// ErrHelp. When PrintErrors is also specified, then the help message
	// will also be automatically printed to os.Stdout unless PrintHelpOnStderr
	// is set.
	HelpFlag = 1 << iota

	// PassDoubleDash passes all arguments after a double dash, --, as
	// remaining command line arguments (i.e. they will not be parsed for
	// flags).
	PassDoubleDash

	// IgnoreUnknown ignores any unknown options and passes them as
	// remaining command line arguments instead of generating an error.
	IgnoreUnknown

	// PrintErrors prints any errors which occurred during parsing to
	// os.Stderr. In the special case of ErrHelp, the message will be printed
	// to os.Stdout unless PrintHelpOnStderr is set.
	PrintErrors

	// PrintHelpOnStderr routes built-in help output (ErrHelp) to os.Stderr
	// when PrintErrors is enabled.
	PrintHelpOnStderr

	// PrintErrorsOnStdout routes non-help parse errors to os.Stdout
	// when PrintErrors is enabled.
	PrintErrorsOnStdout

	// PassAfterNonOption passes all arguments after the first non option
	// as remaining command line arguments. This is equivalent to strict
	// POSIX processing.
	PassAfterNonOption

	// AllowBoolValues allows a user to assign true/false to a boolean value
	// rather than raising an error stating it cannot have an argument.
	AllowBoolValues

	// DefaultsIfEmpty applies tag/env defaults only to options whose current
	// values are empty. This keeps pre-populated option values intact unless
	// they were explicitly set on the command line.
	DefaultsIfEmpty

	// KeepDescriptionWhitespace keeps leading/trailing whitespace in
	// help-rendered descriptions instead of trimming each line before wrapping.
	// This is useful for preserving indentation in lists and code examples.
	KeepDescriptionWhitespace

	// EnvProvisioning auto-generates env keys from long option names when an
	// option does not define an explicit `env` tag. Generated keys are
	// uppercased and punctuation is replaced with underscores.
	EnvProvisioning

	// ShowCommandAliases appends command aliases in the built-in help command
	// list for commands without short descriptions.
	ShowCommandAliases

	// ShowRepeatableInHelp appends a repeatable marker to option descriptions
	// in built-in help output for collection options (slice/map).
	ShowRepeatableInHelp

	// SetTerminalTitle updates terminal window title during ParseArgs using
	// TerminalTitle or parser Name.
	SetTerminalTitle

	// Default is a convenient default set of options which should cover
	// most of the uses of the flags package.
	Default = HelpFlag | PrintErrors | PassDoubleDash
)

type parseState struct {
	lookup lookup
	err    error

	command    *Command
	arg        string
	args       []string
	retargs    []string
	positional []*Arg
}

// Parse is a convenience function to parse command line options with default
// settings. The provided data is a pointer to a struct representing the
// default option group (named "Application Options"). For more control, use
// flags.NewParser.
func Parse(data any) ([]string, error) {
	return NewParser(data, Default).Parse()
}

// ParseArgs is a convenience function to parse command line options with default
// settings. The provided data is a pointer to a struct representing the
// default option group (named "Application Options"). The args argument is
// the list of command line arguments to parse. If you just want to parse the
// default program command line arguments (i.e. os.Args), then use flags.Parse
// instead. For more control, use flags.NewParser.
func ParseArgs(data any, args []string) ([]string, error) {
	return NewParser(data, Default).ParseArgs(args)
}

// NewParser creates a new parser. It uses os.Args[0] as the application
// name and then calls Parser.NewNamedParser (see Parser.NewNamedParser for
// more details). The provided data is a pointer to a struct representing the
// default option group (named "Application Options"), or nil if the default
// group should not be added. The options parameter specifies a set of options
// for the parser.
func NewParser(data any, options Options) *Parser {
	p := NewNamedParser(path.Base(os.Args[0]), options)

	if data != nil {
		g, err := p.AddGroup("Application Options", "", data)

		if err == nil {
			g.parent = p
		}

		p.internalError = err
	}

	return p
}

// NewNamedParser creates a new parser. The appname is used to display the
// executable name in the built-in help message. Option groups and commands can
// be added to this parser by using AddGroup and AddCommand.
func NewNamedParser(appname string, options Options) *Parser {
	p := &Parser{
		Command:               newCommand(appname, "", "", nil),
		Options:               options,
		NamespaceDelimiter:    ".",
		EnvNamespaceDelimiter: "_",
		lookupGeneration:      1,
		flagTags:              NewFlagTags(),
		optionSort:            OptionSortByDeclaration,
		optionTypeRank:        defaultOptionTypeRank(),
		TagListDelimiter:      ';',
	}

	p.parent = p

	return p
}

func (p *Parser) invalidateLookupCache() {
	p.lookupGeneration++
}

// SetTagPrefix configures a common prefix for all struct tags used by the parser.
// It rescans already attached top-level groups and commands.
func (p *Parser) SetTagPrefix(prefix string) error {
	return p.SetFlagTags(NewFlagTagsWithPrefix(prefix))
}

// SetFlagTags configures custom struct tag names used by the parser.
// It rescans already attached top-level groups and commands.
func (p *Parser) SetFlagTags(tags FlagTags) error {
	p.flagTags = tags.withDefaults()
	return p.rebuildTree()
}

// SetEnvPrefix configures a global prefix for all environment variable keys.
// For example, with prefix "MY_APP" and delimiter "_", env key "PORT" becomes
// "MY_APP_PORT", and grouped keys become "MY_APP_<GROUP>_<KEY>".
func (p *Parser) SetEnvPrefix(prefix string) {
	p.EnvPrefix = prefix
}

// SetTagListDelimiter sets delimiter for list-based struct tags such as
// defaults/choices/aliases and rescans attached groups/commands.
func (p *Parser) SetTagListDelimiter(delimiter rune) error {
	if delimiter == 0 {
		return errors.New("tag list delimiter cannot be NUL")
	}

	p.TagListDelimiter = delimiter

	return p.rebuildTree()
}

func (p *Parser) normalizeStructTag(mtag *multiTag) {
	c := mtag.cached()

	alias := map[string]string{
		p.flagTags.Short:               FlagTagShort,
		p.flagTags.Long:                FlagTagLong,
		p.flagTags.Required:            FlagTagRequired,
		p.flagTags.Description:         FlagTagDescription,
		p.flagTags.LongDescription:     FlagTagLongDescription,
		p.flagTags.NoFlag:              FlagTagNoFlag,
		p.flagTags.Optional:            FlagTagOptional,
		p.flagTags.OptionalValue:       FlagTagOptionalValue,
		p.flagTags.Order:               FlagTagOrder,
		p.flagTags.Default:             FlagTagDefault,
		p.flagTags.Defaults:            FlagTagDefaults,
		p.flagTags.DefaultMask:         FlagTagDefaultMask,
		p.flagTags.Env:                 FlagTagEnv,
		p.flagTags.AutoEnv:             FlagTagAutoEnv,
		p.flagTags.EnvDelim:            FlagTagEnvDelim,
		p.flagTags.ValueName:           FlagTagValueName,
		p.flagTags.Choice:              FlagTagChoice,
		p.flagTags.Choices:             FlagTagChoices,
		p.flagTags.Hidden:              FlagTagHidden,
		p.flagTags.Base:                FlagTagBase,
		p.flagTags.IniName:             FlagTagIniName,
		p.flagTags.NoIni:               FlagTagNoIni,
		p.flagTags.Group:               FlagTagGroup,
		p.flagTags.Namespace:           FlagTagNamespace,
		p.flagTags.EnvNamespace:        FlagTagEnvNamespace,
		p.flagTags.Command:             FlagTagCommand,
		p.flagTags.SubCommandsOptional: FlagTagSubCommandsOptional,
		p.flagTags.Alias:               FlagTagAlias,
		p.flagTags.Aliases:             FlagTagAliases,
		p.flagTags.PositionalArgs:      FlagTagPositionalArgs,
		p.flagTags.PositionalArgName:   FlagTagPositionalArgName,
		p.flagTags.KeyValueDelimiter:   FlagTagKeyValueDelimiter,
		p.flagTags.PassAfterNonOption:  FlagTagPassAfterNonOption,
		p.flagTags.Unquote:             FlagTagUnquote,
		p.flagTags.Terminator:          FlagTagTerminator,
	}

	for source, target := range alias {
		if source == "" || source == target {
			continue
		}

		values, ok := c[source]
		if !ok {
			continue
		}

		if _, exists := c[target]; !exists {
			c[target] = values
		}
	}
}

// SetOptionSort configures option order mode for grouped option presentation.
func (p *Parser) SetOptionSort(mode OptionSortMode) {
	p.optionSort = mode
}

// SetOptionTypeOrder customizes type rank used by OptionSortByType.
func (p *Parser) SetOptionTypeOrder(order []OptionTypeClass) error {
	rank, err := buildOptionTypeRank(order)
	if err != nil {
		return err
	}
	p.optionTypeRank = rank
	return nil
}

type groupSpec struct {
	data             any
	shortDescription string
	longDescription  string
	namespace        string
	envNamespace     string
	hidden           bool
}

type commandSpec struct {
	data                any
	name                string
	shortDescription    string
	longDescription     string
	aliases             []string
	subcommandsOptional bool
	passAfterNonOption  bool
	hidden              bool
}

func (p *Parser) rebuildTree() error {
	groups := make([]groupSpec, 0, len(p.groups))
	commands := make([]commandSpec, 0, len(p.commands))
	rootOptions := append([]*Option(nil), p.options...)

	for _, g := range p.groups {
		if g.isBuiltinHelp {
			continue
		}

		groups = append(groups, groupSpec{
			shortDescription: g.ShortDescription,
			longDescription:  g.LongDescription,
			namespace:        g.Namespace,
			envNamespace:     g.EnvNamespace,
			hidden:           g.Hidden,
			data:             g.data,
		})
	}

	for _, c := range p.commands {
		commands = append(commands, commandSpec{
			name:                c.Name,
			shortDescription:    c.ShortDescription,
			longDescription:     c.LongDescription,
			aliases:             append([]string(nil), c.Aliases...),
			subcommandsOptional: c.SubcommandsOptional,
			passAfterNonOption:  c.PassAfterNonOption,
			hidden:              c.Hidden,
			data:                c.data,
		})
	}

	p.groups = nil
	p.commands = nil
	p.options = rootOptions
	p.args = nil
	p.Active = nil
	p.hasBuiltinHelpGroup = false
	p.lookupCacheValid = false

	for _, g := range groups {
		ng, err := p.AddGroup(g.shortDescription, g.longDescription, g.data)
		if err != nil {
			return fmt.Errorf("failed to rescan group %q: %w", g.shortDescription, err)
		}
		ng.Namespace = g.namespace
		ng.EnvNamespace = g.envNamespace
		ng.Hidden = g.hidden
	}

	for _, c := range commands {
		nc, err := p.AddCommand(c.name, c.shortDescription, c.longDescription, c.data)
		if err != nil {
			return fmt.Errorf("failed to rescan command %q: %w", c.name, err)
		}
		nc.Aliases = c.aliases
		nc.SubcommandsOptional = c.subcommandsOptional
		nc.PassAfterNonOption = c.passAfterNonOption
		nc.Hidden = c.hidden
	}

	p.invalidateLookupCache()
	return nil
}

// Parse parses the command line arguments from os.Args using Parser.ParseArgs.
// For more detailed information see ParseArgs.
func (p *Parser) Parse() ([]string, error) {
	return p.ParseArgs(os.Args[1:])
}

// ParseArgs parses the command line arguments according to the option groups that
// were added to the parser. On successful parsing of the arguments, the
// remaining, non-option, arguments (if any) are returned. The returned error
// indicates a parsing error and can be used with PrintError to display
// contextual information on where the error occurred exactly.
//
// When the common help group has been added (AddHelp) and either -h or --help
// was specified in the command line arguments, a help message will be
// automatically printed if the PrintErrors option is enabled.
// Furthermore, the special error type ErrHelp is returned.
// It is up to the caller to exit the program if so desired.
func (p *Parser) ParseArgs(args []string) ([]string, error) {
	if p.internalError != nil {
		return nil, p.internalError
	}

	p.eachOption(func(_ *Command, _ *Group, option *Option) {
		option.clearReferenceBeforeSet = true
		if !option.defaultLiteralInitialized {
			option.updateDefaultLiteral()
			option.defaultLiteralInitialized = true
		}
	})

	// Add built-in help group to all commands if necessary
	if (p.Options & HelpFlag) != None {
		p.addHelpGroups(p.showBuiltinHelp)
	}

	compval := os.Getenv("GO_FLAGS_COMPLETION")

	if len(compval) != 0 {
		comp := &completion{parser: p}
		items := comp.complete(args)

		if p.CompletionHandler != nil {
			p.CompletionHandler(items)
		} else {
			comp.print(items, compval == "verbose")
			os.Exit(0)
		}

		return nil, nil
	}

	p.applyTerminalTitle()

	s := &parseState{
		args:    args,
		retargs: make([]string, 0, len(args)),
	}

	p.fillParseState(s)

	for !s.eof() {
		var err error
		arg := s.pop()

		// When PassDoubleDash is set and we encounter a --, then
		// simply append all the rest as arguments and break out
		if (p.Options&PassDoubleDash) != None && arg == "--" {
			if err = s.addArgs(s.args...); err != nil {
				break
			}
			break
		}

		if !argumentIsOption(arg) {
			if ((p.Options&PassAfterNonOption) != None || s.command.PassAfterNonOption) && s.lookup.commands[arg] == nil {
				// If PassAfterNonOption is set then all remaining arguments
				// are considered positional
				if err = s.addArgs(s.arg); err != nil {
					break
				}

				if err = s.addArgs(s.args...); err != nil {
					break
				}

				break
			}

			// Note: this also sets s.err, so we can just check for
			// nil here and use s.err later
			if p.parseNonOption(s) != nil {
				break
			}

			continue
		}

		prefix, optname, islong := stripOptionPrefix(arg)
		optname, _, argument, hasArgument := splitOption(prefix, optname, islong)

		if islong {
			err = p.parseLong(s, optname, argument, hasArgument)
		} else {
			err = p.parseShort(s, optname, argument, hasArgument)
		}

		if err != nil {
			ignoreUnknown := (p.Options & IgnoreUnknown) != None
			parseErr := wrapError(err)

			if parseErr.Type != ErrUnknownFlag || (!ignoreUnknown && p.UnknownOptionHandler == nil) {
				s.err = parseErr
				break
			}

			if ignoreUnknown {
				if err = s.addArgs(arg); err != nil {
					s.err = err
					break
				}
			} else if p.UnknownOptionHandler != nil {
				modifiedArgs, err := p.UnknownOptionHandler(optname, strArgument{
					value:   argument,
					present: hasArgument,
				}, s.args)

				if err != nil {
					s.err = err
					break
				}

				s.args = modifiedArgs
			}
		}
	}

	if s.err == nil {
		p.eachOption(func(_ *Command, _ *Group, option *Option) {
			if (p.Options&DefaultsIfEmpty) != None && !option.isEmpty() {
				return
			}

			err := option.clearDefault()
			if err != nil {
				if _, ok := err.(*Error); !ok {
					err = p.marshalError(option, err)
				}
				s.err = err
			}
		})

		if reqErr := s.checkRequired(p); reqErr != nil {
			s.err = reqErr
		}
	}

	var reterr error

	if s.err != nil {
		reterr = s.err
	} else if len(s.command.commands) != 0 && !s.command.SubcommandsOptional {
		reterr = s.estimateCommand()
	} else if cmd, ok := s.command.data.(Commander); ok {
		if p.CommandHandler != nil {
			reterr = p.CommandHandler(cmd, s.retargs)
		} else {
			reterr = cmd.Execute(s.retargs)
		}
	} else if p.CommandHandler != nil {
		reterr = p.CommandHandler(nil, s.retargs)
	}

	if reterr != nil {
		var retargs []string

		if ourErr, ok := reterr.(*Error); !ok || ourErr.Type != ErrHelp {
			retargs = append([]string{s.arg}, s.args...)
		} else {
			retargs = s.args
		}

		return retargs, p.printError(reterr)
	}

	return s.retargs, nil
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
				if !option.isSet && option.Required {
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

						reqnames = append(reqnames, "`"+arg.Name+" (at least "+strconv.Itoa(arg.Required)+" "+arguments+")`")
					} else if arg.RequiredMaximum != -1 && arg.value.Len() > arg.RequiredMaximum {
						if arg.RequiredMaximum == 0 {
							reqnames = append(reqnames, "`"+arg.Name+" (zero arguments)`")
						} else {
							var arguments string

							if arg.RequiredMaximum > 1 {
								arguments = "arguments, but got " + strconv.Itoa(arg.value.Len())
							} else {
								arguments = "argument"
							}

							reqnames = append(reqnames, "`"+arg.Name+" (at most "+strconv.Itoa(arg.RequiredMaximum)+" "+arguments+")`")
						}
					}
				} else {
					reqnames = append(reqnames, "`"+arg.Name+"`")
				}
			}

			if len(reqnames) == 0 {
				return nil
			}

			var msg string

			if len(reqnames) == 1 {
				msg = fmt.Sprintf("the required argument %s was not provided", reqnames[0])
			} else {
				msg = fmt.Sprintf("the required arguments %s and %s were not provided",
					strings.Join(reqnames[:len(reqnames)-1], ", "), reqnames[len(reqnames)-1])
			}

			p.err = newError(ErrRequired, msg)
			return p.err
		}

		return nil
	}

	names := make([]string, 0, len(required))

	for _, k := range required {
		names = append(names, "`"+k.String()+"'")
	}

	sort.Strings(names)

	var msg string

	if len(names) == 1 {
		msg = fmt.Sprintf("the required flag %s was not specified", names[0])
	} else {
		msg = fmt.Sprintf("the required flags %s and %s were not specified",
			strings.Join(names[:len(names)-1], ", "), names[len(names)-1])
	}

	p.err = newError(ErrRequired, msg)
	return p.err
}

func (p *parseState) estimateCommand() error {
	commands := p.command.sortedVisibleCommands()
	cmdnames := make([]string, len(commands))

	for i, v := range commands {
		cmdnames[i] = v.Name
	}

	var msg string
	var errtype ErrorType

	if len(p.retargs) != 0 {
		c, l := closestChoice(p.retargs[0], cmdnames)
		msg = fmt.Sprintf("Unknown command `%s'", p.retargs[0])
		errtype = ErrUnknownCommand

		switch {
		case float32(l)/float32(len(c)) < 0.5:
			msg = fmt.Sprintf("%s, did you mean `%s'?", msg, c)
		case len(cmdnames) == 1:
			msg = fmt.Sprintf("%s. You should use the %s command",
				msg,
				cmdnames[0])
		case len(cmdnames) > 1:
			msg = fmt.Sprintf("%s. Please specify one command of: %s or %s",
				msg,
				strings.Join(cmdnames[:len(cmdnames)-1], ", "),
				cmdnames[len(cmdnames)-1])
		}
	} else {
		errtype = ErrCommandRequired

		switch {
		case len(cmdnames) == 1:
			msg = fmt.Sprintf("Please specify the %s command", cmdnames[0])
		case len(cmdnames) > 1:
			msg = fmt.Sprintf("Please specify one command of: %s or %s",
				strings.Join(cmdnames[:len(cmdnames)-1], ", "),
				cmdnames[len(cmdnames)-1])
		}
	}

	return newError(errtype, msg)
}

func (p *Parser) parseOption(s *parseState, _ string, option *Option, canarg bool, argument string, hasArgument bool) (err error) {
	switch {
	case !option.canArgument():
		if hasArgument && (p.Options&AllowBoolValues) == None {
			return newErrorf(ErrNoArgumentForBool, "bool flag `%s' cannot have an argument", option)
		}
		var value *string
		if hasArgument {
			value = &argument
		}
		err = option.Set(value)
	case option.isTerminated():
		if hasArgument {
			return newErrorf(ErrExpectedArgument, "terminated option flag `%s' cannot use inline argument syntax", option)
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
				return newErrorf(ErrExpectedArgument, "expected argument for flag `%s', but got double dash `--'", option)
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
		err = newErrorf(ErrExpectedArgument, "expected argument for flag `%s'", option)
	}

	if err != nil {
		if _, ok := err.(*Error); !ok {
			err = p.marshalError(option, err)
		}
	}

	return err
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
	s := "invalid argument for flag `%s'"

	expected := p.expectedType(option)

	if expected != "" {
		s = s + " (expected " + expected + ")"
	}

	return newErrorf(ErrMarshal, s+": %s",
		option,
		err.Error())
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

	return newErrorf(ErrUnknownFlag, "unknown flag `%s'", name)
}

func (p *Parser) splitShortConcatArg(s *parseState, optname string) (string, *string) {
	c, n := utf8.DecodeRuneInString(optname)

	if n == len(optname) {
		return optname, nil
	}

	first := string(c)

	if option := s.lookup.shortNames[first]; option != nil && option.canArgument() {
		arg := optname[n:]
		return first, &arg
	}

	return optname, nil
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
			return newErrorf(ErrUnknownFlag, "unknown flag `%s'", shortname)
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

		if err := convert(args[0], arg.value, arg.tag); err != nil {
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

func (p *Parser) parseNonOption(s *parseState) error {
	if len(s.positional) > 0 {
		return s.addArgs(s.arg)
	}

	if len(s.command.commands) > 0 && len(s.retargs) == 0 {
		if cmd := s.lookup.commands[s.arg]; cmd != nil {
			s.command.Active = cmd
			cmd.fillParseState(s)

			return nil
		} else if !s.command.SubcommandsOptional {
			if err := s.addArgs(s.arg); err != nil {
				return err
			}
			return newErrorf(ErrUnknownCommand, "Unknown command `%s'", s.arg)
		}
	}

	return s.addArgs(s.arg)
}

func (p *Parser) showBuiltinHelp() error {
	var b bytes.Buffer

	p.WriteHelp(&b)
	return newError(ErrHelp, b.String())
}

func (p *Parser) printError(err error) error {
	if err != nil && (p.Options&PrintErrors) != None {
		_, _ = fmt.Fprintln(p.errorWriter(err), err)
	}

	return err
}

func (p *Parser) errorWriter(err error) io.Writer {
	flagsErr, ok := err.(*Error)

	if ok && flagsErr.Type == ErrHelp {
		if (p.Options & PrintHelpOnStderr) != None {
			return os.Stderr
		}

		return os.Stdout
	}

	if (p.Options & PrintErrorsOnStdout) != None {
		return os.Stdout
	}

	return os.Stderr
}
