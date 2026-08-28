package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

type runtimeImageRecord struct {
	File   string `json:"file"`
	Source string `json:"source"`
	Block  uint8  `json:"block"`
	Item   int    `json:"item"`
}

type runtimeImageManifest struct {
	Version int                             `json:"version"`
	Tiles   []runtimeImageRecord            `json:"tiles"`
	Combat  map[string][]runtimeImageRecord `json:"combat"`
	Symbols []runtimeImageRecord            `json:"symbols"`
	Sky     []runtimeImageRecord            `json:"sky"`
	Walls   []runtimeWallRecord             `json:"walls"`
}

type runtimeWallRecord struct {
	Source string  `json:"source"`
	Block  uint8   `json:"block"`
	Data   []uint8 `json:"data"`
}

type runtimeImageCatalog struct {
	root     string
	manifest runtimeImageManifest
}

func loadRuntimeImageCatalog(root string) (*runtimeImageCatalog, error) {
	data, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("read runtime image manifest: %w", err)
	}
	var manifest runtimeImageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decode runtime image manifest: %w", err)
	}
	if manifest.Version != 1 {
		return nil, fmt.Errorf("runtime image manifest version %d, want 1", manifest.Version)
	}
	return &runtimeImageCatalog{root: root, manifest: manifest}, nil
}

func (c *runtimeImageCatalog) image(record runtimeImageRecord) (*ebiten.Image, error) {
	path := filepath.Join(c.root, filepath.FromSlash(record.File))
	h, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoded, err := png.Decode(h)
	closeErr := h.Close()
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return ebiten.NewImageFromImage(decoded), nil
}

func (c *runtimeImageCatalog) images(records []runtimeImageRecord) ([]*ebiten.Image, error) {
	images := make([]*ebiten.Image, 0, len(records))
	for _, record := range records {
		image, err := c.image(record)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}

func (c *runtimeImageCatalog) maskedSymbolImages(records []runtimeImageRecord) ([]*ebiten.Image, error) {
	images := make([]*ebiten.Image, 0, len(records))
	for _, record := range records {
		path := filepath.Join(c.root, filepath.FromSlash(record.File))
		h, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		decoded, err := png.Decode(h)
		h.Close()
		if err != nil {
			return nil, err
		}
		bounds := decoded.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, decoded, bounds.Min, draw.Src)
		mask := color.RGBAModel.Convert(gfx.EGA16[13]).(color.RGBA)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				if color.RGBAModel.Convert(rgba.At(x, y)).(color.RGBA) == mask {
					rgba.SetRGBA(x, y, color.RGBA{})
				}
			}
		}
		images = append(images, ebiten.NewImageFromImage(rgba))
	}
	return images, nil
}

func (c *runtimeImageCatalog) symbolBlock(symbolFile string, block uint8, wantItems int) ([]*ebiten.Image, error) {
	source := strings.ToUpper(strings.TrimSuffix(symbolFile, filepath.Ext(symbolFile)))
	records := make([]runtimeImageRecord, 0, wantItems)
	for _, record := range c.manifest.Symbols {
		if record.Source == source && record.Block == block {
			records = append(records, record)
		}
	}
	if len(records) != wantItems {
		return nil, fmt.Errorf("PNG manifest %s block 0x%02X items=%d, want %d", source, block, len(records), wantItems)
	}
	return c.maskedSymbolImages(records)
}

func (c *runtimeImageCatalog) symbolBlockAtLeast(symbolFile string, block uint8, minimum int) ([]*ebiten.Image, error) {
	source := strings.ToUpper(strings.TrimSuffix(symbolFile, filepath.Ext(symbolFile)))
	var records []runtimeImageRecord
	for _, record := range c.manifest.Symbols {
		if record.Source == source && record.Block == block {
			records = append(records, record)
		}
	}
	if len(records) < minimum {
		return nil, fmt.Errorf("PNG manifest %s block 0x%02X items=%d, want at least %d", source, block, len(records), minimum)
	}
	return c.maskedSymbolImages(records)
}

func (c *runtimeImageCatalog) wallData(wallFile string, block uint8) ([]uint8, bool) {
	source := strings.ToUpper(strings.TrimSuffix(wallFile, filepath.Ext(wallFile)))
	for _, record := range c.manifest.Walls {
		if record.Source == source && record.Block == block {
			return append([]uint8(nil), record.Data...), true
		}
	}
	return nil, false
}

func (c *runtimeImageCatalog) wallSymbol(symbolFile string, block uint8, item int) (*ebiten.Image, bool) {
	source := strings.ToUpper(strings.TrimSuffix(symbolFile, filepath.Ext(symbolFile)))
	for _, record := range c.manifest.Symbols {
		if record.Source == source && record.Block == block && record.Item == item {
			images, err := c.maskedSymbolImages([]runtimeImageRecord{record})
			if err != nil || len(images) != 1 {
				return nil, false
			}
			return images[0], true
		}
	}
	return nil, false
}

func (c *runtimeImageCatalog) skyImages(skyFile string, blocks [3]uint8) ([3]*ebiten.Image, error) {
	var result [3]*ebiten.Image
	source := strings.ToUpper(strings.TrimSuffix(skyFile, filepath.Ext(skyFile)))
	for index, block := range blocks {
		found := false
		for _, record := range c.manifest.Sky {
			if record.Source == source && record.Block == block {
				image, err := c.image(record)
				if err != nil {
					return result, err
				}
				result[index] = image
				found = true
				break
			}
		}
		if !found {
			return result, fmt.Errorf("PNG manifest %s has no block 0x%02X", source, block)
		}
	}
	return result, nil
}
