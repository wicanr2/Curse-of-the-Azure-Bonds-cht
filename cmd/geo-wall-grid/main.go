// Command geo-wall-grid 把一張第一人稱地圖的**牆面有無**印成 16×16 的十六進位
// 指紋，每格一個 nibble：北 1、東 2、南 4、西 8。
//
// ★ 存在的理由：要判「原版現在站在哪一張圖」，得有一份與 remake 的算圖完全無關
// 的參考。這一支只讀原版的 `GEO*.DAX`，不碰任何算圖程式碼——「拿 remake 畫的
// 圖去比，像就算數」是循環論證，會把「量測 remake 對不對」變成自我證明
// （spec 1185）。
//
// 用法：
//
//	go run ./cmd/geo-wall-grid -set 2 -block 1
//	go run ./cmd/geo-wall-grid -all
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
	set := flag.Int("set", 2, "GEO 檔集（1..6）")
	block := flag.Int("block", 1, "區塊編號")
	all := flag.Bool("all", false, "列出目錄裡每一張圖的指紋")
	types := flag.Bool("types", false, "改印每一面牆的**牆型值**（兩位十六進位，四面以 / 分隔）")
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "原版 image ZIP")
	flag.Parse()

	archive, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	catalog := geo.NewCatalog()
	for member := 1; member <= 6; member++ {
		name := fmt.Sprintf("GEO%d.DAX", member)
		for _, file := range archive.File {
			if !strings.EqualFold(file.Name, name) {
				continue
			}
			handle, openErr := file.Open()
			if openErr != nil {
				log.Fatal(openErr)
			}
			payload, readErr := io.ReadAll(handle)
			handle.Close()
			if readErr != nil {
				log.Fatal(readErr)
			}
			if addErr := catalog.AddDAX(uint8(member), payload); addErr != nil {
				log.Printf("%s：%v", name, addErr)
			}
		}
	}

	if *all {
		for _, ref := range catalog.Refs() {
			printGrid(catalog, ref, *types)
		}
		return
	}
	printGrid(catalog, geo.MapRef{Set: uint8(*set), BlockID: uint8(*block)}, *types)
}

func printGrid(catalog geo.Catalog, ref geo.MapRef, types bool) {
	grid, ok := catalog.Lookup(ref)
	if !ok {
		log.Fatalf("目錄裡沒有 GEO%d 段 0x%02X", ref.Set, ref.BlockID)
	}
	fmt.Printf("== GEO%d 段 0x%02X ==\n", ref.Set, ref.BlockID)
	if types {
		// ⚠ 牆「有沒有」與「擋不擋得住」是兩件事：實測提爾佛頓 (7,13) 四面都有牆，
		// 但往西走得出去——那一面是門。要用走得動與否當地圖指紋，得看牆型值。
		for y := 0; y < geo.Height; y++ {
			cells := make([]string, 0, geo.Width)
			for x := 0; x < geo.Width; x++ {
				sides := make([]string, 0, 4)
				for _, direction := range []int{0, 2, 4, 6} {
					wall, _ := grid.WallWrapped(x, y, direction)
					sides = append(sides, fmt.Sprintf("%02X", wall))
				}
				cells = append(cells, strings.Join(sides, "/"))
			}
			fmt.Printf("y=%2d %s\n", y, strings.Join(cells, " "))
		}
		return
	}
	for y := 0; y < geo.Height; y++ {
		row := make([]string, 0, geo.Width)
		for x := 0; x < geo.Width; x++ {
			bits := 0
			// 方向編碼與引擎相同：0 北、2 東、4 南、6 西。
			for index, direction := range []int{0, 2, 4, 6} {
				if wall, found := grid.WallWrapped(x, y, direction); found && wall != 0 {
					bits |= 1 << index
				}
			}
			row = append(row, fmt.Sprintf("%X", bits))
		}
		fmt.Println(strings.Join(row, ""))
	}
}
