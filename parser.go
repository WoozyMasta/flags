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
//
// A Parser and its bound option struct are not safe for concurrent use;
// see the package doc's "Concurrency" section.
type Parser struct {
	// Internal parser scan/setup error returned by Parse/ParseArgs.
	internalError error

	// Embedded, see Command for more information
	*Command

	// UnknownOptionHandler is a function which gets called when the parser
	// encounters an unknown option. The function receives the unknown option
	// name, a SplitArgument which specifies its value if set with an argument
	// separator, and the remaining command line arguments.
	// It should return a new list of remaining arguments to continue parsing,
	// or an error to indicate a parse failure.
	UnknownOptionHandler func(option string, arg SplitArgument, args []string) ([]string, error)

	// UnknownCommandHandler is called when an unrecognized command token is
	// encountered. It receives the unknown command name and the remaining
	// unparsed arguments. Return a non-nil error to fail parsing; return nil
	// to add the token to retargs and continue.
	// This handler is invoked whenever an unknown command would produce
	// ErrUnknownCommand (i.e. StrictCommands/StrictSubcommands is active, or
	// SubcommandsOptional is false and no positional slots are available).
	UnknownCommandHandler func(command string, args []string) error

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

	// Raw text blocks rendered around built-in help output.
	helpHeader string
	banner     string
	helpFooter string

	// .env file path used when DotEnv/DotEnvOverride/DotEnvFlags is set.
	dotEnvFile string

	// Long names of the three optional pre-configured dotenv flags.
	dotEnvFileFlagName     string
	dotEnvDisableFlagName  string
	dotEnvOverrideFlagName string

	// Long name of the --config flag when ConfigFlags is enabled.
	configFileFlagName string

	// Default config file path used when ConfigFlags is set but --config is not passed.
	configFile string

	// JSON key naming function used by the built-in config command.
	// When nil, JSONKeyLong (identity) is used.
	jsonKeyName JSONKeyFunc

	// Cached version metadata (auto-detected and/or overridden).
	versionInfo VersionInfo

	// SEE ALSO cross-references rendered in all doc formats.
	seeAlso []string

	// Man-page-specific configuration (section number).
	manConfig manConfig

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

	// Number of options or positional args that carry validate-* rules.
	// When zero the post-parse validation sweep is skipped entirely.
	validationRuleCount int

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

	// When set by WriteHelpWithOptions, bypasses hidden filtering in WriteHelp.
	helpIncludeHidden bool

	// Indicates that post-scan configurators should be applied before parse.
	configDirty bool

	// Indicates that duplicate metadata validation must be re-run.
	validationDirty bool

	// Prevents recursive configurator execution.
	configuring bool

	// Set by built-in version option handler during parse.
	versionRequested bool

	// When true, --version/-v prints only the bare version string (like version --short).
	versionShort bool

	// Set when any immediate option/group is requested during parse.
	immediateRequested bool

	// When true, variable expansion ($VAR / ${VAR} / $${VAR}) in .env is disabled.
	dotEnvNoExpand bool

	// Guard: unified config/env options group already materialised.
	hasConfigGroup bool

	// Guard: dotenv options group already materialised.
	hasDotEnvGroup bool
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

// Options provides parser options that change the behavior of the option parser.
type Options uint64

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

	// StrictCommands causes the parser to return ErrUnknownCommand for any
	// unrecognized command token, even when the active command has positional
	// arguments defined or SubcommandsOptional is true.
	// Per-command override: Command.StrictSubcommands.
	StrictCommands

	// StrictPositionalArgs causes the parser to return ErrUnexpectedArgument
	// when more positional arguments are passed than are declared on the
	// active command.
	// Per-command override: Command.StrictArgs.
	StrictPositionalArgs

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

	// SilenceDeprecationWarnings suppresses the stderr warning emitted when
	// a deprecated option or command is used during parsing.
	SilenceDeprecationWarnings

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

	// ConfigIni enables INI format in the built-in config command.
	// Has no effect unless at least one of ConfigIni or ConfigJSON is set.
	ConfigIni

	// ConfigJSON enables JSON format in the built-in config command.
	// Has no effect unless at least one of ConfigIni or ConfigJSON is set.
	ConfigJSON

	// DetectShellFlagStyle enables shell-based flag style rendering in help
	// and doc output when no explicit render style is set.
	DetectShellFlagStyle

	// DetectShellEnvStyle enables shell-based env placeholder rendering in
	// help and doc output when no explicit render style is set.
	DetectShellEnvStyle

	// CommandChain executes every active command implementing Commander from
	// parent to leaf. Without this option only the last active command runs.
	CommandChain

	// DotEnv enables automatic loading of the .env file before ParseArgs.
	// The default file name is ".env" in the current working directory.
	// When the file does not exist, parsing continues silently.
	// Use SetDotEnvFile to change the file path.
	DotEnv

	// DotEnvOverride is like DotEnv but overrides existing environment
	// variables. Without this option values already set in the environment
	// take precedence over .env entries.
	DotEnvOverride

	// DotEnvFlags adds a built-in "Env Options" group with three flags:
	//   --env-file FILE   path to the .env file
	//   --no-env          disable .env loading for this run
	//   --env-override    override existing env vars from the .env file
	// These flags are pre-scanned before loading the .env file so they take
	// effect even when the file is specified on the command line.
	DotEnvFlags

	// ConfigFlags adds a -c/--config FILE flag to the "Config Options" group.
	// The flag is pre-scanned before parsing so the config file is loaded first.
	// Format is detected from the file extension (.ini, .json)
	// or the first byte of the file ('{' = JSON, otherwise INI).
	// Requires ConfigIni or ConfigJSON to be set to determine which formats are accepted.
	// Use SetConfigFile to set a default config path used when --config is absent.
	// When DotEnvFlags is also set, all flags share one "Config Options" group.
	ConfigFlags

	// Default is a convenient default set of options which should cover
	// most of the uses of the flags package.
	Default = HelpFlag | PrintErrors | PassDoubleDash

	// ConfigCommand adds a built-in config command that can render INI and/or JSON configuration.
	// The command is only registered when at least one of ConfigIni or ConfigJSON is also set.
	// Use ConfigIni or ConfigJSON alone to restrict the command to a single format.
	ConfigCommand = ConfigIni | ConfigJSON

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

	// unknownCmdHandled is set when UnknownCommandHandler returned nil, meaning
	// the caller deliberately accepted the unknown token. This suppresses the
	// automatic estimateCommand() call at the end of parsing.
	unknownCmdHandled bool
}

// CommandDescriptions contains short and long user-facing command descriptions.
type CommandDescriptions struct {
	Short string
	Long  string
}

// CommandDescriptionI18nKeys contains localization keys for command descriptions.
type CommandDescriptionI18nKeys struct {
	Short string
	Long  string
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
		validationDirty:            true,
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
// On failure the previous flag tags are restored and the tree is left untouched
// (rebuildTree never mutates the parser unless every group/command rescans successfully).
func (p *Parser) SetFlagTags(tags FlagTags) error {
	prev := p.flagTags
	p.flagTags = tags.withDefaults()

	if err := p.rebuildTree(); err != nil {
		p.flagTags = prev
		return err
	}

	return nil
}

// SetEnvPrefix configures a global prefix for all environment variable keys.
// For example, with prefix "MY_APP" and delimiter "_", env key "PORT" becomes
// "MY_APP_PORT", and grouped keys become "MY_APP_<GROUP>_<KEY>".
func (p *Parser) SetEnvPrefix(prefix string) {
	p.EnvPrefix = prefix
}

// SetBanner configures raw banner text rendered after the help header and
// before normal built-in help output.
func (p *Parser) SetBanner(text string) {
	p.banner = text
}

// SetHelpHeader configures raw text rendered before the banner and normal
// built-in help output.
func (p *Parser) SetHelpHeader(text string) {
	p.helpHeader = text
}

// SetHelpFooter configures raw text rendered after normal built-in help output.
func (p *Parser) SetHelpFooter(text string) {
	p.helpFooter = text
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

// SetCommandOptionIndent configures extra spaces before command option rows in built-in help output.
// The default is 0, so top-level and command options use the same indentation.
func (p *Parser) SetCommandOptionIndent(indent int) error {
	if indent < 0 {
		return ErrNegativeCommandOptionIndent
	}

	p.commandOptionIndent = indent
	return nil
}

// SetHelpWidth configures built-in help output wrapping width.
// When unset, help uses the current terminal width with a fallback of 80 columns.
// Width 0 disables wrapping.
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
// immediately. On failure the previous limit is restored and the tree is
// left untouched (rebuildTree never mutates the parser unless every
// group/command rescans successfully).
func (p *Parser) SetMaxLongNameLength(length int) error {
	if length < 0 {
		return ErrNegativeMaxLongNameLength
	}

	prev := p.MaxLongNameLength
	p.MaxLongNameLength = length

	if err := p.rebuildTree(); err != nil {
		p.MaxLongNameLength = prev
		return err
	}

	return nil
}

// SetTagListDelimiter sets delimiter for list-based struct tags such as
// defaults/choices/aliases and rescans attached groups/commands. On failure
// the previous delimiter is restored and the tree is left untouched
// (rebuildTree never mutates the parser unless every group/command rescans
// successfully).
func (p *Parser) SetTagListDelimiter(delimiter rune) error {
	if delimiter == 0 {
		return ErrNULTagListDelimiter
	}

	prev := p.TagListDelimiter
	p.TagListDelimiter = delimiter

	if err := p.rebuildTree(); err != nil {
		p.TagListDelimiter = prev
		return err
	}

	return nil
}

// SetOptionSort configures option order mode for grouped option presentation.
func (p *Parser) SetOptionSort(mode OptionSortMode) {
	p.optionSort = mode
}

// SetCommandSort configures command order mode for help/docs presentation.
func (p *Parser) SetCommandSort(mode CommandSortMode) {
	p.commandSort = mode
}

// SetCommandShortDescriptions updates short descriptions for multiple commands.
// Missing command names are ignored.
func (p *Parser) SetCommandShortDescriptions(descriptions map[string]string) {
	for commandName, description := range descriptions {
		if cmd := p.Find(commandName); cmd != nil {
			cmd.SetShortDescription(description)
		}
	}
}

// SetCommandLongDescriptions updates long descriptions for multiple commands.
// Missing command names are ignored.
func (p *Parser) SetCommandLongDescriptions(descriptions map[string]string) {
	for commandName, description := range descriptions {
		if cmd := p.Find(commandName); cmd != nil {
			cmd.SetLongDescription(description)
		}
	}
}

// SetCommandDescriptions updates short and long descriptions for multiple commands.
// Missing command names are ignored.
func (p *Parser) SetCommandDescriptions(descriptions map[string]CommandDescriptions) {
	for commandName, description := range descriptions {
		if cmd := p.Find(commandName); cmd != nil {
			cmd.SetShortDescription(description.Short)
			cmd.SetLongDescription(description.Long)
		}
	}
}

// SetCommandShortDescriptionI18nKeys updates short description i18n keys
// for multiple commands. Missing command names are ignored.
func (p *Parser) SetCommandShortDescriptionI18nKeys(keys map[string]string) {
	for commandName, key := range keys {
		if cmd := p.Find(commandName); cmd != nil {
			cmd.SetShortDescriptionI18nKey(key)
		}
	}
}

// SetCommandLongDescriptionI18nKeys updates long description i18n keys
// for multiple commands. Missing command names are ignored.
func (p *Parser) SetCommandLongDescriptionI18nKeys(keys map[string]string) {
	for commandName, key := range keys {
		if cmd := p.Find(commandName); cmd != nil {
			cmd.SetLongDescriptionI18nKey(key)
		}
	}
}

// SetCommandDescriptionI18nKeys updates short and long description i18n keys
// for multiple commands. Missing command names are ignored.
func (p *Parser) SetCommandDescriptionI18nKeys(keys map[string]CommandDescriptionI18nKeys) {
	for commandName, key := range keys {
		if cmd := p.Find(commandName); cmd != nil {
			cmd.SetShortDescriptionI18nKey(key.Short)
			cmd.SetLongDescriptionI18nKey(key.Long)
		}
	}
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

// SetDefaultCommand sets the root subcommand to activate when no explicit
// command token is provided. Equivalent to calling SetDefaultSubcommand on the
// root Command. The named subcommand must already be registered.
// Call with an empty string to clear the default.
func (p *Parser) SetDefaultCommand(name string) {
	p.SetDefaultSubcommand(name)
}

// ParseArgs parses the command line arguments according to the option groups that were added to the parser.
// On successful parsing of the arguments, the remaining, non-option, arguments (if any) are returned.
// The returned error indicates a parsing error and can be used with PrintError
// to display contextual information on where the error occurred exactly.
//
// When the common help group has been added (AddHelp)
// and either -h or --help was specified in the command line arguments,
// a help message will be automatically printed if the PrintErrors option is enabled.
// Furthermore, the special error type ErrHelp is returned.
// It is up to the caller to exit the program if so desired.
//
// The parser may be reused: calling ParseArgs more than once on the same Parser does not reset it.
// Each call layers its arguments on top of the currently bound values,
// so a value set by an earlier explicit CLI flag, or pre-filled into the bound struct before the first call,
// is not overwritten by defaults/env/config on a later call unless that later call sets the option explicitly again.
//
// This is what makes patterns like config-then-CLI
// (populate the struct from a config file, then call ParseArgs to let CLI flags override it)
// and ConfiguredValues (see its doc comment) work.
// Directly mutating a bound struct field outside the parser
// does not reset the parser's own bookkeeping for that option.
func (p *Parser) ParseArgs(args []string) ([]string, error) {
	if p.internalError != nil {
		return nil, p.printError(p.internalError)
	}

	if err := p.applyConfigurators(); err != nil {
		return nil, p.printError(err)
	}

	// Load config file when ConfigFlags is active. Must happen before dotenv
	// and EnsureBuiltinOptions so that config values are available early.
	if (p.Options & ConfigFlags) != None {
		if err := p.EnsureBuiltinConfigOptions(); err != nil {
			return nil, p.printError(err)
		}

		if err := p.applyConfigFile(args); err != nil {
			return nil, p.printError(err)
		}
	}

	// Load .env file when any DotEnv option is active. This must happen before
	// EnsureBuiltinOptions so env vars are available when option defaults are
	// resolved from the environment.
	if (p.Options & (DotEnv | DotEnvOverride | DotEnvFlags)) != None {
		if err := p.EnsureBuiltinDotEnvOptions(); err != nil {
			return nil, p.printError(err)
		}

		if err := p.applyDotEnv(args); err != nil {
			return nil, p.printError(err)
		}
	}

	// Add built-in help/version group before duplicate validation so their
	// flags cannot silently shadow application or command flags.
	p.EnsureBuiltinOptions()
	if err := p.EnsureBuiltinCommands(); err != nil {
		return nil, p.printError(err)
	}

	if p.validationDirty {
		if err := p.validateDuplicateFlags(); err != nil {
			return nil, p.printError(err)
		}
		if err := p.validateDuplicateCommands(); err != nil {
			return nil, p.printError(err)
		}
		if err := p.validateRequiresProvides(); err != nil {
			return nil, p.printError(err)
		}
		p.validationDirty = false
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

		// An option unknown to the current command may belong to a not-yet-activated default command
		// (e.g. `app --token x` where --token is owned by the default command).
		// Lazily activate the default command chain and retry,
		// mirroring the analogous fallback for non-option tokens in parseNonOption.
		// Loop to cascade through nested defaults.
		for err != nil && wrapError(err).Type == ErrUnknownFlag && s.command.Active == nil &&
			s.command.defaultCommand != "" && len(s.command.commands) > 0 {
			cmd := s.lookup.commands[s.command.defaultCommand]
			if cmd == nil {
				break
			}

			s.command.Active = cmd
			cmd.fillParseState(s)

			if islong {
				err = p.parseLong(s, optname, argument, hasArgument)
			} else {
				err = p.parseShort(s, optname, argument, hasArgument)
			}
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

	// If no explicit command was selected and a default is configured,
	// activate it now, before defaults/validators run, so that required/validator/relation checks
	// and positional defaults apply to the default command's own options and args.
	// Loop to cascade through nested default commands.
	for s.err == nil && s.command.Active == nil &&
		s.command.defaultCommand != "" && len(s.command.commands) > 0 {
		cmd := s.lookup.commands[s.command.defaultCommand]
		if cmd == nil {
			break
		}
		s.command.Active = cmd
		cmd.fillParseState(s)
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

		if s.err == nil {
			if validationErr := s.checkValueValidators(p); validationErr != nil {
				s.err = validationErr
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

	switch {
	case s.err != nil:
		reterr = s.err
	case p.shouldSkipCommandExecution():
		return s.retargs, nil
	case len(s.command.commands) != 0 && !s.command.SubcommandsOptional && !s.unknownCmdHandled:
		reterr = s.estimateCommand()
	default:
		reterr = p.executeCommands(s.command, s.retargs)
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

func (p *Parser) executeCommands(command *Command, args []string) error {
	if (p.Options & CommandChain) != None {
		return p.executeCommandChain(command, args)
	}

	if cmd, ok := command.data.(Commander); ok {
		return p.executeCommand(cmd, args)
	}

	if p.CommandHandler != nil {
		return p.CommandHandler(nil, args)
	}

	return nil
}

func (p *Parser) executeCommandChain(command *Command, args []string) error {
	executed := false

	for _, active := range activeCommandPath(command) {
		cmd, ok := active.data.(Commander)
		if !ok {
			continue
		}

		executed = true

		if err := p.executeCommand(cmd, args); err != nil {
			return err
		}
	}

	if !executed && p.CommandHandler != nil {
		return p.CommandHandler(nil, args)
	}

	return nil
}

func (p *Parser) executeCommand(command Commander, args []string) error {
	if p.CommandHandler != nil {
		return p.CommandHandler(command, args)
	}

	return command.Execute(args)
}

func activeCommandPath(command *Command) []*Command {
	var reversed []*Command

	for command != nil {
		reversed = append(reversed, command)

		parent, ok := command.parent.(*Command)
		if !ok {
			break
		}

		command = parent
	}

	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}

	return reversed
}

func (p *Parser) normalizeStructTag(mtag *multiTag) {
	c := mtag.cached()

	normalizeTagAlias(c, p.flagTags.Short, FlagTagShort)
	normalizeTagAlias(c, p.flagTags.Long, FlagTagLong)
	normalizeTagAlias(c, p.flagTags.Required, FlagTagRequired)
	normalizeTagAlias(c, p.flagTags.Xor, FlagTagXor)
	normalizeTagAlias(c, p.flagTags.And, FlagTagAnd)
	normalizeTagAlias(c, p.flagTags.Or, FlagTagOr)
	normalizeTagAlias(c, p.flagTags.Nand, FlagTagNand)
	normalizeTagAlias(c, p.flagTags.Requires, FlagTagRequires)
	normalizeTagAlias(c, p.flagTags.Provides, FlagTagProvides)
	normalizeTagAlias(c, p.flagTags.Counter, FlagTagCounter)
	normalizeTagAlias(c, p.flagTags.IO, FlagTagIO)
	normalizeTagAlias(c, p.flagTags.IOKind, FlagTagIOKind)
	normalizeTagAlias(c, p.flagTags.IOStream, FlagTagIOStream)
	normalizeTagAlias(c, p.flagTags.IOOpen, FlagTagIOOpen)
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
	normalizeTagAlias(c, p.flagTags.Deprecated, FlagTagDeprecated)
	normalizeTagAlias(c, p.flagTags.Secret, FlagTagSecret)
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
	normalizeTagAlias(c, p.flagTags.GroupXor, FlagTagGroupXor)
	normalizeTagAlias(c, p.flagTags.GroupAnd, FlagTagGroupAnd)
	normalizeTagAlias(c, p.flagTags.GroupOr, FlagTagGroupOr)
	normalizeTagAlias(c, p.flagTags.GroupNand, FlagTagGroupNand)
	normalizeTagAlias(c, p.flagTags.GroupRequires, FlagTagGroupRequires)
	normalizeTagAlias(c, p.flagTags.GroupProvides, FlagTagGroupProvides)
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
	normalizeTagAlias(c, p.flagTags.ValidateExistingFile, FlagTagValidateExistingFile)
	normalizeTagAlias(c, p.flagTags.ValidateExistingDir, FlagTagValidateExistingDir)
	normalizeTagAlias(c, p.flagTags.ValidateReadable, FlagTagValidateReadable)
	normalizeTagAlias(c, p.flagTags.ValidateWritable, FlagTagValidateWritable)
	normalizeTagAlias(c, p.flagTags.ValidateNonEmpty, FlagTagValidateNonEmpty)
	normalizeTagAlias(c, p.flagTags.ValidateRegex, FlagTagValidateRegex)
	normalizeTagAlias(c, p.flagTags.ValidateMinLen, FlagTagValidateMinLen)
	normalizeTagAlias(c, p.flagTags.ValidateMaxLen, FlagTagValidateMaxLen)
	normalizeTagAlias(c, p.flagTags.ValidateMin, FlagTagValidateMin)
	normalizeTagAlias(c, p.flagTags.ValidateMax, FlagTagValidateMax)
	normalizeTagAlias(c, p.flagTags.ValidatePathAbs, FlagTagValidatePathAbs)
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
	p.validationDirty = true
}

type groupSpec struct {
	data                    any
	shortDescription        string
	longDescription         string
	namespace               string
	envNamespace            string
	iniName                 string
	shortDescriptionI18nKey string
	longDescriptionI18nKey  string
	xorGroups               []string
	andGroups               []string
	orGroups                []string
	nandGroups              []string
	requiresTokens          []string
	providesTokens          []string
	hidden                  bool
	immediate               bool
	required                bool
}

type commandSpec struct {
	data                    any
	name                    string
	shortDescription        string
	longDescription         string
	iniName                 string
	shortDescriptionI18nKey string
	longDescriptionI18nKey  string
	commandGroup            string
	commandGroupI18nKey     string
	deprecated              string
	defaultCommand          string
	aliases                 []string
	order                   int
	subcommandsOptional     bool
	passAfterNonOption      bool
	hidden                  bool
	immediate               bool
	strictSubcommands       bool
	argsRequired            bool
	strictArgs              bool
}

// rebuildTree rescans attached top-level groups/commands
// after a setting that changes tag-scanning rules
// (tag list delimiter, flag tag names, max long-name length).
// It snapshots the current runtime metadata for each top-level group/command
// and reapplies it after rescanning.
//
// Known limitation: a command declared as a nested command-tagged struct field is transitively recreated
// by its owning group's own rescan (which already reflects the new settings)
// before this function's command loop runs.
//
// That loop deliberately skips reapplying its stale snapshot onto such a command,
// because doing so would overwrite freshly-rescanned tag-derived fields
// (e.g. Aliases split with the previous delimiter) with pre-rebuild data.
//
// So metadata set via Set*() calls on a tag-declared command
// (as opposed to one added directly through AddCommand) does not survive a rebuild;
// see TestSetTagListDelimiterDoesNotPreserveTagDeclaredCommandMetadata.
func (p *Parser) rebuildTree() error {
	groups := make([]groupSpec, 0, len(p.groups))
	commands := make([]commandSpec, 0, len(p.commands))
	rootOptions := append([]*Option(nil), p.options...)

	for _, g := range p.groups {
		if g.isBuiltinHelp {
			continue
		}

		groups = append(groups, groupSpec{
			shortDescription:        g.ShortDescription,
			longDescription:         g.LongDescription,
			namespace:               g.Namespace,
			envNamespace:            g.EnvNamespace,
			iniName:                 g.IniName,
			shortDescriptionI18nKey: g.ShortDescriptionI18nKey,
			longDescriptionI18nKey:  g.LongDescriptionI18nKey,
			hidden:                  g.Hidden,
			immediate:               g.Immediate,
			data:                    g.data,
			required:                g.Required,
			xorGroups:               append([]string(nil), g.XorGroups...),
			andGroups:               append([]string(nil), g.AndGroups...),
			orGroups:                append([]string(nil), g.OrGroups...),
			nandGroups:              append([]string(nil), g.NandGroups...),
			requiresTokens:          append([]string(nil), g.Requires...),
			providesTokens:          append([]string(nil), g.Provides...),
		})
	}

	for _, c := range p.commands {
		commands = append(commands, commandSpec{
			name:                    c.Name,
			shortDescription:        c.ShortDescription,
			longDescription:         c.LongDescription,
			iniName:                 c.IniName,
			shortDescriptionI18nKey: c.ShortDescriptionI18nKey,
			longDescriptionI18nKey:  c.LongDescriptionI18nKey,
			commandGroup:            c.CommandGroup,
			commandGroupI18nKey:     c.CommandGroupI18nKey,
			deprecated:              c.Deprecated,
			defaultCommand:          c.defaultCommand,
			aliases:                 append([]string(nil), c.Aliases...),
			order:                   c.Order,
			subcommandsOptional:     c.SubcommandsOptional,
			passAfterNonOption:      c.PassAfterNonOption,
			hidden:                  c.Hidden,
			immediate:               c.Immediate,
			strictSubcommands:       c.StrictSubcommands,
			argsRequired:            c.ArgsRequired,
			strictArgs:              c.StrictArgs,
			data:                    c.data,
		})
	}

	// Build the new tree on an isolated scratch parser instead of mutating p directly:
	// scratch shares all of p's scan settings (flag tags, tag list/ delimiter, max long-name length, i18n, ...)
	// via a shallow copy, but owns a brand-new root Command,
	// so groups/commands added to it never touch p.groups/p.commands.
	//
	// If any rescan step below fails, we return the error immediately and p is left completely untouched, including p.Active.
	// Only once every group/command has been rescanned/ successfully do we adopt the scratch tree onto p.
	scratchVal := *p
	scratchVal.Command = newCommand(p.Name, p.ShortDescription, p.LongDescription, nil)
	scratchVal.parent = &scratchVal
	scratchVal.validationRuleCount = 0
	scratch := &scratchVal

	for _, g := range groups {
		ng, err := scratch.AddGroup(g.shortDescription, g.longDescription, g.data)
		if err != nil {
			return fmt.Errorf("failed to rescan group %q: %w", g.shortDescription, err)
		}
		ng.Namespace = g.namespace
		ng.EnvNamespace = g.envNamespace
		ng.IniName = g.iniName
		ng.Hidden = g.hidden
		ng.Immediate = g.immediate
		ng.ShortDescriptionI18nKey = g.shortDescriptionI18nKey
		ng.LongDescriptionI18nKey = g.longDescriptionI18nKey
		ng.Required = g.required
		ng.XorGroups = g.xorGroups
		ng.AndGroups = g.andGroups
		ng.OrGroups = g.orGroups
		ng.NandGroups = g.nandGroups
		ng.Requires = g.requiresTokens
		ng.Provides = g.providesTokens
	}

	for _, c := range commands {
		// A command declared as a nested struct field inside another group's data (the common case)
		// is already recreated by that group's AddGroup scan above, via the same struct-tag reflection,
		// and that fresh scan already reflects current parser settings (tag list delimiter, flag tag names, ...).
		// Restoring the pre-rebuild snapshot onto it would reintroduce stale tag-derived data
		// (e.g. Aliases split with the old TagListDelimiter), so leave it alone.
		if existing := scratch.Find(c.name); existing != nil && sameCommandData(existing.data, c.data) {
			continue
		}

		nc, err := scratch.AddCommand(c.name, c.shortDescription, c.longDescription, c.data)
		if err != nil {
			return fmt.Errorf("failed to rescan command %q: %w", c.name, err)
		}
		nc.Aliases = c.aliases
		nc.IniName = c.iniName
		nc.SubcommandsOptional = c.subcommandsOptional
		nc.PassAfterNonOption = c.passAfterNonOption
		nc.Hidden = c.hidden
		nc.Immediate = c.immediate
		nc.ShortDescriptionI18nKey = c.shortDescriptionI18nKey
		nc.LongDescriptionI18nKey = c.longDescriptionI18nKey
		nc.CommandGroup = c.commandGroup
		nc.CommandGroupI18nKey = c.commandGroupI18nKey
		nc.Deprecated = c.deprecated
		nc.defaultCommand = c.defaultCommand
		nc.Order = c.order
		nc.StrictSubcommands = c.strictSubcommands
		nc.ArgsRequired = c.argsRequired
		nc.StrictArgs = c.strictArgs
	}

	// Every rescan above succeeded: adopt the scratch tree.
	// Top-level groups and commands still have their parent set to scratch.Command;
	// repoint them at p.Command so Group.parser()/Command.parser() resolve to the real, live parser again
	// (nested descendants need no fix-up, since their parent chain already resolves through these top-level objects).
	p.groups = scratch.groups
	p.commands = scratch.commands
	p.options = rootOptions
	p.args = scratch.args
	p.Active = nil
	p.hasBuiltinHelpGroup = false
	p.validationRuleCount = scratch.validationRuleCount

	for _, g := range p.groups {
		g.parent = p.Command
	}
	for _, c := range p.commands {
		c.parent = p.Command
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
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return av.Pointer() == bv.Pointer()
	default:
		if av.Type().Comparable() {
			return a == b
		}
		return false
	}
}
