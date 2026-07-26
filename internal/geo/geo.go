// Package geo decodes the 16x16 isometric map geometry blocks used by the
// Gold Box wilderness and city-map renderer. Rendering tiles and wall art are
// separate layers; this package only exposes the bytes proven by the GEO
// block layout.
package geo

import "fmt"

const (
	Width       = 16
	Height      = 16
	PayloadSize = 0x400
	BlockSize   = 2 + PayloadSize
)

// Cell is one map cell. WallDirections and DetailDirections use the original
// direction order 0, 2, 4, 6. Wall values are packed nibbles; detail values
// are packed 2-bit fields. Terrain is the unmodified byte at payload 0x200.
type Cell struct {
	WallDirections   [4]uint8
	Terrain          uint8
	DetailDirections [4]uint8
}

// Grid is a decoded 16x16 GEO block.
type Grid struct {
	BlockID uint8
	Cells   [Height][Width]Cell
}

// Parse decodes one DAX-decoded GEO block. SSI GEO blocks observed in the
// original image contain a two-byte prefix followed by four 0x100-byte
// planes. Accepting a bare 0x400-byte payload also keeps the parser useful for
// callers that have already removed that prefix.
func Parse(blockID uint8, data []byte) (Grid, error) {
	if len(data) != BlockSize && len(data) != PayloadSize {
		return Grid{}, fmt.Errorf("GEO block 0x%02X has %d bytes, want 0x%X or 0x%X", blockID, len(data), BlockSize, PayloadSize)
	}
	payload := data
	if len(data) == BlockSize {
		payload = data[2:]
	}
	var grid Grid
	grid.BlockID = blockID
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			index := y*Width + x
			walls := payload[index]
			extra := payload[0x300+index]
			grid.Cells[y][x] = Cell{
				WallDirections: [4]uint8{
					(walls >> 4) & 0x0F,
					walls & 0x0F,
					(payload[0x100+index] >> 4) & 0x0F,
					payload[0x100+index] & 0x0F,
				},
				Terrain: payload[0x200+index],
				DetailDirections: [4]uint8{
					extra & 0x03,
					(extra >> 2) & 0x03,
					(extra >> 4) & 0x03,
					(extra >> 6) & 0x03,
				},
			}
		}
	}
	return grid, nil
}

func (g Grid) Cell(x, y int) (Cell, bool) {
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return Cell{}, false
	}
	return g.Cells[y][x], true
}

// Wall returns the raw wall type for one of the four cardinal directions
// used by the original engine: 0=north, 2=east, 4=south, 6=west.
func (g Grid) Wall(x, y, direction int) (uint8, bool) {
	cell, ok := g.Cell(x, y)
	if !ok {
		return 0, false
	}
	index := -1
	switch direction {
	case 0:
		index = 0
	case 2:
		index = 1
	case 4:
		index = 2
	case 6:
		index = 3
	}
	if index < 0 {
		return 0, false
	}
	return cell.WallDirections[index], true
}

// CanMove checks only the proven GEO wall fields. It deliberately does not
// apply background-tile movement cost, encounters, or party placement rules.
func (g Grid) CanMove(x, y, direction int) bool {
	if direction != 0 && direction != 2 && direction != 4 && direction != 6 {
		return false
	}
	dx, dy := 0, 0
	switch direction {
	case 0:
		dy = -1
	case 2:
		dx = 1
	case 4:
		dy = 1
	case 6:
		dx = -1
	}
	if _, ok := g.Cell(x+dx, y+dy); !ok {
		return false
	}
	wall, _ := g.Wall(x, y, direction)
	opposite := (direction + 4) % 8
	other, _ := g.Wall(x+dx, y+dy, opposite)
	return wall == 0 && other == 0
}
