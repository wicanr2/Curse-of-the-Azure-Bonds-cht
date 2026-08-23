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
func ReachableOnLoad(data []byte, block uint8) ([]Instruction, error) {
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
		walkOnLoad(offsets, position, byOffset, start, block, reached)
	}
	result := make([]Instruction, 0, len(reached))
	for _, offset := range offsets {
		if instruction, ok := reached[offset]; ok {
			result = append(result, instruction)
		}
	}
	return result, nil
}

func walkOnLoad(offsets []int, position map[int]int, byOffset map[int]Instruction,
	index int, block uint8, reached map[int]Instruction) {
	compared, equal := false, false
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
			compared = false
			continue
		case compareOpcode:
			compared, equal = comparesLastECLBlock(instruction, block)
		}

		if instruction.Command.Opcode >= firstIfOpcode && instruction.Command.Opcode <= lastIfOpcode {
			if index+1 >= len(offsets) {
				return
			}
			guarded := byOffset[offsets[index+1]]
			take, decided := evaluateLoadTimeIf(instruction.Command.Opcode, compared, equal)
			compared = false
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

// comparesLastECLBlock 認得「拿 `4BF2h` 跟一個常數比」這一種 COMPARE。
func comparesLastECLBlock(instruction Instruction, block uint8) (compared, equal bool) {
	if len(instruction.Operands) < 2 {
		return false, false
	}
	if !instruction.Operands[0].WordSet || instruction.Operands[0].Word != lastECLBlockCell {
		return false, false
	}
	value, ok := ConstantOperandValue(instruction.Operands[1])
	if !ok {
		return false, false
	}
	return true, value == uint16(block)
}

// evaluateLoadTimeIf 在「只知道相不相等」的前提下判斷分支。
//
// ⚠ 相等時六個比較都判得出來；**不相等時只判得出 `=` 與 `<>`**，大小關係要看
// 實際數值，這裡不猜。
func evaluateLoadTimeIf(opcode uint8, compared, equal bool) (take, decided bool) {
	if !compared {
		return false, false
	}
	switch opcode {
	case 0x16: // IF =
		return equal, true
	case 0x17: // IF <>
		return !equal, true
	case 0x18, 0x19: // IF < / IF >
		if equal {
			return false, true
		}
	case 0x1A, 0x1B: // IF <= / IF >=
		if equal {
			return true, true
		}
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
