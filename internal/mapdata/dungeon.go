package mapdata

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"

// DungeonFloor is the 50x25 background-entry buffer used by the reference
// combat/map renderer. A 13x5 window around MapX/MapY is built from GEO wall
// and door fields, matching SetupDungeonFloor's write coordinates.
type DungeonFloor struct {
	Tiles [WildernessHeight][WildernessWidth]uint8
	MapX  int
	MapY  int
}

// GenerateDungeon builds the four proven background-tile quadrants for the
// 13x5 dungeon window. The table/chair decoration pass is intentionally
// separate because it depends on the reference wall-x2 field and dice state.
func GenerateDungeon(grid geo.Grid, mapX, mapY int) DungeonFloor {
	floor := DungeonFloor{MapX: mapX, MapY: mapY}
	for row := -2; row <= 2; row++ {
		for column := -6; column <= 6; column++ {
			x, y := mapX+column, mapY+row
			dir0 := dungeonDirFlags(grid, 0, x, y, mapY)
			dir6 := dungeonDirFlags(grid, 6, x, y, mapY)
			dir2 := dungeonDirFlags(grid, 2, x, y, mapY)

			buildDungeonTiles1(&floor, column, row, dir6)
			buildDungeonTiles2(&floor, column, row, dir0)
			buildDungeonTiles3(&floor, grid, column, row, x, y, dir0, dir6)
			buildDungeonTiles4(&floor, grid, column, row, x, y, dir0, dir2, mapY)
		}
	}
	return floor
}

func (floor DungeonFloor) Entry(x, y int) (BackgroundTile, bool) {
	if x < 0 || x >= WildernessWidth || y < 0 || y >= WildernessHeight {
		return BackgroundTile{}, false
	}
	return Lookup(int(floor.Tiles[y][x]))
}

func (floor DungeonFloor) CanEnter(x, y int) bool {
	tile, ok := floor.Entry(x, y)
	return ok && tile.Passable()
}

func buildDungeonTiles1(floor *DungeonFloor, column, row, dir6 int) {
	for y := 2; y <= 4; y++ {
		for x := 0; x <= 5; x++ {
			setDungeonRelative(floor, column, row, x, y, 22)
		}
	}
	if dir6 == 1 {
		for value := 2; value <= 4; value++ {
			setDungeonRelative(floor, column, row, value-1, value, 4)
			setDungeonRelative(floor, column, row, value, value, 3)
			setDungeonRelative(floor, column, row, value+1, value, 13)
		}
	} else if dir6 == 3 {
		setDungeonRelative(floor, column, row, 1, 2, 8)
		setDungeonRelative(floor, column, row, 5, 4, 0)
	}
}

func buildDungeonTiles2(floor *DungeonFloor, column, row, dir0 int) {
	if dir0 == 1 {
		setDungeonRelative(floor, column, row, 3, 0, 5)
		setDungeonRelative(floor, column, row, 4, 0, 5)
		setDungeonRelative(floor, column, row, 3, 1, 10)
		setDungeonRelative(floor, column, row, 4, 1, 10)
		return
	}
	setDungeonRelative(floor, column, row, 3, 0, 22)
	setDungeonRelative(floor, column, row, 4, 0, 22)
	setDungeonRelative(floor, column, row, 3, 1, 22)
	setDungeonRelative(floor, column, row, 4, 1, 22)
}

func buildDungeonTiles3(floor *DungeonFloor, grid geo.Grid, column, row, x, y, dir0, dir6 int) {
	clear := dungeonDirFlags(grid, 6, x, y-1, floor.MapY) == 0 && dungeonDirFlags(grid, 0, x-1, y, floor.MapY) == 0
	var a, b, c, d int
	if dir0 == 0 {
		switch dir6 {
		case 0:
			a = 0x16
		case 3:
			a = 0x0D
		case 1:
			if clear {
				a = 0
			} else {
				a = 0x0D
			}
		}
	} else if dir0 == 3 || dir0 == 1 {
		if dir6 == 0 {
			if clear {
				a = 0x0F
			} else {
				a = 5
			}
		} else if clear {
			a = 0x12
		} else {
			a = 2
		}
	}
	if dir0 == 0 {
		b = 0x16
	} else if dir0 == 3 {
		b = 0x11
	} else if dir0 == 1 {
		b = 5
	}
	switch dir6 {
	case 0:
		if dir0 == 0 {
			c = 0x16
		} else if clear {
			c = 0x10
		} else {
			c = 0x0A
		}
	case 3:
		if clear {
			c = 0x14
		} else {
			c = 7
		}
	case 1:
		if clear {
			c = 1
		} else {
			c = 3
		}
	}
	if dir6 == 0 || dir6 == 3 {
		if dir0 == 0 {
			d = 0x16
		} else if dir0 == 3 {
			d = 0x17
		} else {
			d = 0x0A
		}
	} else if dir0 == 0 {
		d = 0x0D
	} else if dir0 == 3 {
		d = 0x15
	} else {
		d = 6
	}
	setDungeonRelative(floor, column, row, 1, 0, a)
	setDungeonRelative(floor, column, row, 2, 0, b)
	setDungeonRelative(floor, column, row, 1, 1, c)
	setDungeonRelative(floor, column, row, 2, 1, d)
}

func buildDungeonTiles4(floor *DungeonFloor, grid geo.Grid, column, row, x, y, dir0, dir2, mapY int) {
	var7 := dungeonDirFlags(grid, 2, x, y-1, mapY)
	var8 := dungeonDirFlags(grid, 0, x+1, y, mapY)
	clear := var7 == 0 && var8 == 0
	var a, b, c, d int
	if dir0 == 0 {
		if var7 == 1 {
			a = 4
		} else {
			a = 0x16
		}
	} else if dir0 == 3 {
		a = 0x0F
	} else {
		a = 5
	}
	if dir0 == 0 {
		if var7 == 0 {
			b = 0x16
		} else if var7 == 3 {
			if dir2 == 0 && var8 != 0 {
				b = 0x18
			} else {
				b = 1
			}
		} else if dir2 == 0 {
			if var8 != 0 {
				b = 0x0B
			} else {
				b = 7
			}
		} else {
			b = 3
		}
	} else if dir2 != 0 {
		b = 9
	} else if var8 != 0 {
		b = 5
	} else if clear {
		b = 0x11
	} else {
		b = 0x13
	}
	if dir0 == 0 {
		c = 0x16
	} else if dir0 == 3 {
		c = 0x10
	} else {
		c = 0x0A
	}
	if dir0 == 0 {
		if var7 == 0 {
			d = 0x16
		} else if dir2 != 0 {
			d = 4
		} else if var8 == 0 {
			d = 8
		} else {
			d = 0x0C
		}
	} else if dir2 != 0 {
		d = 0x0E
	} else if var8 == 0 {
		d = 0x17
	} else {
		d = 0x0A
	}
	setDungeonRelative(floor, column, row, 5, 0, a)
	setDungeonRelative(floor, column, row, 6, 0, b)
	setDungeonRelative(floor, column, row, 5, 1, c)
	setDungeonRelative(floor, column, row, 6, 1, d)
}

func setDungeonRelative(floor *DungeonFloor, column, row, x, y, tileID int) {
	// Local coordinates are added to the same 13x5 origin as the reference.
	mapX := column*6 + row*5 + 21 + x
	mapY := row*5 + 10 + y
	if tileID >= 0 && mapX >= 0 && mapX < WildernessWidth && mapY >= 0 && mapY < WildernessHeight {
		floor.Tiles[mapY][mapX] = uint8(tileID + 1)
	}
}

func dungeonDirFlags(grid geo.Grid, direction, x, y, centerY int) int {
	current := dungeonOneDirection(grid, direction, x, y, centerY)
	dx, dy := directionDelta(direction)
	neighbor := dungeonOneDirection(grid, (direction+4)%8, x+dx, y+dy, centerY)
	return current | neighbor
}

func dungeonOneDirection(grid geo.Grid, direction, x, y, centerY int) int {
	if x < 0 || x >= geo.Width || y < 0 || y >= geo.Height {
		if y == centerY && (direction == 2 || direction == 6) {
			return 0
		}
		return 1
	}
	wall, _ := grid.Wall(x, y, direction)
	if wall == 0 {
		return 0
	}
	cell, _ := grid.Cell(x, y)
	detail := uint8(0)
	switch direction {
	case 0:
		detail = cell.DetailDirections[0]
	case 2:
		detail = cell.DetailDirections[1]
	case 4:
		detail = cell.DetailDirections[2]
	case 6:
		detail = cell.DetailDirections[3]
	}
	if detail == 0 {
		return 1
	}
	return 3
}

func directionDelta(direction int) (int, int) {
	switch direction {
	case 0:
		return 0, -1
	case 2:
		return 1, 0
	case 4:
		return 0, 1
	case 6:
		return -1, 0
	default:
		return 0, 0
	}
}
