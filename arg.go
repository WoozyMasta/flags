// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"reflect"
)

// Arg represents a positional argument on the command line.
type Arg struct {
	tag   multiTag
	value reflect.Value

	// The name of the positional argument (used in the help)
	Name string

	// A description of the positional argument (used in the help)
	Description string

	// The minimal number of required positional arguments
	Required int

	// The maximum number of required positional arguments
	RequiredMaximum int
}

func (a *Arg) isRemaining() bool {
	return a.value.Type().Kind() == reflect.Slice
}
