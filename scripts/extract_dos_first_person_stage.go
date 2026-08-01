// Command extract_dos_first_person_stage derives the fixed grey inset around
// the DOS first-person/PIC viewport. It retains only UI-frame pixels, leaving
// the 88x88 scene interior transparent.
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
	const destination = "internal/gfx/assets/dos-first-person-stage-frame.png"

	input, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	oracle, err := png.Decode(input)
	if err != nil {
		log.Fatal(err)
	}
	if oracle.Bounds() != image.Rect(0, 0, 320, 200) {
		log.Fatalf("oracle is %v, want 320x200", oracle.Bounds())
	}

	frame := image.NewNRGBA(oracle.Bounds())
	for y := 21; y < 121; y++ {
		for x := 21; x < 121; x++ {
			if x >= 24 && x < 118 && y >= 24 && y < 118 {
				continue
			}
			frame.Set(x, y, oracle.At(x, y))
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
