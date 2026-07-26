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

func TestFormationTileSeparatesPartyAndEnemyRows(t *testing.T) {
	if got := FormationTile(SideParty, 2); got != (TilePoint{X: 2, Y: 0}) {
		t.Fatalf("party tile=%#v", got)
	}
	if got := FormationTile(SideEnemy, 2); got != (TilePoint{X: 9, Y: 0}) {
		t.Fatalf("enemy tile=%#v", got)
	}
}
