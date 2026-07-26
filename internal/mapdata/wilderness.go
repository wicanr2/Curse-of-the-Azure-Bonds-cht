package mapdata

import "math/rand"

const (
	WildernessWidth  = 50
	WildernessHeight = 25
)

// WildernessFloor is the 50x25 background-entry map built by the original
// SetupWildernessFloor routine. Values are indices into BackgroundTiles, not
// direct indices into TILES.DAX.
type WildernessFloor struct {
	Tiles     [WildernessHeight][WildernessWidth]uint8
	CityFlags int
	Seed      int64
}

// CityFlags is the reference engine's city-info table. The caller can use a
// city index when it has decoded the area pointer; GenerateWilderness also
// accepts raw flags for tests and future map loaders.
var CityFlags = [...]int{
	0x00, 0x18, 0x11, 0x15, 0x01, 0x01, 0x60, 0x14,
	0x08, 0x01, 0x00, 0x21, 0x71, 0x09, 0x06, 0x04,
	0x01, 0x09, 0x09, 0x08, 0x59, 0x00, 0x11, 0x11,
	0x00, 0x00, 0x01, 0x11, 0x00, 0x00, 0x20, 0x20,
	0x0A,
}

func CityInfo(index int) (int, bool) {
	if index < 0 || index >= len(CityFlags) {
		return 0, false
	}
	return CityFlags[index], true
}

// GenerateWilderness reproduces the reference wilderness-floor construction
// with an explicit seed. The original uses 1-based dice (roll_dice), so the
// same rule order is retained while making output deterministic for tests and
// replay.
func GenerateWilderness(cityFlags int, seed int64) WildernessFloor {
	floor := WildernessFloor{CityFlags: cityFlags, Seed: seed}
	for y := 0; y < WildernessHeight; y++ {
		for x := 0; x < WildernessWidth; x++ {
			floor.Tiles[y][x] = 23
		}
	}
	rng := rand.New(rand.NewSource(seed))
	roll := func(sides int) int { return rng.Intn(sides) + 1 }

	// SetupWildernessFloor01: a possible vertical road/terrain strip.
	chance := 0
	if cityFlags&0x20 != 0 {
		chance = 0x23
	}
	if cityFlags&0x10 != 0 {
		chance = 0x4B
	}
	if roll(100) <= chance {
		mapX := 0x22 - roll(4) - roll(4) - roll(4) - roll(4) - roll(4)
		for (mapX+2)%7 > 0 {
			mapX--
		}
		for mapY := 0; mapY <= 0x18; mapY++ {
			if mapX <= 0x31 {
				if inBounds(mapX, mapY) {
					floor.Tiles[mapY][mapX] = uint8(roll(2) + 0x3B)
				}
				if mapX < 0x31 && inBounds(mapX+1, mapY) {
					floor.Tiles[mapY][mapX+1] = uint8(roll(2) + 0x3D)
				}
				if roll(20) == 1 {
					setGroundTile40(&floor, mapX, mapY)
				}
				mapX++
			}
		}
	}

	// SetupWildernessFloor02: convert long runs of tile_index 22.
	if cityFlags&0x80 == 0 {
		neededRoll := 10
		if cityFlags&2 != 0 {
			neededRoll -= 5
		}
		if cityFlags&4 != 0 {
			neededRoll -= 2
		}
		if cityFlags&0x40 != 0 {
			neededRoll += 5
		}
		if cityFlags&8 != 0 {
			neededRoll += 10
		}
		if neededRoll < 0 {
			neededRoll = 1
		}
		for x := 0; x < WildernessWidth; x++ {
			for y := 1; y < WildernessHeight; y++ {
				if tileIndex(floor.Tiles[y][x]) == 22 && tileIndex(floor.Tiles[y-1][x]) == 22 && neededRoll >= roll(100) {
					if neededRoll >= roll(100) {
						floor.Tiles[y][x] = uint8(roll(2) + 0x29)
					} else {
						floor.Tiles[y-1][x] = uint8(roll(5) + 0x1F)
						floor.Tiles[y][x] = uint8(roll(5) + 0x24)
					}
				}
			}
		}
	}

	// SetupWildernessFloor03: probabilistic terrain groups based on city flags.
	var group int = 50
	if cityFlags&0x10 != 0 {
		group += 10
	}
	if cityFlags&0x20 != 0 {
		group += 30
	}
	if cityFlags&0x40 != 0 {
		group += 20
	}
	if cityFlags&4 != 0 {
		group -= 10
	}
	if cityFlags&2 != 0 {
		group -= 20
	}
	if cityFlags&0x80 != 0 {
		group -= 50
	}
	for x := 0; x < 50; x++ {
		for y := 0; y < 25; y++ {
			if tileIndex(floor.Tiles[y][x]) != 22 {
				continue
			}
			switch {
			case group >= -30 && group <= 9:
				setGroupMapStepped(&floor, x, y, 15, 30, 0, 0, 0, roll)
			case group >= 10 && group <= 29:
				setGroupMapStepped(&floor, x, y, 10, 14, 5, 1, 0, roll)
			case group >= 30 && group <= 69:
				setGroupMapStepped(&floor, x, y, 5, 10, 5, 2, 0, roll)
			case group >= 60 && group <= 89:
				setGroupMapStepped(&floor, x, y, 1, 10, 10, 2, 10, roll)
			case group >= 90 && group <= 110:
				setGroupMapStepped(&floor, x, y, 1, 10, 15, 5, 15, roll)
			}
		}
	}
	return floor
}

func (floor WildernessFloor) Entry(x, y int) (BackgroundTile, bool) {
	if x < 0 || x >= WildernessWidth || y < 0 || y >= WildernessHeight {
		return BackgroundTile{}, false
	}
	return Lookup(int(floor.Tiles[y][x]))
}

func (floor WildernessFloor) CanEnter(x, y int) bool {
	tile, ok := floor.Entry(x, y)
	return ok && tile.Passable()
}

func inBounds(x, y int) bool {
	return x >= 0 && x < WildernessWidth && y >= 0 && y < WildernessHeight
}

func setGroundTile40(floor *WildernessFloor, x, y int) {
	if x < 0x31 && inBounds(x+1, y) {
		floor.Tiles[y][x+1] = 0x40
	}
	if y < 0x18 && x < 0x31 && inBounds(x+1, y+1) {
		floor.Tiles[y+1][x+1] = 0x41
	}
}

func tileIndex(entry uint8) uint8 {
	tile, ok := Lookup(int(entry))
	if !ok {
		return 0xFF
	}
	return tile.TileIndex
}

func setGroupMapStepped(floor *WildernessFloor, x, y, stepE, stepD, stepC, stepB, stepA int, roll func(int) int) {
	rollValue := roll(100)
	switch {
	case rollValue <= stepA:
		floor.Tiles[y][x] = uint8(roll(2) + 0x39)
	case rollValue <= stepA+stepB:
		floor.Tiles[y][x] = uint8(roll(2) + 0x2F)
	case rollValue <= stepA+stepB+stepC:
		floor.Tiles[y][x] = uint8(roll(4) + 0x2B)
	case rollValue <= stepA+stepB+stepC+stepD:
		floor.Tiles[y][x] = uint8(roll(3) + 0x36)
	case rollValue <= stepA+stepB+stepC+stepD+stepE:
		floor.Tiles[y][x] = uint8(roll(4) + 0x31)
	}
}
