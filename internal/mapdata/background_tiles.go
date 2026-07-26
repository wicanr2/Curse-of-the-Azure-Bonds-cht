// Package mapdata contains tables used after GEO geometry has selected a map
// cell. These entries describe the background-floor layer; they are separate
// from GEO wall bytes and from the TILES pixel pictures.
package mapdata

type BackgroundTile struct {
	MoveCost  uint8
	Height    uint8
	Field     uint8
	TileIndex uint8
}

// BackgroundTiles is the recovered reference-engine table. 0xFF in MoveCost
// is the original impassable sentinel; unknown/reserved tail entries are kept
// instead of being silently discarded.
var BackgroundTiles = []BackgroundTile{
	{1, 0, 0xFF, 0}, {0xFF, 1, 2, 0}, {0xFF, 1, 2, 1}, {0xFF, 1, 2, 2},
	{0xFF, 1, 2, 3}, {1, 1, 0, 4}, {0xFF, 1, 2, 5}, {0xFF, 1, 2, 6},
	{0xFF, 1, 2, 7}, {1, 1, 0, 8}, {0xFF, 1, 2, 9}, {1, 1, 0, 0x0A},
	{0xFF, 1, 2, 0x0B}, {1, 1, 0, 0x0C}, {0xFF, 1, 2, 0x0D}, {1, 1, 0, 0x0E},
	{0xFF, 1, 2, 0x0F}, {1, 1, 0, 0x10}, {0xFF, 1, 2, 0x11}, {0xFF, 1, 2, 0x12},
	{0xFF, 1, 2, 0x13}, {0xFF, 1, 2, 0x14}, {0xFF, 1, 2, 0x15}, {1, 1, 0, 0x16},
	{1, 1, 0, 0x17}, {0xFF, 1, 2, 0x18}, {2, 2, 0, 0x22}, {1, 1, 0, 0x23},
	{1, 1, 0, 0x24}, {1, 1, 0, 0x25}, {1, 1, 0, 0x26}, {1, 1, 0, 0x27},
	{0xFF, 1, 2, 0}, {0xFF, 1, 2, 1}, {0xFF, 1, 2, 2}, {0xFF, 1, 2, 3},
	{0xFF, 1, 2, 4}, {1, 1, 0, 5}, {1, 1, 0, 6}, {1, 1, 0, 7},
	{1, 1, 0, 8}, {1, 1, 0, 9}, {0xFF, 1, 2, 10}, {0xFF, 1, 2, 11},
	{1, 1, 0, 12}, {1, 1, 0, 13}, {1, 1, 0, 14}, {1, 1, 0, 15},
	{2, 1, 0, 16}, {2, 1, 0, 17}, {2, 1, 0, 18}, {2, 1, 0, 19},
	{2, 1, 0, 20}, {2, 1, 0, 21}, {1, 1, 0, 22}, {1, 1, 0, 0x17},
	{1, 1, 0, 0x18}, {1, 1, 0, 0x19}, {2, 1, 0, 0x1A}, {2, 1, 0, 0x1B},
	{4, 0, 0, 0x1C}, {4, 0, 0, 0x1D}, {4, 0, 0, 0x1E}, {4, 0, 0, 0x1F},
	{1, 1, 0, 0x20}, {1, 1, 0, 0x21}, {0, 0, 0xFF, 0xFF}, {0xFF, 0xFF, 0xFF, 0xFF},
	{0, 0, 0, 1}, {0xFF, 0xFF, 0xFF, 0xFF}, {0, 0, 1, 0}, {0xFF, 0xFF, 0xFF, 0xFF},
	{0, 0, 1, 0}, {0, 1, 1, 1},
}

func Lookup(index int) (BackgroundTile, bool) {
	if index < 0 || index >= len(BackgroundTiles) {
		return BackgroundTile{}, false
	}
	return BackgroundTiles[index], true
}

func (tile BackgroundTile) Passable() bool { return tile.MoveCost != 0xFF }
