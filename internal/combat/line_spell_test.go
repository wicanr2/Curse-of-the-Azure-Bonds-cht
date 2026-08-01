package combat

import "testing"

func TestCastReflectingLineSpellContinuesThroughTargetsAndReflects(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		lineSpellFighter("caster", SideParty, 1, 1, 100, 0),
		lineSpellFighter("target", SideEnemy, 2, 1, 100, 0),
		lineSpellFighter("far", SideEnemy, 4, 1, 100, 0),
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	terrain := func(x, y int) LineCell {
		return LineCell{Valid: x >= 0 && x < 10 && y >= 0 && y < 4, Reflect: x == 5}
	}
	result, err := battle.CastReflectingLineSpell(
		"caster", 0x33, TilePoint{X: 2, Y: 1}, 3,
		ReflectingLineOptions{WeightedBudget: 14}, terrain,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.TravelImpacts != 1 || len(result.Impacts) != 5 {
		t.Fatalf("travel impacts=%d impacts=%+v", result.TravelImpacts, result.Impacts)
	}
	wantTargets := []string{"target", "far", "far", "target", "caster"}
	for index, want := range wantTargets {
		if result.Impacts[index].TargetID != want {
			t.Fatalf("impact[%d]=%+v want target %q", index, result.Impacts[index], want)
		}
	}
	if len(result.Segments) != 5 {
		t.Fatalf("segments=%+v", result.Segments)
	}
	if !result.Segments[1].Reflected || result.Segments[1].To != (TilePoint{X: 5, Y: 1}) {
		t.Fatalf("reflection segment=%+v", result.Segments[1])
	}
}

func TestCastReflectingLineSpellHitsLargeFootprintOnceUntilReentry(t *testing.T) {
	large := lineSpellFighter("large", SideEnemy, 3, 1, 100, 3)
	battle, err := NewBattle([]Fighter{
		lineSpellFighter("caster", SideParty, 1, 1, 100, 0),
		large,
	}, 11)
	if err != nil {
		t.Fatal(err)
	}
	terrain := func(x, y int) LineCell {
		return LineCell{Valid: x >= 0 && x <= 8 && y >= 0 && y < 4}
	}
	result, err := battle.CastReflectingLineSpell(
		"caster", 0x33, TilePoint{X: 2, Y: 1}, 2,
		ReflectingLineOptions{WeightedBudget: 14}, terrain,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Impacts) != 1 || result.Impacts[0].TargetID != "large" {
		t.Fatalf("impacts=%+v", result.Impacts)
	}
}

func TestCastReflectingLineSpellHonorsOperationalEffect87ElectricProtection(t *testing.T) {
	protected := lineSpellFighter("protected", SideEnemy, 2, 1, 40, 0)
	protected.MonsterAffects = []MonsterAffect{{Kind: 0x87, Innate: true}}
	battle, err := NewBattle([]Fighter{
		lineSpellFighter("caster", SideParty, 1, 1, 40, 0),
		protected,
	}, 29)
	if err != nil {
		t.Fatal(err)
	}
	terrain := func(x, y int) LineCell {
		return LineCell{Valid: x >= 0 && x <= 5 && y == 1}
	}
	result, err := battle.CastReflectingLineSpell(
		"caster", 0x33, TilePoint{X: 2, Y: 1}, 3,
		ReflectingLineOptions{
			WeightedBudget: 8,
			DamageFlags:    DamageFlagElectricity | DamageFlagMagic,
		}, terrain,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Impacts) != 1 || !result.Impacts[0].Protected ||
		result.Impacts[0].Damage != 0 || result.Impacts[0].TargetHP != 40 {
		t.Fatalf("electric protection impacts=%+v", result.Impacts)
	}
}

func TestCastReflectingLineSpellAppliesCloseFirstReflectionPenalty(t *testing.T) {
	newBattle := func() *Battle {
		battle, err := NewBattle([]Fighter{
			lineSpellFighter("caster", SideParty, 1, 1, 100, 0),
			lineSpellFighter("target", SideEnemy, 2, 1, 100, 0),
		}, 13)
		if err != nil {
			t.Fatal(err)
		}
		return battle
	}
	terrain := func(x, y int) LineCell {
		return LineCell{Valid: x >= 0 && x <= 6 && y == 1, Reflect: x == 4}
	}
	withoutPenalty, err := newBattle().CastReflectingLineSpell(
		"caster", 0x33, TilePoint{X: 2, Y: 1}, 1,
		ReflectingLineOptions{WeightedBudget: 14}, terrain,
	)
	if err != nil {
		t.Fatal(err)
	}
	withPenalty, err := newBattle().CastReflectingLineSpell(
		"caster", 0x33, TilePoint{X: 2, Y: 1}, 1,
		ReflectingLineOptions{
			WeightedBudget: 14, FirstReflectionOriginThreshold: 8, FirstReflectionPenalty: 8,
		}, terrain,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(withoutPenalty.Impacts) <= len(withPenalty.Impacts) {
		t.Fatalf("without penalty impacts=%d, with penalty=%d", len(withoutPenalty.Impacts), len(withPenalty.Impacts))
	}
	if len(withPenalty.Impacts) != 1 {
		t.Fatalf("penalized impacts=%+v", withPenalty.Impacts)
	}
}

func TestCastReflectingLineSpellUsesSharedDamageAndIndependentSaves(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		lineSpellFighter("caster", SideParty, 1, 1, 100, 0),
		lineSpellFighter("one", SideEnemy, 2, 1, 100, 0),
		lineSpellFighter("two", SideEnemy, 4, 1, 100, 0),
	}, 19)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastReflectingLineSpell(
		"caster", 0x33, TilePoint{X: 2, Y: 1}, 4,
		ReflectingLineOptions{WeightedBudget: 14},
		func(x, y int) LineCell { return LineCell{Valid: x >= 0 && x < 10 && y == 1} },
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Impacts) != 2 {
		t.Fatalf("impacts=%+v", result.Impacts)
	}
	for _, impact := range result.Impacts {
		want := result.BaseDamage
		if impact.Saved {
			want /= 2
		}
		if impact.Damage != want {
			t.Fatalf("impact=%+v base damage=%d", impact, result.BaseDamage)
		}
	}
}

func lineSpellFighter(id string, side Side, x, y, hp int, size uint8) Fighter {
	return Fighter{
		ID: id, Name: id, Side: side, HitPoints: hp, MaxHitPoints: hp,
		ArmorClass: 0, DamageDiceCount: 1, DamageDiceSides: 4,
		HasCombatPosition: true, CombatX: x, CombatY: y, CombatSize: size,
		SavingThrows: []uint8{30, 30, 30, 30, 30},
	}
}
