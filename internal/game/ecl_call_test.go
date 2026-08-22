package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

func TestApplyECLCallSignalsMovesForwardAndWraps(t *testing.T) {
	tests := []struct {
		name      string
		x, y      int
		direction uint8
		wantX     int
		wantY     int
	}{
		{name: "north wraps", x: 7, y: 0, direction: 0, wantX: 7, wantY: 15},
		{name: "east wraps", x: 15, y: 6, direction: 2, wantX: 0, wantY: 6},
		{name: "south wraps", x: 7, y: 15, direction: 4, wantX: 7, wantY: 0},
		{name: "west wraps", x: 0, y: 6, direction: 6, wantX: 15, wantY: 6},
		{name: "odd direction unchanged", x: 7, y: 6, direction: 1, wantX: 7, wantY: 6},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := State{DungeonX: test.x, DungeonY: test.y, DungeonDirection: test.direction}
			state.applyECLCallSignals(ecl.RunResult{CallAddresses: []uint16{0xC01E}})
			if state.DungeonX != test.wantX || state.DungeonY != test.wantY {
				t.Fatalf("position=(%d,%d), want (%d,%d)", state.DungeonX, state.DungeonY, test.wantX, test.wantY)
			}
			if got := state.ConsumeECLCallRequests(); !reflect.DeepEqual(got, []uint16{0xC01E}) {
				t.Fatalf("requests=%#v", got)
			}
			if got := state.ConsumeECLCallRequests(); len(got) != 0 {
				t.Fatalf("requests were not one-shot: %#v", got)
			}
		})
	}
}

// scriptView 用 VM 那支 `ViewMirror.Store` 建出鏡射——**不要在測試裡手寫鏡射
// 的欄位**，那會變成第二份規則，改了 VM 也不會紅（spec 1172）。
// scriptView 造一份 `2Dh` 的鏡射快照：**先把隊伍現在的位置 Adopt 進去**，
// 再套上腳本這一次的寫入。
//
// ⚠ 不要從零值開始。原作的 `720Fh`／`7210h`／`7211h` 永遠等於真實位置——引擎
// 每動一步就回寫；remake 這一側對應的是 `syncDungeonECLRegisters()`。從零值起
// 造出來的快照等於假設「腳本沒寫的格子是 `(0,0)` 朝北」，那個狀態在真實執行
// 裡不存在，用它當輸入會把「只寫一格的重畫該怎麼投影」問錯（spec 1172）。
func scriptView(state *State, block uint8, writes ...[2]uint16) ecl.ViewMirror {
	var view ecl.ViewMirror
	view.Adopt(0xC04B, uint16(state.DungeonX))
	view.Adopt(0xC04C, uint16(state.DungeonY))
	view.Adopt(0xC04D, uint16(state.DungeonDirection/2))
	for _, write := range writes {
		view.Store(write[0], write[1], block)
	}
	return view
}

func TestApplyECLCallSignalsPreservesOrderAndDefaultSound(t *testing.T) {
	state := State{}
	state.applyECLCallSignals(ecl.RunResult{CallAddresses: []uint16{0x2E10, 0xB200, 0x9999}})

	if got, want := state.ConsumeECLCallRequests(), []uint16{0x2E10, 0xB200, 0x9999}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requests=%#v, want %#v", got, want)
	}
	if got, want := state.ConsumeSoundEvents(), []SoundEvent{SoundStep}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sound=%#v, want %#v", got, want)
	}
}

func TestApplyECLCallSignalsRedrawProjectsSameBlockDungeonRegisters(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
	}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 11)
	session.SetMemoryValue(0xC04C, 10)
	session.SetMemoryValue(0xC04D, 2)
	state := State{
		session:          session,
		DungeonX:         12,
		DungeonY:         9,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x42,
		SessionEndBlockID:    0x42,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x1096, BlockID: 0x42, Sequence: 4,
				View: scriptView(&state, 0x42,
					[2]uint16{0xC04B, 11}, [2]uint16{0xC04C, 10}, [2]uint16{0xC04D, 2})},
		},
	})

	if state.DungeonX != 11 || state.DungeonY != 10 ||
		state.DungeonDirection != 4 || state.MapX != 11 || state.MapY != 10 {
		t.Fatalf("redraw position=(%d,%d,%d) map=(%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.MapX, state.MapY)
	}
}

func TestApplyECLCallSignalsRedrawIgnoresStaleDungeonRegisters(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
	}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 11)
	session.SetMemoryValue(0xC04C, 10)
	session.SetMemoryValue(0xC04D, 2)
	state := State{
		session:          session,
		DungeonX:         3,
		DungeonY:         4,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{CallAddresses: []uint16{0x2E10}})

	if state.DungeonX != 3 || state.DungeonY != 4 || state.DungeonDirection != 6 {
		t.Fatalf("stale redraw changed position=(%d,%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}

func TestApplyECLCallSignalsRedrawIgnoresPriorBlockTransaction(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
		0x43: {0, 0},
	}, 0x43)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 11)
	session.SetMemoryValue(0xC04C, 10)
	session.SetMemoryValue(0xC04D, 2)
	state := State{
		session:          session,
		DungeonX:         3,
		DungeonY:         4,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x42,
		SessionEndBlockID:    0x43,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x1096, BlockID: 0x42, Sequence: 4},
		},
		SaveWrites: []ecl.MemoryWrite{
			{Address: 0xC04B, Value: 11, PC: 0x1084, BlockID: 0x42, Sequence: 1},
			{Address: 0xC04C, Value: 10, PC: 0x108A, BlockID: 0x42, Sequence: 2},
			{Address: 0xC04D, Value: 2, PC: 0x1090, BlockID: 0x42, Sequence: 3},
		},
	})

	if state.DungeonX != 3 || state.DungeonY != 4 || state.DungeonDirection != 6 {
		t.Fatalf("prior-block redraw changed position=(%d,%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}

// 只寫其中一格的重畫：沒被寫的那兩格沿用鏡射裡的值，而鏡射本來就等於隊伍
// 目前的位置（引擎每動一步都回寫）。所以結果是「X 換成腳本寫的、Y 與朝向
// 不變」——原作讀滿三格，不需要「只投影寫過的」這種遮罩（spec 1172）。
func TestApplyECLCallSignalsRedrawProjectsTheWholeMirror(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x42: {0, 0},
	}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 10)
	session.SetMemoryValue(0xC04C, 2)
	session.SetMemoryValue(0xC04D, 0)
	state := State{
		session:          session,
		DungeonX:         9,
		DungeonY:         2,
		DungeonDirection: 6,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x42,
		SessionEndBlockID:    0x42,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x13DB, BlockID: 0x42, Sequence: 3,
				View: scriptView(&state, 0x42, [2]uint16{0xC04B, 10}, [2]uint16{0xC04D, 0})},
		},
	})

	if state.DungeonX != 10 || state.DungeonY != 2 ||
		state.DungeonDirection != 0 || state.MapX != 10 || state.MapY != 2 {
		t.Fatalf("partial redraw position=(%d,%d,%d) map=(%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.MapX, state.MapY)
	}
}

// 只寫 `C04B`／`C04C`、不寫 `C04D` 的重畫仍然是真實移動：原作「退回上一格」
// 的出口（`ECL2/0x01:1444h`「THEY SEND YOU BACK. YOU MOVE AWAY.」等 15 處，
// spec 1157）就是這個形狀，朝向刻意保持不變。
func TestApplyECLCallSignalsRedrawWithoutFacingStillMovesTheParty(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{
		0x01: {0, 0},
	}, 0x01)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(0xC04B, 0)
	session.SetMemoryValue(0xC04C, 0)
	state := State{
		session:          session,
		DungeonX:         6,
		DungeonY:         5,
		DungeonDirection: 2,
	}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x2E10},
		SessionStartBlockID:  0x01,
		SessionEndBlockID:    0x01,
		SessionBlockRangeSet: true,
		CallRequests: []ecl.CallRequest{
			{Address: 0x2E10, PC: 0x1452, BlockID: 0x01, Sequence: 3,
				View: scriptView(&state, 0x01, [2]uint16{0xC04B, 0}, [2]uint16{0xC04C, 0})},
		},
	})

	if state.DungeonX != 0 || state.DungeonY != 0 || state.DungeonDirection != 2 ||
		state.MapX != 0 || state.MapY != 0 {
		t.Fatalf("restore-previous-cell position=(%d,%d,%d) map=(%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection,
			state.MapX, state.MapY)
	}
}

// `2Dh CALL 6803h` 把圖片序列往前推一格；換圖時原作的 `LOADSEQUENCE` 會把游標
// 設回第 1 格，所以有換圖的執行取代、沒換圖的執行累加（spec 1150）。
func TestApplyECLCallSignalsCarriesPictureFrameCursor(t *testing.T) {
	state := State{}
	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x6803, 0x6803},
		PictureFrameAdvances: 2,
	})
	if state.PictureFrameAdvances() != 2 {
		t.Fatalf("游標推了 %d 格，應該是 2", state.PictureFrameAdvances())
	}

	// 沒換圖 ⇒ 累加。
	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x6803},
		PictureFrameAdvances: 1,
	})
	if state.PictureFrameAdvances() != 3 {
		t.Fatalf("累加之後是 %d 格，應該是 3", state.PictureFrameAdvances())
	}

	// 換圖 ⇒ 取代（VM 已經在 `0Eh PICTURE` 把計數歸零）。
	state.applyECLCallSignals(ecl.RunResult{
		CallAddresses:        []uint16{0x6803},
		PictureRequested:     true,
		PictureBlock:         0x1D,
		PictureFrameAdvances: 1,
	})
	if state.PictureFrameAdvances() != 1 {
		t.Fatalf("換圖之後是 %d 格，應該是 1", state.PictureFrameAdvances())
	}
}

// 執行序這件事現在由 VM 的鏡射結構性地保證（`CALL` 當下複製一份），
// 回歸擋板搬到 `internal/ecl` 的 `TestViewMirrorSnapshotIsTakenAtTheCall`，
// 那一支跑的是真的 bytecode 而不是手寫的 fixture（spec 1172）。

// ★ 「寫了座標卻沒有 `2E10h`」的執行也要搬隊伍。
//
// 原作的 `STOREVALUE` 一寫 `C04B` 就當場改 `720Fh`——隊伍在那一刻就在新格子上，
// 重畫只是重畫。remake 的投影掛在 `2Dh CALL 2E10h` 上，所以這種執行要靠
// `RunResult.FinalView` 在收尾補一次（spec 1172）。
//
// 形狀取自 `ECL3/0x10:0C7Eh`「指揮官帶你走側門」：連寫 `C04B=2`／`C04C=5`／
// `C04D=1` 之後直接 `GOTO` 回主分派器，**沒有** `CALL 2E10h`；隔壁那條搶劫
// 結局（`0D24h`）寫完同樣三格才有重畫。
func TestApplyECLCallSignalsProjectsCoordinatesWrittenWithoutARedraw(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x10: {0, 0}}, 0x10)
	if err != nil {
		t.Fatal(err)
	}
	state := State{session: session, DungeonX: 1, DungeonY: 3, DungeonDirection: 6}
	state.Area.InDungeon = true

	state.applyECLCallSignals(ecl.RunResult{
		SessionStartBlockID:  0x10,
		SessionEndBlockID:    0x10,
		SessionBlockRangeSet: true,
		FinalView: scriptView(&state, 0x10,
			[2]uint16{0xC04B, 2}, [2]uint16{0xC04C, 5}, [2]uint16{0xC04D, 1}),
	})

	if state.DungeonX != 2 || state.DungeonY != 5 || state.DungeonDirection != 2 ||
		state.MapX != 2 || state.MapY != 5 {
		t.Fatalf("沒有重畫的座標寫入位置 ＝ (%d,%d,%d) map=(%d,%d)，want (2,5,2)",
			state.DungeonX, state.DungeonY, state.DungeonDirection, state.MapX, state.MapY)
	}
}

// ⚠ 收尾投影**不是「再重畫一次」**：`2E10h` 已經投影過的執行，髒旗標在 VM 裡
// 就被清掉了，`FinalView` 進不來那一段。用一份乾淨的鏡射確認它不動位置。
func TestApplyECLCallSignalsFinalViewDoesNotReprojectAfterARedraw(t *testing.T) {
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x10: {0, 0}}, 0x10)
	if err != nil {
		t.Fatal(err)
	}
	state := State{session: session, DungeonX: 4, DungeonY: 7, DungeonDirection: 0}
	state.Area.InDungeon = true

	view := scriptView(&state, 0x10, [2]uint16{0xC04B, 9}, [2]uint16{0xC04C, 9})
	view.ClearDirty()
	state.applyECLCallSignals(ecl.RunResult{
		SessionStartBlockID:  0x10,
		SessionEndBlockID:    0x10,
		SessionBlockRangeSet: true,
		FinalView:            view,
	})

	if state.DungeonX != 4 || state.DungeonY != 7 || state.DungeonDirection != 0 {
		t.Fatalf("重畫過的鏡射不該再投影一次：位置 ＝ (%d,%d,%d)",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}
