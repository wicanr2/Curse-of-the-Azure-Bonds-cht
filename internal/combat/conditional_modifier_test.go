package combat

import (
	"testing"

	enginemodifier "github.com/wicanr2/golden-box-remake-engine/combat/modifier"
)

func testConditionalRules() []enginemodifier.Rule {
	return []enginemodifier.Rule{
		{ID: "effect-08-evil", EffectKind: 0x08, Predicate: enginemodifier.PredicateValueIn, Values: []uint8{2, 5, 8}, AttackRollDelta: -2, SavingThrowDelta: 2},
		{ID: "effect-09-good", EffectKind: 0x09, Predicate: enginemodifier.PredicateValueIn, Values: []uint8{0, 3, 6}, AttackRollDelta: -2, SavingThrowDelta: 2},
	}
}

func TestMonsterConditionalModifierUsesInteractionAlignment(t *testing.T) {
	target := Fighter{
		ID: "target", Side: SideEnemy,
		MonsterAffects:           []MonsterAffect{{Kind: 0x09, Innate: true}},
		ConditionalModifierRules: testConditionalRules(),
	}
	good := Fighter{ID: "good", Side: SideParty, Alignment: AlignmentNeutralGood, AlignmentKnown: true}
	evil := Fighter{ID: "evil", Side: SideParty, Alignment: AlignmentLawfulEvil, AlignmentKnown: true}
	if result := target.MonsterConditionalModifierAgainst(good); result.AttackRollDelta != -2 || result.SavingThrowDelta != 2 {
		t.Fatalf("good interaction result=%+v", result)
	}
	if result := target.MonsterConditionalModifierAgainst(evil); result.AttackRollDelta != 0 || result.SavingThrowDelta != 0 {
		t.Fatalf("evil interaction result=%+v", result)
	}
	unknown := Fighter{ID: "unknown"}
	if result := target.MonsterConditionalModifierAgainst(unknown); result.AttackRollDelta != 0 || result.SavingThrowDelta != 0 {
		t.Fatalf("unknown interaction result=%+v", result)
	}
}

func TestConditionalAttackModifierChangesAttackRollOnlyAtInteractionBoundary(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "good", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10,
			AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1,
			Alignment: AlignmentNeutralGood, AlignmentKnown: true},
		{ID: "protected", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10,
			MonsterAffects: []MonsterAffect{{Kind: 0x09, Innate: true}}},
	}, 499)
	if err != nil {
		t.Fatal(err)
	}
	battle.SetConditionalModifierRules(testConditionalRules())
	result, err := battle.ResolveAttack("good", "protected", 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 8 || result.Hit {
		t.Fatalf("conditional attack result=%+v", result)
	}
	protected, ok := battle.Fighter("protected")
	if !ok || protected.HitPoints != 20 {
		t.Fatalf("protected fighter=%+v found=%v", protected, ok)
	}
}
