// Command campaign-snapshot-walk 從**主線存檔**出發把那一段走遍。
//
// ★ 存在的理由：走訪未達成的那幾十個索引，成因幾乎全是「站得上去，但從段入口
// 走不到」——門要**劇情旗標**才開得了（`cmd/cell-reachability` 的成因分類，
// spec 1193）。冷走（`cmd/dungeon-walk-probe`）每一段都開一支新隊伍、沒有旗標，
// 所以那些門對它永遠是牆；主線有旗標，但**主線只走它要走的路**，不會把房間走遍。
//
// 這一支把兩邊接起來：拿主線在各段落存下來的快照（`COAB_CAMPAIGN_SNAPSHOT_DIR`，
// 就是一般的存檔格式），**帶著那一刻的旗標**在那一段做廣度優先走訪。
//
// ⚠ 這**不是**「玩家走得到」的證明，是**帶著劇情旗標的幾何可達性**：快照那一刻
// 隊伍真的在那一段、旗標真的是那樣，所以那些門真的開得了；但走訪本身仍是機器
// 的走法，不是劇情路線。
//
// ⚠ 選單策略要跑好幾種再取聯集（spec 1193）：挑第一項會被收費關卡擋在門外，
// 挑最後一項會在「要離開嗎」直接走人。**單一策略的結果看起來都很合理。**
//
// 用法：
//
//	rm -rf workplace/campaign-frames/snapshots && mkdir -p workplace/campaign-frames/snapshots
//	COAB_CAMPAIGN_SNAPSHOT_DIR=/src/workplace/campaign-frames/snapshots \
//	    tools/go.sh test ./internal/game/ -count=1 \
//	    -run 'TestRealNewGameRunsToTheEnding|TestTilvertonRouteIsWalkableAndLocalized'
//	go run ./cmd/campaign-snapshot-walk -cells-json workplace/campaign-frames/snapshot-cells.json
//
// ⚠ **兩支測試都要跑。** 快照有兩個產生者：主線與提爾佛頓路線測試。只跑主線的話
// `0x01`／`0x02`／`0x03` 三段**一份快照都不會有**（那三段的段內快照全來自路線測試），
// 而少掉的段不會有任何錯誤訊息——報表只會少幾列，看起來和「那幾段沒東西可走」
// 一模一樣。
//
// ⚠ **目錄要先清掉。** 這裡讀的是目錄裡的**每一個檔**，不是這一次跑出來的那些。
// 沿用舊目錄會把上一次（不同程式、不同測試組合）留下的快照一起走進來，
// 得到一個**比現在的程式真的產得出來還大**的數字。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gamecorpus"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

type cellRecord struct {
	Block   uint8 `json:"block"`
	Terrain uint8 `json:"terrain"`
}

type point struct{ x, y int }

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "遊戲 image")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "語系檔")
	dir := flag.String("snapshots", "workplace/campaign-frames/snapshots", "主線快照目錄")
	steps := flag.Int("steps", 4000, "每一份快照最多走幾步")
	cellsOut := flag.String("cells-json", "", "把走得到的 (block, 地形碼) 寫成 JSON")
	output := flag.String("output", "", "Markdown 輸出路徑")
	flag.Parse()

	data, err := gamecorpus.Load(*image, *localePath)
	if err != nil {
		log.Fatal(err)
	}
	entries, err := os.ReadDir(*dir)
	if err != nil {
		log.Fatalf("讀不到快照目錄 %s：%v\n"+
			"先跑：COAB_CAMPAIGN_SNAPSHOT_DIR=/src/%s tools/go.sh test ./internal/game/ "+
			"-run TestRealNewGameRunsToTheEnding -count=1", *dir, err, *dir)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		log.Fatalf("%s 裡沒有快照", *dir)
	}

	type row struct {
		name    string
		block   uint8
		cells   int
		terrains int
		note    string
	}
	rows := make([]row, 0, len(names))
	terrains := map[uint8]map[uint8]bool{}
	for _, name := range names {
		item := row{name: strings.TrimSuffix(name, ".json")}
		found := map[int]bool{}
		cells := 0
		for _, pick := range []int{0, 1, 2, -1} {
			block, walked, reached, err := walkSnapshot(data, filepath.Join(*dir, name), *steps, pick)
			if err != nil {
				if item.note == "" {
					item.note = err.Error()
				}
				continue
			}
			item.block = block
			if terrains[block] == nil {
				terrains[block] = map[uint8]bool{}
			}
			for terrain := range reached {
				terrains[block][terrain] = true
			}
			for index := range reached {
				found[int(index)] = true
			}
			if walked > cells {
				cells = walked
			}
		}
		item.cells, item.terrains = cells, len(found)
		rows = append(rows, item)
	}

	records := make([]cellRecord, 0, 256)
	for block, set := range terrains {
		for terrain := range set {
			records = append(records, cellRecord{Block: block, Terrain: terrain})
		}
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].Block != records[right].Block {
			return records[left].Block < records[right].Block
		}
		return records[left].Terrain < records[right].Terrain
	})
	if *cellsOut != "" {
		payload, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*cellsOut, append(payload, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}

	var report strings.Builder
	report.WriteString("# 帶著劇情旗標走：從主線快照出發的段內走訪\n\n")
	report.WriteString("由 `cmd/campaign-snapshot-walk` 產生，不要手改。\n\n")
	report.WriteString("★ 走訪未達成的那幾十個索引，成因幾乎全是「站得上去，但從段入口走不到」" +
		"——門要**劇情旗標**才開得了（spec 1193）。冷走每一段都開一支新隊伍、沒有旗標，" +
		"所以那些門對它永遠是牆；主線有旗標，但**主線只走它要走的路**。" +
		"這一份拿主線各段的快照，**帶著那一刻的旗標**把那一段走遍。\n\n")
	report.WriteString("⚠ 這**不是**「玩家走得到」的證明，是**帶著劇情旗標的幾何可達性**：" +
		"快照那一刻隊伍真的在那一段、旗標真的是那樣，所以那些門真的開得了；" +
		"但走訪本身仍是機器的走法，不是劇情路線。\n\n")
	report.WriteString("⚠ 選單策略跑四種取聯集（第 1／2／3／最後項）：挑第一項會被收費關卡擋在門外，" +
		"挑最後一項會在「要離開嗎」直接走人。**單一策略的結果看起來都很合理。**\n\n")
	report.WriteString("| 快照 | block | 走到的格子 | 走到的地形碼 | 備註 |\n|---|---:|---:|---:|---|\n")
	totalIndices := 0
	for _, item := range rows {
		note := item.note
		if note == "" {
			note = "—"
		}
		fmt.Fprintf(&report, "| `%s` | %d | %d | %d | %s |\n",
			item.name, item.block, item.cells, item.terrains, note)
		totalIndices += item.terrains
	}
	fmt.Fprintf(&report, "\n合計 %d 份快照、%d 個 (block, 地形碼) 組合。\n",
		len(rows), len(records))

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "snapshots=%d cells=%d\n", len(rows), len(records))
}

// walkSnapshot 讀一份快照，帶著它的旗標把那一段走遍。
func walkSnapshot(data gamecorpus.Corpus, path string, steps, pick int) (uint8, int, map[uint8]bool, error) {
	state, err := data.NewParty()
	if err != nil {
		return 0, 0, nil, err
	}
	if err := state.LoadPartyFile(path); err != nil {
		return 0, 0, nil, fmt.Errorf("讀不回來：%v", err)
	}
	if err := settleWith(&state, pick); err != nil {
		return 0, 0, nil, fmt.Errorf("讀回來推不進地城：%v", err)
	}
	// ⚠ **不要用 `Area.LastECLBlockID`**：那一格是存檔格式裡的欄位，遊戲跑的時候
	// 從來沒有人寫它，讀出來是 0——而 0 是一個**合法的 block 編號**，查下去會拿到
	// 完全不相干的地圖而且不會報錯。存檔有帶 ECL session，段要跟 session 拿。
	block, ok := state.CurrentECLBlockID()
	if !ok {
		return 0, 0, nil, fmt.Errorf("沒有 ECL session")
	}
	payload, hasBlock := data.Blocks[block]
	if !hasBlock {
		return block, 0, nil, fmt.Errorf("corpus 裡沒有 block 0x%02X", block)
	}
	dispatch := eclcells.Analyze(payload)
	if !dispatch.Found {
		return block, 0, nil, fmt.Errorf("block 0x%02X 沒有地形分派", block)
	}
	// 地圖同理用存檔記下來的那一張。
	geoBlock := state.Area.Current3DMapBlockID
	if geoBlock == 0 {
		geoBlock = dispatch.GeoBlock
	}
	grid, has := gridFor(data, block, geoBlock)
	if !has {
		return block, 0, nil, fmt.Errorf("讀不到 GEO")
	}

	terrains := map[uint8]bool{}
	start := point{state.DungeonX, state.DungeonY}
	// ⚠ **踏進一格的方向會改變結果**：樓梯事件是「站對方向踏上去」才觸發的
	// （`ECL5/0x33:090Ch` 用地形碼查表拿方向，朝向不對就直接 `EXIT`，
	// 畫面上什麼都不會發生，spec 1161）。
	//
	// 只用格子當 `seen` 的鍵，第一次從錯的方向踏上去就把那一格封死，
	// 另外三個方向永遠不會再試——**靠樓梯才進得去的樓層因此永遠走不到**，
	// 而報表只會顯示「那幾格走不到」，看不出是走法的問題。
	//
	// 所以邊界要記 **(格子, 進入方向)**；「這一格去過沒有」另外記。
	type entry struct {
		at        point
		direction int
	}
	visited := map[point]bool{start: true}
	tried := map[entry]bool{}
	queue := []point{start}
	cells := 1
	terrains[state.DungeonWallRoof] = true
	used := 0
	for len(queue) > 0 && used < steps {
		current := queue[0]
		queue = queue[1:]
		for _, direction := range []int{0, 2, 4, 6} {
			if used >= steps {
				break
			}
			deltaX, deltaY := headingDelta(direction)
			next := point{current.x + deltaX, current.y + deltaY}
			if next.x < 0 || next.x >= geo.Width || next.y < 0 || next.y >= geo.Height {
				continue
			}
			if tried[entry{next, direction}] {
				continue
			}
			tried[entry{next, direction}] = true
			state.SetDungeonGeometryView(current.x, current.y, uint8(direction))
			state.DungeonWallRoof = grid.CellWrapped(current.x, current.y).Terrain
			if !state.CanMoveDungeon(grid, deltaX, deltaY, direction) {
				continue
			}
			used++
			if err := state.MoveDungeon(grid, deltaX, deltaY, direction); err != nil {
				continue
			}
			if !visited[next] {
				visited[next] = true
				cells++
			}
			if err := settleWith(&state, pick); err != nil {
				continue
			}
			terrains[state.DungeonWallRoof] = true
			queue = append(queue, next)
			// 事件把隊伍搬走時要從**落點**繼續，否則靠樓梯才進得去的樓層永遠
			// 走不到（spec 1193）。
			landed := point{state.DungeonX, state.DungeonY}
			if landed != next && !visited[landed] {
				visited[landed] = true
				cells++
				queue = append(queue, landed)
			}
		}
	}
	return block, cells, terrains, nil
}

func gridFor(data gamecorpus.Corpus, block, geoBlock uint8) (geo.Grid, bool) {
	for set, catalog := range data.Geo {
		if grid, ok := catalog.Lookup(geo.MapRef{Set: set, BlockID: geoBlock}); ok {
			return grid, true
		}
	}
	return geo.Grid{}, false
}

func headingDelta(heading int) (int, int) {
	switch heading {
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

// settleWith 把事件／選單／戰鬥推完，直到回到地城模式。`pick` 是選單策略
// （`-1` ＝ 最後一項）。
func settleWith(state *game.State, pick int) error {
	for step := 0; step < 60 && state.Mode != game.ModeDungeon; step++ {
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					return err
				}
			}
			continue
		}
		choice := 0
		if count := len(state.Choices); count > 0 {
			if pick < 0 || pick >= count {
				choice = count - 1
			} else {
				choice = pick
			}
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
