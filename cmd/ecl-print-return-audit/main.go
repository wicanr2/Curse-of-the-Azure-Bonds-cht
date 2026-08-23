// Command ecl-print-return-audit 回答「原作的硬換行落在哪裡，其中幾處會空行」。
//
// ★ 存在的理由：`33h PRINT RETURN` 還是 `partial`，理由寫的是「remake 沒有游標
// 模型」。但**缺口有多大**一直沒有數字——而那個數字決定它是要補的表現層工作，
// 還是可以標成「玩家看不出來」。原作的語意是硬換行（`65A0h := 1`、
// `inc 65A1h`，spec 1147），所以：
//
//	單獨一條          → 換行，接在文字後面就是一般的斷行
//	連續兩條以上      → **空行**（多推一列而那一列沒有東西）
//
// 這一支只數**走得到的**指令：走訪與 `cmd/ecl-cell-refs` 同一條路（五個生命
// 週期入口 ＋ 跟跳躍）。
//
// ⚠ 「連續」是指**執行序上相鄰**，不是位元組相鄰：中間夾著 `GOTO` 的兩條
// `33h` 在某條路徑上仍然可能連著跑。這一支只認**位元組相鄰**（前一條指令的
// 結束位移就是這一條的開始），所以是**下界**。
//
// 用法：
//
//	go run ./cmd/ecl-print-return-audit
//	go run ./cmd/ecl-print-return-audit -output docs/audit/ecl-print-return.md
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

const printReturnOpcode = 0x33

type site struct {
	Member int
	Block  uint8
	Offset int
	// Run 是這一處起連著幾條 `33h`。1 是單純換行，2 以上會空行。
	Run int
	// Before／After 是**位元組上**前後那一條指令的助憶碼。空行看不看得見要看
	// 它們：夾在兩段文字中間才是玩家看得到的留白。
	//
	// ⚠ `Before` 不一定是執行序上的前一條——這一對 `33h` 可能是被跳進來的
	// （`ECL2/0x02:0CD4h` 的前一條是 `EXIT` 就是這種）。`After` 是落下去的
	// 下一條，那個方向可靠。
	Before string
	After  string
	// BeforeText／AfterText 是位元組上最近的一段 `PRINT`／`PRINTCLEAR` 文字。
	// 有了它才知道空行該加在譯文的哪一句之間。
	BeforeText string
	AfterText  string
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	sites := make([]site, 0, 64)
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			log.Fatalf("image 裡沒有 ECL%d.DAX", member)
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		for _, block := range blocks {
			found, scanErr := scanBlock(member, block.Entry.ID, block.Data)
			if scanErr != nil {
				log.Fatalf("ECL%d/0x%02X: %v", member, block.Entry.ID, scanErr)
			}
			sites = append(sites, found...)
		}
	}

	total, blank := 0, 0
	for _, item := range sites {
		total += item.Run
		if item.Run > 1 {
			blank++
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# 原作的硬換行落在哪裡（`33h PRINT RETURN`）\n\n")
	fmt.Fprintf(&report, "由 `cmd/ecl-print-return-audit` 產生，不要手改。"+
		"語意見 spec 1147：`65A0h := 1`、`inc 65A1h` ⇒ 欄歸位、列前進。\n\n")
	fmt.Fprintf(&report, "| 項目 | 數 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 走得到的 `33h` 指令 | %d |\n", total)
	fmt.Fprintf(&report, "| 其中連著兩條以上的段落（**會空行**）| %d |\n", blank)
	fmt.Fprintf(&report, "| 換行段落合計 | %d |\n\n", len(sites))

	if blank == 0 {
		fmt.Fprintf(&report, "⇒ **走得到的碼裡沒有任何一處連續換行**，"+
			"所以「連續 `33h` 會空行」在原作的劇情文字上碰不到；"+
			"remake 缺游標模型看不出差別。\n\n")
	} else {
		fmt.Fprintf(&report, "⇒ 有 %d 個段落會空行。那幾處 remake 目前會把空行擠掉。\n\n", blank)
		fmt.Fprintf(&report, "| ECL/區塊 | 位移 | 連續幾條 | 位元組上的前一條 | 後一條 |\n|---|---:|---:|---|---|\n")
		for _, item := range sites {
			if item.Run < 2 {
				continue
			}
			fmt.Fprintf(&report, "| ECL%d/0x%02X | `%04Xh` | %d | %s | %s |\n",
				item.Member, item.Block, item.Offset, item.Run,
				label(item.Before), label(item.After))
		}
		report.WriteString("\n空行落在哪兩句之間（位元組上最近的兩段文字）：\n\n")
		for _, item := range sites {
			if item.Run < 2 {
				continue
			}
			fmt.Fprintf(&report, "- **ECL%d/0x%02X `%04Xh`**\n  - 前：%s\n  - 後：%s\n",
				item.Member, item.Block, item.Offset,
				label(item.BeforeText), label(item.AfterText))
		}
		report.WriteString("\n")
		report.WriteString("\n")
	}
	fmt.Fprintf(&report, "⚠ 走訪跟不到的碼不在分母裡，而且「連續」只認**位元組相鄰**"+
		"（中間夾著跳躍、執行序上仍相鄰的算不到）⇒ 兩個方向都是**下界**。\n")

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "print-return=%d runs=%d blank-runs=%d\n", total, len(sites), blank)
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
	// 同一條指令可能被走訪碰到很多次；用位移收斂成一份。
	unique := map[int]ecl.Instruction{}
	for _, instruction := range graph.Instructions {
		unique[instruction.Offset] = instruction
	}
	offsets := make([]int, 0, len(unique))
	for offset := range unique {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)

	sites := make([]site, 0, 4)
	index := 0
	for index < len(offsets) {
		offset := offsets[index]
		if unique[offset].Command.Opcode != printReturnOpcode {
			index++
			continue
		}
		run := 1
		next := index + 1
		// `33h` 沒有運算元，所以位元組長度是 1；相鄰就是「執行序上一定連著」。
		for next < len(offsets) && offsets[next] == offsets[next-1]+1 &&
			unique[offsets[next]].Command.Opcode == printReturnOpcode {
			run++
			next++
		}
		item := site{Member: member, Block: id, Offset: offset, Run: run}
		if index > 0 {
			item.Before = unique[offsets[index-1]].Command.Name
			for back := index - 1; back >= 0; back-- {
				if text, ok := printedText(unique[offsets[back]]); ok {
					item.BeforeText = text
					break
				}
			}
		}
		if next < len(offsets) {
			item.After = unique[offsets[next]].Command.Name
			for ahead := next; ahead < len(offsets); ahead++ {
				if text, ok := printedText(unique[offsets[ahead]]); ok {
					item.AfterText = text
					break
				}
			}
		}
		sites = append(sites, item)
		index = next
	}
	return sites, nil
}

// printedText 取出 `PRINT`／`PRINTCLEAR` 的字面文字。運算元不是打包字串
// （變數插入、數值）時回 false——那種內容要看執行時的值，不是這一支能決定的。
func printedText(instruction ecl.Instruction) (string, bool) {
	switch instruction.Command.Opcode {
	case 0x11, 0x12:
	default:
		return "", false
	}
	if len(instruction.Operands) != 1 || instruction.Operands[0].Code != 0x80 {
		return "", false
	}
	return ecl.DecodePackedText(instruction.Operands[0].Packed), true
}

// label 讓「走訪到這裡就沒有下一條了」與「下一條叫空字串」分得開。
func label(name string) string {
	if name == "" {
		return "—"
	}
	return name
}

func memberPayload(archive *zip.ReadCloser, name string) []byte {
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
