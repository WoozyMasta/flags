// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"errors"
	"os"

	"github.com/woozymasta/flags"
)

type PrefixTagsOptions struct {
	Path string `flag-short:"p" flag-long:"path" flag-description:"Path to input file"`
}

type CustomTagsOptions struct {
	Path  string `x-short:"p" long:"path" description:"Path to input file"`
	Level int    `long:"level" default:"1" description:"Verbosity level"`
}

func runPrefixTagsDemo(args []string) error {
	opts := PrefixTagsOptions{}
	p := flags.NewNamedParser("prefix-tags-demo", flags.Default)
	if _, err := p.AddGroup("Application Options", "Tag prefix demo", &opts); err != nil {
		return err
	}
	if err := p.SetTagPrefix("flag-"); err != nil {
		return err
	}

	_, err := p.ParseArgs(args)
	return err
}

func runCustomTagsDemo(args []string) error {
	opts := CustomTagsOptions{}
	p := flags.NewNamedParser("custom-tags-demo", flags.Default)
	if _, err := p.AddGroup("Application Options", "Custom tags demo", &opts); err != nil {
		return err
	}

	tags := flags.NewFlagTags()
	tags.Short = "x-short"
	if err := p.SetFlagTags(tags); err != nil {
		return err
	}

	_, err := p.ParseArgs(args)
	return err
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	mode := os.Args[1]
	args := os.Args[2:]

	var err error
	switch mode {
	case "prefix":
		err = runPrefixTagsDemo(args)
	case "custom":
		err = runCustomTagsDemo(args)
	default:
		os.Exit(1)
	}

	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
