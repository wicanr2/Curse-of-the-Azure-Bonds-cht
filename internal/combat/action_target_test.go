package combat

import (
	"encoding/json"
	"testing"
)

func TestQuickFightClearsOnlySameTeamActionTarget(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "pc", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "ally", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "enemy", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
	}, 509)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.SetActionTarget("pc", "ally"); err != nil {
		t.Fatal(err)
	}
	if err := battle.SetQuickFightWithPolicy("pc", true); err != nil {
		t.Fatal(err)
	}
	pc, _ := battle.Fighter("pc")
	if pc.CombatAction.ActionTargetID != "" || !pc.QuickFight {
		t.Fatalf("same-team target was not cleared: %+v", pc)
	}

	if err := battle.SetActionTarget("pc", "enemy"); err != nil {
		t.Fatal(err)
	}
	if err := battle.SetQuickFightWithPolicy("pc", true); err != nil {
		t.Fatal(err)
	}
	pc, _ = battle.Fighter("pc")
	if pc.CombatAction.ActionTargetID != "enemy" {
		t.Fatalf("opposing-team target was incorrectly cleared: %+v", pc.CombatAction)
	}
}

func TestQuickFightPolicyCanPreserveSameTeamActionTarget(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "pc", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "ally", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "enemy", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
	}, 509)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.SetActionTarget("pc", "ally"); err != nil {
		t.Fatal(err)
	}
	if err := battle.SetQuickFightWithPolicy("pc", false); err != nil {
		t.Fatal(err)
	}
	pc, _ := battle.Fighter("pc")
	if pc.CombatAction.ActionTargetID != "ally" {
		t.Fatalf("disabled policy changed same-team target: %+v", pc.CombatAction)
	}
}

func TestActionTargetSurvivesBattleSnapshotRoundTrip(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "pc", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "ally", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "enemy", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
	}, 509)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.SetActionTarget("pc", "ally"); err != nil {
		t.Fatal(err)
	}
	snapshot, err := battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded BattleSnapshot
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreBattle(decoded)
	if err != nil {
		t.Fatal(err)
	}
	pc, _ := restored.Fighter("pc")
	if pc.CombatAction.ActionTargetID != "ally" {
		t.Fatalf("action target was not preserved by snapshot: %+v", pc.CombatAction)
	}
}
