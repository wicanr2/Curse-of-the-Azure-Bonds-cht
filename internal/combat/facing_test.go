package combat

import "testing"

// combatDirection 是原作 `overlay-13:29A2h` 的八扇形分類。這一組樣本直接對著
// 反組譯的門檻算：`0x26A/0x100 = 2.4140625`（tan 67.5°）與
// `0x6A/0x100 = 0.4140625`（tan 22.5°），而且乘除是 16-bit 整數、先乘後除。
//
// 座標系是螢幕格：Y 往下增加。函式回傳的是**從 (x2,y2) 看向 (x1,y1)** 的方向。
func TestCombatDirectionMatchesTheOriginalSectors(t *testing.T) {
	cases := []struct {
		name           string
		x1, y1, x2, y2 int
		want           uint8
	}{
		{"正上方", 5, 1, 5, 5, 0},
		{"右上 45 度", 9, 1, 5, 5, 1},
		{"正右方", 9, 5, 5, 5, 2},
		{"右下 45 度", 9, 9, 5, 5, 3},
		{"正下方", 5, 9, 5, 5, 4},
		{"左下 45 度", 1, 9, 5, 5, 5},
		{"正左方", 1, 5, 5, 5, 6},
		{"左上 45 度", 1, 1, 5, 5, 7},
		{"同一格落在方向 0", 5, 5, 5, 5, 0},
		// 22.5 度門檻：dx=4 時 4*106/256 = 1，dy 要 <= 1 才算「正東」。
		{"扁一格仍是正右", 9, 6, 5, 5, 2},
		{"再扁一點就跨到右下", 9, 7, 5, 5, 3},
		// 67.5 度門檻：dx=4 時 4*618/256 = 9。八個扇形是**依序**試的，
		// 所以「右下」先被檢到，dy 要超過 9 才輪得到「正下」。
		{"dy = 9 仍是右下", 9, 14, 5, 5, 3},
		{"dy = 10 才是正下", 9, 15, 5, 5, 4},
	}
	for _, test := range cases {
		if got := combatDirection(test.x1, test.y1, test.x2, test.y2); got != test.want {
			t.Fatalf("%s：combatDirection(%d,%d,%d,%d) = %d，預期 %d",
				test.name, test.x1, test.y1, test.x2, test.y2, got, test.want)
		}
	}
}

// 最短轉法是 `(a − b + 8) mod 8`，大於 4 再取 `8 − 差`：結果 0..4，順逆不分。
func TestTurnDistanceTakesTheShorterWay(t *testing.T) {
	cases := []struct{ a, b, want uint8 }{
		{0, 0, 0}, {2, 0, 2}, {4, 0, 4}, {5, 0, 3}, {7, 0, 1}, {0, 7, 1}, {1, 6, 3},
	}
	for _, test := range cases {
		if got := turnDistance(test.a, test.b); got != test.want {
			t.Fatalf("turnDistance(%d,%d) = %d，預期 %d", test.a, test.b, got, test.want)
		}
	}
}

func facingBattle(t *testing.T) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10,
			AttackBonus: 5, DamageDiceCount: 1, DamageDiceSides: 6,
			HasCombatPosition: true, CombatX: 5, CombatY: 5},
		{ID: "orc", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10,
			AttackBonus: 5, DamageDiceCount: 1, DamageDiceSides: 6,
			HasCombatPosition: true, CombatX: 9, CombatY: 5},
	}, 1137)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// `AccountTurn` 是原作 `sub_194A`：動作計數加一，累計轉向加上最短轉法。
// ⚠ 它**不寫面向**——原作那一支也沒寫，寫面向的是別的呼叫點。
func TestAccountTurnCountsActionsAndAccumulatesTurning(t *testing.T) {
	battle := facingBattle(t)
	if err := battle.AccountTurn("hero", "orc"); err != nil {
		t.Fatal(err)
	}
	hero, _ := battle.Fighter("hero")
	// 從 orc(9,5) 看向 hero(5,5) 是正左方 ＝ 6；hero 的面向初始 0，最短轉法 2。
	if hero.CombatActionCount != 1 || hero.CombatTurnTotal != 2 {
		t.Fatalf("第一次轉向之後 動作計數=%d 累計轉向=%d，預期 1／2", hero.CombatActionCount, hero.CombatTurnTotal)
	}
	if hero.CombatFacing != 0 {
		t.Fatalf("AccountTurn 不該動面向，實際變成 %d", hero.CombatFacing)
	}
	for i := 0; i < 3; i++ {
		if err := battle.AccountTurn("hero", "orc"); err != nil {
			t.Fatal(err)
		}
	}
	hero, _ = battle.Fighter("hero")
	// 累計 2×4 = 8，對 8 取餘之後回到 0——它是環狀量，不是單調成本。
	if hero.CombatActionCount != 4 || hero.CombatTurnTotal != 0 {
		t.Fatalf("四次之後 動作計數=%d 累計轉向=%d，預期 4／0", hero.CombatActionCount, hero.CombatTurnTotal)
	}
}

// 攻擊動作的轉向段只在動作計數小於 2 時發生（spec 1019）。
func TestFaceTargetOnlyWhileActionCountIsBelowTwo(t *testing.T) {
	battle := facingBattle(t)
	turned, err := battle.FaceTarget("hero", "orc")
	if err != nil || !turned {
		t.Fatalf("第一次應該轉向：turned=%v err=%v", turned, err)
	}
	hero, _ := battle.Fighter("hero")
	// 從 orc 看向 hero 是 6，DOS 再取 (6+4) mod 8 = 2。
	if hero.CombatFacing != 2 {
		t.Fatalf("面向 = %d，預期 2", hero.CombatFacing)
	}
	hero.CombatActionCount = 2
	battle.fighters["hero"] = hero
	hero.CombatFacing = 0
	battle.fighters["hero"] = hero
	turned, err = battle.FaceTarget("hero", "orc")
	if err != nil {
		t.Fatal(err)
	}
	if turned {
		t.Fatal("動作計數已經是 2，不該再轉向")
	}
	hero, _ = battle.Fighter("hero")
	if hero.CombatFacing != 0 {
		t.Fatalf("不轉向時面向被改成 %d", hero.CombatFacing)
	}
}

// 三道條件全成立才選第二個 AC 欄位；缺一道就不成立。
func TestRearAttackNeedsAllThreeConditions(t *testing.T) {
	battle := facingBattle(t)
	defender, _ := battle.Fighter("orc")
	// 從 orc(9,5) 看向 hero(5,5) 是 6；把 orc 的面向設成 6 讓第二道成立。
	defender.CombatFacing = 6
	defender.CombatActionCount = 2
	defender.CombatTurnTotal = 5
	battle.fighters["orc"] = defender
	if applies, err := battle.RearAttackApplies("hero", "orc"); err != nil || !applies {
		t.Fatalf("三道都成立時應該成立：applies=%v err=%v", applies, err)
	}
	for _, broken := range []struct {
		name  string
		apply func(*Fighter)
	}{
		{"動作計數只有 1", func(f *Fighter) { f.CombatActionCount = 1 }},
		{"累計轉向只有 4", func(f *Fighter) { f.CombatTurnTotal = 4 }},
		{"面向對不上", func(f *Fighter) { f.CombatFacing = 2 }},
	} {
		probe, _ := battle.Fighter("orc")
		probe.CombatFacing, probe.CombatActionCount, probe.CombatTurnTotal = 6, 2, 5
		broken.apply(&probe)
		battle.fighters["orc"] = probe
		applies, err := battle.RearAttackApplies("hero", "orc")
		if err != nil {
			t.Fatal(err)
		}
		if applies {
			t.Fatalf("%s 時不該成立", broken.name)
		}
	}
}

// ★ 第二個 AC 欄位的**算法還沒解讀**，所以沒有標成已知時攻擊結算不得改用它。
// 這條擋住「先填一個看起來合理的值」——那會讓命中率悄悄偏掉。
func TestSecondArmorClassIsIgnoredUntilItIsDecoded(t *testing.T) {
	count := func(known bool) int {
		battle := facingBattle(t)
		defender, _ := battle.Fighter("orc")
		defender.CombatFacing, defender.CombatActionCount, defender.CombatTurnTotal = 6, 2, 5
		// 30 是**探針值**，不是原作會算出來的第二個 AC（那一定比正面那格小）。
		// 這裡要量的是「這一格到底有沒有被讀」，所以刻意挑一個好分辨的值。
		defender.ArmorClassFacing, defender.ArmorClassFacingKnown = 30, known
		battle.fighters["orc"] = defender
		hits := 0
		for i := 0; i < 200; i++ {
			attack, err := battle.Attack("hero", "orc")
			if err != nil {
				t.Fatal(err)
			}
			if attack.Hit {
				hits++
			}
			healed, _ := battle.Fighter("orc")
			healed.HitPoints = healed.MaxHitPoints
			battle.fighters["orc"] = healed
		}
		return hits
	}
	ignored, used := count(false), count(true)
	// ⚠ 自然 20 一律命中，與 AC 無關（`critical` 那一路），所以「用了 AC 30」
	// 不會是零命中，而是掉到只剩重擊。
	if used >= ignored {
		t.Fatalf("標成已知時應該明顯難打中：命中 %d 次 vs 不用第二個 AC 的 %d 次", used, ignored)
	}
	if ignored == 0 {
		t.Fatal("不用第二個 AC 時一次都沒命中，這組樣本測不出差別")
	}
	if used > ignored/4 {
		t.Fatalf("AC 30 應該只剩重擊打得中，實際命中 %d 次（對照 %d 次）", used, ignored)
	}
}

// 開場面向：`表{7,2,3,6}[隊伍朝向 div 2]`，敵方再轉 180 度（spec 805）。
// ⚠ 索引是**除以 2**，所以八個朝向只對到四筆——0 與 1 得到同一個面向。
func TestApplyInitialFacingUsesTheFourEntryTable(t *testing.T) {
	cases := []struct{ direction, party, enemy uint8 }{
		{0, 7, 3}, {1, 7, 3}, {2, 2, 6}, {4, 3, 7}, {6, 6, 2},
	}
	for _, test := range cases {
		battle := facingBattle(t)
		battle.ApplyInitialFacing(test.direction)
		hero, _ := battle.Fighter("hero")
		orc, _ := battle.Fighter("orc")
		if hero.CombatFacing != test.party || orc.CombatFacing != test.enemy {
			t.Fatalf("朝向 %d：隊伍 %d／敵方 %d，預期 %d／%d",
				test.direction, hero.CombatFacing, orc.CombatFacing, test.party, test.enemy)
		}
	}
}

// 防禦反擊之前，原作先對**被打的那個人**記一次轉向（spec 817），
// 所以計數在正式流程裡真的會動——不是只有單元測試自己塞值。
func TestGuardReactionAccountsTheMoversTurn(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "guard", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "enemy", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 2},
	}, 817)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.GuardAction("guard"); err != nil {
		t.Fatal(err)
	}
	if _, err := battle.MoveWithFreeAttacks("enemy", -1, 0); err != nil {
		t.Fatal(err)
	}
	enemy, _ := battle.Fighter("enemy")
	if enemy.CombatActionCount != 1 {
		t.Fatalf("被反擊的一方動作計數 = %d，預期 1", enemy.CombatActionCount)
	}
}

// 回合開始清掉動作計數與累計轉向，但**不清面向**（spec 804）。
func TestRoundStartClearsCountersButKeepsFacing(t *testing.T) {
	battle := facingBattle(t)
	hero, _ := battle.Fighter("hero")
	hero.CombatFacing, hero.CombatActionCount, hero.CombatTurnTotal = 5, 3, 6
	battle.fighters["hero"] = hero
	battle.initializeRoundDelays()
	hero, _ = battle.Fighter("hero")
	if hero.CombatActionCount != 0 || hero.CombatTurnTotal != 0 {
		t.Fatalf("回合開始沒清乾淨：動作計數=%d 累計轉向=%d", hero.CombatActionCount, hero.CombatTurnTotal)
	}
	if hero.CombatFacing != 5 {
		t.Fatalf("面向被清掉了，變成 %d；原作只清 +0Fh 與 +12h", hero.CombatFacing)
	}
}

// 原作的第二個 AC 一定比第一個小（`+19Bh ＝ +19Ah − 敏捷 − 盾牌 − 2`，
// spec 1000 §七），而命中判定是 `attackTotal >= AC`——**數字小才好打**。
// 所以背後攻擊必須比正面攻擊容易命中。
//
// ⚠ 這條擋的是符號錯誤：`monster.CombatArmorClass` 會把儲存值反轉，
// 直接搬 `+19Bh` 的絕對值會讓背後攻擊變成更難打，而那個錯誤在單看數字時
// 完全看不出來——兩邊都是「有換到第二個 AC」。
func TestRearAttackIsEasierThanFacingTheAttacker(t *testing.T) {
	count := func(rearPenalty int) int {
		battle := facingBattle(t)
		defender, _ := battle.Fighter("orc")
		defender.CombatFacing, defender.CombatActionCount, defender.CombatTurnTotal = 6, 2, 5
		defender.ArmorClass = 15
		defender.ArmorClassFacing = defender.ArmorClass - rearPenalty
		defender.ArmorClassFacingKnown = rearPenalty > 0
		battle.fighters["orc"] = defender
		hits := 0
		for i := 0; i < 400; i++ {
			attack, err := battle.Attack("hero", "orc")
			if err != nil {
				t.Fatal(err)
			}
			if attack.Hit {
				hits++
			}
			healed, _ := battle.Fighter("orc")
			healed.HitPoints = healed.MaxHitPoints
			battle.fighters["orc"] = healed
		}
		return hits
	}
	front, rear := count(0), count(6)
	if rear <= front {
		t.Fatalf("背後攻擊命中 %d 次、正面 %d 次——第二個 AC 的符號反了", rear, front)
	}
}
