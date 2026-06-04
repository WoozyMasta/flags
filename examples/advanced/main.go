// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

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

const advancedHelpHeader = "Advanced CLI example"

const advancedHelpBanner = `
██████ ▄▄ ▄▄  ▄▄▄  ▄▄   ▄▄ ▄▄▄▄  ▄▄    ▄▄▄▄▄   ▄████▄ █████▄ █████▄
██▄▄   ▀█▄█▀ ██▀██ ██▀▄▀██ ██▄█▀ ██    ██▄▄    ██▄▄██ ██▄▄█▀ ██▄▄█▀
██▄▄▄▄ ██ ██ ██▀██ ██   ██ ██    ██▄▄▄ ██▄▄▄   ██  ██ ██     ██

`

const advancedHelpFooter = "Project: https://github.com/woozymasta/flags"

type DynamicToken string

func (d *DynamicToken) Default() ([]string, error) {
	// DefaultProvider is evaluated during parsing, so defaults can come from
	// runtime state instead of fixed struct tags.
	return []string{"token-from-provider"}, nil
}

type ServiceLabel string

func (l *ServiceLabel) UnmarshalText(text []byte) error {
	// encoding.TextUnmarshaler is enough for custom value parsing when the
	// type does not need the flags-specific Marshaler interface.
	if len(text) == 0 {
		return errors.New("service label cannot be empty")
	}
	*l = ServiceLabel(strings.ToLower(string(text)))
	return nil
}

func (l ServiceLabel) MarshalText() ([]byte, error) {
	// MarshalText is used when the parser renders current/default values.
	if l == "" {
		return nil, errors.New("service label cannot be empty")
	}
	return []byte(strings.ToUpper(string(l))), nil
}

type AdvancedOptions struct {
	Env              AdvancedEnvCommand     `description:"Print greeting using USER from environment" command:"env"`
	Publish          AdvancedPublishCommand `description:"Publish artifact to registry" command:"publish" deprecated:"use deploy instead"`
	Alpha            string                 `description:"Example string flag for sort demo" long:"alpha" default:"a"`
	Profile          string                 `description:"Runtime profile" long:"profile" default:"dev" auto-env:"true"`
	Region           string                 `description:"Cloud region" long:"region" default:"eu-west-1" env:"APP_REGION"`
	Zone             string                 `description:"Cloud zone" long:"zone" deprecated:"use --region with a zone suffix, e.g. eu-west-1a"`
	Strategy         string                 `description:"Deployment strategy selector with long value" long:"deployment-strategy-with-very-long-name" default:"rolling-update-with-pre-drain-and-post-verify" value-name:"STRATEGY_PROFILE_NAME"`
	FormatPolicy     string                 `description:"Output format negotiation policy for generated artifacts" long:"output-format-negotiation-policy-for-generated-artifacts" default:"prefer-machine-readable-json-with-stable-field-order" value-name:"OUTPUT_FORMAT_NEGOTIATION_POLICY_IDENTIFIER" choices:"prefer-human-readable-markdown-with-inline-metadata;prefer-machine-readable-json-with-stable-field-order;prefer-manpage-compatible-plain-text-with-unicode-disabled"`
	TemplateStrategy string                 `description:"Profile template selection strategy for runtime environments" long:"profile-template-selection-strategy-for-runtime-environments" value-name:"PROFILE_TEMPLATE_SELECTION_STRATEGY_NAME" choices:"prefer-latest-template-compatible-with-runtime-features;prefer-template-locked-to-application-major-version;prefer-template-selected-by-explicit-environment-marker" optional:"yes" optional-value:"prefer-latest-template-compatible-with-runtime-features"`
	ManualEnvOnly    string                 `description:"Explicit opt-out from global auto env" long:"manual-env-only" default:"local" order:"-40" auto-env:"false"`
	SecretKey        string                 `description:"Hidden secret key for debugging deployments" long:"secret-key" hidden:"yes"`
	HelpColor        string                 `description:"Color scheme for built-in help output" long:"help-color" default:"none" choices:"none;default;contrast;gray;light"`
	Token            DynamicToken           `description:"Dynamic default token" long:"token"`
	Demo             AdvancedDemoOptions    `group:"Demo Options" immediate:"true"`
	Deploy           AdvancedDeployCommand  `description:"Deploy selected targets" command:"deploy" long-description:"Run deployment workflow with validation checks.\n\nExamples:\n  advanced-cli deploy --force target artifact\n  advanced-cli deploy --plan target artifact"`
	Verbose          []bool                 `description:"Increase verbosity level" long:"verbose" order:"100" short:"V"`
	Labels           []ServiceLabel         `description:"Service labels" long:"label"`
	Exec             []string               `description:"Collect args until ';' terminator" long:"exec" order:"-30" terminator:";"`
	Network          AdvancedNetworkOptions `group:"Network Options" namespace:"net" env-namespace:"NET"`
	Count            int                    `description:"Example number flag for sort demo" long:"count" default:"7"`
	Delay            time.Duration          `description:"Example duration flag for sort demo" long:"delay" default:"2s"`
	Zeta             bool                   `description:"Example bool flag for sort demo" long:"zeta"`
}

type AdvancedNetworkOptions struct {
	Endpoint string        `long:"endpoint" description:"Service endpoint" auto-env:"true"`
	Mode     string        `long:"mode"     description:"Network mode"`
	Timeout  time.Duration `long:"timeout"  description:"Request timeout" default:"10s"`
	Retries  int           `long:"retries"  description:"Retry attempts" default:"3"`
	TLS      bool          `long:"tls"      description:"Enable TLS" order:"50"`
}

// AdvancedDemoOptions is tagged as an immediate group in AdvancedOptions, so
// these render/demo flags can run without satisfying normal required values.
type AdvancedDemoOptions struct {
	Help       string `long:"demo-help"       description:"Render built-in help with selected sort mode and exit"        value-name:"MODE"   choices:"decl;name-asc;name-desc;type"`
	Completion string `long:"demo-completion" description:"Render shell completion script and exit"                      value-name:"SHELL"  choices:"bash;zsh;pwsh"`
	DocFormat  string `long:"demo-doc-format" description:"Render documentation in selected format and exit"             value-name:"FORMAT" choices:"markdown;html;man"`
	DocStyle   string `long:"demo-doc-style"  description:"Render markdown style variant for --demo-doc-format=markdown" value-name:"STYLE"  choices:"list;table;code"`
	INI        bool   `long:"demo-ini"        description:"Render example INI and exit"`
}

type AdvancedDeployCommand struct {
	ReleaseID  string                       `long:"release-id" description:"Release identifier for audit trail" required:"yes" value-name:"RELEASE_IDENTIFIER"`
	Positional AdvancedDeployPositionalArgs `required:"yes" positional-args:"yes"`
	Force      bool                         `long:"force"      description:"Force deployment"`
	Plan       bool                         `long:"plan"       description:"Show execution plan only"`
}

type AdvancedDeployPositionalArgs struct {
	Target   string `positional-arg-name:"target"   description:"Target service name or host"`
	Artifact string `positional-arg-name:"artifact" description:"Artifact path or reference"`
}

type AdvancedPublishCommand struct {
	Positional struct {
		Artifact string `description:"Artifact path or reference" positional-arg-name:"artifact"`
	} `positional-args:"yes"`

	Registry string `long:"registry" description:"Target registry URL"`
	Tag      string `long:"tag"      description:"Image tag" default:"latest"`
}

// AdvancedEnvCommand demonstrates .env file loading. USER is populated from
// the .env file located in examples/advanced/ only when the binary runs with
// that directory as the working directory:
//
//	From the project root - .env is NOT loaded:
//	  go run ./examples/advanced/ env     →  profile: dev
//
//	From this directory - .env IS loaded:
//	  cd examples/advanced && go run . env  →  profile: production
type AdvancedEnvCommand struct {
	// opts is wired in main() before Parse so Execute can read parsed flags.
	opts *AdvancedOptions
}

func (c *AdvancedEnvCommand) Execute(_ []string) error {
	user := os.Getenv("USER")
	if user == "" {
		user = "nobody"
	}
	if _, err := fmt.Printf("Hello %s\nprofile: %s\n", user, c.opts.Profile); err != nil {
		return err
	}
	return nil
}

func newParser(opts *AdvancedOptions) *flags.Parser {
	// NewNamedParser is useful for examples/tests because the displayed command
	// name is stable and not derived from os.Args[0].
	p := flags.NewNamedParser(
		"advanced-cli",
		flags.Default|
			flags.HelpCommands|
			flags.VersionFlag|
			flags.ColorErrors|
			flags.EnvProvisioning|
			flags.KeepDescriptionWhitespace|
			flags.DetectShellFlagStyle|
			flags.DetectShellEnvStyle|
			flags.DotEnv|
			flags.DotEnvFlags,
	)
	p.LongDescription = "Example of advanced go-flags features:\n  - dynamic defaults\n  - env provisioning and auto-env\n  - terminated options\n  - option sorting per group block\n  - .env file support with variable expansion"
	p.SetHelpHeader(advancedHelpHeader)
	p.SetBanner(advancedHelpBanner)
	p.SetHelpFooter(advancedHelpFooter)
	// EnvPrefix composes with env-namespace/auto-env tags, keeping all derived
	// environment variables under one application prefix.
	p.SetEnvPrefix("DEMO_APP")
	p.SetVersionURL("https://github.com/woozymasta/flags")
	p.SetVersionFields(flags.VersionFieldsAll)
	// Long option names are intentionally used here to demonstrate wrapping;
	// production CLIs can opt into a larger limit when needed.
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
	// Sorting is configured on the parser and then reflected in every rendered
	// help/doc surface that uses grouped option presentation.
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
	// ColorHelp is opt-in and can be combined with a custom scheme. The "light"
	// branch shows that applications can define every style role themselves.
	switch mode {
	case "", "none":
		p.Options &^= flags.ColorHelp
	case "default":
		p.Options |= flags.ColorHelp
		p.SetHelpColorScheme(flags.DefaultHelpColorScheme())
	case "contrast":
		p.Options |= flags.ColorHelp
		p.SetHelpColorScheme(flags.HighContrastHelpColorScheme())
	case "gray":
		p.Options |= flags.ColorHelp
		p.SetHelpColorScheme(flags.GrayHelpColorScheme())
	case "light":
		p.Options |= flags.ColorHelp
		p.SetHelpColorScheme(flags.HelpColorScheme{
			BaseText:             flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack, UseBG: true, BG: flags.ColorBrightWhite},
			OptionShort:          flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			OptionLong:           flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			OptionDesc:           flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack},
			OptionEnv:            flags.HelpTextStyle{UseFG: true, FG: flags.ColorCyan},
			OptionDefault:        flags.HelpTextStyle{UseFG: true, FG: flags.ColorMagenta},
			OptionChoices:        flags.HelpTextStyle{UseFG: true, FG: flags.ColorGreen, Bold: true},
			VersionLabel:         flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, Bold: true},
			VersionValue:         flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack},
			UsageHeader:          flags.HelpTextStyle{UseFG: true, FG: flags.ColorRed, Bold: true},
			UsageText:            flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack, Bold: true},
			CommandSectionHeader: flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlack, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			CommandGroupHeader:   flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlack, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			CommandName:          flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, UseBG: true, BG: flags.ColorBrightYellow, Bold: true},
			CommandDesc:          flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlack, UseBG: true, BG: flags.ColorBrightYellow},
			ArgumentsHeader:      flags.HelpTextStyle{UseFG: true, FG: flags.ColorRed, Bold: true},
			ArgumentName:         flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, Bold: true},
			ArgumentDesc:         flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack},
			GroupHeader:          flags.HelpTextStyle{UseFG: true, FG: flags.ColorRed, Bold: true, Underline: true},
			HelpHeader:           flags.HelpTextStyle{UseFG: true, FG: flags.ColorGreen, Bold: true},
			Banner:               flags.HelpTextStyle{UseFG: true, FG: flags.ColorBlue, Bold: true},
			HelpFooter:           flags.HelpTextStyle{UseFG: true, FG: flags.ColorBrightBlack},
		})
	default:
		return fmt.Errorf("unknown help color mode %q", mode)
	}

	return nil
}

func detectHelpColorArg(args []string) (string, bool) {
	// This option is inspected before Parse so the selected color scheme is
	// already active when a later --help request renders built-in help.
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
	// Demo flags are normal parsed options. After parsing, this helper turns
	// them into generated artifacts and reports whether normal app flow should
	// stop.
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
	// WriteDoc separates the output format from the template name. That lets an
	// app expose a small UX vocabulary while still using built-in templates.
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
	opts.Env.opts = opts
	p := newParser(opts)

	// Pre-parse hooks like this are a practical way to let a flag influence
	// parser rendering behavior before Parse handles --help.
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
