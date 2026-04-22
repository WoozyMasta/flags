// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

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
	// FlagTagDescription provides short help text.
	FlagTagDescription = "description"
	// FlagTagLongDescription provides extended text (used in man/help contexts).
	FlagTagLongDescription = "long-description"
	// FlagTagNoFlag excludes the field from command-line flag parsing.
	FlagTagNoFlag = "no-flag"
	// FlagTagOptional marks option argument as optional.
	FlagTagOptional = "optional"
	// FlagTagOptionalValue defines fallback value when optional arg is omitted.
	FlagTagOptionalValue = "optional-value"
	// FlagTagOrder sets display/completion priority order for options.
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
	// FlagTagChoice restricts allowed option values (repeatable).
	FlagTagChoice = "choice"
	// FlagTagChoices restricts allowed option values as a delimiter-separated list.
	FlagTagChoices = "choices"
	// FlagTagHidden hides option/group/command from help and completion output.
	FlagTagHidden = "hidden"
	// FlagTagBase sets radix for integer parsing.
	FlagTagBase = "base"
	// FlagTagIniName overrides key name used for INI parse/write.
	FlagTagIniName = "ini-name"
	// FlagTagNoIni excludes the field from INI parse/write.
	FlagTagNoIni = "no-ini"
	// FlagTagGroup turns a nested struct field into an option group.
	FlagTagGroup = "group"
	// FlagTagNamespace prefixes long option names for grouped options.
	FlagTagNamespace = "namespace"
	// FlagTagEnvNamespace prefixes environment variable names for grouped options.
	FlagTagEnvNamespace = "env-namespace"
	// FlagTagCommand turns a field into a subcommand.
	FlagTagCommand = "command"
	// FlagTagSubCommandsOptional makes child subcommands optional.
	FlagTagSubCommandsOptional = "subcommands-optional"
	// FlagTagAlias adds extra command names (repeatable).
	FlagTagAlias = "alias"
	// FlagTagAliases adds extra command names as a delimiter-separated list.
	FlagTagAliases = "aliases"
	// FlagTagPositionalArgs marks a struct as positional arguments container.
	FlagTagPositionalArgs = "positional-args"
	// FlagTagPositionalArgName sets display name for a positional argument.
	FlagTagPositionalArgName = "positional-arg-name"
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
	// Description maps to short help text tag (default: "description").
	Description string
	// LongDescription maps to extended help/man text tag (default: "long-description").
	LongDescription string
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
	// Choice maps to allowed-values tag (default: "choice").
	Choice string
	// Choices maps to multi allowed-values tag (default: "choices").
	Choices string
	// Hidden maps to hide-from-help tag (default: "hidden").
	Hidden string
	// Base maps to integer radix tag (default: "base").
	Base string
	// IniName maps to INI key override tag (default: "ini-name").
	IniName string
	// NoIni maps to INI exclusion tag (default: "no-ini").
	NoIni string
	// Group maps to group declaration tag (default: "group").
	Group string
	// Namespace maps to long-name namespace tag (default: "namespace").
	Namespace string
	// EnvNamespace maps to env-name namespace tag (default: "env-namespace").
	EnvNamespace string
	// Command maps to subcommand declaration tag (default: "command").
	Command string
	// SubCommandsOptional maps to optional-subcommands tag (default: "subcommands-optional").
	SubCommandsOptional string
	// Alias maps to command alias tag (default: "alias").
	Alias string
	// Aliases maps to multi command aliases tag (default: "aliases").
	Aliases string
	// PositionalArgs maps to positional args struct tag (default: "positional-args").
	PositionalArgs string
	// PositionalArgName maps to positional display-name tag (default: "positional-arg-name").
	PositionalArgName string
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
		Description:         prefix + FlagTagDescription,
		LongDescription:     prefix + FlagTagLongDescription,
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
		Choice:              prefix + FlagTagChoice,
		Choices:             prefix + FlagTagChoices,
		Hidden:              prefix + FlagTagHidden,
		Base:                prefix + FlagTagBase,
		IniName:             prefix + FlagTagIniName,
		NoIni:               prefix + FlagTagNoIni,
		Group:               prefix + FlagTagGroup,
		Namespace:           prefix + FlagTagNamespace,
		EnvNamespace:        prefix + FlagTagEnvNamespace,
		Command:             prefix + FlagTagCommand,
		SubCommandsOptional: prefix + FlagTagSubCommandsOptional,
		Alias:               prefix + FlagTagAlias,
		Aliases:             prefix + FlagTagAliases,
		PositionalArgs:      prefix + FlagTagPositionalArgs,
		PositionalArgName:   prefix + FlagTagPositionalArgName,
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
	if t.Description != "" {
		d.Description = t.Description
	}
	if t.LongDescription != "" {
		d.LongDescription = t.LongDescription
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
	if t.Choice != "" {
		d.Choice = t.Choice
	}
	if t.Choices != "" {
		d.Choices = t.Choices
	}
	if t.Hidden != "" {
		d.Hidden = t.Hidden
	}
	if t.Base != "" {
		d.Base = t.Base
	}
	if t.IniName != "" {
		d.IniName = t.IniName
	}
	if t.NoIni != "" {
		d.NoIni = t.NoIni
	}
	if t.Group != "" {
		d.Group = t.Group
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
	if t.SubCommandsOptional != "" {
		d.SubCommandsOptional = t.SubCommandsOptional
	}
	if t.Alias != "" {
		d.Alias = t.Alias
	}
	if t.Aliases != "" {
		d.Aliases = t.Aliases
	}
	if t.PositionalArgs != "" {
		d.PositionalArgs = t.PositionalArgs
	}
	if t.PositionalArgName != "" {
		d.PositionalArgName = t.PositionalArgName
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
