package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 原作的「保留的機會攻擊」是戰鬥狀態 `+07h`：GUARD 設 1、相鄰者進來時消費、
// 回合開始清 0（spec 1136 把三端接起來了）。remake 這一側用的是
// `Fighter.CombatAction.Guarding`。
//
// ⚠ 行為早就有測試釘住（只打一次、定身壓制）；**沒被驗過的是它存不存得下去**。
// 戰鬥中存檔再讀回來，防禦如果掉了，玩家會少掉一次反擊而且沒有任何錯誤訊息。
func TestGuardReservationSurvivesASaveRoundTrip(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	// ⚠ 要**兩名隊員**。只有一名時，`CombatGuard()` 回來之前那名隊員的下一回合
	// 已經開始、防禦照設計被清掉了（spec 1136：防禦撐到自己下一次被選中）——
	// 那是正確行為，不是 bug，但也讓旗標觀察不到。
	abilities := party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10,
		Dexterity: 12, Constitution: 14, Charisma: 10}
	state.partyRoster = party.Roster{
		{ID: "p1", Name: "亞勇", Race: party.RaceHuman, Class: party.ClassFighter,
			Level: 3, Abilities: abilities},
		{ID: "p2", Name: "第二人", Race: party.RaceHuman, Class: party.ClassFighter,
			Level: 3, Abilities: abilities},
	}
	result := ecl.RunResult{CombatRequested: true,
		MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 7, Count: 1, IconBlock: 81}}}
	records := map[uint8]monster.Record{
		7: {Name: "FIRE KNIFE", HitPoints: 8, MaxHitPoints: 8, AttackBlows: [2]uint8{2, 0}},
	}
	partyFighters := []combat.Fighter{
		{ID: "p1", Name: "亞勇", Side: combat.SideParty,
			HitPoints: 20, MaxHitPoints: 20, InitiativeBonus: 99},
		{ID: "p2", Name: "第二人", Side: combat.SideParty,
			HitPoints: 20, MaxHitPoints: 20, InitiativeBonus: 98},
	}
	if err := state.StartEncounter(result, records, partyFighters, 11); err != nil {
		t.Fatal(err)
	}
	attacker, ok := state.combatPartyTurn()
	if !ok {
		t.Skip("這一局的先攻沒有讓隊員先動，換不到 GUARD 的時機")
	}
	guardID := attacker.ID
	if err := state.CombatGuard(); err != nil {
		t.Fatalf("GUARD：%v", err)
	}
	if !fighterGuarding(&state, guardID) {
		t.Skipf("%s 下了 GUARD 之後輪回自己，防禦已被回合開始清掉——這一局觀察不到",
			guardID)
	}

	path := filepath.Join(t.TempDir(), "guard.json")
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(locale.Catalog{Language: "zh-TW", Strings: map[string]string{"title": "t"}})
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}
	if loaded.battle == nil {
		t.Fatal("讀回來沒有進行中的戰鬥")
	}
	if !fighterGuarding(&loaded, guardID) {
		t.Error("讀回來之後防禦掉了——玩家會少掉一次反擊而且沒有任何錯誤訊息")
	}
}

func fighterGuarding(state *State, id string) bool {
	if state.battle == nil {
		return false
	}
	for _, fighter := range state.battle.Fighters() {
		if fighter.ID == id {
			return fighter.CombatAction.Guarding
		}
	}
	return false
}
