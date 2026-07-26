package gfx

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// PieceSet is one of the three wall/symbol slots populated by the original
// LOAD PIECES command. WallDefs and Symbols intentionally remain indexed data;
// a later renderer decides how a wall definition composes its 8x8 symbols.
type PieceSet struct {
	SetID          uint8
	Selector       uint8
	WallDefs       []WallDef
	SymbolSetIDs   []uint8
	SymbolBlockIDs []uint8
	Symbols        map[uint8]Picture
}

// WallStamp is one 8x8D item placed by a reference 3D wall viewport layout.
// Row and Column are logical 8x8 cells; Item is the local item index inside
// the selected 8x8D picture block.
type WallStamp struct {
	Row       int
	Column    int
	SymbolID  uint8
	Item      int
	SymbolSet uint8
	Picture   Picture
}

var wallLayoutIndex = [...]int{0, 2, 6, 10, 22, 38, 54, 110, 132, 154}
var wallLayoutColumns = [...]int{1, 1, 1, 3, 2, 2, 7, 2, 2, 1}
var wallLayoutRows = [...]int{2, 4, 4, 4, 8, 8, 8, 11, 11, 2}

// SymbolSetBase is the reference global-symbol base for the three WALLDEF
// slots populated by LOAD PIECES. The separate 8X8D set 0 (base 0x01) is
// used by the area-map path and is intentionally outside this table.
var SymbolSetBase = [...]uint16{0, 0x2E, 0x74, 0xBA}

// ParsePieceSet applies the LOAD PIECES mapping recovered from the reference
// engine's LoadWalldef routine. A single WALLDEF record uses the selector as
// its 8X8D block ID. For concatenated WALLDEF records, subsequent symbol sets
// use selector*10 + recordIndex + 1.
func ParsePieceSet(setID, selector uint8, wallBlocks, symbolBlocks []dax.Block) (PieceSet, error) {
	if setID < 1 || setID > 3 {
		return PieceSet{}, fmt.Errorf("wall symbol set %d is outside 1..3", setID)
	}
	var wallBlock *dax.Block
	for index := range wallBlocks {
		if wallBlocks[index].Entry.ID == selector {
			wallBlock = &wallBlocks[index]
			break
		}
	}
	if wallBlock == nil {
		return PieceSet{}, fmt.Errorf("WALLDEF selector %d is not present", selector)
	}
	walls, err := ParseWallDefs(wallBlock.Data)
	if err != nil {
		return PieceSet{}, fmt.Errorf("WALLDEF selector %d: %w", selector, err)
	}

	result := PieceSet{
		SetID:          setID,
		Selector:       selector,
		WallDefs:       walls,
		SymbolSetIDs:   make([]uint8, len(walls)),
		SymbolBlockIDs: make([]uint8, len(walls)),
		Symbols:        make(map[uint8]Picture, len(walls)),
	}
	for index := range walls {
		globalSet := int(setID) + index
		if globalSet < 1 || globalSet >= len(SymbolSetBase) {
			return PieceSet{}, fmt.Errorf("piece set %d record %d exceeds symbol-set range", setID, index)
		}
		walls[index].OffsetSymbols(int(SymbolSetBase[globalSet]) - int(SymbolSetBase[1]))
		symbolID := selector
		if len(walls) > 1 {
			value := int(selector)*10 + index + 1
			if value > 0xFF {
				return PieceSet{}, fmt.Errorf("symbol block %d overflows byte", value)
			}
			symbolID = uint8(value)
		}
		var symbolBlock *dax.Block
		for blockIndex := range symbolBlocks {
			if symbolBlocks[blockIndex].Entry.ID == symbolID {
				symbolBlock = &symbolBlocks[blockIndex]
				break
			}
		}
		if symbolBlock == nil {
			return PieceSet{}, fmt.Errorf("8X8D symbol block %d is not present", symbolID)
		}
		picture, err := ParsePicture(symbolBlock.Data, false, 0)
		if err != nil {
			return PieceSet{}, fmt.Errorf("8X8D symbol block %d: %w", symbolID, err)
		}
		result.SymbolBlockIDs[index] = symbolID
		result.SymbolSetIDs[index] = uint8(globalSet)
		result.Symbols[symbolID] = picture
	}
	return result, nil
}

// WallSymbol resolves one WALLDEF cell into its 8X8D picture item. The
// returned picture retains all items so callers can choose the item index
// using globalID without losing the source bitmap.
func (p PieceSet) WallSymbol(record, row, column int) (picture Picture, globalID uint8, ok bool) {
	picture, _, globalID, ok = p.WallSymbolItem(record, row, column)
	return picture, globalID, ok
}

// WallSymbolItem is WallSymbol with the local 8x8D item index exposed for a
// renderer. It keeps the global ID and the source picture together so callers
// cannot accidentally index a symbol with the wrong set base.
func (p PieceSet) WallSymbolItem(record, row, column int) (picture Picture, item int, globalID uint8, ok bool) {
	if record < 0 || record >= len(p.WallDefs) || record >= len(p.SymbolSetIDs) || record >= len(p.SymbolBlockIDs) {
		return Picture{}, 0, 0, false
	}
	id, ok := p.WallDefs[record].ID(row, column)
	if !ok || id == 0 {
		return Picture{}, 0, 0, false
	}
	picture, ok = p.Symbols[p.SymbolBlockIDs[record]]
	if !ok {
		return Picture{}, 0, 0, false
	}
	item = int(id) - int(SymbolSetBase[p.SymbolSetIDs[record]])
	if item < 0 || item >= int(picture.ItemCount) {
		return Picture{}, 0, 0, false
	}
	return picture, item, id, true
}

// BuildWallLayout reproduces the index/window portion of reference
// draw_3D_8x8_titles. wallType values 1..15 select WALLDEF sets 1..3 and
// slices 0..4; offsetIndex selects one of the ten proven viewport shapes.
func BuildWallLayout(piece PieceSet, wallType uint8, offsetIndex, rowStart, columnStart int) ([]WallStamp, error) {
	if wallType < 1 || wallType > 15 {
		return nil, fmt.Errorf("wall type %d is outside 1..15", wallType)
	}
	if offsetIndex < 0 || offsetIndex >= len(wallLayoutIndex) {
		return nil, fmt.Errorf("wall layout %d is outside 0..%d", offsetIndex, len(wallLayoutIndex)-1)
	}
	wallSet := int((wallType-1)/5) + 1
	record := wallSet - int(piece.SetID)
	if record < 0 || record >= len(piece.WallDefs) {
		return nil, fmt.Errorf("piece set %d has no WALLDEF record for wall type %d", piece.SetID, wallType)
	}
	slice := int((wallType - 1) % 5)
	stamps := make([]WallStamp, 0, wallLayoutColumns[offsetIndex]*wallLayoutRows[offsetIndex])
	index := wallLayoutIndex[offsetIndex]
	for row := 0; row < wallLayoutRows[offsetIndex]; row++ {
		for column := 0; column < wallLayoutColumns[offsetIndex]; column++ {
			picture, item, symbolID, ok := piece.WallSymbolItem(record, slice, index)
			if ok {
				stamps = append(stamps, WallStamp{
					Row:       rowStart + row,
					Column:    columnStart + column,
					SymbolID:  symbolID,
					Item:      item,
					SymbolSet: piece.SymbolSetIDs[record],
					Picture:   picture,
				})
			}
			index++
		}
	}
	return stamps, nil
}
