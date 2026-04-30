// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"os"
	"path"
	"reflect"
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

	// Optional i18n runtime config and resolvers.
	i18n *i18nState

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

	// Display group assigned to built-in commands in help/docs.
	builtinCommandGroup string

	// Optional i18n key for builtinCommandGroup.
	builtinCommandGroupI18nKey string

	// Cached version metadata (auto-detected and/or overridden).
	versionInfo VersionInfo

	// MaxLongNameLength limits allowed rune length of option `long` names.
	// Zero disables the limit.
	MaxLongNameLength int

	// Monotonic generation used to invalidate cached lookup maps.
	lookupGeneration uint64

	// Option flags changing the behavior of the parser.
	Options Options

	// Configured set of fields rendered by built-in version output.
	versionFields VersionFields

	// Extra spaces added before command option rows in built-in help output.
	commandOptionIndent int

	// Explicit help output width. Zero means unlimited when helpWidthSet is true.
	helpWidth int

	// Built-in command options that have already been attached.
	builtinCommandsAdded Options

	// TagListDelimiter splits values for list-based struct tags such as
	// defaults/choices/aliases.
	TagListDelimiter rune

	// Active color scheme for built-in help rendering.
	helpColorScheme HelpColorScheme

	// Active color scheme for parser errors.
	errorColorScheme ErrorColorScheme

	// Runtime gate for help color output, set per writer in WriteHelp.
	helpColorEnabled bool

	// Active option sorting mode for grouped option presentation.
	optionSort OptionSortMode

	// Active command sorting mode for command presentation in help/docs.
	commandSort CommandSortMode

	// Preferred rendering style for flags in help/doc output.
	helpFlagStyle RenderStyle

	// Preferred rendering style for env placeholders in help/doc output.
	helpEnvStyle RenderStyle

	// Tracks whether helpWidth was explicitly configured.
	helpWidthSet bool

	// Indicates that post-scan configurators should be applied before parse.
	configDirty bool

	// Prevents recursive configurator execution.
	configuring bool

	// Set by built-in version option handler during parse.
	versionRequested bool

	// Set when any immediate option/group is requested during parse.
	immediateRequested bool
}

// SplitArgument represents the argument value of an option that was passed using
// an argument separator.
type SplitArgument interface {
	// String returns the option's value as a string, and a boolean indicating
	// if the option was present.
	Value() (string, bool)
}

// Configurer can be implemented by option/group/command data structs
// to programmatically adjust parser metadata after tag scanning.
//
// ConfigureFlags is called before parsing when parser topology changes
// (for example after AddGroup/AddCommand, SetTagPrefix, SetFlagTags).
type Configurer interface {
	ConfigureFlags(parser *Parser) error
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

// CommandSortMode configures how commands are ordered in help/docs.
type CommandSortMode uint8

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

const (
	// CommandSortByDeclaration keeps original declaration order.
	CommandSortByDeclaration CommandSortMode = iota
	// CommandSortByNameAsc sorts by command name ascending.
	CommandSortByNameAsc
	// CommandSortByNameDesc sorts by command name descending.
	CommandSortByNameDesc
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

	// VersionFlag adds built-in -v/--version option to the Help Options group.
	// When specified, parser returns ErrVersion and the version message.
	VersionFlag

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

	// PrintHelpOnInputErrors prints built-in help before common user-input
	// parser errors (for example required/unknown flags or command issues).
	PrintHelpOnInputErrors

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

	// RequiredFromValues treats pre-populated non-empty option values as
	// satisfying `required` checks, even when the option was not explicitly
	// set by CLI/env/default processing.
	RequiredFromValues

	// KeepDescriptionWhitespace keeps leading/trailing whitespace in
	// help-rendered descriptions instead of trimming each line before wrapping.
	// This is useful for preserving indentation in lists and code examples.
	KeepDescriptionWhitespace

	// EnvProvisioning auto-generates env keys from long option names when an
	// option does not define an explicit `env` tag. Generated keys are
	// uppercased and punctuation is replaced with underscores.
	EnvProvisioning

	// ShowCommandAliases forces alias display in the built-in "Available commands"
	// list even when a command has no short description. Without this flag,
	// aliases are shown only for commands that already render a short description.
	ShowCommandAliases

	// ShowRepeatableInHelp appends a repeatable marker to option descriptions
	// (slice/map options) and repeatable positional arguments (slice).
	ShowRepeatableInHelp

	// ShowChoiceListInHelp forces rendering choices as a vertical list
	// in built-in help output.
	ShowChoiceListInHelp

	// AutoShowChoiceListInHelp enables adaptive rendering of choices as
	// a vertical list in built-in help output when available width is tight.
	// ShowChoiceListInHelp has priority and always forces list rendering.
	AutoShowChoiceListInHelp

	// HideEnvInHelp suppresses environment variable placeholders in built-in
	// help output.
	HideEnvInHelp

	// ColorHelp enables ANSI-colored built-in help output.
	ColorHelp

	// ColorErrors enables ANSI-colored parser errors according to error severity.
	ColorErrors

	// SetTerminalTitle updates terminal window title during ParseArgs using
	// TerminalTitle or parser Name.
	SetTerminalTitle

	// HelpCommand adds a built-in `help` command that writes parser help.
	HelpCommand

	// VersionCommand adds a built-in `version` command that writes version info.
	VersionCommand

	// CompletionCommand adds a built-in `completion` command that writes shell
	// completion scripts.
	CompletionCommand

	// DocsCommand adds a built-in `docs` command with format subcommands.
	DocsCommand

	// ConfigCommand adds a built-in `config` command that writes an example INI
	// configuration.
	ConfigCommand

	// DetectShellFlagStyle enables shell-based flag style rendering in help
	// and doc output when no explicit render style is set.
	DetectShellFlagStyle

	// DetectShellEnvStyle enables shell-based env placeholder rendering in
	// help and doc output when no explicit render style is set.
	DetectShellEnvStyle

	// Default is a convenient default set of options which should cover
	// most of the uses of the flags package.
	Default = HelpFlag | PrintErrors | PassDoubleDash

	// HelpCommands enables all built-in help-related commands.
	HelpCommands = HelpCommand | VersionCommand | CompletionCommand | DocsCommand | ConfigCommand

	// ConfiguredValues is a convenience mode for config-first flows
	// (for example YAML/JSON prefill before Parse):
	//   - keep prefilled values intact (DefaultsIfEmpty)
	//   - treat non-empty prefilled values as satisfying `required`
	//     (RequiredFromValues)
	ConfiguredValues = DefaultsIfEmpty | RequiredFromValues
)

const (
	// DefaultMaxLongNameLength is the default maximum length for `long` names.
	// Set parser MaxLongNameLength to 0 to disable this limit.
	DefaultMaxLongNameLength = 32
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
			g.SetShortDescriptionI18nKey("help.group.application_options")
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
		Command:                    newCommand(appname, "", "", nil),
		Options:                    options,
		NamespaceDelimiter:         ".",
		EnvNamespaceDelimiter:      "_",
		MaxLongNameLength:          DefaultMaxLongNameLength,
		lookupGeneration:           1,
		flagTags:                   NewFlagTags(),
		helpColorScheme:            DefaultHelpColorScheme(),
		errorColorScheme:           DefaultErrorColorScheme(),
		helpColorEnabled:           true,
		optionSort:                 OptionSortByDeclaration,
		commandSort:                CommandSortByNameAsc,
		optionTypeRank:             defaultOptionTypeRank(),
		TagListDelimiter:           ';',
		helpFlagStyle:              RenderStyleAuto,
		helpEnvStyle:               RenderStyleAuto,
		versionFields:              VersionFieldsCore,
		builtinCommandGroup:        "Help Commands",
		builtinCommandGroupI18nKey: "help.command_group.help_commands",
		configDirty:                true,
	}

	p.parent = p

	return p
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

// SetHelpColorScheme configures color roles used by built-in help rendering.
func (p *Parser) SetHelpColorScheme(scheme HelpColorScheme) {
	p.helpColorScheme = scheme
}

// SetErrorColorScheme configures color roles used by parser error rendering.
func (p *Parser) SetErrorColorScheme(scheme ErrorColorScheme) {
	p.errorColorScheme = scheme
}

// SetHelpFlagRenderStyle configures how flag tokens are rendered in built-in
// help and doc templates.
func (p *Parser) SetHelpFlagRenderStyle(style RenderStyle) {
	p.helpFlagStyle = style
}

// SetHelpEnvRenderStyle configures how env placeholders are rendered in
// built-in help and doc templates.
func (p *Parser) SetHelpEnvRenderStyle(style RenderStyle) {
	p.helpEnvStyle = style
}

// SetCommandOptionIndent configures extra spaces before command option rows in
// built-in help output. The default is 0, so top-level and command options use
// the same indentation.
func (p *Parser) SetCommandOptionIndent(indent int) error {
	if indent < 0 {
		return ErrNegativeCommandOptionIndent
	}

	p.commandOptionIndent = indent
	return nil
}

// SetHelpWidth configures built-in help output wrapping width. When unset,
// help uses the current terminal width with a fallback of 80 columns. Width 0
// disables wrapping.
func (p *Parser) SetHelpWidth(width int) error {
	if width < 0 {
		return ErrNegativeHelpWidth
	}

	p.helpWidth = width
	p.helpWidthSet = true
	return nil
}

// SetMaxLongNameLength sets the maximum allowed length for option `long` names.
// Value 0 disables the limit. Negative values are rejected.
// Existing parser groups/commands are rescanned so the new rule is applied
// immediately.
func (p *Parser) SetMaxLongNameLength(length int) error {
	if length < 0 {
		return ErrNegativeMaxLongNameLength
	}

	prev := p.MaxLongNameLength
	p.MaxLongNameLength = length

	if err := p.rebuildTree(); err != nil {
		p.MaxLongNameLength = prev
		_ = p.rebuildTree()
		return err
	}

	return nil
}

// SetTagListDelimiter sets delimiter for list-based struct tags such as
// defaults/choices/aliases and rescans attached groups/commands.
func (p *Parser) SetTagListDelimiter(delimiter rune) error {
	if delimiter == 0 {
		return ErrNULTagListDelimiter
	}

	p.TagListDelimiter = delimiter

	return p.rebuildTree()
}

// SetOptionSort configures option order mode for grouped option presentation.
func (p *Parser) SetOptionSort(mode OptionSortMode) {
	p.optionSort = mode
}

// SetCommandSort configures command order mode for help/docs presentation.
func (p *Parser) SetCommandSort(mode CommandSortMode) {
	p.commandSort = mode
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

// Parse parses the command line arguments from os.Args using Parser.ParseArgs.
// For more detailed information see ParseArgs.
func (p *Parser) Parse() ([]string, error) {
	return p.ParseArgs(os.Args[1:])
}

// EnsureBuiltinOptions materializes built-in help/version options (when enabled)
// so they can be discovered and tuned before parsing.
func (p *Parser) EnsureBuiltinOptions() {
	if (p.Options&(HelpFlag|VersionFlag)) != None && p.needsHelpGroups() {
		p.addHelpGroups(p.showBuiltinHelp, p.markVersionRequested)
	}
}

// EnsureBuiltinCommands materializes built-in commands when enabled.
func (p *Parser) EnsureBuiltinCommands() error {
	return p.ensureBuiltinCommands()
}

// SetBuiltinCommandGroup configures the display group used by built-in
// commands in help/docs. Use an empty string to render them without a group.
func (p *Parser) SetBuiltinCommandGroup(group string) {
	p.builtinCommandGroup = group
	p.builtinCommandGroupI18nKey = ""
	for _, commandName := range []string{"help", "version", "completion", "docs", "config"} {
		if cmd := p.Find(commandName); cmd != nil {
			cmd.CommandGroup = group
			cmd.CommandGroupI18nKey = ""
		}
	}
}

// BuiltinHelpOption returns built-in help option when HelpFlag is enabled.
// It materializes built-in options lazily and returns nil when unavailable.
func (p *Parser) BuiltinHelpOption() *Option {
	p.EnsureBuiltinOptions()
	return p.FindOptionByLongName("help")
}

// BuiltinVersionOption returns built-in version option when VersionFlag is enabled.
// It materializes built-in options lazily and returns nil when unavailable.
func (p *Parser) BuiltinVersionOption() *Option {
	p.EnsureBuiltinOptions()
	return p.FindOptionByLongName("version")
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
		return nil, p.printError(p.internalError)
	}

	if err := p.applyConfigurators(); err != nil {
		return nil, p.printError(err)
	}

	// Add built-in help/version group before duplicate validation so their
	// flags cannot silently shadow application or command flags.
	p.EnsureBuiltinOptions()
	if err := p.EnsureBuiltinCommands(); err != nil {
		return nil, p.printError(err)
	}

	if err := p.validateDuplicateFlags(); err != nil {
		return nil, p.printError(err)
	}
	if err := p.validateDuplicateCommands(); err != nil {
		return nil, p.printError(err)
	}

	p.eachOption(func(_ *Command, _ *Group, option *Option) {
		option.clearReferenceBeforeSet = true
		if !option.defaultLiteralInitialized {
			option.updateDefaultLiteral()
			option.defaultLiteralInitialized = true
		}
	})

	p.versionRequested = false
	p.immediateRequested = false

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
				if (p.Options & VersionFlag) == None {
					break
				}
				continue
			}

			if ignoreUnknown {
				if err = s.addArgs(arg); err != nil {
					s.err = err
					if (p.Options & VersionFlag) == None {
						break
					}
					continue
				}
			} else if p.UnknownOptionHandler != nil {
				modifiedArgs, err := p.UnknownOptionHandler(optname, strArgument{
					value:   argument,
					present: hasArgument,
				}, s.args)

				if err != nil {
					s.err = err
					if (p.Options & VersionFlag) == None {
						break
					}
					continue
				}

				s.args = modifiedArgs
			}
		}
	}

	if p.versionRequested {
		if s.err != nil {
			if flagsErr, ok := s.err.(*Error); ok && flagsErr.Type == ErrHelp {
				return nil, p.printError(s.err)
			}
		}

		return nil, p.printError(p.showBuiltinVersion())
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

		if s.err == nil {
			if err := s.applyPositionalDefaults(p, (p.Options&DefaultsIfEmpty) != None); err != nil {
				s.err = err
			}
		}

		if s.err == nil && !p.shouldSkipRequiredValidation() {
			if reqErr := s.checkRequired(p); reqErr != nil {
				s.err = reqErr
			}
		}

		if s.err == nil && !p.shouldSkipRequiredValidation() {
			if relationErr := s.checkOptionRelations(p); relationErr != nil {
				s.err = relationErr
			}
		}
	}

	var reterr error

	if s.err != nil {
		reterr = s.err
	} else if p.shouldSkipCommandExecution() {
		return s.retargs, nil
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

func (p *Parser) normalizeStructTag(mtag *multiTag) {
	c := mtag.cached()

	normalizeTagAlias(c, p.flagTags.Short, FlagTagShort)
	normalizeTagAlias(c, p.flagTags.Long, FlagTagLong)
	normalizeTagAlias(c, p.flagTags.Required, FlagTagRequired)
	normalizeTagAlias(c, p.flagTags.Xor, FlagTagXor)
	normalizeTagAlias(c, p.flagTags.And, FlagTagAnd)
	normalizeTagAlias(c, p.flagTags.Counter, FlagTagCounter)
	normalizeTagAlias(c, p.flagTags.Description, FlagTagDescription)
	normalizeTagAlias(c, p.flagTags.DescriptionI18n, FlagTagDescriptionI18n)
	normalizeTagAlias(c, p.flagTags.LongDescription, FlagTagLongDescription)
	normalizeTagAlias(c, p.flagTags.LongDescriptionI18n, FlagTagLongDescriptionI18n)
	normalizeTagAlias(c, p.flagTags.NoFlag, FlagTagNoFlag)
	normalizeTagAlias(c, p.flagTags.Optional, FlagTagOptional)
	normalizeTagAlias(c, p.flagTags.OptionalValue, FlagTagOptionalValue)
	normalizeTagAlias(c, p.flagTags.Order, FlagTagOrder)
	normalizeTagAlias(c, p.flagTags.Default, FlagTagDefault)
	normalizeTagAlias(c, p.flagTags.Defaults, FlagTagDefaults)
	normalizeTagAlias(c, p.flagTags.DefaultMask, FlagTagDefaultMask)
	normalizeTagAlias(c, p.flagTags.Env, FlagTagEnv)
	normalizeTagAlias(c, p.flagTags.AutoEnv, FlagTagAutoEnv)
	normalizeTagAlias(c, p.flagTags.EnvDelim, FlagTagEnvDelim)
	normalizeTagAlias(c, p.flagTags.ValueName, FlagTagValueName)
	normalizeTagAlias(c, p.flagTags.ValueNameI18n, FlagTagValueNameI18n)
	normalizeTagAlias(c, p.flagTags.Choice, FlagTagChoice)
	normalizeTagAlias(c, p.flagTags.Choices, FlagTagChoices)
	normalizeTagAlias(c, p.flagTags.Completion, FlagTagCompletion)
	normalizeTagAlias(c, p.flagTags.Hidden, FlagTagHidden)
	normalizeTagAlias(c, p.flagTags.Immediate, FlagTagImmediate)
	normalizeTagAlias(c, p.flagTags.Base, FlagTagBase)
	normalizeTagAlias(c, p.flagTags.IniName, FlagTagIniName)
	normalizeTagAlias(c, p.flagTags.IniGroup, FlagTagIniGroup)
	normalizeTagAlias(c, p.flagTags.NoIni, FlagTagNoIni)
	normalizeTagAlias(c, p.flagTags.Group, FlagTagGroup)
	normalizeTagAlias(c, p.flagTags.GroupI18n, FlagTagGroupI18n)
	normalizeTagAlias(c, p.flagTags.Namespace, FlagTagNamespace)
	normalizeTagAlias(c, p.flagTags.EnvNamespace, FlagTagEnvNamespace)
	normalizeTagAlias(c, p.flagTags.Command, FlagTagCommand)
	normalizeTagAlias(c, p.flagTags.CommandI18n, FlagTagCommandI18n)
	normalizeTagAlias(c, p.flagTags.CommandGroup, FlagTagCommandGroup)
	normalizeTagAlias(c, p.flagTags.SubCommandsOptional, FlagTagSubCommandsOptional)
	normalizeTagAlias(c, p.flagTags.Alias, FlagTagAlias)
	normalizeTagAlias(c, p.flagTags.Aliases, FlagTagAliases)
	normalizeTagAlias(c, p.flagTags.LongAlias, FlagTagLongAlias)
	normalizeTagAlias(c, p.flagTags.LongAliases, FlagTagLongAliases)
	normalizeTagAlias(c, p.flagTags.ShortAlias, FlagTagShortAlias)
	normalizeTagAlias(c, p.flagTags.ShortAliases, FlagTagShortAliases)
	normalizeTagAlias(c, p.flagTags.PositionalArgs, FlagTagPositionalArgs)
	normalizeTagAlias(c, p.flagTags.PositionalArgName, FlagTagPositionalArgName)
	normalizeTagAlias(c, p.flagTags.ArgNameI18n, FlagTagArgNameI18n)
	normalizeTagAlias(c, p.flagTags.ArgDescriptionI18n, FlagTagArgDescriptionI18n)
	normalizeTagAlias(c, p.flagTags.KeyValueDelimiter, FlagTagKeyValueDelimiter)
	normalizeTagAlias(c, p.flagTags.PassAfterNonOption, FlagTagPassAfterNonOption)
	normalizeTagAlias(c, p.flagTags.Unquote, FlagTagUnquote)
	normalizeTagAlias(c, p.flagTags.Terminator, FlagTagTerminator)
}

func normalizeTagAlias(tags map[string][]string, source string, target string) {
	if source == "" || source == target {
		return
	}

	values, ok := tags[source]
	if !ok {
		return
	}

	if _, exists := tags[target]; !exists {
		tags[target] = values
	}
}

func (p *Parser) invalidateLookupCache() {
	p.lookupGeneration++
	p.configDirty = true
}

type groupSpec struct {
	data             any
	shortDescription string
	longDescription  string
	namespace        string
	envNamespace     string
	iniName          string
	hidden           bool
}

type commandSpec struct {
	data                any
	name                string
	shortDescription    string
	longDescription     string
	iniName             string
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
			iniName:          g.IniName,
			hidden:           g.Hidden,
			data:             g.data,
		})
	}

	for _, c := range p.commands {
		commands = append(commands, commandSpec{
			name:                c.Name,
			shortDescription:    c.ShortDescription,
			longDescription:     c.LongDescription,
			iniName:             c.IniName,
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
		ng.IniName = g.iniName
		ng.Hidden = g.hidden
	}

	for _, c := range commands {
		if existing := p.Find(c.name); existing != nil && sameCommandData(existing.data, c.data) {
			continue
		}

		nc, err := p.AddCommand(c.name, c.shortDescription, c.longDescription, c.data)
		if err != nil {
			return fmt.Errorf("failed to rescan command %q: %w", c.name, err)
		}
		nc.Aliases = c.aliases
		nc.IniName = c.iniName
		nc.SubcommandsOptional = c.subcommandsOptional
		nc.PassAfterNonOption = c.passAfterNonOption
		nc.Hidden = c.hidden
	}

	p.invalidateLookupCache()
	return nil
}

func sameCommandData(a any, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	if av.Type() != bv.Type() {
		return false
	}

	switch av.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return av.Pointer() == bv.Pointer()
	default:
		if av.Type().Comparable() {
			return a == b
		}
		return false
	}
}
