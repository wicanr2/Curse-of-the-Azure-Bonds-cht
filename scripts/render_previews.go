// Command render_previews creates deterministic README evidence from the
// original DAX assets and the same renderer-neutral parsers used by the game.
package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

func main() {
	if err := os.MkdirAll("docs/screenshots", 0755); err != nil {
		panic(err)
	}
	data, err := readMember("curseoftheazurebonds.zip", "TILES.DAX")
	if err != nil {
		panic(err)
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		panic(err)
	}
	var pictures []gfx.Picture
	for _, block := range blocks {
		picture, err := gfx.ParsePicture(block.Data, false, 0)
		if err != nil {
			panic(err)
		}
		pictures = append(pictures, picture)
	}
	if err := renderTiles(pictures); err != nil {
		panic(err)
	}
	data, err = readMember("curseoftheazurebonds.zip", "GEO2.DAX")
	if err != nil {
		panic(err)
	}
	blocks, err = dax.Parse(data)
	if err != nil || len(blocks) == 0 {
		panic(fmt.Errorf("parse GEO2.DAX: %v", err))
	}
	grid, err := geo.Parse(blocks[0].Entry.ID, blocks[0].Data)
	if err != nil {
		panic(err)
	}
	if err := renderGeo(grid); err != nil {
		panic(err)
	}
}

func readMember(path, name string) ([]byte, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		r, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	}
	return nil, fmt.Errorf("member %s not found", name)
}

func renderTiles(pictures []gfx.Picture) error {
	const scale, cellW, cellH = 2, 72, 52
	dst := image.NewRGBA(image.Rect(0, 0, 8*cellW, 6*cellH+32))
	fill(dst, color.RGBA{12, 18, 28, 255})
	index := 0
	for _, picture := range pictures {
		for item := 0; item < int(picture.ItemCount) && index < 48; item++ {
			rgba, err := picture.RGBA(item, gfx.EGA16)
			if err != nil {
				return err
			}
			x := (index%8)*cellW + 8
			y := (index/8)*cellH + 36
			drawScaled(dst, rgba, x, y, scale)
			index++
		}
	}
	return writePNG("docs/screenshots/tiles-gallery.png", dst)
}

func renderGeo(grid geo.Grid) error {
	const originX, originY, cell = 24, 32, 20
	wallColor := color.RGBA{80, 220, 235, 255}
	dst := image.NewRGBA(image.Rect(0, 0, 368, 368))
	fill(dst, color.RGBA{8, 20, 26, 255})
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			fillRect(dst, originX+x*cell+1, originY+y*cell+1, cell-1, cell-1, color.RGBA{18, 42, 48, 255})
			walls := grid.Cells[y][x].WallDirections
			if walls[0] != 0 {
				line(dst, originX+x*cell, originY+y*cell, originX+(x+1)*cell, originY+y*cell, wallColor)
			}
			if walls[1] != 0 {
				line(dst, originX+(x+1)*cell, originY+y*cell, originX+(x+1)*cell, originY+(y+1)*cell, wallColor)
			}
			if walls[2] != 0 {
				line(dst, originX+x*cell, originY+(y+1)*cell, originX+(x+1)*cell, originY+(y+1)*cell, wallColor)
			}
			if walls[3] != 0 {
				line(dst, originX+x*cell, originY+y*cell, originX+x*cell, originY+(y+1)*cell, wallColor)
			}
		}
	}
	fillRect(dst, originX+2, originY+2, 6, 6, color.RGBA{255, 220, 80, 255})
	return writePNG("docs/screenshots/geo-geometry.png", dst)
}

func fill(dst *image.RGBA, c color.Color) {
	fillRect(dst, 0, 0, dst.Bounds().Dx(), dst.Bounds().Dy(), c)
}
func fillRect(dst *image.RGBA, x, y, w, h int, c color.Color) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			dst.Set(xx, yy, c)
		}
	}
}
func drawScaled(dst *image.RGBA, src image.Image, x, y, scale int) {
	for yy := 0; yy < src.Bounds().Dy(); yy++ {
		for xx := 0; xx < src.Bounds().Dx(); xx++ {
			c := src.At(xx, yy)
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					dst.Set(x+xx*scale+sx, y+yy*scale+sy, c)
				}
			}
		}
	}
}
func line(dst *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	if x1 == x2 {
		for y := y1; y <= y2; y++ {
			dst.Set(x1, y, c)
		}
		return
	}
	for x := x1; x <= x2; x++ {
		dst.Set(x, y1, c)
	}
}
func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
