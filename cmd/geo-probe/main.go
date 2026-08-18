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
	"io"
	"log"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "原版 DOS image ZIP")
	set := flag.Int("set", 2, "GEO 檔號（2..6）")
	block := flag.Int("block", 1, "GEO 區塊編號")
	minOpen := flag.Int("open", 2, "只列出至少有這麼多個可走方向的格子")
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
		log.Fatalf("%s 不在 %s 裡", name, *imagePath)
	}
	catalog := geo.NewCatalog()
	if err := catalog.AddDAX(uint8(*set), payload); err != nil {
		log.Fatal(err)
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: uint8(*set), BlockID: uint8(*block)})
	if !ok {
		log.Fatalf("GEO%d 沒有區塊 0x%02X", *set, *block)
	}

	// 方向編號與遊戲一致：0=北、2=東、4=南、6=西。
	dirs := []int{0, 2, 4, 6}
	letters := []string{"N", "E", "S", "W"}
	fmt.Printf("GEO%d 區塊 0x%02X（%d×%d）\n", *set, *block, geo.Width, geo.Height)
	fmt.Println("地圖（. 四面全通、# 全封、數字為可走方向數）：")
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
	fmt.Printf("\n可走方向 >= %d 的格子：\n", *minOpen)
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
			fmt.Printf("  (%2d,%2d) terrain=0x%02X 牆=%v 通=%s\n",
				x, y, cell.Terrain, cell.WallDirections, strings.Join(open, ""))
		}
	}
}
