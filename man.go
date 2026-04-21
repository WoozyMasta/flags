// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func manQuoteLines(s string) string {
	lines := strings.Split(s, "\n")
	parts := make([]string, 0, len(lines))

	for _, line := range lines {
		parts = append(parts, manQuote(line))
	}

	return strings.Join(parts, "\n")
}

func manQuote(s string) string {
	return strings.ReplaceAll(s, "\\", "\\\\")
}

func formatForMan(wr io.Writer, s string, quoter func(s string) string) {
	for {
		idx := strings.IndexRune(s, '`')

		if idx < 0 {
			_, _ = fmt.Fprintf(wr, "%s", quoter(s))
			break
		}

		_, _ = fmt.Fprintf(wr, "%s", quoter(s[:idx]))

		s = s[idx+1:]
		idx = strings.IndexRune(s, '\'')

		if idx < 0 {
			_, _ = fmt.Fprintf(wr, "%s", quoter(s))
			break
		}

		_, _ = fmt.Fprintf(wr, "\\fB%s\\fP", quoter(s[:idx]))
		s = s[idx+1:]
	}
}

func writeManPageOptions(wr io.Writer, grp *Group) {
	grp.eachGroup(func(group *Group) {
		if !group.showInHelp() {
			return
		}

		// If the parent (grp) has any subgroups, display their descriptions as
		// subsection headers similar to the output of --help.
		if group.ShortDescription != "" && len(grp.groups) > 0 {
			_, _ = fmt.Fprintf(wr, ".SS %s\n", group.ShortDescription)

			if group.LongDescription != "" {
				formatForMan(wr, group.LongDescription, manQuoteLines)
				_, _ = fmt.Fprintln(wr, "")
			}
		}

		for _, opt := range group.options {
			if !opt.showInHelp() {
				continue
			}

			_, _ = fmt.Fprintln(wr, ".TP")
			_, _ = fmt.Fprintf(wr, "\\fB")

			if opt.ShortName != 0 {
				_, _ = fmt.Fprintf(wr, "\\fB\\-%c\\fR", opt.ShortName)
			}

			if len(opt.LongName) != 0 {
				if opt.ShortName != 0 {
					_, _ = fmt.Fprintf(wr, ", ")
				}

				_, _ = fmt.Fprintf(wr, "\\fB\\-\\-%s\\fR", manQuote(opt.LongNameWithNamespace()))
			}

			if len(opt.ValueName) != 0 || opt.OptionalArgument {
				if opt.OptionalArgument {
					_, _ = fmt.Fprintf(wr, " [\\fI%s=%s\\fR]", manQuote(opt.ValueName), manQuote(strings.Join(quoteV(opt.OptionalValue), ", ")))
				} else {
					_, _ = fmt.Fprintf(wr, " \\fI%s\\fR", manQuote(opt.ValueName))
				}
			}

			if len(opt.Default) != 0 {
				_, _ = fmt.Fprintf(wr, " <default: \\fI%s\\fR>", manQuote(strings.Join(quoteV(opt.Default), ", ")))
			} else if len(opt.EnvKeyWithNamespace()) != 0 {
				if runtime.GOOS == "windows" {
					_, _ = fmt.Fprintf(wr, " <default: \\fI%%%s%%\\fR>", manQuote(opt.EnvKeyWithNamespace()))
				} else {
					_, _ = fmt.Fprintf(wr, " <default: \\fI$%s\\fR>", manQuote(opt.EnvKeyWithNamespace()))
				}
			}

			if opt.Required {
				_, _ = fmt.Fprintf(wr, " (\\fIrequired\\fR)")
			}

			_, _ = fmt.Fprintln(wr, "\\fP")

			if len(opt.Description) != 0 {
				formatForMan(wr, opt.Description, manQuoteLines)
				_, _ = fmt.Fprintln(wr, "")
			}
		}
	})
}

func writeManPageSubcommands(wr io.Writer, name string, usagePrefix string, root *Command) {
	commands := root.sortedVisibleCommands()

	for _, c := range commands {
		var nn string

		if c.Hidden {
			continue
		}

		if len(name) != 0 {
			nn = name + " " + c.Name
		} else {
			nn = c.Name
		}

		writeManPageCommand(wr, nn, usagePrefix, c)
	}
}

func writeManPageCommand(wr io.Writer, name string, usagePrefix string, command *Command) {
	_, _ = fmt.Fprintf(wr, ".SS %s\n", name)
	_, _ = fmt.Fprintln(wr, command.ShortDescription)

	if len(command.LongDescription) > 0 {
		_, _ = fmt.Fprintln(wr, "")

		cmdstart := fmt.Sprintf("The %s command", manQuote(command.Name))

		if strings.HasPrefix(command.LongDescription, cmdstart) {
			_, _ = fmt.Fprintf(wr, "The \\fI%s\\fP command", manQuote(command.Name))

			formatForMan(wr, command.LongDescription[len(cmdstart):], manQuoteLines)
			_, _ = fmt.Fprintln(wr, "")
		} else {
			formatForMan(wr, command.LongDescription, manQuoteLines)
			_, _ = fmt.Fprintln(wr, "")
		}
	}

	var pre = usagePrefix + " " + command.Name

	var usage string
	if us, ok := command.data.(Usage); ok {
		usage = us.Usage()
	} else if command.hasHelpOptions() {
		usage = fmt.Sprintf("[%s-OPTIONS]", command.Name)
	}

	var nextPrefix = pre
	if len(usage) > 0 {
		_, _ = fmt.Fprintf(wr, "\n\\fBUsage\\fP: %s %s\n.TP\n", manQuote(pre), manQuote(usage))
		nextPrefix = pre + " " + usage
	}

	if len(command.Aliases) > 0 {
		_, _ = fmt.Fprintf(wr, "\n\\fBAliases\\fP: %s\n\n", manQuote(strings.Join(command.Aliases, ", ")))
	}

	writeManPageOptions(wr, command.Group)
	writeManPageSubcommands(wr, name, nextPrefix, command)
}

// WriteManPage writes a basic man page in groff format to the specified
// writer.
func (p *Parser) WriteManPage(wr io.Writer) {
	t := time.Now()
	sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH")
	if sourceDateEpoch != "" {
		sde, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid SOURCE_DATE_EPOCH: %s", err))
		}
		t = time.Unix(sde, 0)
	}

	_, _ = fmt.Fprintf(wr, ".TH %s 1 \"%s\"\n", manQuote(p.Name), t.Format("2 January 2006"))
	_, _ = fmt.Fprintln(wr, ".SH NAME")
	_, _ = fmt.Fprintf(wr, "%s \\- %s\n", manQuote(p.Name), manQuoteLines(p.ShortDescription))
	_, _ = fmt.Fprintln(wr, ".SH SYNOPSIS")

	usage := p.Usage

	if len(usage) == 0 {
		usage = "[OPTIONS]"
	}

	_, _ = fmt.Fprintf(wr, "\\fB%s\\fP %s\n", manQuote(p.Name), manQuote(usage))
	_, _ = fmt.Fprintln(wr, ".SH DESCRIPTION")

	formatForMan(wr, p.LongDescription, manQuoteLines)
	_, _ = fmt.Fprintln(wr, "")

	_, _ = fmt.Fprintln(wr, ".SH OPTIONS")

	writeManPageOptions(wr, p.Group)

	if len(p.visibleCommands()) > 0 {
		_, _ = fmt.Fprintln(wr, ".SH COMMANDS")

		writeManPageSubcommands(wr, "", p.Name+" "+usage, p.Command)
	}
}
