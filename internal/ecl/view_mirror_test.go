package ecl

import "testing"

// viewProgram 是一段真的 bytecode：寫座標 → `CALL 2E10h` → 再寫一輪 → 再 `CALL`。
//
// ★ 用 bytecode 而不是手寫 `RunResult`：這一條要擋的是「快照在什麼時候取」，
// 手寫 fixture 會把答案直接寫進輸入裡。
var viewProgram = []byte{
	0, 0,
	0x09, 0x00, 0x03, 0x02, 0x4B, 0xC0, // SAVE 3 -> C04B
	0x09, 0x00, 0x04, 0x02, 0x4C, 0xC0, // SAVE 4 -> C04C
	0x09, 0x00, 0x00, 0x02, 0x4D, 0xC0, // SAVE 0 -> C04D
	0x2D, 0x02, 0x10, 0x2E, //             CALL 2E10h
	0x09, 0x00, 0x07, 0x02, 0x4B, 0xC0, // SAVE 7 -> C04B
	0x09, 0x00, 0x08, 0x02, 0x4C, 0xC0, // SAVE 8 -> C04C
	0x09, 0x00, 0x01, 0x02, 0x4D, 0xC0, // SAVE 1 -> C04D
	0x2D, 0x02, 0x10, 0x2E, //             CALL 2E10h
	0x00, //                               EXIT
}

// ★★ 快照在 `CALL` 執行的那一刻取，後面再寫也不會回頭改它；而重畫會把髒旗標
// 清掉，所以第二次 `CALL` 的髒旗標只反映它自己前面那一輪。
func TestViewMirrorSnapshotIsTakenAtTheCall(t *testing.T) {
	result, err := RunSubset(viewProgram, 0, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CallRequests) != 2 {
		t.Fatalf("應該有兩條 CALL，實際 %d", len(result.CallRequests))
	}
	first, second := result.CallRequests[0].View, result.CallRequests[1].View
	if first.X != 3 || first.Y != 4 || first.Facing != 0 {
		t.Fatalf("第一次快照 ＝ (%d,%d,%d)，want (3,4,0)", first.X, first.Y, first.Facing)
	}
	if second.X != 7 || second.Y != 8 || second.Facing != 2 {
		t.Fatalf("第二次快照 ＝ (%d,%d,%d)，want (7,8,2)", second.X, second.Y, second.Facing)
	}
	for index, view := range []ViewMirror{first, second} {
		if view.Dirty&ViewDirtyCoords == 0 {
			t.Fatalf("第 %d 次快照沒有立 `8B68h`", index+1)
		}
	}
}

// `C04D` 要折成 0／2／4／6，而且只有三格會弄髒座標那一條旗標。
func TestViewMirrorFoldsFacingAndSeparatesFlags(t *testing.T) {
	var view ViewMirror
	for value, want := range map[uint16]uint16{0: 0, 1: 2, 2: 4, 3: 6, 5: 2} {
		view.Store(0xC04D, value, 1)
		if view.Facing != want {
			t.Fatalf("`C04D` ＝ %d 折成 %d，want %d", value, view.Facing, want)
		}
	}
	view = ViewMirror{}
	view.Store(0xC059, 1, 1)
	if view.Dirty != ViewDirtyCell {
		t.Fatalf("`C059` 應該只立 `8B67h`：dirty=%02x", view.Dirty)
	}
	view = ViewMirror{}
	view.Store(0x4BFD, 1, 1)
	if view.Dirty != ViewDirtyPartyCell {
		t.Fatalf("`4BFD` 應該只立 `8B6Ah`：dirty=%02x", view.Dirty)
	}
}

// `CALL C01Eh` 當場改地圖暫存器——同一次執行裡排在後面的重畫看得到走完的位置。
func TestViewMirrorStepForwardWrapsLikeTheOriginal(t *testing.T) {
	for _, probe := range []struct {
		facing, x, y, wantX, wantY uint16
	}{
		{facing: 0, x: 5, y: 0, wantX: 5, wantY: 15},
		{facing: 2, x: 15, y: 5, wantX: 0, wantY: 5},
		{facing: 4, x: 5, y: 15, wantX: 5, wantY: 0},
		{facing: 6, x: 0, y: 5, wantX: 15, wantY: 5},
	} {
		view := ViewMirror{X: probe.x, Y: probe.y, Facing: probe.facing, Known: true}
		view.StepForward()
		if view.X != probe.wantX || view.Y != probe.wantY {
			t.Fatalf("朝向 %d 從 (%d,%d) 走到 (%d,%d)，want (%d,%d)",
				probe.facing, probe.x, probe.y, view.X, view.Y, probe.wantX, probe.wantY)
		}
		if view.Dirty != 0 {
			t.Fatal("走一步不該弄髒——原作那一支不經過 STOREVALUE")
		}
	}
	// 還沒有人寫過三格時不能亂走。
	var unknown ViewMirror
	unknown.StepForward()
	if unknown.Known || unknown.X != 0 || unknown.Y != 0 {
		t.Fatal("沒有已知座標時 StepForward 不該動")
	}
}

// ★ 「寫了座標卻沒有 `2E10h`」的執行必須留著髒旗標，收尾投影才接得到。
//
// 原作的 `STOREVALUE` 一寫 `C04B` 就當場改 `720Fh`，隊伍在那一刻就搬了；
// remake 的投影掛在重畫上，所以這種執行要靠 `RunResult.FinalView` 補。實例是
// `ECL3/0x10:0C7Eh`「指揮官帶你走側門」：連寫三格之後直接 `GOTO` 回主分派器，
// 隔壁那條搶劫結局（`0D24h`）才有 `CALL 2E10h`（spec 1172）。
func TestViewMirrorKeepsCoordsDirtyWithoutARedraw(t *testing.T) {
	var view ViewMirror
	view.Adopt(0xC04B, 1)
	view.Adopt(0xC04C, 3)
	view.Adopt(0xC04D, 3)
	if view.Dirty != 0 {
		t.Fatalf("引擎自己搬隊伍不該弄髒：dirty=%02x", view.Dirty)
	}
	view.Store(0xC04B, 2, 0x10)
	view.Store(0xC04C, 5, 0x10)
	view.Store(0xC04D, 1, 0x10)
	if view.X != 2 || view.Y != 5 || view.Facing != 2 {
		t.Fatalf("三格 ＝ (%d,%d,%d)，want (2,5,2)", view.X, view.Y, view.Facing)
	}
	if view.Dirty&ViewDirtyCoords == 0 {
		t.Fatal("沒有重畫的座標寫入必須留著 `8B68h`")
	}
	// 重畫才清；清掉之後同一筆寫入不會再被投影一次。
	view.ClearDirty()
	if view.Dirty != 0 {
		t.Fatalf("重畫之後 dirty ＝ %02x", view.Dirty)
	}
}
