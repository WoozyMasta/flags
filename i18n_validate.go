// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// I18nCoverageReport describes missing translation keys for one locale.
type I18nCoverageReport struct {
	Locale                string
	MissingKeys           []string
	PlaceholderMismatches []I18nPlaceholderMismatch
}

// I18nPlaceholderMismatch describes placeholder differences for one key.
type I18nPlaceholderMismatch struct {
	Key  string
	Base []string
	Got  []string
}

// I18nCoverageConfig configures merged catalog coverage validation.
type I18nCoverageConfig struct {
	ModuleCatalog     I18nCatalog
	UserCatalog       I18nCatalog
	BaseLocale        string
	Locales           []string
	CheckPlaceholders bool
}

type i18nMessageSpec struct {
	Key          string
	Source       string
	Placeholders []string
}

var _ = builtinI18nMessageRegistry

// CheckCatalogCoverage compares locale key sets against base locale keys.
// It reports keys that are present in base locale but missing in each target locale.
// If no target locales are provided, all catalog locales except base are checked.
func CheckCatalogCoverage(catalog I18nCatalog, baseLocale string, locales ...string) ([]I18nCoverageReport, error) {
	jsonCat, ok := catalog.(*jsonCatalog)
	if !ok {
		return nil, fmt.Errorf("unsupported catalog type %T: only JSON catalogs are enumerable", catalog)
	}

	base := normalizeLocale(baseLocale)
	if base == "" {
		return nil, fmt.Errorf("invalid base locale %q", baseLocale)
	}

	baseMessages, ok := jsonCat.locales[base]
	if !ok {
		return nil, fmt.Errorf("base locale %q not found in catalog", base)
	}

	targetLocales := collectCoverageLocales(jsonCat.locales, base, locales)
	reports := make([]I18nCoverageReport, 0, len(targetLocales))

	for _, locale := range targetLocales {
		messages := jsonCat.locales[locale]
		missing := make([]string, 0)

		for key := range baseMessages {
			if _, has := messages[key]; has {
				continue
			}
			missing = append(missing, key)
		}

		slices.Sort(missing)
		reports = append(reports, I18nCoverageReport{
			Locale:      locale,
			MissingKeys: missing,
		})
	}

	return reports, nil
}

// CheckMergedCatalogCoverage compares effective locale key sets after applying
// user catalog overrides over module catalog messages.
func CheckMergedCatalogCoverage(cfg I18nCoverageConfig) ([]I18nCoverageReport, error) {
	moduleCatalog := cfg.ModuleCatalog
	if moduleCatalog == nil {
		moduleCatalog = builtinI18nCatalog
	}

	base := normalizeLocale(cfg.BaseLocale)
	if base == "" {
		return nil, fmt.Errorf("invalid base locale %q", cfg.BaseLocale)
	}

	baseMessages := mergedLocaleMessages(moduleCatalog, cfg.UserCatalog, base)
	if len(baseMessages) == 0 {
		return nil, fmt.Errorf("base locale %q not found in merged catalog", base)
	}

	targetLocales := collectMergedCoverageLocales(moduleCatalog, cfg.UserCatalog, base, cfg.Locales)
	reports := make([]I18nCoverageReport, 0, len(targetLocales))

	for _, locale := range targetLocales {
		messages := mergedLocaleMessages(moduleCatalog, cfg.UserCatalog, locale)
		missing := make([]string, 0)
		mismatches := make([]I18nPlaceholderMismatch, 0)

		for key, baseText := range baseMessages {
			text, has := messages[key]
			if !has {
				missing = append(missing, key)
				continue
			}

			if !cfg.CheckPlaceholders {
				continue
			}

			basePlaceholders := placeholderNames(baseText)
			gotPlaceholders := placeholderNames(text)
			if !slices.Equal(basePlaceholders, gotPlaceholders) {
				mismatches = append(mismatches, I18nPlaceholderMismatch{
					Key:  key,
					Base: basePlaceholders,
					Got:  gotPlaceholders,
				})
			}
		}

		slices.Sort(missing)
		slices.SortFunc(mismatches, func(a, b I18nPlaceholderMismatch) int {
			return strings.Compare(a.Key, b.Key)
		})

		reports = append(reports, I18nCoverageReport{
			Locale:                locale,
			MissingKeys:           missing,
			PlaceholderMismatches: mismatches,
		})
	}

	return reports, nil
}

func builtinI18nMessageRegistry() []i18nMessageSpec {
	specs := make([]i18nMessageSpec, 0, len(builtinI18nMessages))

	for _, spec := range builtinI18nMessages {
		specs = append(specs, i18nMessageSpec{
			Key:          spec.Key,
			Source:       spec.Source,
			Placeholders: placeholderNames(spec.Source),
		})
	}

	slices.SortFunc(specs, func(a, b i18nMessageSpec) int {
		return strings.Compare(a.Key, b.Key)
	})

	return specs
}

func mergedLocaleMessages(moduleCatalog, userCatalog I18nCatalog, locale string) map[string]string {
	out := make(map[string]string)

	if moduleMessages := localeMessages(moduleCatalog, locale); len(moduleMessages) > 0 {
		maps.Copy(out, moduleMessages)
	}
	if userMessages := localeMessages(userCatalog, locale); len(userMessages) > 0 {
		maps.Copy(out, userMessages)
	}

	return out
}

func localeMessages(catalog I18nCatalog, locale string) map[string]string {
	jsonCat, ok := catalog.(*jsonCatalog)
	if !ok || jsonCat == nil {
		return nil
	}

	messages := jsonCat.locales[locale]
	if len(messages) == 0 {
		return nil
	}

	out := make(map[string]string, len(messages))
	maps.Copy(out, messages)
	return out
}

func collectCoverageLocales(locales map[string]map[string]string, base string, requested []string) []string {
	if len(requested) == 0 {
		keys := make([]string, 0, len(locales))
		for locale := range locales {
			if locale == base {
				continue
			}
			keys = append(keys, locale)
		}
		slices.Sort(keys)
		return keys
	}

	out := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))

	for _, raw := range requested {
		normalized := normalizeLocale(raw)
		if normalized == "" || normalized == base {
			continue
		}
		if _, ok := locales[normalized]; !ok {
			continue
		}
		if _, dup := seen[normalized]; dup {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	slices.Sort(out)
	return out
}

func collectMergedCoverageLocales(moduleCatalog, userCatalog I18nCatalog, base string, requested []string) []string {
	if len(requested) > 0 {
		return collectRequestedCoverageLocales(moduleCatalog, userCatalog, base, requested)
	}

	seen := make(map[string]struct{})
	collectCatalogLocales(seen, moduleCatalog, base)
	collectCatalogLocales(seen, userCatalog, base)

	out := make([]string, 0, len(seen))
	for locale := range seen {
		out = append(out, locale)
	}
	slices.Sort(out)
	return out
}

func collectRequestedCoverageLocales(moduleCatalog, userCatalog I18nCatalog, base string, requested []string) []string {
	out := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))

	for _, raw := range requested {
		normalized := normalizeLocale(raw)
		if normalized == "" || normalized == base {
			continue
		}
		if _, dup := seen[normalized]; dup {
			continue
		}
		if len(mergedLocaleMessages(moduleCatalog, userCatalog, normalized)) == 0 {
			continue
		}

		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	slices.Sort(out)
	return out
}

func collectCatalogLocales(out map[string]struct{}, catalog I18nCatalog, base string) {
	jsonCat, ok := catalog.(*jsonCatalog)
	if !ok || jsonCat == nil {
		return
	}

	for locale := range jsonCat.locales {
		if locale == base {
			continue
		}
		out[locale] = struct{}{}
	}
}

func placeholderNames(text string) []string {
	if text == "" {
		return nil
	}

	seen := make(map[string]struct{})
	for {
		start := strings.IndexByte(text, '{')
		if start < 0 {
			break
		}

		text = text[start+1:]
		end := strings.IndexByte(text, '}')
		if end < 0 {
			break
		}

		name := strings.TrimSpace(text[:end])
		if name != "" {
			seen[name] = struct{}{}
		}

		text = text[end+1:]
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
