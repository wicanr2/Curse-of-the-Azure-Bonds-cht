package main

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func iconTestImage(colors ...color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, len(colors), 1))
	for x, value := range colors {
		img.Set(x, 0, value)
	}
	return img
}

func TestPaintedThemeComposesAnUncachedHeadBodyPairAndRecolorsIt(t *testing.T) {
	key := "party-head-01-body-02.png" // deliberately absent as a precomposite
	head, body := "chead-block-01-item-00.png", "cbody-block-02-item-00.png"
	modern, err := loadOrComposeIcon(filepath.Join("..", "..", "assets", "modern-a6", "sprites"), key, head, body)
	if err != nil {
		t.Fatal(err)
	}
	guide, err := loadOrComposeIcon(filepath.Join("..", "..", "assets", "sprites"), key, head, body)
	if err != nil {
		t.Fatal(err)
	}
	if modern.Bounds().Dx() <= guide.Bounds().Dx() {
		t.Fatalf("painted=%v guide=%v", modern.Bounds(), guide.Bounds())
	}
	colors := [6]uint8{0x22, 0x33, 0x44, 0x55, 0x66, 0x77}
	got := recolorPartyPixels(modern, guide, colors, true)
	changed := false
	for y := 0; y < got.Bounds().Dy() && !changed; y++ {
		for x := 0; x < got.Bounds().Dx(); x++ {
			if got.At(x, y) != modern.At(x, y) {
				changed = true
				break
			}
		}
	}
	if !changed {
		t.Fatal("painted arbitrary head/body pair ignored all six colour fields")
	}
}

func TestOriginalIconRecolorChangesOnlyTheSelectedSemanticSlot(t *testing.T) {
	guide := iconTestImage(gfx.EGA16[1], gfx.EGA16[2], color.RGBA{R: 82, G: 82, B: 82, A: 255})
	colors := party.DefaultIconColors
	colors[0] = 0x95
	got := recolorPartyPixels(guide, guide, colors, false)
	if got.At(0, 0) != gfx.EGA16[5] {
		t.Fatalf("body first colour=%v", got.At(0, 0))
	}
	if got.At(1, 0) != gfx.EGA16[2] {
		t.Fatalf("arm was changed=%v", got.At(1, 0))
	}
	if got.At(2, 0) != (color.RGBA{R: 82, G: 82, B: 82, A: 255}) {
		t.Fatalf("outline was changed=%v", got.At(2, 0))
	}
}

func TestPaintedIconUsesOriginalGuideAndPreservesAlpha(t *testing.T) {
	guide := iconTestImage(gfx.EGA16[1], gfx.EGA16[2])
	painted := iconTestImage(
		color.RGBA{R: 210, G: 91, B: 69, A: 170}, color.RGBA{R: 100, G: 194, B: 197, A: 90},
		color.RGBA{R: 205, G: 86, B: 64, A: 160}, color.RGBA{R: 95, G: 189, B: 192, A: 80},
	)
	colors := party.DefaultIconColors
	colors[0] = 0x92
	got := recolorPartyPixels(painted, guide, colors, true)
	if _, _, _, alpha := got.At(0, 0).RGBA(); alpha>>8 != 170 {
		t.Fatalf("painted alpha=%d", alpha>>8)
	}
	if got.At(0, 0) == painted.At(0, 0) {
		t.Fatal("painted body slot did not change")
	}
	if got.At(2, 0) != painted.At(2, 0) {
		t.Fatal("painted arm slot changed with body colour")
	}
}
