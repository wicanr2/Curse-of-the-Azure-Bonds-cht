package gfx

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/png"
)

//go:embed assets/dos-character-creation-frame.png
var characterCreationFramePNG []byte

// CharacterCreationFrame returns the fixed single-panel chrome extracted from
// the local DOS character-creation runtime capture. Unlike AdventureFrame it
// deliberately has no 128px roster divider.
func CharacterCreationFrame() image.Image {
	frame, err := png.Decode(bytes.NewReader(characterCreationFramePNG))
	if err != nil {
		panic(err)
	}
	return frame
}

// ExtendedCharacterCreationFrame keeps the original 184 native top rows of
// the single-panel character screen, then reuses the already reconstructed
// full-width CJK message/command extension. This preserves source pixels in
// the original panel and command band while clearly limiting the added lower
// information area to layout-reconstructed chrome.
func ExtendedCharacterCreationFrame() image.Image {
	source := CharacterCreationFrame()
	lower := ExtendedAdventureFrame()
	result := image.NewRGBA(image.Rect(0, 0, 320, 240))
	draw.Draw(result, image.Rect(0, 0, 320, 184), source, image.Point{}, draw.Src)
	draw.Draw(result, image.Rect(0, 184, 320, 240), lower, image.Pt(0, 184), draw.Src)
	return result
}
