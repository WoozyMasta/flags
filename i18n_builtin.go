// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"embed"
)

//go:embed i18n/*.json
var builtinI18nCatalogFS embed.FS

var builtinI18nCatalog = func() I18nCatalog {
	catalog, err := NewJSONCatalogDirFS(builtinI18nCatalogFS, "i18n")
	if err != nil {
		// Keep parser behavior stable even if embedded catalog is malformed.
		return nil
	}

	return catalog
}()
