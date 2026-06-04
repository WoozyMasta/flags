// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"os"
	"strings"
)

// expandDotEnvValue expands variable references in a .env value string.
//
// When noExpand is true every dollar sign is treated as a literal character
// and no substitution is performed.
//
// Otherwise the following syntax is recognised:
//
//   - $VAR        - substitute value of VAR
//   - ${VAR}      - substitute value of VAR (braced form)
//   - ${VAR:=def} - substitute value of VAR; if VAR is unset or empty, set it to "def" in vars and return "def"
//   - $${VAR}     - literal ${VAR} (escaped, no substitution)
//
// Variable lookup checks vars first, then os.Getenv.
// An unrecognised variable expands to an empty string.
func expandDotEnvValue(v string, vars map[string]string, noExpand bool) string {
	if noExpand || !strings.ContainsAny(v, "$") {
		return v
	}

	return expandDotEnvInner(v, vars)
}

// maskStart / maskEnd are sentinel bytes used to hide $${...} sequences
// during expansion so they are never substituted.
const (
	dotEnvMaskStart = "\x00"
	dotEnvMaskEnd   = "\x01"
)

// expandDotEnvInner performs the actual expansion after noExpand has been
// checked. It uses a mask-then-expand-then-unmask approach inspired by
// jamle/expand.go to correctly handle $${VAR} escaping.
func expandDotEnvInner(v string, vars map[string]string) string {
	masked := dotEnvMaskEscaped(v)

	// Up to 10 passes to resolve nested substitutions, e.g. A=${B} B=${C}.
	for range 10 {
		next, changed := dotEnvReplaceVars(masked, vars)
		if !changed {
			break
		}

		masked = next
	}

	// Expand simple $VAR forms that may remain after ${} processing.
	masked = dotEnvExpandSimple(masked, vars)

	if !strings.ContainsRune(masked, '\x00') {
		return masked
	}

	// Unmask \x00...\x01 → ${...}
	return dotEnvUnmask(masked)
}

// dotEnvMaskEscaped replaces $${...} sequences (with balanced braces) with
// \x00content\x01 markers so that the content is never expanded.
func dotEnvMaskEscaped(in string) string {
	if !strings.Contains(in, "$${") {
		return in
	}

	var out strings.Builder
	out.Grow(len(in))

	for i := 0; i < len(in); {
		if i+2 < len(in) && in[i] == '$' && in[i+1] == '$' && in[i+2] == '{' {
			j, ok := dotEnvFindClose(in, i+2)
			if ok {
				out.WriteString(dotEnvMaskStart)
				out.WriteString(in[i+3 : j])
				out.WriteString(dotEnvMaskEnd)
				i = j + 1
				continue
			}
		}

		out.WriteByte(in[i])
		i++
	}

	return out.String()
}

// dotEnvUnmask restores masked sequences back to ${...} literals.
func dotEnvUnmask(in string) string {
	var out strings.Builder
	out.Grow(len(in))

	i := 0
	for i < len(in) {
		if in[i] == '\x00' {
			j := strings.IndexByte(in[i+1:], '\x01')
			if j >= 0 {
				out.WriteString("${")
				out.WriteString(in[i+1 : i+1+j])
				out.WriteByte('}')
				i = i + 1 + j + 1
				continue
			}
		}

		out.WriteByte(in[i])
		i++
	}

	return out.String()
}

// dotEnvFindClose returns the index of the matching '}' for a '{' at openPos.
func dotEnvFindClose(in string, openPos int) (int, bool) {
	if openPos < 0 || openPos >= len(in) || in[openPos] != '{' {
		return 0, false
	}

	depth := 1

	for i := openPos + 1; i < len(in); i++ {
		switch in[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}

	return 0, false
}

// dotEnvReplaceVars finds and replaces the innermost ${...} expressions.
// Returns the updated string and whether any substitution was made.
func dotEnvReplaceVars(in string, vars map[string]string) (string, bool) {
	ranges := dotEnvFindInnermost(in)
	if len(ranges) == 0 {
		return in, false
	}

	var out strings.Builder
	out.Grow(len(in))

	cursor := 0
	changed := false

	for _, r := range ranges {
		content := in[r[0]+2 : r[1]]
		resolved := dotEnvResolveVar(content, vars)

		out.WriteString(in[cursor:r[0]])
		out.WriteString(resolved)
		cursor = r[1] + 1
		changed = true
	}

	out.WriteString(in[cursor:])
	return out.String(), changed
}

// dotEnvFindInnermost returns [start, end] pairs for innermost ${...} blocks.
// Masked segments (starting with \x00) are never opened as variable refs.
func dotEnvFindInnermost(in string) [][2]int {
	type frame struct{ pos int }

	stack := make([]frame, 0, 8)
	result := make([][2]int, 0, 8)
	hasChild := make([]bool, 0, 8)

	for i := 0; i < len(in); i++ {
		if in[i] == '\x00' {
			// Skip masked segment entirely.
			j := strings.IndexByte(in[i+1:], '\x01')
			if j >= 0 {
				i = i + 1 + j
			}

			continue
		}

		if i+1 < len(in) && in[i] == '$' && in[i+1] == '{' {
			if len(hasChild) > 0 {
				hasChild[len(hasChild)-1] = true
			}

			stack = append(stack, frame{i})
			hasChild = append(hasChild, false)
			i++ // skip '{'
			continue
		}

		if in[i] == '}' && len(stack) > 0 {
			f := stack[len(stack)-1]
			isLeaf := !hasChild[len(hasChild)-1]
			stack = stack[:len(stack)-1]
			hasChild = hasChild[:len(hasChild)-1]

			if isLeaf {
				result = append(result, [2]int{f.pos, i})
			}
		}
	}

	return result
}

// dotEnvResolveVar resolves the content of a ${...} expression.
//
// Supports:
//   - VAR - simple lookup
//   - VAR:=def - assign default when unset or empty
func dotEnvResolveVar(content string, vars map[string]string) string {
	name, after, hasColon := strings.Cut(content, ":")

	val, exists := dotEnvLookup(name, vars)

	if !hasColon || after == "" {
		if exists {
			return val
		}

		return ""
	}

	if len(after) == 0 || after[0] != '=' {
		// Unrecognised operator - treat as plain lookup.
		if exists {
			return val
		}

		return ""
	}

	// ${VAR:=default}
	defaultVal := after[1:]

	if exists && val != "" {
		return val
	}

	vars[name] = defaultVal
	return defaultVal
}

// dotEnvExpandSimple handles $VAR (unbraced) substitution in the remaining
// string after all ${...} forms have been processed. It scans left-to-right
// and replaces $NAME sequences where NAME matches [A-Za-z_][A-Za-z0-9_]*.
func dotEnvExpandSimple(in string, vars map[string]string) string {
	if !strings.ContainsRune(in, '$') {
		return in
	}

	var out strings.Builder
	out.Grow(len(in))

	i := 0
	for i < len(in) {
		if in[i] != '$' {
			out.WriteByte(in[i])
			i++
			continue
		}

		// '$' found - check for masked segment (already handled) or braced form.
		if i+1 < len(in) && (in[i+1] == '{' || in[i+1] == '$') {
			// Leave ${...} and $${...} remnants (shouldn't normally reach here).
			out.WriteByte(in[i])
			i++
			continue
		}

		// Collect variable name: [A-Za-z_][A-Za-z0-9_]*
		j := i + 1
		if j >= len(in) || !dotEnvIsNameStart(in[j]) {
			out.WriteByte('$')
			i++
			continue
		}

		for j < len(in) && dotEnvIsNameChar(in[j]) {
			j++
		}

		name := in[i+1 : j]
		val, _ := dotEnvLookup(name, vars)
		out.WriteString(val)
		i = j
	}

	return out.String()
}

// dotEnvLookup returns the value for name: vars map first, then os.Getenv.
func dotEnvLookup(name string, vars map[string]string) (string, bool) {
	if v, ok := vars[name]; ok {
		return v, true
	}

	v, ok := os.LookupEnv(name)
	return v, ok
}

func dotEnvIsNameStart(c byte) bool {
	return c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func dotEnvIsNameChar(c byte) bool {
	return dotEnvIsNameStart(c) || (c >= '0' && c <= '9')
}
