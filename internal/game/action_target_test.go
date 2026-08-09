package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

func TestCombatQuickUsesPackPolicyToClearSameTeamActionTarget(t *testing.T) {
	state := NewState(testCatalog())
	partyFighters := []combat.Fighter{
		{ID: "pc", Side: combat.SideParty, ControlMorale: 0, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, InitiativeBonus: 20},
		{ID: "ally", Side: combat.SideParty, ControlMorale: 0, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 10},
	}
	if err := state.StartCombat(partyFighters, []combat.Fighter{{ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 1}}, 509); err != nil {
		t.Fatal(err)
	}
	if err := state.battle.SetActionTarget("pc", "ally"); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatQuick(); err != nil {
		t.Fatal(err)
	}
	pc, ok := state.fighter("pc")
	if !ok || !pc.QuickFight || pc.CombatAction.ActionTargetID != "" {
		t.Fatalf("pack quick policy did not clear same-team action target: %+v", pc)
	}
}
