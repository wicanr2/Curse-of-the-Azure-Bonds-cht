package mapdata

import "testing"

func TestReferenceBackgroundTileTable(t *testing.T) {
	if len(BackgroundTiles) != 74 {
		t.Fatalf("table length=%d, want 74 recovered entries", len(BackgroundTiles))
	}
	first, ok := Lookup(0)
	if !ok || first.MoveCost != 1 || first.TileIndex != 0 || first.Passable() != true {
		t.Fatalf("tile 0=%+v, want passable floor entry", first)
	}
	blocked, ok := Lookup(1)
	if !ok || blocked.MoveCost != 0xFF || blocked.TileIndex != 0 || blocked.Passable() {
		t.Fatalf("tile 1=%+v, want blocked entry", blocked)
	}
	reserved, ok := Lookup(73)
	if !ok || reserved.MoveCost != 0 || reserved.TileIndex != 1 {
		t.Fatalf("tail entry=%+v, want preserved reserved metadata", reserved)
	}
}

func TestBackgroundTileLookupBounds(t *testing.T) {
	if _, ok := Lookup(-1); ok {
		t.Fatal("negative index should fail")
	}
	if _, ok := Lookup(len(BackgroundTiles)); ok {
		t.Fatal("past-end index should fail")
	}
}
