// Command extract_dos_character_stage derives the fixed inset portrait frame
// from the supplied DOS runtime oracle. It preserves only UI-frame strips,
// quantizes display-capture noise back to the EGA palette, and never includes
// character pixels.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

var ega16 = []color.NRGBA{
	{0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0xaa, 0xff},
	{0x00, 0xaa, 0x00, 0xff}, {0x00, 0xaa, 0xaa, 0xff},
	{0xaa, 0x00, 0x00, 0xff}, {0xaa, 0x00, 0xaa, 0xff},
	{0xaa, 0x55, 0x00, 0xff}, {0xaa, 0xaa, 0xaa, 0xff},
	{0x55, 0x55, 0x55, 0xff}, {0x55, 0x55, 0xff, 0xff},
	{0x55, 0xff, 0x55, 0xff}, {0x55, 0xff, 0xff, 0xff},
	{0xff, 0x55, 0x55, 0xff}, {0xff, 0x55, 0xff, 0xff},
	{0xff, 0xff, 0x55, 0xff}, {0xff, 0xff, 0xff, 0xff},
}

func nearestEGA(source color.Color) color.NRGBA {
	r, g, b, _ := source.RGBA()
	best, bestDistance := ega16[0], int64(1<<62)
	for _, candidate := range ega16 {
		dr := int64(int(r>>8) - int(candidate.R))
		dg := int64(int(g>>8) - int(candidate.G))
		db := int64(int(b>>8) - int(candidate.B))
		distance := dr*dr + dg*dg + db*db
		if distance < bestDistance {
			best, bestDistance = candidate, distance
		}
	}
	return best
}

func main() {
	const source = "docs/reference/original-dos/character-head-body-oracle.png"
	const destination = "internal/gfx/assets/dos-character-stage-frame.png"

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
	for y := 16; y < 117; y++ {
		for x := 21; x < 122; x++ {
			// The four seven-pixel strips are the fixed yellow/grey inset.
			// The centre remains transparent, so HEAD/BODY pixels cannot leak
			// into this reusable frame raster.
			if x >= 28 && x < 115 && y >= 23 && y < 110 {
				continue
			}
			frame.SetNRGBA(x, y, nearestEGA(oracle.At(x, y)))
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
