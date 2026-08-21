package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// dungeonMoveFixture 借用真實的 GEO4／0x21 地城，找一格「連續往南走兩步都通」
// 的位置站上去。用真實地圖而不是自造格子，是因為 MoveDungeon 的牆面判定
// 與外部出口判定都吃真實資料。
func dungeonMoveFixture(t *testing.T) (*State, geo.Grid) {
	t.Helper()
	state, grid := newRealZhentilShrineState(t)
	for y := 0; y+2 < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			if !grid.CanMoveDungeonWrapped(x, y, 4) || !grid.CanMoveDungeonWrapped(x, y+1, 4) {
				continue
			}
			setZhentilShrineCell(state, grid, x, y, 4)
			return state, grid
		}
	}
	t.Fatal("GEO4／0x21 裡找不到連續往南兩步都通的格子")
	return nil, geo.Grid{}
}

// TestMoveDungeonSnapshotsPreMoveCoordinates 釘住 ECL 格 `4BF0`／`4BF1` 的
// producer：原作的地城主迴圈在呼叫 `MOVEPARTY` **之前**把當時的座標存進
// `bank0^[1E0h]`／`[1E2h]`（DOS `overlay-02:3984h`，spec 1045），而 spec 1098
// 的區 0 換算把那兩個 word 對回 `4BF0`／`4BF1`（spec 1155）。
//
// ⚠ 存的是**移動前**的座標。存成移動後的話，腳本那 15 處
// `SAVE 4BF0 C04B; SAVE 4BF1 C04C`（把隊伍退回上一格）就等於什麼都沒做，
// 而且測試會全綠——所以這裡要同時檢查「不等於移動後的座標」。
func TestMoveDungeonSnapshotsPreMoveCoordinates(t *testing.T) {
	state, grid := dungeonMoveFixture(t)
	beforeX, beforeY, _ := state.DungeonGeometryView()
	if err := state.MoveDungeon(grid, 0, 1, 4); err != nil {
		t.Fatal(err)
	}
	afterX, afterY, _ := state.DungeonGeometryView()
	if afterX == beforeX && afterY == beforeY {
		t.Fatalf("這一步沒有移動，測不到快照：(%d,%d)", afterX, afterY)
	}
	gotX, okX := state.session.MemoryValue(eclPreMoveX)
	gotY, okY := state.session.MemoryValue(eclPreMoveY)
	if !okX || !okY {
		t.Fatalf("4BF0／4BF1 沒有被寫過：%v／%v", okX, okY)
	}
	if int(int16(gotX)) != beforeX || int(int16(gotY)) != beforeY {
		t.Fatalf("快照 ＝ (%d,%d)，預期移動**前**的 (%d,%d)",
			int(int16(gotX)), int(int16(gotY)), beforeX, beforeY)
	}
}

// TestMoveDungeonSnapshotTracksEveryStep 釘住它每一步都更新，不是只在第一步。
func TestMoveDungeonSnapshotTracksEveryStep(t *testing.T) {
	state, grid := dungeonMoveFixture(t)
	for step := 0; step < 2; step++ {
		beforeX, beforeY, _ := state.DungeonGeometryView()
		if err := state.MoveDungeon(grid, 0, 1, 4); err != nil {
			t.Fatalf("第 %d 步：%v", step+1, err)
		}
		gotX, _ := state.session.MemoryValue(eclPreMoveX)
		gotY, _ := state.session.MemoryValue(eclPreMoveY)
		if int(int16(gotX)) != beforeX || int(int16(gotY)) != beforeY {
			t.Fatalf("第 %d 步的快照 ＝ (%d,%d)，預期 (%d,%d)",
				step+1, int(int16(gotX)), int(int16(gotY)), beforeX, beforeY)
		}
	}
}
