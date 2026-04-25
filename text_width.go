// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"unicode"

	"golang.org/x/text/width"
)

func textWidth(s string) int {
	total := 0
	inEscape := false

	for _, r := range s {
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}

		if r == '\x1b' {
			inEscape = true
			continue
		}

		total += runeTextWidth(r)
	}

	return total
}

func runeTextWidth(r rune) int {
	switch {
	case r == '\t':
		return 4
	case r == '\n' || r == '\r':
		return 0
	case unicode.IsControl(r):
		return 0
	case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r):
		return 0
	}

	switch width.LookupRune(r).Kind() {
	case width.EastAsianFullwidth, width.EastAsianWide:
		return 2
	default:
		return 1
	}
}

func splitTextWidth(s string, maxWidth int) (int, bool) {
	if maxWidth <= 0 {
		return 0, false
	}

	width := 0
	lastSpace := -1
	for idx, r := range s {
		nextWidth := width + runeTextWidth(r)
		if unicode.IsSpace(r) {
			lastSpace = idx
		}
		if nextWidth > maxWidth {
			if lastSpace >= 0 {
				return lastSpace, true
			}
			if idx == 0 {
				return len(string(r)), false
			}
			return idx, false
		}
		width = nextWidth
	}

	return len(s), false
}
