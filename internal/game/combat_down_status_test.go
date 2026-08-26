package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// ★ 原作只有一支共用的 SAVEDAMAGE（PC-98 `overlay-24:2658h`）在決定倒下的
// 形式（spec 1205）：溢出 >9 → 死亡；1..9 → 瀕死並記出血；剛好歸零 → 昏迷。
// ECL 傷害路徑早已照抄（`ApplyDamageWithHealthStatus`），這裡釘住戰鬥路徑
// 那一半：戰鬥擊倒也要落進同一個階梯，而不是「HP 0、狀態還是 OK」。
func TestCombatDownsProjectSaveDamageLadder(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{
		{ID: "a", Name: "甲", HitPoints: 5, MaxHitPoints: 5},
		{ID: "b", Name: "乙", HitPoints: 5, MaxHitPoints: 5},
		{ID: "c", Name: "丙", HitPoints: 5, MaxHitPoints: 5},
	}
	member := func(id string, x int) combat.Fighter {
		return combat.Fighter{ID: id, Name: id, Side: combat.SideParty,
			HitPoints: 5, MaxHitPoints: 5, ArmorClass: 10,
			HasCombatPosition: true, CombatX: x, CombatY: 1}
	}
	if err := state.StartCombat(
		[]combat.Fighter{member("a", 1), member("b", 2), member("c", 3)},
		[]combat.Fighter{{ID: "g", Name: "哥布林", Side: combat.SideEnemy,
			HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10, AttackBonus: 30,
			DamageDiceCount: 1, DamageDiceSides: 20,
			HasCombatPosition: true, CombatX: 2, CombatY: 2}}, 7); err != nil {
		t.Fatal(err)
	}
	for _, blow := range []struct {
		id   string
		roll int
	}{
		{"a", 15}, // 溢出 10 → 死亡
		{"b", 10}, // 溢出 5 → 瀕死、出血 5
		{"c", 5},  // 剛好歸零 → 昏迷
	} {
		result, err := state.battle.ResolveAttack("g", blow.id, 19, blow.roll)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Hit || result.Damage != 5 {
			t.Fatalf("%s: hit=%v damage=%d", blow.id, result.Hit, result.Damage)
		}
	}
	state.syncCombatDownStatuses()
	for index, want := range []struct {
		status party.HealthStatus
		bleed  int
	}{
		{party.HealthStatusDead, 0},
		{party.HealthStatusDying, 5},
		{party.HealthStatusUnconscious, 0},
	} {
		got := state.partyRoster[index]
		if got.HealthStatus != want.status || got.Bleeding != want.bleed || got.HitPoints != 0 {
			t.Fatalf("%s: status=%v bleeding=%d hp=%d, want %v/%d/0",
				got.ID, got.HealthStatus, got.Bleeding, got.HitPoints, want.status, want.bleed)
		}
	}
	// POSTCOM 的 PARTYDEAD：乙還在瀕死 ⇒ 不算全滅（spec 1204）。
	if state.postCombatPartyDead() {
		t.Fatal("瀕死成員在場，POSTCOM 不算全滅")
	}
	// 瀕死的乙現在包紮得到（原作包紮的對象就是 SAVEDAMAGE 標的瀕死者）。
	if !state.CombatCanBandage() {
		t.Fatal("戰鬥擊倒的瀕死者要能包紮")
	}
	// 投影是冪等的：再跑一次不能把死亡改判。
	state.syncCombatDownStatuses()
	if state.partyRoster[0].HealthStatus != party.HealthStatusDead {
		t.Fatal("重跑投影不可改判")
	}
}
