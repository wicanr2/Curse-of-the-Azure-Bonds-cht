// Command dungeon-walk-probe 回答「從這一段的入口**用走的**，走得到哪些格子」。
//
// ★ 為什麼需要它：可達性報表（`cmd/cell-reachability`）說主線實跑只踏到 250 個
// 分派索引裡的 81 個，其中提爾佛頓整整三段（48 個索引）一格都沒踏到。
// 但那句話有兩種完全不同的成因：
//
//	(a) 主線不經過那裡      → 是路線的選擇，不是缺陷
//	(b) 那些格子走不進去    → 是缺陷，而且玩家會撞到
//
// 逐格實測分不出來——它把隊伍**直接放**到目標格上。這一支從入口出發，
// 只用 `CanMoveDungeon`／`MoveDungeon` 做廣度優先，記下**真的走得到**的格子。
//
// ⚠ 這仍然不是「從新遊戲玩到那裡」：進段是直接進的。它答的是**段內**可達性，
// 也就是「人已經在這張地圖上，走得到幾格」。段與段之間怎麼串是主線測試的事。
//
// ⚠ 走的時候會踩到事件、戰鬥、被劇情推走。推不回地城就停在那一條分支上——
// 所以結果是**下界**：走不到不代表不可達，可能只是被某個事件擋在半路。
//
// 用法：
//
//	# 先錄主線在每一段的到達取樣（門要劇情旗標才開得了）
//	rm -rf workplace/campaign-frames/arrivals
//	mkdir -p workplace/campaign-frames/arrivals
//	COAB_ARRIVAL_SNAPSHOT_DIR=/src/workplace/campaign-frames/arrivals \
//	    tools/go.sh test ./internal/game/ -count=1 \
//	    -run 'TestRealNewGameRunsToTheEnding|TestTilvertonRouteIsWalkableAndLocalized'
//	# 主線的路線紀錄（隊伍真的站過哪些格）
//	COAB_DECISION_LOG=/src/workplace/campaign-frames/route.json \
//	    tools/go.sh test ./internal/game/ -run TestRealNewGameRunsToTheEnding -count=1
//	go run ./cmd/dungeon-walk-probe -arrivals workplace/campaign-frames/arrivals \
//	    -route workplace/campaign-frames/route.json \
//	    -output docs/audit/dungeon-walk.md
//
// ⚠ **兩個旗標都不給會低估很多**，而且都不會報錯：
//
//	沒有            220 個分派索引，可達性聯集 214
//	-arrivals       225，聯集 224   （門要劇情旗標才開得了）
//	＋ -route       256，聯集 233   （樓梯／傳送才進得去的連通分量）
//
// 少給不會有錯誤訊息，只會少幾格——**而少幾格和「那裡本來就走不到」長得一模一樣**
// （spec 1193／1195）。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gamecorpus"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

type point struct{ x, y int }

// loadRouteStarts 把主線路線紀錄裡的移動起點按 ECL 段分組（去重）。
//
// ⚠ 記的是**起點**：終點由下一步的起點涵蓋，最後一步的終點會漏掉（下界）。
func loadRouteStarts(path string) map[uint8][]point {
	out := map[uint8][]point{}
	if path == "" {
		return out
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	var route []struct {
		Kind    string `json:"kind"`
		Segment int    `json:"segment"`
		FromX   int    `json:"from_x"`
		FromY   int    `json:"from_y"`
	}
	if json.Unmarshal(raw, &route) != nil {
		return out
	}
	seen := map[[3]int]bool{}
	for _, step := range route {
		if step.Kind != "move" {
			continue
		}
		key := [3]int{step.Segment, step.FromX, step.FromY}
		if seen[key] {
			continue
		}
		seen[key] = true
		out[uint8(step.Segment)] = append(out[uint8(step.Segment)], point{step.FromX, step.FromY})
	}
	return out
}

// dungeonDelta 是原作的八向朝向碼換成格子位移。只走正交四向：
// 斜向在原作的地城裡不是移動方向。
func dungeonDelta(direction int) (int, int) {
	switch direction {
	case 0:
		return 0, -1
	case 2:
		return 1, 0
	case 4:
		return 0, 1
	case 6:
		return -1, 0
	}
	return 0, 0
}

type segmentWalk struct {
	id       string
	block    uint8
	note     string
	reached  map[int]bool
	cells    int
	blocked  int
	// teleports 是「踏上去之後被事件搬到別處」的次數——樓梯與傳送就是這樣進到
	// 別的連通分量的。
	teleports int
	// fromRoute 是「從主線站過的格子起走」跑了幾趟。0 代表那一段第一輪就走完了
	// （或根本沒有路線紀錄）。
	fromRoute int
	// doorsOpened 是這一段開了幾扇門。⚠ 開門是**盤點**用的：直接解鎖、不擲骰。
	doorsOpened int
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "遊戲 image")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "語系檔")
	steps := flag.Int("steps", 4000, "每一段最多走幾步")
	cellsOut := flag.String("cells-json", "", "把走得到的 (block, 地形碼) 寫成 JSON")
	// ★ 「走得到的」與「圖上有的」是兩件事。沒有後者就分不出「走不進去」與
	// **「這個地形碼在這張圖上根本不存在」**——後者不論誰來走都踏不到，
	// 把它算進未達成會讓覆蓋率永遠差一截，而且看起來像還有事情可做。
	onMapOut := flag.String("on-map-json", "", "把**圖上出現過**的 (block, 地形碼) 寫成 JSON")
	// ★ 「站得上去的」比「從入口走得到的」多：地圖常常分成好幾塊互不相連的區域
	// （巫師塔每一層都是獨立房間）。這一份是**走路能到的上限**。
	componentsOut := flag.String("walkable-json", "", "把**站得上去**（任一連通分量內）的 (block, 地形碼) 寫成 JSON")
	routePath := flag.String("route", "",
		"主線路線紀錄（`COAB_DECISION_LOG`）：從隊伍真的站過的格子再走一次，"+
			"補「幾何上斷開、樓梯／傳送才進得去」那一類缺口")
	openDoors := flag.Bool("open-doors", true,
		"撞到門就開（**盤點用：直接解鎖不擲骰**）——門在 `CanMoveDungeon` 眼裡是牆")
	arrivals := flag.String("arrivals", "",
		"到達取樣目錄：冷走前先把主線在那一段的劇情旗標鋪上（spec 1195），門才開得了")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	data, err := gamecorpus.Load(*image, *localePath)
	if err != nil {
		log.Fatal(err)
	}

	walks := make([]segmentWalk, 0, 16)
	type cellRecord struct {
		Block   uint8 `json:"block"`
		Terrain uint8 `json:"terrain"`
	}
	records := make([]cellRecord, 0, 256)
	onMap := make([]cellRecord, 0, 256)
	routeStarts := loadRouteStarts(*routePath)
	components := make([]cellRecord, 0, 256)
	for _, seg := range segment.All() {
		payload, ok := data.Blocks[seg.Block]
		if !ok {
			continue
		}
		dispatch := eclcells.Analyze(payload)
		if !dispatch.Found {
			continue
		}
		catalog, has := data.Geo[seg.Member]
		if !has {
			continue
		}
		grid, hasGrid := catalog.Lookup(geo.MapRef{Set: seg.Member, BlockID: dispatch.GeoBlock})
		if !hasGrid {
			continue
		}
		walk := segmentWalk{id: seg.ID, block: seg.Block, reached: map[int]bool{}}
		terrains := map[uint8]bool{}
		// 兩種走法都跑，聯集才是這一段真正走得到的（見 `walkSegment` 的說明）。
		// ★ 走訪策略要**跑好幾種再取聯集**。單一策略的結果看起來都很合理，
		// 但每一種都會被某一類岔路擋住：選第一項會被收費關卡擋在外面，
		// 選最後一項會在「要離開嗎」那種提示上直接走人。
		for _, policy := range []struct {
			follow bool
			pick   int
		}{
			{false, 0}, {true, 0}, {false, 1}, {true, 1},
			{false, 2}, {true, 2}, {false, -1}, {true, -1},
		} { // ⚠ 試過再加第 4 項：一個索引都沒多，跑一趟卻多花兩分半 ⇒ 收斂了。
			// ⚠ 每一份旗標取樣各走一趟：門什麼時候開是**時間點**的問題，
			// 一份取樣只代表一個時間點（spec 1193）。
			for _, handoff := range handoffsFor(*arrivals, seg.Block) {
				pass := segmentWalk{id: seg.ID, block: seg.Block, reached: map[int]bool{}}
				if err := walkSegment(data, seg, grid, dispatch.Mask, *steps, &pass, terrains,
					policy.follow, policy.pick, handoff, nil, *openDoors); err != nil {
					if walk.note == "" {
						walk.note = err.Error()
					}
					continue
				}
				for index := range pass.reached {
					walk.reached[index] = true
				}
				if pass.cells > walk.cells {
					walk.cells = pass.cells
				}
				walk.blocked += pass.blocked
				walk.teleports += pass.teleports
			}
		}
		// ★ 第二輪：從**主線真的站過**的格子起走。
		//
		// ⚠ 只挑「那一格的索引，第一輪沒走到」的位置——其餘的起點只會把同一塊
		// 連通分量再走一遍，時間翻倍而一格都不會多。
		for _, at := range routeStarts[seg.Block] {
			index := int(grid.CellWrapped(at.x, at.y).Terrain) & dispatch.Mask
			if walk.reached[index] {
				continue
			}
			for _, follow := range []bool{false, true} {
				pass := segmentWalk{id: seg.ID, block: seg.Block, reached: map[int]bool{}}
				start := at
				if err := walkSegment(data, seg, grid, dispatch.Mask, *steps, &pass, terrains,
					follow, 0, handoffsFor(*arrivals, seg.Block)[0], &start, *openDoors); err != nil {
					continue
				}
				for index := range pass.reached {
					walk.reached[index] = true
				}
				walk.fromRoute++
			}
		}
		for terrain := range terrains {
			records = append(records, cellRecord{Block: seg.Block, Terrain: terrain})
		}
		// 整張圖掃一遍：這個 block 的地圖上到底出現過哪些地形碼。
		present := map[uint8]bool{}
		for y := 0; y < geo.Height; y++ {
			for x := 0; x < geo.Width; x++ {
				present[grid.CellWrapped(x, y).Terrain] = true
			}
		}
		for terrain := range present {
			onMap = append(onMap, cellRecord{Block: seg.Block, Terrain: terrain})
		}
		// ★ 連通分量：一張圖常常不是一整片。巫師塔每一層在 GEO 上是獨立的小房間，
		// 層與層之間只靠樓梯**事件**接（spec 1161）——從入口用走的永遠到不了別層。
		// 分不出「走不進去」是**幾何上斷開**還是**門擋著**，就不知道要補路線
		// 還是補劇情旗標。
		for terrain := range componentTerrains(data, grid) {
			components = append(components, cellRecord{Block: seg.Block, Terrain: terrain})
		}
		walks = append(walks, walk)
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].Block != records[right].Block {
			return records[left].Block < records[right].Block
		}
		return records[left].Terrain < records[right].Terrain
	})
	sort.Slice(onMap, func(left, right int) bool {
		if onMap[left].Block != onMap[right].Block {
			return onMap[left].Block < onMap[right].Block
		}
		return onMap[left].Terrain < onMap[right].Terrain
	})
	sort.Slice(components, func(left, right int) bool {
		if components[left].Block != components[right].Block {
			return components[left].Block < components[right].Block
		}
		return components[left].Terrain < components[right].Terrain
	})
	if *componentsOut != "" {
		payload, err := json.MarshalIndent(components, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*componentsOut, append(payload, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	if *onMapOut != "" {
		payload, err := json.MarshalIndent(onMap, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*onMapOut, append(payload, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	if *cellsOut != "" {
		payload, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*cellsOut, payload, 0o644); err != nil {
			log.Fatal(err)
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# 段內可達性：從入口用走的，走得到哪些格子\n\n")
	fmt.Fprintf(&report, "由 `cmd/dungeon-walk-probe` 產生，不要手改。\n\n")
	fmt.Fprintf(&report, "★ 可達性報表說主線實跑只踏到 250 個分派索引裡的 81 個，"+
		"其中提爾佛頓整整三段一格都沒踏到。那句話有兩種完全不同的成因："+
		"**主線不經過那裡**（路線的選擇）或**那些格子走不進去**（缺陷）。"+
		"逐格實測分不出來——它把隊伍**直接放**到目標格上。這一支從入口出發，"+
		"只用 `CanMoveDungeon`／`MoveDungeon` 廣度優先地走。\n\n")
	fmt.Fprintf(&report, "⚠ 這**不是**「從新遊戲玩到那裡」：進段是直接進的。"+
		"它答的是**段內**可達性——人已經在這張地圖上，走得到幾格。\n\n")
	fmt.Fprintf(&report, "⚠ 走的時候會踩到事件、戰鬥、被劇情推走；推不回地城就停在那條分支上。"+
		"⇒ **下界**：走不到不代表不可達，可能只是被某個事件擋在半路。\n\n")
	fmt.Fprintf(&report, "| 段 | ECL block | 走到的格子 | 走到的分派索引 | 撞牆次數 | 備註 |\n")
	fmt.Fprintf(&report, "|---|---:|---:|---:|---:|---|\n")
	totalIndices := 0
	for _, walk := range walks {
		note := walk.note
		if note == "" {
			note = "—"
		}
		fmt.Fprintf(&report, "| `%s` | %d | %d | %d | %d | %s |\n",
			walk.id, walk.block, walk.cells, len(walk.reached), walk.blocked, note)
		totalIndices += len(walk.reached)
	}
	fmt.Fprintf(&report, "\n合計走得到 **%d** 個分派索引。\n", totalIndices)

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "segments=%d indices=%d cells=%d\n", len(walks), totalIndices, len(records))
}

// walkSegment 進段之後從落點做廣度優先，只走 `CanMoveDungeon` 允許的邊。
//
// ⚠ 每走一步都要把事件推完再繼續，否則下一次 `MoveDungeon` 會在事件模式上失敗
// ——那會被記成「走不到」，而其實只是沒把畫面推回地城。
// walkSegment 走一遍。`followTeleports` 決定踏上去被事件搬走時要不要**從落點繼續**。
//
// ★ 兩種走法**互不涵蓋**，所以兩種都要跑再取聯集：
//
//	不跟傳送  走訪順序穩定，once-only 事件按幾何順序觸發
//	跟傳送    樓梯進得去別的連通分量（`ECL6/0x43` 的格子 191 → 235），
//	          但順序一變，某些 once-only 事件就在別的地方觸發了 ⇒ 反而少踩到兩個索引
//
// ⚠ 只跑其中一種都會低估，而且**兩邊看起來都很合理**——這正是「下界看起來和
// 全集一樣合理」那個坑（spec 1186）。
// handoffsFor 讀那一段**所有**的旗標取樣：剛換到那一份，加上段內沿路的幾份。
//
// ★ 為什麼不只用到達那一份：主線在一段裡走著走著才會開的門（拿到鑰匙、打完某一
// 場、觸發某個開關），在**到達那一刻**還沒開。每一份各走一趟再取聯集才問得完。
//
// ⚠ 目錄沒給或那一段沒有取樣就回一份 nil——冷走照舊跑一趟，只是門打不開。
func handoffsFor(dir string, block uint8) []map[uint16]uint16 {
	if dir == "" {
		return []map[uint16]uint16{nil}
	}
	names, _ := filepath.Glob(filepath.Join(dir, fmt.Sprintf("arrival-block-%02X.json", block)))
	more, _ := filepath.Glob(filepath.Join(dir, fmt.Sprintf("flags-block-%02X-*.json", block)))
	names = append(names, more...)
	sort.Strings(names)
	out := make([]map[uint16]uint16, 0, len(names)+1)
	for _, name := range names {
		if memory := readSample(name); memory != nil {
			out = append(out, memory)
		}
	}
	if len(out) == 0 {
		return []map[uint16]uint16{nil}
	}
	return out
}

func readSample(path string) map[uint16]uint16 {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var file struct {
		Schema string            `json:"schema"`
		Memory map[string]uint16 `json:"memory"`
	}
	if json.Unmarshal(raw, &file) != nil || file.Schema != "coab-arrival-sample/1" {
		return nil
	}
	out := make(map[uint16]uint16, len(file.Memory))
	for key, value := range file.Memory {
		address, err := strconv.ParseUint(key, 10, 16)
		if err != nil {
			continue
		}
		out[uint16(address)] = value
	}
	return out
}

func walkSegment(data gamecorpus.Corpus, seg segment.Segment, grid geo.Grid, mask, steps int,
	walk *segmentWalk, terrains map[uint8]bool, followTeleports bool, pick int,
	handoff map[uint16]uint16, from *point, openDoors bool) error {
	state, err := data.NewParty()
	if err != nil {
		return err
	}
	// ★ 冷走每一段都開一支新隊伍、**沒有任何劇情旗標**，所以要旗標才開得了的門
	// 對它永遠是牆——那正是可達性缺口的主要成因（spec 1193）。有了主線在那一段
	// 的到達取樣，就可以把當時的旗標鋪上去再走（spec 1195）。
	//
	// ⚠ 這仍然**不是**「玩家走得到」的證明，是**帶著劇情旗標的幾何可達性**。
	if len(handoff) > 0 {
		state.SeedHandoffMemory(handoff)
	}
	if err := state.EnterSegment(seg); err != nil {
		return fmt.Errorf("進不去：%v", err)
	}
	// ⚠ 隊伍要撐起來：入口伏擊會讓臨時角色死在門口，整段就走不了。**只給盤點用。**
	if err := gamecorpus.BoostParty(&state); err != nil {
		return err
	}
	if err := settleWith(&state, pick); err != nil {
		return fmt.Errorf("入口推不動：%v", err)
	}
	// ★ `from` 非 nil 就從**主線真的站過的那一格**開始走。
	//
	// 剩下的可達性缺口成因是「幾何上斷開——樓梯／傳送事件才進得去」：那些格子
	// 站得上去，但從段入口走不過去。主線的路線紀錄（`COAB_DECISION_LOG`）裡有
	// 隊伍**真的站過**的座標，從那裡起走就進得去那一塊連通分量（spec 1193）。
	//
	// ⚠ 這仍然是**幾何可達性**：它證明「劇情把你放到那裡之後，那一塊走得完」，
	// 不是「你走得到那裡」。走得到那裡是主線紀錄本身的事。
	if from != nil {
		state.DungeonX, state.DungeonY = from.x, from.y
		state.DungeonWallRoof = grid.CellWrapped(from.x, from.y).Terrain
	}
	start := point{state.DungeonX, state.DungeonY}
	// ⚠ **踏進一格的方向會改變結果**：樓梯事件是「站對方向踏上去」才觸發的
	// （`ECL5/0x33:090Ch` 用地形碼查表拿方向，朝向不對就直接 `EXIT`，畫面上
	// 什麼都不會發生，spec 1161）。只用格子當鍵，第一次從錯的方向踏上去就把
	// 那一格封死，另外三個方向永遠不會再試——**靠樓梯才進得去的樓層因此永遠
	// 走不到**，而報表只會顯示「那幾格走不到」，看不出是走法的問題。
	// ⇒ 邊界記 **(格子, 進入方向)**；「這一格去過沒有」另外記。
	type entry struct {
		at        point
		direction int
	}
	seen := map[point]bool{start: true}
	tried := map[entry]bool{}
	queue := []point{start}
	record := func(state *game.State) {
		terrains[state.DungeonWallRoof] = true
		walk.reached[int(state.DungeonWallRoof)&mask] = true
	}
	record(&state)
	walk.cells = 1
	used := 0
	for len(queue) > 0 && used < steps {
		current := queue[0]
		queue = queue[1:]
		for _, direction := range []int{0, 2, 4, 6} {
			if used >= steps {
				break
			}
			deltaX, deltaY := dungeonDelta(direction)
			next := point{current.x + deltaX, current.y + deltaY}
			if next.x < 0 || next.x >= geo.Width || next.y < 0 || next.y >= geo.Height {
				continue
			}
			if tried[entry{next, direction}] {
				continue
			}
			tried[entry{next, direction}] = true
			// 每一次都從入口重走到 current 太貴；改成直接把視角放到 current
			// 再試那一步。⚠ 這是**幾何可達性**：牆擋不擋得住。ECL 把隊伍推回
			// 上一格那種「走過去又被送回來」不算在內（那是內容不是幾何）。
			state.SetDungeonGeometryView(current.x, current.y, uint8(direction))
			state.DungeonWallRoof = grid.CellWrapped(current.x, current.y).Terrain
			if !state.CanMoveDungeon(grid, deltaX, deltaY, direction) {
				// ★ **門在 `CanMoveDungeon` 眼裡是牆。** 玩家的做法是撞上去、
				// 開選單、撬鎖／撞開／唸開鎖術；冷走只問「牆擋不擋得住」，
				// 於是**門後面的房間永遠走不到**——猶拉什那兩段缺的十幾格，
				// 有一半是圖上散落的**單格房間**（spec 1193）。
				//
				// ⚠ 這裡直接把門解鎖，**不擲骰**：這是**盤點**，問的是
				// 「門開了之後走得到嗎」，不是「這支隊伍開不開得了這扇門」。
				// 開鎖成功率是另一件事（`dungeon.PickLock`／`BashDoor`）。
				if !openDoors {
					walk.blocked++
					continue
				}
				flags, ok := grid.WallDoorFlagsWrapped(current.x, current.y, direction)
				if !ok || (flags != 2 && flags != 3) {
					walk.blocked++
					continue
				}
				if !grid.UnlockDoorWrapped(current.x, current.y, direction) {
					walk.blocked++
					continue
				}
				walk.doorsOpened++
				if !state.CanMoveDungeon(grid, deltaX, deltaY, direction) {
					walk.blocked++
					continue
				}
			}
			used++
			if err := state.MoveDungeon(grid, deltaX, deltaY, direction); err != nil {
				walk.blocked++
				continue
			}
			if err := settleWith(&state, pick); err != nil {
				// 推不回地城就別再從這一格往外走，但已經到過的算數。
				if !seen[next] {
					seen[next] = true
					walk.cells++
				}
				continue
			}
			if !seen[next] {
				seen[next] = true
				walk.cells++
			}
			record(&state)
			queue = append(queue, next)
			// ★ 事件可能把隊伍搬走（樓梯、傳送）。原本只把**打算走到**的那一格
			// 排進佇列，落點就被丟掉了——而**樓梯正是進到別的連通分量的唯一辦法**
			// （巫師塔每一層在 GEO 上是獨立房間，spec 1161）。丟掉落點等於把
			// 「走上樓」記成「走到樓梯口」，那些樓層永遠不會被走到。
			if !followTeleports {
				continue
			}
			landedX, landedY, _ := state.DungeonGeometryView()
			landed := point{landedX, landedY}
			if landed != next && !seen[landed] {
				seen[landed] = true
				walk.cells++
				walk.teleports++
				queue = append(queue, landed)
			}
		}
	}
	return nil
}

// settle 把事件／選單／戰鬥推完，直到回到地城模式。
// settle 把事件／選單／戰鬥推完，直到回到地城模式。
//
// ⚠ `preferLast` 決定選單挑第一項還是最後一項。**這不是美觀問題**：地城裡有
// 收費關卡（下水道的奧提尤格要食物）與「要進去嗎」的岔路，挑第一項會被擋在
// 外面，而被擋住的那一側**看起來就像沒有路**。兩種都跑再取聯集。
func settleWith(state *game.State, pick int) error {
	for step := 0; step < 40 && state.Mode != game.ModeDungeon; step++ {
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					return err
				}
			}
			continue
		}
		// `pick < 0` ＝ 最後一項；否則取第 `pick` 項（超出範圍就取最後一項）。
		choice := 0
		if count := len(state.Choices); count > 0 {
			switch {
			case pick < 0 || pick >= count:
				choice = count - 1
			default:
				choice = pick
			}
		}
		if state.Mode == game.ModePlace && len(state.Choices) > 0 {
			choice = len(state.Choices) - 1
		}
		if err := state.Continue(); err != nil {
			if selectErr := state.Select(choice); selectErr != nil {
				return fmt.Errorf("continue=%v select=%v", err, selectErr)
			}
		}
	}
	if state.Mode != game.ModeDungeon {
		return fmt.Errorf("停在%v", state.Mode)
	}
	return nil
}

// componentTerrains 回傳「**站得上去**的格子」出現過的地形碼——也就是任何一個
// 連通分量裡的格子，不限於入口那一塊。
//
// ★ 判準是「至少有一條邊進得來或出得去」：孤立到四面都不通的格子玩家永遠站不上。
//
// ⚠ 這裡**只看幾何**（`CanMoveDungeon`），不跑 ECL。走過去又被劇情推回來那種
// 不算在內——那是內容不是幾何，混進來會讓這個上限失去意義。
func componentTerrains(data gamecorpus.Corpus, grid geo.Grid) map[uint8]bool {
	out := map[uint8]bool{}
	state, err := data.NewParty()
	if err != nil {
		return out
	}
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			for _, direction := range []int{0, 2, 4, 6} {
				deltaX, deltaY := dungeonDelta(direction)
				state.SetDungeonGeometryView(x, y, uint8(direction))
				state.DungeonWallRoof = grid.CellWrapped(x, y).Terrain
				if !state.CanMoveDungeon(grid, deltaX, deltaY, direction) {
					continue
				}
				out[grid.CellWrapped(x, y).Terrain] = true
				out[grid.CellWrapped(x+deltaX, y+deltaY).Terrain] = true
			}
		}
	}
	return out
}

// settle 是預設策略（選單挑第一項），保留給既有呼叫點。
func settle(state *game.State) error { return settleWith(state, 0) }
