package ecl

import "sort"

// 載入存檔那一刻，一段 ECL 會跑到哪裡。
//
// ★ 為什麼需要這一支。 從 corpus 掃一張「每一段會做什麼」的表時，**「掃得到這條
// 指令」不等於「這條指令會跑」**。CoAB 的段開頭幾乎都有一道
// `COMPARE 4BF2h, <自己的段號>`（`4BF2h` ＝ 存檔的 `LastECLBlockID`），
// 而**兩個方向都有人用**：
//
//	ECL4/0x21  IF =  → GOTO 8112h   本來就在這一段 ⇒ 跳過整段前置
//	ECL3/0x11  IF =  → EXIT         本來就在這一段 ⇒ 直接結束
//	ECL5/0x33  IF <> → GOTO 8045h   不在這一段才跳走 ⇒ 前置照跑
//	ECL3/0x15  （沒有閘門）          一定跑
//
// remake 套這些表的時機正是「載入一份記著這一段的存檔」⇒ 那一刻
// `4BF2h == 段號`。所以要在**這個條件成立的前提下**走一次，才知道那條指令會不會
// 被執行。
//
// ⚠ 只掃指令存在與否會多收：天空選色那張表第一版就是這樣多收八段，把四張圖的
// 天空改錯 32 萬格，而症狀是整片純色換成另一片純色，看不出異常（spec 1185）。
const (
	lastECLBlockCell = 0x4BF2

	exitOpcode    = 0x00
	gotoOpcode    = 0x01
	gosubOpcode   = 0x02
	compareOpcode = 0x03
	returnOpcode  = 0x13

	// cellOperandCode 是「這個運算元是一個記憶體格子」。`0x00` 是立即位元組、
	// `0x02` 是立即字組，兩者都是常數（`ConstantOperandValue`）。
	cellOperandCode = 0x01

	firstIfOpcode = 0x16
	lastIfOpcode  = 0x1B

	loadTimeStepLimit = 4096
	loadTimeEntries   = 5
)

// ReachableOnLoad 回傳「載入一份 `LastECLBlockID == block` 的存檔時，這一段會
// 執行到」的指令，依位移排序。
//
// ⚠ 這**不是**完整的可達性分析：它只走「一定會跑到的前置」那條路。
//
//   - `IF` 只守著下一條指令。判得出來就照判；判不出來而守的是一般指令，就跳過
//     那一條（與原作「條件不成立就跳過」一致，而跳過是保守的選擇）。
//   - 判不出來而守的是控制流（`GOTO`／`GOSUB`／`EXIT`）⇒ **停下來**，不要猜。
//   - 走回頭路也停：這一支不解迴圈。
//
// 五個進入點都會走：段的前置不一定掛在第 0 個（`ECL3/0x15` 的 `LOAD FILES` 在
// `0x14`，而 `ECL5/0x33` 的在 `0x51`）。
// LoadTimeMemory 讓呼叫端把**載檔當下的記憶體**餵進來，於是 `IF` 不只判得出
// 「是不是這一段」，也判得出腳本拿別的格子做的比較。
//
// ★ 為什麼需要。 有些段的進入碼會依**當下站在哪一格**決定要寫什麼：
// `ECL2/0x04` 的 `0x0022 GOSUB 0x96F4` 進去之後是
// `COMPARE C04F 0x95 / IF = / SAVE 0A 4BFE / IF <> / SAVE 09 4BFE`——
// 天花板顏色由那一格的地形碼決定（`C04Fh` ＝ 牆頂）。靜態的表表達不了
// 「看情況」，所以那一格的紅色天花板在 remake 是白的（差 13,748 格，spec 1185）。
type LoadTimeMemory func(address uint16) (uint16, bool)

// ReachableOnLoad 是不帶記憶體的版本：只判得出 `4BF2h` 的那一道閘門。
func ReachableOnLoad(data []byte, block uint8) ([]Instruction, error) {
	return ReachableOnLoadWithMemory(data, block, nil)
}

// ReachableOnLoadWithMemory 同上，但 `IF` 會拿 `memory` 去判。
func ReachableOnLoadWithMemory(data []byte, block uint8, memory LoadTimeMemory) ([]Instruction, error) {
	points, _, err := EntryPoints(data, loadTimeEntries)
	if err != nil {
		return nil, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-CodeAddressBase)
	}
	graph, err := TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil, err
	}
	byOffset := map[int]Instruction{}
	for _, instruction := range graph.Instructions {
		byOffset[instruction.Offset] = instruction
	}
	offsets := make([]int, 0, len(byOffset))
	for offset := range byOffset {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)
	position := map[int]int{}
	for index, offset := range offsets {
		position[offset] = index
	}

	reached := map[int]Instruction{}
	for _, point := range points {
		start, ok := position[int(point)-CodeAddressBase]
		if !ok {
			continue
		}
		walkOnLoad(offsets, position, byOffset, start,
			loadTimeContext{block: block, memory: memory}, reached)
	}
	result := make([]Instruction, 0, len(reached))
	for _, offset := range offsets {
		if instruction, ok := reached[offset]; ok {
			result = append(result, instruction)
		}
	}
	return result, nil
}

// loadTimeContext 是判斷 `IF` 時知道的東西。
type loadTimeContext struct {
	block  uint8
	memory LoadTimeMemory
}

// resolve 把一個運算元換成值。常數直接用；`4BF2h` 用段號（載檔當下就是它）；
// 其餘格子交給呼叫端餵進來的記憶體。
func (c loadTimeContext) resolve(operand Operand) (uint16, bool) {
	if value, ok := ConstantOperandValue(operand); ok {
		return value, true
	}
	if operand.Code != cellOperandCode || !operand.WordSet {
		return 0, false
	}
	if operand.Word == lastECLBlockCell {
		return uint16(c.block), true
	}
	if c.memory == nil {
		return 0, false
	}
	return c.memory(operand.Word)
}

// loadTimeComparison 是一次 `COMPARE` 的結果。
type loadTimeComparison struct {
	known       bool
	left, right uint16
}

func walkOnLoad(offsets []int, position map[int]int, byOffset map[int]Instruction,
	index int, context loadTimeContext, reached map[int]Instruction) {
	var comparison loadTimeComparison
	// ⚠ `GOSUB` 要跟進去：`ECL2/0x04` 的進入碼在 `LOAD FILES`／`LOAD PIECES`
	// 之後就 `GOSUB` 一支設天花板顏色的子程式。不跟的話那一段的寫入整個看不到。
	stack := make([]int, 0, 8)
	visited := map[int]bool{}
	for step := 0; step < loadTimeStepLimit && index >= 0 && index < len(offsets); step++ {
		if visited[index] {
			return
		}
		visited[index] = true
		instruction := byOffset[offsets[index]]
		reached[instruction.Offset] = instruction

		switch instruction.Command.Opcode {
		case exitOpcode:
			return
		case gotoOpcode:
			target, ok := loadTimeJumpTarget(instruction, position)
			if !ok {
				return
			}
			index = target
			comparison = loadTimeComparison{}
			continue
		case gosubOpcode:
			target, ok := loadTimeJumpTarget(instruction, position)
			if !ok {
				return
			}
			stack = append(stack, index+1)
			index = target
			comparison = loadTimeComparison{}
			continue
		case returnOpcode:
			if len(stack) == 0 {
				return
			}
			index = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			comparison = loadTimeComparison{}
			continue
		case compareOpcode:
			comparison = compareOnLoad(instruction, context)
		}

		if instruction.Command.Opcode >= firstIfOpcode && instruction.Command.Opcode <= lastIfOpcode {
			if index+1 >= len(offsets) {
				return
			}
			guarded := byOffset[offsets[index+1]]
			// ⚠ **一個 `COMPARE` 後面可以跟好幾個 `IF`**，那是原作的慣用法：
			// `ECL2/0x04` 的 `COMPARE C04F 0x95` 之後接 `IF =`（紅）與
			// `IF <>`（白）。判完就把比較結果清掉的話，第二個 `IF` 會變成
			// 「判不出來」而被跳過——於是那一段永遠只收得到一半。
			take, decided := evaluateLoadTimeIf(instruction.Command.Opcode, comparison)
			if !decided {
				if isLoadTimeControlFlow(guarded.Command.Opcode) {
					return
				}
				index += 2
				continue
			}
			if !take {
				index += 2
				continue
			}
			index++
			continue
		}
		index++
	}
}

func loadTimeJumpTarget(instruction Instruction, position map[int]int) (int, bool) {
	if len(instruction.Operands) == 0 || !instruction.Operands[0].WordSet {
		return 0, false
	}
	index, ok := position[int(instruction.Operands[0].Word)-CodeAddressBase]
	return index, ok
}

// compareOnLoad 把一次 `COMPARE` 的兩邊都換成值；換不出來就標成不知道。
func compareOnLoad(instruction Instruction, context loadTimeContext) loadTimeComparison {
	if len(instruction.Operands) < 2 {
		return loadTimeComparison{}
	}
	left, leftOK := context.resolve(instruction.Operands[0])
	right, rightOK := context.resolve(instruction.Operands[1])
	if !leftOK || !rightOK {
		return loadTimeComparison{}
	}
	return loadTimeComparison{known: true, left: left, right: right}
}

// evaluateLoadTimeIf 判斷分支。兩邊的值都知道時六個比較都判得出來；
// 有一邊不知道就不猜。
func evaluateLoadTimeIf(opcode uint8, comparison loadTimeComparison) (take, decided bool) {
	if !comparison.known {
		return false, false
	}
	left, right := comparison.left, comparison.right
	switch opcode {
	case 0x16: // IF =
		return left == right, true
	case 0x17: // IF <>
		return left != right, true
	case 0x18: // IF <
		return left < right, true
	case 0x19: // IF >
		return left > right, true
	case 0x1A: // IF <=
		return left <= right, true
	case 0x1B: // IF >=
		return left >= right, true
	}
	return false, false
}

func isLoadTimeControlFlow(opcode uint8) bool {
	return opcode == exitOpcode || opcode == gotoOpcode || opcode == gosubOpcode
}

// ConstantOperandValue 只認靜態就知道值的兩種運算元。
func ConstantOperandValue(operand Operand) (uint16, bool) {
	switch operand.Code {
	case 0x00:
		return uint16(operand.Low), true
	case 0x02:
		if !operand.WordSet {
			return 0, false
		}
		return operand.Word, true
	default:
		return 0, false
	}
}
