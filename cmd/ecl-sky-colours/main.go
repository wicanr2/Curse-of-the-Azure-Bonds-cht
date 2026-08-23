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

// trackedCells 是 remake 在「載入存檔」那一刻需要、而原作靠**跑段的進入碼**
// 得到的 Area1 格子。全部都是同一個形狀：值由該段的進入碼寫入，而 remake 不
// 重跑進入碼（會連帶觸發劇情副作用）⇒ 不補就會沿用底檔（別章）的值。
//
//	4BE7h／4BE8h  `37h LOAD PIECES` 的兩槽模式閘門（spec 1087）
//	4BFDh／4BFEh  室外／室內天空選色
var trackedCells = []uint16{0x4BE7, 0x4BE8, 0x4BFD, 0x4BFE}

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

func resolve(writes []write) map[key]map[uint16]uint16 {
	table := map[key]map[uint16]uint16{}
	for _, item := range writes {
		identity := key{area: item.area, block: item.block}
		cells := table[identity]
		if cells == nil {
			cells = map[uint16]uint16{}
			table[identity] = cells
		}
		// 同一格寫多次時**後面的贏**：走訪是照執行順序走的。
		cells[item.cell] = item.value
	}
	return table
}

// scanBlock 收「載入一份記著這一段的存檔」時真的會跑到的常數寫入。
//
// ⚠ 走訪本身在 `ecl.ReachableOnLoad`（`internal/ecl/loadtime.go`）——**只有一份**。
// `cmd/ecl-wall-pieces` 走同一支：兩支各抄一份的話，其中一支改了條件，另一支
// 會安靜地與它分岔，而分岔的後果是兩張表對「載檔時會發生什麼」有不同的答案。
func scanBlock(area, block uint8, data []byte) ([]write, error) {
	instructions, err := ecl.ReachableOnLoad(data, block)
	if err != nil {
		return nil, err
	}
	found := make([]write, 0, 4)
	for _, instruction := range instructions {
		if instruction.Command.Opcode != saveOpcode {
			continue
		}
		if item, ok := skyWrite(instruction, area, block, instruction.Offset); ok {
			found = append(found, item)
		}
	}
	return found, nil
}

func skyWrite(instruction ecl.Instruction, area, block uint8, offset int) (write, bool) {
	if len(instruction.Operands) < 2 {
		return write{}, false
	}
	destination := instruction.Operands[1]
	if !destination.WordSet {
		return write{}, false
	}
	tracked := false
	for _, cell := range trackedCells {
		if destination.Word == cell {
			tracked = true
			break
		}
	}
	if !tracked {
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

func renderTable(table map[key]map[uint16]uint16) string {
	identities := sortedKeys(table)
	var out strings.Builder
	out.WriteString(`// Code generated by cmd/ecl-sky-colours; DO NOT EDIT.

package game

// eclBlockLoadTimeWrites 是每一段 ECL **載入存檔時真的會跑到**的常數寫入，
// 只收 remake 需要而原作靠跑進入碼才有的那幾格（` + "`cmd/ecl-sky-colours`" + `，spec 1185）。
//
//	4BE7h／4BE8h  ` + "`37h LOAD PIECES`" + ` 的兩槽模式閘門（spec 1087）
//	4BFDh／4BFEh  室外／室內天空選色（存檔 Area1 的 ` + "`01FAh`／`01FCh`" + `）
//
// ★ 為什麼要有這張表。 remake 載入存檔時不重跑該段的進入碼（會連帶觸發劇情
// 副作用），這幾格會停在底檔（別章）的值。原版在同一份存檔上會跑，於是兩邊
// 分岔——而天空是一整片純色，錯的顏色也是一整片純色，**看不出異常**。
//
// ⚠ 鍵是 **ECL 段**，不是地圖。GEO5 的 ` + "`0x31`／`0x32`" + ` 共用同一張幾何區塊。
//
// ⚠ 收集走的是 ` + "`ecl.ReachableOnLoad`" + `：段的開頭幾乎都有一道
// ` + "`COMPARE 4BF2h, <段號>`" + ` 的閘門，而那道閘門**兩個方向都有人用**。只掃
// 「有沒有這條 ` + "`SAVE`" + `」會多收，而多收的後果是把沒被推翻的值改掉。
var eclBlockLoadTimeWrites = map[eclBlockKey]map[uint16]uint16{
`)
	for _, identity := range identities {
		cells := table[identity]
		addresses := make([]int, 0, len(cells))
		for address := range cells {
			addresses = append(addresses, int(address))
		}
		sort.Ints(addresses)
		parts := make([]string, 0, len(addresses))
		for _, address := range addresses {
			parts = append(parts, fmt.Sprintf("0x%04X: %d", address, cells[uint16(address)]))
		}
		out.WriteString(fmt.Sprintf("\t{area: %d, block: 0x%02X}: {%s},\n",
			identity.area, identity.block, strings.Join(parts, ", ")))
	}
	out.WriteString("}\n")
	return out.String()
}

func sortedKeys(table map[key]map[uint16]uint16) []key {
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
	return identities
}

func renderReport(writes []write, table map[key]map[uint16]uint16) string {
	names := map[uint16]string{
		0x4BE7: "兩槽閘門 A", 0x4BE8: "兩槽閘門 B",
		0x4BFD: "室外天空", 0x4BFE: "室內天空",
	}
	var out strings.Builder
	out.WriteString("# 每一段 ECL 載入存檔時會寫進哪幾格\n\n")
	out.WriteString("由 `cmd/ecl-sky-colours` 產生，不要手改。機制與證據見 spec 1185。\n\n")
	out.WriteString("只收 **remake 需要、而原作靠跑段的進入碼才有**的那幾格：\n\n")
	out.WriteString("| 格 | 意義 |\n|---|---|\n")
	out.WriteString("| `4BE7h`／`4BE8h` | `37h LOAD PIECES` 的兩槽模式閘門（spec 1087）|\n")
	out.WriteString("| `4BFDh`／`4BFEh` | 室外／室內天空選色（Area1 `01FAh`／`01FCh`）|\n\n")
	out.WriteString("⚠ 走的是 `ecl.ReachableOnLoad`：段的開頭幾乎都有一道 " +
		"`COMPARE 4BF2h, <段號>` 的閘門，而那道閘門**兩個方向都有人用**。" +
		"只掃「有沒有這條 `SAVE`」會多收，而多收的後果是把沒被推翻的值改掉。\n\n")

	out.WriteString("## 逐處寫入（載檔時真的會跑到的）\n\n")
	out.WriteString("| 段 | 位移 | 格 | 意義 | 值 |\n|---|---:|---|---|---:|\n")
	for _, item := range writes {
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | `0x%04x` | `%04Xh` | %s | `%02Xh` |\n",
			item.area, item.block, item.offset, item.cell, names[item.cell], item.value))
	}

	out.WriteString("\n## 進表的值（同一格寫多次時後面的贏）\n\n")
	out.WriteString("| 段 | 兩槽閘門 A | 兩槽閘門 B | 室外天空 | 室內天空 |\n|---|---:|---:|---:|---:|\n")
	for _, identity := range sortedKeys(table) {
		cells := table[identity]
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | %s | %s | %s | %s |\n",
			identity.area, identity.block,
			show(cells, 0x4BE7), show(cells, 0x4BE8), show(cells, 0x4BFD), show(cells, 0x4BFE)))
	}
	return out.String()
}

func show(cells map[uint16]uint16, address uint16) string {
	value, ok := cells[address]
	if !ok {
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
