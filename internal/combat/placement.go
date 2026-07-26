package combat

// TilePoint is a data-neutral combat-map coordinate. The reference engine
// stores combatants on a tile grid and derives screen positions from it.
type TilePoint struct {
	X int
	Y int
}

// DirectionDelta returns the reference MapDirectionDelta order:
// north, northeast, east, southeast, south, southwest, west, northwest.
func DirectionDelta(direction uint8) (TilePoint, bool) {
	deltas := [...]TilePoint{
		{X: 0, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 0}, {X: 1, Y: 1},
		{X: 0, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: 0}, {X: -1, Y: -1},
	}
	if direction >= uint8(len(deltas)) {
		return TilePoint{}, false
	}
	return deltas[direction], true
}

// FormationTile is the temporary deterministic placement used until ECL /
// combat-map records provide real positions. It keeps party and enemy rows
// on distinct halves of the current combat viewport.
func FormationTile(side Side, ordinal int) TilePoint {
	if ordinal < 0 {
		ordinal = 0
	}
	column, row := ordinal%3, ordinal/3
	if side == SideEnemy {
		return TilePoint{X: 7 + column, Y: row}
	}
	return TilePoint{X: column, Y: row}
}

// ReferencePlacement applies the coordinate formula used by the reference
// try_place_combatant routine. candidateRow/Column are the free cells in the
// current placement window; groupRow and teamX/Y come from the team layout.
func ReferencePlacement(teamX, teamY, groupRow, candidateRow, candidateColumn int) TilePoint {
	return TilePoint{
		X: candidateColumn + teamX*6 + groupRow*5 + 22,
		Y: candidateRow + teamY*5 + 10,
	}
}
