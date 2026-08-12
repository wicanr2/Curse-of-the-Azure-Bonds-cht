package main

import (
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

func TestParseGEOPathRequest(t *testing.T) {
	start, target, err := parseGEOPathRequest("6,12:13,14")
	if err != nil {
		t.Fatal(err)
	}
	if start.X != 6 || start.Y != 12 || target.X != 13 || target.Y != 14 {
		t.Fatalf("path request start=%+v target=%+v", start, target)
	}
	for _, invalid := range []string{"", "6,12", "-1,0:1,1", "0,0:16,1", "6,12:13,14junk"} {
		if _, _, err := parseGEOPathRequest(invalid); err == nil {
			t.Fatalf("parseGEOPathRequest(%q) succeeded, want error", invalid)
		}
	}
}

func TestShortestGEOPathUsesWrappedDungeonMovement(t *testing.T) {
	grid := geo.Grid{}
	path, found := shortestGEOPath(
		grid,
		geoPathStep{X: 0, Y: 0},
		geoPathStep{X: geo.Width - 1, Y: 0},
	)
	if !found {
		t.Fatal("wrapped western neighbour is unreachable")
	}
	want := []geoPathStep{
		{X: 0, Y: 0, Direction: -1},
		{X: geo.Width - 1, Y: 0, Direction: 6},
	}
	if !reflect.DeepEqual(path, want) {
		t.Fatalf("path=%+v, want %+v", path, want)
	}
}

func TestShortestGEOPathWithDoorsDoesNotInventDetailZeroPassage(t *testing.T) {
	grid := geo.Grid{}
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			for direction := range grid.Cells[y][x].WallDirections {
				grid.Cells[y][x].WallDirections[direction] = 9
			}
		}
	}
	if _, found := shortestGEOPathWithDoors(grid, geoPathStep{X: 0, Y: 0}, geoPathStep{X: 1, Y: 0}, true); found {
		t.Fatal("detail-zero wall became passable")
	}
	grid.Cells[0][0].DetailDirections[1] = 2
	grid.Cells[0][1].DetailDirections[3] = 2
	path, found := shortestGEOPathWithDoors(grid, geoPathStep{X: 0, Y: 0}, geoPathStep{X: 1, Y: 0}, true)
	if !found || len(path) != 2 || !path[1].Door {
		t.Fatalf("ordinary locked-door path=%+v found=%v", path, found)
	}
}
