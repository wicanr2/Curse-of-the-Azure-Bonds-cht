// Command geo-move-mask 印出每一格「四個方向走不走得動」的遮罩：
// bit0 北、bit1 東、bit2 南、bit3 西，1 代表走得動。
//
// ★ 存在的理由：要判「原版現在站在哪一張圖」，需要一把與**第一人稱算圖無關**
// 的尺。畫面比對是循環論證（拿 remake 畫的圖去比，等於用待測物當標準）；
// AREA 自動地圖只有 63.3% 對得上 GEO（spec 1185）。走得動與否是**另一個子系統**
// ——移動規則，而且原版那一側只要按一次方向鍵、把座標讀回來就量得到
// （`tools/dos-oracle-move-probe.sh`）。第一人稱算錯不會讓移動也跟著錯，
// 所以拿它當識別依據不構成循環。
//
// ⚠ 「有牆」不等於「走不動」：實測提爾佛頓 (7,13) 四面都有牆（牆型 04/06/04/05），
// 但**往西走得出去**——那一面是門。所以指紋要用移動規則算，不能用牆的有無。
//
// 用法：
//
//	go run ./cmd/geo-move-mask -set 2 -block 1
//	go run ./cmd/geo-move-mask -all -cells 7,13:3,3:12,9
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

func main() {
	set := flag.Int("set", 2, tooltext.Text("h.cfd9421a6aad"))
	block := flag.Int("block", 1, tooltext.Text("h.e3d3f40d66aa"))
	all := flag.Bool("all", false, tooltext.Text("h.39ee1d20af32"))
	cells := flag.String("cells", "", tooltext.Text("h.2c17ad557326"))
	imagePath := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.79f855c8b433"))
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

	var wanted [][2]int
	if *cells != "" {
		for _, item := range strings.Split(*cells, ":") {
			parts := strings.Split(item, ",")
			if len(parts) != 2 {
				log.Fatal(tooltext.Format("h.604f9ddffb26", item))
			}
			x, err1 := strconv.Atoi(parts[0])
			y, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				log.Fatal(tooltext.Format("h.604f9ddffb26", item))
			}
			wanted = append(wanted, [2]int{x, y})
		}
	}

	refs := []geo.MapRef{{Set: uint8(*set), BlockID: uint8(*block)}}
	if *all {
		refs = catalog.Refs()
	}
	for _, ref := range refs {
		grid, ok := catalog.Lookup(ref)
		if !ok {
			continue
		}
		if len(wanted) > 0 {
			parts := make([]string, 0, len(wanted))
			for _, cell := range wanted {
				parts = append(parts, fmt.Sprintf("(%d,%d)=%X", cell[0], cell[1], mask(grid, cell[0], cell[1])))
			}
			fmt.Print(tooltext.Format("h.65884ddead2e", ref.Set, ref.BlockID, strings.Join(parts, " ")))
			continue
		}
		fmt.Print(tooltext.Format("h.09c8a5a38e78", ref.Set, ref.BlockID))
		for y := 0; y < geo.Height; y++ {
			row := make([]string, 0, geo.Width)
			for x := 0; x < geo.Width; x++ {
				row = append(row, fmt.Sprintf("%X", mask(grid, x, y)))
			}
			fmt.Println(strings.Join(row, ""))
		}
	}
}

// mask 用引擎的地城移動規則算四個方向。方向編碼 0 北、2 東、4 南、6 西。
func mask(grid geo.Grid, x, y int) int {
	bits := 0
	for index, direction := range []int{0, 2, 4, 6} {
		if grid.CanMoveDungeonWrapped(x, y, direction) {
			bits |= 1 << index
		}
	}
	return bits
}
