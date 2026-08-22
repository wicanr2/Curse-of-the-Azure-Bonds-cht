package main

import (
	"testing"

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

func at(value int) *int { return &value }

// ★ 第一版的比對只看「這一組有寫的欄位」，於是**只寫朝向**的那一組（走位收尾
// 的 `SAVE 01 C04D`）因為 X／Y 都是 nil 而無條件相符——猶拉什就被判成
// 「宣告值等於腳本寫的」，而那正是最需要被標成 `mismatch` 的一張（spec 1184）。
func TestPlacementNeedsBothCoordinatesToMatch(t *testing.T) {
	spawn := goldenbox.MapSpawn{X: 0, Y: 3, Direction: 2}

	facingOnly := placement{Offset: 0x1991, Facing: at(1)}
	if facingOnly.matches(spawn) {
		t.Fatal("只寫朝向的那一組不算「等於宣告值」")
	}
	xOnly := placement{Offset: 0x11F2, X: at(0)}
	if xOnly.matches(spawn) {
		t.Fatal("只寫 X 的那一組也不算")
	}
	full := placement{Offset: 0x006C, X: at(0), Y: at(3), Facing: at(1)}
	if !full.matches(spawn) {
		t.Fatal("座標與朝向都對得上就該算相符")
	}
}

// 朝向沒寫仍然不比對：只寫座標、朝向刻意保持不變是原作的常見形狀
// （spec 1157 的「退回上一格」共 23 處）。
func TestPlacementIgnoresUnwrittenFacing(t *testing.T) {
	spawn := goldenbox.MapSpawn{X: 6, Y: 15, Direction: 4}
	coordsOnly := placement{Offset: 0x088B, X: at(6), Y: at(15)}
	if !coordsOnly.matches(spawn) {
		t.Fatal("朝向沒寫就不該拿它判不相符")
	}
	wrongFacing := placement{Offset: 0x088B, X: at(6), Y: at(15), Facing: at(3)}
	if wrongFacing.matches(spawn) {
		t.Fatal("朝向有寫而且對不上就是不相符")
	}
}

// `C04D` 在暫存器裡是折過的 0..3，畫面上的八向是它乘二（spec 1150）。
// 少折一次會讓每一張圖都判成對不上。
func TestPlacementFoldsFacingLikeTheOriginal(t *testing.T) {
	for register, eight := range map[int]uint8{0: 0, 1: 2, 2: 4, 3: 6} {
		spawn := goldenbox.MapSpawn{X: 1, Y: 1, Direction: eight}
		group := placement{Offset: 0, X: at(1), Y: at(1), Facing: at(register)}
		if !group.matches(spawn) {
			t.Fatalf("`C04D`=%d 應該折成八向 %d", register, eight)
		}
	}
}

// 判定表的五種結果各自要走得到，尤其 `mismatch`——那是要人去看的那一類。
func TestVerdictCovers(t *testing.T) {
	spawn := goldenbox.MapSpawn{X: 5, Y: 7, Direction: 6}
	hit := placement{Offset: 0, X: at(5), Y: at(7), Facing: at(3)}
	miss := placement{Offset: 0, X: at(12), Y: at(7), Facing: at(3)}
	for _, item := range []struct {
		row  row
		want string
	}{
		{row{}, "none"},
		{row{Found: []placement{miss}}, "script-only"},
		{row{Spawn: &spawn}, "declared-only"},
		{row{Spawn: &spawn, Found: []placement{miss}}, "mismatch"},
		{row{Spawn: &spawn, Found: []placement{miss, hit}}, "script-agrees"},
	} {
		if got := verdict(item.row); got != item.want {
			t.Fatalf("verdict ＝ %q，want %q", got, item.want)
		}
	}
}
