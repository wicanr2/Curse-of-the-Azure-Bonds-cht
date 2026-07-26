package mapdata

import "testing"

func TestWildernessFloorIsDeterministicAndBounded(t *testing.T) {
	a := GenerateWilderness(0x11, 42)
	b := GenerateWilderness(0x11, 42)
	if a != b {
		t.Fatal("same city flags and seed must reproduce the wilderness floor")
	}
	if a.Tiles[0][0] >= uint8(len(BackgroundTiles)) {
		t.Fatalf("tile entry out of background table: %d", a.Tiles[0][0])
	}
	if _, ok := a.Entry(-1, 0); ok {
		t.Fatal("negative coordinate should be outside wilderness")
	}
	if _, ok := a.Entry(WildernessWidth, WildernessHeight); ok {
		t.Fatal("past-end coordinate should be outside wilderness")
	}
}

func TestWildernessFloorUsesBackgroundEntryThenTileIndex(t *testing.T) {
	floor := GenerateWilderness(0, 7)
	entry, ok := floor.Entry(0, 0)
	if !ok {
		t.Fatal("origin should have a background entry")
	}
	if floor.Tiles[0][0] != 23 || entry.TileIndex != 0x16 {
		t.Fatalf("origin entry=%d data=%+v, want reference default entry 23/tile 0x16", floor.Tiles[0][0], entry)
	}
	if !floor.CanEnter(0, 0) {
		t.Fatal("reference default wilderness floor should be enterable")
	}
}

func TestCityInfoTable(t *testing.T) {
	if flags, ok := CityInfo(0); !ok || flags != 0 {
		t.Fatalf("city 0 flags=%#x, want 0", flags)
	}
	if _, ok := CityInfo(-1); ok {
		t.Fatal("negative city index should fail")
	}
}
