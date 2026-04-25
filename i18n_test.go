// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

type mapCatalog map[string]map[string]string

func (m mapCatalog) Lookup(locale, key string) (string, bool) {
	if keys, ok := m[locale]; ok {
		value, ok := keys[key]
		return value, ok
	}

	return "", false
}

func TestI18nResolverOrderUserBeatsModule(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{
		Locale: "ru-RU",
		UserCatalog: mapCatalog{
			"ru": {"help.usage": "Пользовательское использование"},
		},
		ModuleCatalog: mapCatalog{
			"ru": {"help.usage": "Использование"},
		},
	})

	got := parser.i18nText("help.usage", "Usage")
	if got != "Пользовательское использование" {
		t.Fatalf("unexpected localized text: %q", got)
	}
}

func TestI18nResolverOrderUsesBaseAndFallback(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{
		Locale:          "it-IT",
		FallbackLocales: []string{"de-DE", "en"},
		ModuleCatalog: mapCatalog{
			"de": {"help.usage": "Verwendung"},
		},
	})

	got := parser.i18nText("help.usage", "Usage")
	if got != "Verwendung" {
		t.Fatalf("unexpected fallback localized text: %q", got)
	}
}

func TestI18nResolverFallsBackToSourceText(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{
		Locale:        "ru-RU",
		ModuleCatalog: mapCatalog{},
		UserCatalog:   mapCatalog{},
	})

	got := parser.i18nText("help.usage", "Usage")
	if got != "Usage" {
		t.Fatalf("expected source fallback, got %q", got)
	}
}

func TestSetI18nUsesBuiltinModuleCatalogByDefault(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{Locale: "ru-RU"})

	got := parser.i18nText("help.usage", "Usage")
	if got != "Использование" {
		t.Fatalf("unexpected built-in catalog value: %q", got)
	}
}

func TestDetectLocaleFromEnvPriority(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	_ = os.Setenv("LC_ALL", "")
	_ = os.Setenv("LC_MESSAGES", "de_DE.UTF-8")
	_ = os.Setenv("LANG", "ru_RU.UTF-8")
	_ = os.Setenv("LANGUAGE", "fr:en")

	got := detectLocaleFromEnv()
	if got != "de-DE" {
		t.Fatalf("expected LC_MESSAGES locale, got %q", got)
	}

	_ = os.Setenv("LC_MESSAGES", "")
	got = detectLocaleFromEnv()
	if got != "ru-RU" {
		t.Fatalf("expected LANG locale, got %q", got)
	}

	_ = os.Setenv("LANG", "")
	got = detectLocaleFromEnv()
	if got != "fr" {
		t.Fatalf("expected LANGUAGE first locale, got %q", got)
	}
}

func TestNormalizeLocale(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "en_US.UTF-8", want: "en-US"},
		{raw: "ru_RU@latin", want: "ru-RU"},
		{raw: "zh-Hant_TW", want: "zh-Hant-TW"},
		{raw: "___", want: ""},
	}

	for _, test := range tests {
		got := normalizeLocale(test.raw)
		if got != test.want {
			t.Fatalf("normalizeLocale(%q) = %q, want %q", test.raw, got, test.want)
		}
	}
}

func TestI18nDefaultFallbackLocaleIsEnglish(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{
		Locale: "it-IT",
		ModuleCatalog: mapCatalog{
			"en": {"help.usage": "Usage"},
		},
	})

	got := parser.i18nText("help.usage", "Usage fallback")
	if got != "Usage" {
		t.Fatalf("expected default en fallback, got %q", got)
	}
}

func TestParserLocaleChainUsesConfiguredAndFallbackLocales(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{
		Locale:          "ru_RU.UTF-8",
		FallbackLocales: []string{"de-DE", "en"},
	})

	got := parser.LocaleChain()
	want := []string{"ru-RU", "ru", "de-DE", "de", "en"}

	if len(got) != len(want) {
		t.Fatalf("unexpected locale chain len: got=%v want=%v", got, want)
	}

	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected locale chain at %d: got=%q want=%q full=%v", idx, got[idx], want[idx], got)
		}
	}
}

func TestParserLocalizeResolvesUserCatalogAndFormatsData(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{
		Locale: "ru",
		UserCatalog: mapCatalog{
			"ru": {"app.greeting": "Привет, {target}"},
		},
	})

	got := parser.Localize("app.greeting", "Hello, {target}", map[string]string{"target": "мир"})
	if got != "Привет, мир" {
		t.Fatalf("unexpected localized formatted text: %q", got)
	}
}

func TestLocalizerResolvesUserCatalogAndFormatsData(t *testing.T) {
	localizer := NewLocalizer(I18nConfig{
		Locale: "ru-RU",
		UserCatalog: mapCatalog{
			"ru": {"app.greeting": "Привет, {target}"},
		},
	})

	got := localizer.Localize("app.greeting", "Hello, {target}", map[string]string{"target": "мир"})
	if got != "Привет, мир" {
		t.Fatalf("unexpected localized formatted text: %q", got)
	}
}

func TestLocalizerLocaleChainReturnsCopy(t *testing.T) {
	localizer := NewLocalizer(I18nConfig{
		Locale:          "it-IT",
		FallbackLocales: []string{"ru"},
	})

	got := localizer.LocaleChain()
	want := []string{"it-IT", "it", "ru", "en"}

	if len(got) != len(want) {
		t.Fatalf("unexpected locale chain len: got=%v want=%v", got, want)
	}

	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected locale chain at %d: got=%q want=%q full=%v", idx, got[idx], want[idx], got)
		}
	}

	got[0] = "mutated"
	if chain := localizer.LocaleChain(); chain[0] != "it-IT" {
		t.Fatalf("locale chain was mutated through returned slice: %v", chain)
	}
}

func TestSetI18nFallbackLocalesUpdatesLocaleChain(t *testing.T) {
	parser := NewNamedParser("i18n", None)
	parser.SetI18n(I18nConfig{Locale: "it-IT"})
	parser.SetI18nFallbackLocales("ru-RU", "de")

	got := parser.LocaleChain()
	want := []string{"it-IT", "it", "ru-RU", "ru", "de", "en"}

	if len(got) != len(want) {
		t.Fatalf("unexpected locale chain len: got=%v want=%v", got, want)
	}

	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected locale chain at %d: got=%q want=%q full=%v", idx, got[idx], want[idx], got)
		}
	}
}

func TestSetI18nFallbackLocalesEnablesI18nWhenDisabled(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()
	_ = os.Setenv("LC_ALL", "")
	_ = os.Setenv("LC_MESSAGES", "")
	_ = os.Setenv("LANG", "")
	_ = os.Setenv("LANGUAGE", "")

	oldOSFallback := detectLocaleOSFallbackFunc
	defer func() { detectLocaleOSFallbackFunc = oldOSFallback }()
	detectLocaleOSFallbackFunc = func() string { return "" }

	parser := NewNamedParser("i18n", None)
	parser.SetI18nFallbackLocales("ru")

	got := parser.i18nText("help.usage", "Usage fallback")
	if got != "Использование" {
		t.Fatalf("expected built-in fallback locale translation, got %q", got)
	}
}

func TestNewJSONCatalogMultiLocale(t *testing.T) {
	catalog, err := NewJSONCatalog([]byte(`{
		"en": {"help.usage": "Usage"},
		"ru": {"help.usage": "Использование"}
	}`))
	if err != nil {
		t.Fatalf("unexpected NewJSONCatalog error: %v", err)
	}

	value, ok := catalog.Lookup("ru", "help.usage")
	if !ok || value != "Использование" {
		t.Fatalf("unexpected lookup result: value=%q ok=%v", value, ok)
	}
}

func TestNewJSONCatalogFSFlatMapWithPathLocale(t *testing.T) {
	fsys := fstest.MapFS{
		"i18n/en.json": &fstest.MapFile{Data: []byte(`{"help.usage":"Usage"}`)},
	}

	catalog, err := NewJSONCatalogFS(fsys, "i18n/en.json")
	if err != nil {
		t.Fatalf("unexpected NewJSONCatalogFS error: %v", err)
	}

	value, ok := catalog.Lookup("en", "help.usage")
	if !ok || value != "Usage" {
		t.Fatalf("unexpected lookup result: value=%q ok=%v", value, ok)
	}
}

func TestNewJSONCatalogAcceptsObjectValues(t *testing.T) {
	catalog, err := NewJSONCatalog([]byte(`{
		"en": {
			"help.usage": { "other": "Usage" },
			"help.arguments": { "translation": "Arguments" }
		}
	}`))
	if err != nil {
		t.Fatalf("unexpected NewJSONCatalog error: %v", err)
	}

	gotUsage, ok := catalog.Lookup("en", "help.usage")
	if !ok || gotUsage != "Usage" {
		t.Fatalf("unexpected usage lookup result: value=%q ok=%v", gotUsage, ok)
	}

	gotArgs, ok := catalog.Lookup("en", "help.arguments")
	if !ok || gotArgs != "Arguments" {
		t.Fatalf("unexpected arguments lookup result: value=%q ok=%v", gotArgs, ok)
	}
}

func TestNewJSONCatalogDirFSMergesLocales(t *testing.T) {
	fsys := fstest.MapFS{
		"i18n/en.json": &fstest.MapFile{Data: []byte(`{"help.usage":"Usage","help.arguments":"Arguments"}`)},
		"i18n/ru.json": &fstest.MapFile{Data: []byte(`{"help.usage":"Использование"}`)},
	}

	catalog, err := NewJSONCatalogDirFS(fsys, "i18n")
	if err != nil {
		t.Fatalf("unexpected NewJSONCatalogDirFS error: %v", err)
	}

	if value, ok := catalog.Lookup("en", "help.usage"); !ok || value != "Usage" {
		t.Fatalf("unexpected en lookup result: value=%q ok=%v", value, ok)
	}

	if value, ok := catalog.Lookup("ru", "help.usage"); !ok || value != "Использование" {
		t.Fatalf("unexpected ru lookup result: value=%q ok=%v", value, ok)
	}
}

func TestNewJSONCatalogDirFSErrorsOnEmptyDir(t *testing.T) {
	fsys := fstest.MapFS{
		"i18n/.keep": &fstest.MapFile{Data: []byte("")},
	}

	_, err := NewJSONCatalogDirFS(fsys, "i18n")
	if err == nil {
		t.Fatalf("expected error for empty i18n dir")
	}
}

func TestHelpUsesLocalizedBuiltinsWhenI18nEnabled(t *testing.T) {
	var opts struct {
		Value string `long:"value" default:"demo" description:"Desc"`
	}

	parser := NewNamedParser("demo", None)
	if _, err := parser.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("add group error: %v", err)
	}

	parser.SetI18n(I18nConfig{Locale: "ru"})

	var out bytes.Buffer
	parser.WriteHelp(&out)
	got := out.String()

	if !strings.Contains(got, "Использование:") {
		t.Fatalf("expected localized usage label, got:\n%s", got)
	}
	if !strings.Contains(got, "по умолчанию: demo") {
		t.Fatalf("expected localized default marker, got:\n%s", got)
	}
}

func TestRequiredFlagErrorUsesLocalizedBuiltinsWhenI18nEnabled(t *testing.T) {
	var opts struct {
		Required string `long:"required" required:"true"`
	}

	parser := NewNamedParser("demo", None)
	if _, err := parser.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("add group error: %v", err)
	}
	parser.SetI18n(I18nConfig{Locale: "ru"})

	_, err := parser.ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if flagsErr.Type != ErrRequired {
		t.Fatalf("expected ErrRequired, got %v", flagsErr.Type)
	}

	if !strings.Contains(flagsErr.Message, "обязательный флаг") {
		t.Fatalf("expected localized required-flag message, got %q", flagsErr.Message)
	}
}

func TestHelpUsesLocalizedUserTagsWhenI18nEnabled(t *testing.T) {
	var opts struct {
		Mode  string `long:"mode" value-name:"MODE" value-name-i18n:"opt.mode.value" description:"Mode fallback" description-i18n:"opt.mode.desc"`
		Group struct {
			Level int `long:"level" description:"Level fallback" description-i18n:"opt.level.desc"`
		} `group:"General" ini-group:"general" group-i18n:"group.general"`
	}

	parser := NewNamedParser("demo", None)
	if _, err := parser.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("add group error: %v", err)
	}

	parser.SetI18n(I18nConfig{
		Locale: "ru",
		UserCatalog: mapCatalog{
			"ru": {
				"opt.mode.desc":  "Режим работы",
				"opt.mode.value": "РЕЖИМ",
				"opt.level.desc": "Уровень детализации",
				"group.general":  "Общие настройки",
			},
		},
	})

	var out bytes.Buffer
	parser.WriteHelp(&out)
	got := out.String()

	if !strings.Contains(got, "Режим работы") {
		t.Fatalf("expected localized option description, got:\n%s", got)
	}

	if !strings.Contains(got, "РЕЖИМ") {
		t.Fatalf("expected localized value placeholder, got:\n%s", got)
	}

	if !strings.Contains(got, "Уровень детализации") {
		t.Fatalf("expected localized subgroup option description, got:\n%s", got)
	}

	if !strings.Contains(got, "Общие настройки:") {
		t.Fatalf("expected localized group header, got:\n%s", got)
	}
}

func TestIniExampleUsesLocalizedBuiltinsWhenI18nEnabled(t *testing.T) {
	var opts struct {
		Required string         `long:"required" required:"true" description:"Primary required value"`
		Mode     string         `long:"mode" description:"Execution mode" choices:"fast;safe"`
		Labels   map[string]int `long:"label" description:"Label map" key-value-delimiter:"="`
	}

	parser := NewNamedParser("demo", None)
	if _, err := parser.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("add group error: %v", err)
	}
	parser.SetI18n(I18nConfig{Locale: "ru"})

	if _, err := parser.ParseArgs([]string{"--required=req", "--mode=safe", "--label", "a=2"}); err != nil {
		t.Fatalf("parse args error: %v", err)
	}

	inip := NewIniParser(parser)
	var out bytes.Buffer
	inip.WriteExample(&out)
	got := out.String()

	if !strings.Contains(got, "; Primary required value (обязательный)") {
		t.Fatalf("expected localized required marker, got:\n%s", got)
	}

	if !strings.Contains(got, "; Допустимые значения: fast, safe.") {
		t.Fatalf("expected localized choices label, got:\n%s", got)
	}

	if !strings.Contains(got, "Детали: повторяемый;") || !strings.Contains(got, "разделитель: \"=\".") {
		t.Fatalf("expected localized details text, got:\n%s", got)
	}
}

func TestProgrammaticI18nKeySetters(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" value-name:"MODE" description:"Mode fallback"`
	}

	parser := NewNamedParser("demo", None)
	group, err := parser.AddGroup("Application Options", "", &opts)
	if err != nil {
		t.Fatalf("add group error: %v", err)
	}

	group.SetShortDescriptionI18nKey("group.app")

	modeOption := parser.FindOptionByLongName("mode")
	if modeOption == nil {
		t.Fatalf("expected mode option")
	}

	modeOption.SetDescriptionI18nKey("opt.mode.desc")
	modeOption.SetValueNameI18nKey("opt.mode.value")

	parser.SetI18n(I18nConfig{
		Locale: "ru",
		UserCatalog: mapCatalog{
			"ru": {
				"group.app":      "Настройки",
				"opt.mode.desc":  "Режим",
				"opt.mode.value": "РЕЖИМ",
			},
		},
	})

	var out bytes.Buffer
	parser.WriteHelp(&out)
	got := out.String()

	if !strings.Contains(got, "Настройки:") {
		t.Fatalf("expected localized group header, got:\n%s", got)
	}

	if !strings.Contains(got, "Режим") {
		t.Fatalf("expected localized option description, got:\n%s", got)
	}

	if !strings.Contains(got, "РЕЖИМ") {
		t.Fatalf("expected localized value name, got:\n%s", got)
	}
}

func TestBuiltinHelpGroupAndDescriptionsAreLocalized(t *testing.T) {
	var opts struct {
		Locale string `short:"l" long:"locale" choices:"en;ru;eo" description:"Override locale"`
	}

	parser := NewNamedParser("i18n-demo", Default|VersionFlag)
	group, err := parser.AddGroup("Application Options", "", &opts)
	if err != nil {
		t.Fatalf("add group error: %v", err)
	}
	group.SetShortDescriptionI18nKey("help.group.application_options")
	parser.SetI18n(I18nConfig{Locale: "ru"})

	var out bytes.Buffer
	parser.WriteHelp(&out)
	got := out.String()

	if !strings.Contains(got, "Опции приложения:") {
		t.Fatalf("expected localized application group, got:\n%s", got)
	}
	if !strings.Contains(got, "Опции справки:") {
		t.Fatalf("expected localized help group, got:\n%s", got)
	}
	if !strings.Contains(got, "Показать это сообщение справки") {
		t.Fatalf("expected localized help option description, got:\n%s", got)
	}
	if !strings.Contains(got, "Показать информацию о версии") {
		t.Fatalf("expected localized version option description, got:\n%s", got)
	}
	if !strings.Contains(got, "допустимые значения:") {
		t.Fatalf("expected localized valid-values label, got:\n%s", got)
	}
}

func TestWriteVersionUsesLocalizedLabels(t *testing.T) {
	parser := NewNamedParser("i18n-version", None)
	parser.SetI18n(I18nConfig{Locale: "ru"})

	var out bytes.Buffer
	parser.WriteVersion(&out, VersionFieldVersion|VersionFieldCommit|VersionFieldPath)
	got := out.String()

	if !strings.Contains(got, "версия:") {
		t.Fatalf("expected localized version label, got:\n%s", got)
	}
	if !strings.Contains(got, "коммит:") {
		t.Fatalf("expected localized commit label, got:\n%s", got)
	}
	if !strings.Contains(got, "путь:") {
		t.Fatalf("expected localized path label, got:\n%s", got)
	}
}

func TestMarshalErrorUsesLocalizedBuiltinsWhenI18nEnabled(t *testing.T) {
	var opts struct {
		Port int `long:"port"`
	}

	parser := NewNamedParser("demo", None)
	if _, err := parser.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("add group error: %v", err)
	}
	parser.SetI18n(I18nConfig{Locale: "ru"})

	_, err := parser.ParseArgs([]string{"--port", "bad"})
	if err == nil {
		t.Fatalf("expected parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if flagsErr.Type != ErrMarshal {
		t.Fatalf("expected ErrMarshal, got %v", flagsErr.Type)
	}

	if !strings.Contains(flagsErr.Message, "недопустимый аргумент для флага") {
		t.Fatalf("expected localized marshal error, got %q", flagsErr.Message)
	}
}

func TestInvalidChoiceErrorUsesLocalizedChoiceList(t *testing.T) {
	var opts struct {
		Mode string `long:"mode" choice:"fast" choice:"safe"`
	}

	parser := NewNamedParser("demo", None)
	if _, err := parser.AddGroup("Application Options", "", &opts); err != nil {
		t.Fatalf("add group error: %v", err)
	}
	parser.SetI18n(I18nConfig{Locale: "ru"})

	_, err := parser.ParseArgs([]string{"--mode", "bad"})
	if err == nil {
		t.Fatalf("expected parse error")
	}

	flagsErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if flagsErr.Type != ErrInvalidChoice {
		t.Fatalf("expected ErrInvalidChoice, got %v", flagsErr.Type)
	}

	if !strings.Contains(flagsErr.Message, "Допустимые значения: fast или safe") {
		t.Fatalf("expected localized choice list, got %q", flagsErr.Message)
	}
}

func TestCheckCatalogCoverageReportsMissingKeys(t *testing.T) {
	catalog, err := NewJSONCatalog([]byte(`{
		"en": {
			"a": "A",
			"b": "B",
			"c": "C"
		},
		"ru": {
			"a": "А"
		},
		"eo": {
			"a": "A",
			"b": "B"
		}
	}`))
	if err != nil {
		t.Fatalf("unexpected NewJSONCatalog error: %v", err)
	}

	reports, err := CheckCatalogCoverage(catalog, "en")
	if err != nil {
		t.Fatalf("unexpected CheckCatalogCoverage error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf("unexpected report len: %d", len(reports))
	}

	if reports[0].Locale != "eo" || len(reports[0].MissingKeys) != 1 || reports[0].MissingKeys[0] != "c" {
		t.Fatalf("unexpected eo report: %+v", reports[0])
	}

	if reports[1].Locale != "ru" || len(reports[1].MissingKeys) != 2 || reports[1].MissingKeys[0] != "b" || reports[1].MissingKeys[1] != "c" {
		t.Fatalf("unexpected ru report: %+v", reports[1])
	}
}

func TestCheckCatalogCoverageRequestedLocales(t *testing.T) {
	catalog, err := NewJSONCatalog([]byte(`{
		"en": {"a": "A", "b": "B"},
		"ru": {"a": "А"},
		"de": {"a": "A", "b": "B"}
	}`))
	if err != nil {
		t.Fatalf("unexpected NewJSONCatalog error: %v", err)
	}

	reports, err := CheckCatalogCoverage(catalog, "en", "ru", "missing", "de")
	if err != nil {
		t.Fatalf("unexpected CheckCatalogCoverage error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf("unexpected report len: %d", len(reports))
	}

	if reports[0].Locale != "de" || len(reports[0].MissingKeys) != 0 {
		t.Fatalf("unexpected de report: %+v", reports[0])
	}

	if reports[1].Locale != "ru" || len(reports[1].MissingKeys) != 1 || reports[1].MissingKeys[0] != "b" {
		t.Fatalf("unexpected ru report: %+v", reports[1])
	}
}

func TestCheckMergedCatalogCoverageUsesUserOverrides(t *testing.T) {
	moduleCatalog, err := NewJSONCatalog([]byte(`{
		"en": {"a": "A", "b": "B"},
		"ru": {"a": "А"}
	}`))
	if err != nil {
		t.Fatalf("unexpected module catalog error: %v", err)
	}

	userCatalog, err := NewJSONCatalog([]byte(`{
		"ru": {"b": "Б"}
	}`))
	if err != nil {
		t.Fatalf("unexpected user catalog error: %v", err)
	}

	reports, err := CheckMergedCatalogCoverage(I18nCoverageConfig{
		BaseLocale:    "en",
		Locales:       []string{"ru"},
		ModuleCatalog: moduleCatalog,
		UserCatalog:   userCatalog,
	})
	if err != nil {
		t.Fatalf("unexpected CheckMergedCatalogCoverage error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("unexpected report len: %d", len(reports))
	}

	if reports[0].Locale != "ru" || len(reports[0].MissingKeys) != 0 {
		t.Fatalf("unexpected merged ru report: %+v", reports[0])
	}
}

func TestCheckMergedCatalogCoverageReportsPlaceholderMismatches(t *testing.T) {
	catalog, err := NewJSONCatalog([]byte(`{
		"en": {"err.open": "open {path}: {reason}"},
		"ru": {"err.open": "открыть {path}: {cause}"}
	}`))
	if err != nil {
		t.Fatalf("unexpected catalog error: %v", err)
	}

	reports, err := CheckMergedCatalogCoverage(I18nCoverageConfig{
		BaseLocale:        "en",
		Locales:           []string{"ru"},
		ModuleCatalog:     catalog,
		CheckPlaceholders: true,
	})
	if err != nil {
		t.Fatalf("unexpected CheckMergedCatalogCoverage error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("unexpected report len: %d", len(reports))
	}

	mismatches := reports[0].PlaceholderMismatches
	if len(mismatches) != 1 {
		t.Fatalf("unexpected placeholder mismatches: %+v", reports[0])
	}

	if mismatches[0].Key != "err.open" {
		t.Fatalf("unexpected mismatch key: %+v", mismatches[0])
	}

	if strings.Join(mismatches[0].Base, ",") != "path,reason" {
		t.Fatalf("unexpected base placeholders: %+v", mismatches[0])
	}

	if strings.Join(mismatches[0].Got, ",") != "cause,path" {
		t.Fatalf("unexpected translated placeholders: %+v", mismatches[0])
	}
}

func TestBuiltinI18nMessageRegistryMatchesEnglishCatalog(t *testing.T) {
	specs := builtinI18nMessageRegistry()
	if len(specs) == 0 {
		t.Fatalf("expected built-in i18n message registry")
	}

	enMessages := mergedLocaleMessages(builtinI18nCatalog, nil, "en")
	if len(specs) != len(enMessages) {
		t.Fatalf("registry len %d does not match en catalog len %d", len(specs), len(enMessages))
	}

	for _, spec := range specs {
		source, ok := enMessages[spec.Key]
		if !ok {
			t.Fatalf("registry key %q missing from en catalog", spec.Key)
		}
		if spec.Source != source {
			t.Fatalf("registry source mismatch for %q: got %q want %q", spec.Key, spec.Source, source)
		}
		if strings.Join(spec.Placeholders, ",") != strings.Join(placeholderNames(source), ",") {
			t.Fatalf("registry placeholder mismatch for %q", spec.Key)
		}
	}

	registryKeys := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		if _, duplicate := registryKeys[spec.Key]; duplicate {
			t.Fatalf("duplicate registry key %q", spec.Key)
		}
		registryKeys[spec.Key] = struct{}{}
	}

	for key := range enMessages {
		if _, ok := registryKeys[key]; !ok {
			t.Fatalf("en catalog key %q missing from registry", key)
		}
	}
}

func TestBuiltinCatalogCoverageAndPlaceholders(t *testing.T) {
	reports, err := CheckMergedCatalogCoverage(I18nCoverageConfig{
		BaseLocale:        "en",
		CheckPlaceholders: true,
	})
	if err != nil {
		t.Fatalf("unexpected CheckMergedCatalogCoverage error: %v", err)
	}

	for _, report := range reports {
		if len(report.MissingKeys) == 0 && len(report.PlaceholderMismatches) == 0 {
			continue
		}

		t.Fatalf("unexpected built-in i18n coverage gaps for %s: missing=%v placeholder mismatches=%+v",
			report.Locale,
			report.MissingKeys,
			report.PlaceholderMismatches,
		)
	}
}

func TestBuiltinNonEnglishCatalogsTranslateRuntimeHelpKeys(t *testing.T) {
	keys := []string{
		"err.list.disjunction",
		"err.marshal.argument_default",
		"err.marshal.expected",
		"err.marshal.option",
		"help.builtin.show_help",
		"help.builtin.show_version",
		"help.group.application_options",
		"help.group.help_options",
		"help.meta.valid_values",
		"ini.err.unknown_group",
		"ini.err.unknown_option",
	}

	enMessages := mergedLocaleMessages(builtinI18nCatalog, nil, "en")
	for _, locale := range []string{"ru", "de", "it", "fr", "zh", "es", "uk", "kk", "cs", "pl", "ja", "pt", "hi", "id", "ko", "tr", "vi"} {
		messages := mergedLocaleMessages(builtinI18nCatalog, nil, locale)
		if len(messages) == 0 {
			t.Fatalf("expected built-in messages for locale %q", locale)
		}

		for _, key := range keys {
			if messages[key] == enMessages[key] {
				t.Fatalf("locale %q leaves %q as English source text %q", locale, key, messages[key])
			}
		}
	}
}

func TestBuiltinDocLabelsAvoidObviousEnglishFallbacks(t *testing.T) {
	checks := map[string][]string{
		"cs": {
			"doc.tmpl.man.label.base",
			"doc.tmpl.man.label.terminator",
			"doc.tmpl.markdown.label.base",
			"doc.tmpl.markdown.label.base_lower",
			"doc.tmpl.markdown.label.terminator",
			"doc.tmpl.markdown.label.terminator_lower",
		},
		"ja": {
			"doc.tmpl.man.label.base",
			"doc.tmpl.markdown.label.base",
			"doc.tmpl.markdown.label.base_lower",
			"doc.tmpl.markdown.label.terminator_lower",
		},
		"kk": {
			"doc.tmpl.man.label.base",
			"doc.tmpl.man.label.terminator",
			"doc.tmpl.markdown.label.base",
			"doc.tmpl.markdown.label.base_lower",
			"doc.tmpl.markdown.label.terminator",
			"doc.tmpl.markdown.label.terminator_lower",
		},
		"pl": {
			"doc.tmpl.man.label.base",
			"doc.tmpl.man.label.terminator",
			"doc.tmpl.markdown.label.base",
			"doc.tmpl.markdown.label.base_lower",
			"doc.tmpl.markdown.label.terminator",
			"doc.tmpl.markdown.label.terminator_lower",
		},
		"uk": {
			"doc.tmpl.man.label.base",
			"doc.tmpl.man.label.terminator",
			"doc.tmpl.markdown.label.base",
			"doc.tmpl.markdown.label.base_lower",
			"doc.tmpl.markdown.label.terminator",
			"doc.tmpl.markdown.label.terminator_lower",
		},
	}

	enMessages := mergedLocaleMessages(builtinI18nCatalog, nil, "en")
	for locale, keys := range checks {
		messages := mergedLocaleMessages(builtinI18nCatalog, nil, locale)
		for _, key := range keys {
			if messages[key] == enMessages[key] {
				t.Fatalf("locale %q leaves %q as English source text %q", locale, key, messages[key])
			}
		}
	}
}

func TestRussianDocLabelsUseFullTerms(t *testing.T) {
	messages := mergedLocaleMessages(builtinI18nCatalog, nil, "ru")

	for _, key := range []string{
		"doc.tmpl.code.tag.def",
		"doc.tmpl.code.tag.delim",
		"doc.tmpl.code.tag.env",
		"doc.tmpl.code.tag.optional",
		"doc.tmpl.code.tag.order",
		"doc.tmpl.code.tag.req",
		"doc.tmpl.code.tag.term",
		"doc.tmpl.html.meta.env",
		"doc.tmpl.man.label.environment",
		"doc.tmpl.markdown.label.environment",
		"doc.tmpl.markdown.table.env",
	} {
		value := messages[key]
		for _, bad := range []string{"умолч.", "разд.", "окруж.", "необяз.", "поряд.", "обяз.", "терм."} {
			if strings.Contains(value, bad) {
				t.Fatalf("russian label %q still uses abbreviation %q in %q", key, bad, value)
			}
		}
	}

	if got := messages["doc.tmpl.man.label.environment"]; got != "Переменные окружения" {
		t.Fatalf("unexpected russian environment label: %q", got)
	}

	for _, key := range []string{
		"doc.tmpl.code.tag.kv",
		"doc.tmpl.html.meta.key_value_delimiter",
		"doc.tmpl.man.label.key_value_delimiter",
		"doc.tmpl.markdown.label.key_value_delimiter",
		"ini.example.meta.key_value_delimiter",
	} {
		if got := messages[key]; got != "разделитель" {
			t.Fatalf("unexpected russian delimiter label for %q: %q", key, got)
		}
	}
}
