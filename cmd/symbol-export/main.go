// symbol-export 把某個 `8X8D*.DAX` 的每個區塊、每個 8×8 符號匯成一張總覽 PNG。
//
// ★ 存在的理由是「用資料回答，不要用印象回答」。第一次用它是為了回答
// 「第一人稱視野裡那些沒被畫出來的牆磚是什麼」：`WALLDEF` 的符號編號把
// 0..255 切成四段（1..45、46..115、116..185、186..255），前三段對應
// `LOAD PIECES` 選到的三個區塊，**第一段 1..45 是另一個區塊**——先前只載了
// 後三段，於是所有低編號的磚（牆的斜角收邊、天空、地板）全部沒有畫出來。
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	member := flag.String("member", "8X8D2.DAX", "DAX member to export")
	out := flag.String("out", "", "output directory")
	scale := flag.Int("scale", 4, "nearest-neighbour scale for the contact sheet")
	columns := flag.Int("columns", 16, "symbols per contact-sheet row")
	mask := flag.Bool("mask", true, "treat EGA index 13 as transparent")
	flag.Parse()
	if *out == "" {
		log.Fatal("-out is required")
	}
	reader, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	var data []byte
	for _, file := range reader.File {
		if !strings.EqualFold(file.Name, *member) {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		data, err = io.ReadAll(handle)
		handle.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
	if len(data) == 0 {
		log.Fatalf("%s has no member %s", *imagePath, *member)
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	for _, block := range blocks {
		picture, err := gfx.ParsePicture(block.Data, *mask, 13)
		if err != nil {
			fmt.Printf("block 0x%02X: %v\n", block.Entry.ID, err)
			continue
		}
		count := int(picture.ItemCount)
		fmt.Printf("%s block 0x%02X %dx%d items=%d\n",
			*member, block.Entry.ID, picture.Width(), picture.Height(), count)
		rows := (count + *columns - 1) / *columns
		cell := picture.Width() * *scale
		sheet := image.NewRGBA(image.Rect(0, 0, *columns*cell, rows*picture.Height()**scale))
		for item := 0; item < count; item++ {
			rgba, err := picture.RGBA(item, gfx.EGA16)
			if err != nil {
				log.Fatal(err)
			}
			originX := (item % *columns) * cell
			originY := (item / *columns) * picture.Height() * *scale
			for y := 0; y < picture.Height()**scale; y++ {
				for x := 0; x < cell; x++ {
					sheet.Set(originX+x, originY+y, rgba.At(x/(*scale), y/(*scale)))
				}
			}
		}
		draw.Draw(sheet, sheet.Bounds(), sheet, image.Point{}, draw.Src)
		name := fmt.Sprintf("%s-block-%02X.png",
			strings.TrimSuffix(strings.ToUpper(*member), ".DAX"), block.Entry.ID)
		handle, err := os.Create(filepath.Join(*out, name))
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(handle, sheet); err != nil {
			log.Fatal(err)
		}
		handle.Close()
	}
}
