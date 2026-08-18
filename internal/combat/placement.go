package combat

// TilePoint is a data-neutral combat-map coordinate. The reference engine
// stores combatants on a tile grid and derives screen positions from it.
type TilePoint struct {
	X int
	Y int
}

// EncounterLayout contains the reference inputs used to separate the two
// combat teams. Distance comes from SETUP MONSTER's max-distance value; the
// map direction is kept separate because it belongs to the Area/CombatMap
// state, not to an individual fighter.
type EncounterLayout struct {
	Distance     int
	MapDirection uint8
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

// EncounterTeamStart returns the reference team origin and facing group.
// The reference engine keeps the party at (0, 0), while the enemy team is
// offset by encounter distance along MapDirectionDelta. Facing is stored in
// four cardinal groups (the reference divides the eight-way direction by 2).
// This intentionally does not choose an occupied candidate cell; that still
// requires the decoded combat-map occupancy table.
func EncounterTeamStart(layout EncounterLayout, side Side) (TilePoint, uint8, bool) {
	if layout.Distance < 0 || layout.MapDirection >= 8 {
		return TilePoint{}, 0, false
	}
	if side == SideParty {
		return TilePoint{}, layout.MapDirection / 2, true
	}
	if side != SideEnemy {
		return TilePoint{}, 0, false
	}
	delta, _ := DirectionDelta(layout.MapDirection)
	return TilePoint{
		X: layout.Distance * delta.X,
		Y: layout.Distance * delta.Y,
	}, ((layout.MapDirection + 4) % 8) / 2, true
}

// IconDirectionForTeam reproduces SetupCombatActions' HalfDirToIso mapping.
// It returns the full eight-way icon direction used by CombatIcon.GetIcon;
// enemy direction is the party direction rotated 180 degrees.
func IconDirectionForTeam(mapDirection uint8, side Side) (uint8, bool) {
	if mapDirection >= 8 || (side != SideParty && side != SideEnemy) {
		return 0, false
	}
	direction := [...]uint8{7, 2, 3, 6}[mapDirection/2]
	if side == SideEnemy {
		direction = (direction + 4) % 8
	}
	return direction, true
}

// FormationTile is the temporary deterministic placement used until ECL /
// combat-map records provide real positions. It keeps party and enemy rows
// on distinct halves of the current combat viewport.
//
// ★ 敵方縱帶是 4..6，不是 7..9。視野只有 7 欄（0..6），放在 7..9 表示**兩隊
// 永遠不可能同時入鏡**：相機必然被迫開啟，然後對準哪一隊，另一隊就被推出
// 畫面。這與本函式自己寫的「distinct halves of the current combat viewport」
// 相矛盾，Burial Glen 的戰鬥截圖也因此只剩一個人、四隻蜘蛛全在畫面外。
//
// 列從第 2 列起算，讓兩條縱帶落在七列視野的**中段**而不是貼著上緣。人數多到
// 排不下時往下延伸，相機自然接手。
func FormationTile(side Side, ordinal int) TilePoint {
	if ordinal < 0 {
		ordinal = 0
	}
	column, row := ordinal%3, formationFirstRow+ordinal/3
	if side == SideEnemy {
		return TilePoint{X: 4 + column, Y: row}
	}
	return TilePoint{X: column, Y: row}
}

// formationFirstRow 是編制帶的起始列。
const formationFirstRow = 2

// ReferencePlacement applies the coordinate formula used by the reference
// try_place_combatant routine. candidateRow/Column are the free cells in the
// current placement window; groupRow and teamX/Y come from the team layout.
func ReferencePlacement(teamX, teamY, groupRow, candidateRow, candidateColumn int) TilePoint {
	return TilePoint{
		X: candidateColumn + teamX*6 + groupRow*5 + 22,
		Y: candidateRow + teamY*5 + 10,
	}
}
