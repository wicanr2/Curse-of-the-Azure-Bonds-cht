package ecl

import (
	"reflect"
	"testing"
)

// `37h LOAD PIECES` 的三支分派（DOS `overlay-02:0C15h`，spec 1087／1153）。
//
// ★ 全 corpus 23 處都走第三支——`7Fh` 一次都沒出現，而槽 2 每一處都帶著真實
// 片號。前兩支照 handler 寫，沒有實機路徑背書，所以更要用測試把形狀釘住。
func TestWallSetAssignmentsFollowsTheThreeBranches(t *testing.T) {
	for _, item := range []struct {
		name   string
		pieces [3]uint16
		memory map[uint16]uint16
		want   []WallSetAssignment
	}{{
		name:   "逐槽迴圈：0FFh 寫哨兵、其餘照載",
		pieces: [3]uint16{0x0E, 0x0F, 0xFF},
		want: []WallSetAssignment{
			{Slot: 1, Piece: 0x0E}, {Slot: 2, Piece: 0x0F}, {Slot: 3, Sentinel: true},
		},
	}, {
		// ⚠ `7Fh` 那一支只呼叫 `LOADWALLSET(1, 0)`：片號是 **0** 不是 `7Fh`，
		// 而且槽 2／3 **完全不動**——連哨兵都不寫。
		name:   "7Fh：只碰槽 1，片號 0",
		pieces: [3]uint16{0x7F, 0x0F, 0x10},
		want:   []WallSetAssignment{{Slot: 1, Piece: 0}},
	}, {
		name:   "兩槽分支：兩個閘門都非零時只載槽 1 與槽 3",
		pieces: [3]uint16{0x0E, 0x0F, 0x10},
		memory: map[uint16]uint16{wallSetTwoSlotGateA: 1, wallSetTwoSlotGateB: 1},
		want:   []WallSetAssignment{{Slot: 1, Piece: 0x0E}, {Slot: 3, Piece: 0x10}},
	}, {
		// ⚠ 兩槽分支的 `0FFh` **不寫哨兵**——原作那一支只有 `<> 0FFh 才載`，
		// 沒有 else。
		name:   "兩槽分支：0FFh 就整個跳過，不寫哨兵",
		pieces: [3]uint16{0xFF, 0x0F, 0xFF},
		memory: map[uint16]uint16{wallSetTwoSlotGateA: 3, wallSetTwoSlotGateB: 9},
		want:   []WallSetAssignment{},
	}, {
		name:   "閘門只有一個非零就不算數",
		pieces: [3]uint16{0x0E, 0x0F, 0x10},
		memory: map[uint16]uint16{wallSetTwoSlotGateA: 1},
		want: []WallSetAssignment{
			{Slot: 1, Piece: 0x0E}, {Slot: 2, Piece: 0x0F}, {Slot: 3, Piece: 0x10},
		},
	}} {
		got := WallSetAssignmentsFor(item.pieces, item.memory)
		if !reflect.DeepEqual(got, item.want) {
			t.Errorf("%s：得到 %+v，want %+v", item.name, got, item.want)
		}
	}
}

// `7Fh` 的優先序高於兩槽分支：原作是 `if ... else if ...`，不是兩個獨立判斷。
func TestWallSetAssignmentsChecksSpecialPieceFirst(t *testing.T) {
	memory := map[uint16]uint16{wallSetTwoSlotGateA: 1, wallSetTwoSlotGateB: 1}
	got := WallSetAssignmentsFor([3]uint16{0x7F, 0x0F, 0x10}, memory)
	want := []WallSetAssignment{{Slot: 1, Piece: 0}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("兩個閘門都開時 `7Fh` 仍要優先：得到 %+v，want %+v", got, want)
	}
}
