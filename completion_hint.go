// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import "strings"

type completionHint uint8

const (
	completionHintAuto completionHint = iota
	completionHintNone
	completionHintFile
	completionHintDir
)

func parseCompletionHint(raw string, fieldName string) (completionHint, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "":
		return completionHintAuto, nil
	case "none":
		return completionHintNone, nil
	case "file":
		return completionHintFile, nil
	case "dir":
		return completionHintDir, nil
	default:
		return completionHintAuto, newErrorf(
			ErrInvalidTag,
			"invalid completion value `%s' for tag `%s' on field `%s' (expected file, dir, or none)",
			raw,
			FlagTagCompletion,
			fieldName,
		)
	}
}
