// ecl-cell-events 產生「哪一格演哪一場」的全遊戲對照表。
//
// ★ 原作每個地城 block 的每格事件都是同一個形狀：`AND C04F, <遮罩>` ＋
// `ON GOTO`。分派器的解讀在 `internal/eclcells`；這支只負責把它與 GEO 的地形碼
// join 起來、排版成 markdown。
//
// **這比在地圖上逐格走便宜太多**：那些圖不是連通的一整片，走錯一步就被送回
// 世界地圖，而每次試探要重跑一整條 session。
//
// ⚠ 對照表只說「站上那一格會跳到哪支處理常式」。處理常式自己可能還有守衛
// （once-only 旗標、`RANDOM`、前置劇情、`SEARCH`），所以「有對照」不等於
// 「走過去就會演」。實際站上去演了什麼由 `cmd/cell-sweep` 量。
//
// 用法：
//
//	go run ./cmd/ecl-cell-events
//	go run ./cmd/ecl-cell-events -out docs/audit/ecl-cell-events.md
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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

type cellEvent struct {
	index int
	text  string
	cells []string
}

type blockReport struct {
	member   int
	block    uint8
	geoSet   uint8
	geoBlock uint8
	mask     int
	found    bool
	note     string
	events   []cellEvent
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	out := flag.String("out", "docs/audit/ecl-cell-events.md", "輸出的 markdown")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	catalogs := map[uint8]geo.Catalog{}
	for set := 1; set <= 6; set++ {
		payload := memberPayload(archive, fmt.Sprintf("GEO%d.DAX", set))
		if payload == nil {
			continue
		}
		catalog := geo.NewCatalog()
		if err := catalog.AddDAX(uint8(set), payload); err != nil {
			log.Fatalf("GEO%d: %v", set, err)
		}
		catalogs[uint8(set)] = catalog
	}

	reports := make([]blockReport, 0, 25)
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			log.Fatalf("image 裡沒有 ECL%d.DAX", member)
		}
		blocks, err := dax.Parse(payload)
		if err != nil {
			log.Fatal(err)
		}
		for _, block := range blocks {
			reports = append(reports, describe(catalogs, member, block.Entry.ID, block.Data))
		}
	}
	sort.Slice(reports, func(a, b int) bool {
		if reports[a].member != reports[b].member {
			return reports[a].member < reports[b].member
		}
		return reports[a].block < reports[b].block
	})
	if err := os.WriteFile(*out, []byte(render(reports)), 0o644); err != nil {
		log.Fatal(err)
	}
	dispatchers, mapped := 0, 0
	for _, report := range reports {
		if !report.found {
			continue
		}
		dispatchers++
		for _, event := range report.events {
			if len(event.cells) > 0 {
				mapped++
			}
		}
	}
	fmt.Printf("block=%d 有地形分派=%d 對到格子的場景=%d → %s\n",
		len(reports), dispatchers, mapped, *out)
}

func describe(catalogs map[uint8]geo.Catalog, member int, id uint8, data []byte) blockReport {
	report := blockReport{member: member, block: id}
	dispatch := eclcells.Analyze(data)
	if !dispatch.Found {
		report.note = "沒有以地形碼分派的每格事件"
		if dispatch.TableForm {
			// ⚠ 第二種形狀：`GETTABLE <表>, <索引>, <目的地>` ＋ `ON GOTO`。
			// 那是「位置 → 查表 → 索引」，要先解出表的內容與索引怎麼算。
			report.note += "；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）"
		}
		return report
	}
	report.found = true
	report.geoSet = uint8(member)
	report.geoBlock = dispatch.GeoBlock
	report.mask = dispatch.Mask

	terrainCells := terrainIndexCells(catalogs, report.geoSet, report.geoBlock, dispatch.Mask)
	for _, index := range dispatch.Indexes {
		// 索引 0 是「沒有事件的地面」，全圖大半都是它，列出來只是雜訊。
		cells := []string(nil)
		if index != 0 {
			cells = terrainCells[index]
		}
		report.events = append(report.events, cellEvent{
			index: index, text: dispatch.Texts[index], cells: cells,
		})
	}
	return report
}

// terrainIndexCells 把一張地圖的每一格依「地形碼 & 遮罩」分堆。
func terrainIndexCells(catalogs map[uint8]geo.Catalog, set, block uint8, mask int) map[int][]string {
	cells := map[int][]string{}
	catalog, ok := catalogs[set]
	if !ok {
		return cells
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: set, BlockID: block})
	if !ok {
		return cells
	}
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			index := int(grid.CellWrapped(x, y).Terrain) & mask
			cells[index] = append(cells[index], fmt.Sprintf("(%d,%d)", x, y))
		}
	}
	return cells
}

func render(reports []blockReport) string {
	var out strings.Builder
	out.WriteString("# 每格事件對照表（哪一格演哪一場）\n\n" +
		"由 `cmd/ecl-cell-events` 產生，不要手改。\n\n" +
		"原作地城 block 的每格事件是 `AND C04F, <遮罩>` ＋ `ON GOTO`：**地形碼就是\n" +
		"索引**（0 起算）。這張表把 `ON GOTO` 的目的地與 GEO 的地形碼 join 起來。\n\n" +
		"⚠ 有對照**不等於**走過去就會演：處理常式自己可能還有守衛（once-only 旗標、\n" +
		"`RANDOM`、前置劇情、`SEARCH`）。實際站上去演出來的敘述見\n" +
		"`docs/audit/cell-sweep.md`。\n\n" +
		"⚠ 索引 0 是「沒有事件的地面」，格子欄一律留白（那是全圖大半）。\n\n")
	for _, report := range reports {
		out.WriteString(fmt.Sprintf("## ECL%d／`0x%02X`\n\n", report.member, report.block))
		if !report.found {
			out.WriteString(report.note + "\n\n")
			continue
		}
		out.WriteString(fmt.Sprintf("地圖：`GEO%d/0x%02X`；索引 ＝ 地形碼 `& 0x%02X`\n\n",
			report.geoSet, report.geoBlock, report.mask))
		out.WriteString("| 索引 | 遮罩後 | 格子 | 那一場的第一句 |\n|---:|---|---|---|\n")
		for _, event := range report.events {
			cells := "—"
			if len(event.cells) > 0 {
				cells = "`" + strings.Join(event.cells, "`、`") + "`"
			}
			text := event.text
			if text == "" {
				text = "—"
			} else {
				text = "「" + text + "」"
			}
			out.WriteString(fmt.Sprintf("| %d | `%02X` | %s | %s |\n",
				event.index, event.index, cells, text))
		}
		out.WriteString("\n")
	}
	return out.String()
}

func memberPayload(archive *zip.ReadCloser, member string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, member) {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil
		}
		defer reader.Close()
		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}
		return payload
	}
	return nil
}
