package combat

import "testing"

func TestDirectionDeltaUsesReferenceEightWayOrder(t *testing.T) {
	want := [...]TilePoint{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}
	for direction, expected := range want {
		got, ok := DirectionDelta(uint8(direction))
		if !ok || got != expected {
			t.Fatalf("direction %d=%#v ok=%t, want %#v", direction, got, ok, expected)
		}
	}
	if _, ok := DirectionDelta(8); ok {
		t.Fatal("direction 8 should be rejected")
	}
}

func TestEncounterTeamStartUsesDistanceAndOppositeFacing(t *testing.T) {
	layout := EncounterLayout{Distance: 3, MapDirection: 3}
	party, partyFacing, ok := EncounterTeamStart(layout, SideParty)
	if !ok || party != (TilePoint{}) || partyFacing != 1 {
		t.Fatalf("party start=%#v facing=%d ok=%t", party, partyFacing, ok)
	}
	enemy, enemyFacing, ok := EncounterTeamStart(layout, SideEnemy)
	if !ok || enemy != (TilePoint{X: 3, Y: 3}) || enemyFacing != 3 {
		t.Fatalf("enemy start=%#v facing=%d ok=%t", enemy, enemyFacing, ok)
	}
}

func TestEncounterTeamStartRejectsInvalidInputs(t *testing.T) {
	if _, _, ok := EncounterTeamStart(EncounterLayout{Distance: -1}, SideEnemy); ok {
		t.Fatal("negative distance should be rejected")
	}
	if _, _, ok := EncounterTeamStart(EncounterLayout{MapDirection: 8}, SideEnemy); ok {
		t.Fatal("direction 8 should be rejected")
	}
	if _, _, ok := EncounterTeamStart(EncounterLayout{}, Side(2)); ok {
		t.Fatal("unknown side should be rejected")
	}
}

func TestIconDirectionForTeamUsesReferenceHalfDirToIso(t *testing.T) {
	party, ok := IconDirectionForTeam(0, SideParty)
	if !ok || party != 7 {
		t.Fatalf("party direction=%d ok=%t, want 7", party, ok)
	}
	enemy, ok := IconDirectionForTeam(0, SideEnemy)
	if !ok || enemy != 3 {
		t.Fatalf("enemy direction=%d ok=%t, want 3", enemy, ok)
	}
	if _, ok := IconDirectionForTeam(8, SideParty); ok {
		t.Fatal("direction 8 should be rejected")
	}
}

func TestFormationTileSeparatesPartyAndEnemyRows(t *testing.T) {
	if got := FormationTile(SideParty, 2); got != (TilePoint{X: 2, Y: 2}) {
		t.Fatalf("party tile=%#v", got)
	}
	if got := FormationTile(SideEnemy, 2); got != (TilePoint{X: 6, Y: 2}) {
		t.Fatalf("enemy tile=%#v", got)
	}
}

func TestReferencePlacementMatchesTryPlaceCombatantFormula(t *testing.T) {
	got := ReferencePlacement(2, 3, 1, 4, 5)
	if got != (TilePoint{X: 5 + 2*6 + 1*5 + 22, Y: 4 + 3*5 + 10}) {
		t.Fatalf("reference placement=%#v", got)
	}
}
