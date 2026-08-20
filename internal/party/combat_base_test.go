package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

func combatBaseTables(t *testing.T) CombatBaseLookup {
	t.Helper()
	tables, err := gamepack.CombatBase()
	if err != nil {
		t.Fatal(err)
	}
	return tables
}

// 命中能力表（`DS:3E38h`）逐格對回 AD&D 1e 的攻擊進程。
// 職業槽的名字來自 game pack 的 `class_slots`，不是這裡自己排的。
func TestClassAttackTableMatchesTheOriginalProgressions(t *testing.T) {
	tables := combatBaseTables(t)
	cases := []struct {
		slot  int
		level int
		thac0 int
	}{
		{0, 1, 20}, {0, 3, 20}, {0, 4, 18}, {0, 7, 16}, {0, 10, 14}, // 牧師
		{2, 1, 20}, {2, 2, 20}, {2, 3, 18}, {2, 4, 17}, {2, 12, 9}, // 戰士
		{5, 1, 21}, {5, 5, 21}, {5, 6, 19}, {5, 11, 17}, // 法師
		{6, 1, 20}, {6, 5, 19}, {6, 9, 16}, // 盜賊
	}
	for _, test := range cases {
		stored, ok := tables.ClassAttackValue(test.slot, test.level)
		if !ok {
			t.Fatalf("職業槽 %d 等級 %d 查不到", test.slot, test.level)
		}
		if got := 60 - stored; got != test.thac0 {
			t.Errorf("職業槽 %d 等級 %d 的 THAC0 是 %d，預期 %d", test.slot, test.level, got, test.thac0)
		}
	}
}

// 八個職業槽全部跑過再取最大值。空槽是跳過不是提早結束——聖騎士只有槽 3
// 有等級，讀成「碰到 0 就 break」會在槽 0 就結束、什麼都算不出來。
func TestAttackAbilitySkipsEmptySlotsAndTakesTheBest(t *testing.T) {
	tables := combatBaseTables(t)
	paladin := [8]uint8{0, 0, 0, 1, 0, 0, 0, 0}
	value, err := AttackAbilityFrom(tables, paladin)
	if err != nil {
		t.Fatal(err)
	}
	if 60-int(value) != 20 {
		t.Fatalf("一級聖騎士的 THAC0 是 %d，預期 20", 60-int(value))
	}
	// 法師 5 級（THAC0 21）配戰士 3 級（THAC0 18）⇒ 取比較好的 18。
	multi := [8]uint8{0, 0, 3, 0, 0, 5, 0, 0}
	value, err = AttackAbilityFrom(tables, multi)
	if err != nil {
		t.Fatal(err)
	}
	if 60-int(value) != 18 {
		t.Fatalf("多職角色的 THAC0 是 %d，預期取最好的 18", 60-int(value))
	}
	if _, err := AttackAbilityFrom(tables, [8]uint8{}); err == nil {
		t.Fatal("八個槽都 0 級應該回錯誤，不是回一個看起來合理的值")
	}
}

// 帶了原作欄位的角色，AC 與命中加值走原作的算式；沒帶的退回本作近似。
// 建角預設 `+124h = 32h` ⇒ AC 10，敏捷 18 的防禦調整 −4 ⇒ AC 6。
func TestFighterUsesTheOriginalFieldsWhenPresent(t *testing.T) {
	character := Character{
		ID: "pc", Name: "測試", Race: RaceHuman, Class: ClassFighter, Level: 1,
		Abilities: Abilities{Strength: 10, Intelligence: 10, Wisdom: 10,
			Dexterity: 18, Constitution: 10, Charisma: 10},
	}
	approximate, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	character.AttackAbility = 0x2C // THAC0 16
	character.BaseArmorClass = 0x32
	exact, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	if exact.ArmorClass != 6 {
		t.Fatalf("AC 是 %d，預期 6（10 − 敏捷 18 的防禦調整 4）", exact.ArmorClass)
	}
	if exact.AttackBonus != combat.DisplayAttackBonus(0x2C) || exact.AttackBonus != 4 {
		t.Fatalf("命中加值是 %d，預期 4（THAC0 16）", exact.AttackBonus)
	}
	if approximate.ArmorClass == exact.ArmorClass && approximate.AttackBonus == exact.AttackBonus {
		t.Fatal("近似式與原作算式給出同一組值，這組樣本測不出差別")
	}
}
