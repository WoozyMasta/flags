// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/woozymasta/flags"
)

type EditorOptions struct {
	Input  flags.Filename `short:"i" long:"input" description:"Input file" default:"-"`
	Output flags.Filename `short:"o" long:"output" description:"Output file" default:"-"`
}

type Point struct {
	X, Y int
}

func (p *Point) UnmarshalFlag(value string) error {
	// Implementing flags.Unmarshaler lets a domain type own its CLI syntax.
	// The parser calls this method for --point=1,2 instead of requiring a
	// separate string option and manual conversion after parsing.
	parts := strings.Split(value, ",")

	if len(parts) != 2 {
		return errors.New("expected two numbers separated by a ,")
	}

	x, err := strconv.ParseInt(parts[0], 10, 32)

	if err != nil {
		return err
	}

	y, err := strconv.ParseInt(parts[1], 10, 32)

	if err != nil {
		return err
	}

	p.X = int(x)
	p.Y = int(y)

	return nil
}

//nolint:unparam // required by flags.Marshaler interface
func (p Point) MarshalFlag() (string, error) {
	// MarshalFlag controls how defaults and current values for custom types
	// are rendered back into help, docs, and other generated output.
	return fmt.Sprintf("%d,%d", p.X, p.Y), nil
}

type Options struct {
	// Map values can be pre-populated in Go code. default-mask keeps help text
	// stable without duplicating all map defaults in struct tags.
	Users map[string]string `long:"users" description:"User e-mail map" default-mask:"system:system@example.org, admin:admin@example.org"`

	// Nested structs become grouped options in help output.
	Editor EditorOptions `group:"Editor Options"`

	// optional-value allows both --user and --user=name forms.
	User string `short:"u" long:"user" description:"User name" optional:"yes" optional-value:"pancake"`

	// Repeating -v appends to the slice, so len(Verbose) is the level.
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`

	// Point uses the MarshalFlag/UnmarshalFlag methods above.
	Point Point `long:"point" description:"A x,y point" default:"1,2"`
}

var options = Options{
	Users: map[string]string{
		"system": "system@example.org",
		"admin":  "admin@example.org",
	},
}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		// Help is delivered as a typed parser error so applications can
		// distinguish a successful help/version flow from real failures.
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
