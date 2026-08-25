// geo-probe 傾印一張原版 GEO 地圖的牆面與地形，並列出可以走動的格子。
//
// ★ 存在的理由：拿原版當第一人稱 oracle 時，得先知道「哪些格子走得到」。
// 開場那一格是密室（劇情要求先找出口），拿它當起點會誤以為方向鍵沒進去。
//
// 用法：
//
//	go run ./cmd/geo-probe -set 2 -block 1
//	go run ./cmd/geo-probe -set 2 -block 1 -open 12   # 只列開放度 >= 2 的格
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.5d993e050e0a"))
	set := flag.Int("set", 2, tooltext.Text("h.d23c7c76b5e0"))
	block := flag.Int("block", 1, tooltext.Text("h.b96a27801725"))
	minOpen := flag.Int("open", 2, tooltext.Text("h.285d53102035"))
	prefix := flag.Bool("prefix", false, tooltext.Text("h.082b2fe926e4"))
	flag.Parse()

	archive, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	name := fmt.Sprintf("GEO%d.DAX", *set)
	var payload []byte
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
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
		log.Fatal(tooltext.Format("h.f04182cf799c", name, *imagePath))
	}
	if *prefix {
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		for _, block := range blocks {
			if len(block.Data) < 2 {
				continue
			}
			fmt.Print(tooltext.Format("h.fd0ca3bcde27", *set, block.Entry.ID, block.Entry.ID, block.Data[0], block.Data[1], block.Data[0], block.Data[1], len(block.Data)))
		}
		return
	}
	catalog := geo.NewCatalog()
	if err := catalog.AddDAX(uint8(*set), payload); err != nil {
		log.Fatal(err)
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: uint8(*set), BlockID: uint8(*block)})
	if !ok {
		log.Fatal(tooltext.Format("h.a0f15ebb36d3", *set, *block))
	}

	// 方向編號與遊戲一致：0=北、2=東、4=南、6=西。
	dirs := []int{0, 2, 4, 6}
	letters := []string{"N", "E", "S", "W"}
	fmt.Print(tooltext.Format("h.1ae2ba52c9d5", *set, *block, geo.Width, geo.Height))
	fmt.Println(tooltext.Text("h.01f1a8d27227"))
	for y := 0; y < geo.Height; y++ {
		var row strings.Builder
		for x := 0; x < geo.Width; x++ {
			open := 0
			for _, d := range dirs {
				if grid.CanMoveDungeonWrapped(x, y, d) {
					open++
				}
			}
			switch open {
			case 0:
				row.WriteString("#")
			case 4:
				row.WriteString(".")
			default:
				row.WriteString(fmt.Sprintf("%d", open))
			}
		}
		fmt.Printf("%2d %s\n", y, row.String())
	}
	fmt.Print(tooltext.Format("h.050c2a99c800", *minOpen))
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			var open []string
			for i, d := range dirs {
				if grid.CanMoveDungeonWrapped(x, y, d) {
					open = append(open, letters[i])
				}
			}
			if len(open) < *minOpen {
				continue
			}
			cell := grid.CellWrapped(x, y)
			fmt.Print(tooltext.Format("h.dc42b523d5ec", x, y, cell.Terrain, cell.WallDirections, strings.Join(open, "")))
		}
	}
}
