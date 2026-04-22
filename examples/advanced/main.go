// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

// Package main demonstrates advanced parser features.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/woozymasta/flags"
)

type DynamicToken string

func (d *DynamicToken) Default() ([]string, error) {
	return []string{"token-from-provider"}, nil
}

type ServiceLabel string

func (l *ServiceLabel) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return errors.New("service label cannot be empty")
	}
	*l = ServiceLabel(strings.ToLower(string(text)))
	return nil
}

func (l ServiceLabel) MarshalText() ([]byte, error) {
	if l == "" {
		return nil, errors.New("service label cannot be empty")
	}
	return []byte(strings.ToUpper(string(l))), nil
}

type AdvancedOptions struct {
	Positional struct {
		Target   string `positional-arg-name:"target" description:"Target service name or host"`
		Artifact string `positional-arg-name:"artifact" description:"Artifact path or reference"`
	} `positional-args:"yes" required:"yes"`

	Alpha    string       `long:"alpha" description:"Example string flag for sort demo" default:"a"`
	Profile  string       `long:"profile" description:"Runtime profile" default:"dev" auto-env:"true"`
	Region   string       `long:"region" description:"Cloud region" env:"APP_REGION" default:"eu-west-1"`
	Token    DynamicToken `long:"token" description:"Dynamic default token"`
	Strategy string       `long:"deployment-strategy-with-very-long-name" value-name:"STRATEGY_PROFILE_NAME" description:"Deployment strategy selector with long value" default:"rolling-update-with-pre-drain-and-post-verify"`

	ManualEnvOnly string `long:"manual-env-only" description:"Explicit opt-out from global auto env" auto-env:"false" default:"local" order:"-40"`
	ReleaseID     string `long:"release-id" value-name:"RELEASE_IDENTIFIER" description:"Release identifier for audit trail" required:"yes"`
	SecretKey     string `long:"secret-key" description:"Hidden secret key for debugging deployments" hidden:"yes"`

	HelpColor string `long:"help-color" choices:"none;default;contrast;light" default:"none" description:"Color scheme for built-in help output"`

	Verbose []bool `short:"v" long:"verbose" description:"Increase verbosity level" order:"100"`

	Labels []ServiceLabel `long:"label" description:"Service labels"`
	Exec   []string       `long:"exec" description:"Collect args until ';' terminator" terminator:";" order:"-30"`

	Network struct {
		Endpoint string        `long:"endpoint" description:"Service endpoint" auto-env:"true"`
		Mode     string        `long:"mode" description:"Network mode"`
		Timeout  time.Duration `long:"timeout" description:"Request timeout" default:"10s"`
		Retries  int           `long:"retries" description:"Retry attempts" default:"3"`
		TLS      bool          `long:"tls" description:"Enable TLS" order:"50"`
	} `group:"Network Options" namespace:"net" env-namespace:"NET"`

	Count int           `long:"count" description:"Example number flag for sort demo" default:"7"`
	Delay time.Duration `long:"delay" description:"Example duration flag for sort demo" default:"2s"`

	Deploy struct {
		Force bool `long:"force" description:"Force deployment"`
		Plan  bool `long:"plan" description:"Show execution plan only"`
	} `command:"deploy" description:"Deploy selected targets" long-description:"Run deployment workflow with validation checks.\n\nExamples:\n  advanced-cli deploy --force target artifact\n  advanced-cli deploy --plan target artifact" pass-after-non-option:"yes"`
	Zeta bool `long:"zeta" description:"Example bool flag for sort demo"`
}

func newParser(opts *AdvancedOptions) *flags.Parser {
	p := flags.NewNamedParser("advanced-cli", flags.Default|flags.EnvProvisioning|flags.KeepDescriptionWhitespace)
	p.LongDescription = "Example of advanced go-flags features:\n  - dynamic defaults\n  - env provisioning and auto-env\n  - terminated options\n  - option sorting per group block"
	p.SetEnvPrefix("DEMO_APP")

	_, err := p.AddGroup("Application Options", "Advanced feature demo", opts)
	if err != nil {
		panic(err)
	}

	return p
}

func applySortMode(p *flags.Parser, mode string) error {
	switch mode {
	case "decl":
		p.SetOptionSort(flags.OptionSortByDeclaration)
	case "name-asc":
		p.SetOptionSort(flags.OptionSortByNameAsc)
	case "name-desc":
		p.SetOptionSort(flags.OptionSortByNameDesc)
	case "type":
		p.SetOptionSort(flags.OptionSortByType)
		return p.SetOptionTypeOrder([]flags.OptionTypeClass{
			flags.OptionTypeString,
			flags.OptionTypeBool,
			flags.OptionTypeDuration,
			flags.OptionTypeNumber,
			flags.OptionTypeCollection,
			flags.OptionTypeCustom,
		})
	default:
		return fmt.Errorf("unknown sort mode %q", mode)
	}

	return nil
}

func applyHelpColorMode(p *flags.Parser, mode string) error {
	switch mode {
	case "", "none":
		p.Options &^= flags.ColorHelp
	case "default":
		p.Options |= flags.ColorHelp
		p.SetHelpColorScheme(flags.DefaultHelpColorScheme())
	case "contrast":
		p.Options |= flags.ColorHelp
		p.SetHelpColorScheme(flags.HighContrastHelpColorScheme())
	case "light":
		p.Options |= flags.ColorHelp
		p.SetHelpColorScheme(flags.HelpColorScheme{
			BaseText:        flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack, UseBG: true, BG: flags.ColorBrightWhite},
			OptionShort:     flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			OptionLong:      flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			OptionDesc:      flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack},
			OptionEnv:       flags.HelpTextStyle{UseFG: true, FG: flags.ColorCyan},
			OptionDefault:   flags.HelpTextStyle{UseFG: true, FG: flags.ColorMagenta},
			OptionChoices:   flags.HelpTextStyle{UseFG: true, FG: flags.ColorGreen, Bold: true},
			UsageHeader:     flags.HelpTextStyle{UseFG: true, FG: flags.ColorRed, Bold: true},
			UsageText:       flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack, Bold: true},
			CommandsHeader:  flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlack, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			CommandName:     flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			CommandDesc:     flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlack, UseBG: true, BG: flags.ColorBrightYellow},
			ArgumentsHeader: flags.HelpTextStyle{UseFG: true, FG: flags.ColorRed, Bold: true},
			ArgumentName:    flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, Bold: true},
			ArgumentDesc:    flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack},
			GroupHeader:     flags.HelpTextStyle{UseFG: true, FG: flags.ColorRed, Bold: true, Underline: true},
		})
	default:
		return fmt.Errorf("unknown help color mode %q", mode)
	}

	return nil
}

func detectHelpColorArg(args []string) (string, bool) {
	for i, arg := range args {
		if v, ok := strings.CutPrefix(arg, "--help-color="); ok {
			return v, true
		}

		if arg == "--help-color" && i+1 < len(args) {
			return args[i+1], true
		}
	}

	return "", false
}

func demoOutput(args []string, p *flags.Parser) (bool, error) {
	docFormat := ""
	docStyle := ""

	if mode, ok := detectHelpColorArg(args); ok {
		if err := applyHelpColorMode(p, mode); err != nil {
			return true, err
		}
	}

	for _, arg := range args {
		if mode, ok := strings.CutPrefix(arg, "--demo-help="); ok {
			if err := applySortMode(p, mode); err != nil {
				return true, err
			}
			p.WriteHelp(os.Stdout)
			return true, nil
		}

		if shell, ok := strings.CutPrefix(arg, "--demo-completion="); ok {
			if err := p.WriteNamedCompletion(os.Stdout, flags.CompletionShell(shell), "advanced-cli"); err != nil {
				return true, err
			}
			return true, nil
		}

		if v, ok := strings.CutPrefix(arg, "--demo-doc-format="); ok {
			docFormat = v
			continue
		}

		if v, ok := strings.CutPrefix(arg, "--demo-doc-style="); ok {
			docStyle = v
			continue
		}
	}

	if docFormat != "" {
		format, tmpl, err := resolveDocMode(docFormat, docStyle)
		if err != nil {
			return true, err
		}
		if err := p.WriteDoc(os.Stdout, format, flags.WithBuiltinTemplate(tmpl)); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}

func resolveDocMode(format, style string) (flags.DocFormat, string, error) {
	switch format {
	case "html":
		return flags.DocFormatHTML, flags.DocTemplateHTMLDefault, nil
	case "man":
		return flags.DocFormatMan, flags.DocTemplateManDefault, nil
	case "markdown":
		switch style {
		case "", "list":
			return flags.DocFormatMarkdown, flags.DocTemplateMarkdownList, nil
		case "table":
			return flags.DocFormatMarkdown, flags.DocTemplateMarkdownTable, nil
		case "code":
			return flags.DocFormatMarkdown, flags.DocTemplateMarkdownCode, nil
		default:
			return "", "", fmt.Errorf("unknown markdown style %q", style)
		}
	default:
		return "", "", fmt.Errorf("unknown doc format %q", format)
	}
}

func main() {
	opts := &AdvancedOptions{}
	p := newParser(opts)
	if mode, ok := detectHelpColorArg(os.Args[1:]); ok {
		if err := applyHelpColorMode(p, mode); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	handled, err := demoOutput(os.Args[1:], p)
	if handled {
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if _, err := p.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
