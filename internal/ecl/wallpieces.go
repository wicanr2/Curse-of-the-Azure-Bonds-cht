package ecl

// 一段 ECL 的 `37h LOAD PIECES` 有**兩個不同的問題**，兩個都有人問，答案不一樣：
//
//	BlockWallPieces        這一段**宣告**的三元組是什麼？
//	BlockWallPiecesOnLoad  載入一份記著這一段的存檔時，這一條**會不會被執行**？
//
// ⚠ 兩者不能互相替代。段的開頭幾乎都有一道 `COMPARE 4BF2h, <段號>` 的閘門，
// 載檔時 `4BF2h == 段號` ⇒ 有些段的 `LOAD PIECES` 一次都不會發（spec 1185）。
//
//   - **自製存檔換圖**要的是「宣告」：`cmd/dos-save-export` 是在**組一份存檔**，
//     宣告「這張圖的牆是這三塊」，好讓原版載進去之後拿正確的選圖去
//     `WALLDEF<章>.DAX` 找。跟載檔時會不會重發無關——不宣告的話原版會拿**上一張
//     圖的**選圖去新章的檔案裡找，印 `Unable to load 1 from WALLDEF4.`。
//   - **remake 補回選圖**要的是「會不會被執行」：會，就以段的宣告為準（原版那一條
//     一跑就把三個槽重設）；不會，就以存檔記的 `OLDWALL` 為準。
//
// ⚠ 兩支都只認**常數**運算元（`00h` 位元組、`02h` 字組）。`01h`／`03h` 是記憶體
// 參照，值要看執行時的記憶體，靜態看不出來——那種情況回 false，讓呼叫端自己
// 決定，不要猜一個值出來。

// BlockWallPieces 回報一段 ECL **宣告**的牆磚選圖三元組。
//
// 走的是整張 trace graph（五個生命週期入口跟著跳躍走），不看載檔時的閘門。
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
		if seen[instruction.Offset] {
			continue
		}
		seen[instruction.Offset] = true
		if triple, ok := loadPiecesTriple(instruction); ok {
			return triple, true
		}
	}
	return [3]uint16{}, false
}

// BlockWallPiecesOnLoad 回報一段 ECL **載入存檔時真的會發**的牆磚選圖三元組。
// 發不到就回 false ⇒ 那一段的牆面要以存檔記的 `OLDWALL` 為準。
func BlockWallPiecesOnLoad(data []byte, block uint8) ([3]uint16, bool) {
	instructions, err := ReachableOnLoad(data, block)
	if err != nil {
		return [3]uint16{}, false
	}
	for _, instruction := range instructions {
		if triple, ok := loadPiecesTriple(instruction); ok {
			return triple, true
		}
	}
	return [3]uint16{}, false
}

func loadPiecesTriple(instruction Instruction) ([3]uint16, bool) {
	if instruction.Command.Opcode != 0x37 {
		return [3]uint16{}, false
	}
	var triple [3]uint16
	for index, operand := range instruction.Operands {
		if index > 2 {
			break
		}
		value, constant := ConstantOperandValue(operand)
		if !constant {
			return [3]uint16{}, false
		}
		triple[index] = value
	}
	return triple, true
}
