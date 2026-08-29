package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type framePalette struct {
	stone, light, shadow color.RGBA
}

func paletteForFrameStyle(style string) framePalette {
	switch style {
	case "B":
		return framePalette{color.RGBA{89, 84, 82, 255}, color.RGBA{151, 143, 132, 255}, color.RGBA{30, 27, 29, 255}}
	case "C":
		return framePalette{color.RGBA{221, 204, 164, 255}, color.RGBA{255, 239, 196, 255}, color.RGBA{112, 87, 50, 255}}
	default:
		return framePalette{color.RGBA{203, 194, 170, 255}, color.RGBA{249, 239, 207, 255}, color.RGBA{64, 55, 43, 255}}
	}
}

func drawConfigurableOuterFrame(screen *ebiten.Image, width, height int, settings uiSettings) {
	sx, sy := float64(width)/640, float64(height)/480
	outer := float64(settings.OuterBorderPX)
	palette := paletteForFrameStyle(settings.FrameStyle)
	line := func(x, y, w, h float64, ink color.Color) { ebitenutil.DrawRect(screen, x*sx, y*sy, w*sx, h*sy, ink) }
	line(0, 0, 640, outer, palette.stone)
	line(0, 480-outer, 640, outer, palette.shadow)
	line(0, 0, outer, 480, palette.stone)
	line(640-outer, 0, outer, 480, palette.shadow)
	line(2, 2, 636, 2, palette.light)
	step := 18.0
	if settings.FrameStyle == "B" {
		step = 14
	}
	if settings.FrameStyle == "C" {
		step = 24
	}
	for x := outer; x < 640-outer-6; x += step {
		line(x, outer-3, 6, 1, palette.shadow)
		line(x+2, outer-2, 6, 1, palette.light)
	}
}

func drawConfigurableAdventureFrame(screen *ebiten.Image, width, height int, settings uiSettings) {
	drawConfigurableOuterFrame(screen, width, height, settings)
	sx, sy := float64(width)/640, float64(height)/480
	inner := float64(settings.InnerBorderPX)
	palette := paletteForFrameStyle(settings.FrameStyle)
	gold, glint, dark := color.RGBA{255, 213, 45, 255}, color.RGBA{255, 250, 176, 255}, color.RGBA{126, 69, 2, 255}
	line := func(x, y, w, h float64, ink color.Color) { ebitenutil.DrawRect(screen, x*sx, y*sy, w*sx, h*sy, ink) }
	line(264, float64(settings.OuterBorderPX), inner, 256-float64(settings.OuterBorderPX), palette.stone)
	line(float64(settings.OuterBorderPX), 256, 630-float64(settings.OuterBorderPX), inner, palette.stone)
	line(float64(settings.OuterBorderPX), 454, 630-float64(settings.OuterBorderPX), inner, palette.shadow)
	line(264, float64(settings.OuterBorderPX), 1, 256-float64(settings.OuterBorderPX), palette.light)
	line(264+inner-1, float64(settings.OuterBorderPX), 1, 256-float64(settings.OuterBorderPX), palette.shadow)
	// Scene inset remains inside its dedicated empty margin, independent of the
	// outer width, so even the 20px maximum cannot cover portrait or text.
	line(46, 34, 194, 2, dark)
	line(46, 232, 194, 2, dark)
	line(46, 34, 2, 200, dark)
	line(238, 34, 2, 200, dark)
	line(48, 36, 190, 2, glint)
	line(48, 230, 190, 2, gold)
	line(48, 36, 2, 196, glint)
	line(236, 36, 2, 196, gold)
}

func drawConfigurableCombatFrame(screen *ebiten.Image, width, height int, settings uiSettings) {
	drawConfigurableOuterFrame(screen, width, height, settings)
	sx, sy := float64(width)/640, float64(height)/480
	inner := float64(settings.InnerBorderPX)
	palette := paletteForFrameStyle(settings.FrameStyle)
	gold, glint := color.RGBA{255, 213, 45, 255}, color.RGBA{255, 250, 176, 255}
	line := func(x, y, w, h float64, ink color.Color) { ebitenutil.DrawRect(screen, x*sx, y*sy, w*sx, h*sy, ink) }
	line(358, 16, inner, 344, palette.stone)
	line(16, 358, 608, inner, palette.stone)
	if inner >= 3 {
		line(359, 16, 1, 344, glint)
		line(361, 16, 1, 344, gold)
		line(16, 359, 608, 1, glint)
		line(16, 361, 608, 1, gold)
	}
}
