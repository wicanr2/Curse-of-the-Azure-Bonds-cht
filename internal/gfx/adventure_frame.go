package gfx

import (
	"bytes"
	_ "embed"
	"image"
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
