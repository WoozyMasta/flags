// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unicode/utf8"
)

const (
	optionInterfaceUnknown int8 = iota
	optionInterfaceAbsent
	optionInterfacePresent
)

// Option flag information. Contains a description of the option, short and
// long name as well as a default value and whether an argument for this
// flag is optional.
type Option struct {
	// The group which the option belongs to
	group *Group
	io    argIOConfig

	// Parsed struct tags associated with this option.
	tag multiTag

	// The struct field value which the option represents.
	value reflect.Value

	// The description of the option flag. This description is shown
	// automatically in the built-in help.
	Description string

	// Optional i18n key for Description.
	DescriptionI18nKey string

	// The long name of the option. If not "", the option flag can be
	// activated using --<LongName>. Either ShortName or LongName needs
	// to be non-empty.
	LongName string

	// The optional environment default value key name.
	EnvDefaultKey string

	// The optional delimiter string for EnvDefaultKey values.
	EnvDefaultDelim string

	// A name for the value of an option shown in the Help as --flag [ValueName]
	ValueName string

	// Optional i18n key for ValueName.
	ValueNameI18nKey string

	// A mask value to show in the help instead of the default value. This
	// is useful for hiding sensitive information in the help, such as
	// passwords.
	DefaultMask string

	// Cached default literal shown in help/man output.
	defaultLiteral string

	// If non-empty, the option consumes arguments until this exact token is
	// reached (or until end-of-input). This supports find -exec style argument
	// blocks. Only slice and slice-of-slices options are valid with terminator.
	Terminator string

	// Additional long names for the option (without namespace prefix in tags).
	LongAliases []string

	// The default value of the option.
	Default []string

	// The optional value of the option. The optional value is used when
	// the option flag is marked as having an OptionalArgument. This means
	// that when the flag is specified, but no option argument is given,
	// the value of the field this option represents will be set to
	// OptionalValue. This is only valid for non-boolean options.
	OptionalValue []string

	// If non empty, only a certain set of values is allowed for an option.
	Choices []string

	// Relation groups where only one option can be used.
	XorGroups []string

	// Relation groups where all options must be used together.
	AndGroups []string

	// Additional short names for the option.
	ShortAliases []rune

	// The struct field which the option represents.
	field reflect.StructField

	// Display and completion priority within the option's group block.
	// Positive values move options to the top, negative to the bottom,
	// and zero keeps them in the normal sort mode.
	Order int

	// The short name of the option (a single character). If not 0, the
	// option flag can be 'activated' using -<ShortName>. Either ShortName
	// or LongName needs to be non-empty.
	ShortName rune

	// completionHint controls fallback completion mode (file, dir, none).
	completionHint completionHint

	defaultLiteralInitialized bool

	// If true, specifies that the argument to an option flag is optional.
	// When no argument to the flag is specified on the command line, the
	// value of OptionalValue will be set in the field this option represents.
	// This is only valid for non-boolean options.
	OptionalArgument bool

	// If true, the option _must_ be specified on the command line. If the
	// option is not specified, the parser will generate an ErrRequired type
	// error.
	Required bool

	// If true, the option acts as a counter:
	// each flag occurrence increments by 1,
	// and explicit numeric values increment by that amount.
	Counter bool

	// If true, the option is not displayed in the help or man page
	Hidden bool

	// If true, this option participates in immediate parse mode.
	Immediate bool

	// Determines if the option will be always quoted in the INI output
	iniQuote bool

	// Whether the option has been explicitly set by parsing.
	isSet bool

	// Whether the current value was set from defaults.
	isSetDefault bool

	// Whether applying parser defaults is disabled for this option.
	preventDefault bool

	// Whether map/slice values should be cleared before the next set.
	clearReferenceBeforeSet bool

	unmarshalerState    int8
	valueValidatorState int8
}

// SetDescription updates option description used in help/docs output.
func (option *Option) SetDescription(description string) {
	option.Description = description
}

// SetDescriptionI18nKey sets i18n key used to localize option description.
func (option *Option) SetDescriptionI18nKey(key string) {
	option.DescriptionI18nKey = key
}

// SetRequired enables or disables required option validation.
func (option *Option) SetRequired(required bool) {
	option.Required = required
}

// SetHidden controls whether option is shown in help/completion/docs.
func (option *Option) SetHidden(hidden bool) {
	option.Hidden = hidden
}

// SetImmediate enables or disables immediate parse mode for the option.
func (option *Option) SetImmediate(immediate bool) {
	option.Immediate = immediate
}

// IsImmediate reports whether this option is immediate directly or via parent groups.
func (option *Option) IsImmediate() bool {
	if option.Immediate {
		return true
	}

	g := option.group
	for g != nil {
		if g.Immediate {
			return true
		}

		switch parent := g.parent.(type) {
		case *Command:
			g = parent.Group
		case *Group:
			g = parent
		default:
			g = nil
		}
	}

	return false
}

// SetLongName updates canonical long option name.
func (option *Option) SetLongName(name string) error {
	if err := option.validateLongNameLength(name); err != nil {
		return err
	}

	prev := option.LongName
	option.LongName = name

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.LongName = prev
		return err
	}

	option.touchLookupCache()
	return nil
}

// SetShortName updates canonical short option name.
func (option *Option) SetShortName(name rune) error {
	prev := option.ShortName
	option.ShortName = name

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.ShortName = prev
		return err
	}

	option.touchLookupCache()
	return nil
}

// SetLongAliases replaces all long option aliases.
func (option *Option) SetLongAliases(aliases ...string) error {
	for _, alias := range aliases {
		if err := option.validateLongNameLength(alias); err != nil {
			return err
		}
	}

	prev := append([]string(nil), option.LongAliases...)
	option.LongAliases = append(option.LongAliases[:0], aliases...)

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.LongAliases = prev
		return err
	}

	option.touchLookupCache()
	return nil
}

// AddLongAlias appends one long option alias.
func (option *Option) AddLongAlias(alias string) error {
	if err := option.validateLongNameLength(alias); err != nil {
		return err
	}

	prev := append([]string(nil), option.LongAliases...)
	option.LongAliases = append(option.LongAliases, alias)

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.LongAliases = prev
		return err
	}

	option.touchLookupCache()
	return nil
}

// SetShortAliases replaces all short option aliases.
func (option *Option) SetShortAliases(aliases ...rune) error {
	prev := append([]rune(nil), option.ShortAliases...)
	option.ShortAliases = append(option.ShortAliases[:0], aliases...)

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.ShortAliases = prev
		return err
	}

	option.touchLookupCache()
	return nil
}

// AddShortAlias appends one short option alias.
func (option *Option) AddShortAlias(alias rune) error {
	prev := append([]rune(nil), option.ShortAliases...)
	option.ShortAliases = append(option.ShortAliases, alias)

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.ShortAliases = prev
		return err
	}

	option.touchLookupCache()
	return nil
}

// SetDefault replaces option default values.
func (option *Option) SetDefault(values ...string) {
	option.Default = append(option.Default[:0], values...)
	option.defaultLiteralInitialized = false
}

// SetDefaultMask sets displayed default mask used in help/docs.
func (option *Option) SetDefaultMask(mask string) {
	option.DefaultMask = mask
}

// SetValueNameI18nKey sets i18n key used to localize option value placeholder.
func (option *Option) SetValueNameI18nKey(key string) {
	option.ValueNameI18nKey = key
}

// SetEnv sets environment key and optional split delimiter for env value.
func (option *Option) SetEnv(key string, delim string) error {
	prevKey := option.EnvDefaultKey
	prevDelim := option.EnvDefaultDelim
	option.EnvDefaultKey = key
	option.EnvDefaultDelim = delim

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.EnvDefaultKey = prevKey
		option.EnvDefaultDelim = prevDelim
		return err
	}

	return nil
}

// SetBase configures numeric radix for integer parse/format.
// Use 0 for automatic base detection from prefixes (Go-style).
func (option *Option) SetBase(base int) error {
	if base != 0 && (base < 2 || base > 36) {
		return newErrorf(ErrInvalidTag, "invalid base %d; expected 0 or range 2..36", base)
	}

	option.tag.Set(FlagTagBase, strconv.Itoa(base))
	return nil
}

// SetKeyValueDelimiter configures map key/value separator (default is ":").
func (option *Option) SetKeyValueDelimiter(delimiter string) {
	option.tag.Set(FlagTagKeyValueDelimiter, delimiter)
}

// SetUnquote controls automatic unquoting for quoted string arguments.
func (option *Option) SetUnquote(enabled bool) {
	option.tag.Set(FlagTagUnquote, strconv.FormatBool(enabled))
}

// SetIniName overrides key name used for INI read/write.
func (option *Option) SetIniName(name string) {
	option.tag.Set(FlagTagIniName, name)
}

// SetNoIni enables or disables INI read/write participation for the option.
func (option *Option) SetNoIni(disabled bool) {
	if disabled {
		option.tag.Set(FlagTagNoIni, "true")
		return
	}

	option.tag.Set(FlagTagNoIni, "")
}

// SetAutoEnv enables or disables env-key derivation from long option name.
// When enabled and env key is currently empty, a key is derived automatically.
func (option *Option) SetAutoEnv(enabled bool) error {
	prevTag := option.tag.Get(FlagTagAutoEnv)
	prevKey := option.EnvDefaultKey
	option.tag.Set(FlagTagAutoEnv, strconv.FormatBool(enabled))

	if !enabled || option.EnvDefaultKey != "" {
		return nil
	}

	if option.LongName == "" {
		option.tag.Set(FlagTagAutoEnv, prevTag)
		return newErrorf(
			ErrInvalidTag,
			"auto env for flag `%s' requires a long flag name",
			option.shortAndLongName(),
		)
	}

	option.EnvDefaultKey = autoEnvKeyFromLongName(option.LongName)

	if err := option.validateProgrammaticUpdate(); err != nil {
		option.tag.Set(FlagTagAutoEnv, prevTag)
		option.EnvDefaultKey = prevKey
		return err
	}

	return nil
}

// SetChoices replaces allowed option values.
func (option *Option) SetChoices(values ...string) {
	option.Choices = append(option.Choices[:0], values...)
}

// SetXorGroups replaces mutually exclusive relation groups for this option.
func (option *Option) SetXorGroups(groups ...string) {
	option.XorGroups = append(option.XorGroups[:0], groups...)
}

// SetAndGroups replaces all-or-none relation groups for this option.
func (option *Option) SetAndGroups(groups ...string) {
	option.AndGroups = append(option.AndGroups[:0], groups...)
}

// SetOptional configures optional argument behavior and fallback value(s).
func (option *Option) SetOptional(optional bool, values ...string) {
	option.OptionalArgument = optional
	option.OptionalValue = append(option.OptionalValue[:0], values...)
}

// SetValueName updates help placeholder for option argument.
func (option *Option) SetValueName(name string) {
	option.ValueName = name
}

// SetOrder updates help/completion sorting priority for this option.
func (option *Option) SetOrder(order int) {
	option.Order = order
}

// SetTerminator configures terminated-argument mode for slice options.
func (option *Option) SetTerminator(terminator string) {
	option.Terminator = terminator
}

// LongNameWithNamespace returns the option's long name with the group namespaces
// prepended by walking up the option's group tree. Namespaces and the long name
// itself are separated by the parser's namespace delimiter. If the long name is
// empty an empty string is returned.
func (option *Option) LongNameWithNamespace() string {
	if len(option.LongName) == 0 {
		return ""
	}

	return option.longNameWithNamespace(option.LongName)
}

// LongAliasesWithNamespace returns option long aliases with group namespaces applied.
func (option *Option) LongAliasesWithNamespace() []string {
	if len(option.LongAliases) == 0 {
		return nil
	}

	out := make([]string, 0, len(option.LongAliases))
	for _, alias := range option.LongAliases {
		if alias == "" {
			continue
		}
		out = append(out, option.longNameWithNamespace(alias))
	}

	return out
}

// EnvKeyWithNamespace returns the option's env key with the group namespaces
// prepended by walking up the option's group tree. Namespaces and the env key
// itself are separated by the parser's namespace delimiter. If the env key is
// empty an empty string is returned.
func (option *Option) EnvKeyWithNamespace() string {
	if len(option.EnvDefaultKey) == 0 {
		return ""
	}

	// fetch the namespace delimiter from the parser which is always at the
	// end of the group hierarchy
	namespaceDelimiter := ""
	envPrefix := ""
	g := option.group

	for {
		if p, ok := g.parent.(*Parser); ok {
			namespaceDelimiter = p.EnvNamespaceDelimiter
			envPrefix = p.EnvPrefix

			break
		}

		switch i := g.parent.(type) {
		case *Command:
			g = i.Group
		case *Group:
			g = i
		}
	}

	// concatenate long name with namespace
	key := option.EnvDefaultKey
	g = option.group

	for g != nil {
		if g.EnvNamespace != "" {
			key = g.EnvNamespace + namespaceDelimiter + key
		}

		switch i := g.parent.(type) {
		case *Command:
			g = i.Group
		case *Group:
			g = i
		case *Parser:
			g = nil
		}
	}

	if envPrefix != "" {
		key = envPrefix + namespaceDelimiter + key
	}

	return key
}

// String converts an option to a human friendly readable string describing the
// option.
func (option *Option) String() string {
	var s string
	var short string

	if option.ShortName != 0 {
		data := make([]byte, utf8.RuneLen(option.ShortName))
		utf8.EncodeRune(data, option.ShortName)
		short = string(data)

		if len(option.LongName) != 0 {
			s = fmt.Sprintf("%s%s, %s%s",
				string(defaultShortOptDelimiter), short,
				defaultLongOptDelimiter, option.LongNameWithNamespace())
		} else {
			s = fmt.Sprintf("%s%s", string(defaultShortOptDelimiter), short)
		}
	} else if len(option.LongName) != 0 {
		s = fmt.Sprintf("%s%s", defaultLongOptDelimiter, option.LongNameWithNamespace())
	}

	return s
}

// Value returns the option value as an interface{}.
func (option *Option) Value() any {
	return option.value.Interface()
}

// Field returns the reflect struct field of the option.
func (option *Option) Field() reflect.StructField {
	return option.field
}

// IsSet returns true if option has been set
func (option *Option) IsSet() bool {
	return option.isSet
}

// IsSetDefault returns true if option has been set via the default option tag
func (option *Option) IsSetDefault() bool {
	return option.isSetDefault
}

// Set the value of an option to the specified value. An error will be returned
// if the specified value could not be converted to the corresponding option
// value type.
func (option *Option) Set(value *string) error {
	kind := option.value.Type().Kind()

	if (kind == reflect.Map || kind == reflect.Slice) && option.clearReferenceBeforeSet {
		option.empty()
	}

	option.isSet = true
	option.preventDefault = true
	option.clearReferenceBeforeSet = false

	if value != nil {
		if option.io.role != "" {
			normalized, err := normalizeIOValue(option.io, *value)
			if err != nil {
				return err
			}
			value = &normalized
		}
		if err := option.validateChoice(*value); err != nil {
			return err
		}
	}

	if option.isFunc() {
		return option.call(value)
	} else if value != nil {
		return convert(*value, option.value, option.tag)
	}

	return convert("", option.value, option.tag)
}

func (option *Option) applyCounterDelta(delta uint64) error {
	kind := option.value.Kind()
	option.isSet = true
	option.preventDefault = true
	option.clearReferenceBeforeSet = false

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		current := option.value.Int()
		if delta > uint64(math.MaxInt64) {
			return ErrCounterIncrementTooLarge
		}
		next := current + int64(delta)
		if option.value.OverflowInt(next) || next < current {
			return ErrCounterOverflow
		}
		option.value.SetInt(next)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		current := option.value.Uint()
		next := current + delta
		if option.value.OverflowUint(next) || next < current {
			return ErrCounterOverflow
		}
		option.value.SetUint(next)
		return nil

	default:
		return ErrCounterInvalidType
	}
}

// SetTerminated sets all values collected for a terminated option.
// For []T options this replaces the current slice value.
// For [][]T options this appends one collected argument batch.
func (option *Option) SetTerminated(values []string) error {
	tp := option.value.Type()

	if tp.Kind() != reflect.Slice {
		return newErrorf(ErrInvalidTag,
			"terminated flag `%s' must be a slice or slice of slices",
			option.shortAndLongName())
	}

	option.isSet = true
	option.preventDefault = true
	option.clearReferenceBeforeSet = false

	elemTp := tp.Elem()

	if elemTp.Kind() == reflect.Slice {
		elemVal := reflect.New(elemTp).Elem()
		elemVal.Set(reflect.MakeSlice(elemTp, 0, len(values)))

		for _, v := range values {
			if err := option.validateChoice(v); err != nil {
				return err
			}
			if err := convert(v, elemVal, option.tag); err != nil {
				return err
			}
		}

		option.value.Set(reflect.Append(option.value, elemVal))
		return nil
	}

	option.empty()

	for _, v := range values {
		if err := option.validateChoice(v); err != nil {
			return err
		}
		if err := convert(v, option.value, option.tag); err != nil {
			return err
		}
	}

	return nil
}
