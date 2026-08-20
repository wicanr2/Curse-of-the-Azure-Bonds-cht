package game

import (
	"archive/zip"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// 墓園每一格的事件由地形碼分派（`ON GOTO (C04F & 0x7F)`，見 spec 1142 的同一個
// 機制）。這一條逐格站上去、跑一次生命週期，驗那一格的敘述是 game pack 的中文。
//
// ⚠ 每一格都**重新進段**：once-only 旗標與戰鬥結果會互相污染，接續著跑會讓
// 後面的格子看起來沒內容。這一條驗的是**內容與語系**，不是可達性——
// 走得到走不到由 `TestRealNewGameRunsToTheEnding` 的墓園段負責。
func TestBurialGlenCellEventsAreLocalized(t *testing.T) {
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks := map[uint8][]byte{}
	for chapter := 1; chapter <= 6; chapter++ {
		parsed, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range parsed {
			blocks[block.Entry.ID] = block.Data
		}
	}
	records := map[uint8]map[uint8]monster.Record{}
	for chapter := 1; chapter <= 6; chapter++ {
		parsed, parseErr := dax.Parse(zipData(t, image, "MON"+strconv.Itoa(chapter)+"CHA.DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		chapterRecords := map[uint8]monster.Record{}
		for _, block := range parsed {
			record, recordErr := monster.Parse(block.Data)
			if recordErr != nil {
				t.Fatalf("MON%dCHA block %#02x: %v", chapter, block.Entry.ID, recordErr)
			}
			chapterRecords[block.Entry.ID] = record
		}
		records[uint8(chapter)] = chapterRecords
	}
	grid := loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x40)
	glen, ok := segment.Lookup("ECL6/0x40")
	if !ok {
		t.Fatal("註冊表沒有 ECL6/0x40")
	}

	cells := []struct {
		x, y      int
		messageID string
		// random 為真代表那一格的事件由 `RANDOM` 擋著（盜墓者是 100 抽 32），
		// 站上去一次不一定演；重進段重抽，不是放寬條件。
		random bool
	}{
		{6, 14, "myth-drannor.red-web", false},
		{13, 14, "myth-drannor.daemir.offer", false},
		{10, 4, "myth-drannor.grave.thri-kreen", true},
		{12, 8, "myth-drannor.phase-spider-wall", false},
		{14, 8, "myth-drannor.phase-spider-glowing", false},
		{14, 10, "myth-drannor.phase-spider-bones", false},
		{3, 13, "myth-drannor.elf-spirit.greeting", false},
		{10, 7, "myth-drannor.thri-kreen.entrance", false},
		{9, 9, "myth-drannor.thri-kreen.guards", false},
		{7, 9, "myth-drannor.thri-kreen.bivouac", false},
		{9, 2, "myth-drannor.spider-mausoleum", false},
		{10, 1, "myth-drannor.spider-funnel", false},
		{4, 8, "myth-drannor.clan-figure.greeting", false},
		{13, 6, "myth-drannor.red-plume.gesture", false},
		{3, 1, "myth-drannor.court.armor", false},
	}
	for _, cell := range cells {
		t.Run(cell.messageID, func(t *testing.T) {
			attempts := 1
			if cell.random {
				attempts = 24
			}
			var last string
			for attempt := 0; attempt < attempts; attempt++ {
				state := newBurialGlenState(t, blocks, records, glen)
				state.SetECLSeed(int64(attempt) + 1)
				state.SetDungeonGeometryView(cell.x, cell.y, 0)
				state.DungeonWallRoof = grid.CellWrapped(cell.x, cell.y).Terrain
				if err := state.RunDungeonLifecycle(); err != nil {
					t.Fatalf("(%d,%d) 的生命週期：%v", cell.x, cell.y, err)
				}
				// 遭遇選單的敘述走的是 Prompt 那一行，不是 Message。
				want := requireGamePackText(t, &state, cell.messageID)
				if state.Message == want || state.Prompt == want {
					return
				}
				last = state.Message + "／" + state.Prompt
			}
			t.Fatalf("(%d,%d) 試了 %d 次，最後一次的敘述／提示是 %q",
				cell.x, cell.y, attempts, last)
		})
	}
}

func newBurialGlenState(t *testing.T, blocks map[uint8][]byte,
	records map[uint8]map[uint8]monster.Record, glen segment.Segment) State {
	t.Helper()
	state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, 0x50)
	for chapter, chapterRecords := range records {
		state.SetMonsterRecordsForECL(chapter, chapterRecords)
	}
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.AddCreationCharacter(0); err != nil {
		t.Fatal(err)
	}
	if err := state.FinishCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.EnterSegment(glen); err != nil {
		t.Fatal(err)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("進墓園之後模式是 %v", state.Mode)
	}
	return state
}
