// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

//go:build windows

package flags

import "golang.org/x/sys/windows"

func init() {
	detectLocaleOSFallbackFunc = detectLocaleOSFallbackWindows
}

func detectLocaleOSFallbackWindows() string {
	locales, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil || len(locales) == 0 {
		return ""
	}

	for _, candidate := range locales {
		if normalized := normalizeLocale(candidate); normalized != "" {
			return normalized
		}
	}

	return ""
}
