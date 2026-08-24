// Command map-atlas 把每一張第一人稱地圖畫成「事件發生在哪一格、從哪幾格走得
// 出去」的對照圖。
//
// ★ 存在的理由：先前每次要回答「這一格演什麼」「這張圖怎麼離開」都得再讀一次
// ECL。那件事**已經有答案了**，只是散在三個地方——`ON GOTO` 的分派表
// （`internal/eclcells`）、GEO 的移動遮罩（`internal/geo`）與 game pack 宣告的
// 出口。這一支把三份接起來畫成一張圖，之後就不必為了同一個問題重跑反組譯。
//
// 圖怎麼讀：
//
//	.       沒有事件的地面
//	1..9    每格事件的索引（≥10 用 a..z）
//	#       四面都走不出去（走路站不上去的格子）
//	^v<>    這一格往那個方向**走得出地圖**（圖緣交接）
//	@       game pack 宣告的進場錨點
//
// ⚠ 「有索引」不等於「站上去就會演」：處理常式自己可能還有守衛。實際演出來什麼
// 由 `cmd/cell-sweep` 量，這一支只說**位置**。
//
// 用法：
//
//	go run ./cmd/map-atlas -out docs/reference/maps
//	go run ./cmd/map-atlas -only tilverton.fire-knife-hideout.first-person
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/eclcells"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// headings 是四個基本方向與它們的位移；順序固定，輸出才可重現。
var headings = []struct {
	direction int
	deltaX    int
	deltaY    int
	mark      byte
	name      string
}{
	{0, 0, -1, '^', "北"},
	{2, 1, 0, '>', "東"},
	{4, 0, 1, 'v', "南"},
	{6, -1, 0, '<', "西"},
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "遊戲 image zip")
	out := flag.String("out", "docs/reference/maps", "輸出目錄")
	only := flag.String("only", "", "只畫這一張地圖（pack 的 map id）")
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

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	written, skipped := 0, 0
	for _, definition := range pack.Maps {
		if definition.Kind != "first_person" || definition.GeometryFile == "" {
			continue
		}
		if *only != "" && definition.ID != *only {
			continue
		}
		report, renderErr := render(archive, pack, definition)
		if renderErr != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", definition.ID, renderErr)
			skipped++
			continue
		}
		path := filepath.Join(*out, definition.ID+".md")
		if err := os.WriteFile(path, []byte(report), 0o644); err != nil {
			log.Fatal(err)
		}
		written++
	}
	fmt.Fprintf(os.Stderr, "maps=%d skipped=%d → %s\n", written, skipped, *out)
}

func render(archive *zip.ReadCloser, pack *goldenbox.Pack,
	definition goldenbox.MapDefinition) (string, error) {
	grid, err := loadGrid(archive, definition.GeometryFile, definition.GeometryBlock)
	if err != nil {
		return "", err
	}
	dispatch := eclcells.Dispatch{}
	if definition.ScriptBlock != nil {
		block, blockErr := loadECLBlock(archive, definition.AreaID, *definition.ScriptBlock)
		if blockErr == nil {
			dispatch = eclcells.Analyze(block)
		}
	}

	// 地形碼 → 索引。⚠ 遮罩取分派器量到的那一個，不要寫死：`0x7F` 與 `0x3F`
	// 都出現過（`internal/eclcells`）。
	cellIndex := map[[2]int]int{}
	indexCells := map[int][][2]int{}
	if dispatch.Found {
		for y := 0; y < geo.Height; y++ {
			for x := 0; x < geo.Width; x++ {
				cell, _ := grid.Cell(x, y)
				index := int(cell.Terrain) & dispatch.Mask
				if index == 0 {
					continue
				}
				if _, declared := dispatch.Targets[index]; !declared {
					continue
				}
				cellIndex[[2]int{x, y}] = index
				indexCells[index] = append(indexCells[index], [2]int{x, y})
			}
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# %s\n\n", definition.ID)
	fmt.Fprintf(&report, "由 `cmd/map-atlas` 產生，不要手改。"+
		"`GEO%d` 區塊 `0x%02X`", definition.AreaID, definition.GeometryBlock)
	if definition.ScriptBlock != nil {
		fmt.Fprintf(&report, "；腳本 `ECL%d/0x%02X`", definition.AreaID, *definition.ScriptBlock)
	}
	report.WriteString("。\n\n")
	report.WriteString("```text\n")
	report.WriteString(renderGrid(grid, definition, cellIndex))
	report.WriteString("```\n\n")
	report.WriteString("`.` 沒有事件的地面　`#` 四面都走不出去　" +
		"`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點\n\n")

	report.WriteString(renderExits(grid, definition))
	report.WriteString(renderEvents(dispatch, indexCells))
	return report.String(), nil
}

// renderGrid 畫格子本身。索引 ≥10 用 `a` 起算的字母，一格才佔一個字。
func renderGrid(grid geo.Grid, definition goldenbox.MapDefinition,
	cellIndex map[[2]int]int) string {
	var out strings.Builder
	out.WriteString("      ")
	for x := 0; x < geo.Width; x++ {
		fmt.Fprintf(&out, "%2d ", x)
	}
	out.WriteString("\n")
	for y := 0; y < geo.Height; y++ {
		fmt.Fprintf(&out, "y=%2d  ", y)
		for x := 0; x < geo.Width; x++ {
			out.WriteString(" " + string(cellGlyph(grid, definition, cellIndex, x, y)) + " ")
		}
		out.WriteString(" " + edgeMarks(grid, x0(), y) + "\n")
	}
	return out.String()
}

// x0 只是讓 edgeMarks 的簽章讀起來對稱；圖緣標記是逐列算的。
func x0() int { return 0 }

// edgeMarks 回傳這一列有哪幾格走得出地圖，寫在列尾。
func edgeMarks(grid geo.Grid, _ int, y int) string {
	parts := make([]string, 0, 4)
	for x := 0; x < geo.Width; x++ {
		for _, heading := range headings {
			nextX, nextY := x+heading.deltaX, y+heading.deltaY
			if nextX >= 0 && nextX < geo.Width && nextY >= 0 && nextY < geo.Height {
				continue
			}
			if !grid.CanMoveDungeonWrapped(x, y, heading.direction) {
				continue
			}
			parts = append(parts, fmt.Sprintf("(%d,%d)%c", x, y, heading.mark))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "← 走得出去：" + strings.Join(parts, " ")
}

func cellGlyph(grid geo.Grid, definition goldenbox.MapDefinition,
	cellIndex map[[2]int]int, x, y int) byte {
	if definition.Spawn != nil && definition.Spawn.X == x && definition.Spawn.Y == y {
		return '@'
	}
	if index, ok := cellIndex[[2]int{x, y}]; ok {
		if index < 10 {
			return byte('0' + index)
		}
		if index-10 < 26 {
			return byte('a' + index - 10)
		}
		return '+'
	}
	for _, heading := range headings {
		if grid.CanMoveDungeonWrapped(x, y, heading.direction) {
			return '.'
		}
	}
	return '#'
}

func renderExits(grid geo.Grid, definition goldenbox.MapDefinition) string {
	declared := map[[3]int]string{}
	for _, exit := range definition.ExternalExits {
		declared[[3]int{exit.X, exit.Y, int(exit.Direction)}] = exit.ID
	}
	type row struct {
		x, y, direction int
		name            string
		id              string
	}
	rows := make([]row, 0, 8)
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			for _, heading := range headings {
				nextX, nextY := x+heading.deltaX, y+heading.deltaY
				if nextX >= 0 && nextX < geo.Width && nextY >= 0 && nextY < geo.Height {
					continue
				}
				if !grid.CanMoveDungeonWrapped(x, y, heading.direction) {
					continue
				}
				rows = append(rows, row{x, y, heading.direction, heading.name,
					declared[[3]int{x, y, heading.direction}]})
			}
		}
	}
	var out strings.Builder
	out.WriteString("## 離開這張圖的格子\n\n")
	if len(rows) == 0 {
		out.WriteString("這張圖沒有走得出去的邊界格。\n\n")
		return out.String()
	}
	out.WriteString("⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，" +
		"remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。\n\n")
	out.WriteString("| 格子 | 方向 | game pack 宣告 |\n|---|---|---|\n")
	for _, item := range rows {
		mark := "**沒宣告**"
		if item.id != "" {
			mark = "`" + item.id + "`"
		}
		fmt.Fprintf(&out, "| `(%d,%d)` | %s | %s |\n", item.x, item.y, item.name, mark)
	}
	out.WriteString("\n")
	return out.String()
}

func renderEvents(dispatch eclcells.Dispatch, indexCells map[int][][2]int) string {
	var out strings.Builder
	out.WriteString("## 每格事件\n\n")
	if !dispatch.Found {
		out.WriteString("這張圖沒有以地形碼分派的每格事件。\n")
		return out.String()
	}
	out.WriteString("⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛" +
		"（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看" +
		"[`cell-sweep`](../../audit/cell-sweep.md)。\n\n")
	out.WriteString("| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |\n|---|---:|---|---|---|\n")
	for _, index := range dispatch.Indexes {
		if index == 0 {
			continue
		}
		cells := indexCells[index]
		sort.Slice(cells, func(left, right int) bool {
			if cells[left][1] != cells[right][1] {
				return cells[left][1] < cells[right][1]
			}
			return cells[left][0] < cells[right][0]
		})
		names := make([]string, 0, len(cells))
		for _, cell := range cells {
			names = append(names, fmt.Sprintf("`(%d,%d)`", cell[0], cell[1]))
		}
		where := strings.Join(names, "、")
		if where == "" {
			where = "圖上沒有這個地形碼"
		}
		glyph := string(rune(cellGlyphForIndex(index)))
		text := dispatch.Texts[index]
		if text == "" {
			text = "—"
		}
		guard := dispatch.Guards[index]
		if guard == "" {
			guard = "—"
		}
		fmt.Fprintf(&out, "| `%s` | %d | %s | `%s` | %s |\n", glyph, index, where, guard, text)
	}
	return out.String()
}

func cellGlyphForIndex(index int) byte {
	if index < 10 {
		return byte('0' + index)
	}
	if index-10 < 26 {
		return byte('a' + index - 10)
	}
	return '+'
}

func loadGrid(archive *zip.ReadCloser, member string, block uint8) (geo.Grid, error) {
	payload := memberBytes(archive, member)
	if payload == nil {
		return geo.Grid{}, fmt.Errorf("image 裡沒有 %s", member)
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		return geo.Grid{}, err
	}
	for _, candidate := range blocks {
		if candidate.Entry.ID != block {
			continue
		}
		return geo.Parse(candidate.Entry.ID, candidate.Data)
	}
	return geo.Grid{}, fmt.Errorf("%s 沒有區塊 0x%02X", member, block)
}

func loadECLBlock(archive *zip.ReadCloser, area, block uint8) ([]byte, error) {
	payload := memberBytes(archive, fmt.Sprintf("ECL%d.DAX", area))
	if payload == nil {
		return nil, fmt.Errorf("image 裡沒有 ECL%d.DAX", area)
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		return nil, err
	}
	for _, candidate := range blocks {
		if candidate.Entry.ID == block {
			return candidate.Data, nil
		}
	}
	return nil, fmt.Errorf("ECL%d.DAX 沒有區塊 0x%02X", area, block)
}

func memberBytes(archive *zip.ReadCloser, name string) []byte {
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
