// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

// Package main demonstrates flags usage with subcommands.
package main

import (
	"fmt"
)

type AddCommand struct {
	// Command structs use the same option tags as root option structs.
	All bool `short:"a" long:"all" description:"Add all files"`
}

var addCommand AddCommand

//nolint:unparam // required by flags.Commander interface
func (x *AddCommand) Execute(args []string) error {
	// Execute receives the remaining positional arguments after this command's
	// options have been parsed.
	fmt.Printf("Adding (all=%v): %#v\n", x.All, args)
	return nil
}

func init() {
	// This file shares the package-level parser from examples/main.go. Splitting
	// commands across files mirrors how larger CLIs can register subcommands
	// from separate packages or files.
	_, _ = parser.AddCommand("add",
		"Add a file",
		"The add command adds a file to the repository. Use -a to add all files.",
		&addCommand)
}
