// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type docParser struct {
	GeneratedAt      time.Time         `json:"-"`
	Meta             *docParserMeta    `json:"meta,omitempty"`
	Name             string            `json:"name"`
	ShortDescription string            `json:"short_description,omitempty"`
	LongDescription  string            `json:"long_description,omitempty"`
	Usage            string            `json:"usage,omitempty"`
	Args             []docArg          `json:"args,omitempty"`
	Groups           []docGroup        `json:"groups,omitempty"`
	Commands         []docCommand      `json:"commands,omitempty"`
	CommandGroups    []docCommandGroup `json:"-"`
}

// docParserMeta holds project metadata shared across all doc format templates.
// Man-page-specific fields (Section, SeeAlso) are also stored here but are
// only rendered by the man page template.
type docParserMeta struct {
	Author   string   `json:"author,omitempty"`
	Homepage string   `json:"homepage,omitempty"`
	BugsURL  string   `json:"bugs_url,omitempty"`
	License  string   `json:"license,omitempty"`
	SeeAlso  []string `json:"see_also,omitempty"`
	Section  int      `json:"section,omitempty"`
}

type docCommand struct {
	Name                string            `json:"name"`
	ShortDescription    string            `json:"short_description,omitempty"`
	LongDescription     string            `json:"long_description,omitempty"`
	UsageLine           string            `json:"usage_line,omitempty"`
	Group               string            `json:"group,omitempty"`
	Deprecated          string            `json:"deprecated,omitempty"`
	Aliases             []string          `json:"aliases,omitempty"`
	Args                []docArg          `json:"args,omitempty"`
	Groups              []docGroup        `json:"groups,omitempty"`
	Commands            []docCommand      `json:"commands,omitempty"`
	CommandGroups       []docCommandGroup `json:"-"`
	SubcommandsOptional bool              `json:"subcommands_optional,omitempty"`
	PassAfterNonOption  bool              `json:"pass_after_non_option,omitempty"`
	Hidden              bool              `json:"hidden,omitempty"`
}

type docArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type docCommandGroup struct {
	Name     string       `json:"name"`
	Commands []docCommand `json:"commands,omitempty"`
}

type docGroup struct {
	ShortDescription string      `json:"short_description,omitempty"`
	LongDescription  string      `json:"long_description,omitempty"`
	Namespace        string      `json:"namespace,omitempty"`
	EnvNamespace     string      `json:"env_namespace,omitempty"`
	Options          []docOption `json:"options,omitempty"`
	Hidden           bool        `json:"hidden,omitempty"`
}

type docOption struct {
	Tags          map[string][]string `json:"-"`
	Short         string              `json:"short,omitempty"`
	Long          string              `json:"long,omitempty"`
	Env           string              `json:"env,omitempty"`
	ValueName     string              `json:"value_name,omitempty"`
	OptionalVal   string              `json:"optional_value,omitempty"`
	Default       string              `json:"-"`
	Description   string              `json:"description,omitempty"`
	Signature     string              `json:"-"`
	EnvDelim      string              `json:"env_delimiter,omitempty"`
	EnvSignature  string              `json:"-"`
	IniName       string              `json:"ini_name,omitempty"`
	DefaultMask   string              `json:"-"`
	KeyValueDelim string              `json:"key_value_delimiter,omitempty"`
	Terminator    string              `json:"terminator,omitempty"`
	Base          string              `json:"base,omitempty"`
	AutoEnvTag    string              `json:"-"`
	UnquoteTag    string              `json:"-"`
	Deprecated    string              `json:"deprecated,omitempty"`
	Choices       []string            `json:"choices,omitempty"`
	DefaultRaw    []string            `json:"default_raw,omitempty"`
	Order         int                 `json:"-"`
	TypeClass     OptionTypeClass     `json:"-"`
	Hidden        bool                `json:"hidden,omitempty"`
	NoIni         bool                `json:"no_ini,omitempty"`
	NoFlag        bool                `json:"no_flag,omitempty"`
	Optional      bool                `json:"optional,omitempty"`
	Required      bool                `json:"required,omitempty"`
	Secret        bool                `json:"secret,omitempty"`
}

func (p *Parser) buildDocModel(cfg docRenderOptions) docParser {
	format := p.optionRenderFormat()
	if cfg.hasRenderStyle {
		format = p.optionRenderFormatForStyles(cfg.renderStyle, cfg.renderStyle)
	}
	usage := p.Usage
	if usage == "" {
		usage = "[OPTIONS]"
	}
	programName := p.Name
	if cfg.programName != "" {
		programName = cfg.programName
	}

	model := docParser{
		Name:             programName,
		ShortDescription: docDescriptionText(p.localizedShortDescription(), programName, cfg.trimDescriptions),
		LongDescription:  docDescriptionText(p.localizedLongDescription(), programName, cfg.trimDescriptions),
		GeneratedAt:      docNow(),
		Usage:            usage,
		Args:             buildDocArgs(p.Command, programName, cfg.includeHidden, cfg.trimDescriptions),
		Groups:           buildDocGroups(p.Group, programName, true, cfg.includeHidden, format, cfg.trimDescriptions, false),
		Meta:             p.buildDocMeta(),
	}

	skipBuiltinHelpInSubs := !cfg.includeBuiltinHelpInSubcommands

	commands := docCommands(p.Command, cfg.includeHidden)
	if cfg.includeBuiltinCommands != nil {
		filtered := make([]*Command, 0, len(commands))
		for _, cmd := range commands {
			if !isBuiltinCommandData(cmd.Data()) {
				filtered = append(filtered, cmd)
				continue
			}
			if slices.Contains(cfg.includeBuiltinCommands, cmd.Name) {
				filtered = append(filtered, cmd)
			}
		}
		commands = filtered
	}
	for _, cmd := range commands {
		model.Commands = append(model.Commands, buildDocCommand("", programName+" "+usage, programName, cmd, cfg.includeHidden, format, cfg.trimDescriptions, skipBuiltinHelpInSubs))
	}
	model.CommandGroups = buildDocCommandGroups(model.Commands)

	return model
}

func buildDocCommand(
	parentName string,
	usagePrefix string,
	programName string,
	cmd *Command,
	includeHidden bool,
	format optionRenderFormat,
	trimDescriptions bool,
	skipBuiltinHelpGroup bool,
) docCommand {
	fullName := cmd.Name
	if parentName != "" {
		fullName = parentName + " " + cmd.Name
	}

	usage := ""
	if us, ok := cmd.data.(Usage); ok {
		usage = us.Usage()
	} else if cmd.hasHelpOptions() {
		usage = fmt.Sprintf("[%s-OPTIONS]", cmd.Name)
	}

	usageLine := usagePrefix + " " + cmd.Name
	nextPrefix := usageLine
	if usage != "" {
		usageLine = usageLine + " " + usage
		nextPrefix = usageLine
	}

	doc := docCommand{
		Name:                fullName,
		ShortDescription:    docDescriptionText(cmd.localizedShortDescription(), programName, trimDescriptions),
		LongDescription:     docDescriptionText(cmd.localizedLongDescription(), programName, trimDescriptions),
		UsageLine:           usageLine,
		Aliases:             append([]string(nil), cmd.Aliases...),
		SubcommandsOptional: cmd.SubcommandsOptional,
		PassAfterNonOption:  cmd.PassAfterNonOption,
		Hidden:              cmd.Hidden,
		Deprecated:          cmd.Deprecated,
		Group:               cmd.localizedCommandGroup(),
		Args:                buildDocArgs(cmd, programName, includeHidden, trimDescriptions),
		Groups:              buildDocGroups(cmd.Group, programName, true, includeHidden, format, trimDescriptions, skipBuiltinHelpGroup),
	}

	for _, sub := range docCommands(cmd, includeHidden) {
		doc.Commands = append(doc.Commands, buildDocCommand(fullName, nextPrefix, programName, sub, includeHidden, format, trimDescriptions, skipBuiltinHelpGroup))
	}
	doc.CommandGroups = buildDocCommandGroups(doc.Commands)

	return doc
}

func buildDocGroups(
	root *Group,
	programName string,
	includeRoot bool,
	includeHidden bool,
	format optionRenderFormat,
	trimDescriptions bool,
	skipBuiltinHelp bool,
) []docGroup {
	var groups []docGroup

	root.eachGroup(func(group *Group) {
		if skipBuiltinHelp && group.isBuiltinHelp {
			return
		}
		if !includeHidden && !group.showInHelp() {
			return
		}
		if !includeRoot && group == root {
			return
		}

		docGroup := docGroup{
			ShortDescription: docDescriptionText(group.localizedShortDescription(), programName, trimDescriptions),
			LongDescription:  docDescriptionText(group.localizedLongDescription(), programName, trimDescriptions),
			Namespace:        group.Namespace,
			EnvNamespace:     group.EnvNamespace,
			Hidden:           group.Hidden,
		}

		for _, opt := range group.sortedOptionsForDisplay() {
			if opt.ShortName == 0 && len(opt.LongName) == 0 {
				continue
			}
			if !includeHidden && !opt.showInHelp() {
				continue
			}
			docGroup.Options = append(docGroup.Options, buildDocOption(opt, programName, format, trimDescriptions))
		}

		if len(docGroup.Options) > 0 {
			groups = append(groups, docGroup)
		}
	})

	return groups
}

func buildDocOption(opt *Option, programName string, format optionRenderFormat, trimDescriptions bool) docOption {
	doc := docOption{
		Long:          opt.LongNameWithNamespace(),
		ValueName:     opt.localizedValueName(),
		Optional:      opt.OptionalArgument,
		Required:      opt.Required,
		Description:   docDescriptionText(opt.localizedDescription(), programName, trimDescriptions),
		TypeClass:     optionTypeClass(opt),
		Choices:       opt.displayChoices(),
		DefaultRaw:    opt.displayValues(opt.Default),
		EnvDelim:      opt.EnvDefaultDelim,
		IniName:       opt.tag.Get(FlagTagIniName),
		DefaultMask:   opt.displayDefaultMask(),
		KeyValueDelim: opt.tag.Get(FlagTagKeyValueDelimiter),
		Terminator:    opt.Terminator,
		Base:          opt.tag.Get(FlagTagBase),
		AutoEnvTag:    opt.tag.Get(FlagTagAutoEnv),
		UnquoteTag:    opt.tag.Get(FlagTagUnquote),
		Tags:          copyDocOptionTags(opt),
		Order:         opt.Order,
		Hidden:        opt.Hidden,
		NoIni:         parseDocBoolTag(opt.tag.Get(FlagTagNoIni)),
		NoFlag:        parseDocBoolTag(opt.tag.Get(FlagTagNoFlag)),
		Secret:        opt.Secret,
		Deprecated:    opt.Deprecated,
	}

	if opt.ShortName != 0 {
		doc.Short = string(opt.ShortName)
	}

	if len(opt.OptionalValue) > 0 {
		doc.OptionalVal = opt.displayValueList(opt.OptionalValue)
	}

	if env := opt.EnvKeyWithNamespace(); env != "" {
		doc.Env = env
		doc.EnvSignature = format.envPrefix + env + format.envSuffix
	}

	if len(opt.Default) > 0 {
		doc.Default = opt.displayValueList(opt.Default)
	}

	doc.Signature = optionSignature(opt, format)
	return doc
}

func buildDocArgs(cmd *Command, programName string, includeHidden bool, trimDescriptions bool) []docArg {
	args := cmd.Args()
	ret := make([]docArg, 0, len(args))
	for _, arg := range args {
		argDescription := docDescriptionText(arg.localizedDescription(), programName, trimDescriptions)
		if !includeHidden && argDescription == "" {
			continue
		}
		required := arg.Required != -1 || arg.RequiredMaximum != -1
		ret = append(ret, docArg{
			Name:        arg.localizedName(),
			Description: argDescription,
			Required:    required,
		})
	}
	return ret
}

func docDescriptionText(text, programName string, trim bool) string {
	text = replaceDocProgramNamePlaceholder(text, programName)
	if !trim {
		return text
	}
	if text == "" {
		return text
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for idx := range lines {
		lines[idx] = strings.TrimSpace(lines[idx])
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func replaceDocProgramNamePlaceholder(text, programName string) string {
	return strings.ReplaceAll(text, DocProgramNamePlaceholder, programName)
}

func buildDocCommandGroups(commands []docCommand) []docCommandGroup {
	groups := make([]docCommandGroup, 0)
	index := make(map[string]int)

	for _, command := range commands {
		idx, ok := index[command.Group]
		if !ok {
			idx = len(groups)
			index[command.Group] = idx
			groups = append(groups, docCommandGroup{Name: command.Group})
		}
		groups[idx].Commands = append(groups[idx].Commands, command)
	}

	return groups
}

func parseDocBoolTag(raw string) bool {
	v, _, err := parseBoolTagValue(raw)
	return err == nil && v
}

func copyDocTags(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}

	out := make(map[string][]string, len(in))
	for k, v := range in {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

func copyDocOptionTags(opt *Option) map[string][]string {
	tags := copyDocTags(opt.tag.cached())
	if !opt.Secret || len(tags) == 0 {
		return tags
	}

	for _, key := range []string{
		FlagTagDefault,
		FlagTagDefaults,
		FlagTagDefaultMask,
		FlagTagOptionalValue,
		FlagTagChoice,
		FlagTagChoices,
	} {
		if len(tags[key]) > 0 {
			tags[key] = []string{secretValueMask}
		}
	}

	return tags
}

func docCommands(c *Command, includeHidden bool) []*Command {
	if includeHidden {
		ret := make([]*Command, len(c.commands))
		copy(ret, c.commands)
		if p := c.parser(); p != nil {
			if p.shouldSortCommandsForDisplay(ret) {
				sort.SliceStable(ret, func(i, j int) bool {
					return p.compareCommands(ret[i], ret[j]) < 0
				})
			}
		} else {
			sort.Slice(ret, func(i, j int) bool {
				return ret[i].Name < ret[j].Name
			})
		}
		return ret
	}
	return c.sortedVisibleCommands()
}

func optionSignature(opt *Option, format optionRenderFormat) string {
	var b strings.Builder

	if opt.ShortName != 0 {
		b.WriteRune(format.shortDelimiter)
		b.WriteRune(opt.ShortName)
	}

	if opt.LongName != "" {
		if opt.ShortName != 0 {
			b.WriteString(", ")
		}
		b.WriteString(format.longDelimiter)
		b.WriteString(opt.LongNameWithNamespace())
	}

	valueName := opt.localizedValueName()
	if len(valueName) != 0 || opt.OptionalArgument {
		if opt.OptionalArgument {
			optionalText := strings.Join(quoteV(opt.OptionalValue), ", ")
			if opt.Secret {
				optionalText = secretValueMask
			}
			fmt.Fprintf(&b, " [%s=%s]", valueName, optionalText)
		} else {
			fmt.Fprintf(&b, " %s", valueName)
		}
	}

	return b.String()
}

// SetSeeAlso sets the SEE ALSO cross-reference entries rendered in all
// documentation formats. In man pages each entry should follow the
// man page convention, e.g. "grep(1)".
func (p *Parser) SetSeeAlso(refs ...string) {
	p.seeAlso = refs
}

// buildDocMeta returns a *docParserMeta populated from parser configuration
// and VersionInfo. Returns nil when nothing is explicitly set so the field
// is omitted from JSON output.
func (p *Parser) buildDocMeta() *docParserMeta {
	ev := p.versionInfo
	// Use explicit overrides only for the nil check so that auto-detected
	// module URL from debug.ReadBuildInfo does not trigger output.
	if p.manConfig.Section == 0 && len(p.seeAlso) == 0 &&
		ev.Author == "" && ev.BugsURL == "" && ev.License == "" && ev.URL == "" {
		return nil
	}

	vi := p.VersionInfo() // merged (auto-detected + explicit overrides)
	section := p.manConfig.Section
	if section == 0 {
		section = 1
	}

	return &docParserMeta{
		Section:  section,
		Author:   vi.Author,
		Homepage: vi.URL,
		BugsURL:  vi.BugsURL,
		SeeAlso:  p.seeAlso,
		License:  vi.License,
	}
}

// docNow returns the timestamp used for generated documentation,
// honoring SOURCE_DATE_EPOCH for reproducible builds when it is set to a valid Unix timestamp.
// An invalid value is ignored (falls back to the current time)
// rather than aborting doc generation:
// a malformed environment variable should not crash the process.
func docNow() time.Time {
	t := time.Now()
	sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH")
	if sourceDateEpoch == "" {
		return t
	}

	sde, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
	if err != nil {
		return t
	}

	return time.Unix(sde, 0).UTC()
}
