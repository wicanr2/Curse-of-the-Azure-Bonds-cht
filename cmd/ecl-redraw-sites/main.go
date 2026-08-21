// ecl-redraw-sites 把全 corpus 可達的 `2Dh CALL 2E10h`（重畫）逐處分類：
// 這一條 `CALL` 之前有沒有寫過隊伍座標，寫的是哪幾格，來源是常數、鄰格加減，
// 還是 `4BF0`／`4BF1`（移動前快照 ⇒ 退回上一格）。
//
// ★ 存在的理由：remake 沒有原作那五個髒旗標，改用「`CALL` 當下回頭掃同 block、
// 執行序在前的座標寫入」。要判斷這個近似在哪裡會和原作分岔，得先有一份**可重生**
// 的清冊——spec 1155／1157 的分類原本是手工做的，數字對不上就查不出是漏了哪一處。
//
// ⚠ 走訪用 `ecl.TraceGraph` 配 `ecl.EntryPoints(data, 5)` 的五個進入點。只種一個
// 進入點會少掉大半個 block（spec 1141／599／610 各踩過一次）。
//
// 用法：
//
//	go run ./cmd/ecl-redraw-sites
//	go run ./cmd/ecl-redraw-sites -kind move    # 只列「明確移動」那一類
//	go run ./cmd/ecl-redraw-sites -window 8
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

const (
	cellX        = 0xC04B
	cellY        = 0xC04C
	cellFacing   = 0xC04D
	preMoveX     = 0x4BF0
	preMoveY     = 0x4BF1
	redrawTarget = 0x2E10
)

// site 是一處 `CALL 2E10h` 與它前面那個視窗裡的座標寫入。
type site struct {
	member  string
	block   uint8
	offset  int
	kind    string
	writes  []string
	guard   string
	preview string
}

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "原版 DOS image ZIP")
	window := flag.Int("window", 6, "往前看幾條指令（spec 1155 用 6）")
	only := flag.String("kind", "", "只列這一類：move／restore／picture／none／facing")
	flag.Parse()

	archive, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	var sites []site
	for member := 1; member <= 6; member++ {
		name := fmt.Sprintf("ECL%d.DAX", member)
		payload := memberPayload(archive, name)
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatalf("%s：%v", name, parseErr)
		}
		for _, block := range blocks {
			sites = append(sites, classifyBlock(name, block.Entry.ID, block.Data, *window)...)
		}
	}

	counts := map[string]int{}
	for _, found := range sites {
		counts[found.kind]++
	}
	for _, found := range sites {
		if *only != "" && found.kind != *only {
			continue
		}
		fmt.Printf("%s/%#02x:%04Xh  %-8s  %s\n", found.member, found.block, found.offset,
			found.kind, strings.Join(found.writes, "; "))
		if found.guard != "" {
			fmt.Printf("        守衛：%s\n", found.guard)
		}
		if found.preview != "" {
			fmt.Printf("        文字：%s\n", found.preview)
		}
	}
	fmt.Printf("\n可達的 CALL 2E10h 共 %d 處：", len(sites))
	kinds := make([]string, 0, len(counts))
	for kind := range counts {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		fmt.Printf("%s=%d ", kind, counts[kind])
	}
	fmt.Println()
}

// classifyBlock 走完一個 block，回報每一處 `CALL 2E10h` 的分類。
func classifyBlock(member string, block uint8, data []byte, window int) []site {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil
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

	var found []site
	for index, offset := range offsets {
		instruction := unique[offset]
		if instruction.Command.Opcode != 0x2D || !matchesOperand(instruction, redrawTarget) {
			continue
		}
		low := index - window
		if low < 0 {
			low = 0
		}
		current := site{member: member, block: block, offset: offset, kind: "none"}
		var wroteFacing, wroteRestore, wroteMove, wrotePicture bool
		for cursor := low; cursor < index; cursor++ {
			previous := unique[offsets[cursor]]
			destination, ok := coordinateDestination(previous)
			if !ok {
				if previous.Command.Opcode == 0x0E {
					wrotePicture = true
				}
				if strings.HasPrefix(previous.Command.Name, "COMPARE") {
					current.guard = format(previous)
				}
				if current.preview == "" {
					current.preview = firstText(previous)
				}
				continue
			}
			current.writes = append(current.writes, format(previous))
			switch destination {
			case cellFacing:
				wroteFacing = true
			default:
				if sourceIsPreMoveSnapshot(previous) {
					wroteRestore = true
				} else {
					wroteMove = true
				}
			}
		}
		switch {
		case wroteFacing:
			current.kind = "facing"
		case wroteRestore:
			current.kind = "restore"
		case wroteMove:
			current.kind = "move"
		case wrotePicture:
			current.kind = "picture"
		}
		found = append(found, current)
	}
	return found
}

// coordinateDestination 回報這一條指令有沒有把值寫進 `C04B`／`C04C`／`C04D`。
// ⚠ 目的地是**最後一個**運算元（`SAVE <來源> <目的地>`、
// `ADD <左> <右> <目的地>`），不是第一個。
func coordinateDestination(instruction ecl.Instruction) (uint16, bool) {
	if len(instruction.Operands) == 0 {
		return 0, false
	}
	// ⚠ 寫入路徑不只 `09h SAVE`。原作全部走 `STOREVALUE`，所以算術、位元、
	// `08h RANDOM` 與 `2Ah GETTABLE` 也都算——`ECL5/0x33:096Ah` 就是一處用
	// 查表把目的地寫進 `C04B` 的傳送。
	switch instruction.Command.Opcode {
	case 0x09, 0x04, 0x05, 0x06, 0x07, 0x08, 0x2A, 0x2F, 0x30:
	default:
		return 0, false
	}
	last := instruction.Operands[len(instruction.Operands)-1]
	if !last.WordSet {
		return 0, false
	}
	switch last.Word {
	case cellX, cellY, cellFacing:
		return last.Word, true
	}
	return 0, false
}

// sourceIsPreMoveSnapshot 認出「退回上一格」：來源是移動前快照 `4BF0`／`4BF1`。
func sourceIsPreMoveSnapshot(instruction ecl.Instruction) bool {
	for _, operand := range instruction.Operands[:len(instruction.Operands)-1] {
		if operand.WordSet && (operand.Word == preMoveX || operand.Word == preMoveY) {
			return true
		}
	}
	return false
}

func firstText(instruction ecl.Instruction) string {
	for _, operand := range instruction.Operands {
		if len(operand.Packed) == 0 {
			continue
		}
		text := ecl.DecodePackedText(operand.Packed)
		if text != "" {
			return text
		}
	}
	return ""
}

func matchesOperand(instruction ecl.Instruction, want int) bool {
	if len(instruction.Operands) == 0 {
		return false
	}
	first := instruction.Operands[0]
	if first.WordSet {
		return int(first.Word) == want
	}
	return int(first.Low) == want
}

func format(instruction ecl.Instruction) string {
	parts := make([]string, 0, len(instruction.Operands)+1)
	parts = append(parts, instruction.Command.Name)
	for _, operand := range instruction.Operands {
		switch {
		case len(operand.Packed) > 0:
			parts = append(parts, fmt.Sprintf("%q", ecl.DecodePackedText(operand.Packed)))
		case operand.WordSet:
			parts = append(parts, fmt.Sprintf("%#04x", operand.Word))
		default:
			parts = append(parts, fmt.Sprintf("%#02x", operand.Low))
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
