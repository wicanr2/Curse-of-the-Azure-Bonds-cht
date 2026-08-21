package combat

import "testing"

// specChargedItems 是 spec 1169 那張表的獨立謄本（骰子與形狀），故意不共用
// `chargedItemBehaviours`——兩份各自出錯的機率才低。
var specChargedItems = map[uint8]struct {
	shape  ChargedItemShape
	dice   string
	center string
}{
	0x2F: {shape: ChargedItemNamedSpell},
	0x33: {shape: ChargedItemNamedSpell},
	0x57: {shape: ChargedItemNamedSpell},
	0x39: {shape: ChargedItemEffect},
	0x3B: {shape: ChargedItemGiantStrength},
	0x3D: {shape: ChargedItemEffect},
	0x3F: {shape: ChargedItemEffect},
	0x40: {shape: ChargedItemAreaDamage, dice: "?d6", center: "aim"},
	0x41: {shape: ChargedItemDamage, dice: "2d4+2"},
	0x61: {shape: ChargedItemEffect},
	0x62: {shape: ChargedItemAreaDamage, dice: "6d6", center: "user"},
	0x63: {shape: ChargedItemHeal, dice: "2d4+2"},
}

func TestChargedItemBehavioursMatchTheHandlerReadings(t *testing.T) {
	if len(chargedItemBehaviours) != len(specChargedItems) {
		t.Fatalf("表有 %d 筆，謄本有 %d 筆", len(chargedItemBehaviours), len(specChargedItems))
	}
	for effect, want := range specChargedItems {
		got, ok := LookupChargedItemBehaviour(effect)
		if !ok {
			t.Errorf("效果 %02Xh 不在表裡", effect)
			continue
		}
		if got.Shape != want.shape {
			t.Errorf("效果 %02Xh 形狀 ＝ %d，want %d", effect, got.Shape, want.shape)
		}
		center := ""
		if got.Shape == ChargedItemAreaDamage {
			center = "aim"
			if got.CenterOnUser {
				center = "user"
			}
		}
		if center != want.center {
			t.Errorf("效果 %02Xh 圓心 ＝ %q，want %q", effect, center, want.center)
		}
		if want.dice == "" {
			continue
		}
		shown := "?d" + itoa(got.Dice.Sides)
		if !got.RandomDiceCount {
			shown = itoa(got.Dice.Count) + "d" + itoa(got.Dice.Sides)
			if got.Dice.Bonus != 0 {
				shown += "+" + itoa(got.Dice.Bonus)
			}
		}
		if shown != want.dice {
			t.Errorf("效果 %02Xh 骰子 ＝ %q，want %q", effect, shown, want.dice)
		}
	}
}

// ⚠ 「表裡沒有」代表**那一支 handler 還沒讀**，不是「沒有效果」。
// 呼叫端要看得出差別，所以查詢一定要回兩個值。
func TestUnknownChargedItemEffectIsReportedAsUnread(t *testing.T) {
	if _, ok := LookupChargedItemBehaviour(0x01); ok {
		t.Fatal("`01h` 不是物品效果，不該查得到")
	}
}

// 飛彈項鍊的骰數是 `(1d3 × 2 ＋ 1)`：只可能是 3、5、7 顆。
func TestNecklaceDiceCountIsOddAndBounded(t *testing.T) {
	behaviour, ok := LookupChargedItemBehaviour(0x40)
	if !ok {
		t.Fatal("`40h` 不在表裡")
	}
	seen := map[int]bool{}
	for seed := int64(1); seed <= 60; seed++ {
		battle, err := NewBattle([]Fighter{{ID: "a", Side: SideParty, HitPoints: 5, MaxHitPoints: 5}}, seed)
		if err != nil {
			t.Fatal(err)
		}
		dice, rollErr := battle.RollChargedItemDice(behaviour)
		if rollErr != nil {
			t.Fatal(rollErr)
		}
		if dice.Sides != 6 {
			t.Fatalf("面數 ＝ %d，want 6", dice.Sides)
		}
		if dice.Count != 3 && dice.Count != 5 && dice.Count != 7 {
			t.Fatalf("骰數 ＝ %d，只可能是 3／5／7", dice.Count)
		}
		seen[dice.Count] = true
	}
	if len(seen) != 3 {
		t.Fatalf("60 顆種子只擲出 %d 種骰數：%v", len(seen), seen)
	}
}
