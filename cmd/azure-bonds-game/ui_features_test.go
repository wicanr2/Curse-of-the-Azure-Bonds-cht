package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestUISettingsRoundTripAndRejectUnsupportedResolution(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui-settings.json")
	want := uiSettings{Schema: "coab-ui-settings/1", Theme: "original", Language: "zh-TW", Width: 1024, Height: 768, SpoilerWarning: true, Explored: map[string][]string{}}
	if err := saveUISettings(path, want); err != nil {
		t.Fatal(err)
	}
	if got := loadUISettings(path); !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
	want.Width, want.Height = 800, 600
	if err := saveUISettings(path, want); err != nil {
		t.Fatal(err)
	}
	got := loadUISettings(path)
	if got.Width != 640 || got.Height != 480 {
		t.Fatalf("unsupported resolution should fail closed to 640x480, got %dx%d", got.Width, got.Height)
	}
}

func TestF1ToF6AreGlobalAndResolutionCycles(t *testing.T) {
	state := game.NewState(locale.Catalog{})
	keys := newScriptedKeys()
	path := filepath.Join(t.TempDir(), "ui-settings.json")
	application := &app{state: &state, keys: keys, ui: newUIRuntime(defaultUISettings(), path), locales: map[string]locale.Catalog{
		"zh-CN": {Language: "zh-CN", Strings: map[string]string{"title": "青色枷的诅咒"}},
	}}

	keys.press(ebiten.KeyF1)
	if err := application.Update(); err != nil || !application.ui.helpOpen {
		t.Fatalf("F1 did not open Help: open=%v err=%v", application.ui.helpOpen, err)
	}
	keys.press(ebiten.KeyF2)
	if err := application.Update(); err != nil || application.ui.settings.Theme != "original" {
		t.Fatalf("F2 did not switch theme: theme=%q err=%v", application.ui.settings.Theme, err)
	}
	keys.press(ebiten.KeyF3)
	if err := application.Update(); err != nil || !application.ui.guideOpen || application.ui.helpOpen {
		t.Fatalf("F3 did not open guide exclusively: guide=%v help=%v err=%v", application.ui.guideOpen, application.ui.helpOpen, err)
	}
	keys.press(ebiten.KeyF4)
	if err := application.Update(); err != nil {
		t.Fatal(err)
	}
	if width, height := application.Layout(0, 0); width != 1024 || height != 768 {
		t.Fatalf("first F4 = %dx%d, want 1024x768", width, height)
	}
	keys.press(ebiten.KeyF4)
	_ = application.Update()
	keys.press(ebiten.KeyF4)
	_ = application.Update()
	if width, height := application.Layout(0, 0); width != 640 || height != 480 {
		t.Fatalf("third F4 = %dx%d, want 640x480", width, height)
	}
	keys.press(ebiten.KeyF6)
	if err := application.Update(); err != nil || state.LocaleLanguage() != "zh-CN" || application.ui.settings.Language != "zh-CN" {
		t.Fatalf("F6 did not switch and persist locale: state=%q settings=%q err=%v", state.LocaleLanguage(), application.ui.settings.Language, err)
	}
}

func TestGuideFullModeRequiresOneTimeSpoilerAcknowledgement(t *testing.T) {
	state := game.NewState(locale.Catalog{})
	keys := newScriptedKeys()
	application := &app{state: &state, keys: keys, ui: newUIRuntime(defaultUISettings(), filepath.Join(t.TempDir(), "ui.json"))}
	application.ui.guideOpen = true

	keys.press(ebiten.KeyV)
	if err := application.Update(); err != nil || !application.ui.spoilerAsk || application.ui.guideFull {
		t.Fatalf("first V should ask, not reveal: ask=%v full=%v err=%v", application.ui.spoilerAsk, application.ui.guideFull, err)
	}
	keys.press(ebiten.KeyEnter)
	if err := application.Update(); err != nil || !application.ui.guideFull || !application.ui.settings.SpoilerWarning {
		t.Fatalf("confirmation did not reveal and persist: full=%v ack=%v err=%v", application.ui.guideFull, application.ui.settings.SpoilerWarning, err)
	}
}

func TestGuideCatalogAndExplorationUseCurrentGEOCoordinates(t *testing.T) {
	catalog, err := loadGuideCatalog(filepath.Join("..", "..", "assets", "guide", "maps.zh-TW.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Maps) < 8 || len(catalog.Maps["2/01"].Points) == 0 {
		t.Fatalf("guide catalog is unexpectedly sparse: maps=%d tilverton=%d", len(catalog.Maps), len(catalog.Maps["2/01"].Points))
	}
	state := game.NewState(locale.Catalog{})
	state.GeoMapSet, state.GeoMapBlock = 2, 1
	state.Mode = game.ModeDungeon
	state.SetDungeonGeometryView(6, 5, 2)
	application := &app{state: &state, ui: newUIRuntime(defaultUISettings(), filepath.Join(t.TempDir(), "ui.json")), guide: catalog}
	application.rememberCurrentGuideCell()
	application.rememberCurrentGuideCell()
	if got := application.ui.settings.Explored["2/01"]; !reflect.DeepEqual(got, []string{"6,5"}) {
		t.Fatalf("explored = %v, want one current GEO cell", got)
	}
	if !application.ui.settingsDirty || application.ui.exploredSinceSave != 1 {
		t.Fatalf("first unique cell dirty=%v count=%d", application.ui.settingsDirty, application.ui.exploredSinceSave)
	}
	if _, err := os.Stat(application.ui.settingsPath); !os.IsNotExist(err) {
		t.Fatalf("first cell unexpectedly wrote settings: %v", err)
	}
}

func TestF10WithoutPartyTerminates(t *testing.T) {
	state := game.NewState(locale.Catalog{})
	keys := newScriptedKeys()
	application := &app{state: &state, keys: keys, ui: newUIRuntime(defaultUISettings(), filepath.Join(t.TempDir(), "ui.json"))}
	keys.press(ebiten.KeyF10)
	err := application.Update()
	if !errors.Is(err, ebiten.Termination) {
		t.Fatalf("F10 without progress = %v, want ebiten.Termination", err)
	}
}

func TestF10SavesBeforeTerminationAndFailsClosed(t *testing.T) {
	makeApp := func(path string) (*app, *scriptedKeys) {
		state := game.NewState(locale.Catalog{})
		abilities := party.Abilities{Strength: 10, Intelligence: 10, Wisdom: 10, Dexterity: 10, Constitution: 10, Charisma: 10}
		if err := state.SetPartyRoster(party.Roster{{ID: "quit-save", Name: "測試隊員", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1, Abilities: abilities, HitPoints: 10, MaxHitPoints: 10}}); err != nil {
			t.Fatal(err)
		}
		keys := newScriptedKeys()
		return &app{state: &state, keys: keys, partyPath: path, ui: newUIRuntime(defaultUISettings(), filepath.Join(t.TempDir(), "ui.json"))}, keys
	}

	savePath := filepath.Join(t.TempDir(), "party.json")
	application, keys := makeApp(savePath)
	keys.press(ebiten.KeyF10)
	if err := application.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("successful F10 = %v, want termination", err)
	}
	if _, err := os.Stat(savePath); err != nil {
		t.Fatalf("F10 did not create save: %v", err)
	}

	application, keys = makeApp(t.TempDir()) // A directory cannot be atomically replaced by the save file.
	keys.press(ebiten.KeyF10)
	if err := application.Update(); err != nil {
		t.Fatalf("failed save must keep running, got %v", err)
	}
	if application.state.Message == "" {
		t.Fatal("failed F10 save must report an error")
	}
}
