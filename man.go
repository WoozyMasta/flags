// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

// manConfig holds man-page-specific configuration fields.
// Fields shared with other doc formats (Author, BugsURL, etc.) live in
// VersionInfo. The SEE ALSO cross-references live in Parser.seeAlso.
type manConfig struct {
	Section int
}

// SetManSection sets the man page section number used in the .TH header line.
// Valid values are 1–9. The default is 1 (user commands).
func (p *Parser) SetManSection(section int) {
	p.manConfig.Section = section
}
