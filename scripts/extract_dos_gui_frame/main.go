// Command extract_dos_gui_frame derives transparent UI chrome from a local
// native DOS runtime capture. It never redistributes game data beyond the
// fixed interface pixels used by the remake.
package main

import (
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func main() {
	const source = "docs/reference/original-dos/tilverton-first-person-demo.png"
	const destination = "internal/gfx/assets/dos-adventure-frame.png"

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

	frame := image.NewNRGBA(image.Rect(0, 0, 320, 200))
	retain := func(x, y int) bool {
		return x < 8 || x >= 312 ||
			y < 8 ||
			(x >= 128 && x < 136 && y < 136) ||
			(y >= 128 && y < 136) ||
			(y >= 184 && y < 192)
	}
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			if retain(x, y) {
				frame.Set(x, y, oracle.At(x, y))
			}
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
