package combat

import "testing"

func effectBattle(t *testing.T, saves []uint8) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "caster", Side: SideParty, HitPoints: 10, MaxHitPoints: 10},
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			SavingThrows: saves},
	}, 4)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 效果、持續時間、豁免類別全部由呼叫端從法術主表帶進來；這一層只負責記上去。
func TestCastEffectSpellRecordsTheEffectWithItsDuration(t *testing.T) {
	battle := effectBattle(t, []uint8{20, 20, 20, 20, 20})
	result, err := battle.CastEffectSpell("caster", []string{"orc"}, EffectSpellRequest{
		SpellID: 23, EffectKind: 0x34, Duration: 6, SaveKind: 1, SaveCategory: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Impacts) != 1 || !result.Impacts[0].Rolled {
		t.Fatalf("result=%+v，`+8h ≠ 0` 應該擲豁免", result)
	}
	orc, _ := battle.Fighter("orc")
	if !result.Impacts[0].Applied {
		if len(orc.MonsterAffects) != 0 {
			t.Fatalf("豁免成功卻還是記了效果：%+v", orc.MonsterAffects)
		}
		return
	}
	if len(orc.MonsterAffects) != 1 {
		t.Fatalf("affects=%+v", orc.MonsterAffects)
	}
	affect := orc.MonsterAffects[0]
	// Duration 與 Value 兩格都要寫；Strength 不是時間（0FFh 是永久標記）。
	if affect.Kind != 0x34 || affect.Duration != 6 || affect.Value != 6 || !affect.Active {
		t.Fatalf("affect=%+v", affect)
	}
	if affect.Strength == 0xFF {
		t.Fatal("Strength 被寫成 0FFh，那是「永久」的標記，效果永遠不會過期")
	}
	if !orc.MonsterIsHeld() {
		t.Fatal("效果碼 0x34 記上去了，MonsterIsHeld 卻不認")
	}
}

// `+8h = 0` 的法術完全不擲豁免——不是「擲了但一定失敗」。
func TestCastEffectSpellSkipsTheSaveWhenTheTableSaysZero(t *testing.T) {
	battle := effectBattle(t, []uint8{2, 2, 2, 2, 2})
	result, err := battle.CastEffectSpell("caster", []string{"orc"}, EffectSpellRequest{
		SpellID: 48, EffectKind: 0x27, Duration: 6, SaveKind: 0, SaveCategory: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Impacts[0].Rolled || !result.Impacts[0].Applied {
		t.Fatalf("impact=%+v，`+8h = 0` 不擲豁免且一定套上", result.Impacts[0])
	}
}

// 沒有豁免門檻的目標算**失敗**，不是自動成功——怪物記錄缺豁免值時該是弱點。
func TestMissingSavingThrowCountsAsAFailedSave(t *testing.T) {
	battle := effectBattle(t, nil)
	result, err := battle.CastEffectSpell("caster", []string{"orc"}, EffectSpellRequest{
		SpellID: 23, EffectKind: 0x34, Duration: 6, SaveKind: 1, SaveCategory: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Impacts[0].Saved || !result.Impacts[0].Applied {
		t.Fatalf("impact=%+v，沒有豁免門檻應該算失敗", result.Impacts[0])
	}
}

// 效果碼 0 不該走這條路——那代表呼叫端挑錯了法術。
func TestCastEffectSpellRejectsSpellsWithoutAnEffect(t *testing.T) {
	battle := effectBattle(t, []uint8{10, 10, 10, 10, 10})
	if _, err := battle.CastEffectSpell("caster", []string{"orc"}, EffectSpellRequest{
		SpellID: 47, EffectKind: 0, Duration: 6,
	}); err == nil {
		t.Fatal("效果碼 0 應該回錯誤")
	}
}
