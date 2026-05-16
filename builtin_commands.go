// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"time"
)

type builtinCommand interface {
	isBuiltinCommand()
}

type builtinHelpCommand struct {
	parser *Parser
}

func (c *builtinHelpCommand) Execute(args []string) error {
	return withTemporaryActivePath(c.parser, args, func() error {
		c.parser.WriteHelp(os.Stdout)
		return nil
	})
}

type builtinVersionCommand struct {
	parser *Parser

	Short  bool `long:"short" description:"Print version number only" description-i18n:"help.builtin.command.version.short.desc" auto-env:"false"`
	Commit bool `long:"commit" description:"Print commit SHA only" description-i18n:"help.builtin.command.version.commit.desc" auto-env:"false"`
	JSON   bool `long:"json" description:"Print version information as JSON" description-i18n:"help.builtin.command.version.json.desc" auto-env:"false"`
}

func (c *builtinVersionCommand) Execute(_ []string) error {
	if c.Short {
		info := c.parser.VersionInfo()
		version := info.Version
		if version == "" {
			version = c.parser.i18nText("version.value.unknown", "unknown")
		}
		_, _ = fmt.Fprintln(os.Stdout, version)
		return nil
	}

	if c.Commit {
		info := c.parser.VersionInfo()
		commit := info.Revision
		if commit == "" {
			commit = c.parser.i18nText("version.value.unknown", "unknown")
		}
		_, _ = fmt.Fprintln(os.Stdout, commit)
		return nil
	}

	if c.JSON {
		info := c.parser.VersionInfo()
		out := struct {
			File      string     `json:"file,omitempty"`
			Version   string     `json:"version,omitempty"`
			Commit    string     `json:"commit,omitempty"`
			Built     *time.Time `json:"built,omitempty"`
			URL       string     `json:"url,omitempty"`
			Path      string     `json:"path,omitempty"`
			Module    string     `json:"module,omitempty"`
			GoVersion string     `json:"go,omitempty"`
			GOOS      string     `json:"goos,omitempty"`
			GOARCH    string     `json:"goarch,omitempty"`
			Modified  bool       `json:"modified,omitempty"`
			License   string     `json:"license,omitempty"`
		}{
			File:      info.File,
			Version:   info.Version,
			Commit:    info.Revision,
			URL:       info.URL,
			Path:      info.Path,
			Module:    info.ModulePath,
			GoVersion: info.GoVersion,
			GOOS:      info.GOOS,
			GOARCH:    info.GOARCH,
			Modified:  info.Modified,
			License:   info.License,
		}

		if !info.RevisionTime.IsZero() {
			out.Built = &info.RevisionTime
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	c.parser.WriteVersion(os.Stdout, c.parser.versionFields)
	return nil
}

type builtinCompletionCommand struct {
	parser *Parser

	Shell CompletionShell `long:"shell" choices:"bash;zsh;pwsh" description:"Shell completion format" description-i18n:"help.builtin.command.completion.shell.desc" auto-env:"false"`

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`
}

func (c *builtinCompletionCommand) Execute(_ []string) error {
	shell := c.Shell
	if shell == "" {
		shell = DetectCompletionShell()
	}

	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteCompletion(w, shell)
	})
}

type builtinDocsCommand struct {
	HTML     builtinDocHTMLCommand     `command:"html" ini-group:"docs.html" description:"Generate HTML documentation" description-i18n:"help.builtin.command.docs.html.desc"`
	Man      builtinDocManCommand      `command:"man" ini-group:"docs.man" description:"Generate man page documentation" description-i18n:"help.builtin.command.docs.man.desc"`
	JSON     builtinDocJSONCommand     `command:"json" ini-group:"docs.json" description:"Generate JSON documentation manifest" description-i18n:"help.builtin.command.docs.json.desc"`
	MD       builtinDocMarkdownCommand `command:"md" ini-group:"docs.md" description:"Generate Markdown documentation" description-i18n:"help.builtin.command.docs.md.desc"`
	Template builtinDocTemplateCommand `command:"template" ini-group:"docs.template" description:"Export or render documentation templates" description-i18n:"help.builtin.command.docs.template.desc"`
}

type builtinDocProgramNameOption struct {
	ProgramName string `long:"program-name" value-name:"NAME" description:"Override program name used in generated documentation templates" auto-env:"false"`
}

type builtinDocRenderStyleOption struct {
	Style string `long:"style" choices:"auto;posix;windows;shell" description:"Override flag and environment render style used in generated documentation" auto-env:"false"`
}

type builtinDocBuiltinCommandsOption struct {
	Builtins []string `long:"builtins" choices:"help;version;completion;docs;config" description:"Built-in commands to include in generated output" description-i18n:"help.builtin.command.docs.builtins.desc" auto-env:"false"`
}

type builtinDocHelpGroupOption struct {
	HelpInSubcommands bool `long:"help-in-subcommands" description:"Include builtin help options (-h/--help) in nested command sections" auto-env:"false"`
}

type builtinDocManCommand struct {
	parser *Parser

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`

	builtinDocProgramNameOption
	builtinDocRenderStyleOption
	builtinDocBuiltinCommandsOption
	builtinDocHelpGroupOption
	TrimDescriptions bool `long:"trim-descriptions" description:"Trim description whitespace in generated output" auto-env:"false"`

	IncludeHidden bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc" auto-env:"false"`
	MarkHidden    bool `long:"mark-hidden" description:"Mark hidden entities in documentation output" description-i18n:"help.builtin.command.docs.mark_hidden.desc" auto-env:"false"`
}

func (c *builtinDocManCommand) Execute(_ []string) error {
	opts := []DocOption{
		WithBuiltinTemplate(DocTemplateManDefault),
		WithProgramName(c.ProgramName),
		WithTrimDescriptions(c.TrimDescriptions),
		WithIncludeHidden(c.IncludeHidden),
		WithMarkHidden(c.MarkHidden),
		WithBuiltinCommands(c.Builtins),
		WithBuiltinHelpInSubcommands(c.HelpInSubcommands),
	}
	opts = appendBuiltinDocRenderStyleOption(opts, c.Style)
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteDoc(w, DocFormatMan, opts...)
	})
}

type builtinDocHTMLCommand struct {
	parser   *Parser
	Template string `long:"template" choices:"default;styled" default:"default" description:"HTML documentation template" description-i18n:"help.builtin.command.docs.template_html.desc" auto-env:"false"`

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`
	builtinDocProgramNameOption
	builtinDocRenderStyleOption
	builtinDocBuiltinCommandsOption
	builtinDocHelpGroupOption
	TOC              bool `long:"toc" description:"Include table of contents in output" auto-env:"false"`
	TrimDescriptions bool `long:"trim-descriptions" description:"Trim description whitespace in generated output" auto-env:"false"`

	IncludeHidden bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc" auto-env:"false"`
	MarkHidden    bool `long:"mark-hidden" description:"Mark hidden entities in documentation output" description-i18n:"help.builtin.command.docs.mark_hidden.desc" auto-env:"false"`
}

func (c *builtinDocHTMLCommand) Execute(_ []string) error {
	templateName := DocTemplateHTMLDefault
	if c.Template == "styled" {
		templateName = DocTemplateHTMLStyled
	}

	opts := []DocOption{
		WithBuiltinTemplate(templateName),
		WithProgramName(c.ProgramName),
		WithTOC(c.TOC),
		WithTrimDescriptions(c.TrimDescriptions),
		WithIncludeHidden(c.IncludeHidden),
		WithMarkHidden(c.MarkHidden),
		WithBuiltinCommands(c.Builtins),
		WithBuiltinHelpInSubcommands(c.HelpInSubcommands),
	}
	opts = appendBuiltinDocRenderStyleOption(opts, c.Style)
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteDoc(w, DocFormatHTML, opts...)
	})
}

type builtinDocMarkdownCommand struct {
	parser   *Parser
	Template string `long:"template" choices:"list;table;code" default:"list" description:"Markdown documentation template" description-i18n:"help.builtin.command.docs.template_markdown.desc" auto-env:"false"`

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`
	builtinDocProgramNameOption
	builtinDocRenderStyleOption
	builtinDocBuiltinCommandsOption
	builtinDocHelpGroupOption
	WrapWidth int `long:"wrap-width" value-name:"COLUMNS" default:"80" description:"Maximum width for wrapped Markdown text; zero disables wrapping" auto-env:"false"`

	TOC              bool `long:"toc" description:"Include table of contents in output" auto-env:"false"`
	TrimDescriptions bool `long:"trim-descriptions" description:"Trim description whitespace in generated output" auto-env:"false"`

	IncludeHidden bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc" auto-env:"false"`
	MarkHidden    bool `long:"mark-hidden" description:"Mark hidden entities in documentation output" description-i18n:"help.builtin.command.docs.mark_hidden.desc" auto-env:"false"`
}

func (c *builtinDocMarkdownCommand) Execute(_ []string) error {
	templateName := DocTemplateMarkdownList
	switch c.Template {
	case "table":
		templateName = DocTemplateMarkdownTable
	case "code":
		templateName = DocTemplateMarkdownCode
	}

	opts := []DocOption{
		WithBuiltinTemplate(templateName),
		WithProgramName(c.ProgramName),
		WithTOC(c.TOC),
		WithTrimDescriptions(c.TrimDescriptions),
		WithDocWrapWidth(c.WrapWidth),
		WithIncludeHidden(c.IncludeHidden),
		WithMarkHidden(c.MarkHidden),
		WithBuiltinCommands(c.Builtins),
		WithBuiltinHelpInSubcommands(c.HelpInSubcommands),
	}
	opts = appendBuiltinDocRenderStyleOption(opts, c.Style)
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteDoc(w, DocFormatMarkdown, opts...)
	})
}

type builtinDocJSONCommand struct {
	parser *Parser

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`

	builtinDocProgramNameOption
	builtinDocRenderStyleOption
	builtinDocBuiltinCommandsOption
	builtinDocHelpGroupOption
	TrimDescriptions bool `long:"trim-descriptions" description:"Trim description whitespace in generated output" auto-env:"false"`
	IncludeHidden    bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc" auto-env:"false"`
	Compact          bool `long:"compact" description:"Emit compact JSON without indentation" auto-env:"false"`
}

func (c *builtinDocJSONCommand) Execute(_ []string) error {
	opts := []DocOption{
		WithProgramName(c.ProgramName),
		WithTrimDescriptions(c.TrimDescriptions),
		WithIncludeHidden(c.IncludeHidden),
		withJSONCompact(c.Compact),
		WithBuiltinCommands(c.Builtins),
		WithBuiltinHelpInSubcommands(c.HelpInSubcommands),
	}
	opts = appendBuiltinDocRenderStyleOption(opts, c.Style)
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteDoc(w, DocFormatJSON, opts...)
	})
}

type builtinDocTemplateCommand struct {
	Export builtinDocTemplateExportCommand `command:"export" ini-group:"docs.template.export" description:"Export a built-in documentation template" description-i18n:"help.builtin.command.docs.template.export.desc"`
	Render builtinDocTemplateRenderCommand `command:"render" ini-group:"docs.template.render" description:"Render documentation using a custom template" description-i18n:"help.builtin.command.docs.template.render.desc"`
}

type builtinDocTemplateExportCommand struct {
	Name string `long:"name" description:"Built-in template name to export" description-i18n:"help.builtin.command.docs.template.name.desc" required:"yes" auto-env:"false"`

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`
}

func (c *builtinDocTemplateExportCommand) Execute(_ []string) error {
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return WriteBuiltinTemplate(w, c.Name)
	})
}

type builtinDocTemplateRenderCommand struct {
	parser *Parser

	Format string `long:"format" choices:"markdown;html;man;json" default:"markdown" description:"Output format for template rendering" description-i18n:"help.builtin.command.docs.template.format.desc" auto-env:"false"`

	Inputs struct {
		Template string `positional-arg-name:"template" arg-name-i18n:"help.builtin.command.docs.template.input.name" description:"Template file path or - for stdin" arg-description-i18n:"help.builtin.command.docs.template.input.desc"`
		Output   string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`

	builtinDocProgramNameOption
	builtinDocRenderStyleOption
	builtinDocBuiltinCommandsOption
	builtinDocHelpGroupOption
	WrapWidth        int  `long:"wrap-width" value-name:"COLUMNS" default:"80" description:"Maximum width for wrapped Markdown text; zero disables wrapping" auto-env:"false"`
	TOC              bool `long:"toc" description:"Include table of contents in output" auto-env:"false"`
	TrimDescriptions bool `long:"trim-descriptions" description:"Trim description whitespace in generated output" auto-env:"false"`
	IncludeHidden    bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc" auto-env:"false"`
	MarkHidden       bool `long:"mark-hidden" description:"Mark hidden entities in documentation output" description-i18n:"help.builtin.command.docs.mark_hidden.desc" auto-env:"false"`
	Compact          bool `long:"compact" description:"Emit compact JSON without indentation" auto-env:"false"`
}

func (c *builtinDocTemplateRenderCommand) Execute(_ []string) error {
	var tplBytes []byte
	var err error

	if c.Inputs.Template == "" || c.Inputs.Template == "-" {
		tplBytes, err = io.ReadAll(os.Stdin)
	} else {
		tplBytes, err = os.ReadFile(c.Inputs.Template)
	}
	if err != nil {
		return err
	}

	var format DocFormat
	switch c.Format {
	case "html":
		format = DocFormatHTML
	case "man":
		format = DocFormatMan
	case "json":
		format = DocFormatJSON
	default:
		format = DocFormatMarkdown
	}

	opts := []DocOption{
		WithTemplateBytes(tplBytes),
		WithProgramName(c.ProgramName),
		WithTOC(c.TOC),
		WithTrimDescriptions(c.TrimDescriptions),
		WithDocWrapWidth(c.WrapWidth),
		WithIncludeHidden(c.IncludeHidden),
		WithMarkHidden(c.MarkHidden),
		WithBuiltinCommands(c.Builtins),
		WithBuiltinHelpInSubcommands(c.HelpInSubcommands),
	}
	if c.Compact {
		opts = append(opts, withJSONCompact(c.Compact))
	}
	opts = appendBuiltinDocRenderStyleOption(opts, c.Style)

	return writeBuiltinCommandOutput(c.Inputs.Output, func(w io.Writer) error {
		return c.parser.WriteDoc(w, format, opts...)
	})
}

// SetBuiltinCommandHidden controls visibility of an enabled built-in command in
// help, completion, and generated documentation.
func (p *Parser) SetBuiltinCommandHidden(name string, hidden bool) error {
	if p == nil {
		return nil
	}

	if err := p.EnsureBuiltinCommands(); err != nil {
		return err
	}

	cmd := p.Find(name)
	if cmd == nil || !isBuiltinCommandData(cmd.Data()) {
		return newErrorf(ErrUnknownCommand, "unknown built-in command `%s`", name)
	}

	cmd.SetHidden(hidden)
	return nil
}

func isBuiltinCommandData(data any) bool {
	_, ok := data.(builtinCommand)
	return ok
}

func (p *Parser) configureDocBuiltinCommandsOption() {
	enabled := make([]string, 0, 5)
	if (p.Options & HelpCommand) != None {
		enabled = append(enabled, "help")
	}
	if (p.Options & VersionCommand) != None {
		enabled = append(enabled, "version")
	}
	if (p.Options & CompletionCommand) != None {
		enabled = append(enabled, "completion")
	}
	if (p.Options & DocsCommand) != None {
		enabled = append(enabled, "docs")
	}
	if (p.Options & ConfigCommand) != None {
		enabled = append(enabled, "config")
	}
	// docs is in choices but not in defaults: excluded from default output

	preferred := []string{"help", "version", "completion"}
	defaults := make([]string, 0, 3)
	for _, name := range preferred {
		if slices.Contains(enabled, name) {
			defaults = append(defaults, name)
		}
	}

	docCmd := p.Find("docs")
	if docCmd == nil {
		return
	}
	for _, subName := range []string{"html", "man", "md", "json"} {
		subCmd := docCmd.Find(subName)
		if subCmd == nil {
			continue
		}
		if opt := subCmd.Group.FindOptionByLongName("builtins"); opt != nil {
			opt.SetChoices(enabled...)
			opt.SetDefault(defaults...)
		}
	}

	templateCmd := docCmd.Find("template")
	if templateCmd == nil {
		return
	}
	if renderCmd := templateCmd.Find("render"); renderCmd != nil {
		if opt := renderCmd.Group.FindOptionByLongName("builtins"); opt != nil {
			opt.SetChoices(enabled...)
			opt.SetDefault(defaults...)
		}
	}
	if exportCmd := templateCmd.Find("export"); exportCmd != nil {
		if opt := exportCmd.Group.FindOptionByLongName("name"); opt != nil {
			opt.SetChoices(ListBuiltinTemplates()...)
		}
	}
}

func appendBuiltinDocRenderStyleOption(opts []DocOption, style string) []DocOption {
	switch style {
	case "posix":
		return append(opts, WithDocRenderStyle(RenderStylePOSIX))
	case "windows":
		return append(opts, WithDocRenderStyle(RenderStyleWindows))
	case "shell":
		return append(opts, WithDocRenderStyle(RenderStyleShell))
	case "auto":
		return append(opts, WithDocRenderStyle(RenderStyleAuto))
	default:
		return opts
	}
}

type builtinConfigCommand struct {
	parser *Parser

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`

	CommentWidth int `long:"comment-width" value-name:"COLUMNS" value-name-i18n:"help.builtin.command.value.columns" default:"80" description:"Maximum width for wrapped comments" description-i18n:"help.builtin.command.config.comment_width.desc" auto-env:"false"`
}

func (c *builtinConfigCommand) Execute(_ []string) error {
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		NewIniParser(c.parser).WriteExampleWithOptions(w, IniExampleOptions{
			CommentWidth: c.CommentWidth,
		})
		return nil
	})
}

type activeStateEntry struct {
	command *Command
	active  *Command
}

func snapshotActiveState(p *Parser) []activeStateEntry {
	if p == nil {
		return nil
	}

	ret := make([]activeStateEntry, 0)
	p.eachCommand(func(cmd *Command) {
		ret = append(ret, activeStateEntry{
			command: cmd,
			active:  cmd.Active,
		})
	})

	return ret
}

func restoreActiveState(state []activeStateEntry) {
	for _, entry := range state {
		entry.command.Active = entry.active
	}
}

func clearActiveState(p *Parser) {
	if p == nil {
		return
	}

	p.eachCommand(func(cmd *Command) {
		cmd.Active = nil
	})
}

func setActivePath(root *Command, path []string) error {
	if root == nil {
		return nil
	}

	current := root
	for _, name := range path {
		next := current.Find(name)
		if next == nil {
			parser := root.parser()
			msg := "Unknown command `" + name + "`"
			if parser != nil {
				msg = parser.i18nTextf(
					"err.command.unknown",
					"Unknown command `{command}`",
					map[string]string{"command": name},
				)
			}
			return newError(ErrUnknownCommand, msg)
		}
		current.Active = next
		current = next
	}

	return nil
}

func withTemporaryActivePath(p *Parser, path []string, fn func() error) error {
	if p == nil || fn == nil {
		return nil
	}

	state := snapshotActiveState(p)
	defer restoreActiveState(state)

	clearActiveState(p)
	if err := setActivePath(p.Command, path); err != nil {
		return err
	}

	return fn()
}

func writeBuiltinCommandOutput(path string, write func(io.Writer) error) (err error) {
	if path == "" {
		return write(os.Stdout)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	return write(file)
}

func (p *Parser) ensureBuiltinCommands() error {
	if p == nil {
		return nil
	}

	wanted := p.Options & HelpCommands
	missing := wanted &^ p.builtinCommandsAdded
	if missing == None {
		return nil
	}

	if (missing & HelpCommand) != None {
		if err := p.addBuiltinCommand("help", "Show help", "help.builtin.command.help.desc", &builtinHelpCommand{parser: p}); err != nil {
			return err
		}
	}
	if (missing & VersionCommand) != None {
		if err := p.addBuiltinCommand("version", "Show version information", "help.builtin.command.version.desc", &builtinVersionCommand{parser: p}); err != nil {
			return err
		}
	}
	if (missing & CompletionCommand) != None {
		if err := p.addBuiltinCommand("completion", "Generate shell completion", "help.builtin.command.completion.desc", &builtinCompletionCommand{parser: p}); err != nil {
			return err
		}
	}
	if (missing & DocsCommand) != None {
		docs := &builtinDocsCommand{
			Man:  builtinDocManCommand{parser: p},
			HTML: builtinDocHTMLCommand{parser: p},
			MD:   builtinDocMarkdownCommand{parser: p},
			JSON: builtinDocJSONCommand{parser: p},
			Template: builtinDocTemplateCommand{
				Render: builtinDocTemplateRenderCommand{parser: p},
			},
		}
		if err := p.addBuiltinCommand("docs", "Generate documentation", "help.builtin.command.docs.desc", docs); err != nil {
			return err
		}
		p.configureDocBuiltinCommandsOption()
	}
	if (missing & ConfigCommand) != None {
		if err := p.addBuiltinCommand("config", "Generate INI configuration example", "help.builtin.command.config.desc", &builtinConfigCommand{parser: p}); err != nil {
			return err
		}
	}

	p.builtinCommandsAdded |= missing
	if (p.Options & (HelpFlag | VersionFlag)) != None {
		p.addHelpGroups(p.showBuiltinHelp, p.markVersionRequested)
	}

	return nil
}

func (p *Parser) addBuiltinCommand(name string, shortDescription string, shortDescriptionI18n string, data any) error {
	if existing := p.Find(name); existing != nil {
		return newErrorf(ErrDuplicatedFlag, "command `%s` conflicts with built-in command `%s`", existing.Name, name)
	}

	cmd, err := p.AddCommand(name, shortDescription, "", data)
	if err != nil {
		return err
	}
	cmd.CommandGroup = p.builtinCommandGroup
	cmd.CommandGroupI18nKey = p.builtinCommandGroupI18nKey
	cmd.ShortDescriptionI18nKey = shortDescriptionI18n
	cmd.Order = builtinCommandOrder(name)

	return nil
}

func builtinCommandOrder(name string) int {
	switch name {
	case "help":
		return 200
	case "version":
		return 190
	default:
		return 0
	}
}

func (*builtinHelpCommand) isBuiltinCommand()              {}
func (*builtinVersionCommand) isBuiltinCommand()           {}
func (*builtinCompletionCommand) isBuiltinCommand()        {}
func (*builtinDocsCommand) isBuiltinCommand()              {}
func (*builtinDocManCommand) isBuiltinCommand()            {}
func (*builtinDocHTMLCommand) isBuiltinCommand()           {}
func (*builtinDocMarkdownCommand) isBuiltinCommand()       {}
func (*builtinDocJSONCommand) isBuiltinCommand()           {}
func (*builtinDocTemplateCommand) isBuiltinCommand()       {}
func (*builtinDocTemplateExportCommand) isBuiltinCommand() {}
func (*builtinDocTemplateRenderCommand) isBuiltinCommand() {}
func (*builtinConfigCommand) isBuiltinCommand()            {}
