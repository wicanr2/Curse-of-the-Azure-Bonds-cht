package monster

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

// weaponCatalog：`0x10` 是槽 0 的近戰武器、`0x20` 是槽 10 的弓、`0x30` 是槽 1 的盾。
func weaponCatalog() BaseItemCatalog {
	items := make([]BaseItem, 128)
	for index := range items {
		items[index] = BaseItem{Type: uint8(index), Slot: 0x0F}
	}
	items[0x10] = BaseItem{Type: 0x10, Slot: 0,
		SmallDamageDice: 1, SmallDamageSides: 6, SmallDamageBonus: 0,
		LargeDamageDice: 1, LargeDamageSides: 8, LargeDamageBonus: 1}
	items[0x20] = BaseItem{Type: 0x20, Slot: 10,
		SmallDamageDice: 1, SmallDamageSides: 4, LargeDamageDice: 1, LargeDamageSides: 4}
	items[0x30] = BaseItem{Type: 0x30, Slot: 1}
	return BaseItemCatalog{Items: items}
}

// ★★ 槽 0 的武器**無條件**蓋掉記錄的骰子——記錄那組是放下武器時的天生攻擊
// （spec 1174）。`BUGBEAR` 就是這個形狀：記錄 `2d4`、武器 `1d6`。
func TestMonsterWeaponOverridesTheRecordDice(t *testing.T) {
	fighter := combat.Fighter{
		ID: "m1", DamageDiceCount: 2, DamageDiceSides: 4,
		MonsterItems: []combat.MonsterItem{
			{Name: "Bow", Type: 0x20, Readied: true},
			{Name: "Morning Star", Type: 0x10, Readied: true, Plus: 1},
		},
	}
	projected := ProjectMonsterWeapon(fighter, weaponCatalog())
	if projected.DamageDiceCount != 1 || projected.DamageDiceSides != 6 ||
		projected.DamageBonus != 1 {
		t.Fatalf("小型傷害 ＝ %dd%d＋%d，want 1d6＋1",
			projected.DamageDiceCount, projected.DamageDiceSides, projected.DamageBonus)
	}
	if projected.LargeDamageDiceCount != 1 || projected.LargeDamageDiceSides != 8 ||
		projected.LargeDamageBonus != 2 {
		t.Fatalf("大型傷害 ＝ %dd%d＋%d，want 1d8＋2",
			projected.LargeDamageDiceCount, projected.LargeDamageDiceSides,
			projected.LargeDamageBonus)
	}
	if !projected.HasSlotZeroWeapon {
		t.Fatal("有槽 0 的武器就要標起來，大／小體型的切換靠它")
	}
}

// 只有非槽 0 的裝備時整支不動——原作 `sub_25` 第一行就返回。
func TestMonsterWithoutSlotZeroWeaponKeepsItsNaturalAttack(t *testing.T) {
	fighter := combat.Fighter{
		ID: "m2", DamageDiceCount: 2, DamageDiceSides: 8, DamageBonus: 3,
		MonsterItems: []combat.MonsterItem{
			{Name: "Shield", Type: 0x30, Readied: true},
			{Name: "Sword", Type: 0x10}, // 沒有裝備中
		},
	}
	projected := ProjectMonsterWeapon(fighter, weaponCatalog())
	if projected.DamageDiceCount != 2 || projected.DamageDiceSides != 8 ||
		projected.DamageBonus != 3 {
		t.Fatalf("天生攻擊被動到了：%dd%d＋%d",
			projected.DamageDiceCount, projected.DamageDiceSides, projected.DamageBonus)
	}
	if projected.HasSlotZeroWeapon {
		t.Fatal("沒有槽 0 的武器不該標起來")
	}
}

// ★ 六章的怪物資料裡，記錄骰與武器骰不一樣的那 13 隻是真實存在的形狀。
// 這一條把「投影之後打出來的是武器的骰」釘在真實資料上。
func TestBugbearShapeUsesTheWeaponNotTheRecord(t *testing.T) {
	fighter := combat.Fighter{
		ID: "bugbear", DamageDiceCount: 2, DamageDiceSides: 4,
		MonsterItems: []combat.MonsterItem{{Name: "Weapon", Type: 0x10, Readied: true}},
	}
	projected := ProjectMonsterWeapon(fighter, weaponCatalog())
	small := combat.Fighter{}
	large := combat.Fighter{LargeTarget: true}
	if count, sides, _ := projected.WeaponDamageAgainst(small); count != 1 || sides != 6 {
		t.Fatalf("打小型 ＝ %dd%d，want 1d6", count, sides)
	}
	if count, sides, _ := projected.WeaponDamageAgainst(large); count != 1 || sides != 8 {
		t.Fatalf("打大型 ＝ %dd%d，want 1d8", count, sides)
	}
}
