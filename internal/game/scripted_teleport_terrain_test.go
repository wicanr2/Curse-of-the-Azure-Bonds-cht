package game

import (
	"archive/zip"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// 原作的 `2Dh CALL 2E10h` 重畫是**自己去地圖讀**目前這一格的牆面與地形，寫進
// `C04E`／`C04F`；每格事件的分派器 (`AND C04F, <遮罩>` ＋ `ON GOTO`) 就靠
// `C04F`。所以腳本傳送（只寫 `C04B`／`C04C`）之後那兩格一定要跟著換
// （spec 1161）。
//
// 這一條兩面都釘：有寫座標的重畫要換地形碼，沒寫座標的重畫不能動它。
// 只釘一邊的話「每次重畫都重讀地形」與「永遠不重讀」各會過一邊。
func TestRedrawRefreshesTerrainOnlyWhenTheScriptMovedTheParty(t *testing.T) {
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	catalog := geo.NewCatalog()
	if err := catalog.AddDAX(2, zipData(t, image, "GEO2.DAX")); err != nil {
		t.Fatal(err)
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: 2, BlockID: 1})
	if !ok {
		t.Fatal("GEO2 block 1 is absent")
	}
	// 提爾弗頓 `(5,2)` 是訓練所（地形碼 0x8C），`(5,3)` 是招牌格（0x0B）。
	from, to := grid.CellWrapped(5, 2).Terrain, grid.CellWrapped(5, 3).Terrain
	if from == to {
		t.Fatalf("fixture needs two different terrains, both are %#x", from)
	}

	for _, probe := range []struct {
		name      string
		writes    []ecl.MemoryWrite
		wantX     int
		wantY     int
		wantRoof  uint8
		reasoning string
	}{
		{
			name: "腳本搬了隊伍",
			writes: []ecl.MemoryWrite{
				{Address: 0xC04B, Value: 5, PC: 0x1000, BlockID: 0x01, Sequence: 1},
				{Address: 0xC04C, Value: 3, PC: 0x1006, BlockID: 0x01, Sequence: 2},
			},
			wantX: 5, wantY: 3, wantRoof: to,
			reasoning: "重畫要把 (5,3) 的地形碼讀回來",
		},
		{
			name:   "腳本沒有寫座標",
			writes: nil,
			wantX:  5, wantY: 2, wantRoof: from,
			reasoning: "沒有位移就不重讀，地形碼維持 (5,2)",
		},
	} {
		t.Run(probe.name, func(t *testing.T) {
			session, err := ecl.NewBlockSession(map[uint8][]byte{0x01: {0, 0}}, 0x01)
			if err != nil {
				t.Fatal(err)
			}
			state := State{session: session, DungeonX: 5, DungeonY: 2, DungeonDirection: 4}
			state.Area.InDungeon = true
			state.Area.GameArea = 2
			state.GeoMapSet, state.GeoMapBlock = 2, 1
			state.SetGeoCatalog(catalog)
			state.DungeonWallRoof = from
			state.applyECLCallSignals(ecl.RunResult{
				CallAddresses:        []uint16{0x2E10},
				SessionStartBlockID:  0x01,
				SessionEndBlockID:    0x01,
				SessionBlockRangeSet: true,
				CallRequests: []ecl.CallRequest{
					{Address: 0x2E10, PC: 0x100C, BlockID: 0x01, Sequence: 3},
				},
				SaveWrites: probe.writes,
			})
			if state.DungeonX != probe.wantX || state.DungeonY != probe.wantY ||
				state.DungeonWallRoof != probe.wantRoof {
				t.Fatalf("%s：位置=(%d,%d) 地形碼=%#x，要 (%d,%d)／%#x",
					probe.reasoning, state.DungeonX, state.DungeonY, state.DungeonWallRoof,
					probe.wantX, probe.wantY, probe.wantRoof)
			}
			roof, ok := session.MemoryValue(0xC04F)
			if probe.writes == nil {
				return
			}
			if !ok || uint8(roof) != probe.wantRoof {
				t.Fatalf("C04F=%#x,%v，要 %#x", roof, ok, probe.wantRoof)
			}
		})
	}
}
