// ecl-block-graph 產生 ECL block 之間的轉移圖：每個 block 的 `NEWECL` 出邊，
// 以及寫進 `4BF2h`（`bank0 +1E4h`，引擎主迴圈用來決定載哪一個 block）的立即值。
//
// ★ 可達性用 `ecl.TraceGraph`，它**會跟 `ON GOTO`／`ON GOSUB` 的每一個目的地**。
// `docs/audit/ecl-event-catalog.md` 那份不跟，所以它看到的邊只有這裡的一小部分
// ——那是假零，不能拿來當「這個 block 沒有出口」的證據。
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

type blockKey struct{ member, block string }

type blockGraph struct {
	instructions int
	payload      int
	newECL       map[int]int
	lastECL      map[int]int
	// mapBlock 是寫進 `4BC5h`（`bank0 +18Ah`，目前的 3D 地圖區塊編號）的立即值。
	mapBlock map[int]int
	// inDungeon 是寫進 `4BE6h`（`bank0 +1CCh`）的立即值。
	inDungeon map[int]int
	// loadFiles 是 `21h LOAD FILES` 的三個運算元組合。
	loadFiles map[string]int
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
- 「3D 地圖」是寫進 ` + "`4BC5h`" + `（` + "`bank0 +18Ah`" + `，目前的 3D 地圖區塊編號）
  的立即值，也就是**這個 block 會把玩家放進哪一張地城地圖**。空白代表它不換地圖
  （世界地圖 hub、純劇情 block 都是這樣）。
- ` + "`InDungeon`" + ` 是 ` + "`4BE6h`" + `（` + "`bank0 +1CCh`" + `）：
  0 是野外／世界地圖，非 0 是地城。
- ` + "`LOAD FILES`" + ` 的三個運算元決定那一段用哪一組 DAX 檔。
- ` + "`SAVE→4BF2h`" + ` 是寫進 ` + "`bank0 +1E4h`" + `（LastECL）的立即值。引擎主迴圈
  （` + "`overlay-02:3772`" + `）載入的就是這一格指的 block，並在載完之後把它寫回去，
  所以 block 寫自己的編號是「記錄我是誰」，不是轉移。
- ⚠ 可達性用 ` + "`ecl.TraceGraph`" + `，**會跟 ` + "`ON GOTO`" + ` 的每一個目的地**。
  ` + "`docs/audit/ecl-event-catalog.md`" + ` 那份不跟，它看到的邊只有這裡的一小部分。
  拿那一份判斷「這個 block 沒有出口」會得到假零。

| member | block | 可達指令 | ` + "`NEWECL`" + ` 出邊 | ` + "`SAVE→4BF2h`" + ` | 3D 地圖（` + "`4BC5h`" + `） | InDungeon（` + "`4BE6h`" + `） | ` + "`LOAD FILES`" + ` |
|---|---|---:|---|---|---|---|---|
`)
	edges, noExit := 0, []string{}
	for _, key := range keys {
		g := graphs[key]
		edges += len(g.newECL)
		if len(g.newECL) == 0 {
			noExit = append(noExit, key.member+"／"+key.block)
		}
		report.WriteString(fmt.Sprintf("| `%s` | `%s` | %d | %s | %s | %s | %s | %s |\n",
			key.member, key.block, g.instructions,
			formatTargets(g.newECL), formatTargets(g.lastECL),
			formatTargets(g.mapBlock), formatTargets(g.inDungeon), formatStrings(g.loadFiles)))
	}
	report.WriteString(segmentTable(keys, graphs))
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
	result := &blockGraph{payload: len(data), newECL: map[int]int{}, lastECL: map[int]int{},
		mapBlock: map[int]int{}, inDungeon: map[int]int{}, loadFiles: map[string]int{}}
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
			if len(operands) < 2 || !operands[1].WordSet || operands[0].WordSet {
				continue
			}
			switch operands[1].Word {
			case 0x4BF2: // bank0 +1E4h：LastECL
				result.lastECL[int(operands[0].Low)]++
			case 0x4BC5: // bank0 +18Ah：目前的 3D 地圖區塊
				result.mapBlock[int(operands[0].Low)]++
			case 0x4BE6: // bank0 +1CCh：InDungeon
				result.inDungeon[int(operands[0].Low)]++
			}
		case 0x21: // LOAD FILES
			if len(operands) >= 3 {
				result.loadFiles[fmt.Sprintf("%02X/%02X/%02X",
					operands[0].Low, operands[1].Low, operands[2].Low)]++
			}
		}
	}
	result.instructions = len(seen)
	return result
}

// segmentTable 把轉移圖與 game pack 的地圖宣告接起來：**`area_id` 就是 ECL
// 成員編號、`script_block` 就是 block 編號**，所以兩邊 join 得起來。
// 這張表是分段驗證的段落清單（`docs/plan/mainline-segmented-verification.md`）。
func segmentTable(keys []blockKey, graphs map[blockKey]*blockGraph) string {
	pack, err := gamepack.Default()
	if err != nil {
		return "\n（讀不到 game pack，略過段落清單）\n"
	}
	byBlock := map[[2]int][]string{}
	for _, m := range pack.Maps {
		if m.ScriptBlock == nil {
			continue
		}
		key := [2]int{int(m.AreaID), int(*m.ScriptBlock)}
		byBlock[key] = append(byBlock[key], m.ID)
	}
	incoming := map[int][]string{}
	for _, key := range keys {
		for target := range graphs[key].newECL {
			incoming[target] = append(incoming[target], key.block)
		}
	}
	var out strings.Builder
	labels := segmentLabels()
	out.WriteString("\n## 段落清單（block ↔ 地圖）\n\n" +
		"`area_id` 就是 ECL 成員編號、`script_block` 就是 block 編號，所以 game pack 的\n" +
		"地圖宣告與轉移圖 join 得起來。**沒有地圖的 block 不是缺漏**：世界地圖 hub 與\n" +
		"開場不需要 3D 地圖（`LOAD FILES` 是 `7F/7F/7F`），另外幾個沿用上一段的檔案。\n\n" +
		"段的 id 一律是 `ECL{成員}/0x{block}`（機械且穩定）；標籤取自\n" +
		"`docs/plan/segment-labels.json`，證據見 `docs/plan/seg-03-verification-report.md`。\n\n" +
		"| 段 | 標籤 | 進入自 | 離開到 | game pack 地圖 |\n|---|---|---|---|---|\n")
	for _, key := range keys {
		member := 0
		fmt.Sscanf(key.member, "ECL%d.DAX", &member)
		block := 0
		fmt.Sscanf(key.block, "0x%X", &block)
		names := byBlock[[2]int{member, block}]
		sort.Strings(names)
		label := "—"
		if len(names) > 0 {
			label = "`" + strings.Join(names, "`、`") + "`"
		}
		in := incoming[block]
		sort.Strings(in)
		inLabel := "—"
		if len(in) > 0 {
			inLabel = "`" + strings.Join(in, "`、`") + "`"
		}
		id := fmt.Sprintf("ECL%d/%s", member, key.block)
		out.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s |\n",
			id, labelOr(labels, id), inLabel, formatTargets(graphs[key].newECL), label))
	}
	return out.String()
}

// segmentLabels 讀人類可讀的段落標籤。讀不到就留白——標籤是給人看的，
// 缺了不影響這張表的機械內容。
func segmentLabels() map[string]string {
	payload, err := os.ReadFile("docs/plan/segment-labels.json")
	if err != nil {
		return nil
	}
	var parsed struct {
		Labels map[string]string `json:"labels"`
	}
	if json.Unmarshal(payload, &parsed) != nil {
		return nil
	}
	return parsed.Labels
}

func labelOr(labels map[string]string, id string) string {
	if label, ok := labels[id]; ok {
		return label
	}
	return "—"
}

func formatStrings(values map[string]int) string {
	if len(values) == 0 {
		return "—"
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, "`"+key+"`")
	}
	return strings.Join(parts, "、")
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
