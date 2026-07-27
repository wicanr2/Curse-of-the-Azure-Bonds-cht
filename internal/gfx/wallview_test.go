package gfx

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

func TestTraverseWallViewUsesReferenceFarCenter(t *testing.T) {
	var grid geo.Grid
	grid.Cells[6][8].WallDirections[0] = 1
	view, err := TraverseWallView(grid, 0, 8, 8)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, call := range view.Calls {
		if call.Depth == 2 && call.Layout == 0 && call.MapX == 8 && call.MapY == 6 && call.RowStart == 4 && call.ColStart == 5 && call.WallType == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("far center call not found in %+v", view.Calls)
	}
}

func TestTraverseWallViewReturnsOrderedDepths(t *testing.T) {
	var grid geo.Grid
	for y := 5; y <= 7; y++ {
		grid.Cells[y][8].WallDirections[0] = 1
	}
	view, err := TraverseWallView(grid, 0, 8, 8)
	if err != nil {
		t.Fatal(err)
	}
	last := 2
	for _, call := range view.Calls {
		if call.Depth > last {
			t.Fatalf("depth order regressed from %d to %d", last, call.Depth)
		}
		last = call.Depth
	}
}

func TestTraverseWallViewUsesFarMidNearCoordinates(t *testing.T) {
	var grid geo.Grid
	for _, point := range [][2]int{{8, 6}, {6, 7}, {10, 7}, {7, 8}, {9, 8}} {
		grid.Cells[point[1]][point[0]].WallDirections[0] = 1
	}
	view, err := TraverseWallView(grid, 0, 8, 8)
	if err != nil {
		t.Fatal(err)
	}
	wants := map[[3]int]bool{
		{2, 8, 6}:  true,
		{1, 6, 7}:  true,
		{1, 10, 7}: true,
		{0, 7, 8}:  true,
		{0, 9, 8}:  true,
	}
	for _, call := range view.Calls {
		delete(wants, [3]int{call.Depth, call.MapX, call.MapY})
	}
	if len(wants) != 0 {
		t.Fatalf("missing Far/Mid/Near coordinates: %v", wants)
	}
}

func TestTraverseWallViewRejectsInvalidDirection(t *testing.T) {
	if _, err := TraverseWallView(geo.Grid{}, 8, 0, 0); err == nil {
		t.Fatal("expected invalid direction error")
	}
}

func TestTraverseWallViewWrappedReadsAcrossMapEdge(t *testing.T) {
	var grid geo.Grid
	grid.Cells[geo.Height-2][8].WallDirections[0] = 3
	view, err := TraverseWallViewWrapped(grid, 0, 8, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, call := range view.Calls {
		if call.Depth == 2 && call.MapX == 8 && call.MapY == -2 && call.WallType == 3 {
			return
		}
	}
	t.Fatalf("wrapped far wall not found in %+v", view.Calls)
}
