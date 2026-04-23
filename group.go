// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Group represents an option group. Option groups can be used to logically
// group options together under a description. Groups are only used to provide
// more structure to options both for the user (as displayed in the help message)
// and for you, since groups can be nested.
type Group struct {
	// The parent of the group or nil if it has no parent
	parent any

	// The user-provided struct pointer backing this group.
	data any

	// A short description of the group. The
	// short description is primarily used in the built-in generated help
	// message
	ShortDescription string

	// A long description of the group. The long
	// description is primarily used to present information on commands
	// (Command embeds Group) in the built-in generated help and man pages.
	LongDescription string

	// The namespace of the group
	Namespace string

	// The environment namespace of the group
	EnvNamespace string

	// All the options in the group
	options []*Option

	// All the subgroups
	groups []*Group

	// If true, the group is not displayed in the help or man page
	Hidden bool

	// If true, options in this group bypass required checks and command execution.
	Immediate bool

	// Whether the group represents the built-in help group
	isBuiltinHelp bool
}

type scanHandler func(reflect.Value, *reflect.StructField) (bool, error)

func (g *Group) touchLookupCache() {
	if p := g.parser(); p != nil {
		p.invalidateLookupCache()
	}
}

// AddGroup adds a new group to the command with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the group.
func (g *Group) AddGroup(shortDescription string, longDescription string, data any) (*Group, error) {
	group := newGroup(shortDescription, longDescription, data)

	group.parent = g

	if err := group.scan(); err != nil {
		return nil, err
	}

	g.groups = append(g.groups, group)

	if p := g.parser(); p != nil {
		p.invalidateLookupCache()
	}

	return group, nil
}

// SetShortDescription updates group short description used in help/docs output.
func (g *Group) SetShortDescription(description string) {
	g.ShortDescription = description
}

// SetLongDescription updates group long description used in docs output.
func (g *Group) SetLongDescription(description string) {
	g.LongDescription = description
}

// SetNamespace updates long-option namespace prefix for child options.
func (g *Group) SetNamespace(namespace string) {
	g.Namespace = namespace
	g.touchLookupCache()
}

// SetEnvNamespace updates env-variable namespace prefix for child options.
func (g *Group) SetEnvNamespace(namespace string) {
	g.EnvNamespace = namespace
	g.touchLookupCache()
}

// SetHidden controls group visibility in help/completion/docs.
func (g *Group) SetHidden(hidden bool) {
	g.Hidden = hidden
}

// SetImmediate controls immediate parse behavior for this group subtree.
func (g *Group) SetImmediate(immediate bool) {
	g.Immediate = immediate
}

// AddOption adds a new option to this group.
func (g *Group) AddOption(option *Option, data any) {
	option.value = reflect.ValueOf(data)
	option.group = g
	g.options = append(g.options, option)

	if p := g.parser(); p != nil {
		p.invalidateLookupCache()
	}
}

// Groups returns the list of groups embedded in this group.
func (g *Group) Groups() []*Group {
	return g.groups
}

// Options returns the list of options in this group.
func (g *Group) Options() []*Option {
	ret := make([]*Option, len(g.options))
	copy(ret, g.options)

	if p := g.parser(); p != nil {
		return p.sortedOptions(ret)
	}

	return ret
}

// Data returns the user-provided struct pointer backing this group.
func (g *Group) Data() any {
	return g.data
}

// Find locates the subgroup with the given short description and returns it.
// If no such group can be found Find will return nil. Note that the description
// is matched case insensitively.
func (g *Group) Find(shortDescription string) *Group {
	lshortDescription := strings.ToLower(shortDescription)

	var ret *Group

	g.eachGroup(func(gg *Group) {
		if gg != g && strings.ToLower(gg.ShortDescription) == lshortDescription {
			ret = gg
		}
	})

	return ret
}

func (g *Group) findOption(matcher func(*Option) bool) (option *Option) {
	g.eachGroup(func(g *Group) {
		for _, opt := range g.options {
			if option == nil && matcher(opt) {
				option = opt
			}
		}
	})

	return option
}

// FindOptionByLongName finds an option that is part of the group, or any of its
// subgroups, by matching its long name (including the option namespace).
func (g *Group) FindOptionByLongName(longName string) *Option {
	return g.findOption(func(option *Option) bool {
		if option.LongNameWithNamespace() == longName {
			return true
		}
		return slices.Contains(option.LongAliasesWithNamespace(), longName)
	})
}

// FindOptionByShortName finds an option that is part of the group, or any of
// its subgroups, by matching its short name.
func (g *Group) FindOptionByShortName(shortName rune) *Option {
	return g.findOption(func(option *Option) bool {
		if option.ShortName == shortName {
			return true
		}
		return slices.Contains(option.ShortAliases, shortName)
	})
}

func newGroup(shortDescription string, longDescription string, data any) *Group {
	return &Group{
		ShortDescription: shortDescription,
		LongDescription:  longDescription,

		data: data,
	}
}

func (g *Group) parser() *Parser {
	var parent = g.parent

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

func (g *Group) optionByName(name string, namematch func(*Option, string) bool) *Option {
	prio := 0
	var retopt *Option

	g.eachGroup(func(g *Group) {
		for _, opt := range g.options {
			if namematch != nil && namematch(opt, name) && prio < 4 {
				retopt = opt
				prio = 4
			}

			if name == opt.field.Name && prio < 3 {
				retopt = opt
				prio = 3
			}

			if name == opt.LongNameWithNamespace() && prio < 2 {
				retopt = opt
				prio = 2
			}

			if slices.Contains(opt.LongAliasesWithNamespace(), name) && prio < 2 {
				retopt = opt
				prio = 2
			}

			if opt.ShortName != 0 && name == string(opt.ShortName) && prio < 1 {
				retopt = opt
				prio = 1
			}

			if containsShortAliasByString(opt.ShortAliases, name) && prio < 1 {
				retopt = opt
				prio = 1
			}
		}
	})

	return retopt
}

func containsShortAliasByString(aliases []rune, value string) bool {
	for _, alias := range aliases {
		if string(alias) == value {
			return true
		}
	}
	return false
}

func (g *Group) showInHelp() bool {
	if g.Hidden {
		return false
	}
	for _, opt := range g.options {
		if opt.showInHelp() {
			return true
		}
	}
	return false
}

func (g *Group) eachGroup(f func(*Group)) {
	f(g)

	for _, gg := range g.groups {
		gg.eachGroup(f)
	}
}

func parseBoolTagValue(raw string) (bool, bool, error) {
	if raw == "" {
		return false, false, nil
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "t", "yes", "y", "on":
		return true, true, nil
	case "0", "false", "f", "no", "n", "off":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("unsupported boolean value %q", raw)
	}
}

func parseStructBoolTag(mtag multiTag, tagName string, fieldName string) (bool, bool, error) {
	raw := mtag.Get(tagName)
	value, set, err := parseBoolTagValue(raw)
	if err != nil {
		return false, false, newErrorf(ErrInvalidTag,
			"invalid boolean value `%s' for tag `%s' on field `%s'",
			raw, tagName, fieldName)
	}

	return value, set, nil
}

func autoEnvKeyFromLongName(longName string) string {
	upper := strings.ToUpper(longName)

	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}

		return '_'
	}, upper)
}

func parseOptionShortAliases(fieldName string, aliases []string) ([]rune, error) {
	if len(aliases) == 0 {
		return nil, nil
	}

	shortAliases := make([]rune, 0, len(aliases))

	for _, alias := range aliases {
		rc := utf8.RuneCountInString(alias)
		if rc != 1 {
			return nil, newErrorf(
				ErrShortNameTooLong,
				"short names can only be 1 character long, not `%s' (field `%s')",
				alias,
				fieldName,
			)
		}

		r, _ := utf8.DecodeRuneInString(alias)
		shortAliases = append(shortAliases, r)
	}

	return shortAliases, nil
}

func (g *Group) scanStruct(realval reflect.Value, sfield *reflect.StructField, handler scanHandler) error {
	stype := realval.Type()

	if sfield != nil {
		if ok, err := handler(realval, sfield); err != nil {
			return err
		} else if ok {
			return nil
		}
	}

	for i := 0; i < stype.NumField(); i++ {
		field := stype.Field(i)

		// PkgName is set only for non-exported fields, which we ignore
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		mtag := newMultiTag(string(field.Tag))

		if err := mtag.Parse(); err != nil {
			return err
		}
		if p := g.parser(); p != nil {
			p.normalizeStructTag(&mtag)
		}

		// Skip fields with the no-flag tag
		if mtag.Get(FlagTagNoFlag) != "" {
			continue
		}

		// Dive deep into structs or pointers to structs
		kind := field.Type.Kind()
		fld := realval.Field(i)

		if kind == reflect.Struct {
			if err := g.scanStruct(fld, &field, handler); err != nil {
				return err
			}
		} else if kind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			flagCountBefore := len(g.options) + len(g.groups)

			if fld.IsNil() {
				fld = reflect.New(fld.Type().Elem())
			}

			if err := g.scanStruct(reflect.Indirect(fld), &field, handler); err != nil {
				return err
			}

			if len(g.options)+len(g.groups) != flagCountBefore {
				realval.Field(i).Set(fld)
			}
		}

		longname := mtag.Get(FlagTagLong)
		shortname := mtag.Get(FlagTagShort)

		if longname != "" {
			maxLong := 0
			if p := g.parser(); p != nil {
				maxLong = p.MaxLongNameLength
			}

			if maxLong > 0 && utf8.RuneCountInString(longname) > maxLong {
				return newErrorf(ErrInvalidTag,
					"long flag name `%s' exceeds max length %d (use SetMaxLongNameLength to override)",
					longname, maxLong)
			}
		}

		// Need at least either a short or long name
		if longname == "" && shortname == "" && mtag.Get(FlagTagIniName) == "" {
			continue
		}

		short := rune(0)
		rc := utf8.RuneCountInString(shortname)

		if rc > 1 {
			return newErrorf(ErrShortNameTooLong,
				"short names can only be 1 character long, not `%s'",
				shortname)
		} else if rc == 1 {
			short, _ = utf8.DecodeRuneInString(shortname)
		}

		description := mtag.Get(FlagTagDescription)
		delimiter := parserTagListDelimiter(g.parser())
		def, err := collectTagValues(mtag, FlagTagDefault, FlagTagDefaults, field.Name, delimiter)
		if err != nil {
			return err
		}

		optionalValue := mtag.GetMany(FlagTagOptionalValue)
		valueName := mtag.Get(FlagTagValueName)
		defaultMask := mtag.Get(FlagTagDefaultMask)
		order := 0
		if rawOrder := mtag.Get(FlagTagOrder); rawOrder != "" {
			parsedOrder, convErr := strconv.Atoi(rawOrder)
			if convErr != nil {
				return newErrorf(ErrInvalidTag,
					"invalid integer value `%s' for tag `%s' on field `%s'",
					rawOrder, FlagTagOrder, field.Name)
			}
			order = parsedOrder
		}

		optional, _, err := parseStructBoolTag(mtag, FlagTagOptional, field.Name)
		if err != nil {
			return err
		}

		required, _, err := parseStructBoolTag(mtag, FlagTagRequired, field.Name)
		if err != nil {
			return err
		}

		choices, err := collectTagValues(mtag, FlagTagChoice, FlagTagChoices, field.Name, delimiter)
		if err != nil {
			return err
		}
		longAliases, err := collectTagValues(mtag, FlagTagLongAlias, FlagTagLongAliases, field.Name, delimiter)
		if err != nil {
			return err
		}
		if len(longAliases) > 0 {
			maxLong := 0
			if p := g.parser(); p != nil {
				maxLong = p.MaxLongNameLength
			}
			for _, alias := range longAliases {
				if maxLong > 0 && utf8.RuneCountInString(alias) > maxLong {
					return newErrorf(
						ErrInvalidTag,
						"long flag alias `%s' exceeds max length %d (use SetMaxLongNameLength to override)",
						alias,
						maxLong,
					)
				}
			}
		}
		shortAliasValues, err := collectTagValues(mtag, FlagTagShortAlias, FlagTagShortAliases, field.Name, delimiter)
		if err != nil {
			return err
		}
		shortAliases, err := parseOptionShortAliases(field.Name, shortAliasValues)
		if err != nil {
			return err
		}
		hidden, _, err := parseStructBoolTag(mtag, FlagTagHidden, field.Name)
		if err != nil {
			return err
		}
		immediate, _, err := parseStructBoolTag(mtag, FlagTagImmediate, field.Name)
		if err != nil {
			return err
		}

		envKey := mtag.Get(FlagTagEnv)
		autoEnv, hasAutoEnvTag, err := parseStructBoolTag(mtag, FlagTagAutoEnv, field.Name)
		if err != nil {
			return err
		}

		if p := g.parser(); p != nil && (p.Options&EnvProvisioning) != None && !hasAutoEnvTag {
			autoEnv = true
		}

		if envKey == "" && autoEnv {
			if longname == "" {
				return newErrorf(ErrInvalidTag,
					"auto env for field `%s' requires a long flag name",
					field.Name)
			}
			envKey = autoEnvKeyFromLongName(longname)
		}

		option := &Option{
			Description:      description,
			ShortName:        short,
			ShortAliases:     shortAliases,
			LongName:         longname,
			LongAliases:      longAliases,
			Default:          def,
			EnvDefaultKey:    envKey,
			EnvDefaultDelim:  mtag.Get(FlagTagEnvDelim),
			OptionalArgument: optional,
			OptionalValue:    optionalValue,
			Required:         required,
			ValueName:        valueName,
			DefaultMask:      defaultMask,
			Choices:          choices,
			Hidden:           hidden,
			Immediate:        immediate,
			Terminator:       mtag.Get(FlagTagTerminator),
			Order:            order,

			group: g,

			field: field,
			value: realval.Field(i),
			tag:   mtag,
		}

		if option.isBool() && option.Default != nil {
			return newErrorf(ErrInvalidTag,
				"boolean flag `%s' may not have default values, they always default to `false' and can only be turned on",
				option.shortAndLongName())
		}

		if option.isTerminated() {
			optionType := option.value.Type()
			if optionType.Kind() != reflect.Slice {
				return newErrorf(ErrInvalidTag,
					"terminated flag `%s' must be a slice or slice of slices",
					option.shortAndLongName())
			}
		}

		if defaults, ok, err := dynamicOptionDefault(option.value); err != nil {
			return newErrorf(ErrMarshal,
				"could not get dynamic default for flag `%s': %v",
				option.shortAndLongName(),
				err)
		} else if ok {
			option.Default = defaults
		}

		g.options = append(g.options, option)
	}

	return nil
}

func (g *Group) checkForDuplicateFlags() *Error {
	shortNames := make(map[rune]*Option)
	longNames := make(map[string]*Option)

	var duplicateError *Error

	g.eachGroup(func(g *Group) {
		for _, option := range g.options {
			if option.LongName != "" {
				longName := option.LongNameWithNamespace()

				if otherOption, ok := longNames[longName]; ok {
					duplicateError = newErrorf(ErrDuplicatedFlag, "option `%s' uses the same long name as option `%s'", option, otherOption)
					return
				}
				longNames[longName] = option
			}
			for _, alias := range option.LongAliasesWithNamespace() {
				if otherOption, ok := longNames[alias]; ok {
					duplicateError = newErrorf(ErrDuplicatedFlag, "option `%s' uses the same long alias `%s' as option `%s'", option, alias, otherOption)
					return
				}
				longNames[alias] = option
			}
			if option.ShortName != 0 {
				if otherOption, ok := shortNames[option.ShortName]; ok {
					duplicateError = newErrorf(ErrDuplicatedFlag, "option `%s' uses the same short name as option `%s'", option, otherOption)
					return
				}
				shortNames[option.ShortName] = option
			}
			for _, alias := range option.ShortAliases {
				if otherOption, ok := shortNames[alias]; ok {
					duplicateError = newErrorf(ErrDuplicatedFlag, "option `%s' uses the same short alias `%c' as option `%s'", option, alias, otherOption)
					return
				}
				shortNames[alias] = option
			}
		}
	})

	return duplicateError
}

func (g *Group) scanSubGroupHandler(realval reflect.Value, sfield *reflect.StructField) (bool, error) {
	mtag := newMultiTag(string(sfield.Tag))

	if err := mtag.Parse(); err != nil {
		return true, err
	}
	if p := g.parser(); p != nil {
		p.normalizeStructTag(&mtag)
	}

	subgroup := mtag.Get(FlagTagGroup)

	if len(subgroup) != 0 {
		var ptrval reflect.Value

		if realval.Kind() == reflect.Ptr {
			ptrval = realval

			if ptrval.IsNil() {
				ptrval.Set(reflect.New(ptrval.Type()))
			}
		} else {
			ptrval = realval.Addr()
		}

		description := mtag.Get(FlagTagDescription)

		group, err := g.AddGroup(subgroup, description, ptrval.Interface())

		if err != nil {
			return true, err
		}

		group.Namespace = mtag.Get(FlagTagNamespace)
		group.EnvNamespace = mtag.Get(FlagTagEnvNamespace)
		hidden, _, err := parseStructBoolTag(mtag, FlagTagHidden, sfield.Name)
		if err != nil {
			return true, err
		}
		group.Hidden = hidden
		immediate, _, err := parseStructBoolTag(mtag, FlagTagImmediate, sfield.Name)
		if err != nil {
			return true, err
		}
		group.Immediate = immediate

		return true, nil
	}

	return false, nil
}

func (g *Group) scanType(handler scanHandler) error {
	// Get all the public fields in the data struct
	ptrval := reflect.ValueOf(g.data)

	if ptrval.Type().Kind() != reflect.Ptr {
		panic(ErrNotPointerToStruct)
	}

	stype := ptrval.Type().Elem()

	if stype.Kind() != reflect.Struct {
		panic(ErrNotPointerToStruct)
	}

	realval := reflect.Indirect(ptrval)

	if err := g.scanStruct(realval, nil, handler); err != nil {
		return err
	}

	if err := g.checkForDuplicateFlags(); err != nil {
		return err
	}

	return nil
}

func (g *Group) scan() error {
	return g.scanType(g.scanSubGroupHandler)
}

func (g *Group) groupByName(name string) *Group {
	if len(name) == 0 {
		return g
	}

	return g.Find(name)
}

func dynamicOptionDefault(value reflect.Value) ([]string, bool, error) {
	v := value

	for v.IsValid() && v.CanInterface() {
		if provider, ok := v.Interface().(DefaultProvider); ok {
			def, err := provider.Default()
			return def, true, err
		}

		if !v.CanAddr() {
			break
		}

		v = v.Addr()
	}

	return nil, false, nil
}
