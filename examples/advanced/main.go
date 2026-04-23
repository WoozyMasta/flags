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
	Positional AdvancedPositionalArgs `positional-args:"yes" required:"yes"`

	Alpha            string                 `long:"alpha" description:"Example string flag for sort demo" default:"a"`
	Profile          string                 `long:"profile" description:"Runtime profile" default:"dev" auto-env:"true"`
	Region           string                 `long:"region" description:"Cloud region" env:"APP_REGION" default:"eu-west-1"`
	Token            DynamicToken           `long:"token" description:"Dynamic default token"`
	Strategy         string                 `long:"deployment-strategy-with-very-long-name" value-name:"STRATEGY_PROFILE_NAME" description:"Deployment strategy selector with long value" default:"rolling-update-with-pre-drain-and-post-verify"`
	FormatPolicy     string                 `long:"output-format-negotiation-policy-for-generated-artifacts" value-name:"OUTPUT_FORMAT_NEGOTIATION_POLICY_IDENTIFIER" description:"Output format negotiation policy for generated artifacts" choices:"prefer-human-readable-markdown-with-inline-metadata;prefer-machine-readable-json-with-stable-field-order;prefer-manpage-compatible-plain-text-with-unicode-disabled" default:"prefer-machine-readable-json-with-stable-field-order"`
	TemplateStrategy string                 `long:"profile-template-selection-strategy-for-runtime-environments" value-name:"PROFILE_TEMPLATE_SELECTION_STRATEGY_NAME" description:"Profile template selection strategy for runtime environments" optional:"yes" optional-value:"prefer-latest-template-compatible-with-runtime-features" choices:"prefer-latest-template-compatible-with-runtime-features;prefer-template-locked-to-application-major-version;prefer-template-selected-by-explicit-environment-marker"`
	ManualEnvOnly    string                 `long:"manual-env-only" description:"Explicit opt-out from global auto env" auto-env:"false" default:"local" order:"-40"`
	ReleaseID        string                 `long:"release-id" value-name:"RELEASE_IDENTIFIER" description:"Release identifier for audit trail" required:"yes"`
	SecretKey        string                 `long:"secret-key" description:"Hidden secret key for debugging deployments" hidden:"yes"`
	HelpColor        string                 `long:"help-color" choices:"none;default;contrast;light" default:"none" description:"Color scheme for built-in help output"`
	Verbose          []bool                 `short:"v" long:"verbose" description:"Increase verbosity level" order:"100"`
	Labels           []ServiceLabel         `long:"label" description:"Service labels"`
	Exec             []string               `long:"exec" description:"Collect args until ';' terminator" terminator:";" order:"-30"`
	Network          AdvancedNetworkOptions `group:"Network Options" namespace:"net" env-namespace:"NET"`
	Demo             AdvancedDemoOptions    `group:"Demo Options" immediate:"true"`
	Count            int                    `long:"count" description:"Example number flag for sort demo" default:"7"`
	Delay            time.Duration          `long:"delay" description:"Example duration flag for sort demo" default:"2s"`
	Deploy           AdvancedDeployCommand  `command:"deploy" description:"Deploy selected targets" long-description:"Run deployment workflow with validation checks.\n\nExamples:\n  advanced-cli deploy --force target artifact\n  advanced-cli deploy --plan target artifact" pass-after-non-option:"yes"`
	Zeta             bool                   `long:"zeta" description:"Example bool flag for sort demo"`
}

type AdvancedPositionalArgs struct {
	Target   string `positional-arg-name:"target" description:"Target service name or host"`
	Artifact string `positional-arg-name:"artifact" description:"Artifact path or reference"`
}

type AdvancedNetworkOptions struct {
	Endpoint string        `long:"endpoint" description:"Service endpoint" auto-env:"true"`
	Mode     string        `long:"mode" description:"Network mode"`
	Timeout  time.Duration `long:"timeout" description:"Request timeout" default:"10s"`
	Retries  int           `long:"retries" description:"Retry attempts" default:"3"`
	TLS      bool          `long:"tls" description:"Enable TLS" order:"50"`
}

type AdvancedDemoOptions struct {
	Help       string `long:"demo-help" value-name:"MODE" choices:"decl;name-asc;name-desc;type" description:"Render built-in help with selected sort mode and exit"`
	Completion string `long:"demo-completion" value-name:"SHELL" choices:"bash;zsh" description:"Render shell completion script and exit"`
	DocFormat  string `long:"demo-doc-format" value-name:"FORMAT" choices:"markdown;html;man" description:"Render documentation in selected format and exit"`
	DocStyle   string `long:"demo-doc-style" value-name:"STYLE" choices:"list;table;code" description:"Render markdown style variant for --demo-doc-format=markdown"`
	INI        bool   `long:"demo-ini" description:"Render example INI and exit"`
}

type AdvancedDeployCommand struct {
	Force bool `long:"force" description:"Force deployment"`
	Plan  bool `long:"plan" description:"Show execution plan only"`
}

func newParser(opts *AdvancedOptions) *flags.Parser {
	p := flags.NewNamedParser(
		"advanced-cli",
		flags.Default|
			flags.VersionFlag|
			flags.ColorErrors|
			flags.EnvProvisioning|
			flags.KeepDescriptionWhitespace|
			flags.DetectShellFlagStyle|
			flags.DetectShellEnvStyle,
	)
	p.LongDescription = "Example of advanced go-flags features:\n  - dynamic defaults\n  - env provisioning and auto-env\n  - terminated options\n  - option sorting per group block"
	p.SetEnvPrefix("DEMO_APP")
	p.SetVersionURL("https://github.com/woozymasta/flags")
	p.SetVersionFields(flags.VersionFieldsAll)
	if err := p.SetMaxLongNameLength(256); err != nil {
		panic(err)
	}

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

func demoOutput(opts *AdvancedOptions, p *flags.Parser) (bool, error) {
	if opts.Demo.Help != "" {
		if err := applySortMode(p, opts.Demo.Help); err != nil {
			return true, err
		}
		p.WriteHelp(os.Stdout)
		return true, nil
	}

	if opts.Demo.Completion != "" {
		if err := p.WriteNamedCompletion(os.Stdout, flags.CompletionShell(opts.Demo.Completion), "advanced-cli"); err != nil {
			return true, err
		}
		return true, nil
	}

	if opts.Demo.DocFormat != "" {
		format, tmpl, err := resolveDocMode(opts.Demo.DocFormat, opts.Demo.DocStyle)
		if err != nil {
			return true, err
		}
		if err := p.WriteDoc(os.Stdout, format, flags.WithBuiltinTemplate(tmpl)); err != nil {
			return true, err
		}
		return true, nil
	}

	if opts.Demo.INI {
		flags.NewIniParser(p).WriteExample(os.Stdout)
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

	if _, err := p.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && (flagsErr.Type == flags.ErrHelp || flagsErr.Type == flags.ErrVersion) {
			os.Exit(0)
		}
		os.Exit(1)
	}

	handled, err := demoOutput(opts, p)
	if handled {
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
}
