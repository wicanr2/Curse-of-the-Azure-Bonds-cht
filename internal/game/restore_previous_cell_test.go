package game

import (
	"archive/zip"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// 原作有一批「退回上一格」的出口：`SAVE 4BF0 C04B; SAVE 4BF1 C04C; CALL 2E10`，
// 把這一步移動之前的座標寫回去、朝向不動（spec 1157 列出 15 處）。火刀據點冰凍
// 房間的 `RETREAT` 就是其中一處：`ECL2/0x04:0E65h` 的 `ON GOTO` 第 0 項跳到
// `160Ch`。
//
// 這一條兩邊都釘：選 `RETREAT` 要退回原來那一格，選 `INTERROGATE` 要留在房裡。
// 只釘一邊的話，「所有重畫都不搬座標」也會過。
func TestFireKnifeFrozenRoomRetreatReturnsToPreviousCell(t *testing.T) {
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)
	grid := loadGeoCampaignGrid(t, image, 2, "GEO2.DAX", 0x04)
	hideout, ok := segment.Lookup("ECL2/0x04")
	if !ok {
		t.Fatal("註冊表沒有 ECL2/0x04")
	}

	for _, option := range []struct {
		id     string
		wantX  int
		wantY  int
		reason string
	}{
		{"ecl-option.retreat", 5, 2, "退回上一格"},
		{"ecl-option.interrogate", 4, 2, "留在冰凍房間"},
	} {
		t.Run(option.id, func(t *testing.T) {
			state := newSegmentDungeonState(t, blocks, records, hideout)
			state.SetECLSeed(1)
			// 從 (5,2) 往西踏進 (4,2) 的冰凍房間；`MoveDungeon` 會先把
			// 移動前的座標存進 `4BF0`／`4BF1`，那正是退回去時用的來源。
			state.SetDungeonGeometryView(5, 2, 6)
			state.DungeonWallRoof = grid.CellWrapped(5, 2).Terrain
			if err := state.MoveDungeon(grid, -1, 0, 6); err != nil {
				t.Fatalf("走進冰凍房間：%v", err)
			}
			want := requireGamePackText(t, &state, "fire-knife.frozen-room")
			if state.Message != want && state.Prompt != want {
				t.Fatalf("冰凍房間的敘述＝%q／%q", state.Message, state.Prompt)
			}
			if err := state.Select(requireGamePackOptionIndex(t, &state, option.id)); err != nil {
				t.Fatalf("選 %s：%v", option.id, err)
			}
			for step := 0; step < 8 && state.Mode != ModeDungeon; step++ {
				if state.CombatActive() {
					for turn := 0; turn < 400 && state.CombatActive(); turn++ {
						if err := state.CombatAct(); err != nil {
							t.Fatalf("冰凍房間戰鬥：%v", err)
						}
					}
					continue
				}
				if err := state.Continue(); err != nil {
					if selectErr := state.Select(0); selectErr != nil {
						t.Fatalf("推不動：continue=%v select=%v", err, selectErr)
					}
				}
			}
			if state.DungeonX != option.wantX || state.DungeonY != option.wantY {
				t.Fatalf("%s 之後在 (%d,%d)，要 (%d,%d)", option.reason,
					state.DungeonX, state.DungeonY, option.wantX, option.wantY)
			}
			// 朝向不在那兩句 `SAVE` 裡，所以退回去之後仍然朝西。
			if state.DungeonDirection != 6 {
				t.Fatalf("%s 之後朝向＝%d，要 6", option.reason, state.DungeonDirection)
			}
		})
	}
}
