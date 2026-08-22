package ecl

// BlockWallPieces 回報一段 ECL 進入時發出的牆磚選圖三元組（`37h LOAD PIECES`）。
//
// ★ 為什麼要有一支共用的。 這個值有兩個用途：remake 載入原版存檔時要補回
// `LoadPieces`（否則每一面牆都畫成空氣），而自製存檔換圖時要一併換掉存檔裡的
// 牆面參數（否則原版會拿舊圖的選圖去新的 `WALLDEF` 檔裡找，印
// `Unable to load 1 from WALLDEF4.`）。兩邊要用同一份讀法，否則會漂移。
//
// ⚠ 只認**常數**運算元（`00h` 位元組、`02h` 字組）。`01h`／`03h` 是記憶體參照，
// 值要看執行時的記憶體，靜態看不出來——那種情況回 false，讓呼叫端自己決定，
// 不要猜一個值出來。
//
// 全 corpus 掃過的結果是 19 段各只發一組常數三元組（`cmd/ecl-wall-pieces`）。
func BlockWallPieces(data []byte) ([3]uint16, bool) {
	points, _, err := EntryPoints(data, 5)
	if err != nil {
		return [3]uint16{}, false
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-CodeAddressBase)
	}
	graph, err := TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return [3]uint16{}, false
	}
	seen := map[int]bool{}
	for _, instruction := range graph.Instructions {
		if instruction.Command.Opcode != 0x37 || seen[instruction.Offset] {
			continue
		}
		seen[instruction.Offset] = true
		var triple [3]uint16
		ok := true
		for index, operand := range instruction.Operands {
			if index > 2 {
				break
			}
			value, constant := constantOperandValue(operand)
			if !constant {
				ok = false
				break
			}
			triple[index] = value
		}
		if ok {
			return triple, true
		}
	}
	return [3]uint16{}, false
}

// constantOperandValue 只認靜態就知道值的兩種運算元，讀法與 operandValue 相同。
func constantOperandValue(operand Operand) (uint16, bool) {
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
