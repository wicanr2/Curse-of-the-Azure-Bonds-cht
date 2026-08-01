package gfx

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

//go:embed assets/dos-first-person-stage-frame.png
var dosFirstPersonStageFramePNG []byte

// FirstPersonStageFrame returns the exact DOS grey inset surrounding the
// 88x88 first-person/PIC scene. Its centre is transparent so scene pixels can
// be rendered first and clipped independently from HEAD/BODY characters.
func FirstPersonStageFrame() image.Image {
	frame, err := png.Decode(bytes.NewReader(dosFirstPersonStageFramePNG))
	if err != nil {
		panic(err)
	}
	return frame
}
