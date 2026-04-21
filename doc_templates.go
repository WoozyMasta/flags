// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import (
	"embed"
	"fmt"
	"io/fs"
)

const (
	// DocTemplateHTMLDefault is the built-in HTML template name.
	DocTemplateHTMLDefault = "html/default"
	// DocTemplateHTMLStyled is the built-in styled HTML template name.
	DocTemplateHTMLStyled = "html/styled"
	// DocTemplateManDefault is the built-in man page template name.
	DocTemplateManDefault = "man/default"
	// DocTemplateMarkdownCode is the built-in markdown code-block template name.
	DocTemplateMarkdownCode = "markdown/code"
	// DocTemplateMarkdownList is the built-in markdown list template name.
	DocTemplateMarkdownList = "markdown/list"
	// DocTemplateMarkdownTable is the built-in markdown table template name.
	DocTemplateMarkdownTable = "markdown/table"
)

var builtinTemplateFiles = map[string]string{
	DocTemplateHTMLDefault:   "html-default.tmpl",
	DocTemplateHTMLStyled:    "html-styled.tmpl",
	DocTemplateManDefault:    "man-default.tmpl",
	DocTemplateMarkdownCode:  "markdown-code.tmpl",
	DocTemplateMarkdownList:  "markdown-list.tmpl",
	DocTemplateMarkdownTable: "markdown-table.tmpl",
}

var builtinTemplates = loadBuiltinTemplates()

//go:embed templates/*
var embeddedDocTemplates embed.FS

func loadBuiltinTemplates() map[string][]byte {
	templates := make(map[string][]byte, len(builtinTemplateFiles))

	for name, file := range builtinTemplateFiles {
		content, err := readBuiltinTemplate(file)
		if err != nil {
			panic(fmt.Sprintf("failed to load builtin template %q from %q: %v", name, file, err))
		}
		templates[name] = content
	}

	return templates
}

func readBuiltinTemplate(name string) ([]byte, error) {
	return fs.ReadFile(embeddedDocTemplates, "templates/"+name)
}
