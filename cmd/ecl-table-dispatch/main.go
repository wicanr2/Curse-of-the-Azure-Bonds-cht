// Command ecl-table-dispatch 盤點 ECL 的**查表分派**：`25h ON GOTO`／`26h ON GOSUB`
// 與 `2Ah GETTABLE`。
//
// ★ 存在的理由：`docs/audit/remake-status.md` 把「全城市／全房間走訪」列為沒有
// 分母的項目，而它點名的第一個缺口就是「世界地圖那種 `GETTABLE` ＋ `ON GOTO`
// 的查表分派」。以地形碼分派的每格事件早就有分母了，查表分派沒有。
//
// 量什麼：分派點幾處、分支目標幾個、相異的目標位址幾個。**相異目標**才是真正的
// 分母——同一個目標被兩個索引指到，只是一條路。
//
// ⚠ 這是**分母**，不是覆蓋率。要知道「玩家實際走到幾條」得有實跑資料；
// 這一支不宣稱那件事。把分支數當成「已完成幾條」會把沒跑過的路算成跑過。
//
// ⚠ 只數靜態追得到的分派點。`TraceGraph` 到不了的程式碼裡若還有分派，這裡看不到
// ——所以這個分母是**下界**。
//
// 用法：
//
//	go run ./cmd/ecl-table-dispatch -output docs/audit/ecl-table-dispatch.md
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

type blockStat struct {
	member       string
	block        uint8
	onSites      int
	onTargets    int
	distinct     int
	getTableSite int
}

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "原版 image ZIP")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	archive, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	stats := make([]blockStat, 0, 64)
	for member := 1; member <= 6; member++ {
		name := fmt.Sprintf("ECL%d.DAX", member)
		payload := zipMember(&archive.Reader, name)
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatalf("%s：%v", name, parseErr)
		}
		for _, block := range blocks {
			if stat, ok := scan(name, block.Entry.ID, block.Data); ok {
				stats = append(stats, stat)
			}
		}
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].member != stats[j].member {
			return stats[i].member < stats[j].member
		}
		return stats[i].block < stats[j].block
	})

	totalOn, totalTargets, totalDistinct, totalGet := 0, 0, 0, 0
	for _, stat := range stats {
		totalOn += stat.onSites
		totalTargets += stat.onTargets
		totalDistinct += stat.distinct
		totalGet += stat.getTableSite
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# ECL 的查表分派：`ON GOTO`／`ON GOSUB` 與 `GETTABLE`\n\n")
	fmt.Fprintf(&report, "由 `cmd/ecl-table-dispatch` 產生，不要手改。\n\n")
	fmt.Fprintf(&report, "⚠ 這是**分母**，不是覆蓋率。要知道玩家實際走到幾條分支，"+
		"得有實跑資料——這一支不宣稱那件事。⚠ 只數靜態追得到的分派點，所以是**下界**。\n\n")
	fmt.Fprintf(&report, "| 總計 | 數量 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 有查表分派的 ECL 段 | %d |\n", len(stats))
	fmt.Fprintf(&report, "| `25h`／`26h` 分派點 | %d |\n", totalOn)
	fmt.Fprintf(&report, "| 分支目標（含重複）| %d |\n", totalTargets)
	fmt.Fprintf(&report, "| **相異的分支目標** | **%d** |\n", totalDistinct)
	fmt.Fprintf(&report, "| `2Ah GETTABLE` 取值點 | %d |\n\n", totalGet)

	fmt.Fprintf(&report, "| ECL 檔 | 段 | 分派點 | 目標 | 相異目標 | `GETTABLE` |\n")
	fmt.Fprintf(&report, "|---|---:|---:|---:|---:|---:|\n")
	for _, stat := range stats {
		fmt.Fprintf(&report, "| `%s` | `0x%02X` | %d | %d | %d | %d |\n",
			stat.member, stat.block, stat.onSites, stat.onTargets, stat.distinct, stat.getTableSite)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "blocks=%d on-sites=%d targets=%d distinct=%d gettable=%d\n",
		len(stats), totalOn, totalTargets, totalDistinct, totalGet)
}

func scan(member string, block uint8, data []byte) (blockStat, bool) {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return blockStat{}, false
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return blockStat{}, false
	}
	stat := blockStat{member: member, block: block}
	distinct := map[int]bool{}
	seen := map[int]bool{}
	for _, instruction := range graph.Instructions {
		if seen[instruction.Offset] {
			continue
		}
		seen[instruction.Offset] = true
		switch instruction.Command.Opcode {
		case 0x2A:
			stat.getTableSite++
		case 0x25, 0x26:
			stat.onSites++
			// 目標清單接在兩個 head 運算元之後；`TraceGraph` 已經把可達的邊
			// 走過一遍，這裡直接數這一條指令的邊。
			for _, edge := range graph.Edges {
				if edge.From == instruction.Offset {
					stat.onTargets++
					distinct[edge.To] = true
				}
			}
		}
	}
	stat.distinct = len(distinct)
	if stat.onSites == 0 && stat.getTableSite == 0 {
		return blockStat{}, false
	}
	return stat, true
}

func zipMember(archive *zip.Reader, name string) []byte {
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
