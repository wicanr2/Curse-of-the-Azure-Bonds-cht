package gfx

import "fmt"

const (
	CombatTileHeaderSize = 17
	CombatTileWidth      = 24
	CombatTileHeight     = 24
	combatTilePackedSize = CombatTileWidth * CombatTileHeight / 2
)

// CombatTileSet is the fixed 24x24 4bpp terrain atlas stored in DUNGCOM,
// WILDCOM and RANDCOM. The 17-byte prefix is retained until every metadata
// field is named; byte 8 is the standard SSI picture item count.
type CombatTileSet struct {
	Header [CombatTileHeaderSize]byte
	Tiles  []Picture
}

func ParseCombatTiles(data []byte) (CombatTileSet, error) {
	if len(data) < CombatTileHeaderSize {
		return CombatTileSet{}, fmt.Errorf("combat tiles are %d bytes, shorter than header", len(data))
	}
	var result CombatTileSet
	copy(result.Header[:], data[:CombatTileHeaderSize])
	count := int(result.Header[8])
	if count == 0 {
		return CombatTileSet{}, fmt.Errorf("combat tile count is zero")
	}
	expected := CombatTileHeaderSize + count*combatTilePackedSize
	if len(data) != expected {
		return CombatTileSet{}, fmt.Errorf("combat tiles are %d bytes, want %d for %d 24x24 items", len(data), expected, count)
	}
	result.Tiles = make([]Picture, 0, count)
	for item := 0; item < count; item++ {
		picture := Picture{
			WidthUnits:  CombatTileWidth / 8,
			HeightUnits: CombatTileHeight,
			ItemCount:   1,
			Pixels:      make([]uint8, CombatTileWidth*CombatTileHeight),
		}
		offset := CombatTileHeaderSize + item*combatTilePackedSize
		for pixel := 0; pixel < len(picture.Pixels); pixel += 2 {
			packed := data[offset+pixel/2]
			first, second := packed>>4, packed&0x0F
			if first == 0 {
				first = 16
			}
			if second == 0 {
				second = 16
			}
			picture.Pixels[pixel] = first
			picture.Pixels[pixel+1] = second
		}
		result.Tiles = append(result.Tiles, picture)
	}
	return result, nil
}
