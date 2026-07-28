package gfx

import (
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	engineview "github.com/wicanr2/golden-box-remake-engine/viewport"
)

type WallLayoutCall = engineview.WallLayoutCall
type WallView = engineview.WallView

func TraverseWallView(grid geo.Grid, partyDirection uint8, partyX, partyY int) (WallView, error) {
	return engineview.TraverseWallView(grid, partyDirection, partyX, partyY)
}

func TraverseWallViewWrapped(grid geo.Grid, partyDirection uint8, partyX, partyY int) (WallView, error) {
	return engineview.TraverseWallViewWrapped(grid, partyDirection, partyX, partyY)
}
