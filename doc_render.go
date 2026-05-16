// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

// DocFormat identifies output format for parser documentation rendering.
type DocFormat string

const (
	// DocFormatMan renders classic man page output.
	DocFormatMan DocFormat = "man"
	// DocFormatMarkdown renders markdown documentation.
	DocFormatMarkdown DocFormat = "markdown"
	// DocFormatHTML renders HTML documentation.
	DocFormatHTML DocFormat = "html"
	// DocFormatJSON renders the doc model as a JSON manifest.
	DocFormatJSON DocFormat = "json"
)

type docRenderOptions struct {
	templateData                    map[string]any
	builtinTemplate                 string
	templateText                    string
	programName                     string
	includeBuiltinCommands          []string // nil = include all; non-nil = include only these names
	wrapWidth                       int
	renderStyle                     RenderStyle
	toc                             bool
	trimDescriptions                bool
	includeHidden                   bool
	markHidden                      bool
	jsonCompact                     bool
	hasRenderStyle                  bool
	hasWrapWidth                    bool
	includeBuiltinHelpInSubcommands bool
}

// DocOption configures WriteDoc behavior.
type DocOption func(*docRenderOptions) error

// WithBuiltinTemplate selects a built-in template by name.
func WithBuiltinTemplate(name string) DocOption {
	return func(o *docRenderOptions) error {
		o.builtinTemplate = name
		return nil
	}
}

// WithTemplateString sets custom template content.
func WithTemplateString(text string) DocOption {
	return func(o *docRenderOptions) error {
		o.templateText = text
		return nil
	}
}

// WithTemplateBytes sets custom template content from bytes.
func WithTemplateBytes(data []byte) DocOption {
	return func(o *docRenderOptions) error {
		o.templateText = string(data)
		return nil
	}
}

// WithTemplateData injects additional template data.
func WithTemplateData(data map[string]any) DocOption {
	return func(o *docRenderOptions) error {
		o.templateData = data
		return nil
	}
}

// WithProgramName overrides program/binary name in the generated doc model.
// It affects all templates/formats through Doc.Name and usage lines.
func WithProgramName(name string) DocOption {
	return func(o *docRenderOptions) error {
		o.programName = name
		return nil
	}
}

// WithDocRenderStyle configures how flag tokens and environment placeholders
// are rendered in generated documentation for this WriteDoc call.
func WithDocRenderStyle(style RenderStyle) DocOption {
	return func(o *docRenderOptions) error {
		o.renderStyle = style
		o.hasRenderStyle = true
		return nil
	}
}

// WithDocWrapWidth configures markdown text wrapping width for one WriteDoc
// call. A width of zero disables wrapping.
func WithDocWrapWidth(width int) DocOption {
	return func(o *docRenderOptions) error {
		if width < 0 {
			return errors.New("doc wrap width must be non-negative")
		}
		o.wrapWidth = width
		o.hasWrapWidth = true
		return nil
	}
}

// WithTOC enables table-of-contents rendering for templates supporting it.
func WithTOC(enabled bool) DocOption {
	return func(o *docRenderOptions) error {
		o.toc = enabled
		return nil
	}
}

// WithTrimDescriptions forces description trimming in doc model rendering.
// When enabled, leading/trailing spaces are trimmed per line for
// parser/command/group/option/argument descriptions.
func WithTrimDescriptions(enabled bool) DocOption {
	return func(o *docRenderOptions) error {
		o.trimDescriptions = enabled
		return nil
	}
}

// WithBuiltinCommands sets the list of built-in command names to include in
// the documentation output. Pass nil to include all built-in commands.
// Pass an empty slice to exclude all built-in commands.
// The default when using the built-in docs command is help, version, completion.
func WithBuiltinCommands(names []string) DocOption {
	return func(o *docRenderOptions) error {
		o.includeBuiltinCommands = names
		return nil
	}
}

// WithBuiltinHelpInSubcommands controls whether the builtin help options
// (-h/--help) group is included in nested command sections in generated
// documentation. By default (false) the help group is only shown for the
// root command, matching the behavior of the interactive --help output.
func WithBuiltinHelpInSubcommands(include bool) DocOption {
	return func(o *docRenderOptions) error {
		o.includeBuiltinHelpInSubcommands = include
		return nil
	}
}

// WithIncludeHidden controls whether hidden options/groups/commands are included.
func WithIncludeHidden(include bool) DocOption {
	return func(o *docRenderOptions) error {
		o.includeHidden = include
		return nil
	}
}

// WithMarkHidden controls hidden markers in rendered output.
// It does not include hidden entities by itself. Use WithIncludeHidden(true)
// to include hidden groups/options/commands in the rendered model.
func WithMarkHidden(mark bool) DocOption {
	return func(o *docRenderOptions) error {
		o.markHidden = mark
		return nil
	}
}

// ListBuiltinTemplates returns sorted names of built-in templates.
func ListBuiltinTemplates() []string {
	names := make([]string, 0, len(builtinTemplates))
	for k := range builtinTemplates {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// WriteBuiltinTemplate writes a built-in template source by name.
func WriteBuiltinTemplate(w io.Writer, name string) error {
	tpl, ok := builtinTemplates[name]
	if !ok {
		return fmt.Errorf("unknown builtin template %q", name)
	}

	_, err := w.Write(tpl)
	return err
}

// WriteDoc renders parser documentation in the selected format.
func (p *Parser) WriteDoc(w io.Writer, format DocFormat, opts ...DocOption) error {
	if w == nil {
		return ErrNilWriter
	}

	cfg := docRenderOptions{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return err
		}
	}

	switch format {
	case DocFormatMan:
		return p.writeDocMan(w, cfg)
	case DocFormatMarkdown:
		return p.writeDocMarkdown(w, cfg)
	case DocFormatHTML:
		return p.writeDocHTML(w, cfg)
	case DocFormatJSON:
		return p.writeDocJSON(w, cfg)
	default:
		return fmt.Errorf("unsupported doc format %q", format)
	}
}

// WriteManPage writes a basic man page in groff format to the specified
// writer.
func (p *Parser) WriteManPage(w io.Writer) {
	_ = p.WriteDoc(w, DocFormatMan, WithBuiltinTemplate(DocTemplateManDefault))
}

func withJSONCompact(compact bool) DocOption {
	return func(o *docRenderOptions) error {
		o.jsonCompact = compact
		return nil
	}
}

func (p *Parser) writeDocJSON(w io.Writer, cfg docRenderOptions) error {
	model := p.buildDocModel(cfg)

	enc := json.NewEncoder(w)
	if !cfg.jsonCompact {
		enc.SetIndent("", "  ")
	}

	return enc.Encode(model)
}
