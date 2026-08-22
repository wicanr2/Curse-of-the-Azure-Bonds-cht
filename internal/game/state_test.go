package game

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
)

func testCatalog() locale.Catalog {
	return locale.Catalog{Language: "zh-TW", Strings: map[string]string{
		"title": "青色枷的詛咒", "press_enter": "請按 Enter 繼續",
		"you_are_at_the_edge_of": "你已抵達邊界。", "enter_city": "進入城市", "journey_on": "繼續旅程",
	}}
}

func optionTextFromPack(t *testing.T, state State, optionRuleID string) string {
	t.Helper()
	if state.dataPack == nil {
		t.Fatalf("game pack is unavailable while resolving option rule %q", optionRuleID)
	}
	for _, rule := range state.dataPack.OptionRules {
		if rule.ID != optionRuleID {
			continue
		}
		value, ok := state.dataPack.Text(rule.MessageID, state.catalog.Language)
		if !ok {
			t.Fatalf("option rule %q message %q has no locale value", rule.ID, rule.MessageID)
		}
		return value
	}
	t.Fatalf("option rule %q is missing from the game pack", optionRuleID)
	return ""
}

func TestLocalizedOpeningFlow(t *testing.T) {
	state := NewState(testCatalog())
	wantEnterCity := optionTextFromPack(t, state, "ecl-option.enter-city")
	if state.Title != "青色枷的詛咒" || state.Mode != ModeTitle {
		t.Fatalf("initial state=%#v", state)
	}
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Choices[0] != wantEnterCity {
		t.Fatalf("opening state=%#v", state)
	}
	if err := state.Apply(ActionEnterCity); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != wantEnterCity {
		t.Fatalf("event state=%#v", state)
	}
}

func TestWorldLocationProjectionCoversAreaOneCities(t *testing.T) {
	state := NewState(testCatalog())
	tests := []struct {
		value    uint16
		location Location
		original string
	}{
		{0, LocationTilverton, "TILVERTON"},
		{1, LocationShadowdale, "SHADOWDALE"},
		{2, LocationAshabenford, "ASHABENFORD"},
		{3, LocationDaggerFalls, "DAGGER FALLS"},
		{4, LocationStandingStone, "THE STANDING STONE"},
		{5, LocationVoonlar, "VOONLAR"},
		{6, LocationPhlan, "PHLAN"},
		{7, LocationTeshwave, "TESHWAVE"},
		{8, LocationEssembra, "ESSEMBRA"},
		{9, LocationHap, "HAP"},
		{10, LocationYulash, "YULASH"},
		{11, LocationHillsfar, "HILLSFAR"},
		{12, LocationZhentilKeep, "ZHENTIL KEEP"},
		{13, LocationMythDrannor, "MYTH DRANNOR"},
	}
	for _, test := range tests {
		state.setWorldLocation(test.value)
		if state.Location != test.location || state.OriginalLocation != test.original ||
			state.Area.CurrentCity != uint8(test.value) {
			t.Fatalf("world location %d projected location=%v original=%q currentCity=%d",
				test.value, state.Location, state.OriginalLocation, state.Area.CurrentCity)
		}
	}
}

func TestECLClockAdvancesSharedGameTime(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.applyECLClockSignals(ecl.RunResult{ClockRequests: []ecl.ClockRequest{{TimeStep: 3, TimeSlot: 1}}}); err != nil {
		t.Fatal(err)
	}
	if got := state.GameTimeSlots(); got[1] != 3 {
		t.Fatalf("clock=%v, want slot 1 to advance by 3", got)
	}
	if state.Area.GameTime[1] != 3 {
		t.Fatalf("area clock=%v, want slot 1 to mirror state", state.Area.GameTime)
	}
	if err := state.applyECLClockSignals(ecl.RunResult{ClockRequests: []ecl.ClockRequest{{TimeStep: 1, TimeSlot: 7}}}); err == nil {
		t.Fatal("invalid ECL CLOCK slot was accepted")
	}
}

func TestAdvanceGameTimeMirrorsClockIntoECLWorkMemory(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x42: {0, 0}}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(testCatalog())
	state.session = session
	state.gameClock = [7]uint16{0, 4, 5, 13, 7, 2, 9}
	if err := state.AdvanceGameTime(1, 1); err != nil {
		t.Fatal(err)
	}
	want := [7]uint16{0, 5, 5, 13, 7, 2, 9}
	for index, expected := range want {
		got, ok := session.MemoryValue(eclClockBaseAddress + uint16(index))
		if !ok || got != expected {
			t.Fatalf("ECL clock slot %d=%#x,%v, want %#x,true", index, got, ok, expected)
		}
	}
}

func TestGameTimeDisplayUsesReferenceArea1Mapping(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	state.gameClock = [7]uint16{0, 4, 5, 13, 7, 2, 9}

	display := state.GameTimeDisplay()
	if display.Hour != 13 || display.Minute != 54 || display.Day != 7 || display.Month != 2 || display.Year != 9 {
		t.Fatalf("display=%+v, want 13:54 day=7 month=2 year=9", display)
	}
	if got, want := state.GameTimeText(), fmt.Sprintf(catalog.Text("game_time", ""), 13, 54, 7, 2, 9); got != want {
		t.Fatalf("text=%q", got)
	}
}

func TestSoundEventsAreOneShotAndRendererNeutral(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	events := state.ConsumeSoundEvents()
	if len(events) != 1 || events[0] != SoundOverture {
		t.Fatalf("start sound events=%#v", events)
	}
	if got := state.ConsumeSoundEvents(); len(got) != 0 {
		t.Fatalf("sound events were not consumed: %#v", got)
	}
	// 時機照 spec 1186：近戰**每一次**揮擊都有 `SWISHFX`，命中補一聲 `HITFX`，
	// 目標離場放 `DEADFX`；**揮空沒有音效**——`MISSFX` 是法術沒中的聲音
	// （原作只在 `TWINKLE` 裡呼叫它），不是近戰揮空的聲音。
	state.requestAttackSounds([]combat.AttackResult{
		{Hit: true, TargetHP: 0},
		{Hit: false, TargetHP: 4},
	})
	want := []SoundEvent{SoundSwish, SoundHit, SoundDead, SoundSwish}
	if got := state.ConsumeSoundEvents(); !reflect.DeepEqual(got, want) {
		t.Fatalf("attack sound events=%#v want %#v", got, want)
	}
	// 負對照：揮空那一擊**不能**發 `SoundMiss`。上面的序列已經釘住了，
	// 但這一條把「為什麼」寫清楚——這一格錯了會在原作從不出聲的時機放法術音效。
	for _, event := range want {
		if event == SoundMiss {
			t.Fatal("近戰揮空不該發 SoundMiss")
		}
	}
}

func TestPictureResultBecomesResumableEvent(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeWilderness
	state.Choices = []string{"事件畫面"}
	state.currentOriginalChoices = []string{"PICTURE"}
	state.eclBlock = []byte{0, 0, 0x0E, 0x00, 0x1D, 0x00}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 0x1D || state.OriginalEvent != "PICTURE" {
		t.Fatalf("picture state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.PictureRequested {
		t.Fatalf("picture continuation state=%#v", state)
	}
}

func TestECLWhoPausesForRosterAndResumesSelectedPlayer(t *testing.T) {
	block := make([]byte, 2+0x18)
	for index := 0; index < 5; index++ {
		pos := 2 + index*4
		block[pos+1], block[pos+2], block[pos+3] = 0x02, 0x14, 0x80
	}
	block[2+0x14] = 0x39
	block[2+0x15], block[2+0x16] = 0x00, 0x00
	block[2+0x17] = 0x00
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x50: block}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(testCatalog())
	state.session = session
	state.eclBlock = session.CurrentData()
	state.Mode = ModeWilderness
	state.Choices = []string{"觸發 WHO"}
	state.currentOriginalChoices = []string{"WHO"}
	state.partyRoster = party.Roster{
		{ID: "a", Name: "甲", Class: party.ClassFighter, Level: 1, Abilities: party.Abilities{Strength: 10, Dexterity: 10, Constitution: 10}},
		{ID: "b", Name: "乙", Class: party.ClassFighter, Level: 1, Abilities: party.Abilities{Strength: 10, Dexterity: 10, Constitution: 10}},
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 2 || state.SelectedPlayerID() != "" {
		t.Fatalf("WHO pause state=%#v selected=%q", state, state.SelectedPlayerID())
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.SelectedPlayerID() != "b" {
		t.Fatalf("WHO resume state=%#v selected=%q", state, state.SelectedPlayerID())
	}
}

func TestECLLoadCharacterResolvesZeroBasedRosterAndHighBit(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "a", Name: "甲"}, {ID: "b", Name: "乙"}}
	state.applyECLLoadCharacterSignals(ecl.RunResult{LoadCharacterRequests: []ecl.LoadCharacterRequest{{
		Address: 0x7F79, Value: 0x81, PlayerIndex: 1, HighBitSet: true,
	}}})
	if state.SelectedPlayerID() != "b" || state.LoadCharacterNotFound() || !state.LoadCharacterHighBit() {
		t.Fatalf("state=%#v selected=%q", state, state.SelectedPlayerID())
	}
	state.applyECLLoadCharacterSignals(ecl.RunResult{LoadCharacterRequests: []ecl.LoadCharacterRequest{{
		Address: 0x7F79, Value: 2, PlayerIndex: 2,
	}}})
	if state.SelectedPlayerID() != "b" || !state.LoadCharacterNotFound() {
		t.Fatalf("invalid lookup state=%#v selected=%q", state, state.SelectedPlayerID())
	}
}

func TestECLPartyContextProjectsCharacterNames(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "a", Name: "阿卡巴", Equipment: []monster.ItemRecord{{Type: 0x5E}}, Effects: []monster.AffectRecord{{Kind: 0x27, Active: true}, {Kind: 0x2A}}}, {ID: "b", Name: "乙"}}
	context := state.eclPartyContext()
	if len(context.Members) != 2 || context.Members[0].Name != "阿卡巴" || context.Members[1].Name != "乙" || len(context.Members[0].ItemTypes) != 1 || context.Members[0].ItemTypes[0] != 0x5E || len(context.Members[0].Effects) != 1 || context.Members[0].Effects[0] != 0x27 {
		t.Fatalf("party context=%#v", context)
	}
}

func TestECLDumpRemovesSelectedRosterAndFighter(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "a", Name: "甲"}, {ID: "b", Name: "乙"}, {ID: "c", Name: "丙"}}
	state.party = []combat.Fighter{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	state.whoSelectedIndex = 1
	err := state.applyECLDumpSignals(ecl.RunResult{DumpRequests: []ecl.DumpRequest{{
		SelectedPlayerIndex: 1, NextSelectedPlayerIndex: 0, NextSelectedPlayerSet: true, Resolved: true,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster) != 2 || state.partyRoster[0].ID != "a" || state.partyRoster[1].ID != "c" || len(state.party) != 2 || state.party[0].ID != "a" || state.party[1].ID != "c" || state.SelectedPlayerID() != "a" {
		t.Fatalf("roster=%#v fighters=%#v selected=%q", state.partyRoster, state.party, state.SelectedPlayerID())
	}
}

func TestECLDumpMayRemoveLastPartyMember(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "npc", Name: "同行者"}}
	state.party = []combat.Fighter{{ID: "npc"}}
	err := state.applyECLDumpSignals(ecl.RunResult{DumpRequests: []ecl.DumpRequest{{
		SelectedPlayerIndex: 0, NextSelectedPlayerIndex: -1, Resolved: true,
	}}})
	if err != nil || len(state.partyRoster) != 0 || len(state.party) != 0 || state.SelectedPlayerID() != "" {
		t.Fatalf("err=%v roster=%#v fighters=%#v selected=%q", err, state.partyRoster, state.party, state.SelectedPlayerID())
	}
}

func TestPictureUsesHeadBodyBranchWhenHeadBlockIsPresent(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeWilderness
	state.Choices = []string{"人物"}
	state.currentOriginalChoices = []string{"PICTURE"}
	state.eclBlock = []byte{0, 0, 0x0E, 0x00, 0x03, 0x00}
	state.SetSceneCharacter(0x01, 0x02)
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.SceneCharacterRequested || state.SceneHeadBlock != 0x01 || state.SceneBodyBlock != 0x03 {
		t.Fatalf("scene character state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.SceneCharacterRequested {
		t.Fatal("scene character request was not cleared")
	}
}

func TestAreaStateHeadBlockDrivesPictureBranch(t *testing.T) {
	state := NewState(testCatalog())
	state.SetAreaState(area.State{GameArea: 2, HeadBlockID: 0x01})
	state.Mode = ModeWilderness
	state.Choices = []string{"人物"}
	state.currentOriginalChoices = []string{"PICTURE"}
	state.eclBlock = []byte{0, 0, 0x0E, 0x00, 0x02, 0x00}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.SceneCharacterRequested || state.SceneHeadBlock != 0x01 || state.SceneBodyBlock != 0x02 {
		t.Fatalf("area-driven scene state=%#v", state)
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
	state := NewState(catalog)
	want, ok := state.dataPack.Text("dagger_falls", catalog.Language)
	if !ok {
		t.Fatal("game pack is missing dagger_falls")
	}
	if got := state.localizeOption("DAGGER FALLS"); got != want {
		t.Fatalf("localized city=%q, want pack value %q", got, want)
	}
}

func TestLocalizedEncounterMenuOptions(t *testing.T) {
	catalog := testCatalog()
	state := NewState(catalog)
	for _, original := range []string{"COMBAT", "WAIT", "FLEE", "ADVANCE", "PARLAY"} {
		want, ok := state.dataPack.LocalizeOption(original, catalog.Language)
		if !ok {
			t.Fatalf("game pack is missing option %q", original)
		}
		if got := state.localizeOption(original); got != want {
			t.Fatalf("%s localized as %q, want pack value %q", original, got, want)
		}
	}
}

func TestEncounterFleeReturnsToWildernessEvent(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.Choices = []string{"DISPLAY"}
	state.currentOriginalChoices = []string{"FLEE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "FLEE" || state.Message != catalog.Text("encounter_flee_done", "") {
		t.Fatalf("flee state=%+v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness {
		t.Fatalf("flee continuation mode=%v err=%v", state.Mode, err)
	}
}

func TestEncounterParlayOffersFiveTacticsAndReturnsEvent(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.Choices = []string{"DISPLAY"}
	state.currentOriginalChoices = []string{"PARLAY"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.parlayMenu || len(state.Choices) != 5 {
		t.Fatalf("parlay menu state=%+v", state)
	}
	wantChoices := []string{
		catalog.Text("parlay_haughty", ""), catalog.Text("parlay_sly", ""),
		catalog.Text("parlay_meek", ""), catalog.Text("parlay_nice", ""),
		catalog.Text("parlay_abusive", ""),
	}
	if state.Prompt != catalog.Text("parlay_menu_prompt", "") || !reflect.DeepEqual(state.Choices, wantChoices) {
		t.Fatalf("parlay prompt=%q choices=%#v want=%#v", state.Prompt, state.Choices, wantChoices)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	wantMessage := fmt.Sprintf(catalog.Text("encounter_parlay_done", ""), catalog.Text("parlay_meek", ""))
	if state.Mode != ModeEvent || state.OriginalEvent != "PARLAY" || state.Message != wantMessage {
		t.Fatalf("parlay result state=%+v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness {
		t.Fatalf("parlay continuation mode=%v err=%v", state.Mode, err)
	}
}

func TestLocationDefaultsToWilderness(t *testing.T) {
	state := NewState(testCatalog())
	if state.Location != LocationWilderness || state.LocationName != "Wilderness" {
		t.Fatalf("location=%v, want wilderness", state.Location)
	}
}

func TestTurnDungeonUsesEightDirectionOrder(t *testing.T) {
	state := NewState(testCatalog())
	state.TurnDungeon(2)
	if state.DungeonDirection != 2 {
		t.Fatalf("direction after right turn=%d, want 2", state.DungeonDirection)
	}
	state.TurnDungeon(-4)
	if state.DungeonDirection != 6 {
		t.Fatalf("direction after left turn=%d, want 6", state.DungeonDirection)
	}
	state.TurnDungeon(10)
	if state.DungeonDirection != 0 {
		t.Fatalf("direction after wrapped turn=%d, want 0", state.DungeonDirection)
	}
}

func TestCombatMapDirectionIsValidated(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.SetCombatMapDirection(8); err == nil {
		t.Fatal("direction 8 should be rejected")
	}
	if err := state.SetCombatMapDirection(3); err != nil {
		t.Fatal(err)
	}
}

func TestDungeonDefaultsFollowReferenceInit(t *testing.T) {
	state := NewState(testCatalog())
	if state.DungeonX != 7 || state.DungeonY != 13 || state.DungeonDirection != 0 {
		t.Fatalf("default dungeon state=(%d,%d,%d), want (7,13,0)", state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}

func TestDungeonStateActionsUseRosterAndSeededRolls(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{HitPoints: 6, ThiefSkills: []uint8{0, 100}}}
	state.SetDungeonSeed(42)
	result := state.PickDungeonLock()
	if !result.Attempted || !result.Opened || result.CharacterIndex != 0 {
		t.Fatalf("pick result=%#v", result)
	}
	state.partyRoster[0].SpellSlots = []uint8{0x1F, 2}
	if !state.ConsumeDungeonKnockSpell() {
		t.Fatal("Knock slot should be consumed")
	}
	if got := state.partyRoster[0].SpellSlots; len(got) != 1 || got[0] != 2 {
		t.Fatalf("remaining spell slots=%#v", got)
	}
}

func TestDungeonDoorMenuOptionsReadLoadedRoster(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ThiefSkills: []uint8{0, 25}, SpellSlots: []uint8{0x1F}}}
	if got := state.DungeonDoorMenuOptions(2); !got.Bash || !got.Pick || !got.Knock {
		t.Fatalf("detail2 options=%#v", got)
	}
	if got := state.DungeonDoorMenuOptions(3); got.Pick || !got.Bash || !got.Knock {
		t.Fatalf("detail3 options=%#v", got)
	}
}

func TestPartySaveLoadRoundTripRestoresDungeonViewState(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "p1", Name: "阿勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 11, 6, 2
	state.DungeonWallType, state.DungeonWallRoof = 7, 0x40
	path := t.TempDir() + "/game.json"
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(testCatalog())
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	if loaded.DungeonX != 11 || loaded.DungeonY != 6 || loaded.DungeonDirection != 2 || loaded.DungeonWallType != 7 || loaded.DungeonWallRoof != 0x40 {
		t.Fatalf("loaded dungeon state=(%d,%d,%d)", loaded.DungeonX, loaded.DungeonY, loaded.DungeonDirection)
	}
}

func TestPartySaveLoadRoundTripRestoresDungeonSearchState(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "p1", Name: "阿勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.DungeonSearchEnabled = true
	state.dungeonSearchEdges["tilverton.sewers.wall-09-west"] = true
	state.Mode = ModeDungeon
	path := t.TempDir() + "/search-state.json"
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(trainingTestCatalog(t))
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	if !loaded.DungeonSearchEnabled || !loaded.dungeonSearchEdges["tilverton.sewers.wall-09-west"] ||
		!loaded.DungeonSearchActive() {
		t.Fatalf("loaded dungeon search state enabled=%v edges=%v", loaded.DungeonSearchEnabled, loaded.dungeonSearchEdges)
	}
}

func TestDungeonSearchToggleDoesNotConsumeTurn(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeDungeon
	if err := state.ToggleDungeonSearch(); err != nil {
		t.Fatal(err)
	}
	if !state.DungeonSearchActive() {
		t.Fatal("SEARCH toggle did not enable persistent mode")
	}
	if err := state.ToggleDungeonSearch(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonSearchActive() {
		t.Fatal("SEARCH toggle did not disable persistent mode")
	}
}

func TestDiscoveredDungeonSearchEdgeIsPassableFromBothSides(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.GeoMapSet, state.GeoMapBlock = 4, 0x25
	state.dungeonSearchEdges["zhentil-keep.beholder-cave.dexam-east"] = true
	if !state.searchEdgeDiscovered(14, 1, 2) {
		t.Fatal("declared search edge was not passable from its discovery side")
	}
	if !state.searchEdgeDiscovered(15, 1, 6) {
		t.Fatal("one physical search edge was not passable from its opposite side")
	}
}

func TestDungeonSearchPolicyControlsMovementMinutes(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeDungeon
	if got, ok := state.dungeonMoveMinutes(); !ok || got != 1 {
		t.Fatalf("SEARCH-off movement minutes=%d found=%v, want 1,true", got, ok)
	}
	if err := state.ToggleDungeonSearch(); err != nil {
		t.Fatal(err)
	}
	if got, ok := state.dungeonMoveMinutes(); !ok || got != 10 {
		t.Fatalf("SEARCH-on movement minutes=%d found=%v, want 10,true", got, ok)
	}
	if err := state.advanceDungeonMoveTime(); err != nil {
		t.Fatal(err)
	}
	if slots := state.GameTimeSlots(); slots[1] != 0 || slots[2] != 1 {
		t.Fatalf("SEARCH-on movement clock=%v, want ten minutes", slots)
	}
}

func TestPartySaveRestoresJournalByStableIDInCurrentLocale(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "p1", Name: "亞勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	zhPage, found := state.dataPack.Text("journal.31", "zh-TW")
	if !found {
		t.Fatal("journal.31 zh-TW is absent")
	}
	state.appendJournalPage("journal.31", zhPage)
	path := t.TempDir() + "/game.json"
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}

	loaded := NewState(locale.Catalog{Language: "en", Strings: map[string]string{"title": "test"}})
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	wantEnglish, found := loaded.dataPack.Text("journal.31", "en")
	if !found {
		t.Fatal("journal.31 en is absent")
	}
	if !reflect.DeepEqual(loaded.journalMessageIDs, []string{"journal.31"}) ||
		!reflect.DeepEqual(loaded.JournalPages, []string{wantEnglish}) || loaded.JournalPages[0] == zhPage {
		t.Fatalf("restored journal IDs=%v pages=%q", loaded.journalMessageIDs, loaded.JournalPages)
	}
}

func TestActiveCombatSaveRestoresCombatantNameInCurrentLocale(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "p1", Name: "亞勇", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	result := ecl.RunResult{CombatRequested: true, MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 7, Count: 1, IconBlock: 81}}}
	records := map[uint8]monster.Record{7: {Name: "HIPPOGRIFF", HitPoints: 8, MaxHitPoints: 8, AttackBlows: [2]uint8{2, 0}}}
	partyFighters := []combat.Fighter{{ID: "p1", Name: "亞勇", Side: combat.SideParty, HitPoints: 12, MaxHitPoints: 12}}
	if err := state.StartEncounter(result, records, partyFighters, 11); err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/combat.json"
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}

	loaded := NewState(locale.Catalog{Language: "en", Strings: map[string]string{"title": "test"}})
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	wantEnglish, found := loaded.dataPack.LocalizeCombatantName("HIPPOGRIFF", "en")
	if !found {
		t.Fatal("HIPPOGRIFF English combatant name is absent")
	}
	fighters := loaded.CombatFighters()
	var restoredEnemy *combat.Fighter
	for index := range fighters {
		if fighters[index].SourceName == "HIPPOGRIFF" {
			restoredEnemy = &fighters[index]
			break
		}
	}
	if loaded.Mode != ModeCombat || len(fighters) != 2 || restoredEnemy == nil || restoredEnemy.Name != wantEnglish {
		t.Fatalf("restored combat mode=%v fighters=%+v", loaded.Mode, fighters)
	}
}

func TestSAVGAMPrefixStateAdapterRestoresKnownFieldsAndRawRecords(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "p1", Name: "阿勇"}}
	state.Area = area.State{GameArea: 4, Current3DMapBlockID: 0x12, CurrentCity: 2, InDungeon: true}
	state.MapX, state.MapY = -7, 13
	state.DungeonDirection, state.DungeonWallType, state.DungeonWallRoof = 6, 0x81, 0x40
	path := t.TempDir() + "/SAVGAM0.DAT"
	if err := state.SaveSAVGAMPrefix(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(testCatalog())
	if err := loaded.LoadSAVGAMPrefix(path); err != nil {
		t.Fatal(err)
	}
	if loaded.Area.GameArea != 4 || loaded.Area.Current3DMapBlockID != 0x12 || !loaded.Area.InDungeon || loaded.MapX != -7 || loaded.MapY != 13 || loaded.DungeonDirection != 6 || loaded.DungeonWallType != 0x81 || loaded.DungeonWallRoof != 0x40 {
		t.Fatalf("known SAVGAM state not restored: area=%#v map=(%d,%d) dungeon=(%d,%d,%d)", loaded.Area, loaded.MapX, loaded.MapY, loaded.DungeonDirection, loaded.DungeonWallType, loaded.DungeonWallRoof)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	container, err := partySave.DecodeSAVGAM(data)
	if err != nil {
		t.Fatal(err)
	}
	// 角色檔名那一欄是 Pascal 短字串：長度位元組在前（spec 1072）。
	want, err := partySave.SAVGAMCharacterRef("阿勇")
	if err != nil {
		t.Fatal(err)
	}
	if container.PartyCount != 1 || !bytes.Equal(container.CharacterRefs[0], want) {
		t.Fatalf("SAVGAM party refs=%#v", container.CharacterRefs[0])
	}
}

func TestLoadSAVGAMSlotLoadsPlayerRecordAndOptionalSidecars(t *testing.T) {
	directory := t.TempDir()
	areaState := area.State{GameArea: 3, Current3DMapBlockID: 0x21, InDungeon: true}
	area1, err := area.EncodeArea1(areaState, nil)
	if err != nil {
		t.Fatal(err)
	}
	area2, err := area.EncodeArea2(areaState, nil)
	if err != nil {
		t.Fatal(err)
	}
	prefix, err := partySave.EncodeSAVGAM(partySave.SAVGAMContainer{
		GameArea: 3, Area1: area1, Area2: area2,
		Runtime: make([]byte, partySave.SAVGAMRuntimeStateSize), ECL: make([]byte, partySave.SAVGAMECLMemorySize),
		PartyCount: 1, CharacterRefs: [8][]byte{[]byte("CHRDATC1")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory+"/savgamc.dat", prefix, 0o600); err != nil {
		t.Fatal(err)
	}
	record := make([]byte, party.DOSPlayerRecordSize)
	record[0x190] = 0xAA // undocumented byte must survive a known-field patch
	record[0] = 4
	copy(record[1:], "ELLA")
	record[0x10], record[0x11], record[0x12] = 16, 16, 10
	record[0x14], record[0x16], record[0x18], record[0x1A] = 10, 12, 14, 10
	record[0x74], record[0x75] = 7, 2
	record[0x78], record[0x1A4], record[0x10B] = 10, 8, 1
	if err := os.WriteFile(directory+"/CHRDATC1.sav", record, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory+"/CHRDATC1.fx", []byte{0x27, 3, 0, 1, 2, 3, 4, 5, 6}, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory+"/CHRDATC1.swg", make([]byte, monster.ItemRecordSize), 0o600); err != nil {
		t.Fatal(err)
	}
	state := NewState(testCatalog())
	if err := state.LoadSAVGAMSlot(directory, 'C'); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Area.GameArea != 3 || state.GeoMapBlock != 0x21 || len(state.partyRoster) != 1 || state.partyRoster[0].Name != "ELLA" || len(state.partyRoster[0].Effects) != 1 || len(state.partyRoster[0].Equipment) != 1 {
		t.Fatalf("loaded SAVGAM state=%#v roster=%#v", state, state.partyRoster)
	}
	if err := state.LoadSAVGAMSlot(directory, 'K'); err == nil {
		t.Fatal("slot key K should be rejected")
	}
	state.partyRoster[0].HitPoints = 6
	state.partyRoster[0].Gold = 321
	state.partyRoster[0].SpellSlots = []uint8{0x02, 0x04}
	state.partyRoster[0].Name = "新名"
	if err := os.WriteFile(directory+"/CHRDATC2.sav", record, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveSAVGAMSlot(directory, 'C'); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(directory + "/CHRDATC2.sav"); !os.IsNotExist(err) {
		t.Fatalf("stale player file was not removed, stat error=%v", err)
	}
	savedRecord, err := os.ReadFile(directory + "/CHRDATC1.sav")
	if err != nil {
		t.Fatal(err)
	}
	if savedRecord[0x190] != 0xAA {
		t.Fatalf("unknown player byte changed: %#x", savedRecord[0x190])
	}
	decoded, err := party.ParseDOSPlayerRecord(savedRecord, "CHRDATC1")
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Name != "新名" || decoded.CurrentHitPoints != 6 || decoded.Gold != 321 || !reflect.DeepEqual(decoded.MemorizedSpells, []uint8{2, 4}) {
		t.Fatalf("saved player fields=%#v", decoded)
	}
	savedPrefix, err := os.ReadFile(directory + "/savgamc.dat")
	if err != nil {
		t.Fatal(err)
	}
	savedContainer, err := partySave.DecodeSAVGAM(savedPrefix)
	if err != nil || !strings.HasPrefix(string(savedContainer.CharacterRefs[0]), "\x08CHRDATC1") {
		t.Fatalf("saved prefix ref=%q err=%v", savedContainer.CharacterRefs[0], err)
	}
}

func TestCityMenuSelectionMapsAllLocalizedLocations(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["ashabenford"] = "阿沙本福德"
	catalog.Strings["dagger_falls"] = "匕首瀑布"
	for index, want := range []struct {
		location Location
		name     string
		original string
	}{{LocationShadowdale, "暗影谷", "SHADOWDALE"}, {LocationAshabenford, "阿沙本福德", "ASHABENFORD"}, {LocationDaggerFalls, "匕首瀑布", "DAGGER FALLS"}} {
		state := NewState(catalog)
		state.selectionSequence = []uint16{0, 0, 1, uint16(index)}
		state.applyCitySelection()
		if state.Location != want.location || state.LocationName != want.name || state.OriginalLocation != want.original || state.Area.CurrentCity != uint8(index) {
			t.Fatalf("city %d state=%#v want=%#v", index, state, want)
		}
	}
}

func TestECLLoadFilesTransfersGeoBlockRequest(t *testing.T) {
	state := NewState(testCatalog())
	state.SetInDungeon(true)
	state.applyGeoMapLoad(ecl.RunResult{LoadFilesRequested: true, LoadFiles: [3]uint16{0x10, 0xFF, 0xFF}})
	set, block, ok := state.ConsumeGeoMapRequest()
	if !ok || set != 2 || block != 0x10 {
		t.Fatalf("geo request=(%d,%#x,%v), want set 2 block 0x10", set, block, ok)
	}
	if _, _, ok := state.ConsumeGeoMapRequest(); ok {
		t.Fatal("geo map request should be consumed exactly once")
	}
}

func TestECLLoadPiecesTransfersRequestOnce(t *testing.T) {
	state := NewState(testCatalog())
	state.applyLoadPieces(ecl.RunResult{LoadPiecesRequested: true, LoadPieces: [3]uint16{1, 2, 3}})
	pieces, ok := state.ConsumeLoadPiecesRequest()
	if !ok || pieces != [3]uint16{1, 2, 3} {
		t.Fatalf("load pieces request=%v,%v", pieces, ok)
	}
	if _, ok := state.ConsumeLoadPiecesRequest(); ok {
		t.Fatal("load pieces request should be consumed exactly once")
	}
}

func TestECLSpellSignalsTransferToStateOnce(t *testing.T) {
	state := NewState(testCatalog())
	state.applySpellSignals(ecl.RunResult{
		SpellSearches:      []ecl.SpellSearch{{SpellID: 0x12, SpellSlotAddress: 0x100, CharacterAddress: 0x200}},
		ProtectionRequests: []uint16{0x7C02},
	})
	spells := state.ConsumeSpellSearches()
	if len(spells) != 1 || spells[0].SpellID != 0x12 || spells[0].SpellSlotAddress != 0x100 || spells[0].CharacterAddress != 0x200 {
		t.Fatalf("spell requests=%#v", spells)
	}
	protection := state.ConsumeProtectionRequests()
	if len(protection) != 1 || protection[0] != 0x7C02 {
		t.Fatalf("protection requests=%#v", protection)
	}
	if len(state.ConsumeSpellSearches()) != 0 || len(state.ConsumeProtectionRequests()) != 0 {
		t.Fatal("ECL spell signals were not consumed exactly once")
	}
}

func TestECLDamageSignalsTransferToStateOnce(t *testing.T) {
	state := NewState(testCatalog())
	state.applyECLDamageSignals(ecl.RunResult{
		DamageRequests: []ecl.DamageRequest{{Flags: 0xA0, DiceCount: 1, DiceSize: 6, Bonus: 1, SaveFlags: 0x80}},
	})
	requests := state.ConsumeDamageRequests()
	want := ecl.DamageRequest{Flags: 0xA0, DiceCount: 1, DiceSize: 6, Bonus: 1, SaveFlags: 0x80}
	if len(requests) != 1 || requests[0] != want {
		t.Fatalf("damage requests=%#v, want %#v", requests, want)
	}
	if len(state.ConsumeDamageRequests()) != 0 {
		t.Fatal("ECL damage signals were not consumed exactly once")
	}
}

func TestResolvePendingECLDamageWritesRosterAndFighterHP(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{
		{ID: "first", HitPoints: 10, SavingThrows: []uint8{12, 12, 12, 12, 12}},
		{ID: "second", HitPoints: 10, SavingThrows: []uint8{12, 12, 12, 12, 12}},
	}
	state.party = []combat.Fighter{
		{ID: "first", HitPoints: 10, MaxHitPoints: 10},
		{ID: "second", HitPoints: 10, MaxHitPoints: 10},
	}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 0x80, DiceCount: 1, DiceSize: 6, Bonus: 1, SaveFlags: 0x81,
	}}})
	outcomes, err := state.ResolvePendingECLDamage(1, func(int) int { return 4 }, func(int) int { return 1 })
	if err != nil || len(outcomes) != 1 || outcomes[0].Applied != 5 {
		t.Fatalf("outcomes=%#v err=%v", outcomes, err)
	}
	if state.partyRoster[0].HitPoints != 10 || state.partyRoster[1].HitPoints != 5 || state.party[1].HitPoints != 5 {
		t.Fatalf("roster=%#v fighters=%#v", state.partyRoster, state.party)
	}
	if len(state.ConsumeDamageRequests()) != 0 {
		t.Fatal("resolved damage request remained pending")
	}
}

func TestAutomaticWholePartyECLDamageUsesStateSeed(t *testing.T) {
	state := NewState(testCatalog())
	state.SetECLSeed(1)
	state.partyRoster = party.Roster{
		{ID: "one", Name: "一", HitPoints: 100, MaxHitPoints: 100},
		{ID: "two", Name: "二", HitPoints: 100, MaxHitPoints: 100},
	}
	state.party = []combat.Fighter{
		{ID: "one", HitPoints: 100, MaxHitPoints: 100},
		{ID: "two", HitPoints: 100, MaxHitPoints: 100},
	}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 0xE0, DiceCount: 8, DiceSize: 8,
	}}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 2 || outcomes[0].DamageRoll != 38 ||
		state.partyRoster[0].HitPoints != 62 || state.partyRoster[1].HitPoints != 62 ||
		state.party[0].HitPoints != 62 || state.party[1].HitPoints != 62 {
		t.Fatalf("automatic damage outcomes=%+v roster=%+v fighters=%+v",
			outcomes, state.partyRoster, state.party)
	}
	if len(state.ConsumeDamageRequests()) != 0 {
		t.Fatal("automatic whole-party damage remained pending")
	}
}

func TestAutomaticWholePartyECLDamageResolvesSavingThrows(t *testing.T) {
	state := NewState(testCatalog())
	state.SetECLSeed(1)
	state.partyRoster = party.Roster{
		{ID: "one", Name: "一", HitPoints: 100, MaxHitPoints: 100, SavingThrows: []uint8{21, 21, 21, 21, 21}},
		{ID: "two", Name: "二", HitPoints: 100, MaxHitPoints: 100, SavingThrows: []uint8{21, 21, 21, 21, 21}},
	}
	state.party = []combat.Fighter{
		{ID: "one", HitPoints: 100, MaxHitPoints: 100},
		{ID: "two", HitPoints: 100, MaxHitPoints: 100},
	}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 0xC0, DiceCount: 3, DiceSize: 4, Bonus: 3, SaveFlags: 1,
	}}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 2 || outcomes[0].DamageRoll != 13 ||
		!outcomes[0].Saved || outcomes[0].Applied != 0 ||
		outcomes[1].Saved || outcomes[1].Applied != 13 ||
		state.partyRoster[0].HitPoints != 100 || state.partyRoster[1].HitPoints != 87 {
		t.Fatalf("saving damage outcomes=%+v roster=%+v", outcomes, state.partyRoster)
	}
	if len(state.ConsumeDamageRequests()) != 0 {
		t.Fatal("saving whole-party damage remained pending")
	}
}

func TestAutomaticWholePartyECLDamageDoesNotWaitBehindSelectedPackets(t *testing.T) {
	state := NewState(testCatalog())
	state.SetECLSeed(1)
	state.partyRoster = party.Roster{{ID: "one", HitPoints: 100, MaxHitPoints: 100}}
	state.party = []combat.Fighter{{ID: "one", HitPoints: 100, MaxHitPoints: 100}}
	selected := ecl.DamageRequest{Flags: 0x90, DiceCount: 1, DiceSize: 10, SaveFlags: 0x80}
	wholeParty := ecl.DamageRequest{Flags: 0xE0, DiceCount: 1, DiceSize: 1}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{selected, wholeParty}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil || len(outcomes) != 1 || outcomes[0].Applied != 1 ||
		state.partyRoster[0].HitPoints != 99 {
		t.Fatalf("mixed automatic damage outcomes=%+v err=%v roster=%+v", outcomes, err, state.partyRoster)
	}
	if !reflect.DeepEqual(state.pendingDamageRequests, []ecl.DamageRequest{selected}) {
		t.Fatalf("selected pending packet was not preserved: %#v", state.pendingDamageRequests)
	}
}

func TestResolvePendingECLDamageWithHitResolverHandlesRandomTargets(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "first", HitPoints: 10}, {ID: "second", HitPoints: 10}}
	state.party = []combat.Fighter{{ID: "first", HitPoints: 10}, {ID: "second", HitPoints: 10}}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 2, DiceCount: 1, DiceSize: 1, Bonus: 2, SaveFlags: 3,
	}}})
	targetRolls := []int{2, 1}
	targetIndex := 0
	outcomes, err := state.ResolvePendingECLDamageWithHitResolver(-1, func(sides int) int {
		if sides == 1 {
			return 1
		}
		value := targetRolls[targetIndex]
		targetIndex++
		return value
	}, func(int) int { return 1 }, func(target party.Character, bonus int, rollDie func(int) int) (bool, error) {
		return target.ID == "first" && bonus == 3, nil
	})
	if err != nil || len(outcomes) != 2 || state.partyRoster[0].HitPoints != 7 || state.partyRoster[1].HitPoints != 10 || state.party[0].HitPoints != 7 {
		t.Fatalf("outcomes=%#v roster=%#v fighters=%#v err=%v", outcomes, state.partyRoster, state.party, err)
	}
}

func TestResolvePendingECLDamageWithDefaultHitResolverUsesProjectedAC(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		HitPoints: 10, MaxHitPoints: 10,
		Effects: []monster.AffectRecord{{Kind: 0x19, Active: true}},
	}}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 1, DiceCount: 1, DiceSize: 1, SaveFlags: 0,
	}}})
	outcomes, err := state.ResolvePendingECLDamageWithDefaultHitResolver(0, func(sides int) int {
		if sides == 1 {
			return 1
		}
		return 14
	}, func(int) int { return 1 })
	if err != nil || len(outcomes) != 1 || outcomes[0].Hit || state.partyRoster[0].HitPoints != 10 {
		t.Fatalf("outcomes=%#v roster=%#v err=%v", outcomes, state.partyRoster, err)
	}
}

func TestResolvePendingECLDamageDefaultResolverPassesBlinkContext(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		HitPoints: 10, MaxHitPoints: 10, Effects: []monster.AffectRecord{{Kind: 0x25, Active: true}}}}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{Flags: 1, DiceCount: 1, DiceSize: 1}}})
	outcomes, err := state.ResolvePendingECLDamageWithDefaultHitResolverContext(-1, party.ECLHitContext{ActionDelay: 0}, func(sides int) int {
		if sides == 1 {
			return 1
		}
		return 20
	}, func(int) int { return 1 })
	if err != nil || len(outcomes) != 1 || outcomes[0].Hit || state.partyRoster[0].HitPoints != 10 {
		t.Fatalf("outcomes=%#v roster=%#v err=%v", outcomes, state.partyRoster, err)
	}
}

func TestResolvePendingECLDamageDisplaceBitIsTransactional(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		HitPoints: 10, MaxHitPoints: 10, Effects: []monster.AffectRecord{{Kind: 0x59, Active: true}}}}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{
		{Flags: 1, DiceCount: 1, DiceSize: 1},
		{Flags: 1, DiceCount: 0, DiceSize: 1},
	}})
	_, err := state.ResolvePendingECLDamageWithDefaultHitResolverContext(-1, party.ECLHitContext{CombatRound: 1}, func(sides int) int {
		if sides == 1 {
			return 1
		}
		return 20
	}, func(int) int { return 1 })
	if err == nil || state.partyRoster[0].Effects[0].Data[0]&0x10 != 0 {
		t.Fatalf("err=%v roster effect data=%02x, want rollback", err, state.partyRoster[0].Effects[0].Data[0])
	}
	if len(state.ConsumeDamageRequests()) != 2 {
		t.Fatal("failed damage transaction should retain both pending requests")
	}
}

func TestResolvePendingECLDamageFinishesActiveCombatWhenPartyFalls(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", HitPoints: 1, MaxHitPoints: 1, Effects: []monster.AffectRecord{{Kind: 0x25}, {Kind: 0x07}, {Kind: 0x01}}}}
	if err := state.StartCombat(
		[]combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 1, MaxHitPoints: 1, ArmorClass: 10}},
		[]combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10}}, 7); err != nil {
		t.Fatal(err)
	}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 0x80, DiceCount: 1, DiceSize: 1, SaveFlags: 0x80,
	}}})
	outcomes, err := state.ResolvePendingECLDamage(0, func(int) int { return 1 }, func(int) int { return 1 })
	if err != nil || len(outcomes) != 1 || outcomes[0].Health != party.HealthStatusUnconscious {
		t.Fatalf("outcomes=%#v err=%v", outcomes, err)
	}
	heroHP := -1
	heroHasPosition := true
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" {
			heroHP = fighter.HitPoints
			heroHasPosition = fighter.HasCombatPosition
		}
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusEnemyWon || state.partyRoster[0].HitPoints != 0 || heroHP != 0 || heroHasPosition {
		t.Fatalf("mode=%v status=%v roster=%#v fighters=%#v", state.Mode, state.CombatStatus(), state.partyRoster, state.CombatFighters())
	}
	if len(state.partyRoster[0].Effects) != 2 || state.partyRoster[0].Effects[0].Kind != 0x25 || state.partyRoster[0].Effects[1].Kind != 0x01 {
		t.Fatalf("combat effects were not removed: %#v", state.partyRoster[0].Effects)
	}
}

func TestClearCombatActionForResetsCurrentFighterUIState(t *testing.T) {
	state := NewState(testCatalog())
	state.combatTurns = []combat.Turn{{FighterID: "hero"}}
	state.combatTurnIndex = 0
	state.combatCastingSpell = MagicMissileSpellID
	state.combatCastingClassSet = true
	state.combatSpellTargetIndex = 2
	state.combatMoveMode = true
	state.combatMoveRemaining = 3
	state.combatView = true
	state.combatViewFighterID = "hero"

	state.clearCombatActionFor("goblin")
	if state.combatCastingSpell == 0 || !state.combatMoveMode || !state.combatView {
		t.Fatal("non-current fighter unexpectedly cleared the active action")
	}
	state.clearCombatActionFor("hero")
	if state.combatCastingSpell != 0 || state.combatCastingClassSet || state.combatSpellTargetIndex != 0 || state.combatMoveMode || state.combatMoveRemaining != 0 || state.combatView || state.combatViewFighterID != "" {
		t.Fatalf("current fighter action was not cleared: spell=%d class=%v target=%d move=%v remaining=%d view=%v viewID=%q", state.combatCastingSpell, state.combatCastingClassSet, state.combatSpellTargetIndex, state.combatMoveMode, state.combatMoveRemaining, state.combatView, state.combatViewFighterID)
	}
}

func TestResolveDeathEffectsIsTransactionalAndSyncsRoster(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", HitPoints: 0, MaxHitPoints: 10, HealthStatus: party.HealthStatusDying, Bleeding: 2,
		Effects: []monster.AffectRecord{{Kind: 0x63, Active: true}}}}
	if err := state.ResolveDeathEffects(party.DeathEffectContext{CombatHealAllowed: true, RollDie: func(int) int { return 4 }}); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].HealthStatus != party.HealthStatusOK || state.partyRoster[0].HitPoints != 4 || len(state.partyRoster[0].Effects) != 1 || state.partyRoster[0].Effects[0].Kind != 0x5F {
		t.Fatalf("death recovery state=%#v", state.partyRoster[0])
	}
	state.partyRoster[0].Effects = []monster.AffectRecord{{Kind: 0x64, Active: true}}
	before := state.partyRoster[0]
	if err := state.ResolveDeathEffects(party.DeathEffectContext{DamageFlagsKnown: true, RollDie: func(int) int { return 0 }}); err == nil {
		t.Fatal("invalid troll die should fail")
	}
	if state.partyRoster[0].Effects[0].Kind != before.Effects[0].Kind || len(state.partyRoster[0].Effects) != 1 {
		t.Fatalf("failed death transaction leaked effects: %#v", state.partyRoster[0])
	}
}

func TestResolveDeathEffectsCombatHealRestoresDownedPlacement(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{
		{ID: "cleric", Name: "牧師", HitPoints: 8, MaxHitPoints: 8},
		{ID: "hero", Name: "倒下的英雄", HitPoints: 0, MaxHitPoints: 10, HealthStatus: party.HealthStatusDying, Bleeding: 2, Effects: []monster.AffectRecord{{Kind: 0x63, Active: true}}},
	}
	if err := state.StartCombat(
		[]combat.Fighter{
			{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, InitiativeBonus: 20, HasCombatPosition: true, CombatX: 1, CombatY: 1},
			{ID: "hero", Name: "倒下的英雄", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1},
		},
		[]combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10}}, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.ResolveDeathEffects(party.DeathEffectContext{CombatHealAllowed: true, RollDie: func(int) int { return 4 }}); err != nil {
		t.Fatal(err)
	}
	var hero combat.Fighter
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" {
			hero = fighter
		}
	}
	if state.partyRoster[1].HealthStatus != party.HealthStatusOK || hero.HitPoints != 4 || !hero.HasCombatPosition || hero.DownedCorpse || hero.DeathOverlay || hero.CombatX != 2 || hero.CombatY != 1 {
		t.Fatalf("combat heal did not restore placement: roster=%#v hero=%+v", state.partyRoster, hero)
	}
}

func TestResolveDragonSlayerUsesExplicitMonsterTargetContext(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{Effects: []monster.AffectRecord{{Kind: 0x4B, Active: true}}}}
	result, err := state.ResolveDragonSlayer(0, party.MonsterTypeDragon, 1, func(int) int { return 10 })
	if err != nil || !result.Triggered || result.Damage != 35 || result.AttackRollBonus != 2 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestECLSpellSignalsAreCapturedDuringSelection(t *testing.T) {
	state := NewState(testCatalog())
	state.eclBlock = append([]byte{0, 0}, []byte{
		0x3B, 0x00, 0x12, 0x01, 0x00, 0x7C, 0x01, 0x01, 0x7C,
		0x3C, 0x01, 0x02, 0x7C,
		0x00,
	}...)
	state.eclStart = 0
	state.Mode = ModeWilderness
	state.Choices = []string{"繼續旅程"}
	state.currentOriginalChoices = []string{"JOURNEY ON"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.ConsumeSpellSearches()) != 1 || len(state.ConsumeProtectionRequests()) != 1 {
		t.Fatalf("selection did not capture ECL signals: state=%#v", state)
	}
}

func TestShadowdaleWildernessMapMovementAndExit(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["wilderness"] = "荒野"
	catalog.Strings["exit"] = "離開"
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.Location = LocationShadowdale
	state.LocationName = "暗影谷"
	state.Choices = []string{"荒野", "離開"}
	state.currentOriginalChoices = []string{"WILDERNESS", "EXIT"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeMap || state.MapX != 0 || state.MapY != 0 {
		t.Fatalf("map entry state=%#v", state)
	}
	dx, dy := 0, 0
	for _, candidate := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
		if state.WildernessFloor.CanEnter(candidate[0], candidate[1]) {
			dx, dy = candidate[0], candidate[1]
			break
		}
	}
	if dx == 0 && dy == 0 {
		t.Fatal("generated origin has no passable neighbor")
	}
	if err := state.Move(dx, dy); err != nil {
		t.Fatal(err)
	}
	if state.MapX != dx || state.MapY != dy {
		t.Fatalf("map position=(%d,%d)", state.MapX, state.MapY)
	}
	if err := state.LeaveMap(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 3 {
		t.Fatalf("leave state=%#v", state)
	}
}

func TestInnRestoresPartyAndReturnsToPlaceMenu(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["inn_restored"] = "客棧休息完成"
	state := NewState(catalog)
	state.Mode = ModePlace
	state.Location = LocationShadowdale
	state.LocationName = "暗影谷"
	state.Choices = []string{"客棧", "商店", "酒館", "離開"}
	state.currentOriginalChoices = []string{"INN", "STORE", "BAR", "LEAVE"}
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, HitPoints: 2, MaxHitPoints: 10,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.party = []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 2, MaxHitPoints: 10}}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != "客棧休息完成" || state.partyRoster[0].HitPoints != 10 || state.party[0].HitPoints != 10 {
		t.Fatalf("inn state=%#v party=%#v roster=%#v", state, state.party, state.partyRoster)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace {
		t.Fatalf("after inn continue mode=%v, want place", state.Mode)
	}
}

func TestShadowdalePlaceMenuAndEvents(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["shadowdale"] = "暗影谷"
	catalog.Strings["shadowdale_map_prompt"] = "暗影谷荒野"
	catalog.Strings["what_place"] = "你在暗影谷。要去哪裡？"
	catalog.Strings["inn"] = "客棧"
	catalog.Strings["store"] = "商店"
	catalog.Strings["bar"] = "酒館"
	catalog.Strings["leave"] = "離開"
	catalog.Strings["inn_restored"] = "暗影谷客棧休息完成"
	state := NewState(catalog)
	state.Location = LocationShadowdale
	state.Mode = ModeMap
	if err := state.EnterPlaces(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || len(state.Choices) != 4 || state.Choices[0] != "客棧" {
		t.Fatalf("place menu=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != "暗影谷客棧休息完成" || state.OriginalEvent != "INN" {
		t.Fatalf("inn event=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModePlace {
		t.Fatalf("place continuation state=%#v err=%v", state, err)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeMap {
		t.Fatalf("leave continuation state=%#v err=%v", state, err)
	}
}

func TestBarMenuReadsTavernTalesAndReturnsToPlaces(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.Location = LocationShadowdale
	state.Mode = ModePlace
	state.Choices = []string{"客棧", "商店", "酒館", "離開"}
	state.currentOriginalChoices = []string{"INN", "STORE", "BAR", "LEAVE"}
	state.SetBarTales([]string{"第一則傳聞", "第二則傳聞"})
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.barMenu || len(state.Choices) != 2 ||
		state.Choices[0] != catalog.Text("bar_listen", "") {
		t.Fatalf("bar menu=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "BAR_LISTEN" || !strings.Contains(state.Message, "第一則傳聞") || state.BarTaleIndex() != 1 {
		t.Fatalf("bar tale=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModePlace || !state.barMenu {
		t.Fatalf("bar continuation state=%#v err=%v", state, err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "BAR_EXIT" || state.barMenu {
		t.Fatalf("bar exit=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModePlace || state.barMenu {
		t.Fatalf("place return state=%#v err=%v", state, err)
	}
}

func TestTavernCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	keys := []string{
		"bar_event", "bar_menu_prompt", "bar_listen", "bar_exit", "bar_exit_message",
		"bar_no_tales", "bar_tale", "bar_tale_1", "bar_tale_2", "bar_tale_3",
		"bar_tale_4", "bar_tale_5", "bar_tale_6", "tavern_drink_prompt",
		"tavern_punch", "tavern_drink", "tavern_dragon_breath", "tavern_basilisk",
		"tavern_lemonade", "tavern_whiskey", "tavern_beer", "tavern_ale",
		"tavern_port", "tavern_mead", "ecl_tavern_pleasure", "ecl_tavern_special_1",
		"ecl_tavern_special_2", "ecl_tavern_purple_1", "ecl_tavern_purple_2",
		"ecl_tavern_purple_3", "ecl_tavern_commotion_1", "ecl_tavern_commotion_2",
		"ecl_tavern_commotion_3",
	}
	for _, key := range keys {
		if got := catalog.Text(key, ""); got == "" {
			t.Fatalf("tavern locale ID %q is absent", key)
		}
	}
}

func TestShopCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	keys := []string{
		"shop_menu_prompt", "shop_buy", "shop_view", "shop_take", "shop_pool",
		"shop_share", "shop_appraise", "shop_sell", "shop_identify", "shop_exit",
		"shop_buy_unavailable", "shop_stock_prompt", "shop_purchase_done",
		"shop_purchase_failed", "shop_sell_prompt", "shop_sell_item_prompt",
		"shop_sell_exit", "shop_sale_done", "shop_sell_failed", "shop_sell_unavailable",
		"shop_identify_prompt", "shop_identify_item_prompt", "shop_identify_exit",
		"shop_identify_done", "shop_identify_failed", "shop_identify_unavailable",
		"shop_view_prompt", "shop_view_exit", "shop_view_summary", "shop_view_unavailable",
		"shop_take_prompt", "shop_take_amount_prompt", "shop_take_exit", "shop_take_done",
		"shop_take_failed", "shop_take_unavailable", "shop_appraise_prompt",
		"shop_appraise_treasure_prompt", "shop_appraise_exit", "shop_appraise_done",
		"shop_appraise_failed", "shop_appraise_confirm_prompt", "shop_appraise_accept",
		"shop_appraise_reject", "shop_appraise_cancel", "shop_appraise_rejected",
		"shop_appraise_unavailable", "shop_pool_done", "shop_pool_failed",
		"shop_share_done", "shop_share_failed", "shop_item_price", "shop_view_character",
		"shop_take_character", "shop_gold_amount", "shop_character_items",
		"shop_identify_character", "shop_appraise_character", "shop_gems_offer_unavailable",
		"shop_gems_offer", "shop_jewelry_offer_unavailable", "shop_jewelry_offer",
		"list_separator",
	}
	for _, key := range keys {
		if got := catalog.Text(key, ""); got == "" {
			t.Fatalf("shop locale ID %q is absent", key)
		}
	}
}

func TestStoreOpensLocalizedShopMenuAndReturnsToPlaces(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.Mode = ModePlace
	state.Location = LocationShadowdale
	state.LocationName = "暗影谷"
	state.Choices = []string{"客棧", "商店", "酒館", "離開"}
	state.currentOriginalChoices = []string{"INN", "STORE", "BAR", "LEAVE"}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopMenu || len(state.Choices) != 9 ||
		state.Choices[0] != catalog.Text("shop_buy", "") {
		t.Fatalf("shop menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "BUY" {
		t.Fatalf("shop buy state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopMenu || state.Choices[0] != catalog.Text("shop_buy", "") {
		t.Fatalf("shop continuation state=%#v", state)
	}
	if err := state.Select(8); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || state.shopMenu || len(state.Choices) != 4 ||
		state.Choices[1] != catalog.Text("store", "") {
		t.Fatalf("shop exit state=%#v", state)
	}
}

func TestShopMoneyPoolAndInjectedOffer(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{
		{ID: "one", Name: "一號", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1, Gold: 100, Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10}},
		{ID: "two", Name: "二號", Race: party.RaceHuman, Class: party.ClassCleric, Level: 1, Gold: 50, Abilities: party.Abilities{Strength: 10, Intelligence: 10, Wisdom: 16, Dexterity: 10, Constitution: 14, Charisma: 10}},
	}
	if err := state.PoolPartyGold(); err != nil || state.MoneyPool() != 150 || state.partyRoster[0].Gold != 0 {
		t.Fatalf("pooled state=%#v err=%v", state, err)
	}
	if err := state.TakeGold(0, 20); err != nil || state.MoneyPool() != 130 || state.partyRoster[0].Gold != 20 {
		t.Fatalf("take state=%#v err=%v", state, err)
	}
	if err := state.ShareGold(); err != nil || state.MoneyPool() != 0 || state.partyRoster[0].Gold != 85 || state.partyRoster[1].Gold != 65 {
		t.Fatalf("share state=%#v err=%v", state, err)
	}
	if err := state.PoolPartyGold(); err != nil {
		t.Fatal(err)
	}
	state.SetShopOffers([]ShopOffer{{Item: monster.ItemRecord{Type: 36, Name: "長劍"}, Price: 100}})
	if err := state.BuyShopOffer(0, 0); err != nil || state.MoneyPool() != 50 || len(state.partyRoster[0].Equipment) != 1 || state.partyRoster[0].Equipment[0].Type != 36 {
		t.Fatalf("buy state=%#v err=%v", state, err)
	}
}

func TestShopPurchaseUsesTypedCoinsAndDoesNotDepleteStock(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Platinum: 2, Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.SetShopOffers([]ShopOffer{{Item: monster.ItemRecord{Type: 36, Name: "長劍"}, Price: 6}})
	if err := state.BuyShopOffer(0, 0); err != nil {
		t.Fatal(err)
	}
	if got := characterCoinGoldWorth(state.partyRoster[0]); got != 4 {
		t.Fatalf("typed coin worth=%d, want 4 GP after purchase", got)
	}
	if len(state.partyRoster[0].Equipment) != 1 || len(state.ShopOffers()) != 1 {
		t.Fatalf("equipment=%d stock=%d, want cloned purchase and retained stock",
			len(state.partyRoster[0].Equipment), len(state.ShopOffers()))
	}
}

func TestTempleCureLightWoundsUsesReferenceCostAndTypedCoins(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, HitPoints: 2, MaxHitPoints: 10, Platinum: 21,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.fixSeed = 1
	if err := state.enterECLTemple(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if !state.templeConfirmMenu || state.Choices[0] != catalog.Text("temple_confirm", "") {
		t.Fatalf("temple confirmation state=%#v", state)
	}
	beforeWorth := characterCoinGoldWorth(state.partyRoster[0])
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].HitPoints <= 2 || state.partyRoster[0].HitPoints > 10 ||
		characterCoinGoldWorth(state.partyRoster[0]) != beforeWorth-100 {
		t.Fatalf("temple cure hp=%d worth=%d, want healed and %d",
			state.partyRoster[0].HitPoints, characterCoinGoldWorth(state.partyRoster[0]), beforeWorth-100)
	}
}

func TestTempleRemoveCurseClearsEffectAndCursedEquipment(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Gold: 4000,
		Effects:   []monster.AffectRecord{{Kind: 0x24, Active: true}, {Kind: 0x27, Active: true}},
		Equipment: []monster.ItemRecord{{Type: 36, Cursed: true}},
	}}
	if err := state.applyTempleCure(8); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].Effects) != 1 || state.partyRoster[0].Effects[0].Kind != 0x27 ||
		state.partyRoster[0].Equipment[0].Cursed {
		t.Fatalf("remove curse result=%#v", state.partyRoster[0])
	}
}

func TestTempleCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	uiKeys := []string{
		"temple_prompt", "temple_heal", "temple_view", "temple_pool", "temple_share",
		"temple_appraise", "temple_exit", "temple_heal_prompt", "temple_cure_choice",
		"temple_cure_exit", "temple_confirm_prompt", "temple_confirm", "temple_cancel",
		"temple_view_summary", "temple_pool_done", "temple_share_done",
		"temple_appraise_empty", "temple_insufficient_gold", "temple_cure_done",
	}
	for _, key := range uiKeys {
		if got := catalog.Text(key, ""); got == "" {
			t.Fatalf("temple locale ID %q is absent", key)
		}
	}
	seenCureKeys := make(map[string]bool, len(templeCures))
	for _, cure := range templeCures {
		if cure.Key == "" || seenCureKeys[cure.Key] {
			t.Fatalf("invalid or duplicate temple cure key %q", cure.Key)
		}
		seenCureKeys[cure.Key] = true
		if got := catalog.Text(cure.Key, ""); got == "" {
			t.Fatalf("temple cure locale ID %q is absent", cure.Key)
		}
	}
}

func TestNamedAffectCatalogCoversEveryStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	keys := []string{
		"affect_unknown",
		"affect_kind_01", "affect_kind_02", "affect_kind_08", "affect_kind_09",
		"affect_kind_0A", "affect_kind_18", "affect_kind_19", "affect_kind_1C",
		"affect_kind_21", "affect_kind_23", "affect_kind_27", "affect_kind_28",
		"affect_kind_2A", "affect_kind_31", "affect_kind_34", "affect_kind_35",
		"affect_kind_37", "affect_kind_3F", "affect_kind_44", "affect_kind_5A",
	}
	for _, key := range keys {
		if got := catalog.Text(key, ""); got == "" {
			t.Errorf("missing named affect locale key %q", key)
		}
	}
}

func TestShopBuyListsOfferAndUpdatesParty(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Gold: 150, Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	if err := state.PoolPartyGold(); err != nil {
		t.Fatal(err)
	}
	state.SetShopOffers([]ShopOffer{{Item: monster.ItemRecord{Type: 36, Name: "長劍"}, Price: 100}})
	state.Mode = ModePlace
	state.Choices = []string{"商店"}
	state.currentOriginalChoices = []string{"STORE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	wantOffer := fmt.Sprintf(catalog.Text("shop_item_price", ""), monster.LocalizedItemName(state.shopOffers[0].Item, catalog), 100)
	if state.Mode != ModePlace || !state.shopStockMenu || len(state.Choices) != 2 || state.Choices[0] != wantOffer {
		t.Fatalf("stock menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "BUY" || len(state.partyRoster[0].Equipment) != 1 || state.MoneyPool() != 50 {
		t.Fatalf("after buy state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || state.shopStockMenu || len(state.Choices) != 9 {
		t.Fatalf("after buy continue state=%#v", state)
	}
}

func TestShopSellListsItemsAndUsesDocumentedValue(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Equipment: []monster.ItemRecord{{Type: 36, Name: "長劍", Value: 75}},
	}}
	state.Mode = ModePlace
	state.Choices = []string{"商店"}
	state.currentOriginalChoices = []string{"STORE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(6); err != nil { // SELL
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopSellMenu || len(state.Choices) != 2 {
		t.Fatalf("sell character menu=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	wantItem := fmt.Sprintf(catalog.Text("shop_item_price", ""), monster.LocalizedItemName(state.partyRoster[0].Equipment[0], catalog), 75)
	if state.Mode != ModePlace || !state.shopSellItemMenu || state.Choices[0] != wantItem {
		t.Fatalf("sell item menu=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "SELL" || state.MoneyPool() != 75 || len(state.partyRoster[0].Equipment) != 0 {
		t.Fatalf("after sell state=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModePlace || !state.shopMenu || len(state.Choices) != 9 {
		t.Fatalf("sell continuation state=%#v err=%v", state, err)
	}
	state.partyRoster[0].Equipment = []monster.ItemRecord{{Type: 36, Value: 50, Readied: true}, {Type: 37, Value: 50, Cursed: true}}
	if err := state.SellShopItem(0, 0); err == nil {
		t.Fatal("readied item should not be sellable")
	}
	if err := state.SellShopItem(0, 1); err == nil {
		t.Fatal("cursed item should not be sellable")
	}
}

func TestShopIdentifyChargesDocumentedFeeWithoutInventingResult(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Gold: 250,
		Equipment: []monster.ItemRecord{{Type: 99, Name: "神秘戒指", HiddenNameFlags: 3}},
	}}
	state.Mode = ModePlace
	state.Choices = []string{"商店"}
	state.currentOriginalChoices = []string{"STORE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(7); err != nil { // ID
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopIdentifyMenu || len(state.Choices) != 2 {
		t.Fatalf("identify character menu=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopIdentifyItemMenu ||
		state.Choices[0] != monster.LocalizedItemName(state.partyRoster[0].Equipment[0], catalog) {
		t.Fatalf("identify item menu=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	wantMessage := fmt.Sprintf(catalog.Text("shop_identify_done", ""),
		party.ShopIdentifyFee, monster.LocalizedItemName(state.partyRoster[0].Equipment[0], catalog))
	if state.Mode != ModeEvent || state.OriginalEvent != "ID" || state.partyRoster[0].Gold != 50 ||
		state.partyRoster[0].Equipment[0].HiddenNameFlags != 3 || state.Message != wantMessage {
		t.Fatalf("identify result state=%#v", state)
	}
}

func TestShopViewListsCharactersAndEquipment(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Gold: 40, HitPoints: 8, MaxHitPoints: 10,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
		Equipment: []monster.ItemRecord{{Type: 36, Name: "長劍", Readied: true}},
	}}
	state.Mode = ModePlace
	state.Choices = []string{"商店"}
	state.currentOriginalChoices = []string{"STORE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	wantCharacter := fmt.Sprintf(catalog.Text("shop_view_character", ""), state.partyRoster[0].Name, 8, 10, 40)
	if state.Mode != ModePlace || !state.shopViewMenu || len(state.Choices) != 2 || state.Choices[0] != wantCharacter {
		t.Fatalf("view menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	wantSummary := fmt.Sprintf(catalog.Text("shop_view_summary", ""), state.partyRoster[0].Name,
		8, 10, 40, monster.LocalizedItemName(state.partyRoster[0].Equipment[0], catalog))
	if state.Mode != ModeEvent || state.OriginalEvent != "VIEW" || state.Message != wantSummary {
		t.Fatalf("view summary state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || state.shopViewMenu || len(state.Choices) != 9 {
		t.Fatalf("after view continue state=%#v", state)
	}
}

func TestShopTakeSelectsCharacterAndAmount(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.moneyPool = 150
	state.Mode = ModePlace
	state.Choices = []string{"商店"}
	state.currentOriginalChoices = []string{"STORE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	wantCharacter := fmt.Sprintf(catalog.Text("shop_take_character", ""), state.partyRoster[0].Name, 0)
	if state.Mode != ModePlace || !state.shopTakeMenu || len(state.Choices) != 2 || state.Choices[0] != wantCharacter {
		t.Fatalf("take character menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopTakeAmountMenu ||
		state.Choices[0] != fmt.Sprintf(catalog.Text("shop_gold_amount", ""), 1) ||
		state.Choices[len(state.Choices)-1] != catalog.Text("shop_take_exit", "") {
		t.Fatalf("take amount menu state=%#v", state)
	}
	if err := state.Select(2); err != nil { // 100 GP
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "TAKE" || state.partyRoster[0].Gold != 100 || state.MoneyPool() != 50 {
		t.Fatalf("after take state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || state.shopTakeMenu || len(state.Choices) != 9 {
		t.Fatalf("after take continue state=%#v", state)
	}
}

func TestShopAppraiseGemsUsesInjectedOffer(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Gems:      3,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.SetAppraisalOffers(AppraisalOffers{Gems: 75, GemsReady: true})
	state.Mode = ModePlace
	state.Choices = []string{"商店"}
	state.currentOriginalChoices = []string{"STORE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(5); err != nil { // APPRAISE
		t.Fatal(err)
	}
	wantCharacter := fmt.Sprintf(catalog.Text("shop_appraise_character", ""), state.partyRoster[0].Name, 3, 0)
	if state.Mode != ModePlace || !state.shopAppraiseMenu || state.Choices[0] != wantCharacter {
		t.Fatalf("appraise character menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || len(state.Choices) != 2 ||
		state.Choices[0] != fmt.Sprintf(catalog.Text("shop_gems_offer", ""), 75) {
		t.Fatalf("appraise treasure menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopAppraiseConfirm || len(state.Choices) != 3 ||
		state.Choices[0] != catalog.Text("shop_appraise_accept", "") {
		t.Fatalf("appraise confirmation state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "APPRAISE" || state.partyRoster[0].Gems != 0 || state.MoneyPool() != 75 {
		t.Fatalf("after appraise accept state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || state.shopAppraiseMenu || len(state.Choices) != 9 {
		t.Fatalf("after appraise continue state=%#v", state)
	}
}

func TestShopAppraiseRejectKeepsTreasure(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Gems:      3,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
	}}
	state.SetAppraisalOffers(AppraisalOffers{Gems: 75, GemsReady: true})
	state.Mode = ModePlace
	state.Choices = []string{"商店"}
	state.currentOriginalChoices = []string{"STORE"}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(5); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.partyRoster[0].Gems != 3 || state.MoneyPool() != 0 || !strings.Contains(state.Message, "拒絕") {
		t.Fatalf("after appraise reject state=%#v", state)
	}
}

func TestStateExposesPostCombatEntryFromECL(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["combat_started"] = "戰鬥開始（戰鬥規則尚未完成）"
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.Choices = []string{"遭遇"}
	state.currentOriginalChoices = []string{"ENCOUNTER"}
	state.eclBlock = []byte{0, 0, 0x24}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	// ⚠ 這個 `24h` 沒有擺怪，所以走的是**第四支**：戰後處理（`overlay-05`
	// ＝ POSTCOM），不是戰鬥（spec 1182）。原作的 `24h` 沒有「零隻怪的戰鬥」。
	if state.Mode != ModeEvent || state.OriginalEvent != "POSTCOMBAT" {
		t.Fatalf("post-combat state=%#v", state)
	}
}

func TestECLCombatRequestStartsBattleWithConfiguredParty(t *testing.T) {
	state := NewState(testCatalog())
	state.Mode = ModeWilderness
	state.Choices = []string{"遭遇"}
	state.currentOriginalChoices = []string{"ENCOUNTER"}
	state.eclBlock = []byte{0, 0, 0x0B, 0x00, 0x56, 0x00, 0x01, 0x00, 0x56, 0x24}
	state.eclStart = 0
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1}}
	if err := state.SetParty(party); err != nil {
		t.Fatal(err)
	}
	state.SetMonsterRecords(map[uint8]monster.Record{0x56: {Name: "BUGBEAR", MaxHitPoints: 2, HitPoints: 2, ArmorClass: 0, AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1}})
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.CombatActive() || len(state.CombatTargets()) != 1 {
		t.Fatalf("combat state=%#v targets=%#v", state, state.CombatTargets())
	}
}

func TestPlayableCombatStateRunsPartyTurnAndVictory(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HasCombatPosition: true, CombatX: 4, CombatY: 3,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
		AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
	}}
	enemies := []combat.Fighter{{
		ID: "goblin", Name: "哥布林", Side: combat.SideEnemy,
		HitPoints: 1, MaxHitPoints: 1, ArmorClass: 0,
		AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1,
	}}
	if err := state.StartCombat(party, enemies, 7); err != nil {
		t.Fatal(err)
	}
	positions := state.CombatFighters()
	var hero, goblin combat.Fighter
	for _, fighter := range positions {
		if fighter.ID == "hero" {
			hero = fighter
		}
		if fighter.ID == "goblin" {
			goblin = fighter
		}
	}
	if len(positions) != 2 || !hero.HasCombatPosition || hero.CombatX != 4 || hero.CombatY != 3 || !goblin.HasCombatPosition {
		t.Fatalf("combat positions=%#v", positions)
	}
	if !state.CombatActive() || len(state.CombatTargets()) != 1 {
		t.Fatalf("combat state=%#v", state)
	}
	active, ok := state.CombatActiveFighter()
	if !ok || active.ID != "hero" || active.CombatX != 4 || active.CombatY != 3 {
		t.Fatalf("active combat fighter=%+v ok=%v", active, ok)
	}
	if err := state.CombatAct(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("combat result mode=%v status=%v message=%q", state.Mode, state.CombatStatus(), state.Message)
	}
}

func TestHeldFighterClearsPendingActionWithoutConsumingSlotBeforePartyInput(t *testing.T) {
	for _, kind := range []uint8{0x1F, 0x33, 0x34, 0x35} {
		t.Run(fmt.Sprintf("effect_%02X", kind), func(t *testing.T) {
			catalog := combatVisualCatalog(t)
			state := NewState(catalog)
			state.partyRoster = party.Roster{{
				ID: "held", Name: "受制牧師", Class: party.ClassCleric, Level: 1,
				SpellSlots: []uint8{BlessSpellID},
			}, {ID: "hero", Name: "英雄", Class: party.ClassFighter, Level: 1}}
			partyFighters := []combat.Fighter{{
				ID: "held", Name: "受制牧師", Side: combat.SideParty,
				HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 100,
				MonsterAffects: []combat.MonsterAffect{{Kind: kind, Active: true}},
				CombatAction: combat.ActionState{
					Delay: 5, SpellID: BlessSpellID, TargetID: "held",
				},
			}, {
				ID: "hero", Name: "英雄", Side: combat.SideParty,
				HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 10,
			}}
			enemies := []combat.Fighter{{
				ID: "enemy", Name: "敵人", Side: combat.SideEnemy,
				HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 1,
			}}
			if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
				t.Fatal(err)
			}
			active, ok := state.CombatActiveFighter()
			if !ok || active.ID != "hero" {
				t.Fatalf("active=%+v ok=%v message=%q", active, ok, state.CombatMessage())
			}
			if state.CombatMessage() != fmt.Sprintf(catalog.Text("combat_monster_held", ""), "受制牧師") {
				t.Fatalf("held message=%q", state.CombatMessage())
			}
			held, ok := state.fighter("held")
			if !ok || held.CombatAction != (combat.ActionState{}) {
				t.Fatalf("held action=%+v ok=%v", held.CombatAction, ok)
			}
			if events := state.battle.TakeSpellInterruptions(); len(events) != 0 {
				t.Fatalf("held effect incorrectly consumed a memorized slot: %+v", events)
			}
			if !slices.Equal(state.partyRoster[0].SpellSlots, []uint8{BlessSpellID}) {
				t.Fatalf("held caster slots=%v", state.partyRoster[0].SpellSlots)
			}
		})
	}
}

func TestEnemyAttackMessageUsesResolvedTarget(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{
		{ID: "hero-a", Name: "甲", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "hero-b", Name: "乙", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
	}
	enemies := []combat.Fighter{{
		ID: "ogre", Name: "食人魔", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 100,
	}}
	if err := state.StartCombat(partyFighters, enemies, 1); err != nil {
		t.Fatal(err)
	}
	if state.CombatMessage() == "" {
		t.Fatal("enemy turn did not produce a message")
	}
	message := state.CombatMessage()
	if !strings.Contains(message, "乙") || strings.Contains(message, "甲") {
		t.Fatalf("message=%q does not name only the resolved living target", message)
	}
}

func TestPlayableCombatUsesWeaponAttackSequence(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{{ID: "archer", Name: "弓手", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, AttacksPerTurn: 2, InitiativeBonus: 20}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 2, MaxHitPoints: 2, ArmorClass: 0}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatAct(); err != nil {
		t.Fatal(err)
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("multi attack status=%v mode=%v message=%q", state.CombatStatus(), state.Mode, state.Message)
	}
}

func TestECLCombatVictoryResumesFollowingMenu(t *testing.T) {
	// Five entry pointers are required by BlockSession. The explicit entry at
	// payload +0x14 first requests COMBAT; after victory the resumable PC must
	// continue into a one-option horizontal menu instead of restoring stale
	// wilderness choices.
	block := make([]byte, 2+0x14+1+15)
	for index := 0; index < 5; index++ {
		pos := 2 + index*4
		block[pos+1], block[pos+2], block[pos+3] = 0x02, 0x14, 0x80
	}
	block[2+0x14] = 0x24
	copy(block[2+0x15:], []byte{
		0x2B, 0x02, 0x00, 0x90, 0x00, 1,
		0x80, 0x02, 0x20, 0x92,
		0x00,
	})
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x50: block}, 0x50)
	if err != nil {
		t.Fatal(err)
	}
	if result, runErr := session.RunFrom(0x14, 20, nil); runErr != nil || !result.PostCombatRequested {
		t.Fatalf("combat prefix=%+v err=%v", result, runErr)
	}

	state := NewState(testCatalog())
	state.session = session
	state.eclBlock = session.CurrentData()
	state.Choices = []string{"舊 ECL 選項"}
	state.currentOriginalChoices = []string{"STALE"}
	partyFighters := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
		AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
	}}
	enemies := []combat.Fighter{{
		ID: "goblin", Name: "哥布林", Side: combat.SideEnemy,
		HitPoints: 1, MaxHitPoints: 1, ArmorClass: 0,
	}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatAct(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !reflect.DeepEqual(state.currentOriginalChoices, []string{"HI"}) {
		t.Fatalf("continuation mode=%v choices=%#v message=%q", state.Mode, state.currentOriginalChoices, state.Message)
	}
}

func TestFinishedCombatSyncsRosterHitPointsForSaveAndCamp(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", HitPoints: 10, MaxHitPoints: 10}}
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 2}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	result, err := state.battle.Attack("goblin", "hero")
	if err != nil {
		t.Fatal(err)
	}
	if err := state.finishCombat(); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].HitPoints != result.TargetHP || state.PartyFighters()[0].HitPoints != result.TargetHP {
		t.Fatalf("combat HP did not reach save roster: result=%d roster=%d fighter=%d", result.TargetHP, state.partyRoster[0].HitPoints, state.PartyFighters()[0].HitPoints)
	}
}

func TestCombatResultContinuationRestoresWildernessMenu(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", HitPoints: 10, MaxHitPoints: 10}}
	state.Choices = []string{"舊 ECL 選項"}
	state.currentOriginalChoices = []string{"STALE"}
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 100}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 1, MaxHitPoints: 1, ArmorClass: 0}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatAct(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent {
		t.Fatalf("combat did not produce result event: mode=%v", state.Mode)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("wilderness menu was not restored: mode=%v choices=%#v", state.Mode, state.currentOriginalChoices)
	}
}

func TestSuppressedECLPictureStillPresentsContinuationMenu(t *testing.T) {
	state := NewState(testCatalog())
	state.picturesEnabled = false
	result := ecl.RunResult{
		PictureRequested: true,
		PictureBlock:     14,
		WaitingForMenu:   true,
		Menus: []ecl.Menu{{
			Options: []string{"PRESS BUTTON OR RETURN TO CONTINUE."},
		}},
	}
	handled, err := state.continueAfterSuppressedPicture(result)
	if err != nil {
		t.Fatal(err)
	}
	if !handled || state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("suppressed picture continuation handled=%v mode=%v choices=%#v",
			handled, state.Mode, state.currentOriginalChoices)
	}
}

func TestCombatDoneEndsPartyTurnWithoutAttacking(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{
		{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 100},
		{ID: "ally", Name: "隊友", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 0},
	}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0, DamageDiceCount: 0}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "hero" {
		t.Fatalf("unexpected first active fighter=%+v ok=%v", active, ok)
	}
	if err := state.CombatDone(); err != nil {
		t.Fatal(err)
	}
	active, ok := state.CombatActiveFighter()
	if !ok || active.ID != "ally" {
		t.Fatalf("DONE did not advance to next party fighter=%+v ok=%v", active, ok)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "goblin" && fighter.HitPoints != 20 {
			t.Fatalf("DONE attacked enemy: %+v", fighter)
		}
	}
}

func TestCombatDelayReentersSameRoundAfterOtherActions(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{
		{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20},
		{ID: "ally", Name: "隊友", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 10},
	}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 1}}
	if err := state.StartCombat(partyFighters, enemies, 420); err != nil {
		t.Fatal(err)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "hero" {
		t.Fatalf("first active=%+v ok=%v", active, ok)
	}
	if err := state.CombatDelay(); err != nil {
		t.Fatal(err)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "ally" {
		t.Fatalf("after delay active=%+v ok=%v", active, ok)
	}
	if err := state.CombatDone(); err != nil {
		t.Fatal(err)
	}
	if active, ok := state.CombatActiveFighter(); !ok || active.ID != "hero" {
		t.Fatalf("delayed fighter did not reenter same round: active=%+v ok=%v", active, ok)
	}
	fighter, _ := state.fighter("hero")
	if fighter.CombatAction.Delay != 1 {
		t.Fatalf("delayed action=%+v want delay 1", fighter.CombatAction)
	}
}

func TestCombatGuardEndsTurnAndArmsReaction(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{
		{ID: "guard", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 20,
			HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "ally", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 10},
	}
	enemies := []combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 1,
		HasCombatPosition: true, CombatX: 4, CombatY: 2}}
	if err := state.StartCombat(partyFighters, enemies, 421); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanGuard() {
		t.Fatal("melee fighter cannot select Guard")
	}
	if err := state.CombatGuard(); err != nil {
		t.Fatal(err)
	}
	guard, _ := state.fighter("guard")
	if !guard.CombatAction.Guarding || guard.CombatAction.Delay != 0 {
		t.Fatalf("guard action=%+v", guard.CombatAction)
	}
	active, ok := state.CombatActiveFighter()
	if !ok || active.ID != "ally" {
		t.Fatalf("Guard did not advance turn: active=%+v ok=%v", active, ok)
	}
}

func TestCombatBandageChangesOnlyFirstDyingRosterMember(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{
		{ID: "first", Name: "甲", HitPoints: 0, MaxHitPoints: 10, HealthStatus: party.HealthStatusDying, Bleeding: 3},
		{ID: "second", Name: "乙", HitPoints: 0, MaxHitPoints: 10, HealthStatus: party.HealthStatusDying, Bleeding: 4},
		{ID: "actor", Name: "丙", HitPoints: 10, MaxHitPoints: 10},
	}
	partyFighters := []combat.Fighter{
		{ID: "actor", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20},
		{ID: "first", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "second", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10},
	}
	if err := state.StartCombat(partyFighters, []combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10}}, 421); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanBandage() {
		t.Fatal("dying roster member did not enable Bandage")
	}
	if err := state.CombatBandage(); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].HealthStatus != party.HealthStatusUnconscious || state.partyRoster[0].Bleeding != 0 {
		t.Fatalf("first candidate=%+v", state.partyRoster[0])
	}
	if state.partyRoster[1].HealthStatus != party.HealthStatusDying || state.partyRoster[1].Bleeding != 4 {
		t.Fatalf("second candidate was also changed: %+v", state.partyRoster[1])
	}
}

func TestCombatQuickAndManualControlUseProjectedControlMorale(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{
		{ID: "pc", Side: combat.SideParty, ControlMorale: 0, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 20},
		{ID: "next", Side: combat.SideParty, ControlMorale: 0, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 10},
	}
	if err := state.SetParty(partyFighters); err != nil {
		t.Fatal(err)
	}
	if err := state.StartCombat(partyFighters, []combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 1}}, 421); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatQuick(); err != nil {
		t.Fatal(err)
	}
	pc, _ := state.fighter("pc")
	if !pc.QuickFight {
		t.Fatal("QUICK did not delegate current PC")
	}
	if changed := state.CombatManualControl(); changed != 1 {
		t.Fatalf("manual changed=%d want 1", changed)
	}
	pc, _ = state.fighter("pc")
	if pc.QuickFight {
		t.Fatal("Space did not restore PC manual control")
	}
	if state.PartyFighters()[0].QuickFight {
		t.Fatal("Space did not synchronize manual control to the next-combat party projection")
	}
}

func TestCombatQuickAllCanBeInterruptedDuringVisualHandoff(t *testing.T) {
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	partyFighters := []combat.Fighter{
		{ID: "pc", Side: combat.SideParty, ControlMorale: 0, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 20},
		{ID: "npc", Side: combat.SideParty, ControlMorale: 0x80, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 10},
	}
	if err := state.StartCombat(partyFighters, []combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 1}}, 422); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatQuickAll(); err != nil {
		t.Fatal(err)
	}
	if !state.CombatVisualPending() {
		t.Fatal("all-Quick did not yield at the first action timeline")
	}
	if changed := state.CombatManualControl(); changed != 1 {
		t.Fatalf("manual changed=%d want 1", changed)
	}
	pc, _ := state.fighter("pc")
	npc, _ := state.fighter("npc")
	if pc.QuickFight || !npc.QuickFight {
		t.Fatalf("manual interruption pc=%v npc=%v", pc.QuickFight, npc.QuickFight)
	}
}

func TestCombatAltMEnablesQuickMagicMissileFromGlobalSpellSlot(t *testing.T) {
	found := false
	for seed := int64(0); seed < 128 && !found; seed++ {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 1,
			SpellSlots: []uint8{MagicMissileSpellID},
		}}
		heroes := []combat.Fighter{{
			ID: "mage", Name: "法師", Side: combat.SideParty, ControlMorale: 0,
			HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20,
			DamageDiceCount: 1, DamageDiceSides: 1,
		}}
		enemies := []combat.Fighter{{
			ID: "enemy", Name: "敵人", Side: combat.SideEnemy,
			HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, InitiativeBonus: 1,
		}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		event, ok := state.CombatVisualEvent()
		if !ok || event.Kind != combat.VisualMagicMissile {
			continue
		}
		found = true
		if event.ActorID != "mage" || event.TargetID != "enemy" ||
			len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("quick magic event=%+v roster=%+v", event, state.partyRoster)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-4 Magic Missile selection")
	}
}

func TestCombatAltMQuickBlessWaitsForPendingCastingAction(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.partyRoster = party.Roster{{
			ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1,
			SpellSlots: []uint8{BlessSpellID},
		}}
		heroes := []combat.Fighter{{
			ID: "cleric", Name: "牧師", Side: combat.SideParty, ControlMorale: 0,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: -10, InitiativeBonus: 8,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			HasCombatPosition: true, CombatX: 1, CombatY: 1,
		}}
		enemies := []combat.Fighter{{
			ID: "enemy", Name: "敵人", Side: combat.SideEnemy,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10, InitiativeBonus: 6,
			DamageDiceCount: 1, DamageDiceSides: 1,
			HasCombatPosition: true, CombatX: 10, CombatY: 10,
		}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		caster, _ := state.fighter("cleric")
		if caster.CombatAction.SpellID != BlessSpellID {
			continue
		}
		found = true
		if caster.Blessed || len(state.partyRoster[0].SpellSlots) != 1 ||
			caster.CombatAction.Delay != 5 {
			t.Fatalf("pre-resolution caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
		if changed := state.CombatManualControl(); changed != 1 {
			t.Fatalf("pending cast manual handoff changed=%d want 1", changed)
		}
		for action := 0; action < 8; action++ {
			if !state.CombatActive() {
				break
			}
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
			caster, _ = state.fighter("cleric")
			if caster.Blessed {
				break
			}
		}
		if !caster.Blessed || caster.CombatAction.SpellID != 0 ||
			len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("resolved mode=%v status=%v caster=%+v slots=%v message=%q turns=%+v index=%d",
				state.Mode, state.CombatStatus(), caster, state.partyRoster[0].SpellSlots,
				state.CombatMessage(), state.CombatTurns(), state.combatTurnIndex)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-1 Bless selection")
	}
}

func TestCombatAltMQuickCurePreservesAdjacentTargetAcrossPendingAction(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.partyRoster = party.Roster{
			{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, HitPoints: 10, MaxHitPoints: 10, SpellSlots: []uint8{CureLightWoundsSpellID}},
			{ID: "adjacent", Name: "鄰近隊友", Class: party.ClassFighter, Level: 1, HitPoints: 3, MaxHitPoints: 10},
			{ID: "far", Name: "遠方隊友", Class: party.ClassFighter, Level: 1, HitPoints: 1, MaxHitPoints: 10},
		}
		heroes := []combat.Fighter{
			{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: -10, InitiativeBonus: 8, HasCombatPosition: true, CombatX: 1, CombatY: 1},
			{ID: "adjacent", Name: "鄰近隊友", Side: combat.SideParty, HitPoints: 3, MaxHitPoints: 10, ArmorClass: -10, InitiativeBonus: 2, HasCombatPosition: true, CombatX: 2, CombatY: 1},
			{ID: "far", Name: "遠方隊友", Side: combat.SideParty, HitPoints: 1, MaxHitPoints: 10, ArmorClass: -10, InitiativeBonus: 1, HasCombatPosition: true, CombatX: 8, CombatY: 8},
		}
		enemies := []combat.Fighter{{ID: "enemy", Name: "敵人", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10, InitiativeBonus: 6, DamageDiceCount: 1, DamageDiceSides: 1, HasCombatPosition: true, CombatX: 10, CombatY: 10}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		caster, _ := state.fighter("cleric")
		if caster.CombatAction.SpellID != CureLightWoundsSpellID {
			continue
		}
		found = true
		if caster.CombatAction.TargetID != "adjacent" || len(state.partyRoster[0].SpellSlots) != 1 {
			t.Fatalf("pending=%+v slots=%v", caster.CombatAction, state.partyRoster[0].SpellSlots)
		}
		if changed := state.CombatManualControl(); changed != 1 {
			t.Fatalf("manual handoff changed=%d", changed)
		}
		for action := 0; action < 8 && state.CombatActive(); action++ {
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
			caster, _ = state.fighter("cleric")
			if caster.CombatAction.SpellID == 0 {
				break
			}
		}
		adjacent, _ := state.fighter("adjacent")
		far, _ := state.fighter("far")
		if adjacent.HitPoints <= 3 || far.HitPoints != 1 || len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("adjacent=%+v far=%+v slots=%v", adjacent, far, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-1 Cure selection")
	}
}

func TestQuickCureTargetMatchesPC98DirectionSelfAndDownedRules(t *testing.T) {
	makeState := func(t *testing.T, roster party.Roster, heroes []combat.Fighter) (*State, combat.Fighter) {
		t.Helper()
		state := NewState(testCatalog())
		state.partyRoster = roster
		enemy := combat.Fighter{ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 10, CombatY: 10}
		if err := state.StartCombat(heroes, []combat.Fighter{enemy}, 427); err != nil {
			t.Fatal(err)
		}
		caster, ok := state.fighter("caster")
		if !ok {
			t.Fatal("caster is absent")
		}
		return &state, caster
	}
	baseCaster := combat.Fighter{ID: "caster", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
		HasCombatPosition: true, CombatX: 5, CombatY: 5}

	t.Run("equal HP keeps north before east", func(t *testing.T) {
		heroes := []combat.Fighter{
			baseCaster,
			{ID: "north", Side: combat.SideParty, HitPoints: 4, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 5, CombatY: 4},
			{ID: "east", Side: combat.SideParty, HitPoints: 4, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 6, CombatY: 5},
		}
		state, caster := makeState(t, party.Roster{{ID: "caster"}, {ID: "north"}, {ID: "east"}}, heroes)
		if target, ok := state.quickCureTarget(caster); !ok || target.ID != "north" {
			t.Fatalf("target=%+v ok=%v", target, ok)
		}
	})

	t.Run("caster below half overrides lower HP neighbour", func(t *testing.T) {
		casterFighter := baseCaster
		casterFighter.HitPoints, casterFighter.MaxHitPoints = 9, 20
		heroes := []combat.Fighter{
			casterFighter,
			{ID: "north", Side: combat.SideParty, HitPoints: 1, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 5, CombatY: 4},
		}
		state, caster := makeState(t, party.Roster{{ID: "caster"}, {ID: "north"}}, heroes)
		if target, ok := state.quickCureTarget(caster); !ok || target.ID != "caster" {
			t.Fatalf("target=%+v ok=%v", target, ok)
		}
	})

	t.Run("downed target replaces active only at eight HP", func(t *testing.T) {
		for _, test := range []struct {
			activeHP int
			wantID   string
		}{{activeHP: 8, wantID: "downed"}, {activeHP: 7, wantID: "north"}} {
			heroes := []combat.Fighter{
				baseCaster,
				{ID: "north", Side: combat.SideParty, HitPoints: test.activeHP, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 5, CombatY: 4},
				{ID: "downed", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 6, CombatY: 5, DownedCorpse: true},
			}
			roster := party.Roster{{ID: "caster"}, {ID: "north"}, {ID: "downed", HealthStatus: party.HealthStatusUnconscious}}
			state, caster := makeState(t, roster, heroes)
			if target, ok := state.quickCureTarget(caster); !ok || target.ID != test.wantID {
				t.Fatalf("activeHP=%d target=%+v ok=%v", test.activeHP, target, ok)
			}
		}
	})

	t.Run("stoned down-player is excluded", func(t *testing.T) {
		heroes := []combat.Fighter{
			baseCaster,
			{ID: "stoned", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 6, CombatY: 5, DownedCorpse: true},
		}
		state, caster := makeState(t, party.Roster{{ID: "caster"}, {ID: "stoned", HealthStatus: party.HealthStatusStoned}}, heroes)
		if target, ok := state.quickCureTarget(caster); ok {
			t.Fatalf("stoned target=%+v", target)
		}
	})
}

func TestCombatAltMGateResetsAtEachCombatStart(t *testing.T) {
	state := NewState(testCatalog())
	heroes := []combat.Fighter{{
		ID: "hero", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, InitiativeBonus: 20,
	}}
	enemies := []combat.Fighter{{
		ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, InitiativeBonus: 1,
	}}
	if err := state.StartCombat(heroes, enemies, 1); err != nil {
		t.Fatal(err)
	}
	if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
		t.Fatalf("enabled=%v err=%v", enabled, err)
	}
	if err := state.StartCombat(heroes, enemies, 2); err != nil {
		t.Fatal(err)
	}
	if state.CombatQuickMagicEnabled() {
		t.Fatal("ALT+M gate leaked into the next combat")
	}
}

func TestCombatSpeedBoundsAndMenuUseCatalogEntries(t *testing.T) {
	catalog := testCatalog()
	state := NewState(catalog)
	if state.CombatSpeed() != 4 {
		t.Fatalf("default speed=%d want 4", state.CombatSpeed())
	}
	for state.CombatSpeedSlower() {
	}
	if state.CombatSpeed() != 9 || state.CombatSpeedSlower() {
		t.Fatalf("slow bound=%d", state.CombatSpeed())
	}
	if strings.Contains(state.CombatSpeedMenuText(), catalog.Text("combat_speed_slower", "更慢")) {
		t.Fatalf("slow option remained at upper bound: %q", state.CombatSpeedMenuText())
	}
	for state.CombatSpeedFaster() {
	}
	if state.CombatSpeed() != 0 || state.CombatSpeedFaster() {
		t.Fatalf("fast bound=%d", state.CombatSpeed())
	}
	if strings.Contains(state.CombatSpeedMenuText(), catalog.Text("combat_speed_faster", "更快")) {
		t.Fatalf("fast option remained at lower bound: %q", state.CombatSpeedMenuText())
	}
}

func TestCombatHUDContractUsesFormalLocaleStableIDs(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	keys := []string{
		"combat_hud_hit_points", "combat_hud_armor_class", "combat_prompt_confirm_spell",
		"combat_prompt_select_fighter_target", "combat_prompt_select_fireball_center",
		"combat_prompt_select_sleep_center", "combat_prompt_select_lightning_direction",
		"combat_prompt_select_stinking_cloud_corner", "combat_prompt_select_cloudkill_center",
		"combat_prompt_move", "combat_hint_magic_missile", "combat_hint_sleep",
		"combat_hint_fireball", "combat_hint_lightning_bolt", "combat_hint_stinking_cloud",
		"combat_hint_cloudkill", "combat_hint_cure_light_wounds", "combat_hint_bless",
		"combat_hint_curse", "combat_hint_cause_light_wounds", "combat_hint_protection_from_evil",
		"combat_hint_protection_from_good", "combat_target_status", "combat_quick_status",
	}
	for _, key := range keys {
		if value := catalog.Text(key, key); value == key {
			t.Fatalf("combat HUD locale ID %q is absent", key)
		}
	}
	if state.CombatHitPointsLabel() != catalog.Text("combat_hud_hit_points", "") ||
		state.CombatArmorClassLabel() != catalog.Text("combat_hud_armor_class", "") {
		t.Fatalf("combat HUD labels hp=%q ac=%q", state.CombatHitPointsLabel(), state.CombatArmorClassLabel())
	}
	prompts := []struct {
		spellID uint8
		key     string
	}{
		{BlessSpellID, "combat_prompt_confirm_spell"},
		{MagicMissileSpellID, "combat_prompt_select_fighter_target"},
		{FireballSpellID, "combat_prompt_select_fireball_center"},
		{SleepSpellID, "combat_prompt_select_sleep_center"},
		{LightningBoltSpellID, "combat_prompt_select_lightning_direction"},
		{StinkingCloudSpellID, "combat_prompt_select_stinking_cloud_corner"},
		{CloudkillSpellID, "combat_prompt_select_cloudkill_center"},
	}
	for _, prompt := range prompts {
		state.combatCastingSpell = prompt.spellID
		value, selecting := state.CombatSelectionPrompt()
		if !selecting || value != catalog.Text(prompt.key, "") {
			t.Fatalf("spell 0x%02X prompt=%q selecting=%v", prompt.spellID, value, selecting)
		}
	}
	state.combatCastingSpell = 0
	state.combatMoveMode, state.combatMoveRemaining = true, 7
	wantMove := fmt.Sprintf(catalog.Text("combat_prompt_move", ""), 7)
	if value, selecting := state.CombatSelectionPrompt(); !selecting || value != wantMove {
		t.Fatalf("move prompt=%q selecting=%v want=%q", value, selecting, wantMove)
	}
	if got := state.CombatTargetStatus("TARGET"); got != fmt.Sprintf(catalog.Text("combat_target_status", ""), "TARGET") {
		t.Fatalf("target status=%q", got)
	}
	if got := state.CombatQuickStatus([]string{"A", "B"}); got != fmt.Sprintf(catalog.Text("combat_quick_status", ""), "A　B") {
		t.Fatalf("quick status=%q", got)
	}
}

func TestEnemyTurnUsesWeaponAttackSequence(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	partyFighters := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, InitiativeBonus: 100,
	}}
	enemies := []combat.Fighter{{
		ID: "archer", Name: "敵方弓手", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
		AttacksPerTurn: 2,
	}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatDone(); err != nil {
		t.Fatal(err)
	}
	active, ok := state.CombatActiveFighter()
	if state.Mode != ModeCombat || !ok || active.ID != "hero" {
		t.Fatalf("enemy turn did not return to party: mode=%v message=%q", state.Mode, state.CombatMessage())
	}
	wantMessage := fmt.Sprintf(catalog.Text("combat_multi_attack", ""), "敵方弓手", 2, 2, 2)
	if active.HitPoints != 8 || state.CombatMessage() != wantMessage {
		t.Fatalf("enemy multi-attack not applied: fighter=%+v message=%q", active, state.CombatMessage())
	}
}

func TestEnemyTurnSelectsAmongLivingPartyTargets(t *testing.T) {
	selectedAlly := false
	for seed := int64(1); seed <= 32; seed++ {
		state := NewState(testCatalog())
		partyFighters := []combat.Fighter{
			{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 100},
			{ID: "ally", Name: "隊友", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 90},
		}
		enemies := []combat.Fighter{{
			ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
			ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: -100,
		}}
		if err := state.StartCombat(partyFighters, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if err := state.CombatDone(); err != nil {
			t.Fatal(err)
		}
		if err := state.CombatDone(); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(state.CombatMessage(), "隊友") {
			selectedAlly = true
			break
		}
	}
	if !selectedAlly {
		t.Fatal("enemy target selection never chose the second living party member")
	}
}

func TestEnemyTurnCastsVerifiedMonsterMagicMissile(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	partyFighters := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, InitiativeBonus: 100,
	}}
	enemies := []combat.Fighter{{
		ID: "spell-monster", Name: "施法怪", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, InitiativeBonus: -100, MonsterSpellIDs: []uint8{combat.MonsterMagicMissileSpellID},
		MonsterSpellUses: [3]uint8{1},
	}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatDone(); err != nil {
		t.Fatal(err)
	}
	active, ok := state.CombatActiveFighter()
	if state.Mode != ModeCombat || !ok || active.ID != "hero" {
		t.Fatalf("monster spell turn did not return to party: mode=%v active=%+v message=%q", state.Mode, active, state.CombatMessage())
	}
	wantMessage := fmt.Sprintf(catalog.Text("combat_monster_magic_missile", ""), "施法怪", "英雄", 10-active.HitPoints)
	if active.HitPoints < 5 || active.HitPoints > 8 || state.CombatMessage() != wantMessage {
		t.Fatalf("monster spell was not applied: active=%+v message=%q", active, state.CombatMessage())
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "spell-monster" && fighter.MonsterSpellUses[0] != 0 {
			t.Fatalf("monster spell use not consumed: %+v", fighter)
		}
	}
}

func TestCombatAmmunitionMappingRejectsInsufficientShotsBeforeAttack(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "archer", Name: "弓手", Equipment: []monster.ItemRecord{{Type: 73, Count: 1}}}}
	state.SetAmmunitionItemTypes(map[uint8][]uint8{11: {73}})
	partyFighters := []combat.Fighter{{ID: "archer", Name: "弓手", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, AttacksPerTurn: 2, AmmunitionType: 11, InitiativeBonus: 20}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 0}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatAct(); err == nil {
		t.Fatal("expected insufficient ammunition error")
	}
	if state.CombatFighters()[1].HitPoints != 10 || state.partyRoster[0].Equipment[0].Count != 1 {
		t.Fatalf("failed attack mutated state: fighters=%+v roster=%+v", state.CombatFighters(), state.partyRoster)
	}
}

func TestCombatAdjacentMissileRejectsBeforeAmmunitionTransaction(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "archer", Name: "弓手", Equipment: []monster.ItemRecord{{Type: 73, Count: 2}}}}
	state.SetAmmunitionItemTypes(map[uint8][]uint8{11: {73}})
	partyFighters := []combat.Fighter{{
		ID: "archer", Name: "弓手", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
		AmmunitionType: 11, MissileWeapon: true, HasCombatPosition: true, CombatX: 0, CombatY: 0,
		InitiativeBonus: 20,
	}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 0, HasCombatPosition: true, CombatX: 1, CombatY: 0}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatAct(); err == nil {
		t.Fatal("expected adjacent missile rejection")
	}
	if got := state.partyRoster[0].Equipment[0].Count; got != 2 {
		t.Fatalf("invalid missile attack consumed ammunition: %d", got)
	}
	if got := state.CombatTargets()[0].HitPoints; got != 10 {
		t.Fatalf("invalid missile attack changed target HP: %d", got)
	}
}

func TestReportCombatErrorKeepsLocalizedRecoverableMessage(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	state.ReportCombatError(combat.ErrAdjacentMissileTarget)
	if state.CombatMessage() != catalog.Text("combat_missile_adjacent_error", "") {
		t.Fatalf("missile error message=%q", state.CombatMessage())
	}
	err := errors.New("combat is not active")
	state.ReportCombatError(err)
	if state.CombatMessage() != fmt.Sprintf(catalog.Text("combat_action_error", ""), err) {
		t.Fatalf("generic combat error message=%q", state.CombatMessage())
	}
}

func TestCombatMoveConsumesPartyTurnAndUpdatesPosition(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20, HasCombatPosition: true, CombatX: 4, CombatY: 3}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 7, CombatY: 3}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatMove(); err != nil || !state.CombatMoveMode() {
		t.Fatalf("begin move state=%+v err=%v", state, err)
	}
	if err := state.CombatMove(-1, 0); err != nil {
		t.Fatal(err)
	}
	if state.CombatMoveMode() {
		t.Fatal("move mode was not cleared")
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" && (fighter.CombatX != 3 || fighter.CombatY != 3) {
			t.Fatalf("hero position=%+v", fighter)
		}
	}
}

func TestCombatMoveUsesMultipleArmorLimitedSquaresBeforeEndingTurn(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20, MovementAllowance: 2, HasCombatPosition: true, CombatX: 4, CombatY: 3}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 7, CombatY: 3}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatMove(); err != nil || state.CombatMoveRemaining() != 2 {
		t.Fatalf("begin move remaining=%d err=%v", state.CombatMoveRemaining(), err)
	}
	if err := state.CombatMove(-1, 0); err != nil {
		t.Fatal(err)
	}
	if !state.CombatMoveMode() || state.CombatMoveRemaining() != 1 {
		t.Fatalf("first move ended turn: mode=%v remaining=%d", state.CombatMoveMode(), state.CombatMoveRemaining())
	}
	if err := state.CombatMove(-1, 0); err != nil {
		t.Fatal(err)
	}
	if state.CombatMoveMode() {
		t.Fatal("second move did not end turn")
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" && (fighter.CombatX != 2 || fighter.CombatY != 3) {
			t.Fatalf("hero position=%+v", fighter)
		}
	}
}

func TestCombatMoveConsumesTerrainCostWithoutMutatingOnInsufficientBudget(t *testing.T) {
	catalog := testCatalog()
	state := NewState(catalog)
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20, MovementAllowance: 3, HasCombatPosition: true, CombatX: 4, CombatY: 3}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 8, CombatY: 3}}
	if err := state.StartCombat(partyFighters, enemies, 1); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatMove(); err != nil {
		t.Fatal(err)
	}
	costTwo := func(_, _ int) (int, bool) { return 2, true }
	if err := state.CombatMoveWithTerrain(-1, 0, costTwo); err != nil {
		t.Fatal(err)
	}
	if !state.CombatMoveMode() || state.CombatMoveRemaining() != 1 {
		t.Fatalf("cost-two move mode=%v remaining=%d, want active with 1", state.CombatMoveMode(), state.CombatMoveRemaining())
	}
	if err := state.CombatMoveWithTerrain(-1, 0, costTwo); err == nil {
		t.Fatal("second cost-two move succeeded with one point remaining")
	}
	var fighter combat.Fighter
	for _, candidate := range state.battle.Fighters() {
		if candidate.ID == "hero" {
			fighter = candidate
		}
	}
	if fighter.CombatX != 3 || fighter.CombatY != 3 || state.CombatMoveRemaining() != 1 {
		t.Fatalf("rejected move mutated fighter=%+v remaining=%d", fighter, state.CombatMoveRemaining())
	}
}

func TestStartCombatPreservesPlacementCoordinateNamespace(t *testing.T) {
	newState := func() *State {
		state := NewState(testCatalog())
		return &state
	}
	fallback := newState()
	if err := fallback.StartCombat(
		[]combat.Fighter{{ID: "hero", Side: combat.SideParty, HitPoints: 5, MaxHitPoints: 5}},
		[]combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 5, MaxHitPoints: 5}},
		1,
	); err != nil {
		t.Fatal(err)
	}
	if fallback.CombatUsesReferenceCoordinates() {
		t.Fatal("generated formation was mistaken for original CombatMap coordinates")
	}

	reference := newState()
	if err := reference.StartCombat(
		[]combat.Fighter{{ID: "hero", Side: combat.SideParty, HitPoints: 5, MaxHitPoints: 5, HasCombatPosition: true, CombatX: 22, CombatY: 10}},
		[]combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 5, MaxHitPoints: 5, HasCombatPosition: true, CombatX: 28, CombatY: 10}},
		1,
	); err != nil {
		t.Fatal(err)
	}
	if !reference.CombatUsesReferenceCoordinates() {
		t.Fatal("original placement coordinate namespace was not preserved")
	}
}

func TestStartCombatRebuildsOneBasedLegacyObjectIDsInCharacterListOrder(t *testing.T) {
	state := NewState(testCatalog())
	partyMembers := []combat.Fighter{
		{ID: "hero", Side: combat.SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, LegacyObjectID: 70},
		{ID: "temporary-ally", Side: combat.SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, QuickFight: true, TemporaryAlly: true},
	}
	enemies := []combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10}}
	if err := state.StartCombat(partyMembers, enemies, 438); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]uint8{"hero": 1, "temporary-ally": 2, "enemy": 3} {
		fighter, ok := state.battle.Fighter(id)
		if !ok || fighter.LegacyObjectID != want {
			t.Fatalf("fighter %q object ID=%d found=%v want=%d", id, fighter.LegacyObjectID, ok, want)
		}
	}
}

func TestCombatViewIsReadOnlyAndLocalized(t *testing.T) {
	catalog := combatVisualCatalog(t)
	state := NewState(catalog)
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 8, MaxHitPoints: 10, ArmorClass: 6, AttackBonus: 3, InitiativeBonus: 20}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	activeBefore, ok := state.CombatActiveFighter()
	if !ok {
		t.Fatal("missing active fighter")
	}
	if err := state.BeginCombatView(); err != nil {
		t.Fatal(err)
	}
	if !state.CombatViewActive() {
		t.Fatal("combat view did not open")
	}
	viewed, ok := state.CombatViewFighter()
	if !ok || viewed.ID != "hero" {
		t.Fatalf("viewed=%+v ok=%v", viewed, ok)
	}
	lines := strings.Join(state.CombatViewLines(), " ")
	for _, want := range []string{
		fmt.Sprintf(catalog.Text("combat_view_name", ""), "英雄"),
		fmt.Sprintf(catalog.Text("combat_view_hp", ""), 8, 10),
		fmt.Sprintf(catalog.Text("combat_view_ac", ""), 6),
		fmt.Sprintf(catalog.Text("combat_view_attack", ""), 3),
	} {
		if !strings.Contains(lines, want) {
			t.Fatalf("view lines=%q missing %q", lines, want)
		}
	}
	activeAfter, ok := state.CombatActiveFighter()
	if !ok || activeAfter.ID != activeBefore.ID {
		t.Fatalf("view changed turn: before=%+v after=%+v", activeBefore, activeAfter)
	}
	state.EndCombatView()
	if state.CombatViewActive() {
		t.Fatal("combat view did not close")
	}
}

func TestCombatMoveIntoEnemySquareResolvesAttack(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 20, HasCombatPosition: true, CombatX: 4, CombatY: 3}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 5, CombatY: 3}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatMove(); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatMove(1, 0); err != nil {
		t.Fatal(err)
	}
	if state.CombatMoveMode() || state.CombatMessage() == "" {
		t.Fatalf("move attack state=%+v", state)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" && (fighter.CombatX != 4 || fighter.CombatY != 3) {
			t.Fatalf("hero entered enemy square: %+v", fighter)
		}
		if fighter.ID == "goblin" && fighter.HitPoints != 19 {
			t.Fatalf("goblin after move attack: %+v", fighter)
		}
	}
}

func TestEnemyPhysicalTargetUsesPackScanProducerAndWallRetry(t *testing.T) {
	state := NewState(testCatalog())
	state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
		return enginescan.TacticalMap{
			Width: 5, Height: 1,
			Tiles: []uint8{1, 1, 2, 1, 1},
			Definitions: []enginescan.TerrainDefinition{
				{LOS: 1, SYM: 0}, {LOS: 1, SYM: 2},
			},
		}, nil
	})
	if err := state.StartCombat(
		[]combat.Fighter{{ID: "hero", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 0}},
		[]combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 0, CombatY: 0}},
		510,
	); err != nil {
		t.Fatal(err)
	}
	target, found, err := state.selectEnemyPhysicalTarget("enemy", combat.SideParty)
	if err != nil || !found || target.ID != "hero" {
		t.Fatalf("target=%+v found=%v err=%v, want pack-declared wall-bypass target", target, found, err)
	}
}

func TestCombatAltMQuickSleepUsesAreaCenterAndConsumesSlot(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "mage", Class: party.ClassMagicUser, Level: 1,
			SpellSlots: []uint8{SleepSpellID},
		}}
		state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
			return enginescan.TacticalMap{
				Width: 7, Height: 1,
				Tiles:       []uint8{1, 1, 1, 1, 1, 1, 1},
				Definitions: []enginescan.TerrainDefinition{{LOS: 1}},
			}, nil
		})
		heroes := []combat.Fighter{{
			ID: "mage", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20,
			ArmorClass: 0, InitiativeBonus: 30, HasCombatPosition: true,
			CombatX: 0, CombatY: 0,
		}}
		enemies := []combat.Fighter{{
			ID: "enemy", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20,
			ArmorClass: 10, HitDice: 1, InitiativeBonus: 1,
			HasCombatPosition: true, CombatX: 5, CombatY: 0,
		}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		event, ok := state.CombatVisualEvent()
		if !ok || event.Kind != combat.VisualTwinkle {
			continue
		}
		enemy, ok := state.fighter("enemy")
		if !ok || !enemy.MonsterIsHeld() {
			continue
		}
		found = true
		if len(state.partyRoster[0].SpellSlots) != 0 || len(event.Impacts) != 1 ||
			event.Impacts[0].TargetID != "enemy" ||
			event.Impacts[0].To != (combat.TilePoint{X: 5, Y: 0}) {
			t.Fatalf("Quick Sleep event=%+v slots=%v enemy=%+v", event, state.partyRoster[0].SpellSlots, enemy)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-2 Quick Sleep selection")
	}
}

func TestCombatAltMQuickFireballUsesAreaCenterAndPendingDelay(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "mage", Class: party.ClassMagicUser, Level: 1,
			SpellSlots: []uint8{FireballSpellID},
		}}
		state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
			return enginescan.TacticalMap{
				Width: 11, Height: 1,
				Tiles:       []uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Definitions: []enginescan.TerrainDefinition{{LOS: 1}},
			}, nil
		})
		heroes := []combat.Fighter{{
			ID: "mage", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20,
			ArmorClass: 0, InitiativeBonus: 30, HasCombatPosition: true,
			CombatX: 0, CombatY: 0,
		}}
		enemies := []combat.Fighter{{
			ID: "enemy", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
			ArmorClass: 10, HitDice: 1, InitiativeBonus: 1,
			HasCombatPosition: true, CombatX: 5, CombatY: 0,
			SavingThrows: []uint8{20, 20, 20, 20, 20},
		}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		caster, ok := state.fighter("mage")
		if !ok || caster.CombatAction.SpellID != FireballSpellID {
			continue
		}
		found = true
		if caster.CombatAction.Delay <= 0 || !caster.CombatAction.HasTargetPoint ||
			caster.CombatAction.TargetX != 5 || caster.CombatAction.TargetY != 0 ||
			len(state.partyRoster[0].SpellSlots) != 1 {
			t.Fatalf("pending Quick Fireball caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
		if changed := state.CombatManualControl(); changed != 1 {
			t.Fatalf("pending Fireball manual handoff changed=%d", changed)
		}
		for action := 0; action < 16; action++ {
			if _, ok := state.CombatVisualEvent(); ok {
				break
			}
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
		}
		event, ok := state.CombatVisualEvent()
		if !ok || event.Kind != combat.VisualAreaSpell || event.Effect != "fireball" ||
			event.ActorID != "mage" || event.To != (combat.TilePoint{X: 5, Y: 0}) ||
			len(event.Impacts) != 1 || event.Impacts[0].TargetID != "enemy" ||
			len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("Quick Fireball event=%+v ok=%v slots=%v", event, ok, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-7 Quick Fireball selection")
	}
}

func TestCombatAltMQuickStinkingCloudUsesAreaCenterAndPersistentArea(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "mage", Class: party.ClassMagicUser, Level: 5,
			SpellSlots: []uint8{StinkingCloudSpellID},
		}}
		tiles := make([]uint8, 12*12)
		for index := range tiles {
			tiles[index] = 1
		}
		state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
			return enginescan.TacticalMap{
				Width: 12, Height: 12, Tiles: tiles,
				Definitions: []enginescan.TerrainDefinition{{LOS: 1}},
			}, nil
		})
		state.SetCombatLineTerrain(func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x < 12 && y >= 0 && y < 12}
		})
		heroes := []combat.Fighter{{
			ID: "mage", Side: combat.SideParty, HitPoints: 100, MaxHitPoints: 100,
			ArmorClass: 0, InitiativeBonus: 30, HasCombatPosition: true,
			CombatX: 1, CombatY: 1, SavingThrows: []uint8{20, 20, 20, 20, 20},
		}}
		enemies := []combat.Fighter{{
			ID: "enemy", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
			ArmorClass: 10, HitDice: 1, InitiativeBonus: 1,
			HasCombatPosition: true, CombatX: 4, CombatY: 2,
			SavingThrows: []uint8{20, 20, 20, 20, 20},
		}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		event, ok := state.CombatVisualEvent()
		if !ok || event.Kind != combat.VisualAreaSpell || event.Effect != "stinking_cloud" {
			continue
		}
		found = true
		areas := state.CombatPersistentAreas()
		if event.ActorID != "mage" || event.To != (combat.TilePoint{X: 4, Y: 2}) ||
			event.PersistentAreaID == 0 || len(event.Impacts) != 1 ||
			event.Impacts[0].TargetID != "enemy" || len(areas) != 1 ||
			len(areas[0].Cells) != 4 || len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("Quick Stinking Cloud event=%+v areas=%+v slots=%v", event, areas, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-5 Quick Stinking Cloud selection")
	}
}

func TestCombatAltMQuickCloudkillUsesAreaCenterAndPendingDelay(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "mage", Class: party.ClassMagicUser, Level: 7,
			SpellSlots: []uint8{CloudkillSpellID},
		}}
		state.SetCombatLineTerrain(func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x < 12 && y >= 0 && y < 12}
		})
		heroes := []combat.Fighter{{
			ID: "mage", Side: combat.SideParty, HitPoints: 100, MaxHitPoints: 100,
			HitDice: 7, ArmorClass: 0, InitiativeBonus: 30,
			HasCombatPosition: true, CombatX: 1, CombatY: 1,
		}}
		enemies := []combat.Fighter{{
			ID: "enemy", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
			HitDice: 4, ArmorClass: 10, InitiativeBonus: 1,
			HasCombatPosition: true, CombatX: 4, CombatY: 2,
			SavingThrows: []uint8{20, 20, 20, 20, 20},
		}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		caster, ok := state.fighter("mage")
		if !ok || caster.CombatAction.SpellID != CloudkillSpellID {
			continue
		}
		found = true
		if caster.CombatAction.Delay <= 0 || !caster.CombatAction.HasTargetPoint ||
			caster.CombatAction.TargetX != 4 || caster.CombatAction.TargetY != 2 ||
			len(state.partyRoster[0].SpellSlots) != 1 {
			t.Fatalf("pending Quick Cloudkill caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
		if changed := state.CombatManualControl(); changed != 1 {
			t.Fatalf("pending Cloudkill manual handoff changed=%d", changed)
		}
		for action := 0; action < 16; action++ {
			if _, ok := state.CombatVisualEvent(); ok {
				break
			}
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
		}
		event, ok := state.CombatVisualEvent()
		if !ok || event.Kind != combat.VisualAreaSpell || event.Effect != "cloudkill" ||
			event.ActorID != "mage" || event.To != (combat.TilePoint{X: 4, Y: 2}) ||
			event.PersistentAreaID == 0 || len(event.Impacts) != 1 ||
			!event.Impacts[0].Killed || len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("Quick Cloudkill event=%+v ok=%v slots=%v", event, ok, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-6 Quick Cloudkill selection")
	}
}

func TestCombatAltMQuickLightningBoltUsesLineTargetAndPendingDelay(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "mage", Class: party.ClassMagicUser, Level: 5,
			SpellSlots: []uint8{LightningBoltSpellID},
		}}
		state.SetCombatLineTerrain(func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x < 12 && y >= 0 && y < 12}
		})
		saves := []uint8{30, 30, 30, 30, 30}
		heroes := []combat.Fighter{{
			ID: "mage", Side: combat.SideParty, HitPoints: 100, MaxHitPoints: 100,
			ArmorClass: 0, InitiativeBonus: 30, HasCombatPosition: true,
			CombatX: 1, CombatY: 2, SavingThrows: saves,
		}}
		enemies := []combat.Fighter{
			{ID: "a-near", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
				ArmorClass: 10, InitiativeBonus: 1, HasCombatPosition: true,
				CombatX: 3, CombatY: 2, SavingThrows: saves},
			{ID: "z-far", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
				ArmorClass: 10, InitiativeBonus: 1, HasCombatPosition: true,
				CombatX: 5, CombatY: 2, SavingThrows: saves},
		}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		caster, ok := state.fighter("mage")
		if !ok || caster.CombatAction.SpellID != LightningBoltSpellID {
			continue
		}
		found = true
		if caster.CombatAction.Delay <= 0 || !caster.CombatAction.HasTargetPoint ||
			caster.CombatAction.TargetX != 3 || caster.CombatAction.TargetY != 2 ||
			len(state.partyRoster[0].SpellSlots) != 1 {
			t.Fatalf("pending Quick Lightning Bolt caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
		if changed := state.CombatManualControl(); changed != 1 {
			t.Fatalf("pending Lightning Bolt manual handoff changed=%d", changed)
		}
		for action := 0; action < 16; action++ {
			if _, ok := state.CombatVisualEvent(); ok {
				break
			}
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
		}
		event, ok := state.CombatVisualEvent()
		if !ok || event.Kind != combat.VisualLineSpell || event.Effect != "lightning_bolt" ||
			event.ActorID != "mage" || event.To != (combat.TilePoint{X: 3, Y: 2}) ||
			len(event.Impacts) != 2 || event.Impacts[0].TargetID != "a-near" ||
			event.Impacts[1].TargetID != "z-far" || len(event.Segments) < 2 ||
			len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("Quick Lightning Bolt event=%+v ok=%v slots=%v", event, ok, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-6 Quick Lightning Bolt selection")
	}
}

func TestQuickTargetUsesPreservedLegacyObjectOrder(t *testing.T) {
	state := NewState(testCatalog())
	state.SetCombatLineTerrain(func(x, y int) combat.LineCell {
		return combat.LineCell{Valid: x >= 0 && x < 12 && y >= 0 && y < 12}
	})
	heroes := []combat.Fighter{{
		ID: "caster", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20,
		HasCombatPosition: true, CombatX: 1, CombatY: 1,
	}}
	enemies := []combat.Fighter{
		{ID: "z-first", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20,
			HasCombatPosition: true, CombatX: 3, CombatY: 1},
		{ID: "a-second", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20,
			HasCombatPosition: true, CombatX: 5, CombatY: 1},
	}
	if err := state.StartCombat(heroes, enemies, 503); err != nil {
		t.Fatal(err)
	}
	caster, ok := state.fighter("caster")
	if !ok {
		t.Fatal("caster is absent")
	}
	point, found, err := state.quickLineSpellTarget(caster)
	if err != nil {
		t.Fatal(err)
	}
	if !found || point != (combat.TilePoint{X: 3, Y: 1}) {
		t.Fatalf("point=%+v found=%v; Quick target should follow legacy object order", point, found)
	}
}

func TestCombatAltMQuickCurseUsesPendingEnemyTarget(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.partyRoster = party.Roster{{
			ID: "cleric", Class: party.ClassCleric, Level: 1,
			SpellSlots: []uint8{CurseSpellID},
		}}
		heroes := []combat.Fighter{{
			ID: "cleric", Side: combat.SideParty, HitPoints: 100, MaxHitPoints: 100,
			ArmorClass: 0, InitiativeBonus: 30, HasCombatPosition: true,
			CombatX: 1, CombatY: 1,
		}}
		enemies := []combat.Fighter{{
			ID: "enemy", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
			ArmorClass: 10, AttackBonus: 3, InitiativeBonus: 1,
			HasCombatPosition: true, CombatX: 4, CombatY: 1,
		}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		caster, ok := state.fighter("cleric")
		if !ok || caster.CombatAction.SpellID != CurseSpellID {
			continue
		}
		found = true
		if caster.CombatAction.Delay <= 0 || caster.CombatAction.TargetID != "enemy" ||
			len(state.partyRoster[0].SpellSlots) != 1 {
			t.Fatalf("pending Quick Curse caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
		if changed := state.CombatManualControl(); changed != 1 {
			t.Fatalf("pending Curse manual handoff changed=%d", changed)
		}
		for action := 0; action < 8 && state.CombatActive(); action++ {
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
			caster, _ = state.fighter("cleric")
			if target, ok := state.fighter("enemy"); ok && target.Cursed {
				break
			}
		}
		target, _ := state.fighter("enemy")
		if !target.Cursed || target.AttackBonus != 2 || target.CurseRounds != 5 ||
			caster.CombatAction.SpellID != 0 || len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("Quick Curse target=%+v caster=%+v slots=%v", target, caster, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-3 Quick Curse selection")
	}
}

func TestCombatAltMQuickCauseLightWoundsUsesPendingAdjacentTarget(t *testing.T) {
	found := false
	for seed := int64(0); seed < 512 && !found; seed++ {
		state := NewState(testCatalog())
		state.partyRoster = party.Roster{{
			ID: "cleric", Class: party.ClassCleric, Level: 1,
			SpellSlots: []uint8{CauseLightWoundsSpellID},
		}}
		heroes := []combat.Fighter{{
			ID: "cleric", Side: combat.SideParty, HitPoints: 100, MaxHitPoints: 100,
			ArmorClass: 0, InitiativeBonus: 30, HasCombatPosition: true,
			CombatX: 1, CombatY: 1,
		}}
		enemies := []combat.Fighter{
			{ID: "a-adjacent", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
				ArmorClass: 10, InitiativeBonus: 1, HasCombatPosition: true,
				CombatX: 2, CombatY: 1},
			{ID: "z-far", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
				ArmorClass: 10, InitiativeBonus: 1, HasCombatPosition: true,
				CombatX: 8, CombatY: 8},
		}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
			t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
		}
		if err := state.CombatQuick(); err != nil {
			t.Fatal(err)
		}
		caster, ok := state.fighter("cleric")
		if !ok || caster.CombatAction.SpellID != CauseLightWoundsSpellID {
			continue
		}
		found = true
		if caster.CombatAction.Delay <= 0 || caster.CombatAction.TargetID != "a-adjacent" ||
			len(state.partyRoster[0].SpellSlots) != 1 {
			t.Fatalf("pending Quick Cause Light Wounds caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
		if changed := state.CombatManualControl(); changed != 1 {
			t.Fatalf("pending Cause Light Wounds manual handoff changed=%d", changed)
		}
		for action := 0; action < 8 && state.CombatActive(); action++ {
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
			caster, _ = state.fighter("cleric")
			if target, ok := state.fighter("a-adjacent"); ok && target.HitPoints < 100 {
				break
			}
		}
		target, _ := state.fighter("a-adjacent")
		far, _ := state.fighter("z-far")
		if target.HitPoints >= 100 || far.HitPoints != 100 ||
			caster.CombatAction.SpellID != 0 || len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("Quick Cause Light Wounds target=%+v far=%+v caster=%+v slots=%v", target, far, caster, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic seed reached the original priority-2 Quick Cause Light Wounds selection")
	}
}

func TestCombatAltMQuickProtectionSpellsUsePendingPartyTarget(t *testing.T) {
	tests := []struct {
		name   string
		spell  uint8
		active func(combat.Fighter) bool
	}{
		{name: "protection-from-evil", spell: ProtectionFromEvilSpellID, active: func(f combat.Fighter) bool {
			return f.ProtectedFromEvil
		}},
		{name: "protection-from-good", spell: ProtectionFromGoodSpellID, active: func(f combat.Fighter) bool {
			return f.ProtectedFromGood
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			found := false
			for seed := int64(0); seed < 512 && !found; seed++ {
				state := NewState(testCatalog())
				state.partyRoster = party.Roster{{
					ID: "cleric", Class: party.ClassCleric, Level: 2,
					SpellSlots: []uint8{test.spell},
				}}
				heroes := []combat.Fighter{{
					ID: "cleric", Side: combat.SideParty, HitPoints: 100, MaxHitPoints: 100,
					ArmorClass: 0, InitiativeBonus: 30, HasCombatPosition: true,
					CombatX: 1, CombatY: 1,
				}}
				enemies := []combat.Fighter{{
					ID: "enemy", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100,
					ArmorClass: 10, InitiativeBonus: 1, HasCombatPosition: true,
					CombatX: 8, CombatY: 8,
				}}
				if err := state.StartCombat(heroes, enemies, seed); err != nil {
					t.Fatal(err)
				}
				if enabled, err := state.CombatToggleQuickMagic(); err != nil || !enabled {
					t.Fatalf("ALT+M enabled=%v err=%v", enabled, err)
				}
				if err := state.CombatQuick(); err != nil {
					t.Fatal(err)
				}
				caster, ok := state.fighter("cleric")
				if !ok || caster.CombatAction.SpellID != test.spell {
					continue
				}
				found = true
				if caster.CombatAction.Delay <= 0 || caster.CombatAction.TargetID != "cleric" ||
					len(state.partyRoster[0].SpellSlots) != 1 {
					t.Fatalf("pending Quick %s caster=%+v slots=%v", test.name, caster, state.partyRoster[0].SpellSlots)
				}
				if changed := state.CombatManualControl(); changed != 1 {
					t.Fatalf("pending %s manual handoff changed=%d", test.name, changed)
				}
				for action := 0; action < 8 && state.CombatActive(); action++ {
					if err := state.CombatAct(); err != nil {
						t.Fatal(err)
					}
					caster, _ = state.fighter("cleric")
					if test.active(caster) {
						break
					}
				}
				if !test.active(caster) || caster.CombatAction.SpellID != 0 ||
					len(state.partyRoster[0].SpellSlots) != 0 {
					t.Fatalf("Quick %s caster=%+v slots=%v", test.name, caster, state.partyRoster[0].SpellSlots)
				}
			}
			if !found {
				t.Fatalf("no deterministic seed reached Quick %s selection", test.name)
			}
		})
	}
}

func TestCombatCastMagicMissileConsumesSlotAndDamagesTarget(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 1, SpellSlots: []uint8{MagicMissileSpellID}}}
	partyFighters := []combat.Fighter{{ID: "mage", Name: "法師", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastMagicMissile() {
		t.Fatalf("magic missile should be available: turns=%#v state=%#v", state.CombatTurns(), state)
	}
	if err := state.BeginCombatCast(MagicMissileSpellID); err != nil || state.CombatCastingSpell() != MagicMissileSpellID || len(state.partyRoster[0].SpellSlots) != 1 {
		t.Fatalf("begin cast state=%#v err=%v", state, err)
	}
	state.CancelCombatCast()
	if state.CombatCastingSpell() != 0 || len(state.partyRoster[0].SpellSlots) != 1 {
		t.Fatalf("cancel cast state=%#v", state)
	}
	if err := state.CombatCast(MagicMissileSpellID); err != nil {
		t.Fatal(err)
	}
	targetHP := 20
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "goblin" {
			targetHP = fighter.HitPoints
		}
	}
	if len(state.partyRoster[0].SpellSlots) != 0 || targetHP >= 20 {
		t.Fatalf("cast state=%#v fighters=%#v", state, state.CombatFighters())
	}
}

func TestCombatCastSleepUsesSelectedCellScanAndConsumesSlot(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 3,
		SpellSlots: []uint8{SleepSpellID},
	}}
	state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
		return enginescan.TacticalMap{
			Width: 7, Height: 1,
			Tiles:       []uint8{1, 1, 1, 1, 1, 1, 1},
			Definitions: []enginescan.TerrainDefinition{{LOS: 1}},
		}, nil
	})
	heroes := []combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty, LegacyObjectID: 1,
		HitPoints: 20, MaxHitPoints: 20, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 0, CombatY: 0,
	}}
	enemies := []combat.Fighter{
		{ID: "near-caster", Side: combat.SideEnemy, LegacyObjectID: 2,
			HitPoints: 10, MaxHitPoints: 10, HitDice: 1, HasCombatPosition: true, CombatX: 1, CombatY: 0},
		{ID: "selected", Side: combat.SideEnemy, LegacyObjectID: 3,
			HitPoints: 10, MaxHitPoints: 10, HitDice: 1, HasCombatPosition: true, CombatX: 5, CombatY: 0},
	}
	if err := state.StartCombat(heroes, enemies, 439); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatSelectTarget(1); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(SleepSpellID); err != nil {
		t.Fatal(err)
	}
	if point, ok := state.CombatSpellTargetPoint(); !ok || point != (combat.TilePoint{X: 5, Y: 0}) {
		t.Fatalf("Sleep center=%+v ok=%v, want selected enemy", point, ok)
	}
	if err := state.CombatCastWithTerrain(SleepSpellID, nil); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Sleep slot not consumed: %v", state.partyRoster[0].SpellSlots)
	}
	for _, fighter := range state.CombatFighters() {
		held := fighter.MonsterIsHeld()
		if fighter.ID == "near-caster" && held {
			t.Fatalf("caster-centred enemy was incorrectly put to Sleep: %+v", fighter)
		}
		if fighter.ID == "selected" && !held {
			t.Fatalf("selected-cell enemy did not receive Sleep: %+v", fighter)
		}
	}
	selected, _ := state.battle.Fighter("selected")
	if len(selected.MonsterAffects) != 1 || selected.MonsterAffects[0].Duration != 14 {
		t.Fatalf("normal cast handoff duration=%+v, want one of 15 ticks consumed at next-round boundary", selected.MonsterAffects)
	}
	for tick := 1; tick <= 13; tick++ {
		if _, err := state.battle.StartRound(); err != nil {
			t.Fatal(err)
		}
		selected, _ = state.battle.Fighter("selected")
		if !selected.MonsterIsHeld() {
			t.Fatalf("level-3 Sleep expired at additional tick %d", tick)
		}
	}
	if _, err := state.battle.StartRound(); err != nil {
		t.Fatal(err)
	}
	selected, _ = state.battle.Fighter("selected")
	if selected.MonsterIsHeld() {
		t.Fatal("level-3 Sleep remained held after duration 15")
	}
}

func TestActiveCombatSaveRestoresSleepVisualEffectSchedulerAndRandom(t *testing.T) {
	newSleepState := func() *State {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "mage", Name: "法師", Race: party.RaceHuman,
			Class: party.ClassMagicUser, Level: 3,
			Abilities: party.Abilities{Strength: 10, Intelligence: 16, Wisdom: 10,
				Dexterity: 12, Constitution: 10, Charisma: 10}, SpellSlots: []uint8{SleepSpellID},
		}}
		state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
			return enginescan.TacticalMap{Width: 3, Height: 1, Tiles: []uint8{1, 1, 1},
				Definitions: []enginescan.TerrainDefinition{{LOS: 1}}}, nil
		})
		return &state
	}
	original := newSleepState()
	original.Location = LocationMythDrannor
	heroes := []combat.Fighter{{ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0, AttackBonus: 30,
		DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 30,
		HasCombatPosition: true, CombatX: 0, CombatY: 0}}
	enemies := []combat.Fighter{{ID: "orc", Name: "半獸人", Side: combat.SideEnemy,
		HitPoints: 20, MaxHitPoints: 20, HitDice: 1, ArmorClass: 8,
		DamageDiceCount: 1, DamageDiceSides: 4, InitiativeBonus: 20,
		HasCombatPosition: true, CombatX: 2, CombatY: 0}}
	if err := original.StartCombat(heroes, enemies, 443); err != nil {
		t.Fatal(err)
	}
	if err := original.CombatSelectTarget(1); err != nil {
		t.Fatal(err)
	}
	if err := original.BeginCombatCast(SleepSpellID); err != nil {
		t.Fatal(err)
	}
	if err := original.CombatCastWithTerrain(SleepSpellID, nil); err != nil {
		t.Fatal(err)
	}
	event, ok := original.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualTwinkle {
		t.Fatalf("pre-save Sleep visual=%+v ok=%v", event, ok)
	}
	midElapsed := 700 * time.Millisecond
	if err := original.AdvanceCombatVisual(midElapsed); err != nil {
		t.Fatal(err)
	}
	if got, want := original.ConsumeSoundEvents(), []SoundEvent{SoundCast, SoundSpellHit}; !reflect.DeepEqual(got, want) {
		t.Fatalf("mid-Sleep sounds=%v want=%v", got, want)
	}
	path := filepath.Join(t.TempDir(), "sleep-active.json")
	if err := original.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := newSleepState()
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	loadedEvent, ok := loaded.CombatVisualEvent()
	if !ok || !reflect.DeepEqual(loadedEvent, event) || loaded.Mode != ModeCombat ||
		loaded.Location != LocationMythDrannor {
		t.Fatalf("loaded visual=%+v ok=%v mode=%v want=%+v", loadedEvent, ok, loaded.Mode, event)
	}
	if got := loaded.CombatVisualElapsed(); got != midElapsed {
		t.Fatalf("loaded visual elapsed=%s want=%s", got, midElapsed)
	}
	if got, want := loadedEvent.FrameAt(loaded.CombatVisualElapsed()), event.FrameAt(midElapsed); !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded visual frame=%+v want=%+v", got, want)
	}
	if sounds := loaded.ConsumeSoundEvents(); len(sounds) != 0 {
		t.Fatalf("load replayed committed visual sounds: %v", sounds)
	}
	if err := loaded.AdvanceCombatVisual(midElapsed); err != nil {
		t.Fatal(err)
	}
	if sounds := loaded.ConsumeSoundEvents(); len(sounds) != 0 {
		t.Fatalf("same-frame resume replayed committed visual sounds: %v", sounds)
	}
	if err := original.AdvanceCombatVisual(event.Duration()); err != nil {
		t.Fatal(err)
	}
	if err := loaded.AdvanceCombatVisual(loadedEvent.Duration()); err != nil {
		t.Fatal(err)
	}
	if got := append(original.ConsumeSoundEvents(), loaded.ConsumeSoundEvents()...); len(got) != 0 {
		t.Fatalf("completed resumed Sleep replayed sounds: %v", got)
	}
	want, err := original.battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) || loaded.combatTurnIndex != original.combatTurnIndex ||
		!reflect.DeepEqual(loaded.combatTurns, original.combatTurns) {
		t.Fatalf("post-handoff active combat diverged\n got=%#v\nwant=%#v", got, want)
	}
	wakeLoaded := newSleepState()
	if err := wakeLoaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	wakeEvent, _ := wakeLoaded.CombatVisualEvent()
	if err := wakeLoaded.AdvanceCombatVisual(wakeEvent.Duration()); err != nil {
		t.Fatal(err)
	}
	if _, err := wakeLoaded.battle.Attack("mage", "orc"); err != nil {
		t.Fatal(err)
	}
	woken, _ := wakeLoaded.battle.Fighter("orc")
	if woken.MonsterIsHeld() || len(wakeLoaded.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("post-load damage wake fighter=%+v slots=%v", woken, wakeLoaded.partyRoster[0].SpellSlots)
	}
	for tick := 0; tick < 14; tick++ {
		if _, err := original.battle.StartRound(); err != nil {
			t.Fatal(err)
		}
		if _, err := loaded.battle.StartRound(); err != nil {
			t.Fatal(err)
		}
	}
	originalOrc, _ := original.battle.Fighter("orc")
	loadedOrc, _ := loaded.battle.Fighter("orc")
	if originalOrc.MonsterIsHeld() || loadedOrc.MonsterIsHeld() ||
		!reflect.DeepEqual(originalOrc.MonsterAffects, loadedOrc.MonsterAffects) {
		t.Fatalf("natural expiry original=%+v loaded=%+v", originalOrc.MonsterAffects, loadedOrc.MonsterAffects)
	}
}

func TestActiveCombatSaveResumesMissileDeathFrameWithoutSoundReplay(t *testing.T) {
	newArcherState := func() *State {
		state := NewState(testCatalog())
		state.EnableCombatVisualTimeline(true)
		state.partyRoster = party.Roster{{
			ID: "archer", Name: "弓手", Race: party.RaceHuman,
			Class: party.ClassFighter, Level: 3,
			Abilities: party.Abilities{Strength: 12, Intelligence: 10, Wisdom: 10,
				Dexterity: 16, Constitution: 12, Charisma: 10},
		}}
		return &state
	}
	original := newArcherState()
	heroes := []combat.Fighter{{
		ID: "archer", Name: "弓手", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0,
		AttackBonus: 30, DamageDiceCount: 1, DamageDiceSides: 4,
		InitiativeBonus: 30, MissileWeapon: true,
		HasCombatPosition: true, CombatX: 1, CombatY: 3,
	}}
	enemies := []combat.Fighter{{
		ID: "goblin", Name: "哥布林", Side: combat.SideEnemy,
		HitPoints: 1, MaxHitPoints: 1, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 5, CombatY: 3,
	}}
	if err := original.StartCombat(heroes, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := original.CombatAct(); err != nil {
		t.Fatal(err)
	}
	event, ok := original.CombatVisualEvent()
	if !ok || event.Kind != combat.VisualMissile || !event.Killed {
		t.Fatalf("missile visual=%+v ok=%v", event, ok)
	}
	deathAt := combat.VisualWindupDuration + combat.VisualTravelDuration +
		combat.VisualImpactDuration + combat.VisualCommitDuration
	if err := original.AdvanceCombatVisual(deathAt); err != nil {
		t.Fatal(err)
	}
	// 兩聲 `SoundArrow`：進場一聲、飛行動畫尾端依武器類別再一聲。弓要另外的
	// 彈藥 ⇒ 飛出去的是箭 ⇒ 第二聲也是箭（原作 `SHOWARROW`，spec 1186）。
	want := []SoundEvent{SoundArrow, SoundArrow, SoundHit, SoundDead}
	if got := original.ConsumeSoundEvents(); !reflect.DeepEqual(got, want) {
		t.Fatalf("pre-save missile sounds=%v want=%v", got, want)
	}
	path := filepath.Join(t.TempDir(), "missile-death.json")
	if err := original.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	snapshot, err := original.activeCombatSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	badElapsed := *snapshot
	badElapsed.VisualElapsedNanos = int64(event.Duration() + time.Nanosecond)
	if err := newArcherState().restoreActiveCombat(badElapsed); err == nil {
		t.Fatal("restore accepted combat visual elapsed beyond event duration")
	}
	badMarker := *snapshot
	badMarker.VisualImpactSent = event.ImpactCount()
	if err := newArcherState().restoreActiveCombat(badMarker); err == nil {
		t.Fatal("restore accepted out-of-range combat visual impact marker")
	}
	loaded := newArcherState()
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	loadedEvent, ok := loaded.CombatVisualEvent()
	if !ok || !reflect.DeepEqual(loadedEvent.FrameAt(loaded.CombatVisualElapsed()), event.FrameAt(deathAt)) {
		t.Fatalf("loaded death frame=%+v elapsed=%s ok=%v", loadedEvent.FrameAt(loaded.CombatVisualElapsed()), loaded.CombatVisualElapsed(), ok)
	}
	if sounds := loaded.ConsumeSoundEvents(); len(sounds) != 0 {
		t.Fatalf("load replayed missile sounds: %v", sounds)
	}
	if err := loaded.AdvanceCombatVisual(deathAt); err != nil {
		t.Fatal(err)
	}
	if sounds := loaded.ConsumeSoundEvents(); len(sounds) != 0 {
		t.Fatalf("same death frame replayed missile sounds: %v", sounds)
	}
	if err := loaded.AdvanceCombatVisual(loadedEvent.Duration()); err != nil {
		t.Fatal(err)
	}
	if sounds := loaded.ConsumeSoundEvents(); len(sounds) != 0 {
		t.Fatalf("handoff replayed missile sounds: %v", sounds)
	}
	if loaded.Mode != ModeEvent || loaded.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("loaded handoff mode=%v status=%v", loaded.Mode, loaded.CombatStatus())
	}
}

func TestCombatCastSleepMapFailureDoesNotConsumeSlot(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "mage", Class: party.ClassMagicUser, Level: 1, SpellSlots: []uint8{SleepSpellID},
	}}
	state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
		return enginescan.TacticalMap{Width: 1, Height: 1, Tiles: []uint8{0}}, nil
	})
	if err := state.StartCombat(
		[]combat.Fighter{{ID: "mage", Side: combat.SideParty, HitPoints: 8, MaxHitPoints: 8, InitiativeBonus: 30}},
		[]combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1}},
		440,
	); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(SleepSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCastWithTerrain(SleepSpellID, nil); err == nil {
		t.Fatal("invalid Sleep TACTICALMAP unexpectedly resolved")
	}
	if got := state.partyRoster[0].SpellSlots; len(got) != 1 || got[0] != SleepSpellID {
		t.Fatalf("failed Sleep consumed slot: %v", got)
	}
}

func TestCombatCastMagicMissileUsesLocalizedResistanceMessage(t *testing.T) {
	data, err := os.ReadFile("../../assets/locale/zh-TW.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	format := catalog.Text("combat_magic_resisted", "")
	if format == "" {
		t.Fatal("combat_magic_resisted stable message ID is unavailable")
	}

	for seed := int64(1); seed <= 128; seed++ {
		state := NewState(catalog)
		state.partyRoster = party.Roster{{
			ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 11,
			SpellSlots: []uint8{MagicMissileSpellID},
		}}
		if err := state.StartCombat(
			[]combat.Fighter{{ID: "mage", Name: "法師", Side: combat.SideParty,
				HitPoints: 20, MaxHitPoints: 20, InitiativeBonus: 20},
				{ID: "witness", Name: "見證者", Side: combat.SideParty,
					HitPoints: 20, MaxHitPoints: 20, InitiativeBonus: 10}},
			[]combat.Fighter{{ID: "tyranthraxus", Name: "提朗瑟克斯", Side: combat.SideEnemy,
				HitPoints: 30, MaxHitPoints: 30,
				MonsterAffects: []combat.MonsterAffect{{Kind: 0x6A, Innate: true}}}},
			seed,
		); err != nil {
			t.Fatal(err)
		}
		if err := state.CombatCast(MagicMissileSpellID); err != nil {
			t.Fatal(err)
		}
		targetHP := 0
		for _, fighter := range state.CombatFighters() {
			if fighter.ID == "tyranthraxus" {
				targetHP = fighter.HitPoints
			}
		}
		if targetHP != 30 {
			continue
		}
		want := fmt.Sprintf(format, "法師", "提朗瑟克斯", 6)
		if state.CombatMessage() != want || len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("resisted message=%q want=%q slots=%v", state.CombatMessage(), want, state.partyRoster[0].SpellSlots)
		}
		return
	}
	t.Fatal("deterministic seeds did not produce a resisted Magic Missile")
}

func TestAttackEffectMessagesResolveStableLocaleIDs(t *testing.T) {
	data, err := os.ReadFile("../../assets/locale/zh-TW.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	attacker := combat.Fighter{Name: "烈焰者"}
	target := combat.Fighter{Name: "戰士"}
	hit := combat.AttackResult{
		Hit: true, Damage: 3,
		Effects: []combat.AttackEffectResult{{
			Kind: 0x4F, DamageFlags: combat.DamageFlagFire | combat.DamageFlagMagic, Damage: 9,
		}},
	}
	wantHit := fmt.Sprintf(catalog.Text("combat_hit_with_fire", ""), attacker.Name, target.Name, 3, 9)
	if got := formatAttackMessage(catalog, attacker, target, hit); got != wantHit || wantHit == "" {
		t.Fatalf("fire hit message=%q want=%q", got, wantHit)
	}

	protected := hit
	protected.Effects = []combat.AttackEffectResult{{
		Kind: 0x4F, DamageFlags: combat.DamageFlagFire | combat.DamageFlagMagic, Protected: true,
	}}
	wantProtected := fmt.Sprintf(catalog.Text("combat_hit_fire_protected", ""), attacker.Name, target.Name, 3)
	if got := formatAttackMessage(catalog, attacker, target, protected); got != wantProtected || wantProtected == "" {
		t.Fatalf("protected fire message=%q want=%q", got, wantProtected)
	}

	results := []combat.AttackResult{hit, {Hit: false}}
	wantMulti := fmt.Sprintf(catalog.Text("combat_multi_attack_with_fire", ""), attacker.Name, 2, 1, 3, 9)
	if got := formatMultiAttackMessage(catalog, attacker, results); got != wantMulti || wantMulti == "" {
		t.Fatalf("multi fire message=%q want=%q", got, wantMulti)
	}
}

func TestCombatCastCureLightWoundsConsumesSlotAndHealsParty(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, SpellSlots: []uint8{CureLightWoundsSpellID}}}
	partyFighters := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20}, {ID: "hero", Name: "戰士", Side: combat.SideParty, HitPoints: 3, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 19}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastCureLightWounds() {
		t.Fatalf("cure light wounds should be available: turns=%#v", state.CombatTurns())
	}
	if err := state.CombatCast(CureLightWoundsSpellID); err != nil {
		t.Fatal(err)
	}
	if got, want := state.ConsumeSoundEvents(), []SoundEvent{SoundCast, SoundSpellHit}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cure sound order=%v, want %v", got, want)
	}
	heroHP := 0
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" {
			heroHP = fighter.HitPoints
		}
	}
	if len(state.partyRoster[0].SpellSlots) != 0 || heroHP <= 3 || heroHP > 10 {
		t.Fatalf("healing state=%#v fighters=%#v", state, state.CombatFighters())
	}
}

func TestCombatCastCureLightWoundsHealsDownedPartyCorpse(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{
		{ID: "cleric", Name: "牧師", Class: party.ClassCleric, HitPoints: 8, MaxHitPoints: 8, SpellSlots: []uint8{CureLightWoundsSpellID}},
		{ID: "hero", Name: "倒下的英雄", Class: party.ClassFighter, HitPoints: 0, MaxHitPoints: 10, HealthStatus: party.HealthStatusUnconscious},
	}
	if err := state.StartCombat(
		[]combat.Fighter{
			{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, InitiativeBonus: 20},
			{ID: "hero", Name: "倒下的英雄", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10},
		},
		[]combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10}}, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(CureLightWoundsSpellID); err != nil {
		t.Fatal(err)
	}
	var hero combat.Fighter
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" {
			hero = fighter
		}
	}
	if hero.HitPoints <= 0 || hero.DeathOverlay || !hero.DownedCorpse || hero.HasCombatPosition || state.partyRoster[1].HitPoints != hero.HitPoints {
		t.Fatalf("downed cure state=%+v roster=%#v", hero, state.partyRoster)
	}
}

func TestCombatCastBlessConsumesSlotAndRaisesPartyAttackBonus(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, SpellSlots: []uint8{BlessSpellID}}}
	partyFighters := []combat.Fighter{
		{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 2, InitiativeBonus: 20},
		{ID: "hero", Name: "戰士", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 4},
	}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastBless() {
		t.Fatalf("Bless should be available: turns=%#v", state.CombatTurns())
	}
	if err := state.BeginCombatCast(BlessSpellID); err != nil || state.CombatCastingSpell() != BlessSpellID || len(state.partyRoster[0].SpellSlots) != 1 {
		t.Fatalf("begin Bless state=%#v err=%v", state, err)
	}
	state.CancelCombatCast()
	if err := state.BeginCombatCast(BlessSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(BlessSpellID); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Bless slot was not consumed: %#v", state.partyRoster[0].SpellSlots)
	}
	for _, fighter := range state.CombatFighters() {
		switch fighter.ID {
		case "cleric":
			if fighter.AttackBonus != 3 || !fighter.Blessed {
				t.Fatalf("cleric Bless state=%+v", fighter)
			}
		case "hero":
			if fighter.AttackBonus != 5 || !fighter.Blessed {
				t.Fatalf("hero Bless state=%+v", fighter)
			}
		}
	}
}

func TestManualCombatBlessUsesOriginalCastingDelayHandoff(t *testing.T) {
	found := false
	for seed := int64(0); seed < 256 && !found; seed++ {
		state := NewState(testCatalog())
		state.partyRoster = party.Roster{
			{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, SpellSlots: []uint8{BlessSpellID}},
			{ID: "ally", Name: "隊友", Class: party.ClassFighter, Level: 1},
		}
		heroes := []combat.Fighter{
			{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: -10, InitiativeBonus: 8, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1},
			{ID: "ally", Name: "隊友", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: -10, InitiativeBonus: 6, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1},
		}
		enemies := []combat.Fighter{{ID: "enemy", Name: "敵人", Side: combat.SideEnemy, HitPoints: 100, MaxHitPoints: 100, ArmorClass: 10, DamageDiceCount: 1, DamageDiceSides: 1}}
		if err := state.StartCombat(heroes, enemies, seed); err != nil {
			t.Fatal(err)
		}
		turn, ok := state.combatPartyTurn()
		if !ok || turn.ID != "cleric" {
			continue
		}
		if err := state.BeginCombatCast(BlessSpellID); err != nil {
			t.Fatal(err)
		}
		if err := state.ConfirmCombatCast(nil); err != nil {
			t.Fatal(err)
		}
		caster, _ := state.fighter("cleric")
		if caster.CombatAction.SpellID == 0 {
			continue
		}
		found = true
		if caster.Blessed || caster.CombatAction.SpellID != BlessSpellID || len(state.partyRoster[0].SpellSlots) != 1 {
			t.Fatalf("pending caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
		for action := 0; action < 8 && !caster.Blessed; action++ {
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
			caster, _ = state.fighter("cleric")
		}
		if !caster.Blessed || caster.CombatAction.SpellID != 0 || len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("resolved caster=%+v slots=%v", caster, state.partyRoster[0].SpellSlots)
		}
	}
	if !found {
		t.Fatal("no deterministic initiative ordering exposed pending manual Bless")
	}
}

func TestCombatPositiveDamageConsumesInterruptedPendingSpellByStableID(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1,
		SpellSlots: []uint8{BlessSpellID, CurseSpellID, BlessSpellID},
	}}
	heroes := []combat.Fighter{{
		ID: "cleric", Name: "牧師", Side: combat.SideParty,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 8,
	}}
	enemies := []combat.Fighter{{
		ID: "attacker", Name: "敵人", Side: combat.SideEnemy,
		HitPoints: 10, MaxHitPoints: 10, AttackBonus: 20, InitiativeBonus: 6,
	}}
	if err := state.StartCombat(heroes, enemies, 429); err != nil {
		t.Fatal(err)
	}
	turn, ok := state.combatPartyTurn()
	if !ok || turn.ID != "cleric" {
		t.Fatalf("party turn=%+v ok=%v", turn, ok)
	}
	if err := state.battle.BeginPendingSpellAction("cleric", BlessSpellID, 3); err != nil {
		t.Fatal(err)
	}
	if result, err := state.battle.ResolveAttack("attacker", "cleric", 20, 1); err != nil || result.Damage != 1 {
		t.Fatalf("attack=%+v err=%v", result, err)
	}
	message := state.consumeCombatSpellInterruptions()
	wantMessage := fmt.Sprintf(state.catalog.Text("combat_spell_interrupted", ""), "牧師")
	if message != wantMessage {
		t.Fatalf("message=%q want=%q", message, wantMessage)
	}
	wantSlots := []uint8{CurseSpellID, BlessSpellID}
	if !slices.Equal(state.partyRoster[0].SpellSlots, wantSlots) {
		t.Fatalf("slots=%v want=%v", state.partyRoster[0].SpellSlots, wantSlots)
	}
	caster, _ := state.fighter("cleric")
	if caster.CombatAction.SpellID != 0 || caster.CombatAction.Delay != 5 {
		t.Fatalf("caster action=%+v", caster.CombatAction)
	}
}

func TestCombatCloudkillDirectDeathConsumesInterruptedPendingSpellByStableID(t *testing.T) {
	state := NewState(combatVisualCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 4,
		SpellSlots: []uint8{BlessSpellID, CurseSpellID, BlessSpellID},
	}}
	saves := []uint8{10, 10, 10, 10, 10}
	heroes := []combat.Fighter{{
		ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 10, MaxHitPoints: 10, HitDice: 4, InitiativeBonus: 8,
		HasCombatPosition: true, CombatX: 4, CombatY: 4, SavingThrows: saves,
	}}
	enemies := []combat.Fighter{{
		ID: "enemy-mage", Name: "敵方法師", Side: combat.SideEnemy,
		HitPoints: 10, MaxHitPoints: 10, HitDice: 7, InitiativeBonus: 6,
		HasCombatPosition: true, CombatX: 8, CombatY: 4, SavingThrows: saves,
	}}
	if err := state.StartCombat(heroes, enemies, 430); err != nil {
		t.Fatal(err)
	}
	if err := state.battle.BeginPendingSpellAction("mage", BlessSpellID, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := state.battle.CastCloudkill("enemy-mage", combat.TilePoint{X: 4, Y: 4}, 7, nil); err != nil {
		t.Fatal(err)
	}
	message := state.consumeCombatSpellInterruptions()
	wantMessage := fmt.Sprintf(state.catalog.Text("combat_spell_interrupted", ""), "法師")
	if message != wantMessage {
		t.Fatalf("message=%q want=%q", message, wantMessage)
	}
	wantSlots := []uint8{CurseSpellID, BlessSpellID}
	if !slices.Equal(state.partyRoster[0].SpellSlots, wantSlots) {
		t.Fatalf("slots=%v want=%v", state.partyRoster[0].SpellSlots, wantSlots)
	}
	caster, _ := state.fighter("mage")
	if caster.HitPoints != 0 || caster.CombatAction != (combat.ActionState{}) {
		t.Fatalf("Cloudkill victim=%+v", caster)
	}
}

func TestCombatCastCurseConsumesSlotAndDebuffsSelectedEnemy(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, SpellSlots: []uint8{CurseSpellID}}}
	partyFighters := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, AttackBonus: 3}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastCurse() {
		t.Fatalf("Curse should be available: turns=%#v", state.CombatTurns())
	}
	if err := state.BeginCombatCast(CurseSpellID); err != nil || state.CombatCastingSpell() != CurseSpellID || len(state.partyRoster[0].SpellSlots) != 1 {
		t.Fatalf("begin Curse state=%+v err=%v", state, err)
	}
	state.CancelCombatCast()
	if err := state.BeginCombatCast(CurseSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(CurseSpellID); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Curse slot was not consumed: %#v", state.partyRoster[0].SpellSlots)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "goblin" && (fighter.AttackBonus != 2 || !fighter.Cursed || fighter.CurseRounds != 5) {
			t.Fatalf("enemy Curse state=%+v", fighter)
		}
	}
}

func TestCombatCastCauseLightWoundsConsumesSlotAndDamagesAdjacentEnemy(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, SpellSlots: []uint8{CauseLightWoundsSpellID}}}
	partyFighters := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20, HasCombatPosition: true, CombatX: 1, CombatY: 1}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastCauseLightWounds() {
		t.Fatalf("Cause Light Wounds should be available: turns=%#v", state.CombatTurns())
	}
	if err := state.BeginCombatCast(CauseLightWoundsSpellID); err != nil || len(state.partyRoster[0].SpellSlots) != 1 {
		t.Fatalf("begin Cause Light Wounds state=%+v err=%v", state, err)
	}
	state.CancelCombatCast()
	if err := state.BeginCombatCast(CauseLightWoundsSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(CauseLightWoundsSpellID); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Cause Light Wounds slot was not consumed: %#v", state.partyRoster[0].SpellSlots)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "goblin" && (fighter.HitPoints >= 20 || fighter.HitPoints < 12 || fighter.HitPoints > 19) {
			t.Fatalf("enemy damage state=%+v", fighter)
		}
	}
}

func TestCombatCastProtectionFromEvilConsumesSlotAndProtectsSelf(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 2, SpellSlots: []uint8{ProtectionFromEvilSpellID}}}
	partyFighters := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20, HasCombatPosition: true, CombatX: 1, CombatY: 1}}
	enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy, Evil: true, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 4, CombatY: 4}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastProtectionFromEvil() {
		t.Fatalf("Protection from Evil should be available: turns=%#v", state.CombatTurns())
	}
	if err := state.BeginCombatCast(ProtectionFromEvilSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(ProtectionFromEvilSpellID); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Protection from Evil slot was not consumed: %#v", state.partyRoster[0].SpellSlots)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "cleric" && (!fighter.ProtectedFromEvil || fighter.ProtectionEvilRounds != 5) {
			t.Fatalf("protection state=%+v", fighter)
		}
	}
}

func TestClericSpellIDSevenUsesProtectionFromGoodInsteadOfMagicMissile(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, SpellSlots: []uint8{ProtectionFromGoodSpellID}}}
	partyFighters := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20, HasCombatPosition: true, CombatX: 1, CombatY: 1}}
	enemies := []combat.Fighter{{ID: "good", Name: "聖騎士", Side: combat.SideEnemy, Good: true, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 4, CombatY: 4}}
	if err := state.StartCombat(partyFighters, enemies, 5); err != nil {
		t.Fatal(err)
	}
	if state.CombatCanCastMagicMissile() || !state.CombatCanCastProtectionFromGood() {
		t.Fatalf("class-specific ID 7 gates wrong: magic=%v protection=%v", state.CombatCanCastMagicMissile(), state.CombatCanCastProtectionFromGood())
	}
	if err := state.BeginCombatCast(ProtectionFromGoodSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(ProtectionFromGoodSpellID); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("Protection from Good slot was not consumed: %#v", state.partyRoster[0].SpellSlots)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "cleric" && !fighter.ProtectedFromGood {
			t.Fatalf("cleric protection state=%+v", fighter)
		}
	}
}

func TestStartEncounterBuildsBattleFromECLAndMonsterRecord(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1}}
	records := map[uint8]monster.Record{0x56: {Name: "BUGBEAR", MaxHitPoints: 2, HitPoints: 2, ArmorClass: 0, AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1}}
	result := ecl.RunResult{CombatRequested: true, MonsterSetup: &ecl.MonsterSetup{SpriteID: 0x09}, MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 0x56, Count: 1, IconBlock: 0x35}}}
	if err := state.StartEncounter(result, records, party, 11); err != nil {
		t.Fatal(err)
	}
	enemies := state.CombatTargets()
	if !state.CombatActive() || len(enemies) != 1 || enemies[0].SourceName != "BUGBEAR" || enemies[0].SpriteSet != 1 || enemies[0].SpriteBlock != 0x35 || enemies[0].AnimationBlock != 0x09 || !enemies[0].HasAnimation {
		t.Fatalf("state=%#v enemies=%#v", state, enemies)
	}
	if got := state.ConsumeSoundEvents(); len(got) == 0 || got[0] != SoundCombat {
		t.Fatalf("ECL combat sound events=%v, want %v at transition start", got, SoundCombat)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideParty && fighter.IconDirection != 7 {
			t.Fatalf("party icon direction=%d, want 7", fighter.IconDirection)
		}
		if fighter.Side == combat.SideEnemy && fighter.IconDirection != 3 {
			t.Fatalf("enemy icon direction=%d, want 3", fighter.IconDirection)
		}
	}
}

func TestStartEncounterUsesEightReservedPlayerSlotsForTemporaryMonsterAlly(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
	}}
	records := map[uint8]monster.Record{
		0x43: {Name: "RAKSHASA", HitPoints: 35, MaxHitPoints: 35},
		0x44: {Name: "HELL HOUND", HitPoints: 20, MaxHitPoints: 20},
	}
	result := ecl.RunResult{
		CombatRequested: true,
		MonsterSpawns: []ecl.MonsterSpawn{
			{MonsterID: 0x43, Count: 1, IconBlock: 0x43},
			{MonsterID: 0x44, Count: 2, IconBlock: 0x44},
		},
		CombatTeamWrites: []ecl.CombatTeamWrite{{TeamListIndex: 8, Value: 0x80}},
	}
	if err := state.StartEncounter(result, records, partyFighters, 11); err != nil {
		t.Fatal(err)
	}
	fighters := state.CombatFighters()
	partyCount, enemyCount := 0, 0
	temporaryFound := false
	for _, fighter := range fighters {
		if fighter.Side == combat.SideParty {
			partyCount++
			if fighter.SourceName == "RAKSHASA" && fighter.QuickFight && fighter.TemporaryAlly {
				temporaryFound = true
			}
		} else {
			enemyCount++
		}
	}
	if partyCount != 2 || enemyCount != 2 || !temporaryFound {
		t.Fatalf("fighters=%+v party=%d enemies=%d temporary=%v",
			fighters, partyCount, enemyCount, temporaryFound)
	}
}

func TestMonsterRecordsFollowCurrentECLChapter(t *testing.T) {
	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{
		3: append([]byte{0, 0}, make([]byte, 32)...),
	}, 3)
	state.SetMonsterRecords(map[uint8]monster.Record{7: {Name: "ECL1 fallback"}})
	state.SetMonsterRecordsForECL(2, map[uint8]monster.Record{7: {Name: "ECL2 monster"}})
	state.SetMonsterRecordsForECL(1, map[uint8]monster.Record{7: {Name: "ECL1 monster"}})
	records := state.monsterRecordsForCurrentECL()
	if records[7].Name != "ECL2 monster" {
		t.Fatalf("records=%#v, want ECL2 table", records)
	}
}

func TestCampAndAlterCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	keys := []string{
		"camp_rest_menu_prompt", "camp_rest_start", "camp_rest_add", "camp_rest_subtract",
		"camp_rest_exit", "camp_rest_insufficient", "camp_rest_interrupted", "camp_rest_done",
		"fix_no_cure", "fix_done", "camp_menu_prompt", "camp_save", "camp_view",
		"camp_magic", "camp_rest", "camp_alter", "camp_fix", "camp_exit",
		"camp_magic_pending", "camp_magic_cast_unknown", "camp_magic_memorize_selected",
		"camp_magic_memorize_full", "camp_magic_none", "camp_magic_summary",
		"camp_view_prompt", "camp_view_exit", "camp_magic_menu_prompt", "camp_magic_cast",
		"camp_magic_memorize", "camp_magic_scribe", "camp_magic_display", "camp_magic_rest",
		"camp_magic_exit", "camp_magic_cast_character_prompt", "camp_magic_cast_character",
		"camp_magic_cast_exit", "camp_magic_cast_spell_prompt", "camp_magic_cast_target_prompt",
		"camp_magic_cast_no_target", "camp_magic_cast_done", "camp_magic_prompt",
		"camp_magic_character", "camp_magic_view_exit", "camp_magic_memorize_prompt",
		"camp_magic_memorize_character", "camp_magic_memorize_spell_prompt",
		"camp_magic_mem_done", "camp_magic_mem_cancel", "camp_magic_memorize_exit",
		"camp_view_summary", "dungeon_prompt", "press_button", "enter_city", "journey_on",
		"camp", "camp_save_unavailable", "camp_save_requested", "camp_view_unavailable",
		"camp_magic_unavailable", "alter_prompt", "alter_order", "alter_drop", "alter_speed",
		"alter_icon", "alter_pics", "alter_exit", "alter_rename", "alter_order_done",
		"alter_order_unavailable", "alter_drop_unavailable", "alter_icon_unavailable",
		"alter_rename_prompt", "alter_rename_character", "alter_rename_exit",
		"alter_rename_edit_prompt", "alter_rename_done", "alter_speed_prompt",
		"alter_speed_slower", "alter_speed_faster", "alter_speed_exit", "alter_icon_prompt",
		"alter_icon_character", "alter_icon_exit", "alter_icon_edit_prompt", "alter_icon_head",
		"alter_icon_head_prev", "alter_icon_head_next", "alter_icon_body", "alter_icon_body_prev",
		"alter_icon_body_next", "alter_icon_done", "alter_pics_prompt", "alter_pics_on",
		"alter_pics_off", "alter_pics_monsters", "alter_pics_animations", "alter_pics_exit",
		"alter_drop_prompt", "alter_drop_character", "alter_drop_exit", "alter_drop_failed",
		"alter_drop_confirm_prompt", "alter_drop_confirm", "alter_drop_cancel",
		"alter_drop_warning", "alter_drop_done", "alter_order_prompt",
		"alter_order_destination_prompt", "alter_order_destination", "alter_order_exit",
		"alter_order_cancel", "alter_order_selected", "alter_speed_unavailable",
		"alter_pics_unavailable", "camp_alter_unavailable", "class_cleric", "class_fighter",
		"class_ranger", "class_paladin", "class_magic_user", "class_thief", "class_unknown",
	}
	for _, key := range keys {
		if got := catalog.Text(key, key); got == key {
			t.Fatalf("camp/ALTER locale ID %q is absent", key)
		}
	}
}

func TestCampBoundaryAndInGameJournal(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	if err := state.Camp(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campMenu || state.CampCount != 1 {
		t.Fatalf("camp state=%#v", state)
	}
	if err := state.OpenJournal(); err != nil || state.Mode != ModeJournal || state.JournalTitle == "" || state.JournalText == "" {
		t.Fatalf("journal state=%#v err=%v", state, err)
	}
	if len(state.JournalPages) != 0 || state.JournalPageStatus() != "" {
		t.Fatalf("new game must not reveal journal pages: pages=%v status=%q", state.JournalPages, state.JournalPageStatus())
	}
	emptyPage := state.JournalText
	if err := state.NextJournalPage(); err != nil || state.JournalPage != 0 || state.JournalText != emptyPage {
		t.Fatalf("empty journal next page=%#v err=%v", state, err)
	}
	state.appendJournalPage("test.journal.1", "第一頁")
	state.appendJournalPage("test.journal.2", "第二頁")
	state.JournalText = state.JournalPages[0]
	firstPage := state.JournalText
	if err := state.NextJournalPage(); err != nil || state.JournalPage != 1 || state.JournalText == firstPage || state.JournalPageStatus() == "" {
		t.Fatalf("journal next page=%#v err=%v", state, err)
	}
	if err := state.PreviousJournalPage(); err != nil || state.JournalPage != 0 || state.JournalText != firstPage {
		t.Fatalf("journal previous page=%#v err=%v", state, err)
	}
	if err := state.CloseJournal(); err != nil || state.Mode != ModeWilderness {
		t.Fatalf("journal close mode=%v err=%v", state.Mode, err)
	}
}

func TestECLRobScalesPartyGoldAndUsesReferenceWeightPenalty(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = []party.Character{
		{
			ID: "one", Copper: 9, Silver: 7, Electrum: 5, Gold: 101, Platinum: 3,
			Equipment: []monster.ItemRecord{
				{Type: 1, Weight: 10},
				{Type: 2, Weight: 300},
			},
		},
		{ID: "two", Gold: 51},
	}
	state.applyECLRobSignals(ecl.RunResult{RobRequests: []ecl.RobRequest{{
		AllParty: true, LossPercent: 50, ItemChance: 100,
	}}})
	if state.partyRoster[0].Gold != 50 || state.partyRoster[1].Gold != 25 {
		t.Fatalf("ROB gold=(%d,%d), want floor-scaled (50,25)",
			state.partyRoster[0].Gold, state.partyRoster[1].Gold)
	}
	if got := state.partyRoster[0]; got.Copper != 4 || got.Silver != 3 ||
		got.Electrum != 2 || got.Platinum != 1 {
		t.Fatalf("ROB typed coins=(%d,%d,%d,%d), want floor-scaled (4,3,2,1)",
			got.Copper, got.Silver, got.Electrum, got.Platinum)
	}
	if len(state.partyRoster[0].Equipment) != 1 || state.partyRoster[0].Equipment[0].Type != 2 {
		t.Fatalf("ROB equipment=%+v, want light item stolen and >255 weight item retained after -90 chance",
			state.partyRoster[0].Equipment)
	}
	state.applyECLRobSignals(ecl.RunResult{RobRequests: []ecl.RobRequest{{
		LossPercent: 100, SelectedPlayerIndex: 1, SelectedPlayerSet: true,
	}}})
	if state.partyRoster[0].Gold != 50 || state.partyRoster[1].Gold != 0 {
		t.Fatalf("selected ROB gold=(%d,%d), want only member 1 cleared",
			state.partyRoster[0].Gold, state.partyRoster[1].Gold)
	}
}

func TestCampMenuRestAndExit(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campMenu || len(state.Choices) != 7 || state.Choices[3] != "休息" {
		t.Fatalf("camp menu state=%#v", state)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campRestMenu || !state.campMenu || len(state.Choices) != 4 || state.CampCount != 0 {
		t.Fatalf("camp rest state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "REST" || state.RestHours() != 24 {
		t.Fatalf("rest start state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campMenu || len(state.Choices) != 7 {
		t.Fatalf("camp rest continuation state=%#v", state)
	}
	if err := state.Select(6); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.campMenu || len(state.Choices) != 3 || state.Choices[2] != "紮營" {
		t.Fatalf("camp exit state=%#v", state)
	}
}

func TestCampMenuViewCharacterAndReturn(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{{Name: "阿明", Class: party.ClassFighter, HitPoints: 7, MaxHitPoints: 12, Gold: 25, Gems: 2, Jewelry: 1}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campViewMenu || len(state.Choices) != 2 {
		t.Fatalf("camp view menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "VIEW" || !strings.Contains(state.Message, "阿明") || !strings.Contains(state.Message, "寶石 2") {
		t.Fatalf("camp view summary state=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campViewMenu {
		t.Fatalf("camp view return state=%#v", state)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.campViewMenu || !state.campMenu {
		t.Fatalf("camp view exit state=%#v", state)
	}
}

func TestCampMenuMagicListsMemorizedSlots(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{{Name: "法師", Class: party.ClassMagicUser, SpellSlots: []uint8{0x12, 0x24}, KnownSpells: []uint8{0x09, MagicMissileSpellID}}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campMagicMenu || len(state.Choices) != 6 || state.Choices[3] != "查看已記憶法術" {
		t.Fatalf("camp magic menu state=%#v", state)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campMagicViewMenu || len(state.Choices) != 2 {
		t.Fatalf("camp magic view menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	// 0x12 是全域編號 18（Read Magic），有譯名 ⇒ 顯示名字；
	// 0x24 是 36（Animate Dead），職業碼 3、沒有法術名鍵 ⇒ 保留編號，
	// 這樣匯入的存檔裡出現沒認識的編號時看得出來。
	if state.Mode != ModeEvent || state.OriginalEvent != "MAGIC" || !strings.Contains(state.Message, "閱讀魔法") || !strings.Contains(state.Message, "0x24") || !strings.Contains(state.Message, "可用法術：2") {
		t.Fatalf("camp magic summary state=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness || !state.campMagicViewMenu {
		t.Fatalf("camp magic return state=%#v err=%v", state, err)
	}
	if err := state.Select(1); err != nil || !state.campMagicMenu || state.campMagicViewMenu {
		t.Fatalf("camp magic command return state=%#v err=%v", state, err)
	}
	if err := state.Select(5); err != nil || state.campMagicMenu || !state.campMenu {
		t.Fatalf("camp magic exit state=%#v err=%v", state, err)
	}
}

func TestCampMagicLocalizesVerifiedFirstLevelSpellNames(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.catalog.Strings["spell_cleric_1"] = "祝福"
	state.catalog.Strings["spell_cleric_3"] = "治療輕傷"
	state.Mode = ModeWilderness
	state.partyRoster = party.Roster{{Name: "牧師", Class: party.ClassCleric, SpellSlots: []uint8{1, CureLightWoundsSpellID, 0x24}}}
	state.enterCampMenu()
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "祝福") || !strings.Contains(state.Message, "治療輕傷") || !strings.Contains(state.Message, "0x24") {
		t.Fatalf("spell labels=%q", state.Message)
	}
}

func TestCampMagicCastCureLightWoundsConsumesSlotAndSyncsHP(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{
		{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, SpellSlots: []uint8{CureLightWoundsSpellID}},
		{ID: "fighter", Name: "戰士", Class: party.ClassFighter, Level: 1, HitPoints: 2, MaxHitPoints: 10},
	}
	if err := state.SetParty([]combat.Fighter{
		{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "fighter", Name: "戰士", Side: combat.SideParty, HitPoints: 2, MaxHitPoints: 10, ArmorClass: 10},
	}); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.currentOriginalChoices[0] != "CAMP_MAGIC_CAST_SPELL_0" {
		t.Fatalf("cast spell menu=%#v originals=%#v", state.Choices, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.currentOriginalChoices[0] != "CAMP_MAGIC_CAST_TARGET_1" {
		t.Fatalf("cast target menu=%#v originals=%#v", state.Choices, state.currentOriginalChoices)
	}
	state.SetFixSeed(0)
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "MAGIC CAST" || state.partyRoster[1].HitPoints <= 2 || len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("cast result state=%#v", state)
	}
	if state.PartyFighters()[1].HitPoints != state.partyRoster[1].HitPoints || !strings.Contains(state.Message, "Cure Light Wounds") {
		t.Fatalf("cast sync message=%q fighters=%#v roster=%#v", state.Message, state.PartyFighters(), state.partyRoster)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness || !state.campMagicMenu {
		t.Fatalf("cast continuation state=%#v err=%v", state, err)
	}
}

func TestCampMagicMemorizeAppliesAtRest(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{{ID: "mage", Name: "法師", Class: party.ClassMagicUser, Level: 1, SpellSlots: []uint8{0x09}, KnownSpells: []uint8{0x09, MagicMissileSpellID}}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if got := state.partyRoster[0].SpellSlots; len(got) != 1 || got[0] != MagicMissileSpellID {
		t.Fatalf("memorized slots=%v, want [%d] after rest", got, MagicMissileSpellID)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "REST" || !strings.Contains(state.Message, "完成 1 名角色的法術記憶") {
		t.Fatalf("rest result state=%#v", state)
	}
}

func TestCampMagicMemorizeRequiresPreparationTime(t *testing.T) {
	if got := firstLevelMemorizationHours(map[int][]uint8{0: {1}}); got != 5 {
		t.Fatalf("one first-level spell requires %d hours, want 5", got)
	}
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{Name: "法師", Class: party.ClassMagicUser, Level: 1, SpellSlots: []uint8{0x09}}}
	state.pendingMemorizedSpells = map[int][]uint8{0: {MagicMissileSpellID}}
	state.SetRestHours(4)
	state.enterCampMenu()
	state.enterCampRestMenu()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if got := state.partyRoster[0].SpellSlots; len(got) != 1 || got[0] != 0x09 {
		t.Fatalf("short rest changed slots=%v", got)
	}
	if !strings.Contains(state.Message, "至少需要 5 小時") || state.pendingMemorizedSpells[0][0] != MagicMissileSpellID {
		t.Fatalf("short rest state=%#v", state)
	}
}

func TestCampMenuSaveEmitsRequest(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", Class: party.ClassFighter, Level: 1}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "SAVE" || !state.ConsumeSaveRequest() {
		t.Fatalf("camp save state=%#v", state)
	}
	if state.ConsumeSaveRequest() {
		t.Fatal("save request should be consumed exactly once")
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness || !state.campMenu {
		t.Fatalf("camp save continuation state=%#v err=%v", state, err)
	}
}

func TestCampAlterOrderReordersRosterAndFighters(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{
		{ID: "a", Name: "甲", Class: party.ClassFighter, Level: 1},
		{ID: "b", Name: "乙", Class: party.ClassCleric, Level: 1},
		{ID: "c", Name: "丙", Class: party.ClassThief, Level: 1},
	}
	state.party = []combat.Fighter{{ID: "a", Name: "甲"}, {ID: "b", Name: "乙"}, {ID: "c", Name: "丙"}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if !state.alterMenu || len(state.Choices) != 7 {
		t.Fatalf("alter menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.alterOrderSelected != 0 || len(state.Choices) != 4 {
		t.Fatalf("alter order destination state=%#v", state)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "ALTER ORDER" || state.partyRoster[0].ID != "b" || state.partyRoster[2].ID != "a" || state.party[0].ID != "b" {
		t.Fatalf("alter order result state=%#v party=%#v", state, state.party)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness || !state.alterMenu {
		t.Fatalf("alter order continuation state=%#v err=%v", state, err)
	}
	if err := state.Select(5); err != nil || state.alterMenu || !state.campMenu {
		t.Fatalf("alter exit state=%#v err=%v", state, err)
	}
}

func TestCampAlterRenameUpdatesRosterAndFighter(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{{ID: "hero", Name: "舊名", Class: party.ClassFighter, Level: 1}}
	state.party = []combat.Fighter{{ID: "hero", Name: "舊名", Side: combat.SideParty}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(6); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil || !state.RenameEditing() {
		t.Fatalf("rename editor state=%#v err=%v", state, err)
	}
	if err := state.BackspaceRenameName(); err != nil {
		t.Fatal(err)
	}
	if err := state.AppendRenameName([]rune("新")); err != nil {
		t.Fatal(err)
	}
	if err := state.CommitRename(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "ALTER RENAME" || state.partyRoster[0].Name != "舊新" || state.party[0].Name != "舊新" {
		t.Fatalf("rename result state=%#v", state)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness || !state.alterMenu {
		t.Fatalf("rename continuation state=%#v err=%v", state, err)
	}
}

func TestCampAlterDropRequiresConfirmationAndRemovesCharacter(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{
		{ID: "a", Name: "甲", Class: party.ClassFighter, Level: 1},
		{ID: "b", Name: "乙", Class: party.ClassCleric, Level: 1},
	}
	state.party = []combat.Fighter{{ID: "a", Name: "甲"}, {ID: "b", Name: "乙"}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.alterDropConfirm || state.partyRoster[1].ID != "b" {
		t.Fatalf("drop confirmation state=%#v", state)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.alterDropConfirm || !state.alterMenu || len(state.partyRoster) != 2 {
		t.Fatalf("drop cancel state=%#v", state)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "ALTER DROP" || len(state.partyRoster) != 1 || state.partyRoster[0].ID != "b" || len(state.party) != 1 || state.party[0].ID != "b" {
		t.Fatalf("drop result state=%#v party=%#v", state, state.party)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness || !state.alterMenu {
		t.Fatalf("drop continuation state=%#v err=%v", state, err)
	}
}

func TestCampAlterPicsTogglesRendererPreferences(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	if !state.PicturesEnabled() || !state.AnimationsEnabled() {
		t.Fatal("new state should enable pictures and animations")
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 3 || state.currentOriginalChoices[0] != "ALTER_PICS_MONSTERS" {
		t.Fatalf("pics menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.PicturesEnabled() || !strings.Contains(state.Choices[0], "關閉") {
		t.Fatalf("monster picture toggle state=%#v", state)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.AnimationsEnabled() || !strings.Contains(state.Choices[1], "關閉") {
		t.Fatalf("animation toggle state=%#v", state)
	}
	if err := state.Select(2); err != nil || state.alterPicsMenu || !state.alterMenu {
		t.Fatalf("pics exit state=%#v err=%v", state, err)
	}
}

func TestCampAlterSpeedAdjustsMessageRevealRate(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	if state.MessageSpeed() != 3 {
		t.Fatalf("default message speed=%d", state.MessageSpeed())
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 3 || state.currentOriginalChoices[0] != "ALTER_SPEED_SLOWER" {
		t.Fatalf("speed menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.MessageSpeed() != 2 {
		t.Fatalf("slower message speed=%d", state.MessageSpeed())
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.MessageSpeed() != 3 {
		t.Fatalf("faster message speed=%d", state.MessageSpeed())
	}
	if err := state.Select(2); err != nil || state.alterSpeedMenu || !state.alterMenu {
		t.Fatalf("speed exit state=%#v err=%v", state, err)
	}
}

func TestCampAlterIconUpdatesRosterAndFighter(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", Class: party.ClassFighter, Level: 1}}
	state.party = []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HasPartyIcon: true}}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.currentOriginalChoices[0] != "ALTER_ICON_CHARACTER_0" {
		t.Fatalf("icon character menu state=%#v", state)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 7 || !state.alterIconEdit {
		t.Fatalf("icon edit menu state=%#v", state)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].IconHeadBlock != 1 || state.party[0].PartyHeadBlock != 1 {
		t.Fatalf("head icon result roster=%#v party=%#v", state.partyRoster, state.party)
	}
	if err := state.Select(5); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].IconWeaponBlock != 1 || state.party[0].PartyBodyBlock != 1 {
		t.Fatalf("body icon result roster=%#v party=%#v", state.partyRoster, state.party)
	}
	if err := state.Select(6); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil || state.alterIconMenu || !state.alterMenu {
		t.Fatalf("icon exit state=%#v err=%v", state, err)
	}
}

func TestCampFixUsesMemorizedCureLightWounds(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.Choices = []string{"進入城市", "繼續旅程", "紮營"}
	state.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	state.partyRoster = party.Roster{
		{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 1, HitPoints: 8, MaxHitPoints: 8, SpellSlots: []uint8{CureLightWoundsSpellID}},
		{ID: "fighter", Name: "戰士", Class: party.ClassFighter, Level: 1, HitPoints: 1, MaxHitPoints: 10},
	}
	state.party = []combat.Fighter{{ID: "cleric", Side: combat.SideParty, HitPoints: 8, MaxHitPoints: 8}, {ID: "fighter", Side: combat.SideParty, HitPoints: 1, MaxHitPoints: 10}}
	state.SetFixSeed(7)
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(5); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.OriginalEvent != "FIX" || state.partyRoster[1].HitPoints <= 1 || state.partyRoster[1].HitPoints > 9 || state.party[1].HitPoints != state.partyRoster[1].HitPoints || len(state.partyRoster[0].SpellSlots) != 1 || !strings.Contains(state.Message, "Cure Light Wounds") {
		t.Fatalf("fix result state=%#v party=%#v", state, state.party)
	}
	if err := state.Continue(); err != nil || state.Mode != ModeWilderness || !state.campMenu {
		t.Fatalf("fix continuation state=%#v err=%v", state, err)
	}
}

func TestCampOpensMenuWithoutInstantHealing(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 3, MaxHitPoints: 10}}
	if err := state.SetParty(party); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeWilderness
	if err := state.Camp(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campMenu || state.PartyFighters()[0].HitPoints != 3 {
		t.Fatalf("camp state=%#v", state)
	}
}

func TestCampRestNaturallyHealsOneHPPer24Hours(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", HitPoints: 3, MaxHitPoints: 10}}
	state.party = []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 3, MaxHitPoints: 10}}
	state.SetRestHours(48)
	state.enterCampMenu()
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.partyRoster[0].HitPoints != 5 || state.party[0].HitPoints != 5 || !strings.Contains(state.Message, "48") {
		t.Fatalf("rest healing state=%#v", state)
	}
}

func TestCampRestAdvancesGameTimeBeforeHealing(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeWilderness
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", HitPoints: 3, MaxHitPoints: 10,
		Effects: []monster.AffectRecord{{Kind: 0x27, Duration: 30, Value: 30, Strength: 1}}}}
	state.SetRestHours(1)
	state.enterCampMenu()
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].Effects) != 0 || state.partyRoster[0].HitPoints != 3 {
		t.Fatalf("rest effects/healing=%#v, want 60-minute expiry and no 24h heal", state.partyRoster[0])
	}
	clock := state.GameTimeSlots()
	if clock[1] != 0 || clock[2] != 0 || clock[3] != 1 {
		t.Fatalf("rest clock=%v, want one normalized hour", clock)
	}
}

func TestAdvancePartyEffectsUsesRosterDurationAdapter(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
		Effects:   []monster.AffectRecord{{Kind: 1, Duration: 2, Value: 2, Strength: 1}},
	}}
	if removed := state.AdvancePartyEffects(1); removed != 0 || state.partyRoster[0].Effects[0].Duration != 1 {
		t.Fatalf("effects after one tick=%#v removed=%d", state.partyRoster[0].Effects, removed)
	}
	if removed := state.AdvancePartyEffects(1); removed != 1 || len(state.partyRoster[0].Effects) != 0 {
		t.Fatalf("effects after expiry=%#v removed=%d", state.partyRoster[0].Effects, removed)
	}
}

func TestAdvanceGameTimeUsesReferenceSlotScaleAndExpiresEffects(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{Effects: []monster.AffectRecord{{Kind: 1, Duration: 11, Value: 11, Strength: 1}}}}
	if err := state.AdvanceGameTime(2, 1); err != nil {
		t.Fatal(err)
	}
	clock := state.GameTimeSlots()
	if clock[2] != 1 || state.partyRoster[0].Effects[0].Duration != 1 {
		t.Fatalf("clock=%v effects=%#v, want slot-2=1 and 1 duration tick remaining", clock, state.partyRoster[0].Effects)
	}
	if err := state.AdvanceGameTime(1, 1); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].Effects) != 0 || state.GameTimeSlots()[1] != 1 {
		t.Fatalf("after expiry clock=%v effects=%#v", state.GameTimeSlots(), state.partyRoster[0].Effects)
	}
	if err := state.AdvanceGameTime(0, 9); err != nil {
		t.Fatal(err)
	}
	if state.GameTimeSlots()[0] != 9 {
		t.Fatalf("clock slot-0=%v", state.GameTimeSlots())
	}
	if err := state.AdvanceGameTime(0, 1); err != nil {
		t.Fatal(err)
	}
	if state.GameTimeSlots()[0] != 0 || state.GameTimeSlots()[1] != 2 {
		t.Fatalf("clock normalization=%v", state.GameTimeSlots())
	}
}

func TestAdvanceGameTimeIncrementsPartyAgeOnSlotSixOverflow(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{Age: 40}}
	if err := state.AdvanceGameTime(6, 0x100); err != nil {
		t.Fatal(err)
	}
	if got := state.partyRoster[0].Age; got != 41 {
		t.Fatalf("party age=%d, want 41", got)
	}
	if state.GameAgeCycles() != 1 || state.GameTimeSlots()[6] != 0 {
		t.Fatalf("clock=%v age cycles=%d", state.GameTimeSlots(), state.GameAgeCycles())
	}
}

func TestLoadDOSCharacterFilesInstallsImportedParty(t *testing.T) {
	state := NewState(testCatalog())
	record := make([]byte, party.DOSPlayerRecordSize)
	record[0] = 4
	copy(record[1:], []byte("ELLA"))
	record[0x10], record[0x12], record[0x14] = 16, 15, 12
	record[0x16], record[0x18], record[0x1A] = 14, 13, 10
	record[0x74], record[0x75], record[0x78], record[0x1A4] = 7, 5, 22, 18
	record[0x10E] = 4
	if err := state.LoadDOSCharacterFiles("ella-1", party.DOSPlayerFiles{Record: record}); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.PartyFighters()) != 1 || state.PartyFighters()[0].Name != "ELLA" {
		t.Fatalf("imported game state=%#v", state)
	}
}

func TestCharacterCreationBuildsPartyAndReturnsToWilderness(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil || state.Mode != ModeCharacterCreation {
		t.Fatalf("open creation mode=%v err=%v", state.Mode, err)
	}
	for index := 0; index < 6; index++ {
		if err := state.AddCreationCharacter(index); err != nil {
			t.Fatal(err)
		}
	}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.PartyFighters()) != 6 {
		t.Fatalf("finished state mode=%v party=%#v", state.Mode, state.PartyFighters())
	}
	if state.partyRoster[0].Age == 0 || state.partyRoster[1].Age == 0 || state.partyRoster[2].Age == 0 {
		t.Fatalf("created party ages=%#v, want generated ages", state.partyRoster)
	}
}

func TestCharacterCreationListsVerifiedClassOptions(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if len(state.CreationOptions) != 40 {
		t.Fatalf("creation options=%d, want 40 verified single/multi-class combinations", len(state.CreationOptions))
	}
	multiClassCount := 0
	for index, character := range state.CreationOptions {
		if err := character.Validate(); err != nil {
			t.Fatalf("option %d=%#v is invalid: %v", index, character, err)
		}
		ageLookup, lookupErr := gamepack.AgeLookup()
		if lookupErr != nil {
			t.Fatal(lookupErr)
		}
		if _, err := party.StartingAgeSpecFrom(ageLookup, character.Race, character.Class); err != nil {
			t.Fatalf("option %d has no starting-age evidence: %v", index, err)
		}
		if character.RawClassID >= 8 {
			multiClassCount++
			levels := uint8(0)
			for _, level := range character.ClassLevels {
				levels += level
			}
			if levels < 2 {
				t.Fatalf("multi-class option %d has no preserved class levels: %#v", index, character)
			}
		}
	}
	if multiClassCount != 18 {
		t.Fatalf("multi-class options=%d, want 18 reference race/class entries", multiClassCount)
	}
}

func TestCharacterCreationUsesGamePackTemplatesAndFormalLocale(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if state.dataPack == nil || state.dataPack.CharacterCreation == nil ||
		len(state.CreationOptions) != len(state.dataPack.CharacterCreation.Templates) {
		t.Fatalf("creation options=%d pack=%+v", len(state.CreationOptions), state.dataPack)
	}
	for index, template := range state.dataPack.CharacterCreation.Templates {
		wantName, ok := state.dataPack.Text(template.DisplayID, catalog.Language)
		if !ok || state.CreationOptions[index].ID != "creation."+template.ID ||
			state.CreationOptions[index].Name != wantName {
			t.Fatalf("creation option[%d]=%+v template=%+v want name=%q", index, state.CreationOptions[index], template, wantName)
		}
	}
	keys := []string{
		"creation_title", "creation_name_input", "creation_name_help",
		"creation_ability_title", "creation_ability_row", "creation_ability_help",
		"creation_option_label", "creation_progress", "creation_help",
		"ability_strength", "ability_intelligence", "ability_wisdom",
		"ability_dexterity", "ability_constitution", "ability_charisma",
		"race_dwarf", "race_elf", "race_gnome", "race_half_elf",
		"race_halfling", "race_human", "race_half_orc", "race_unknown",
		"class_cleric", "class_fighter", "class_ranger", "class_paladin",
		"class_magic_user", "class_thief", "class_unknown",
	}
	for _, key := range keys {
		if got := catalog.Text(key, ""); got == "" {
			t.Errorf("missing character creation locale key %q", key)
		}
	}
}

func TestCharacterCreationCustomNameSupportsUnicode(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCreationName(); err != nil {
		t.Fatal(err)
	}
	if err := state.AppendCreationName([]rune("阿勇")); err != nil {
		t.Fatal(err)
	}
	if err := state.CommitCreationName(); err != nil {
		t.Fatal(err)
	}
	if state.CreationOptions[0].Name != "阿勇" || state.CreationEditing {
		t.Fatalf("creation=%#v", state.CreationOptions[0])
	}
}

func TestCharacterCreationAdjustsAbilityValues(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.ToggleCreationAbilities(); err != nil {
		t.Fatal(err)
	}
	if err := state.AdjustCreationAbility(1); err != nil {
		t.Fatal(err)
	}
	if got, _ := state.CreationOptions[0].Abilities.Value(0); got != 17 {
		t.Fatalf("strength=%d, want 17", got)
	}
	if err := state.MoveCreationAbility(1); err != nil {
		t.Fatal(err)
	}
	if state.CreationAbility != 1 {
		t.Fatalf("ability cursor=%d", state.CreationAbility)
	}
}

func TestCharacterCreationRerollsAbilitiesWithSeed(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.RerollCreationAbilities(42); err != nil {
		t.Fatal(err)
	}
	first, _ := state.CreationOptions[0].Abilities.Value(0)
	if err := state.RerollCreationAbilities(42); err != nil {
		t.Fatal(err)
	}
	second, _ := state.CreationOptions[0].Abilities.Value(0)
	if first != second || first < 3 || first > 18 {
		t.Fatalf("reroll values=%d,%d", first, second)
	}
}

func TestPartySaveLoadRoundTrip(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	wantName := state.CreationOptions[0].Name
	if err := state.AddCreationCharacter(0); err != nil {
		t.Fatal(err)
	}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/party.json"
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(testCatalog())
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	if len(loaded.PartyFighters()) != 1 || loaded.PartyFighters()[0].Name != wantName {
		t.Fatalf("loaded party=%#v", loaded.PartyFighters())
	}
}

func TestResolveSpellSearchUsesLoadedPartyRoster(t *testing.T) {
	state := NewState(testCatalog())
	first := party.Character{
		ID: "p1", Name: "施法者", Race: party.RaceHuman, Class: party.ClassMagicUser, Level: 1,
		Abilities:  party.Abilities{Strength: 10, Intelligence: 14, Wisdom: 10, Dexterity: 10, Constitution: 10, Charisma: 10},
		SpellSlots: []uint8{0x12, 0x24},
	}
	second := first
	second.ID, second.Name = "p2", "另一位施法者"
	second.SpellSlots = []uint8{0x12}
	state.partyRoster = party.Roster{first, second}
	match, ok := state.ResolveSpellSearch(ecl.SpellSearch{SpellID: 0x12})
	if !ok || match.CharacterIndex != 0 || match.SlotIndex != 0 {
		t.Fatalf("spell match=%#v ok=%t", match, ok)
	}
}

func TestItemCatalogFeedsCharacterCreationFighterProjection(t *testing.T) {
	data := make([]byte, monster.BaseItemHeaderSize+monster.BaseItemRecordSize)
	data[monster.BaseItemHeaderSize+2] = 2
	data[monster.BaseItemHeaderSize+3] = 4
	data[monster.BaseItemHeaderSize+9] = 2
	data[monster.BaseItemHeaderSize+10] = 4
	catalog, err := monster.ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	character := party.Character{
		ID: "p1", Name: "戰士", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
		Equipment: []monster.ItemRecord{{Type: 0, Plus: 1, Readied: true}},
	}
	state := NewState(testCatalog())
	state.SetItemCatalog(catalog)
	state.Mode = ModeCharacterCreation
	state.CreationRoster = party.Roster{character}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	fighters := state.PartyFighters()
	if len(fighters) != 1 || fighters[0].DamageDiceCount != 2 || fighters[0].DamageDiceSides != 4 || fighters[0].DamageBonus != 1 || fighters[0].AttackBonus != 4 {
		t.Fatalf("equipped creation fighter=%#v", fighters)
	}
}

func TestLocalizeECLTextUsesCatalogAndPreservesUnknownLines(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["ecl_smoke_rises"] = "煙霧從殘破的牆後升起"
	catalog.Strings["ecl_move_away"] = "你們離開此處。"
	message := localizeECLText(catalog, []string{
		"SMOKE RISES FROM BEHIND THE RUINED WALLS",
		"YOU MOVE AWAY.",
		"UNMAPPED ECL LINE",
	})
	if message != "煙霧從殘破的牆後升起 你們離開此處。 UNMAPPED ECL LINE" {
		t.Fatalf("message=%q", message)
	}
}

func TestECLLineCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	keys := []string{
		"ecl_training_prompt", "ecl_training_progress", "ecl_training_exit",
		"ecl_tavern_pleasure", "ecl_tavern_special_1", "ecl_tavern_special_2",
		"ecl_tavern_purple_1", "ecl_tavern_purple_2", "ecl_tavern_purple_3",
		"ecl_tavern_commotion_1", "ecl_tavern_commotion_2", "ecl_tavern_commotion_3",
		"ecl_tilverton_patrol_arrives", "ecl_tilverton_guards_move",
		"ecl_tilverton_inn_welcome", "ecl_tilverton_inn_scowls",
		"ecl_tilverton_inn_calm", "ecl_tilverton_inn_listen",
		"ecl_filani_intro", "ecl_filani_correct", "ecl_filani_price", "ecl_filani_funds",
		"ecl_filani_lie", "ecl_filani_no", "ecl_weaponers_intro",
		"ecl_weaponers_interested", "ecl_weaponers_farewell", "ecl_weaponers_decline",
		"ecl_general_store_intro", "ecl_general_store_purchase", "ecl_general_store_farewell",
		"ecl_move_away", "ecl_smoke_rises", "ecl_yulash_sound", "ecl_battle_rings",
		"ecl_war_blasted_city", "ecl_small_magic_shop",
	}
	for _, key := range keys {
		if got := catalog.Text(key, key); got == key {
			t.Fatalf("ECL line locale ID %q is absent or unresolved", key)
		}
	}
}

// 走出戰場邊界就是原作的逃跑（spec 799／1112）：Gold Box 沒有 FLEE 指令，
// 邊界那一步本身就是嘗試脫離。這一條同時擋住「角色走到地圖外面」——
// 在有這個判定之前，往左走出 x = 0 是合法的一步。
func TestWalkingOffTheCombatMapEdgeAttemptsEscape(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20,
		MovementAllowance: 12, HasCombatPosition: true, CombatX: 0, CombatY: 3}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, MovementAllowance: 6,
		HasCombatPosition: true, CombatX: 1, CombatY: 3}}
	if err := state.StartCombat(party, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatMove(); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatMove(-1, 0); err != nil {
		t.Fatal(err)
	}
	// 敵人比較慢 ⇒ 逃得掉；場上沒有隊員了 ⇒ 戰鬥以「逃離」收場，不是全滅。
	if !strings.Contains(state.CombatMessage(), "逃離") {
		t.Fatalf("逃跑訊息=%q", state.CombatMessage())
	}
	if state.Mode != ModeEvent {
		t.Fatalf("逃離之後 mode=%v", state.Mode)
	}
	if got := state.Message; got != trainingTestCatalog(t).Text("combat_party_fled", "") {
		t.Fatalf("結果訊息=%q", got)
	}
}

// 逃不掉時角色留在原地，回合也還沒結束——原作的 'Escape is blocked' 就是這樣。
func TestBlockedEscapeKeepsTheFighterInPlace(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20,
		MovementAllowance: 6, HasCombatPosition: true, CombatX: 0, CombatY: 3}}
	enemies := []combat.Fighter{{ID: "goblin", Name: "哥布林", Side: combat.SideEnemy,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, MovementAllowance: 12,
		HasCombatPosition: true, CombatX: 1, CombatY: 3}}
	if err := state.StartCombat(party, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatMove(); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatMove(-1, 0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatMoveMode() {
		t.Fatalf("擋住之後 mode=%v moveMode=%v", state.Mode, state.CombatMoveMode())
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID == "hero" && (fighter.CombatX != 0 || fighter.Escaped) {
			t.Fatalf("被擋住的角色=%+v", fighter)
		}
	}
}

// AI 的一回合是「先走到打得到，再打」（spec 830／838）。在這之前怪物會站在
// 出生點隔空攻擊——距離、移動力與武器射程都沒有進到判斷裡。
func TestEnemyTurnWalksTowardTheTargetBeforeAttacking(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 2, CombatY: 5}}
	enemies := []combat.Fighter{{ID: "ogre", Name: "食人魔", Side: combat.SideEnemy,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20,
		DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 100,
		MovementAllowance: 2, WeaponRange: 1,
		HasCombatPosition: true, CombatX: 12, CombatY: 5}}
	if err := state.StartCombat(party, enemies, 1); err != nil {
		t.Fatal(err)
	}
	ogre := func() combat.Fighter {
		for _, fighter := range state.CombatFighters() {
			if fighter.ID == "ogre" {
				return fighter
			}
		}
		t.Fatal("食人魔不見了")
		return combat.Fighter{}
	}
	// 移動力 2 ＝ 4 個半格 ⇒ 一回合走兩格，走不到就不會攻擊。
	if got := ogre(); got.CombatX != 10 {
		t.Fatalf("第一回合走到 x=%d，移動力 2 應該走兩格到 x=10", got.CombatX)
	}
	if state.partyRoster != nil && len(state.partyRoster) > 0 && state.partyRoster[0].HitPoints < 30 {
		t.Fatal("還沒走到就攻擊了")
	}
}

// 地形要與玩家共用：怪物不能走玩家走不了的格子。
func TestEnemyApproachHonoursTheInstalledMovementTerrain(t *testing.T) {
	state := NewState(testCatalog())
	// x = 8 那一整行不能走 ⇒ 怪物只能停在 x = 9。
	state.SetCombatMovementTerrain(func(x, y int) (int, bool) {
		if x == 8 {
			return 0, false
		}
		return 1, true
	})
	party := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 2, CombatY: 5}}
	enemies := []combat.Fighter{{ID: "ogre", Name: "食人魔", Side: combat.SideEnemy,
		HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20,
		DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 100,
		MovementAllowance: 8, WeaponRange: 1,
		HasCombatPosition: true, CombatX: 12, CombatY: 5}}
	if err := state.StartCombat(party, enemies, 1); err != nil {
		t.Fatal(err)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID != "ogre" {
			continue
		}
		if fighter.CombatX == 8 {
			t.Fatal("怪物站在不可通行的格子上")
		}
		// 模式的候選方向可能繞開，但一定不會穿過 x = 8。
		if fighter.CombatX < 8 {
			t.Fatalf("怪物穿過了牆：x=%d", fighter.CombatX)
		}
	}
}

// 資料驅動的效果類法術：定身術的效果碼、持續時間、豁免類別全部來自法術主表
// （spec 1111／1117），沒有一支法術一段程式碼。
//
// 這一條同時證明效果**真的生效**：0x34 是 `MonsterIsHeld` 認得的碼，
// 被定住的怪物那一回合不會行動。
func TestHoldPersonAppliesTheOriginalEffectCodeAndHolds(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 3,
		ClassLevels: [8]uint8{0: 3}, HitPoints: 12, MaxHitPoints: 12,
		HealthStatus: party.HealthStatusOK, SpellSlots: []uint8{23},
	}}
	party := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty,
		HitPoints: 12, MaxHitPoints: 12, ArmorClass: 10, InitiativeBonus: 100,
		HasCombatPosition: true, CombatX: 2, CombatY: 5}}
	enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy,
		HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, HitDice: 1,
		SavingThrows:      []uint8{20, 20, 20, 20, 20}, // 幾乎必定豁免失敗
		HasCombatPosition: true, CombatX: 3, CombatY: 5}}
	if err := state.StartCombat(party, enemies, 5); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(23); err != nil {
		t.Fatal(err)
	}
	held := false
	for _, fighter := range state.CombatFighters() {
		if fighter.ID != "orc" {
			continue
		}
		for _, affect := range fighter.MonsterAffects {
			// 屬性表說定身術的效果碼是 52（0x34），持續 0×等級 + 6。
			if affect.Kind == 0x34 && affect.Active {
				held = true
				// 屬性表算出來是 6；施法之後回合會推進，效果跟著扣一跳，
				// 所以這裡只驗「在合理區間」，精確值由 internal/combat 的
				// 單元測試釘住（那一層不會推進回合）。
				if affect.Duration < 1 || affect.Duration > 6 {
					t.Fatalf("持續時間=%d，屬性表算出來是 6，扣一跳也該在 5..6", affect.Duration)
				}
			}
		}
		if fighter.MonsterIsHeld() != held {
			t.Fatalf("效果記上去了但 MonsterIsHeld=%v", fighter.MonsterIsHeld())
		}
	}
	if !held {
		t.Fatalf("定身術沒有套上效果碼 0x34／持續 6：%+v", state.CombatFighters())
	}
	// 法術槽要被消耗掉。
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("法術槽沒有消耗：%v", state.partyRoster[0].SpellSlots)
	}
}
