package game

import (
	"archive/zip"
	"io"
	"path/filepath"
	"reflect"
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
	// Keep the single generated integration-test hero alive while five real
	// Royal Guards take their opening turns. A normal campaign has a full party.
	hero := state.PartyFighters()[0]
	hero.HitPoints, hero.MaxHitPoints = 200, 200
	hero.ArmorClass = -10
	hero.InitiativeBonus = 100
	hero.AttackBonus = 100
	hero.DamageDiceCount, hero.DamageDiceSides, hero.DamageBonus = 1, 1, 100
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		t.Fatal(err)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 1, 0, 0
	state.DungeonWallType, _ = grid.WallWrapped(1, 0, 0)
	state.DungeonWallRoof = grid.CellWrapped(1, 0).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "皇家衛兵") || !strings.Contains(state.Message, "暫時封閉") {
		t.Fatalf("first city-gate block mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.DungeonX != 1 || state.DungeonY != 0 {
		t.Fatalf("first city-gate return mode=%v position=(%d,%d)",
			state.Mode, state.DungeonX, state.DungeonY)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 11 ||
		state.SceneHeadBlock != 0xFF || !strings.Contains(state.Message, "皇家馬車") {
		t.Fatalf("royal carriage picture mode=%v picture=%v:%d head=%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.SceneHeadBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "讓路") {
		t.Fatalf("royal carriage first pause mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "青色印記") || !strings.Contains(state.Message, "迫使") {
		t.Fatalf("royal carriage compulsion message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "不是真正的國王") || !strings.Contains(state.Message, "又來了") {
		t.Fatalf("royal carriage false king message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "警鐘") || !strings.Contains(state.Message, "拔劍") {
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
		!strings.Contains(state.Message, "紅袍人") || !strings.Contains(state.Message, "小巷") {
		t.Fatalf("Royal Guard victory status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 2 || state.Choices[0] != "是" || state.Choices[1] != "否" ||
		!strings.Contains(state.Message, "投降") {
		t.Fatalf("surrender prompt choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if len(state.Choices) != 1 || !strings.Contains(state.Message, "牢房") {
		t.Fatalf("jail pause choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 2 ||
		state.SceneHeadBlock != 2 || state.SceneBodyBlock != 2 ||
		!strings.Contains(state.Message, "盜賊") || !strings.Contains(state.Message, "裝備") {
		t.Fatalf("thief rescue picture mode=%v picture=%v:%d head/body=%d/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock,
			state.SceneHeadBlock, state.SceneBodyBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 ||
		!strings.Contains(state.Message, "盜賊") {
		t.Fatalf("thief rescue pause mode=%v choices=%v message=%q",
			state.Mode, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "隱密通道") || !strings.Contains(state.Message, "盜賊公會") {
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
	state.SetDungeonGeometryView(geometryX, geometryY, geometryDirection)
	if state.DungeonX != 1 || state.DungeonY != 12 || state.DungeonDirection != 0 {
		t.Fatalf("guild geometry-to-script inverse=(%d,%d,%d), want (1,12,0)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
	state.DungeonWallType, _ = grid.WallWrapped(geometryX, geometryY, int(geometryDirection))
	state.DungeonWallRoof = grid.CellWrapped(geometryX, geometryY).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
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
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
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
		!strings.Contains(state.Message, "尤拉什") || !strings.Contains(state.Message, "手札第 4 條") {
		t.Fatalf("guild victory mode=%v status=%v message=%q choices=%v",
			state.Mode, state.CombatStatus(), state.Message, state.Choices)
	}
	foundJournal4 := false
	for _, page := range state.JournalPages {
		if strings.HasPrefix(page, "手札條目 4：") &&
			strings.Contains(page, "下水道地圖") && strings.Contains(page, "火刀據點") {
			foundJournal4 = true
			break
		}
	}
	if !foundJournal4 {
		t.Fatalf("Journal Entry 4 was not unlocked in-game: pages=%v", state.JournalPages)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.SetDungeonGeometryView(11, 7, 0)
	state.DungeonWallType, _ = grid.WallWrapped(11, 7, 0)
	state.DungeonWallRoof = grid.CellWrapped(11, 7).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "抱著豎琴的半身人") {
		t.Fatalf("guild halfling event mode=%v message=%q choices=%v", state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.SetDungeonGeometryView(12, 7, 0)
	state.DungeonWallType, _ = grid.WallWrapped(12, 7, 0)
	state.DungeonWallRoof = grid.CellWrapped(12, 7).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "放出了犬群") {
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
		!strings.Contains(state.Message, "被啃咬的骨頭") {
		t.Fatalf("guild kennel result mode=%v status=%v message=%q choices=%v",
			state.Mode, state.CombatStatus(), state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.SetDungeonGeometryView(15, 7, 0)
	state.DungeonWallType, _ = grid.WallWrapped(15, 7, 0)
	state.DungeonWallRoof = grid.CellWrapped(15, 7).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "關猴子的空籠") {
		t.Fatalf("guild cages mode=%v message=%q choices=%v", state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.SetDungeonGeometryView(11, 3, 0)
	state.DungeonWallType, _ = grid.WallWrapped(11, 3, 0)
	state.DungeonWallRoof = grid.CellWrapped(11, 3).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "奧莉芙・拉斯凱托") {
		t.Fatalf("guild guestbook mode=%v message=%q choices=%v", state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.SetDungeonGeometryView(10, 15, 4)
	state.DungeonWallType, _ = grid.WallWrapped(10, 15, 4)
	state.DungeonWallRoof = grid.CellWrapped(10, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !strings.Contains(state.Message, "綠色黏液痕跡") {
		t.Fatalf("guild sewer door mode=%v message=%q", state.Mode, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
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
		!strings.Contains(state.Message, "提爾佛頓") || !strings.Contains(state.Message, "難在這裡靈活作戰") {
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
	state.DungeonX, state.DungeonY, state.DungeonDirection = 1, 8, 2
	state.DungeonWallType, _ = sewerGrid.WallWrapped(1, 8, 2)
	state.DungeonWallRoof = sewerGrid.CellWrapped(1, 8).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "火刀要求你們立刻投降") ||
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
		!strings.Contains(state.Message, "火刀的屍體藏了起來") {
		t.Fatalf("sewer checkpoint status=%v mode=%v message=%q",
			state.CombatStatus(), state.Mode, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 13, 10, 2
	state.DungeonWallType, _ = sewerGrid.WallWrapped(13, 10, 2)
	state.DungeonWallRoof = sewerGrid.CellWrapped(13, 10).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "迷斯卓諾騎士團") {
		t.Fatalf("sewer knight first pause mode=%v terrain=%#x message=%q choices=%v",
			state.Mode, state.DungeonWallRoof, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "你們效忠誰") ||
		len(state.Choices) != 3 || state.Choices[1] != "娜卡西亞公主" {
		t.Fatalf("sewer knight allegiance mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || !strings.Contains(state.Message, "拿著戰鎚的牧師") ||
		!strings.Contains(state.Message, "讓你們通過") {
		t.Fatalf("sewer knight Princess branch mode=%v message=%q choices=%v",
			state.Mode, state.Message, state.Choices)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 13, 10, 2
	state.DungeonWallType, _ = sewerGrid.WallWrapped(13, 10, 2)
	state.DungeonWallRoof = sewerGrid.CellWrapped(13, 10).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.Message != "" || len(state.Choices) != 0 {
		t.Fatalf("sewer knight revisit mode=%v message=%q choices=%v, want consumed event",
			state.Mode, state.Message, state.Choices)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 8, 15, 4
	state.DungeonWallType, _ = sewerGrid.WallWrapped(8, 15, 4)
	state.DungeonWallRoof = sewerGrid.CellWrapped(8, 15).Terrain
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness || state.session.CurrentBlockID() != 4 ||
		state.GeoMapSet != 2 || state.GeoMapBlock != 4 ||
		state.DungeonX != 6 || state.DungeonY != 1 || state.DungeonDirection != 4 ||
		state.LoadPieces != [3]uint16{1, 2, 4} ||
		!strings.Contains(state.Message, "火刀據點") {
		t.Fatalf("Fire Knife hideout entry mode=%v block=%#x script=(%d,%d,%d) geo=%d/%d pieces=%v message=%q choices=%v",
			state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
			state.DungeonDirection, state.GeoMapSet, state.GeoMapBlock, state.LoadPieces,
			state.Message, state.Choices)
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
	if got := localizeECLText(testCatalog(), prompt.Text); !strings.Contains(got, "刀刃") ||
		!strings.Contains(got, "金屬嗡鳴") {
		t.Fatalf("localized blade barrier=%q", got)
	}
	if got := localizeECLText(testCatalog(), wait.Text); !strings.Contains(got, "逐漸放慢") ||
		!strings.Contains(got, "完全止息") {
		t.Fatalf("localized blade barrier aftermath=%q", got)
	}
	if got := []string{
		localizeOption(testCatalog(), "ENTER THE BLADES"),
		localizeOption(testCatalog(), "WAIT"),
		localizeOption(testCatalog(), "RETREAT"),
	}; strings.Join(got, "/") != "闖入刀刃/等待/撤退" {
		t.Fatalf("localized blade barrier choices=%v", got)
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
	if len(state.Choices) != 3 || state.Choices[0] != "闖入刀刃" {
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
		!strings.Contains(state.Message, "完全止息") {
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
	if got := []string{
		localizeOption(testCatalog(), "RETREAT"),
		localizeOption(testCatalog(), "INTERROGATE"),
		localizeOption(testCatalog(), "KILL"),
	}; strings.Join(got, "/") != "撤退/審問/殺死" {
		t.Fatalf("localized frozen-room choices=%v", got)
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
		!strings.Contains(state.Message, "交戰姿勢") {
		t.Fatalf("playable frozen room choices=%v message=%q", state.Choices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "有用情報") ||
		!strings.Contains(state.Message, "第 26 條") {
		t.Fatalf("playable frozen interrogation message=%q", state.Message)
	}
	foundJournal := false
	for _, page := range state.JournalPages {
		if strings.HasPrefix(page, "手札條目 26：") &&
			strings.Contains(page, "入侵牧師") &&
			strings.Contains(page, "南方首領房") {
			foundJournal = true
		}
	}
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

	state := NewStateFromECLBlocks(testCatalog(), map[uint8][]byte{4: hideout}, 4)
	if err := state.session.Reset(4); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.DungeonX, state.DungeonY, state.DungeonDirection = 14, 11, 0
	state.DungeonWallRoof = 0x9B
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "火刀") || !strings.Contains(state.Message, "辦公室") {
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
	foundJournal := false
	for _, page := range state.JournalPages {
		if strings.HasPrefix(page, "手札條目 9：") &&
			strings.Contains(page, "燃燒靈氣") &&
			strings.Contains(page, "光芒之池") {
			foundJournal = true
		}
	}
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
		localized     string
		continuations []uint16
	}{
		{0x9C, 0x4C11, "STRANGE SMOKY SCENT", "奇怪的煙味", []uint16{0}},
		{0x9D, 0x4C12, "UNSEEN SERVANTS", "看不見的僕人", []uint16{0}},
		{0x9E, 0x4C13, "CHARRED BODY", "焦屍", []uint16{0, 0}},
		{0x9F, 0x4C14, "NOTHING ESCAPED DESTRUCTION", "逃過毀滅", []uint16{0}},
		{0xA0, 0x4C15, "TWO ROWS OF SHROUDED BODIES", "待復活", []uint16{0}},
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
		if !strings.Contains(state.Message, test.localized) {
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
			foundJournal := false
			for _, page := range state.JournalPages {
				if strings.HasPrefix(page, "手札條目 29：") &&
					strings.Contains(page, "異次元力量") &&
					strings.Contains(page, "泰蘭索斯") {
					foundJournal = true
				}
			}
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
	state := NewState(testCatalog())
	state.session = session
	state.eclBlock = session.CurrentData()
	state.eclStart, err = session.InitialEntry()
	if err != nil {
		t.Fatal(err)
	}
	state.selectionSequence = []uint16{0}
	state.eclMenuReturnMode = ModeDungeon
	state.eventReturnMode = ModeDungeon
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄", HitPoints: 10, MaxHitPoints: 10}}
	state.applyECLTreasureSignals(encounter)
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
	for step := 0; step < 6; step++ {
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		}
		if state.Mode != ModeWilderness || len(state.Choices) == 0 {
			t.Fatalf("continuation step %d mode=%v picture=%v choices=%v message=%q",
				step, state.Mode, state.PictureRequested, state.currentOriginalChoices, state.Message)
		}
		if err := state.Select(0); err != nil {
			t.Fatal(err)
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
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Tilverton edge mode=%v choices=%#v message=%q", state.Mode, state.currentOriginalChoices, state.Message)
	}
	for address, want := range map[uint16]uint16{0x4CFF: 1, 0x4C2A: 1, 0x7F12: 1} {
		if got, ok := session.MemoryValue(address); !ok || got != want {
			t.Fatalf("memory[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}
	for _, entry := range []string{"手札條目 54", "手札條目 53"} {
		found := false
		for _, page := range state.JournalPages {
			found = found || strings.HasPrefix(page, entry)
		}
		if !found {
			t.Fatalf("%s was not unlocked: %v", entry, state.JournalPages)
		}
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
	if !strings.Contains(state.Message, "提爾隘口") ||
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
	for _, fighter := range fighters[1:] {
		if fighter.Name != "鷹馬" || fighter.SpriteBlock != 81 || fighter.Side != combat.SideEnemy {
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
		!strings.Contains(state.Message, "阿沙本福德城外") {
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
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 80 ||
		!strings.Contains(state.Message, "阿沙本福德") {
		t.Fatalf("Ashabenford entry mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
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
		!strings.Contains(state.Message, "河畔酒館") {
		t.Fatalf("Ashabenford ale house choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		!strings.Contains(state.Message, "兩艘") || !strings.Contains(state.Message, "暗影谷") {
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
	if !strings.Contains(state.Message, "偽裝成巡邏兵的火刀") ||
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
	if !strings.Contains(state.Message, "灰袍男子") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Standing Stone arrival choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"THANK HIM", "ATTACK", "LEAVE"}) ||
		!strings.Contains(state.Message, "四位主人") {
		t.Fatalf("Standing Stone counsel choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "南方") || !strings.Contains(state.Message, "紅色之人") ||
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
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ASHABENFORD", "ESSEMBRA", "HILLSFAR"}) {
		t.Fatalf("Standing Stone destinations=%#v", state.currentOriginalChoices)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) ||
		!strings.Contains(state.Message, "艾森布拉") {
		t.Fatalf("Essembra routes=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Location != LocationEssembra || !strings.Contains(state.Message, "艾森布拉城外") {
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
		!strings.Contains(state.Message, "哈普") {
		t.Fatalf("Hap routes=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "三隻") || !strings.Contains(state.Message, "黑龍") {
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
	for _, fighter := range dragons[1:] {
		if fighter.Name != "黑龍" || fighter.SpriteBlock != 0x35 || fighter.Side != combat.SideEnemy {
			t.Fatalf("Hap dragon=%+v", fighter)
		}
	}
	for turn := 0; turn < 16 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Location != LocationHap || !strings.Contains(state.Message, "哈普村外") ||
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
		state.GeoMapSet != 5 || !strings.Contains(state.Message, "破敗的村莊") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap entry block=%#x area=%+v geo=%d choices=%#v message=%q",
			session.CurrentBlockID(), state.Area, state.GeoMapSet, state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.Prompt != "哈普村　↑：前進　K／M：轉向　S：搜索　E：紮營" ||
		state.LoadPieces != [3]uint16{12, 0xFF, 0xFF} {
		t.Fatalf("Hap dungeon mode=%v prompt=%q pieces=%v", state.Mode, state.Prompt, state.LoadPieces)
	}
	state.DungeonWallRoof = 0x84
	session.SetMemoryValue(0x4BC9, 15)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 50 ||
		!strings.Contains(state.Message, "幾名村民驚慌退縮") {
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
	if !strings.Contains(state.Message, "村民奪門而出") ||
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
	state.eclSeed = 3
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
	if state.Mode != ModeCombat || !strings.Contains(state.Message, "死得痛快") {
		t.Fatalf("Hap patrol combat mode=%v message=%q", state.Mode, state.Message)
	}
	patrolFighters := state.CombatFighters()
	if len(patrolFighters) != 5 {
		t.Fatalf("Hap patrol fighters=%d, want hero, three fighters, and one mage", len(patrolFighters))
	}
	for _, fighter := range patrolFighters[1:4] {
		if fighter.Name != "黑暗精靈戰士" || fighter.SpriteBlock != 0x31 {
			t.Fatalf("Hap dark elf fighter=%+v", fighter)
		}
	}
	if patrolFighters[4].Name != "黑暗精靈法師" || patrolFighters[4].SpriteBlock != 0x32 {
		t.Fatalf("Hap dark elf mage=%+v", patrolFighters[4])
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
		!strings.Contains(state.Message, "阿卡巴・貝爾・阿卡什") {
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
		!strings.Contains(state.Message, "保持低調") {
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
	if !strings.Contains(state.Message, "伊弗利特") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap efreet barn choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) ||
		!strings.Contains(state.Message, "伊弗利特隆隆吼道") {
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
	if !strings.Contains(state.Message, "村莊與一處洞穴") ||
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
		!strings.Contains(state.Message, "整座村莊隨即充滿歡呼") {
		t.Fatalf("Hap liberation picture=%v/%d message=%q", state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "哈普圖斯永遠歡迎") {
		t.Fatalf("Hap elder thanks=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "附近法師塔控制") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap elder tower choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "祕密商路") ||
		!reflect.DeepEqual(state.currentOriginalChoices, []string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
		t.Fatalf("Hap Akabar secret route choices=%#v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("Hap liberation returned mode=%v, want dungeon", state.Mode)
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
