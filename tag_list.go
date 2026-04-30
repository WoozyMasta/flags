// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import "strings"

func splitTagListValues(raw []string, delimiter rune) []string {
	if len(raw) == 0 {
		return nil
	}

	delim := string(delimiter)
	out := make([]string, 0, len(raw))

	for _, item := range raw {
		for part := range strings.SplitSeq(item, delim) {
			out = append(out, strings.TrimSpace(part))
		}
	}

	return out
}

func parserTagListDelimiter(p *Parser) rune {
	if p != nil && p.TagListDelimiter != 0 {
		return p.TagListDelimiter
	}

	return ';'
}

func collectTagValues(
	mtag multiTag,
	singleTag string,
	listTag string,
	fieldName string,
	delimiter rune,
) ([]string, error) {
	single := mtag.GetMany(singleTag)
	list := mtag.GetMany(listTag)

	if len(single) > 0 && len(list) > 0 {
		return nil, newErrorf(
			ErrInvalidTag,
			"field `%s` cannot mix `%s` and `%s` tags",
			fieldName,
			singleTag,
			listTag,
		)
	}

	if len(list) > 0 {
		if len(list) > 1 {
			return nil, newErrorf(
				ErrInvalidTag,
				"field `%s` uses non-repeatable `%s` tag more than once",
				fieldName,
				listTag,
			)
		}
		return splitTagListValues(list, delimiter), nil
	}

	return single, nil
}

func collectDelimitedTagValues(
	mtag multiTag,
	tagName string,
	fieldName string,
	delimiter rune,
) ([]string, error) {
	values := mtag.GetMany(tagName)
	if len(values) == 0 {
		return nil, nil
	}

	if len(values) > 1 {
		return nil, newErrorf(
			ErrInvalidTag,
			"field `%s` uses non-repeatable `%s` tag more than once",
			fieldName,
			tagName,
		)
	}

	return uniqueTagListValues(splitTagListValues(values, delimiter)), nil
}

func uniqueTagListValues(values []string) []string {
	if len(values) < 2 {
		return values
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))

	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	return out
}
