// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"io"
	"os"
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
}

func (c *builtinVersionCommand) Execute(_ []string) error {
	c.parser.WriteVersion(os.Stdout, c.parser.versionFields)
	return nil
}

type builtinCompletionCommand struct {
	parser *Parser

	Shell CompletionShell `long:"shell" value-name:"SHELL" value-name-i18n:"help.builtin.command.value.shell" choices:"bash;zsh;pwsh" description:"Shell completion format" description-i18n:"help.builtin.command.completion.shell.desc"`

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
	Man  builtinDocManCommand      `command:"man" ini-group:"docs.man" description:"Generate man page documentation" description-i18n:"help.builtin.command.docs.man.desc"`
	HTML builtinDocHTMLCommand     `command:"html" ini-group:"docs.html" description:"Generate HTML documentation" description-i18n:"help.builtin.command.docs.html.desc"`
	MD   builtinDocMarkdownCommand `command:"md" ini-group:"docs.md" description:"Generate Markdown documentation" description-i18n:"help.builtin.command.docs.md.desc"`
}

type builtinDocProgramNameOption struct {
	ProgramName string `long:"program-name" value-name:"NAME" description:"Override program name used in generated documentation templates"`
}

type builtinDocManCommand struct {
	parser *Parser

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`

	builtinDocProgramNameOption

	IncludeHidden bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc"`
	MarkHidden    bool `long:"mark-hidden" description:"Mark hidden entities in documentation output" description-i18n:"help.builtin.command.docs.mark_hidden.desc"`
}

func (c *builtinDocManCommand) Execute(_ []string) error {
	opts := []DocOption{
		WithBuiltinTemplate(DocTemplateManDefault),
		WithProgramName(c.ProgramName),
		WithIncludeHidden(c.IncludeHidden),
		WithMarkHidden(c.MarkHidden),
	}
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteDoc(w, DocFormatMan, opts...)
	})
}

type builtinDocHTMLCommand struct {
	parser   *Parser
	Template string `long:"template" value-name:"TEMPLATE" value-name-i18n:"help.builtin.command.value.template" choices:"default;styled" default:"default" description:"HTML documentation template" description-i18n:"help.builtin.command.docs.template_html.desc"`

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`
	builtinDocProgramNameOption
	IncludeHidden bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc"`
	MarkHidden    bool `long:"mark-hidden" description:"Mark hidden entities in documentation output" description-i18n:"help.builtin.command.docs.mark_hidden.desc"`
}

func (c *builtinDocHTMLCommand) Execute(_ []string) error {
	templateName := DocTemplateHTMLDefault
	if c.Template == "styled" {
		templateName = DocTemplateHTMLStyled
	}

	opts := []DocOption{
		WithBuiltinTemplate(templateName),
		WithProgramName(c.ProgramName),
		WithIncludeHidden(c.IncludeHidden),
		WithMarkHidden(c.MarkHidden),
	}
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteDoc(w, DocFormatHTML, opts...)
	})
}

type builtinDocMarkdownCommand struct {
	parser   *Parser
	Template string `long:"template" value-name:"TEMPLATE" value-name-i18n:"help.builtin.command.value.template" choices:"list;table;code" default:"list" description:"Markdown documentation template" description-i18n:"help.builtin.command.docs.template_markdown.desc"`

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`
	builtinDocProgramNameOption
	IncludeHidden bool `long:"include-hidden" description:"Include hidden options, groups and commands" description-i18n:"help.builtin.command.docs.include_hidden.desc"`
	MarkHidden    bool `long:"mark-hidden" description:"Mark hidden entities in documentation output" description-i18n:"help.builtin.command.docs.mark_hidden.desc"`
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
		WithIncludeHidden(c.IncludeHidden),
		WithMarkHidden(c.MarkHidden),
	}
	return writeBuiltinCommandOutput(c.Output.Path, func(w io.Writer) error {
		return c.parser.WriteDoc(w, DocFormatMarkdown, opts...)
	})
}

type builtinConfigCommand struct {
	parser *Parser

	Output struct {
		Path string `positional-arg-name:"output" arg-name-i18n:"help.builtin.command.output.name" description:"Output file path" arg-description-i18n:"help.builtin.command.output.desc"`
	} `positional-args:"yes"`

	CommentWidth int `long:"comment-width" value-name:"COLUMNS" value-name-i18n:"help.builtin.command.value.columns" default:"80" description:"Maximum width for wrapped comments" description-i18n:"help.builtin.command.config.comment_width.desc"`
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
		}
		if err := p.addBuiltinCommand("docs", "Generate documentation", "help.builtin.command.docs.desc", docs); err != nil {
			return err
		}
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

func (*builtinHelpCommand) isBuiltinCommand()        {}
func (*builtinVersionCommand) isBuiltinCommand()     {}
func (*builtinCompletionCommand) isBuiltinCommand()  {}
func (*builtinDocsCommand) isBuiltinCommand()        {}
func (*builtinDocManCommand) isBuiltinCommand()      {}
func (*builtinDocHTMLCommand) isBuiltinCommand()     {}
func (*builtinDocMarkdownCommand) isBuiltinCommand() {}
func (*builtinConfigCommand) isBuiltinCommand()      {}
