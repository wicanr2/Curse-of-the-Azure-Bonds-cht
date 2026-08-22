package combat

import "testing"

// `ADJUSTBLOWS`（DOS `overlay-13:0F12h`）的完整真值表。
//
// ★ 重點是**奇數**那幾格：半次值 3 ＝ 一回合半次，原作靠回合奇偶交替成 2、1、2、1，
// 平均剛好 1.5。寫成 `半次值 / 2` 會**每一回合都少半次**而且不會有任何徵兆。
func TestAdjustBlowsAlternatesOnOddHalfBlows(t *testing.T) {
	for _, item := range []struct {
		blows int
		want  [4]int // 回合 0、1、2、3
	}{
		{blows: 0, want: [4]int{0, 0, 0, 0}},
		{blows: 1, want: [4]int{0, 1, 0, 1}},
		{blows: 2, want: [4]int{1, 1, 1, 1}},
		{blows: 3, want: [4]int{1, 2, 1, 2}},
		{blows: 4, want: [4]int{2, 2, 2, 2}},
		{blows: 5, want: [4]int{2, 3, 2, 3}},
		{blows: 8, want: [4]int{4, 4, 4, 4}},
	} {
		for round := 0; round < 4; round++ {
			if got := AdjustBlows(item.blows, round); got != item.want[round] {
				t.Errorf("半次值 %d 在回合 %d ＝ %d，want %d",
					item.blows, round, got, item.want[round])
			}
		}
	}
	// ⚠ 平均值必須真的是「半次值 ÷ 2」：連續兩回合加起來就是半次值本身。
	for blows := 0; blows <= 9; blows++ {
		if sum := AdjustBlows(blows, 0) + AdjustBlows(blows, 1); sum != blows {
			t.Errorf("半次值 %d 的兩回合合計 ＝ %d，want %d", blows, sum, blows)
		}
	}
}

func blowsBattle(t *testing.T, blows [2]int) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "beast", Side: SideEnemy, HitPoints: 200, MaxHitPoints: 200, HitDice: 4,
			ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			AttackBlows: blows, HasCombatPosition: true, CombatX: 2, CombatY: 1},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 槽 0 是 0 就用槽 1：`GIANT SPIDER`／`PHASE SPIDER` 的 `BASEATTBLOWS` 是 `{0, 2}`，
// 牠們用第二個武器槽咬（原作 `if 打手^[11Ch] = 0 then 槽 := 2`，spec 1010）。
func TestAttacksThisRoundFallsBackToTheSecondSlot(t *testing.T) {
	battle := blowsBattle(t, [2]int{0, 2})
	if got := battle.AttacksThisRound(battle.fighters["beast"]); got != 1 {
		t.Fatalf("槽 0 為 0 時 ＝ %d，want 1（改用槽 1）", got)
	}
	empty := blowsBattle(t, [2]int{0, 0})
	if got := empty.AttacksThisRound(empty.fighters["beast"]); got != 0 {
		t.Fatalf("兩槽都是 0 時 ＝ %d，want 0（原作不夾下限）", got)
	}
}

// 一回合四次的怪物（`THRI-KREEN` 的 `BASEATTBLOWS[0]` ＝ 8）真的要打四下。
func TestAttackSequenceUsesTheRecordBlows(t *testing.T) {
	battle := blowsBattle(t, [2]int{8, 2})
	results, err := battle.AttackSequence("beast", "hero")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 4 {
		t.Fatalf("打了 %d 下，want 4", len(results))
	}
}

// 遠程武器已經投影成整數次數，那一路不再走半次換算（原作是**取代**基準值，
// 而 CoAB 的射速全是偶數，所以兩種寫法結果相同，spec 1180）。
func TestAttacksThisRoundPrefersTheProjectedMissileRate(t *testing.T) {
	battle := blowsBattle(t, [2]int{2, 0})
	beast := battle.fighters["beast"]
	beast.AttacksPerTurn = 2
	battle.fighters["beast"] = beast
	if got := battle.AttacksThisRound(beast); got != 2 {
		t.Fatalf("架著弓時 ＝ %d，want 2", got)
	}
}

// 遠程攻擊次數被剩下的彈藥壓住（spec 808）。
//
// ⚠ **數量 0 不會壓成 1**：原作的 `m := 1` 被第二個條件擋掉了。這一格看起來像
// 疏漏，但照抄才是對的——所以兩邊都釘。
func TestCapByAmmunition(t *testing.T) {
	for _, item := range []struct {
		attacks, count, want int
	}{
		{attacks: 2, count: 0, want: 2}, // 沒有數量欄位 ⇒ 不壓
		{attacks: 2, count: 1, want: 1},
		{attacks: 4, count: 3, want: 3},
		{attacks: 2, count: 5, want: 2}, // 彈藥有餘 ⇒ 不動
		{attacks: 1, count: 1, want: 1},
	} {
		if got := capByAmmunition(item.attacks, item.count); got != item.want {
			t.Errorf("次數 %d、彈藥 %d ＝ %d，want %d",
				item.attacks, item.count, got, item.want)
		}
	}
}

// 整條走完：架著弓（射速投影成 2 次）而箭只剩 1 支 ⇒ 這一回合只射一次。
func TestAttacksThisRoundIsCappedByTheReadiedAmmunition(t *testing.T) {
	battle := blowsBattle(t, [2]int{2, 0})
	archer := battle.fighters["beast"]
	archer.AttacksPerTurn = 2
	archer.AmmunitionCount = 1
	battle.fighters["beast"] = archer
	if got := battle.AttacksThisRound(archer); got != 1 {
		t.Fatalf("剩一支箭時 ＝ %d，want 1", got)
	}
	archer.AmmunitionCount = 0
	if got := battle.AttacksThisRound(archer); got != 2 {
		t.Fatalf("數量欄位是 0 時 ＝ %d，want 2（原作不壓）", got)
	}
}
