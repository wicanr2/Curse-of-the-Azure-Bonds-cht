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
// ★★★ `-covered` 把已經擷取的原版畫面換算成簽章，於是分母有了**分子**：
// `docs/reference/original-dos/first-person/index.tsv` 那 20 張蓋掉幾種、還剩幾種。
// 它同時列出**接下來該拍哪幾格**——照「這一格能一次蓋掉幾個還沒蓋到的簽章」排序，
// 那就是 oracle session 的工作清單。
//
// 用法：
//
//	./tools/go.sh run ./cmd/fp-screen-plan -output docs/audit/fp-screen-plan.md \
//	  -covered docs/reference/original-dos/first-person/index.tsv
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

// sealed 回答「這一格四面都是實牆嗎」。
//
// ⚠ **四面都是牆不代表玩家不會站在那裡。** 第一版拿它當過濾器把那種格子排除在
// 分母外，而開新遊戲的起點 `GEO2/0x01 (7,13)` 正是這種格子——劇情要求先找出口，
// 所以它是密室，但它是**每個玩家看到的第一張畫面**。`-covered` 因此有四張擷取
// 對不到任何格子，工具照實回報，那個矛盾才浮出來。
//
// 現在分母收**全部 256 格**；這一支只用來回報「其中有幾格是密室」，讓讀的人
// 知道分母的形狀，而不是拿來過濾。
//
// ⚠ 真正的「玩家站得到哪些格」要從入口做連通性走訪，還要考慮腳本傳送
// （spec 1183：進場落點是腳本自己寫的）。這支刻意不做，也不假裝有做。
func sealed(grid geo.Grid, x, y int) bool {
	for _, direction := range []int{0, 2, 4, 6} {
		if wall, ok := grid.WallWrapped(x, y, direction); ok && wall == 0 {
			return false
		}
	}
	return true
}

// coveredEntry 是一張已經擷取的原版畫面。
type coveredEntry struct {
	Set    uint8
	Block  uint8
	X, Y   int
	Facing int
}

// referenceName 從檔名認出 GEO 檔號與區塊：`geo2-b01-x07-y13-N.png`。
// ⚠ 索引檔本身沒有這兩欄，而**猜錯會讓簽章整批算在錯的地圖上**，看起來像
// 「已經蓋掉很多」。認不出來就整列跳過並回報，不要預設成某一張圖。
var referenceName = regexp.MustCompile(`^geo(\d+)-b([0-9a-fA-F]+)-`)

var facingLetter = map[string]int{"N": 0, "E": 2, "S": 4, "W": 6}

func readCovered(path string) ([]coveredEntry, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := make([]coveredEntry, 0, 32)
	for _, line := range strings.Split(string(payload), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		match := referenceName.FindStringSubmatch(parts[0])
		if match == nil {
			fmt.Fprintf(os.Stderr, "skip %s: filename does not name a GEO set/block\n", parts[0])
			continue
		}
		set, _ := strconv.ParseUint(match[1], 10, 8)
		block, _ := strconv.ParseUint(match[2], 16, 8)
		x, errX := strconv.Atoi(parts[1])
		y, errY := strconv.Atoi(parts[2])
		facing, ok := facingLetter[strings.ToUpper(parts[3])]
		if errX != nil || errY != nil || !ok {
			fmt.Fprintf(os.Stderr, "skip %s: cannot parse %q\n", parts[0], line)
			continue
		}
		out = append(out, coveredEntry{Set: uint8(set), Block: uint8(block), X: x, Y: y, Facing: facing})
	}
	return out, nil
}

// placement 是「哪一張圖的哪一格朝哪個方向」。朝向留空時代表整格（四個朝向）。
type placement struct {
	mapID  string
	set    uint8
	block  uint8
	x, y   int
	facing int
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
	// Sealed 是四面都是實牆的格子數。**不是**用來過濾的，只是讓分母的形狀
	// 看得見——開新遊戲的起點就是這種格子。
	Sealed int
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	output := flag.String("output", "", "Markdown output path (empty prints to stdout)")
	covered := flag.String("covered", "", "index.tsv of already-captured original screens; adds the numerator")
	only := flag.String("only", "", "restrict the greedy capture plan to map IDs containing this substring")
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
	// signatureAt 記「哪一張圖的哪一格朝哪個方向是這個簽章」，`-covered` 要靠它
	// 把擷取換算成簽章，也要靠它排出「接下來拍哪一格最划算」。
	signatureAt := map[string][]placement{}
	byPlacement := map[placement]string{}
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
				plan.Cells++
				if sealed(grid, x, y) {
					plan.Sealed++
				}
				for _, facing := range []int{0, 2, 4, 6} {
					plan.Screens++
					key := signature(grid, x, y, facing)
					spot := placement{mapID: definition.ID, set: definition.AreaID,
						block: definition.GeometryBlock, x: x, y: y, facing: facing}
					signatureAt[key] = append(signatureAt[key], spot)
					byPlacement[spot] = key
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

	totalCells, totalScreens, totalSealed := 0, 0, 0
	for _, plan := range plans {
		totalCells += plan.Cells
		totalScreens += plan.Screens
		totalSealed += plan.Sealed
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
		"for reachability. Every cell is in the denominator, including the ones sealed on all four sides: "+
		"the new-game start cell is exactly that (the story requires finding the exit) and it is the first "+
		"screen every player sees. Which cells the party can actually reach needs a connectivity walk plus "+
		"script placement, and this tool does not do that. Comparing every signature would still not prove "+
		"the first-person view is correct: sky colour, wall-tile selection and palette follow the map "+
		"declaration and need their own comparison (spec 1134's black ceiling was exactly that).\n\n")
	fmt.Fprintf(&report, "| total | value |\n|---|---:|\n")
	fmt.Fprintf(&report, "| first-person maps | %d |\n", len(plans))
	fmt.Fprintf(&report, "| cells | %d |\n", totalCells)
	fmt.Fprintf(&report, "| of which sealed on all four sides | %d |\n", totalSealed)
	fmt.Fprintf(&report, "| cells x facings | %d |\n", totalScreens)
	fmt.Fprintf(&report, "| **distinct wall-presence signatures (geometry)** | **%d** |\n", len(seen))
	fmt.Fprintf(&report, "| **distinct wall-type values (tiles)** | **%d** |\n\n", len(allWallTypes))
	fmt.Fprintf(&report, "| map | GEO | block | cells | sealed | screens | signatures | new here | wall types |\n|---|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, plan := range plans {
		fmt.Fprintf(&report, "| `%s` | %d | `0x%02X` | %d | %d | %d | %d | %d | %d |\n",
			plan.ID, plan.Set, plan.Block, plan.Cells, plan.Sealed, plan.Screens,
			plan.Signatures, plan.NewHere, plan.WallTypes)
	}

	if *covered != "" {
		entries, coverErr := readCovered(*covered)
		if coverErr != nil {
			log.Fatal(coverErr)
		}
		hit := map[string]bool{}
		unresolved := 0
		for _, entry := range entries {
			spot := placement{}
			found := false
			// 索引只給 GEO 檔號與區塊，而同一個區塊可能被多張地圖宣告
			// （`geo3/0x11` 就是兩層共用），簽章相同所以取第一個對得上的。
			for candidate, key := range byPlacement {
				if candidate.set == entry.Set && candidate.block == entry.Block &&
					candidate.x == entry.X && candidate.y == entry.Y &&
					candidate.facing == entry.Facing {
					spot, found = candidate, true
					hit[key] = true
					break
				}
			}
			if !found {
				unresolved++
				fmt.Fprintf(os.Stderr, "unresolved capture: GEO%d block %#02x (%d,%d) facing %d\n",
					entry.Set, entry.Block, entry.X, entry.Y, entry.Facing)
			}
			_ = spot
		}
		fmt.Fprintf(&report, "\n## Coverage\n\n")
		fmt.Fprintf(&report, "| item | value |\n|---|---:|\n")
		fmt.Fprintf(&report, "| captured screens in `%s` | %d |\n", filepath.Base(*covered), len(entries))
		fmt.Fprintf(&report, "| captures that map to no standable cell | %d |\n", unresolved)
		fmt.Fprintf(&report, "| **distinct signatures covered** | **%d / %d** |\n\n", len(hit), len(seen))
		if unresolved > 0 {
			fmt.Fprintf(&report, "A capture resolves to nothing when its cell is sealed on all four sides, "+
				"which means either the index or the GEO block is wrong. It is reported rather than "+
				"silently dropped.\n\n")
		}

		// 接下來拍哪一格：貪婪法，每次挑「一次能蓋掉最多還沒蓋到的簽章」的格子。
		// 一格拍四個朝向，所以單位是格不是畫面。
		type candidate struct {
			spot placement
			gain int
			keys []string
		}
		remaining := map[string]bool{}
		for key := range seen {
			if !hit[key] {
				remaining[key] = true
			}
		}
		cells := map[placement][]string{}
		for spot, key := range byPlacement {
			cell := placement{mapID: spot.mapID, set: spot.set, block: spot.block, x: spot.x, y: spot.y}
			cells[cell] = append(cells[cell], key)
		}
		// -only 把貪婪清單限制在**取得到畫面的那張圖**上。跨圖取 oracle 目前
		// 走不通（原版的第一人稱地圖由存檔裡的 ECL 狀態決定，不是由存檔的地圖
		// 編號決定，第 675 輪實測），所以「下一格拍哪裡」如果只按全域增益排序，
		// 排在前面的每一格都拍不到，等於沒有可執行的下一步。
		if *only != "" {
			for cell := range cells {
				if !strings.Contains(cell.mapID, *only) {
					delete(cells, cell)
				}
			}
		}
		picks := make([]candidate, 0, 64)
		for len(remaining) > 0 && len(picks) < 64 {
			best := candidate{}
			for cell, keys := range cells {
				gain, fresh := 0, make([]string, 0, 4)
				for _, key := range keys {
					if remaining[key] {
						gain++
						fresh = append(fresh, key)
					}
				}
				if gain > best.gain ||
					(gain == best.gain && gain > 0 && lessPlacement(cell, best.spot)) {
					best = candidate{spot: cell, gain: gain, keys: fresh}
				}
			}
			if best.gain == 0 {
				break
			}
			for _, key := range best.keys {
				delete(remaining, key)
			}
			picks = append(picks, best)
		}
		fmt.Fprintf(&report, "### Next cells worth capturing\n\n")
		fmt.Fprintf(&report, "Greedy order: each row is the cell that knocks out the most still-uncovered "+
			"signatures. One capture session covers four facings, so the unit is a cell, not a screen.\n\n")
		fmt.Fprintf(&report, "| # | map | GEO | block | cell | new signatures | still uncovered after |\n")
		fmt.Fprintf(&report, "|---:|---|---:|---:|---|---:|---:|\n")
		left := len(seen) - len(hit)
		for index, pick := range picks {
			left -= pick.gain
			fmt.Fprintf(&report, "| %d | `%s` | %d | `0x%02X` | `(%d,%d)` | %d | %d |\n",
				index+1, pick.spot.mapID, pick.spot.set, pick.spot.block,
				pick.spot.x, pick.spot.y, pick.gain, left)
		}
		fmt.Fprintf(os.Stderr, "covered=%d/%d unresolved=%d\n", len(hit), len(seen), unresolved)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "maps=%d cells=%d sealed=%d screens=%d signatures=%d wall-types=%d\n",
		len(plans), totalCells, totalSealed, totalScreens, len(seen), len(allWallTypes))
}

// lessPlacement 讓貪婪法在同分時有穩定的順序——否則同一份輸入每次跑出來的
// 「接下來拍哪格」都不一樣，那張表就沒辦法當工作清單引用。
func lessPlacement(a, b placement) bool {
	if a.set != b.set {
		return a.set < b.set
	}
	if a.block != b.block {
		return a.block < b.block
	}
	if a.y != b.y {
		return a.y < b.y
	}
	return a.x < b.x
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
