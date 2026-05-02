// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

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

// WriteHelp writes a help message containing all the possible options and
// their descriptions to the provided writer. Note that the HelpFlag parser
// option provides a convenient way to add a -h/--help option group to the
// command line parser which will automatically show the help messages using
// this method.
func (p *Parser) WriteHelp(writer io.Writer) {
	if writer == nil {
		return
	}

	prevHelpColorEnabled := p.helpColorEnabled
	p.helpColorEnabled = (p.Options&ColorHelp) != None && DetectColorSupport(writer)
	defer func() {
		p.helpColorEnabled = prevHelpColorEnabled
	}()

	// Keep WriteHelp behavior consistent with ParseArgs:
	// when builtin help/version flags are enabled, ensure
	// corresponding options are present in help output.
	if (p.Options & (HelpFlag | VersionFlag)) != None {
		p.addHelpGroups(p.showBuiltinHelp, p.markVersionRequested)
	}
	_ = p.EnsureBuiltinCommands()

	wr := bufio.NewWriter(writer)
	basePrefix := ""
	if (p.Options&ColorHelp) != None && p.helpColorEnabled {
		basePrefix = helpStylePrefix(p.helpColorScheme.BaseText)
		if basePrefix != "" {
			_, _ = wr.WriteString(basePrefix)
		}
	}
	aligninfo := p.getAlignmentInfo()
	format := p.optionRenderFormat()
	trimDescriptions := (p.Options & KeepDescriptionWhitespace) == None

	cmd := p.Command

	for cmd.Active != nil {
		cmd = cmd.Active
	}

	if p.Name != "" {
		usageLabel := p.i18nText("help.usage", "Usage") + ":"
		_, _ = wr.WriteString(p.colorizeHelp(usageLabel, p.helpColorScheme.UsageHeader))
		if basePrefix != "" && p.helpColorScheme.BaseText.UseBG && !aligninfo.unlimitedWidth {
			pad := aligninfo.terminalColumns - textWidth(usageLabel)
			if pad > 0 {
				_, _ = wr.WriteString(strings.Repeat(" ", pad))
			}
		}
		_, _ = wr.WriteString("\n")
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
				_, _ = fmt.Fprintf(
					wr,
					" %s %s",
					p.colorizeHelp(allcmd.Name, p.helpColorScheme.UsageText),
					p.colorizeHelp(usage, p.helpColorScheme.UsageText),
				)
			} else {
				_, _ = fmt.Fprintf(wr, " %s", p.colorizeHelp(allcmd.Name, p.helpColorScheme.UsageText))
			}

			if len(allcmd.args) > 0 {
				_, _ = fmt.Fprintf(wr, " ")
			}

			for i, arg := range allcmd.args {
				if i != 0 {
					_, _ = fmt.Fprintf(wr, " ")
				}

				name := arg.localizedName()

				if arg.isRemaining() {
					name += "..."
				}

				if !allcmd.ArgsRequired {
					if arg.Required > 0 {
						_, _ = fmt.Fprintf(wr, "%s", p.colorizeHelp(name, p.helpColorScheme.UsageText))
					} else {
						_, _ = fmt.Fprintf(wr, "[%s]", p.colorizeHelp(name, p.helpColorScheme.UsageText))
					}
				} else {
					_, _ = fmt.Fprintf(wr, "%s", p.colorizeHelp(name, p.helpColorScheme.UsageText))
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
					_, _ = fmt.Fprintf(
						wr,
						" %s%s%s",
						co,
						p.colorizeHelp(p.i18nText("help.command_placeholder", "command"), p.helpColorScheme.UsageText),
						cc,
					)
				} else {
					subcommands := allcmd.sortedVisibleCommands()
					names := make([]string, len(subcommands))

					for i, subc := range subcommands {
						names[i] = subc.Name
					}

					_, _ = fmt.Fprintf(
						wr,
						" %s%s%s",
						co,
						p.colorizeHelp(strings.Join(names, " | "), p.helpColorScheme.UsageText),
						cc,
					)
				}
			}

			allcmd = allcmd.Active
		}

		_, _ = fmt.Fprintln(wr)

		longDescription := cmd.localizedLongDescription()
		if len(longDescription) != 0 {
			_, _ = fmt.Fprintln(wr)

			t := wrapText(longDescription,
				aligninfo.terminalColumns,
				"",
				trimDescriptions)

			_, _ = fmt.Fprintln(wr, p.colorizeHelp(t, p.helpColorScheme.LongDescription))
		}
	}

	c := p.Command

	for c != nil {
		printcmd := c != p.Command
		optionAlignInfo := aligninfo
		if printcmd {
			optionAlignInfo.reserveShort = commandHasVisibleShortOption(c)
			optionAlignInfo.indent = p.commandOptionIndent
		}

		c.eachGroup(func(grp *Group) {
			first := true

			// Skip built-in help group for all commands except the top-level
			// parser
			if grp.Hidden || (grp.isBuiltinHelp && c != p.Command) {
				return
			}

			for _, info := range c.sortedOptionsForGroup(grp) {
				if !info.showInHelp() {
					continue
				}

				if printcmd {
					header := p.i18nTextf(
						"help.command.options_header",
						"[{command} command options]",
						map[string]string{"command": c.Name},
					)
					_, _ = fmt.Fprintf(
						wr,
						"\n%s\n",
						p.colorizeHelp(header, p.helpColorScheme.SubcommandOptionsHeader),
					)
					printcmd = false
				}

				if first && cmd.Group != grp {
					_, _ = fmt.Fprintln(wr)

					if optionAlignInfo.indent > 0 {
						_, _ = wr.WriteString(strings.Repeat(" ", optionAlignInfo.indent))
					}

					_, _ = fmt.Fprintf(wr, "%s:\n", p.colorizeHelp(grp.localizedShortDescription(), p.helpColorScheme.GroupHeader))
					first = false
				}

				p.writeHelpOption(wr, info, optionAlignInfo, trimDescriptions, format)
			}
		})

		var args []*Arg
		for _, arg := range c.args {
			if arg.localizedDescription() != "" {
				args = append(args, arg)
			}
		}

		if len(args) > 0 {
			if c == p.Command {
				_, _ = fmt.Fprintf(
					wr,
					"\n%s:\n",
					p.colorizeHelp(p.i18nText("help.arguments", "Arguments"), p.helpColorScheme.ArgumentsHeader),
				)
			} else {
				header := p.i18nTextf(
					"help.command.arguments_header",
					"[{command} command arguments]",
					map[string]string{"command": c.Name},
				)
				_, _ = fmt.Fprintf(wr, "\n%s\n", p.colorizeHelp(header, p.helpColorScheme.ArgumentsHeader))
			}

			for _, arg := range args {
				p.writeHelpArgument(wr, arg, &aligninfo, paddingBeforeOption, trimDescriptions)
			}
		}

		c = c.Active
	}

	scommands := cmd.sortedVisibleCommands()

	if len(scommands) > 0 {
		maxnamelen := maxCommandLength(scommands)

		_, _ = fmt.Fprintln(wr)
		_, _ = fmt.Fprintln(
			wr,
			p.colorizeHelp(
				p.i18nText("help.available_commands", "Available commands")+":",
				p.helpColorScheme.CommandSectionHeader,
			),
		)

		commandGroups := groupedHelpCommands(p, scommands)
		commandIndent := paddingBeforeOption
		for gi, group := range commandGroups {
			if len(commandGroups) > 1 || group.name != "" {
				if group.name != "" {
					_, _ = fmt.Fprintf(
						wr,
						"%s%s:\n",
						strings.Repeat(" ", paddingBeforeOption),
						p.colorizeHelp(group.name, p.helpColorScheme.CommandGroupHeader),
					)
					commandIndent = paddingBeforeOption * 2
				} else {
					commandIndent = paddingBeforeOption
				}
			}

			for _, c := range group.commands {
				_, _ = fmt.Fprintf(wr, "%s%s", strings.Repeat(" ", commandIndent), p.colorizeHelp(c.Name, p.helpColorScheme.CommandName))

				shortDescription := c.localizedShortDescription()
				if len(shortDescription) > 0 {
					pad := strings.Repeat(" ", maxnamelen-textWidth(c.Name))
					_, _ = fmt.Fprintf(wr, "%s  %s", pad, p.colorizeHelp(shortDescription, p.helpColorScheme.CommandDesc))
				}

				if len(c.Aliases) > 0 &&
					(len(shortDescription) > 0 || (p.Options&ShowCommandAliases) != None) {
					aliases := p.i18nTextf(
						"help.command.aliases_suffix",
						" (aliases: {aliases})",
						map[string]string{"aliases": strings.Join(c.Aliases, ", ")},
					)
					_, _ = fmt.Fprintf(
						wr,
						"%s",
						p.colorizeHelp(aliases, p.helpColorScheme.CommandAliases),
					)
				}

				_, _ = fmt.Fprintln(wr)
			}
			if gi < len(commandGroups)-1 {
				_, _ = fmt.Fprintln(wr)
			}
		}
	}

	if (p.Options&ColorHelp) != None && p.helpColorEnabled {
		writeANSIReset(wr)
	}

	_ = wr.Flush()
}

type alignmentInfo struct {
	maxLeftLen      int
	terminalColumns int
	reserveShort    bool
	unlimitedWidth  bool
	indent          int
}

const (
	paddingBeforeOption                 = 2
	distanceBetweenOptionAndDescription = 2
	minHelpLeftWidth                    = 8
	minHelpDescriptionWidth             = 10
)

func (a *alignmentInfo) descriptionStart() int {
	return a.maxLeftLen + distanceBetweenOptionAndDescription
}

func (a *alignmentInfo) optionDescriptionStart() int {
	descstart := a.descriptionStart()

	if a.unlimitedWidth {
		return descstart
	}

	if a.terminalColumns <= 0 {
		a.terminalColumns = defaultTermSize
	}

	minDescStart := paddingBeforeOption + minHelpLeftWidth + distanceBetweenOptionAndDescription
	maxLeftHalf := max(a.terminalColumns/2, minDescStart)
	maxDescStart := max(a.terminalColumns-minHelpDescriptionWidth, minDescStart)
	maxAllowed := min(maxLeftHalf, maxDescStart)
	if descstart <= maxAllowed {
		return descstart
	}

	return maxAllowed
}

func (a *alignmentInfo) updateLen(name string, indent int) {
	l := textWidth(name)

	l += indent

	if l > a.maxLeftLen {
		a.maxLeftLen = l
	}
}

func (p *Parser) getAlignmentInfo() alignmentInfo {
	ret := alignmentInfo{
		maxLeftLen:      0,
		reserveShort:    false,
		terminalColumns: p.helpColumns(),
	}
	if p.helpWidthSet && p.helpWidth == 0 {
		ret.terminalColumns = unlimitedHelpWidth
		ret.unlimitedWidth = true
	}

	p.eachActiveGroup(func(_ *Command, grp *Group) {
		if !grp.showInHelp() {
			return
		}
		for _, info := range grp.options {
			if !info.showInHelp() {
				continue
			}

			if info.ShortName != 0 {
				ret.reserveShort = true
			}
		}
	})

	format := p.optionRenderFormat()
	var prevcmd *Command

	p.eachActiveGroup(func(c *Command, grp *Group) {
		indent := 0
		if c != p.Command {
			indent = p.commandOptionIndent
		}

		if c != prevcmd {
			for _, arg := range c.args {
				argLeft := strings.Repeat(" ", paddingBeforeOption) + arg.localizedName()
				if arg.localizedDescription() != "" {
					argLeft += ":"
				}
				ret.updateLen(argLeft, 0)
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

			prefix := paddingBeforeOption + indent
			if info.ShortName == 0 && ret.reserveShort {
				prefix += 4
			}

			leftRaw, _, _, _ := p.buildHelpOptionLeft(info, format, prefix)
			ret.updateLen(leftRaw, 0)
		}
	})

	return ret
}

const unlimitedHelpWidth = 1 << 30

func (p *Parser) helpColumns() int {
	if p.helpWidthSet {
		return p.helpWidth
	}

	width, _ := DetectTerminalSize()
	if width <= 0 {
		return defaultTermSize
	}

	return width
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

		for textWidth(line) > l {
			suffix := ""
			pos, splitOnSpace := splitTextWidth(line, l)
			if !splitOnSpace {
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

func wrapTextNoHyphen(s string, l int) string {
	var ret string

	if l < 10 {
		l = 10
	}

	lines := strings.SplitSeq(s, "\n")

	for line := range lines {
		var retline string

		for textWidth(line) > l {
			pos, splitOnSpace := splitTextWidth(line, l)
			if len(retline) != 0 {
				retline += "\n"
			}

			part := line[:pos]
			switch {
			case splitOnSpace:
				line = line[pos+1:]
			default:
				line = line[pos:]
			}

			retline += part
		}

		if len(line) > 0 {
			if len(retline) != 0 {
				retline += "\n"
			}

			retline += line
		}

		if len(ret) > 0 {
			ret += "\n"

			// no-op: no additional prefix for wrapped continuation lines
		}

		ret += retline
	}

	return ret
}

func optionIsRepeatable(option *Option) bool {
	kind := option.value.Type().Kind()
	return kind == reflect.Slice || kind == reflect.Map
}

func renderChoiceToken(option *Option) string {
	if len(option.Choices) == 0 {
		return ""
	}

	return renderChoiceInline(option.Choices)
}

type optionTailLine struct {
	Text     string
	IsChoice bool
	IsValue  bool
}

type choiceRenderMode int

const (
	choiceRenderInline choiceRenderMode = iota
	choiceRenderPacked
	choiceRenderList
)

type helpLeftLineKind int

const (
	helpLeftLineDefault helpLeftLineKind = iota
	helpLeftLineChoice
	helpLeftLineValue
)

type helpLeftLine struct {
	Text string
	Kind helpLeftLineKind
}

type helpOptionLayout struct {
	LeftPrefix  int
	DescStart   int
	LeftWidth   int
	DescWidth   int
	ShouldSplit bool
}

func renderChoicePipeLines(choices []string, width int) []string {
	if len(choices) == 0 {
		return nil
	}
	if width < 10 {
		width = 10
	}

	lines := make([]string, 0, len(choices))
	current := "[" + choices[0]

	for _, choice := range choices[1:] {
		token := "|" + choice
		if textWidth(current+token+"]") <= width {
			current += token
			continue
		}

		lines = append(lines, current)
		current = "|" + choice
	}

	if textWidth(current+"]") <= width {
		current += "]"
		lines = append(lines, current)
		return lines
	}

	avail := max(width-1, 1)
	wrapped := strings.Split(wrapTextNoHyphen(current, avail), "\n")
	if len(wrapped) == 0 {
		return []string{"]"}
	}
	for i, line := range wrapped {
		if i == len(wrapped)-1 {
			lines = append(lines, line+"]")
			continue
		}
		lines = append(lines, line)
	}

	return lines
}

func renderChoiceListLines(choices []string, width int, label string) []string {
	lines := []string{label + ":"}

	appendChoice := func(choice string) {
		prefix := "> "
		avail := width - textWidth(prefix)
		if avail < 1 {
			lines = append(lines, prefix+choice)
			return
		}

		wrapped := strings.Split(wrapTextNoHyphen(choice, avail), "\n")
		if len(wrapped) == 0 {
			return
		}

		lines = append(lines, prefix+wrapped[0])
		for _, line := range wrapped[1:] {
			lines = append(lines, "  "+line)
		}
	}

	for _, choice := range choices {
		appendChoice(choice)
	}

	return lines
}

func renderChoiceInline(choices []string) string {
	if len(choices) == 0 {
		return ""
	}

	return "[" + strings.Join(choices, "|") + "]"
}

func chooseChoiceRenderMode(
	choices []string,
	width int,
	forceChoiceList bool,
	autoChoiceList bool,
) choiceRenderMode {
	if len(choices) == 0 {
		return choiceRenderInline
	}
	if forceChoiceList {
		return choiceRenderList
	}

	inline := renderChoiceInline(choices)
	if textWidth(inline) <= width {
		return choiceRenderInline
	}

	if !autoChoiceList {
		return choiceRenderPacked
	}

	packed := renderChoicePipeLines(choices, width)
	// Auto list mode is a last resort for very narrow layouts where
	// packed by-separator output becomes too tall to read comfortably.
	if width < 24 && len(packed) > 3 {
		return choiceRenderList
	}

	return choiceRenderPacked
}

func splitOptionTailLines(
	valueName string,
	choices []string,
	width int,
	forceChoiceList bool,
	autoChoiceList bool,
	choiceListLabel string,
) []optionTailLine {
	if width < 10 {
		width = 10
	}

	if len(choices) == 0 {
		if valueName == "" {
			return nil
		}
		raw := strings.Split(wrapTextNoHyphen(valueName, width), "\n")
		lines := make([]optionTailLine, 0, len(raw))
		for _, line := range raw {
			lines = append(lines, optionTailLine{Text: line, IsValue: true})
		}
		return lines
	}

	lines := make([]optionTailLine, 0, len(choices)+2)
	if valueName != "" {
		for line := range strings.SplitSeq(wrapTextNoHyphen(valueName, width), "\n") {
			lines = append(lines, optionTailLine{Text: line, IsValue: true})
		}
	}

	switch chooseChoiceRenderMode(choices, width, forceChoiceList, autoChoiceList) {
	case choiceRenderList:
		lines = appendChoiceTailLines(lines, renderChoiceListLines(choices, width, choiceListLabel))
	case choiceRenderInline:
		lines = append(lines, optionTailLine{Text: renderChoiceInline(choices), IsChoice: true})
	default:
		lines = appendChoiceTailLines(lines, renderChoicePipeLines(choices, width))
	}

	return lines
}

func appendChoiceTailLines(lines []optionTailLine, choiceLines []string) []optionTailLine {
	for _, line := range choiceLines {
		lines = append(lines, optionTailLine{Text: line, IsChoice: true})
	}

	return lines
}

func splitAdaptiveLeftBody(
	option *Option,
	format optionRenderFormat,
	leftBody string,
	leftPrefix int,
	leftWidth int,
	forceChoiceList bool,
	autoChoiceList bool,
	choiceListLabel string,
) []helpLeftLine {
	indentPrefix := strings.Repeat(" ", leftPrefix)
	lines := make([]helpLeftLine, 0, 6)
	wrapWidth := max(leftWidth, minHelpLeftWidth)
	appendWrapped := func(text string, kind helpLeftLineKind) {
		for line := range strings.SplitSeq(wrapTextNoHyphen(text, wrapWidth), "\n") {
			lines = append(lines, helpLeftLine{Text: indentPrefix + line, Kind: kind})
		}
	}

	if !option.canArgument() {
		appendWrapped(leftBody, helpLeftLineDefault)
		return lines
	}

	idx := strings.IndexRune(leftBody, format.nameDelimiter)
	if idx < 0 || idx+1 >= len(leftBody) {
		appendWrapped(leftBody, helpLeftLineDefault)
		return lines
	}

	head := leftBody[:idx+1]
	valueName := option.localizedValueName()
	if valueName != "" {
		candidate := head + valueName
		if textWidth(candidate) <= leftWidth {
			head = candidate
			valueName = ""
		}
	}

	appendWrapped(head, helpLeftLineDefault)

	continuationPrefix := strings.Repeat(" ", leftPrefix+2)
	tailWidth := max(leftWidth-2, minHelpLeftWidth)
	for _, line := range splitOptionTailLines(
		valueName,
		option.Choices,
		tailWidth,
		forceChoiceList,
		autoChoiceList,
		choiceListLabel,
	) {
		kind := helpLeftLineDefault
		if line.IsChoice {
			kind = helpLeftLineChoice
		} else if line.IsValue {
			kind = helpLeftLineValue
		}
		lines = append(lines, helpLeftLine{Text: continuationPrefix + line.Text, Kind: kind})
	}

	return lines
}

func (p *Parser) buildHelpOptionLayout(
	option *Option,
	info alignmentInfo,
	format optionRenderFormat,
	forceSplit bool,
) (helpOptionLayout, string, string, string, string, string) {
	layout := helpOptionLayout{}
	prefix := paddingBeforeOption + info.indent
	layout.LeftPrefix = prefix
	if option.ShortName == 0 && info.reserveShort {
		layout.LeftPrefix += 4
	}

	layout.DescStart = info.optionDescriptionStart()
	layout.LeftWidth = max(
		layout.DescStart-layout.LeftPrefix-distanceBetweenOptionAndDescription,
		minHelpLeftWidth,
	)

	layout.DescWidth = max(info.terminalColumns-layout.DescStart, minHelpDescriptionWidth)

	leftRaw, shortToken, longToken, choicesToken := p.buildHelpOptionLeft(
		option,
		format,
		layout.LeftPrefix,
	)

	indentPrefix := strings.Repeat(" ", layout.LeftPrefix)
	leftBody := strings.TrimPrefix(leftRaw, indentPrefix)
	leftLen := textWidth(leftBody)
	layout.ShouldSplit = forceSplit || leftLen > layout.LeftWidth

	return layout, leftRaw, leftBody, shortToken, longToken, choicesToken
}

func (p *Parser) renderHelpOptionLeftLines(
	option *Option,
	format optionRenderFormat,
	layout helpOptionLayout,
	leftBody string,
	shortToken string,
	longToken string,
	choicesToken string,
) ([]string, []string) {
	forceChoiceList := (p.Options & ShowChoiceListInHelp) != None
	autoChoiceList := (p.Options & AutoShowChoiceListInHelp) != None
	leftLinesMeta := splitAdaptiveLeftBody(
		option,
		format,
		leftBody,
		layout.LeftPrefix,
		layout.LeftWidth,
		forceChoiceList,
		autoChoiceList,
		p.i18nText("help.meta.valid_values", "valid values"),
	)

	leftLines := make([]string, len(leftLinesMeta))
	leftLinesPlain := make([]string, len(leftLinesMeta))
	for i, lineMeta := range leftLinesMeta {
		leftLinesPlain[i] = lineMeta.Text
		colored := p.colorizeHelpOptionLeftLine(lineMeta.Text, option, format, shortToken, longToken, choicesToken)
		switch lineMeta.Kind {
		case helpLeftLineChoice:
			colored = p.colorizeHelp(colored, p.helpColorScheme.OptionChoices)
		case helpLeftLineValue:
			colored = p.colorizeHelp(colored, p.helpColorScheme.OptionValueName)
		}
		colored = p.colorizeOptionPunctuation(colored, format)
		leftLines[i] = colored
	}

	return leftLines, leftLinesPlain
}

func (p *Parser) colorizeHelpOptionLeftLine(
	line string,
	option *Option,
	format optionRenderFormat,
	shortToken string,
	longToken string,
	choicesToken string,
) string {
	if shortToken != "" {
		line = strings.Replace(line, shortToken, p.colorizeHelp(shortToken, p.helpColorScheme.OptionShort), 1)
	}
	if longToken != "" {
		coloredLong := format.longDelimiter + longToken
		line = strings.Replace(line, coloredLong, p.colorizeHelp(coloredLong, p.helpColorScheme.OptionLong), 1)
	}
	if choicesToken != "" {
		line = strings.Replace(line, choicesToken, p.colorizeHelp(choicesToken, p.helpColorScheme.OptionChoices), 1)
	}
	if option.canArgument() {
		valueName := option.localizedValueName()
		if valueName != "" {
			line = strings.Replace(line, valueName, p.colorizeHelp(valueName, p.helpColorScheme.OptionValueName), 1)
		}
	}

	return line
}

func (p *Parser) renderHelpOptionRows(
	writer *bufio.Writer,
	leftLines []string,
	leftLinesPlain []string,
	descLines []string,
	layout helpOptionLayout,
	trimDescriptions bool,
	defaultFrag string,
	envFrag string,
	repeatableFrag string,
) {
	if len(descLines) == 0 {
		for _, line := range leftLines {
			_, _ = writer.WriteString(line)
			_, _ = writer.WriteString("\n")
		}
		return
	}

	desc := strings.Join(descLines, "\n")
	descWrapped := wrapText(desc, layout.DescWidth, "", trimDescriptions)

	if defaultFrag != "" {
		descWrapped = strings.Replace(descWrapped, defaultFrag, p.colorizeHelp(defaultFrag, p.helpColorScheme.OptionDefault), 1)
	}
	if envFrag != "" {
		descWrapped = strings.Replace(descWrapped, envFrag, p.colorizeHelp(envFrag, p.helpColorScheme.OptionEnv), 1)
	}
	if repeatableFrag != "" {
		descWrapped = strings.Replace(descWrapped, repeatableFrag, p.colorizeHelp(repeatableFrag, p.helpColorScheme.OptionChoices), 1)
	}

	descWrapped = p.colorizeHelp(descWrapped, p.helpColorScheme.OptionDesc)
	wrappedDescLines := strings.Split(descWrapped, "\n")
	total := max(len(wrappedDescLines), len(leftLines))

	for i := range total {
		leftPlain := ""
		leftLine := ""
		if i < len(leftLinesPlain) {
			leftPlain = leftLinesPlain[i]
			leftLine = leftLines[i]
		}

		_, _ = writer.WriteString(leftLine)
		if i < len(wrappedDescLines) {
			pad := max(layout.DescStart-textWidth(leftPlain), distanceBetweenOptionAndDescription)
			_, _ = writer.WriteString(strings.Repeat(" ", pad))
			_, _ = writer.WriteString(wrappedDescLines[i])
		}
		_, _ = writer.WriteString("\n")
	}
}

func (p *Parser) buildHelpOptionLeft(option *Option, format optionRenderFormat, prefix int) (string, string, string, string) {
	left := &bytes.Buffer{}
	shortToken := ""
	longToken := ""
	choicesToken := ""

	left.WriteString(strings.Repeat(" ", prefix))

	if option.ShortName != 0 {
		left.WriteRune(format.shortDelimiter)
		left.WriteRune(option.ShortName)
		shortToken = string(format.shortDelimiter) + string(option.ShortName)
	}

	if len(option.LongName) > 0 {
		if option.ShortName != 0 {
			left.WriteString(", ")
		}

		left.WriteString(format.longDelimiter)
		longToken = option.LongNameWithNamespace()
		left.WriteString(longToken)
	}

	if option.canArgument() {
		left.WriteRune(format.nameDelimiter)

		valueName := option.localizedValueName()
		if len(valueName) > 0 {
			left.WriteString(valueName)
		}

		choicesToken = renderChoiceToken(option)
		if choicesToken != "" {
			left.WriteString(choicesToken)
		}
	}

	return left.String(), shortToken, longToken, choicesToken
}

func (p *Parser) buildHelpOptionDescription(
	option *Option,
	format optionRenderFormat,
	descWidth int,
) ([]string, string, string, string) {
	description := option.localizedDescription()
	if description == "" {
		return nil, "", "", ""
	}

	def := ""
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

	defaultFrag := ""
	if def != "" {
		defaultFrag = fmt.Sprintf("%s: %v", p.i18nText("help.meta.default", "default"), def)
	}

	envFrag := ""
	if (p.Options&HideEnvInHelp) == None && option.EnvKeyWithNamespace() != "" {
		envFrag = format.envPrefix + option.EnvKeyWithNamespace() + format.envSuffix
	}

	repeatableFrag := ""
	if (p.Options&ShowRepeatableInHelp) != None && optionIsRepeatable(option) {
		repeatableFrag = p.i18nText("help.meta.repeatable", "repeatable")
	}

	extras := make([]string, 0, 3)
	if defaultFrag != "" {
		extras = append(extras, "("+defaultFrag+")")
	}
	if envFrag != "" {
		extras = append(extras, "["+envFrag+"]")
	}
	if repeatableFrag != "" {
		extras = append(extras, "("+repeatableFrag+")")
	}

	desc := description
	if len(extras) == 0 {
		return strings.Split(desc, "\n"), defaultFrag, envFrag, repeatableFrag
	}

	lines := strings.Split(desc, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	if descWidth <= 0 {
		lines = append(lines, extras...)
		return lines, defaultFrag, envFrag, repeatableFrag
	}

	for _, extra := range extras {
		last := lines[len(lines)-1]
		candidate := extra
		if last != "" {
			candidate = last + " " + extra
		}

		// Keep each extra as an atomic block: if it does not fit, move it
		// entirely to the next line instead of splitting it.
		if textWidth(candidate) <= descWidth {
			lines[len(lines)-1] = candidate
			continue
		}

		lines = append(lines, extra)
	}

	return lines, defaultFrag, envFrag, repeatableFrag
}

func isHelpTextStyleSet(style HelpTextStyle) bool {
	return style.UseFG || style.UseBG || style.Bold || style.Italic || style.Underline
}

func (p *Parser) colorizeOptionPunctuation(text string, format optionRenderFormat) string {
	style := p.helpColorScheme.OptionPunctuation
	if text == "" || !isHelpTextStyleSet(style) {
		return text
	}

	for _, token := range []string{
		",",
		string(format.nameDelimiter),
	} {
		colored := p.colorizeHelp(token, style)
		text = strings.ReplaceAll(text, token, colored)
	}

	return text
}

func (p *Parser) writeLaidOutHelpOption(
	writer *bufio.Writer,
	option *Option,
	info alignmentInfo,
	trimDescriptions bool,
	format optionRenderFormat,
	forceSplit bool,
) {
	layout, leftRaw, leftBody, shortToken, longToken, choicesToken := p.buildHelpOptionLayout(
		option,
		info,
		format,
		forceSplit,
	)
	var leftLines []string
	var leftLinesPlain []string
	if layout.ShouldSplit {
		leftLines, leftLinesPlain = p.renderHelpOptionLeftLines(
			option,
			format,
			layout,
			leftBody,
			shortToken,
			longToken,
			choicesToken,
		)
	} else {
		leftLinesPlain = []string{leftRaw}
		leftLine := p.colorizeHelpOptionLeftLine(leftRaw, option, format, shortToken, longToken, choicesToken)
		leftLines = []string{p.colorizeOptionPunctuation(leftLine, format)}
	}

	descLines, defaultFrag, envFrag, repeatableFrag := p.buildHelpOptionDescription(
		option,
		format,
		layout.DescWidth,
	)
	p.renderHelpOptionRows(
		writer,
		leftLines,
		leftLinesPlain,
		descLines,
		layout,
		trimDescriptions,
		defaultFrag,
		envFrag,
		repeatableFrag,
	)
}

func (p *Parser) writeHelpOption(
	writer *bufio.Writer,
	option *Option,
	info alignmentInfo,
	trimDescriptions bool,
	format optionRenderFormat,
) {
	if option.Hidden {
		return
	}

	forceChoiceListSplit := (p.Options&ShowChoiceListInHelp) != None &&
		len(option.Choices) > 0 &&
		option.canArgument()
	p.writeLaidOutHelpOption(writer, option, info, trimDescriptions, format, forceChoiceListSplit)
}

func (p *Parser) writeHelpArgument(
	wr *bufio.Writer,
	arg *Arg,
	aligninfo *alignmentInfo,
	indent int,
	trimDescriptions bool,
) {
	descStart := aligninfo.optionDescriptionStart()
	argPrefix := strings.Repeat(" ", indent)
	argPrefix += arg.localizedName()

	argDescriptionText := arg.localizedDescription()
	if len(argDescriptionText) > 0 {
		argPrefix += ":"
		_, _ = wr.WriteString(p.colorizeHelp(argPrefix, p.helpColorScheme.ArgumentName))

		descPadding := strings.Repeat(
			" ",
			max(descStart-textWidth(argPrefix), 1),
		)
		descWidth := aligninfo.terminalColumns - 1 - descStart
		descPrefix := strings.Repeat(" ", descStart)

		_, _ = wr.WriteString(descPadding)
		argDescription := argDescriptionText
		if def := arg.defaultLiteral(); def != "" {
			defaultLabel := p.i18nText("help.meta.default", "default")
			argDescription += " (" + defaultLabel + ": " + def + ")"
		}
		if (p.Options&ShowRepeatableInHelp) != None && arg.isRemaining() {
			repeatableLabel := p.i18nText("help.meta.repeatable", "repeatable")
			argDescription += " (" + repeatableLabel + ")"
		}

		argDesc := wrapText(argDescription, descWidth, descPrefix, trimDescriptions)
		_, _ = wr.WriteString(p.colorizeHelp(argDesc, p.helpColorScheme.ArgumentDesc))
	} else {
		_, _ = wr.WriteString(p.colorizeHelp(argPrefix, p.helpColorScheme.ArgumentName))
	}

	_, _ = fmt.Fprintln(wr)
}

func commandHasVisibleShortOption(c *Command) bool {
	for _, grp := range c.groups {
		if !grp.showInHelp() || grp.isBuiltinHelp {
			continue
		}

		for _, opt := range grp.options {
			if opt.showInHelp() && opt.ShortName != 0 {
				return true
			}
		}
	}

	return false
}

func maxCommandLength(s []*Command) int {
	if len(s) == 0 {
		return 0
	}

	ret := textWidth(s[0].Name)

	for _, v := range s[1:] {
		l := textWidth(v.Name)

		if l > ret {
			ret = l
		}
	}

	return ret
}

type helpCommandGroup struct {
	name     string
	commands []*Command
	sortRank int
}

func groupedHelpCommands(p *Parser, commands []*Command) []helpCommandGroup {
	groups := make([]helpCommandGroup, 0)
	index := make(map[string]int)

	hasNamedGroups := false
	for _, command := range commands {
		if command.localizedCommandGroup() != "" {
			hasNamedGroups = true
			break
		}
	}

	for _, command := range commands {
		name := command.localizedCommandGroup()
		rank := 1
		if hasNamedGroups && name == "" {
			name = p.i18nText("help.command_group.main_commands", "Main Commands")
			rank = 0
		}
		if _, ok := command.data.(builtinCommand); ok {
			rank = 2
		}

		key := fmt.Sprintf("%d\x00%s", rank, name)
		idx, ok := index[key]
		if !ok {
			idx = len(groups)
			index[key] = idx
			groups = append(groups, helpCommandGroup{name: name, sortRank: rank})
		}
		groups[idx].commands = append(groups[idx].commands, command)
	}

	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].sortRank != groups[j].sortRank {
			return groups[i].sortRank < groups[j].sortRank
		}
		return groups[i].name < groups[j].name
	})

	return groups
}
