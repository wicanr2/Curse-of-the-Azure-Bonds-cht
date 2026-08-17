package combat

import "testing"

func moraleBattle(t *testing.T, seed int64) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			// 33h 的最高位沒設 ⇒ 逃走那一段會把它補成 0B3h。
			ControlMorale: 0x33, HasCombatPosition: true, CombatX: 3, CombatY: 3},
		{ID: "orc2", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1},
		{ID: "hero", Side: SideParty, HitPoints: 12, MaxHitPoints: 12},
	}, seed)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 四段的邊界：1..10 / 11..60 / 61..80 / 81..100，加起來剛好 100。
func TestMoraleFailureCoversEveryRollExactlyOnce(t *testing.T) {
	seen := map[int]MoraleOutcome{}
	for seed := int64(1); seed <= 600; seed++ {
		battle := moraleBattle(t, seed)
		result, err := battle.ResolveMoraleFailure("orc")
		if err != nil {
			t.Fatal(err)
		}
		if result.Roll < 1 || result.Roll > 100 {
			t.Fatalf("骰出 %d，1d100 的範圍是 1..100", result.Roll)
		}
		if previous, ok := seen[result.Roll]; ok && previous != result.Outcome {
			t.Fatalf("骰 %d 對到兩種結果：%v 與 %v", result.Roll, previous, result.Outcome)
		}
		seen[result.Roll] = result.Outcome
		want := MoraleEnraged
		switch {
		case result.Roll <= 10:
			want = MoraleRunsAway
		case result.Roll <= 60:
			want = MoraleConfused
		case result.Roll <= 80:
			want = MoraleBerserk
		}
		if result.Outcome != want {
			t.Fatalf("骰 %d → %v，want %v", result.Roll, result.Outcome, want)
		}
		if result.MessageID == "" {
			t.Fatalf("骰 %d 沒有訊息鍵", result.Roll)
		}
	}
	if len(seen) < 90 {
		t.Fatalf("600 次只蓋到 %d 種骰值，分佈太窄，這條測試沒在測東西", len(seen))
	}
}

// 逃走那一段做的五件事，逐項檢查。
func TestRunningAwayLeavesTheFieldAndRefillsMorale(t *testing.T) {
	checked := false
	for seed := int64(1); seed <= 200 && !checked; seed++ {
		battle := moraleBattle(t, seed)
		before := battle.fighters["orc"]
		before.CombatAction.SetActionTarget("hero")
		battle.fighters["orc"] = before
		result, err := battle.ResolveMoraleFailure("orc")
		if err != nil {
			t.Fatal(err)
		}
		if result.Outcome != MoraleRunsAway {
			continue
		}
		checked = true
		after := battle.fighters["orc"]
		if !after.Escaped {
			t.Error("逃走了卻還在場上")
		}
		if after.RawCombatState10 != 1 {
			t.Errorf("`+18Dh^[10h]` ＝ %d，want 1", after.RawCombatState10)
		}
		if after.ControlMorale != 0xB3 {
			t.Errorf("士氣 ＝ %02Xh，want B3h（最高位沒設就補滿）", after.ControlMorale)
		}
		if after.CombatAction.ActionTargetID != "" {
			t.Error("目前目標沒被清掉")
		}
		if !fighterHasAnyAffect(after, []uint8{moraleFleeEffectKind}) {
			t.Error("沒掛上效果碼 23h")
		}
	}
	if !checked {
		t.Fatal("200 次都沒擲到 1..10——10% 的段落抽不到，測試等於沒跑")
	}
}

// 最高位已經設起來的士氣**不覆寫**（原作是 `if +0F7h <= 7Fh`）。
func TestRunningAwayLeavesAnAlreadyValidMoraleAlone(t *testing.T) {
	for seed := int64(1); seed <= 200; seed++ {
		battle := moraleBattle(t, seed)
		fighter := battle.fighters["orc"]
		fighter.ControlMorale = 0x8A
		battle.fighters["orc"] = fighter
		result, err := battle.ResolveMoraleFailure("orc")
		if err != nil {
			t.Fatal(err)
		}
		if result.Outcome != MoraleRunsAway {
			continue
		}
		if got := battle.fighters["orc"].ControlMorale; got != 0x8A {
			t.Fatalf("士氣被改成 %02Xh，8Ah 的最高位已經設了不該覆寫", got)
		}
		return
	}
	t.Fatal("200 次都沒擲到逃走那一段")
}

// 另外三段**不動戰鬥員的狀態**——原作是各自呼叫別支處理，那幾支還沒解讀。
func TestTheOtherThreeOutcomesDoNotTouchTheFighter(t *testing.T) {
	for seed := int64(1); seed <= 200; seed++ {
		battle := moraleBattle(t, seed)
		before := battle.fighters["orc"]
		result, err := battle.ResolveMoraleFailure("orc")
		if err != nil {
			t.Fatal(err)
		}
		if result.Outcome == MoraleRunsAway {
			continue
		}
		after := battle.fighters["orc"]
		if after.Escaped || after.RawCombatState10 != before.RawCombatState10 ||
			after.ControlMorale != before.ControlMorale || len(after.MonsterAffects) != len(before.MonsterAffects) {
			t.Fatalf("%v 動到了戰鬥員狀態", result.Outcome)
		}
	}
}

// 敵人全部逃光就是隊伍勝利——`Escaped` 不算「還在場上」。
func TestBattleEndsWhenEveryEnemyRunsAway(t *testing.T) {
	battle := moraleBattle(t, 3)
	for _, id := range []string{"orc", "orc2"} {
		fighter := battle.fighters[id]
		fighter.Escaped = true
		battle.fighters[id] = fighter
	}
	battle.updateStatus()
	if battle.Status() != StatusPartyWon {
		t.Fatalf("status=%v，敵人跑光了該算隊伍勝利", battle.Status())
	}
}
