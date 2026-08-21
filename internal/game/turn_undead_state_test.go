package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// turnUndeadState 起一場「牧師對上一隻最弱的不死生物」。先攻靠 InitiativeBonus
// 壓成牧師先動，這樣才觀察得到選單與訊息。
func turnUndeadState(t *testing.T, clericLevel int, undeadType uint8) *State {
	t.Helper()
	state := NewState(trainingTestCatalog(t))
	abilities := party.Abilities{Strength: 12, Intelligence: 10, Wisdom: 16,
		Dexterity: 12, Constitution: 14, Charisma: 10}
	state.partyRoster = party.Roster{
		{ID: "p1", Name: "祭司", Race: party.RaceHuman, Class: party.ClassCleric,
			Level: clericLevel, Abilities: abilities},
	}
	raw := make([]byte, monster.RecordSize)
	raw[0x0E9] = undeadType
	result := ecl.RunResult{CombatRequested: true,
		MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 7, Count: 1, IconBlock: 81}}}
	records := map[uint8]monster.Record{
		7: {Name: "DRACOLICH", HitPoints: 8, MaxHitPoints: 8, AttacksPerTurn: 1,
			UndeadType: undeadType, Raw: raw},
	}
	cleric, err := state.partyRoster[0].Fighter()
	if err != nil {
		t.Fatal(err)
	}
	cleric.HitPoints, cleric.MaxHitPoints, cleric.InitiativeBonus = 20, 20, 99
	if err := state.StartEncounter(result, records, []combat.Fighter{cleric}, 11); err != nil {
		t.Fatal(err)
	}
	return &state
}

// 選單是動態組合的：牧師看得到「退散」，用過就不見（spec 905）。
func TestCombatMenuShowsTurnUndeadOnlyForAClericWhoHasNotUsedIt(t *testing.T) {
	state := turnUndeadState(t, 5, 1)
	if _, ok := state.combatPartyTurn(); !ok {
		t.Skip("這一局的先攻沒有讓牧師先動")
	}
	if !state.CombatCanTurnUndead() {
		t.Fatal("牧師應該看得到「退散」")
	}
	menuWith := state.CombatMainMenuText()
	if !strings.Contains(menuWith, "退散") {
		t.Fatalf("選單裡沒有「退散」：%q", menuWith)
	}
	if err := state.CombatTurnUndead(); err != nil {
		t.Fatalf("驅散：%v", err)
	}
	if state.CombatCanTurnUndead() {
		t.Fatal("用過之後不該再看得到")
	}
	if menuAfter := state.CombatMainMenuText(); strings.Contains(menuAfter, "退散") {
		t.Fatalf("用過之後選單還留著「退散」：%q", menuAfter)
	}
	if menuWith == state.CombatMainMenuText() {
		t.Fatal("兩種排列應該是不同的文字規則")
	}
}

// 不是牧師就完全看不到，連叫都不該成功。
func TestCombatTurnUndeadIsClericOnly(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	abilities := party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10,
		Dexterity: 12, Constitution: 14, Charisma: 10}
	state.partyRoster = party.Roster{
		{ID: "p1", Name: "戰士", Race: party.RaceHuman, Class: party.ClassFighter,
			Level: 5, Abilities: abilities},
	}
	result := ecl.RunResult{CombatRequested: true,
		MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 7, Count: 1, IconBlock: 81}}}
	records := map[uint8]monster.Record{
		7: {Name: "DRACOLICH", HitPoints: 8, MaxHitPoints: 8, AttacksPerTurn: 1, UndeadType: 1},
	}
	fighter, err := state.partyRoster[0].Fighter()
	if err != nil {
		t.Fatal(err)
	}
	fighter.HitPoints, fighter.MaxHitPoints, fighter.InitiativeBonus = 20, 20, 99
	if err := state.StartEncounter(result, records, []combat.Fighter{fighter}, 11); err != nil {
		t.Fatal(err)
	}
	if _, ok := state.combatPartyTurn(); !ok {
		t.Skip("這一局的先攻沒有讓戰士先動")
	}
	if state.CombatCanTurnUndead() {
		t.Fatal("戰士不該看得到「退散」")
	}
	if strings.Contains(state.CombatMainMenuText(), "退散") {
		t.Fatalf("戰士的選單不該有「退散」：%q", state.CombatMainMenuText())
	}
	if err := state.CombatTurnUndead(); err == nil {
		t.Fatal("戰士驅散應該報錯")
	}
}

// 場上沒有不死生物時，原作照樣讓你按下去——條件裡沒有「要有目標」，
// 而且那一次行動照樣用掉。
func TestCombatTurnUndeadWithNoUndeadStillConsumesTheAction(t *testing.T) {
	state := turnUndeadState(t, 5, 0)
	if _, ok := state.combatPartyTurn(); !ok {
		t.Skip("這一局的先攻沒有讓牧師先動")
	}
	if !state.CombatCanTurnUndead() {
		t.Fatal("條件裡沒有「場上要有不死生物」")
	}
	if err := state.CombatTurnUndead(); err != nil {
		t.Fatalf("驅散：%v", err)
	}
	if state.CombatCanTurnUndead() {
		t.Fatal("沒有目標也算用掉了")
	}
}

// 訊息由四句組成（spec 834）。⚠ 不能在 `CombatTurnUndead()` 之後讀
// `CombatMessage()`——那一支結束之後輪到敵方，訊息已經被下一件事蓋掉了。
func TestTurnUndeadMessageComposesTheOriginalFourLines(t *testing.T) {
	state := turnUndeadState(t, 5, 1)
	cleric := combat.Fighter{ID: "p1", Name: "祭司"}
	target, ok := state.fighter(state.CombatFighters()[len(state.CombatFighters())-1].ID)
	if !ok {
		t.Fatal("找不到目標")
	}

	nothing := state.turnUndeadMessage(cleric, combat.TurnUndeadResult{ClericID: "p1"})
	if !strings.Contains(nothing, requireCatalogText(t, state, "combat_turn_undead_nothing")) {
		t.Fatalf("一個都沒成時要印「什麼都沒發生」：%q", nothing)
	}

	turned := state.turnUndeadMessage(cleric, combat.TurnUndeadResult{
		ClericID: "p1",
		Outcomes: []combat.TurnUndeadOutcome{{TargetID: target.ID, UndeadType: 1, Threshold: 4}},
	})
	if strings.Contains(turned, requireCatalogText(t, state, "combat_turn_undead_nothing")) {
		t.Fatalf("有成的時候不該印「什麼都沒發生」：%q", turned)
	}
	if !strings.Contains(turned, target.Name) {
		t.Fatalf("訊息裡沒有目標的名字：%q", turned)
	}

	destroyed := state.turnUndeadMessage(cleric, combat.TurnUndeadResult{
		ClericID: "p1",
		Outcomes: []combat.TurnUndeadOutcome{{TargetID: target.ID, UndeadType: 1, Destroyed: true}},
	})
	if destroyed == turned {
		t.Fatalf("驅離與摧毀應該是不同的句子：%q", destroyed)
	}
}

func requireCatalogText(t *testing.T, state *State, key string) string {
	t.Helper()
	text := state.catalog.Text(key, key)
	if text == key {
		t.Fatalf("locale 缺 %q", key)
	}
	return text
}
