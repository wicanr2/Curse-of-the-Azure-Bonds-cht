package gfx

import (
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	engineview "github.com/wicanr2/golden-box-remake-engine/viewport"
)

type WallLayoutCall = engineview.WallLayoutCall
type WallView = engineview.WallView
type Background = engineview.Background
type BackgroundRect = engineview.BackgroundRect
type StageInset = engineview.StageInset
type StageInsetFill = engineview.StageInsetFill
type SkyOverlayKind = engineview.SkyOverlayKind

const (
	SkyNorth   = engineview.SkyNorth
	SkyTransit = engineview.SkyTransit
	SkyGround  = engineview.SkyGround
)

var BuildBackground = engineview.BuildBackground
var FillBackgroundToStageInset = engineview.FillBackgroundToStageInset

func TraverseWallView(grid geo.Grid, partyDirection uint8, partyX, partyY int) (WallView, error) {
	return engineview.TraverseWallView(grid, partyDirection, partyX, partyY)
}

func TraverseWallViewWrapped(grid geo.Grid, partyDirection uint8, partyX, partyY int) (WallView, error) {
	return engineview.TraverseWallViewWrapped(grid, partyDirection, partyX, partyY)
}
