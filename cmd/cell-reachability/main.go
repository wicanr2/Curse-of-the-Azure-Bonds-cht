// Command cell-reachability 回答走訪那一列最後缺的那個問題：
// **逐格實測試過的那些格子，正常隊伍走得到幾個？**
//
// ★ 為什麼這一支不能靠靜態分析：`cmd/cell-sweep` 把隊伍**直接放**到目標格上，
// 還把隊伍撐起來以免入口戰鬥擋住盤點——所以它答的是「這一格演不演得出來」，
// 不是「走不走得到」。可達性只有實跑資料答得了，而
// `TestRealNewGameRunsToTheEnding`（從角色建立打到提朗瑟克斯）就是實跑資料。
//
// 資料怎麼來：
//
//	COAB_CAMPAIGN_CELLS_PATH=... go test -run TestRealNewGameRunsToTheEnding
//
// 那條 session 每次觀測都記下當時的 (ECL block, 地形碼)。這一支把它與逐格實測
// 的分派索引對起來。
//
// ⚠ **「沒被踏到」不等於「走不到」。** 主線只走它要走的路：支線房間、可選對話、
// 另一條分歧本來就不會被踏到。這一支給的是**下界**——它能證明「這些走得到」，
// 不能證明「其餘走不到」。
//
// ⚠ 記錄點是觀測迴圈，不是每一步移動。走過去又馬上被劇情推走的格子可能沒被取樣。
// 同樣是**下界**。
//
// 用法：
//
//	go run ./cmd/cell-reachability -output docs/audit/cell-reachability.md
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

)

// sweepIndex 是逐格實測輸出的一筆分母。**這一支不重算分母**——
// 兩邊各自算一定會漂：第一版自己掃全圖地形得到 299，而實測是 250，
// 兩個數都自洽，放在一起就是假的覆蓋率。
type sweepIndex struct {
	Segment string `json:"segment"`
	Block   uint8  `json:"block"`
	Mask    int    `json:"mask"`
	Index   int    `json:"index"`
	Played  bool   `json:"played"`
}

type visitedCell struct {
	Block   uint8 `json:"block"`
	Terrain uint8 `json:"terrain"`
}

func main() {
	sweepPath := flag.String("sweep-indices", "workplace/campaign-frames/sweep-indices.json",
		"`cmd/cell-sweep -index-json` 輸出的分母")
	walkPath := flag.String("walk-cells", "workplace/campaign-frames/walk-cells.json",
		"`cmd/dungeon-walk-probe -cells-json` 輸出的段內可達格子（可省略）")
	// ★ 這兩份把「沒被踏到」拆成三種完全不同的成因。少了它們，未達成只是一個
	// 數字，看不出**要補路線、要補劇情旗標、還是根本補不了**。
	walkablePath := flag.String("walkable-cells", "workplace/campaign-frames/walkable-cells.json",
		"`-walkable-json` 輸出的「站得上去」格子（走路的上限，可省略）")
	onMapPath := flag.String("on-map-cells", "workplace/campaign-frames/on-map-cells.json",
		"`-on-map-json` 輸出的「圖上出現過」格子（可省略）")
	visitedPath := flag.String("visited", "workplace/campaign-frames/visited-cells.json",
		"主線實跑導出的格子（見本檔頂端）")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	outputJSON := flag.String("json", "", "JSON 輸出路徑，給 `cmd/remake-status` 取用")
	flag.Parse()

	raw, err := os.ReadFile(*visitedPath)
	if err != nil {
		log.Fatalf("讀不到實跑資料 %s：%v\n"+
			"先跑：COAB_CAMPAIGN_CELLS_PATH=/src/%s tools/go.sh test ./internal/game/ "+
			"-run TestRealNewGameRunsToTheEnding -count=1", *visitedPath, err, *visitedPath)
	}
	var visited []visitedCell
	if err := json.Unmarshal(raw, &visited); err != nil {
		log.Fatal(err)
	}
	if len(visited) == 0 {
		log.Fatal("實跑資料是空的——記錄點沒被叫到")
	}
	// block → 踏過的地形碼。
	walked := map[uint8]map[uint8]bool{}
	for _, cell := range visited {
		if walked[cell.Block] == nil {
			walked[cell.Block] = map[uint8]bool{}
		}
		walked[cell.Block][cell.Terrain] = true
	}

	sweepRaw, err := os.ReadFile(*sweepPath)
	if err != nil {
		log.Fatalf("讀不到逐格實測的索引清單 %s：%v\n"+
			"先跑：tools/go.sh run ./cmd/cell-sweep -index-json %s", *sweepPath, err, *sweepPath)
	}
	var sweep []sweepIndex
	if err := json.Unmarshal(sweepRaw, &sweep); err != nil {
		log.Fatal(err)
	}
	if len(sweep) == 0 {
		log.Fatal("逐格實測的索引清單是空的")
	}

	// 段內走得到的格子（可省略）。有它才分得出「主線不經過」與「走不進去」。
	walkable := map[uint8]map[uint8]bool{}
	if raw, err := os.ReadFile(*walkPath); err == nil {
		var cells []visitedCell
		if json.Unmarshal(raw, &cells) == nil {
			for _, cell := range cells {
				if walkable[cell.Block] == nil {
					walkable[cell.Block] = map[uint8]bool{}
				}
				walkable[cell.Block][cell.Terrain] = true
			}
		}
	}

	loadCells := func(path string) map[uint8]map[uint8]bool {
		out := map[uint8]map[uint8]bool{}
		raw, err := os.ReadFile(path)
		if err != nil {
			return out
		}
		var cells []visitedCell
		if json.Unmarshal(raw, &cells) != nil {
			return out
		}
		for _, cell := range cells {
			if out[cell.Block] == nil {
				out[cell.Block] = map[uint8]bool{}
			}
			out[cell.Block][cell.Terrain] = true
		}
		return out
	}
	standable := loadCells(*walkablePath)
	onMap := loadCells(*onMapPath)
	matches := func(table map[uint8]map[uint8]bool, block uint8, mask, index int) bool {
		for terrain := range table[block] {
			if int(terrain)&mask == index {
				return true
			}
		}
		return false
	}

	type blockRow struct {
		id       string
		block    uint8
		indices  int
		reached  int
		onFoot   int
		either   int
		walkedIn int
		// 未達成的三種成因（見報表說明）。
		gatedOff  int
		noEdge    int
		notOnMap  int
	}
	bySegment := map[string]*blockRow{}
	order := make([]string, 0, 16)
	totalIndices, totalReached, totalOnFoot, totalEither := 0, 0, 0, 0
	totalGated, totalNoEdge, totalNotOnMap, totalStandable := 0, 0, 0, 0
	for _, item := range sweep {
		row, ok := bySegment[item.Segment]
		if !ok {
			row = &blockRow{id: item.Segment, block: item.Block}
			bySegment[item.Segment] = row
			order = append(order, item.Segment)
		}
		row.indices++
		totalIndices++
		if matches(standable, item.Block, item.Mask, item.Index) {
			totalStandable++
		}
		// 主線在這個 block 踏過的地形碼，換算成索引之後對得上嗎。
		reachedHere := false
		for terrain := range walked[item.Block] {
			if int(terrain)&item.Mask == item.Index {
				reachedHere = true
				break
			}
		}
		if reachedHere {
			row.reached++
			totalReached++
		}
		onFoot := false
		for terrain := range walkable[item.Block] {
			if int(terrain)&item.Mask == item.Index {
				onFoot = true
				break
			}
		}
		if onFoot {
			row.onFoot++
			totalOnFoot++
		}
		// ⚠ 兩者**互不涵蓋**：主線有劇情旗標，開得了冷走打不開的門；
		// 冷走沒有劇情擋路，走得到主線繞過的地方。所以聯集才是目前最好的下界。
		if onFoot || reachedHere {
			row.either++
			totalEither++
			continue
		}
		// ── 沒達成的分三種，處置完全不同 ────────────────────────────
		switch {
		case matches(standable, item.Block, item.Mask, item.Index):
			// 站得上去，只是從段入口走不過去：幾何上斷開（樓梯／傳送事件才進得去）
			// 或門擋著。要補的是**路線**。
			row.gatedOff++
			totalGated++
		case matches(onMap, item.Block, item.Mask, item.Index):
			// 圖上有這個地形碼，但那些格子四面都不通——走路永遠站不上去。
			row.noEdge++
			totalNoEdge++
		default:
			// 這個 block 的地圖上根本沒有這個地形碼：分派表有、圖上沒有。
			row.notOnMap++
			totalNotOnMap++
		}
	}
	rows := make([]blockRow, 0, len(order))
	for _, id := range order {
		row := bySegment[id]
		row.walkedIn = len(walked[row.block])
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(left, right int) bool { return rows[left].id < rows[right].id })

	var report strings.Builder
	fmt.Fprintf(&report, "# 走訪可達性：主線實際踏到了逐格實測的哪些格子\n\n")
	fmt.Fprintf(&report, "由 `cmd/cell-reachability` 產生，不要手改。\n\n")
	fmt.Fprintf(&report, "★ 走訪那一列最後缺的問題是**可達性**。`cmd/cell-sweep` 把隊伍"+
		"**直接放**到目標格上、還把隊伍撐起來，所以它答的是「這一格演不演得出來」，"+
		"不是「正常隊伍走不走得到」。可達性只有實跑資料答得了——"+
		"而 `TestRealNewGameRunsToTheEnding`（從角色建立打到提朗瑟克斯）與 "+
		"`TestTilvertonRouteIsWalkableAndLocalized`（開場那三段）就是實跑資料。\n\n")
	fmt.Fprintf(&report, "⚠ **「沒被踏到」不等於「走不到」。** 實跑路線只走它要走的路："+
		"支線房間、可選對話、另一條分歧本來就不會被踏到。這一份能證明「這些走得到」，"+
		"**不能證明「其餘走不到」** ⇒ 覆蓋率是**下界**。\n\n")
	fmt.Fprintf(&report, "⚠ 記錄點是觀測迴圈不是每一步移動，走過去又馬上被劇情推走的"+
		"可能沒被取樣；同樣是下界。\n\n")

	fmt.Fprintf(&report, "| 指標 | 數字 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 有地形分派、地圖上也有格子的 block | %d |\n", len(rows))
	fmt.Fprintf(&report, "| 分派索引（**直接取自逐格實測的輸出**）| %d |\n", totalIndices)
	fmt.Fprintf(&report, "| **實跑路線踏到的** | %d |\n", totalReached)
	fmt.Fprintf(&report, "| **從段入口用走的走得到的** | %d |\n", totalOnFoot)
	fmt.Fprintf(&report, "| **兩者的聯集**（目前最好的下界）| %d |\n", totalEither)
	fmt.Fprintf(&report, "| **走路的上限**（站得上去的格子，不限從入口走得到）| %d |\n",
		totalStandable)
	if totalIndices > 0 {
		fmt.Fprintf(&report, "| 聯集覆蓋率（下界）| %.0f%% |\n",
			100*float64(totalEither)/float64(totalIndices))
		fmt.Fprintf(&report, "| 佔走路上限的比例 | %.0f%% |\n",
			100*float64(totalEither)/float64(totalStandable))
	}
	report.WriteString("\n★ **「走路的上限」才是走訪這條路的天花板**：它是站得上去的格子，" +
		"不管要不要靠事件才進得去那一塊。分母 250 裡差的那一格站不上去（見下），" +
		"所以走訪能做到的最好就是這個數字——**不是 250**。\n")
	fmt.Fprintf(&report, "\n## 沒達成的那些，是哪一種沒達成\n\n")
	report.WriteString("★ 未達成只是一個數字的話，看不出**要補路線、要補劇情旗標、還是根本補不了**。" +
		"這三種的處置完全不同：\n\n")
	fmt.Fprintf(&report, "| 成因 | 索引 | 意思 | 要做什麼 |\n|---|---:|---|---|\n")
	fmt.Fprintf(&report, "| 站得上去，但從段入口走不到 | %d | 幾何上斷開（樓梯／傳送**事件**才進得去，"+
		"巫師塔每一層都是獨立房間）或門擋著 | 補路線：從那一塊的入口格另外走一趟 |\n", totalGated)
	fmt.Fprintf(&report, "| 圖上有，但四面都不通 | %d | 那個地形碼只出現在沒有任何可通行邊的格子上 | "+
		"**走路永遠站不上去**；要嘛靠事件傳送，要嘛它本來就不是玩家會站的格子 |\n", totalNoEdge)
	fmt.Fprintf(&report, "| 圖上根本沒有這個地形碼 | %d | 分派表有這一格，地圖上沒有 | "+
		"分母裡的死索引，補不了也不用補 |\n\n", totalNotOnMap)
	report.WriteString("⚠ 這一段用的是**純幾何**（`CanMoveDungeon`），不跑 ECL：" +
		"走過去又被劇情推回來那種不算「走不到」——那是內容不是幾何，混進來會讓分類失去意義。\n")

	fmt.Fprintf(&report, "\n⚠ **「實跑路線踏到」裡有兩種東西**，不要把它讀成「玩到的比例」："+
		"`TestRealNewGameRunsToTheEnding` 是照劇情走的主線，"+
		"而 `TestTilvertonRouteIsWalkableAndLocalized` 是**廣度優先的走訪**——"+
		"後者用的方法和下面那個「冷走」一樣，所以它一加進來，這一欄就會往冷走那一欄收斂，"+
		"而**聯集幾乎不動**（101→152 但聯集只從 177 到 178）。"+
		"⇒ 那條路線的價值不在數字，在於它是**帶著語系不變量的測試**："+
		"第一次跑就抓到兩句玩家看得到的英文。\n")
	fmt.Fprintf(&report, "\n⚠ **兩把尺互不涵蓋，所以要看聯集。** 主線有劇情旗標，"+
		"開得了冷走打不開的門（`ECL4/0x22` 主線 6、冷走 1）；冷走沒有劇情擋路，"+
		"走得到主線繞過的地方（`ECL6/0x40` 主線 2、冷走 22）。"+
		"只報其中一個都會低估。\n")
	fmt.Fprintf(&report, "\n| 段 | ECL block | 分派索引 | 實跑踏到 | 走得到 | 聯集 |\n")
	fmt.Fprintf(&report, "|---|---:|---:|---:|---:|---:|\n")
	for _, row := range rows {
		fmt.Fprintf(&report, "| `%s` | %d | %d | %d | %d | %d |\n",
			row.id, row.block, row.indices, row.reached, row.onFoot, row.either)
	}
	fmt.Fprintf(&report, "\n")

	// 整段 0 的要單獨點出來：那不是「覆蓋率低」，是**這條路線根本沒去過那裡**。
	untouched := make([]blockRow, 0, 4)
	missing := 0
	for _, row := range rows {
		if row.reached == 0 {
			untouched = append(untouched, row)
			missing += row.indices
		}
	}
	fmt.Fprintf(&report, "## 實跑路線一格都沒踏到的段\n\n")
	if len(untouched) == 0 {
		fmt.Fprintf(&report, "（沒有。）\n")
	} else {
		fmt.Fprintf(&report, "這幾段**整段沒被任何實跑路線走過**，共 %d 個分派索引"+
			"（占分母的 %.0f%%）。那不是「覆蓋率低」，是這條路線根本沒去過那裡——"+
			"逐格實測驗過它們演得出來，但**沒有任何一條實跑路徑經過**。\n\n",
			missing, 100*float64(missing)/float64(totalIndices))
		fmt.Fprintf(&report, "| 段 | ECL block | 分派索引 |\n|---|---:|---:|\n")
		for _, row := range untouched {
			fmt.Fprintf(&report, "| `%s` | %d | %d |\n", row.id, row.block, row.indices)
		}
		onFoot := 0
		for _, row := range untouched {
			onFoot += row.onFoot
		}
		fmt.Fprintf(&report, "\n★ **那不是走不進去。** 同樣這幾段，從段入口用走的走得到 **%d** 個索引"+
			"（`cmd/dungeon-walk-probe`）。⇒ 成因是**主線不經過那裡**，不是格子不可達："+
			"主線測試從艾森布拉開始（`runNormalNewGameToEssembra`），提爾佛頓那一段序章"+
			"不在它的路線上。**要不要補一條走那裡的實跑路線是另一個決定**，"+
			"這一份負責把「路線的選擇」與「缺陷」分開。\n", onFoot)
	}

	if *outputJSON != "" {
		encoded, err := json.MarshalIndent(struct {
			Schema    string `json:"schema"`
			Indices   int    `json:"indices"`
			Reached   int    `json:"walked_by_campaign"`
			OnFoot    int    `json:"reached_on_foot"`
			Either    int    `json:"union"`
			Standable int    `json:"walkable_ceiling"`
			Gated     int    `json:"gated_off"`
			NoEdge    int    `json:"no_passable_edge"`
			NotOnMap  int    `json:"not_on_map"`
		}{
			Schema: "coab-cell-reachability/1", Indices: totalIndices, Reached: totalReached,
			OnFoot: totalOnFoot, Either: totalEither, Standable: totalStandable,
			Gated: totalGated, NoEdge: totalNoEdge, NotOnMap: totalNotOnMap,
		}, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*outputJSON, append(encoded, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "blocks=%d indices=%d reached=%d\n", len(rows), totalIndices, totalReached)
}
