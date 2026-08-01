package combat

import "testing"

func TestSelectCombatTargetUsesSeededCandidateSelection(t *testing.T) {
	newBattle := func(seed int64) *Battle {
		battle, err := NewBattle([]Fighter{
			{ID: "ogre", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10},
			{ID: "cleric", Side: SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10},
			{ID: "fighter", Side: SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10},
		}, seed)
		if err != nil {
			t.Fatal(err)
		}
		return battle
	}

	first, err := newBattle(41).SelectCombatTarget("ogre", SideParty)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := newBattle(41).SelectCombatTarget("ogre", SideParty)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != replay.ID {
		t.Fatalf("same seed selected different targets: %q vs %q", first.ID, replay.ID)
	}
	seen := map[string]bool{first.ID: true}
	for seed := int64(42); seed < 64; seed++ {
		target, err := newBattle(seed).SelectCombatTarget("ogre", SideParty)
		if err != nil {
			t.Fatal(err)
		}
		seen[target.ID] = true
	}
	if len(seen) != 2 {
		t.Fatalf("seeded target selection never varied candidates: %v", seen)
	}
}

func TestSetHitPointsMarksDeathOverlayAndClearsCombatPosition(t *testing.T) {
	battle, err := NewBattle([]Fighter{{
		ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, HasCombatPosition: true, CombatX: 4, CombatY: 3,
		CombatAction: ActionState{Delay: 4, Move: 2, SpellID: 7, Guarding: true},
	}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.SetHitPoints("hero", 0); err != nil {
		t.Fatal(err)
	}
	fighter := battle.Fighters()[0]
	if fighter.HasCombatPosition || !fighter.DeathOverlay || fighter.CombatX != 4 || fighter.CombatY != 3 || fighter.CombatAction != (ActionState{}) {
		t.Fatalf("death state lost position anchor: %+v", fighter)
	}
	if err := battle.SetHitPoints("hero", 3); err != nil {
		t.Fatal(err)
	}
	fighter = battle.Fighters()[0]
	if fighter.DeathOverlay || fighter.HasCombatPosition {
		t.Fatalf("healing unexpectedly restored placement or overlay: %+v", fighter)
	}
}

func TestNewBattleNormalizesInitiallyDownedFighter(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 2, CombatY: 2,
			CombatAction: ActionState{Delay: 3, Move: 1, SpellID: 4, Guarding: true}},
		{ID: "goblin", Side: SideEnemy, HitPoints: 5, MaxHitPoints: 5, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 3, CombatY: 2},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	turns, err := battle.StartRound()
	if err != nil {
		t.Fatal(err)
	}
	var hero Fighter
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "hero" {
			hero = fighter
			break
		}
	}
	if hero.HasCombatPosition || !hero.DeathOverlay || hero.CombatAction != (ActionState{}) || len(turns) != 1 || turns[0].FighterID != "goblin" {
		t.Fatalf("initial downed fighter was not normalized: hero=%+v turns=%+v", hero, turns)
	}
}

func TestCureLightWoundsCanHealDownedPartyWithoutRestoringPlacement(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Side: SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, HasCombatPosition: true},
		{ID: "hero", Side: SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true},
		{ID: "goblin", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, HasCombatPosition: true},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastCureLightWounds("cleric", "hero")
	if err != nil {
		t.Fatal(err)
	}
	var hero Fighter
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "hero" {
			hero = fighter
		}
	}
	if result.Healing <= 0 || hero.HitPoints <= 0 || hero.DeathOverlay || !hero.DownedCorpse || hero.HasCombatPosition {
		t.Fatalf("downed cure changed wrong boundaries: result=%+v hero=%+v", result, hero)
	}
}

func TestRestoreCombatantStandsUpAtExplicitPlacement(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Side: SideParty, HitPoints: 8, MaxHitPoints: 8, ArmorClass: 10, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "hero", Side: SideParty, HitPoints: 0, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.SetHitPoints("hero", 4); err != nil {
		t.Fatal(err)
	}
	if err := battle.RestoreCombatant("hero", TilePoint{X: 2, Y: 1}); err != nil {
		t.Fatal(err)
	}
	fighters := battle.Fighters()
	for _, fighter := range fighters {
		if fighter.ID == "hero" && (!fighter.HasCombatPosition || fighter.DownedCorpse || fighter.DeathOverlay || fighter.CombatX != 2 || fighter.CombatY != 1) {
			t.Fatalf("restore did not stand hero up: %+v", fighter)
		}
	}
}

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

func TestStartRoundPreservesConstructionTeamListOrder(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "z-first", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, Dexterity: 15},
		{ID: "a-second", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, Dexterity: 15},
		{ID: "m-third", Side: SideEnemy, HitPoints: 0, MaxHitPoints: 10, Dexterity: 15},
	}, 419)
	if err != nil {
		t.Fatal(err)
	}
	if got := battle.fighterOrder; len(got) != 3 || got[0] != "z-first" || got[1] != "a-second" || got[2] != "m-third" {
		t.Fatalf("TeamList order=%v", got)
	}
	turns, err := battle.StartRound()
	if err != nil || len(turns) != 2 {
		t.Fatalf("turns=%v err=%v", turns, err)
	}
	for _, turn := range turns {
		fighter, _ := battle.Fighter(turn.FighterID)
		if fighter.CombatAction.Delay != turn.Initiative || turn.Initiative < 1 || turn.Initiative > 6 {
			t.Fatalf("turn=%+v fighter=%+v", turn, fighter)
		}
	}
}

func TestMoveChangesPositionAndRejectsOccupiedDestination(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "ally", Name: "Ally", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 3, CombatY: 2},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	moved, err := battle.Move("hero", -1, 0)
	if err != nil || moved.CombatX != 1 || moved.CombatY != 2 {
		t.Fatalf("moved=%+v err=%v", moved, err)
	}
	if _, err := battle.Move("hero", 1, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := battle.Move("hero", 1, 0); err == nil {
		t.Fatal("expected occupied destination error")
	}
	if _, err := battle.Move("hero", 2, 0); err == nil {
		t.Fatal("expected multi-step error")
	}
}

func TestMoveIntoEnemySquareResolvesAttackWithoutChangingPosition(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 3, CombatY: 2},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.MoveWithFreeAttacks("hero", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Attack == nil || result.Attack.AttackerID != "hero" || result.Attack.TargetID != "goblin" || len(result.FreeAttacks) != 0 {
		t.Fatalf("move attack=%+v free attacks=%+v", result.Attack, result.FreeAttacks)
	}
	if result.Fighter.CombatX != 2 || result.Fighter.CombatY != 2 {
		t.Fatalf("fighter moved onto enemy square: %+v", result.Fighter)
	}
}

func TestMoveWithFreeAttacksTriggersWhenLeavingEnemyAdjacency(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, DamageDiceCount: 1, DamageDiceSides: 1, AttackBonus: 20, HasCombatPosition: true, CombatX: 3, CombatY: 2},
	}, 6)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.MoveWithFreeAttacks("hero", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FreeAttacks) != 1 || result.FreeAttacks[0].AttackerID != "goblin" || result.FreeAttacks[0].TargetID != "hero" {
		t.Fatalf("free attacks=%+v", result.FreeAttacks)
	}
}

func TestGuardActionAttacksOnceWhenEnemyEntersAdjacency(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "guard", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "enemy", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 2},
	}, 421)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.GuardAction("guard"); err != nil {
		t.Fatal(err)
	}
	result, err := battle.MoveWithFreeAttacks("enemy", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.GuardAttacks) != 1 || result.GuardAttacks[0].AttackerID != "guard" || result.GuardAttacks[0].TargetID != "enemy" {
		t.Fatalf("guard attacks=%+v", result.GuardAttacks)
	}
	guard, _ := battle.Fighter("guard")
	if guard.CombatAction.Guarding {
		t.Fatal("guard flag was not consumed before reaction attack")
	}
	result, err = battle.MoveWithFreeAttacks("enemy", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.GuardAttacks) != 0 {
		t.Fatalf("guard reacted twice: %+v", result.GuardAttacks)
	}
}

func TestGuardActionHeldSuppressionAndMissileEligibility(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "held", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			MonsterAffects:    []MonsterAffect{{Kind: 0x1F, Active: true}},
			HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "archer", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			MissileWeapon: true, HasCombatPosition: true, CombatX: 2, CombatY: 4},
		{ID: "enemy", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 2},
	}, 421)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.GuardAction("archer"); err == nil {
		t.Fatal("pure missile weapon unexpectedly allowed Guard")
	}
	if err := battle.GuardAction("held"); err != nil {
		t.Fatal(err)
	}
	result, err := battle.MoveWithFreeAttacks("enemy", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.GuardAttacks) != 0 {
		t.Fatalf("held guard reacted: %+v", result.GuardAttacks)
	}
	guard, _ := battle.Fighter("held")
	if !guard.CombatAction.Guarding {
		t.Fatal("suppressed guard should remain armed until its next turn")
	}
}

func TestQuickFightManualControlUsesControlMoraleNamespace(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "pc", Side: SideParty, QuickFight: true, ControlMorale: 0x00, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "npc", Side: SideParty, QuickFight: true, ControlMorale: 0x80, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "enemy", Side: SideEnemy, QuickFight: true, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
	}, 421)
	if err != nil {
		t.Fatal(err)
	}
	if changed := battle.SetPlayerCharactersManual(); changed != 1 {
		t.Fatalf("changed=%d want 1", changed)
	}
	pc, _ := battle.Fighter("pc")
	npc, _ := battle.Fighter("npc")
	enemy, _ := battle.Fighter("enemy")
	if pc.QuickFight || !npc.QuickFight || !enemy.QuickFight {
		t.Fatalf("control projection pc=%v npc=%v enemy=%v", pc.QuickFight, npc.QuickFight, enemy.QuickFight)
	}
}

func TestAllQuickFightPreservesTeamListAndConsumesTwentyToNineteenHandoff(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "pc", Side: SideParty, ControlMorale: 0, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "npc", Side: SideParty, ControlMorale: 0x80, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "enemy", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
	}, 422)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.SetAllQuickFight("pc"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"pc", "npc", "enemy"} {
		fighter, _ := battle.Fighter(id)
		if !fighter.QuickFight {
			t.Fatalf("%s was not delegated", id)
		}
	}
	pc, _ := battle.Fighter("pc")
	if pc.CombatAction.Delay != 20 {
		t.Fatalf("handoff delay=%d want 20", pc.CombatAction.Delay)
	}
	if err := battle.BeginQuickFightAction("pc"); err != nil {
		t.Fatal(err)
	}
	pc, _ = battle.Fighter("pc")
	if pc.CombatAction.Delay != 19 {
		t.Fatalf("AI entry delay=%d want 19", pc.CombatAction.Delay)
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

func TestAttackDispatchesOperational4FAfterLivingTargetHit(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "flamed", Side: SideEnemy, HitPoints: 30, MaxHitPoints: 30,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
			MonsterAffects: []MonsterAffect{{Kind: 0x4F, Innate: true}}},
		{ID: "hero", Side: SideParty, HitPoints: 50, MaxHitPoints: 50, ArmorClass: 10},
	}, 414)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.Attack("flamed", "hero")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Hit || result.Damage != 1 || len(result.Effects) != 1 {
		t.Fatalf("post-hit result=%+v", result)
	}
	effect := result.Effects[0]
	if effect.Kind != 0x4F || effect.DamageFlags != DamageFlagFire|DamageFlagMagic ||
		effect.RolledDamage < 2 || effect.RolledDamage > 20 || effect.Damage != effect.RolledDamage || effect.Protected ||
		result.TargetHP != 49-effect.Damage || effect.TargetHP != result.TargetHP {
		t.Fatalf("4F effect=%+v result=%+v", effect, result)
	}
}

func TestAttack4FRequiresOperationalEffectHitAndLivingTarget(t *testing.T) {
	tests := []struct {
		name     string
		affect   MonsterAffect
		targetHP int
		wantHit  bool
	}{
		{name: "inactive", affect: MonsterAffect{Kind: 0x4F}, targetHP: 10, wantHit: true},
		{name: "physical kill", affect: MonsterAffect{Kind: 0x4F, Innate: true}, targetHP: 1, wantHit: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			battle, err := NewBattle([]Fighter{
				{ID: "attacker", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20,
					AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1,
					MonsterAffects: []MonsterAffect{test.affect}},
				{ID: "target", Side: SideParty, HitPoints: test.targetHP, MaxHitPoints: test.targetHP, ArmorClass: 10},
			}, 414)
			if err != nil {
				t.Fatal(err)
			}
			result, err := battle.Attack("attacker", "target")
			if err != nil {
				t.Fatal(err)
			}
			if result.Hit != test.wantHit || len(result.Effects) != 0 {
				t.Fatalf("result=%+v", result)
			}
		})
	}

	// ResolveAttack is the injected physical boundary. The original 4F caller
	// lives in the higher attack-slot scheduler, so a miss there cannot dispatch
	// a post-hit effect.
	battle, err := NewBattle([]Fighter{
		{ID: "attacker", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, AttackBonus: -100,
			DamageDiceCount: 1, DamageDiceSides: 1, MonsterAffects: []MonsterAffect{{Kind: 0x4F, Innate: true}}},
		{ID: "target", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	miss, err := battle.ResolveAttack("attacker", "target", 2, 1)
	if err != nil || miss.Hit || len(miss.Effects) != 0 {
		t.Fatalf("miss=%+v err=%v", miss, err)
	}
}

func TestAttack4FHonorsFireProtectionAndOnlyFirstTwoSlots(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "flamed", Side: SideEnemy, HitPoints: 30, MaxHitPoints: 30,
			AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, AttacksPerTurn: 3,
			MonsterAffects: []MonsterAffect{{Kind: 0x4F, Innate: true}}},
		{ID: "warded", Side: SideParty, HitPoints: 50, MaxHitPoints: 50, ArmorClass: 10,
			MonsterAffects: []MonsterAffect{{Kind: 0x70, Innate: true}}},
	}, 414)
	if err != nil {
		t.Fatal(err)
	}
	results, err := battle.AttackSequence("flamed", "warded")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 || len(results[0].Effects) != 1 || len(results[1].Effects) != 1 || len(results[2].Effects) != 0 {
		t.Fatalf("slot results=%+v", results)
	}
	for index := 0; index < 2; index++ {
		if !results[index].Effects[0].Protected || results[index].Effects[0].Damage != 0 ||
			results[index].Effects[0].RolledDamage < 2 || results[index].Effects[0].RolledDamage > 20 {
			t.Fatalf("slot %d effect=%+v", index+1, results[index].Effects[0])
		}
	}
	if results[2].TargetHP != 47 {
		t.Fatalf("protected sequence final HP=%d, want 47", results[2].TargetHP)
	}
}

func TestMonsterThrowsLightningRequiresOperationalEffect84(t *testing.T) {
	if (Fighter{MonsterAffects: []MonsterAffect{{Kind: 0x84}}}).MonsterThrowsLightning() {
		t.Fatal("inactive effect 84 became operational")
	}
	if !(Fighter{MonsterAffects: []MonsterAffect{{Kind: 0x84, Innate: true}}}).MonsterThrowsLightning() {
		t.Fatal("innate effect 84 was not projected")
	}
	if !(Fighter{MonsterAffects: []MonsterAffect{{Kind: 0x84, Active: true}}}).MonsterThrowsLightning() {
		t.Fatal("active effect 84 was not projected")
	}
}

func TestSideAttackRollModifierIsBattleScopedAndSigned(t *testing.T) {
	battle := testBattle(t)
	if err := battle.SetSideAttackRollModifier(SideParty, 2); err != nil {
		t.Fatal(err)
	}
	if err := battle.SetSideAttackRollModifier(SideEnemy, -2); err != nil {
		t.Fatal(err)
	}
	result, err := battle.ResolveAttack("hero", "goblin", 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Hit || result.Total != 10 {
		t.Fatalf("party modifier result=%+v", result)
	}
	hero, _ := battle.Fighter("hero")
	if hero.AttackBonus != 3 || battle.SideAttackRollModifier(SideParty) != 2 {
		t.Fatalf("persistent fighter was mutated: hero=%+v modifier=%d", hero, battle.SideAttackRollModifier(SideParty))
	}
	if err := battle.SetSideAttackRollModifier(Side(99), 1); err == nil {
		t.Fatal("invalid combat side accepted")
	}
}

func TestResolveAttackProjectsMonsterInvisibilityACBonus(t *testing.T) {
	base := []Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1},
		{ID: "invisible", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, MonsterAffects: []MonsterAffect{{Kind: 0x19, Active: true}}},
	}
	battle, err := NewBattle(base, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.ResolveAttack("hero", "invisible", 13, 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Hit {
		t.Fatalf("invisible target was hit at adjusted AC boundary: %#v", result)
	}
	base[1].MonsterAffects[0].Active = false
	battle, err = NewBattle(base, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err = battle.ResolveAttack("hero", "invisible", 13, 1)
	if err != nil || !result.Hit {
		t.Fatalf("inactive invisibility should not alter hit: result=%#v err=%v", result, err)
	}
}

func TestResolveAttackInnateDetectInvisibleBypassesInvisibilityACBonus(t *testing.T) {
	fighters := []Fighter{
		{ID: "seer", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1,
			MonsterAffects: []MonsterAffect{{Kind: 0x18, Innate: true}}},
		{ID: "invisible", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			MonsterAffects: []MonsterAffect{{Kind: 0x19, Active: true}}},
	}
	battle, err := NewBattle(fighters, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.ResolveAttack("seer", "invisible", 10, 1)
	if err != nil || !result.Hit {
		t.Fatalf("innate detect-invisible result=%#v err=%v", result, err)
	}

	fighters[0].MonsterAffects[0].Innate = false
	battle, err = NewBattle(fighters, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err = battle.ResolveAttack("seer", "invisible", 10, 1)
	if err != nil || result.Hit {
		t.Fatalf("inactive non-innate detect-invisible result=%#v err=%v", result, err)
	}
}

func TestResolveAttackDetectInvisibleDoesNotBypassEffect47(t *testing.T) {
	fighters := []Fighter{
		{ID: "seer", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1,
			MonsterAffects: []MonsterAffect{{Kind: 0x18, Innate: true}}},
		{ID: "hidden", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			MonsterAffects: []MonsterAffect{{Kind: 0x47, Active: true}}},
	}
	battle, err := NewBattle(fighters, 417)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.ResolveAttack("seer", "hidden", 10, 1)
	if err != nil || result.Hit {
		t.Fatalf("effect 47 must retain its AC bonus against effect 18: result=%#v err=%v", result, err)
	}
}

func TestResolveAttackBlinkOverridesNatural20UntilActionDelay(t *testing.T) {
	fighters := []Fighter{
		{ID: "attacker", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10,
			AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1},
		{ID: "blinking", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 0,
			MonsterAffects: []MonsterAffect{{Kind: 0x25, Active: true}}},
	}
	battle, err := NewBattle(fighters, 418)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.ResolveAttack("attacker", "blinking", 20, 1)
	if err != nil || result.Hit || result.Critical {
		t.Fatalf("zero-delay blink must override natural 20: result=%#v err=%v", result, err)
	}
	fighters[1].CombatAction.Delay = 1
	battle, err = NewBattle(fighters, 418)
	if err != nil {
		t.Fatal(err)
	}
	result, err = battle.ResolveAttack("attacker", "blinking", 20, 1)
	if err != nil || !result.Hit || !result.Critical {
		t.Fatalf("nonzero-delay blink result=%#v err=%v", result, err)
	}
}

func TestStartRoundAndCompleteActionProjectDelayLifecycle(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "fast", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, InitiativeBonus: 10},
		{ID: "slow", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, InitiativeBonus: -100},
	}, 418)
	if err != nil {
		t.Fatal(err)
	}
	turns, err := battle.StartRound()
	if err != nil || len(turns) != 2 {
		t.Fatalf("turns=%+v err=%v", turns, err)
	}
	for _, turn := range turns {
		fighter, _ := battle.Fighter(turn.FighterID)
		if fighter.CombatAction.Delay <= 0 {
			t.Fatalf("scheduled fighter has zero delay: %+v", fighter)
		}
	}
	if err := battle.CompleteAction(turns[0].FighterID); err != nil {
		t.Fatal(err)
	}
	completed, _ := battle.Fighter(turns[0].FighterID)
	pending, _ := battle.Fighter(turns[1].FighterID)
	if completed.CombatAction.Delay != 0 || pending.CombatAction.Delay <= 0 {
		t.Fatalf("delay lifecycle completed=%+v pending=%+v", completed.CombatAction, pending.CombatAction)
	}
}

func TestDynamicRoundDelayActionReentersAtTierOne(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "first", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, InitiativeBonus: 20},
		{ID: "other", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, InitiativeBonus: 10},
	}, 420)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.BeginScheduledRound(); err != nil {
		t.Fatal(err)
	}
	first, ok, err := battle.NextScheduledTurn()
	if err != nil || !ok || first.FighterID != "first" {
		t.Fatalf("first=%+v ok=%v err=%v", first, ok, err)
	}
	if err := battle.DelayAction("first"); err != nil {
		t.Fatal(err)
	}
	other, ok, err := battle.NextScheduledTurn()
	if err != nil || !ok || other.FighterID != "other" {
		t.Fatalf("other=%+v ok=%v err=%v", other, ok, err)
	}
	if err := battle.CompleteAction("other"); err != nil {
		t.Fatal(err)
	}
	delayed, ok, err := battle.NextScheduledTurn()
	if err != nil || !ok || delayed.FighterID != "first" || delayed.Initiative != 1 {
		t.Fatalf("delayed=%+v ok=%v err=%v", delayed, ok, err)
	}
}

func TestPendingSpellActionSubtractsCastingDelayAndReenters(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "caster", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, InitiativeBonus: 8},
		{ID: "other", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, InitiativeBonus: 6},
	}, 425)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.BeginScheduledRound(); err != nil {
		t.Fatal(err)
	}
	first, ok, err := battle.NextScheduledTurn()
	if err != nil || !ok || first.FighterID != "caster" || first.Initiative != 8 {
		t.Fatalf("first=%+v ok=%v err=%v", first, ok, err)
	}
	if err := battle.BeginPendingSpellAction("caster", 1, 3); err != nil {
		t.Fatal(err)
	}
	pending, _ := battle.Fighter("caster")
	if pending.CombatAction.SpellID != 1 || pending.CombatAction.Delay != 5 {
		t.Fatalf("pending action=%+v", pending.CombatAction)
	}
	other, ok, err := battle.NextScheduledTurn()
	if err != nil || !ok || other.FighterID != "other" {
		t.Fatalf("other=%+v ok=%v err=%v", other, ok, err)
	}
	if err := battle.CompleteAction("other"); err != nil {
		t.Fatal(err)
	}
	resumed, ok, err := battle.NextScheduledTurn()
	if err != nil || !ok || resumed.FighterID != "caster" || resumed.Initiative != 5 {
		t.Fatalf("resumed=%+v ok=%v err=%v", resumed, ok, err)
	}
	spellID, err := battle.TakePendingSpellAction("caster")
	if err != nil || spellID != 1 {
		t.Fatalf("take spell=0x%02X err=%v", spellID, err)
	}
}

func TestResolveAttackAnimalInvisibilityKeepsPenaltyWhenDetected(t *testing.T) {
	fighters := []Fighter{
		{ID: "animal", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10,
			MonsterType: MonsterTypeAnimal, AttackBonus: 0, DamageDiceCount: 1, DamageDiceSides: 1,
			MonsterAffects: []MonsterAffect{{Kind: 0x18, Innate: true}}},
		{ID: "hidden", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10,
			MonsterAffects: []MonsterAffect{{Kind: 0x45, Active: true}}},
	}
	battle, err := NewBattle(fighters, 418)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.ResolveAttack("animal", "hidden", 10, 1)
	if err != nil || result.Hit {
		t.Fatalf("effect 45 must keep -4 against detecting animal: result=%#v err=%v", result, err)
	}
	if !fighters[1].VisibleTo(fighters[0]) {
		t.Fatal("operational effect 18 should prevent effect 45 target-hidden flag")
	}
}

func TestResolveAttackHeldMonsterIsAlwaysHit(t *testing.T) {
	fighters := []Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: -20, DamageDiceCount: 1, DamageDiceSides: 1},
		{ID: "sleeping", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, MonsterAffects: []MonsterAffect{{Kind: 0x35, Active: true}}},
	}
	battle, err := NewBattle(fighters, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.ResolveAttack("hero", "sleeping", 1, 1)
	if err != nil || !result.Hit || result.Damage != 1 {
		t.Fatalf("held target result=%#v err=%v", result, err)
	}
	if !battle.Fighters()[1].MonsterIsHeld() {
		t.Fatal("held effect was unexpectedly consumed")
	}
}

func TestAdvanceMonsterAffectsUsesFiniteMinutesAndPermanentMarker(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "monster", Side: SideEnemy, HitPoints: 5, MaxHitPoints: 5,
			MonsterAffects: []MonsterAffect{
				{Kind: 0x35, Duration: 3, Value: 3, Strength: 1, Active: true},
				{Kind: 0x27, Duration: 3, Value: 3, Strength: 0xFF, Active: true},
			}}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if removed := battle.AdvanceMonsterAffects(2); removed != 0 {
		t.Fatalf("removed=%d after two minutes", removed)
	}
	fighter := battle.Fighters()[0]
	if fighter.MonsterAffects[0].Duration != 1 || len(fighter.MonsterAffects) != 2 {
		t.Fatalf("after two minutes=%#v", fighter.MonsterAffects)
	}
	if removed := battle.AdvanceMonsterAffects(1); removed != 1 || len(battle.Fighters()[0].MonsterAffects) != 1 {
		t.Fatalf("expiry removed=%d effects=%#v", removed, battle.Fighters()[0].MonsterAffects)
	}
}

func TestResolveAttackRejectsAdjacentMissileButAllowsDartException(t *testing.T) {
	missile, err := NewBattle([]Fighter{
		{ID: "archer", Name: "Archer", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, HasCombatPosition: true, CombatX: 0, CombatY: 0, WeaponRange: 22, MissileWeapon: true},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 0, HasCombatPosition: true, CombatX: 1, CombatY: 0},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := missile.Attack("archer", "goblin"); err == nil {
		t.Fatal("expected adjacent missile attack rejection")
	}
	dart, err := NewBattle([]Fighter{
		{ID: "dart", Name: "Dart", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, HasCombatPosition: true, CombatX: 0, CombatY: 0, WeaponRange: 6, MissileWeapon: true, ThrownWeapon: true},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 0, HasCombatPosition: true, CombatX: 1, CombatY: 0},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dart.Attack("dart", "goblin"); err != nil {
		t.Fatalf("dart adjacent exception rejected: %v", err)
	}
}

func TestValidateAttackRejectsAdjacentMissileBeforeRandomRoll(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "archer", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, MissileWeapon: true, HasCombatPosition: true, CombatX: 0, CombatY: 0},
		{ID: "goblin", Side: SideEnemy, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, HasCombatPosition: true, CombatX: 1, CombatY: 0},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.ValidateAttack("archer", "goblin"); err == nil {
		t.Fatal("expected adjacent missile preflight rejection")
	}
}

func TestAttackSequenceUsesAttacksPerTurnAndStopsAtDefeat(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "archer", Name: "Archer", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 1, AttacksPerTurn: 3},
		{ID: "goblin", Name: "Goblin", Side: SideEnemy, HitPoints: 2, MaxHitPoints: 2, ArmorClass: 0},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	results, err := battle.AttackSequence("archer", "goblin")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || !results[0].Hit || !results[1].Hit || results[1].TargetHP != 0 || battle.Status() != StatusPartyWon {
		t.Fatalf("sequence=%+v status=%v", results, battle.Status())
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

func TestMagicResistanceChanceUsesPC98LevelAdjustment(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{level: 10, want: 20},
		{level: 11, want: 15},
		{level: 12, want: 10},
	}
	for _, test := range tests {
		if got := MagicResistanceChance(15, test.level); got != test.want {
			t.Fatalf("level %d chance=%d, want %d", test.level, got, test.want)
		}
	}
}

func TestCastMagicMissileHonorsOperationalEffect6A(t *testing.T) {
	resisted := false
	notResisted := false
	for seed := int64(1); seed <= 128 && (!resisted || !notResisted); seed++ {
		battle, err := NewBattle([]Fighter{
			{ID: "mage", Side: SideParty, HitPoints: 20, MaxHitPoints: 20},
			{ID: "tyranthraxus", Side: SideEnemy, HitPoints: 30, MaxHitPoints: 30,
				MonsterAffects: []MonsterAffect{{Kind: 0x6A, Innate: true}}},
		}, seed)
		if err != nil {
			t.Fatal(err)
		}
		result, err := battle.CastMagicMissile("mage", "tyranthraxus", 11)
		if err != nil {
			t.Fatal(err)
		}
		if result.Resisted {
			resisted = true
			if result.Damage != 0 || result.TargetHP != 30 {
				t.Fatalf("resisted result=%+v", result)
			}
		} else {
			notResisted = true
			if result.Damage == 0 || result.TargetHP >= 30 {
				t.Fatalf("unresisted result=%+v", result)
			}
		}
	}
	if !resisted || !notResisted {
		t.Fatalf("deterministic seeds did not cover both outcomes: resisted=%v unresisted=%v", resisted, notResisted)
	}

	inactive, err := NewBattle([]Fighter{
		{ID: "mage", Side: SideParty, HitPoints: 20, MaxHitPoints: 20},
		{ID: "target", Side: SideEnemy, HitPoints: 30, MaxHitPoints: 30,
			MonsterAffects: []MonsterAffect{{Kind: 0x6A}}},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := inactive.CastMagicMissile("mage", "target", 1)
	if err != nil || result.Resisted || result.Damage == 0 {
		t.Fatalf("inactive effect 6A result=%+v err=%v", result, err)
	}
}

func TestCastFireballUsesOneDamageRollAndHitsBothSidesInRadius(t *testing.T) {
	saves := []uint8{20, 20, 20, 20, 20}
	battle, err := NewBattle([]Fighter{
		{ID: "mage", Name: "Mage", Side: SideParty, HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 0, CombatY: 0, SavingThrows: saves},
		{ID: "ally", Name: "Ally", Side: SideParty, HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 3, CombatY: 2, SavingThrows: saves},
		{ID: "near", Name: "Near", Side: SideEnemy, HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 3, SavingThrows: saves},
		{ID: "large", Name: "Large", Side: SideEnemy, HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 4, CombatSize: 2, SavingThrows: saves},
		{ID: "far", Name: "Far", Side: SideEnemy, HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 8, CombatY: 8, SavingThrows: saves},
		{ID: "corner", Name: "Corner", Side: SideEnemy, HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 6, CombatY: 5, SavingThrows: saves},
	}, 17)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastFireball("mage", TilePoint{X: 4, Y: 3}, 5)
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != FireballSpellID || result.BaseDamage < 5 || result.BaseDamage > 30 ||
		len(result.Impacts) != 3 {
		t.Fatalf("fireball result=%+v", result)
	}
	hit := map[string]AreaSpellImpact{}
	for _, impact := range result.Impacts {
		hit[impact.TargetID] = impact
		wantDamage := result.BaseDamage
		if impact.Saved {
			wantDamage /= 2
		}
		if impact.Damage != wantDamage {
			t.Fatalf("impact=%+v, shared damage=%d", impact, result.BaseDamage)
		}
	}
	if _, ok := hit["ally"]; !ok {
		t.Fatal("friendly combatant inside Fireball was not hit")
	}
	if _, ok := hit["near"]; !ok {
		t.Fatal("enemy at center was not hit")
	}
	if _, ok := hit["large"]; !ok {
		t.Fatal("large target footprint intersecting radius was not hit")
	}
	for _, fighter := range battle.Fighters() {
		if (fighter.ID == "mage" || fighter.ID == "far" || fighter.ID == "corner") && fighter.HitPoints != 40 {
			t.Fatalf("outside fighter was hit: %+v", fighter)
		}
	}
}

func TestCastFireballHonorsOperationalEffect70FireProtection(t *testing.T) {
	saves := []uint8{20, 20, 20, 20, 20}
	battle, err := NewBattle([]Fighter{
		{ID: "mage", Side: SideParty, HitPoints: 20, MaxHitPoints: 20,
			HasCombatPosition: true, CombatX: 0, CombatY: 0, SavingThrows: saves},
		{ID: "protected", Side: SideEnemy, HitPoints: 40, MaxHitPoints: 40,
			HasCombatPosition: true, CombatX: 2, CombatY: 2, SavingThrows: saves,
			MonsterAffects: []MonsterAffect{{Kind: 0x70, Innate: true}}},
		{ID: "inactive", Side: SideEnemy, HitPoints: 40, MaxHitPoints: 40,
			HasCombatPosition: true, CombatX: 3, CombatY: 2, SavingThrows: saves,
			MonsterAffects: []MonsterAffect{{Kind: 0x70}}},
	}, 23)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastFireball("mage", TilePoint{X: 2, Y: 2}, 3)
	if err != nil {
		t.Fatal(err)
	}
	impacts := make(map[string]AreaSpellImpact)
	for _, impact := range result.Impacts {
		impacts[impact.TargetID] = impact
	}
	if impact := impacts["protected"]; !impact.Protected || impact.Damage != 0 || impact.TargetHP != 40 {
		t.Fatalf("protected fireball impact=%+v", impact)
	}
	if impact := impacts["inactive"]; impact.Protected || impact.Damage == 0 || impact.TargetHP == 40 {
		t.Fatalf("inactive fire protection impact=%+v", impact)
	}
}

func TestCastMonsterMagicMissileConsumesRawLevelOneUse(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "mage-monster", Name: "施法怪", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10,
			MonsterSpellIDs: []uint8{MonsterMagicMissileSpellID}, MonsterSpellUses: [3]uint8{1}},
		{ID: "hero", Name: "Hero", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10},
	}, 9)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastMonsterMagicMissile("mage-monster", "hero")
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != MonsterMagicMissileSpellID || result.Missiles != 1 || result.Damage < 2 || result.Damage > 5 {
		t.Fatalf("monster spell result=%+v", result)
	}
	if _, err := battle.CastMonsterMagicMissile("mage-monster", "hero"); err == nil {
		t.Fatal("monster spell use was not consumed")
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

func TestProtectionFromEvilAddsConditionalACAndExpires(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "orc", Name: "Orc", Side: SideEnemy, Evil: true, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, AttackBonus: 8},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastProtectionFromEvil("cleric", "cleric", 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != 6 || result.Targets != 1 {
		t.Fatalf("protection result=%+v", result)
	}
	var protected Fighter
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "cleric" {
			protected = fighter
		}
	}
	if !protected.ProtectedFromEvil || protected.ProtectionEvilRounds != 3 || protected.ArmorClass != 10 {
		t.Fatalf("protected fighter=%+v", protected)
	}
	miss, err := battle.ResolveAttack("orc", "cleric", 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if miss.Hit {
		t.Fatalf("evil attacker ignored protection: %+v", miss)
	}
	for round := 0; round < 3; round++ {
		if _, err := battle.StartRound(); err != nil {
			t.Fatal(err)
		}
	}
	for _, fighter := range battle.Fighters() {
		if fighter.ID == "cleric" && (fighter.ProtectedFromEvil || fighter.ProtectionEvilRounds != 0) {
			t.Fatalf("protection did not expire: %+v", fighter)
		}
	}
}

func TestProtectionFromGoodUsesClassSpecificSpellIDAndConditionalAC(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "cleric", Name: "Cleric", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10},
		{ID: "paladin", Name: "Paladin", Side: SideEnemy, Good: true, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, AttackBonus: 8},
	}, 3)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastProtectionFromGood("cleric", "cleric", 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellID != 7 || result.Targets != 1 {
		t.Fatalf("protection from good result=%+v", result)
	}
	miss, err := battle.ResolveAttack("paladin", "cleric", 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if miss.Hit {
		t.Fatalf("good attacker ignored protection: %+v", miss)
	}
}
