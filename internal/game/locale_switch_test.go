package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

func TestSetLocaleCatalogRetranslatesVisibleStateAndOriginalChoices(t *testing.T) {
	state := NewState(locale.Catalog{Language: "zh-TW", Strings: map[string]string{
		"title": "標題", "press_enter": "繼續", "wilderness": "荒野",
	}})
	state.Message = "繼續"
	state.currentOriginalChoices = []string{"ENTER CITY"}
	state.Choices = []string{"ENTER CITY"}
	state.SetLocaleCatalog(locale.Catalog{Language: "en", Strings: map[string]string{
		"title": "Title", "press_enter": "Continue", "wilderness": "Wilderness",
	}})
	if state.Title != "Title" || state.Message != "Continue" || state.LocationName != "Wilderness" {
		t.Fatalf("visible state was not retranslated: title=%q message=%q location=%q", state.Title, state.Message, state.LocationName)
	}
	if state.LocaleLanguage() != "en" {
		t.Fatalf("language = %q, want en", state.LocaleLanguage())
	}
}
