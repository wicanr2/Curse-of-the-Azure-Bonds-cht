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

func TestCastBlessImprovesLivingPartyAttackBonusOnce(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 2},
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, AttackBonus: 4},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastBless("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != 1 || result.Targets != 2 {
		t.Fatalf("result=%+v", result)
	}
	for _, fighter := range battle.Fighters() {
		if fighter.Side == SideParty && (!fighter.Blessed || (fighter.ID == "cleric" && fighter.AttackBonus != 3) || (fighter.ID == "hero" && fighter.AttackBonus != 5)) {
			t.Fatalf("blessed fighter=%+v", fighter)
		}
	}
	result, err = battle.CastBless("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if result.Targets != 0 {
		t.Fatalf("Bless stacked unexpectedly: %+v", result)
	}
}

func TestCastBlessSkipsAdjacentPartyAndExpiresAfterSixRounds(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 2, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, AttackBonus: 4, HasCombatPosition: true, CombatX: 6, CombatY: 6},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastBless("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if result.Targets != 1 {
		t.Fatalf("Bless affected=%d, want distant hero only", result.Targets)
	}
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "cleric" && (fighter.Blessed || fighter.AttackBonus != 2) {
			t.Fatalf("adjacent cleric was blessed: %+v", fighter)
		}
		if fighter.ID == "hero" && (fighter.BlessRounds != 6 || fighter.AttackBonus != 5) {
			t.Fatalf("distant hero=%+v", fighter)
		}
	}
	for round := 0; round < 5; round++ {
		if _, err := battle.StartRound(); err != nil {
			t.Fatal(err)
		}
	}
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "hero" && (fighter.BlessRounds != 1 || !fighter.Blessed) {
			t.Fatalf("hero before expiry=%+v", fighter)
		}
	}
	if _, err := battle.StartRound(); err != nil {
		t.Fatal(err)
	}
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "hero" && (fighter.BlessRounds != 0 || fighter.Blessed || fighter.AttackBonus != 4) {
			t.Fatalf("hero after expiry=%+v", fighter)
		}
	}
}

func TestCastCurseSkipsAdjacentEnemyAndExpiresAfterSixRounds(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "near", Name: "Near", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, AttackBonus: 3, HasCombatPosition: true, CombatX: 2, CombatY: 1},
		{ID: "far", Name: "Far", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, AttackBonus: 4, HasCombatPosition: true, CombatX: 6, CombatY: 6},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastCurse("cleric", "near")
	if err != nil {
		t.Fatal(err)
	}
	nearCursed := false
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "near" {
			nearCursed = fighter.Cursed
		}
	}
	if result.Targets != 0 || nearCursed {
		t.Fatalf("adjacent target was cursed: result=%+v fighters=%+v", result, battle.Fighters())
	}
	result, err = battle.CastCurse("cleric", "far")
	if err != nil {
		t.Fatal(err)
	}
	if result.Targets != 1 {
		t.Fatalf("distant target result=%+v", result)
	}
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "far" && (fighter.AttackBonus != 3 || fighter.CurseRounds != 6 || !fighter.Cursed) {
			t.Fatalf("cursed target=%+v", fighter)
		}
	}
	for round := 0; round < 6; round++ {
		if _, err := battle.StartRound(); err != nil {
			t.Fatal(err)
		}
	}
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "far" && (fighter.AttackBonus != 4 || fighter.Cursed || fighter.CurseRounds != 0) {
			t.Fatalf("expired curse=%+v", fighter)
		}
	}
}

func TestCastCauseLightWoundsDealsOneToEightTouchDamage(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "near", Name: "Near", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1},
	}, 4)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastCauseLightWounds("cleric", "near")
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != 4 || result.Damage < 1 || result.Damage > 8 || result.TargetHP != 20-result.Damage {
		t.Fatalf("cause light wounds result=%+v", result)
	}
	farBattle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "far", Name: "Far", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 4, CombatY: 4},
	}, 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := farBattle.CastCauseLightWounds("cleric", "far"); err == nil {
		t.Fatal("expected touch-range failure")
	}
}
