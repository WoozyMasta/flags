// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"reflect"
	"strings"
)

// Arg represents a positional argument on the command line.
type Arg struct {
	tag   multiTag
	value reflect.Value

	// The name of the positional argument (used in the help)
	Name string

	// A description of the positional argument (used in the help)
	Description string

	// The default value(s) of the positional argument.
	Default []string

	// The minimal number of required positional arguments
	Required int

	// The maximum number of required positional arguments
	RequiredMaximum int
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
