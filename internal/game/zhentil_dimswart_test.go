package game

import (
	"archive/zip"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

func TestRealZhentilRecruitDimswart(t *testing.T) {
	state, grid := newRealZhentilShrineState(t)

	setZhentilShrineCell(state, grid, 6, 13, 0)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 42 ||
		state.Message != requireGamePackText(t, state, "zhentil.olive_cell_hint") {
		t.Fatalf("Olive direction mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Message != requireGamePackText(t, state, "zhentil.olive_repeats_dimswart") {
		t.Fatalf("Olive repeated explanation message=%q", state.Message)
	}
	for step := 0; step < 8 && state.Mode != ModeDungeon; step++ {
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
	if state.Mode != ModeDungeon {
		t.Fatalf("Olive direction did not return to dungeon: mode=%v message=%q", state.Mode, state.Message)
	}

	setZhentilShrineCell(state, grid, 2, 14, 0)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 35 ||
		state.Message != requireGamePackText(t, state, "zhentil.dimswart_appears") {
		t.Fatalf("Dimswart meeting mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	for _, id := range []string{
		"journal.12.1", "journal.12.2", "journal.12.3",
		"journal.12.4", "journal.12.5", "journal.12.6",
	} {
		if !slices.Contains(state.JournalPages, requireGamePackText(t, state, id)) {
			t.Fatalf("Journal 12 missing %s: %q", id, state.JournalPages)
		}
	}

	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(state.Choices, "/"); got != "是/否" ||
		state.Message != requireGamePackText(t, state, "zhentil.dimswart_join") {
		t.Fatalf("Dimswart choice=%q message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.GeoMapSet != 4 || state.GeoMapBlock != 0x21 {
		t.Fatalf("Dimswart acceptance mode=%v geo=%d/%d message=%q",
			state.Mode, state.GeoMapSet, state.GeoMapBlock, state.Message)
	}
	if fighters := state.PartyFighters(); len(fighters) != 1 || fighters[0].ID != "hero" {
		t.Fatalf("Dimswart must remain a story escort, not consume a fighter slot: %+v", fighters)
	}

	// The original ECL flag makes the accepted encounter one-shot. Re-entering
	// the cell must not replay Journal 12 or ask the recruitment question.
	setZhentilShrineCell(state, grid, 2, 14, 0)
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.PictureRequested || state.Message != "" {
		t.Fatalf("Dimswart cell replayed after acceptance: mode=%v picture=%v message=%q",
			state.Mode, state.PictureRequested, state.Message)
	}
}

func newRealZhentilShrineState(t *testing.T) (*State, geo.Grid) {
	t.Helper()
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	t.Cleanup(func() { image.Close() })

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
	grid, ok := geometry.Lookup(geo.MapRef{Set: 4, BlockID: 0x21})
	if !ok {
		t.Fatal("GEO4/21 missing")
	}

	state := NewStateFromECLBlocks(testCatalog(), all, 0x21)
	state.SetMonsterRecordsForECL(4, records)
	if err := state.SetParty([]combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 60, MaxHitPoints: 60, ArmorClass: 0,
		AttackBonus: 20, DamageDiceCount: 4, DamageDiceSides: 12,
	}}); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.Area.GameArea, state.Area.InDungeon = 4, true
	state.GeoMapSet, state.GeoMapBlock = 4, 0x21
	return &state, grid
}

func setZhentilShrineCell(state *State, grid geo.Grid, x, y int, direction uint8) {
	state.Mode = ModeDungeon
	state.DungeonX, state.DungeonY, state.DungeonDirection = x, y, direction
	state.DungeonWallType, _ = grid.WallWrapped(x, y, int(direction))
	state.DungeonWallRoof = grid.CellWrapped(x, y).Terrain
}
