// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

// Package main demonstrates custom flag tag mapping.
package main

import (
	"errors"
	"os"

	"github.com/woozymasta/flags"
)

type PrefixTagsOptions struct {
	// SetTagPrefix("flag-") maps these tags back to the standard short/long
	// and description tag meanings.
	Path string `flag-short:"p" flag-long:"path" flag-description:"Path to input file"`
}

type CustomTagsOptions struct {
	// SetFlagTags can replace individual tag names while leaving all other
	// standard tags unchanged.
	Path  string `x-short:"p" long:"path" description:"Path to input file"`
	Level int    `long:"level" default:"1" description:"Verbosity level"`
}

type PrefixCommand struct{}

func (c *PrefixCommand) Execute(args []string) error {
	return runPrefixTagsDemo(args)
}

type CustomCommand struct{}

func (c *CustomCommand) Execute(args []string) error {
	return runCustomTagsDemo(args)
}

type RootOptions struct {
	Prefix PrefixCommand `command:"prefix" description:"Run parser with SetTagPrefix(\"flag-\") mapping"`
	Custom CustomCommand `command:"custom" description:"Run parser with SetFlagTags(...) custom mapping"`
}

func runPrefixTagsDemo(args []string) error {
	opts := PrefixTagsOptions{}
	p := flags.NewNamedParser("prefix-tags-demo", flags.Default)
	if _, err := p.AddGroup("Application Options", "Tag prefix demo", &opts); err != nil {
		return err
	}
	// SetTagPrefix rescans already registered groups, so it can be called after
	// AddGroup when a parser is assembled by reusable setup code.
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
	// Only the short tag is customized here. NewFlagTags supplies defaults for
	// every tag that is not explicitly changed.
	if err := p.SetFlagTags(tags); err != nil {
		return err
	}

	_, err := p.ParseArgs(args)
	return err
}

func main() {
	root := RootOptions{}
	p := flags.NewNamedParser("custom-flag-tags", flags.Default)
	// The root parser only selects which tag-mapping demo to run. Each command
	// creates a separate child parser so the two mappings do not interfere.
	p.Usage = "[OPTIONS] <prefix|custom> [ARGS]"
	p.LongDescription = "Subcommands:\n  prefix  Demonstrates SetTagPrefix(\"flag-\")\n  custom  Demonstrates SetFlagTags(...)"

	if _, err := p.AddGroup("Application Options", "Mode selector", &root); err != nil {
		os.Exit(1)
	}

	_, err := p.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		}
		os.Exit(1)
	}
}
