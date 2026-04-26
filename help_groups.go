// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"sort"
)

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

type helpArgGroup struct {
	name     string
	args     []*Arg
	sortRank int
}

type helpCommandGroup struct {
	name     string
	commands []*Command
	sortRank int
}

func groupedHelpArgs(p *Parser, args []*Arg) []helpArgGroup {
	groups := make([]helpArgGroup, 0)
	index := make(map[string]int)

	hasNamedGroups := false
	for _, arg := range args {
		if arg.Group != "" {
			hasNamedGroups = true
			break
		}
	}

	for _, arg := range args {
		name := arg.Group
		rank := 1
		if hasNamedGroups && name == "" {
			name = p.i18nText("help.arg_group.main_arguments", "Main Arguments")
			rank = 0
		}

		key := fmt.Sprintf("%d\x00%s", rank, name)
		idx, ok := index[key]
		if !ok {
			idx = len(groups)
			index[key] = idx
			groups = append(groups, helpArgGroup{name: name, sortRank: rank})
		}
		groups[idx].args = append(groups[idx].args, arg)
	}

	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].sortRank != groups[j].sortRank {
			return groups[i].sortRank < groups[j].sortRank
		}
		return groups[i].name < groups[j].name
	})

	return groups
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
