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

func moraleCheckBattle(t *testing.T, morale uint8, hp, maxHP, speed, enemySpeed int) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "orc", Side: SideEnemy, HitPoints: hp, MaxHitPoints: maxHP,
			ControlMorale: morale, MovementAllowance: speed},
		{ID: "hero", Side: SideParty, HitPoints: 12, MaxHitPoints: 12,
			MovementAllowance: enemySpeed},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 門檻是 `100 − HP%`：滿血幾乎一定過，重傷才會崩。
func TestMoraleThresholdTracksHowHurtTheFighterIs(t *testing.T) {
	for _, item := range []struct {
		name      string
		hp        int
		threshold int
		passed    bool
	}{
		{"滿血", 20, 0, true},
		{"半血", 10, 50, true}, // 士氣 102 ≥ 50
		{"剩一成", 2, 90, true}, // 102 ≥ 90
	} {
		battle := moraleCheckBattle(t, 0xB3, item.hp, 20, 6, 6)
		result, err := battle.CheckMorale("orc")
		if err != nil {
			t.Fatal(err)
		}
		if !result.Checked || result.Threshold != item.threshold || result.Passed != item.passed {
			t.Errorf("%s：%+v，want 門檻 %d、過關 %v", item.name, result, item.threshold, item.passed)
		}
	}
	// 士氣 32（90h）在剩一成血時過不了。
	battle := moraleCheckBattle(t, 0x90, 2, 20, 6, 6)
	result, err := battle.CheckMorale("orc")
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed {
		t.Fatalf("士氣 %d、門檻 %d 卻過關了", result.Morale, result.Threshold)
	}
}

// 沒有士氣資料（最高位沒設）就不檢定——**不是當成 0**。
func TestMissingMoraleDataSkipsTheCheckInsteadOfFailingIt(t *testing.T) {
	battle := moraleCheckBattle(t, 0x33, 1, 20, 6, 6)
	result, err := battle.CheckMorale("orc")
	if err != nil {
		t.Fatal(err)
	}
	if result.Checked || result.Withdrew {
		t.Fatalf("沒有士氣資料卻做了檢定：%+v", result)
	}
}

// 崩了之後跑不跑得掉看移動率，而且是**折半之後**比。
func TestBrokenMoraleWithdrawsOnlyWhenFastEnough(t *testing.T) {
	slow := moraleCheckBattle(t, 0x90, 2, 20, 4, 12)
	slowResult, err := slow.CheckMorale("orc")
	if err != nil {
		t.Fatal(err)
	}
	if slowResult.Withdrew {
		t.Fatal("移動率 4 對上 12，跑不掉才對")
	}
	fast := moraleCheckBattle(t, 0x90, 2, 20, 12, 4)
	fastResult, err := fast.CheckMorale("orc")
	if err != nil {
		t.Fatal(err)
	}
	if !fastResult.Withdrew {
		t.Fatal("移動率 12 對上 4，該跑得掉")
	}
	if !fast.fighters["orc"].RetreatedThisRound {
		t.Fatal("`+14h` 沒有被設起來")
	}
	// 整數除法：5 與 4 折半之後都是 2，算跑得掉。
	tie := moraleCheckBattle(t, 0x90, 2, 20, 4, 5)
	tieResult, err := tie.CheckMorale("orc")
	if err != nil {
		t.Fatal(err)
	}
	if !tieResult.Withdrew {
		t.Fatal("4 與 5 折半之後都是 2，原作的整數除法算跑得掉")
	}
}

// 每回合開頭要清掉上一回合的 `+14h`。
func TestMoraleCheckClearsLastRoundsRetreatFlag(t *testing.T) {
	battle := moraleCheckBattle(t, 0xB3, 20, 20, 6, 6)
	fighter := battle.fighters["orc"]
	fighter.RetreatedThisRound = true
	battle.fighters["orc"] = fighter
	if _, err := battle.CheckMorale("orc"); err != nil {
		t.Fatal(err)
	}
	if battle.fighters["orc"].RetreatedThisRound {
		t.Fatal("上一回合的撤退旗標沒清掉，訊息會每回合重印")
	}
}

// 恐慌訊息只在「本回合撤退」而且「還沒正式逃走」時印。
func TestPanicMessageNeedsBothConditions(t *testing.T) {
	for _, item := range []struct {
		name      string
		retreated bool
		fled      uint8
		want      bool
	}{
		{"本回合撤退、還沒逃走", true, 0, true},
		{"本回合撤退、但已經逃走了", true, 1, false},
		{"沒有撤退", false, 0, false},
	} {
		_, got := Fighter{RetreatedThisRound: item.retreated, RawCombatState10: item.fled}.PanicMessageID()
		if got != item.want {
			t.Errorf("%s：got %v，want %v", item.name, got, item.want)
		}
	}
}

// 混亂術（效果碼 `23h`）走的就是士氣崩潰那張四段表——這支法術不需要新規則。
func TestAreaMoraleBreakRunsTheFourWayTableOnFailedSaves(t *testing.T) {
	build := func(threshold uint8) *Battle {
		battle, err := NewBattle([]Fighter{
			{ID: "mage", Side: SideParty, HitPoints: 20, MaxHitPoints: 20,
				HasCombatPosition: true, CombatX: 1, CombatY: 1},
			{ID: "orc-a", Side: SideEnemy, HitPoints: 12, MaxHitPoints: 12,
				ControlMorale:     0x33,
				SavingThrows:      []uint8{threshold, threshold, threshold, threshold, threshold},
				HasCombatPosition: true, CombatX: 8, CombatY: 8},
			{ID: "orc-b", Side: SideEnemy, HitPoints: 12, MaxHitPoints: 12,
				ControlMorale:     0x33,
				SavingThrows:      []uint8{threshold, threshold, threshold, threshold, threshold},
				HasCombatPosition: true, CombatX: 9, CombatY: 9},
			{ID: "orc-far", Side: SideEnemy, HitPoints: 12, MaxHitPoints: 12,
				ControlMorale:     0x33,
				SavingThrows:      []uint8{threshold, threshold, threshold, threshold, threshold},
				HasCombatPosition: true, CombatX: 25, CombatY: 2},
		}, 21)
		if err != nil {
			t.Fatal(err)
		}
		return battle
	}
	// 門檻 99 ⇒ 幾乎一定沒過（天然 20 除外）。
	failing, err := build(99).CastAreaMoraleBreak("mage", TilePoint{X: 8, Y: 8}, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(failing.Impacts) != 2 {
		t.Fatalf("命中 %d 個目標，半徑 3 應該只蓋到貼近的兩個", len(failing.Impacts))
	}
	for _, impact := range failing.Impacts {
		if impact.Saved {
			continue
		}
		if impact.Outcome < MoraleRunsAway || impact.Outcome > MoraleEnraged {
			t.Fatalf("%s 沒過豁免卻沒有四段表的結果：%+v", impact.TargetID, impact)
		}
		if impact.MessageID == "" {
			t.Fatalf("%s 沒有訊息鍵", impact.TargetID)
		}
	}
	// 門檻 1 ⇒ 幾乎一定過，過了就不跑那張表。
	saving, err := build(1).CastAreaMoraleBreak("mage", TilePoint{X: 8, Y: 8}, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, impact := range saving.Impacts {
		if impact.Saved && impact.MessageID != "" {
			t.Fatalf("%s 過了豁免卻還是跑了四段表", impact.TargetID)
		}
	}
}
