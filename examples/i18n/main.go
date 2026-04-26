// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

// Package main demonstrates full i18n wiring for flags and app messages.
package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/woozymasta/flags"
)

//go:embed catalog/*.json
var catalogFS embed.FS

type Options struct {
	// i18n tags point to catalog keys. The fallback text lives in code paths
	// that call Localize, while parser metadata resolves through SetI18n.
	Greet struct {
		Excited bool `short:"e" long:"excited" description-i18n:"opt.excited.desc"`
		Args    struct {
			Target string `arg-name-i18n:"arg.target.name" arg-description-i18n:"arg.target.desc" default:"world"`
		} `positional-args:"yes"`
	} `command:"greet" command-i18n:"cmd.greet.desc" long-description-i18n:"cmd.greet.long" ini-group:"greet"`

	Locale string `short:"l" long:"locale" choices:"en;ru;eo" description-i18n:"opt.locale.desc" value-name-i18n:"opt.locale.value"`

	Display struct {
		Style string `long:"style" default:"plain" choices:"plain;fancy" description-i18n:"opt.style.desc"`
	} `group:"Display" group-i18n:"group.display" long-description-i18n:"group.display.long" ini-group:"display"`
	Verbose bool `short:"V" long:"verbose" description-i18n:"opt.verbose.desc"`
}

func main() {
	var opts Options
	localeOverride := detectLocaleArg(os.Args[1:])

	parser := flags.NewNamedParser("i18n-demo", flags.Default|flags.VersionFlag|flags.HelpCommands)
	parser.SetLongDescriptionI18nKey("app.description")

	// Groups can also be localized after registration. This keeps ordinary
	// library defaults readable while still allowing application-specific text.
	group, err := parser.AddGroup("Application Options", "", &opts)
	if err != nil {
		os.Exit(1)
	}
	group.SetShortDescriptionI18nKey("help.group.application_options")

	userCatalog, err := flags.NewJSONCatalogDirFS(catalogFS, "catalog")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load i18n catalog: %v\n", err)
		os.Exit(1)
	}

	i18nConfig := flags.I18nConfig{
		Locale:          localeOverride,
		FallbackLocales: []string{"en"},
		UserCatalog:     userCatalog,
	}
	// The parser uses this config for help/errors/docs/INI output. The
	// standalone Localizer below uses the same config for application messages.
	parser.SetI18n(i18nConfig)
	localizer := flags.NewLocalizer(i18nConfig)

	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && (flagsErr.Type == flags.ErrHelp || flagsErr.Type == flags.ErrVersion) {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if parser.Active == nil {
		return
	}

	switch parser.Active.Name {
	case "greet":
		runGreet(localizer, &opts)
	}
}

func runGreet(localizer *flags.Localizer, opts *Options) {
	target := opts.Greet.Args.Target
	if target == "" {
		target = "world"
	}

	line := localizer.Localize(
		// Localize supports placeholder substitution for application text that
		// is not part of parser metadata.
		"app.greeting",
		"Hello, {target}",
		map[string]string{"target": target},
	)

	if opts.Greet.Excited {
		line += "!"
	}

	if opts.Display.Style == "fancy" {
		line = ">>> " + line + " <<<"
	}

	if opts.Verbose {
		activeLocale := "en"
		if chain := localizer.LocaleChain(); len(chain) > 0 {
			activeLocale = chain[0]
		}
		_, _ = fmt.Printf("[locale=%s style=%s]\n", activeLocale, opts.Display.Style)
	}

	_, _ = fmt.Fprintln(os.Stdout, line)
}

func detectLocaleArg(args []string) string {
	// Locale must be known before parser.Parse so parse errors and --help can
	// be localized on the first render.
	for idx := range args {
		arg := args[idx]

		if after, ok := strings.CutPrefix(arg, "--locale="); ok {
			value := strings.TrimSpace(after)
			if value != "" {
				return value
			}
			continue
		}

		if arg == "--locale" || arg == "-l" {
			if idx+1 < len(args) {
				value := strings.TrimSpace(args[idx+1])
				if value != "" && !strings.HasPrefix(value, "-") {
					return value
				}
			}
		}
	}

	return ""
}
