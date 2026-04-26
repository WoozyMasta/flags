// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"reflect"
	"strings"
)

// Arg represents a positional argument on the command line.
type Arg struct {
	tag   multiTag
	value reflect.Value
	cmd   *Command

	// The name of the positional argument (used in the help)
	Name string

	// Optional i18n key for Name.
	NameI18nKey string

	// A description of the positional argument (used in the help)
	Description string

	// Optional i18n key for Description.
	DescriptionI18nKey string

	// Group name used to organize positional arguments in help/docs.
	Group string

	// The default value(s) of the positional argument.
	Default []string

	// completionHint controls fallback completion mode (file, dir, none).
	completionHint completionHint

	// The minimal number of required positional arguments
	Required int

	// The maximum number of required positional arguments
	RequiredMaximum int
}

// SetName updates positional argument name used in usage/help placeholders.
func (a *Arg) SetName(name string) {
	a.Name = name
}

// SetNameI18nKey sets i18n key used to localize positional argument name.
func (a *Arg) SetNameI18nKey(key string) {
	a.NameI18nKey = key
}

// SetDescription updates positional argument description used in help/docs.
func (a *Arg) SetDescription(description string) {
	a.Description = description
}

// SetDescriptionI18nKey sets i18n key used to localize positional argument description.
func (a *Arg) SetDescriptionI18nKey(key string) {
	a.DescriptionI18nKey = key
}

// SetGroup updates help/docs group used for this positional argument.
func (a *Arg) SetGroup(group string) {
	a.Group = group
}

// SetDefault replaces positional argument default values.
func (a *Arg) SetDefault(values ...string) {
	a.Default = append(a.Default[:0], values...)
}

// SetRequired toggles required state for positional argument.
func (a *Arg) SetRequired(required bool) {
	if required {
		a.Required = 1
		a.RequiredMaximum = -1
		return
	}

	a.Required = -1
	a.RequiredMaximum = -1
}

// SetRequiredRange sets positional required bounds.
// Use requiredMax = -1 for "at least requiredMin".
func (a *Arg) SetRequiredRange(requiredMin int, requiredMax int) error {
	if requiredMin < 0 {
		return newErrorf(ErrInvalidTag, "required min must be >= 0, got %d", requiredMin)
	}
	if requiredMax < -1 {
		return newErrorf(ErrInvalidTag, "required max must be >= -1, got %d", requiredMax)
	}
	if requiredMax != -1 && requiredMax < requiredMin {
		return newErrorf(ErrInvalidTag, "required max %d must be >= min %d", requiredMax, requiredMin)
	}

	a.Required = requiredMin
	a.RequiredMaximum = requiredMax
	return nil
}

func (a *Arg) isRemaining() bool {
	return a.value.Type().Kind() == reflect.Slice
}

func (a *Arg) emptyValue() reflect.Value {
	tp := a.value.Type()

	if tp.Kind() == reflect.Map {
		return reflect.MakeMap(tp)
	}

	return reflect.Zero(tp)
}

func (a *Arg) empty() {
	a.value.Set(a.emptyValue())
}

func (a *Arg) isEmpty() bool {
	switch a.value.Kind() {
	case reflect.String, reflect.Slice, reflect.Map:
		return a.value.Len() == 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return a.value.IsNil()
	default:
		return a.value.IsZero()
	}
}

func (a *Arg) applyDefault(defaultsIfEmpty bool) error {
	if len(a.Default) == 0 {
		return nil
	}

	if defaultsIfEmpty && !a.isEmpty() {
		return nil
	}

	a.empty()

	for _, d := range a.Default {
		if err := convert(d, a.value, a.tag); err != nil {
			return err
		}
	}

	return nil
}

func (a *Arg) defaultLiteral() string {
	if len(a.Default) == 0 {
		return ""
	}

	parts := make([]string, 0, len(a.Default))
	for _, d := range a.Default {
		parts = append(parts, quoteIfNeeded(d))
	}

	return strings.Join(parts, ", ")
}
