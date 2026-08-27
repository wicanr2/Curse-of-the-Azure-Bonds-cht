package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 骰子表釘住三支同族法術的差別**只有骰數與加值**：`1d8`／`2d8 ＋ 1`／`3d8 ＋ 3`。
// 這是「一支法術一段程式碼」被換成資料的直接證據，也守住加值不能漏——
// 少了加值三支的治療量全部偏低，而且低得像是合理的數字。
func TestCureFamilyDiffersOnlyInDiceAndBonus(t *testing.T) {
	table, err := gamepack.SpellDamage()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		spell uint8
		count int
		bonus int
	}{{3, 1, 0}, {58, 2, 1}, {71, 3, 3}} {
		count, sides, bonus, ok := table.Dice(item.spell)
		if !ok {
			t.Fatalf("法術 %d 在骰子表裡沒有可用的骰", item.spell)
		}
		if count != item.count || sides != 8 || bonus != item.bonus {
			t.Fatalf("法術 %d 是 %dd%d ＋ %d，want %dd8 ＋ %d",
				item.spell, count, sides, bonus, item.count, item.bonus)
		}
	}
}

// 骰數由施法者等級算出來的那幾支不在機械表裡，走逐支讀完的算式
// （spec 1124）。寒冰錐是 `等級d4 ＋ 等級`、電擊之握是 `1d8 ＋ 等級`、
// 燃燒之手根本不擲骰。
func TestCasterLevelFormulasComeFromHandlers(t *testing.T) {
	table, err := gamepack.SpellDamage()
	if err != nil {
		t.Fatal(err)
	}
	for _, spellID := range []uint8{9, 20, 92} {
		if _, _, _, ok := table.Dice(spellID); ok {
			t.Fatalf("法術 %d 的骰數含施法者等級，機械表不該給數字", spellID)
		}
	}
	for _, item := range []struct {
		spell uint8
		want  combat.SpellDamageFormula
	}{
		{9, combat.SpellDamageFormula{Bonus: 7}},
		{20, combat.SpellDamageFormula{Count: 1, Sides: 8, Bonus: 7}},
		{92, combat.SpellDamageFormula{Count: 7, Sides: 4, Bonus: 7}},
	} {
		got, ok := combat.SpellDamageRoll(item.spell, 7)
		if !ok || got != item.want {
			t.Fatalf("法術 %d 在 7 級是 %+v（ok=%v），want %+v", item.spell, got, ok, item.want)
		}
	}
}

func TestBurningHandsUsesItsOwnLocalizedName(t *testing.T) {
	value := NewState(combatVisualCatalog(t))
	state := &value
	state.partyRoster = party.Roster{{
		ID: "mage", Name: "魔法師", Class: party.ClassMagicUser, Level: 5,
		ClassLevels: [8]uint8{5: 5}, SpellSlots: []uint8{BurningHandsSpellID},
	}}
	if err := state.StartCombat(
		[]combat.Fighter{{ID: "mage", Name: "魔法師", Side: combat.SideParty,
			HitPoints: 15, MaxHitPoints: 15, InitiativeBonus: 20,
			HasCombatPosition: true, CombatX: 1, CombatY: 1},
			{ID: "ally", Name: "隊友", Side: combat.SideParty,
				HitPoints: 20, MaxHitPoints: 20, InitiativeBonus: 19,
				HasCombatPosition: true, CombatX: 0, CombatY: 1}},
		[]combat.Fighter{{ID: "guard", Name: "衛兵", Side: combat.SideEnemy,
			HitPoints: 20, MaxHitPoints: 20,
			HasCombatPosition: true, CombatX: 2, CombatY: 1}},
		11,
	); err != nil {
		t.Fatal(err)
	}
	if !state.CombatCanCastBurningHands() {
		active, _ := state.CombatActiveFighter()
		t.Fatalf("相鄰敵人前的五級法師應可施放燃燒之手：active=%+v fighters=%+v choices=%+v",
			active, state.CombatFighters(), state.CombatSpellChoices())
	}
	if err := state.BeginCombatCast(BurningHandsSpellID); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(BurningHandsSpellID); err != nil {
		t.Fatal(err)
	}
	if got := state.CombatMessage(); got != "魔法師 對 衛兵 施放燃燒之手，造成 5 點傷害。" {
		t.Fatalf("燃燒之手訊息=%q", got)
	}
}

// handler 裡擲了不只一次骰的（火球）不給數字——寧可沒有也不要錯的。
func TestAmbiguousHandlersDoNotHandOutDice(t *testing.T) {
	table, err := gamepack.SpellDamage()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, ok := table.Dice(47); ok {
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

// 治療重傷：治 2d8，範圍 2..16。
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

// 造成重傷：打 2d8。
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

// 範圍傷害只擲**一次**骰，套給半徑內的每一個人；每人各自擲豁免。
// 每個目標各擲一次傷害會讓方差與原版不同。
func TestIceStormRollsDamageOnceForTheWholeArea(t *testing.T) {
	value := NewState(testCatalog())
	state := &value
	state.partyRoster = party.Roster{{ID: "mage", Name: "法師",
		Class: party.ClassMagicUser, Level: 8, HitPoints: 20, MaxHitPoints: 20,
		ClassLevels:  [8]uint8{5: 8},
		HealthStatus: party.HealthStatusOK,
		Abilities: party.Abilities{Strength: 10, Intelligence: 17, Wisdom: 10,
			Dexterity: 12, Constitution: 12, Charisma: 10},
		SpellSlots: []uint8{87}}}
	partyFighters := []combat.Fighter{{ID: "mage", Name: "法師", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, InitiativeBonus: 20,
		HasCombatPosition: true, CombatX: 1, CombatY: 1}}
	enemies := []combat.Fighter{
		{ID: "orc-a", Name: "獸人甲", Side: combat.SideEnemy, HitPoints: 60, MaxHitPoints: 60,
			ArmorClass: 10, SavingThrows: []uint8{99, 99, 99, 99, 99},
			HasCombatPosition: true, CombatX: 8, CombatY: 8},
		{ID: "orc-b", Name: "獸人乙", Side: combat.SideEnemy, HitPoints: 60, MaxHitPoints: 60,
			ArmorClass: 10, SavingThrows: []uint8{99, 99, 99, 99, 99},
			HasCombatPosition: true, CombatX: 9, CombatY: 8},
		{ID: "orc-far", Name: "獸人丙", Side: combat.SideEnemy, HitPoints: 60, MaxHitPoints: 60,
			ArmorClass: 10, SavingThrows: []uint8{99, 99, 99, 99, 99},
			HasCombatPosition: true, CombatX: 20, CombatY: 2},
	}
	if err := state.StartCombat(partyFighters, enemies, 13); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(87); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(87); err != nil {
		t.Fatal(err)
	}
	damage := map[string]int{}
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideEnemy {
			damage[fighter.ID] = 60 - fighter.HitPoints
		}
	}
	if damage["orc-a"] == 0 || damage["orc-a"] != damage["orc-b"] {
		t.Fatalf("同一次冰風暴打出不同傷害：%v（骰只擲一次）", damage)
	}
	if damage["orc-a"] < 3 || damage["orc-a"] > 30 {
		t.Fatalf("傷害 %d 不在 3d10 的範圍 3..30", damage["orc-a"])
	}
	if damage["orc-far"] != 0 {
		t.Fatalf("半徑外的人也被打了 %d 點", damage["orc-far"])
	}
}

// 抗寒的人在冰風暴裡只吃一半——傷害屬性旗標接上了才會這樣。
func TestResistColdHalvesIceStorm(t *testing.T) {
	build := func(resist bool) *State {
		value := NewState(testCatalog())
		state := &value
		state.partyRoster = party.Roster{{ID: "mage", Name: "法師",
			Class: party.ClassMagicUser, Level: 8, HitPoints: 20, MaxHitPoints: 20,
			ClassLevels:  [8]uint8{5: 8},
			HealthStatus: party.HealthStatusOK,
			Abilities: party.Abilities{Strength: 10, Intelligence: 17, Wisdom: 10,
				Dexterity: 12, Constitution: 12, Charisma: 10},
			SpellSlots: []uint8{87}}}
		orc := combat.Fighter{ID: "orc", Name: "獸人", Side: combat.SideEnemy,
			HitPoints: 60, MaxHitPoints: 60, ArmorClass: 10,
			SavingThrows:      []uint8{99, 99, 99, 99, 99},
			HasCombatPosition: true, CombatX: 8, CombatY: 8}
		if resist {
			orc.MonsterAffects = []combat.MonsterAffect{{Kind: 0x0A, Active: true}}
		}
		partyFighters := []combat.Fighter{{ID: "mage", Name: "法師", Side: combat.SideParty,
			HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, InitiativeBonus: 20,
			HasCombatPosition: true, CombatX: 1, CombatY: 1}}
		if err := state.StartCombat(partyFighters, []combat.Fighter{orc}, 13); err != nil {
			t.Fatal(err)
		}
		if err := state.BeginCombatCast(87); err != nil {
			t.Fatal(err)
		}
		if err := state.CombatCast(87); err != nil {
			t.Fatal(err)
		}
		return state
	}
	damageOf := func(state *State) int {
		for _, fighter := range state.CombatFighters() {
			if fighter.ID == "orc" {
				return 60 - fighter.HitPoints
			}
		}
		return -1
	}
	plain, resisted := damageOf(build(false)), damageOf(build(true))
	if resisted != plain/2 {
		t.Fatalf("抗寒吃了 %d 點、沒抗寒 %d 點，應該剛好一半", resisted, plain)
	}
}
