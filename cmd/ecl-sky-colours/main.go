// Command ecl-sky-colours 掃出每一段 ECL 進入時設定的天空／地面選色，產生
// `internal/game/skycolors_table.go`。
//
// ★ 存在的理由。 載入原版存檔時 remake 不重跑該段的進入碼，於是 Area1 的
// `OutdoorSkyColor`／`IndoorSkyColor` 停在底檔（別章）的值。原版在同一份存檔上
// 會跑那段進入碼，把兩格改成這一段自己的選色 ⇒ 兩邊的天空顏色不同，而**畫面
// 看起來完全正常**：天空是一整片純色，錯的顏色也是一整片純色。
//
// 第 749 輪量到的差距：`geo5-b35` 二十張畫面每一張都差 2,000～3,400 格，全部是
// 同一個代換（EGA 4 → 15），佔那四張圖總差異的 66%。
//
// 兩格就是 ECL 記憶體的 `4BFDh`／`4BFEh`：Area1 的字組索引 `位址 − 4B00h`，
// 位元組位移 `× 2` ⇒ `01FAh`／`01FCh`，正是 `internal/area` 的
// `area1OutdoorSky`／`area1IndoorSky`（spec 1163／1185）。
//
// ⚠ 值是**選色索引**不是 EGA 色號：引擎的 `skyPalette` 查表之後才是 EGA
// （`0Ah` → `skyPalette[2]` ＝ EGA 4）。
//
// ⚠ 只收**沒有條件守著**的寫入。`ECL5/0x33` 的 `4BFDh` 先寫 `0Bh`，再由
// `IF >` 守著一條 `08h`——那一條要看執行時的 `4BC9h`，靜態決定不了。有條件的
// 寫入會列進報表，不會進表也不會被無聲丟掉。
//
// 用法：
//
//	go run ./cmd/ecl-sky-colours -output internal/game/skycolours_table.go \
//	  -report docs/audit/ecl-sky-colours.md
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

const (
	outdoorSkyCell = 0x4BFD
	indoorSkyCell  = 0x4BFE
	// lastECLCell 是存檔記下來的「上一段 ECL」（Area1 `01E4h`）。段的開頭常拿它
	// 跟自己的段號比，用來分辨「第一次進來」與「本來就在這一段」。
	lastECLCell   = 0x4BF2
	exitOpcode    = 0x00
	gotoOpcode    = 0x01
	gosubOpcode   = 0x02
	compareOpcode = 0x03
	saveOpcode    = 0x09
	firstIfOpcode = 0x16
	lastIfOpcode  = 0x1B
	walkStepLimit = 4096
)

type write struct {
	area, block uint8
	offset      int
	cell        uint16
	value       uint16
	guarded     bool
	guard       string
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "原版 image ZIP")
	output := flag.String("output", "", "產生的 Go 表路徑（留白就不寫）")
	report := flag.String("report", "", "Markdown 報表路徑（留白就印到 stdout）")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	writes := make([]write, 0, 64)
	for area := 1; area <= 6; area++ {
		payload := member(archive, fmt.Sprintf("ECL%d.DAX", area))
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatalf("ECL%d.DAX: %v", area, parseErr)
		}
		for _, block := range blocks {
			found, scanErr := scanBlock(uint8(area), block.Entry.ID, block.Data)
			if scanErr != nil {
				log.Fatalf("ECL%d/0x%02X: %v", area, block.Entry.ID, scanErr)
			}
			writes = append(writes, found...)
		}
	}
	sort.SliceStable(writes, func(left, right int) bool {
		if writes[left].area != writes[right].area {
			return writes[left].area < writes[right].area
		}
		if writes[left].block != writes[right].block {
			return writes[left].block < writes[right].block
		}
		return writes[left].offset < writes[right].offset
	})

	table := resolve(writes)
	if *output != "" {
		if err := os.WriteFile(*output, []byte(renderTable(table)), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	text := renderReport(writes, table)
	if *report == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*report, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	guarded := 0
	for _, item := range writes {
		if item.guarded {
			guarded++
		}
	}
	fmt.Fprintf(os.Stderr, "writes=%d guarded=%d blocks=%d\n", len(writes), guarded, len(table))
}

// key 是 ECL 段，不是地圖：GEO5 的 `0x31` 與 `0x32` 共用同一張幾何區塊，選色
// 卻不同（`0Bh`／`0Ah` vs 只有 `0Ah`）。照地圖查會在這兩段給出自洽但錯的天空。
type key struct{ area, block uint8 }

type colours struct {
	outdoor, indoor       int
	outdoorAt, indoorAt   int
	outdoorAlt, indoorAlt []int
}

func resolve(writes []write) map[key]*colours {
	table := map[key]*colours{}
	for _, item := range writes {
		identity := key{area: item.area, block: item.block}
		entry := table[identity]
		if entry == nil {
			entry = &colours{outdoor: -1, indoor: -1}
			table[identity] = entry
		}
		switch item.cell {
		case outdoorSkyCell:
			if item.guarded {
				entry.outdoorAlt = append(entry.outdoorAlt, int(item.value))
				continue
			}
			entry.outdoor, entry.outdoorAt = int(item.value), item.offset
		case indoorSkyCell:
			if item.guarded {
				entry.indoorAlt = append(entry.indoorAlt, int(item.value))
				continue
			}
			entry.indoor, entry.indoorAt = int(item.value), item.offset
		}
	}
	for identity, entry := range table {
		if entry.outdoor < 0 && entry.indoor < 0 {
			delete(table, identity)
		}
	}
	return table
}

// scanBlock 走一次「載入存檔之後，這一段實際會跑到的路」，收沿路寫進兩格天空
// 選色的常數。
//
// ★ 為什麼不能只掃「有沒有這條指令」。 段的開頭幾乎都有一道
// `COMPARE 4BF2h, <自己的段號>` 的閘門，而**兩個方向都有人用**：
//
//	ECL4/0x21  IF =  → GOTO 8112h   { 本來就在這一段 ⇒ 跳過整段前置 }
//	ECL3/0x11  IF =  → EXIT         { 本來就在這一段 ⇒ 直接結束 }
//	ECL5/0x33  IF <> → GOTO 8045h   { 不在這一段才跳走 ⇒ 前置照跑 }
//	ECL3/0x15  （沒有閘門）          { 一定跑 }
//
// `4BF2h` 就是存檔的 `LastECLBlockID`，而 remake 套這張表的時機正是「載入一份
// 記著這一段的存檔」⇒ 那一刻 `4BF2h == 段號`。所以要在**這個條件成立**的前提下
// 走一次，才知道那條 `SAVE` 到底會不會被執行。
//
// ⚠ 只掃指令存在與否會多收：`ECL4/0x21`／`ECL4/0x25`／`ECL3/0x10`／`ECL3/0x11`／
// `ECL2/0x03` 五段的天空寫入全都在閘門後面，載檔時**一條都不會跑**。第 749 輪
// 就是這樣把四張圖的天空改錯——而症狀是整片純色換成另一片純色，看不出異常
// （spec 1185）。
//
// ⚠ 判斷不了的分支一律**停下來**，不要猜。停下來的段會在報表裡標出來。
func scanBlock(area, block uint8, data []byte) ([]write, error) {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return nil, nil
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
	index := map[int]int{}
	for position, offset := range offsets {
		index[offset] = position
	}

	// ⚠ **五個進入點都要走。** 段的前置不一定掛在第 0 個：`ECL3/0x15` 的
	// `LOAD FILES` 在 `0x14`，而 `ECL5/0x33` 的 `0x14` 是另一段判斷，
	// `LOAD FILES` 在 `0x51`。只走第 0 個會漏掉一半的段，而漏掉的後果是那一段
	// 安靜地沿用別章的天空。
	found := make([]write, 0, 4)
	seen := map[int]bool{}
	for _, point := range points {
		start, ok := index[int(point)-ecl.CodeAddressBase]
		if !ok {
			continue
		}
		for _, item := range walkFrom(offsets, index, unique, start, area, block) {
			if seen[item.offset] {
				continue
			}
			seen[item.offset] = true
			found = append(found, item)
		}
	}
	return found, nil
}

// walkFrom 從一個進入點走一條路，回沿路收到的天空選色寫入。
func walkFrom(offsets []int, index map[int]int, unique map[int]ecl.Instruction,
	position int, area, block uint8) []write {
	found := make([]write, 0, 4)
	// compared 記「上一條 COMPARE 是不是拿 4BF2h 跟一個常數比」，以及比的結果。
	compared, equal := false, false
	visited := map[int]bool{}
	for step := 0; step < walkStepLimit && position < len(offsets); step++ {
		if visited[position] {
			// 走回頭路就停：這一支只要「載檔那一刻一定會跑到的前置」，
			// 不是完整的可達性分析。
			return found
		}
		visited[position] = true
		instruction := unique[offsets[position]]
		switch instruction.Command.Opcode {
		case exitOpcode:
			return found
		case gotoOpcode:
			next, jumped := jumpTarget(instruction, index)
			if !jumped {
				return found
			}
			position = next
			compared = false
			continue
		case compareOpcode:
			compared, equal = compareAgainstLastECL(instruction, block)
		case saveOpcode:
			if item, okWrite := skyWrite(instruction, area, block, offsets[position]); okWrite {
				found = append(found, item)
			}
		}
		if instruction.Command.Opcode >= firstIfOpcode && instruction.Command.Opcode <= lastIfOpcode {
			// `IF` 只守著**下一條**指令。
			if position+1 >= len(offsets) {
				return found
			}
			guarded := unique[offsets[position+1]]
			take, decided := evaluateIf(instruction.Command.Opcode, compared, equal)
			compared = false
			if !decided {
				// 判斷不了。守的是控制流就停（走哪一條都不確定）；守的是一般
				// 指令就跳過它，其餘照走——這與原作「條件不成立就跳過那一條」
				// 的語意一致，而跳過一條 `SAVE` 正是保守的選擇。
				if isControlFlow(guarded.Command.Opcode) {
					return found
				}
				position += 2
				continue
			}
			if !take {
				position += 2
				continue
			}
			position++
			continue
		}
		position++
	}
	return found
}

// jumpTarget 把 `GOTO` 的目的地換成指令表裡的位置。
func jumpTarget(instruction ecl.Instruction, index map[int]int) (int, bool) {
	if len(instruction.Operands) == 0 || !instruction.Operands[0].WordSet {
		return 0, false
	}
	position, ok := index[int(instruction.Operands[0].Word)-ecl.CodeAddressBase]
	return position, ok
}

// compareAgainstLastECL 認得「拿 `4BF2h` 跟一個常數比」這一種 COMPARE。
func compareAgainstLastECL(instruction ecl.Instruction, block uint8) (compared, equal bool) {
	if len(instruction.Operands) < 2 {
		return false, false
	}
	if !instruction.Operands[0].WordSet || instruction.Operands[0].Word != lastECLCell {
		return false, false
	}
	value, constant := constantOperand(instruction.Operands[1])
	if !constant {
		return false, false
	}
	return true, value == uint16(block)
}

// evaluateIf 在「只知道相不相等」的前提下判斷分支。
//
// ⚠ 相等時六個比較都判得出來；**不相等時只判得出 `=` 與 `<>`**，大小關係要看
// 實際數值，這裡不猜。
func evaluateIf(opcode uint8, compared, equal bool) (take, decided bool) {
	if !compared {
		return false, false
	}
	switch opcode {
	case 0x16: // IF =
		return equal, true
	case 0x17: // IF <>
		return !equal, true
	case 0x18: // IF <
		if equal {
			return false, true
		}
	case 0x19: // IF >
		if equal {
			return false, true
		}
	case 0x1A: // IF <=
		if equal {
			return true, true
		}
	case 0x1B: // IF >=
		if equal {
			return true, true
		}
	}
	return false, false
}

func isControlFlow(opcode uint8) bool {
	return opcode == exitOpcode || opcode == gotoOpcode || opcode == gosubOpcode
}

// skyWrite 認得「把常數寫進兩格天空選色」的 SAVE。
func skyWrite(instruction ecl.Instruction, area, block uint8, offset int) (write, bool) {
	if len(instruction.Operands) < 2 {
		return write{}, false
	}
	destination := instruction.Operands[1]
	if !destination.WordSet {
		return write{}, false
	}
	if destination.Word != outdoorSkyCell && destination.Word != indoorSkyCell {
		return write{}, false
	}
	value, constant := constantOperand(instruction.Operands[0])
	if !constant {
		return write{}, false
	}
	return write{area: area, block: block, offset: offset, cell: destination.Word, value: value}, true
}

func constantOperand(operand ecl.Operand) (uint16, bool) {
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

func renderTable(table map[key]*colours) string {
	identities := make([]key, 0, len(table))
	for identity := range table {
		identities = append(identities, identity)
	}
	sort.Slice(identities, func(left, right int) bool {
		if identities[left].area != identities[right].area {
			return identities[left].area < identities[right].area
		}
		return identities[left].block < identities[right].block
	})

	var out strings.Builder
	out.WriteString(`// Code generated by cmd/ecl-sky-colours; DO NOT EDIT.

package game

// eclBlockSkyColours 是每一段 ECL 進入時寫進 ` + "`4BFDh`／`4BFEh`" + ` 的天空選色，
// 逐段從原版 ECL 位元組碼掃出來（` + "`cmd/ecl-sky-colours`" + `，spec 1185）。
//
// ★ 為什麼要有這張表。 那兩格就是存檔 Area1 的 ` + "`OutdoorSkyColor`／`IndoorSkyColor`" + `
// （位元組位移 ` + "`01FAh`／`01FCh`" + `）。載入原版存檔時 remake 不重跑該段的進入碼，
// 兩格會停在底檔（別章）的值；原版在同一份存檔上會跑，於是天空整片是另一個顏色，
// 而畫面看起來完全正常——純色錯成另一個純色，沒有任何破圖可以提示。
//
// ⚠ 鍵是 **ECL 段**，不是地圖。GEO5 的 ` + "`0x31`／`0x32`" + ` 共用同一張幾何區塊，
// 選色卻不同。
//
// ⚠ 值是**選色索引**，引擎再查 ` + "`skyPalette`" + ` 才得到 EGA 色號。
// ` + "`-1`" + ` ＝ 這一段沒有寫那一格，維持存檔帶進來的值。
var eclBlockSkyColours = map[eclBlockKey]skyColours{
`)
	for _, identity := range identities {
		entry := table[identity]
		out.WriteString(fmt.Sprintf("\t{area: %d, block: 0x%02X}: {Outdoor: %d, Indoor: %d},\n",
			identity.area, identity.block, entry.outdoor, entry.indoor))
	}
	out.WriteString(`}

// skyColours 是一段 ECL 的兩個選色索引；-1 表示那一段沒有設定。
type skyColours struct {
	Outdoor int
	Indoor  int
}
`)
	return out.String()
}

func renderReport(writes []write, table map[key]*colours) string {
	var out strings.Builder
	out.WriteString("# 每一段 ECL 進入時設定的天空選色\n\n")
	out.WriteString("由 `cmd/ecl-sky-colours` 產生，不要手改。機制與證據見 spec 1185。\n\n")
	out.WriteString("`4BFDh` ＝ 室外、`4BFEh` ＝ 室內；地圖的屋頂位元組 `>= 80h` 時用室內那一個。\n")
	out.WriteString("選色索引經引擎的 `skyPalette` 查表才是 EGA 色號。\n\n")

	out.WriteString("## 逐處寫入\n\n")
	out.WriteString("| 段 | 位移 | 格 | 值 | 守衛 |\n|---|---:|---|---:|---|\n")
	for _, item := range writes {
		guard := "—"
		if item.guarded {
			guard = "`" + item.guard + "`"
		}
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | `0x%04x` | `%04Xh` | `%02Xh` | %s |\n",
			item.area, item.block, item.offset, item.cell, item.value, guard))
	}

	out.WriteString("\n## 進表的值\n\n")
	out.WriteString("有條件的寫入不進表：它要看執行時的記憶體，靜態決定不了。\n\n")
	out.WriteString("| 段 | 室外 | 室內 | 被跳過的有條件寫入 |\n|---|---:|---:|---|\n")
	identities := make([]key, 0, len(table))
	for identity := range table {
		identities = append(identities, identity)
	}
	sort.Slice(identities, func(left, right int) bool {
		if identities[left].area != identities[right].area {
			return identities[left].area < identities[right].area
		}
		return identities[left].block < identities[right].block
	})
	for _, identity := range identities {
		entry := table[identity]
		skipped := make([]string, 0, 2)
		for _, value := range entry.outdoorAlt {
			skipped = append(skipped, fmt.Sprintf("室外 `%02Xh`", value))
		}
		for _, value := range entry.indoorAlt {
			skipped = append(skipped, fmt.Sprintf("室內 `%02Xh`", value))
		}
		note := "—"
		if len(skipped) > 0 {
			note = strings.Join(skipped, "、")
		}
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | %s | %s | %s |\n",
			identity.area, identity.block, cell(entry.outdoor), cell(entry.indoor), note))
	}
	return out.String()
}

func cell(value int) string {
	if value < 0 {
		return "—"
	}
	return fmt.Sprintf("`%02Xh`", value)
}

func member(archive *zip.ReadCloser, name string) []byte {
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
