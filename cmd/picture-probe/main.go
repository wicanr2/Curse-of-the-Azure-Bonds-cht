// picture-probe 把某個 DAX 區塊當成一張 masked picture 解出來，印出尺寸與
// 指定像素的調色盤索引，需要時一併輸出 PNG。
//
// ★ 存在的理由：逐格比對原版畫面時，差異會落在「某一張圖的某一格」上
// （spec 1134 §七 就是這樣一格）。要判斷那是解碼錯還是遮罩錯，得看得到
// **索引**，不是看 RGB——遮罩鍵是索引 13，轉成 RGB 之後就分不出來了。
//
// 用法：
//
//	go run ./cmd/picture-probe -member SKY.DAX -block 252 -x 64 -y 6
//	go run ./cmd/picture-probe -member SKY.DAX -block 252 -png workplace/sky-fc.png
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.5d993e050e0a"))
	member := flag.String("member", "SKY.DAX", tooltext.Text("h.ad0d12f74987"))
	block := flag.Int("block", 252, tooltext.Text("h.e3d3f40d66aa"))
	item := flag.Int("item", 0, tooltext.Text("h.7f351c62aa76"))
	maskIndex := flag.Int("mask", 13, tooltext.Text("h.846b8c0fbbce"))
	probeX := flag.Int("x", -1, tooltext.Text("h.cc4562afdad5"))
	probeY := flag.Int("y", -1, tooltext.Text("h.29e36b86b08f"))
	radius := flag.Int("radius", 3, tooltext.Text("h.6d4b02329cb9"))
	pngPath := flag.String("png", "", tooltext.Text("h.6d8e6733e5d6"))
	flag.Parse()

	archive, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	var payload []byte
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, *member) {
			continue
		}
		handle, openErr := file.Open()
		if openErr != nil {
			log.Fatal(openErr)
		}
		payload, err = io.ReadAll(handle)
		handle.Close()
		if err != nil {
			log.Fatal(err)
		}
		break
	}
	if payload == nil {
		log.Fatal(tooltext.Format("h.f04182cf799c", *member, *imagePath))
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		log.Fatal(err)
	}
	var data []byte
	for _, entry := range blocks {
		if int(entry.Entry.ID) == *block {
			data = entry.Data
			break
		}
	}
	if data == nil {
		log.Fatal(tooltext.Format("h.99998dd4110c", *member, *block))
	}
	masked := *maskIndex >= 0
	key := uint8(0)
	if masked {
		key = uint8(*maskIndex)
	}
	picture, err := gfx.ParsePicture(data, masked, key)
	if err != nil {
		log.Fatal(err)
	}
	width := int(picture.WidthUnits) * 8
	height := int(picture.HeightUnits)
	fmt.Print(tooltext.Format("h.fde27ae385c5", *member, *block, width, height, picture.WidthUnits, picture.HeightUnits, picture.ItemCount, len(picture.Pixels)))

	if *probeX >= 0 && *probeY >= 0 {
		fmt.Print(tooltext.Format("h.23536a7aa40a", *probeX, *probeY))
		for y := *probeY - *radius; y <= *probeY+*radius; y++ {
			if y < 0 || y >= height {
				continue
			}
			var row strings.Builder
			for x := *probeX - *radius; x <= *probeX+*radius; x++ {
				if x < 0 || x >= width {
					row.WriteString("   ")
					continue
				}
				offset := (*item)*width*height + y*width + x
				if offset >= len(picture.Pixels) {
					row.WriteString(" ??")
					continue
				}
				fmt.Fprintf(&row, "%3X", picture.Pixels[offset])
			}
			marker := ""
			if y == *probeY {
				marker = "  <<<"
			}
			fmt.Printf("  y=%3d %s%s\n", y, row.String(), marker)
		}
	}

	if *pngPath != "" {
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				offset := (*item)*width*height + y*width + x
				if offset >= len(picture.Pixels) {
					continue
				}
				index := picture.Pixels[offset]
				if int(index) >= len(gfx.EGA16) {
					img.Set(x, y, color.RGBA{255, 0, 255, 255})
					continue
				}
				img.Set(x, y, gfx.EGA16[index])
			}
		}
		handle, createErr := os.Create(*pngPath)
		if createErr != nil {
			log.Fatal(createErr)
		}
		defer handle.Close()
		if err := png.Encode(handle, img); err != nil {
			log.Fatal(err)
		}
		fmt.Print(tooltext.Format("h.d77032aa340c", *pngPath))
	}
}
