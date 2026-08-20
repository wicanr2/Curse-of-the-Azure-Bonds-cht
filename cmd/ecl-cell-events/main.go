// ecl-cell-events 產生「哪一格演哪一場」的全遊戲對照表。
//
// ★ 原作每個地城 block 的每格事件都是同一個形狀：
//
//	AND     C04F, 0x7F → 7F79     { C04F 就是那一格的地形碼 }
//	ON GOTO 7F79                   { 0 起算，每個索引一支處理常式 }
//
// 所以把 `ON GOTO` 的表拆開、再拿 GEO 的地形碼反查格子，就得到完整對照。
// **這比在地圖上逐格走便宜太多**：那些圖不是連通的一整片，走錯一步就被送回
// 世界地圖，而每次試探要重跑一整條 session。
//
// ⚠ 對照表只說「站上那一格會跳到哪支處理常式」。處理常式自己可能還有守衛
// （once-only 旗標、`RANDOM`、前置劇情、`SEARCH`），所以「有對照」不等於
// 「走過去就會演」。
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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
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
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		report.note = "取不到生命週期入口"
		return report
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		report.note = "走訪失敗"
		return report
	}
	unique := map[int]ecl.Instruction{}
	for _, instruction := range graph.Instructions {
		unique[instruction.Offset] = instruction
	}
	offsets := make([]int, 0, len(unique))
	for offset := range unique {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)

	dispatcher, mask, ok := findTerrainDispatcher(unique, offsets)
	if !ok {
		report.note = "沒有以地形碼分派的每格事件"
		if hasTableDispatcher(unique, offsets) {
			// ⚠ 第二種形狀：`GETTABLE <表>, <索引>, <目的地>` ＋ `ON GOTO`。
			// 那是「位置 → 查表 → 索引」，要先解出表的內容與索引怎麼算。
			report.note += "；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）"
		}
		return report
	}
	report.found = true
	report.geoSet = uint8(member)
	report.geoBlock, _ = firstGeometryBlock(unique, offsets)
	report.mask = mask

	targets := decodeTable(data, unique, offsets, dispatcher)
	terrainCells := map[int][]string{}
	if catalog, hasCatalog := catalogs[report.geoSet]; hasCatalog {
		if grid, hasGrid := catalog.Lookup(geo.MapRef{Set: report.geoSet, BlockID: report.geoBlock}); hasGrid {
			for y := 0; y < geo.Height; y++ {
				for x := 0; x < geo.Width; x++ {
					index := int(grid.CellWrapped(x, y).Terrain) & mask
					terrainCells[index] = append(terrainCells[index],
						fmt.Sprintf("(%d,%d)", x, y))
				}
			}
		}
	}
	for index, target := range targets {
		// 索引 0 是「沒有事件的地面」，全圖大半都是它，列出來只是雜訊。
		cells := []string(nil)
		if index != 0 {
			cells = terrainCells[index]
		}
		report.events = append(report.events, cellEvent{
			index: index, text: firstTextAt(unique, offsets, target), cells: cells,
		})
	}
	return report
}

// findTerrainDispatcher 找「`AND C04F, <遮罩>` 之後緊接著的 `ON GOTO`」，
// 並回傳那個遮罩。⚠ 遮罩不是固定的：目前量到 `0x7F` 與 `0x3F` 兩種。
func findTerrainDispatcher(unique map[int]ecl.Instruction, offsets []int) (int, int, bool) {
	for index, offset := range offsets {
		instruction := unique[offset]
		if instruction.Command.Name != "AND" || len(instruction.Operands) < 3 {
			continue
		}
		if !mentions(instruction, 0xC04F) {
			continue
		}
		mask := -1
		for _, operand := range instruction.Operands[:2] {
			value := int(operand.Low)
			if operand.WordSet {
				value = int(operand.Word)
			}
			if value != 0xC04F {
				mask = value
			}
		}
		if mask <= 0 {
			continue
		}
		for cursor := index + 1; cursor < len(offsets) && cursor <= index+2; cursor++ {
			if unique[offsets[cursor]].Command.Opcode == 0x25 {
				return offsets[cursor], mask, true
			}
		}
	}
	return 0, 0, false
}

// hasTableDispatcher 判斷這個 block 是不是用 `GETTABLE` ＋ `ON GOTO` 分派。
func hasTableDispatcher(unique map[int]ecl.Instruction, offsets []int) bool {
	for index, offset := range offsets {
		if unique[offset].Command.Name != "GETTABLE" {
			continue
		}
		for cursor := index + 1; cursor < len(offsets) && cursor <= index+2; cursor++ {
			if unique[offsets[cursor]].Command.Opcode == 0x25 {
				return true
			}
		}
	}
	return false
}

func mentions(instruction ecl.Instruction, value uint16) bool {
	for _, operand := range instruction.Operands {
		if operand.WordSet && operand.Word == value {
			return true
		}
		if !operand.WordSet && uint16(operand.Low) == value {
			return true
		}
	}
	return false
}

// decodeTable 把 `ON GOTO` 的版面拆成索引 → 目的地位移。
func decodeTable(data []byte, unique map[int]ecl.Instruction, offsets []int, offset int) map[int]int {
	end := unique[offset].Next
	for _, candidate := range offsets {
		if candidate > offset {
			end = candidate
			break
		}
	}
	// ⚠ 走訪用的位移與 payload 差兩個位元組（區塊標頭）。
	low, high := offset+2, end+2
	if low < 0 || high > len(data) || high <= low {
		return nil
	}
	raw := data[low:high]
	cursor := 1
	cursor += operandWidth(raw[cursor:])
	if cursor+1 >= len(raw) {
		return nil
	}
	count := int(raw[cursor+1])
	cursor += operandWidth(raw[cursor:])
	targets := map[int]int{}
	for index := 0; index < count && cursor+2 < len(raw); index++ {
		targets[index] = int(raw[cursor+1]) | int(raw[cursor+2])<<8 - ecl.CodeAddressBase
		cursor += 3
	}
	return targets
}

func operandWidth(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	if raw[0] == 0x00 {
		return 2
	}
	return 3
}

// firstGeometryBlock 回傳這個 block 第一個 `LOAD FILES` 的第一個運算元，
// 也就是它載的地圖區塊（spec：`SEG-02`）。
func firstGeometryBlock(unique map[int]ecl.Instruction, offsets []int) (uint8, bool) {
	for _, offset := range offsets {
		instruction := unique[offset]
		if instruction.Command.Opcode != 0x21 || len(instruction.Operands) == 0 {
			continue
		}
		operand := instruction.Operands[0]
		if operand.WordSet {
			return uint8(operand.Word), true
		}
		return operand.Low, true
	}
	return 0, false
}

func firstTextAt(unique map[int]ecl.Instruction, offsets []int, start int) string {
	index := sort.SearchInts(offsets, start)
	for cursor := index; cursor < len(offsets) && cursor < index+12; cursor++ {
		for _, operand := range unique[offsets[cursor]].Operands {
			if len(operand.Packed) > 0 {
				text := ecl.DecodePackedText(operand.Packed)
				text = strings.ReplaceAll(text, "|", "／")
				if len([]rune(text)) > 56 {
					text = string([]rune(text)[:56]) + "…"
				}
				return text
			}
		}
	}
	return ""
}

func render(reports []blockReport) string {
	var out strings.Builder
	out.WriteString("# 每格事件對照表（哪一格演哪一場）\n\n" +
		"由 `cmd/ecl-cell-events` 產生，不要手改。\n\n" +
		"原作地城 block 的每格事件是 `AND C04F, 0x7F` ＋ `ON GOTO`：**地形碼就是索引**\n" +
		"（0 起算）。這張表把 `ON GOTO` 的目的地與 GEO 的地形碼 join 起來。\n\n" +
		"⚠ 有對照**不等於**走過去就會演：處理常式自己可能還有守衛（once-only 旗標、\n" +
		"`RANDOM`、前置劇情、`SEARCH`）。逐格實際驗過的清單見\n" +
		"`docs/plan/seg-22-chapter6-cell-sweep-report.md`。\n\n" +
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
