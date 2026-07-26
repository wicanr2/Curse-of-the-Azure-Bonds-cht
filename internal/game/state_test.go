package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

func testCatalog() locale.Catalog {
	return locale.Catalog{Language: "zh-TW", Strings: map[string]string{
		"title": "青色枷的詛咒", "press_enter": "請按 Enter 繼續",
		"you_are_at_the_edge_of": "你已抵達邊界。", "enter_city": "進入城市", "journey_on": "繼續旅程",
	}}
}

func TestLocalizedOpeningFlow(t *testing.T) {
	state := NewState(testCatalog())
	if state.Title != "青色枷的詛咒" || state.Mode != ModeTitle {
		t.Fatalf("initial state=%#v", state)
	}
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Choices[0] != "進入城市" {
		t.Fatalf("opening state=%#v", state)
	}
	if err := state.Apply(ActionEnterCity); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != "進入城市" {
		t.Fatalf("event state=%#v", state)
	}
}

func TestRejectsWrongModeAction(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.Apply(ActionEnterCity); err == nil {
		t.Fatal("expected invalid action")
	}
}

func TestOpeningStateRecordsOriginalECLText(t *testing.T) {
	// 0x80 length + packed "YOU ARE AT THE EDGE OF" candidate.
	packed := []byte{0x64, 0xf5, 0x60, 0x05, 0x21, 0x60, 0x05, 0x48, 0x14, 0x20, 0x58, 0x05, 0x10, 0x71, 0x60, 0x3c, 0x68, 0x00}
	block := append([]byte{0, 0, 0x80, byte(len(packed))}, packed...)
	state := NewStateFromECL(testCatalog(), block)
	if state.OriginalOpening != "YOU ARE AT THE EDGE OF" {
		t.Fatalf("original=%q", state.OriginalOpening)
	}
}

func TestStateUsesECLInitialMenuChoices(t *testing.T) {
	block := make([]byte, 2+47)
	for i := 0; i < 5; i++ {
		pos := 2 + i*4
		block[pos+1], block[pos+2], block[pos+3] = 0x02, 0x14, 0x80
	}
	payload := block[2:]
	payload[20] = 0x2B
	payload[21], payload[22], payload[23] = 0x02, 0x00, 0x90
	payload[24], payload[25] = 0x00, 0x02
	// Packed "ENTER CITY" and "JOURNEY ON" from the original ECL string records.
	copy(payload[26:], []byte{0x80, 0x08, 0x14, 0xE5, 0x05, 0x4A, 0x00, 0xC9, 0x51, 0x90})
	copy(payload[36:], []byte{0x80, 0x08, 0x28, 0xF5, 0x52, 0x38, 0x56, 0x60, 0x3C, 0xE0})
	payload[46] = 0x00
	state := NewStateFromECL(testCatalog(), block)
	if len(state.OriginalChoices) != 2 || state.OriginalChoices[0] != "ENTER CITY" || state.Choices[1] != "繼續旅程" {
		t.Fatalf("state=%#v", state)
	}
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil || state.Mode != ModeEvent || state.Message != "繼續旅程" {
		t.Fatalf("selected state=%#v err=%v", state, err)
	}
}

func TestLocalizedCityMenuOptions(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["ashabenford"] = "阿沙本福德"
	catalog.Strings["dagger_falls"] = "匕首瀑布"
	if got := localizeOption(catalog, "DAGGER FALLS"); got != "匕首瀑布" {
		t.Fatalf("localized city=%q", got)
	}
}
