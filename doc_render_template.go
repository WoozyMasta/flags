// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

type docTemplateContext struct {
	Data       map[string]any
	Doc        docParser
	MarkHidden bool
}

func (p *Parser) writeDocMarkdown(w io.Writer, cfg docRenderOptions) error {
	templateText := cfg.templateText
	if templateText == "" {
		name := cfg.builtinTemplate
		if name == "" {
			name = DocTemplateMarkdownList
		}

		tpl, ok := builtinTemplates[name]
		if !ok {
			return fmt.Errorf("unknown builtin template %q", name)
		}
		templateText = string(tpl)
	}

	var raw bytes.Buffer
	if err := p.executeDocTemplate(&raw, templateText, cfg.templateData, cfg); err != nil {
		return err
	}

	clean := normalizeMarkdownRender(raw.String())
	_, err := io.WriteString(w, clean)
	return err
}

func (p *Parser) writeDocMan(w io.Writer, cfg docRenderOptions) error {
	templateText := cfg.templateText
	if templateText == "" {
		name := cfg.builtinTemplate
		if name == "" {
			name = DocTemplateManDefault
		}

		tpl, ok := builtinTemplates[name]
		if !ok {
			return fmt.Errorf("unknown builtin template %q", name)
		}
		templateText = string(tpl)
	}

	return p.executeDocTemplate(w, templateText, cfg.templateData, cfg)
}

func (p *Parser) writeDocHTML(w io.Writer, cfg docRenderOptions) error {
	templateText := cfg.templateText
	if templateText == "" {
		name := cfg.builtinTemplate
		if name == "" {
			name = DocTemplateHTMLDefault
		}

		tpl, ok := builtinTemplates[name]
		if !ok {
			return fmt.Errorf("unknown builtin template %q", name)
		}
		templateText = string(tpl)
	}

	return p.executeDocTemplate(w, templateText, cfg.templateData, cfg)
}

func (p *Parser) executeDocTemplate(w io.Writer, templateText string, data map[string]any, cfg docRenderOptions) error {
	tpl, err := template.New("doc").Funcs(docTemplateFuncs(cfg.markHidden, p.optionRenderFormat())).Parse(templateText)
	if err != nil {
		return err
	}

	ctx := docTemplateContext{
		Doc:        p.buildDocModel(cfg),
		Data:       data,
		MarkHidden: cfg.markHidden,
	}

	return tpl.Execute(w, ctx)
}

func docTemplateFuncs(markHidden bool, format optionRenderFormat) template.FuncMap {
	return template.FuncMap{
		"hiddenMark": func(hidden bool) bool {
			return markHidden && hidden
		},

		"optionForms": func(opt docOption) []string {
			forms := make([]string, 0, 2)
			addValue := func(base string) string {
				if opt.Optional {
					val := opt.OptionalVal
					if val == "" {
						val = "VALUE"
					}
					return fmt.Sprintf("%s [=%s]", base, val)
				}
				if opt.ValueName != "" {
					return fmt.Sprintf("%s %s", base, opt.ValueName)
				}
				return base
			}
			if opt.Short != "" {
				forms = append(forms, addValue(string(format.shortDelimiter)+opt.Short))
			}
			if opt.Long != "" {
				forms = append(forms, addValue(format.longDelimiter+opt.Long))
			}
			return forms
		},

		"codeJoin": func(items []string) string {
			if len(items) == 0 {
				return ""
			}
			out := make([]string, len(items))
			for i, it := range items {
				out[i] = "`" + it + "`"
			}
			return strings.Join(out, ", ")
		},

		"quoteMarkdown": func(s string) string {
			return strings.ReplaceAll(s, "\\", "\\\\")
		},

		"quoteMan": manQuote,

		"manInline": manInline,

		"quoteHTML": func(s string) string {
			replacer := strings.NewReplacer(
				"&", "&amp;",
				"<", "&lt;",
				">", "&gt;",
				`"`, "&quot;",
				"'", "&#39;",
			)
			return replacer.Replace(s)
		},

		"join": strings.Join,

		"wrap": func(s string, width int) string {
			if width <= 0 {
				return s
			}
			return wrapText(s, width, "", true)
		},

		"markdownWrap": func(s string, width ...int) string {
			maxWidth := 76
			if len(width) > 0 && width[0] > 0 {
				maxWidth = width[0]
			}
			return wrapMarkdownText(s, maxWidth)
		},

		"indent": func(s string, spaces int) string {
			if s == "" {
				return ""
			}
			pad := strings.Repeat(" ", spaces)
			lines := strings.Split(s, "\n")
			for i := range lines {
				lines[i] = pad + lines[i]
			}
			return strings.Join(lines, "\n")
		},

		"defaultValue": func(v string) string {
			if v == "" {
				return ""
			}
			return " (default: " + v + ")"
		},

		"code": func(s string) string {
			return "`" + s + "`"
		},

		"codeFenceOpen": func() string {
			return "```text"
		},

		"codeFenceClose": func() string {
			return "```"
		},

		"isRequired": func(opt docOption) bool {
			return opt.Required
		},

		"hasDefault": func(opt docOption) bool {
			return opt.Default != ""
		},

		"hasEnv": func(opt docOption) bool {
			return opt.Env != ""
		},

		"isBool": func(opt docOption) bool {
			return opt.TypeClass == OptionTypeBool
		},

		"isCollection": func(opt docOption) bool {
			return opt.TypeClass == OptionTypeCollection
		},
	}
}

func manQuoteLines(s string) string {
	lines := strings.Split(s, "\n")
	parts := make([]string, 0, len(lines))

	for _, line := range lines {
		parts = append(parts, manQuote(line))
	}

	return strings.Join(parts, "\n")
}

func manQuote(s string) string {
	return strings.ReplaceAll(s, "\\", "\\\\")
}

func manInline(s string) string {
	var b strings.Builder

	for {
		idx := strings.IndexRune(s, '`')

		if idx < 0 {
			b.WriteString(manQuoteLines(s))
			break
		}

		b.WriteString(manQuoteLines(s[:idx]))

		s = s[idx+1:]
		idx = strings.IndexRune(s, '\'')

		if idx < 0 {
			b.WriteString(manQuoteLines(s))
			break
		}

		b.WriteString("\\fB")
		b.WriteString(manQuoteLines(s[:idx]))
		b.WriteString("\\fP")
		s = s[idx+1:]
	}

	return b.String()
}

func wrapMarkdownText(s string, width int) string {
	if s == "" {
		return s
	}

	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	inFence := false

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		compact := strings.TrimSpace(trimmed)

		if strings.HasPrefix(compact, "```") || strings.HasPrefix(compact, "~~~") {
			inFence = !inFence
			out = append(out, trimmed)
			continue
		}

		if inFence {
			out = append(out, trimmed)
			continue
		}

		if compact == "" {
			out = append(out, "")
			continue
		}

		if strings.HasPrefix(trimmed, "    ") ||
			strings.HasPrefix(trimmed, "\t") ||
			strings.HasPrefix(compact, "|") {
			out = append(out, trimmed)
			continue
		}

		out = append(out, wrapText(compact, width, "", true))
	}

	return strings.Join(out, "\n")
}

func normalizeMarkdownRender(in string) string {
	in = strings.ReplaceAll(in, "\r\n", "\n")
	lines := strings.Split(in, "\n")
	out := make([]string, 0, len(lines))
	inFence := false
	isBulletLine := func(s string) bool {
		return strings.HasPrefix(s, "- ") || strings.HasPrefix(s, "* ")
	}

	nextNonEmpty := func(start int) string {
		for i := start; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				return strings.TrimSpace(lines[i])
			}
		}
		return ""
	}

	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		compact := strings.TrimSpace(trimmed)

		if strings.HasPrefix(compact, "```") || strings.HasPrefix(compact, "~~~") {
			inFence = !inFence
			out = append(out, trimmed)
			continue
		}

		if compact == "" {
			if inFence {
				continue
			}

			prev := ""
			for j := len(out) - 1; j >= 0; j-- {
				if strings.TrimSpace(out[j]) != "" {
					prev = strings.TrimSpace(out[j])
					break
				}
			}

			next := nextNonEmpty(i + 1)
			if isBulletLine(prev) && isBulletLine(next) {
				continue
			}

			if strings.HasPrefix(prev, "|") && strings.HasPrefix(next, "|") {
				continue
			}
		}

		out = append(out, trimmed)
	}

	result := strings.Join(out, "\n")
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return result
}
