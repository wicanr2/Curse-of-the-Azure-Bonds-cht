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
	SymbolBlockIDs []uint8
	Symbols        map[uint8]Picture
}

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
		SymbolBlockIDs: make([]uint8, len(walls)),
		Symbols:        make(map[uint8]Picture, len(walls)),
	}
	for index := range walls {
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
		result.Symbols[symbolID] = picture
	}
	return result, nil
}
