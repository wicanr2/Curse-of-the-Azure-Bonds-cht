package game

import (
	"archive/zip"
	"io"
	"path/filepath"
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

func TestRealNewGameBeginsAtGlobalBlockOne(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
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
	state := NewStateFromECLBlocks(testCatalog(), all, 0x50)
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
		!strings.Contains(state.Message, "小房間") || !strings.Contains(state.Message, "所有裝備都不見了") {
		t.Fatalf("new game first pause: mode=%v choices=%v message=%q", state.Mode, state.Choices, state.Message)
	}
	if state.LoadPieces != [3]uint16{1, 2, 3} {
		t.Fatalf("new game LOAD PIECES=%v", state.LoadPieces)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested ||
		!strings.Contains(state.Message, "持劍的手臂") || !strings.Contains(state.Message, "相同的印記") {
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
	state.DungeonX, state.DungeonY, state.DungeonDirection = 6, 13, 6
	state.DungeonWallType, _ = grid.WallWrapped(6, 13, 6)
	state.DungeonWallRoof = grid.CellWrapped(6, 13).Terrain
	if state.DungeonWallRoof != 0x86 {
		t.Fatalf("Windlord's Inn GEO selector=%#x, want 0x86", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 3 ||
		!state.SceneCharacterRequested || state.SceneHeadBlock != 3 || state.SceneBodyBlock != 3 ||
		!strings.Contains(state.Message, "歡迎來到美麗的提爾佛頓") ||
		!strings.Contains(state.Message, "旅店老闆娘") {
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
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "冒險手札") || !strings.Contains(state.Message, "賢者") {
		t.Fatalf("Windlord's Inn journal pause mode=%v choices=%v message=%q", state.Mode, state.Choices, state.Message)
	}
	if len(state.JournalPages) != 9 || !strings.Contains(state.JournalPages[8], "手札條目 31") ||
		!strings.Contains(state.JournalPages[8], "菲拉妮") {
		t.Fatalf("Journal Entry 31 was not unlocked in-game: pages=%v", state.JournalPages)
	}
	if err := state.OpenJournal(); err != nil {
		t.Fatal(err)
	}
	for state.JournalPage+1 < len(state.JournalPages) {
		if err := state.NextJournalPage(); err != nil {
			t.Fatal(err)
		}
	}
	if state.JournalPage != 8 || !strings.Contains(state.JournalText, "手札條目 31") {
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
	state.DungeonX, state.DungeonY, state.DungeonDirection = 7, 13, 2
	state.DungeonWallType, state.DungeonWallRoof = 0, 0
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
	if clock := state.GameTimeDisplay(); clock.Hour != 1 || clock.Minute != 0 {
		t.Fatalf("unsafe interrupted rest clock=%+v, want exactly one completed hour", clock)
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
	state.DungeonX, state.DungeonY, state.DungeonDirection = 6, 5, 0
	state.DungeonWallType, _ = grid.WallWrapped(6, 5, 0)
	state.DungeonWallRoof = grid.CellWrapped(6, 5).Terrain
	if state.DungeonWallRoof != 0x8A {
		t.Fatalf("Filani GEO selector=%#x, want 0x8A", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
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
	if len(state.JournalPages) != 12 ||
		!strings.HasPrefix(state.JournalPages[9], "手札條目 38（1/3）") ||
		!strings.Contains(state.JournalPages[11], "暗影谷") {
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
	state.partyRoster[0].Platinum = 0xFFFF
	state.DungeonX, state.DungeonY, state.DungeonDirection = 2, 12, 0
	state.DungeonWallType, _ = grid.WallWrapped(2, 12, 0)
	state.DungeonWallRoof = grid.CellWrapped(2, 12).Terrain
	if state.DungeonWallRoof != 0x84 {
		t.Fatalf("Weaponers GEO selector=%#x, want 0x84", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
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
	beforeWorth := characterCoinGoldWorth(state.partyRoster[0])
	beforeEquipment := len(state.partyRoster[0].Equipment)
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.shopStockMenu || len(state.Choices) < 2 {
		t.Fatalf("Weaponers stock menu=%v choices=%v", state.shopStockMenu, state.Choices)
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
	state.DungeonX, state.DungeonY, state.DungeonDirection = 0, 7, 0
	state.DungeonWallType, _ = grid.WallWrapped(0, 7, 0)
	state.DungeonWallRoof = grid.CellWrapped(0, 7).Terrain
	if state.DungeonWallRoof != 0x92 {
		t.Fatalf("Gond altar GEO selector=%#x, want 0x92", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
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
	character := &state.partyRoster[0]
	character.Class = party.ClassFighter
	character.Level = 1
	character.ClassLevels = [8]uint8{2: 1}
	character.Experience = 2001
	character.HitPoints, character.MaxHitPoints = 10, 10
	character.HealthStatus = party.HealthStatusOK
	character.Platinum = 0xFFFF
	beforeWorth = characterCoinGoldWorth(*character)
	state.DungeonX, state.DungeonY, state.DungeonDirection = 5, 2, 0
	state.DungeonWallType, _ = grid.WallWrapped(5, 2, 0)
	state.DungeonWallRoof = grid.CellWrapped(5, 2).Terrain
	if state.DungeonWallRoof != 0x8C {
		t.Fatalf("training hall GEO selector=%#x, want 0x8C", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
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
	state.DungeonX, state.DungeonY, state.DungeonDirection = 6, 10, 0
	state.DungeonWallType, _ = grid.WallWrapped(6, 10, 0)
	state.DungeonWallRoof = grid.CellWrapped(6, 10).Terrain
	if state.DungeonWallRoof != 0x88 {
		t.Fatalf("tavern GEO selector=%#x, want 0x88", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 4 ||
		!state.SceneCharacterRequested || state.SceneHeadBlock != 4 || state.SceneBodyBlock != 4 ||
		!strings.Contains(state.Message, "幾位想來點什麼") {
		t.Fatalf("tavern picture mode=%v picture=%v:%d head/body=%d/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 3 ||
		state.Choices[0] != "揍酒保" || state.Choices[1] != "喝一杯" || state.Choices[2] != "離開" {
		t.Fatalf("tavern action menu mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 4 ||
		state.Choices[0] != "龍息酒" || state.Choices[1] != "石化蜥蜴酒" ||
		state.Choices[2] != "檸檬水" || state.Choices[3] != "威士忌" {
		t.Fatalf("tavern drink menu mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(2); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.Choices[0] != "是" || state.Choices[1] != "否" ||
		!strings.Contains(state.Message, "特別的客人") {
		t.Fatalf("tavern special-customer prompt choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 1 || !strings.Contains(state.Message, "紫色腰帶") {
		t.Fatalf("tavern purple-sash pause choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.Choices[0] != "是" || state.Choices[1] != "否" ||
		!strings.Contains(state.Message, "騷動") || !strings.Contains(state.Message, "調查") {
		t.Fatalf("tavern investigate prompt choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 1 || !strings.Contains(state.Message, "華麗") ||
		!strings.Contains(state.Message, "匕首") || !strings.Contains(state.Message, "第 17 條") {
		t.Fatalf("tavern knife pause choices=%v message=%q", state.Choices, state.Message)
	}
	foundJournal17 := false
	for _, page := range state.JournalPages {
		if strings.HasPrefix(page, "手札條目 17：") && strings.Contains(page, "火刀") {
			foundJournal17 = true
			break
		}
	}
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
	state.DungeonX, state.DungeonY, state.DungeonDirection = 1, 10, 0
	state.DungeonWallType, _ = grid.WallWrapped(1, 10, 0)
	state.DungeonWallRoof = grid.CellWrapped(1, 10).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x8F || state.Mode != ModeEvent || !state.PictureRequested ||
		state.PictureBlock != 6 || state.SceneHeadBlock != 6 || state.SceneBodyBlock != 6 ||
		!strings.Contains(state.Message, "高階祭司") {
		t.Fatalf("high priest introduction selector=%#x mode=%v picture=%v:%d head/body=%d/%d message=%q",
			state.DungeonWallRoof, state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 2 ||
		state.Choices[0] != "是" || state.Choices[1] != "否" {
		t.Fatalf("high priest story prompt mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "移除詛咒") || !strings.Contains(state.Message, "第 19 條") {
		t.Fatalf("high priest remove-curse pause mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	foundJournal19 := false
	for _, page := range state.JournalPages {
		if strings.HasPrefix(page, "手札條目 19：") &&
			strings.Contains(page, "藍色火焰") && strings.Contains(page, "剛德") {
			foundJournal19 = true
			break
		}
	}
	if !foundJournal19 {
		t.Fatalf("Journal Entry 19 was not unlocked in-game: pages=%v", state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "離開此處") {
		t.Fatalf("high priest departure pause mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 1 || state.DungeonY != 10 ||
		len(state.Choices) != 0 {
		t.Fatalf("high priest continuation mode=%v position=(%d,%d) message=%q choices=%v originals=%v, want same cell",
			state.Mode, state.DungeonX, state.DungeonY, state.Message, state.Choices, state.currentOriginalChoices)
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
