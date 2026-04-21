// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
)

type RmCommand struct {
	Force bool `short:"f" long:"force" description:"Force removal of files"`
}

var rmCommand RmCommand

//nolint:unparam // required by flags.Commander interface
func (x *RmCommand) Execute(args []string) error {
	fmt.Printf("Removing (force=%v): %#v\n", x.Force, args)
	return nil
}

func init() {
	_, _ = parser.AddCommand("rm",
		"Remove a file",
		"The rm command removes a file to the repository. Use -f to force removal of files.",
		&rmCommand)
}
