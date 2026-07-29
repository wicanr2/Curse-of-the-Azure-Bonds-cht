package combat

import "testing"

func TestCastStinkingCloudUsesTargetAnchoredTwoByTwoAndDeduplicatesFootprint(t *testing.T) {
	battle := newCloudBattle(t, 2)
	result, err := battle.CastStinkingCloud("mage", TilePoint{X: 4, Y: 3}, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []PersistentAreaCell{{4, 3}, {5, 3}, {5, 4}, {4, 4}}
	if len(result.Area.Cells) != len(want) {
		t.Fatalf("cells=%#v", result.Area.Cells)
	}
	for index := range want {
		if result.Area.Cells[index] != want[index] {
			t.Fatalf("cell[%d]=%#v want %#v", index, result.Area.Cells[index], want[index])
		}
	}
	if len(result.Impacts) != 1 || result.Impacts[0].TargetID != "ogre" {
		t.Fatalf("large footprint should be affected once: %#v", result.Impacts)
	}
}

func TestCastStinkingCloudFiltersWallsAndRejectsEmptyArea(t *testing.T) {
	battle := newCloudBattle(t, 2)
	result, err := battle.CastStinkingCloud("mage", TilePoint{X: 4, Y: 3}, 3, func(x, y int) bool {
		return x == 4 && y == 3
	})
	if err != nil || len(result.Area.Cells) != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	empty := newCloudBattle(t, 2)
	if _, err = empty.CastStinkingCloud("mage", TilePoint{}, 3, func(int, int) bool { return false }); err == nil {
		t.Fatal("all-wall cloud should fail")
	}
	if len(empty.PersistentAreas()) != 0 {
		t.Fatal("failed cast must not leave an area")
	}
}

func TestStinkingCloudAppliesSavedCoughOrFailedHelpless(t *testing.T) {
	saved := newCloudBattle(t, 2)
	target := saved.fighters["ogre"]
	target.SavingThrowBonus = 100
	saved.fighters["ogre"] = target
	result, err := saved.CastStinkingCloud("mage", TilePoint{X: 4, Y: 3}, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Impacts) != 1 || !result.Impacts[0].Saved || result.Impacts[0].CoughingTurns != 1 {
		t.Fatalf("saved impact=%#v", result.Impacts)
	}
	fighter, _ := saved.Fighter("ogre")
	if !fighter.CloudIncapacitated() {
		t.Fatal("saved target should cough")
	}
	if fighter, err = saved.ConsumeCloudIncapacitation("ogre"); err != nil || fighter.CloudIncapacitated() {
		t.Fatalf("consumed fighter=%#v err=%v", fighter, err)
	}

	failed := newCloudBattle(t, 1)
	target = failed.fighters["ogre"]
	target.SavingThrowBonus = -100
	failed.fighters["ogre"] = target
	result, err = failed.CastStinkingCloud("mage", TilePoint{X: 4, Y: 3}, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Impacts) != 1 || result.Impacts[0].Saved || result.Impacts[0].HelplessTurns < 2 || result.Impacts[0].HelplessTurns > 5 {
		t.Fatalf("failed impact=%#v", result.Impacts)
	}
}

func TestPersistentCloudOverlapAndExpiry(t *testing.T) {
	battle := newCloudBattle(t, 2)
	first, err := battle.CastStinkingCloud("mage", TilePoint{X: 3, Y: 2}, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := battle.CastStinkingCloud("mage", TilePoint{X: 4, Y: 2}, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Area.ID == second.Area.ID {
		t.Fatal("area IDs must be unique")
	}
	if _, ok := battle.PersistentAreaAt(4, 2); !ok {
		t.Fatal("overlap should be covered")
	}
	if _, err := battle.StartRound(); err != nil {
		t.Fatal(err)
	}
	if len(battle.PersistentAreas()) != 1 {
		t.Fatalf("areas after expiry=%#v", battle.PersistentAreas())
	}
	if _, ok := battle.PersistentAreaAt(4, 2); !ok {
		t.Fatal("second cloud should preserve overlap")
	}
}

func newCloudBattle(t *testing.T, seed int64) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "mage", Side: SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 5,
			HasCombatPosition: true, CombatX: 1, CombatY: 1, CombatSize: 1,
			SavingThrows: []uint8{10, 10, 10, 10, 10}},
		{ID: "ogre", Side: SideEnemy, HitPoints: 30, MaxHitPoints: 30, ArmorClass: 5,
			HasCombatPosition: true, CombatX: 4, CombatY: 3, CombatSize: 2,
			SavingThrows: []uint8{10, 10, 10, 10, 10}},
	}, seed)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}
