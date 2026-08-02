package combat

import (
	"reflect"
	"testing"

	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
)

func TestOrderScanTargetIDsUsesLegacyObjectDistanceAndIgnoresDirection(t *testing.T) {
	records := []ScanTargetRecord{
		{ObjectID: 8, TargetID: "eight", Distance: 5, Direction: 0},
		{ObjectID: 7, TargetID: "seven", Distance: 5, Direction: 7},
		{ObjectID: 6, TargetID: "six", Distance: 5, Direction: 6},
		{ObjectID: 3, TargetID: "near", Distance: 2, Direction: 5},
	}
	got, err := OrderScanTargetIDs(records)
	want := []string{"near", "six", "eight", "seven"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("ordered=%v want=%v err=%v", got, want, err)
	}
	if records[0].TargetID != "eight" {
		t.Fatalf("input records mutated: %v", records)
	}
}

func TestOrderScanTargetIDsRejectsAmbiguousProjection(t *testing.T) {
	tests := [][]ScanTargetRecord{
		{{ObjectID: 0, TargetID: "zero"}},
		{{ObjectID: 1, TargetID: ""}},
		{{ObjectID: 1, TargetID: "one"}, {ObjectID: 1, TargetID: "two"}},
		{{ObjectID: 1, TargetID: "same"}, {ObjectID: 2, TargetID: "same"}},
	}
	for _, records := range tests {
		if _, err := OrderScanTargetIDs(records); err == nil {
			t.Fatalf("accepted ambiguous SCAN records %v", records)
		}
	}
}

func TestBuildScanTargetIDsFeedsSleepWithoutDisplayTextOrSliceIdentity(t *testing.T) {
	tacticalMap := enginescan.TacticalMap{
		Width: 6, Height: 1,
		Tiles: []uint8{1, 1, 1, 2, 1, 1},
		Definitions: []enginescan.TerrainDefinition{
			{LOS: 1, SYM: 0},
			{LOS: 1, SYM: 2},
		},
	}
	ordered, err := BuildScanTargetIDs(
		tacticalMap,
		1,
		[]enginescan.Point{{X: 0}},
		[]ScanTargetCandidate{
			{ObjectID: 9, TargetID: "blocked", Cells: []enginescan.Point{{X: 5}}},
			{ObjectID: 3, TargetID: "near", Cells: []enginescan.Point{{X: 1}}},
		},
		6,
		func(from, to enginescan.Point) (uint8, error) { return 0, nil },
	)
	if err != nil || !reflect.DeepEqual(ordered, []string{"near"}) {
		t.Fatalf("ordered=%v err=%v", ordered, err)
	}

	battle, err := NewBattle([]Fighter{
		{ID: "caster", Side: SideParty, HitPoints: 10, MaxHitPoints: 10},
		{ID: "near", Side: SideEnemy, HitPoints: 4, MaxHitPoints: 4, HitDice: 1},
		{ID: "blocked", Side: SideEnemy, HitPoints: 4, MaxHitPoints: 4, HitDice: 1},
	}, 436)
	if err != nil {
		t.Fatal(err)
	}
	result, err := battle.CastSleepOrdered("caster", ordered, 5)
	if err != nil || len(result.Impacts) != 1 || result.Impacts[0].TargetID != "near" {
		t.Fatalf("Sleep result=%+v err=%v", result, err)
	}
	blocked, _ := battle.Fighter("blocked")
	for _, effect := range blocked.MonsterAffects {
		if effect.Kind == 0x35 {
			t.Fatal("terrain-blocked target received Sleep")
		}
	}
}

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

func TestFighterVisibilityProjectsEffects18_19And47(t *testing.T) {
	ordinary := Fighter{ID: "ordinary"}
	detector := Fighter{ID: "detector", MonsterAffects: []MonsterAffect{{Kind: 0x18, Innate: true}}}
	inactiveDetector := Fighter{ID: "inactive", MonsterAffects: []MonsterAffect{{Kind: 0x18}}}
	effect19 := Fighter{ID: "effect-19", MonsterAffects: []MonsterAffect{{Kind: 0x19, Active: true}}}
	effect47 := Fighter{ID: "effect-47", MonsterAffects: []MonsterAffect{{Kind: 0x47, Active: true}}}
	inactive := Fighter{ID: "inactive-hidden", MonsterAffects: []MonsterAffect{{Kind: 0x19}}}

	if effect19.VisibleTo(ordinary) || !effect19.VisibleTo(detector) || effect19.VisibleTo(inactiveDetector) {
		t.Fatal("effect 19 visibility did not follow operational effect 18")
	}
	if effect47.VisibleTo(ordinary) || effect47.VisibleTo(detector) {
		t.Fatal("effect 47 must remain hidden even from effect 18")
	}
	if !inactive.VisibleTo(ordinary) {
		t.Fatal("inactive non-innate visibility effect became operational")
	}
}

func TestFighterVisibilityProjectsBlinkAndAnimalInvisibility(t *testing.T) {
	nonAnimal := Fighter{ID: "humanoid"}
	animal := Fighter{ID: "animal", MonsterType: MonsterTypeAnimal}
	detectingAnimal := Fighter{ID: "detecting-animal", MonsterType: MonsterTypeAnimal,
		MonsterAffects: []MonsterAffect{{Kind: 0x18, Innate: true}}}
	blink := Fighter{ID: "blink", MonsterAffects: []MonsterAffect{{Kind: 0x25, Active: true}}}
	animalHidden := Fighter{ID: "animal-hidden", MonsterAffects: []MonsterAffect{{Kind: 0x45, Active: true}}}

	if blink.VisibleTo(nonAnimal) {
		t.Fatal("zero-delay blink target remained visible")
	}
	blink.CombatAction.Delay = 1
	if !blink.VisibleTo(nonAnimal) {
		t.Fatal("nonzero-delay blink target remained hidden")
	}
	if !animalHidden.VisibleTo(nonAnimal) || animalHidden.VisibleTo(animal) || !animalHidden.VisibleTo(detectingAnimal) {
		t.Fatal("effect 45 did not follow animal type and effect 18")
	}
}
