package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 骰子表釘住三支同族法術的差別**只有骰數**：`1d8`／`2d8`／`3d8`。
// 這是「一支法術一段程式碼」被換成資料的直接證據。
func TestCureFamilyDiffersOnlyInDiceCount(t *testing.T) {
	table, err := gamepack.SpellDamage()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		spell uint8
		count int
	}{{3, 1}, {58, 2}, {71, 3}} {
		count, sides, ok := table.Dice(item.spell, 5)
		if !ok {
			t.Fatalf("法術 %d 在骰子表裡沒有可用的骰", item.spell)
		}
		if count != item.count || sides != 8 {
			t.Fatalf("法術 %d 是 %dd%d，want %dd8", item.spell, count, sides, item.count)
		}
	}
}

// 骰數 0 代表用施法者等級（魔法飛彈）。
func TestZeroDiceCountMeansCasterLevel(t *testing.T) {
	table, err := gamepack.SpellDamage()
	if err != nil {
		t.Fatal(err)
	}
	count, sides, ok := table.Dice(15, 7)
	if !ok {
		t.Fatal("魔法飛彈在骰子表裡沒有可用的骰")
	}
	if count != 7 || sides != 4 {
		t.Fatalf("7 級的魔法飛彈是 %dd%d，want 7d4", count, sides)
	}
}

// handler 裡擲了不只一次骰的（火球）不給數字——寧可沒有也不要錯的。
func TestAmbiguousHandlersDoNotHandOutDice(t *testing.T) {
	table, err := gamepack.SpellDamage()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, ok := table.Dice(47, 5); ok {
		t.Fatal("火球的 handler 擲了不只一次骰，不該給數字")
	}
}

func diceSpellState(t *testing.T, spellID uint8, hurt bool) *State {
	t.Helper()
	value := NewState(testCatalog())
	state := &value
	hp := 30
	if hurt {
		hp = 5
	}
	state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師",
		Class: party.ClassCleric, Level: 5, HitPoints: hp, MaxHitPoints: 30,
		ClassLevels:  [8]uint8{0: 5},
		HealthStatus: party.HealthStatusOK,
		Abilities: party.Abilities{Strength: 12, Intelligence: 10, Wisdom: 16,
			Dexterity: 12, Constitution: 14, Charisma: 10},
		SpellSlots: []uint8{spellID}}}
	partyFighters := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty,
		HitPoints: hp, MaxHitPoints: 30, ArmorClass: 10, InitiativeBonus: 20,
		HasCombatPosition: true, CombatX: 1, CombatY: 1}}
	enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy,
		HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 2, CombatY: 1}}
	if err := state.StartCombat(partyFighters, enemies, 11); err != nil {
		t.Fatal(err)
	}
	return state
}

// 中度治療術：治 2d8，範圍 2..16。
func TestCureSeriousWoundsHealsTwoD8(t *testing.T) {
	state := diceSpellState(t, 58, true)
	if err := state.BeginCombatCast(58); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(58); err != nil {
		t.Fatal(err)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("法術位沒被消耗：%#v", state.partyRoster[0].SpellSlots)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID != "cleric" {
			continue
		}
		healed := fighter.HitPoints - 5
		if healed < 2 || healed > 16 {
			t.Fatalf("治了 %d 點，2d8 的範圍是 2..16", healed)
		}
	}
}

// 中度致傷術：打 2d8。
func TestCauseSeriousWoundsDealsTwoD8(t *testing.T) {
	state := diceSpellState(t, 66, false)
	if err := state.BeginCombatCast(66); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(66); err != nil {
		t.Fatal(err)
	}
	for _, fighter := range state.CombatFighters() {
		if fighter.ID != "orc" {
			continue
		}
		damage := 40 - fighter.HitPoints
		if damage < 2 || damage > 16 {
			t.Fatalf("打了 %d 點，2d8 的範圍是 2..16", damage)
		}
	}
}

// 沒有人受傷就不能施治療。
func TestHealDiceIsUnavailableWithAFullHealthParty(t *testing.T) {
	state := diceSpellState(t, 58, false)
	if state.combatCanCastSpellDice(58, true) {
		t.Fatal("全隊滿血卻說治療術可以施")
	}
	if err := state.BeginCombatCast(58); err == nil {
		t.Fatal("全隊滿血卻開始施法了")
	}
}

// 火焰打擊有「豁免過了減半」（`+8h = 2`）。折半要在套上去之前，
// 否則「打死了又活過來」會出現，而且回報的數字對不上實際掉的血。
func TestFlameStrikeHalvesOnASuccessfulSave(t *testing.T) {
	build := func(threshold uint8, seed int64) *State {
		value := NewState(testCatalog())
		state := &value
		state.partyRoster = party.Roster{{ID: "cleric", Name: "牧師",
			Class: party.ClassCleric, Level: 9, HitPoints: 30, MaxHitPoints: 30,
			ClassLevels:  [8]uint8{0: 9},
			HealthStatus: party.HealthStatusOK,
			Abilities: party.Abilities{Strength: 12, Intelligence: 10, Wisdom: 16,
				Dexterity: 12, Constitution: 14, Charisma: 10},
			SpellSlots: []uint8{74}}}
		partyFighters := []combat.Fighter{{ID: "cleric", Name: "牧師", Side: combat.SideParty,
			HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10, InitiativeBonus: 20,
			HasCombatPosition: true, CombatX: 1, CombatY: 1}}
		enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy,
			HitPoints: 200, MaxHitPoints: 200, ArmorClass: 10,
			SavingThrows:      []uint8{threshold, threshold, threshold, threshold, threshold},
			HasCombatPosition: true, CombatX: 2, CombatY: 1}}
		if err := state.StartCombat(partyFighters, enemies, seed); err != nil {
			t.Fatal(err)
		}
		return state
	}
	damageOf := func(state *State) int {
		for _, fighter := range state.CombatFighters() {
			if fighter.ID == "orc" {
				return 200 - fighter.HitPoints
			}
		}
		return -1
	}
	// 門檻 1 ＝ 幾乎一定過；門檻 99 ＝ 幾乎一定不過（天然 20 除外）。
	alwaysSaves, neverSaves := 0, 0
	for seed := int64(1); seed <= 30; seed++ {
		saved := build(1, seed)
		if err := saved.BeginCombatCast(74); err != nil {
			t.Fatal(err)
		}
		if err := saved.CombatCast(74); err != nil {
			t.Fatal(err)
		}
		alwaysSaves += damageOf(saved)

		failed := build(99, seed)
		if err := failed.BeginCombatCast(74); err != nil {
			t.Fatal(err)
		}
		if err := failed.CombatCast(74); err != nil {
			t.Fatal(err)
		}
		neverSaves += damageOf(failed)
	}
	if alwaysSaves >= neverSaves {
		t.Fatalf("過豁免共 %d 點、沒過共 %d 點——減半沒有生效", alwaysSaves, neverSaves)
	}
	// 6d8 的範圍是 6..48，折半後 3..24。30 次的總和不該超過上界太多。
	if neverSaves > 30*48 || alwaysSaves < 30*3 {
		t.Fatalf("傷害總和超出 6d8 的合理範圍：過 %d／沒過 %d", alwaysSaves, neverSaves)
	}
}
