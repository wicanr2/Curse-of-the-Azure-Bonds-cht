// Command route-cells 把主線錄下來的**每一步移動**換成 (ECL 段, 地形碼)。
//
// ★ 存在的理由：可達性報表的「實跑踏到」來自**觀測迴圈**，不是每一步移動——
// 走過去又馬上被劇情推走的格子沒被取樣，所以那一欄是下界（`cmd/cell-reachability`
// 自己標明了這一點）。`COAB_DECISION_LOG` 錄的是 `MoveDungeon` 的**每一次**呼叫，
// 正好補上那個取樣缺口。
//
// ⚠ 錄的是**起點座標**，所以這一支輸出的是「隊伍站上去過的格子」——包含那一步
// 的起點；終點會由下一步的起點涵蓋，最後一步的終點則會漏掉（下界，不是全集）。
//
// 用法：
//
//	COAB_DECISION_LOG=/src/workplace/campaign-frames/route.json \
//	    tools/go.sh test ./internal/game/ -run TestRealNewGameRunsToTheEnding -count=1
//	go run ./cmd/route-cells -route workplace/campaign-frames/route.json \
//	    -cells-json workplace/campaign-frames/route-cells.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gamecorpus"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

type cellRecord struct {
	Block   uint8 `json:"block"`
	Terrain uint8 `json:"terrain"`
}

type decision struct {
	Kind      string `json:"kind"`
	Segment   int    `json:"segment"`
	FromX     int    `json:"from_x"`
	FromY     int    `json:"from_y"`
	Direction int    `json:"direction"`
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "遊戲 image")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "語系檔")
	routePath := flag.String("route", "workplace/campaign-frames/route.json", "主線錄下來的路線")
	cellsOut := flag.String("cells-json", "", "輸出 (block, 地形碼) JSON")
	flag.Parse()

	raw, err := os.ReadFile(*routePath)
	if err != nil {
		log.Fatalf("讀不到路線 %s：%v", *routePath, err)
	}
	var route []decision
	if err := json.Unmarshal(raw, &route); err != nil {
		log.Fatal(err)
	}
	data, err := gamecorpus.Load(*image, *localePath)
	if err != nil {
		log.Fatal(err)
	}

	// 每一段各查一次 GEO：地形分派決定用哪一張圖。
	grids := map[uint8]*geo.Grid{}
	gridFor := func(block uint8) *geo.Grid {
		if grid, done := grids[block]; done {
			return grid
		}
		grids[block] = nil
		payload, ok := data.Blocks[block]
		if !ok {
			return nil
		}
		dispatch := eclcells.Analyze(payload)
		if !dispatch.Found {
			return nil
		}
		for set, catalog := range data.Geo {
			if grid, has := catalog.Lookup(geo.MapRef{Set: set, BlockID: dispatch.GeoBlock}); has {
				grids[block] = &grid
				return grids[block]
			}
		}
		return nil
	}

	seen := map[cellRecord]bool{}
	moves, missing := 0, map[uint8]int{}
	for _, step := range route {
		if step.Kind != "move" {
			continue
		}
		moves++
		block := uint8(step.Segment)
		grid := gridFor(block)
		if grid == nil {
			missing[block]++
			continue
		}
		if step.FromX < 0 || step.FromX >= geo.Width || step.FromY < 0 || step.FromY >= geo.Height {
			continue
		}
		seen[cellRecord{Block: block, Terrain: grid.CellWrapped(step.FromX, step.FromY).Terrain}] = true
	}

	records := make([]cellRecord, 0, len(seen))
	for record := range seen {
		records = append(records, record)
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
	blocks := make([]string, 0, len(missing))
	for block := range missing {
		blocks = append(blocks, fmt.Sprintf("0x%02X×%d", block, missing[block]))
	}
	sort.Strings(blocks)
	fmt.Fprintf(os.Stderr, "moves=%d cells=%d 查不到圖的段=%v\n", moves, len(records), blocks)
}
