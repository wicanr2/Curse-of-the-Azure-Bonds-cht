package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 目錄：類別 0 ＝ 短劍（1d6、單手）、1 ＝ 長劍（1d8、單手）、
// 2 ＝ 弓（1d6、雙手、射速 4、射程 20、要彈藥槽 A）、11 ＝ 箭（彈藥槽 A）。
func autoEquipTestCatalog(t *testing.T) monster.BaseItemCatalog {
	t.Helper()
	data := make([]byte, monster.BaseItemHeaderSize+12*monster.BaseItemRecordSize)
	set := func(index int, fields map[int]byte) {
		offset := monster.BaseItemHeaderSize + index*monster.BaseItemRecordSize
		for position, value := range fields {
			data[offset+position] = value
		}
	}
	set(0, map[int]byte{0: 0, 1: 1, 9: 1, 10: 6, 13: 0xFF})
	set(1, map[int]byte{0: 0, 1: 1, 9: 1, 10: 8, 13: 0xFF})
	set(2, map[int]byte{0: 0, 1: 2, 5: 4, 9: 1, 10: 6, 12: 20, 13: 0xFF, 14: 0x09})
	set(11, map[int]byte{0: 11, 1: 0, 13: 0xFF})
	catalog, err := monster.ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func autoEquipState(t *testing.T, equipment []monster.ItemRecord, enemyX int) *State {
	t.Helper()
	value := NewState(testCatalog())
	state := &value
	state.SetItemCatalog(autoEquipTestCatalog(t))
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄",
		Class: party.ClassFighter, Level: 3, HitPoints: 20, MaxHitPoints: 20,
		// 戰士在 ClassLevels 裡的索引是 2（`HasClass` 的對照）。
		ClassLevels:  [8]uint8{2: 3},
		HealthStatus: party.HealthStatusOK,
		Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10,
			Dexterity: 12, Constitution: 14, Charisma: 10},
		Equipment: equipment}}
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, MovementAllowance: 6,
		HasCombatPosition: true, CombatX: 1, CombatY: 1}}
	enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy,
		HitPoints: 12, MaxHitPoints: 12, ArmorClass: 10, MovementAllowance: 6,
		HasCombatPosition: true, CombatX: enemyX, CombatY: 1}}
	if err := state.StartCombat(partyFighters, enemies, 3); err != nil {
		t.Fatal(err)
	}
	return state
}

func readiedType(t *testing.T, state *State) (uint8, bool) {
	t.Helper()
	for _, item := range state.partyRoster[0].Equipment {
		if item.Readied {
			base, ok := state.itemCatalog.Lookup(item.Type)
			if ok && base.Slot == 0 {
				return item.Type, true
			}
		}
	}
	return 0, false
}

// 拿著短劍、背包裡有長劍：AI 回合會換成長劍。
func TestAutoEquipSwapsToTheHigherScoringWeapon(t *testing.T) {
	state := autoEquipState(t, []monster.ItemRecord{
		{Name: "短劍", Type: 0, Readied: true, Count: 1},
		{Name: "長劍", Type: 1, Count: 1},
	}, 8)
	swapped, err := state.autoEquipBeforeAITurn("hero")
	if err != nil {
		t.Fatal(err)
	}
	if !swapped {
		t.Fatal("長劍分數比短劍高，該換")
	}
	if got, ok := readiedType(t, state); !ok || got != 1 {
		t.Fatalf("裝備中的武器類別 %d（found=%v），want 1（長劍）", got, ok)
	}
}

// 詛咒的武器換不掉（spec 1004 §四）。
func TestAutoEquipLeavesACursedWeaponAlone(t *testing.T) {
	state := autoEquipState(t, []monster.ItemRecord{
		{Name: "詛咒短劍", Type: 0, Readied: true, Cursed: true, Count: 1},
		{Name: "長劍", Type: 1, Count: 1},
	}, 8)
	swapped, err := state.autoEquipBeforeAITurn("hero")
	if err != nil {
		t.Fatal(err)
	}
	if swapped {
		t.Fatal("詛咒的裝備被換掉了")
	}
	if got, _ := readiedType(t, state); got != 0 {
		t.Fatalf("裝備中的武器類別 %d，詛咒的那把該留著", got)
	}
}

// 敵人貼身就不會換成弓；離得遠、而且有箭，才會。
func TestAutoEquipTakesTheBowOnlyWhenNobodyIsAdjacentAndAmmunitionIsReadied(t *testing.T) {
	equipment := func() []monster.ItemRecord {
		return []monster.ItemRecord{
			{Name: "短劍", Type: 0, Readied: true, Count: 1},
			{Name: "弓", Type: 2, Count: 1},
			{Name: "箭", Type: 11, Readied: true, Count: 20},
		}
	}
	adjacent := autoEquipState(t, equipment(), 2)
	if _, err := adjacent.autoEquipBeforeAITurn("hero"); err != nil {
		t.Fatal(err)
	}
	if got, _ := readiedType(t, adjacent); got == 2 {
		t.Fatal("敵人貼身還是換成弓")
	}
	far := autoEquipState(t, equipment(), 9)
	if _, err := far.autoEquipBeforeAITurn("hero"); err != nil {
		t.Fatal(err)
	}
	if got, _ := readiedType(t, far); got != 2 {
		t.Fatalf("離得遠、有箭，卻沒換成弓（拿的是類別 %d）", got)
	}
	// 沒有箭就不該換。
	noAmmo := autoEquipState(t, []monster.ItemRecord{
		{Name: "短劍", Type: 0, Readied: true, Count: 1},
		{Name: "弓", Type: 2, Count: 1},
	}, 9)
	if _, err := noAmmo.autoEquipBeforeAITurn("hero"); err != nil {
		t.Fatal(err)
	}
	if got, _ := readiedType(t, noAmmo); got == 2 {
		t.Fatal("沒有箭卻換成弓")
	}
}

// 換完之後戰鬥員的傷害骰要跟著改——不然換了等於沒換。
func TestAutoEquipReprojectsTheFightersDamage(t *testing.T) {
	state := autoEquipState(t, []monster.ItemRecord{
		{Name: "短劍", Type: 0, Readied: true, Count: 1},
		{Name: "長劍", Type: 1, Count: 1},
	}, 8)
	before, _ := state.fighter("hero")
	if _, err := state.autoEquipBeforeAITurn("hero"); err != nil {
		t.Fatal(err)
	}
	after, _ := state.fighter("hero")
	if after.DamageDiceSides == before.DamageDiceSides {
		t.Fatalf("換武器之後傷害骰還是 %d 面——衍生值沒有重算", after.DamageDiceSides)
	}
	if after.HitPoints != before.HitPoints || after.CombatX != before.CombatX {
		t.Fatal("換武器把生命或位置也一起換掉了")
	}
}
