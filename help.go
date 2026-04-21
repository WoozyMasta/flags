// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"
	"unicode/utf8"
)

type alignmentInfo struct {
	maxLongLen      int
	terminalColumns int
	hasShort        bool
	hasValueName    bool
	indent          bool
}

const (
	paddingBeforeOption                 = 2
	distanceBetweenOptionAndDescription = 2
)

func (a *alignmentInfo) descriptionStart() int {
	ret := a.maxLongLen + distanceBetweenOptionAndDescription

	if a.hasShort {
		ret += 2
	}

	if a.maxLongLen > 0 {
		ret += 4
	}

	if a.hasValueName {
		ret += 3
	}

	return ret
}

func (a *alignmentInfo) updateLen(name string, indent bool) {
	l := utf8.RuneCountInString(name)

	if indent {
		l += 4
	}

	if l > a.maxLongLen {
		a.maxLongLen = l
	}
}

func (p *Parser) getAlignmentInfo() alignmentInfo {
	ret := alignmentInfo{
		maxLongLen:      0,
		hasShort:        false,
		hasValueName:    false,
		terminalColumns: getTerminalColumns(),
	}

	if ret.terminalColumns <= 0 {
		ret.terminalColumns = 80
	}

	var prevcmd *Command

	p.eachActiveGroup(func(c *Command, grp *Group) {
		if c != prevcmd {
			for _, arg := range c.args {
				ret.updateLen(arg.Name, c != p.Command)
			}
			prevcmd = c
		}
		if !grp.showInHelp() {
			return
		}
		for _, info := range grp.options {
			if !info.showInHelp() {
				continue
			}

			if info.ShortName != 0 {
				ret.hasShort = true
			}

			if len(info.ValueName) > 0 {
				ret.hasValueName = true
			}

			l := info.LongNameWithNamespace() + info.ValueName

			if len(info.Choices) != 0 {
				l += "[" + strings.Join(info.Choices, "|") + "]"
			}

			ret.updateLen(l, c != p.Command)
		}
	})

	return ret
}

func wrapText(s string, l int, prefix string, trimWhitespace bool) string {
	var ret string

	if l < 10 {
		l = 10
	}

	// Basic text wrapping of s at spaces to fit in l
	lines := strings.SplitSeq(s, "\n")

	for line := range lines {
		var retline string

		if trimWhitespace {
			line = strings.TrimSpace(line)
		}

		for len(line) > l {
			// Try to split on space
			suffix := ""

			pos := strings.LastIndex(line[:l], " ")

			splitOnSpace := pos >= 0

			if !splitOnSpace {
				pos = l - 1
				suffix = "-\n"
			}

			if len(retline) != 0 {
				retline += "\n" + prefix
			}

			part := line[:pos]
			switch {
			case trimWhitespace:
				part = strings.TrimSpace(part)
				line = strings.TrimSpace(line[pos:])
			case splitOnSpace:
				// Consume the separator so it is not processed repeatedly.
				line = line[pos+1:]
			default:
				line = line[pos:]
			}

			retline += part + suffix
		}

		if len(line) > 0 {
			if len(retline) != 0 {
				retline += "\n" + prefix
			}

			retline += line
		}

		if len(ret) > 0 {
			ret += "\n"

			if len(retline) > 0 {
				ret += prefix
			}
		}

		ret += retline
	}

	return ret
}

func (p *Parser) writeHelpOption(writer *bufio.Writer, option *Option, info alignmentInfo, trimDescriptions bool) {
	line := &bytes.Buffer{}

	prefix := paddingBeforeOption

	if info.indent {
		prefix += 4
	}

	if option.Hidden {
		return
	}

	line.WriteString(strings.Repeat(" ", prefix))

	if option.ShortName != 0 {
		line.WriteRune(defaultShortOptDelimiter)
		line.WriteRune(option.ShortName)
	} else if info.hasShort {
		line.WriteString("  ")
	}

	descstart := info.descriptionStart() + paddingBeforeOption

	if len(option.LongName) > 0 {
		if option.ShortName != 0 {
			line.WriteString(", ")
		} else if info.hasShort {
			line.WriteString("  ")
		}

		line.WriteString(defaultLongOptDelimiter)
		line.WriteString(option.LongNameWithNamespace())
	}

	if option.canArgument() {
		line.WriteRune(defaultNameArgDelimiter)

		if len(option.ValueName) > 0 {
			line.WriteString(option.ValueName)
		}

		if len(option.Choices) > 0 {
			line.WriteString("[" + strings.Join(option.Choices, "|") + "]")
		}
	}

	written := line.Len()
	_, _ = line.WriteTo(writer)

	if option.Description != "" {
		dw := descstart - written
		_, _ = writer.WriteString(strings.Repeat(" ", dw))

		var def string

		if len(option.DefaultMask) != 0 {
			if option.DefaultMask != "-" {
				def = option.DefaultMask
			}
		} else {
			if !option.defaultLiteralInitialized {
				option.updateDefaultLiteral()
				option.defaultLiteralInitialized = true
			}
			def = option.defaultLiteral
		}

		var envDef string
		if option.EnvKeyWithNamespace() != "" {
			var envPrintable string
			if runtime.GOOS == "windows" {
				envPrintable = "%" + option.EnvKeyWithNamespace() + "%"
			} else {
				envPrintable = "$" + option.EnvKeyWithNamespace()
			}
			envDef = fmt.Sprintf(" [%s]", envPrintable)
		}

		var desc string

		if def != "" {
			desc = fmt.Sprintf("%s (default: %v)%s", option.Description, def, envDef)
		} else {
			desc = option.Description + envDef
		}

		_, _ = writer.WriteString(wrapText(desc,
			info.terminalColumns-descstart,
			strings.Repeat(" ", descstart),
			trimDescriptions))
	}

	_, _ = writer.WriteString("\n")
}

func maxCommandLength(s []*Command) int {
	if len(s) == 0 {
		return 0
	}

	ret := len(s[0].Name)

	for _, v := range s[1:] {
		l := len(v.Name)

		if l > ret {
			ret = l
		}
	}

	return ret
}

// WriteHelp writes a help message containing all the possible options and
// their descriptions to the provided writer. Note that the HelpFlag parser
// option provides a convenient way to add a -h/--help option group to the
// command line parser which will automatically show the help messages using
// this method.
func (p *Parser) WriteHelp(writer io.Writer) {
	if writer == nil {
		return
	}

	wr := bufio.NewWriter(writer)
	aligninfo := p.getAlignmentInfo()
	trimDescriptions := (p.Options & KeepDescriptionWhitespace) == None

	cmd := p.Command

	for cmd.Active != nil {
		cmd = cmd.Active
	}

	if p.Name != "" {
		_, _ = wr.WriteString("Usage:\n")
		_, _ = wr.WriteString(" ")

		allcmd := p.Command

		for allcmd != nil {
			var usage string

			if allcmd == p.Command {
				if len(p.Usage) != 0 {
					usage = p.Usage
				} else if p.Options&HelpFlag != 0 {
					usage = "[OPTIONS]"
				}
			} else if us, ok := allcmd.data.(Usage); ok {
				usage = us.Usage()
			} else if allcmd.hasHelpOptions() {
				usage = fmt.Sprintf("[%s-OPTIONS]", allcmd.Name)
			}

			if len(usage) != 0 {
				_, _ = fmt.Fprintf(wr, " %s %s", allcmd.Name, usage)
			} else {
				_, _ = fmt.Fprintf(wr, " %s", allcmd.Name)
			}

			if len(allcmd.args) > 0 {
				_, _ = fmt.Fprintf(wr, " ")
			}

			for i, arg := range allcmd.args {
				if i != 0 {
					_, _ = fmt.Fprintf(wr, " ")
				}

				name := arg.Name

				if arg.isRemaining() {
					name += "..."
				}

				if !allcmd.ArgsRequired {
					if arg.Required > 0 {
						_, _ = fmt.Fprintf(wr, "%s", name)
					} else {
						_, _ = fmt.Fprintf(wr, "[%s]", name)
					}
				} else {
					_, _ = fmt.Fprintf(wr, "%s", name)
				}
			}

			if allcmd.Active == nil && len(allcmd.commands) > 0 {
				var co, cc string

				if allcmd.SubcommandsOptional {
					co, cc = "[", "]"
				} else {
					co, cc = "<", ">"
				}

				visibleCommands := allcmd.visibleCommands()

				if len(visibleCommands) > 3 {
					_, _ = fmt.Fprintf(wr, " %scommand%s", co, cc)
				} else {
					subcommands := allcmd.sortedVisibleCommands()
					names := make([]string, len(subcommands))

					for i, subc := range subcommands {
						names[i] = subc.Name
					}

					_, _ = fmt.Fprintf(wr, " %s%s%s", co, strings.Join(names, " | "), cc)
				}
			}

			allcmd = allcmd.Active
		}

		_, _ = fmt.Fprintln(wr)

		if len(cmd.LongDescription) != 0 {
			_, _ = fmt.Fprintln(wr)

			t := wrapText(cmd.LongDescription,
				aligninfo.terminalColumns,
				"",
				trimDescriptions)

			_, _ = fmt.Fprintln(wr, t)
		}
	}

	c := p.Command

	for c != nil {
		printcmd := c != p.Command

		c.eachGroup(func(grp *Group) {
			first := true

			// Skip built-in help group for all commands except the top-level
			// parser
			if grp.Hidden || (grp.isBuiltinHelp && c != p.Command) {
				return
			}

			for _, info := range grp.Options() {
				if !info.showInHelp() {
					continue
				}

				if printcmd {
					_, _ = fmt.Fprintf(wr, "\n[%s command options]\n", c.Name)
					aligninfo.indent = true
					printcmd = false
				}

				if first && cmd.Group != grp {
					_, _ = fmt.Fprintln(wr)

					if aligninfo.indent {
						_, _ = wr.WriteString("    ")
					}

					_, _ = fmt.Fprintf(wr, "%s:\n", grp.ShortDescription)
					first = false
				}

				p.writeHelpOption(wr, info, aligninfo, trimDescriptions)
			}
		})

		var args []*Arg
		for _, arg := range c.args {
			if arg.Description != "" {
				args = append(args, arg)
			}
		}

		if len(args) > 0 {
			if c == p.Command {
				_, _ = fmt.Fprintf(wr, "\nArguments:\n")
			} else {
				_, _ = fmt.Fprintf(wr, "\n[%s command arguments]\n", c.Name)
			}

			descStart := aligninfo.descriptionStart() + paddingBeforeOption

			for _, arg := range args {
				argPrefix := strings.Repeat(" ", paddingBeforeOption)
				argPrefix += arg.Name

				if len(arg.Description) > 0 {
					argPrefix += ":"
					_, _ = wr.WriteString(argPrefix)

					// Space between "arg:" and the description start
					descPadding := strings.Repeat(" ", descStart-len(argPrefix))
					// How much space the description gets before wrapping
					descWidth := aligninfo.terminalColumns - 1 - descStart
					// Whitespace to which we can indent new description lines
					descPrefix := strings.Repeat(" ", descStart)

					_, _ = wr.WriteString(descPadding)
					_, _ = wr.WriteString(wrapText(arg.Description, descWidth, descPrefix, trimDescriptions))
				} else {
					_, _ = wr.WriteString(argPrefix)
				}

				_, _ = fmt.Fprintln(wr)
			}
		}

		c = c.Active
	}

	scommands := cmd.sortedVisibleCommands()

	if len(scommands) > 0 {
		maxnamelen := maxCommandLength(scommands)

		_, _ = fmt.Fprintln(wr)
		_, _ = fmt.Fprintln(wr, "Available commands:")

		for _, c := range scommands {
			_, _ = fmt.Fprintf(wr, "  %s", c.Name)

			if len(c.ShortDescription) > 0 {
				pad := strings.Repeat(" ", maxnamelen-len(c.Name))
				_, _ = fmt.Fprintf(wr, "%s  %s", pad, c.ShortDescription)

				if len(c.Aliases) > 0 {
					_, _ = fmt.Fprintf(wr, " (aliases: %s)", strings.Join(c.Aliases, ", "))
				}
			}

			_, _ = fmt.Fprintln(wr)
		}
	}

	_ = wr.Flush()
}

// WroteHelp is a helper to test the error from ParseArgs() to
// determine if the help message was written. It is safe to
// call without first checking that error is nil.
func WroteHelp(err error) bool {
	if err == nil { // No error
		return false
	}

	flagError, ok := err.(*Error)
	if !ok { // Not a go-flag error
		return false
	}

	if flagError.Type != ErrHelp { // Did not print the help message
		return false
	}

	return true
}
