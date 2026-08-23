package gfx

import (
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	enginegfx "github.com/wicanr2/golden-box-remake-engine/graphics"
)

type PieceSet = enginegfx.PieceSet
type WallStamp = enginegfx.WallStamp

var SymbolSetBase = enginegfx.SymbolSetBase
var BuildWallLayout = enginegfx.BuildWallLayout
var RawWallStamps = enginegfx.RawWallStamps

func ParsePieceSet(setID, selector uint8, wallBlocks, symbolBlocks []dax.Block) (PieceSet, error) {
	return enginegfx.ParsePieceSet(setID, selector, decodedBlocks(wallBlocks), decodedBlocks(symbolBlocks))
}

func decodedBlocks(blocks []dax.Block) map[uint8][]byte {
	result := make(map[uint8][]byte, len(blocks))
	for _, block := range blocks {
		result[block.Entry.ID] = block.Data
	}
	return result
}
