// Command extract_dos_character_creation_frame derives transparent single-panel
// character-creation chrome from a local native DOS runtime capture. It keeps
// only fixed frame pixels; original character text and values are never copied
// into the remake asset.
package main

import (
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func main() {
	const source = "docs/reference/original-dos/character-age-create.png"
	const commandDividerSource = "internal/gfx/assets/dos-adventure-frame.png"
	const destination = "internal/gfx/assets/dos-character-creation-frame.png"

	input, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	oracle, err := png.Decode(input)
	if err != nil {
		log.Fatal(err)
	}
	if oracle.Bounds().Dx() != 320 || oracle.Bounds().Dy() != 200 {
		log.Fatalf("oracle is %v, want 320x200", oracle.Bounds())
	}
	dividerInput, err := os.Open(commandDividerSource)
	if err != nil {
		log.Fatal(err)
	}
	defer dividerInput.Close()
	divider, err := png.Decode(dividerInput)
	if err != nil {
		log.Fatal(err)
	}
	if divider.Bounds().Dx() != 320 || divider.Bounds().Dy() != 200 {
		log.Fatalf("divider source is %v, want 320x200", divider.Bounds())
	}

	frame := image.NewNRGBA(image.Rect(0, 0, 320, 200))
	retain := func(x, y int) bool {
		// The DOS character sheet has one full-width upper panel. Its fixed
		// outer stone edge occupies the outer eight pixels. The source screenshot
		// has live STATUS text immediately above its lower divider, so that
		// divider is copied separately from the text-free adventure command band.
		return y < 176 && (x < 8 || x >= 312 || y < 8)
	}
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			if retain(x, y) {
				frame.Set(x, y, oracle.At(x, y))
			}
		}
	}
	for y := 176; y < 184; y++ {
		for x := 0; x < 320; x++ {
			frame.Set(x, y, divider.At(x, y+8))
		}
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		log.Fatal(err)
	}
	output, err := os.Create(destination)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()
	if err := png.Encode(output, frame); err != nil {
		log.Fatal(err)
	}
}
