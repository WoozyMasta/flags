// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

// .env file parser adapted from github.com/joho/godotenv (MIT licence).
// Key differences: variable expansion is handled by dotenv_expand.go which
// adds ${VAR:=default} assignment and $${VAR} escaping on top of the
// original $VAR / ${VAR} substitution.

package flags

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// sentinel errors returned by the .env parser.
var (
	errDotEnvZeroLength     = errors.New("zero length string")
	errDotEnvUnexpectedChar = errors.New("unexpected character")
	errDotEnvUnterminatedQ  = errors.New("unterminated quoted value")
)

const (
	dotEnvComment     = '#'
	dotEnvSingleQuote = '\''
	dotEnvDoubleQuote = '"'
	dotEnvExport      = "export"
)

// parseDotEnvBytes parses src and returns a key->value map.
// When noExpand is true, dollar-sign sequences are left as-is.
func parseDotEnvBytes(src []byte, noExpand bool) (map[string]string, error) {
	src = bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	out := make(map[string]string)
	cutset := src

	for {
		cutset = dotEnvStatementStart(cutset)
		if cutset == nil {
			break
		}

		key, left, err := dotEnvLocateKey(cutset)
		if err != nil {
			return nil, err
		}

		value, left, err := dotEnvExtractValue(left, out, noExpand)
		if err != nil {
			return nil, err
		}

		out[key] = value
		cutset = left
	}

	return out, nil
}

// dotEnvStatementStart skips whitespace and comment lines, returning the
// slice starting at the next statement, or nil at EOF.
func dotEnvStatementStart(src []byte) []byte {
	for {
		pos := bytes.IndexFunc(src, func(r rune) bool { return !unicode.IsSpace(r) })
		if pos == -1 {
			return nil
		}

		src = src[pos:]
		if src[0] != dotEnvComment {
			return src
		}

		// Skip rest of comment line.
		pos = bytes.IndexByte(src, '\n')
		if pos == -1 {
			return nil
		}

		src = src[pos:]
	}
}

// dotEnvLocateKey extracts the variable name and returns the remainder.
func dotEnvLocateKey(src []byte) (key string, rest []byte, err error) {
	src = bytes.TrimLeftFunc(src, dotEnvIsSpace)

	// Strip optional "export " prefix.
	if after, ok := bytes.CutPrefix(src, []byte(dotEnvExport)); ok {
		trimmed := after
		if len(trimmed) > 0 && dotEnvIsSpace(rune(trimmed[0])) {
			src = bytes.TrimLeftFunc(trimmed, dotEnvIsSpace)
		}
	}

	offset := 0
loop:
	for i, ch := range src {
		r := rune(ch)
		if dotEnvIsSpace(r) {
			continue
		}

		switch ch {
		case '=', ':':
			key = string(src[:i])
			offset = i + 1
			break loop
		case '_':
		default:
			if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '.' || r == '-' {
				continue
			}

			return "", nil, fmt.Errorf("%w %q near %q",
				errDotEnvUnexpectedChar, string(ch), string(src))
		}
	}

	if len(src) == 0 {
		return "", nil, errDotEnvZeroLength
	}

	key = strings.TrimRightFunc(key, unicode.IsSpace)
	rest = bytes.TrimLeftFunc(src[offset:], dotEnvIsSpace)
	return key, rest, nil
}

// dotEnvExtractValue reads the value (quoted or unquoted) and applies
// expansion, returning the value and the remaining slice.
func dotEnvExtractValue(src []byte, vars map[string]string, noExpand bool) (value string, rest []byte, err error) {
	if len(src) == 0 {
		return "", src, nil
	}

	quote, isQuoted := dotEnvQuotePrefix(src)
	if !isQuoted {
		return dotEnvExtractUnquoted(src, vars, noExpand)
	}

	return dotEnvExtractQuoted(src, quote, vars, noExpand)
}

// dotEnvExtractUnquoted reads an unquoted value up to EOL, strips trailing
// whitespace and inline comments, then applies expansion.
func dotEnvExtractUnquoted(src []byte, vars map[string]string, noExpand bool) (string, []byte, error) {
	eol := bytes.IndexFunc(src, func(r rune) bool { return r == '\n' || r == '\r' })
	if eol == -1 {
		eol = len(src)
	}

	line := []rune(string(src[:eol]))
	end := len(line)

	// Strip inline comment: "# ..." preceded by whitespace.
	for i := 1; i < end; i++ {
		if line[i] == dotEnvComment && dotEnvIsSpace(line[i-1]) {
			end = i
			break
		}
	}

	raw := strings.TrimFunc(string(line[:end]), dotEnvIsSpace)
	return expandDotEnvValue(raw, vars, noExpand), src[eol:], nil
}

// dotEnvExtractQuoted reads a single- or double-quoted value. Double-quoted
// values receive escape processing (\n, \r, \\) and expansion; single-quoted
// values are returned literally with no expansion.
func dotEnvExtractQuoted(src []byte, quote byte, vars map[string]string, noExpand bool) (string, []byte, error) {
	for i := 1; i < len(src); i++ {
		if src[i] != quote {
			continue
		}

		// Count preceding backslashes - an odd number means this quote is escaped.
		bs := 0
		for j := i - 1; j >= 0 && src[j] == '\\'; j-- {
			bs++
		}

		if bs%2 == 1 {
			continue
		}

		// Found the closing quote.
		inner := string(src[1:i])

		if quote == dotEnvDoubleQuote {
			inner = dotEnvUnescapeDouble(inner)
			inner = expandDotEnvValue(inner, vars, noExpand)
		}
		// Single-quoted: no unescaping, no expansion.

		return inner, src[i+1:], nil
	}

	// No closing quote found - find EOL for error context.
	end := bytes.IndexByte(src, '\n')
	if end == -1 {
		end = len(src)
	}

	return "", nil, fmt.Errorf("%w: %s", errDotEnvUnterminatedQ, src[:end])
}

// dotEnvUnescapeDouble processes backslash sequences inside double-quoted values.
// Handles \n, \r, \\; all other \X sequences are passed through unchanged.
func dotEnvUnescapeDouble(s string) string {
	var out strings.Builder
	out.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] != '\\' || i+1 >= len(s) {
			out.WriteByte(s[i])
			i++
			continue
		}

		switch s[i+1] {
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case '\\':
			out.WriteByte('\\')
		default:
			out.WriteByte(s[i])
			out.WriteByte(s[i+1])
		}

		i += 2
	}

	return out.String()
}

// dotEnvQuotePrefix reports whether src begins with a quote character and
// returns that character.
func dotEnvQuotePrefix(src []byte) (byte, bool) {
	if len(src) == 0 {
		return 0, false
	}

	switch src[0] {
	case dotEnvDoubleQuote, dotEnvSingleQuote:
		return src[0], true
	}

	return 0, false
}

// dotEnvIsSpace reports whether r is horizontal whitespace (not a newline).
func dotEnvIsSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	}

	return false
}
