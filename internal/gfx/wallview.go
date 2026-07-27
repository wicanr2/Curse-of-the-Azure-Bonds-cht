package gfx

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// WallLayoutCall describes one reference draw_3D_8x8_titles invocation before
// it is expanded into WallStamps. Coordinates are kept in map space so a
// renderer can apply its own wrapping, clipping, or camera policy.
type WallLayoutCall struct {
	Depth     int
	Direction uint8
	MapX      int
	MapY      int
	WallType  uint8
	Layout    int
	RowStart  int
	ColStart  int
}

// WallView is the ordered Far/Mid/Near wall traversal for one party-facing
// direction. Calls are ordered in the same draw order as the reference code.
type WallView struct {
	PartyDirection uint8
	Calls          []WallLayoutCall
}

var wallViewDelta = [...]struct{ x, y int }{
	{0, -1}, {1, -1}, {1, 0}, {1, 1},
	{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
}

func wallViewDirection(direction uint8) (struct{ x, y int }, bool) {
	if direction >= uint8(len(wallViewDelta)) {
		return struct{ x, y int }{}, false
	}
	return wallViewDelta[direction], true
}

func wallViewWall(grid geo.Grid, direction uint8, x, y int) uint8 {
	value, ok := grid.Wall(x, y, int(direction))
	if !ok {
		return 0
	}
	return value
}

func addWallCall(calls *[]WallLayoutCall, grid geo.Grid, depth int, direction uint8, x, y, layout, row, column int) {
	wallType := wallViewWall(grid, direction, x, y)
	if wallType == 0 {
		return
	}
	*calls = append(*calls, WallLayoutCall{
		Depth: depth, Direction: direction, MapX: x, MapY: y,
		WallType: wallType, Layout: layout, RowStart: row, ColStart: column,
	})
}

func addWallCallValue(calls *[]WallLayoutCall, depth int, direction uint8, x, y int, wallType uint8, layout, row, column int) {
	if wallType == 0 {
		return
	}
	*calls = append(*calls, WallLayoutCall{
		Depth: depth, Direction: direction, MapX: x, MapY: y,
		WallType: wallType, Layout: layout, RowStart: row, ColStart: column,
	})
}

func drawWallFar(grid geo.Grid, calls *[]WallLayoutCall, partyDirection, left, right uint8, x, y int) {
	startX, startY := x, y
	leftDelta, _ := wallViewDirection(left)
	var previous uint8
	column := 0
	for index := 0; index < 4; index++ {
		current := wallViewWall(grid, partyDirection, x, y)
		if current != 0 {
			if previous != 0 {
				addWallCallValue(calls, 2, partyDirection, x, y, previous, 9, 4, 5+column+1)
			}
			previous = current
			addWallCall(calls, grid, 2, partyDirection, x, y, 0, 4, 5+column)
		} else {
			if previous != 0 && wallViewWall(grid, left, x-leftDelta.x, y-leftDelta.y) != 0 {
				addWallCallValue(calls, 2, partyDirection, x, y, previous, 9, 4, 5+column+1)
			}
			previous = 0
		}
		column -= 2
		x += leftDelta.x
		y += leftDelta.y
	}

	rightDelta, _ := wallViewDirection(right)
	previous = 0
	column = 0
	x, y = startX, startY
	for index := 0; index < 4; index++ {
		current := wallViewWall(grid, partyDirection, x, y)
		if current != 0 {
			if previous != 0 {
				addWallCallValue(calls, 2, partyDirection, x, y, previous, 9, 4, 5+column-1)
			}
			previous = current
			addWallCall(calls, grid, 2, partyDirection, x, y, 0, 4, 5+column)
		} else {
			if previous != 0 && wallViewWall(grid, right, x-rightDelta.x, y-rightDelta.y) != 0 {
				addWallCallValue(calls, 2, partyDirection, x, y, previous, 9, 4, 5+column-1)
			}
			previous = 0
		}
		column += 2
		x += rightDelta.x
		y += rightDelta.y
	}

	x, y = startX, startY
	column = 0
	for index := 0; index < 3; index++ {
		sideColumn := 4 + column
		if index > 0 {
			sideColumn--
		}
		addWallCall(calls, grid, 2, left, x, y, 1, 3, sideColumn)
		column -= 2
		x += leftDelta.x
		y += leftDelta.y
	}

	x, y = startX, startY
	column = 0
	for index := 0; index < 3; index++ {
		sideColumn := 6 + column
		if index > 0 {
			sideColumn++
		}
		addWallCall(calls, grid, 2, right, x, y, 2, 3, sideColumn)
		column += 2
		x += rightDelta.x
		y += rightDelta.y
	}
}

func drawWallMid(grid geo.Grid, calls *[]WallLayoutCall, partyDirection, left, right uint8, x, y int) {
	baseX, baseY := x, y
	leftDelta, _ := wallViewDirection(left)
	rightDelta, _ := wallViewDirection(right)
	x, y = baseX+2*leftDelta.x, baseY+2*leftDelta.y
	column := -6
	for index := 0; index < 3; index++ {
		addWallCall(calls, grid, 1, partyDirection, x, y, 3, 3, 4+column)
		addWallCall(calls, grid, 1, left, x, y, 4, 1, 2+column)
		column += 3
		x += rightDelta.x
		y += rightDelta.y
	}

	x, y = baseX+2*rightDelta.x, baseY+2*rightDelta.y
	column = 6
	for index := 0; index < 3; index++ {
		addWallCall(calls, grid, 1, partyDirection, x, y, 3, 3, 4+column)
		addWallCall(calls, grid, 1, right, x, y, 5, 1, 7+column)
		column -= 3
		x += leftDelta.x
		y += leftDelta.y
	}
}

func drawWallNear(grid geo.Grid, calls *[]WallLayoutCall, partyDirection, left, right uint8, x, y int) {
	baseX, baseY := x, y
	leftDelta, _ := wallViewDirection(left)
	rightDelta, _ := wallViewDirection(right)
	x, y = baseX+leftDelta.x, baseY+leftDelta.y
	column := -7
	for index := 0; index < 2; index++ {
		addWallCall(calls, grid, 0, partyDirection, x, y, 6, 1, 2+column)
		addWallCall(calls, grid, 0, left, x, y, 7, 0, column)
		column += 7
		x += rightDelta.x
		y += rightDelta.y
	}

	x, y = baseX+rightDelta.x, baseY+rightDelta.y
	column = 7
	for index := 0; index < 2; index++ {
		addWallCall(calls, grid, 0, partyDirection, x, y, 6, 1, 2+column)
		addWallCall(calls, grid, 0, right, x, y, 8, 0, 9+column)
		column -= 7
		x += leftDelta.x
		y += leftDelta.y
	}
}

// TraverseWallView reproduces the map-coordinate traversal of reference
// Draw3dWorld. It intentionally does not wrap coordinates; the original
// engine's wrap decision depends on ECL context and belongs to an outer Area
// adapter.
func TraverseWallView(grid geo.Grid, partyDirection uint8, partyX, partyY int) (WallView, error) {
	if partyDirection >= uint8(len(wallViewDelta)) {
		return WallView{}, fmt.Errorf("party direction %d is outside 0..7", partyDirection)
	}
	left := (partyDirection + 6) % 8
	behind := (partyDirection + 4) % 8
	right := (partyDirection + 2) % 8
	partyDelta, _ := wallViewDirection(partyDirection)
	behindDelta, _ := wallViewDirection(behind)
	view := WallView{PartyDirection: partyDirection, Calls: make([]WallLayoutCall, 0)}
	x := partyX + 2*partyDelta.x
	y := partyY + 2*partyDelta.y
	for depth := 2; depth >= 0; depth-- {
		switch depth {
		case 2:
			drawWallFar(grid, &view.Calls, partyDirection, left, right, x, y)
		case 1:
			drawWallMid(grid, &view.Calls, partyDirection, left, right, x, y)
		case 0:
			drawWallNear(grid, &view.Calls, partyDirection, left, right, x, y)
		}
		x += behindDelta.x
		y += behindDelta.y
	}
	return view, nil
}
