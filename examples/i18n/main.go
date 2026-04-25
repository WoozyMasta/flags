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

	Doc struct {
		Format        string `long:"format" default:"markdown" choices:"markdown;html;man" description-i18n:"opt.doc.format.desc" value-name-i18n:"opt.doc.format.value"`
		Template      string `long:"template" choices:"list;table;code;default;styled" description-i18n:"opt.doc.template.desc" value-name-i18n:"opt.doc.template.value"`
		IncludeHidden bool   `long:"include-hidden" description-i18n:"opt.doc.include_hidden.desc"`
		MarkHidden    bool   `long:"mark-hidden" description-i18n:"opt.doc.mark_hidden.desc"`
	} `command:"doc" command-i18n:"cmd.doc.desc" long-description-i18n:"cmd.doc.long" ini-group:"doc"`

	INI struct {
		Mode            string `long:"mode" default:"example" choices:"example;current" description-i18n:"opt.ini.mode.desc" value-name-i18n:"opt.ini.mode.value"`
		CommentWidth    int    `long:"comment-width" default:"88" description-i18n:"opt.ini.comment_width.desc" value-name-i18n:"opt.ini.comment_width.value"`
		IncludeDefaults bool   `long:"include-defaults" description-i18n:"opt.ini.include_defaults.desc"`
		IncludeComments bool   `long:"include-comments" description-i18n:"opt.ini.include_comments.desc"`
		CommentDefaults bool   `long:"comment-defaults" description-i18n:"opt.ini.comment_defaults.desc"`
	} `command:"ini" command-i18n:"cmd.ini.desc" long-description-i18n:"cmd.ini.long" ini-group:"ini"`
	Verbose bool `short:"v" long:"verbose" description-i18n:"opt.verbose.desc"`
}

func main() {
	var opts Options
	localeOverride := detectLocaleArg(os.Args[1:])

	parser := flags.NewNamedParser("i18n-demo", flags.Default|flags.VersionFlag)
	parser.SetLongDescriptionI18nKey("app.description")

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
	case "doc":
		if err := runDoc(parser, localizer, &opts); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "ini":
		if err := runINI(parser, localizer, &opts); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func runGreet(localizer *flags.Localizer, opts *Options) {
	target := opts.Greet.Args.Target
	if target == "" {
		target = "world"
	}

	line := localizer.Localize(
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

func runDoc(parser *flags.Parser, localizer *flags.Localizer, opts *Options) error {
	format, templateName, err := resolveDocMode(localizer, opts.Doc.Format, opts.Doc.Template)
	if err != nil {
		return err
	}

	var renderOpts []flags.DocOption
	renderOpts = append(renderOpts, flags.WithBuiltinTemplate(templateName))

	if opts.Doc.IncludeHidden {
		renderOpts = append(renderOpts, flags.WithIncludeHidden(true))
	}

	if opts.Doc.MarkHidden {
		renderOpts = append(renderOpts, flags.WithMarkHidden(true))
	}

	return parser.WriteDoc(os.Stdout, format, renderOpts...)
}

func runINI(parser *flags.Parser, localizer *flags.Localizer, opts *Options) error {
	ini := flags.NewIniParser(parser)

	switch opts.INI.Mode {
	case "example":
		ini.WriteExampleWithOptions(os.Stdout, flags.IniExampleOptions{
			CommentWidth: opts.INI.CommentWidth,
		})
		return nil
	case "current":
		mask := flags.IniNone
		if opts.INI.IncludeComments {
			mask |= flags.IniIncludeComments
		}
		if opts.INI.IncludeDefaults {
			mask |= flags.IniIncludeDefaults
		}
		if opts.INI.CommentDefaults {
			mask |= flags.IniCommentDefaults
		}

		ini.Write(os.Stdout, mask)
		return nil
	default:
		return fmt.Errorf(
			"%s: %s",
			localizer.Localize("app.error.invalid_ini_mode", "invalid ini mode", nil),
			opts.INI.Mode,
		)
	}
}

func resolveDocMode(localizer *flags.Localizer, format string, template string) (flags.DocFormat, string, error) {
	switch format {
	case "markdown":
		switch template {
		case "", "list":
			return flags.DocFormatMarkdown, flags.DocTemplateMarkdownList, nil
		case "table":
			return flags.DocFormatMarkdown, flags.DocTemplateMarkdownTable, nil
		case "code":
			return flags.DocFormatMarkdown, flags.DocTemplateMarkdownCode, nil
		default:
			return "", "", fmt.Errorf(
				"%s: %s",
				localizer.Localize("app.error.invalid_markdown_template", "invalid markdown template", nil),
				template,
			)
		}
	case "html":
		switch template {
		case "", "default":
			return flags.DocFormatHTML, flags.DocTemplateHTMLDefault, nil
		case "styled":
			return flags.DocFormatHTML, flags.DocTemplateHTMLStyled, nil
		default:
			return "", "", fmt.Errorf(
				"%s: %s",
				localizer.Localize("app.error.invalid_html_template", "invalid html template", nil),
				template,
			)
		}
	case "man":
		switch template {
		case "", "default":
			return flags.DocFormatMan, flags.DocTemplateManDefault, nil
		default:
			return "", "", fmt.Errorf(
				"%s: %s",
				localizer.Localize("app.error.invalid_man_template", "invalid man template", nil),
				template,
			)
		}
	default:
		return "", "", fmt.Errorf(
			"%s: %s",
			localizer.Localize("app.error.invalid_doc_format", "invalid doc format", nil),
			format,
		)
	}
}

func detectLocaleArg(args []string) string {
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
