// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/woozymasta/flags"
)

type Options struct {
	Name string `short:"n" long:"name" description:"Name to greet" default:"world"`
}

func main() {
	var options Options
	parser := flags.NewParser(&options, flags.Default)

	_, err := parser.Parse()
	if err == nil {
		fmt.Printf("Hello, %s!\n", options.Name)
	}

	// A double-clicked .exe gets its own console window that closes the instant the process exits,
	// so any output above (help, greeting, error) would flash and disappear unread.
	// Pause once, regardless of the outcome, before the process would otherwise exit.
	if flags.LaunchedFromExplorer() {
		_ = flags.WaitForEnter(os.Stdout, os.Stdin, "\nPress Enter to exit...")
	}

	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
