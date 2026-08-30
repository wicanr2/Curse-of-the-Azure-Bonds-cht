package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

var iconTemplateIndexes = [6][2]uint8{{1, 9}, {2, 10}, {3, 11}, {4, 12}, {6, 14}, {7, 15}}

func iconTemplateSlot(r, g, b byte) (part, component int, ok bool) {
	for p, pair := range iconTemplateIndexes {
		for c, index := range pair {
			want := color.RGBAModel.Convert(gfx.EGA16[index]).(color.RGBA)
			if r == want.R && g == want.G && b == want.B {
				return p, c, true
			}
		}
	}
	return 0, 0, false
}

func recolorPartyPixels(source, guide image.Image, colors [6]uint8, painted bool) *image.RGBA {
	sb, gb := source.Bounds(), guide.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, sb.Dx(), sb.Dy()))
	for y := 0; y < sb.Dy(); y++ {
		for x := 0; x < sb.Dx(); x++ {
			src := color.RGBAModel.Convert(source.At(sb.Min.X+x, sb.Min.Y+y)).(color.RGBA)
			if src.A == 0 {
				out.SetRGBA(x, y, src)
				continue
			}
			gx, gy := x*gb.Dx()/sb.Dx(), y*gb.Dy()/sb.Dy()
			mask := color.RGBAModel.Convert(guide.At(gb.Min.X+gx, gb.Min.Y+gy)).(color.RGBA)
			part, component, ok := iconTemplateSlot(mask.R, mask.G, mask.B)
			if !ok || mask.A == 0 {
				out.SetRGBA(x, y, src)
				continue
			}
			selected := colors[part] & 0x0F
			if component == 1 {
				selected = colors[part] >> 4
			}
			if selected == iconTemplateIndexes[part][component] {
				out.SetRGBA(x, y, src)
				continue
			}
			target := color.RGBAModel.Convert(gfx.EGA16[selected]).(color.RGBA)
			if !painted {
				target.A = src.A
				out.SetRGBA(x, y, target)
				continue
			}
			lum := 0.2126*float64(src.R) + 0.7152*float64(src.G) + 0.0722*float64(src.B)
			targetLum := 0.2126*float64(target.R) + 0.7152*float64(target.G) + 0.0722*float64(target.B)
			if targetLum < 1 {
				shade := byte(lum * .12)
				out.SetRGBA(x, y, color.RGBA{shade, shade, shade, src.A})
				continue
			}
			factor := lum / targetLum
			out.SetRGBA(x, y, color.RGBA{
				R: byte(math.Min(255, float64(target.R)*factor)),
				G: byte(math.Min(255, float64(target.G)*factor)),
				B: byte(math.Min(255, float64(target.B)*factor)), A: src.A,
			})
		}
	}
	return out
}

func loadIconPNG(path string) (image.Image, error) {
	handle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	return png.Decode(handle)
}

func loadOrComposeIcon(directory, compositeKey, headKey, bodyKey string) (image.Image, error) {
	if decoded, err := loadIconPNG(filepath.Join(directory, compositeKey)); err == nil {
		return decoded, nil
	}
	head, err := loadIconPNG(filepath.Join(directory, headKey))
	if err != nil {
		return nil, err
	}
	body, err := loadIconPNG(filepath.Join(directory, bodyKey))
	if err != nil {
		return nil, err
	}
	bounds := body.Bounds()
	result := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(result, result.Bounds(), body, bounds.Min, draw.Src)
	draw.Draw(result, result.Bounds(), head, head.Bounds().Min, draw.Over)
	return result, nil
}

func (a *app) coloredPartySprite(theme, key, headKey, bodyKey string, colors [6]uint8) *ebiten.Image {
	if a.partyIconColorCache == nil {
		a.partyIconColorCache = make(map[string]*ebiten.Image)
	}
	cacheKey := iconColorCacheKey(theme, key, colors)
	if cached := a.partyIconColorCache[cacheKey]; cached != nil {
		return cached
	}
	guide, err := loadOrComposeIcon(filepath.Join("assets", "sprites"), key, headKey, bodyKey)
	if err != nil {
		return nil
	}
	directory, painted := filepath.Join("assets", "sprites"), false
	if theme == "modern-a6" {
		directory, painted = filepath.Join("assets", "modern-a6", "sprites"), true
	}
	source, err := loadOrComposeIcon(directory, key, headKey, bodyKey)
	if err != nil {
		return nil
	}
	result := ebiten.NewImageFromImage(recolorPartyPixels(source, guide, colors, painted))
	a.partyIconColorCache[cacheKey] = result
	return result
}

func iconColorCacheKey(theme, spriteKey string, colors [6]uint8) string {
	return fmt.Sprintf("%s:%s:%x", theme, spriteKey, colors)
}
