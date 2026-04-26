// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"os"
	"slices"
	"strings"

	"golang.org/x/text/language"
)

// I18nCatalog resolves localized text by locale and key.
type I18nCatalog interface {
	Lookup(locale, key string) (string, bool)
}

// I18nConfig configures parser localization behavior.
type I18nConfig struct {
	ModuleCatalog   I18nCatalog
	UserCatalog     I18nCatalog
	Locale          string
	FallbackLocales []string
}

// Localizer resolves localized text from an I18nConfig without requiring a
// Parser. It is useful for application messages that share the same catalogs
// and locale chain as CLI help and parser errors.
type Localizer struct {
	i18n *i18nState
}

type i18nState struct {
	cfg     I18nConfig
	locales []string
}

// NewLocalizer creates a standalone localizer with the provided config.
func NewLocalizer(cfg I18nConfig) *Localizer {
	return &Localizer{i18n: newI18nState(cfg)}
}

// SetI18n enables parser localization with the provided config.
func (p *Parser) SetI18n(cfg I18nConfig) {
	p.i18n = newI18nState(cfg)
}

// SetI18nFallbackLocales updates i18n fallback locale chain.
// If i18n is not enabled yet, it enables i18n with default module catalog.
func (p *Parser) SetI18nFallbackLocales(locales ...string) {
	if p == nil {
		return
	}

	if p.i18n == nil {
		p.SetI18n(I18nConfig{FallbackLocales: locales})
		return
	}

	p.i18n.cfg.FallbackLocales = append([]string(nil), locales...)
	p.i18n.resetLocaleCache()
}

// DisableI18n disables parser localization.
func (p *Parser) DisableI18n() {
	p.i18n = nil
}

func (p *Parser) i18nText(key, fallback string) string {
	if p == nil || p.i18n == nil {
		return fallback
	}

	return p.i18n.text(key, fallback)
}

func (p *Parser) i18nTextf(key, fallback string, data map[string]string) string {
	return formatI18nText(p.i18nText(key, fallback), data)
}

// Localize returns localized text for key with source fallback and optional
// placeholder substitutions.
func (p *Parser) Localize(key, fallback string, data map[string]string) string {
	return p.i18nTextf(key, fallback, data)
}

// Localize returns localized text for key with source fallback and optional
// placeholder substitutions.
func (l *Localizer) Localize(key, fallback string, data map[string]string) string {
	if l == nil || l.i18n == nil {
		return formatI18nText(fallback, data)
	}

	return l.i18n.textf(key, fallback, data)
}

// LocaleChain returns resolved locale lookup order for the parser.
func (p *Parser) LocaleChain() []string {
	if p == nil || p.i18n == nil {
		return nil
	}

	return p.i18n.chainCopy()
}

// LocaleChain returns resolved locale lookup order for the localizer.
func (l *Localizer) LocaleChain() []string {
	if l == nil || l.i18n == nil {
		return nil
	}

	return l.i18n.chainCopy()
}

func newI18nState(cfg I18nConfig) *i18nState {
	if cfg.ModuleCatalog == nil {
		cfg.ModuleCatalog = builtinI18nCatalog
	}
	cfg.FallbackLocales = append([]string(nil), cfg.FallbackLocales...)

	state := &i18nState{cfg: cfg}
	state.resetLocaleCache()

	return state
}

func (s *i18nState) text(key, fallback string) string {
	if s == nil {
		return fallback
	}

	locales := s.localeChain()
	for _, locale := range locales {
		if text, ok := lookupI18nCatalog(s.cfg.UserCatalog, locale, key); ok {
			return text
		}
		if text, ok := lookupI18nCatalog(s.cfg.ModuleCatalog, locale, key); ok {
			return text
		}
	}

	return fallback
}

func (s *i18nState) textf(key, fallback string, data map[string]string) string {
	return formatI18nText(s.text(key, fallback), data)
}

func (s *i18nState) chainCopy() []string {
	chain := s.localeChain()
	if len(chain) == 0 {
		return nil
	}

	return append([]string(nil), chain...)
}

func formatI18nText(template string, data map[string]string) string {
	if len(data) == 0 {
		return template
	}

	for k, v := range data {
		template = strings.ReplaceAll(template, "{"+k+"}", v)
	}

	return template
}

func lookupI18nCatalog(catalog I18nCatalog, locale, key string) (string, bool) {
	if catalog == nil || locale == "" || key == "" {
		return "", false
	}

	value, ok := catalog.Lookup(locale, key)
	if !ok || value == "" {
		return "", false
	}

	return value, true
}

func (s *i18nState) localeChain() []string {
	if s == nil {
		return nil
	}

	if s.locales != nil {
		return s.locales
	}

	return s.buildLocaleChain()
}

func (s *i18nState) resetLocaleCache() {
	if s == nil {
		return
	}

	if s.cfg.Locale == "" {
		s.locales = nil
		return
	}

	s.locales = s.buildLocaleChain()
}

func (s *i18nState) buildLocaleChain() []string {
	var locales []string
	appendLocale := func(candidate string) {
		normalized := normalizeLocale(candidate)
		if normalized == "" {
			return
		}

		if !strings.Contains(normalized, "-") {
			locales = appendUnique(locales, normalized)
			return
		}

		locales = appendUnique(locales, normalized)
		if base := baseLocale(normalized); base != "" {
			locales = appendUnique(locales, base)
		}
	}

	if s.cfg.Locale != "" {
		appendLocale(s.cfg.Locale)
	} else if detected := DetectLocale(); detected != "" {
		appendLocale(detected)
	}

	for _, fallback := range s.cfg.FallbackLocales {
		appendLocale(fallback)
	}

	appendLocale("en")

	return locales
}

func appendUnique(items []string, item string) []string {
	if slices.Contains(items, item) {
		return items
	}

	return append(items, item)
}

func normalizeLocale(raw string) string {
	cleaned := cleanLocaleToken(raw)
	if cleaned == "" {
		return ""
	}

	tag, err := language.Parse(cleaned)
	if err != nil {
		return ""
	}

	return tag.String()
}

func baseLocale(locale string) string {
	if locale == "" {
		return ""
	}

	tag, err := language.Parse(locale)
	if err != nil {
		return ""
	}

	base, confidence := tag.Base()
	if confidence == language.No || base.String() == "und" {
		return ""
	}

	return base.String()
}

func cleanLocaleToken(raw string) string {
	token := strings.TrimSpace(raw)
	if token == "" {
		return ""
	}

	if idx := strings.IndexByte(token, ':'); idx >= 0 {
		token = token[:idx]
	}
	if idx := strings.IndexByte(token, '.'); idx >= 0 {
		token = token[:idx]
	}
	if idx := strings.IndexByte(token, '@'); idx >= 0 {
		token = token[:idx]
	}

	token = strings.ReplaceAll(token, "_", "-")
	token = strings.TrimSpace(token)

	return token
}

// DetectLocale returns detected locale using environment variables and
// OS-specific fallback.
func DetectLocale() string {
	if locale := detectLocaleFromEnv(); locale != "" {
		return locale
	}

	if locale := detectLocaleOSFallbackFunc(); locale != "" {
		return locale
	}

	return ""
}

func detectLocaleFromEnv() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		value, ok := os.LookupEnv(key)
		if !ok {
			continue
		}

		if key == "LANGUAGE" {
			for candidate := range strings.SplitSeq(value, ":") {
				if normalized := normalizeLocale(candidate); normalized != "" {
					return normalized
				}
			}
			continue
		}

		if normalized := normalizeLocale(value); normalized != "" {
			return normalized
		}
	}

	return ""
}

var detectLocaleOSFallbackFunc = func() string {
	return ""
}
