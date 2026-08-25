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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
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
	// Page／PageRule／PageText 是這個空行落在哪一個**顯示頁**上
	// （`docs/audit/ecl-text-coverage.json`，由 `cmd/ecl-text-coverage` 產生）。
	Page     int
	PageRule string
	PageText string
	// InPage 為真 ＝ 前後兩段文字**都**出現在同一頁裡 ⇒ 玩家真的看得到那個留白。
	// 落在頁緣或被跳進來的那幾段兩側對不上同一頁，`InPage` 是 false。
	InPage bool
}

// coveragePage 是 `ecl-text-coverage.json` 的一頁。
type coveragePage struct {
	Member string `json:"member"`
	Block  string `json:"block"`
	Offset string `json:"offset"`
	Text   string `json:"text"`
	RuleID string `json:"rule_id"`
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	coverage := flag.String("coverage", "docs/audit/ecl-text-coverage.json",
		tooltext.Text("h.e387de426580"))
	jsonPath := flag.String("json", "", tooltext.Text("h.0572347d71ad"))
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
			log.Fatal(tooltext.Format("h.f74a43cb81d6", member))
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

	if pages, loadErr := loadCoveragePages(*coverage); loadErr == nil {
		attachPages(sites, pages)
	} else {
		fmt.Fprint(os.Stderr, tooltext.Format("h.8a966de19f86", *coverage, loadErr))
	}

	total, blank := 0, 0
	for _, item := range sites {
		total += item.Run
		if item.Run > 1 {
			blank++
		}
	}

	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.6407375837e2"))
	fmt.Fprint(&report, tooltext.Text("h.d1c752bfa04e")+
		tooltext.Text("h.e1595eec91b3"))
	fmt.Fprint(&report, tooltext.Format("h.e811720ed9a3"))
	fmt.Fprint(&report, tooltext.Format("h.97ca48347bd7", total))
	fmt.Fprint(&report, tooltext.Format("h.6afc3e516e8b", blank))
	fmt.Fprint(&report, tooltext.Format("h.cac23a8e78b2", len(sites)))

	if blank == 0 {
		fmt.Fprint(&report, tooltext.Text("h.cd3b87273d91")+
			tooltext.Text("h.b4c715133f8b")+
			tooltext.Text("h.9860b1e031d6"))
	} else {
		fmt.Fprint(&report, tooltext.Format("h.c9e115dc63df", blank))
		visible := 0
		for _, item := range sites {
			if item.Run > 1 && item.InPage {
				visible++
			}
		}
		fmt.Fprintf(&report, tooltext.Text("h.ec34dd756579")+
			tooltext.Text("h.08b822b88015"), visible)
		fmt.Fprint(&report, tooltext.Format("h.9818638c4453"))
		fmt.Fprintf(&report, "|---|---:|---:|---|---|---|---|\n")
		for _, item := range sites {
			if item.Run < 2 {
				continue
			}
			mark := "—"
			if item.InPage {
				mark = "✅"
			}
			fmt.Fprintf(&report, "| ECL%d/0x%02X | `%04Xh` | %d | %s | %s | %s | %s |\n",
				item.Member, item.Block, item.Offset, item.Run,
				label(item.Before), label(item.After), mark, label(item.PageRule))
		}
		report.WriteString(tooltext.Text("h.943ebc7e980c"))
		for _, item := range sites {
			if item.Run < 2 {
				continue
			}
			fmt.Fprint(&report, tooltext.Format("h.54d347782b96", item.Member, item.Block, item.Offset, label(item.BeforeText), label(item.AfterText)))
		}
		report.WriteString("\n")
		report.WriteString("\n")
	}
	fmt.Fprint(&report, tooltext.Text("h.7c454bdf0809")+
		tooltext.Text("h.f392252d307e"))

	if *jsonPath != "" {
		type visibleRun struct {
			Member int    `json:"member"`
			Block  uint8  `json:"block"`
			Offset int    `json:"offset"`
			Run    int    `json:"run"`
			Rule   string `json:"rule_id"`
			Before string `json:"before_text"`
			After  string `json:"after_text"`
		}
		payload := struct {
			Instructions int          `json:"instructions"`
			Runs         int          `json:"runs"`
			BlankRuns    int          `json:"blank_runs"`
			Visible      []visibleRun `json:"visible_blank_runs"`
		}{Instructions: total, Runs: len(sites), BlankRuns: blank}
		for _, item := range sites {
			if item.Run < 2 || !item.InPage {
				continue
			}
			payload.Visible = append(payload.Visible, visibleRun{
				Member: item.Member, Block: item.Block, Offset: item.Offset,
				Run: item.Run, Rule: item.PageRule,
				Before: item.BeforeText, After: item.AfterText,
			})
		}
		encoded, marshalErr := json.MarshalIndent(payload, "", "  ")
		if marshalErr != nil {
			log.Fatal(marshalErr)
		}
		if writeErr := os.WriteFile(*jsonPath, append(encoded, '\n'), 0o644); writeErr != nil {
			log.Fatal(writeErr)
		}
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "print-return=%d runs=%d blank-runs=%d\n", total, len(sites), blank)
}

// loadCoveragePages 讀 `cmd/ecl-text-coverage` 的 JSON。
func loadCoveragePages(path string) ([]coveragePage, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Groups []coveragePage `json:"groups"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Groups) == 0 {
		return nil, tooltext.Errorf("h.dbc19fa8aa18")
	}
	return parsed.Groups, nil
}

// attachPages 把每一段換行對到它所在的顯示頁。
//
// ⚠ 頁的界線由走訪決定，報表只記**起點**位移 ⇒ 這裡取「同一個 block 裡起點
// 不大於這一段、而且最靠近的那一頁」。這是推論，不是宣告，所以**還要驗**：
// `InPage` 要求前後兩段文字都出現在那一頁的文字裡，對不上就不算。
func attachPages(sites []site, pages []coveragePage) {
	byBlock := map[string][]coveragePage{}
	for _, page := range pages {
		key := fmt.Sprintf("%s|%s", page.Member, page.Block)
		byBlock[key] = append(byBlock[key], page)
	}
	for key := range byBlock {
		list := byBlock[key]
		sort.SliceStable(list, func(left, right int) bool {
			return pageOffset(list[left]) < pageOffset(list[right])
		})
		byBlock[key] = list
	}
	for index := range sites {
		item := &sites[index]
		key := fmt.Sprintf("ECL%d.DAX|0x%02X", item.Member, item.Block)
		best := -1
		for position, page := range byBlock[key] {
			if pageOffset(page) <= item.Offset {
				best = position
				continue
			}
			break
		}
		if best < 0 {
			continue
		}
		page := byBlock[key][best]
		item.Page = pageOffset(page)
		item.PageRule = page.RuleID
		item.PageText = page.Text
		item.InPage = item.BeforeText != "" && item.AfterText != "" &&
			strings.Contains(page.Text, item.BeforeText) &&
			strings.Contains(page.Text, item.AfterText)
	}
}

func pageOffset(page coveragePage) int {
	value, err := strconv.ParseInt(strings.TrimPrefix(
		strings.TrimPrefix(page.Offset, "0x"), "0X"), 16, 32)
	if err != nil {
		return -1
	}
	return int(value)
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
