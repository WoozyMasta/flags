// SPDX-FileType: SOURCE
// SPDX-FileCopyrightText: 2012 Jesse van den Kieboom
// SPDX-FileCopyrightText: 2026 Maxim Levchenko (WoozyMasta)
// SPDX-License-Identifier: BSD-3-Clause

package flags

import "strings"

var terminalTitleSetter = setTerminalTitle

func normalizeTerminalTitle(title string) string {
	title = strings.ReplaceAll(title, "\x00", "")
	title = strings.ReplaceAll(title, "\a", "")
	title = strings.ReplaceAll(title, "\x1b", "")
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.ReplaceAll(title, "\n", " ")
	return strings.TrimSpace(title)
}

func (p *Parser) terminalTitleText() string {
	if p == nil {
		return ""
	}

	title := p.TerminalTitle
	if title == "" {
		title = p.Name
	}

	return normalizeTerminalTitle(title)
}

func (p *Parser) applyTerminalTitle() {
	if p == nil || (p.Options&SetTerminalTitle) == None {
		return
	}

	title := p.terminalTitleText()
	if title == "" {
		return
	}

	_ = terminalTitleSetter(title)
}
