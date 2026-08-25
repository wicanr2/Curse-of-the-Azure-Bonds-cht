// ecl-window 印出某個 ECL block 裡「某條指令前後」的指令窗，用來看某個轉移或
// 副作用是被什麼條件守住的。
//
// ★ 存在的理由：在地圖上逐格試探「哪一格會觸發轉移」很貴——墓園那張圖不是連通
// 的一整片，走錯一步就被送回世界地圖。守衛條件寫在 bytecode 裡，讀它比撞它便宜。
//
// ⚠ 走訪用 `ecl.TraceGraph`，它會跟 `ON GOTO`／`ON GOSUB` 的每一個目的地。
// 指令窗依 payload 位移排序，所以「前面幾條」是位址上的前面，不保證是執行順序。
//
// 用法：
//
//	go run ./cmd/ecl-window -member ECL6.DAX -block 40 -opcode 20
//	go run ./cmd/ecl-window -member ECL6.DAX -block 40 -opcode 20 -operand 42 -before 12
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

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	member := flag.String("member", "ECL6.DAX", tooltext.Text("h.7fc9c661ba53"))
	block := flag.String("block", "40", tooltext.Text("h.2713767dcc40"))
	opcode := flag.String("opcode", "20", tooltext.Text("h.a707fee6b434"))
	operand := flag.String("operand", "", tooltext.Text("h.6bf8bcc70c7f"))
	text := flag.String("text", "", tooltext.Text("h.b8048ccd0252"))
	into := flag.String("into", "", tooltext.Text("h.50dca18a4330"))
	table := flag.String("table", "", tooltext.Text("h.34b9ecb021e2"))
	getTable := flag.String("gettable", "", tooltext.Text("h.8667ce09b72d"))
	getCount := flag.Int("gettable-count", 32, tooltext.Text("h.9e1e87318ff2"))
	at := flag.String("at", "", tooltext.Text("h.7f56596244b1"))
	raw := flag.Bool("raw", false, tooltext.Text("h.a44eb63d8142"))
	before := flag.Int("before", 8, tooltext.Text("h.a5f00ce09029"))
	after := flag.Int("after", 2, tooltext.Text("h.13950847a639"))
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	payload := memberPayload(archive, *member)
	if payload == nil {
		log.Fatal(tooltext.Format("h.13356c0db288", *member))
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		log.Fatal(err)
	}
	wanted, err := strconv.ParseUint(strings.TrimPrefix(*block, "0x"), 16, 8)
	if err != nil {
		log.Fatal(err)
	}
	var data []byte
	for _, candidate := range blocks {
		if candidate.Entry.ID == uint8(wanted) {
			data = candidate.Data
		}
	}
	if data == nil {
		log.Fatal(tooltext.Format("h.99998dd4110c", *member, wanted))
	}
	target, err := strconv.ParseUint(strings.TrimPrefix(*opcode, "0x"), 16, 8)
	if err != nil {
		log.Fatal(err)
	}
	var operandValue int = -1
	if *operand != "" {
		parsed, parseErr := strconv.ParseUint(strings.TrimPrefix(*operand, "0x"), 16, 16)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		operandValue = int(parsed)
	}

	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		log.Fatal(err)
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		log.Fatal(err)
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

	if *getTable != "" {
		dumpGetTable(*member, uint8(wanted), data, unique, offsets, *getTable, *getCount)
		return
	}

	if *table != "" {
		wantOffset, parseErr := strconv.ParseInt(strings.TrimPrefix(*table, "0x"), 16, 32)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		instruction, ok := unique[int(wantOffset)]
		if !ok {
			log.Fatal(tooltext.Format("h.e0979e0a37a8", wantOffset))
		}
		end := instruction.Next
		for _, offset := range offsets {
			if offset > int(wantOffset) {
				end = offset
				break
			}
		}
		// ⚠ 走訪用的位移與 payload 差兩個位元組（區塊標頭）。
		raw := data[int(wantOffset)+2 : end+2]
		fmt.Print(tooltext.Format("h.ba85062756eb", *member, wanted, instruction.Command.Name, wantOffset, raw))
		// 版面是：opcode、被分派的運算元、個數、然後每個索引一個 word 目的地。
		cursor := 1
		cursor += operandWidth(raw[cursor:])
		count := int(raw[cursor+1])
		cursor += operandWidth(raw[cursor:])
		for index := 0; cursor+2 < len(raw); index++ {
			target := int(raw[cursor+1]) | int(raw[cursor+2])<<8
			cursor += 3
			payloadOffset := target - ecl.CodeAddressBase
			fmt.Print(tooltext.Format("h.6c0bbe56e1b2", index, payloadOffset, firstTextAt(unique, offsets, payloadOffset)))
			if index+1 >= count {
				break
			}
		}
		return
	}

	if *into != "" {
		wantOffset, parseErr := strconv.ParseInt(strings.TrimPrefix(*into, "0x"), 16, 32)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		sources := map[int]bool{}
		for _, edge := range graph.Edges {
			if edge.To == int(wantOffset) {
				sources[edge.From] = true
			}
		}
		keys := make([]int, 0, len(sources))
		for source := range sources {
			keys = append(keys, source)
		}
		sort.Ints(keys)
		for _, source := range keys {
			index := sort.SearchInts(offsets, source)
			low := index - *before
			if low < 0 {
				low = 0
			}
			fmt.Print(tooltext.Format("h.a484cb752d50", wantOffset, source))
			for cursor := low; cursor <= index && cursor < len(offsets); cursor++ {
				marker := "   "
				if cursor == index {
					marker = "→  "
				}
				fmt.Printf("%s%#04x  %s\n", marker, offsets[cursor], format(unique[offsets[cursor]]))
			}
			fmt.Println()
		}
		fmt.Print(tooltext.Format("h.4e2cf09eeb9b", *member, wanted, wantOffset, len(keys)))
		return
	}

	if *at != "" {
		wantOffset, parseErr := strconv.ParseInt(strings.TrimPrefix(*at, "0x"), 16, 32)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		index := sort.SearchInts(offsets, int(wantOffset))
		if index >= len(offsets) {
			log.Fatal(tooltext.Format("h.a25b0fb3fee9", wantOffset))
		}
		low := index - *before
		if low < 0 {
			low = 0
		}
		high := index + *after
		if high >= len(offsets) {
			high = len(offsets) - 1
		}
		for cursor := low; cursor <= high; cursor++ {
			marker := "   "
			if cursor == index {
				marker = "→  "
			}
			fmt.Printf("%s%#04x  %s\n", marker, offsets[cursor], format(unique[offsets[cursor]]))
		}
		return
	}

	hits := 0
	for index, offset := range offsets {
		instruction := unique[offset]
		if *text != "" {
			if !containsText(instruction, *text) {
				continue
			}
		} else {
			if instruction.Command.Opcode != byte(target) {
				continue
			}
			if operandValue >= 0 && !matchesOperand(instruction, operandValue) {
				continue
			}
		}
		hits++
		fmt.Print(tooltext.Format("h.46f1794923ea", hits, offset, instruction.Command.Name))
		low := index - *before
		if low < 0 {
			low = 0
		}
		high := index + *after
		if high >= len(offsets) {
			high = len(offsets) - 1
		}
		for cursor := low; cursor <= high; cursor++ {
			marker := "   "
			if cursor == index {
				marker = "→  "
			}
			instruction := unique[offsets[cursor]]
			fmt.Printf("%s%#04x  %s\n", marker, offsets[cursor], format(instruction))
			if *raw {
				end := instruction.Next
				if cursor+1 < len(offsets) && offsets[cursor+1] > end {
					end = offsets[cursor+1]
				}
				// ⚠ 走訪用的位移與 payload 之間差兩個位元組（區塊標頭），
				// 所以印原始位元組要補回來。
				low, high := instruction.Offset+2, end+2
				if low >= 0 && high <= len(data) && high > low {
					fmt.Print(tooltext.Format("h.d2f692e229bd", data[low:high]))
				}
			}
		}
		fmt.Println()
	}
	fmt.Print(tooltext.Format("h.181715c2f368", *member, wanted, len(offsets), hits))
}

// operandWidth 回傳一個運算元佔幾個位元組。0x00 是位元組運算元（`00 vv`），
// 其餘（0x01 等）是 word 運算元（`cc lo hi`）。
func operandWidth(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	if raw[0] == 0x00 {
		return 2
	}
	return 3
}

// firstTextAt 從某個位移往後找，回傳最先出現的那一句打包文字。
func firstTextAt(unique map[int]ecl.Instruction, offsets []int, start int) string {
	index := sort.SearchInts(offsets, start)
	for cursor := index; cursor < len(offsets) && cursor < index+12; cursor++ {
		instruction := unique[offsets[cursor]]
		for _, operand := range instruction.Operands {
			if len(operand.Packed) > 0 {
				return fmt.Sprintf("%q", ecl.DecodePackedText(operand.Packed))
			}
		}
	}
	return tooltext.Text("h.5c3ffe87c47a")
}

func containsText(instruction ecl.Instruction, needle string) bool {
	for _, operand := range instruction.Operands {
		if len(operand.Packed) == 0 {
			continue
		}
		if strings.Contains(ecl.DecodePackedText(operand.Packed), needle) {
			return true
		}
	}
	return false
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

// dumpGetTable 拆 `GETTABLE <表>, <索引格>, <目的地>`：表的位址在 block 自己的
// 資料裡（`位址 − 0x8000`），逐個索引就是一個位元組。
//
// ★ 這是每格事件的**第二種分派形狀**。前一種是 `AND C04F, <遮罩>` ＋ `ON GOTO`
// （地形碼直接當索引）；這一種多一層查表，索引來自別的格子（世界地圖是所在地點
// `4C9D`，地城是位置編號）。
//
// ⚠ 表沒有宣告長度。後面緊接的 `ON GOTO` 有個數，可以當上界；沒有 `ON GOTO`
// 時就照 `-gettable-count` 印，多印的部分是相鄰資料，不要當成表的內容。
func dumpGetTable(member string, block uint8, data []byte,
	unique map[int]ecl.Instruction, offsets []int, want string, count int) {
	wantOffset, err := strconv.ParseInt(strings.TrimPrefix(want, "0x"), 16, 32)
	if err != nil {
		log.Fatal(err)
	}
	instruction, ok := unique[int(wantOffset)]
	if !ok {
		log.Fatal(tooltext.Format("h.e0979e0a37a8", wantOffset))
	}
	if instruction.Command.Name != "GETTABLE" || len(instruction.Operands) < 3 {
		log.Fatal(tooltext.Format("h.a4e3b4b7097e", wantOffset, instruction.Command.Name))
	}
	base := int(instruction.Operands[0].Word) - ecl.CodeAddressBase
	source := operandLabel(instruction.Operands[1])
	dest := operandLabel(instruction.Operands[2])
	fmt.Print(tooltext.Format("h.c40c8f04b93d", member, block, wantOffset, base, source, dest))

	// 後面緊接的 `ON GOTO` 決定「查到的值」能對到哪一支處理常式。
	targets := map[int]int{}
	onGotoCount := 0
	for index, offset := range offsets {
		if offset != int(wantOffset) {
			continue
		}
		for cursor := index + 1; cursor < len(offsets) && cursor <= index+2; cursor++ {
			if unique[offsets[cursor]].Command.Opcode != 0x25 {
				continue
			}
			targets, onGotoCount = decodeOnGoto(data, unique, offsets, offsets[cursor])
			fmt.Print(tooltext.Format("h.3753a22d7e7d", offsets[cursor], onGotoCount))
		}
		break
	}
	if onGotoCount > 0 && count > onGotoCount {
		// 表的長度沒有宣告，但值必須落在 `ON GOTO` 的範圍內才有意義。
		fmt.Print(tooltext.Format("h.276ebc1f2eb6", onGotoCount))
	}
	fmt.Println(tooltext.Text("h.963e8c6ab149"))
	fmt.Println("|---:|---:|---|")
	for index := 0; index < count; index++ {
		cell := base + index + 2 // ⚠ payload 與走訪位移差兩個位元組（區塊標頭）
		if cell < 0 || cell >= len(data) {
			break
		}
		value := int(data[cell])
		handler := "—"
		if target, ok := targets[value]; ok {
			text := firstTextAt(unique, offsets, target)
			handler = fmt.Sprintf("%#04x %s", target, text)
		} else if onGotoCount > 0 {
			handler = tooltext.Text("h.f13af05aecb8")
		}
		fmt.Printf("| %d | %d | %s |\n", index, value, handler)
	}
}

// decodeOnGoto 把 `ON GOTO` 的版面拆成索引 → 目的地位移，並回傳宣告的個數。
func decodeOnGoto(data []byte, unique map[int]ecl.Instruction, offsets []int,
	offset int) (map[int]int, int) {
	end := unique[offset].Next
	for _, candidate := range offsets {
		if candidate > offset {
			end = candidate
			break
		}
	}
	low, high := offset+2, end+2
	if low < 0 || high > len(data) || high <= low {
		return nil, 0
	}
	raw := data[low:high]
	cursor := 1
	cursor += operandWidth(raw[cursor:])
	if cursor+1 >= len(raw) {
		return nil, 0
	}
	count := int(raw[cursor+1])
	cursor += operandWidth(raw[cursor:])
	targets := map[int]int{}
	for index := 0; index < count && cursor+2 < len(raw); index++ {
		targets[index] = int(raw[cursor+1]) | int(raw[cursor+2])<<8 - ecl.CodeAddressBase
		cursor += 3
	}
	return targets, count
}

func operandLabel(operand ecl.Operand) string {
	if operand.WordSet {
		return fmt.Sprintf("%04X", operand.Word)
	}
	return fmt.Sprintf("%02X", operand.Low)
}

func format(instruction ecl.Instruction) string {
	parts := make([]string, 0, len(instruction.Operands)+1)
	parts = append(parts, fmt.Sprintf("%-16s", instruction.Command.Name))
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
		if file.Name != member {
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
