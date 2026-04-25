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
	"strings"
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

func (a *alignmentInfo) optionDescriptionStart() int {
	descstart := a.descriptionStart() + paddingBeforeOption

	if a.maxLongLen <= a.terminalColumns/2 {
		return descstart
	}

	maxDescStart := min(max(a.terminalColumns/2, paddingBeforeOption+12), a.terminalColumns-10)

	if descstart > maxDescStart {
		descstart = maxDescStart
	}

	return descstart
}

func (a *alignmentInfo) updateLen(name string, indent bool) {
	l := textWidth(name)

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
				ret.updateLen(arg.localizedName(), c != p.Command)
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

			valueName := info.localizedValueName()
			if len(valueName) > 0 {
				ret.hasValueName = true
			}

			l := info.LongNameWithNamespace() + valueName
			// Choices are rendered as a tail block and can move to their own
			// lines in adaptive mode; keep base alignment focused on flag/value
			// tokens to avoid over-expanding description gap.

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

func renderChoiceToken(option *Option, leftWidth int, prefixLen int) string {
	if len(option.Choices) == 0 {
		return ""
	}

	compact := "[" + strings.Join(option.Choices, "|") + "]"
	if prefixLen+textWidth(compact) <= leftWidth {
		return compact
	}

	return "[" + strings.Join(option.Choices, " | ") + "]"
}

type optionTailLine struct {
	Text     string
	IsChoice bool
}

func renderChoicePipeLines(choices []string, width int) []string {
	lines := make([]string, 0, len(choices)+1)

	appendChoice := func(prefix string, continuationPrefix string, choice string) {
		avail := width - textWidth(prefix)
		if avail < 1 {
			lines = append(lines, prefix)
			prefix = continuationPrefix
			avail = width - textWidth(prefix)
			if avail < 1 {
				avail = width
				prefix = ""
			}
		}

		wrapped := strings.Split(wrapTextNoHyphen(choice, avail), "\n")
		if len(wrapped) == 0 {
			return
		}

		lines = append(lines, prefix+wrapped[0])

		for _, line := range wrapped[1:] {
			lines = append(lines, continuationPrefix+line)
		}
	}

	appendChoice("[", "  ", choices[0])

	for _, choice := range choices[1:] {
		appendChoice("| ", "  ", choice)
	}

	if len(lines) > 0 {
		last := lines[len(lines)-1]
		if textWidth(last)+1 <= width {
			lines[len(lines)-1] = last + "]"
		} else {
			lines = append(lines, "]")
		}
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

func splitOptionTailLines(
	valueName string,
	choices []string,
	width int,
	forceChoiceList bool,
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
			lines = append(lines, optionTailLine{Text: line})
		}
		return lines
	}

	lines := make([]optionTailLine, 0, len(choices)+2)
	if valueName != "" {
		for line := range strings.SplitSeq(wrapTextNoHyphen(valueName, width), "\n") {
			lines = append(lines, optionTailLine{Text: line})
		}
	}

	pipeLines := renderChoicePipeLines(choices, width)
	choiceLineCount := 0
	for _, line := range pipeLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "]" {
			continue
		}
		choiceLineCount++
	}
	threshold := (len(choices) + 1) / 2 // ceil(len/2)
	useChoiceList := forceChoiceList || (len(choices) >= 3 && choiceLineCount > threshold)
	if useChoiceList {
		for _, line := range renderChoiceListLines(choices, width, choiceListLabel) {
			lines = append(lines, optionTailLine{Text: line, IsChoice: true})
		}
		return lines
	}

	for _, line := range pipeLines {
		lines = append(lines, optionTailLine{Text: line, IsChoice: true})
	}
	return lines
}

func (p *Parser) buildHelpOptionLeft(option *Option, format optionRenderFormat, leftWidth int, prefix int) (string, string, string, string) {
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

		choicesToken = renderChoiceToken(option, leftWidth, textWidth(left.String()))
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

func (p *Parser) adaptiveWriteHelpOption(
	writer *bufio.Writer,
	option *Option,
	info alignmentInfo,
	trimDescriptions bool,
	format optionRenderFormat,
	forceSplit bool,
) bool {
	prefix := paddingBeforeOption
	if info.indent {
		prefix += 4
	}

	leftPrefix := prefix
	if option.ShortName == 0 && info.hasShort {
		leftPrefix += 4
	}

	descstart := info.optionDescriptionStart()

	leftWidth := descstart - leftPrefix
	if leftWidth < 10 {
		return false
	}

	leftRaw, shortToken, longToken, choicesToken := p.buildHelpOptionLeft(option, format, leftWidth, leftPrefix)
	indentPrefix := strings.Repeat(" ", leftPrefix)
	leftBody := strings.TrimPrefix(leftRaw, indentPrefix)

	leftLen := textWidth(leftBody)
	if !forceSplit && leftLen <= leftWidth+8 {
		return false
	}

	leftLinesPlain := make([]string, 0, 4)
	leftLinesChoice := make([]bool, 0, 4)
	if option.canArgument() {
		if idx := strings.IndexRune(leftBody, format.nameDelimiter); idx >= 0 && idx+1 < len(leftBody) {
			head := leftBody[:idx+1]
			headWrapped := wrapTextNoHyphen(head, leftWidth)
			for line := range strings.SplitSeq(headWrapped, "\n") {
				leftLinesPlain = append(leftLinesPlain, indentPrefix+line)
				leftLinesChoice = append(leftLinesChoice, false)
			}

			continuationPrefix := strings.Repeat(" ", leftPrefix+2)
			tailWidth := max(leftWidth-2, 10)

			forceChoiceList := (p.Options & ShowChoiceListInHelp) != None
			for _, line := range splitOptionTailLines(
				option.localizedValueName(),
				option.Choices,
				tailWidth,
				forceChoiceList,
				p.i18nText("help.meta.valid_values", "valid values"),
			) {
				leftLinesPlain = append(leftLinesPlain, continuationPrefix+line.Text)
				leftLinesChoice = append(leftLinesChoice, line.IsChoice)
			}
		}
	}

	if len(leftLinesPlain) == 0 {
		leftWrapped := wrapTextNoHyphen(leftBody, leftWidth)
		for line := range strings.SplitSeq(leftWrapped, "\n") {
			leftLinesPlain = append(leftLinesPlain, indentPrefix+line)
			leftLinesChoice = append(leftLinesChoice, false)
		}
	}

	leftLines := make([]string, len(leftLinesPlain))

	for i, line := range leftLinesPlain {
		colored := line
		if shortToken != "" {
			colored = strings.Replace(colored, shortToken, p.colorizeHelp(shortToken, p.helpColorScheme.OptionShort), 1)
		}
		if longToken != "" {
			coloredLong := format.longDelimiter + longToken
			colored = strings.Replace(colored, coloredLong, p.colorizeHelp(coloredLong, p.helpColorScheme.OptionLong), 1)
		}
		if choicesToken != "" {
			colored = strings.Replace(colored, choicesToken, p.colorizeHelp(choicesToken, p.helpColorScheme.OptionChoices), 1)
		}
		if i < len(leftLinesChoice) && leftLinesChoice[i] {
			colored = p.colorizeHelp(colored, p.helpColorScheme.OptionChoices)
		}
		leftLines[i] = colored
	}

	descWidth := info.terminalColumns - descstart
	descLines, defaultFrag, envFrag, repeatableFrag := p.buildHelpOptionDescription(option, format, descWidth)

	if len(descLines) == 0 {
		for _, line := range leftLines {
			_, _ = writer.WriteString(line)
			_, _ = writer.WriteString("\n")
		}
		return true
	}

	desc := strings.Join(descLines, "\n")
	descWrapped := wrapText(desc, info.terminalColumns-descstart, "", trimDescriptions)

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
			pad := max(descstart-textWidth(leftPlain), 1)
			_, _ = writer.WriteString(strings.Repeat(" ", pad))
			_, _ = writer.WriteString(wrappedDescLines[i])
		}

		_, _ = writer.WriteString("\n")
	}

	return true
}

func (p *Parser) writeHelpOption(
	writer *bufio.Writer,
	option *Option,
	info alignmentInfo,
	trimDescriptions bool,
	format optionRenderFormat,
) {
	var line strings.Builder
	shortToken := ""
	longToken := ""
	choicesToken := ""

	prefix := paddingBeforeOption

	if info.indent {
		prefix += 4
	}

	if option.Hidden {
		return
	}

	forceChoiceListSplit := (p.Options&ShowChoiceListInHelp) != None &&
		len(option.Choices) > 0 &&
		option.canArgument()
	if p.adaptiveWriteHelpOption(writer, option, info, trimDescriptions, format, forceChoiceListSplit) {
		return
	}

	line.Grow(64)
	line.WriteString(strings.Repeat(" ", prefix))

	if option.ShortName != 0 {
		line.WriteRune(format.shortDelimiter)
		line.WriteRune(option.ShortName)
		shortToken = string(format.shortDelimiter) + string(option.ShortName)
	} else if info.hasShort {
		line.WriteString("  ")
	}

	descstart := info.optionDescriptionStart()

	if len(option.LongName) > 0 {
		if option.ShortName != 0 {
			line.WriteString(", ")
		} else if info.hasShort {
			line.WriteString("  ")
		}

		line.WriteString(format.longDelimiter)
		longToken = option.LongNameWithNamespace()
		line.WriteString(longToken)
	}

	if option.canArgument() {
		line.WriteRune(format.nameDelimiter)

		valueName := option.localizedValueName()
		if len(valueName) > 0 {
			line.WriteString(valueName)
		}

		if len(option.Choices) > 0 {
			choicesToken = "[" + strings.Join(option.Choices, "|") + "]"
			line.WriteString(choicesToken)
		}
	}

	written := textWidth(line.String())
	lineText := line.String()
	if shortToken != "" {
		lineText = strings.Replace(lineText, shortToken, p.colorizeHelp(shortToken, p.helpColorScheme.OptionShort), 1)
	}
	if longToken != "" {
		coloredLong := format.longDelimiter + longToken
		lineText = strings.Replace(lineText, coloredLong, p.colorizeHelp(coloredLong, p.helpColorScheme.OptionLong), 1)
	}
	if choicesToken != "" {
		lineText = strings.Replace(lineText, choicesToken, p.colorizeHelp(choicesToken, p.helpColorScheme.OptionChoices), 1)
	}
	_, _ = writer.WriteString(lineText)

	if option.localizedDescription() != "" {
		dw := max(descstart-written, 1)
		_, _ = writer.WriteString(strings.Repeat(" ", dw))
		descWidth := info.terminalColumns - descstart
		descLines, defaultFrag, envFrag, repeatableFrag := p.buildHelpOptionDescription(option, format, descWidth)
		desc := wrapText(strings.Join(descLines, "\n"),
			descWidth,
			strings.Repeat(" ", descstart),
			trimDescriptions)

		if defaultFrag != "" {
			desc = strings.Replace(desc, defaultFrag, p.colorizeHelp(defaultFrag, p.helpColorScheme.OptionDefault), 1)
		}
		if envFrag != "" {
			desc = strings.Replace(desc, envFrag, p.colorizeHelp(envFrag, p.helpColorScheme.OptionEnv), 1)
		}
		if repeatableFrag != "" {
			desc = strings.Replace(desc, repeatableFrag, p.colorizeHelp(repeatableFrag, p.helpColorScheme.OptionChoices), 1)
		}

		desc = p.colorizeHelp(desc, p.helpColorScheme.OptionDesc)
		_, _ = writer.WriteString(desc)
	}

	_, _ = writer.WriteString("\n")
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
	p.helpColorEnabled = p.shouldUseColors(writer)
	defer func() {
		p.helpColorEnabled = prevHelpColorEnabled
	}()

	// Keep WriteHelp behavior consistent with ParseArgs:
	// when builtin help/version flags are enabled, ensure
	// corresponding options are present in help output.
	if (p.Options & (HelpFlag | VersionFlag)) != None {
		p.addHelpGroups(p.showBuiltinHelp, p.markVersionRequested)
	}

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
		if basePrefix != "" && p.helpColorScheme.BaseText.UseBG {
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
					aligninfo.indent = true
					printcmd = false
				}

				if first && cmd.Group != grp {
					_, _ = fmt.Fprintln(wr)

					if aligninfo.indent {
						_, _ = wr.WriteString("    ")
					}

					_, _ = fmt.Fprintf(wr, "%s:\n", p.colorizeHelp(grp.localizedShortDescription(), p.helpColorScheme.GroupHeader))
					first = false
				}

				p.writeHelpOption(wr, info, aligninfo, trimDescriptions, format)
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

			descStart := aligninfo.optionDescriptionStart()

			for _, arg := range args {
				argPrefix := strings.Repeat(" ", paddingBeforeOption)
				argPrefix += arg.localizedName()

				argDescriptionText := arg.localizedDescription()
				if len(argDescriptionText) > 0 {
					argPrefix += ":"
					_, _ = wr.WriteString(p.colorizeHelp(argPrefix, p.helpColorScheme.ArgumentName))

					// Space between "arg:" and the description start
					descPadding := strings.Repeat(
						" ",
						max(descStart-textWidth(argPrefix), 1),
					)
					// How much space the description gets before wrapping
					descWidth := aligninfo.terminalColumns - 1 - descStart
					// Whitespace to which we can indent new description lines
					descPrefix := strings.Repeat(" ", descStart)

					_, _ = wr.WriteString(descPadding)
					argDescription := argDescriptionText
					if def := arg.defaultLiteral(); def != "" {
						defaultLabel := p.i18nText("help.meta.default", "default")
						argDescription += " (" + defaultLabel + ": " + def + ")"
					}

					argDesc := wrapText(argDescription, descWidth, descPrefix, trimDescriptions)
					_, _ = wr.WriteString(p.colorizeHelp(argDesc, p.helpColorScheme.ArgumentDesc))
				} else {
					_, _ = wr.WriteString(p.colorizeHelp(argPrefix, p.helpColorScheme.ArgumentName))
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
		_, _ = fmt.Fprintln(
			wr,
			p.colorizeHelp(p.i18nText("help.available_commands", "Available commands")+":", p.helpColorScheme.CommandsHeader),
		)

		for _, c := range scommands {
			_, _ = fmt.Fprintf(wr, "  %s", p.colorizeHelp(c.Name, p.helpColorScheme.CommandName))

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
	}

	if basePrefix != "" {
		_, _ = wr.WriteString("\x1b[0m")
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
