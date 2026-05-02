// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
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
)

type docRenderOptions struct {
	templateData    map[string]any
	builtinTemplate string
	templateText    string
	programName     string
	includeHidden   bool
	markHidden      bool
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
	default:
		return fmt.Errorf("unsupported doc format %q", format)
	}
}

// WriteManPage writes a basic man page in groff format to the specified
// writer.
func (p *Parser) WriteManPage(w io.Writer) {
	_ = p.WriteDoc(w, DocFormatMan, WithBuiltinTemplate(DocTemplateManDefault))
}
