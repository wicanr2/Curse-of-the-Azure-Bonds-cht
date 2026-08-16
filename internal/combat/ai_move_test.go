package combat

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// 方向表逐 byte 對回常駐資料段（`DS:2B6h`，spec 837）。
//
// ★ 為什麼要對 dump 而不是對規格文字：表抄進 Go 之後，規格與程式碼會各自演化，
// 而**原始位元組不會**。抄錯一格的後果是某個模式的繞行方向反了——玩起來只是
// 「怪物有點笨」，不會有任何測試變紅。
func TestAIDirectionTableMatchesTheResidentDump(t *testing.T) {
	const base = 0x2B6
	path := filepath.Join("..", "..", "workplace", "re-sweep", "dos", "dseg", "dos-dseg-dseg.bin")
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("資料段 dump 不在：%v", err)
	}
	for mode, offsets := range aiDirectionOffsets {
		for index, want := range offsets {
			// 序號從 1 起算：`base + 模式 × 5 + 序號`。
			at := base + mode*5 + index + 1
			if at >= len(blob) {
				t.Fatalf("模式 %d 序號 %d 落在 dump 之外", mode, index+1)
			}
			if got := int(blob[at]); got != want {
				t.Fatalf("模式 %d 的第 %d 個候選是 %d，資料段是 %d", mode, index+1, want, got)
			}
		}
	}
	// 表的前五個 byte（「模式 0」那一列）是前一張表的尾巴，值全部落在 1..8 之外
	// ——這正好反證模式從 1 起算、序號也從 1 起算。
	// ⚠ 只有這一列能這樣驗：模式 2 的「序號 0」其實是模式 1 的序號 5，
	// 它當然是合法方向。
	for index, value := range blob[base : base+5] {
		if value >= 1 && value <= 8 {
			t.Fatalf("表前面第 %d 個 byte 是 %d，落在合法方向範圍內；"+
				"若表真的從模式 0 起算，這條測試的前提要重看", index, value)
		}
	}
	if got := blob[base+5]; got != 0 {
		t.Fatalf("模式 1 的序號 0 是 %d，預期是分隔用的 0", got)
	}
}

// 模式重骰：1..4 有 3/4 機率延續、5／6 每回合必重骰，而且重骰時那一擲 1d8
// 的點數本身被丟掉（spec 830）。
func TestRollAIModeKeepsModesOneToFourThreeQuartersOfTheTime(t *testing.T) {
	// 第一擲 1d4 回 2（≠ 1）⇒ 保持原模式，之後不再擲。
	rolls := []int{2}
	index := 0
	roll := func(sides int) int {
		if index >= len(rolls) {
			t.Fatalf("多擲了一次 %dd", sides)
		}
		value := rolls[index]
		index++
		return value
	}
	if got := RollAIMode(roll, 3); got != 3 {
		t.Fatalf("模式=%d，1d4 ≠ 1 時應該保持 3", got)
	}
	if index != 1 {
		t.Fatalf("擲了 %d 次，保持原模式只該擲 1 次", index)
	}
}

func TestRollAIModeRerollsAndBurnsTheD8(t *testing.T) {
	// 1d4 ＝ 1 ⇒ 重骰；1d8 ＝ 8 ⇒ 走 5／6 那一條；1d2 ＝ 2 ⇒ 模式 6。
	rolls := []int{1, 8, 2}
	index := 0
	roll := func(int) int { value := rolls[index]; index++; return value }
	if got := RollAIMode(roll, 2); got != 6 {
		t.Fatalf("模式=%d，1d2 ＝ 2 時應該是 6", got)
	}
	if index != 3 {
		t.Fatalf("擲了 %d 次，重骰到 5／6 要 1d4 ＋ 1d8 ＋ 1d2 共 3 次", index)
	}
	// 1d8 不是 8 ⇒ 走 1d4，那一擲的點數（這裡是 5）被丟掉。
	rolls = []int{1, 5, 3}
	index = 0
	if got := RollAIMode(roll, 2); got != 3 {
		t.Fatalf("模式=%d，1d4 ＝ 3 時應該是 3", got)
	}
	if index != 3 {
		t.Fatalf("擲了 %d 次；1d8 那一擲不能省，否則亂數序列會偏掉", index)
	}
	// 模式 5／6 一律重骰，連「保持」的 1d4 都不擲。
	rolls = []int{7, 4}
	index = 0
	if got := RollAIMode(roll, 5); got != 4 {
		t.Fatalf("模式=%d", got)
	}
	if index != 2 {
		t.Fatalf("模式 5 應該直接重骰（1d8 ＋ 1d4 ＝ 2 擲），實際 %d 擲", index)
	}
}

func approachBattle(t *testing.T, monsterX, monsterY, allowance, weaponRange int) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: monsterX, CombatY: monsterY,
			MovementAllowance: allowance, WeaponRange: weaponRange},
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 10, CombatY: 5},
	}, 11)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 模式 1 的第一個候選是「不轉」，所以空曠地形上就是直線走過去，
// 走到打得到為止——不是走到移動力用完。
func TestApproachWalksStraightUntilTheTargetIsInReach(t *testing.T) {
	battle := approachBattle(t, 4, 5, 6, 1)
	result, err := battle.MonsterApproach("orc", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.InWeaponRange || result.Blocked {
		t.Fatalf("result=%+v", result)
	}
	orc, _ := battle.Fighter("orc")
	// 射程 1 ⇒ 加權距離 ≤ 3 就停；從 (4,5) 走到 (8,5) 距離 4 還不夠，(9,5) 距離 2 才停。
	if orc.CombatX != 9 || orc.CombatY != 5 {
		t.Fatalf("停在 (%d,%d)", orc.CombatX, orc.CombatY)
	}
	if result.HalfTilesUsed != len(result.Steps)*2 {
		t.Fatalf("正向每步應該花 2 個半格：steps=%d used=%d",
			len(result.Steps), result.HalfTilesUsed)
	}
}

// 移動力不夠就停在半路，而且不是「走到剩 0」——最後一步花不起就停。
func TestApproachStopsWhenMovementRunsOut(t *testing.T) {
	battle := approachBattle(t, 0, 5, 2, 1)
	result, err := battle.MonsterApproach("orc", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.InWeaponRange || !result.Blocked {
		t.Fatalf("result=%+v，移動力 2（＝ 4 個半格）走不到 (10,5)", result)
	}
	if len(result.Steps) != 2 || result.HalfTilesUsed != 4 {
		t.Fatalf("steps=%v used=%d，移動力 2 應該正好走兩步",
			result.Steps, result.HalfTilesUsed)
	}
}

// 射程長的怪不必貼上來：加權距離 ≤ 射程 × 2 ＋ 1 就停手。
func TestLongWeaponRangeStopsFurtherAway(t *testing.T) {
	battle := approachBattle(t, 0, 5, 12, 4)
	result, err := battle.MonsterApproach("orc", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.InWeaponRange {
		t.Fatalf("result=%+v", result)
	}
	orc, _ := battle.Fighter("orc")
	if orc.CombatX != 6 {
		t.Fatalf("射程 4 應該停在加權距離 ≤ 9 的第一格（x=6），實際 x=%d", orc.CombatX)
	}
}

// 正前方被牆擋住時換下一個候選方向——這就是模式存在的理由。
// 模式 1 的候選順序是 前、左 45°、左 90°、右 45°、右 90°。
func TestBlockedFrontFallsBackToTheNextCandidateDirection(t *testing.T) {
	battle := approachBattle(t, 4, 5, 6, 1)
	// x = 5 那一整行不能走，只留 y = 4 的缺口。
	terrain := func(x, y int) (int, bool) {
		if x == 5 && y != 4 {
			return 0, false
		}
		return 1, true
	}
	result, err := battle.MonsterApproach("orc", "hero", 1, terrain)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) == 0 {
		t.Fatalf("result=%+v，應該繞過去而不是卡住", result)
	}
	if result.Steps[0] != (TilePoint{X: 5, Y: 4}) {
		t.Fatalf("第一步走到 %+v，正前方被擋時模式 1 的下一個候選是左 45°", result.Steps[0])
	}
	// 斜向那一步花 3 個半格。
	if result.HalfTilesUsed < 3 {
		t.Fatalf("used=%d，斜向一步就要 3 個半格", result.HalfTilesUsed)
	}
}

// 四面都走不了就回 Blocked，不會空轉到 20 次上限。
func TestApproachReportsBlockedInsteadOfSpinning(t *testing.T) {
	battle := approachBattle(t, 4, 5, 6, 1)
	terrain := func(x, y int) (int, bool) {
		if x == 4 && y == 5 {
			return 1, true
		}
		return 0, false
	}
	result, err := battle.MonsterApproach("orc", "hero", 1, terrain)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || len(result.Steps) != 0 || result.Iterations != 1 {
		t.Fatalf("result=%+v，走不動應該立刻回報，不是繞到上限", result)
	}
}

// 別人站的格子不能踩。
func TestApproachDoesNotWalkThroughAnotherCombatant(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: 4, CombatY: 5, MovementAllowance: 6, WeaponRange: 1},
		{ID: "wall-orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: 5, CombatY: 5},
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 10, CombatY: 5},
	}, 11)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.MonsterApproach("orc", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range result.Steps {
		if step == (TilePoint{X: 5, Y: 5}) {
			t.Fatalf("踩到別人站的格子：%+v", result.Steps)
		}
	}
}

// 模式不在 1..6 就回錯誤，不要默默當成 1——模式是原作的資料，
// 猜一個預設值會讓「模式沒接上」看起來像正常運作。
func TestApproachRejectsAnUnknownMode(t *testing.T) {
	battle := approachBattle(t, 4, 5, 6, 1)
	if _, err := battle.MonsterApproach("orc", "hero", 0, nil); err == nil {
		t.Fatal("模式 0 應該回錯誤")
	}
	if _, err := battle.MonsterApproach("orc", "hero", 7, nil); err == nil {
		t.Fatal("模式 7 應該回錯誤（表裡有那一列，但 AI 的範圍是 1..6）")
	}
}

// 模式 5／6 是繞行：一路往同一邊轉。這裡只驗「候選順序真的被照用」——
// 目標在正東、正東被擋時模式 5 的下一個候選是北（左 45°）、模式 6 是南。
func TestWallHuggingModesTurnToOppositeSides(t *testing.T) {
	terrain := func(x, y int) (int, bool) {
		if x == 5 && y == 5 {
			return 0, false
		}
		return 1, true
	}
	for _, item := range []struct {
		mode int
		want TilePoint
	}{
		{5, TilePoint{X: 5, Y: 4}},
		{6, TilePoint{X: 5, Y: 6}},
	} {
		battle := approachBattle(t, 4, 5, 6, 1)
		result, err := battle.MonsterApproach("orc", "hero", item.mode, terrain)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Steps) == 0 || result.Steps[0] != item.want {
			t.Fatalf("模式 %d 的第一步是 %+v，want %+v", item.mode, result.Steps, item.want)
		}
	}
}

// PRNG 走 Battle 自己那一條，模式重骰不另開亂數源。
func TestRollAIModeUsesTheCallerSuppliedRoller(t *testing.T) {
	const seed int64 = 99
	reference := rand.New(rand.NewSource(seed))
	first := reference.Intn(4) + 1
	stream := rand.New(rand.NewSource(seed))
	roll := func(sides int) int { return stream.Intn(sides) + 1 }
	mode := RollAIMode(roll, 2)
	if first != 1 && mode != 2 {
		t.Fatalf("1d4 ＝ %d（≠ 1）⇒ 應該保持模式 2，實際 %d", first, mode)
	}
}
