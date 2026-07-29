package game

import (
	"archive/zip"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// TestRealZhentilOliveSecretPassage preserves the first main-story path after
// entering Zhentil Keep: meet Olive, read Journals 50/51 and follow her from
// ECL4/GEO4 block 0x20 into the Dark Shrine at block 0x21.
func TestRealZhentilOliveSecretPassage(t *testing.T) {
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
	grid, ok := geometry.Lookup(geo.MapRef{Set: 4, BlockID: 0x20})
	if !ok {
		t.Fatal("GEO4/20 missing")
	}

	state := NewStateFromECLBlocks(testCatalog(), all, 0x20)
	state.SetMonsterRecordsForECL(4, records)
	if err := state.SetParty([]combat.Fighter{{
		ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 60, MaxHitPoints: 60, ArmorClass: 0,
		AttackBonus: 20, DamageDiceCount: 4, DamageDiceSides: 12,
	}}); err != nil {
		t.Fatal(err)
	}
	state.Mode = ModeDungeon
	state.Area.GameArea = 4
	state.Area.InDungeon = true
	state.GeoMapSet, state.GeoMapBlock = 4, 0x20
	state.DungeonX, state.DungeonY, state.DungeonDirection = 10, 11, 0
	state.DungeonWallType, _ = grid.WallWrapped(10, 11, 0)
	state.DungeonWallRoof = grid.CellWrapped(10, 11).Terrain

	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeEvent || !state.PictureRequested || state.PictureBlock != 42 ||
		!strings.Contains(state.Message, "半身人女子從隱密壁龕中現身") {
		t.Fatalf("Olive meeting mode=%v picture=%v/%d message=%q",
			state.Mode, state.PictureRequested, state.PictureBlock, state.Message)
	}
	journals := strings.Join(state.JournalPages, "\n")
	if !strings.Contains(journals, "手札條目 50（1/2）") ||
		!strings.Contains(journals, "洛山達護符") ||
		!strings.Contains(journals, "迪姆斯沃特") {
		t.Fatalf("Journal 50 was not unlocked as readable Chinese pages: %q", journals)
	}

	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(state.Choices, "/"); got != "是/否" ||
		!strings.Contains(state.Message, "跟她走") {
		t.Fatalf("follow prompt choices=%q message=%q", state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if state.session.CurrentBlockID() != 0x21 || !state.PictureRequested ||
		state.PictureBlock != 42 || !strings.Contains(state.Message, "穿牆進入幽暗神殿") {
		t.Fatalf("shrine transition block=0x%02x picture=%v/%d message=%q",
			state.session.CurrentBlockID(), state.PictureRequested, state.PictureBlock, state.Message)
	}

	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "手札第 51 條") {
		t.Fatalf("Olive explanation message=%q", state.Message)
	}
	journals = strings.Join(state.JournalPages, "\n")
	if !strings.Contains(journals, "手札條目 51（1/2）") ||
		!strings.Contains(journals, "弗佐爾・錢布瑞爾") ||
		!strings.Contains(journals, "貝恩祭司") {
		t.Fatalf("Journal 51 was not unlocked as readable Chinese pages: %q", journals)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "門的另一頭") {
		t.Fatalf("Dimswart door message=%q", state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.Message, "微微一笑") {
		t.Fatalf("Olive departure message=%q", state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x21 ||
		state.GeoMapSet != 4 || state.GeoMapBlock != 0x21 {
		t.Fatalf("final mode=%v block=0x%02x geo=%d/%d coords=%d,%d,%d message=%q",
			state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection, state.Message)
	}
	t.Logf("Olive path returns at GEO4/21 (%d,%d,%d)",
		state.DungeonX, state.DungeonY, state.DungeonDirection)
}
