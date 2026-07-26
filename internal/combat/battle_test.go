package combat

import "testing"

func testBattle(t *testing.T) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 5, AttackBonus: 3, DamageDiceCount: 1, DamageDiceSides: 8},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 6, MaxHitPoints: 6, ArmorClass: 10, AttackBonus: 1, DamageDiceCount: 1, DamageDiceSides: 6},
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

func TestStartRoundIsDeterministicAndCoversLivingFighters(t *testing.T) {
	first := testBattle(t)
	second := testBattle(t)
	one, err := first.StartRound()
	if err != nil {
		t.Fatal(err)
	}
	two, err := second.StartRound()
	if err != nil {
		t.Fatal(err)
	}
	if len(one) != 2 || one[0] != two[0] || one[1] != two[1] || first.Round() != 1 {
		t.Fatalf("first=%v second=%v round=%d", one, two, first.Round())
	}
}

func TestResolveAttackNaturalOneMissesAndNaturalTwentyHits(t *testing.T) {
	battle := testBattle(t)
	miss, err := battle.ResolveAttack("hero", "goblin", 1, 8)
	if err != nil {
		t.Fatal(err)
	}
	if miss.Hit || miss.Damage != 0 || miss.TargetHP != 6 {
		t.Fatalf("natural one=%+v", miss)
	}
	hit, err := battle.ResolveAttack("hero", "goblin", 20, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !hit.Hit || !hit.Critical || hit.Damage != 6 || battle.Status() != StatusPartyWon {
		t.Fatalf("natural twenty=%+v status=%v", hit, battle.Status())
	}
}

func TestResolveAttackRequiresArmorClass(t *testing.T) {
	battle := testBattle(t)
	result, err := battle.ResolveAttack("hero", "goblin", 7, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Hit || result.Total != 10 || result.Damage != 4 {
		t.Fatalf("result=%+v", result)
	}
}

func TestCastMagicMissileUsesVerifiedDamageAndLevelScaling(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "mage", Name: "Mage", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10},
	}, 9)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastMagicMissile("mage", "goblin", 3)
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != 7 || result.Missiles != 2 || result.Damage < 4 || result.Damage > 10 || result.TargetHP != 30-result.Damage {
		t.Fatalf("spell result=%+v", result)
	}
}

func TestCastCureLightWoundsHealsOneToEightAndCapsAtMaxHP(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 3, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10},
	}, 3)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastCureLightWounds("cleric", "hero")
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != 3 || result.Healing < 1 || result.Healing > 7 || result.TargetHP != 3+result.Healing {
		t.Fatalf("healing result=%+v", result)
	}
}
