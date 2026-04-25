// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build !windows || forceposix

package flags

import (
	"strings"
)

const (
	defaultShortOptDelimiter = '-'
	defaultLongOptDelimiter  = "--"
	defaultNameArgDelimiter  = '='
)

func argumentStartsOption(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

func argumentIsOption(arg string) bool {
	if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' {
		return true
	}

	if len(arg) > 2 && arg[0] == '-' && arg[1] == '-' && arg[2] != '-' {
		return true
	}

	return false
}

// stripOptionPrefix returns the option without the prefix and whether or
// not the option is a long option or not.
func stripOptionPrefix(optname string) (prefix string, name string, islong bool) {
	if strings.HasPrefix(optname, "--") {
		return "--", optname[2:], true
	} else if strings.HasPrefix(optname, "-") {
		return "-", optname[1:], false
	}

	return "", optname, false
}

// splitOption attempts to split the passed option into a name and an argument.
// When there is no argument specified, hasArgument will be false.
func splitOption(_ string, option string, islong bool) (name string, split string, argument string, hasArgument bool) {
	pos := strings.Index(option, "=")

	if (islong && pos >= 0) || (!islong && pos == 1) {
		return option[:pos], "=", option[pos+1:], true
	}

	return option, "", "", false
}

// addHelpGroup adds a new group that contains default help/version parameters.
func (c *Command) addHelpGroup(showHelp func() error, showVersion func() error) *Group {
	includeVersion := false
	if p := c.parser(); p != nil && (p.Options&VersionFlag) != None && showVersion != nil {
		includeVersion = true
	}

	if includeVersion {
		var help struct {
			ShowHelp    func() error `short:"h" long:"help" description:"Show this help message" description-i18n:"help.builtin.show_help" auto-env:"false" immediate:"true"`
			ShowVersion func() error `short:"v" long:"version" description:"Show version information" description-i18n:"help.builtin.show_version" auto-env:"false" immediate:"true"`
		}

		help.ShowHelp = showHelp
		help.ShowVersion = showVersion
		ret, err := c.AddGroup("Help Options", "", &help)
		if err != nil {
			return nil
		}
		ret.SetShortDescriptionI18nKey("help.group.help_options")
		ret.isBuiltinHelp = true

		return ret
	}

	var help struct {
		ShowHelp func() error `short:"h" long:"help" description:"Show this help message" description-i18n:"help.builtin.show_help" auto-env:"false" immediate:"true"`
	}

	help.ShowHelp = showHelp
	ret, err := c.AddGroup("Help Options", "", &help)
	if err != nil {
		return nil
	}
	ret.SetShortDescriptionI18nKey("help.group.help_options")
	ret.isBuiltinHelp = true

	return ret
}
