// Package eclcells 解讀 ECL block 的「每格事件分派器」。
//
// 原作地城 block 的每格事件都是同一個形狀：
//
//	AND     C04F, <遮罩> → 7F79     { C04F 就是那一格的地形碼 }
//	ON GOTO 7F79                     { 0 起算，每個索引一支處理常式 }
//
// 拆開 `ON GOTO` 的表就得到「地形碼 → 處理常式」的完整對照。
//
// ⚠ 遮罩**不是固定的**：目前量到 `0x7F` 與 `0x3F` 兩種。寫死一種會讓另一種
// 整批 block 看起來「沒有每格事件」——那是假零，不是沒有內容。
//
// ⚠ 第二種分派形狀是 `GETTABLE <表>, <索引>` ＋ `ON GOTO`（位置查表），本套件
// 只把它標出來（TableForm），沒有解讀。
package eclcells

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// Dispatch 是一個 block 的每格事件分派結果。
type Dispatch struct {
	// Found 為真代表找到以地形碼分派的每格事件。
	Found bool
	// TableForm 為真代表這個 block 改用 `GETTABLE` ＋ `ON GOTO` 查表分派。
	// 它與 Found 互斥：兩者都為假就是這個 block 沒有每格事件。
	TableForm bool
	// Mask 是 `AND C04F, <遮罩>` 的遮罩。
	Mask int
	// GeoBlock 是這個 block 第一個 `LOAD FILES` 載的地圖區塊。
	GeoBlock uint8
	// Indexes 是 `ON GOTO` 表的索引，**升冪排序**。表本身是 0..N-1 連號，
	// 排序是為了讓產出的對照表每次跑都一樣（map 走訪順序是隨機的）。
	Indexes []int
	// Targets 是索引 → 處理常式的起始位移。
	Targets map[int]int
	// Texts 是索引 → 那支處理常式的第一句原作文字（沒有就是空字串）。
	//
	// ⚠ 這是**依位址往後找**的推測，不是執行結果：處理常式開頭可能是個
	// `GOTO`，往後找就會撈到位址上相鄰、執行上無關的另一支的台詞。真的演了
	// 什麼要跑 VM 量（`cmd/cell-sweep`），兩者不一致時以實測為準。
	Texts map[int]string
	// Guards 是索引 → 處理常式開頭在講話之前先做的判斷，用來回答「這一格
	// 為什麼站上去沒反應」。空字串代表一進去就講話，沒有守衛。
	Guards map[int]string
}

// maxTextLookahead 是「往下找幾條指令算第一句話」的上限。處理常式的開頭常常
// 先設幾個旗標才講話，但找太遠會撈到下一支處理常式的台詞。
const maxTextLookahead = 12

// maxTextRunes 是對照表裡每一句話的截斷長度。
const maxTextRunes = 56

// Analyze 走訪一個 block 的五個生命週期入口，找出每格事件的分派器。
func Analyze(data []byte) Dispatch {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return Dispatch{}
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return Dispatch{}
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
		return Dispatch{TableForm: hasTableDispatcher(unique, offsets)}
	}
	out := Dispatch{Found: true, Mask: mask}
	out.GeoBlock, _ = firstGeometryBlock(unique, offsets)
	out.Targets = decodeTable(data, unique, offsets, dispatcher)
	out.Texts, out.Guards = map[int]string{}, map[int]string{}
	for index, target := range out.Targets {
		out.Indexes = append(out.Indexes, index)
		out.Texts[index] = firstTextAt(unique, offsets, target)
		out.Guards[index] = guardAt(unique, offsets, target)
	}
	sort.Ints(out.Indexes)
	return out
}

// findTerrainDispatcher 找「`AND C04F, <遮罩>` 之後緊接著的 `ON GOTO`」，
// 並回傳那個遮罩。
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

// maxGuardInstructions 是守衛欄列出的指令數上限。
const maxGuardInstructions = 4

// guardAt 回傳處理常式在講話之前先做的判斷。掃到第一條帶文字的指令就停——
// 那之後是內容，不是守衛。
func guardAt(unique map[int]ecl.Instruction, offsets []int, start int) string {
	index := sort.SearchInts(offsets, start)
	parts := make([]string, 0, maxGuardInstructions)
	for cursor := index; cursor < len(offsets) && len(parts) < maxGuardInstructions; cursor++ {
		instruction := unique[offsets[cursor]]
		if hasText(instruction) {
			break
		}
		parts = append(parts, compact(instruction))
	}
	return strings.Join(parts, " / ")
}

func hasText(instruction ecl.Instruction) bool {
	for _, operand := range instruction.Operands {
		if len(operand.Packed) > 0 {
			return true
		}
	}
	return false
}

// compact 把一條指令排成一行，運算元一律十六進位。
func compact(instruction ecl.Instruction) string {
	parts := make([]string, 0, len(instruction.Operands)+1)
	parts = append(parts, instruction.Command.Name)
	for _, operand := range instruction.Operands {
		if operand.WordSet {
			parts = append(parts, fmt.Sprintf("%04X", operand.Word))
			continue
		}
		parts = append(parts, fmt.Sprintf("%02X", operand.Low))
	}
	return strings.Join(parts, " ")
}

func firstTextAt(unique map[int]ecl.Instruction, offsets []int, start int) string {
	index := sort.SearchInts(offsets, start)
	for cursor := index; cursor < len(offsets) && cursor < index+maxTextLookahead; cursor++ {
		for _, operand := range unique[offsets[cursor]].Operands {
			if len(operand.Packed) == 0 {
				continue
			}
			text := ecl.DecodePackedText(operand.Packed)
			text = strings.ReplaceAll(text, "|", "／")
			if len([]rune(text)) > maxTextRunes {
				text = string([]rune(text)[:maxTextRunes]) + "…"
			}
			return text
		}
	}
	return ""
}
