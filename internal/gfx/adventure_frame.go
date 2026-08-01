package gfx

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/png"
)

//go:embed assets/dos-adventure-frame.png
var adventureFramePNG []byte

// AdventureFrame returns the exact fixed chrome pixels sampled from the local
// 320x200 DOS runtime oracle. Panel interiors are transparent.
func AdventureFrame() image.Image {
	frame, err := png.Decode(bytes.NewReader(adventureFramePNG))
	if err != nil {
		panic(err)
	}
	return frame
}

// ExtendedAdventureFrame keeps the DOS 320x200 geometry through the message
// panel, inserts forty native rows into that panel, and moves the original
// command strip to y=224. At the renderer's integer 2x scale this becomes the
// 640x480 CJK layout without distorting any original raster pixel.
//
// The top 184 rows and final 16 command rows remain exact source material. The
// inserted side-wall band is layout-reconstructed from the final forty rows of
// the original message panel; callers must not label the whole 320x240 image
// pixel-exact.
func ExtendedAdventureFrame() image.Image {
	source := AdventureFrame()
	result := image.NewRGBA(image.Rect(0, 0, 320, 240))
	draw.Draw(result, image.Rect(0, 0, 320, 184), source, image.Point{}, draw.Src)
	draw.Draw(result, image.Rect(0, 184, 320, 224), source, image.Pt(0, 144), draw.Src)
	draw.Draw(result, image.Rect(0, 224, 320, 240), source, image.Pt(0, 184), draw.Src)
	return result
}
