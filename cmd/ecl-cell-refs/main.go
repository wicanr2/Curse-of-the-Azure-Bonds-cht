// ecl-cell-refs 回答「誰讀寫這個 ECL 記憶體格子」。
//
// ★ 存在的理由：追一個旗標為什麼是現在這個值，靠人去 grep 每個 block 太貴，
// 而拿 `docs/audit/ecl-event-catalog.json` 查會**少報**——那份目錄的指令表有
// 可達性缺口（`ECL6/0x43` 廚房那支 `SAVE 01 4C06` 就不在裡面）。這支直接走
// `ecl.TraceGraph`，跟 `cmd/ecl-window` 同一條路。
//
// ⚠ 走訪仍然是「從五個生命週期入口跟著跳躍走」。跟不到的碼不會出現在結果裡，
// 所以**查無存取代表這條路沒找到，不代表不存在**。
//
// 用法：
//
//	go run ./cmd/ecl-cell-refs -cell 4C06
//	go run ./cmd/ecl-cell-refs -range 4C00-4C1F
//	go run ./cmd/ecl-cell-refs -range 4C00-4C1F -before-load-files
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// reference 是一處對某個格子的存取。
type reference struct {
	member   int
	block    uint8
	offset   int
	name     string
	operands string
	cell     uint16
	// beforeLoadFiles 為真代表這一處在該 block 的第一個 `LOAD FILES` 之前。
	beforeLoadFiles bool
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	cell := flag.String("cell", "", tooltext.Text("h.72f6fff50970"))
	cellRange := flag.String("range", "", tooltext.Text("h.cee99cc2005d"))
	onlyBefore := flag.Bool("before-load-files", false,
		tooltext.Text("h.3394d9c8cd7c"))
	flag.Parse()

	low, high, err := parseTargets(*cell, *cellRange)
	if err != nil {
		log.Fatal(err)
	}
	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	refs, err := collect(archive, low, high)
	if err != nil {
		log.Fatal(err)
	}
	byCell := map[uint16][]reference{}
	for _, ref := range refs {
		if *onlyBefore && !ref.beforeLoadFiles {
			continue
		}
		byCell[ref.cell] = append(byCell[ref.cell], ref)
	}
	cells := make([]int, 0, len(byCell))
	for value := range byCell {
		cells = append(cells, int(value))
	}
	sort.Ints(cells)
	total := 0
	for _, value := range cells {
		list := byCell[uint16(value)]
		total += len(list)
		fmt.Print(tooltext.Format("h.f999a736c08f", value, len(list)))
		for _, ref := range list {
			marker := " "
			if ref.beforeLoadFiles {
				marker = "!" // 在 LOAD FILES 之前
			}
			fmt.Printf("  %s ECL%d/0x%02X %#06x  %-14s %s\n",
				marker, ref.member, ref.block, ref.offset, ref.name, ref.operands)
		}
	}
	fmt.Print(tooltext.Format("h.8147a75e678d", total))
}

func parseTargets(cell, cellRange string) (uint16, uint16, error) {
	switch {
	case cellRange != "":
		parts := strings.SplitN(cellRange, "-", 2)
		if len(parts) != 2 {
			return 0, 0, tooltext.Errorf("h.a207ba3d940f", cellRange)
		}
		low, err := strconv.ParseUint(strings.TrimPrefix(parts[0], "0x"), 16, 16)
		if err != nil {
			return 0, 0, err
		}
		high, err := strconv.ParseUint(strings.TrimPrefix(parts[1], "0x"), 16, 16)
		if err != nil {
			return 0, 0, err
		}
		return uint16(low), uint16(high), nil
	case cell != "":
		value, err := strconv.ParseUint(strings.TrimPrefix(cell, "0x"), 16, 16)
		if err != nil {
			return 0, 0, err
		}
		return uint16(value), uint16(value), nil
	}
	return 0, 0, tooltext.Errorf("h.1f6196ba6530")
}

func collect(archive *zip.ReadCloser, low, high uint16) ([]reference, error) {
	refs := make([]reference, 0, 64)
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			return nil, tooltext.Errorf("h.f74a43cb81d6", member)
		}
		blocks, err := dax.Parse(payload)
		if err != nil {
			return nil, err
		}
		for _, block := range blocks {
			found, err := scanBlock(member, block.Entry.ID, block.Data, low, high)
			if err != nil {
				return nil, fmt.Errorf("ECL%d/0x%02X: %w", member, block.Entry.ID, err)
			}
			refs = append(refs, found...)
		}
	}
	return refs, nil
}

func scanBlock(member int, id uint8, data []byte, low, high uint16) ([]reference, error) {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil, err
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

	// LOAD FILES（`21h`）是那一段載地圖檔的地方。旗標在它之前還是之後被寫，
	// 決定「換地圖時整批換掉」這個模型站不站得住。
	loadFiles := -1
	for _, offset := range offsets {
		if unique[offset].Command.Opcode == 0x21 {
			loadFiles = offset
			break
		}
	}
	refs := make([]reference, 0, 8)
	for _, offset := range offsets {
		instruction := unique[offset]
		// 同一條指令可能提到同一格兩次（`MULTIPLY 4C06 04 4C06`），只算一處。
		counted := map[uint16]bool{}
		for _, operand := range instruction.Operands {
			if !operand.WordSet || operand.Word < low || operand.Word > high {
				continue
			}
			if counted[operand.Word] {
				continue
			}
			counted[operand.Word] = true
			refs = append(refs, reference{
				member: member, block: id, offset: offset,
				name:     instruction.Command.Name,
				operands: format(instruction),
				cell:     operand.Word,
				// LOAD FILES 找不到時一律當成「不在它之前」，不要無中生有。
				beforeLoadFiles: loadFiles >= 0 && offset < loadFiles,
			})
		}
	}
	return refs, nil
}

func format(instruction ecl.Instruction) string {
	parts := make([]string, 0, len(instruction.Operands))
	for _, operand := range instruction.Operands {
		switch {
		case len(operand.Packed) > 0:
			text := ecl.DecodePackedText(operand.Packed)
			if len([]rune(text)) > 32 {
				text = string([]rune(text)[:32]) + "…"
			}
			parts = append(parts, fmt.Sprintf("%q", text))
		case operand.WordSet:
			parts = append(parts, fmt.Sprintf("%04X", operand.Word))
		default:
			parts = append(parts, fmt.Sprintf("%02X", operand.Low))
		}
	}
	return strings.Join(parts, " ")
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
