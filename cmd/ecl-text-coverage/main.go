// Command ecl-text-coverage answers one question per ECL block: how many
// player-visible text groups does the original have, and how many of them does
// the CoAB game pack already resolve?
//
// 內容產出（`ENG-01`）先前沒有分母。「艾森布拉只有 6 條 text_rules」講的是
// remake 這側寫了幾條，不是原作有幾條，所以看不出還剩多少。這支把分母算出來。
//
// 群組化與 runtime 一致：`PRINTCLEAR`（`12h`）清框後開始一段，後續的 `PRINT`
// （`11h`）接在同一段裡；引擎的 `MatchText` 也是把累積的文字用空白接起來再比對
// （spec 1104 §一：兩者共用同一支 handler，靠重讀 opcode 分流）。
//
// 用法：
//
//	go run ./cmd/ecl-text-coverage -output docs/audit/ecl-text-coverage.md
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
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

var members = []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"}

type group struct {
	Member string `json:"member"`
	Block  string `json:"block"`
	Offset string `json:"offset"`
	Text   string `json:"text"`
	RuleID string `json:"rule_id,omitempty"`
	// GosubInserts 記錄「這一頁的文字被一支子程式插了一段」的位置——可能在中間，
	// 也可能整頁就是從子程式的 `PRINTCLEAR` 開始（那時它提供的是**前綴**）。
	// 子程式印的字本工具不追（見 Limitations），所以 `Text` 少了那一段；
	// 寫 `all_contains` 的片段**不可以跨過這些位置**，否則實機接不上。
	GosubInserts []string `json:"gosub_inserts,omitempty"`
	// DynamicBranch 表示這一頁所在的 run 有 `ON GOTO`／`ON GOSUB`。分支的文字會接
	// 在同一個 run 後面（街頭傳聞就是這樣），但目的地是動態的，本工具跟不到，
	// 所以拿不到完整的 run 文字。
	DynamicBranch bool `json:"dynamic_branch,omitempty"`
	// VariableInserts 記錄這一頁印過**執行期的值**（角色名、城名、數字）。
	// 那些值只有跑起來才知道，靜態文字裡是空的，所以工具比對不了——
	// 逐城的 `*.edge` 規則需要城名，就是這一類。
	VariableInserts []string `json:"variable_inserts,omitempty"`
	// BranchTail 表示這一段是「被 IF 守衛的 GOTO」跳進來的另一條續寫（`WEST.`／
	// `EAST.`），實機會把它接在同一頁後面，不會單獨出現。
	BranchTail bool `json:"branch_tail,omitempty"`
	// Status 是這一段在內容產出裡的處置：
	//
	//	matched      有規則命中，做完了
	//	unmatched    還沒寫規則——這才是待辦
	//	variable-insert 頁裡印了執行期的值（人名、城名），靜態文字裡沒有，比對不了
	//	subroutine   共用子程式的片段，實機一定被併進呼叫端那一頁
	//
	// `gosub_inserts`／`dynamic_branch` 只是**註記**，不是 status：那兩種插入
	// 已經在 runVariants 展開，展開後沒命中就照樣算 unmatched。
	Status string `json:"status"`
	// RunText 是「這一頁所在的某一份 run 的完整文字」，只在沒接上時填。
	// 寫規則要看的是它，不是單獨一頁——`all_contains` 比對的就是這一份。
	RunText string `json:"run_text,omitempty"`
	// Run 是 offset 線性切法留下的頁序號，只用來排序與比對舊報告。
	Run int `json:"run,omitempty"`
	// appended 記錄這一段是以 `11h PRINT` 起頭（true）還是 `12h PRINTCLEAR`（false）。
	// 只有前者可能是「接在呼叫端那一頁後面」的片段——它自己不開新頁。
	appended bool
	// fragment 表示這一段落在「會被插進別頁的子程式」裡，不管它用什麼起頭。
	fragment bool
	// gosubTargets 是這一頁插入點跳去的 payload offset，展開時要進去取文字。
	gosubTargets []int
}

type report struct {
	Schema      string   `json:"schema"`
	Limitations []string `json:"limitations"`
	Groups      []group  `json:"groups"`
	// OrphanRuns 是**一整條 run 沒有任何規則命中**的路徑。
	//
	// ★ 為什麼要另外算一個數。 `Groups` 的處置是**逐頁**判的，而一頁只要曾經
	// 被某一條命中的 run 經過就算 `matched`。問題是玩家走的是**一條** run，
	// 不是所有 run 的聯集：同一頁在 A 路徑上和兄弟句一起印（規則命中），
	// 在 B 路徑上單獨印（沒有規則）——逐頁的算法只看得到 A，`unmatched` 是 0，
	// 而實機走 B 的玩家看到英文。
	//
	// 這不是假想的失誤。`wizard-tower.efreet-band` 與
	// `tilverton.sewers.spoils-and-master-alone` 兩條規則是實機走出來才發現要補的；
	// 把它們從 pack 拿掉重跑，`unmatched` **仍然是 0**——逐頁的算法對這一類
	// 缺陷是全盲的。`TestOrphanRunAccountingCatchesTheAloneSiblingBug` 把這件事釘住。
	OrphanRuns []orphanRun `json:"orphan_runs,omitempty"`
	Summary    summary     `json:"summary"`
}

// orphanRun 是一條沒有任何規則命中的執行路徑。
type orphanRun struct {
	Member string `json:"member"`
	Block  string `json:"block"`
	// Pages 是這條 run 依序經過的頁（payload offset）。
	Pages []string `json:"pages"`
	// Text 是交給 `MatchText` 的那一份（各頁以空白相接）。
	Text string `json:"text"`
	// Status：`orphan` 才是待辦；其餘是與逐頁同一套的豁免。
	Status string `json:"status"`
}

type summary struct {
	Groups    int `json:"groups"`
	Matched   int `json:"matched"`
	Unmatched int `json:"unmatched"`
	// GosubInsert／BranchInsert 是**註記的計數**（有多少頁帶著插入點），
	// 不是狀態桶。它們與 Matched／Unmatched 有重疊，加起來不等於 Groups。
	GosubInsert    int            `json:"gosub_insert_notes"`
	VariableInsert int            `json:"variable_insert"`
	BranchInsert   int            `json:"branch_insert_notes"`
	BranchTail     int            `json:"branch_tail"`
	Subroutine     int            `json:"subroutine"`
	ByBlock        map[string]int `json:"unmatched_by_block"`

	// 下面四個是 **run** 為單位的帳（見 `report.OrphanRuns`）。文字相同的 run
	// 只算一次：走訪器會在不同位址上吐出同一份文字，那是同一段玩家體驗。
	Runs               int            `json:"runs"`
	RunsMatched        int            `json:"runs_matched"`
	RunsOrphan         int            `json:"runs_orphan"`
	RunsVariableInsert int            `json:"runs_variable_insert"`
	RunsSubroutine     int            `json:"runs_subroutine"`
	RunsGosubInsert    int            `json:"runs_gosub_insert"`
	RunsTruncated      int            `json:"runs_truncated"`
	OrphanByBlock      map[string]int `json:"orphan_runs_by_block"`
}

// truncatedRun 判一條 orphan run 是不是走訪器切出來的半截，而不是玩家真的
// 看得到的一頁。兩個形狀：
//
//	停在句子中間   原作只在一句講完之後才把文字交出去（選單／戰鬥／寶物／輸入）
//	是別條的前綴   同一條路徑的截斷版，完整的那條才是實機會演的
//
// ⚠ 收尾的引號要看**前一個字元**：`HE CONTINUES, '` 收在引號上，但引號前是逗號，
// 那是「話還沒講完」不是句子結束。少了這一關會有半句混進待辦清單。
func truncatedRun(text string, all map[string]bool) bool {
	for other := range all {
		if other != text && strings.HasPrefix(other, text) {
			return true
		}
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return true
	}
	last := runes[len(runes)-1]
	if last == '\'' || last == '"' {
		if len(runes) < 2 {
			return true
		}
		last = runes[len(runes)-2]
	}
	return last != '.' && last != '!' && last != '?'
}

// markdownCell 把一份 run 文字放進表格：跳脫直線、截到看得完的長度。
func markdownCell(text string) string {
	cell := strings.ReplaceAll(strings.TrimSpace(text), "|", "\\|")
	if len(cell) > 200 {
		cell = cell[:200] + "…"
	}
	return cell
}

// classifyRun 判一條**沒命中**的 run 是待辦還是豁免。豁免用的是與逐頁同一套
// 判準，只是提升到 run；順序有意義，前面的先套。
//
//	variable-insert  run 裡有任何一頁印執行期的值 → 靜態比對不到，不是缺陷
//	subroutine       整條 run 就是一頁共用子程式片段 → 實機會被併進呼叫端那一頁
//	gosub-insert     run 裡有插入點 → 手上這份文字**不完整**，不能當缺陷
//	orphan           控制流走得到、文字完整、卻沒有任何規則命中 —— 這才是待辦
func classifyRun(pages []string, notes map[string]group, subroutines, aloneOnly map[string]bool) string {
	switch {
	case anyPageHas(notes, pages, func(item group) bool { return len(item.VariableInserts) > 0 }):
		return "variable-insert"
	case len(pages) == 1 && subroutines[pages[0]] && aloneOnly[pages[0]]:
		return "subroutine"
	case anyPageHas(notes, pages, func(item group) bool { return len(item.GosubInserts) > 0 }):
		return "gosub-insert"
	}
	return "orphan"
}

// anyPageHas 回答「這條 run 經過的頁裡有沒有符合條件的」。
func anyPageHas(notes map[string]group, pages []string, want func(group) bool) bool {
	for _, page := range pages {
		if want(notes[page]) {
			return true
		}
	}
	return false
}

func main() {
	archive := flag.String("archive", "curseoftheazurebonds.zip", "original CoAB archive")
	output := flag.String("output", "", "write the Markdown report to this path")
	outputJSON := flag.String("json", "", "write the machine-readable report to this path")
	block := flag.String("block", "", "limit to one block, e.g. ECL3.DAX/0x10")
	debugPages := flag.Bool("debug-pages", false, tooltext.Text("h.4d3b7a0f753f"))
	flag.Parse()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	reader, err := zip.OpenReader(*archive)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	result := report{
		Schema: "coab-ecl-text-coverage/1",
		Limitations: []string{
			tooltext.Text("h.fbe078e1ab27"),
			tooltext.Text("h.fa64c409b7f0"),
			tooltext.Text("h.ad4976f51b6f"),
			tooltext.Text("h.f20e4cb5d793"),
			tooltext.Text("h.811b1520eb10"),
			tooltext.Text("h.d6be44be96a5"),
			tooltext.Text("h.d703b4e3c535"),
			tooltext.Text("h.450bca06483d"),
			tooltext.Text("h.b39186829581"),
		},
		Summary: summary{ByBlock: map[string]int{}, OrphanByBlock: map[string]int{}},
	}

	for _, member := range members {
		data, err := readMember(&reader.Reader, member)
		if err != nil {
			log.Fatal(err)
		}
		blocks, err := dax.Parse(data)
		if err != nil {
			log.Fatalf("%s: %v", member, err)
		}
		for _, raw := range blocks {
			blockID := fmt.Sprintf("0x%02X", raw.Entry.ID)
			if *block != "" && *block != member+"/"+blockID {
				continue
			}
			annotations, err := blockGroups(member, blockID, raw.Data)
			if err != nil {
				log.Fatalf("%s block %s: %v", member, blockID, err)
			}
			// 分母與比對分開走（`runs.go` 的 walkPages／walkRuns）：
			// 前者不帶文字，狀態小到一定走得完，所以「有哪些頁」是完整的；
			// 後者帶文字，碰到上限只會**少判**，不會讓頁從分母裡消失。
			pageText, err := walkPages(raw.Data)
			if err != nil {
				log.Fatalf("%s block %s page walk: %v", member, blockID, err)
			}
			paths, runPageText, err := walkRuns(member+"/"+blockID, raw.Data)
			if err != nil {
				log.Fatalf("%s block %s walk: %v", member, blockID, err)
			}
			// 分母走訪只記「開頁那一條」的文字（它不累積），完整的一頁要用帶文字
			// 的走訪那一份。兩者都有時取後者，只有它沒走到的頁才退回前者。
			for offset, text := range runPageText {
				if len(text) > len(pageText[offset]) {
					pageText[offset] = text
				}
			}
			// 一條 run 命中，它經過的每一頁就都接上了：runtime 是把整份文字
			// **一次**交給 MatchText，規則命中的是那一份，不是其中某一頁。
			covered := map[string]string{}
			runOf := map[string]string{}
			// aloneOnly 記著「這一頁只曾經自己成為一整份 run」。共用子程式
			// （`WHAT DO YOU DO?`、`UP`／`DOWN`）在實機一定被併進呼叫端那一頁，
			// 走訪器卻會從某些路徑上單獨吐出它們。
			aloneOnly := map[string]bool{}
			// runSeen 讓「文字相同的 run」只算一次：走訪器會在不同位址上吐出
			// 同一份文字（同一段玩家體驗），逐條計數會讓分母被分支數放大。
			runSeen := map[string]bool{}
			type runRecord struct {
				pages   []string
				matched bool
			}
			runs := make([]runRecord, 0, len(paths))
			runText := make([]string, 0, len(paths))
			for _, path := range paths {
				joined := strings.Join(path.Texts, " ")
				matched := pack.MatchText(path.Texts, pack.DefaultLocale)
				if strings.TrimSpace(joined) != "" && !runSeen[joined] {
					runSeen[joined] = true
					runText = append(runText, joined)
					runs = append(runs, runRecord{pages: path.Pages, matched: matched.Matched})
				}
				for _, page := range path.Pages {
					if len(path.Pages) > 1 {
						aloneOnly[page] = false
					} else if _, seen := aloneOnly[page]; !seen {
						aloneOnly[page] = true
					}
					if matched.Matched {
						if _, done := covered[page]; !done {
							covered[page] = matched.RuleID
						}
						continue
					}
					// 沒命中的那一份留一個代表，寫規則時要看的是整份 run 文字，
					// 不是單獨一頁——片段跨不過去的地方就在這裡看得出來。
					if existing, ok := runOf[page]; !ok || len(joined) > len(existing) {
						runOf[page] = joined
					}
				}
			}
			if *debugPages {
				for _, item := range annotations.groups {
					if _, ok := pageText[item.Offset]; !ok {
						fmt.Fprintf(os.Stderr, "unreached %s/%s %s | %.70s\n",
							member, blockID, item.Offset, item.Text)
					}
				}
			}
			notes := map[string]group{}
			for _, item := range annotations.groups {
				notes[item.Offset] = item
			}
			// ── run 級的帳 ───────────────────────────────────────────────
			// 豁免用的是與逐頁同一套判準，只是提升到 run：
			//
			//	variable-insert  run 裡有任何一頁印執行期的值 → 靜態比對不到
			//	subroutine       整條 run 就是一頁共用子程式片段 → 實機會被併進呼叫端
			//	gosub-insert     run 裡有插入點 → 這份文字**不完整**，不能當缺陷
			//
			// 剩下的才是 `orphan`：控制流走得到、文字完整、卻沒有任何規則命中。
			for index, record := range runs {
				result.Summary.Runs++
				if record.matched {
					result.Summary.RunsMatched++
					continue
				}
				status := classifyRun(record.pages, notes, annotations.subroutines, aloneOnly)
				switch status {
				case "variable-insert":
					result.Summary.RunsVariableInsert++
				case "subroutine":
					result.Summary.RunsSubroutine++
				case "gosub-insert":
					result.Summary.RunsGosubInsert++
				default:
					result.Summary.RunsOrphan++
					result.Summary.OrphanByBlock[member+"/"+blockID]++
				}
				if status != "orphan" {
					continue
				}
				result.OrphanRuns = append(result.OrphanRuns, orphanRun{
					Member: member, Block: blockID, Pages: record.pages,
					Text: runText[index], Status: status,
				})
			}
			offsets := make([]string, 0, len(pageText))
			for offset := range pageText {
				offsets = append(offsets, offset)
			}
			sort.Strings(offsets)
			for _, offset := range offsets {
				item := group{
					Member: member, Block: blockID, Offset: offset,
					Text:            pageText[offset],
					GosubInserts:    notes[offset].GosubInserts,
					VariableInserts: notes[offset].VariableInserts,
					DynamicBranch:   notes[offset].DynamicBranch,
				}
				switch rule, matched := covered[offset]; {
				case matched:
					item.RuleID = rule
					item.Status = "matched"
					result.Summary.Matched++
				// ⚠ 只有這一桶是真的「比對不了」：頁裡印的是執行期的值（城名、
				// 人名、數字），靜態文字裡根本沒有那幾個字，規則要靠它才會命中。
				// 子程式插入與 `ON GOTO` 分支都已經由 walkRuns 走過，走完還是
				// 沒命中就是**沒接上**——待辦躲進看起來已處理的桶子，
				// 就是誤判做完的那個形狀。
				case len(item.VariableInserts) > 0:
					item.Status = "variable-insert"
					result.Summary.VariableInsert++
				// 這一頁落在某支被 `GOSUB` 呼叫的子程式裡，而且**從來沒有和別的頁
				// 同屬一份 run**——也就是走訪器只在「從半路進入子程式」那種路徑上
				// 單獨吐出過它。實機它一定是被併進呼叫端那一頁的（`THE STAIRS
				// LEAD ⟨UP｜DOWN⟩ HERE.`），所以既不算做完也不算待辦。
				//
				// ⚠ 兩個條件缺一不可。只看「在子程式裡」會把 `THE DOOR IS LOCKED`
				// 這種共用的**整頁**也吞掉；只看「單獨成 run」會把真正的單句事件
				// 吞掉。**寧可少標也不要多標**：少標只是清單裡多幾筆看起來奇怪的
				// 短句，多標會讓真正的待辦從清單上消失。
				case annotations.subroutines[offset] && aloneOnly[offset]:
					item.Status = "subroutine"
					result.Summary.Subroutine++
				default:
					item.Status = "unmatched"
					item.RunText = runOf[offset]
					result.Summary.Unmatched++
					result.Summary.ByBlock[member+"/"+blockID]++
				}
				if len(item.GosubInserts) > 0 {
					result.Summary.GosubInsert++
				}
				if item.DynamicBranch {
					result.Summary.BranchInsert++
				}
				result.Summary.Groups++
				result.Groups = append(result.Groups, item)
			}
		}
	}

	// ── orphan 再分一次：把走訪器切出來的半句挑掉 ──────────────────────
	//
	// ★ 為什麼可以這樣挑。 原作把累積的文字交給 `MatchText` 的時機是**選單、
	// 戰鬥、寶物、輸入、換 block、離開**——那些點一定落在一句話講完之後。
	// 所以「停在句子中間」的 run 不是玩家看得到的一頁，是走訪器在 `IF` 分岔上
	// 切出來的半截。同理，另一條 run 的**嚴格前綴**也是同一條路徑的截斷版。
	//
	// ⚠ 只降級不刪除：這些 run 仍留在 `orphan_runs` 裡並標成 `truncated`，
	// 需要時看得到。真的要盯的是 `runs_orphan`。
	{
		full := map[string]bool{}
		for _, item := range result.OrphanRuns {
			full[strings.TrimSpace(item.Text)] = true
		}
		kept := result.OrphanRuns[:0]
		for _, item := range result.OrphanRuns {
			text := strings.TrimSpace(item.Text)
			if truncatedRun(text, full) {
				item.Status = "truncated"
				result.Summary.RunsTruncated++
				result.Summary.RunsOrphan--
				result.Summary.OrphanByBlock[item.Member+"/"+item.Block]--
				if result.Summary.OrphanByBlock[item.Member+"/"+item.Block] == 0 {
					delete(result.Summary.OrphanByBlock, item.Member+"/"+item.Block)
				}
			}
			kept = append(kept, item)
		}
		result.OrphanRuns = kept
	}

	if *outputJSON != "" {
		encoded, err := json.MarshalIndent(result, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*outputJSON, append(encoded, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	if *output != "" {
		if err := os.WriteFile(*output, renderMarkdown(result), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprint(os.Stderr, tooltext.Format("h.7a34e313b90d", result.Summary.Groups, result.Summary.Matched, result.Summary.Unmatched, result.Summary.VariableInsert, result.Summary.Subroutine, result.Summary.GosubInsert, result.Summary.BranchInsert))
}

// blockGroups collects the reachable PRINTCLEAR/PRINT runs of one block, plus
// the set of group offsets that live inside a GOSUB'd subroutine.
func blockGroups(member, blockID string, data []byte) (blockText, error) {
	if len(data) < 2 {
		return blockText{}, nil
	}
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return blockText{}, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return blockText{}, err
	}
	instructions := append([]ecl.Instruction(nil), graph.Instructions...)
	sort.Slice(instructions, func(i, j int) bool {
		return instructions[i].Offset < instructions[j].Offset
	})

	var groups []group
	var current *group
	var pendingGosub []string
	guarded := false
	// branching 記著哪些 run 裡有 `ON GOTO`／`ON GOSUB`：那些分支的文字接在同一個
	// run 後面，但目的地是動態的，本工具跟不到。
	branching := map[int]bool{}
	// branchTargets 記著每個 run 的 `ON GOTO`／`ON GOSUB` 會跳到哪些 payload
	// offset。展開比對時要把那邊的文字接到這個 run 後面。
	branchTargets := map[int][]int{}
	// branchRuns 反過來記：這個 run 是從哪一個分支目的地開始的。展開時要把
	// 跳進來之前累積的文字接在前面。
	branchRuns := map[int]int{}
	// run 只在「交出文字」時才換：`PRINTCLEAR` 開新頁但**不**結束 run。
	// 頁尾的 `GOSUB`（翻頁提示、Yes／No）會讓 VM 停下等玩家，那才是 run 邊界；
	// 頁中的 `GOSUB`（插一段文字進來）不是。兩者的差別在於後面還有沒有 PRINT。
	run := 0
	// cleared 記著「剛剛清過框」。沒有文字的 PRINTCLEAR 不會開出段落，但它確實
	// 開了新頁，下一條 PRINT 因此不是接在別人後面的片段。
	cleared := false
	for _, instruction := range instructions {
		opcode := instruction.Command.Opcode
		wasGuarded := guarded
		guarded = opcode >= 0x16 && opcode <= 0x1B
		// 被 `IF` 守衛的指令，條件不成立時整條被跳過（spec 1106 §一），
		// 所以下一條在 offset 順序上仍然是同一頁的延續。
		if !wasGuarded && endsTextGroup(opcode) {
			current, pendingGosub = nil, nil
			// ⚠ 只有「交出文字」才換 run。`GOTO`／`RETURN`／`ON GOTO` 是控制流，
			// 它們讓**頁**在 offset 順序上斷開，但不會把累積的文字交給 MatchText。
			if endsRun(opcode) {
				run++
			}
			if opcode == 0x25 || opcode == 0x26 {
				// 選哪一條是動態的，但**目的地是靜態的字面位址**。那邊印的文字
				// 接在同一個 run 後面，所以要把每個目的地都記下來展開比對。
				branching[run] = true
				targets, _, err := ecl.BranchTargets(data, instruction.Offset)
				if err != nil {
					return blockText{}, err
				}
				branchTargets[run] = append(branchTargets[run], targets...)
			}
			continue
		}
		if opcode == 0x02 {
			// 先記著。只有後面還有 `11h PRINT` 接上來，它才算是插入點；收尾的那個
			// GOSUB（翻頁提示、Yes/No）不算，`12h PRINTCLEAR` 開新頁也把它作廢。
			pendingGosub = append(pendingGosub, fmt.Sprintf("0x%04X", instruction.Offset))
			continue
		}
		if opcode != 0x11 && opcode != 0x12 {
			// 中間夾 SAVE、COMPARE、PICTURE、DELAY、IF 這類指令不會斷開一頁文字。
			continue
		}
		variables := instructionVariables(instruction)
		if len(variables) > 0 && current != nil {
			// 印執行期的值那一條沒有靜態文字，會走下面的 `text == ""`，
			// 所以要先記起來——不然「城名」這一類插入永遠偵測不到。
			current.VariableInserts = append(current.VariableInserts, variables...)
		}
		text := instructionText(instruction)
		if text == "" {
			// 沒有文字的 `PRINTCLEAR` 照樣清框開新頁（原作用它把上一頁擦掉）。
			// 略過它會讓下一條 `PRINT` 看起來像「沒有自己的 PRINTCLEAR」，
			// `ECL2.DAX/0x01 +024Eh` 就是這樣被誤判成子程式片段的。
			if opcode == 0x12 {
				current, pendingGosub, cleared = nil, nil, true
			}
			continue
		}
		// `12h PRINTCLEAR` 會把文字游標重設到左上（spec 1104：多寫 65A0h／65A1h），
		// 也就是**開新的一頁**；`11h PRINT` 接在目前這一頁後面。所以一段 ＝ 一頁，
		// 玩家看到的就是一頁。
		if opcode == 0x12 || current == nil {
			// `11h PRINT` 開頭卻沒有自己的 `PRINTCLEAR` ⇒ 這一頁是從剛才那支
			// 子程式裡開始的，前綴在裡面。`ECL4.DAX/0x25` 的八個遭遇頁都是這個
			// 形狀：`GOSUB 9534h` 印「YOU ARE ATTACKED BY」，回來才印怪物名。
			inserts := pendingGosub
			if opcode == 0x12 {
				// 上一頁尾巴那個 GOSUB 多半是翻頁提示或 Yes／No，VM 會在那裡停住，
				// 所以新的一頁屬於下一個 run。
				if len(pendingGosub) > 0 {
					run++
				}
				inserts = nil
			}
			groups = append(groups, group{
				Member: member, Block: blockID,
				Offset: fmt.Sprintf("0x%04X", instruction.Offset), Text: text,
				Run:          run,
				GosubInserts: inserts,
				appended:     opcode == 0x11 && !cleared,
				gosubTargets: gosubTargetsOf(instructions, inserts),
			})
			current, pendingGosub, cleared = &groups[len(groups)-1], nil, false
			current.VariableInserts = append(current.VariableInserts, variables...)
			continue
		}
		current.Text += " " + text
		current.VariableInserts = append(current.VariableInserts, variables...)
		current.GosubInserts = append(current.GosubInserts, pendingGosub...)
		current.gosubTargets = append(current.gosubTargets, gosubTargetsOf(instructions, pendingGosub)...)
		pendingGosub = nil
	}
	for index := range groups {
		if branching[groups[index].Run] {
			groups[index].DynamicBranch = true
		}
	}
	// 被記成「插入點」的那些 GOSUB，它們的目標整支都是片段——`YOU ARE ATTACKED BY`
	// 用 `PRINTCLEAR` 起頭卻仍然是前綴，靠 `appended` 判不出來，靠呼叫端判得出來。
	inserting := map[uint16]bool{}
	for _, item := range groups {
		for _, insert := range item.GosubInserts {
			var offset int
			fmt.Sscanf(insert, "0x%04X", &offset)
			for _, instruction := range instructions {
				if instruction.Offset == offset && len(instruction.Operands) == 1 &&
					instruction.Operands[0].WordSet {
					inserting[instruction.Operands[0].Word] = true
				}
			}
		}
	}
	fragments := subroutineOffsets(instructions, inserting)
	tails := guardedGotoTargets(instructions)
	for index := range groups {
		if fragments[groups[index].Offset] {
			groups[index].fragment = true
		}
		if groups[index].appended && tails[groups[index].Offset] {
			groups[index].BranchTail = true
		}
	}
	return blockText{
		groups:        groups,
		subroutines:   subroutineOffsets(instructions, nil),
		branchTargets: branchTargets,
		branchRuns:    branchRuns,
		instructions:  instructions,
	}, nil
}

// blockText 是一個 block 的完整文字模型：切好的頁、子程式範圍、以及每個 run 的
// `ON GOTO` 目的地。後兩者讓 main 有辦法把「插進來的文字」展開成完整的 run。
type blockText struct {
	groups        []group
	subroutines   map[string]bool
	branchTargets map[int][]int
	branchRuns    map[int]int
	instructions  []ecl.Instruction
}

// runOrder 依 run 編號由小到大列出這個 block 的所有 run。
func (b blockText) runOrder() []int {
	seen := map[int]bool{}
	var order []int
	for _, item := range b.groups {
		if !seen[item.Run] {
			seen[item.Run] = true
			order = append(order, item.Run)
		}
	}
	sort.Ints(order)
	return order
}

// maxRunVariants 擋住組合爆炸。一個 run 同時有兩個插入點、每個插入點又有兩條
// 分支就已經四種；`ON GOTO` 的目的地可以有十幾個。上限到了就不再展開——
// 代價是可能少判一個命中，方向與「寧可少標也不要多標」相反但同樣安全：
// 少判會讓那一頁留在待辦裡被人看到，多判才會讓待辦悄悄消失。
const maxRunVariants = 48

// runVariants 把一個 run 在實機可能印出的文字序列列出來。
//
// 工具依 offset 順序切頁，切出來的是「這一段位址上寫了什麼」；玩家看到的卻是
// **一次執行**累積的文字，而那份文字有兩個來源不在這一段位址上：
//
//   - `02h GOSUB` 進去的子程式（翻頁提示不算，會印字的那種：`YOU ARE ATTACKED BY`、
//     `UP`／`DOWN`）。
//   - `25h ON GOTO`／`26h ON GOSUB` 跳進來之前，前面那一段已經累積的文字。
//
// 兩者都會讓 `MatchText` 拿到的字比這一段多。不展開就只能把那些頁標成
// 「比對不了」，而那正是**誤判做完的地方**：待辦被歸進一個看起來已處理的桶子。
func (b blockText) runVariants(run int) [][]string {
	variants := b.ownVariants(run)
	if prefix, ok := b.prefixRunOf(run); ok {
		variants = crossVariants(b.ownVariants(prefix), variants)
	}
	return variants
}

// ownVariants 只展開這個 run 自己的頁與它們的 GOSUB 插入點。
func (b blockText) ownVariants(run int) [][]string {
	variants := [][]string{{}}
	for _, item := range b.groups {
		if item.Run != run {
			continue
		}
		variants = appendToAll(variants, item.Text)
		if len(item.gosubTargets) > 0 {
			var inserted [][]string
			for _, target := range item.gosubTargets {
				inserted = append(inserted, b.subroutineTexts(target)...)
			}
			variants = crossVariants(variants, inserted)
		}
	}
	return variants
}

// prefixRunOf 找出「跳進這個 run 的那張 ON GOTO 表」屬於哪一個 run。
// 只往回找一層：分支的分支再往上追會指數成長，而 `all_contains` 是子字串比對，
// 多接一層通常不會多命中什麼。
func (b blockText) prefixRunOf(run int) (int, bool) {
	start, ok := b.branchRuns[run]
	if !ok {
		return 0, false
	}
	for owner, targets := range b.branchTargets {
		for _, target := range targets {
			if target == start && owner != run {
				return owner, true
			}
		}
	}
	return 0, false
}

// runAt 找出某個 payload offset 落在哪一個 run：取第一個 offset 不小於它的頁。
// 分支目的地常常先做幾條 SAVE／COMPARE 才 `PRINTCLEAR`，所以不能要求剛好相等。
func (b blockText) runAt(offset int) (int, bool) {
	best, bestOffset := 0, -1
	for _, item := range b.groups {
		var at int
		fmt.Sscanf(item.Offset, "0x%04X", &at)
		if at < offset {
			continue
		}
		if bestOffset < 0 || at < bestOffset {
			best, bestOffset = item.Run, at
		}
	}
	return best, bestOffset >= 0
}

// subroutineTexts 走一支子程式，列出它可能印出的文字序列。`IF` 的兩條路都要走
// （spec 1106：條件不成立時整條下一個指令被跳過），`UP`／`DOWN` 就是靠這個
// 分成兩種變體——把兩條併成一條會得到現實中不存在的「UP DOWN」。
func (b blockText) subroutineTexts(target int) [][]string {
	index := map[int]ecl.Instruction{}
	for _, instruction := range b.instructions {
		index[instruction.Offset] = instruction
	}
	var walk func(offset int, acc []string, budget int) [][]string
	walk = func(offset int, acc []string, budget int) [][]string {
		if budget <= 0 {
			return [][]string{acc}
		}
		instruction, ok := index[offset]
		if !ok {
			return [][]string{acc}
		}
		opcode := instruction.Command.Opcode
		switch {
		case opcode == 0x11 || opcode == 0x12:
			if printed := instructionText(instruction); printed != "" {
				acc = append(append([]string(nil), acc...), printed)
			}
			return walk(instruction.Next, acc, budget-1)
		case opcode == 0x13 || opcode == 0x00:
			return [][]string{acc}
		case opcode == 0x01:
			if len(instruction.Operands) == 1 {
				if next, ok := ecl.CodeTarget(instruction.Operands[0], 1<<20); ok {
					return walk(next, acc, budget-1)
				}
			}
			return [][]string{acc}
		case opcode >= 0x16 && opcode <= 0x1B:
			// 條件成立：執行下一條。條件不成立：跳過它，從再下一條繼續。
			taken := walk(instruction.Next, acc, budget-1)
			skipped := [][]string{acc}
			if guardedInstruction, ok := index[instruction.Next]; ok {
				skipped = walk(guardedInstruction.Next, acc, budget-1)
			}
			return append(taken, skipped...)
		default:
			return walk(instruction.Next, acc, budget-1)
		}
	}
	return walk(target, nil, 64)
}

func appendToAll(variants [][]string, text string) [][]string {
	for index := range variants {
		variants[index] = append(append([]string(nil), variants[index]...), text)
	}
	return variants
}

// crossVariants 把「這一頁的文字」與「插進來的每一種文字」相乘。
// `all_contains` 是子字串比對，順序不影響結果，所以插入的字接在後面就夠了
// ——前提是規則的片段不跨越插入點（Limitations 已寫明）。
func crossVariants(variants, inserted [][]string) [][]string {
	if len(inserted) == 0 {
		return variants
	}
	var out [][]string
	for _, base := range variants {
		for _, extra := range inserted {
			if len(out) >= maxRunVariants {
				return out
			}
			out = append(out, append(append([]string(nil), base...), extra...))
		}
	}
	return out
}

// gosubTargetsOf 把「GOSUB 指令的 offset」換成「它跳去的 payload offset」。
func gosubTargetsOf(instructions []ecl.Instruction, inserts []string) []int {
	if len(inserts) == 0 {
		return nil
	}
	wanted := map[int]bool{}
	for _, insert := range inserts {
		var offset int
		fmt.Sscanf(insert, "0x%04X", &offset)
		wanted[offset] = true
	}
	var targets []int
	for _, instruction := range instructions {
		if !wanted[instruction.Offset] || len(instruction.Operands) != 1 {
			continue
		}
		if operand := instruction.Operands[0]; operand.WordSet &&
			int(operand.Word) >= ecl.CodeAddressBase {
			targets = append(targets, int(operand.Word)-ecl.CodeAddressBase)
		}
	}
	return targets
}

// guardedGotoTargets 收集「被 IF 守衛的 GOTO」跳過去的位址。那種 GOTO 是
// if／else 的分叉，目標若又以 `11h PRINT` 起頭（沒有自己的 PRINTCLEAR），
// 它就是**同一頁的另一條續寫**，不是獨立的一頁：
//
//	PRINTCLEAR «YOU FIND A SECRET DOOR TO THE »
//	IF =; GOTO east
//	PRINT «WEST.»   ; GOTO after
//	east: PRINT «EAST.»
//
// 實機只會印出其中一條，玩家看到的是完整的一句。工具依 offset 順序切頁，
// 必然把後面那條切成獨立的一段，而那一段的文字（`EAST.`）短到不能寫規則。
// ⇒ 標成 branch-tail：不是待辦，但**它的譯文要跟著主分支一起手寫**，
// 這一點工具驗不到（見 report 的 limitations）。
func guardedGotoTargets(instructions []ecl.Instruction) map[string]bool {
	targets := map[string]bool{}
	guarded := false
	for _, instruction := range instructions {
		if guarded && instruction.Command.Opcode == 0x01 && len(instruction.Operands) == 1 {
			if operand := instruction.Operands[0]; operand.WordSet &&
				int(operand.Word) >= ecl.CodeAddressBase {
				targets[fmt.Sprintf("0x%04X", int(operand.Word)-ecl.CodeAddressBase)] = true
			}
		}
		guarded = instruction.Command.Opcode >= 0x16 && instruction.Command.Opcode <= 0x1B
	}
	return targets
}

// endsRun 是真正把累積文字交給 `MatchText` 的那些指令，取自
// `internal/ecl/runtime.go` 裡會 `return result` 的 case。控制流（`GOTO`、
// `RETURN`、`ON GOTO`）不在此列——那是 `endsTextGroup` 的事。
func endsRun(opcode byte) bool {
	switch opcode {
	case 0x00, 0x20, 0x24, 0x27, 0x15, 0x29, 0x2B, 0x2C, 0x10, 0x0F, 0x39, 0x38:
		return true
	default:
		return false
	}
}

// subroutineOffsets 標出「這一段其實是某支子程式的內容」。這種段落在實機不會
// 單獨出現（它一定是被某一頁 GOSUB 進來的），所以既不算做完也不算待辦——
// 替它們寫規則反而有害：只有一兩個字的 `all_contains` 會攔截到別的文字。
//
// 子程式的範圍：從呼叫目標往後走，遇到**沒被 IF 守衛**的 `RETURN`／`EXIT` 就結束。
// 例外是 `IF; GOTO x` 的 else 分支——`UP`／`DOWN` 那種兩條分支的子程式，第一條
// 印完就 `RETURN`，不把 x 納進來只會算到一半。
//
// ⚠ 這個延伸**只吃被守衛、且跳很近的 GOTO**（256 byte 內）。第一版對所有 GOTO
// 都延伸，結果把 `THE DOOR IS LOCKED`、`RED PLUME GUARDS RUSH AT YOU` 這些真正的
// 頁一起吞進來——**寧可少標也不要多標**：少標只是清單裡多幾筆看起來奇怪的短句，
// 多標會讓真正的待辦從清單上消失。
func subroutineOffsets(instructions []ecl.Instruction, only map[uint16]bool) map[string]bool {
	targets := map[int]bool{}
	for _, instruction := range instructions {
		if instruction.Command.Opcode != 0x02 || len(instruction.Operands) != 1 {
			continue
		}
		operand := instruction.Operands[0]
		if !operand.WordSet {
			continue
		}
		if only != nil && !only[operand.Word] {
			continue
		}
		targets[int(operand.Word)-ecl.CodeAddressBase] = true
	}
	inside := map[string]bool{}
	for target := range targets {
		guarded := false
		limit := target
		for _, instruction := range instructions {
			if instruction.Offset < target {
				continue
			}
			wasGuarded := guarded
			guarded = instruction.Command.Opcode >= 0x16 && instruction.Command.Opcode <= 0x1B
			inside[fmt.Sprintf("0x%04X", instruction.Offset)] = true
			if wasGuarded && instruction.Command.Opcode == 0x01 && len(instruction.Operands) == 1 &&
				instruction.Operands[0].WordSet {
				branch := int(instruction.Operands[0].Word) - ecl.CodeAddressBase
				if branch > limit && branch-instruction.Offset <= 256 {
					limit = branch
				}
			}
			if wasGuarded {
				continue
			}
			if instruction.Command.Opcode == 0x13 || instruction.Command.Opcode == 0x00 {
				if instruction.Offset >= limit {
					break
				}
			}
		}
	}
	return inside
}

// endsTextGroup lists the opcodes after which the next instruction in offset
// order is no longer the next instruction to execute. Everything else keeps
// accumulating into the current page.
//
// 兩類邊界：
//
//   - **交出文字**：`internal/ecl/runtime.go` 裡會 `return result` 的那些 case
//     （離開、換 block、戰鬥、寶物、三種選單、字串輸入、選人、終局）。
//   - **離開直線**：`GOTO`／`RETURN`／`ON GOTO`／`ON GOSUB` 之後的位元組是別條路徑
//     或別支子程式。同 `eclcatalog.endsStraightLine`。
//
// ⚠ 少了第二類會把**互斥分支**的文字併成同一頁。`ECL5.DAX/0x33 +09B2h` 是
// `IF =; GOTO 89BDh`：不成立印 `UP` 後 `RETURN`，成立印 `DOWN` 後 `RETURN`，
// 併起來就得到現實中不存在的「UP DOWN」，任何規則都接不上它。
//
// ⚠⚠ 但**被 `IF` 守衛的**這些指令不算邊界（判斷寫在呼叫端）。ECL 表達 if/else
// 的慣用法就是 `IF; GOTO x`，條件不成立時那個 `GOTO` 整條被跳過——落下來的
// 文字仍是同一頁。把守衛版本也當邊界會切出「YOU ARE AT THE STANDING STONES.」
// 這種半頁，反而讓原本接得上的規則接不上。
//
// `02h GOSUB` 不在此列：它會回來，後面的文字仍屬同一頁。代價寫在 Limitations。
func endsTextGroup(opcode byte) bool {
	switch opcode {
	case 0x00, // EXIT
		0x20, // NEWECL
		0x24, // COMBAT
		0x27, // TREASURE
		0x15, // VERTICAL MENU
		0x29, // ENCOUNTER MENU
		0x2B, // HORIZONTAL MENU
		0x2C, // PARLAY
		0x10, // INPUT STRING
		0x0F, // INPUT NUMBER
		0x39, // WHO
		0x38, // PROGRAM
		0x01, // GOTO
		0x13, // RETURN
		0x25, // ON GOTO
		0x26: // ON GOSUB
		return true
	default:
		return false
	}
}

// instructionVariables 回報這一條 PRINT 印了哪些執行期的值。`80h` 是打包好的
// 靜態文字，其餘（`81h` 字串記憶體、`01h` 變數、立即值）都要跑起來才知道。
func instructionVariables(instruction ecl.Instruction) []string {
	if instruction.Command.Opcode != 0x11 && instruction.Command.Opcode != 0x12 {
		return nil
	}
	var out []string
	for _, operand := range instruction.Operands {
		if operand.Code == 0x80 {
			continue
		}
		// 記**值從哪裡來**，不是記這條指令在哪裡：要判斷這一頁能不能按值列舉
		// （像 `world.night-note.24/35/42`），得先知道那個值是哪一格。
		if operand.WordSet {
			out = append(out, fmt.Sprintf("0x%04X@%04Xh", instruction.Offset, operand.Word))
			continue
		}
		out = append(out, fmt.Sprintf("0x%04X@imm%d", instruction.Offset, operand.Low))
	}
	return out
}

func instructionText(instruction ecl.Instruction) string {
	var parts []string
	for _, operand := range instruction.Operands {
		if operand.Code != 0x80 || len(operand.Packed) == 0 {
			continue
		}
		decoded := strings.TrimSpace(ecl.DecodePackedText(operand.Packed))
		if decoded != "" {
			parts = append(parts, decoded)
		}
	}
	return strings.Join(parts, " ")
}

func readMember(reader *zip.Reader, name string) ([]byte, error) {
	for _, file := range reader.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer handle.Close()
		return io.ReadAll(handle)
	}
	return nil, fmt.Errorf("archive member %s is missing", name)
}

func renderMarkdown(result report) []byte {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.7e7bcf7694d4"))
	out.WriteString(tooltext.Text("h.e33806e5d28a"))
	for _, limitation := range result.Limitations {
		fmt.Fprintf(&out, "- %s\n", limitation)
	}
	fmt.Fprint(&out, tooltext.Format("h.10af87ec2e4b"))
	fmt.Fprint(&out, tooltext.Format("h.a7d38a8ee7ed", result.Summary.Groups))
	fmt.Fprint(&out, tooltext.Format("h.969ccd1ebc9c", result.Summary.Matched))
	fmt.Fprint(&out, tooltext.Format("h.3683b87494bd", result.Summary.Unmatched))
	fmt.Fprint(&out, tooltext.Format("h.e54664af3947", result.Summary.VariableInsert))
	fmt.Fprint(&out, tooltext.Format("h.ffac0daf9c8f", result.Summary.Subroutine))
	out.WriteString(tooltext.Text("h.c259e43fd5a9"))
	fmt.Fprint(&out, tooltext.Format("h.87c105b12a3b"))
	fmt.Fprint(&out, tooltext.Format("h.c37e771df231", result.Summary.GosubInsert))
	fmt.Fprint(&out, tooltext.Format("h.8230e70169b4", result.Summary.BranchInsert))

	blocks := make([]string, 0, len(result.Summary.ByBlock))
	for key := range result.Summary.ByBlock {
		blocks = append(blocks, key)
	}
	sort.Slice(blocks, func(i, j int) bool {
		if result.Summary.ByBlock[blocks[i]] != result.Summary.ByBlock[blocks[j]] {
			return result.Summary.ByBlock[blocks[i]] > result.Summary.ByBlock[blocks[j]]
		}
		return blocks[i] < blocks[j]
	})
	out.WriteString(tooltext.Text("h.5d246149be56"))
	for _, key := range blocks {
		fmt.Fprintf(&out, "| `%s` | %d |\n", key, result.Summary.ByBlock[key])
	}

	fmt.Fprint(&out, tooltext.Format("h.3b827aa6eec5"))
	out.WriteString(tooltext.Text("h.6594090a6501") +
		tooltext.Text("h.6bd68b0a97c2") +
		tooltext.Text("h.c60b95adad5e"))
	out.WriteString(tooltext.Text("h.2ed9df2c8c0e") +
		tooltext.Text("h.3d58f6511ecf") +
		tooltext.Text("h.efacba4510e4") +
		tooltext.Text("h.1d099a62e5f2"))
	fmt.Fprint(&out, tooltext.Format("h.92f593c42d61"))
	fmt.Fprint(&out, tooltext.Format("h.317feaa4937d", result.Summary.Runs))
	fmt.Fprint(&out, tooltext.Format("h.31f9b4f0eb3f", result.Summary.RunsMatched))
	fmt.Fprint(&out, tooltext.Format("h.67b223c410da", result.Summary.RunsOrphan))
	fmt.Fprint(&out, tooltext.Format("h.5bcdd8bed616", result.Summary.RunsVariableInsert))
	fmt.Fprint(&out, tooltext.Format("h.6d3f8d0484cf", result.Summary.RunsSubroutine))
	fmt.Fprint(&out, tooltext.Format("h.59a85fe53434", result.Summary.RunsGosubInsert))
	fmt.Fprint(&out, tooltext.Format("h.b8146d0bd562", result.Summary.RunsTruncated))

	orphanBlocks := make([]string, 0, len(result.Summary.OrphanByBlock))
	for key := range result.Summary.OrphanByBlock {
		orphanBlocks = append(orphanBlocks, key)
	}
	sort.Slice(orphanBlocks, func(i, j int) bool {
		left, right := result.Summary.OrphanByBlock[orphanBlocks[i]], result.Summary.OrphanByBlock[orphanBlocks[j]]
		if left != right {
			return left > right
		}
		return orphanBlocks[i] < orphanBlocks[j]
	})
	out.WriteString("| Block | `orphan` run |\n|---|---:|\n")
	for _, key := range orphanBlocks {
		fmt.Fprintf(&out, "| `%s` | %d |\n", key, result.Summary.OrphanByBlock[key])
	}

	out.WriteString(tooltext.Text("h.7eed76896cd2") +
		tooltext.Text("h.da3ebb355058") +
		tooltext.Text("h.12cc03399aba") +
		tooltext.Text("h.b6cdf7d0bed3"))
	out.WriteString(tooltext.Text("h.4fed16ba219d"))
	for _, item := range result.OrphanRuns {
		start := ""
		if len(item.Pages) > 0 {
			start = item.Pages[0]
		}
		fmt.Fprintf(&out, "| `%s/%s` | `%s` | %d | %s |\n",
			item.Member, item.Block, start, len(item.Pages), markdownCell(item.Text))
	}

	out.WriteString(tooltext.Text("h.a6b7b0aa6256"))
	for _, item := range result.Groups {
		rule := "—"
		if item.RuleID != "" {
			rule = "`" + item.RuleID + "`"
		}
		text := strings.ReplaceAll(item.Text, "|", "\\|")
		if len(text) > 160 {
			text = text[:160] + "…"
		}
		if len(item.GosubInserts) > 0 {
			text += tooltext.Text("h.0e4f348ce2b3") + strings.Join(item.GosubInserts, "`／`") + tooltext.Text("h.5ab59d4f32d6")
		}
		fmt.Fprintf(&out, "| `%s/%s` | `%s` | `%s` | %s | %s |\n",
			item.Member, item.Block, item.Offset, item.Status, rule, text)
	}
	return []byte(out.String())
}
