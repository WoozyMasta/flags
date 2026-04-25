// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type docParser struct {
	Name             string
	ShortDescription string
	LongDescription  string
	GeneratedAt      time.Time
	Usage            string
	Args             []docArg
	Groups           []docGroup
	Commands         []docCommand
}

type docCommand struct {
	Name                string
	ShortDescription    string
	LongDescription     string
	UsageLine           string
	Aliases             []string
	Args                []docArg
	Groups              []docGroup
	Commands            []docCommand
	SubcommandsOptional bool
	PassAfterNonOption  bool
	Hidden              bool
}

type docArg struct {
	Name        string
	Description string
	Required    bool
}

type docGroup struct {
	ShortDescription string
	LongDescription  string
	Namespace        string
	EnvNamespace     string
	Options          []docOption
	Hidden           bool
}

type docOption struct {
	Tags          map[string][]string
	Short         string
	Long          string
	Env           string
	ValueName     string
	OptionalVal   string
	Default       string
	Description   string
	Signature     string
	EnvDelim      string
	IniName       string
	DefaultMask   string
	KeyValueDelim string
	Terminator    string
	Base          string
	AutoEnvTag    string
	UnquoteTag    string
	Choices       []string
	DefaultRaw    []string
	Order         int
	TypeClass     OptionTypeClass
	Hidden        bool
	NoIni         bool
	NoFlag        bool
	Optional      bool
	Required      bool
}

func (p *Parser) buildDocModel(cfg docRenderOptions) docParser {
	format := p.optionRenderFormat()
	usage := p.Usage
	if usage == "" {
		usage = "[OPTIONS]"
	}

	model := docParser{
		Name:             p.Name,
		ShortDescription: p.localizedShortDescription(),
		LongDescription:  p.localizedLongDescription(),
		GeneratedAt:      docNow(),
		Usage:            usage,
		Args:             buildDocArgs(p.Command, cfg.includeHidden),
		Groups:           buildDocGroups(p.Group, true, cfg.includeHidden, format),
	}

	for _, cmd := range p.docCommands(cfg.includeHidden) {
		model.Commands = append(model.Commands, buildDocCommand("", p.Name+" "+usage, cmd, cfg.includeHidden, format))
	}

	return model
}

func buildDocCommand(
	parentName string,
	usagePrefix string,
	cmd *Command,
	includeHidden bool,
	format optionRenderFormat,
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
		ShortDescription:    cmd.localizedShortDescription(),
		LongDescription:     cmd.localizedLongDescription(),
		UsageLine:           usageLine,
		Aliases:             append([]string(nil), cmd.Aliases...),
		SubcommandsOptional: cmd.SubcommandsOptional,
		PassAfterNonOption:  cmd.PassAfterNonOption,
		Hidden:              cmd.Hidden,
		Args:                buildDocArgs(cmd, includeHidden),
		Groups:              buildDocGroups(cmd.Group, true, includeHidden, format),
	}

	for _, sub := range docCommands(cmd, includeHidden) {
		doc.Commands = append(doc.Commands, buildDocCommand(fullName, nextPrefix, sub, includeHidden, format))
	}

	return doc
}

func buildDocGroups(
	root *Group,
	includeRoot bool,
	includeHidden bool,
	format optionRenderFormat,
) []docGroup {
	var groups []docGroup

	root.eachGroup(func(group *Group) {
		if !includeHidden && !group.showInHelp() {
			return
		}
		if !includeRoot && group == root {
			return
		}

		docGroup := docGroup{
			ShortDescription: group.localizedShortDescription(),
			LongDescription:  group.localizedLongDescription(),
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
			docGroup.Options = append(docGroup.Options, buildDocOption(opt, format))
		}

		if len(docGroup.Options) > 0 {
			groups = append(groups, docGroup)
		}
	})

	return groups
}

func buildDocOption(opt *Option, format optionRenderFormat) docOption {
	doc := docOption{
		Long:          opt.LongNameWithNamespace(),
		ValueName:     opt.localizedValueName(),
		Optional:      opt.OptionalArgument,
		Required:      opt.Required,
		Description:   opt.localizedDescription(),
		TypeClass:     optionTypeClass(opt),
		Choices:       append([]string(nil), opt.Choices...),
		DefaultRaw:    append([]string(nil), opt.Default...),
		EnvDelim:      opt.EnvDefaultDelim,
		IniName:       opt.tag.Get(FlagTagIniName),
		DefaultMask:   opt.DefaultMask,
		KeyValueDelim: opt.tag.Get(FlagTagKeyValueDelimiter),
		Terminator:    opt.Terminator,
		Base:          opt.tag.Get(FlagTagBase),
		AutoEnvTag:    opt.tag.Get(FlagTagAutoEnv),
		UnquoteTag:    opt.tag.Get(FlagTagUnquote),
		Tags:          copyDocTags(opt.tag.cached()),
		Order:         opt.Order,
		Hidden:        opt.Hidden,
		NoIni:         parseDocBoolTag(opt.tag.Get(FlagTagNoIni)),
		NoFlag:        parseDocBoolTag(opt.tag.Get(FlagTagNoFlag)),
	}

	if opt.ShortName != 0 {
		doc.Short = string(opt.ShortName)
	}

	if len(opt.OptionalValue) > 0 {
		doc.OptionalVal = strings.Join(opt.OptionalValue, ", ")
	}

	if env := opt.EnvKeyWithNamespace(); env != "" {
		doc.Env = env
	}

	if len(opt.Default) > 0 {
		doc.Default = strings.Join(opt.Default, ", ")
	} else if doc.Env != "" {
		doc.Default = format.envPrefix + doc.Env + format.envSuffix
	}

	doc.Signature = optionSignature(opt, format)
	return doc
}

func buildDocArgs(cmd *Command, includeHidden bool) []docArg {
	args := cmd.Args()
	ret := make([]docArg, 0, len(args))
	for _, arg := range args {
		argDescription := arg.localizedDescription()
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

func docCommands(c *Command, includeHidden bool) []*Command {
	if includeHidden {
		ret := make([]*Command, len(c.commands))
		copy(ret, c.commands)
		sort.Sort(commandList(ret))
		return ret
	}
	return c.sortedVisibleCommands()
}

func (p *Parser) docCommands(includeHidden bool) []*Command {
	return docCommands(p.Command, includeHidden)
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
			fmt.Fprintf(&b, " [%s=%s]", valueName, strings.Join(quoteV(opt.OptionalValue), ", "))
		} else {
			fmt.Fprintf(&b, " %s", valueName)
		}
	}

	return b.String()
}

func docNow() time.Time {
	t := time.Now()
	sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH")
	if sourceDateEpoch == "" {
		return t
	}

	sde, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid SOURCE_DATE_EPOCH: %s", err))
	}

	return time.Unix(sde, 0).UTC()
}
