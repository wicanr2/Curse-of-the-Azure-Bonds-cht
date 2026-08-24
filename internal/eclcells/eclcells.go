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
// 第二種分派形狀是 `GETTABLE <表>, <索引格>` ＋ `ON GOTO`（位置查表）：索引不是
// 地形碼而是別的格子（世界地圖是路段編號 `4C9D`，spec 1143），中間多一層查表。
// 這一種也解得出來（`analyzeTableForm`），結果放在 `TableForm` 為真的 Dispatch 裡；
// `Found` 只代表「地形碼那一種」，兩者互斥。
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
	// GuardCells 是同一段守衛裡出現過的 `COMPARE <格子> <值>`，拆成結構。
	//
	// ★ 它讓「站上去沒反應」的盤點從**要人去讀守衛**變成可以自動分類：
	// 把這些格子設成守衛比對的值再站一次，演得出來就是「需要前置狀態」，
	// 演不出來才是真的沒接（spec 1177）。
	GuardCells map[int][]GuardCompare
	// TableIndexCell 是查表分派的索引取自哪一格（`TableForm` 為真時才有意義）。
	TableIndexCell uint16
	// TableValues 是查表的內容：`TableValues[i]` 是索引 i 查到的值，也就是
	// `ON GOTO` 的索引。表沒有宣告長度，這裡只讀 tableProbeLength 個。
	TableValues []int
}

// GuardCompare 是守衛裡的一次 `COMPARE <格子> <值>`，連同它後面那條 `IF`。
//
// ★ 只有 `Address`／`Value` 是不夠的：`COMPARE 4C01 01 / IF >= / EXIT` 的意思是
// 「`4C01 >= 1` 就離開」，把 `4C01` 設成 1 反而保證演不出來。要避開離開那一路，
// 得知道比較運算子往哪一邊（spec 1177）。
type GuardCompare struct {
	Address uint16
	Value   uint16
	// Operator 是後面那條 `IF` 的運算子（`=`／`<>`／`<`／`>`／`<=`／`>=`）。
	// 空字串代表這一條 `COMPARE` 後面沒有接得上的 `IF`。
	Operator string
	// ExitsOnMatch 為真代表條件成立時下一條是 `EXIT`——那正是「站上去沒反應」
	// 的形狀。
	ExitsOnMatch bool
}

// AvoidValue 回一個**避開離開那一路**的值。第二個回傳值為 false 代表避不開
// （例如 `IF >= 0`：任何無號值都成立）。
func (g GuardCompare) AvoidValue() (uint16, bool) {
	switch g.Operator {
	case "<>":
		// 不等於就離開 ⇒ 設成相等。
		return g.Value, true
	case ">":
		return g.Value, true
	case "<":
		return g.Value, true
	case "=":
		return g.Value + 1, true
	case "<=":
		return g.Value + 1, true
	case ">=":
		if g.Value == 0 {
			return 0, false
		}
		return g.Value - 1, true
	}
	return 0, false
}

// tableProbeLength 是查表分派要讀幾個索引。表沒有宣告長度，讀太多會撈到相鄰
// 資料——所以只讀到足以看出用法的長度，並在文件裡講明這是探測不是宣告。
const tableProbeLength = 48

// minTableRun 是「表的開頭要連續幾個值落在 `ON GOTO` 範圍內」才認定是查表分派。
const minTableRun = 8

// inRangeRun 回傳表的開頭有連續幾個值小於 size。
func inRangeRun(data []byte, base, size int) int {
	run := 0
	for index := 0; index < tableProbeLength; index++ {
		// ⚠ payload 與走訪位移差兩個位元組（區塊標頭）。
		cell := base + index + 2
		if cell < 0 || cell >= len(data) || int(data[cell]) >= size {
			break
		}
		run++
	}
	return run
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

	dispatcher, mask, ok := findTerrainDispatcher(data, unique, offsets)
	if !ok {
		return analyzeTableForm(data, unique, offsets)
	}
	out := Dispatch{Found: true, Mask: mask}
	out.GeoBlock, _ = firstGeometryBlock(unique, offsets)
	out.Targets = decodeTable(data, unique, offsets, dispatcher)
	out.Texts, out.Guards = map[int]string{}, map[int]string{}
	out.GuardCells = map[int][]GuardCompare{}
	for index, target := range out.Targets {
		out.Indexes = append(out.Indexes, index)
		out.Texts[index] = firstTextAt(unique, offsets, target)
		out.Guards[index] = guardAt(unique, offsets, target)
		out.GuardCells[index] = guardCellsAt(unique, offsets, target)
	}
	sort.Ints(out.Indexes)
	return out
}

// findTerrainDispatcher 找「`AND C04F, <遮罩> → <格子>` 配上分派同一個格子的
// `ON GOTO`」，並回傳那個遮罩。
//
// ⚠ 配對條件是**目的地運算元相同**，不是「緊接在後面」。提爾佛頓下水道
// （`ECL2/0x03`）中間隔了 `DIVIDE` 與 `GETTABLE` 兩條，用「緊接兩條內」去找會
// 整個 block 落空——而落空跟「這個 block 沒有每格事件」長得一模一樣。
func findTerrainDispatcher(data []byte, unique map[int]ecl.Instruction, offsets []int) (int, int, bool) {
	bestTarget, bestMask, bestSize := 0, 0, 0
	for _, offset := range offsets {
		if unique[offset].Command.Opcode != 0x25 {
			continue
		}
		cell, ok := onGotoCell(data, offset)
		if !ok {
			continue
		}
		mask, ok := maskFeeding(data, unique, offsets, cell, offset)
		if !ok {
			continue
		}
		// ⚠ 一個 block 有好幾組 `AND C04F, <遮罩>` ＋ `ON GOTO`：`& 0x80` 那種是
		// 「這一格特別嗎」的位元測試，只有兩三個目的地。**每格事件的分派器是把
		// 地形碼逐個列出來的那一組**，所以取表最大的。
		if size := onGotoSize(data, unique, offsets, offset); size > bestSize {
			bestTarget, bestMask, bestSize = offset, mask, size
		}
	}
	return bestTarget, bestMask, bestSize > 0
}

// maskFeeding 找「在 `before` 之前、離它最近、把地形碼遮罩後寫進 `cell`」的
// `AND`，回傳那個遮罩。
//
// ⚠ 同一格會被好幾條 `AND C04F, <遮罩>` 寫過（`0x3F`、`0x7F`、`0x80` 都有），
// 取最近的那一條才會拿到餵給這個 `ON GOTO` 的遮罩。
//
// ⚠ 「之前」要照**執行順序**算，不是位移大小：遮罩可以藏在 `GOSUB` 進去的
// 子程式裡（見 scanOrderBefore）。
func maskFeeding(data []byte, unique map[int]ecl.Instruction, offsets []int,
	cell uint16, before int) (int, bool) {
	mask, found := 0, false
	for _, offset := range scanOrderBefore(data, unique, offsets, before) {
		instruction := unique[offset]
		if instruction.Command.Name != "AND" || len(instruction.Operands) < 3 {
			continue
		}
		if !mentions(instruction, 0xC04F) {
			continue
		}
		if !instruction.Operands[2].WordSet || instruction.Operands[2].Word != cell {
			continue
		}
		candidate := -1
		for _, operand := range instruction.Operands[:2] {
			value := int(operand.Low)
			if operand.WordSet {
				value = int(operand.Word)
			}
			if value != 0xC04F {
				candidate = value
			}
		}
		if candidate > 0 {
			mask, found = candidate, true
		}
	}
	return mask, found
}

// maxSubroutineBody 是往子程式裡看時最多看幾條指令。子程式一律短（設幾個
// 格子就 `RETURN`），上限只是避免解錯時整段掃下去。
const maxSubroutineBody = 24

// scanOrderBefore 回傳「跑到 `before` 之前執行過的指令」的位移，順序照執行順序
// 近似：線性往下，遇到 `GOSUB` 就把被呼叫的那一段插在它後面。
//
// ★ 為什麼不能只用位移大小比較。 火刀據點（`ECL2/0x04`）的分派器是
// `GOSUB 0x970D` ＋ `ON GOTO 4C05`，而 `AND 3Fh, C04F → 4C05` 住在位移 `0x170D`
// 的子程式裡——位置在 `ON GOTO`（`0x00F5`）**後面** 5,656 個位元組。用「位移比
// `ON GOTO` 小」去找遮罩，這個 block 會落空，而落空跟「這個 block 沒有每格事件」
// 在報表上長得一模一樣。那張表原本就這樣少了 33 個索引的一整層。
//
// ⚠ 只往下看一層：目前語料裡遮罩都在被直接呼叫的那一支，多層遞迴只會多撈到
// 執行上無關的 `AND`。
func scanOrderBefore(data []byte, unique map[int]ecl.Instruction, offsets []int,
	before int) []int {
	order := make([]int, 0, len(offsets))
	for _, offset := range offsets {
		if offset >= before {
			break
		}
		order = append(order, offset)
		instruction := unique[offset]
		if instruction.Command.Opcode != 0x02 || len(instruction.Operands) != 1 {
			continue
		}
		target, ok := ecl.CodeTarget(instruction.Operands[0], len(data))
		if !ok {
			continue
		}
		order = append(order, subroutineBody(unique, offsets, target)...)
	}
	return order
}

// subroutineBody 回傳子程式從 `start` 到第一條 `RETURN`／`EXIT` 為止的位移。
func subroutineBody(unique map[int]ecl.Instruction, offsets []int, start int) []int {
	index := sort.SearchInts(offsets, start)
	body := make([]int, 0, maxSubroutineBody)
	for cursor := index; cursor < len(offsets) && len(body) < maxSubroutineBody; cursor++ {
		offset := offsets[cursor]
		body = append(body, offset)
		opcode := unique[offset].Command.Opcode
		if opcode == 0x13 || opcode == 0x00 {
			break
		}
	}
	return body
}

// onGotoSize 回傳 `ON GOTO` 宣告的索引個數。
func onGotoSize(data []byte, unique map[int]ecl.Instruction, offsets []int, offset int) int {
	end := unique[offset].Next
	for _, candidate := range offsets {
		if candidate > offset {
			end = candidate
			break
		}
	}
	low, high := offset+2, end+2
	if low < 0 || high > len(data) || high <= low {
		return 0
	}
	raw := data[low:high]
	cursor := 1
	cursor += operandWidth(raw[cursor:])
	if cursor+1 >= len(raw) {
		return 0
	}
	return int(raw[cursor+1])
}

// analyzeTableForm 解讀第二種分派形狀：`GETTABLE <表>, <索引格> → <格子>` 配上
// 分派同一格的 `ON GOTO`。索引不是地形碼，而是別的格子（世界地圖是路段編號
// `4C9D`），中間多一層查表。
func analyzeTableForm(data []byte, unique map[int]ecl.Instruction, offsets []int) Dispatch {
	bestGet, bestTarget, bestRun := -1, 0, 0
	for _, offset := range offsets {
		instruction := unique[offset]
		if instruction.Command.Name != "GETTABLE" || len(instruction.Operands) < 3 {
			continue
		}
		if !instruction.Operands[2].WordSet {
			continue
		}
		target, ok := findOnGoto(data, unique, offsets, instruction.Operands[2].Word)
		if !ok || target <= offset {
			continue
		}
		size := onGotoSize(data, unique, offsets, target)
		if size == 0 {
			continue
		}
		// ★ 判準是**查到的值要落在 `ON GOTO` 的範圍內**。同一個 block 有好幾組
		// `GETTABLE` ＋ `ON GOTO`，只有真正餵給這個分派的那張表，每個索引查出來
		// 的值才會是合法的處理常式編號。光比表的大小會挑到不相干的一對。
		run := inRangeRun(data, int(instruction.Operands[0].Word)-ecl.CodeAddressBase, size)
		if run > bestRun {
			bestGet, bestTarget, bestRun = offset, target, run
		}
	}
	// 連續合法值太少就不算查表分派——那多半只是剛好相鄰的兩條指令。
	if bestRun < minTableRun {
		return Dispatch{}
	}
	get := unique[bestGet]
	out := Dispatch{TableForm: true, TableIndexCell: get.Operands[1].Word}
	if !get.Operands[1].WordSet {
		out.TableIndexCell = uint16(get.Operands[1].Low)
	}
	out.GeoBlock, _ = firstGeometryBlock(unique, offsets)
	out.Targets = decodeTable(data, unique, offsets, bestTarget)
	out.Texts, out.Guards = map[int]string{}, map[int]string{}
	out.GuardCells = map[int][]GuardCompare{}
	for index, target := range out.Targets {
		out.Indexes = append(out.Indexes, index)
		out.Texts[index] = firstTextAt(unique, offsets, target)
		out.Guards[index] = guardAt(unique, offsets, target)
		out.GuardCells[index] = guardCellsAt(unique, offsets, target)
	}
	sort.Ints(out.Indexes)

	// 查表本身在 block 自己的資料裡（`位址 − 0x8000`），一個索引一個位元組。
	base := int(get.Operands[0].Word) - ecl.CodeAddressBase
	for index := 0; index < tableProbeLength; index++ {
		// ⚠ payload 與走訪位移差兩個位元組（區塊標頭）。
		cell := base + index + 2
		if cell < 0 || cell >= len(data) {
			break
		}
		out.TableValues = append(out.TableValues, int(data[cell]))
	}
	return out
}

// findOnGoto 找分派 `cell` 的第一個 `ON GOTO`。
//
// ⚠ `ON GOTO` 是變長指令，`Command.Arity` 是 0，所以 `Instruction.Operands`
// **是空的**——被分派的那一格只能從 payload 的原始位元組讀。
func findOnGoto(data []byte, unique map[int]ecl.Instruction, offsets []int,
	cell uint16) (int, bool) {
	for _, offset := range offsets {
		if unique[offset].Command.Opcode != 0x25 {
			continue
		}
		if value, ok := onGotoCell(data, offset); ok && value == cell {
			return offset, true
		}
	}
	return 0, false
}

// onGotoCell 讀出 `ON GOTO` 分派的是哪一格。
func onGotoCell(data []byte, offset int) (uint16, bool) {
	// ⚠ 走訪用的位移與 payload 差兩個位元組（區塊標頭）。
	base := offset + 2
	if base < 0 || base+3 >= len(data) {
		return 0, false
	}
	if data[base+1] == 0x00 {
		// 立即數當分派來源沒有意義，當成讀不到。
		return 0, false
	}
	return uint16(data[base+2]) | uint16(data[base+3])<<8, true
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

// guardCellsAt 把守衛裡的 `COMPARE` 拆成「哪一格、比什麼值」。
//
// ⚠ 只收**資料格**：位址落在程式碼視窗（`8000h` 以上）的是跳躍目標不是格子，
// 立即數那一側也只收得到單一位元組。`COMPARE AND` 一條裡有兩組，兩組都收。
func guardCellsAt(unique map[int]ecl.Instruction, offsets []int, start int) []GuardCompare {
	index := sort.SearchInts(offsets, start)
	cells := make([]GuardCompare, 0, 4)
	for cursor := index; cursor < len(offsets) && cursor < index+maxGuardInstructions; cursor++ {
		instruction := unique[offsets[cursor]]
		if hasText(instruction) {
			break
		}
		if !strings.HasPrefix(instruction.Command.Name, "COMPARE") {
			continue
		}
		// 後面那條 `IF` 決定往哪一邊避，再後面那條是不是 `EXIT` 決定值不值得避。
		operator, exits := "", false
		if cursor+1 < len(offsets) {
			next := unique[offsets[cursor+1]].Command.Name
			if strings.HasPrefix(next, "IF ") {
				operator = strings.TrimPrefix(next, "IF ")
				if cursor+2 < len(offsets) {
					exits = unique[offsets[cursor+2]].Command.Name == "EXIT"
				}
			}
		}
		operands := instruction.Operands
		for position := 0; position+1 < len(operands); position += 2 {
			address := operands[position]
			value := operands[position+1]
			if !address.WordSet || address.Word >= ecl.CodeAddressBase {
				continue
			}
			compare := GuardCompare{Address: address.Word, Operator: operator, ExitsOnMatch: exits}
			if value.WordSet {
				compare.Value = value.Word
			} else {
				compare.Value = uint16(value.Low)
			}
			cells = append(cells, compare)
		}
	}
	return cells
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
