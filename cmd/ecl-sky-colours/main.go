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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
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
	image := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.79f855c8b433"))
	output := flag.String("output", "", tooltext.Text("h.716496383ea7"))
	report := flag.String("report", "", tooltext.Text("h.0dd7f41cdcdb"))
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	writes := make([]write, 0, 64)
	conditionals := make([]conditionalBlock, 0, 8)
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
			conditional, condErr := scanConditional(uint8(area), block.Entry.ID, block.Data)
			if condErr != nil {
				log.Fatal(tooltext.Format("h.1a4a2624c413", area, block.Entry.ID, condErr))
			}
			conditionals = append(conditionals, conditional...)
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
	text := renderReport(writes, table) + renderConditional(conditionals)
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
	fmt.Fprintf(os.Stderr, "writes=%d guarded=%d blocks=%d conditional=%d\n",
		len(writes), guarded, len(table), len(conditionals))
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
// terrainProbe 是「拿當下那一格的地形碼去判 `IF`」時試的幾個值。
//
// ★ 為什麼要試：`ECL2/0x04` 的進入碼 `GOSUB` 進去之後是
// `COMPARE C04F 0x95 / IF = / SAVE 0A 4BFE / IF <> / SAVE 09 4BFE`
// ——天花板顏色由**站在哪一格**決定，靜態的表表達不了（spec 1185）。
// 這裡不是要把它收進表，是要把**有這種寫法的段**列出來，讓「表是完整的」
// 這句話不會被誤讀。
var terrainProbe = []uint16{0x00, 0x80, 0x95, 0xC0}

// conditionalBlock 是一段「載檔時的寫入會隨當下格子而變」的紀錄。
type conditionalBlock struct {
	area, block uint8
	cell        uint16
	// byTerrain 是地形碼 → 寫進去的值。
	byTerrain map[uint16]uint16
}

// scanConditional 找出「同一段、同一格，換個地形碼就寫不同值」的地方。
func scanConditional(area, block uint8, data []byte) ([]conditionalBlock, error) {
	perCell := map[uint16]map[uint16]uint16{}
	for _, terrain := range terrainProbe {
		value := terrain
		instructions, err := ecl.ReachableOnLoadWithMemory(data, block,
			func(address uint16) (uint16, bool) {
				if address == terrainRegisterCell {
					return value, true
				}
				return 0, false
			})
		if err != nil {
			return nil, err
		}
		for _, instruction := range instructions {
			if instruction.Command.Opcode != saveOpcode {
				continue
			}
			item, ok := skyWrite(instruction, area, block, instruction.Offset)
			if !ok {
				continue
			}
			if perCell[item.cell] == nil {
				perCell[item.cell] = map[uint16]uint16{}
			}
			perCell[item.cell][terrain] = item.value
		}
	}
	result := make([]conditionalBlock, 0, 2)
	for cell, byTerrain := range perCell {
		if len(byTerrain) < len(terrainProbe) {
			// 有些地形碼下根本沒有寫入 ⇒ 也是「看情況」。
			result = append(result, conditionalBlock{area: area, block: block, cell: cell, byTerrain: byTerrain})
			continue
		}
		first := byTerrain[terrainProbe[0]]
		for _, terrain := range terrainProbe {
			if byTerrain[terrain] != first {
				result = append(result, conditionalBlock{area: area, block: block, cell: cell, byTerrain: byTerrain})
				break
			}
		}
	}
	sort.SliceStable(result, func(left, right int) bool { return result[left].cell < result[right].cell })
	return result, nil
}

// terrainRegisterCell 是 `C04Fh`：由隊伍位置推出來的牆頂／地形碼。
const terrainRegisterCell uint16 = 0xC04F

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
	out.WriteString(tooltext.Text("h.ec3ac323dff8") + "`cmd/ecl-sky-colours`" + `，spec 1185）。
//
//	4BE7h／4BE8h  ` + "`37h LOAD PIECES`" + tooltext.Text("h.03e51a0fabd1") + "`01FAh`／`01FCh`" + tooltext.Text("h.1e22e31ea9bd") + "`0x31`／`0x32`" + tooltext.Text("h.3b44943f9851") + "`ecl.ReachableOnLoad`" + tooltext.Text("h.bf1c02df3db5") + tooltext.Text("h.2065b65257d9") + tooltext.Text("h.d2bfffa055f0") + "`SAVE`" + tooltext.Text("h.3adfff7816f1"))
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

// renderConditional 把「看情況」那幾段列出來。
//
// ★ 為什麼要列：這張表宣稱「載檔時真的會跑到的常數寫入」，而**條件式的寫入
// 進不了表**——不列出來的話，讀者會把「表裡沒有」讀成「原作沒寫」。
// remake 那一側靠 `applyLoadTimeECLWritesFromScript` 在載檔當下重走一次
// （`ecl.ReachableOnLoadWithMemory`）處理它們。
func renderConditional(items []conditionalBlock) string {
	var report strings.Builder
	report.WriteString(tooltext.Text("h.7c4150feb6ab"))
	if len(items) == 0 {
		report.WriteString(tooltext.Text("h.0c76db18381a"))
		return report.String()
	}
	fmt.Fprintf(&report, tooltext.Text("h.cac4027948e8")+
		tooltext.Text("h.3836433e4e44"), terrainProbe)
	report.WriteString(tooltext.Text("h.cc8e6c570508"))
	for _, item := range items {
		parts := make([]string, 0, len(terrainProbe))
		for _, terrain := range terrainProbe {
			if value, ok := item.byTerrain[terrain]; ok {
				parts = append(parts, fmt.Sprintf("`%02Xh`→`%02Xh`", terrain, value))
			} else {
				parts = append(parts, tooltext.Format("h.d6ccb7ac6911", terrain))
			}
		}
		fmt.Fprintf(&report, "| `ECL%d/0x%02X` | `%04Xh` | %s |\n",
			item.area, item.block, item.cell, strings.Join(parts, "、"))
	}
	report.WriteString(tooltext.Text("h.06b75e1c7e5c") +
		tooltext.Text("h.97756325ef34"))
	return report.String()
}

func renderReport(writes []write, table map[key]map[uint16]uint16) string {
	names := map[uint16]string{
		0x4BE7: tooltext.Text("h.04fa08eb07d2"), 0x4BE8: tooltext.Text("h.7abf3fff85d8"),
		0x4BFD: tooltext.Text("h.b2f940763c10"), 0x4BFE: tooltext.Text("h.79f51341b8e2"),
	}
	var out strings.Builder
	out.WriteString(tooltext.Text("h.56ef0a629eba"))
	out.WriteString(tooltext.Text("h.6af0cb806c24"))
	out.WriteString(tooltext.Text("h.b612a1d43792"))
	out.WriteString(tooltext.Text("h.2426643e7da4"))
	out.WriteString(tooltext.Text("h.cfbfcee7d51c"))
	out.WriteString(tooltext.Text("h.fccff3ffdb03"))
	out.WriteString(tooltext.Text("h.6ff3a5c793b5") +
		tooltext.Text("h.936ade4769ee") +
		tooltext.Text("h.6fd47a6b9fc4"))

	out.WriteString(tooltext.Text("h.7850350470c1"))
	out.WriteString(tooltext.Text("h.c62c5359fbff"))
	for _, item := range writes {
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | `0x%04x` | `%04Xh` | %s | `%02Xh` |\n",
			item.area, item.block, item.offset, item.cell, names[item.cell], item.value))
	}

	out.WriteString(tooltext.Text("h.aea306409d18"))
	out.WriteString(tooltext.Text("h.5337e64551b0"))
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
