package mapdata

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

func TestDungeonFloorBuildsReferenceWindow(t *testing.T) {
	grid := geo.Grid{}
	floor := GenerateDungeon(grid, 8, 8)
	seen := 0
	for y := 0; y < WildernessHeight; y++ {
		for x := 0; x < WildernessWidth; x++ {
			if floor.Tiles[y][x] != 0 {
				seen++
			}
			if floor.Tiles[y][x] >= uint8(len(BackgroundTiles)) {
				t.Fatalf("invalid background entry at (%d,%d): %d", x, y, floor.Tiles[y][x])
			}
		}
	}
	if seen == 0 {
		t.Fatal("dungeon builder did not write its 13x5 window")
	}
	if floor.MapX != 8 || floor.MapY != 8 {
		t.Fatalf("map center=(%d,%d)", floor.MapX, floor.MapY)
	}
}

func TestDungeonWallsAffectBackgroundComposition(t *testing.T) {
	open := geo.Grid{}
	walled := geo.Grid{}
	walled.Cells[8][8].WallDirections[0] = 1
	walled.Cells[8][8].DetailDirections[0] = 1
	openFloor := GenerateDungeon(open, 8, 8)
	walledFloor := GenerateDungeon(walled, 8, 8)
	if openFloor.Tiles == walledFloor.Tiles {
		t.Fatal("a GEO wall should alter dungeon background composition")
	}
}
