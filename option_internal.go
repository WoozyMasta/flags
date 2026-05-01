// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"
)

func (option *Option) touchLookupCache() {
	if option.group == nil {
		return
	}

	if p := option.group.parser(); p != nil {
		p.invalidateLookupCache()
	}
}

func (option *Option) parser() *Parser {
	if option.group == nil {
		return nil
	}

	return option.group.parser()
}

func (option *Option) validateLongNameLength(name string) error {
	if option.group == nil || name == "" {
		return nil
	}

	p := option.group.parser()
	if p == nil || p.MaxLongNameLength == 0 {
		return nil
	}

	if utf8.RuneCountInString(name) > p.MaxLongNameLength {
		return newErrorf(
			ErrInvalidTag,
			"long flag name `%s` exceeds max length %d (use SetMaxLongNameLength to override)",
			name,
			p.MaxLongNameLength,
		)
	}

	return nil
}

func (option *Option) validateProgrammaticUpdate() error {
	if option.group == nil {
		return nil
	}

	p := option.group.parser()
	if p == nil {
		return nil
	}

	if err := p.validateDuplicateFlags(); err != nil {
		return err
	}

	if err := p.validateDuplicateEnvKeys(); err != nil {
		return err
	}

	return nil
}

func (option *Option) longNameWithNamespace(name string) string {
	if name == "" {
		return ""
	}

	// fetch the namespace delimiter from the parser which is always at the
	// end of the group hierarchy
	namespaceDelimiter := ""
	g := option.group

	for {
		if p, ok := g.parent.(*Parser); ok {
			namespaceDelimiter = p.NamespaceDelimiter

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
	longName := name
	g = option.group

	for g != nil {
		if g.Namespace != "" {
			longName = g.Namespace + namespaceDelimiter + longName
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

	return longName
}

func (option *Option) setDefault(value *string) error {
	if option.preventDefault {
		return nil
	}

	if err := option.Set(value); err != nil {
		return err
	}

	option.isSetDefault = true
	option.preventDefault = false

	return nil
}

func (option *Option) showInHelp() bool {
	return !option.Hidden && (option.ShortName != 0 || len(option.LongName) != 0)
}

func (option *Option) canArgument() bool {
	if u := option.isUnmarshaler(); u != nil {
		return true
	}

	return !option.isBool()
}

func (option *Option) isTerminated() bool {
	return option.Terminator != ""
}

func (option *Option) hasRelationGroups() bool {
	return len(option.XorGroups) > 0 || len(option.AndGroups) > 0
}

func (option *Option) emptyValue() reflect.Value {
	tp := option.value.Type()

	if tp.Kind() == reflect.Map {
		return reflect.MakeMap(tp)
	}

	return reflect.Zero(tp)
}

func (option *Option) empty() {
	if !option.isFunc() {
		option.value.Set(option.emptyValue())
	}
}

func (option *Option) clearDefault() error {
	if option.preventDefault {
		return nil
	}

	usedDefault := option.Default

	if envKey := option.EnvKeyWithNamespace(); envKey != "" {
		if value, ok := os.LookupEnv(envKey); ok {
			if option.EnvDefaultDelim != "" {
				usedDefault = strings.Split(value, option.EnvDefaultDelim)
			} else {
				usedDefault = []string{value}
			}
		}
	}

	option.isSetDefault = true

	if len(usedDefault) > 0 {
		option.empty()

		for _, d := range usedDefault {
			err := option.setDefault(&d)

			if err != nil {
				return err
			}
		}
	} else {
		tp := option.value.Type()

		switch tp.Kind() {
		case reflect.Map:
			if option.value.IsNil() {
				option.empty()
			}
		case reflect.Slice:
			if option.value.IsNil() {
				option.empty()
			}
		}
	}

	return nil
}

func (option *Option) valueIsDefault() bool {
	// Check if the value of the option corresponds to its
	// default value
	emptyval := option.emptyValue()

	checkvalptr := reflect.New(emptyval.Type())
	checkval := reflect.Indirect(checkvalptr)

	checkval.Set(emptyval)

	if len(option.Default) != 0 {
		for _, v := range option.Default {
			if err := convert(v, checkval, option.tag); err != nil {
				return false
			}
		}
	}

	return reflect.DeepEqual(option.value.Interface(), checkval.Interface())
}

func (option *Option) isUnmarshaler() Unmarshaler {
	if option.unmarshalerState == optionInterfaceAbsent {
		return nil
	}

	v := option.value

	for v.CanInterface() {
		i := v.Interface()

		if u, ok := i.(Unmarshaler); ok {
			option.unmarshalerState = optionInterfacePresent
			return u
		}

		if !v.CanAddr() {
			break
		}

		v = v.Addr()
	}

	option.unmarshalerState = optionInterfaceAbsent
	return nil
}

func (option *Option) isValueValidator() ValueValidator {
	if option.valueValidatorState == optionInterfaceAbsent {
		return nil
	}

	v := option.value

	for v.CanInterface() {
		i := v.Interface()

		if u, ok := i.(ValueValidator); ok {
			option.valueValidatorState = optionInterfacePresent
			return u
		}

		if !v.CanAddr() {
			break
		}

		v = v.Addr()
	}

	option.valueValidatorState = optionInterfaceAbsent
	return nil
}

func (option *Option) isBool() bool {
	tp := option.value.Type()

	for {
		switch tp.Kind() {
		case reflect.Slice, reflect.Pointer:
			tp = tp.Elem()
		case reflect.Bool:
			return true
		case reflect.Func:
			return tp.NumIn() == 0
		default:
			return false
		}
	}
}

func (option *Option) isSignedNumber() bool {
	tp := option.value.Type()

	for {
		switch tp.Kind() {
		case reflect.Slice, reflect.Pointer:
			tp = tp.Elem()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
			return true
		default:
			return false
		}
	}
}

func (option *Option) isFunc() bool {
	return option.value.Type().Kind() == reflect.Func
}

func (option *Option) isEmpty() bool {
	switch option.value.Kind() {
	case reflect.String, reflect.Slice, reflect.Map:
		return option.value.Len() == 0
	case reflect.Pointer, reflect.Interface, reflect.Func:
		return option.value.IsNil()
	default:
		return option.value.IsZero()
	}
}

func (option *Option) isRepeatableValue() bool {
	return isRepeatableOptionValue(option.value)
}

func isRepeatableOptionValue(value reflect.Value) bool {
	tp := value.Type()

	for tp.Kind() == reflect.Pointer {
		tp = tp.Elem()
	}

	switch tp.Kind() {
	case reflect.Map, reflect.Slice:
		return true
	default:
		return false
	}
}

func (option *Option) requiredValueCount() int {
	value := option.value

	for {
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if value.IsNil() {
				return 0
			}
			value = value.Elem()
		case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
			return value.Len()
		default:
			if value.IsZero() {
				return 0
			}
			return 1
		}
	}
}

func (option *Option) call(value *string) error {
	var retval []reflect.Value

	if value == nil {
		retval = option.value.Call(nil)
	} else {
		tp := option.value.Type().In(0)

		val := reflect.New(tp)
		val = reflect.Indirect(val)

		if err := convert(*value, val, option.tag); err != nil {
			return err
		}

		retval = option.value.Call([]reflect.Value{val})
	}

	if len(retval) == 1 && retval[0].Type() == reflect.TypeFor[error]() {
		if retval[0].Interface() == nil {
			return nil
		}

		return retval[0].Interface().(error)
	}

	return nil
}

func (option *Option) updateDefaultLiteral() {
	defs := option.Default
	def := ""

	if len(defs) == 0 && option.canArgument() {
		var showdef bool

		switch option.field.Type.Kind() {
		case reflect.Func, reflect.Pointer:
			showdef = !option.value.IsNil()
		case reflect.Slice, reflect.String, reflect.Array:
			showdef = option.value.Len() > 0
		case reflect.Map:
			showdef = !option.value.IsNil() && option.value.Len() > 0
		default:
			zeroval := reflect.Zero(option.field.Type)
			showdef = !reflect.DeepEqual(zeroval.Interface(), option.value.Interface())
		}

		if showdef {
			def, _ = convertToString(option.value, option.tag)
		}
	} else if len(defs) != 0 {
		l := len(defs) - 1

		var defSb532 strings.Builder
		for i := range l {
			defSb532.WriteString(quoteIfNeeded(defs[i]) + ", ")
		}
		def += defSb532.String()

		def += quoteIfNeeded(defs[l])
	}

	option.defaultLiteral = def
}

func (option *Option) shortAndLongName() string {
	ret := &bytes.Buffer{}

	if option.ShortName != 0 {
		ret.WriteRune(defaultShortOptDelimiter)
		ret.WriteRune(option.ShortName)
	}

	if len(option.LongName) != 0 {
		if option.ShortName != 0 {
			ret.WriteRune('/')
		}

		ret.WriteString(option.LongName)
	}

	return ret.String()
}

func (option *Option) isValidValue(arg string) error {
	if validator := option.isValueValidator(); validator != nil {
		return validator.IsValidValue(arg)
	}
	if argumentIsOption(arg) && (!option.isSignedNumber() || len(arg) <= 1 || arg[0] != '-' || arg[1] < '0' || arg[1] > '9') {
		if p := option.parser(); p != nil {
			return fmt.Errorf(
				"%s",
				p.i18nTextf(
					"err.invalid_argument.option",
					"expected argument for flag `{flag}`, but got option `{option}`",
					map[string]string{
						"flag":   option.String(),
						"option": arg,
					},
				),
			)
		}

		return fmt.Errorf("expected argument for flag `%s`, but got option `%s`", option, arg)
	}
	return nil
}

func (option *Option) validateChoice(value string) error {
	if len(option.Choices) == 0 || slices.Contains(option.Choices, value) {
		return nil
	}

	allowed := option.Choices[0]
	p := option.parser()

	if len(option.Choices) > 1 {
		items := strings.Join(option.Choices[0:len(option.Choices)-1], ", ")
		last := option.Choices[len(option.Choices)-1]
		if p != nil {
			allowed = p.i18nTextf(
				"err.list.disjunction",
				"{items} or {last}",
				map[string]string{
					"items": items,
					"last":  last,
				},
			)
		} else {
			allowed = items + " or " + last
		}
	}

	if p != nil {
		return newError(
			ErrInvalidChoice,
			p.i18nTextf(
				"err.invalid_choice",
				"Invalid value `{value}` for option `{option}`. Allowed values are: {allowed}",
				map[string]string{
					"value":   value,
					"option":  option.String(),
					"allowed": allowed,
				},
			),
		)
	}

	return newErrorf(
		ErrInvalidChoice,
		"Invalid value `%s` for option `%s`. Allowed values are: %s",
		value, option, allowed,
	)
}
