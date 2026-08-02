package combat

import (
	"reflect"
	"testing"

	enginerandom "github.com/wicanr2/golden-box-remake-engine/randomstream"
)

func TestBattleSnapshotRestoresEffectsSchedulerAndRandomContinuation(t *testing.T) {
	fighters := []Fighter{
		{ID: "mage", Side: SideParty, LegacyObjectID: 1, HitPoints: 20, MaxHitPoints: 20,
			ArmorClass: 0, AttackBonus: 30, DamageDiceCount: 1, DamageDiceSides: 4,
			Dexterity: 16, InitiativeBonus: 12, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "orc", Side: SideEnemy, LegacyObjectID: 2, HitPoints: 30, MaxHitPoints: 30,
			HitDice: 1, ArmorClass: 8, AttackBonus: 10, DamageDiceCount: 1, DamageDiceSides: 6,
			Dexterity: 9, InitiativeBonus: 8, HasCombatPosition: true, CombatX: 4, CombatY: 1,
			CombatAction: ActionState{Delay: 8, SpellID: 0x22, TargetID: "mage", Guarding: true}},
	}
	battle, err := NewBattle(fighters, 443)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.SetSideAttackRollModifier(SideParty, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := battle.CastSleepOrdered("mage", []string{"orc"}, 3); err != nil {
		t.Fatal(err)
	}
	if err := battle.BeginScheduledRound(); err != nil {
		t.Fatal(err)
	}
	selected, ok, err := battle.NextScheduledTurn()
	if err != nil || !ok {
		t.Fatalf("select turn=%+v ok=%v err=%v", selected, ok, err)
	}
	snapshot, err := battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreBattle(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restoredSnapshot, err := restored.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restoredSnapshot, snapshot) {
		t.Fatalf("restored snapshot differs\n got=%#v\nwant=%#v", restoredSnapshot, snapshot)
	}
	if err := battle.CompleteAction(selected.FighterID); err != nil {
		t.Fatal(err)
	}
	if err := restored.CompleteAction(selected.FighterID); err != nil {
		t.Fatal(err)
	}
	nextA, okA, errA := battle.NextScheduledTurn()
	nextB, okB, errB := restored.NextScheduledTurn()
	if errA != nil || errB != nil || okA != okB || nextA != nextB {
		t.Fatalf("next continuation original=(%+v,%v,%v) restored=(%+v,%v,%v)", nextA, okA, errA, nextB, okB, errB)
	}
}

func TestRestoreBattleRejectsMalformedSchedulerSelection(t *testing.T) {
	snapshot := BattleSnapshot{
		Version:  BattleSnapshotVersion,
		Fighters: []Fighter{{ID: "hero", Side: SideParty, HitPoints: 1, MaxHitPoints: 1, ArmorClass: 10}},
		Random:   enginerandom.Snapshot{Seed: 1}, Status: StatusActive,
		SchedulerSelected: true,
	}
	if _, err := RestoreBattle(snapshot); err == nil {
		t.Fatal("selected action without scheduler unexpectedly restored")
	}
}

func TestRestoreBattleRejectsUnboundedRandomReplay(t *testing.T) {
	snapshot := BattleSnapshot{
		Version:  BattleSnapshotVersion,
		Fighters: []Fighter{{ID: "hero", Side: SideParty, HitPoints: 1, MaxHitPoints: 1, ArmorClass: 10}},
		Random:   enginerandom.Snapshot{Seed: 1, Draws: enginerandom.MaxReplayDraws + 1},
		Status:   StatusActive,
	}
	if _, err := RestoreBattle(snapshot); err == nil {
		t.Fatal("unbounded random replay unexpectedly restored")
	}
}
