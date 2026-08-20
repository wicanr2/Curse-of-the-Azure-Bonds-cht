// ecl-block-graph 產生 ECL block 之間的轉移圖：每個 block 的 `NEWECL` 出邊，
// 以及寫進 `4BF2h`（`bank0 +1E4h`，引擎主迴圈用來決定載哪一個 block）的立即值。
//
// ★ 可達性用 `ecl.TraceGraph`，它**會跟 `ON GOTO`／`ON GOSUB` 的每一個目的地**。
// `docs/audit/ecl-event-catalog.md` 那份不跟，所以它看到的邊只有這裡的一小部分
// ——那是假零，不能拿來當「這個 block 沒有出口」的證據。
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

type blockKey struct{ member, block string }

type blockGraph struct {
	instructions int
	payload      int
	newECL       map[int]int
	lastECL      map[int]int
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	out := flag.String("out", "docs/audit/ecl-block-graph.md", "輸出的 markdown")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	graphs := map[blockKey]*blockGraph{}
	for chapter := 1; chapter <= 6; chapter++ {
		member := fmt.Sprintf("ECL%d.DAX", chapter)
		payload := memberPayload(archive, member)
		if payload == nil {
			log.Fatalf("image 裡沒有 %s", member)
		}
		blocks, err := daxParse(payload)
		if err != nil {
			log.Fatalf("%s: %v", member, err)
		}
		for _, raw := range blocks {
			key := blockKey{member, fmt.Sprintf("0x%02X", raw.Entry.ID)}
			graphs[key] = traceBlock(raw.Data)
		}
	}

	keys := make([]blockKey, 0, len(graphs))
	for key := range graphs {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].member != keys[j].member {
			return keys[i].member < keys[j].member
		}
		return keys[i].block < keys[j].block
	})

	var report strings.Builder
	report.WriteString(`# ECL block 轉移圖

由 ` + "`cmd/ecl-block-graph`" + ` 產生，不要手改。

- 出邊是 ` + "`20h NEWECL`" + ` 的立即運算元。
- ` + "`SAVE→4BF2h`" + ` 是寫進 ` + "`bank0 +1E4h`" + `（LastECL）的立即值。引擎主迴圈
  （` + "`overlay-02:3772`" + `）載入的就是這一格指的 block，並在載完之後把它寫回去，
  所以 block 寫自己的編號是「記錄我是誰」，不是轉移。
- ⚠ 可達性用 ` + "`ecl.TraceGraph`" + `，**會跟 ` + "`ON GOTO`" + ` 的每一個目的地**。
  ` + "`docs/audit/ecl-event-catalog.md`" + ` 那份不跟，它看到的邊只有這裡的一小部分。
  拿那一份判斷「這個 block 沒有出口」會得到假零。

| member | block | 可達指令 | payload | ` + "`NEWECL`" + ` 出邊 | ` + "`SAVE→4BF2h`" + ` |
|---|---|---:|---:|---|---|
`)
	edges, noExit := 0, []string{}
	for _, key := range keys {
		g := graphs[key]
		edges += len(g.newECL)
		if len(g.newECL) == 0 {
			noExit = append(noExit, key.member+"／"+key.block)
		}
		report.WriteString(fmt.Sprintf("| `%s` | `%s` | %d | %d | %s | %s |\n",
			key.member, key.block, g.instructions, g.payload,
			formatTargets(g.newECL), formatTargets(g.lastECL)))
	}
	report.WriteString(fmt.Sprintf("\n## 摘要\n\n| 項目 | 數量 |\n|---|---:|\n| block | %d |\n"+
		"| 不重複 `NEWECL` 出邊 | %d |\n| 沒有出邊的 block | %d（%s）|\n",
		len(keys), edges, len(noExit), strings.Join(noExit, "、")))
	report.WriteString("\n沒有出邊的 block 是開場與結局，不是資料缺漏。\n")

	if err := os.WriteFile(*out, []byte(report.String()), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("block=%d NEWECL 出邊=%d 沒有出邊=%d → %s\n", len(keys), edges, len(noExit), *out)
}

// daxParse 只是把 dax.Parse 包一層，讓測試不必再 import 一次。
func daxParse(payload []byte) ([]dax.Block, error) { return dax.Parse(payload) }

func memberPayload(archive *zip.ReadCloser, member string) []byte {
	for _, file := range archive.File {
		if file.Name != member {
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

func traceBlock(data []byte) *blockGraph {
	result := &blockGraph{payload: len(data), newECL: map[int]int{}, lastECL: map[int]int{}}
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return result
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return result
	}
	seen := map[int]bool{}
	for _, instruction := range graph.Instructions {
		if seen[instruction.Offset] {
			continue
		}
		seen[instruction.Offset] = true
		operands := instruction.Operands
		switch instruction.Command.Opcode {
		case 0x20: // NEWECL
			if len(operands) >= 1 && !operands[0].WordSet {
				result.newECL[int(operands[0].Low)]++
			}
		case 0x09: // SAVE
			if len(operands) >= 2 && operands[1].WordSet && operands[1].Word == 0x4BF2 &&
				!operands[0].WordSet {
				result.lastECL[int(operands[0].Low)]++
			}
		}
	}
	result.instructions = len(seen)
	return result
}

func formatTargets(targets map[int]int) string {
	if len(targets) == 0 {
		return "—"
	}
	ids := make([]int, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("`0x%02X`", id))
	}
	return strings.Join(parts, "、")
}
