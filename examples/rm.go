// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
