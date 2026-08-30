package game

import (
	"archive/zip"
	"strconv"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// 原作離開墓園時先跨出 GEO 邊界，再由 ECL6/0x40 的 ON GOTO C04D 依面向
// 分派：東邊進外城，其餘方向回世界。所有 GEO 真正可跨出的周界邊都必須宣告，
// 否則 wrap=true 會讓正常按鍵靜默環繞回圖內；主線直接呼叫 lifecycle 曾掩蓋此洞。
func TestBurialGlenDeclaresEveryPassableBoundaryExit(t *testing.T) {
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)
	grid := loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x40)
	glen, ok := segment.Lookup("ECL6/0x40")
	if !ok {
		t.Fatal("註冊表沒有 ECL6/0x40")
	}
	state := newSegmentDungeonState(t, blocks, records, glen)

	type edge struct{ x, y, direction int }
	var edges []edge
	for offset := 0; offset < geo.Width; offset++ {
		edges = append(edges,
			edge{offset, 0, 0}, edge{geo.Width - 1, offset, 2},
			edge{offset, geo.Height - 1, 4}, edge{0, offset, 6})
	}
	for _, candidate := range edges {
		want := grid.CanMoveDungeonWrapped(candidate.x, candidate.y, candidate.direction)
		_, got := state.dungeonExternalExit(candidate.x, candidate.y, uint8(candidate.direction))
		if got != want {
			t.Errorf("boundary (%d,%d,%d) declared=%v, GEO passable=%v",
				candidate.x, candidate.y, candidate.direction, got, want)
		}
	}
}

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
	blocks, records := loadAllECLAndMonsters(t, image)
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
		// ⚠ 遭遇選單的旁白依**距離**挑一句（spec 1144）。這一處的距離上限是 2，
		// 所以站上去看到的是遠距那一句；`red-plume.gesture` 是距離 0 才演的。
		{13, 6, "myth-drannor.lone-red-plume-spotted", false},
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
				state := newSegmentDungeonState(t, blocks, records, glen)
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

func newSegmentDungeonState(t *testing.T, blocks map[uint8][]byte,
	records map[uint8]map[uint8]monster.Record, seg segment.Segment) State {
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
	if err := state.EnterSegment(seg); err != nil {
		t.Fatal(err)
	}
	// ⚠ 盤點用的隊伍一律撐起來：有的段一進去就開打（古熔岩洞的伏擊），
	// 臨時建的一名角色會死在入口，後面一格都盤點不到。**只給盤點用**——
	// 段測試不准這樣做（見 boostSweepParty）。
	boostSweepParty(t, &state)
	// 段的入口不一定停在同一種模式：有的要按繼續，有的停在選單上，有的開打。
	for step := 0; step < 16 && state.Mode != ModeDungeon; step++ {
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					t.Fatalf("進 %s 的入口戰鬥：%v", seg.ID, err)
				}
			}
			continue
		}
		if err := state.Continue(); err != nil {
			if selectErr := state.Select(0); selectErr != nil {
				t.Fatalf("進 %s 的入口推不動：continue=%v select=%v", seg.ID, err, selectErr)
			}
		}
	}
	if state.Mode != ModeDungeon {
		t.Fatalf("進 %s 之後模式是 %v", seg.ID, state.Mode)
	}
	return state
}

// loadAllECLAndMonsters 讀六個 ECL 成員與六章的怪物表。逐格盤點要重複進段，
// 少載一章的怪物表那一段就會安靜地跳過開戰。
func loadAllECLAndMonsters(t *testing.T, image *zip.ReadCloser) (
	map[uint8][]byte, map[uint8]map[uint8]monster.Record) {
	t.Helper()
	blocks := map[uint8][]byte{}
	for chapter := 1; chapter <= 6; chapter++ {
		parsed, err := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range parsed {
			blocks[block.Entry.ID] = block.Data
		}
	}
	records := map[uint8]map[uint8]monster.Record{}
	for chapter := 1; chapter <= 6; chapter++ {
		parsed, err := dax.Parse(zipData(t, image, "MON"+strconv.Itoa(chapter)+"CHA.DAX"))
		if err != nil {
			t.Fatal(err)
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
	return blocks, records
}

// 外城遺跡的每格事件同樣由地形碼分派（`ON GOTO (C04F & 0x7F)`，位移 0x0219）。
func TestOuterRuinsCellEventsAreLocalized(t *testing.T) {
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)
	grid := loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x42)
	outer, ok := segment.Lookup("ECL6/0x42")
	if !ok {
		t.Fatal("註冊表沒有 ECL6/0x42")
	}

	cells := []struct {
		x, y      int
		messageID string
	}{
		{1, 12, "myth-drannor.outer.tirsheya.greeting"},
		{3, 14, "myth-drannor.outer.storehouse.guards"},
		{2, 6, "myth-drannor.outer.fugitive.intro"},
		{3, 6, "myth-drannor.outer.fugitive.remains"},
		{13, 7, "myth-drannor.outer.nameless-warning"},
		{12, 9, "myth-drannor.outer.brush-decoy"},
		{11, 11, "myth-drannor.outer.brush-bloodstains"},
		{1, 3, "myth-drannor.outer.gambling-room"},
		{11, 2, "myth-drannor.outer.margoyle-trap"},
		{5, 6, "myth-drannor.outer.sewer-margoyle"},
		{9, 6, "myth-drannor.outer.rakshasa-residence"},
	}
	for _, cell := range cells {
		t.Run(cell.messageID, func(t *testing.T) {
			state := newSegmentDungeonState(t, blocks, records, outer)
			state.SetDungeonGeometryView(cell.x, cell.y, 0)
			state.DungeonWallRoof = grid.CellWrapped(cell.x, cell.y).Terrain
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatalf("(%d,%d) 的生命週期：%v", cell.x, cell.y, err)
			}
			want := requireGamePackText(t, &state, cell.messageID)
			if state.Message != want && state.Prompt != want {
				t.Fatalf("(%d,%d) roof=%#02x 的敘述是 %q，提示是 %q",
					cell.x, cell.y, state.DungeonWallRoof, state.Message, state.Prompt)
			}
		})
	}
}

// 藏匿處不是走過去就有：守衛是 `COMPARE AND 7ECA,1 4CD5,1`——要**先拿到垂死者
// 的線索**（`4CD5`），而且要在那一格**SEARCH**（`7ECA`）。這一條把整段跑一次。
func TestOuterRuinsFugitiveCache(t *testing.T) {
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)
	grid := loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x42)
	outer, ok := segment.Lookup("ECL6/0x42")
	if !ok {
		t.Fatal("註冊表沒有 ECL6/0x42")
	}
	state := newSegmentDungeonState(t, blocks, records, outer)

	state.SetDungeonGeometryView(2, 6, 0)
	state.DungeonWallRoof = grid.CellWrapped(2, 6).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatalf("垂死者那一格：%v", err)
	}
	if want := requireGamePackText(t, &state, "myth-drannor.outer.fugitive.intro"); state.Message != want {
		t.Fatalf("垂死者的敘述是 %q", state.Message)
	}
	for step := 0; step < 16; step++ {
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					t.Fatalf("救人的戰鬥：%v", err)
				}
			}
			continue
		}
		if state.Mode == ModeDungeon {
			break
		}
		if err := state.Continue(); err != nil {
			if selectErr := state.Select(0); selectErr != nil {
				t.Fatalf("救人的流程推不動：continue=%v select=%v", err, selectErr)
			}
		}
	}
	if flag, _ := state.session.MemoryValue(0x4CD5); flag != 1 {
		t.Fatalf("救完人之後 4CD5=%#x，沒有拿到線索", flag)
	}

	// 拿到線索之後到那一格 SEARCH，藏匿處才出得來。
	state.SetDungeonGeometryView(14, 3, 0)
	state.DungeonWallRoof = grid.CellWrapped(14, 3).Terrain
	state.DungeonSearchEnabled = true
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatalf("藏匿處那一格：%v", err)
	}
	if want := requireGamePackText(t, &state, "myth-drannor.outer.fugitive.cache"); state.Message != want {
		t.Fatalf("藏匿處的敘述是 %q", state.Message)
	}
}

// boostSweepParty 把隊伍撐到足以走完內容盤點的程度。**只給盤點用**：
// 段測試不准這樣做，那會把「這一段打得贏嗎」偷偷換成「這一段演得出來嗎」。
func boostSweepParty(t *testing.T, state *State) {
	t.Helper()
	party := state.PartyFighters()
	if len(party) == 0 {
		t.Fatal("盤點用的隊伍是空的")
	}
	for index := range party {
		party[index].HitPoints, party[index].MaxHitPoints = 999, 999
		party[index].ArmorClass = -10
		party[index].AttackBonus = 100
		party[index].DamageDiceCount, party[index].DamageDiceSides = 1, 1
		party[index].DamageBonus = 100
		party[index].AttacksPerTurn = 8
		party[index].InitiativeBonus = 100
	}
	if err := state.SetParty(party); err != nil {
		t.Fatalf("裝上盤點用隊伍：%v", err)
	}
}

// 古熔岩洞的每格事件同樣由地形碼分派（遮罩 `0x7F`）。對照表由
// `cmd/ecl-cell-events` 產生，在 `docs/audit/ecl-cell-events.md`。
func TestLavaTubeCellEventsAreLocalized(t *testing.T) {
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)
	grid := loadGeoCampaignGrid(t, image, 5, "GEO5.DAX", 0x32)
	cave, ok := segment.Lookup("ECL5/0x32")
	if !ok {
		t.Fatal("註冊表沒有 ECL5/0x32")
	}

	cells := []struct {
		x, y      int
		messageID string
	}{
		{2, 0, "lava-tube.arrow-in-ash"},
		{13, 5, "lava-tube.stone-arrow-south"},
		{4, 0, "lava-tube.four-dark-elves"},
		{0, 3, "lava-tube.violated-our-precinct"},
		{14, 2, "lava-tube.barracks-disturbed"},
		{8, 4, "lava-tube.incense-room"},
		{0, 7, "lava-tube.mage-meditation-room"},
		{4, 2, "lava-tube.guarded-door"},
		{7, 12, "lava-tube.crimdrac-introduces"},
		{13, 15, "lava-tube.guarded-by-efreeti"},
		{15, 13, "lava-tube.tunnel-collapsed"},
	}
	for _, cell := range cells {
		t.Run(cell.messageID, func(t *testing.T) {
			// ⚠ 古熔岩洞有一半的格子要**搜尋**才會演（守衛讀 `7ECA`），
			// 所以先不搜尋跑一次，沒有再搜尋跑一次。
			for _, search := range []bool{false, true} {
				state := newSegmentDungeonState(t, blocks, records, cave)
				state.SetDungeonGeometryView(cell.x, cell.y, 0)
				state.DungeonWallRoof = grid.CellWrapped(cell.x, cell.y).Terrain
				state.DungeonSearchEnabled = search
				if err := state.RunDungeonLifecycle(); err != nil {
					t.Fatalf("(%d,%d) 的生命週期：%v", cell.x, cell.y, err)
				}
				want := requireGamePackText(t, &state, cell.messageID)
				if state.Message == want || state.Prompt == want {
					return
				}
				if search {
					t.Fatalf("(%d,%d) roof=%#02x 搜尋前後都沒出現；最後的敘述是 %q，提示是 %q",
						cell.x, cell.y, state.DungeonWallRoof, state.Message, state.Prompt)
				}
			}
		})
	}
}
