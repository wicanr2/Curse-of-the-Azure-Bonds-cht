package gfx

import (
	"fmt"
	"image"
	"image/color"
)

// EGA16 is the original engine's default 16-colour palette. Keeping it in a
// platform-neutral package lets Ebiten and offline tools share exact pixels.
func ega(r, g, b uint8) color.RGBA { return color.RGBA{R: r, G: g, B: b, A: 255} }

var EGA16 = [16]color.RGBA{
	ega(0, 0, 0), ega(0, 0, 173), ega(0, 173, 0), ega(0, 173, 173),
	ega(173, 0, 0), ega(173, 0, 173), ega(173, 82, 0), ega(173, 173, 173),
	ega(82, 82, 82), ega(82, 82, 255), ega(82, 255, 82), ega(82, 255, 255),
	ega(255, 82, 82), ega(255, 82, 255), ega(255, 255, 82), ega(255, 255, 255),
}

// RGBA renders one indexed picture item. Palette index 16 is transparent,
// matching the masked DAX picture convention; all other indices must be in
// the 16-colour palette.
func (p Picture) RGBA(item int, palette [16]color.RGBA) (*image.RGBA, error) {
	if item < 0 || item >= int(p.ItemCount) {
		return nil, fmt.Errorf("picture item %d is outside 0..%d", item, int(p.ItemCount)-1)
	}
	output := image.NewRGBA(image.Rect(0, 0, p.Width(), p.Height()))
	for y := 0; y < p.Height(); y++ {
		for x := 0; x < p.Width(); x++ {
			value, _ := p.Pixel(item, x, y)
			if value == 16 {
				output.SetRGBA(x, y, color.RGBA{})
				continue
			}
			if value >= uint8(len(palette)) {
				return nil, fmt.Errorf("picture item %d has palette index %d at (%d,%d)", item, value, x, y)
			}
			output.SetRGBA(x, y, palette[value])
		}
	}
	return output, nil
}
