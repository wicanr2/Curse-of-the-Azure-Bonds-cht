package game

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func requireGamePackText(t *testing.T, state *State, messageID string) string {
	t.Helper()
	value, ok := state.dataPack.Text(messageID, state.catalog.Language)
	if !ok || value == "" {
		t.Fatalf("game-pack message %q is unavailable for locale %q", messageID, state.catalog.Language)
	}
	return value
}

func requireGamePackOptionIndex(t *testing.T, state *State, optionID string) int {
	t.Helper()
	for _, rule := range state.dataPack.OptionRules {
		if rule.ID != optionID {
			continue
		}
		if index, ok := state.OriginalChoiceIndex(rule.Source); ok {
			return index
		}
		t.Fatalf("game-pack option %q source %q is not an active choice: %v", optionID, rule.Source, state.currentOriginalChoices)
	}
	t.Fatalf("game-pack option %q is unavailable", optionID)
	return 0
}

func findGamePackOptionIndex(state *State, optionID string) (int, bool) {
	if state == nil || state.dataPack == nil {
		return 0, false
	}
	for _, rule := range state.dataPack.OptionRules {
		if rule.ID != optionID {
			continue
		}
		return state.OriginalChoiceIndex(rule.Source)
	}
	return 0, false
}

func requireCombatantName(t *testing.T, state *State, source string) string {
	t.Helper()
	value, ok := state.dataPack.LocalizeCombatantName(source, state.catalog.Language)
	if !ok || value == "" {
		t.Fatalf("game-pack combatant name %q is unavailable for locale %q", source, state.catalog.Language)
	}
	return value
}

func TestRealECLJourneyDispatchesGeneralStoreService(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	eclData := zipData(t, image, "ECL1.DAX")
	eclBlocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	eclByID := make(map[uint8][]byte, len(eclBlocks))
	for _, block := range eclBlocks {
		eclByID[block.Entry.ID] = block.Data
	}

	monsterData := zipData(t, image, "MON1CHA.DAX")
	monsterBlocks, err := dax.Parse(monsterData)
	if err != nil {
		t.Fatal(err)
	}
	records := make(map[uint8]monster.Record, len(monsterBlocks))
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		records[block.Entry.ID] = record
	}

	state := NewStateFromECLBlocks(testCatalog(), eclByID, 0x51)
	debugSession, err := ecl.NewBlockSession(eclByID, 0x51)
	if err != nil {
		t.Fatal(err)
	}
	debugResult, debugErr := debugSession.RunInteractive(180, []uint16{1, 1})
	if debugErr != nil {
		t.Fatal(debugErr)
	}
	if !debugResult.ShopRequested || debugResult.ShopPriceScale != 0x10 ||
		debugResult.CombatRequested || len(debugResult.TreasureRequests) != 1 ||
		debugResult.TreasureRequests[0].ItemBlock != 5 {
		t.Fatalf("real result=%+v, want General Store service with ITEM1 block 5", debugResult)
	}
	state.SetMonsterRecords(records)
	treasureItems, err := ParseTreasureItemBlocks(map[uint8][]byte{
		1: zipData(t, image, "ITEM1.DAX"),
	})
	if err != nil {
		t.Fatal(err)
	}
	state.SetTreasureItemBlocks(treasureItems)
	state.Area.GameArea = 1
	if err := state.SetParty([]combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 5,
		AttackBonus: 10, DamageDiceCount: 1, DamageDiceSides: 6,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	// The observed ECL1 path is JOURNEY ON, then STORE; the latter reaches
	// CMD_COMBAT with EnterShop set and therefore enters CityShop.
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.OriginalEvent != "PICTURE" || !state.PictureRequested {
		t.Fatalf("journey picture boundary=%#v", state)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopMenu || len(state.ShopOffers()) == 0 {
		t.Fatalf("real ECL path did not dispatch CityShop: mode=%v shop=%v offers=%d",
			state.Mode, state.shopMenu, len(state.ShopOffers()))
	}
}

func runNormalNewGameToEssembra(t *testing.T) *State {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
		return nil
	}
	defer image.Close()

	all := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			all[block.Entry.ID] = block.Data
		}
	}
	state := NewStateFromECLBlocks(trainingTestCatalog(t), all, 0x50)
	guardBlocks, err := dax.Parse(zipData(t, image, "MON2CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	guardRecords := make(map[uint8]monster.Record, len(guardBlocks))
	for _, block := range guardBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		guardRecords[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(2, guardRecords)
	patrolBlocks, err := dax.Parse(zipData(t, image, "MON1CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	patrolRecords := make(map[uint8]monster.Record, len(patrolBlocks))
	for _, block := range patrolBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		patrolRecords[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(1, patrolRecords)
	hapBlocks, err := dax.Parse(zipData(t, image, "MON5CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	hapRecords := make(map[uint8]monster.Record, len(hapBlocks))
	for _, block := range hapBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		hapRecords[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(5, hapRecords)
	yulashBlocks, err := dax.Parse(zipData(t, image, "MON3CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	yulashRecords := make(map[uint8]monster.Record, len(yulashBlocks))
	for _, block := range yulashBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		yulashRecords[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(3, yulashRecords)
	yulashAffectBlocks, err := dax.Parse(zipData(t, image, "MON3SPC.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	yulashAffects := make(map[uint8][]monster.AffectRecord, len(yulashAffectBlocks))
	for _, block := range yulashAffectBlocks {
		affects, parseErr := monster.ParseAffects(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		yulashAffects[block.Entry.ID] = affects
	}
	state.SetMonsterAffectsForECL(3, yulashAffects)
	yulashItemBlocks, err := dax.Parse(zipData(t, image, "MON3ITM.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	yulashItems := make(map[uint8][]monster.ItemRecord, len(yulashItemBlocks))
	for _, block := range yulashItemBlocks {
		items, parseErr := monster.ParseItems(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		yulashItems[block.Entry.ID] = items
	}
	state.SetMonsterItemsForECL(3, yulashItems)
	if err := state.Apply(ActionStart); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCharacterCreation {
		t.Fatalf("party-less production start mode=%v, want character creation", state.Mode)
	}
	if err := state.AddCreationCharacter(0); err != nil {
		t.Fatal(err)
	}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x01 {
		t.Fatalf("new game block=0x%02X, want 0x01", state.session.CurrentBlockID())
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		state.Message != requireGamePackText(t, &state, "opening.new-game-awakening") {
		t.Fatalf("new game first pause: mode=%v choices=%v message=%q", state.Mode, state.Choices, state.Message)
	}
	if state.LoadPieces != [3]uint16{1, 2, 3} {
		t.Fatalf("new game LOAD PIECES=%v", state.LoadPieces)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested ||
		state.Message != requireGamePackText(t, &state, "opening.new-game-marks") {
		t.Fatalf("new game picture text: mode=%v picture=%v message=%q", state.Mode, state.PictureRequested, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 {
		t.Fatalf("new game second pause: mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 7 || state.DungeonY != 13 || state.DungeonDirection != 2 {
		t.Fatalf("new game world entry: mode=%v position=(%d,%d) direction=%d",
			state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
	if state.Location != LocationTilverton || !state.Area.InDungeon ||
		state.Area.Current3DMapBlockID != 1 || len(state.Choices) != 0 {
		t.Fatalf("new game world location=%v area=%+v choices=%v, want Tilverton GEO1 dungeon without synthetic menu",
			state.Location, state.Area, state.Choices)
	}
	for address, want := range map[uint16]uint16{
		0xC04B: 7, 0xC04C: 13, 0xC04D: 1,
	} {
		if got, ok := state.session.MemoryValue(address); !ok || got != want {
			t.Fatalf("world register %#x=%d,%v, want %d,true", address, got, ok, want)
		}
	}
	geoCatalog := geo.NewCatalog()
	if err := geoCatalog.AddDAX(2, zipData(t, image, "GEO2.DAX")); err != nil {
		t.Fatal(err)
	}
	grid, ok := geoCatalog.Lookup(geo.MapRef{Set: 2, BlockID: 1})
	if !ok || !grid.CanMoveDungeonWrapped(7, 13, 6) {
		t.Fatal("Tilverton start has no verified west path to the Windlord's Inn")
	}
	treasureItems, err := ParseTreasureItemBlocks(map[uint8][]byte{
		2: zipData(t, image, "ITEM2.DAX"),
	})
	if err != nil {
		t.Fatal(err)
	}
	state.SetTreasureItemBlocks(treasureItems)
	if err := state.MoveDungeon(grid, -1, 0, 6); err != nil {
		t.Fatal(err)
	}
	if state.DungeonX != 6 || state.DungeonY != 13 || state.DungeonDirection != 6 ||
		state.DungeonWallRoof != 0x86 {
		t.Fatalf("normal west step state=(%d,%d,%d) roof=%#x, want (6,13,6) roof=0x86",
			state.DungeonX, state.DungeonY, state.DungeonDirection, state.DungeonWallRoof)
	}
	wantInnWelcome := state.catalog.Text("ecl_tilverton_inn_welcome", "ecl_tilverton_inn_welcome")
	wantInnScowls := state.catalog.Text("ecl_tilverton_inn_scowls", "ecl_tilverton_inn_scowls")
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 3 ||
		!state.SceneCharacterRequested || state.SceneHeadBlock != 3 || state.SceneBodyBlock != 3 ||
		!strings.Contains(state.Message, wantInnWelcome) ||
		!strings.Contains(state.Message, wantInnScowls) {
		registers := make([]uint16, 0, 5)
		for address := uint16(0xC04B); address <= 0xC04F; address++ {
			value, _ := state.session.MemoryValue(address)
			registers = append(registers, value)
		}
		flags, _ := state.session.MemoryValue(0x4C00)
		mask, _ := state.session.MemoryValue(0x7F7A)
		index, _ := state.session.MemoryValue(0x7F3F)
		quotient, _ := state.session.MemoryValue(0x7F80)
		selector, _ := state.session.MemoryValue(0x7F7B)
		t.Fatalf("Windlord's Inn first pause mode=%v picture=%v:%d message=%q registers=%#v flags=%#x mask=%#x index=%#x quotient=%#x selector=%#x",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message, registers, flags, mask, index, quotient, selector)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		state.Choices[0] != state.catalog.Text("press_button", "請按任意鍵或 Enter 繼續") {
		t.Fatalf("Windlord's Inn first Continue mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	wantInnJournalTrigger := requireGamePackText(t, &state, "journal-trigger.tilverton-inn-31")
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, wantInnJournalTrigger) {
		t.Fatalf("Windlord's Inn journal pause mode=%v choices=%v message=%q", state.Mode, state.Choices, state.Message)
	}
	// 開場甦醒帶出手札 1 的三頁（spec 1108 §二之二），所以旅店那一則是第四頁。
	// 這裡連順序一起釘：玩家一進遊戲手上就有導言，之後才依劇情逐則追加。
	wantPages := []string{
		requireGamePackText(t, &state, "journal.1.1"),
		requireGamePackText(t, &state, "journal.1.2"),
		requireGamePackText(t, &state, "journal.1.3"),
		requireGamePackText(t, &state, "journal.31"),
	}
	if len(state.JournalPages) != len(wantPages) {
		t.Fatalf("Journal pages=%v, want %d pages ending with Entry 31",
			state.JournalPages, len(wantPages))
	}
	for index, want := range wantPages {
		if state.JournalPages[index] != want {
			t.Fatalf("journal page %d = %q, want %q", index, state.JournalPages[index], want)
		}
	}
	if err := state.OpenJournal(); err != nil {
		t.Fatal(err)
	}
	for state.JournalPage+1 < len(state.JournalPages) {
		if err := state.NextJournalPage(); err != nil {
			t.Fatal(err)
		}
	}
	last := len(wantPages) - 1
	if state.JournalPage != last || state.JournalText != wantPages[last] {
		t.Fatalf("unlocked Journal Entry 31 is not reachable in journal UI: page=%d text=%q",
			state.JournalPage, state.JournalText)
	}
	if err := state.CloseJournal(); err != nil || state.Mode != ModeWilderness {
		t.Fatalf("close journal after inn event: mode=%v err=%v", state.Mode, err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 6 || state.DungeonY != 13 {
		t.Fatalf("Windlord's Inn continuation mode=%v position=(%d,%d), want same dungeon cell",
			state.Mode, state.DungeonX, state.DungeonY)
	}
	if err := state.MoveDungeon(grid, 1, 0, 2); err != nil {
		t.Fatal(err)
	}
	if state.DungeonX != 7 || state.DungeonY != 13 || state.DungeonDirection != 2 {
		t.Fatalf("normal east return from inn=(%d,%d,%d), want (7,13,2)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
	if err := state.EnterDungeonCamp(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.campMenu || len(state.Choices) != 7 ||
		state.restEncounterPeriod != 0 || state.restEncounterPercent != 0 {
		t.Fatalf("safe Tilverton camp mode=%v menu=%v choices=%v encounter=%d/%d",
			state.Mode, state.campMenu, state.Choices, state.restEncounterPeriod, state.restEncounterPercent)
	}
	if err := state.Select(6); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 7 || state.DungeonY != 13 {
		t.Fatalf("camp exit mode=%v position=(%d,%d), want original dungeon view", state.Mode, state.DungeonX, state.DungeonY)
	}
	state.DungeonX = 4
	if err := state.EnterDungeonCamp(); err != nil {
		t.Fatal(err)
	}
	if state.restEncounterPeriod != 1 || state.restEncounterPercent != 100 {
		t.Fatalf("unsafe Tilverton camp encounter=%d/%d, want 1/100",
			state.restEncounterPeriod, state.restEncounterPercent)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 || state.campMenu || !strings.Contains(state.Message, "休息突然中斷") ||
		(!strings.Contains(state.Message, "皇家巡邏隊") && !strings.Contains(state.Message, "皇家衛兵")) {
		t.Fatalf("unsafe interrupted rest mode=%v camp=%v message=%q",
			state.Mode, state.campMenu, state.Message)
	}
	if clock := state.GameTimeDisplay(); clock.Hour != 1 || clock.Minute != 2 {
		t.Fatalf("unsafe interrupted rest clock=%+v, want one hour plus two SEARCH-off dungeon minutes", clock)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("interrupted camp continuation mode=%v, want dungeon", state.Mode)
	}

	// Continue the same formal Tilverton session at the sage Filani's GEO
	// selector. The truth branch executes real ROB 1,50,0 before revealing
	// Journal Entry 38.
	state.partyRoster[0].Gold = 101
	cityGateClosedDuringWalk := false
	resumeKnownPathPause := func() {
		if state.Mode != ModeWilderness {
			return
		}
		wantGateClosed := requireGamePackText(t, &state, "tilverton.carriage-gate-closed")
		if state.Message == wantGateClosed {
			if state.DungeonX != 1 || state.DungeonY != 0 || state.DungeonDirection != 6 || len(state.Choices) != 1 {
				t.Fatalf("unexpected city-gate pause mode=%v position=(%d,%d,%d) choices=%v message=%q",
					state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection, state.Choices, state.Message)
			}
			cityGateClosedDuringWalk = true
			if err := state.Select(0); err != nil {
				t.Fatalf("continue past city-gate pause: %v", err)
			}
			if state.Mode == ModeEvent {
				if err := state.Continue(); err != nil {
					t.Fatalf("resume after city-gate pause: %v", err)
				}
			}
			return
		}
		wantSign := state.catalog.Text("ecl_tilverton_temple_sign", "ecl_tilverton_temple_sign")
		wantRumor := state.catalog.Text("ecl_tilverton_sewer_rumor", "ecl_tilverton_sewer_rumor")
		wantGreenRobesRumor := requireGamePackText(t, &state, "tilverton.green-robes-rumor")
		if (state.Message != wantSign && state.Message != wantRumor && state.Message != wantGreenRobesRumor) || len(state.Choices) != 1 {
			t.Fatalf("unexpected normal-path pause mode=%v position=(%d,%d,%d) choices=%v message=%q", state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection, state.Choices, state.Message)
		}
		if err := state.Select(0); err != nil {
			t.Fatalf("continue past normal-path pause: %v", err)
		}
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatalf("resume after normal-path pause: %v", err)
			}
		}
	}
	steps := []struct {
		dx, dy, direction int
	}{
		{0, -1, 0}, // (4,13) -> (4,12), after interrupted unsafe rest
		{0, -1, 0}, // (4,12) -> (4,11)
		{0, -1, 0}, // (4,11) -> (4,10)
		{0, -1, 0}, // (4,10) -> (4,9)
		{0, -1, 0}, // (4,9) -> (4,8)
		{0, -1, 0}, // (4,8) -> (4,7)
		{0, -1, 0}, // (4,7) -> (4,6)
		{0, -1, 0}, // (4,6) -> (4,5)
		{1, 0, 2},  // (4,5) -> (5,5)
		{1, 0, 2},  // (5,5) -> (6,5)
	}
	for index, step := range steps {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path step %d (%d,%d,%d): %v", index, step.dx, step.dy, step.direction, err)
		}
		resumeKnownPathPause()
		if index+1 < len(steps) && state.Mode != ModeDungeon {
			t.Fatalf("normal path step %d stopped at mode=%v position=(%d,%d)",
				index, state.Mode, state.DungeonX, state.DungeonY)
		}
	}
	if state.DungeonX != 6 || state.DungeonY != 5 || state.DungeonDirection != 2 {
		t.Fatalf("normal path reached (%d,%d,%d), want (6,5,2)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
	if state.DungeonWallRoof != 0x8A {
		t.Fatalf("Filani GEO selector=%#x, want 0x8A", state.DungeonWallRoof)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 5 ||
		!state.SceneCharacterRequested || state.SceneHeadBlock != 5 || state.SceneBodyBlock != 5 ||
		!strings.Contains(state.Message, "賢者菲拉妮") || !strings.Contains(state.Message, "印記") {
		t.Fatalf("Filani introduction mode=%v picture=%v:%d head/body=%d/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 2 ||
		state.Choices[0] != "是" || state.Choices[1] != "否" {
		t.Fatalf("Filani first menu mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 3 ||
		state.Choices[0] != "如實相告" || state.Choices[1] != "說謊" ||
		!strings.Contains(state.Message, "一半的財物") {
		t.Fatalf("Filani truth menu mode=%v choices=%v message=%q", state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].Gold != 50 {
		t.Fatalf("Filani ROB gold=%d, want floor(101*0.5)=50", state.partyRoster[0].Gold)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "冒險手札") || !strings.Contains(state.Message, "38") {
		t.Fatalf("Filani journal pause mode=%v choices=%v message=%q", state.Mode, state.Choices, state.Message)
	}
	// 手札 1 的三頁（開場）＋ 旅店的 31 ＋ 菲拉妮的 38 三頁 ＝ 7 頁。
	if len(state.JournalPages) != 7 ||
		state.JournalPages[4] != requireGamePackText(t, &state, "journal.38.1") ||
		state.JournalPages[5] != requireGamePackText(t, &state, "journal.38.2") ||
		state.JournalPages[6] != requireGamePackText(t, &state, "journal.38.3") {
		t.Fatalf("Journal Entry 38 was not unlocked as three readable pages: pages=%v", state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "離開此處") {
		t.Fatalf("Filani departure pause mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 6 || state.DungeonY != 5 {
		t.Fatalf("Filani continuation mode=%v position=(%d,%d), want same dungeon cell",
			state.Mode, state.DungeonX, state.DungeonY)
	}

	// Weaponers of Cormyr uses CMD_COMBAT as CityShop's engine-service
	// boundary after setting Area2.EnterShop and loading ITEM2 block 5.
	facilityCatalog := trainingTestCatalog(t)
	state.catalog = facilityCatalog
	state.partyRoster[0].Platinum = 0xFFFF
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{-1, 0, 6}, // (6,5) -> (5,5)
		{0, 1, 4},  // (5,5) -> (5,6)
		{0, 1, 4},  // (5,6) -> (5,7)
		{0, 1, 4},  // (5,7) -> (5,8)
		{0, 1, 4},  // (5,8) -> (5,9)
		{0, 1, 4},  // (5,9) -> (5,10)
		{0, 1, 4},  // (5,10) -> (5,11)
		{0, 1, 4},  // (5,11) -> (5,12)
		{-1, 0, 6}, // (5,12) -> (4,12)
		{-1, 0, 6}, // (4,12) -> (3,12)
		{-1, 0, 6}, // (3,12) -> (2,12)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to Weaponers step %d: %v", index, err)
		}
		resumeKnownPathPause()
		if index+1 < 11 && state.Mode != ModeDungeon {
			t.Fatalf("normal path to Weaponers step %d stopped at mode=%v position=(%d,%d)",
				index, state.Mode, state.DungeonX, state.DungeonY)
		}
	}
	if state.DungeonWallRoof != 0x84 {
		t.Fatalf("Weaponers GEO selector=%#x, want 0x84", state.DungeonWallRoof)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 4 ||
		!state.SceneCharacterRequested || state.SceneHeadBlock != 4 || state.SceneBodyBlock != 4 ||
		!strings.Contains(state.Message, "科米爾") || !strings.Contains(state.Message, "精鋼") {
		t.Fatalf("Weaponers introduction mode=%v picture=%v:%d head/body=%d/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 2 ||
		state.Choices[0] != "是" || state.Choices[1] != "否" {
		t.Fatalf("Weaponers menu mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	offers := state.ShopOffers()
	if state.Mode != ModePlace || !state.shopMenu || len(offers) == 0 || len(state.Choices) != 9 {
		t.Fatalf("Weaponers service mode=%v shop=%v offers=%d choices=%v",
			state.Mode, state.shopMenu, len(offers), state.Choices)
	}
	if state.Prompt != facilityCatalog.Text("shop_menu_prompt", "") ||
		state.Choices[0] != facilityCatalog.Text("shop_buy", "") ||
		state.Choices[8] != facilityCatalog.Text("shop_exit", "") {
		t.Fatalf("Weaponers service did not resolve stable locale IDs: prompt=%q choices=%v",
			state.Prompt, state.Choices)
	}
	beforeWorth := characterCoinGoldWorth(state.partyRoster[0])
	beforeEquipment := len(state.partyRoster[0].Equipment)
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.shopStockMenu || len(state.Choices) < 2 {
		t.Fatalf("Weaponers stock menu=%v choices=%v", state.shopStockMenu, state.Choices)
	}
	if state.Prompt != facilityCatalog.Text("shop_stock_prompt", "") {
		t.Fatalf("Weaponers stock prompt=%q, want locale ID result", state.Prompt)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].Equipment) != beforeEquipment+1 ||
		characterCoinGoldWorth(state.partyRoster[0]) != beforeWorth-uint32(offers[0].Price) ||
		len(state.ShopOffers()) != len(offers) {
		t.Fatalf("Weaponers purchase equipment=%d worth=%d offers=%d, want %d/%d/%d",
			len(state.partyRoster[0].Equipment), characterCoinGoldWorth(state.partyRoster[0]),
			len(state.ShopOffers()), beforeEquipment+1, beforeWorth-uint32(offers[0].Price), len(offers))
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.shopMenu {
		t.Fatalf("Weaponers purchase continuation mode=%v shop=%v", state.Mode, state.shopMenu)
	}
	if err := state.Select(8); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "準確命中") {
		t.Fatalf("Weaponers departure mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 2 || state.DungeonY != 12 {
		t.Fatalf("Weaponers continuation mode=%v position=(%d,%d), want same dungeon cell",
			state.Mode, state.DungeonX, state.DungeonY)
	}

	// The altar at GEO2 (0,7), terrain 0x92, requests PICTURE 6 and then
	// dispatches temple_shop through the same resumable CMD_COMBAT boundary.
	state.partyRoster[0].HitPoints = 1
	state.partyRoster[0].MaxHitPoints = 10
	state.partyRoster[0].Platinum = 0xFFFF
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{1, 0, 2},  // (2,12) -> (3,12)
		{0, -1, 0}, // (3,12) -> (3,11)
		{1, 0, 2},  // (3,11) -> (4,11)
		{0, -1, 0}, // (4,11) -> (4,10)
		{0, -1, 0}, // (4,10) -> (4,9)
		{0, -1, 0}, // (4,9) -> (4,8)
		{0, -1, 0}, // (4,8) -> (4,7)
		{-1, 0, 6}, // (4,7) -> (3,7)
		{-1, 0, 6}, // (3,7) -> (2,7)
		{-1, 0, 6}, // (2,7) -> (1,7)
		{-1, 0, 6}, // (1,7) -> (0,7)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to Gond temple step %d: %v", index, err)
		}
		resumeKnownPathPause()
		if index+1 < 11 && state.Mode != ModeDungeon {
			t.Fatalf("normal path to Gond temple step %d stopped at mode=%v position=(%d,%d) choices=%v message=%q",
				index, state.Mode, state.DungeonX, state.DungeonY, state.Choices, state.Message)
		}
	}
	if state.DungeonWallRoof != 0x92 {
		t.Fatalf("Gond altar GEO selector=%#x, want 0x92", state.DungeonWallRoof)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 6 ||
		!state.SceneCharacterRequested || state.SceneHeadBlock != 9 || state.SceneBodyBlock != 6 {
		t.Fatalf("Gond altar picture mode=%v picture=%v:%d head/body=%d/%d",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.templeMenu || len(state.Choices) != 6 {
		t.Fatalf("Gond temple service mode=%v temple=%v choices=%v",
			state.Mode, state.templeMenu, state.Choices)
	}
	if state.Prompt != facilityCatalog.Text("temple_prompt", "") ||
		state.Choices[0] != facilityCatalog.Text("temple_heal", "") ||
		state.Choices[5] != facilityCatalog.Text("temple_exit", "") {
		t.Fatalf("Gond temple did not resolve stable locale IDs: prompt=%q choices=%v",
			state.Prompt, state.Choices)
	}
	beforeWorth = characterCoinGoldWorth(state.partyRoster[0])
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].HitPoints <= 1 ||
		characterCoinGoldWorth(state.partyRoster[0]) != beforeWorth-100 {
		t.Fatalf("Gond cure hp=%d worth=%d, want healed and %d",
			state.partyRoster[0].HitPoints, characterCoinGoldWorth(state.partyRoster[0]), beforeWorth-100)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModePlace || !state.templeMenu {
		t.Fatalf("Gond cure continuation mode=%v temple=%v", state.Mode, state.templeMenu)
	}
	if err := state.Select(5); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 0 || state.DungeonY != 7 {
		t.Fatalf("Gond temple continuation mode=%v position=(%d,%d) picture=%v:%d message=%q choices=%v, want same dungeon cell",
			state.Mode, state.DungeonX, state.DungeonY, state.PictureRequested, state.PictureBlock,
			state.Message, state.Choices)
	}

	// The Hall of Training at GEO2 (5,2), terrain 0x8C, reuses PICTURE 4
	// before its YES branch invokes the location-specific PROGRAM 0 service.
	trainingCatalog := facilityCatalog
	character := &state.partyRoster[0]
	character.Class = party.ClassFighter
	character.Level = 1
	character.ClassLevels = [8]uint8{2: 1}
	character.Experience = 2001
	character.HitPoints, character.MaxHitPoints = 10, 10
	character.HealthStatus = party.HealthStatusOK
	character.Platinum = 0xFFFF
	beforeWorth = characterCoinGoldWorth(*character)
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{1, 0, 2},  // (0,7) -> (1,7)
		{1, 0, 2},  // (1,7) -> (2,7)
		{1, 0, 2},  // (2,7) -> (3,7)
		{1, 0, 2},  // (3,7) -> (4,7)
		{0, -1, 0}, // (4,7) -> (4,6)
		{0, -1, 0}, // (4,6) -> (4,5)
		{0, -1, 0}, // (4,5) -> (4,4)
		{0, -1, 0}, // (4,4) -> (4,3)
		{1, 0, 2},  // (4,3) -> (5,3)
		{0, -1, 0}, // (5,3) -> (5,2)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to training hall step %d: %v", index, err)
		}
		resumeKnownPathPause()
		if index+1 < 10 && state.Mode != ModeDungeon {
			t.Fatalf("normal path to training hall step %d stopped at mode=%v position=(%d,%d)",
				index, state.Mode, state.DungeonX, state.DungeonY)
		}
	}
	if state.DungeonWallRoof != 0x8C {
		t.Fatalf("training hall GEO selector=%#x, want 0x8C", state.DungeonWallRoof)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 4 ||
		!strings.Contains(state.Message, "接受訓練") {
		t.Fatalf("training picture mode=%v picture=%v:%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 2 {
		t.Fatalf("training question mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.trainingMenu || len(state.Choices) != 2 {
		t.Fatalf("training service mode=%v menu=%v choices=%v",
			state.Mode, state.trainingMenu, state.Choices)
	}
	if state.Prompt != trainingCatalog.Text("training_select_character", "") ||
		state.Choices[1] != trainingCatalog.Text("training_exit", "") {
		t.Fatalf("training service did not resolve stable locale IDs: prompt=%q choices=%v",
			state.Prompt, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.trainingConfirmMenu || len(state.Choices) != 2 {
		t.Fatalf("training confirmation=%v choices=%v", state.trainingConfirmMenu, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].Level != 2 || state.partyRoster[0].ClassLevels[2] != 2 ||
		state.partyRoster[0].MaxHitPoints <= 10 ||
		characterCoinGoldWorth(state.partyRoster[0]) != beforeWorth-1000 {
		t.Fatalf("training result level=%d classes=%v hp=%d/%d worth=%d, want level 2 and %d",
			state.partyRoster[0].Level, state.partyRoster[0].ClassLevels,
			state.partyRoster[0].HitPoints, state.partyRoster[0].MaxHitPoints,
			characterCoinGoldWorth(state.partyRoster[0]), beforeWorth-1000)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.trainingMenu {
		t.Fatalf("training result continuation mode=%v menu=%v", state.Mode, state.trainingMenu)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 5 || state.DungeonY != 2 {
		t.Fatalf("training departure mode=%v position=(%d,%d), want same dungeon cell",
			state.Mode, state.DungeonX, state.DungeonY)
	}
	if len(state.Choices) != 0 {
		t.Fatalf("training departure retained stale choices %v", state.Choices)
	}
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{0, 1, 4}, // (5,2) -> (5,3)
		{0, 1, 4}, // (5,3) -> (5,4)
		{0, 1, 4}, // (5,4) -> (5,5)
		{0, 1, 4}, // (5,5) -> (5,6)
		{0, 1, 4}, // (5,6) -> (5,7)
		{0, 1, 4}, // (5,7) -> (5,8)
		{0, 1, 4}, // (5,8) -> (5,9)
		{0, 1, 4}, // (5,9) -> (5,10)
		{1, 0, 2}, // (5,10) -> (6,10)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to tavern step %d: %v", index, err)
		}
		resumeKnownPathPause()
		if index+1 < 9 && state.Mode != ModeDungeon {
			t.Fatalf("normal path to tavern step %d stopped at mode=%v position=(%d,%d)",
				index, state.Mode, state.DungeonX, state.DungeonY)
		}
	}
	if state.DungeonWallRoof != 0x88 {
		t.Fatalf("tavern GEO selector=%#x, want 0x88", state.DungeonWallRoof)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 4 ||
		!state.SceneCharacterRequested || state.SceneHeadBlock != 4 || state.SceneBodyBlock != 4 ||
		state.Message != facilityCatalog.Text("ecl_tavern_pleasure", "") {
		t.Fatalf("tavern picture mode=%v picture=%v:%d head/body=%d/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 3 ||
		state.Choices[0] != facilityCatalog.Text("tavern_punch", "") ||
		state.Choices[1] != facilityCatalog.Text("tavern_drink", "") ||
		state.Choices[2] != facilityCatalog.Text("leave", "") {
		t.Fatalf("tavern action menu mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 4 ||
		state.Choices[0] != facilityCatalog.Text("tavern_dragon_breath", "") ||
		state.Choices[1] != facilityCatalog.Text("tavern_basilisk", "") ||
		state.Choices[2] != facilityCatalog.Text("tavern_lemonade", "") ||
		state.Choices[3] != facilityCatalog.Text("tavern_whiskey", "") {
		t.Fatalf("tavern drink menu mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.Choices[0] != facilityCatalog.Text("yes", "") ||
		state.Choices[1] != facilityCatalog.Text("no", "") ||
		!strings.Contains(state.Message, facilityCatalog.Text("ecl_tavern_special_1", "")) {
		t.Fatalf("tavern special-customer prompt choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 1 ||
		!strings.Contains(state.Message, facilityCatalog.Text("ecl_tavern_purple_2", "")) {
		t.Fatalf("tavern purple-sash pause choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.Choices[0] != facilityCatalog.Text("yes", "") ||
		state.Choices[1] != facilityCatalog.Text("no", "") ||
		!strings.Contains(state.Message, facilityCatalog.Text("ecl_tavern_commotion_2", "")) ||
		!strings.Contains(state.Message, facilityCatalog.Text("ecl_tavern_commotion_3", "")) {
		t.Fatalf("tavern investigate prompt choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 1 ||
		state.Message != requireGamePackText(t, &state, "journal-trigger.tavern-knife-17") {
		t.Fatalf("tavern knife pause choices=%v message=%q", state.Choices, state.Message)
	}
	foundJournal17 := slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.17"))
	if !foundJournal17 {
		t.Fatalf("Journal Entry 17 was not unlocked in-game: pages=%v", state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 6 || state.DungeonY != 10 ||
		len(state.Choices) != 0 {
		t.Fatalf("tavern continuation mode=%v position=(%d,%d) picture=%v:%d message=%q choices=%v, want same tavern cell without stale menu",
			state.Mode, state.DungeonX, state.DungeonY, state.PictureRequested, state.PictureBlock,
			state.Message, state.Choices)
	}
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{-1, 0, 6}, // (6,10) -> (5,10)
		{0, -1, 0}, // (5,10) -> (5,9)
		{0, -1, 0}, // (5,9) -> (5,8)
		{0, -1, 0}, // (5,8) -> (5,7)
		{-1, 0, 6}, // (5,7) -> (4,7)
		{-1, 0, 6}, // (4,7) -> (3,7)
		{0, 1, 4},  // (3,7) -> (3,8)
		{0, 1, 4},  // (3,8) -> (3,9)
		{0, 1, 4},  // (3,9) -> (3,10)
		{-1, 0, 6}, // (3,10) -> (2,10)
		{-1, 0, 6}, // (2,10) -> (1,10)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to high priest step %d: %v", index, err)
		}
		resumeKnownPathPause()
		if index+1 < 11 && state.Mode != ModeDungeon {
			t.Fatalf("normal path to high priest step %d stopped at mode=%v position=(%d,%d)",
				index, state.Mode, state.DungeonX, state.DungeonY)
		}
	}
	state.TurnDungeonWithGrid(grid, 2) // arrive facing west; rotate to the north-facing shrine
	if state.DungeonDirection != 0 {
		t.Fatalf("high priest approach direction=%d, want north after turn", state.DungeonDirection)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x8F || state.Mode != ModeDungeon || state.PictureRequested {
		t.Fatalf("normal route revisited high priest group mode=%v picture=%v:%d selector=%#x message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.DungeonWallRoof, state.Message)
	}
	if flag, ok := state.session.MemoryValue(0x4C03); !ok || flag != 0x80 {
		t.Fatalf("normal route high priest group flag=%#x,%v, want 0x80,true", flag, ok)
	}

	// The original map intentionally shares the 0x80 one-shot event group
	// between the temple sign and the high-priest handler.  The normal route
	// above therefore proves the revisit is silent; exercise the independent
	// high-priest branch in a fresh ECL session so this long-path test does not
	// erase the original one-shot semantics just to reach a later assertion.
	highPriest := NewStateFromECLBlocks(facilityCatalog, all, 0x01)
	highPriest.Mode = ModeDungeon
	highPriest.Location = LocationTilverton
	highPriest.LocationName = highPriest.catalog.Text("tilverton", "tilverton")
	highPriest.Area.InDungeon = true
	highPriest.Area.GameArea = 2
	highPriest.Area.Current3DMapBlockID = 1
	highPriest.GeoMapSet = 2
	highPriest.GeoMapBlock = 1
	highPriest.DungeonX, highPriest.DungeonY, highPriest.DungeonDirection = 1, 10, 0
	highPriest.DungeonWallType, _ = grid.WallWrapped(1, 10, 0)
	highPriest.DungeonWallRoof = grid.CellWrapped(1, 10).Terrain
	if err := highPriest.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if highPriest.DungeonWallRoof != 0x8F || highPriest.Mode != ModeEvent || !highPriest.PictureRequested ||
		highPriest.PictureBlock != 6 || highPriest.SceneHeadBlock != 6 || highPriest.SceneBodyBlock != 6 ||
		highPriest.Message != requireGamePackText(t, &highPriest, "tilverton.high-priest-intro") {
		t.Fatalf("high priest introduction selector=%#x mode=%v picture=%v:%d head/body=%d/%d message=%q",
			highPriest.DungeonWallRoof, highPriest.Mode, highPriest.PictureRequested, highPriest.PictureBlock,
			highPriest.SceneHeadBlock, highPriest.SceneBodyBlock, highPriest.Message)
	}
	if err := highPriest.Continue(); err != nil {
		t.Fatal(err)
	}
	if highPriest.Mode != ModeWilderness || len(highPriest.Choices) != 2 ||
		highPriest.Choices[0] != "是" || highPriest.Choices[1] != "否" {
		t.Fatalf("high priest story prompt mode=%v choices=%v message=%q",
			highPriest.Mode, highPriest.Choices, highPriest.Message)
	}
	if err := highPriest.Select(0); err != nil {
		t.Fatal(err)
	}
	if highPriest.Mode != ModeWilderness || len(highPriest.Choices) != 1 ||
		!strings.Contains(highPriest.Message, "移除詛咒") || !strings.Contains(highPriest.Message, "第 19 條") {
		t.Fatalf("high priest remove-curse pause mode=%v choices=%v message=%q",
			highPriest.Mode, highPriest.Choices, highPriest.Message)
	}
	foundJournal19 := slices.Contains(highPriest.JournalPages, requireGamePackText(t, &highPriest, "journal.19"))
	if !foundJournal19 {
		t.Fatalf("Journal Entry 19 was not unlocked in-game: pages=%v", highPriest.JournalPages)
	}
	if err := highPriest.Select(0); err != nil {
		t.Fatal(err)
	}
	if highPriest.Mode != ModeWilderness || len(highPriest.Choices) != 1 ||
		!strings.Contains(highPriest.Message, "離開此處") {
		t.Fatalf("high priest departure pause mode=%v choices=%v message=%q",
			highPriest.Mode, highPriest.Choices, highPriest.Message)
	}
	if err := highPriest.Select(0); err != nil {
		t.Fatal(err)
	}
	if highPriest.Mode != ModeDungeon || highPriest.DungeonX != 1 || highPriest.DungeonY != 10 ||
		len(highPriest.Choices) != 0 {
		t.Fatalf("high priest continuation mode=%v position=(%d,%d) message=%q choices=%v originals=%v, want same cell",
			highPriest.Mode, highPriest.DungeonX, highPriest.DungeonY, highPriest.Message, highPriest.Choices, highPriest.currentOriginalChoices)
	}
	// Keep the single generated integration-test hero alive across this long
	// path's many real encounters. A normal campaign has a full party; this
	// fixture deliberately tests ECL continuation rather than attrition.
	hero := state.PartyFighters()[0]
	hero.HitPoints, hero.MaxHitPoints = 20000, 20000
	hero.ArmorClass = -10
	hero.InitiativeBonus = 100
	hero.AttackBonus = 100
	hero.DamageDiceCount, hero.DamageDiceSides, hero.DamageBonus = 1, 1, 100
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{1, 0, 2},  // (1,10) -> (2,10)
		{1, 0, 2},  // (2,10) -> (3,10)
		{0, -1, 0}, // (3,10) -> (3,9)
		{0, -1, 0}, // (3,9) -> (3,8)
		{0, -1, 0}, // (3,8) -> (3,7)
		{1, 0, 2},  // (3,7) -> (4,7)
		{0, -1, 0}, // (4,7) -> (4,6)
		{0, -1, 0}, // (4,6) -> (4,5)
		{0, -1, 0}, // (4,5) -> (4,4)
		{0, -1, 0}, // (4,4) -> (4,3)
		{0, -1, 0}, // (4,3) -> (4,2)
		{0, -1, 0}, // (4,2) -> (4,1)
		{0, -1, 0}, // (4,1) -> (4,0)
		{-1, 0, 6}, // (4,0) -> (3,0)
		{-1, 0, 6}, // (3,0) -> (2,0)
		{-1, 0, 6}, // (2,0) -> (1,0)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to city gate step %d: %v", index, err)
		}
		resumeKnownPathPause()
		if index+1 < 16 && state.Mode != ModeDungeon {
			t.Fatalf("normal path to city gate step %d stopped at mode=%v position=(%d,%d)",
				index, state.Mode, state.DungeonX, state.DungeonY)
		}
	}
	if !cityGateClosedDuringWalk {
		t.Fatal("normal path did not trigger the city-gate closure")
	}
	if state.DungeonX != 1 || state.DungeonY != 0 || state.DungeonDirection != 6 {
		t.Fatalf("normal path reached city gate (%d,%d,%d), want (1,0,6)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("normal city-gate continuation mode=%v, want dungeon", state.Mode)
	}
	state.TurnDungeonWithGrid(grid, 2)
	if state.DungeonDirection != 0 {
		t.Fatalf("city gate approach direction=%d, want north after turn", state.DungeonDirection)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 11 ||
		state.SceneHeadBlock != 0xFF || state.Message != requireGamePackText(t, &state, "tilverton.carriage-make-way") {
		t.Fatalf("royal carriage picture mode=%v picture=%v:%d head=%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.SceneHeadBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		state.Message != requireGamePackText(t, &state, "tilverton.carriage-make-way") {
		t.Fatalf("royal carriage first pause mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.carriage-bond-compulsion") {
		t.Fatalf("royal carriage compulsion message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.carriage-false-king") {
		t.Fatalf("royal carriage false king message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.carriage-alarm") {
		t.Fatalf("royal carriage alarm message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() || len(state.CombatFighters()) != 6 ||
		len(state.CombatTargets()) != 5 {
		t.Fatalf("Royal Guard combat mode=%v active=%v fighters=%v targets=%v",
			state.Mode, state.CombatActive(), state.CombatFighters(), state.CombatTargets())
	}
	for turn := 0; turn < 10 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeWilderness ||
		state.Message != requireGamePackText(t, &state, "tilverton.carriage-abduction") {
		t.Fatalf("Royal Guard victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.Choices[0] != "是" || state.Choices[1] != "否" ||
		state.Message != requireGamePackText(t, &state, "tilverton.carriage-surrender") {
		t.Fatalf("surrender prompt choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 1 || state.Message != requireGamePackText(t, &state, "tilverton.carriage-jailed") {
		t.Fatalf("jail pause choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 2 ||
		state.SceneHeadBlock != 2 || state.SceneBodyBlock != 2 ||
		state.Message != requireGamePackText(t, &state, "tilverton.carriage-thief-rescue") {
		t.Fatalf("thief rescue picture mode=%v picture=%v:%d head/body=%d/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		state.Message != requireGamePackText(t, &state, "tilverton.carriage-thief-rescue") {
		t.Fatalf("thief rescue pause mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.carriage-guild-arrival") {
		t.Fatalf("thieves guild arrival message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x02 ||
		state.DungeonX != 1 || state.DungeonY != 12 || len(state.Choices) != 0 {
		t.Fatalf("thieves guild transition mode=%v block=%#x position=(%d,%d) choices=%v",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY, state.Choices)
	}
	geometryX, geometryY, geometryDirection := state.DungeonGeometryView()
	if geometryX != 9 || geometryY != 3 || geometryDirection != 4 {
		t.Fatalf("guild script-to-geometry view=(%d,%d,%d), want (9,3,4)",
			geometryX, geometryY, geometryDirection)
	}
	state.TurnDungeonWithGrid(grid, 0)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.guildmaster-greeting") {
		t.Fatalf("guildmaster greeting message=%q", state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.guildmaster-briefing") {
		t.Fatalf("guildmaster briefing message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.guild-breach") {
		t.Fatalf("guild breach message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.guild-fire-knife-command") {
		t.Fatalf("guild Fire Knife command message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.guild-poisoned-dagger") {
		t.Fatalf("guild poisoned dagger message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.guild-battle-joined") {
		t.Fatalf("guild battle joined message=%q", state.Message)
	}
	guildParty, guildEnemies := 0, 0
	guildFighters := state.CombatFighters()
	for _, fighter := range guildFighters {
		if fighter.Side == combat.SideParty {
			guildParty++
		} else {
			guildEnemies++
		}
	}
	if guildParty != 5 || guildEnemies != 13 {
		t.Fatalf("guild mixed battle party/enemy=%d/%d, want hero + 4 allied thieves vs 2 Fire Knives + 11 thieves",
			guildParty, guildEnemies)
	}
	for turn := 0; turn < 100 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeWilderness ||
		state.Message != requireGamePackText(t, &state, "journal-trigger.guildmaster-map-4") {
		t.Fatalf("guild victory mode=%v status=%v message=%q choices=%v",
			state.Mode, state.CombatStatus(), state.Message, state.Choices)
	}
	foundJournal4 := slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.4"))
	if !foundJournal4 {
		t.Fatalf("Journal Entry 4 was not unlocked in-game: pages=%v", state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if geometryX, geometryY, geometryDirection := state.DungeonGeometryView(); geometryX != 9 || geometryY != 3 || geometryDirection != 2 {
		t.Fatalf("guild battle return geometry=(%d,%d,%d), want (9,3,2)", geometryX, geometryY, geometryDirection)
	}
	guestBookSeen := false
	runningThievesWarningSeen := false
	fireKnivesSpotSeen := false
	assassinsAttackSeen := false
	metalAndAnimalsSeen := false
	bodiesAfterBattleSeen := false
	finishDungeonEncounter := func(context string) {
		if state.Mode != ModeCombat {
			return
		}
		if !state.CombatActive() {
			t.Fatalf("%s entered inactive combat", context)
		}
		for turn := 0; turn < 100 && state.CombatActive(); turn++ {
			if err := state.CombatAct(); err != nil {
				t.Fatalf("%s turn %d: %v", context, turn, err)
			}
		}
		if state.Mode == ModeEvent && state.CombatStatus() == combat.StatusPartyWon && state.eventReturnMode == ModeDungeon {
			if err := state.Continue(); err != nil {
				t.Fatalf("%s resume: %v", context, err)
			}
		}
		if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeDungeon {
			t.Fatalf("%s status=%v mode=%v return=%v eclReturn=%v eventReturn=%v inDungeon=%v message=%q choices=%v",
				context, state.CombatStatus(), state.Mode, state.combatReturnMode, state.eclMenuReturnMode, state.eventReturnMode, state.Area.InDungeon, state.Message, state.Choices)
		}
	}
	continueDungeonPress := func(context string) {
		if state.Mode != ModeWilderness || len(state.currentOriginalChoices) != 1 {
			return
		}
		if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.press-button-or-return-to-continue")); err != nil {
			t.Fatalf("%s press continuation: %v", context, err)
		}
		if state.Mode == ModeEvent && state.eventReturnMode == ModeDungeon {
			if err := state.Continue(); err != nil {
				t.Fatalf("%s press resume: %v", context, err)
			}
		}
	}
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{1, 0, 2}, // (9,3) -> (10,3)
		{1, 0, 2}, // (10,3) -> (11,3)
		{0, 1, 4}, // (11,3) -> (11,4)
		{0, 1, 4}, // (11,4) -> (11,5)
		{0, 1, 4}, // (11,5) -> (11,6)
		{0, 1, 4}, // (11,6) -> (11,7)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to guild halfling step %d: %v", index, err)
		}
		finishDungeonEncounter(fmt.Sprintf("normal guild interior random encounter step %d", index))
		if state.Mode != ModeDungeon && index+1 == 6 {
			if state.Message != requireGamePackText(t, &state, "tilverton.guild-halfling") || len(state.Choices) != 1 {
				t.Fatalf("normal path to guild halfling final pause mode=%v message=%q choices=%v",
					state.Mode, state.Message, state.Choices)
			}
		} else if state.Mode != ModeDungeon {
			if state.Message != requireGamePackText(t, &state, "tilverton.guild-guest-book") || len(state.Choices) != 1 {
				t.Fatalf("normal path to guild halfling step %d unexpected pause mode=%v message=%q choices=%v",
					index, state.Mode, state.Message, state.Choices)
			}
			guestBookSeen = true
			if err := state.Select(0); err != nil {
				t.Fatalf("continue past normal guild guest book: %v", err)
			}
			if state.Mode == ModeEvent {
				if err := state.Continue(); err != nil {
					t.Fatalf("resume after normal guild guest book: %v", err)
				}
			}
		}
		if index+1 < 6 && state.Mode != ModeDungeon {
			t.Fatalf("normal path to guild halfling step %d stopped at mode=%v geometry=%v message=%q choices=%v", index, state.Mode, func() string {
				x, y, direction := state.DungeonGeometryView()
				return fmt.Sprintf("(%d,%d,%d)", x, y, direction)
			}(), state.Message, state.Choices)
		}
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "tilverton.guild-halfling") {
		t.Fatalf("guild halfling event mode=%v message=%q choices=%v", state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	doorFlags, doorFlagsOK := grid.WallDoorFlagsWrapped(11, 7, 2)
	if !doorFlagsOK || doorFlags != 2 {
		t.Fatalf("guild kennel doorway flags=%d found=%v, want locked detail 2", doorFlags, doorFlagsOK)
	}
	state.partyRoster[0].Abilities.StrengthFull = 25
	doorResult := state.BashDungeonDoor(doorFlags)
	if !doorResult.Opened || !grid.UnlockDoorWrapped(11, 7, 2) {
		t.Fatalf("guild kennel doorway bash result=%+v", doorResult)
	}
	if err := state.MoveDungeon(grid, 1, 0, 2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "tilverton.guild-kennel-intro") {
		t.Fatalf("guild kennel intro mode=%v message=%q choices=%v", state.Mode, state.Message, state.Choices)
	}
	hero = state.PartyFighters()[0]
	hero.HitPoints, hero.MaxHitPoints = 2000, 2000
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("guild kennel combat mode=%v active=%v fighters=%v",
			state.Mode, state.CombatActive(), state.CombatFighters())
	}
	kennelParty, kennelFireKnives, kennelDogs := 0, 0, 0
	for _, fighter := range state.CombatFighters() {
		switch {
		case fighter.Side == combat.SideParty:
			kennelParty++
		case fighter.Name == "FIRE KNIFE":
			kennelFireKnives++
		case fighter.Name == "FIGHTING DOG":
			kennelDogs++
			if fighter.MonsterType != combat.MonsterTypeAnimal {
				t.Fatalf("FIGHTING DOG monster type=%#x, want animal", fighter.MonsterType)
			}
		}
	}
	if kennelParty != 1 || kennelFireKnives != 1 || kennelDogs < 8 {
		t.Fatalf("guild kennel party/fire-knives/dogs=%d/%d/%d",
			kennelParty, kennelFireKnives, kennelDogs)
	}
	for turn := 0; turn < 30 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeWilderness ||
		state.Message != requireGamePackText(t, &state, "tilverton.guild-kennel-aftermath") {
		t.Fatalf("guild kennel result mode=%v status=%v message=%q choices=%v",
			state.Mode, state.CombatStatus(), state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if geometryX, geometryY, geometryDirection := state.DungeonGeometryView(); geometryX != 12 || geometryY != 7 || geometryDirection != 2 {
		t.Fatalf("guild kennel return geometry=(%d,%d,%d), want (12,7,2)", geometryX, geometryY, geometryDirection)
	}
	state.TurnDungeonWithGrid(grid, -2)
	secondDoorFlags, secondDoorFlagsOK := grid.WallDoorFlagsWrapped(12, 7, 0)
	if !secondDoorFlagsOK || secondDoorFlags != 2 {
		t.Fatalf("guild cages north doorway flags=%d found=%v, want locked detail 2", secondDoorFlags, secondDoorFlagsOK)
	}
	secondDoorResult := state.BashDungeonDoor(secondDoorFlags)
	if !secondDoorResult.Opened || !grid.UnlockDoorWrapped(12, 7, 0) {
		t.Fatalf("guild cages north doorway bash result=%+v", secondDoorResult)
	}
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{0, -1, 0}, // (12,7) -> (12,6)
		{0, -1, 0}, // (12,6) -> (12,5)
		{0, -1, 0}, // (12,5) -> (12,4)
	} {
		if err := state.MoveDungeon(grid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to guild cages north corridor step %d: %v", index, err)
		}
	}
	state.TurnDungeonWithGrid(grid, 2)
	for index := 0; index < 3; index++ {
		if err := state.MoveDungeon(grid, 1, 0, 2); err != nil {
			t.Fatalf("normal path to guild cages east corridor step %d: %v", index, err)
		}
	}
	state.TurnDungeonWithGrid(grid, 2)
	for index := 0; index < 3; index++ {
		if err := state.MoveDungeon(grid, 0, 1, 4); err != nil {
			t.Fatalf("normal path to guild cages south corridor step %d: %v", index, err)
		}
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "tilverton.guild-monkey-cages") {
		t.Fatalf("guild cages mode=%v message=%q choices=%v", state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !guestBookSeen {
		t.Fatal("normal guild interior path did not visit the guest book")
	}
	if geometryX, geometryY, geometryDirection := state.DungeonGeometryView(); geometryX != 15 || geometryY != 7 || geometryDirection != 4 {
		t.Fatalf("guild cages return geometry=(%d,%d,%d), want (15,7,4)", geometryX, geometryY, geometryDirection)
	}
	walkGuildInterior := func(context string, dx, dy, direction int) {
		if err := state.MoveDungeon(grid, dx, dy, direction); err != nil {
			t.Fatalf("%s: %v", context, err)
		}
		assassinsAttackSeen = assassinsAttackSeen || state.MessageContainsGamePackText("tilverton.guild-assassins-attack")
		metalAndAnimalsSeen = metalAndAnimalsSeen || state.MessageContainsGamePackText("tilverton.guild-metal-and-animals")
		bodiesAfterBattleSeen = bodiesAfterBattleSeen || state.MessageContainsGamePackText("tilverton.guild-bodies-after-battle")
		if state.Mode == ModeWilderness && state.MessageContainsGamePackText("tilverton.running-thieves") {
			if len(state.currentOriginalChoices) != 3 {
				t.Fatalf("%s running-thieves choices=%v", context, state.currentOriginalChoices)
			}
			if err := state.Select(requireGamePackOptionIndex(t, &state, "tilverton.option.remain-calm")); err != nil {
				t.Fatalf("%s remain-calm selection: %v", context, err)
			}
			if !state.MessageContainsGamePackText("tilverton.running-thieves-warning") {
				t.Fatalf("%s remain-calm warning was not resolved through game-pack: message=%q", context, state.Message)
			}
			runningThievesWarningSeen = true
			continueDungeonPress(context + " remain-calm")
		}
		if state.Mode == ModeWilderness && state.MessageContainsGamePackText("tilverton.fire-knives-spot-you") {
			fireKnivesSpotSeen = true
		}
		continueDungeonPress(context)
		finishDungeonEncounter(context)
		if state.Mode != ModeDungeon {
			t.Fatalf("%s stopped at mode=%v geometry=(%d,%d,%d) originalEvent=%q originalChoices=%v currentOriginalChoices=%v message=%q choices=%v", context, state.Mode,
				func() int { x, _, _ := state.DungeonGeometryView(); return x }(),
				func() int { _, y, _ := state.DungeonGeometryView(); return y }(),
				func() uint8 { _, _, direction := state.DungeonGeometryView(); return direction }(), state.OriginalEvent, state.OriginalChoices, state.currentOriginalChoices, state.Message, state.Choices)
		}
	}
	state.TurnDungeonWithGrid(grid, -4)
	exitDoorFlags, exitDoorFlagsOK := grid.WallDoorFlagsWrapped(15, 7, 0)
	if !exitDoorFlagsOK || exitDoorFlags != 2 {
		t.Fatalf("guild exit doorway flags=%d found=%v, want locked detail 2", exitDoorFlags, exitDoorFlagsOK)
	}
	exitDoorResult := state.BashDungeonDoor(exitDoorFlags)
	if !exitDoorResult.Opened || !grid.UnlockDoorWrapped(15, 7, 0) {
		t.Fatalf("guild exit doorway bash result=%+v", exitDoorResult)
	}
	for index := 0; index < 3; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages north step %d", index), 0, -1, 0)
	}
	state.TurnDungeonWithGrid(grid, -2)
	for index := 0; index < 3; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages west step %d", index), -1, 0, 6)
	}
	state.TurnDungeonWithGrid(grid, -2)
	for index := 0; index < 3; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages south step %d", index), 0, 1, 4)
	}
	state.TurnDungeonWithGrid(grid, 2)
	walkGuildInterior("normal path from guild cages to central corridor", -1, 0, 6)
	state.TurnDungeonWithGrid(grid, -2)
	for index := 0; index < 2; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages lower corridor step %d", index), 0, 1, 4)
	}
	state.TurnDungeonWithGrid(grid, 2)
	walkGuildInterior("normal path from guild cages west return", -1, 0, 6)
	state.TurnDungeonWithGrid(grid, 2)
	for index := 0; index < 3; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages upper corridor step %d", index), 0, -1, 0)
	}
	state.TurnDungeonWithGrid(grid, 0)
	for index := 0; index < 2; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages west hall step %d", index), -1, 0, 6)
	}
	state.TurnDungeonWithGrid(grid, -2)
	for index := 0; index < 6; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages south hall step %d", index), 0, 1, 4)
	}
	state.TurnDungeonWithGrid(grid, -2)
	for index := 0; index < 2; index++ {
		walkGuildInterior(fmt.Sprintf("normal path from guild cages east hall step %d", index), 1, 0, 2)
	}
	state.TurnDungeonWithGrid(grid, 2)
	sewerDoorFlags, sewerDoorFlagsOK := grid.WallDoorFlagsWrapped(10, 13, 4)
	if !sewerDoorFlagsOK || sewerDoorFlags != 2 {
		t.Fatalf("guild sewer doorway flags=%d found=%v, want locked detail 2", sewerDoorFlags, sewerDoorFlagsOK)
	}
	sewerDoorResult := state.BashDungeonDoor(sewerDoorFlags)
	if !sewerDoorResult.Opened || !grid.UnlockDoorWrapped(10, 13, 4) {
		t.Fatalf("guild sewer doorway bash result=%+v", sewerDoorResult)
	}
	walkGuildInterior("normal path to guild sewer approach", 0, 1, 4)
	walkGuildInterior("normal path through guild sewer doorway", 0, 1, 4)
	if err := state.MoveDungeon(grid, 0, 1, 4); err != nil {
		t.Fatalf("normal path into guild sewer event: %v", err)
	}
	if !runningThievesWarningSeen || !fireKnivesSpotSeen || !assassinsAttackSeen || !metalAndAnimalsSeen || !bodiesAfterBattleSeen {
		t.Fatalf("normal guild interior path warning coverage running-thieves=%v fire-knives=%v assassins=%v metal-and-animals=%v bodies=%v",
			runningThievesWarningSeen, fireKnivesSpotSeen, assassinsAttackSeen, metalAndAnimalsSeen, bodiesAfterBattleSeen)
	}
	if state.Mode != ModeEvent || state.Message != requireGamePackText(t, &state, "tilverton.guild-sewer-traces") {
		t.Fatalf("guild sewer door mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Yulash Pit entrance continue mode=%v return=%v message=%q",
			state.Mode, state.eventReturnMode, state.Message)
	}
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 3 {
		t.Fatalf("guild sewer exit mode=%v block=%#x script=(%d,%d,%d) geo=%d/%d terrain=%#x message=%q choices=%v",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.GeoMapSet, state.GeoMapBlock, state.DungeonWallRoof, state.Message, state.Choices)
	}
	if state.Mode != ModeWilderness || state.GeoMapSet != 2 || state.GeoMapBlock != 3 ||
		state.DungeonX != 0 || state.DungeonY != 1 || state.DungeonDirection != 4 ||
		state.Message != requireGamePackText(t, &state, "tilverton.sewers-entry") {
		t.Fatalf("sewer entry mode=%v block=%#x script=(%d,%d,%d) geo=%d/%d message=%q",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.GeoMapSet, state.GeoMapBlock, state.Message)
	}
	sewerGrid, ok := geoCatalog.Lookup(geo.MapRef{Set: 2, BlockID: 3})
	if !ok {
		t.Fatal("Tilverton sewer GEO2 block 3 is absent")
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{1, 0, 2},  // (0,1) -> (1,1)
		{0, 1, 4},  // (1,1) -> (1,2)
		{0, 1, 4},  // (1,2) -> (1,3)
		{0, 1, 4},  // (1,3) -> (1,4)
		{-1, 0, 6}, // (1,4) -> (0,4)
		{0, 1, 4},  // (0,4) -> (0,5)
		{0, 1, 4},  // (0,5) -> (0,6)
		{1, 0, 2},  // (0,6) -> (1,6)
		{0, 1, 4},  // (1,6) -> (1,7)
		{0, 1, 4},  // (1,7) -> (1,8)
	} {
		if err := state.MoveDungeon(sewerGrid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal path to sewer checkpoint step %d: %v", index, err)
		}
	}
	if state.MessageContainsGamePackText("tilverton.sewers.guild-battle-echoes") {
		if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.press-button-or-return-to-continue")); err != nil {
			t.Fatalf("continue sewer guild battle echoes: %v", err)
		}
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatalf("resume after sewer guild battle echoes: %v", err)
			}
		}
		if state.Mode == ModeDungeon {
			if err := state.SearchDungeonLocation(); err != nil {
				t.Fatalf("search sewer checkpoint after guild battle echoes: %v", err)
			}
		}
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "tilverton.sewers-checkpoint") ||
		len(state.Choices) != 2 {
		t.Fatalf("sewer checkpoint mode=%v terrain=%#x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message, state.Choices)
	}
	hero = state.PartyFighters()[0]
	hero.HitPoints, hero.MaxHitPoints = 2000, 2000
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("sewer checkpoint refusal mode=%v active=%v message=%q",
			state.Mode, state.CombatActive(), state.Message)
	}
	fireKnives := 0
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideEnemy && fighter.Name == "FIRE KNIFE" {
			fireKnives++
		}
	}
	if fireKnives != 5 {
		t.Fatalf("sewer checkpoint Fire Knives=%d, want 5; fighters=%v",
			fireKnives, state.CombatFighters())
	}
	for turn := 0; turn < 30 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeWilderness ||
		state.Message != requireGamePackText(t, &state, "tilverton.sewers-hide-bodies") {
		t.Fatalf("sewer checkpoint status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 13 || state.DungeonY != 10 ||
		state.DungeonDirection != 2 || state.DungeonWallType != 0x0F ||
		state.DungeonWallRoof != 0x83 {
		t.Fatalf("data-pack Fire Knife handoff mode=%v position=(%d,%d,%d) wall=(%#x,%#x)",
			state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.DungeonWallType, state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "tilverton.sewers-knight-appears") {
		t.Fatalf("sewer knight first pause mode=%v terrain=%#x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "tilverton.sewers-knight-allegiance") ||
		len(state.Choices) != 3 || state.Choices[1] != "娜卡西亞公主" {
		t.Fatalf("sewer knight allegiance mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "tilverton.sewers-knight-princess-friend") {
		t.Fatalf("sewer knight Princess branch mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.Message != "" || len(state.Choices) != 0 {
		t.Fatalf("sewer knight revisit mode=%v message=%q choices=%v, want consumed event",
			state.Mode, state.Message, state.Choices)
	}
	if err := state.ToggleDungeonSearch(); err != nil {
		t.Fatalf("enable persistent sewer SEARCH: %v", err)
	}
	if !state.DungeonSearchActive() {
		t.Fatal("persistent sewer SEARCH did not turn on")
	}
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{-1, 0, 6}, // (13,10) -> (12,10)
		{-1, 0, 6}, // (12,10) -> (11,10)
		{-1, 0, 6}, // (11,10) -> (10,10)
		{0, 1, 4},  // (10,10) -> (10,11)
		{0, 1, 4},  // (10,11) -> (10,12)
		{-1, 0, 6}, // searched wall=09: (10,12) -> (9,12)
		{0, 1, 4},  // (9,12) -> (9,13)
		{0, 1, 4},  // (9,13) -> (9,14)
		{0, 1, 4},  // (9,14) -> (9,15)
		{-1, 0, 6}, // (9,15) -> (8,15)
	} {
		if err := state.MoveDungeon(sewerGrid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal sewer Search route step %d: %v", index, err)
		}
	}
	if !state.dungeonSearchEdges["tilverton.sewers.wall-09-west"] {
		t.Fatalf("wall=09 was not discovered by persistent SEARCH: edges=%v", state.dungeonSearchEdges)
	}
	if state.Mode != ModeDungeon {
		if len(state.Choices) == 0 {
			t.Fatalf("sewer route pause at external-exit source has no continuation: mode=%v message=%q", state.Mode, state.Message)
		}
		if err := state.Select(0); err != nil {
			t.Fatalf("continue sewer external-exit source event: %v", err)
		}
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatalf("resume sewer external-exit source event: %v", err)
			}
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("sewer route did not return to dungeon before external exit: mode=%v message=%q choices=%v", state.Mode, state.Message, state.Choices)
	}
	if err := state.MoveDungeon(sewerGrid, 0, 1, 4); err != nil {
		t.Fatalf("normal sewer E2 external exit: %v", err)
	}
	if state.Mode != ModeWilderness || state.session.CurrentBlockID() != 4 ||
		state.GeoMapSet != 2 || state.GeoMapBlock != 4 ||
		state.DungeonX != 6 || state.DungeonY != 1 || state.DungeonDirection != 4 ||
		state.LoadPieces != [3]uint16{1, 2, 4} ||
		state.Message != requireGamePackText(t, &state, "fire-knife.hideout-entry") {
		t.Fatalf("Fire Knife hideout entry mode=%v block=%#x script=(%d,%d,%d) geo=%d/%d pieces=%v message=%q choices=%v",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.GeoMapSet, state.GeoMapBlock, state.LoadPieces,
			state.Message, state.Choices)
	}
	if len(state.Choices) == 0 {
		t.Fatalf("Fire Knife hideout entry has no normal-player continuation: message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatalf("enter Fire Knife hideout from E2 handoff: %v", err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Fire Knife hideout did not enter dungeon mode: mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.Choices)
	}
	// SEARCH was needed to discover the sewer wall=09.  Turn it off here using
	// the same player command a campaign would use, so the interior route is
	// deterministic and only exercises the room events that the player walks
	// through.
	if state.DungeonSearchActive() {
		if err := state.ToggleDungeonSearch(); err != nil {
			t.Fatalf("disable persistent SEARCH before Fire Knife interior route: %v", err)
		}
	}
	hideoutGrid, ok := geoCatalog.Lookup(geo.MapRef{Set: 2, BlockID: 4})
	if !ok {
		t.Fatal("Fire Knife hideout GEO2 block 4 is absent")
	}
	// Keep the route itself normal while giving the single opening character a
	// deterministic combat fixture strong enough to finish the leader battle.
	// This changes no map position, ECL flag, or story choice; it only keeps the
	// long integration test bounded before the real post-victory continuation.
	for index := range state.party {
		state.party[index].HitPoints = 10000
		state.party[index].MaxHitPoints = 10000
		state.party[index].AttackBonus = 100
		state.party[index].DamageDiceCount = 1
		state.party[index].DamageDiceSides = 1
		state.party[index].DamageBonus = 100
		state.party[index].AttacksPerTurn = 8
		state.party[index].InitiativeBonus = 100
	}
	continueHideoutEvent := func(step int) {
		if state.Mode == ModeCombat {
			finishDungeonEncounter(fmt.Sprintf("normal Fire Knife hideout combat step %d", step))
		}
		for pause := 0; state.Mode != ModeDungeon; pause++ {
			if pause >= 8 || len(state.Choices) == 0 {
				t.Fatalf("Fire Knife hideout event did not resume at step %d: mode=%v message=%q choices=%v",
					step, state.Mode, state.Message, state.Choices)
			}
			choice := 0
			if state.Message == requireGamePackText(t, &state, "fire-knife.blade-barrier") {
				choice = requireGamePackOptionIndex(t, &state, "ecl-option.wait")
			}
			if err := state.Select(choice); err != nil {
				t.Fatalf("continue Fire Knife hideout event at step %d: %v", step, err)
			}
			if state.Mode == ModeEvent {
				if err := state.Continue(); err != nil {
					t.Fatalf("resume Fire Knife hideout event at step %d: %v", step, err)
				}
			}
		}
	}
	// This is the raw GEO2 block-4 route from the E2 entrance at (6,1) to
	// terrain 87 at (3,13), the Fire Knife leader encounter.  It intentionally
	// crosses the blade barrier and frozen-room cells instead of injecting a
	// coordinate or ECL selector, so every transition is driven by MoveDungeon
	// and the normal SEARCH/LOOK lifecycle.
	for index, step := range []struct {
		dx, dy, direction int
	}{
		{0, 1, 4},  // (6,1) -> (6,2)
		{-1, 0, 6}, // (6,2) -> (5,2), blade barrier
		{-1, 0, 6}, // (5,2) -> (4,2), frozen room
		{0, 1, 4},  // (4,2) -> (4,3)
		{-1, 0, 6}, // (4,3) -> (3,3), frozen room
		{0, 1, 4},  // (3,3) -> (3,4)
		{1, 0, 2},  // (3,4) -> (4,4)
		{0, 1, 4},  // (4,4) -> (4,5)
		{1, 0, 2},  // (4,5) -> (5,5)
		{0, 1, 4},  // (5,5) -> (5,6)
		{0, 1, 4},  // (5,6) -> (5,7)
		{1, 0, 2},  // (5,7) -> (6,7)
		{1, 0, 2},  // (6,7) -> (7,7)
		{1, 0, 2},  // (7,7) -> (8,7)
		{1, 0, 2},  // (8,7) -> (9,7)
		{0, 1, 4},  // (9,7) -> (9,8)
		{1, 0, 2},  // (9,8) -> (10,8)
		{0, 1, 4},  // (10,8) -> (10,9)
		{0, 1, 4},  // (10,9) -> (10,10)
		{0, 1, 4},  // (10,10) -> (10,11)
		{-1, 0, 6}, // (10,11) -> (9,11)
		{0, 1, 4},  // (9,11) -> (9,12)
		{-1, 0, 6}, // (9,12) -> (8,12)
		{-1, 0, 6}, // (8,12) -> (7,12)
		{-1, 0, 6}, // (7,12) -> (6,12)
		{-1, 0, 6}, // (6,12) -> (5,12)
		{-1, 0, 6}, // (5,12) -> (4,12)
		{0, 1, 4},  // (4,12) -> (4,13)
		{-1, 0, 6}, // (4,13) -> (3,13), leader
	} {
		if err := state.MoveDungeon(hideoutGrid, step.dx, step.dy, step.direction); err != nil {
			t.Fatalf("normal Fire Knife hideout leader route step %d: %v", index, err)
		}
		if index+1 < 29 {
			continueHideoutEvent(index)
		}
	}
	if state.Mode == ModeEvent {
		if state.Message != requireGamePackText(t, &state, "journal-trigger.fire-knives-leader-11") {
			t.Fatalf("normal Fire Knife leader pre-combat pause message=%q choices=%v picture=%v:%d",
				state.Message, state.Choices, state.PictureRequested, state.PictureBlock)
		}
		if err := state.Continue(); err != nil {
			t.Fatalf("continue Fire Knife leader journal pause: %v", err)
		}
	}
	for pause := 0; state.Mode != ModeCombat; pause++ {
		if pause >= 3 {
			t.Fatalf("normal Fire Knife leader did not reach combat: mode=%v message=%q choices=%v",
				state.Mode, state.Message, state.Choices)
		}
		if len(state.Choices) == 0 {
			t.Fatalf("normal Fire Knife leader continuation has no input: mode=%v message=%q",
				state.Mode, state.Message)
		}
		if err := state.Select(0); err != nil {
			t.Fatalf("continue Fire Knife leader pre-combat choice: %v", err)
		}
	}
	if state.Mode != ModeCombat || !state.CombatActive() || state.DungeonX != 3 || state.DungeonY != 13 ||
		state.DungeonWallRoof != 0x87 {
		t.Fatalf("normal Fire Knife leader route mode=%v active=%v position=(%d,%d) terrain=%#x message=%q choices=%v",
			state.Mode, state.CombatActive(), state.DungeonX, state.DungeonY,
			state.DungeonWallRoof, state.Message, state.Choices)
	}
	leaderEnemies := 0
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideEnemy {
			leaderEnemies++
		}
	}
	if leaderEnemies != 21 {
		t.Fatalf("normal Fire Knife leader route enemies=%d fighters=%v, want 21", leaderEnemies, state.CombatFighters())
	}
	for turn := 0; turn < 100 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatalf("normal Fire Knife leader victory turn %d: %v", turn, err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon {
		t.Fatalf("normal Fire Knife leader status=%v mode=%v message=%q choices=%v",
			state.CombatStatus(), state.Mode, state.Message, state.Choices)
	}
	if !state.treasureMenu || len(state.Choices) < 1 {
		t.Fatalf("normal Fire Knife leader victory did not expose loot service: mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatalf("close normal Fire Knife leader loot service: %v", err)
	}
	for step := 0; step < 6; step++ {
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatalf("continue Fire Knife post-victory picture %d: %v", step, err)
			}
		}
		if state.Mode != ModeWilderness || len(state.Choices) == 0 {
			t.Fatalf("Fire Knife post-victory continuation %d mode=%v choices=%v message=%q picture=%v:%d",
				step, state.Mode, state.Choices, state.Message, state.PictureRequested, state.PictureBlock)
		}
		if err := state.Select(0); err != nil {
			t.Fatalf("continue Fire Knife post-victory choice %d: %v", step, err)
		}
	}
	if !state.PictureRequested || !state.BigPictureRequested || state.PictureBlock != 121 {
		t.Fatalf("Fire Knife post-victory world edge picture=%v big=%v block=%d message=%q choices=%v",
			state.PictureRequested, state.BigPictureRequested, state.PictureBlock, state.Message, state.Choices)
	}
	if err := state.Continue(); err != nil {
		t.Fatalf("continue Fire Knife post-victory world edge: %v", err)
	}
	if state.Mode != ModeWilderness || state.Location != LocationTilverton || state.Area.InDungeon ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Fire Knife post-victory world return mode=%v location=%v area=%+v choices=%v",
			state.Mode, state.Location, state.Area, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatalf("continue Tilverton patrol choice: %v", err)
	}
	for turn := 0; turn < 100 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatalf("normal Fire Knife patrol victory turn %d: %v", turn, err)
		}
	}
	if state.Mode != ModeWilderness || !reflect.DeepEqual(state.currentOriginalChoices, []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Tilverton patrol continuation mode=%v choices=%v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatalf("continue Tilverton journey choice: %v", err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"SHADOWDALE", "ASHABENFORD", "DAGGER FALLS"}) {
		t.Fatalf("post-Fire-Knife journey destinations=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatalf("select Ashabenford after Fire Knife: %v", err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("Ashabenford route choices=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatalf("select Ashabenford trail after Fire Knife: %v", err)
	}
	if state.Message != requireGamePackText(t, &state, "ashabenford.tilvers-gap") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Tilver's Gap continuation choices=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatalf("start Tilver's Gap combat after Fire Knife: %v", err)
	}
	if state.Mode != ModeCombat || len(state.CombatFighters()) != 9 {
		t.Fatalf("Tilver's Gap combat mode=%v fighters=%d message=%q",
			state.Mode, len(state.CombatFighters()), state.Message)
	}
	for turn := 0; turn < 100 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatalf("Tilver's Gap victory turn %d: %v", turn, err)
		}
	}
	if state.Mode != ModeWilderness || state.Location != LocationAshabenford ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) ||
		state.Message != requireGamePackText(t, &state, "ashabenford.edge") {
		t.Fatalf("Ashabenford arrival mode=%v location=%v choices=%#v message=%q",
			state.Mode, state.Location, state.currentOriginalChoices, state.Message)
	}
	for address, want := range map[uint16]uint16{0x4C83: 1, 0x4C9B: 2, 0x4CA1: 2} {
		if got, ok := state.session.MemoryValue(address); !ok || got != want {
			t.Fatalf("Ashabenford memory[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}

	// Continue the same new-game session through an actual city visit and the
	// next main-line world handoff.  The fixed fixture below covers the larger
	// city/event corpus; this normal-session tail proves that the Fire Knife
	// victory does not strand the player at the first Ashabenford edge.
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.enter-city")); err != nil {
		t.Fatalf("enter Ashabenford after Fire Knife: %v", err)
	}
	if !state.PictureRequested || state.Message != requireGamePackText(t, &state, "ashabenford.places") {
		t.Fatalf("Ashabenford places picture=%v block=%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatalf("continue Ashabenford places: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.bar")); err != nil {
		t.Fatalf("visit Ashabenford ale house: %v", err)
	}
	if state.Message != requireGamePackText(t, &state, "ashabenford.ale-house") {
		t.Fatalf("Ashabenford ale house message=%q choices=%v", state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.relax")); err != nil {
		t.Fatalf("relax in Ashabenford ale house: %v", err)
	}
	if state.Message != requireGamePackText(t, &state, "ashabenford.tavern-tale-28") {
		t.Fatalf("Ashabenford Tavern Tale 28 message=%q", state.Message)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.press-button-or-return-to-continue")); err != nil {
		t.Fatalf("continue Ashabenford Tavern Tale 28: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.exit")); err != nil {
		t.Fatalf("leave Ashabenford ale house: %v", err)
	}
	if state.PictureRequested {
		if err := state.Continue(); err != nil {
			t.Fatalf("continue Ashabenford places after ale house: %v", err)
		}
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.leave")); err != nil {
		t.Fatalf("leave Ashabenford city: %v", err)
	}
	if state.PictureRequested {
		if err := state.Continue(); err != nil {
			t.Fatalf("continue Ashabenford world edge: %v", err)
		}
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.journey-on")); err != nil {
		t.Fatalf("journey on from Ashabenford: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.the-standing-stone")); err != nil {
		t.Fatalf("travel to Standing Stone: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.trail")); err != nil {
		t.Fatalf("take Standing Stone trail: %v", err)
	}
	if state.Message != requireGamePackText(t, &state, "shadow-gap.fire-knives-patrol") {
		t.Fatalf("Standing Stone patrol message=%q choices=%v", state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.press-button-or-return-to-continue")); err != nil {
		t.Fatalf("continue Standing Stone patrol warning: %v", err)
	}
	if state.Mode != ModeCombat || len(state.CombatFighters()) != 7 {
		t.Fatalf("Standing Stone patrol mode=%v fighters=%d message=%q",
			state.Mode, len(state.CombatFighters()), state.Message)
	}
	for turn := 0; turn < 100 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatalf("Standing Stone patrol turn %d: %v", turn, err)
		}
	}
	if state.Message != requireGamePackText(t, &state, "standing-stone.grey-man") {
		t.Fatalf("Standing Stone grey-man message=%q choices=%v", state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.press-button-or-return-to-continue")); err != nil {
		t.Fatalf("continue Standing Stone grey-man: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.thank-him")); err != nil {
		t.Fatalf("hear Standing Stone masters: %v", err)
	}
	if state.Message != requireGamePackText(t, &state, "standing-stone.seek-red") {
		t.Fatalf("Standing Stone red clue message=%q choices=%v", state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.press-button-or-return-to-continue")); err != nil {
		t.Fatalf("continue Standing Stone red clue: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.journey-on")); err != nil {
		t.Fatalf("journey on from Standing Stone: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.essembra")); err != nil {
		t.Fatalf("travel to Essembra: %v", err)
	}
	if err := state.Select(requireGamePackOptionIndex(t, &state, "ecl-option.trail")); err != nil {
		t.Fatalf("take Essembra trail: %v", err)
	}
	if state.Location != LocationEssembra || state.Message != requireGamePackText(t, &state, "essembra.edge") {
		t.Fatalf("normal main-line Essembra arrival location=%v message=%q choices=%v",
			state.Location, state.Message, state.currentOriginalChoices)
	}
	return &state
}

func TestRealNewGameBeginsAtGlobalBlockOne(t *testing.T) {
	runNormalNewGameToEssembra(t)
}

func TestRealNewGameContinuesIntoHapByDungeonMovement(t *testing.T) {
	state := runNormalNewGameToEssembra(t)
	if state == nil {
		return
	}
	selectOption := func(id string) {
		t.Helper()
		if err := state.Select(requireGamePackOptionIndex(t, state, id)); err != nil {
			t.Fatalf("select %s: %v", id, err)
		}
	}
	selectOption("ecl-option.journey-on")
	selectOption("ecl-option.hap")
	selectOption("ecl-option.trail")
	selectOption("ecl-option.press-button-or-return-to-continue")
	for turn := 0; turn < 128 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatalf("Hap travel combat turn %d: %v", turn, err)
		}
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatalf("continue Hap travel combat: %v", err)
		}
	}
	selectOption("ecl-option.enter-city")
	if state.PictureRequested {
		if err := state.Continue(); err != nil {
			t.Fatalf("continue Hap entry picture: %v", err)
		}
	}
	selectOption("ecl-option.press-button-or-return-to-continue")
	if state.Mode != ModeDungeon || state.session == nil || state.session.CurrentBlockID() != 0x31 ||
		state.GeoMapSet != 5 || state.GeoMapBlock != 0x32 {
		t.Fatalf("Hap dungeon entry mode=%v block=%#x geo=%d/%#x message=%q",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock, state.Message)
	}
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer image.Close()
	geoCatalog := geo.NewCatalog()
	if err := geoCatalog.AddDAX(5, zipData(t, image, "GEO5.DAX")); err != nil {
		t.Fatal(err)
	}
	grid, ok := geoCatalog.Lookup(geo.MapRef{Set: 5, BlockID: 0x32})
	if !ok {
		t.Fatal("missing Hap GEO5 block 0x32")
	}

	resolveBoundary := func() {
		t.Helper()
		for attempt := 0; attempt < 16 && state.Mode != ModeDungeon; attempt++ {
			switch state.Mode {
			case ModeCombat:
				if err := state.CombatAct(); err != nil {
					t.Fatalf("Hap patrol combat: %v", err)
				}
			case ModeEvent:
				if state.PictureRequested {
					if err := state.Continue(); err != nil {
						t.Fatalf("continue Hap event picture: %v", err)
					}
					continue
				}
				if len(state.Choices) > 0 {
					if err := state.Select(0); err != nil {
						t.Fatalf("continue Hap event menu: %v", err)
					}
				} else if err := state.Continue(); err != nil {
					t.Fatalf("continue Hap event: %v", err)
				}
			case ModeWilderness:
				if index, found := findGamePackOptionIndex(state, "option.flee"); found {
					if err := state.Select(index); err != nil {
						t.Fatalf("flee Hap encounter menu: %v", err)
					}
					continue
				}
				if index, found := findGamePackOptionIndex(state, "ecl-option.press-button-or-return-to-continue"); found {
					if err := state.Select(index); err != nil {
						t.Fatalf("continue Hap wilderness event: %v", err)
					}
					continue
				}
				t.Fatalf("unexpected Hap wilderness boundary choices=%v message=%q block=%#x pos=(%d,%d,%d) eclReturn=%v eventReturn=%v",
					state.currentOriginalChoices, state.Message, state.session.CurrentBlockID(),
					state.DungeonX, state.DungeonY, state.DungeonDirection,
					state.eclMenuReturnMode, state.eventReturnMode)
			default:
				t.Fatalf("unexpected Hap boundary mode=%v", state.Mode)
			}
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("Hap boundary did not return to dungeon: mode=%v block=%#x choices=%v message=%q",
				state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices, state.Message)
		}
	}

	for index, step := range [][3]int{
		{0, 1, 4}, {0, 1, 4}, {-1, 0, 6}, {0, 1, 4}, {0, 1, 4},
		{0, 1, 4}, {1, 0, 2}, {0, 1, 4}, {0, 1, 4}, {0, 1, 4},
		{0, 1, 4}, {0, 1, 4}, {0, 1, 4}, {0, 1, 4}, {0, 1, 4},
		{1, 0, 2},
	} {
		if state.Mode != ModeDungeon {
			t.Fatalf("Hap route before step %d mode=%v choices=%v message=%q",
				index, state.Mode, state.currentOriginalChoices, state.Message)
		}
		if err := state.MoveDungeon(grid, step[0], step[1], step[2]); err != nil {
			t.Fatalf("Hap route step %d from (%d,%d): %v", index,
				state.DungeonX, state.DungeonY, err)
		}
		resolveBoundary()
	}
	if state.DungeonX != 4 || state.DungeonY != 13 || state.Mode != ModeDungeon {
		t.Fatalf("Hap route endpoint mode=%v position=(%d,%d,%d) message=%q",
			state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection, state.Message)
	}
}

func TestRealFireKnifeBladeBarrierBranches(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, err := dax.Parse(zipData(t, image, "ECL2.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var hideout []byte
	for _, block := range blocks {
		if block.Entry.ID == 4 {
			hideout = block.Data
			break
		}
	}
	if len(hideout) == 0 {
		t.Fatal("missing ECL2 block 4")
	}
	run := func(selections []uint16) ecl.RunResult {
		session, sessionErr := ecl.NewBlockSession(map[uint8][]byte{4: hideout}, 4)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04F, 0x99)
		result, runErr := session.RunEntrySeedWithPartyContext(
			1, 500, selections, nil, 1, ecl.PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		return result
	}
	prompt := run(nil)
	if !prompt.WaitingForMenu || len(prompt.Menus) != 1 ||
		len(prompt.Menus[0].Options) != 3 ||
		prompt.Menus[0].Options[0] != "ENTER THE BLADES" ||
		prompt.Menus[0].Options[1] != "WAIT" ||
		!strings.Contains(strings.Join(prompt.Text, " "), "CLOUD OF BLADES") {
		t.Fatalf("blade barrier prompt=%+v", prompt)
	}
	wait := run([]uint16{1})
	joined := strings.Join(wait.Text, " ")
	if !wait.WaitingForMenu || len(wait.Menus) != 2 ||
		len(wait.Menus[1].Options) != 1 ||
		wait.Menus[1].Options[0] != "PRESS BUTTON OR RETURN TO CONTINUE." ||
		len(wait.DamageRequests) != 0 ||
		!strings.Contains(joined, "BLADES SLOW DOWN") ||
		!strings.Contains(joined, "FADE AWAY") {
		t.Fatalf("blade barrier wait branch=%+v", wait)
	}
	localizer := NewState(testCatalog())
	if got := localizer.localizeECLText(prompt.Text); got != requireGamePackText(t, &localizer, "fire-knife.blade-barrier") {
		t.Fatalf("localized blade barrier=%q", got)
	}
	if got := localizer.localizeECLText(wait.Text); got != requireGamePackText(t, &localizer, "fire-knife.blade-barrier-fades") {
		t.Fatalf("localized blade barrier aftermath=%q", got)
	}
	for _, source := range []string{"ENTER THE BLADES", "WAIT", "RETREAT"} {
		want, ok := localizer.dataPack.LocalizeOption(source, localizer.catalog.Language)
		if !ok || localizer.localizeOption(source) != want {
			t.Fatalf("blade barrier option %q is not game-pack driven", source)
		}
	}
	enter := run([]uint16{0, 0})
	if len(enter.DamageRequests) != 1 ||
		enter.DamageRequests[0] != (ecl.DamageRequest{
			Flags: 0xE0, DiceCount: 8, DiceSize: 8, Bonus: 0, SaveFlags: 0,
		}) ||
		!strings.Contains(strings.Join(enter.Text, " "), "BLADES TEAR INTO YOU") {
		t.Fatalf("blade barrier enter branch=%+v", enter)
	}

	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{4: hideout}, 4)
	if err := state.session.Reset(4); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.DungeonWallRoof = 0x99
	state.partyRoster = party.Roster{
		{ID: "one", Name: "一", HitPoints: 100, MaxHitPoints: 100},
		{ID: "two", Name: "二", HitPoints: 100, MaxHitPoints: 100},
	}
	state.party = []combat.Fighter{
		{ID: "one", HitPoints: 100, MaxHitPoints: 100},
		{ID: "two", HitPoints: 100, MaxHitPoints: 100},
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 3 || state.Choices[0] != "闖入刀刃" ||
		state.Message != requireGamePackText(t, &state, "fire-knife.blade-barrier") {
		t.Fatalf("playable blade barrier prompt choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].HitPoints != 62 || state.partyRoster[1].HitPoints != 62 ||
		state.party[0].HitPoints != 62 || state.party[1].HitPoints != 62 ||
		state.Message != requireGamePackText(t, &state, "fire-knife.blade-barrier-fades") {
		t.Fatalf("playable blade damage roster=%+v fighters=%+v message=%q choices=%v",
			state.partyRoster, state.party, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || len(state.Choices) != 0 {
		t.Fatalf("blade barrier return mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	state.DungeonWallRoof = 0x99
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || len(state.Choices) != 0 || state.Message != "" {
		t.Fatalf("blade barrier revisit mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
}

func TestRealFireKnifeFrozenRoomBranches(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, err := dax.Parse(zipData(t, image, "ECL2.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var hideout []byte
	for _, block := range blocks {
		if block.Entry.ID == 4 {
			hideout = block.Data
			break
		}
	}
	run := func(selections []uint16) ecl.RunResult {
		session, sessionErr := ecl.NewBlockSession(map[uint8][]byte{4: hideout}, 4)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04F, 0x9A)
		result, runErr := session.RunEntrySeedWithPartyContext(
			1, 500, selections, nil, 1, ecl.PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		return result
	}
	prompt := run(nil)
	if !prompt.WaitingForMenu || len(prompt.Menus) != 1 ||
		strings.Join(prompt.Menus[0].Options, "/") != "RETREAT/INTERROGATE/KILL" ||
		!strings.Contains(strings.Join(prompt.Text, " "), "PEOPLE FROZEN IN") {
		t.Fatalf("frozen room prompt=%+v", prompt)
	}
	retreat := run([]uint16{0})
	if !retreat.Exited || retreat.WaitingForMenu || len(retreat.DamageRequests) != 0 {
		t.Fatalf("frozen room retreat=%+v", retreat)
	}
	interrogate := run([]uint16{1})
	interrogateText := strings.Join(interrogate.Text, " ")
	if !interrogate.WaitingForMenu || len(interrogate.Menus) != 2 ||
		!strings.Contains(interrogateText, "YOU DISARM THE FIRE KNIVES") ||
		!strings.Contains(interrogateText, "JOURNAL ENTRY 26") {
		t.Fatalf("frozen room interrogate=%+v", interrogate)
	}
	kill := run([]uint16{2})
	if !kill.WaitingForMenu ||
		!strings.Contains(strings.Join(kill.Text, " "), "YOU SLAUGHTER THEM") {
		t.Fatalf("frozen room kill=%+v", kill)
	}
	localizer := NewState(testCatalog())
	for _, source := range []string{"RETREAT", "INTERROGATE", "KILL"} {
		want, ok := localizer.dataPack.LocalizeOption(source, localizer.catalog.Language)
		if !ok || localizer.localizeOption(source) != want {
			t.Fatalf("frozen-room option %q is not game-pack driven", source)
		}
	}

	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{4: hideout}, 4)
	if err := state.session.Reset(4); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.DungeonWallRoof = 0x9A
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 3 || state.Choices[1] != "審問" ||
		state.Message != requireGamePackText(t, &state, "fire-knife.frozen-room") {
		t.Fatalf("playable frozen room choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "journal-trigger.frozen-room-26") {
		t.Fatalf("playable frozen interrogation message=%q", state.Message)
	}
	foundJournal := slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.26"))
	if !foundJournal {
		t.Fatalf("Journal Entry 26 was not unlocked: %v", state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("frozen interrogation return mode=%v message=%q", state.Mode, state.Message)
	}
	state.DungeonWallRoof = 0x9A
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.Message != "" || len(state.Choices) != 0 {
		t.Fatalf("frozen room revisit mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.Choices)
	}
}

func TestRealFireKnifeOfficeStages(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, err := dax.Parse(zipData(t, image, "ECL2.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var hideout []byte
	for _, block := range blocks {
		if block.Entry.ID == 4 {
			hideout = block.Data
			break
		}
	}
	run := func(memory map[uint16]uint16, selections []uint16) ecl.RunResult {
		session, sessionErr := ecl.NewBlockSession(map[uint8][]byte{4: hideout}, 4)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04F, 0x9B)
		for address, value := range memory {
			session.SetMemoryValue(address, value)
		}
		result, runErr := session.RunEntrySeedWithPartyContext(
			1, 500, selections, nil, 1, ecl.PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		return result
	}
	fresh := run(nil, nil)
	if !fresh.WaitingForMenu || len(fresh.Menus) != 1 ||
		!strings.Contains(strings.Join(fresh.Text, " "), "ORNATE ROOM") {
		t.Fatalf("fresh Fire Knife office=%+v", fresh)
	}
	ordinaryRevisit := run(map[uint16]uint16{0x4C10: 1}, nil)
	if !ordinaryRevisit.Exited || len(ordinaryRevisit.Text) != 0 {
		t.Fatalf("ordinary office revisit=%+v", ordinaryRevisit)
	}
	search := run(map[uint16]uint16{0x4C10: 1, 0x7ECA: 1}, nil)
	searchText := strings.Join(search.Text, " ")
	if !search.WaitingForMenu || !strings.Contains(searchText, "ROSEWOOD DESK") ||
		!strings.Contains(searchText, "JOURNAL ENTRY 9") {
		t.Fatalf("office search=%+v", search)
	}
	loot := run(map[uint16]uint16{0x4C10: 1, 0x7ECA: 1}, []uint16{0})
	if !loot.CombatRequested || len(loot.TreasureRequests) != 1 ||
		loot.TreasureRequests[0] != (ecl.TreasureRequest{
			Coins: [7]uint16{0, 0, 0, 500, 500, 3, 2}, ItemBlock: 0x82,
		}) {
		t.Fatalf("office loot=%+v", loot)
	}
	consumed := run(map[uint16]uint16{0x4C10: 2, 0x4CFE: 0x80, 0x7ECA: 1}, nil)
	if !consumed.Exited || len(consumed.Text) != 0 || len(consumed.TreasureRequests) != 0 {
		t.Fatalf("consumed office search=%+v", consumed)
	}

	state := NewStateFromECLBlocks(combatVisualCatalog(t), map[uint8][]byte{4: hideout}, 4)
	if err := state.session.Reset(4); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.DungeonX, state.DungeonY, state.DungeonDirection = 14, 11, 0
	state.DungeonWallRoof = 0x9B
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "fire-knife.office") {
		t.Fatalf("playable office intro=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("office intro return mode=%v", state.Mode)
	}
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "花梨木書桌") ||
		!strings.Contains(state.Message, "手札第 9 條") {
		t.Fatalf("localized office search=%q", state.Message)
	}
	foundJournal := slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.9"))
	if !foundJournal {
		t.Fatalf("Journal Entry 9 was not unlocked: %v", state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !state.treasureMenu ||
		len(state.PendingTreasureItems()) != 2 || len(state.Choices) != 3 {
		t.Fatalf("office treasure mode=%v menu=%v items=%v choices=%v",
			state.Mode, state.treasureMenu, state.PendingTreasureItems(), state.Choices)
	}
	if state.Prompt != state.catalog.Text("treasure_prompt", "") ||
		state.Choices[len(state.Choices)-1] != state.catalog.Text("treasure_exit", "") {
		t.Fatalf("office treasure locale prompt=%q choices=%#v", state.Prompt, state.Choices)
	}
	if state.MoneyPool() != 3000 {
		t.Fatalf("office pooled gold=%d, want 3000", state.MoneyPool())
	}
	gems, jewelry := state.TreasurePool()
	if gems != 3 || jewelry != 2 {
		t.Fatalf("office treasure pool gems/jewelry=%d/%d", gems, jewelry)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("office treasure exit mode=%v message=%q", state.Mode, state.Message)
	}
	beforeItems := len(state.PendingTreasureItems())
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || len(state.PendingTreasureItems()) != beforeItems ||
		state.MoneyPool() != 3000 {
		t.Fatalf("office consumed search mode=%v items=%d/%d money=%d",
			state.Mode, len(state.PendingTreasureItems()), beforeItems, state.MoneyPool())
	}
}

func TestRealFireKnifeAshenRooms(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, err := dax.Parse(zipData(t, image, "ECL2.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var hideout []byte
	for _, block := range blocks {
		if block.Entry.ID == 4 {
			hideout = block.Data
			break
		}
	}
	run := func(terrain uint16, selections []uint16) ecl.RunResult {
		session, sessionErr := ecl.NewBlockSession(map[uint8][]byte{4: hideout}, 4)
		if sessionErr != nil {
			t.Fatal(sessionErr)
		}
		session.SetMemoryValue(0xC04F, terrain)
		result, runErr := session.RunEntrySeedWithPartyContext(
			1, 500, selections, nil, 1, ecl.PartyContext{},
		)
		if runErr != nil {
			t.Fatal(runErr)
		}
		return result
	}
	cases := []struct {
		terrain       uint16
		flag          uint16
		raw           string
		messageID     string
		continuations []uint16
	}{
		{0x9C, 0x4C11, "STRANGE SMOKY SCENT", "fire-knife.smoky-hall", []uint16{0}},
		{0x9D, 0x4C12, "UNSEEN SERVANTS", "fire-knife.ordered-bedroom", []uint16{0}},
		{0x9E, 0x4C13, "CHARRED BODY", "fire-knife.burned-library", []uint16{0, 0}},
		{0x9F, 0x4C14, "NOTHING ESCAPED DESTRUCTION", "fire-knife.burned-lab", []uint16{0}},
		{0xA0, 0x4C15, "TWO ROWS OF SHROUDED BODIES", "fire-knife.shrouded-bodies", []uint16{0}},
	}
	for _, test := range cases {
		first := run(test.terrain, nil)
		if !first.WaitingForMenu ||
			!strings.Contains(strings.Join(first.Text, " "), test.raw) {
			t.Fatalf("ashen terrain %#x first=%+v", test.terrain, first)
		}
		completed := run(test.terrain, test.continuations)
		if !completed.Exited || completed.WaitingForMenu {
			t.Fatalf("ashen terrain %#x completed=%+v", test.terrain, completed)
		}

		state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{4: hideout}, 4)
		if err := state.session.Reset(4); err != nil {
			t.Fatal(err)
		}
		state.Mode = ModeDungeon
		state.DungeonWallRoof = uint8(test.terrain)
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if state.Message != requireGamePackText(t, &state, test.messageID) {
			t.Fatalf("localized terrain %#x message=%q", test.terrain, state.Message)
		}
		for range test.continuations {
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("terrain %#x return mode=%v message=%q",
				test.terrain, state.Mode, state.Message)
		}
		if value, ok := state.session.MemoryValue(test.flag); !ok || value != 1 {
			t.Fatalf("terrain %#x visited flag %#x=%d,%v",
				test.terrain, test.flag, value, ok)
		}
		state.DungeonWallRoof = uint8(test.terrain)
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatal(err)
		}
		if state.Mode != ModeDungeon || len(state.Choices) != 0 {
			t.Fatalf("terrain %#x revisit mode=%v choices=%v message=%q",
				test.terrain, state.Mode, state.Choices, state.Message)
		}
		if test.terrain == 0x9E {
			foundJournal := slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.29"))
			if !foundJournal {
				t.Fatalf("Journal Entry 29 was not unlocked: %v", state.JournalPages)
			}
		}
	}
}

func TestRealFireKnifeLeaderEncounterAndBondProgression(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	allBlocks := make(map[uint8][]byte)
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
		blocks, parseErr := dax.Parse(zipData(t, image, member))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			allBlocks[block.Entry.ID] = block.Data
		}
	}
	session, err := ecl.NewBlockSession(allBlocks, 4)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04F, 0x87)
	beforeCombat, err := session.RunEntrySeedWithPartyContext(1, 1000, []uint16{0}, nil, 1, ecl.PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !beforeCombat.CombatRequested || beforeCombat.PictureBlock != 12 ||
		len(beforeCombat.MonsterSpawns) != 2 ||
		beforeCombat.MonsterSpawns[0] != (ecl.MonsterSpawn{MonsterID: 1, Count: 20, IconBlock: 1}) ||
		beforeCombat.MonsterSpawns[1] != (ecl.MonsterSpawn{MonsterID: 3, Count: 1, IconBlock: 3}) {
		t.Fatalf("leader encounter=%+v", beforeCombat)
	}
	if len(beforeCombat.TreasureRequests) != 1 ||
		beforeCombat.TreasureRequests[0] != (ecl.TreasureRequest{
			Coins: [7]uint16{0, 0, 0, 2000, 3000, 8, 4}, ItemBlock: 0x82,
		}) {
		t.Fatalf("leader treasure=%+v", beforeCombat.TreasureRequests)
	}
	session.SetMemoryValue(0x7EC7, 0)
	postSelections := []uint16{0}
	var continuation []ecl.RunResult
	for step := 0; step < 5; step++ {
		postSelections = append(postSelections, 0)
		continued, continueErr := session.RunInteractiveSeed(1000, postSelections, 1)
		if continueErr != nil {
			t.Fatal(continueErr)
		}
		continuation = append(continuation, continued)
	}
	if !strings.Contains(strings.Join(continuation[0].Text, " "), "JOURNAL ENTRY 54") ||
		!strings.Contains(strings.Join(continuation[0].Text, " "), "JOURNAL ENTRY 53") ||
		continuation[0].PictureBlock != 13 {
		t.Fatalf("leader victory continuation=%+v", continuation[0])
	}
	for address, want := range map[uint16]uint16{0x4CFF: 1, 0x4C2A: 1} {
		if got, ok := session.MemoryValue(address); !ok || got != want {
			t.Fatalf("memory[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}
	if !strings.Contains(strings.Join(continuation[1].Text, " "), "FIRST NIGHT OUTSIDE THE CITY") ||
		!continuation[2].BigPictureRequested || continuation[2].PictureBlock != 120 ||
		!strings.Contains(strings.Join(continuation[4].Text, " "), "COLD SWEAT") {
		t.Fatalf("bond dream continuation=%+v", continuation)
	}
	postSelections = append(postSelections, 0)
	chapter, err := session.RunInteractiveSeed(1000, postSelections, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := session.MemoryValue(0x7F12); !ok || got != 1 {
		t.Fatalf("bond progression memory=%#x,%v want 1", got, ok)
	}
	if session.CurrentBlockID() != 0x50 {
		t.Fatalf("post-Fire-Knife block=%#x result=%+v, want ECL1 block 0x50", session.CurrentBlockID(), chapter)
	}
}

func TestFireKnifeLeaderStateVictoryReturnsToTilverton(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	allBlocks := make(map[uint8][]byte)
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
		blocks, parseErr := dax.Parse(zipData(t, image, member))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			allBlocks[block.Entry.ID] = block.Data
		}
	}
	session, err := ecl.NewBlockSession(allBlocks, 4)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04F, 0x87)
	encounter, err := session.RunEntrySeedWithPartyContext(1, 1000, []uint16{0}, nil, 1, ecl.PartyContext{})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(trainingTestCatalog(t))
	state.session = session
	state.eclBlock = session.CurrentData()
	state.eclStart, err = session.InitialEntry()
	if err != nil {
		t.Fatal(err)
	}
	state.selectionSequence = []uint16{0}
	state.eclMenuReturnMode = ModeDungeon
	state.eventReturnMode = ModeDungeon
	state.partyRoster = party.Roster{{
		ID: "hero", Name: "英雄", Race: party.RaceHuman, Class: party.ClassFighter,
		Level: 1, HitPoints: 10, MaxHitPoints: 10,
		Abilities: party.Abilities{
			Strength: 10, Intelligence: 10, Wisdom: 10,
			Dexterity: 10, Constitution: 10, Charisma: 10,
		},
	}}
	state.applyECLTreasureSignals(encounter)
	treasureItems, err := ParseTreasureItemBlocks(map[uint8][]byte{
		1: zipData(t, image, "ITEM1.DAX"),
		2: zipData(t, image, "ITEM2.DAX"),
		3: zipData(t, image, "ITEM3.DAX"),
		4: zipData(t, image, "ITEM4.DAX"),
		5: zipData(t, image, "ITEM5.DAX"),
		6: zipData(t, image, "ITEM6.DAX"),
	})
	if err != nil {
		t.Fatal(err)
	}
	state.SetTreasureItemBlocks(treasureItems)
	hero := combat.Fighter{
		ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
		DamageBonus: 100, AttacksPerTurn: 8, InitiativeBonus: 100,
	}
	state.party = []combat.Fighter{hero}
	monsterBlocks, err := dax.Parse(zipData(t, image, "MON1CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	monsterRecords := make(map[uint8]monster.Record, len(monsterBlocks))
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monsterRecords[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(1, monsterRecords)
	monster5Blocks, err := dax.Parse(zipData(t, image, "MON5CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	monster5Records := make(map[uint8]monster.Record, len(monster5Blocks))
	for _, block := range monster5Blocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monster5Records[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(5, monster5Records)
	monster3Blocks, err := dax.Parse(zipData(t, image, "MON3CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	monster3Records := make(map[uint8]monster.Record, len(monster3Blocks))
	for _, block := range monster3Blocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monster3Records[block.Entry.ID] = record
	}
	state.SetMonsterRecordsForECL(3, monster3Records)
	monster3Affects := make(map[uint8][]monster.AffectRecord)
	monster3AffectBlocks, err := dax.Parse(zipData(t, image, "MON3SPC.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range monster3AffectBlocks {
		parsed, parseErr := monster.ParseAffects(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monster3Affects[block.Entry.ID] = parsed
	}
	state.SetMonsterAffectsForECL(3, monster3Affects)
	monster3Items := make(map[uint8][]monster.ItemRecord)
	monster3ItemBlocks, err := dax.Parse(zipData(t, image, "MON3ITM.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range monster3ItemBlocks {
		parsed, parseErr := monster.ParseItems(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monster3Items[block.Entry.ID] = parsed
	}
	state.SetMonsterItemsForECL(3, monster3Items)
	monster5Affects := make(map[uint8][]monster.AffectRecord)
	monster5AffectBlocks, err := dax.Parse(zipData(t, image, "MON5SPC.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range monster5AffectBlocks {
		parsed, parseErr := monster.ParseAffects(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monster5Affects[block.Entry.ID] = parsed
	}
	state.SetMonsterAffectsForECL(5, monster5Affects)
	monster5Items := make(map[uint8][]monster.ItemRecord)
	monster5ItemBlocks, err := dax.Parse(zipData(t, image, "MON5ITM.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range monster5ItemBlocks {
		parsed, parseErr := monster.ParseItems(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		monster5Items[block.Entry.ID] = parsed
	}
	state.SetMonsterItemsForECL(5, monster5Items)
	if err := state.StartCombat([]combat.Fighter{hero}, []combat.Fighter{{
		ID: "leader", Name: "火刀首領", Side: combat.SideEnemy, HitPoints: 1, MaxHitPoints: 1,
	}}, 7); err != nil {
		t.Fatal(err)
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if !state.treasureMenu || state.MoneyPool() != 17000 || len(state.PendingTreasureItems()) != 2 {
		t.Fatalf("victory loot menu=%v money=%d items=%#v", state.treasureMenu, state.MoneyPool(), state.PendingTreasureItems())
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 14 || !strings.Contains(state.Message, "手札第 54") {
		t.Fatalf("Journal 54 state mode=%v picture=%d message=%q", state.Mode, state.PictureBlock, state.Message)
	}
	dreamMessages := make([]string, 0, 12)
	for step := 0; step < 6; step++ {
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		}
		if state.Message != "" {
			dreamMessages = append(dreamMessages, state.Message)
		}
		if state.Mode != ModeWilderness || len(state.Choices) == 0 {
			t.Fatalf("continuation step %d mode=%v picture=%v choices=%v message=%q",
				step, state.Mode, state.PictureRequested, state.currentOriginalChoices, state.Message)
		}
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if state.Message != "" {
			dreamMessages = append(dreamMessages, state.Message)
		}
	}
	for _, messageID := range []string{
		"bond-dream.first-night",
		"bond-dream.masters-taunt",
		"bond-dream.masters-prophecy",
		"bond-dream.ends",
	} {
		if !slices.Contains(dreamMessages, requireGamePackText(t, &state, messageID)) {
			t.Fatalf("%s was not displayed during bond dream: %q", messageID, dreamMessages)
		}
	}
	if !state.PictureRequested || !state.BigPictureRequested || state.PictureBlock != 121 {
		t.Fatalf("Tilverton edge picture=%v big=%v block=%d message=%q",
			state.PictureRequested, state.BigPictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) ||
		state.Message != requireGamePackText(t, &state, "tilverton.edge") {
		t.Fatalf("Tilverton edge mode=%v choices=%#v message=%q", state.Mode, state.currentOriginalChoices, state.Message)
	}
	savePath := filepath.Join(t.TempDir(), "fire-knife-world-edge.json")
	if err := state.SavePartyFile(savePath); err != nil {
		t.Fatalf("save after Fire Knife world-map return: %v", err)
	}
	loaded := NewStateFromECLBlocks(trainingTestCatalog(t), allBlocks, session.CurrentBlockID())
	if err := loaded.LoadPartyFile(savePath); err != nil {
		t.Fatalf("load after Fire Knife world-map return: %v", err)
	}
	if loaded.Mode != ModeWilderness || loaded.Location != LocationTilverton ||
		loaded.session == nil || loaded.session.CurrentBlockID() != session.CurrentBlockID() ||
		loaded.OriginalLocation != state.OriginalLocation ||
		!reflect.DeepEqual(loaded.currentOriginalChoices, state.currentOriginalChoices) ||
		!reflect.DeepEqual(loaded.Choices, state.Choices) {
		t.Fatalf("Fire Knife world-map save revisit mode=%v/%v location=%v/%v block=%#x/%#x choices=%#v/%#v localized=%#v",
			loaded.Mode, state.Mode, loaded.Location, state.Location, loaded.session.CurrentBlockID(), session.CurrentBlockID(),
			loaded.currentOriginalChoices, state.currentOriginalChoices, loaded.Choices)
	}
	for address, want := range map[uint16]uint16{0x4CFF: 1, 0x4C2A: 1, 0x7F12: 1} {
		if got, ok := session.MemoryValue(address); !ok || got != want {
			t.Fatalf("memory[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}
	for _, messageID := range []string{"journal.54", "journal.53.1", "journal.53.2"} {
		if !slices.Contains(state.JournalPages, requireGamePackText(t, &state, messageID)) {
			t.Fatalf("%s was not unlocked: %v", messageID, state.JournalPages)
		}
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "tilverton.entry-barred") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Tilverton barred choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) ||
		state.Message != requireGamePackText(t, &state, "tilverton.edge") {
		t.Fatalf("Tilverton barred return mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"SHADOWDALE", "ASHABENFORD", "DAGGER FALLS"}) {
		t.Fatalf("post-Fire-Knife journey mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("Ashabenford routes=%#v message=%q", state.currentOriginalChoices, state.Prompt)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "ashabenford.tilvers-gap") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Tilver's Gap event choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("Tilver's Gap encounter mode=%v event=%q message=%q", state.Mode, state.OriginalEvent, state.Message)
	}
	fighters := state.CombatFighters()
	if len(fighters) != 9 {
		t.Fatalf("Tilver's Gap fighters=%d, want hero plus eight hippogriffs", len(fighters))
	}
	wantHippogriff := requireCombatantName(t, &state, "HIPPOGRIFF")
	for _, fighter := range fighters[1:] {
		if fighter.Name != wantHippogriff || fighter.SpriteBlock != 81 || fighter.Side != combat.SideEnemy {
			t.Fatalf("Tilver's Gap enemy=%+v", fighter)
		}
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || state.Location != LocationAshabenford ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) ||
		state.Message != requireGamePackText(t, &state, "ashabenford.edge") {
		t.Fatalf("Ashabenford arrival mode=%v location=%v choices=%#v message=%q",
			state.Mode, state.Location, state.currentOriginalChoices, state.Message)
	}
	if len(state.PendingTreasureItems()) != 0 {
		t.Fatalf("skipped Fire Knife loot leaked into the hippogriff victory: %#v", state.PendingTreasureItems())
	}
	for address, want := range map[uint16]uint16{0x4C83: 1, 0x4C9B: 2, 0x4CA1: 2} {
		if got, ok := session.MemoryValue(address); !ok || got != want {
			t.Fatalf("Ashabenford memory[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}
	state.ConsumeMusicEvents()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 80 ||
		state.Message != requireGamePackText(t, &state, "ashabenford.places") {
		t.Fatalf("Ashabenford entry mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, []MusicEvent{{
		Action: "play", TrackID: "pc98-bgm-selector-06",
	}}) {
		t.Fatalf("Ashabenford enter music=%+v", got)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"INN", "STORE", "HALL", "TEMPLE", "BAR", "LEAVE"}) {
		t.Fatalf("Ashabenford places mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"HAVE A DRINK", "RELAX", "EXIT"}) ||
		state.Message != requireGamePackText(t, &state, "ashabenford.ale-house") {
		t.Fatalf("Ashabenford ale house choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != requireGamePackText(t, &state, "ashabenford.tavern-tale-28") {
		t.Fatalf("Ashabenford Tavern Tale 28 choices=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"HAVE A DRINK", "RELAX", "EXIT"}) {
		t.Fatalf("ale house did not resume after tale: %#v", state.currentOriginalChoices)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.PictureRequested {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"INN", "STORE", "HALL", "TEMPLE", "BAR", "LEAVE"}) {
		t.Fatalf("Ashabenford places did not resume: %#v", state.currentOriginalChoices)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || !state.BigPictureRequested || state.PictureBlock != 121 {
		t.Fatalf("Ashabenford leave picture=%v big=%v block=%d", state.PictureRequested, state.BigPictureRequested, state.PictureBlock)
	}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, []MusicEvent{{
		Action: "play", TrackID: "pc98-bgm-selector-05",
	}}) {
		t.Fatalf("Ashabenford leave music=%+v", got)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TILVERTON", "SHADOWDALE", "THE STANDING STONE"}) ||
		state.Prompt != "從這裡可以前往" {
		t.Fatalf("Ashabenford destinations=%#v prompt=%q", state.currentOriginalChoices, state.Prompt)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("Standing Stone routes=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "shadow-gap.fire-knives-patrol") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Shadow Gap ambush choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("Shadow Gap patrol did not enter combat: mode=%v message=%q", state.Mode, state.Message)
	}
	patrol := state.CombatFighters()
	if len(patrol) != 7 {
		t.Fatalf("Shadow Gap fighters=%d, want hero plus six disguised Fire Knives", len(patrol))
	}
	for _, fighter := range patrol[1:] {
		if fighter.Name != "戰士" || fighter.SpriteBlock != 0x20 || fighter.Side != combat.SideEnemy {
			t.Fatalf("Shadow Gap enemy=%+v", fighter)
		}
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Message != requireGamePackText(t, &state, "standing-stone.grey-man") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Standing Stone arrival choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"THANK HIM", "ATTACK", "LEAVE"}) ||
		state.Message != requireGamePackText(t, &state, "standing-stone.four-masters") {
		t.Fatalf("Standing Stone counsel choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "standing-stone.seek-red") ||
		state.Location != LocationStandingStone {
		t.Fatalf("Standing Stone hint location=%v message=%q", state.Location, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C9B); !ok || got != 4 {
		t.Fatalf("Standing Stone current-location memory=%#x,%v want 4", got, ok)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Standing Stone actions=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ASHABENFORD", "ESSEMBRA", "HILLSFAR", "MYTH DRANNOR"}) {
		t.Fatalf("Standing Stone destinations=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) ||
		state.Message != requireGamePackText(t, &state, "world-route.essembra") {
		t.Fatalf("Essembra routes=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Location != LocationEssembra || state.Message != requireGamePackText(t, &state, "essembra.edge") {
		t.Fatalf("Essembra arrival location=%v message=%q", state.Location, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C9B); !ok || got != 8 {
		t.Fatalf("Essembra current-location memory=%#x,%v want 8", got, ok)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"HAP", "THE STANDING STONE"}) {
		t.Fatalf("Essembra destinations=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) ||
		state.Message != requireGamePackText(t, &state, "world-route.hap") {
		t.Fatalf("Hap routes=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.black-dragons") {
		t.Fatalf("Hap dragon approach message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("Hap dragons did not enter combat: mode=%v message=%q", state.Mode, state.Message)
	}
	if got, ok := session.MemoryValue(0x4CA2); !ok || got != 1 {
		t.Fatalf("Hap dragon encounter flag=%#x,%v want 1 before combat", got, ok)
	}
	dragons := state.CombatFighters()
	if len(dragons) != 4 {
		t.Fatalf("Hap dragon fighters=%d, want hero plus three dragons", len(dragons))
	}
	wantBlackDragon := requireCombatantName(t, &state, "BLACK DRAGON")
	for _, fighter := range dragons[1:] {
		if fighter.Name != wantBlackDragon || fighter.SpriteBlock != 0x35 || fighter.Side != combat.SideEnemy {
			t.Fatalf("Hap dragon=%+v", fighter)
		}
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Location != LocationHap || state.Message != requireGamePackText(t, &state, "hap.edge") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Hap arrival location=%v choices=%#v message=%q",
			state.Location, state.currentOriginalChoices, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C9B); !ok || got != 9 {
		t.Fatalf("Hap current-location memory=%#x,%v want 9", got, ok)
	}
	if got, ok := session.MemoryValue(0x4CA1); !ok || got != 9 {
		t.Fatalf("Hap previous-location mirror=%#x,%v want 9", got, ok)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x31 || state.Area.GameArea != 5 || !state.Area.InDungeon ||
		state.GeoMapSet != 5 || state.Message != requireGamePackText(t, &state, "hap.abandoned-village") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap entry block=%#x area=%+v geo=%d choices=%#v message=%q",
			session.CurrentBlockID(), state.Area, state.GeoMapSet, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.Prompt != "哈普村　↑：前進　K／M：轉向　S：搜尋　L：查看　E：紮營" ||
		state.LoadPieces != [3]uint16{12, 0xFF, 0xFF} {
		t.Fatalf("Hap dungeon mode=%v prompt=%q pieces=%v", state.Mode, state.Prompt, state.LoadPieces)
	}
	state.DungeonWallRoof = 0x84
	session.SetMemoryValue(0x4BC9, 15)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 50 ||
		state.Message != requireGamePackText(t, &state, "hap.hiding-peasants") {
		t.Fatalf("Hap peasants mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C02); !ok || got != 1 {
		t.Fatalf("Hap peasant visited flag=%#x,%v want 1", got, ok)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"LEAVE", "TRY TO TALK FURTHER"}) {
		t.Fatalf("Hap peasants choices=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.peasants-flee") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap talk choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Hap peasant event returned mode=%v, want dungeon", state.Mode)
	}
	state.DungeonWallRoof = 0x80
	state.SetECLSeed(3)
	session.SetMemoryValue(0x4BC9, 15)
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}) ||
		state.Prompt != "一隊黑暗精靈巡邏兵出現了" {
		t.Fatalf("Hap patrol choices=%#v prompt=%q", state.currentOriginalChoices, state.Prompt)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || state.Message != requireGamePackText(t, &state, "hap.dark-elf-attack") {
		t.Fatalf("Hap patrol combat mode=%v message=%q", state.Mode, state.Message)
	}
	patrolFighters := state.CombatFighters()
	if len(patrolFighters) != 5 {
		t.Fatalf("Hap patrol fighters=%d, want hero and four fighters", len(patrolFighters))
	}
	// The menu resume must continue the same seeded PRNG stream. Restarting
	// seed 3 at the COMBAT choice used to fabricate three fighters plus one
	// mage; the uninterrupted ECL sequence produces four fighters here.
	for _, fighter := range patrolFighters[1:] {
		if fighter.Name != "黑暗精靈戰士" || fighter.SpriteBlock != 0x31 {
			t.Fatalf("Hap dark elf fighter=%+v", fighter)
		}
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeEvent || state.Message != "戰鬥勝利！" {
		t.Fatalf("Hap patrol victory mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Hap patrol continuation mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C47); !ok || got != 1 {
		t.Fatalf("Hap defeated patrol count=%#x,%v want 1", got, ok)
	}
	state.DungeonWallRoof = 0x8A
	state.DungeonWallType = 0
	session.SetMemoryValue(0x4BC9, 15)
	session.SetMemoryValue(0x4C02, 0)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 59 ||
		state.Message != requireGamePackText(t, &state, "hap.akabar-join") {
		t.Fatalf("Hap Akabar picture=%v/%d message=%q", state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("Hap Akabar choices=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster) != 2 || state.partyRoster[1].Name != "阿卡巴・貝爾・阿卡什" ||
		state.partyRoster[1].ScriptName != "AKABAR BEL AKAS" ||
		state.partyRoster[1].Class != party.ClassMagicUser || state.partyRoster[1].Level != 5 ||
		len(state.partyRoster[1].Equipment) != 2 || len(state.partyRoster[1].KnownSpells) != 11 {
		t.Fatalf("Hap Akabar roster=%+v", state.partyRoster)
	}
	if !state.PictureRequested || state.PictureBlock != 50 ||
		state.Message != requireGamePackText(t, &state, "hap.inn-before-liberation") {
		t.Fatalf("Hap inn picture=%v/%d message=%q", state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("Hap inn choices=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Hap inn exit mode=%v, want dungeon", state.Mode)
	}
	for address, want := range map[uint16]uint16{0x4C5F: 1, 0x7F7A: 1} {
		if got, ok := session.MemoryValue(address); !ok || got != want {
			t.Fatalf("Hap Akabar flag[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}

	state.party[0].HitPoints, state.party[0].MaxHitPoints = 500, 500
	state.partyRoster[0].HitPoints, state.partyRoster[0].MaxHitPoints = 500, 500
	state.DungeonWallRoof = 0x88
	session.SetMemoryValue(0x4BC9, 15)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.efreet-barn") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap efreet barn choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		state.Message != requireGamePackText(t, &state, "hap.efreet-threat") {
		t.Fatalf("Hap efreet approach choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat {
		t.Fatalf("Hap efreet combat mode=%v event=%q message=%q", state.Mode, state.OriginalEvent, state.Message)
	}
	efreetFighters := state.CombatFighters()
	if len(efreetFighters) != 15 {
		t.Fatalf("Hap efreet fighters=%d, want two party members plus 13 enemies", len(efreetFighters))
	}
	efreetCount, mageCount, clericCount := 0, 0, 0
	for _, fighter := range efreetFighters {
		if fighter.Side == combat.SideParty {
			continue
		}
		switch {
		case fighter.Name == "伊弗利特" && fighter.SpriteBlock == 0x34:
			efreetCount++
		case fighter.Name == "黑暗精靈法師" && fighter.SpriteBlock == 0x32:
			mageCount++
		case fighter.Name == "黑暗精靈牧師" && fighter.SpriteBlock == 0x33:
			clericCount++
		default:
			t.Fatalf("unexpected Hap efreet cohort=%+v", fighter)
		}
	}
	if efreetCount != 1 || mageCount != 6 || clericCount != 6 {
		t.Fatalf("Hap efreet cohort counts efreet=%d mage=%d cleric=%d", efreetCount, mageCount, clericCount)
	}
	for turn := 0; turn < 32 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Message != requireGamePackText(t, &state, "hap.efreet-map") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap efreet map choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	for address, want := range map[uint16]uint16{0x4C01: 5, 0x4C5E: 1} {
		if got, ok := session.MemoryValue(address); !ok || got != want {
			t.Fatalf("Hap efreet flag[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 50 ||
		state.Message != requireGamePackText(t, &state, "hap.liberated-crowd") {
		t.Fatalf("Hap liberation picture=%v/%d message=%q", state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.elder-thanks") {
		t.Fatalf("Hap elder thanks=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.elder-wizard-tower") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap elder tower choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.akabar-secret-routes") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap Akabar secret route choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Hap liberation returned mode=%v, want dungeon", state.Mode)
	}
	state.DungeonDirection = 2
	state.DungeonWallType = 7
	state.DungeonWallRoof = 2
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.leave") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("Hap exit choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hap.map-route") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"CAVES", "WILDERNESS"}) {
		t.Fatalf("Hap map route choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x32 || state.GeoMapSet != 5 || state.GeoMapBlock != 0x32 ||
		state.DungeonX != 15 || state.DungeonY != 5 || state.DungeonDirection != 6 ||
		state.LoadPieces != [3]uint16{8, 0xFF, 0xFF} ||
		state.Message != requireGamePackText(t, &state, "lava-tube.entry") {
		t.Fatalf("lava tube entry mode=%v block=%#x script=(%d,%d,%d) geo=%d/%d pieces=%v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.GeoMapSet, state.GeoMapBlock, state.LoadPieces, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "lava-tube.ambush") {
		t.Fatalf("lava tube ambush message=%q", state.Message)
	}
	if state.Mode != ModeCombat {
		t.Fatalf("lava tube ambush mode=%v, want combat", state.Mode)
	}
	salamanders, darkElves := 0, 0
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideParty {
			continue
		}
		switch {
		case fighter.Name == "火蜥蜴" && fighter.SpriteBlock == 0x39:
			salamanders++
		case fighter.Name == "黑暗精靈戰士" && fighter.SpriteBlock == 0x31:
			darkElves++
		default:
			t.Fatalf("unexpected lava tube enemy=%+v", fighter)
		}
	}
	if salamanders != 4 || darkElves != 3 {
		t.Fatalf("lava tube enemies salamanders=%d dark-elves=%d", salamanders, darkElves)
	}
	for turn := 0; turn < 32 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeEvent || state.Message != "戰鬥勝利！" {
		t.Fatalf("lava tube battle result mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.Message != "" {
		t.Fatalf("lava tube battle continuation mode=%v message=%q", state.Mode, state.Message)
	}
	lavaCatalog := geo.NewCatalog()
	if err := lavaCatalog.AddDAX(5, zipData(t, image, "GEO5.DAX")); err != nil {
		t.Fatal(err)
	}
	lavaGrid, ok := lavaCatalog.Lookup(geo.MapRef{Set: 5, BlockID: 0x32})
	if !ok {
		t.Fatal("GEO5 block 0x32 is unavailable")
	}
	towerGrid, ok := lavaCatalog.Lookup(geo.MapRef{Set: 5, BlockID: 0x33})
	if !ok {
		t.Fatal("GEO5 block 0x33 is unavailable")
	}
	if center, east := lavaGrid.CellWrapped(15, 5).Terrain, lavaGrid.CellWrapped(0, 5).Terrain; center != 0x80 || east != 0x89 {
		t.Fatalf("lava entrance terrains center=%#x east=%#x, want 0x80/0x89", center, east)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 9, 10, 2
	state.DungeonWallType, _ = lavaGrid.WallWrapped(9, 10, 2)
	state.DungeonWallRoof = lavaGrid.CellWrapped(9, 10).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || state.Message != requireGamePackText(t, &state, "lava-tube.guarded-door") {
		flag, _ := session.MemoryValue(0x4C48)
		boss, _ := session.MemoryValue(0x4C60)
		t.Fatalf("lava guarded door mode=%v message=%q flags=%#x boss=%#x picture=%v/%d originals=%#v",
			state.Mode, state.Message, flag, boss, state.PictureRequested, state.PictureBlock, state.currentOriginalChoices)
	}
	doorSalamanders, doorFighters, doorClerics := 0, 0, 0
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideParty {
			continue
		}
		switch fighter.Name {
		case "火蜥蜴":
			doorSalamanders++
		case "黑暗精靈戰士":
			doorFighters++
		case "黑暗精靈牧師":
			doorClerics++
		default:
			t.Fatalf("unexpected guarded-door enemy=%+v", fighter)
		}
	}
	if doorSalamanders != 3 || doorFighters != 3 || doorClerics != 1 {
		t.Fatalf("guarded-door enemies salamanders=%d fighters=%d clerics=%d",
			doorSalamanders, doorFighters, doorClerics)
	}
	for turn := 0; turn < 32 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "lava-tube.dream-warning") {
		t.Fatalf("guarded-door warning mode=%v message=%q", state.Mode, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C48); !ok || got&0x08 == 0 {
		t.Fatalf("guarded-door flag=%#x,%v want bit 0x08", got, ok)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("guarded-door continuation mode=%v, want dungeon", state.Mode)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 0, 5, 0
	state.DungeonWallType, _ = lavaGrid.WallWrapped(0, 5, 0)
	state.DungeonWallRoof = lavaGrid.CellWrapped(0, 5).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 57 ||
		state.Message != requireGamePackText(t, &state, "lava-tube.salamander-pools") {
		t.Fatalf("lava pools picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "PARLAY"}) {
		t.Fatalf("lava pools encounter choices=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices,
		[]string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK", "PARLAY_NICE", "PARLAY_ABUSIVE"}) {
		t.Fatalf("lava pools WAIT parlay choices=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "lava-tube.nice-parlay") {
		t.Fatalf("lava pools nice parlay mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("lava pools nice parlay continuation mode=%v, want dungeon", state.Mode)
	}
	if got, _ := session.MemoryValue(0x4C48); got&0x01 != 0 {
		t.Fatalf("lava pools parlay prematurely set completion flag=%#x", got)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || state.Message != requireGamePackText(t, &state, "lava-tube.intense-heat") {
		t.Fatalf("lava pools combat mode=%v message=%q", state.Mode, state.Message)
	}
	poolSalamanders := 0
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideEnemy {
			if fighter.Name != "火蜥蜴" || fighter.SpriteBlock != 0x39 {
				t.Fatalf("unexpected lava pools enemy=%+v", fighter)
			}
			poolSalamanders++
		}
	}
	if poolSalamanders != 15 {
		t.Fatalf("lava pools salamanders=%d, want 15", poolSalamanders)
	}
	for turn := 0; turn < 48 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, &state, "lava-tube.fireproof-casks") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("lava pools casks mode=%v message=%q choices=%#v",
			state.Mode, state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Prompt != "請選擇角色" || len(state.Choices) != 2 {
		t.Fatalf("lava cask volunteer prompt=%q choices=%#v", state.Prompt, state.Choices)
	}
	// This assertion covers the failed heat check. Reset explicitly instead
	// of relying on the removed per-invocation RNG restart.
	state.SetECLSeed(3)
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "lava-tube.cask-heat-retreat") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("lava cask heat mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("lava cask decline mode=%v, want dungeon", state.Mode)
	}
	if got, ok := session.MemoryValue(0x4C48); !ok || got&0x01 == 0 {
		t.Fatalf("lava pools flag=%#x,%v want bit 0x01", got, ok)
	}
	for _, page := range state.JournalPages {
		if strings.HasPrefix(page, "手札條目 15") {
			t.Fatalf("wizard-tower journal unlocked before event: %q", page)
		}
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 6, 15, 6
	state.DungeonWallType, _ = lavaGrid.WallWrapped(6, 15, 6)
	state.DungeonWallRoof = lavaGrid.CellWrapped(6, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x33 || state.GeoMapSet != 5 || state.GeoMapBlock != 0x33 ||
		state.DungeonX != 7 || state.DungeonY != 15 || state.DungeonDirection != 6 ||
		state.LoadPieces != [3]uint16{0x0E, 0x0F, 0xFF} ||
		!state.PictureRequested || state.PictureBlock != 0x33 ||
		!strings.Contains(state.Message, "五層高塔的庭院") {
		t.Fatalf("wizard-tower transition block=%#x geo=%d/%#x pos=(%d,%d,%d) pieces=%v picture=%v/%d message=%q",
			session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection, state.LoadPieces,
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "德拉坎德羅斯") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "PARLAY"}) {
		t.Fatalf("Dracandros arrival message=%q choices=%#v", state.Message, state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	wizardTowerMessages := []string{
		"一大群黑龍",
		"獨自走上前來",
		"伊爾明斯特",
		"只是一道幻象",
		"手札條目 15",
		"一枚枷印逐漸消退",
	}
	for _, fragment := range wizardTowerMessages {
		for step := 0; step < 4 && !strings.Contains(state.Message, fragment); step++ {
			var err error
			if state.Mode == ModeEvent {
				err = state.Continue()
			} else {
				err = state.Select(0)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		if !strings.Contains(state.Message, fragment) {
			t.Fatalf("wizard-tower continuation want=%q mode=%v message=%q choices=%#v",
				fragment, state.Mode, state.Message, state.currentOriginalChoices)
		}
	}
	for step := 0; step < 4 && !reflect.DeepEqual(state.currentOriginalChoices,
		[]string{"ATTACK DRAGONS", "ATTACK WIZARD", "FLEE", "PARLAY WITH THE DRAGONS"}); step++ {
		var err error
		if state.Mode == ModeEvent {
			err = state.Continue()
		} else {
			err = state.Select(0)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices,
		[]string{"ATTACK DRAGONS", "ATTACK WIZARD", "FLEE", "PARLAY WITH THE DRAGONS"}) ||
		!reflect.DeepEqual(state.Choices, []string{"攻擊龍群", "攻擊法師", "撤退", "與龍群交涉"}) {
		t.Fatalf("wizard-tower decision originals=%#v localized=%#v",
			state.currentOriginalChoices, state.Choices)
	}
	if got, ok := session.MemoryValue(0x4CFF); !ok || got != 1 {
		t.Fatalf("wizard-tower faded-bond flag=%#x,%v want 1", got, ok)
	}
	journal15Pages := 0
	for _, page := range state.JournalPages {
		if strings.HasPrefix(page, "手札條目 15") {
			journal15Pages++
		}
	}
	if journal15Pages != 2 {
		t.Fatalf("wizard-tower journal 15 pages=%d, want 2", journal15Pages)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "人類之間的爭端") || !state.PictureRequested ||
		state.PictureBlock != 0x35 {
		t.Fatalf("attack-wizard dragons depart picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	sawDracandrosCallTroops := strings.Contains(state.Message, "保護我")
	for step := 0; step < 4 && state.Mode != ModeCombat; step++ {
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		} else if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		sawDracandrosCallTroops = sawDracandrosCallTroops || strings.Contains(state.Message, "保護我")
	}
	if state.Mode != ModeCombat || !sawDracandrosCallTroops {
		t.Fatalf("attack-wizard combat mode=%v saw-call=%v message=%q",
			state.Mode, sawDracandrosCallTroops, state.Message)
	}
	towerFighters := state.CombatFighters()
	efreeti, darkElfFighters, darkElfMages := 0, 0, 0
	for _, fighter := range towerFighters {
		if fighter.Side == combat.SideParty {
			continue
		}
		switch {
		case fighter.Name == "伊弗利特" && fighter.SpriteBlock == 0x34:
			efreeti++
		case fighter.Name == "黑暗精靈戰士" && fighter.SpriteBlock == 0x31:
			darkElfFighters++
		case fighter.Name == "黑暗精靈法師" && fighter.SpriteBlock == 0x32:
			darkElfMages++
		default:
			t.Fatalf("unexpected wizard-tower defender=%+v", fighter)
		}
	}
	if efreeti != 1 || darkElfFighters != 2 || darkElfMages != 1 {
		t.Fatalf("wizard-tower defenders efreeti=%d fighters=%d mages=%d",
			efreeti, darkElfFighters, darkElfMages)
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "安全休息") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("wizard-tower victory continuation mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || session.CurrentBlockID() != 0x33 {
		t.Fatalf("wizard-tower roof return mode=%v block=%#x message=%q",
			state.Mode, session.CurrentBlockID(), state.Message)
	}
	session.SetMemoryValue(0x4C60, 0)
	state.PictureRequested = false
	if err := state.StartDungeonStoryPreview(0x33, 0x32, 5); err != nil {
		t.Fatal(err)
	}
	towerDecision := []string{"ATTACK DRAGONS", "ATTACK WIZARD", "FLEE", "PARLAY WITH THE DRAGONS"}
	for step := 0; step < 24 && !reflect.DeepEqual(state.currentOriginalChoices, towerDecision); step++ {
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		selection := 0
		if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "PARLAY"}) {
			selection = 1
		}
		if err := state.Select(selection); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, towerDecision) {
		t.Fatalf("replayed wizard-tower decision mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	parlayTactics := []string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK", "PARLAY_NICE", "PARLAY_ABUSIVE"}
	for step := 0; step < 4 && !reflect.DeepEqual(state.currentOriginalChoices, parlayTactics); step++ {
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		} else if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, parlayTactics) ||
		!reflect.DeepEqual(state.Choices, []string{
			state.catalog.Text("parlay_haughty", ""), state.catalog.Text("parlay_sly", ""),
			state.catalog.Text("parlay_meek", ""), state.catalog.Text("parlay_nice", ""),
			state.catalog.Text("parlay_abusive", ""),
		}) {
		t.Fatalf("wizard-tower parlay originals=%#v localized=%#v",
			state.currentOriginalChoices, state.Choices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	sawDragonsConvinced := strings.Contains(state.Message, "沒有對付龍族的陰謀")
	sawParlayTroops := strings.Contains(state.Message, "保護我")
	for step := 0; step < 8 && state.Mode != ModeCombat; step++ {
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		} else if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		sawDragonsConvinced = sawDragonsConvinced || strings.Contains(state.Message, "沒有對付龍族的陰謀")
		sawParlayTroops = sawParlayTroops || strings.Contains(state.Message, "保護我")
	}
	if state.Mode != ModeCombat || !sawDragonsConvinced || !sawParlayTroops {
		t.Fatalf("wizard-tower successful parlay mode=%v convinced=%v troops=%v message=%q",
			state.Mode, sawDragonsConvinced, sawParlayTroops, state.Message)
	}
	parlayDefenders := state.CombatFighters()
	parlayEfreeti, parlayFighters, parlayMages := 0, 0, 0
	for _, fighter := range parlayDefenders {
		if fighter.Side == combat.SideParty {
			continue
		}
		switch {
		case fighter.Name == "伊弗利特" && fighter.SpriteBlock == 0x34:
			parlayEfreeti++
		case fighter.Name == "黑暗精靈戰士" && fighter.SpriteBlock == 0x31:
			parlayFighters++
		case fighter.Name == "黑暗精靈法師" && fighter.SpriteBlock == 0x32:
			parlayMages++
		default:
			t.Fatalf("unexpected parlay defender=%+v", fighter)
		}
	}
	if parlayEfreeti != 1 || parlayFighters != 2 || parlayMages != 1 {
		t.Fatalf("parlay defenders efreeti=%d fighters=%d mages=%d",
			parlayEfreeti, parlayFighters, parlayMages)
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "安全休息") {
		t.Fatalf("parlay victory continuation mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || session.CurrentBlockID() != 0x33 {
		t.Fatalf("parlay roof return mode=%v block=%#x", state.Mode, session.CurrentBlockID())
	}
	session.SetMemoryValue(0x4C60, 0)
	session.SetMemoryValue(0x4C61, 1)
	session.SetMemoryValue(0x4C64, 0)
	state.PictureRequested = false
	state.partyRoster[0].HitPoints = 500
	state.partyRoster[0].MaxHitPoints = 500
	state.partyRoster[0].SavingThrows = []uint8{21, 21, 21, 21, 21}
	state.party[0].HitPoints = 500
	state.party[0].MaxHitPoints = 500
	if err := state.StartDungeonStoryPreview(0x33, 0x32, 5); err != nil {
		t.Fatal(err)
	}
	for step := 0; step < 24 && !reflect.DeepEqual(state.currentOriginalChoices, towerDecision); step++ {
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		selection := 0
		if reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "WAIT", "FLEE", "PARLAY"}) {
			selection = 1
		}
		if err := state.Select(selection); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, towerDecision) {
		t.Fatalf("attack-dragons replay mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	sawDragonsCondemn := strings.Contains(state.Message, "自行定罪")
	for step := 0; step < 4 && state.Mode != ModeCombat; step++ {
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		} else if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		sawDragonsCondemn = sawDragonsCondemn || strings.Contains(state.Message, "自行定罪")
	}
	if state.Mode != ModeCombat || !sawDragonsCondemn {
		t.Fatalf("attack-dragons combat mode=%v condemn=%v message=%q",
			state.Mode, sawDragonsCondemn, state.Message)
	}
	dragonCount := 0
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideEnemy {
			if fighter.Name != requireCombatantName(t, &state, "BLACK DRAGON") || fighter.SpriteBlock != 0x35 {
				t.Fatalf("unexpected wizard-tower dragon=%+v", fighter)
			}
			dragonCount++
		}
	}
	if dragonCount != 14 {
		t.Fatalf("wizard-tower black dragons=%d, want 14", dragonCount)
	}
	for turn := 0; turn < 32 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "黑龍的屍體") ||
		!strings.Contains(state.Message, "龍心") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("dragon-heart prompt mode=%v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	hpBeforeAcid := state.partyRoster[0].HitPoints
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "酸液") || !strings.Contains(state.Message, "龍心") {
		t.Fatalf("dragon-heart acid message=%q", state.Message)
	}
	state.SetECLSeed(3)
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.partyRoster[0].HitPoints >= hpBeforeAcid || state.party[0].HitPoints != state.partyRoster[0].HitPoints {
		t.Fatalf("dragon-heart acid HP roster=%d fighter=%d before=%d pending=%#v",
			state.partyRoster[0].HitPoints, state.party[0].HitPoints, hpBeforeAcid, state.pendingDamageRequests)
	}
	if got, ok := session.MemoryValue(0x4C64); !ok || got != 1 {
		t.Fatalf("dragon-heart flag=%#x,%v want 1", got, ok)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "安全休息") {
		t.Fatalf("dragon-heart safe roof mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || session.CurrentBlockID() != 0x33 {
		t.Fatalf("dragon-heart roof return mode=%v block=%#x", state.Mode, session.CurrentBlockID())
	}
	if terrain := towerGrid.CellWrapped(7, 15).Terrain; terrain != 1 {
		t.Fatalf("wizard-tower exit terrain=%#x, want 1", terrain)
	}
	state.catalog.Strings["wilderness"] = "荒野"
	state.DungeonX, state.DungeonY, state.DungeonDirection = 7, 15, 2
	state.DungeonWallType, _ = towerGrid.WallWrapped(7, 15, 2)
	state.DungeonWallRoof = towerGrid.CellWrapped(7, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	roofExitChoices := []string{"CAVES", "WILDERNESS", "STAY HERE"}
	if !strings.Contains(state.Message, "直達荒野的祕道") ||
		!reflect.DeepEqual(state.currentOriginalChoices, roofExitChoices) ||
		!reflect.DeepEqual(state.Choices, []string{"洞穴", "荒野", "留在這裡"}) {
		t.Fatalf("wizard-tower roof exit originals=%#v localized=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || session.CurrentBlockID() != 0x33 {
		t.Fatalf("wizard-tower stay mode=%v block=%#x", state.Mode, session.CurrentBlockID())
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 7, 15, 2
	state.DungeonWallRoof = towerGrid.CellWrapped(7, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x32 || state.GeoMapBlock != 0x32 ||
		state.DungeonX != 6 || state.DungeonY != 15 || state.DungeonDirection != 2 {
		t.Fatalf("wizard-tower caves return block=%#x geo=%#x pos=(%d,%d,%d) mode=%v message=%q",
			session.CurrentBlockID(), state.GeoMapBlock, state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.Mode, state.Message)
	}
	session.SetMemoryValue(0x4C60, 1)
	if err := state.StartDungeonStoryPreview(0x33, 0x32, 5); err != nil {
		t.Fatal(err)
	}
	// The preview entry faithfully raises the tower establishing picture.
	// This revisit test is concerned with the already-completed roof exit.
	state.Mode = ModeDungeon
	state.PictureRequested = false
	state.pendingPictureResult = nil
	state.DungeonX, state.DungeonY, state.DungeonDirection = 7, 15, 2
	state.DungeonWallType, _ = towerGrid.WallWrapped(7, 15, 2)
	state.DungeonWallRoof = towerGrid.CellWrapped(7, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "哈普圖斯村") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"VILLAGE", "DEPART"}) ||
		!reflect.DeepEqual(state.Choices, []string{"村莊", "離開此區"}) {
		t.Fatalf("wizard-tower wilderness submenu mode=%v originals=%#v localized=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x31 || state.GeoMapBlock != 0x32 {
		t.Fatalf("wizard-tower village return block=%#x geo=%#x mode=%v message=%q",
			session.CurrentBlockID(), state.GeoMapBlock, state.Mode, state.Message)
	}
	state.partyRoster[0].Equipment = append(state.partyRoster[0].Equipment,
		monster.ItemRecord{Type: 0x5E},
		monster.ItemRecord{Type: 0x60},
		monster.ItemRecord{Type: 0x61},
		monster.ItemRecord{Type: 0x09},
	)
	session.SetMemoryValue(0x4C60, 1)
	if err := state.StartDungeonStoryPreview(0x33, 0x32, 5); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.PictureRequested = false
	state.pendingPictureResult = nil
	state.DungeonX, state.DungeonY, state.DungeonDirection = 7, 15, 2
	state.DungeonWallType, _ = towerGrid.WallWrapped(7, 15, 2)
	state.DungeonWallRoof = towerGrid.CellWrapped(7, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	for _, itemType := range []uint8{0x5E, 0x60, 0x61} {
		found := false
		for _, item := range state.partyRoster[0].Equipment {
			found = found || item.Type == itemType
		}
		if !found {
			t.Fatalf("dark-elf item %#x disappeared before DEPART", itemType)
		}
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x30 ||
		state.Message != requireGamePackText(t, &state, "area5.depart-akabar") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("wizard-tower Akabar farewell block=%#x mode=%v originals=%#v message=%q",
			session.CurrentBlockID(), state.Mode, state.currentOriginalChoices, state.Message)
	}
	if len(state.partyRoster) != 1 || state.partyRoster[0].ID != "hero" {
		t.Fatalf("Akabar departure roster=%#v, want hero only", state.partyRoster)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x30 ||
		state.Message != requireGamePackText(t, &state, "area5.dark-elf-gear-decays") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("wizard-tower dark-elf decay block=%#x mode=%v originals=%#v message=%q",
			session.CurrentBlockID(), state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x50 || !state.PictureRequested || state.PictureBlock != 121 {
		t.Fatalf("wizard-tower depart world return block=%#x mode=%v picture=%v/%d message=%q",
			session.CurrentBlockID(), state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	for _, item := range state.partyRoster[0].Equipment {
		if item.Type == 0x5E || item.Type == 0x60 || item.Type == 0x61 {
			t.Fatalf("sunlight did not destroy dark-elf item %#x", item.Type)
		}
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("wizard-tower depart routes=%#v mode=%v message=%q",
			state.currentOriginalChoices, state.Mode, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ESSEMBRA"}) {
		t.Fatalf("post-wizard destinations location=%v originals=%#v message=%q",
			state.Location, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("post-wizard Essembra routes=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "post-wizard.dracolich") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("post-wizard skeletal encounter originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	dracolichFighters := state.CombatFighters()
	if state.Mode != ModeCombat || len(dracolichFighters) != 2 ||
		dracolichFighters[1].Name != requireCombatantName(t, &state, "DRACOLICH") || dracolichFighters[1].SpriteSet != 5 ||
		dracolichFighters[1].SpriteBlock != 0x3C || dracolichFighters[1].ArmorClass != -6 ||
		dracolichFighters[1].HitPoints != 66 {
		t.Fatalf("post-wizard dracolich combat mode=%v fighters=%#v message=%q",
			state.Mode, dracolichFighters, state.Message)
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || state.Location != LocationEssembra ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) ||
		state.Message != requireGamePackText(t, &state, "essembra.edge") {
		t.Fatalf("post-dracolich arrival mode=%v location=%v originals=%#v message=%q",
			state.Mode, state.Location, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || session.CurrentBlockID() != 0x50 ||
		state.Area.InDungeon || state.Area.GameArea != 1 ||
		!state.PictureRequested || state.PictureBlock != 80 ||
		state.Message != requireGamePackText(t, &state, "essembra.places") {
		t.Fatalf("Essembra enter mode=%v block=0x%02X area=%+v originals=%#v message=%q picture=%v/%d",
			state.Mode, session.CurrentBlockID(), state.Area,
			state.currentOriginalChoices, state.Message, state.PictureRequested, state.PictureBlock)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"INN", "STORE", "HALL", "TEMPLE", "BAR", "LEAVE"}) ||
		state.Message != requireGamePackText(t, &state, "essembra.places") {
		t.Fatalf("Essembra places mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "essembra.branching-oak") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Essembra inn welcome originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"SAVE", "VIEW", "MAGIC", "REST", "ALTER", "FIX", "EXIT"}) {
		t.Fatalf("Essembra inn camp services originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(6); err != nil {
		t.Fatal(err)
	}
	if state.PictureRequested {
		if state.PictureBlock != 80 {
			t.Fatalf("Essembra inn return picture=%d, want PIC1 80", state.PictureBlock)
		}
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || session.CurrentBlockID() != 0x50 ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"INN", "STORE", "HALL", "TEMPLE", "BAR", "LEAVE"}) ||
		state.Message != requireGamePackText(t, &state, "essembra.places") {
		t.Fatalf("Essembra inn return mode=%v block=0x%02X originals=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "essembra.outdoor-bar") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"HAVE A DRINK", "RELAX", "EXIT"}) {
		t.Fatalf("Essembra outdoor bar originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	journalCountBeforeEssembraBar := len(state.JournalPages)
	if state.Message != requireGamePackText(t, &state, "essembra.tavern-tale-60") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Essembra tavern tale originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if len(state.JournalPages) != journalCountBeforeEssembraBar {
		t.Fatalf("Essembra tavern tale incorrectly unlocked journal pages: before=%d after=%d",
			journalCountBeforeEssembraBar, len(state.JournalPages))
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"HAVE A DRINK", "RELAX", "EXIT"}) {
		t.Fatalf("Essembra bar continuation stalled mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "要喝什麼") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"BEER", "ALE", "PORT", "MEAD", "WHISKEY", "EXIT"}) ||
		!reflect.DeepEqual(state.Choices, []string{"啤酒", "愛爾啤酒", "波特酒", "蜂蜜酒", "威士忌", "離開"}) {
		t.Fatalf("Essembra drink menu prompt=%q originals=%#v choices=%#v message=%q",
			state.Prompt, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "essembra.tavern-tale-44") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Essembra beer tale originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"HAVE A DRINK", "RELAX", "EXIT"}) {
		t.Fatalf("Essembra beer continuation mode=%v originals=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"INN", "STORE", "HALL", "TEMPLE", "BAR", "LEAVE"}) ||
		state.Message != requireGamePackText(t, &state, "essembra.places") {
		t.Fatalf("Essembra bar exit mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(5); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 121 ||
		state.Message != requireGamePackText(t, &state, "essembra.edge") {
		t.Fatalf("Essembra leave picture mode=%v block=0x%02X originals=%#v message=%q picture=%v/%d",
			state.Mode, session.CurrentBlockID(), state.currentOriginalChoices,
			state.Message, state.PictureRequested, state.PictureBlock)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Essembra edge mode=%v originals=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"HAP", "THE STANDING STONE"}) {
		t.Fatalf("Essembra departure destinations=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("Standing Stone route choices=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "灰袍男子") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Standing Stone return mode=%v location=%v originals=%#v message=%q",
			state.Mode, state.Location, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"THANK HIM", "ATTACK", "LEAVE"}) {
		t.Fatalf("Standing Stone return counsel originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Standing Stone leave counsel originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ASHABENFORD", "ESSEMBRA", "HILLSFAR", "MYTH DRANNOR"}) {
		t.Fatalf("Standing Stone destinations for Hillsfar=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("Hillsfar route choices=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x51 || !state.PictureRequested || state.PictureBlock != 121 ||
		state.Message != requireGamePackText(t, &state, "hillsfar.fire-knives-ambush") {
		t.Fatalf("Hillsfar trail ambush mode=%v block=0x%02X location=%v originals=%#v message=%q picture=%v/%d",
			state.Mode, session.CurrentBlockID(), state.Location, state.currentOriginalChoices,
			state.Message, state.PictureRequested, state.PictureBlock)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hillsfar ambush continuation mode=%v originals=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	hillsfarPatrol := state.CombatFighters()
	if state.Mode != ModeCombat || len(hillsfarPatrol) != 7 {
		t.Fatalf("Hillsfar trail combat mode=%v fighters=%#v message=%q",
			state.Mode, hillsfarPatrol, state.Message)
	}
	for _, fighter := range hillsfarPatrol[1:] {
		if fighter.Name != "戰士" || fighter.SpriteBlock != 0x20 ||
			fighter.Side != combat.SideEnemy {
			t.Fatalf("Hillsfar trail enemy=%+v", fighter)
		}
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || state.Location != LocationHillsfar ||
		state.Area.CurrentCity != 11 ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) ||
		state.Message != requireGamePackText(t, &state, "hillsfar.edge") {
		t.Fatalf("Hillsfar arrival mode=%v block=0x%02X location=%v currentCity=%d originals=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.Location, state.Area.CurrentCity,
			state.currentOriginalChoices, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C9B); !ok || got != 11 {
		t.Fatalf("Hillsfar current-location memory=%#x,%v want 11", got, ok)
	}
	state.ConsumeMusicEvents()
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 80 ||
		state.Message != requireGamePackText(t, &state, "hillsfar.places") {
		t.Fatalf("Hillsfar enter mode=%v block=0x%02X location=%v originals=%#v message=%q picture=%v/%d",
			state.Mode, session.CurrentBlockID(), state.Location, state.currentOriginalChoices,
			state.Message, state.PictureRequested, state.PictureBlock)
	}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, []MusicEvent{{
		Action: "play", TrackID: "pc98-bgm-selector-06",
	}}) {
		t.Fatalf("Hillsfar enter music=%+v", got)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"INN", "STORE", "HALL", "TEMPLE", "BAR", "LEAVE"}) ||
		state.Message != requireGamePackText(t, &state, "hillsfar.places") {
		t.Fatalf("Hillsfar places mode=%v block=0x%02X originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hillsfar.dockside-bar") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"HAVE A DRINK", "RELAX", "EXIT"}) {
		t.Fatalf("Hillsfar bar mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "hillsfar.red-plumes-spill-drinks") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
		t.Fatalf("Hillsfar Red Plume provocation originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	hillsfarRedPlumes := state.CombatFighters()
	if state.Mode != ModeCombat || len(hillsfarRedPlumes) != 7 {
		t.Fatalf("Hillsfar Red Plume combat mode=%v fighters=%#v message=%q",
			state.Mode, hillsfarRedPlumes, state.Message)
	}
	for _, fighter := range hillsfarRedPlumes[1:] {
		if fighter.Name != "戰士" || fighter.SpriteBlock != 0x20 ||
			fighter.Side != combat.SideEnemy {
			t.Fatalf("Hillsfar Red Plume enemy=%+v", fighter)
		}
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"HAVE A DRINK", "RELAX", "EXIT"}) ||
		state.Message != requireGamePackText(t, &state, "hillsfar.dockside-bar") {
		t.Fatalf("Hillsfar Red Plume victory mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 80 ||
		state.Message != requireGamePackText(t, &state, "hillsfar.places") {
		t.Fatalf("Hillsfar bar exit picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if got := state.ConsumeMusicEvents(); len(got) != 0 {
		t.Fatalf("Hillsfar bar exit replayed town music=%+v", got)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"INN", "STORE", "HALL", "TEMPLE", "BAR", "LEAVE"}) {
		t.Fatalf("Hillsfar bar exit places originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(5); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 121 ||
		state.Message != requireGamePackText(t, &state, "hillsfar.edge") {
		t.Fatalf("Hillsfar leave picture=%v/%d message=%q",
			state.PictureRequested, state.PictureBlock, state.Message)
	}
	if got := state.ConsumeMusicEvents(); !reflect.DeepEqual(got, []MusicEvent{{
		Action: "play", TrackID: "pc98-bgm-selector-05",
	}}) {
		t.Fatalf("Hillsfar leave music=%+v", got)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Hillsfar edge return originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"YULASH", "THE STANDING STONE", "PHLAN"}) ||
		!reflect.DeepEqual(state.Choices, []string{"尤拉什", "立石群", "弗蘭"}) {
		t.Fatalf("Hillsfar destinations originals=%#v choices=%#v prompt=%q",
			state.currentOriginalChoices, state.Choices, state.Prompt)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("Yulash route choices=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "yulash.red-plume-patrol") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Yulash Red Plume patrol originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	yulashPatrol := state.CombatFighters()
	if state.Mode != ModeCombat || len(yulashPatrol) != 13 {
		t.Fatalf("Yulash Red Plume combat mode=%v fighters=%#v message=%q",
			state.Mode, yulashPatrol, state.Message)
	}
	for _, fighter := range yulashPatrol[1:] {
		if fighter.Name != "戰士" || fighter.SpriteBlock != 0x20 ||
			fighter.Side != combat.SideEnemy {
			t.Fatalf("Yulash Red Plume enemy=%+v", fighter)
		}
	}
	for turn := 0; turn < 24 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness || state.Location != LocationYulash ||
		state.Area.CurrentCity != 10 ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) ||
		state.Message != requireGamePackText(t, &state, "yulash.edge") {
		t.Fatalf("Yulash arrival mode=%v block=0x%02X location=%v currentCity=%d originals=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.Location, state.Area.CurrentCity,
			state.currentOriginalChoices, state.Message)
	}
	if got, ok := session.MemoryValue(0x4C9B); !ok || got != 10 {
		t.Fatalf("Yulash current-location memory=%#x,%v want 10", got, ok)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if session.CurrentBlockID() != 0x10 ||
		state.Message != requireGamePackText(t, &state, "yulash.entry") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"SNEAK IN", "ASK PERMISSION", "LEAVE"}) ||
		!reflect.DeepEqual(state.Choices, []string{"潛入", "請求許可", "離開"}) {
		t.Fatalf("Yulash entry mode=%v block=0x%02X area=%+v originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.Area,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "yulash.riders-burst-out") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Yulash riders originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "yulash.checkpoint-halt") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"RUN AWAY", "FIGHT", "PARLAY"}) ||
		!reflect.DeepEqual(state.Choices, []string{"逃走", "戰鬥", "談判"}) {
		t.Fatalf("Yulash checkpoint originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "yulash.see-commander") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"GO WITH GUARDS", "FIGHT", "RUN AWAY"}) ||
		!reflect.DeepEqual(state.Choices, []string{"跟衛兵走", "戰鬥", "逃走"}) {
		t.Fatalf("Yulash checkpoint parlay originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "yulash.waiting-room") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Yulash waiting room originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || session.CurrentBlockID() != 0x10 ||
		state.Area.GameArea != 3 || !state.Area.InDungeon ||
		state.GeoMapSet != 3 || state.GeoMapBlock != 0x10 ||
		!state.geoMapPending ||
		state.DungeonX != 0 || state.DungeonY != 3 || state.DungeonDirection != 2 {
		t.Fatalf("Yulash waiting room transition mode=%v block=0x%02X area=%+v geo=%d/%d pending=%v coords=%d,%d,%d message=%q",
			state.Mode, session.CurrentBlockID(), state.Area, state.GeoMapSet, state.GeoMapBlock,
			state.geoMapPending, state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.Message)
	}
	yulashGeoCatalog := geo.NewCatalog()
	if err := yulashGeoCatalog.AddDAX(3, zipData(t, image, "GEO3.DAX")); err != nil {
		t.Fatal(err)
	}
	yulashGrid, ok := yulashGeoCatalog.Lookup(geo.MapRef{Set: 3, BlockID: 0x10})
	if !ok {
		t.Fatal("Yulash GEO3 block 0x10 is unavailable")
	}
	state.DungeonX = 1
	state.DungeonWallType, _ = yulashGrid.WallWrapped(1, 3, 2)
	state.DungeonWallRoof = yulashGrid.CellWrapped(1, 3).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x9A ||
		state.Message != requireGamePackText(t, &state, "yulash.zhentarim-spies") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"FIGHT THE MEN", "LET THEM GO"}) ||
		!reflect.DeepEqual(state.Choices, []string{"攔下他們戰鬥", "放他們離開"}) {
		t.Fatalf("Yulash Zhentarim-spy event terrain=%#x mode=%v originals=%#v choices=%#v message=%q",
			state.DungeonWallRoof, state.Mode,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	spyFighters := state.CombatFighters()
	if state.Mode != ModeCombat || len(spyFighters) != 12 {
		t.Fatalf("Yulash spy fight mode=%v fighters=%+v message=%q",
			state.Mode, spyFighters, state.Message)
	}
	spyCounts := map[string]int{}
	for _, fighter := range spyFighters[1:] {
		spyCounts[fighter.Name]++
	}
	if !reflect.DeepEqual(spyCounts, map[string]int{
		requireCombatantName(t, &state, "ZHENTRIM CLERIC"): 1,
		requireCombatantName(t, &state, "ZHENTRIM FGHTR"):  8,
		requireCombatantName(t, &state, "ZHENTRIM MAGE"):   2,
	}) {
		t.Fatalf("Yulash spy enemy counts=%#v", spyCounts)
	}
	for turn := 0; turn < 24 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeWilderness ||
		state.Message != requireGamePackText(t, &state, "yulash.led-to-commander") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Yulash post-spy-victory mode=%v block=0x%02X originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(),
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "yulash.commander-business") ||
		!reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK", "PARLAY_NICE", "PARLAY_ABUSIVE"}) {
		t.Fatalf("Yulash commander audience mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "journal-trigger.yulash-commander-22") ||
		!slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.22.1")) ||
		!slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.22.2")) {
		t.Fatalf("Yulash commander approval mode=%v originals=%#v choices=%#v message=%q journals=%#v",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message, state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "yulash.commander-side-door") ||
		!slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.52")) {
		t.Fatalf("Yulash commander pass mode=%v originals=%#v choices=%#v message=%q journals=%#v",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message, state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon ||
		state.DungeonX != 1 || state.DungeonY != 3 || state.DungeonDirection != 2 ||
		len(state.Choices) != 0 || state.Message != "" {
		t.Fatalf("Yulash commander exit mode=%v coords=%d,%d,%d originals=%#v choices=%#v message=%q",
			state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || len(state.Choices) != 0 {
		t.Fatalf("Yulash consumed spy event replayed mode=%v choices=%#v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 11, 0, 0
	state.DungeonWallType, _ = yulashGrid.WallWrapped(11, 0, 0)
	state.DungeonWallRoof = yulashGrid.CellWrapped(11, 0).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x26 || state.Mode != ModeEvent ||
		state.Message != requireGamePackText(t, &state, "yulash.pit-entrance") {
		t.Fatalf("Yulash Pit entrance terrain=%#x wall=%#x mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.DungeonWallRoof, state.DungeonWallType, state.Mode, session.CurrentBlockID(),
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode == ModeWilderness &&
		reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Yulash Pit entrance continuation mode=%v return=%v picture=%v originals=%#v message=%q",
			state.Mode, state.eventReturnMode, state.PictureRequested,
			state.currentOriginalChoices, state.Message)
	}
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || session.CurrentBlockID() != 0x11 ||
		state.GeoMapSet != 3 || state.GeoMapBlock != 0x11 || !state.geoMapPending ||
		state.DungeonX != 0 || state.DungeonY != 0 || state.DungeonDirection != 2 ||
		state.Message != requireGamePackText(t, &state, "pit.opening-dead-cultists") {
		t.Fatalf("Yulash Pit transition terrain=%#x wall=%#x mode=%v block=0x%02x geo=%d/%d pending=%v coords=%d,%d,%d originals=%#v choices=%#v message=%q",
			state.DungeonWallRoof, state.DungeonWallType, state.Mode, session.CurrentBlockID(),
			state.GeoMapSet, state.GeoMapBlock, state.geoMapPending,
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.opening-chosen") {
		t.Fatalf("Pit wounded cleric mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.trapped") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit collapse mode=%v block=0x%02x coords=%d,%d,%d originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.cleric-dies") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit cleric death mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(),
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.ambience") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit ambience mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(),
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || session.CurrentBlockID() != 0x11 ||
		state.GeoMapSet != 3 || state.GeoMapBlock != 0x11 ||
		state.DungeonX != 0 || state.DungeonY != 0 || state.DungeonDirection != 2 {
		t.Fatalf("Pit opening return mode=%v block=0x%02x geo=%d/%d coords=%d,%d,%d originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	pitLevelOne, ok := yulashGeoCatalog.Lookup(geo.MapRef{Set: 3, BlockID: 0x11})
	if !ok {
		t.Fatal("Pit of Moander GEO3 block 0x11 is unavailable")
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 2, 4, 0
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(2, 4, 0)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(2, 4).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	for step := 0; step < 6 && state.Mode == ModeWilderness &&
		reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}); step++ {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Pit monster-remains continuation mode=%v originals=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 1, 4, 0
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(1, 4, 0)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(1, 4).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x85 ||
		state.Message != requireGamePackText(t, &state, "pit.alias-dragonbait-meet") {
		t.Fatalf("Pit Alias/Dragonbait meeting terrain=%#x wall=%#x mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.DungeonWallRoof, state.DungeonWallType, state.Mode, session.CurrentBlockID(),
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.alias-bonded-reaction") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"COMBAT", "ADVANCE", "WAIT", "FLEE", "PARLAY"}) {
		t.Fatalf("Pit Alias/Dragonbait encounter mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, session.CurrentBlockID(), state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices,
		[]string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK", "PARLAY_NICE", "PARLAY_ABUSIVE"}) {
		t.Fatalf("Pit Alias/Dragonbait parlay originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.alias-dragonbait-introduction") {
		t.Fatalf("Pit Alias/Dragonbait introduction mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices,
		[]string{"TELL HER YOUR STORY", "TELL HER YOU'RE HUNTING CULTISTS", "TELL HER IT'S NONE OF HER AFFAIR"}) ||
		!reflect.DeepEqual(state.Choices,
			[]string{"告訴她你們的經歷", "告訴她你們正在追捕邪教徒", "告訴她這不關她的事"}) {
		t.Fatalf("Pit Alias/Dragonbait story choices originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "手札第 3 條") ||
		!slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.3.1")) ||
		!slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.3.2")) ||
		!slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.3.3")) {
		t.Fatalf("Pit Alias/Dragonbait story mode=%v originals=%#v choices=%#v message=%q journals=%#v",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message, state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.alias-dragonbait-join") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"NO", "YES"}) {
		t.Fatalf("Pit Alias/Dragonbait join prompt mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster) != 3 ||
		state.partyRoster[1].Name != "愛麗雅絲" ||
		state.partyRoster[1].ScriptName != "ALIAS" ||
		state.partyRoster[1].Class != party.ClassFighter ||
		state.partyRoster[2].Name != "龍餌" ||
		state.partyRoster[2].ScriptName != "DRAGONBAIT" ||
		state.partyRoster[2].Race != party.RaceSaurial ||
		state.partyRoster[2].Class != party.ClassPaladin ||
		state.Message != requireGamePackText(t, &state, "pit.alias-dragonbait-joined") {
		t.Fatalf("Pit Alias/Dragonbait joined mode=%v originals=%#v choices=%#v message=%q party=%#v",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message, state.partyRoster)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || len(state.partyRoster) != 3 {
		t.Fatalf("Pit Alias/Dragonbait return mode=%v originals=%#v message=%q party=%#v",
			state.Mode, state.currentOriginalChoices, state.Message, state.partyRoster)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || len(state.partyRoster) != 3 {
		t.Fatalf("Pit Alias/Dragonbait event replayed mode=%v originals=%#v message=%q party=%#v",
			state.Mode, state.currentOriginalChoices, state.Message, state.partyRoster)
	}
	state.DungeonX = 15
	state.DungeonY = 11
	state.DungeonDirection = 4
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(15, 11, 4)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(15, 11).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x11 ||
		state.Message != requireGamePackText(t, &state, "pit.stairs-down") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"NO", "YES"}) {
		t.Fatalf("Pit stairs prompt mode=%v block=0x%02x geo=%d pos=(%d,%d,%d) terrain=%#x wall=%#x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.DungeonWallRoof, state.DungeonWallType,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x12 ||
		state.GeoMapSet != 3 || state.GeoMapBlock != 0x11 ||
		state.DungeonX != 15 || state.DungeonY != 14 || state.DungeonDirection != 4 {
		t.Fatalf("Pit stairs transition mode=%v block=0x%02x geo=%d/%d pos=(%d,%d,%d) originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(15, 14, 4)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(15, 14).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.stairs-up") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"NO", "YES"}) {
		t.Fatalf("Pit lower stairs mode=%v block=0x%02x pos=(%d,%d,%d) terrain=%#x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.DungeonWallRoof,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.DungeonX = 14
	state.DungeonY = 15
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(14, 15, 4)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(14, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.dead-zhentrim") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"LEAVE", "EXAMINE CORPSE"}) {
		t.Fatalf("Pit Zhentil corpse mode=%v originals=%#v choices=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.zhentrim-scroll") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit Zhentil note mode=%v block=0x%02x pos=(%d,%d,%d) terrain=%#x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.DungeonWallRoof,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Pit Zhentil note return mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.Message != "" {
		t.Fatalf("Pit Zhentil note replayed mode=%v originals=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	state.DungeonX = 11
	state.DungeonY = 3
	state.DungeonDirection = 4
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(11, 3, 4)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(11, 3).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.mogion-altar") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit Mogion altar arrival mode=%v block=0x%02x pos=(%d,%d,%d) terrain=%#x picture=%v/%d originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.DungeonWallRoof, state.PictureRequested, state.PictureBlock,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.alias-identifies-mogion") {
		t.Fatalf("Pit Alias identifies Mogion originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.PictureRequested || state.PictureBlock != 17 ||
		state.Message != requireGamePackText(t, &state, "pit.mogion-greeting") {
		t.Fatalf("Pit Mogion introduction picture=%v/%d originals=%#v message=%q",
			state.PictureRequested, state.PictureBlock, state.currentOriginalChoices, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ATTACK", "FLEE", "WAIT", "PARLAY"}) {
		t.Fatalf("Pit Mogion encounter originals=%#v choices=%#v message=%q",
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(3); err != nil {
		t.Fatal(err)
	}
	var mogionCeremony []string
	for step := 0; step < 20 &&
		reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}); step++ {
		mogionCeremony = append(mogionCeremony, state.Message)
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, messageID := range []string{
		"pit.bond-paralysis",
		"pit.alias-dragonbait-tendrils",
		"pit.mogion-ritual",
		"pit.dimensional-window",
		"pit.moander-returns",
		"pit.bond-fades",
		"pit.bond-broken",
		"pit.alias-attack-mogion",
	} {
		if !slices.Contains(mogionCeremony, requireGamePackText(t, &state, messageID)) {
			t.Fatalf("Pit Mogion ceremony missing %s stages=%q", messageID, mogionCeremony)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ATTACK", "FLEE"}) {
		t.Fatalf("Pit Mogion ceremony originals=%#v choices=%#v message=%q stages=%q",
			state.currentOriginalChoices, state.Choices, state.Message, mogionCeremony)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	mogionBattle := state.CombatFighters()
	if state.Mode != ModeCombat || len(mogionBattle) != 15 || state.Message != "攻擊" {
		t.Fatalf("Pit Mogion combat mode=%v fighters=%#v message=%q",
			state.Mode, mogionBattle, state.Message)
	}
	mogionCount, cultistCount, moundCount := 0, 0, 0
	wantMogion := requireCombatantName(t, &state, "MOGION")
	wantCultist := requireCombatantName(t, &state, "CULTIST")
	wantMound := requireCombatantName(t, &state, "SHAMBLING MOUND")
	aliasPresent, dragonbaitPresent := false, false
	for _, fighter := range mogionBattle {
		switch fighter.Name {
		case wantMogion:
			mogionCount++
			if fighter.HitPoints != 60 || fighter.SpriteBlock != 0x18 {
				t.Fatalf("Pit Mogion record=%+v", fighter)
			}
		case wantCultist:
			cultistCount++
		case wantMound:
			moundCount++
		case "愛麗雅絲":
			aliasPresent = fighter.Side == combat.SideParty
		case "龍餌":
			dragonbaitPresent = fighter.Side == combat.SideParty
		}
	}
	if mogionCount != 1 || cultistCount != 6 || moundCount != 5 ||
		!aliasPresent || !dragonbaitPresent {
		t.Fatalf("Pit Mogion roster Mogion=%d cultists=%d mounds=%d Alias=%v Dragonbait=%v fighters=%#v",
			mogionCount, cultistCount, moundCount, aliasPresent, dragonbaitPresent, mogionBattle)
	}
	for turn := 0; turn < 120 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeEvent || state.Message != requireGamePackText(t, &state, "pit.rift-closes") {
		t.Fatalf("Pit Mogion victory mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.remnants-scream") {
		t.Fatalf("Pit Moander remnants emerge originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	moanderRemnants := state.CombatFighters()
	if state.Mode != ModeCombat || len(moanderRemnants) != 6 ||
		state.Message != requireGamePackText(t, &state, "pit.remnants-attack") {
		t.Fatalf("Pit Moander remnants mode=%v block=0x%02x message=%q fighters=%#v",
			state.Mode, state.session.CurrentBlockID(), state.Message, moanderRemnants)
	}
	remnantCount := 0
	wantRemnant := requireCombatantName(t, &state, "BIT O' MOANDER")
	for _, fighter := range moanderRemnants {
		if fighter.Name != wantRemnant {
			continue
		}
		remnantCount++
		if fighter.HitPoints != 140 || fighter.SpriteBlock != 0x1A ||
			fighter.CombatSize != 4 {
			t.Fatalf("Pit Moander remnant record=%+v", fighter)
		}
	}
	if remnantCount != 3 {
		t.Fatalf("Pit Moander remnant count=%d fighters=%#v", remnantCount, moanderRemnants)
	}
	for turn := 0; turn < 60 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Message != requireGamePackText(t, &state, "pit.gauntlet") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit Moander gauntlet mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	moanderGauntletFlag, _ := state.session.MemoryValue(0x4C5B)
	if moanderGauntletFlag != 1 {
		t.Fatalf("Pit Moander gauntlet flag4C5B=%#x", moanderGauntletFlag)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.Message != requireGamePackText(t, &state, "pit.priest-flees") {
		t.Fatalf("Pit Moander gauntlet continuation mode=%v block=0x%02x flag4C5B=%#x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), moanderGauntletFlag,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	var pitAfterGauntlet []string
	for step := 0; step < 12 && state.Mode != ModeDungeon; step++ {
		pitAfterGauntlet = append(pitAfterGauntlet, state.Message)
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
			continue
		}
		break
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Pit after gauntlet mode=%v block=0x%02x originals=%#v choices=%#v message=%q stages=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message, pitAfterGauntlet)
	}
	state.DungeonX = 12
	state.DungeonY = 0
	state.DungeonDirection = 2
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(12, 0, 2)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(12, 0).Terrain
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.altar-treasure") {
		t.Fatalf("Pit altar search mode=%v originals=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	gems, jewelry := state.TreasurePool()
	if !state.treasureMenu || state.Mode != ModeWilderness ||
		gems != 28 || jewelry != 10 || len(state.PendingTreasureItems()) == 0 {
		t.Fatalf("Pit altar treasure mode=%v menu=%v block=0x%02x originals=%#v choices=%#v message=%q items=%#v gems=%d jewelry=%d",
			state.Mode, state.treasureMenu, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message, state.PendingTreasureItems(), gems, jewelry)
	}
	if err := state.Select(len(state.Choices) - 1); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "神殿地圖") ||
		!strings.Contains(state.Message, "手札第 20") {
		t.Fatalf("Pit altar map mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message)
	}
	foundJournal20 := slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.20"))
	if !foundJournal20 {
		t.Fatalf("Pit altar Journal 20 was not unlocked: %v", state.JournalPages)
	}
	if reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Pit altar continuation mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message)
	}
	if err := state.SearchDungeonLocation(); err != nil {
		t.Fatal(err)
	}
	if state.Message == requireGamePackText(t, &state, "pit.altar-treasure") {
		t.Fatalf("Pit altar treasure repeated after one-time search: %q", state.Message)
	}
	state.DungeonX = 15
	state.DungeonY = 14
	state.DungeonDirection = 4
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(15, 14, 4)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(15, 14).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	for step := 0; step < 12 && state.Mode != ModeDungeon; step++ {
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if len(state.currentOriginalChoices) > 0 {
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x11 {
		t.Fatalf("Pit upstairs mode=%v block=0x%02x geo=%d/%d pos=(%d,%d,%d) originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	// This path asserts the explicit alias_alive departure branch. Keep Alias
	// alive through the real exit encounter instead of relying on the former
	// approximate initiative order to protect her by accident.
	aliasID := ""
	for index := range state.partyRoster {
		if state.partyRoster[index].ScriptName == "ALIAS" {
			aliasID = state.partyRoster[index].ID
			state.partyRoster[index].HitPoints = 20000
			state.partyRoster[index].MaxHitPoints = 20000
		}
	}
	for index := range state.party {
		if state.party[index].ID == aliasID {
			state.party[index].HitPoints = 20000
			state.party[index].MaxHitPoints = 20000
		}
	}
	if aliasID == "" {
		t.Fatal("Pit alias_alive fixture has no ALIAS party record")
	}
	state.DungeonX = 0
	state.DungeonY = 12
	state.DungeonDirection = 6
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(0, 12, 6)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(0, 12).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "pit.exit-last-stand") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit exit battle warning mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("Pit exit battle mode=%v event=%q message=%q", state.Mode, state.OriginalEvent, state.Message)
	}
	exitEnemies := map[uint8]int{}
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideEnemy {
			exitEnemies[fighter.SpriteBlock]++
		}
	}
	if !reflect.DeepEqual(exitEnemies, map[uint8]int{0x11: 10, 0x1C: 5, 0x19: 5}) {
		t.Fatalf("Pit exit enemies=%#v fighters=%#v", exitEnemies, state.CombatFighters())
	}
	for turn := 0; turn < 100 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Pit exit victory return mode=%v block=0x%02x originals=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices, state.Message)
	}
	state.DungeonX = 0
	state.DungeonY = 11
	state.DungeonDirection = 6
	state.DungeonWallType, _ = pitLevelOne.WallWrapped(0, 11, 6)
	state.DungeonWallRoof = pitLevelOne.CellWrapped(0, 11).Terrain
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.session.CurrentBlockID() != 0x51 ||
		!strings.Contains(state.Message, "愛麗雅絲說") ||
		!strings.Contains(state.Message, "忍冬花香") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Pit departure mode=%v block=0x%02x originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Message)
	}
	if flag, ok := state.session.MemoryValue(0x4C5B); !ok || flag != 0xFF {
		t.Fatalf("Pit departure gauntlet flag=%02x ok=%v", flag, ok)
	}
	if flag, ok := state.session.MemoryValue(0x7F12); !ok || flag != 1 {
		t.Fatalf("Pit departure progress flag=%02x ok=%v", flag, ok)
	}
	for _, character := range state.partyRoster {
		if character.ScriptName == "ALIAS" || character.ScriptName == "DRAGONBAIT" {
			t.Fatalf("Pit departure retained companion %#v", character)
		}
	}
	if len(state.partyRoster) != 1 || state.partyRoster[0].ID != hero.ID ||
		len(state.party) != 1 || state.party[0].ID != hero.ID {
		t.Fatalf("Pit departure party roster=%#v fighters=%#v", state.partyRoster, state.party)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP", "SEARCH AREA"}) ||
		state.Message != requireGamePackText(t, &state, "yulash.edge") {
		t.Fatalf("Pit departure return mode=%v originals=%#v message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"VOONLAR", "TESHWAVE", "ZHENTIL KEEP", "HILLSFAR"}) {
		t.Fatalf("post-Pit destinations mode=%v block=0x%02x originals=%#v choices=%#v prompt=%q message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Prompt, state.Message)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
		t.Fatalf("Zhentil routes mode=%v block=0x%02x originals=%#v choices=%#v prompt=%q message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices,
			state.Choices, state.Prompt, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "zhentil.patrol_pass") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Zhentil patrol mode=%v block=0x%02x originals=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Location != LocationZhentilKeep || state.Area.CurrentCity != 12 ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP", "SEARCH AREA"}) ||
		!reflect.DeepEqual(state.Choices, []string{"進入城市", "繼續旅程", "紮營", "搜索此區"}) ||
		state.Message != requireGamePackText(t, &state, "zhentil.edge") {
		t.Fatalf("Zhentil arrival mode=%v block=0x%02x location=%v currentCity=%d originals=%#v choices=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.Location, state.Area.CurrentCity,
			state.currentOriginalChoices, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.session.CurrentBlockID() != 0x20 ||
		!state.PictureRequested || state.PictureBlock != 32 ||
		state.Message != requireGamePackText(t, &state, "zhentil.guards_question") {
		t.Fatalf("Zhentil enter mode=%v block=0x%02x area=%+v location=%v message=%q picture=%v/%d",
			state.Mode, state.session.CurrentBlockID(), state.Area, state.Location,
			state.Message, state.PictureRequested, state.PictureBlock)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Zhentil questioning continuation originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "zhentil.guards_warning") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Zhentil guards warning originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "zhentil.inner_city") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Zhentil inner city originals=%#v message=%q",
			state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x20 ||
		state.Area.GameArea != 4 || !state.Area.InDungeon ||
		state.GeoMapSet != 4 || state.GeoMapBlock != 0x20 ||
		state.DungeonX != 2 || state.DungeonY != 0 || state.DungeonDirection != 4 ||
		len(state.currentOriginalChoices) != 0 || state.Message != "" {
		t.Fatalf("Zhentil inner-city entry mode=%v block=0x%02x area=%+v geo=%d/%d coords=%d,%d,%d originals=%#v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.Area,
			state.GeoMapSet, state.GeoMapBlock, state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.currentOriginalChoices, state.Message)
	}
	if !slices.Contains(state.JournalPages, requireGamePackText(t, &state, "journal.32")) {
		t.Fatalf("Zhentil entry did not unlock Journal 32: %#v", state.JournalPages)
	}
	if err := session.Reset(1); err != nil {
		t.Fatal(err)
	}
	if err := state.StartDungeonStoryPreview(0x32, 0x31, 5); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "古老的熔岩隧道") {
		t.Fatalf("lava tube preview mode=%v message=%q", state.Mode, state.Message)
	}
}

func TestRealCrossDAXNEWECLReachesECL1Entry(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocksByID := make(map[uint8][]byte)
	for _, member := range []string{"ECL1.DAX", "ECL4.DAX", "ECL5.DAX"} {
		data := zipData(t, image, member)
		blocks, parseErr := dax.Parse(data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			blocksByID[block.Entry.ID] = block.Data
		}
	}
	session, err := ecl.NewBlockSession(blocksByID, 0x25)
	if err != nil {
		t.Fatal(err)
	}
	result, runErr := session.RunFrom(555, 200, nil)
	if runErr != nil {
		t.Fatal(runErr)
	}
	if session.CurrentBlockID() != 0x50 {
		t.Fatalf("cross-DAX block=0x%02X, want ECL1 block 0x50; result=%+v", session.CurrentBlockID(), result)
	}
	if len(result.Text) == 0 || !strings.Contains(strings.Join(result.Text, " "), "YOU ARE AT THE EDGE OF TILVERTON") {
		t.Fatalf("cross-DAX text=%q, want ECL1 opening event", result.Text)
	}
	if !result.WaitingForMenu || len(result.Menus) != 1 || len(result.Menus[0].Options) != 3 {
		t.Fatalf("cross-DAX result=%+v, want opening menu pause", result)
	}
}

func TestRealECL3CallRedrawReachesStateAdapter(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	blocks, err := dax.Parse(zipData(t, image, "ECL3.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var block []byte
	for _, candidate := range blocks {
		if candidate.Entry.ID == 16 {
			block = candidate.Data
			break
		}
	}
	if len(block) == 0 {
		t.Fatal("ECL3 block 16 is absent")
	}
	result, err := ecl.RunSubsetInteractive(block, 0x0198, 40, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CallAddresses) != 1 || result.CallAddresses[0] != 0x2E10 {
		t.Fatalf("real CALL addresses=%#v, want [0x2E10]", result.CallAddresses)
	}
	state := State{}
	state.applyECLCallSignals(result)
	if got := state.ConsumeECLCallRequests(); len(got) != 1 || got[0] != 0x2E10 {
		t.Fatalf("State CALL requests=%#v", got)
	}
}

func TestRealECL1AddNPCBuildsThreePartyMembers(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	eclBlocks, err := dax.Parse(zipData(t, image, "ECL1.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	var opening []byte
	for _, block := range eclBlocks {
		if block.Entry.ID == 0x52 {
			opening = block.Data
			break
		}
	}
	if len(opening) == 0 {
		t.Fatal("ECL1 block 0x52 is absent")
	}
	result, err := ecl.RunSubset(opening, 0x14, 100)
	if err != nil {
		t.Fatal(err)
	}

	records := make(map[uint8]monster.Record)
	monsterBlocks, err := dax.Parse(zipData(t, image, "MON1CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		records[block.Entry.ID] = record
	}
	affects := make(map[uint8][]monster.AffectRecord)
	affectBlocks, err := dax.Parse(zipData(t, image, "MON1SPC.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range affectBlocks {
		parsed, parseErr := monster.ParseAffects(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		affects[block.Entry.ID] = parsed
	}
	items := make(map[uint8][]monster.ItemRecord)
	itemBlocks, err := dax.Parse(zipData(t, image, "MON1ITM.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range itemBlocks {
		parsed, parseErr := monster.ParseItems(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		items[block.Entry.ID] = parsed
	}

	session, err := ecl.NewBlockSession(map[uint8][]byte{0x52: opening}, 0x52)
	if err != nil {
		t.Fatal(err)
	}
	state := State{
		session:     session,
		partyRoster: party.Roster{{ID: "pc", Name: "玩家", IconID: 0, HitPoints: 10, MaxHitPoints: 10}},
		party:       []combat.Fighter{{ID: "pc", Name: "玩家", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10}},
	}
	state.SetMonsterRecordsForECL(1, records)
	state.SetMonsterAffectsForECL(1, affects)
	state.SetMonsterItemsForECL(1, items)
	if err := state.applyECLNPCSignals(result); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster) != 4 || len(state.party) != 4 {
		t.Fatalf("party sizes roster=%d fighters=%d", len(state.partyRoster), len(state.party))
	}
	for index, want := range []string{"RUSTLE", "CYNTHIA", "GRENDEL"} {
		character := state.partyRoster[index+1]
		if character.Name != want || !character.NPC || character.ControlMorale != 0xB2 || character.IconID != uint8(index+1) {
			t.Fatalf("NPC %d=%+v", index, character)
		}
		if state.party[index+1].Name != want || state.party[index+1].Side != combat.SideParty {
			t.Fatalf("NPC fighter %d=%+v", index, state.party[index+1])
		}
	}
	if state.whoSelectedIndex != 3 {
		t.Fatalf("selected index=%d, want last NPC", state.whoSelectedIndex)
	}
	state.Mode = ModeEvent
	state.PictureRequested = true
	state.pendingPictureResult = &result
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCombat || !state.CombatActive() {
		t.Fatalf("picture continuation did not enter real opening combat: mode=%v event=%q", state.Mode, state.OriginalEvent)
	}
}

func TestRealECL2EncounterBuildsBattleFromMON2CHA(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	eclData := zipData(t, image, "ECL2.DAX")
	eclBlocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	var encounterBlock []byte
	for _, block := range eclBlocks {
		if block.Entry.ID == 3 {
			encounterBlock = block.Data
			break
		}
	}
	if len(encounterBlock) == 0 {
		t.Fatal("ECL2 block 3 is absent")
	}
	result, err := ecl.RunSubset(encounterBlock, 0x2B0, 128)
	if err != nil {
		t.Fatal(err)
	}
	if !result.CombatRequested || len(result.MonsterSpawns) != 2 {
		t.Fatalf("ECL2 encounter result=%+v, want COMBAT with two spawn descriptors", result)
	}

	monsterData := zipData(t, image, "MON2CHA.DAX")
	monsterBlocks, err := dax.Parse(monsterData)
	if err != nil {
		t.Fatal(err)
	}
	records := make(map[uint8]monster.Record, len(monsterBlocks))
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		records[block.Entry.ID] = record
	}
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20,
		ArmorClass: 5, AttackBonus: 10, DamageDiceCount: 1, DamageDiceSides: 6, InitiativeBonus: 100,
	}}
	if err := state.StartEncounter(result, records, partyFighters, 37); err != nil {
		t.Fatal(err)
	}
	if !state.CombatActive() || len(state.CombatTargets()) == 0 {
		t.Fatalf("ECL2 battle was not created: fighters=%#v", state.CombatFighters())
	}
}

func zipData(t *testing.T, archive *zip.ReadCloser, name string) []byte {
	t.Helper()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		closeErr := reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		return data
	}
	t.Fatalf("archive member %q not found", name)
	return nil
}
