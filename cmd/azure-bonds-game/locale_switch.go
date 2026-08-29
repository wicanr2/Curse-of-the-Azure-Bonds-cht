package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

var languageCycle = [...]string{"zh-TW", "zh-CN", "ja", "en"}

func loadRuntimeLocales(root string) (map[string]locale.Catalog, error) {
	loaded := make(map[string]locale.Catalog, len(languageCycle))
	for _, language := range languageCycle {
		data, err := osReadFile(filepath.Join(root, language+".json"))
		if err != nil {
			return nil, fmt.Errorf("load %s locale: %w", language, err)
		}
		catalog, err := locale.Load(data)
		if err != nil {
			return nil, fmt.Errorf("load %s locale: %w", language, err)
		}
		if catalog.Language != language {
			return nil, fmt.Errorf("locale %s declares %s", language, catalog.Language)
		}
		loaded[language] = catalog
	}
	return loaded, nil
}

func loadRuntimeGuides(root string) (map[string]guideCatalog, error) {
	loaded := make(map[string]guideCatalog, len(languageCycle))
	for _, language := range languageCycle {
		catalog, err := loadGuideCatalog(filepath.Join(root, "maps."+language+".json"))
		if err != nil {
			return nil, fmt.Errorf("load %s guide: %w", language, err)
		}
		loaded[language] = catalog
	}
	return loaded, nil
}

var osReadFile = os.ReadFile

func (a *app) cycleLanguage() {
	current := a.ui.settings.Language
	index := 0
	for i, language := range languageCycle {
		if language == current {
			index = i
			break
		}
	}
	next := languageCycle[(index+1)%len(languageCycle)]
	catalog, found := a.locales[next]
	if !found {
		a.state.Message = fmt.Sprintf(a.state.LocalizedText("ui_locale_unavailable", "Locale %s is not installed."), next)
		return
	}
	a.state.SetLocaleCatalog(catalog)
	a.ui.settings.Language = next
	a.selectLocaleTypography(next)
	if guide, found := a.guides[next]; found {
		a.guide = guide
	}
	a.persistUISettings()
}
