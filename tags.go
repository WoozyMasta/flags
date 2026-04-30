// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

// FlagTag* constants define canonical struct tag keys used by the parser.
// They correspond to struct field tags read during parser/group/command scan.
const (
	// FlagTagShort configures a one-character short flag name (`-v`).
	FlagTagShort = "short"
	// FlagTagLong configures a long flag name (`--verbose`).
	FlagTagLong = "long"
	// FlagTagRequired marks option or positional argument as required.
	FlagTagRequired = "required"
	// FlagTagXor declares mutually exclusive option relation groups.
	FlagTagXor = "xor"
	// FlagTagAnd declares option relation groups that must be used together.
	FlagTagAnd = "and"
	// FlagTagCounter enables counter mode for integer flags.
	FlagTagCounter = "counter"
	// FlagTagDescription provides short help text.
	FlagTagDescription = "description"
	// FlagTagDescriptionI18n provides i18n key for description text.
	FlagTagDescriptionI18n = "description-i18n"
	// FlagTagLongDescription provides extended text (used in man/help contexts).
	FlagTagLongDescription = "long-description"
	// FlagTagLongDescriptionI18n provides i18n key for long description text.
	FlagTagLongDescriptionI18n = "long-description-i18n"
	// FlagTagNoFlag excludes the field from command-line flag parsing.
	FlagTagNoFlag = "no-flag"
	// FlagTagOptional marks option argument as optional.
	FlagTagOptional = "optional"
	// FlagTagOptionalValue defines fallback value when optional arg is omitted.
	FlagTagOptionalValue = "optional-value"
	// FlagTagOrder sets display/completion priority order for options and commands.
	FlagTagOrder = "order"
	// FlagTagDefault sets default option value (repeatable for slices/maps).
	FlagTagDefault = "default"
	// FlagTagDefaults sets multiple default option values as a delimiter-separated list.
	FlagTagDefaults = "defaults"
	// FlagTagDefaultMask customizes how default is shown in generated help.
	FlagTagDefaultMask = "default-mask"
	// FlagTagEnv maps option default to an environment variable key.
	FlagTagEnv = "env"
	// FlagTagAutoEnv enables deriving env key from long flag name.
	FlagTagAutoEnv = "auto-env"
	// FlagTagEnvDelim splits env-provided list/map values by delimiter.
	FlagTagEnvDelim = "env-delim"
	// FlagTagValueName customizes value placeholder shown in help.
	FlagTagValueName = "value-name"
	// FlagTagValueNameI18n provides i18n key for value placeholder.
	FlagTagValueNameI18n = "value-name-i18n"
	// FlagTagChoice restricts allowed option values (repeatable).
	FlagTagChoice = "choice"
	// FlagTagChoices restricts allowed option values as a delimiter-separated list.
	FlagTagChoices = "choices"
	// FlagTagCompletion configures completion hint (file, dir, none).
	FlagTagCompletion = "completion"
	// FlagTagHidden hides option/group/command from help and completion output.
	FlagTagHidden = "hidden"
	// FlagTagImmediate marks options/groups/commands that should bypass required checks.
	FlagTagImmediate = "immediate"
	// FlagTagBase sets radix for integer parsing.
	FlagTagBase = "base"
	// FlagTagIniName overrides key name used for INI parse/write.
	FlagTagIniName = "ini-name"
	// FlagTagIniGroup overrides section token used for INI group/command blocks.
	FlagTagIniGroup = "ini-group"
	// FlagTagNoIni excludes the field from INI parse/write.
	FlagTagNoIni = "no-ini"
	// FlagTagGroup turns a nested struct field into an option group.
	FlagTagGroup = "group"
	// FlagTagGroupI18n provides i18n key for group display name.
	FlagTagGroupI18n = "group-i18n"
	// FlagTagNamespace prefixes long option names for grouped options.
	FlagTagNamespace = "namespace"
	// FlagTagEnvNamespace prefixes environment variable names for grouped options.
	FlagTagEnvNamespace = "env-namespace"
	// FlagTagCommand turns a field into a subcommand.
	FlagTagCommand = "command"
	// FlagTagCommandI18n provides i18n key for command short description.
	FlagTagCommandI18n = "command-i18n"
	// FlagTagCommandGroup groups commands in help and documentation output.
	FlagTagCommandGroup = "command-group"
	// FlagTagSubCommandsOptional makes child subcommands optional.
	FlagTagSubCommandsOptional = "subcommands-optional"
	// FlagTagAlias adds extra command names (repeatable).
	FlagTagAlias = "alias"
	// FlagTagAliases adds extra command names as a delimiter-separated list.
	FlagTagAliases = "aliases"
	// FlagTagLongAlias adds extra long option names (repeatable).
	FlagTagLongAlias = "long-alias"
	// FlagTagLongAliases adds extra long option names as a delimiter-separated list.
	FlagTagLongAliases = "long-aliases"
	// FlagTagShortAlias adds extra short option names (repeatable).
	FlagTagShortAlias = "short-alias"
	// FlagTagShortAliases adds extra short option names as a delimiter-separated list.
	FlagTagShortAliases = "short-aliases"
	// FlagTagPositionalArgs marks a struct as positional arguments container.
	FlagTagPositionalArgs = "positional-args"
	// FlagTagPositionalArgName sets display name for a positional argument.
	FlagTagPositionalArgName = "positional-arg-name"
	// FlagTagArgNameI18n provides i18n key for positional arg display name.
	FlagTagArgNameI18n = "arg-name-i18n"
	// FlagTagArgDescriptionI18n provides i18n key for positional arg description.
	FlagTagArgDescriptionI18n = "arg-description-i18n"
	// FlagTagKeyValueDelimiter customizes map key/value delimiter.
	FlagTagKeyValueDelimiter = "key-value-delimiter"
	// FlagTagPassAfterNonOption enables command-local strict POSIX behavior.
	FlagTagPassAfterNonOption = "pass-after-non-option"
	// FlagTagUnquote controls automatic unquoting of string arguments.
	FlagTagUnquote = "unquote"
	// FlagTagTerminator marks an option that consumes arguments until terminator token.
	FlagTagTerminator = "terminator"
)

const flagTagReadIniName = "_read-ini-name"

// FlagTags defines all configurable struct tag keys used by the parser.
// Override fields to avoid collisions with other libraries using struct tags.
// Typical usage is NewFlagTagsWithPrefix("flag-") and then optional per-field
// overrides for mixed schemas.
type FlagTags struct {
	// Short maps to short option name tag (default: "short").
	Short string
	// Long maps to long option name tag (default: "long").
	Long string
	// Required maps to required marker tag (default: "required").
	Required string
	// Xor maps to mutually exclusive option relation groups tag.
	Xor string
	// And maps to option relation groups that must be used together tag.
	And string
	// Counter maps to counter mode tag for integer flags.
	Counter string
	// Description maps to short help text tag (default: "description").
	Description string
	// DescriptionI18n maps to i18n key for short help text (default: "description-i18n").
	DescriptionI18n string
	// LongDescription maps to extended help/man text tag (default: "long-description").
	LongDescription string
	// LongDescriptionI18n maps to i18n key for extended help/man text.
	LongDescriptionI18n string
	// NoFlag maps to option exclusion tag (default: "no-flag").
	NoFlag string
	// Optional maps to optional argument marker tag (default: "optional").
	Optional string
	// OptionalValue maps to fallback optional value tag (default: "optional-value").
	OptionalValue string
	// Order maps to option sorting priority tag (default: "order").
	Order string
	// Default maps to default value tag (default: "default").
	Default string
	// Defaults maps to multi-default value tag (default: "defaults").
	Defaults string
	// DefaultMask maps to help default-mask tag (default: "default-mask").
	DefaultMask string
	// Env maps to environment variable key tag (default: "env").
	Env string
	// AutoEnv maps to env auto-derivation toggle tag (default: "auto-env").
	AutoEnv string
	// EnvDelim maps to env list/map delimiter tag (default: "env-delim").
	EnvDelim string
	// ValueName maps to help placeholder tag (default: "value-name").
	ValueName string
	// ValueNameI18n maps to i18n key for value placeholder tag.
	ValueNameI18n string
	// Choice maps to allowed-values tag (default: "choice").
	Choice string
	// Choices maps to multi allowed-values tag (default: "choices").
	Choices string
	// Completion maps to completion hint tag (default: "completion").
	Completion string
	// Hidden maps to hide-from-help tag (default: "hidden").
	Hidden string
	// Immediate maps to immediate-processing tag (default: "immediate").
	Immediate string
	// Base maps to integer radix tag (default: "base").
	Base string
	// IniName maps to INI key override tag (default: "ini-name").
	IniName string
	// IniGroup maps to INI section token override tag (default: "ini-group").
	IniGroup string
	// NoIni maps to INI exclusion tag (default: "no-ini").
	NoIni string
	// Group maps to group declaration tag (default: "group").
	Group string
	// GroupI18n maps to i18n key for group display name tag.
	GroupI18n string
	// Namespace maps to long-name namespace tag (default: "namespace").
	Namespace string
	// EnvNamespace maps to env-name namespace tag (default: "env-namespace").
	EnvNamespace string
	// Command maps to subcommand declaration tag (default: "command").
	Command string
	// CommandI18n maps to i18n key for command short description tag.
	CommandI18n string
	// CommandGroup maps to command group tag (default: "command-group").
	CommandGroup string
	// SubCommandsOptional maps to optional-subcommands tag (default: "subcommands-optional").
	SubCommandsOptional string
	// Alias maps to command alias tag (default: "alias").
	Alias string
	// Aliases maps to multi command aliases tag (default: "aliases").
	Aliases string
	// LongAlias maps to option long alias tag (default: "long-alias").
	LongAlias string
	// LongAliases maps to multi option long aliases tag (default: "long-aliases").
	LongAliases string
	// ShortAlias maps to option short alias tag (default: "short-alias").
	ShortAlias string
	// ShortAliases maps to multi option short aliases tag (default: "short-aliases").
	ShortAliases string
	// PositionalArgs maps to positional args struct tag (default: "positional-args").
	PositionalArgs string
	// PositionalArgName maps to positional display-name tag (default: "positional-arg-name").
	PositionalArgName string
	// ArgNameI18n maps to positional i18n display-name tag.
	ArgNameI18n string
	// ArgDescriptionI18n maps to positional i18n description tag.
	ArgDescriptionI18n string
	// KeyValueDelimiter maps to map key/value delimiter tag (default: "key-value-delimiter").
	KeyValueDelimiter string
	// PassAfterNonOption maps to command-local POSIX behavior tag (default: "pass-after-non-option").
	PassAfterNonOption string
	// Unquote maps to string unquoting control tag (default: "unquote").
	Unquote string
	// Terminator maps to terminated-arguments tag (default: "terminator").
	Terminator string
}

// NewFlagTags returns default tag names.
func NewFlagTags() FlagTags {
	return NewFlagTagsWithPrefix("")
}

// NewFlagTagsWithPrefix returns default tag names with a custom prefix.
func NewFlagTagsWithPrefix(prefix string) FlagTags {
	return FlagTags{
		Short:               prefix + FlagTagShort,
		Long:                prefix + FlagTagLong,
		Required:            prefix + FlagTagRequired,
		Xor:                 prefix + FlagTagXor,
		And:                 prefix + FlagTagAnd,
		Counter:             prefix + FlagTagCounter,
		Description:         prefix + FlagTagDescription,
		DescriptionI18n:     prefix + FlagTagDescriptionI18n,
		LongDescription:     prefix + FlagTagLongDescription,
		LongDescriptionI18n: prefix + FlagTagLongDescriptionI18n,
		NoFlag:              prefix + FlagTagNoFlag,
		Optional:            prefix + FlagTagOptional,
		OptionalValue:       prefix + FlagTagOptionalValue,
		Order:               prefix + FlagTagOrder,
		Default:             prefix + FlagTagDefault,
		Defaults:            prefix + FlagTagDefaults,
		DefaultMask:         prefix + FlagTagDefaultMask,
		Env:                 prefix + FlagTagEnv,
		AutoEnv:             prefix + FlagTagAutoEnv,
		EnvDelim:            prefix + FlagTagEnvDelim,
		ValueName:           prefix + FlagTagValueName,
		ValueNameI18n:       prefix + FlagTagValueNameI18n,
		Choice:              prefix + FlagTagChoice,
		Choices:             prefix + FlagTagChoices,
		Completion:          prefix + FlagTagCompletion,
		Hidden:              prefix + FlagTagHidden,
		Immediate:           prefix + FlagTagImmediate,
		Base:                prefix + FlagTagBase,
		IniName:             prefix + FlagTagIniName,
		IniGroup:            prefix + FlagTagIniGroup,
		NoIni:               prefix + FlagTagNoIni,
		Group:               prefix + FlagTagGroup,
		GroupI18n:           prefix + FlagTagGroupI18n,
		Namespace:           prefix + FlagTagNamespace,
		EnvNamespace:        prefix + FlagTagEnvNamespace,
		Command:             prefix + FlagTagCommand,
		CommandI18n:         prefix + FlagTagCommandI18n,
		CommandGroup:        prefix + FlagTagCommandGroup,
		SubCommandsOptional: prefix + FlagTagSubCommandsOptional,
		Alias:               prefix + FlagTagAlias,
		Aliases:             prefix + FlagTagAliases,
		LongAlias:           prefix + FlagTagLongAlias,
		LongAliases:         prefix + FlagTagLongAliases,
		ShortAlias:          prefix + FlagTagShortAlias,
		ShortAliases:        prefix + FlagTagShortAliases,
		PositionalArgs:      prefix + FlagTagPositionalArgs,
		PositionalArgName:   prefix + FlagTagPositionalArgName,
		ArgNameI18n:         prefix + FlagTagArgNameI18n,
		ArgDescriptionI18n:  prefix + FlagTagArgDescriptionI18n,
		KeyValueDelimiter:   prefix + FlagTagKeyValueDelimiter,
		PassAfterNonOption:  prefix + FlagTagPassAfterNonOption,
		Unquote:             prefix + FlagTagUnquote,
		Terminator:          prefix + FlagTagTerminator,
	}
}

func (t FlagTags) withDefaults() FlagTags {
	d := NewFlagTags()

	if t.Short != "" {
		d.Short = t.Short
	}
	if t.Long != "" {
		d.Long = t.Long
	}
	if t.Required != "" {
		d.Required = t.Required
	}
	if t.Xor != "" {
		d.Xor = t.Xor
	}
	if t.And != "" {
		d.And = t.And
	}
	if t.Counter != "" {
		d.Counter = t.Counter
	}
	if t.Description != "" {
		d.Description = t.Description
	}
	if t.DescriptionI18n != "" {
		d.DescriptionI18n = t.DescriptionI18n
	}
	if t.LongDescription != "" {
		d.LongDescription = t.LongDescription
	}
	if t.LongDescriptionI18n != "" {
		d.LongDescriptionI18n = t.LongDescriptionI18n
	}
	if t.NoFlag != "" {
		d.NoFlag = t.NoFlag
	}
	if t.Optional != "" {
		d.Optional = t.Optional
	}
	if t.OptionalValue != "" {
		d.OptionalValue = t.OptionalValue
	}
	if t.Order != "" {
		d.Order = t.Order
	}
	if t.Default != "" {
		d.Default = t.Default
	}
	if t.Defaults != "" {
		d.Defaults = t.Defaults
	}
	if t.DefaultMask != "" {
		d.DefaultMask = t.DefaultMask
	}
	if t.Env != "" {
		d.Env = t.Env
	}
	if t.AutoEnv != "" {
		d.AutoEnv = t.AutoEnv
	}
	if t.EnvDelim != "" {
		d.EnvDelim = t.EnvDelim
	}
	if t.ValueName != "" {
		d.ValueName = t.ValueName
	}
	if t.ValueNameI18n != "" {
		d.ValueNameI18n = t.ValueNameI18n
	}
	if t.Choice != "" {
		d.Choice = t.Choice
	}
	if t.Choices != "" {
		d.Choices = t.Choices
	}
	if t.Completion != "" {
		d.Completion = t.Completion
	}
	if t.Hidden != "" {
		d.Hidden = t.Hidden
	}
	if t.Immediate != "" {
		d.Immediate = t.Immediate
	}
	if t.Base != "" {
		d.Base = t.Base
	}
	if t.IniName != "" {
		d.IniName = t.IniName
	}
	if t.IniGroup != "" {
		d.IniGroup = t.IniGroup
	}
	if t.NoIni != "" {
		d.NoIni = t.NoIni
	}
	if t.Group != "" {
		d.Group = t.Group
	}
	if t.GroupI18n != "" {
		d.GroupI18n = t.GroupI18n
	}
	if t.Namespace != "" {
		d.Namespace = t.Namespace
	}
	if t.EnvNamespace != "" {
		d.EnvNamespace = t.EnvNamespace
	}
	if t.Command != "" {
		d.Command = t.Command
	}
	if t.CommandI18n != "" {
		d.CommandI18n = t.CommandI18n
	}
	if t.CommandGroup != "" {
		d.CommandGroup = t.CommandGroup
	}
	if t.SubCommandsOptional != "" {
		d.SubCommandsOptional = t.SubCommandsOptional
	}
	if t.Alias != "" {
		d.Alias = t.Alias
	}
	if t.Aliases != "" {
		d.Aliases = t.Aliases
	}
	if t.LongAlias != "" {
		d.LongAlias = t.LongAlias
	}
	if t.LongAliases != "" {
		d.LongAliases = t.LongAliases
	}
	if t.ShortAlias != "" {
		d.ShortAlias = t.ShortAlias
	}
	if t.ShortAliases != "" {
		d.ShortAliases = t.ShortAliases
	}
	if t.PositionalArgs != "" {
		d.PositionalArgs = t.PositionalArgs
	}
	if t.PositionalArgName != "" {
		d.PositionalArgName = t.PositionalArgName
	}
	if t.ArgNameI18n != "" {
		d.ArgNameI18n = t.ArgNameI18n
	}
	if t.ArgDescriptionI18n != "" {
		d.ArgDescriptionI18n = t.ArgDescriptionI18n
	}
	if t.KeyValueDelimiter != "" {
		d.KeyValueDelimiter = t.KeyValueDelimiter
	}
	if t.PassAfterNonOption != "" {
		d.PassAfterNonOption = t.PassAfterNonOption
	}
	if t.Unquote != "" {
		d.Unquote = t.Unquote
	}
	if t.Terminator != "" {
		d.Terminator = t.Terminator
	}

	return d
}
