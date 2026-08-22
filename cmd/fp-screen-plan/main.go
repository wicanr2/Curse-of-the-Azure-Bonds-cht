// Command fp-screen-plan 算出「第一人稱 fidelity 該比哪些畫面」的分母。
//
// ★ 存在的理由：原版截圖 oracle 早就建好了（`tools/dos-oracle-session.sh` 那六支，
// spec 1134），也已經比過提爾佛頓的 20 張。缺的不是工具而是**分母**——沒有一份
// 「該比哪些畫面」的清單，20 這個數字就沒有母數，UI fidelity 也就永遠是一句
// 「尚未逐張比對」。
//
// ★★ 分母**不是**「可走格子 × 四個朝向」。那是幾千張，而且大部分是重複的：
// 第一人稱畫面只由**視野內的牆面配置**決定，兩格的牆鄰域一樣就會畫出同一張圖，
// 比第二張證明不了任何新東西。所以分母取**相異的牆鄰域簽章**。
//
// 簽章取的是原作 `BUILDVIEW` 真正會讀到的那一塊：站在 `(x,y)` 朝某個方向時，
// 前方三排（深度 1..3）× 橫向 −1..+1 的九格，每格記它的左右牆與前方牆。
// 深度取 3 是因為原作的第一人稱只畫到第三排（spec 1130 的幾何與框）。
//
// ⚠ 這是**渲染器**的分母，不是內容的分母。同一個簽章在不同地圖會畫出同一張圖，
// 所以簽章相同的格子只需要比一張；反過來，簽章不同不代表玩家走得到——可達性
// 是另一件事，這支不管。
//
// ⚠ 也不宣稱「比完所有簽章 ＝ 第一人稱完全正確」：天空色、牆磚選圖、調色盤
// 都跟著地圖宣告走（spec 1134 的天花板就是這樣錯的），那些要各自比。
//
// 用法：
//
//	./tools/go.sh run ./cmd/fp-screen-plan -output docs/audit/fp-screen-plan.md
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

// viewDepth 是第一人稱畫得到的排數（spec 1130）。
const viewDepth = 3

// step 是四個朝向各自的前進向量。0＝北、2＝東、4＝南、6＝西。
var step = map[int][2]int{0: {0, -1}, 2: {1, 0}, 4: {0, 1}, 6: {-1, 0}}

// rightOf 是「朝這個方向時，右手邊」的方向。
var rightOf = map[int]int{0: 2, 2: 4, 4: 6, 6: 0}

// signature 把站在 `(x,y)` 朝 `facing` 時視野內的牆面配置壓成一個字串。
//
// ★ 第一人稱的正確與否有**兩個互相獨立的軸**，分母也要分開算：
//
//   - **幾何**：哪裡該畫一面牆、透視怎麼收斂。只跟「有沒有牆」有關，
//     跟那面牆貼哪一張圖無關 ⇒ 簽章取**二元的有無**。
//   - **貼圖**：那面牆該用哪一張牆磚。那是 `LOADWALLSET` 與地圖宣告的事，
//     跟站在哪一格無關 ⇒ 分母是**這張地圖用到的相異牆型值**。
//
// ⚠ 第一版把牆型值直接寫進簽章，結果 17,684 張畫面得到 15,693 個「相異」簽章
// ——去重幾乎沒有作用，因為牆型在每一格都不同。那個數字誠實但沒有用：它把
// 兩個軸混在一起，於是兩邊都量不到。
//
// ⚠ 第二版改成二元的有無，仍然有 14,294 個——因為它簽的是**原始鄰域**而不是
// **看得到的東西**。走廊裡一面牆擋住之後，後面那幾排根本不會畫出來，把它們簽
// 進去等於在區分兩張一模一樣的畫面。這一版做遮蔽：中央那一行走到前方有牆為止，
// 兩側只有側牆是開的那一格看得進去。
//
// ⚠ 用**環繞**取格（`WallWrapped`）：原作的地城是 16×16 環繞的，走出邊界會
// 繞回來，所以視野也會看到繞回來的那一側。
func signature(grid geo.Grid, x, y, facing int) string {
	forward := step[facing]
	right := step[rightOf[facing]]
	left := rightOf[rightOf[rightOf[facing]]]
	var builder strings.Builder
	present := func(cellX, cellY, direction int) bool {
		wall, ok := grid.WallWrapped(cellX, cellY, direction)
		return ok && wall != 0
	}
	mark := func(visible bool) {
		if visible {
			builder.WriteByte('#')
		} else {
			builder.WriteByte('.')
		}
	}
	// 中央那一行：一路往前，直到前方有牆為止。牆後面的東西畫不到。
	for depth := 0; depth <= viewDepth; depth++ {
		cellX := x + forward[0]*depth
		cellY := y + forward[1]*depth
		mark(present(cellX, cellY, left))
		mark(present(cellX, cellY, rightOf[facing]))
		blocked := present(cellX, cellY, facing)
		mark(blocked)
		if blocked {
			break
		}
		builder.WriteByte('|')
	}
	// 左右兩側：**只有側牆是開的那一格看得進去**。側牆一擋，那一側往後就全部
	// 不可見——這正是走廊裡大部分畫面長得一樣的原因。
	for _, lateral := range []int{-1, 1} {
		builder.WriteByte('/')
		for depth := 0; depth <= viewDepth; depth++ {
			centerX := x + forward[0]*depth
			centerY := y + forward[1]*depth
			side := rightOf[facing]
			if lateral < 0 {
				side = left
			}
			if present(centerX, centerY, side) {
				// 側面被擋住：這一格之後的側向都看不到了。
				mark(true)
				break
			}
			mark(false)
			cellX := centerX + right[0]*lateral
			cellY := centerY + right[1]*lateral
			mark(present(cellX, cellY, facing))
			if present(centerX, centerY, facing) {
				break
			}
		}
	}
	return builder.String()
}

// wallTypes 收集這張地圖用到的相異牆型值——貼圖那一軸的分母。
func wallTypes(grid geo.Grid) map[uint8]bool {
	out := map[uint8]bool{}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			for _, direction := range []int{0, 2, 4, 6} {
				if wall, ok := grid.WallWrapped(x, y, direction); ok && wall != 0 {
					out[wall] = true
				}
			}
		}
	}
	return out
}

// walkable 回答「這一格站得住嗎」。四面都是實牆的格子玩家進不去，畫面也就
// 不會出現，收進分母只會虛胖。
//
// ⚠ 判準是「至少有一個方向沒有牆」，不是「走得到」。真正的可達性要從入口做
// 連通性走訪，而那還要考慮腳本傳送——這支刻意不做，也不假裝有做。
func walkable(grid geo.Grid, x, y int) bool {
	for _, direction := range []int{0, 2, 4, 6} {
		if wall, ok := grid.WallWrapped(x, y, direction); ok && wall == 0 {
			return true
		}
	}
	return false
}

type mapPlan struct {
	ID         string
	Set        uint8
	Block      uint8
	Cells      int
	Screens    int
	Signatures int
	// NewHere 是「這張地圖第一次出現」的簽章數：先前的地圖已經涵蓋過的不算。
	NewHere int
	// WallTypes 是這張地圖用到的相異牆型值（貼圖那一軸）。
	WallTypes int
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	output := flag.String("output", "", "Markdown output path (empty prints to stdout)")
	flag.Parse()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	catalog := geo.NewCatalog()
	for set := 2; set <= 6; set++ {
		payload := memberPayload(&archive.Reader, fmt.Sprintf("GEO%d.DAX", set))
		if payload == nil {
			continue
		}
		if err := catalog.AddDAX(uint8(set), payload); err != nil {
			log.Fatalf("GEO%d.DAX: %v", set, err)
		}
	}
	if catalog.Len() == 0 {
		log.Fatal("no GEO block was parsed; check -image")
	}

	seen := map[string]bool{}
	allWallTypes := map[uint8]bool{}
	plans := make([]mapPlan, 0, 24)
	definitions := append([]goldenbox.MapDefinition(nil), pack.Maps...)
	sort.SliceStable(definitions, func(i, j int) bool {
		if definitions[i].AreaID != definitions[j].AreaID {
			return definitions[i].AreaID < definitions[j].AreaID
		}
		return definitions[i].GeometryBlock < definitions[j].GeometryBlock
	})
	for _, definition := range definitions {
		if definition.Kind != "first_person" {
			continue
		}
		grid, ok := catalog.Lookup(geo.MapRef{Set: definition.AreaID, BlockID: definition.GeometryBlock})
		if !ok {
			fmt.Fprintf(os.Stderr, "skip %s: GEO%d block %#02x not in catalog\n",
				definition.ID, definition.AreaID, definition.GeometryBlock)
			continue
		}
		plan := mapPlan{ID: definition.ID, Set: definition.AreaID, Block: definition.GeometryBlock}
		types := wallTypes(grid)
		plan.WallTypes = len(types)
		for value := range types {
			allWallTypes[value] = true
		}
		local := map[string]bool{}
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if !walkable(grid, x, y) {
					continue
				}
				plan.Cells++
				for _, facing := range []int{0, 2, 4, 6} {
					plan.Screens++
					key := signature(grid, x, y, facing)
					if !local[key] {
						local[key] = true
						plan.Signatures++
					}
					if !seen[key] {
						seen[key] = true
						plan.NewHere++
					}
				}
			}
		}
		plans = append(plans, plan)
	}

	totalCells, totalScreens := 0, 0
	for _, plan := range plans {
		totalCells += plan.Cells
		totalScreens += plan.Screens
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# First-person fidelity: which screens are worth comparing\n\n")
	fmt.Fprintf(&report, "Generated by `cmd/fp-screen-plan`; do not hand-edit. Rationale: spec 1185.\n\n")
	fmt.Fprintf(&report, "A first-person screen is determined by the wall layout inside the view, so two "+
		"cells with the same wall neighbourhood render the same picture and comparing the second one "+
		"proves nothing new. The denominator for renderer fidelity is therefore the number of **distinct "+
		"wall signatures**, not cells times facings.\n\n")
	fmt.Fprintf(&report, "The signature covers rows 1..%d ahead by lateral -1..+1, three walls per cell "+
		"(front, left, right), sampled with the original 16x16 wrap, recording only whether a wall is "+
		"**present**.\n\n", viewDepth)
	fmt.Fprintf(&report, "Presence and tile are two independent axes, so they get two denominators. "+
		"Geometry (where a wall is drawn, how the perspective converges) depends only on presence. "+
		"Which tile that wall uses is `LOADWALLSET` plus the map declaration and does not depend on "+
		"which cell you stand in, so its denominator is the set of distinct wall-type values.\n\n")
	fmt.Fprintf(&report, "An earlier version folded the wall-type value into the signature and got 15,693 "+
		"\"distinct\" signatures out of 17,684 screens -- deduplication did almost nothing, because the "+
		"type differs in nearly every cell. That number was honest and useless: mixing the two axes "+
		"measures neither.\n\n")
	fmt.Fprintf(&report, "Caveats: this is the denominator for the **renderer**, not for content, and not "+
		"for reachability -- a cell counts as standable when at least one side has no wall, which is not "+
		"the same as the party being able to walk there. Comparing every signature would still not prove "+
		"the first-person view is correct: sky colour, wall-tile selection and palette follow the map "+
		"declaration and need their own comparison (spec 1134's black ceiling was exactly that).\n\n")
	fmt.Fprintf(&report, "| total | value |\n|---|---:|\n")
	fmt.Fprintf(&report, "| first-person maps | %d |\n", len(plans))
	fmt.Fprintf(&report, "| standable cells | %d |\n", totalCells)
	fmt.Fprintf(&report, "| cells x facings | %d |\n", totalScreens)
	fmt.Fprintf(&report, "| **distinct wall-presence signatures (geometry)** | **%d** |\n", len(seen))
	fmt.Fprintf(&report, "| **distinct wall-type values (tiles)** | **%d** |\n\n", len(allWallTypes))
	fmt.Fprintf(&report, "| map | GEO | block | cells | screens | signatures | new here | wall types |\n|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, plan := range plans {
		fmt.Fprintf(&report, "| `%s` | %d | `0x%02X` | %d | %d | %d | %d | %d |\n",
			plan.ID, plan.Set, plan.Block, plan.Cells, plan.Screens, plan.Signatures,
			plan.NewHere, plan.WallTypes)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "maps=%d cells=%d screens=%d signatures=%d wall-types=%d\n",
		len(plans), totalCells, totalScreens, len(seen), len(allWallTypes))
}

func memberPayload(archive *zip.Reader, name string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer handle.Close()
		payload, readErr := io.ReadAll(handle)
		if readErr != nil {
			log.Fatal(readErr)
		}
		return payload
	}
	return nil
}
