// Command map-spawn-audit 回答「每張第一人稱地圖的進場落點是誰決定的」。
//
// ★ 存在的理由：spec 1183 普查完 `720Fh`／`7210h`／`7211h` 的寫入者之後確定
// **原作沒有「換 ECL block 的引擎進場放置」**——落點是腳本自己寫的，而且不是
// 統一步驟（`ECL2/0x03:00DBh` 還帶條件 `4C2A == 1`）。於是 `ViewMirror.Block`
// 那個擋板到底該不該留，就從模型問題變成**逐張地圖的資料問題**：腳本自己有寫
// 進場座標的地圖應該讓腳本贏，沒寫的才需要 game pack 宣告的 `spawn` 補值。
//
// 這一支把那個問題變成一張可以引用的表。
//
// 判準：在該 block 的控制流走訪裡找 `09h SAVE <立即數> C04B|C04C|C04D`，
// 把位址相鄰的寫入併成一組座標。**只收立即數**——`SAVE 4BF0 C04B` 那種是把
// 別的格子搬進來（退回上一格的出口，spec 1157），不是進場放置。
//
// ⚠ 走訪跟 `cmd/ecl-cell-refs` 同一條路：從五個生命週期入口跟著跳躍走。
// 跟不到的碼不會出現，所以**查無寫入代表這條路沒找到，不代表不存在**。
//
// ⚠ 這支**不判**哪一組才是「進場」那一組。一個 block 通常有好幾組（進場一組、
// 劇情傳送好幾組），要哪一組是進場得看呼叫關係與條件。表列出全部，讓人去看。
//
// 用法：
//
//	./tools/go.sh run ./cmd/map-spawn-audit -output docs/audit/map-spawn-sources.md
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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

// 三個地圖暫存器的 ECL 位址（spec 1150）。
const (
	cellX      uint16 = 0xC04B
	cellY      uint16 = 0xC04C
	cellFacing uint16 = 0xC04D
)

// placement 是一組相鄰的立即數座標寫入。缺的欄位用 nil 表示「這一組沒寫它」。
type placement struct {
	Offset int
	// End 是這一組最後一條 `SAVE` 的位移。★ 判斷「NEWECL 緊接在這一組後面」
	// 要用它：只有第一條的位移時，三條 `SAVE` 的長度得用猜的。
	End    int
	X      *int
	Y      *int
	Facing *int
}

// String 印成 `(x,y,facing)`，沒寫的欄位印 `-`。
func (p placement) String() string {
	part := func(v *int) string {
		if v == nil {
			return "-"
		}
		return fmt.Sprint(*v)
	}
	// 朝向在暫存器裡是折過的 0..3，畫面上的八向是它乘二（spec 1150）。
	facing := "-"
	if p.Facing != nil {
		facing = fmt.Sprintf("%d(=%d)", *p.Facing, (*p.Facing&3)*2)
	}
	return fmt.Sprintf("(%s,%s,%s)", part(p.X), part(p.Y), facing)
}

// matches 回答「這一組等於宣告的 spawn 嗎」。
//
// ⚠ 要求**同時寫過 X 與 Y**。只寫朝向的那種是走位收尾（`ECL3/0x10:1991h`），
// 拿它去比對會因為 X／Y 是 nil 而「無條件相符」——第一版就是這樣把猶拉什
// 判成 `script-agrees`，而那張圖的宣告值其實是走位**終點**，腳本從來沒寫過。
// 朝向沒寫就不比對：只寫座標、朝向刻意不變是原作的常見形狀（spec 1157）。
func (p placement) matches(spawn goldenbox.MapSpawn) bool {
	if p.X == nil || p.Y == nil {
		return false
	}
	if *p.X != spawn.X || *p.Y != spawn.Y {
		return false
	}
	return p.Facing == nil || uint8((*p.Facing&3)*2) == spawn.Direction
}

// gapBytes 是「相鄰」的容忍度。原作那幾組是三條連續的 `SAVE`，每條 6 bytes，
// 所以同一組的位址差不會超過這個值。
const gapBytes = 24

func placements(data []byte) ([]placement, error) {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil, err
	}
	type write struct {
		offset int
		cell   uint16
		value  int
	}
	seen := map[int]bool{}
	writes := make([]write, 0, 8)
	for _, instruction := range graph.Instructions {
		if instruction.Command.Opcode != 0x09 || len(instruction.Operands) != 2 {
			continue
		}
		source, target := instruction.Operands[0], instruction.Operands[1]
		if !target.WordSet || source.WordSet {
			continue
		}
		if target.Word != cellX && target.Word != cellY && target.Word != cellFacing {
			continue
		}
		if seen[instruction.Offset] {
			continue
		}
		seen[instruction.Offset] = true
		writes = append(writes, write{offset: instruction.Offset, cell: target.Word, value: int(source.Low)})
	}
	sort.Slice(writes, func(i, j int) bool { return writes[i].offset < writes[j].offset })

	out := make([]placement, 0, 4)
	for _, item := range writes {
		value := item.value
		if len(out) == 0 || item.offset-out[len(out)-1].Offset > gapBytes {
			out = append(out, placement{Offset: item.offset})
		}
		current := &out[len(out)-1]
		current.End = item.offset
		switch item.cell {
		case cellX:
			current.X = &value
		case cellY:
			current.Y = &value
		case cellFacing:
			current.Facing = &value
		}
	}
	return out, nil
}


// incomingGroup 是**別的 block** 在 `NEWECL` 切到這一張圖之前寫下的座標。
//
// ★ 為什麼需要它。 原作的進場落點是腳本寫的（spec 1183），而寫它的腳本
// **常常在來源那一段**，不在目的地那一段：
//
//	ECL5/0x32  0513 PRINTCLEAR "YOU ARE HEADING UP INTO THE WIZARD'S TOWER."
//	           0537 SAVE 07 → C04B
//	           053D SAVE 0F → C04C
//	           0543 SAVE 03 → C04D
//	           0549 NEWECL 33
//
// 只掃目的地那一段的話，這一組看不到，於是 game pack 宣告的 `(7,15,6)` 會被
// 判成 `mismatch`——而它其實**逐欄相符**，只是證據在另一段裡（spec 1198）。
type incomingGroup struct {
	FromArea  uint8
	FromBlock uint8
	NewECLAt  int
	Group     placement
}

// incomingGapBytes 是「`NEWECL` 緊接在這一組後面」的容忍度。原作那幾處是
// 三條 `SAVE`（各 6 bytes）之後就是 `NEWECL`，所以差距很小；放寬到這個值容得下
// 中間夾一兩條旗標寫入。
const incomingGapBytes = 48

// incomingPlacements 找出全 corpus 裡「切到 (area, dest) 之前剛寫過座標」的地方。
//
// ⚠ 只看**同一個 ECL 成員**：`NEWECL` 的運算元是 block 編號，而 block 編號在
// 成員之間會重複。跨成員的切換走的是另一條路（area 換檔），不是這一個 opcode。
func incomingPlacements(blocks map[uint8]map[uint8][]byte, area, dest uint8) []incomingGroup {
	out := make([]incomingGroup, 0, 2)
	for source, data := range blocks[area] {
		if source == dest {
			continue
		}
		groups, err := placements(data)
		if err != nil || len(groups) == 0 {
			continue
		}
		points, _, err := ecl.EntryPoints(data, 5)
		if err != nil {
			continue
		}
		starts := make([]int, 0, len(points))
		for _, point := range points {
			starts = append(starts, int(point)-ecl.CodeAddressBase)
		}
		graph, err := ecl.TraceGraph(data, starts, len(data)*8)
		if err != nil {
			continue
		}
		seen := map[int]bool{}
		for _, instruction := range graph.Instructions {
			if instruction.Command.Opcode != 0x20 || len(instruction.Operands) != 1 {
				continue
			}
			operand := instruction.Operands[0]
			if operand.WordSet || operand.Low != dest {
				continue
			}
			if seen[instruction.Offset] {
				continue
			}
			seen[instruction.Offset] = true
			for _, group := range groups {
				gap := instruction.Offset - group.End
				if gap <= 0 || gap > incomingGapBytes {
					continue
				}
				out = append(out, incomingGroup{
					FromArea: area, FromBlock: source,
					NewECLAt: instruction.Offset, Group: group,
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FromBlock != out[j].FromBlock {
			return out[i].FromBlock < out[j].FromBlock
		}
		return out[i].NewECLAt < out[j].NewECLAt
	})
	return out
}

type row struct {
	MapID    string
	Area     uint8
	Block    uint8
	Spawn    *goldenbox.MapSpawn
	Found    []placement
	Incoming []incomingGroup
	Verdict  string
	ScanNote string
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	output := flag.String("output", "", "Markdown output path (empty prints to stdout)")
	flag.Parse()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	blocks := map[uint8]map[uint8][]byte{}
	for member := 1; member <= 6; member++ {
		payload := memberPayload(&archive.Reader, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			log.Fatalf("image has no ECL%d.DAX", member)
		}
		parsed, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		blocks[uint8(member)] = map[uint8][]byte{}
		for _, block := range parsed {
			blocks[uint8(member)][block.Entry.ID] = block.Data
		}
	}

	rows := make([]row, 0, 32)
	for _, definition := range pack.Maps {
		if definition.Kind != "first_person" {
			continue
		}
		// ⚠ 沒有 `script_block` 的地圖也要列。`Pack.FindMapByKindScript` 找不到
		// 精確對應時會退回**同 area 的預設地圖**，所以那種宣告一樣會被套用；
		// 只掃有 script block 的會漏掉它們。
		if definition.ScriptBlock == nil {
			rows = append(rows, row{MapID: definition.ID, Area: definition.AreaID,
				Spawn: definition.Spawn, Verdict: "area-default", ScanNote: "no script_block"})
			continue
		}
		item := row{MapID: definition.ID, Area: definition.AreaID, Block: *definition.ScriptBlock,
			Spawn: definition.Spawn}
		data, ok := blocks[definition.AreaID][*definition.ScriptBlock]
		if !ok {
			item.Verdict = "no-such-block"
			item.ScanNote = fmt.Sprintf("ECL%d has no block 0x%02X", definition.AreaID, *definition.ScriptBlock)
			rows = append(rows, item)
			continue
		}
		found, scanErr := placements(data)
		if scanErr != nil {
			item.Verdict = "scan-error"
			item.ScanNote = scanErr.Error()
			rows = append(rows, item)
			continue
		}
		item.Found = found
		item.Incoming = incomingPlacements(blocks, definition.AreaID, *definition.ScriptBlock)
		item.Verdict = verdict(item)
		rows = append(rows, item)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Area != rows[j].Area {
			return rows[i].Area < rows[j].Area
		}
		if rows[i].Block != rows[j].Block {
			return rows[i].Block < rows[j].Block
		}
		return rows[i].MapID < rows[j].MapID
	})

	counts := map[string]int{}
	for _, item := range rows {
		counts[item.Verdict]++
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# Where each first-person map's entry position comes from\n\n")
	fmt.Fprintf(&report, "Generated by `cmd/map-spawn-audit`; do not hand-edit. Rationale: spec 1183/1184.\n\n")
	fmt.Fprintf(&report, "The original has no engine-side placement on ECL block change, so the entry cell "+
		"is whatever the script writes. This table pairs every declared `spawn` with the immediate "+
		"`SAVE <imm> C04B|C04C|C04D` groups reachable in the same block.\n\n")
	fmt.Fprintf(&report, "Verdicts: `script-agrees` the declared spawn equals a script group in the same block; "+
		"`script-agrees-incoming` it equals a group written just before a `NEWECL` into this block from another block; "+
		"`mismatch` the script writes coordinates but none equals the declared spawn; "+
		"`declared-only` no immediate script write was reached, the declaration is the only source; "+
		"`script-only` the script writes coordinates and no spawn is declared; "+
		"`none` neither; "+
		"`area-default` the definition declares no script block, so it only applies as the area fallback.\n\n")
	fmt.Fprintf(&report, "Caveats: the walk follows control flow from the five lifecycle entries, so an "+
		"unreached write is absent from this table and absence is not proof. The table does not decide "+
		"which group is the entry placement -- a block usually has several (entry plus story teleports).\n\n")

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Fprintf(&report, "| verdict | maps |\n|---|---:|\n")
	for _, key := range keys {
		fmt.Fprintf(&report, "| `%s` | %d |\n", key, counts[key])
	}
	fmt.Fprintf(&report, "| total | %d |\n\n", len(rows))

	fmt.Fprintf(&report, "| map | ECL | block | declared spawn | script groups | incoming NEWECL groups | verdict |\n|---|---:|---:|---|---|---|---|\n")
	for _, item := range rows {
		spawn := "-"
		if item.Spawn != nil {
			spawn = fmt.Sprintf("(%d,%d,%d)", item.Spawn.X, item.Spawn.Y, item.Spawn.Direction)
		}
		groups := "-"
		if len(item.Found) > 0 {
			parts := make([]string, 0, len(item.Found))
			for _, found := range item.Found {
				parts = append(parts, fmt.Sprintf("`%04Xh`%s", found.Offset, found))
			}
			groups = strings.Join(parts, "<br>")
		}
		note := ""
		if item.ScanNote != "" {
			note = " " + item.ScanNote
		}
		block := fmt.Sprintf("`0x%02X`", item.Block)
		if item.Verdict == "area-default" {
			block = "-"
		}
		arrivals := "-"
		if len(item.Incoming) > 0 {
			parts := make([]string, 0, len(item.Incoming))
			for _, group := range item.Incoming {
				parts = append(parts, fmt.Sprintf("`0x%02X`@`%04Xh`%s",
					group.FromBlock, group.NewECLAt, group.Group))
			}
			arrivals = strings.Join(parts, "<br>")
		}
		fmt.Fprintf(&report, "| `%s` | %d | %s | %s | %s | %s | `%s`%s |\n",
			item.MapID, item.Area, block, spawn, groups, arrivals, item.Verdict, note)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	for _, key := range keys {
		fmt.Fprintf(os.Stderr, "%s=%d ", key, counts[key])
	}
	fmt.Fprintf(os.Stderr, "total=%d\n", len(rows))
}

func verdict(item row) string {
	switch {
	case item.Spawn == nil && len(item.Found) == 0 && len(item.Incoming) == 0:
		return "none"
	case item.Spawn == nil:
		return "script-only"
	case len(item.Found) == 0 && len(item.Incoming) == 0:
		return "declared-only"
	}
	for _, found := range item.Found {
		if found.matches(*item.Spawn) {
			return "script-agrees"
		}
	}
	// ⚠ 目的地那一段沒寫**不代表**腳本沒寫。切進來的那一段常常先寫座標再
	// `NEWECL`，證據在別的 block 裡（spec 1198 的巫師塔）。
	for _, group := range item.Incoming {
		if group.Group.matches(*item.Spawn) {
			return "script-agrees-incoming"
		}
	}
	return "mismatch"
}

func memberPayload(archive *zip.Reader, name string) []byte {
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
