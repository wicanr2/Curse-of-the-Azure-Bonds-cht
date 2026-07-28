package combat

import "testing"

func TestFootprintForOriginalSizeCodes(t *testing.T) {
	for _, test := range []struct {
		size        uint8
		width, high int
	}{
		{size: 1, width: 1, high: 1},
		{size: 2, width: 1, high: 2},
		{size: 3, width: 2, high: 1},
		{size: 4, width: 2, high: 2},
	} {
		got := FootprintForSize(test.size)
		if got.Width != test.width || got.Height != test.high {
			t.Fatalf("size %d footprint=%+v, want %dx%d", test.size, got, test.width, test.high)
		}
	}
}

func TestLargeFootprintOverlapAndAdjacency(t *testing.T) {
	dragon := Fighter{CombatX: 2, CombatY: 2, CombatSize: 4}
	hero := Fighter{CombatSize: 1}
	if !FootprintsOverlapAt(hero, 3, 3, dragon) {
		t.Fatal("one-cell fighter did not overlap dragon's lower-right cell")
	}
	if FootprintsOverlapAt(hero, 4, 3, dragon) {
		t.Fatal("fighter outside dragon footprint reported overlap")
	}
	hero.CombatX, hero.CombatY = 4, 3
	if !footprintAdjacent(hero, dragon) {
		t.Fatal("fighter beside dragon footprint is not adjacent")
	}
}

func TestRestoreCombatantRejectsAnyOccupiedLargeMonsterCell(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "dragon", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0, CombatSize: 4, HasCombatPosition: true, CombatX: 2, CombatY: 2},
		{ID: "hero", Side: SideParty, HitPoints: 5, MaxHitPoints: 10, ArmorClass: 10},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.RestoreCombatant("hero", TilePoint{X: 3, Y: 3}); err == nil {
		t.Fatal("restored hero inside dragon 2x2 footprint")
	}
}

func TestLargeFootprintMovementChecksEveryTerrainCell(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "dragon", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0, CombatSize: 4, HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "enemy", Side: SideEnemy, HitPoints: 5, MaxHitPoints: 5, ArmorClass: 10, HasCombatPosition: true, CombatX: 6, CombatY: 6},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	terrain := func(x, y int) (int, bool) {
		return 1, x != 3 || y != 2
	}
	if _, err := battle.MoveWithTerrainAndFreeAttacks("dragon", 1, 0, 5, terrain); err == nil {
		t.Fatal("2x2 dragon moved while its lower-right destination cell was impassable")
	}
	var fighter Fighter
	for _, candidate := range battle.Fighters() {
		if candidate.ID == "dragon" {
			fighter = candidate
		}
	}
	if fighter.CombatX != 1 || fighter.CombatY != 1 {
		t.Fatalf("failed terrain move mutated position to (%d,%d)", fighter.CombatX, fighter.CombatY)
	}
}

func TestTerrainMovementUsesHighestFootprintCostAndBudget(t *testing.T) {
	newBattle := func(t *testing.T) *Battle {
		t.Helper()
		battle, err := NewBattle([]Fighter{
			{ID: "wide", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 0, CombatSize: 3, HasCombatPosition: true, CombatX: 1, CombatY: 1},
			{ID: "enemy", Side: SideEnemy, HitPoints: 5, MaxHitPoints: 5, ArmorClass: 10, HasCombatPosition: true, CombatX: 6, CombatY: 6},
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		return battle
	}
	terrain := func(x, y int) (int, bool) {
		if x == 3 && y == 1 {
			return 2, true
		}
		return 1, true
	}
	if _, err := newBattle(t).MoveWithTerrainAndFreeAttacks("wide", 1, 0, 1, terrain); err == nil {
		t.Fatal("wide fighter entered cost-2 terrain with one point remaining")
	}
	result, err := newBattle(t).MoveWithTerrainAndFreeAttacks("wide", 1, 0, 2, terrain)
	if err != nil {
		t.Fatal(err)
	}
	if result.MovementCost != 2 || result.Fighter.CombatX != 2 {
		t.Fatalf("terrain move=%+v, want cost 2 at x=2", result)
	}
}
