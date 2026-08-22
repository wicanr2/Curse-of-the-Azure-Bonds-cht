package party

import "testing"

// 原作 `overlay-25:042Fh` 的三個門檻。
//
// ⚠ **聖武士是 7 級不是規則書的 8 級**——這一條只能從原作讀出來，照 AD&D 填會錯
// 一級（spec 1180）。
func TestBaseAttackBlowsThresholds(t *testing.T) {
	for _, item := range []struct {
		class Class
		level int
		want  int
	}{
		{ClassFighter, 6, 2}, {ClassFighter, 7, 3},
		{ClassPaladin, 6, 2}, {ClassPaladin, 7, 3},
		{ClassRanger, 7, 2}, {ClassRanger, 8, 3},
		{ClassCleric, 20, 2}, {ClassMagicUser, 20, 2}, {ClassThief, 20, 2},
	} {
		character := Character{Class: item.class, Level: item.level}
		if got := character.BaseAttackBlows(); got != item.want {
			t.Errorf("%v 等級 %d ＝ %d 半次，want %d", item.class, item.level, got, item.want)
		}
	}
}

// 多職角色走 `ClassLevels` 的八個槽：任何一個職業過門檻就算數
// （原作的迴圈跑完每一個職業，寫進去就不會被後面的分支蓋掉）。
func TestBaseAttackBlowsUsesEveryClassSlot(t *testing.T) {
	// 槽 2 ＝ 戰士、槽 5 ＝ 法師（`ClassLevel` 的對照）。
	var levels [8]uint8
	levels[2] = 7
	levels[5] = 3
	if got := (Character{ClassLevels: levels}).BaseAttackBlows(); got != 3 {
		t.Fatalf("戰士 7／法師 3 ＝ %d 半次，want 3", got)
	}
	levels[2] = 6
	if got := (Character{ClassLevels: levels}).BaseAttackBlows(); got != 2 {
		t.Fatalf("戰士 6／法師 3 ＝ %d 半次，want 2", got)
	}
}

// ★ 原作寫進 `+11Ch` 的立即數只有 2 與 3：隊員**永遠到不了一回合兩次**。
// 這一條擋住「照 AD&D 13 級補一個 4」的好意——那在這一款是無中生有。
func TestBaseAttackBlowsNeverReachesTwoAttacks(t *testing.T) {
	for level := 1; level <= 40; level++ {
		for _, class := range []Class{ClassCleric, ClassFighter, ClassRanger,
			ClassPaladin, ClassMagicUser, ClassThief} {
			if got := (Character{Class: class, Level: level}).BaseAttackBlows(); got > 3 {
				t.Fatalf("%v 等級 %d ＝ %d 半次，原作最高只有 3", class, level, got)
			}
		}
	}
}
