// Command asset-overview builds a deterministic README contact sheet from the
// already-exported remake PNG assets. It never reads the proprietary archives.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	canvasWidth  = 1280
	canvasHeight = 800
)

type section struct {
	title    string
	patterns []string
	limit    int
	columns  int
	x, y     int
	w, h     int
}

func main() {
	output := "docs/reference/remake-png-asset-overview.png"
	if len(os.Args) > 1 {
		output = os.Args[1]
	}
	sections := []section{
		{"SCENE / PORTRAIT PNG", []string{"assets/sprites/character-*.png", "assets/sprites/pic*-block-*-item-00.png", "assets/sprites/party*.png"}, 18, 6, 24, 64, 604, 344},
		{"COMBAT SPRITE PNG", []string{"assets/sprites/cpic*-block-*-item-00.png", "assets/sprites/comspr-*.png"}, 30, 8, 652, 64, 604, 344},
		{"WORLD / AREA TILESET PNG", []string{"assets/runtime-images/tiles/*.png", "assets/runtime-images/symbols/8x8d1-*.png"}, 56, 8, 24, 432, 604, 344},
		{"COMBAT TILESET PNG", []string{"assets/runtime-images/combat/*.png", "assets/runtime-images/sky/*.png"}, 48, 8, 652, 432, 604, 344},
	}
	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.RGBA{7, 11, 19, 255}}, image.Point{}, draw.Src)
	label(canvas, 24, 34, "CURSE OF THE AZURE BONDS REMAKE - EXPORTED PNG ASSET OVERVIEW", color.RGBA{255, 220, 92, 255})
	for _, current := range sections {
		paths, err := collect(current.patterns, current.limit)
		if err != nil {
			fatal(err)
		}
		drawSection(canvas, current, paths)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		fatal(err)
	}
	handle, err := os.Create(output)
	if err != nil {
		fatal(err)
	}
	if err := png.Encode(handle, canvas); err != nil {
		handle.Close()
		fatal(err)
	}
	if err := handle.Close(); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s\n", output)
}

func collect(patterns []string, limit int) ([]string, error) {
	var all []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		all = append(all, matches...)
	}
	sort.Strings(all)
	if len(all) == 0 {
		return nil, fmt.Errorf("no PNG matched %s", strings.Join(patterns, ", "))
	}
	if len(all) <= limit {
		return all, nil
	}
	selected := make([]string, 0, limit)
	for index := 0; index < limit; index++ {
		selected = append(selected, all[index*(len(all)-1)/(limit-1)])
	}
	return selected, nil
}

func drawSection(dst *image.RGBA, current section, paths []string) {
	border := color.RGBA{105, 90, 58, 255}
	draw.Draw(dst, image.Rect(current.x, current.y, current.x+current.w, current.y+current.h), &image.Uniform{color.RGBA{13, 20, 32, 255}}, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(current.x, current.y, current.x+current.w, current.y+2), &image.Uniform{border}, image.Point{}, draw.Src)
	label(dst, current.x+12, current.y+24, current.title, color.RGBA{235, 224, 198, 255})
	columns := current.columns
	if columns < 1 {
		columns = 8
	}
	rows := (len(paths) + columns - 1) / columns
	cellW := (current.w - 24) / columns
	cellH := (current.h - 46) / rows
	for index, path := range paths {
		source, err := readPNG(path)
		if err != nil {
			fatal(err)
		}
		x := current.x + 12 + index%columns*cellW
		y := current.y + 38 + index/columns*cellH
		drawNearest(dst, source, image.Rect(x+3, y+3, x+cellW-3, y+cellH-3))
	}
}

func readPNG(path string) (image.Image, error) {
	handle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	return png.Decode(handle)
}

func drawNearest(dst *image.RGBA, src image.Image, box image.Rectangle) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if sw == 0 || sh == 0 {
		return
	}
	scale := min(box.Dx()/sw, box.Dy()/sh)
	if scale < 1 {
		scale = 1
	}
	w, h := sw*scale, sh*scale
	if w > box.Dx() || h > box.Dy() {
		w, h = box.Dx(), box.Dy()
	}
	x0, y0 := box.Min.X+(box.Dx()-w)/2, box.Min.Y+(box.Dy()-h)/2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sourceX := src.Bounds().Min.X + x*sw/w
			sourceY := src.Bounds().Min.Y + y*sh/h
			dst.Set(x0+x, y0+y, src.At(sourceX, sourceY))
		}
	}
}

func label(dst draw.Image, x, y int, value string, ink color.Color) {
	(&font.Drawer{Dst: dst, Src: image.NewUniform(ink), Face: basicfont.Face7x13, Dot: fixed.P(x, y)}).DrawString(value)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
