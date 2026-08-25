// ecl-combat-sites 盤點每一處 `24h COMBAT` 走的是三選一裡的哪一支。
//
// ★ `24h` 沒有運算元（arity 0）。它是**服務分派點**（spec 1095）：先看兩個引擎寫的
// 請求旗標，命中就跑商店或營地並跳過戰鬥，否則才進戰鬥。所以「這 199 處都在做
// 什麼」不是看運算元，是看**它前面有沒有把怪物擺上去**。
//
// 判準：沿著控制流**往回走**，看有沒有 `0Bh LOAD MONSTER` 或 `0Ch SETUP MONSTER`
// 到得了這一處 `24h`，而且中間沒有經過 `1Ch CLEARMONSTERS` 把場清掉。
//
// ⚠ 往回走用的是 `ecl.TraceGraph` 的邊，**不是位址順序**。位址上的前一條指令常常
// 屬於別的分支——用位址找會把不相干的 `LOAD MONSTER` 算進來。
//
// ⚠ 「沒有怪物」不等於「一定是商店或營地」：也可能是前一段留在場上的怪
// （`SETUP MONSTER` 在另一個 block），或是走訪跟不到的路徑。這支只把兩堆分開，
// 不替沒有怪物的那一堆下結論。
//
// 用法：
//
//	go run ./cmd/ecl-combat-sites
//	go run ./cmd/ecl-combat-sites -out docs/audit/ecl-combat-sites.md
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

const (
	opExit           = 0x00
	opGoto           = 0x01
	opLoadMonster    = 0x0B
	opSetupMonster   = 0x0C
	opReturn         = 0x13
	opClearMonsters  = 0x1C
	opCombat         = 0x24
	maxBackwardSteps = 4096
)

// fallsThrough 回答「這條指令跑完會不會接著跑下一條」。
//
// ★ `TraceGraph` 的 `Edges` **只記跳躍**，不記循序。反向走如果只用那些邊，
// 多數 `24h` 找不到任何前驅——**0/199 那種假零就是這樣來的**。
func fallsThrough(instruction ecl.Instruction) bool {
	switch instruction.Command.Opcode {
	case opExit, opGoto, opReturn:
		return false
	}
	return true
}

type site struct {
	member    int
	block     uint8
	offset    int
	armed     bool
	viaSetup  bool
	viaLoad   bool
	cleared   bool
	firstText string
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	out := flag.String("out", "docs/audit/ecl-combat-sites.md", tooltext.Text("h.aff4479ab1b9"))
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	sites := make([]site, 0, 256)
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			log.Fatal(tooltext.Format("h.f74a43cb81d6", member))
		}
		blocks, err := dax.Parse(payload)
		if err != nil {
			log.Fatal(err)
		}
		for _, block := range blocks {
			found, err := scanBlock(member, block.Entry.ID, block.Data)
			if err != nil {
				log.Fatalf("ECL%d/0x%02X: %v", member, block.Entry.ID, err)
			}
			sites = append(sites, found...)
		}
	}
	sort.Slice(sites, func(a, b int) bool {
		if sites[a].member != sites[b].member {
			return sites[a].member < sites[b].member
		}
		if sites[a].block != sites[b].block {
			return sites[a].block < sites[b].block
		}
		return sites[a].offset < sites[b].offset
	})
	if err := os.WriteFile(*out, []byte(render(sites)), 0o644); err != nil {
		log.Fatal(err)
	}
	armed := 0
	for _, item := range sites {
		if item.armed {
			armed++
		}
	}
	fmt.Print(tooltext.Format("h.ae8f7d746f14", len(sites), armed, len(sites)-armed, *out))
}

func scanBlock(member int, id uint8, data []byte) ([]site, error) {
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
	unique := map[int]ecl.Instruction{}
	for _, instruction := range graph.Instructions {
		unique[instruction.Offset] = instruction
	}
	offsets := make([]int, 0, len(unique))
	for offset := range unique {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)

	// 反向邊：跳躍的邊 ＋ **循序的前驅**。少了後者會得到假零。
	incoming := map[int][]int{}
	for _, edge := range graph.Edges {
		incoming[edge.To] = append(incoming[edge.To], edge.From)
	}
	for index := 0; index+1 < len(offsets); index++ {
		current := unique[offsets[index]]
		if !fallsThrough(current) {
			continue
		}
		// 只在「下一條可達指令就是它的 Next」時才算循序前驅：中間有洞代表
		// 那一段沒被走訪到，硬接會把不相干的兩條指令串起來。
		if current.Next != offsets[index+1] {
			continue
		}
		incoming[offsets[index+1]] = append(incoming[offsets[index+1]], offsets[index])
	}

	sites := make([]site, 0, 8)
	for _, offset := range offsets {
		if unique[offset].Command.Opcode != opCombat {
			continue
		}
		item := site{member: member, block: id, offset: offset}
		item.viaSetup, item.viaLoad, item.cleared = walkBack(unique, incoming, offset)
		item.armed = item.viaSetup || item.viaLoad
		item.firstText = firstTextAt(unique, offsets, offset)
		sites = append(sites, item)
	}
	return sites, nil
}

// walkBack 沿著反向邊往回走，看有沒有把怪物擺上去的指令到得了這一處。
//
// ⚠ 碰到 `CLEARMONSTERS` 就**停在那一條路上**：場已經清了，再往前的
// `LOAD MONSTER` 餵的不是這一場。
func walkBack(unique map[int]ecl.Instruction, incoming map[int][]int,
	start int) (setup, load, cleared bool) {
	seen := map[int]bool{start: true}
	queue := append([]int(nil), incoming[start]...)
	for steps := 0; len(queue) > 0 && steps < maxBackwardSteps; steps++ {
		offset := queue[0]
		queue = queue[1:]
		if seen[offset] {
			continue
		}
		seen[offset] = true
		switch unique[offset].Command.Opcode {
		case opSetupMonster:
			setup = true
		case opLoadMonster:
			load = true
		case opClearMonsters:
			// 這條路上的場已經清掉了，不再往前追。
			cleared = true
			continue
		}
		queue = append(queue, incoming[offset]...)
	}
	return setup, load, cleared
}

func firstTextAt(unique map[int]ecl.Instruction, offsets []int, start int) string {
	index := sort.SearchInts(offsets, start)
	for cursor := index; cursor < len(offsets) && cursor < index+8; cursor++ {
		for _, operand := range unique[offsets[cursor]].Operands {
			if len(operand.Packed) == 0 {
				continue
			}
			text := ecl.DecodePackedText(operand.Packed)
			text = strings.ReplaceAll(text, "|", "／")
			if len([]rune(text)) > 40 {
				text = string([]rune(text)[:40]) + "…"
			}
			return text
		}
	}
	return ""
}

func render(sites []site) string {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.9a862ceda69b") +
		tooltext.Text("h.e3266f22fa28") +
		tooltext.Text("h.6c18eb21b24d") +
		tooltext.Text("h.670df3d7fbb9") +
		tooltext.Text("h.7d57b2d68b44") +
		tooltext.Text("h.d9d46162e0ff") +
		tooltext.Text("h.df91786fc98c") +
		tooltext.Text("h.84a95e4fc910") +
		tooltext.Text("h.36011f32180a"))
	out.WriteString(tooltext.Text("h.73ef9e2c00d9"))
	out.WriteString("|---|---|---|---|---|\n")
	for _, item := range sites {
		source := "—"
		switch {
		case item.viaSetup && item.viaLoad:
			source = "`SETUP` ＋ `LOAD`"
		case item.viaSetup:
			source = "`SETUP MONSTER`"
		case item.viaLoad:
			source = "`LOAD MONSTER`"
		case item.cleared:
			source = tooltext.Text("h.25ad0121d8b2")
		}
		armed := tooltext.Text("h.0c70665b6eb6")
		if item.armed {
			armed = tooltext.Text("h.b5141d3d19e9")
		}
		text := item.firstText
		if text == "" {
			text = "—"
		} else {
			text = "「" + text + "」"
		}
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | `%#04x` | %s | %s | %s |\n",
			item.member, item.block, item.offset, armed, source, text))
	}
	armed, cleared := 0, 0
	blocks := map[string]bool{}
	for _, item := range sites {
		blocks[fmt.Sprintf("%d/%d", item.member, item.block)] = true
		if item.armed {
			armed++
		}
		if !item.armed && item.cleared {
			cleared++
		}
	}
	out.WriteString(fmt.Sprintf(tooltext.Text("h.e3b2a4f11e09")+
		tooltext.Text("h.f5e1d1e87120")+
		tooltext.Text("h.e1e77582ffac")+
		tooltext.Text("h.dc2349e37dc8"),
		len(sites), len(blocks), armed, len(sites)-armed, cleared))
	return out.String()
}

// syntheticInstruction 只給測試用：組一條只有 opcode 有意義的指令。
func syntheticInstruction(opcode byte) ecl.Instruction {
	return ecl.Instruction{Command: ecl.Command{Opcode: opcode}}
}

func memberPayload(archive *zip.ReadCloser, member string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, member) {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil
		}
		defer reader.Close()
		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}
		return payload
	}
	return nil
}
