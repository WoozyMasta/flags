// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build !forceposix

package flags

import (
	"strings"
)

// Windows uses a front slash for both short and long options.  Also it uses
// a colon for name/argument delimter.
const (
	defaultShortOptDelimiter = '/'
	defaultLongOptDelimiter  = "/"
	defaultNameArgDelimiter  = ':'
)

func argumentStartsOption(arg string) bool {
	return len(arg) > 0 && (arg[0] == '-' || arg[0] == '/')
}

func argumentIsOption(arg string) bool {
	// Windows-style options allow front slash for the option
	// delimiter.
	if len(arg) > 1 && arg[0] == '/' {
		return true
	}

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
	// Determine if the argument is a long option or not.  Windows
	// typically supports both long and short options with a single
	// front slash as the option delimiter, so handle this situation
	// nicely.
	possplit := 0

	switch {
	case strings.HasPrefix(optname, "--"):
		possplit = 2
		islong = true

	case strings.HasPrefix(optname, "-"):
		possplit = 1
		islong = false

	case strings.HasPrefix(optname, "/"):
		possplit = 1
		islong = len(optname) > 2
	}

	return optname[:possplit], optname[possplit:], islong
}

// splitOption attempts to split the passed option into a name and an argument.
// When there is no argument specified, hasArgument will be false.
func splitOption(prefix string, option string, islong bool) (name string, split string, argument string, hasArgument bool) {
	if len(option) == 0 {
		return option, "", "", false
	}

	// Windows typically uses a colon for the option name and argument
	// delimiter while POSIX typically uses an equals.  Support both styles,
	// but don't allow the two to be mixed.  That is to say /foo:bar and
	// --foo=bar are acceptable, but /foo=bar and --foo:bar are not.
	var pos int
	var sp string

	if prefix == "/" {
		sp = ":"
		pos = strings.Index(option, sp)
	} else if len(prefix) > 0 {
		sp = "="
		pos = strings.Index(option, sp)
	}

	if (islong && pos >= 0) || (!islong && pos == 1) {
		return option[:pos], sp, option[pos+1:], true
	}

	return option, "", "", false
}

// addHelpGroup adds a new group that contains default help/version parameters.
func (c *Command) addHelpGroup(showHelp func() error, showVersion func() error) *Group {
	style := RenderStyleWindows
	if parser := c.parser(); parser != nil {
		style = parser.resolveFlagRenderStyle()
	}

	includeVersion := false
	if p := c.parser(); p != nil && (p.Options&VersionFlag) != None && showVersion != nil {
		includeVersion = true
	}

	if style == RenderStylePOSIX {
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

	// Windows CLI applications typically use /? for help, so make both
	// that available as well as the POSIX style h and help.
	if includeVersion {
		var help struct {
			ShowHelpWindows func() error `short:"?" description:"Show this help message" description-i18n:"help.builtin.show_help" auto-env:"false" immediate:"true"`
			ShowHelpPosix   func() error `short:"h" long:"help" description:"Show this help message" description-i18n:"help.builtin.show_help" auto-env:"false" immediate:"true"`
			ShowVersion     func() error `short:"v" long:"version" description:"Show version information" description-i18n:"help.builtin.show_version" auto-env:"false" immediate:"true"`
		}

		help.ShowHelpWindows = showHelp
		help.ShowHelpPosix = showHelp
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
		ShowHelpWindows func() error `short:"?" description:"Show this help message" description-i18n:"help.builtin.show_help" auto-env:"false" immediate:"true"`
		ShowHelpPosix   func() error `short:"h" long:"help" description:"Show this help message" description-i18n:"help.builtin.show_help" auto-env:"false" immediate:"true"`
	}

	help.ShowHelpWindows = showHelp
	help.ShowHelpPosix = showHelp

	ret, err := c.AddGroup("Help Options", "", &help)
	if err != nil {
		return nil
	}
	ret.SetShortDescriptionI18nKey("help.group.help_options")
	ret.isBuiltinHelp = true

	return ret
}
