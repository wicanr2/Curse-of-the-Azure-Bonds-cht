package combat

import "testing"

func TestNewCombatCameraCentersActiveTile(t *testing.T) {
	camera := NewCombatCamera(TilePoint{X: 23, Y: 14}, TilePoint{X: 4, Y: 2}, true)
	if got := camera.Apply(TilePoint{X: 23, Y: 14}); got != (TilePoint{X: 4, Y: 2}) {
		t.Fatalf("active tile=%+v", got)
	}
	if got := camera.Apply(TilePoint{X: 24, Y: 13}); got != (TilePoint{X: 5, Y: 1}) {
		t.Fatalf("translated tile=%+v", got)
	}
}

func TestNewCombatCameraWithoutActivePositionPreservesCoordinates(t *testing.T) {
	camera := NewCombatCamera(TilePoint{X: 23, Y: 14}, TilePoint{X: 4, Y: 2}, false)
	if got := camera.Apply(TilePoint{X: 7, Y: 3}); got != (TilePoint{X: 7, Y: 3}) {
		t.Fatalf("fallback tile=%+v", got)
	}
}
