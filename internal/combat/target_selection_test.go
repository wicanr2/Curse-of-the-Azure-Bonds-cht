package combat

import "testing"

func TestSelectRangedCombatTargetFiltersRangeWallsAndUsesFootprints(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "caster", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, HasCombatPosition: true, CombatX: 1, CombatY: 1, CombatSize: 4},
		{ID: "footprint-edge", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 11, CombatY: 1, CombatSize: 3},
		{ID: "blocked", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 5, CombatY: 4},
		{ID: "too-far", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 13, CombatY: 1},
	}, 416)
	if err != nil {
		t.Fatal(err)
	}
	terrain := func(x, y int) LineCell {
		if x == 5 && y == 4 {
			return LineCell{Valid: true, Reflect: true}
		}
		return LineCell{Valid: x >= 0 && x < 20 && y >= 0 && y < 10}
	}
	target, found, err := battle.SelectRangedCombatTarget("caster", SideParty, TargetSelectionOptions{
		MaxRange: 10, Terrain: terrain,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !found || target.ID != "footprint-edge" {
		t.Fatalf("ranged target=%+v found=%v", target, found)
	}
}

func TestSelectRangedCombatTargetRemovesInvisibleCandidatesBeforeRetry(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "caster", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "hidden-a", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1},
		{ID: "visible", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 3, CombatY: 1},
		{ID: "hidden-b", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 4, CombatY: 1},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	visibilityChecks := 0
	target, found, err := battle.SelectRangedCombatTarget("caster", SideParty, TargetSelectionOptions{
		MaxRange: 10,
		Terrain:  func(x, y int) LineCell { return LineCell{Valid: true} },
		VisibleTo: func(attacker, target Fighter) bool {
			visibilityChecks++
			return target.ID == "visible"
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !found || target.ID != "visible" || visibilityChecks < 1 || visibilityChecks > 3 {
		t.Fatalf("target=%+v found=%v visibility checks=%d", target, found, visibilityChecks)
	}
}

func TestSelectRangedCombatTargetReturnsNormalMissWhenEveryCandidateIsHidden(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "caster", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "hidden", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	target, found, err := battle.SelectRangedCombatTarget("caster", SideParty, TargetSelectionOptions{
		MaxRange: 10,
		Terrain:  func(x, y int) LineCell { return LineCell{Valid: true} },
		VisibleTo: func(attacker, target Fighter) bool {
			return false
		},
	})
	if err != nil || found || target.ID != "" {
		t.Fatalf("target=%+v found=%v err=%v", target, found, err)
	}
}
