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

func TestDungeonDecorationUsesTerrainBitAndSeed(t *testing.T) {
	grid := geo.Grid{}
	cell := &grid.Cells[8][8]
	cell.WallDirections = [4]uint8{1, 1, 1, 1}
	cell.Terrain = 0x40
	plain := GenerateDungeonSeeded(geo.Grid{}, 8, 8, 1)
	found := false
	for seed := int64(1); seed <= 100; seed++ {
		decorated := GenerateDungeonSeeded(grid, 8, 8, seed)
		if decorated == plain {
			continue
		}
		for y := 0; y < WildernessHeight; y++ {
			for x := 0; x < WildernessWidth; x++ {
				if decorated.Tiles[y][x] == 0x1A || decorated.Tiles[y][x] == 0x1B {
					found = true
				}
			}
		}
		if found {
			if decorated != GenerateDungeonSeeded(grid, 8, 8, seed) {
				t.Fatalf("seed %d decoration is not deterministic", seed)
			}
			break
		}
	}
	if !found {
		t.Fatal("terrain bit and open wall flags never produced table/chair decoration")
	}
}
