// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

func (option *Option) localizedDescription() string {
	if option == nil {
		return ""
	}

	if option.DescriptionI18nKey == "" {
		return option.Description
	}

	if p := option.parser(); p != nil {
		return p.i18nText(option.DescriptionI18nKey, option.Description)
	}

	return option.Description
}

func (option *Option) localizedValueName() string {
	if option == nil {
		return ""
	}

	if option.ValueNameI18nKey == "" {
		return option.ValueName
	}

	if p := option.parser(); p != nil {
		return p.i18nText(option.ValueNameI18nKey, option.ValueName)
	}

	return option.ValueName
}

func (group *Group) localizedShortDescription() string {
	if group == nil {
		return ""
	}

	if group.ShortDescriptionI18nKey == "" {
		return group.ShortDescription
	}

	if p := group.parser(); p != nil {
		return p.i18nText(group.ShortDescriptionI18nKey, group.ShortDescription)
	}

	return group.ShortDescription
}

func (group *Group) localizedLongDescription() string {
	if group == nil {
		return ""
	}

	if group.LongDescriptionI18nKey == "" {
		return group.LongDescription
	}

	if p := group.parser(); p != nil {
		return p.i18nText(group.LongDescriptionI18nKey, group.LongDescription)
	}

	return group.LongDescription
}

func (cmd *Command) localizedShortDescription() string {
	if cmd == nil {
		return ""
	}

	return cmd.Group.localizedShortDescription()
}

func (cmd *Command) localizedLongDescription() string {
	if cmd == nil {
		return ""
	}

	return cmd.Group.localizedLongDescription()
}

func (arg *Arg) localizedName() string {
	if arg == nil {
		return ""
	}

	if arg.NameI18nKey == "" {
		return arg.Name
	}

	if arg.cmd != nil {
		if p := arg.cmd.parser(); p != nil {
			return p.i18nText(arg.NameI18nKey, arg.Name)
		}
	}

	return arg.Name
}

func (arg *Arg) localizedDescription() string {
	if arg == nil {
		return ""
	}

	if arg.DescriptionI18nKey == "" {
		return arg.Description
	}

	if arg.cmd != nil {
		if p := arg.cmd.parser(); p != nil {
			return p.i18nText(arg.DescriptionI18nKey, arg.Description)
		}
	}

	return arg.Description
}
