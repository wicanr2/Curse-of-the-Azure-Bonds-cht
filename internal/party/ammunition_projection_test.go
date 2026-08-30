package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

func ammunitionProjectionCatalog(t *testing.T, flags uint8) monster.BaseItemCatalog {
	t.Helper()
	data := make([]byte, monster.BaseItemHeaderSize+0x4A*monster.BaseItemRecordSize)
	set := func(index int, fields map[int]byte) {
		offset := monster.BaseItemHeaderSize + index*monster.BaseItemRecordSize
		for position, value := range fields {
			data[offset+position] = value
		}
	}
	set(2, map[int]byte{0: 0, 5: 4, 9: 1, 10: 6, 13: 0xFF, 14: flags})
	set(11, map[int]byte{0: 11, 13: 0xFF}) // 卷軸槽正對照。
	set(0x1C, map[int]byte{0: 10, 13: 0xFF})
	set(0x49, map[int]byte{0: 10, 13: 0xFF})
	catalog, err := monster.ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func ammunitionProjectionCharacter(equipment []monster.ItemRecord) Character {
	return Character{
		ID: "archer", Name: "弓手", Race: RaceHuman, Class: ClassFighter, Level: 7,
		HitPoints: 20, MaxHitPoints: 20, HealthStatus: HealthStatusOK,
		Abilities: Abilities{Strength: 16, Intelligence: 10, Wisdom: 10,
			Dexterity: 12, Constitution: 14, Charisma: 10},
		Equipment: equipment,
	}
}

func TestFighterWithEquipmentProjectsArrowPointerByItemType(t *testing.T) {
	character := ammunitionProjectionCharacter([]monster.ItemRecord{
		{Name: "弓", Type: 2, Readied: true, Count: 1},
		{Name: "舊箭疊", Type: 0x49, Count: 7},
		{Name: "卷軸", Type: 11, Readied: true, Count: 30},
		{Name: "新箭疊", Type: 0x49, Count: 2},
	})
	fighter, err := character.FighterWithEquipment(ammunitionProjectionCatalog(t, 0x09))
	if err != nil {
		t.Fatal(err)
	}
	if fighter.AmmunitionCount != 2 {
		t.Fatalf("彈藥數量 = %d，want 鏈上最後一疊 Arrow 的 2（不能取卷軸或加總）",
			fighter.AmmunitionCount)
	}
}

func TestFighterWithEquipmentKeepsArrowAndQuarrelPointersSeparate(t *testing.T) {
	equipment := []monster.ItemRecord{
		{Name: "弩", Type: 2, Readied: true, Count: 1},
		{Name: "箭", Type: 0x49, Count: 9},
		{Name: "弩矢", Type: 0x1C, Count: 3},
	}
	fighter, err := ammunitionProjectionCharacter(equipment).
		FighterWithEquipment(ammunitionProjectionCatalog(t, 0x88))
	if err != nil {
		t.Fatal(err)
	}
	if fighter.AmmunitionCount != 3 {
		t.Fatalf("bit 7 武器的彈藥數量 = %d，want Quarrel 的 3", fighter.AmmunitionCount)
	}
}
