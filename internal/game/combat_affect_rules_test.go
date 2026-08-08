package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

func TestCombatAffectRulesLoadFromPackAndReattachAfterRestore(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{
		ID: "hero", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20,
		ArmorClass: 10, Alignment: combat.AlignmentNeutralGood, AlignmentKnown: true,
	}}
	enemies := []combat.Fighter{{
		ID: "cold-resistant", Side: combat.SideEnemy, HitPoints: 20, MaxHitPoints: 20,
		ArmorClass: 10, MonsterAffects: []combat.MonsterAffect{{Kind: 0x0A, Innate: true}, {Kind: 0x09, Innate: true}},
	}}
	if err := state.StartCombat(party, enemies, 498); err != nil {
		t.Fatal(err)
	}
	fighter, ok := state.battle.Fighter("cold-resistant")
	if !ok || len(fighter.DamageRules) != 3 {
		t.Fatalf("start combat damage rules=%+v found=%v", fighter.DamageRules, ok)
	}
	if len(fighter.ConditionalModifierRules) != 2 {
		t.Fatalf("start combat conditional rules=%+v", fighter.ConditionalModifierRules)
	}
	if len(fighter.MagicResistanceRules) != 1 || fighter.MagicResistanceRules[0].Base != 15 {
		t.Fatalf("start combat magic resistance rules=%+v", fighter.MagicResistanceRules)
	}
	if result := fighter.MonsterConditionalModifierAgainst(party[0]); result.AttackRollDelta != -2 || result.SavingThrowDelta != 2 {
		t.Fatalf("start combat conditional result=%+v", result)
	}
	if result := fighter.MonsterDamageAdjustment(combat.DamageFlagCold, 17); result.Damage != 8 || result.Immune {
		t.Fatalf("start combat cold adjustment=%+v", result)
	}

	snapshot, err := state.activeCombatSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	loaded := NewState(testCatalog())
	if err := loaded.restoreActiveCombat(*snapshot); err != nil {
		t.Fatal(err)
	}
	restored, ok := loaded.battle.Fighter("cold-resistant")
	if !ok || len(restored.DamageRules) != 3 {
		t.Fatalf("restored damage rules=%+v found=%v", restored.DamageRules, ok)
	}
	if len(restored.ConditionalModifierRules) != 2 {
		t.Fatalf("restored conditional rules=%+v", restored.ConditionalModifierRules)
	}
	if len(restored.MagicResistanceRules) != 1 || restored.MagicResistanceRules[0].Base != 15 {
		t.Fatalf("restored magic resistance rules=%+v", restored.MagicResistanceRules)
	}
	if result := restored.MonsterConditionalModifierAgainst(party[0]); result.AttackRollDelta != -2 || result.SavingThrowDelta != 2 {
		t.Fatalf("restored conditional result=%+v", result)
	}
	if result := restored.MonsterDamageAdjustment(combat.DamageFlagCold, 17); result.Damage != 8 || result.Immune {
		t.Fatalf("restored cold adjustment=%+v", result)
	}
}
