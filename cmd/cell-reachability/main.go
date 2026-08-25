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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
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
		tooltext.Text("h.91b16ee690c4"))
	walkPath := flag.String("walk-cells", "workplace/campaign-frames/walk-cells.json",
		tooltext.Text("h.195ebca98b59"))
	// ★ 這兩份把「沒被踏到」拆成三種完全不同的成因。少了它們，未達成只是一個
	// 數字，看不出**要補路線、要補劇情旗標、還是根本補不了**。
	walkablePath := flag.String("walkable-cells", "workplace/campaign-frames/walkable-cells.json",
		tooltext.Text("h.78d6363f0eef"))
	onMapPath := flag.String("on-map-cells", "workplace/campaign-frames/on-map-cells.json",
		tooltext.Text("h.dbbbfd29f5ec"))
	// ★ 帶著劇情旗標走出來的格子（`cmd/campaign-snapshot-walk`）。冷走沒有旗標，
	// 門對它永遠是牆；主線有旗標，但只走它要走的路。這一份是兩者的交集能力：
	// **從主線快照出發、把那一段走遍**。
	snapshotPath := flag.String("snapshot-cells", "workplace/campaign-frames/snapshot-cells.json",
		tooltext.Text("h.750ea14b4ebe"))
	// ★ 主線**每一步移動**換算出來的格子（`cmd/route-cells`）。實跑那一欄本來
	// 只採樣觀測迴圈，「走過去又馬上被劇情推走」的格子沒被記到——這一份補上。
	routeCellsPath := flag.String("route-cells", "workplace/campaign-frames/route-cells.json",
		tooltext.Text("h.12b4b740c9cb"))
	visitedPath := flag.String("visited", "workplace/campaign-frames/visited-cells.json",
		tooltext.Text("h.e4e3d23ed386"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	outputJSON := flag.String("json", "", tooltext.Text("h.fa63314394ca"))
	flag.Parse()

	raw, err := os.ReadFile(*visitedPath)
	if err != nil {
		log.Fatalf(tooltext.Text("h.98410c4dd11a")+
			tooltext.Text("h.8d0c6b4ee725")+
			"-run TestRealNewGameRunsToTheEnding -count=1", *visitedPath, err, *visitedPath)
	}
	var visited []visitedCell
	if err := json.Unmarshal(raw, &visited); err != nil {
		log.Fatal(err)
	}
	if len(visited) == 0 {
		log.Fatal(tooltext.Text("h.4245ef915653"))
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
		log.Fatalf(tooltext.Text("h.8ae705333376")+
			tooltext.Text("h.902f6353b92c"), *sweepPath, err, *sweepPath)
	}
	var sweep []sweepIndex
	if err := json.Unmarshal(sweepRaw, &sweep); err != nil {
		log.Fatal(err)
	}
	if len(sweep) == 0 {
		log.Fatal(tooltext.Text("h.95a80cf8a993"))
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
	// 路線走出來的格子併進「實跑踏到」：來源都是主線，只是取樣密度不同。
	for block, set := range loadCells(*routeCellsPath) {
		if walked[block] == nil {
			walked[block] = map[uint8]bool{}
		}
		for terrain := range set {
			walked[block][terrain] = true
		}
	}
	withFlags := loadCells(*snapshotPath)
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
		gatedOff int
		noEdge   int
		notOnMap int
		// withFlags 是「只有帶著劇情旗標才走得到」的索引數。
		withFlags int
	}
	bySegment := map[string]*blockRow{}
	order := make([]string, 0, 16)
	totalIndices, totalReached, totalOnFoot, totalEither := 0, 0, 0, 0
	totalGated, totalNoEdge, totalNotOnMap, totalStandable := 0, 0, 0, 0
	totalGatedButPlayed := 0
	totalWithFlags := 0
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
		// 帶旗標走出來的算進「走得到」：那是同一種走訪，只是門開得了。
		if !onFoot && matches(withFlags, item.Block, item.Mask, item.Index) {
			onFoot = true
			row.withFlags++
			totalWithFlags++
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
			// ★ 走訪走不到，不代表那一格的**事件沒驗過**。逐格實測
			// （`cmd/cell-sweep`）是直接站上去的，它的 `played` 才是「那一格
			// 的事件演不演得出來」。兩件事分開數：走訪問的是**路**，
			// 逐格實測問的是**內容**。
			//
			// ⚠ 混在一起看會把「這個儀器記不到」讀成「這一格沒驗過」。
			if item.Played {
				totalGatedButPlayed++
			}
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
	fmt.Fprint(&report, tooltext.Format("h.743e62f813d8"))
	fmt.Fprint(&report, tooltext.Format("h.c21a2e90c97c"))
	fmt.Fprint(&report, tooltext.Text("h.2bbbb28f9713")+
		tooltext.Text("h.c707d2a9f63c")+
		tooltext.Text("h.b91268717948")+
		tooltext.Text("h.51df12757a90")+
		tooltext.Text("h.384107b0f75d"))
	fmt.Fprint(&report, tooltext.Text("h.01ca755e0101")+
		tooltext.Text("h.adf453a1f1c2")+
		tooltext.Text("h.4a90648f0242"))
	fmt.Fprint(&report, tooltext.Text("h.3b487f729f82")+
		tooltext.Text("h.01075088acb3"))

	fmt.Fprint(&report, tooltext.Format("h.13c83a8a875e"))
	fmt.Fprint(&report, tooltext.Format("h.f8daf874df0e", len(rows)))
	fmt.Fprint(&report, tooltext.Format("h.aaa6e7eb79a2", totalIndices))
	fmt.Fprint(&report, tooltext.Format("h.ed23a40b68ac", totalReached))
	fmt.Fprint(&report, tooltext.Format("h.67f8bafad4cc", totalOnFoot))
	fmt.Fprint(&report, tooltext.Format("h.fad72b478a27", totalWithFlags))
	fmt.Fprint(&report, tooltext.Format("h.3e0adabfa550", totalEither))
	fmt.Fprint(&report, tooltext.Format("h.943386f9ac34", totalStandable))
	if totalIndices > 0 {
		fmt.Fprint(&report, tooltext.Format("h.f154c5107d25", 100*float64(totalEither)/float64(totalIndices)))
		fmt.Fprint(&report, tooltext.Format("h.ef2026cf4698", 100*float64(totalEither)/float64(totalStandable)))
	}
	report.WriteString(tooltext.Text("h.c2995cc05231") +
		tooltext.Text("h.3ea3670e51cd") +
		tooltext.Text("h.3aaedda66d7f"))
	fmt.Fprint(&report, tooltext.Format("h.098f75165049"))
	report.WriteString(tooltext.Text("h.76c3222d68a4") +
		tooltext.Text("h.0e14bcd80d0d"))
	fmt.Fprint(&report, tooltext.Format("h.e0a2af8fb33f"))
	fmt.Fprintf(&report, tooltext.Text("h.61c40e5468e7")+
		tooltext.Text("h.a1af13331f14"), totalGated)
	fmt.Fprintf(&report, tooltext.Text("h.e14d0183596a")+
		tooltext.Text("h.04268a94e089"), totalNoEdge)
	fmt.Fprintf(&report, tooltext.Text("h.8a87b9c18848")+
		tooltext.Text("h.9408ac835ebe"), totalNotOnMap)
	report.WriteString(tooltext.Text("h.95e3e73f5ed4") +
		tooltext.Text("h.6585e4c47bc3"))
	fmt.Fprintf(&report, tooltext.Text("h.a4f73064e697")+
		tooltext.Text("h.18937e5b4810")+
		tooltext.Text("h.f2c8c20e94ae")+
		tooltext.Text("h.7fec156e0938"),
		totalGated+totalNoEdge+totalNotOnMap, totalGatedButPlayed)

	fmt.Fprint(&report, tooltext.Text("h.e49cff34933b")+
		tooltext.Text("h.7b084696b3e0")+
		tooltext.Text("h.384e2d29c510")+
		tooltext.Text("h.7dedc919fe86")+
		tooltext.Text("h.fd2d2b9fca3d")+
		tooltext.Text("h.cadc31e49143")+
		tooltext.Text("h.4d5b49c8e879"))
	// ⚠ 這兩個例子要**算出來**，不能寫死：走法一改動它們就變了，而寫死的數字
	// 不會有任何東西提醒你它已經不對。
	campaignWins, footWins := rows[0], rows[0]
	for _, row := range rows {
		if row.reached-row.onFoot > campaignWins.reached-campaignWins.onFoot {
			campaignWins = row
		}
		if row.onFoot-row.reached > footWins.onFoot-footWins.reached {
			footWins = row
		}
	}
	fmt.Fprintf(&report, tooltext.Text("h.dfd4f5b2e22f")+
		tooltext.Text("h.f1ce795e73c0")+
		tooltext.Text("h.a2f2ba844099")+
		tooltext.Text("h.539eaa7a0abe"),
		campaignWins.id, campaignWins.reached, campaignWins.onFoot,
		footWins.id, footWins.reached, footWins.onFoot)
	report.WriteString(tooltext.Text("h.fab697dcc323") +
		tooltext.Text("h.b8d3454bb420") +
		tooltext.Text("h.59359ab2e940") +
		tooltext.Text("h.848a49ef1b3b") +
		tooltext.Text("h.f28d49603ad9") +
		tooltext.Text("h.f47614da56e3") +
		tooltext.Text("h.c085c7a78e2a"))
	fmt.Fprint(&report, tooltext.Format("h.50a28011d14a"))
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
	fmt.Fprint(&report, tooltext.Format("h.7feed5b93a1b"))
	if len(untouched) == 0 {
		fmt.Fprint(&report, tooltext.Format("h.be6c71d35815"))
	} else {
		fmt.Fprintf(&report, tooltext.Text("h.4e30334a12b4")+
			tooltext.Text("h.697958203959")+
			tooltext.Text("h.aa3f6e760237"),
			missing, 100*float64(missing)/float64(totalIndices))
		fmt.Fprint(&report, tooltext.Format("h.74f97624b4dd"))
		for _, row := range untouched {
			fmt.Fprintf(&report, "| `%s` | %d | %d |\n", row.id, row.block, row.indices)
		}
		onFoot := 0
		for _, row := range untouched {
			onFoot += row.onFoot
		}
		fmt.Fprintf(&report, tooltext.Text("h.3df96eb5a1c7")+
			tooltext.Text("h.ed1129c15d4c")+
			tooltext.Text("h.3c50c395efba")+
			tooltext.Text("h.feee90fb18ae")+
			tooltext.Text("h.8bf4af94b7ee"), onFoot)
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
			// GatedButPlayed 是「走訪走不到，但逐格實測站上去演出來了」。
			GatedButPlayed int `json:"gated_off_but_played"`
			NoEdge         int `json:"no_passable_edge"`
			NotOnMap       int `json:"not_on_map"`
		}{
			Schema: "coab-cell-reachability/1", Indices: totalIndices, Reached: totalReached,
			OnFoot: totalOnFoot, Either: totalEither, Standable: totalStandable,
			Gated: totalGated, GatedButPlayed: totalGatedButPlayed,
			NoEdge: totalNoEdge, NotOnMap: totalNotOnMap,
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
