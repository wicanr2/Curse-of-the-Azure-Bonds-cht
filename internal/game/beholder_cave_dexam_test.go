package game

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

func TestRealBeholderCaveDexamAndZhentilBattles(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
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
	monsterBlocks, err := dax.Parse(zipData(t, image, "MON4CHA.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	records := make(map[uint8]monster.Record, len(monsterBlocks))
	for _, block := range monsterBlocks {
		record, parseErr := monster.Parse(block.Data)
		if parseErr == nil {
			records[block.Entry.ID] = record
		}
	}
	geometry := geo.NewCatalog()
	if err := geometry.AddDAX(4, zipData(t, image, "GEO4.DAX")); err != nil {
		t.Fatal(err)
	}
	grid, ok := geometry.Lookup(geo.MapRef{Set: 4, BlockID: 0x25})
	if !ok {
		t.Fatal("GEO4/25 missing")
	}

	state := NewStateFromECLBlocks(testCatalog(), all, 0x22)
	// Carry the Dimswart escort state into this focused cave regression.
	state.session.SetMemoryValue(0x4C01, 1)
	state.SetMonsterRecordsForECL(4, records)
	testParty := make([]combat.Fighter, 6)
	for index := range testParty {
		testParty[index] = combat.Fighter{
			ID: fmt.Sprintf("hero-%d", index+1), Name: fmt.Sprintf("英雄%d", index+1),
			Side: combat.SideParty, HitPoints: 1000, MaxHitPoints: 1000,
			ArmorClass: -20, AttackBonus: 30, DamageDiceCount: 8, DamageDiceSides: 20,
		}
	}
	if err := state.SetParty(testParty); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.Area.GameArea, state.Area.InDungeon = 4, true
	state.GeoMapSet, state.GeoMapBlock = 4, 0x25
	state.DungeonX, state.DungeonY, state.DungeonDirection = 15, 1, 0
	state.DungeonWallType, _ = grid.WallWrapped(15, 1, 0)
	state.DungeonWallRoof = grid.CellWrapped(15, 1).Terrain

	if state.DungeonWallRoof != 0x90 {
		t.Fatalf("Dexam trigger terrain=%02x, want 90", state.DungeonWallRoof)
	}
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 49 ||
		state.Message != requireGamePackText(t, &state, "dexam.final_reveal") {
		t.Fatalf("Dexam reveal mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
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
	if !state.CombatActive() || state.Message != requireGamePackText(t, &state, "dexam.attack") {
		t.Fatalf("Dexam battle active=%v message=%q", state.CombatActive(), state.Message)
	}
	assertCombatMonsterCounts(t, state.CombatFighters(), map[string]int{
		"monster-27-": 1,
		"monster-28-": 1,
		"monster-29-": 10,
	})

	for turn := 0; turn < 500 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode == ModeCombat {
		t.Fatalf("Dexam battle did not finish after 500 turns: fighters=%+v", state.CombatFighters())
	}
	if !state.treasureMenu || len(state.PendingTreasureItems()) != 4 ||
		len(state.currentOriginalChoices) != 5 ||
		state.currentOriginalChoices[4] != "TREASURE_EXIT" {
		t.Fatalf("Dexam loot mode=%v choices=%q items=%+v",
			state.Mode, state.currentOriginalChoices, state.PendingTreasureItems())
	}
	if err := state.Select(4); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, &state, "dexam.amulet_retrieved") {
		t.Fatalf("amulet retrieval message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !state.CombatActive() || state.Message != requireGamePackText(t, &state, "dexam.zhentil_attack") {
		t.Fatalf("Zhentil battle active=%v message=%q", state.CombatActive(), state.Message)
	}
	assertCombatMonsterCounts(t, state.CombatFighters(), map[string]int{
		"monster-20-": 11,
		"monster-21-": 4,
		"monster-22-": 3,
		"monster-48-": 1,
	})

	for turn := 0; turn < 500 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatal(err)
		}
	}
	if state.Mode == ModeCombat {
		t.Fatalf("Zhentil battle did not finish after 500 turns: fighters=%+v", state.CombatFighters())
	}
	if state.CombatStatus() != combat.StatusPartyWon || state.Mode != ModeEvent {
		t.Fatalf("Zhentil victory mode=%v status=%v message=%q",
			state.Mode, state.CombatStatus(), state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x22 ||
		state.GeoMapSet != 4 || state.GeoMapBlock != 0x25 ||
		state.DungeonX != 15 || state.DungeonY != 1 {
		t.Fatalf("post-battle cave mode=%v ecl=%02x geo=%d/%02x coords=%d,%d",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY)
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 6, 3, 2
	state.DungeonWallType, _ = grid.WallWrapped(6, 3, 2)
	state.DungeonWallRoof = grid.CellWrapped(6, 3).Terrain
	if err := state.RunDungeonExitLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.DungeonWallRoof != 0x93 || state.Mode != ModeEvent ||
		!state.PictureRequested || state.PictureBlock != 42 ||
		state.Message != requireGamePackText(t, &state, "dexam.departure.olive") {
		t.Fatalf("cave exit terrain=%02x mode=%v picture=%v/%d message=%q",
			state.DungeonWallRoof, state.Mode, state.PictureRequested,
			state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	for _, messageID := range []string{
		"dexam.departure.dimswart",
		"dexam.departure.gharri",
		"dexam.departure.riders",
	} {
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if state.Message != requireGamePackText(t, &state, messageID) {
			t.Fatalf("departure stage %s message=%q", messageID, state.Message)
		}
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || state.session.CurrentBlockID() != 0x51 ||
		state.Area.InDungeon || state.Area.GameArea != 1 ||
		state.Message != requireGamePackText(t, &state, "world.shadowdale.edge") {
		t.Fatalf("world return mode=%v block=%02x area=%+v message=%q",
			state.Mode, state.session.CurrentBlockID(), state.Area, state.Message)
	}
	for address, want := range map[uint16]uint16{0x4CE2: 1, 0x7F12: 1} {
		if got, ok := state.session.MemoryValue(address); !ok || got != want {
			t.Fatalf("departure flag %04x=%04x ok=%v, want %04x", address, got, ok, want)
		}
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeWilderness ||
		strings.Join(state.currentOriginalChoices, "/") != "ENTER CITY/JOURNEY ON/CAMP/SEARCH AREA" {
		t.Fatalf("Shadowdale world menu mode=%v originals=%q message=%q",
			state.Mode, state.currentOriginalChoices, state.Message)
	}
}

func assertCombatMonsterCounts(t *testing.T, fighters []combat.Fighter, want map[string]int) {
	t.Helper()
	got := make(map[string]int, len(want))
	for _, fighter := range fighters {
		for prefix := range want {
			if strings.HasPrefix(fighter.ID, prefix) {
				got[prefix]++
			}
		}
	}
	for prefix, count := range want {
		if got[prefix] != count {
			t.Fatalf("combat count %s=%d, want %d; fighters=%+v",
				prefix, got[prefix], count, fighters)
		}
	}
}
