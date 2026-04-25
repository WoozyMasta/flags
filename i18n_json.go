// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

var (
	// ErrI18nInvalidJSONCatalog indicates malformed/unsupported catalog format.
	ErrI18nInvalidJSONCatalog = errors.New("invalid i18n json catalog format")
)

type jsonCatalog struct {
	locales map[string]map[string]string
}

// NewJSONCatalog parses JSON catalog bytes into an I18nCatalog.
func NewJSONCatalog(data []byte) (I18nCatalog, error) {
	return newJSONCatalogWithHint(data, "")
}

// NewJSONCatalogFS parses JSON catalog from fs path into an I18nCatalog.
func NewJSONCatalogFS(fsys fs.FS, path string) (I18nCatalog, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read i18n catalog %q: %w", path, err)
	}

	return newJSONCatalogWithHint(data, inferLocaleFromPath(path))
}

// NewJSONCatalogDirFS loads and merges all .json catalogs from the given fs directory.
// Files are processed in lexical order for deterministic conflict resolution.
func NewJSONCatalogDirFS(fsys fs.FS, dir string) (I18nCatalog, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("read i18n catalog dir %q: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.EqualFold(filepath.Ext(name), ".json") {
			continue
		}

		files = append(files, name)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("%w: no json catalogs found in %q", ErrI18nInvalidJSONCatalog, dir)
	}

	slices.Sort(files)
	locales := make(map[string]map[string]string)

	for _, file := range files {
		filePath := path.Join(dir, file)
		catalog, catalogErr := NewJSONCatalogFS(fsys, filePath)
		if catalogErr != nil {
			return nil, catalogErr
		}

		jsonCat, ok := catalog.(*jsonCatalog)
		if !ok {
			return nil, fmt.Errorf("%w: unexpected catalog type for %q", ErrI18nInvalidJSONCatalog, filePath)
		}

		for locale, messages := range jsonCat.locales {
			dst := locales[locale]
			if dst == nil {
				dst = make(map[string]string, len(messages))
				locales[locale] = dst
			}

			maps.Copy(dst, messages)
		}
	}

	return &jsonCatalog{locales: locales}, nil
}

func (c *jsonCatalog) Lookup(locale, key string) (string, bool) {
	if c == nil || key == "" {
		return "", false
	}

	if locale != "" {
		if m := c.locales[locale]; m != nil {
			value, ok := m[key]
			return value, ok
		}
	}

	return "", false
}

func newJSONCatalogWithHint(data []byte, localeHint string) (I18nCatalog, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrI18nInvalidJSONCatalog, err)
	}

	root, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: root must be object", ErrI18nInvalidJSONCatalog)
	}

	locales := make(map[string]map[string]string)

	// Shape A: {"locale":"en","messages":{"k":"v"}}
	if localeValue, hasLocale := root["locale"]; hasLocale {
		locale, ok := localeValue.(string)
		if !ok {
			return nil, fmt.Errorf("%w: locale must be string", ErrI18nInvalidJSONCatalog)
		}

		messageValue, ok := root["messages"]
		if !ok {
			return nil, fmt.Errorf("%w: messages field is required with locale", ErrI18nInvalidJSONCatalog)
		}

		messages, err := anyToStringMap(messageValue)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrI18nInvalidJSONCatalog, err)
		}

		normalized := normalizeLocale(locale)
		if normalized == "" {
			return nil, fmt.Errorf("%w: invalid locale %q", ErrI18nInvalidJSONCatalog, locale)
		}

		locales[normalized] = messages
		return &jsonCatalog{locales: locales}, nil
	}

	// Shape B: {"en":{"k":"v"},"ru":{"k":"v"}}
	multiLocale := true
	for _, value := range root {
		_, ok := value.(map[string]any)
		if !ok {
			multiLocale = false
			break
		}
	}

	if multiLocale {
		for locale, value := range root {
			normalized := normalizeLocale(locale)
			if normalized == "" {
				return nil, fmt.Errorf("%w: invalid locale key %q", ErrI18nInvalidJSONCatalog, locale)
			}

			messages, err := anyToStringMap(value)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrI18nInvalidJSONCatalog, err)
			}
			locales[normalized] = messages
		}

		return &jsonCatalog{locales: locales}, nil
	}

	// Shape C: {"k":"v"} with locale inferred from filename in FS loader.
	messages, err := anyToStringMap(root)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrI18nInvalidJSONCatalog, err)
	}

	normalizedHint := normalizeLocale(localeHint)
	if normalizedHint == "" {
		return nil, fmt.Errorf("%w: flat message map requires locale context", ErrI18nInvalidJSONCatalog)
	}

	locales[normalizedHint] = messages
	return &jsonCatalog{locales: locales}, nil
}

func anyToStringMap(v any) (map[string]string, error) {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("expected object with string values")
	}

	out := make(map[string]string, len(obj))
	for key, rawValue := range obj {
		switch value := rawValue.(type) {
		case string:
			out[key] = value
		case map[string]any:
			text := firstString(
				value,
				"other",
				"translation",
				"message",
				"text",
			)
			if text == "" {
				return nil, fmt.Errorf("value for key %q object has no supported text fields", key)
			}
			out[key] = text
		default:
			return nil, fmt.Errorf("value for key %q must be string or object", key)
		}
	}

	return out, nil
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		raw, ok := values[key]
		if !ok {
			continue
		}

		text, ok := raw.(string)
		if !ok || text == "" {
			continue
		}

		return text
	}

	return ""
}

func inferLocaleFromPath(path string) string {
	file := filepath.Base(path)
	ext := filepath.Ext(file)
	if ext == "" {
		return ""
	}

	name := strings.TrimSuffix(file, ext)
	if name == "" {
		return ""
	}

	parts := strings.Split(name, ".")
	candidate := parts[len(parts)-1]

	return normalizeLocale(candidate)
}
