// ecl-treasure-clear 回答「`27h TREASURE` 之後還走得到 `1Ch CLEARMONSTERS` 嗎」。
//
// ★ 存在的理由：`1Ch` 的名字只講了一半。原作（DOS `overlay-02:120Eh`，37 條）除了
// 釋放節點鏈與清「有怪要打」旗標，還把 **`DS:6F70h` 起的 28 個位元組歸零**、
// 沿 **`DS:6F8Ch`** 鏈逐節點 `FreeMem(63)`。而那兩塊正是戰利品：
// `DS:6F70h + i × 4`（`i = 0..6`）是七種貨幣／寶石／珠寶的池（spec 1059），
// `DS:6F8Ch` 是 `27h TREASURE` 串進去的物品節點鏈（spec 1087）。
//
// ⇒ **`1Ch` 會把還沒領走的戰利品整堆丟掉。** remake 這一側 `TREASURE` 進的是
// `pendingTreasure` 佇列，`1Ch` 目前不碰它——所以「原作丟掉、remake 照發」這件事
// 有沒有真的發生，取決於 corpus 裡走不走得到那條路。這支就是去量它。
//
// ⚠ 前向走訪要**同時**收跳躍邊與循序前驅：`TraceGraph.Edges` 只記跳躍，只用它
// 會得到「一處都走不到」的假零（`cmd/ecl-combat-sites` 踩過同一個坑）。
//
// 用法：
//
//	go run ./cmd/ecl-treasure-clear
//	go run ./cmd/ecl-treasure-clear -out docs/audit/ecl-treasure-clear.md
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
	opExit           = 0x00
	opGoto           = 0x01
	opReturn         = 0x13
	opClearMonsters  = 0x1C
	opCombat         = 0x24
	opTreasure       = 0x27
	maxForwardSteps  = 8192
)

// pair 是一處「`27h` 之後走得到的 `1Ch`」。
type pair struct {
	member   int
	block    uint8
	treasure int
	clear    int
	// viaCombat 為真代表兩者之間至少一條路上經過 `24h COMBAT`。
	viaCombat bool
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	out := flag.String("out", "docs/audit/ecl-treasure-clear.md", "輸出的 markdown")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	pairs := make([]pair, 0, 32)
	treasures := 0
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			log.Fatalf("image 裡沒有 ECL%d.DAX", member)
		}
		blocks, err := dax.Parse(payload)
		if err != nil {
			log.Fatal(err)
		}
		for _, block := range blocks {
			found, sites, err := scanBlock(member, block.Entry.ID, block.Data)
			if err != nil {
				log.Fatalf("ECL%d/0x%02X: %v", member, block.Entry.ID, err)
			}
			pairs = append(pairs, found...)
			treasures += sites
		}
	}
	sort.Slice(pairs, func(a, b int) bool {
		if pairs[a].member != pairs[b].member {
			return pairs[a].member < pairs[b].member
		}
		if pairs[a].block != pairs[b].block {
			return pairs[a].block < pairs[b].block
		}
		if pairs[a].treasure != pairs[b].treasure {
			return pairs[a].treasure < pairs[b].treasure
		}
		return pairs[a].clear < pairs[b].clear
	})
	if err := os.WriteFile(*out, []byte(render(pairs, treasures)), 0o644); err != nil {
		log.Fatal(err)
	}
	viaCombat := 0
	for _, item := range pairs {
		if item.viaCombat {
			viaCombat++
		}
	}
	fmt.Printf("TREASURE=%d 之後走得到 CLEARMONSTERS 的配對=%d（其中經過 COMBAT=%d）→ %s\n",
		treasures, len(pairs), viaCombat, *out)
}

// fallsThrough 回答「這條指令跑完會不會接著跑下一條」。
func fallsThrough(instruction ecl.Instruction) bool {
	switch instruction.Command.Opcode {
	case opExit, opGoto, opReturn:
		return false
	}
	return true
}

func scanBlock(member int, id uint8, data []byte) ([]pair, int, error) {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil, 0, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil, 0, err
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

	// 前向邊：跳躍 ＋ 循序後繼。少了後者是假零。
	outgoing := map[int][]int{}
	for _, edge := range graph.Edges {
		outgoing[edge.From] = append(outgoing[edge.From], edge.To)
	}
	for index := 0; index+1 < len(offsets); index++ {
		current := unique[offsets[index]]
		if !fallsThrough(current) {
			continue
		}
		if current.Next != offsets[index+1] {
			continue
		}
		outgoing[offsets[index]] = append(outgoing[offsets[index]], offsets[index+1])
	}

	pairs := make([]pair, 0, 4)
	sites := 0
	for _, offset := range offsets {
		if unique[offset].Command.Opcode != opTreasure {
			continue
		}
		sites++
		for clear, viaCombat := range forwardClears(unique, outgoing, offset) {
			pairs = append(pairs, pair{member: member, block: id,
				treasure: offset, clear: clear, viaCombat: viaCombat})
		}
	}
	return pairs, sites, nil
}

// forwardClears 從一處 `27h` 往前走，回傳走得到的每一處 `1Ch`，以及路上有沒有
// 經過 `24h COMBAT`。
//
// ⚠ 「經過 COMBAT」是**存在某一條路經過**，不是每一條都經過——這個量只拿來分類，
// 不拿來下「一定先打完再清」的結論。
func forwardClears(unique map[int]ecl.Instruction, outgoing map[int][]int,
	start int) map[int]bool {
	found := map[int]bool{}
	type node struct {
		offset    int
		viaCombat bool
	}
	seen := map[node]bool{}
	queue := []node{}
	for _, next := range outgoing[start] {
		queue = append(queue, node{offset: next})
	}
	for steps := 0; len(queue) > 0 && steps < maxForwardSteps; steps++ {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		switch unique[current.offset].Command.Opcode {
		case opClearMonsters:
			if _, ok := found[current.offset]; !ok || current.viaCombat {
				found[current.offset] = found[current.offset] || current.viaCombat
			}
			// 清掉之後再往前走碰到的 `1Ch` 已經不是同一堆戰利品了。
			continue
		case opTreasure:
			// 又補了一堆新的，往前走碰到的 `1Ch` 屬於那一堆。
			continue
		case opCombat:
			current.viaCombat = true
		}
		for _, next := range outgoing[current.offset] {
			queue = append(queue, node{offset: next, viaCombat: current.viaCombat})
		}
	}
	return found
}

func render(pairs []pair, treasures int) string {
	var out strings.Builder
	out.WriteString("# `27h TREASURE` 之後走得到的 `1Ch CLEARMONSTERS`\n\n" +
		"由 `cmd/ecl-treasure-clear` 產生，不要手改。\n\n" +
		"`1Ch` 的名字只講了一半。原作（DOS `overlay-02:120Eh`）除了釋放節點鏈與清\n" +
		"「有怪要打」旗標（`8B69h`，spec 1095），還把 `DS:6F70h` 起的 28 個位元組歸零、\n" +
		"沿 `DS:6F8Ch` 鏈逐節點 `FreeMem(63)`。那兩塊正是戰利品——\n" +
		"`DS:6F70h + i × 4`（`i = 0..6`）是七種貨幣／寶石／珠寶的池（spec 1059），\n" +
		"`DS:6F8Ch` 是 `27h` 串進去的物品節點鏈（spec 1087）。\n\n" +
		"⇒ **`1Ch` 會把還沒領走的戰利品整堆丟掉。** 這份表列的是 corpus 裡真的走得到\n" +
		"那條路的地方。\n\n" +
		"⚠ 前向走訪同時收跳躍邊與循序後繼；只用跳躍邊會得到假零。\n" +
		"⚠ 「經過 `COMBAT`」是**存在某一條路經過**，不是每一條都經過。\n\n")
	out.WriteString("| 段 | `27h` 位移 | 走得到的 `1Ch` | 路上經過 `24h` |\n")
	out.WriteString("|---|---|---|---|\n")
	for _, item := range pairs {
		via := "否"
		if item.viaCombat {
			via = "是"
		}
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | `%#04x` | `%#04x` | %s |\n",
			item.member, item.block, item.treasure, item.clear, via))
	}
	if len(pairs) == 0 {
		out.WriteString("| — | — | — | — |\n")
	}
	viaCombat := 0
	for _, item := range pairs {
		if item.viaCombat {
			viaCombat++
		}
	}
	out.WriteString(fmt.Sprintf("\n## 摘要\n\n| 項目 | 數 |\n|---|---:|\n"+
		"| `27h TREASURE` 的處數 | %d |\n"+
		"| 之後走得到 `1Ch` 的配對 | %d |\n"+
		"| 其中路上經過 `24h COMBAT` | %d |\n", treasures, len(pairs), viaCombat))
	return out.String()
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
