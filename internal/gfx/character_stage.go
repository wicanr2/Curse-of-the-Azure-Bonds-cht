package gfx

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

//go:embed assets/dos-character-stage-frame.png
var dosCharacterStageFramePNG []byte

// CharacterStageFrame returns the DOS portrait inset reconstructed from the
// supplied 320x200 runtime oracle. It is separate from the stone chrome because
// first-person scenes and layered HEAD/BODY characters use different stages.
func CharacterStageFrame() image.Image {
	frame, err := png.Decode(bytes.NewReader(dosCharacterStageFramePNG))
	if err != nil {
		panic(err)
	}
	return frame
}
