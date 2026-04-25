// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package main

import (
	"fmt"
)

type RmCommand struct {
	// Command-local flags are only active after the command name is selected.
	Force bool `short:"f" long:"force" description:"Force removal of files"`
}

var rmCommand RmCommand

//nolint:unparam // required by flags.Commander interface
func (x *RmCommand) Execute(args []string) error {
	// args are the non-option values that belong to rm, for example filenames.
	fmt.Printf("Removing (force=%v): %#v\n", x.Force, args)
	return nil
}

func init() {
	// AddCommand wires a command name, help text, and a command data struct.
	// The parser scans the struct immediately and later calls Execute.
	_, _ = parser.AddCommand("rm",
		"Remove a file",
		"The rm command removes a file to the repository. Use -f to force removal of files.",
		&rmCommand)
}
