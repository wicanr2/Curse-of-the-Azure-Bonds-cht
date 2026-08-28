// runtime-image-export 將執行期仍會直接解碼的原版圖像匯成獨立 PNG 與 manifest。
// 原版 ZIP 只在這個 export 步驟使用；remake loader 只消費輸出目錄。
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

type record struct {
	File   string `json:"file"`
	Source string `json:"source"`
	Block  uint8  `json:"block"`
	Item   int    `json:"item"`
}

type manifest struct {
	Version int                 `json:"version"`
	Tiles   []record            `json:"tiles"`
	Combat  map[string][]record `json:"combat"`
	Symbols []record            `json:"symbols"`
	Sky     []record            `json:"sky"`
	Walls   []wallRecord        `json:"walls"`
}

type wallRecord struct {
	Source string  `json:"source"`
	Block  uint8   `json:"block"`
	Data   []uint8 `json:"data"`
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: runtime-image-export ORIGINAL.zip OUTPUT-DIRECTORY")
	}
	reader, err := zip.OpenReader(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	out := os.Args[2]
	if err := os.MkdirAll(out, 0o755); err != nil {
		log.Fatal(err)
	}
	m := manifest{Version: 1, Combat: make(map[string][]record)}

	data := member(reader.File, "TILES.DAX")
	blocks, err := dax.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	for _, block := range blocks {
		picture, err := gfx.ParsePicture(block.Data, false, 0)
		if err != nil {
			log.Fatalf("TILES block %02X: %v", block.Entry.ID, err)
		}
		for item := 0; item < int(picture.ItemCount); item++ {
			rgba, err := picture.RGBA(item, gfx.EGA16)
			if err != nil {
				log.Fatal(err)
			}
			m.Tiles = append(m.Tiles, write(out, "tiles", "TILES", block.Entry.ID, item, rgba))
		}
	}

	for _, source := range []string{"DUNGCOM", "WILDCOM", "RANDCOM"} {
		blocks, err := dax.Parse(member(reader.File, source+".DAX"))
		if err != nil || len(blocks) != 1 {
			log.Fatalf("%s blocks=%d: %v", source, len(blocks), err)
		}
		set, err := gfx.ParseCombatTiles(blocks[0].Data)
		if err != nil {
			log.Fatal(err)
		}
		for item, tile := range set.Tiles {
			rgba, err := tile.RGBA(0, gfx.EGA16)
			if err != nil {
				log.Fatal(err)
			}
			m.Combat[source] = append(m.Combat[source], write(out, "combat", source, blocks[0].Entry.ID, item, rgba))
		}
	}

	for area := 1; area <= 6; area++ {
		source := fmt.Sprintf("8X8D%d", area)
		blocks, err := dax.Parse(member(reader.File, source+".DAX"))
		if err != nil {
			log.Fatal(err)
		}
		for _, block := range blocks {
			picture, err := gfx.ParsePicture(block.Data, false, 0)
			if err != nil || picture.Width() != 8 || picture.Height() != 8 {
				continue
			}
			for item := 0; item < int(picture.ItemCount); item++ {
				rgba, err := picture.RGBA(item, gfx.EGA16)
				if err != nil {
					log.Fatal(err)
				}
				m.Symbols = append(m.Symbols, write(out, "symbols", source, block.Entry.ID, item, rgba))
			}
		}
	}
	for area := 2; area <= 6; area++ {
		source := fmt.Sprintf("WALLDEF%d", area)
		blocks, err := dax.Parse(member(reader.File, source+".DAX"))
		if err != nil {
			log.Fatal(err)
		}
		for _, block := range blocks {
			if _, err := gfx.ParseWallDefs(block.Data); err != nil {
				log.Fatalf("%s block %02X: %v", source, block.Entry.ID, err)
			}
			m.Walls = append(m.Walls, wallRecord{Source: source, Block: block.Entry.ID, Data: append([]uint8(nil), block.Data...)})
		}
	}

	blocks, err = dax.Parse(member(reader.File, "SKY.DAX"))
	if err != nil {
		log.Fatal(err)
	}
	for _, block := range blocks {
		picture, err := gfx.ParsePicture(block.Data, true, 13)
		if err != nil || picture.ItemCount == 0 {
			continue
		}
		rgba, err := picture.RGBA(0, gfx.EGA16)
		if err != nil {
			log.Fatal(err)
		}
		m.Sky = append(m.Sky, write(out, "sky", "SKY", block.Entry.ID, 0, rgba))
	}
	sort.Slice(m.Symbols, func(i, j int) bool {
		if m.Symbols[i].Source != m.Symbols[j].Source {
			return m.Symbols[i].Source < m.Symbols[j].Source
		}
		if m.Symbols[i].Block != m.Symbols[j].Block {
			return m.Symbols[i].Block < m.Symbols[j].Block
		}
		return m.Symbols[i].Item < m.Symbols[j].Item
	})
	handle, err := os.Create(filepath.Join(out, "manifest.json"))
	if err != nil {
		log.Fatal(err)
	}
	encoder := json.NewEncoder(handle)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		handle.Close()
		log.Fatal(err)
	}
	if err := handle.Close(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tiles=%d combat=%d symbols=%d sky=%d walls=%d\n", len(m.Tiles), len(m.Combat["DUNGCOM"])+len(m.Combat["WILDCOM"])+len(m.Combat["RANDCOM"]), len(m.Symbols), len(m.Sky), len(m.Walls))
}

func member(files []*zip.File, name string) []byte {
	for _, file := range files {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		h, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		data, err := io.ReadAll(h)
		h.Close()
		if err != nil {
			log.Fatal(err)
		}
		return data
	}
	log.Fatalf("original image has no %s", name)
	return nil
}

func write(root, kind, source string, block uint8, item int, rgba image.Image) record {
	rel := filepath.ToSlash(filepath.Join(kind, fmt.Sprintf("%s-block-%02X-item-%03d.png", strings.ToLower(source), block, item)))
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatal(err)
	}
	h, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(h, rgba); err != nil {
		h.Close()
		log.Fatal(err)
	}
	if err := h.Close(); err != nil {
		log.Fatal(err)
	}
	return record{File: rel, Source: source, Block: block, Item: item}
}
