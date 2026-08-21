package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// weaponSlotCatalog：`0x10` 是槽 0 的近戰武器（`1d8`），`0x20` 是槽 10 的弓
// （`1d6`）。原作的派生值重算只看槽 0（spec 1174）。
func weaponSlotCatalog() monster.BaseItemCatalog {
	items := make([]monster.BaseItem, 128)
	for index := range items {
		items[index] = monster.BaseItem{Type: uint8(index), Slot: 0x0F}
	}
	items[0x10] = monster.BaseItem{Type: 0x10, Slot: 0,
		SmallDamageDice: 1, SmallDamageSides: 8, LargeDamageDice: 1, LargeDamageSides: 12}
	items[0x20] = monster.BaseItem{Type: 0x20, Slot: 10, RateOfFire: 4,
		SmallDamageDice: 1, SmallDamageSides: 6, LargeDamageDice: 1, LargeDamageSides: 6}
	return monster.BaseItemCatalog{Items: items}
}

// ★★ 弓排在鏈的前面也不能搶走傷害骰：原作讀的是 `+151h[0]`（槽 0），
// 不是「第一件有骰的裝備中物品」。
func TestFighterWithEquipmentTakesTheSlotZeroWeapon(t *testing.T) {
	character := Character{
		ID: "p1", Name: "戰士", Race: RaceHuman, Class: ClassFighter, Level: 3,
		Abilities: Abilities{Strength: 12, Intelligence: 10, Wisdom: 10,
			Dexterity: 10, Constitution: 10, Charisma: 10},
		Equipment: []monster.ItemRecord{
			{Name: "Bow", Type: 0x20, Readied: true},
			{Name: "Sword", Type: 0x10, Readied: true},
		},
	}
	fighter, err := character.FighterWithEquipment(weaponSlotCatalog())
	if err != nil {
		t.Fatal(err)
	}
	if fighter.DamageDiceCount != 1 || fighter.DamageDiceSides != 8 {
		t.Fatalf("傷害骰 ＝ %dd%d，want 1d8（劍，槽 0）",
			fighter.DamageDiceCount, fighter.DamageDiceSides)
	}
	if fighter.MissileWeapon {
		t.Fatal("槽 0 是近戰武器，不該被弓的遠程旗標蓋掉")
	}
}

// 沒有槽 0 的武器時沿用舊規則——原作在那個情況下什麼都不做，但 remake 還沒有
// 徒手的完整模型，拿掉會讓只帶弓的角色沒有攻擊（spec 1174 的明確不宣稱）。
func TestFighterWithEquipmentFallsBackWhenNoSlotZeroWeapon(t *testing.T) {
	character := Character{
		ID: "p1", Name: "遊俠", Race: RaceHuman, Class: ClassFighter, Level: 3,
		Abilities: Abilities{Strength: 12, Intelligence: 10, Wisdom: 10,
			Dexterity: 10, Constitution: 10, Charisma: 10},
		Equipment: []monster.ItemRecord{{Name: "Bow", Type: 0x20, Readied: true}},
	}
	fighter, err := character.FighterWithEquipment(weaponSlotCatalog())
	if err != nil {
		t.Fatal(err)
	}
	if fighter.DamageDiceCount != 1 || fighter.DamageDiceSides != 6 {
		t.Fatalf("傷害骰 ＝ %dd%d，want 1d6（弓）",
			fighter.DamageDiceCount, fighter.DamageDiceSides)
	}
}
