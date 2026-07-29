package gfx

import (
	"image"
	"image/draw"
)

const (
	CombatFrameWidth  = 320
	CombatFrameHeight = 184
)

// CombatFrame composes the combat geometry from stone pixels sampled from the
// native DOS adventure oracle. This preserves SSI's actual bevel, stipple,
// palette, and crack material while moving the centre divider to combat x=176.
// It is a material-exact/layout-reconstructed frame until a local DOS combat
// runtime capture can replace the composition oracle.
func CombatFrame() *image.RGBA {
	source := AdventureFrame()
	frame := image.NewRGBA(image.Rect(0, 0, CombatFrameWidth, CombatFrameHeight))

	copyRect := func(destination image.Point, sourceRect image.Rectangle) {
		draw.Draw(frame, sourceRect.Sub(sourceRect.Min).Add(destination), source, sourceRect.Min, draw.Src)
	}
	copyRect(image.Pt(0, 0), image.Rect(0, 0, 320, 8))
	copyRect(image.Pt(0, 0), image.Rect(0, 0, 8, 176))
	copyRect(image.Pt(176, 0), image.Rect(0, 0, 8, 176))
	copyRect(image.Pt(312, 0), image.Rect(312, 0, 320, 176))
	copyRect(image.Pt(0, 176), image.Rect(0, 184, 320, 192))
	return frame
}
