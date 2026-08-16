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
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
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
	// Status 是這一段在內容產出裡的處置：
	//
	//	matched      有規則命中，做完了
	//	unmatched    還沒寫規則——這才是待辦
	//	gosub-insert 文字被子程式插過，本工具比對不了；規則可能早就寫好了
	//	subroutine   這一段本身就是某支子程式的內容，實機不會單獨出現
	Status string `json:"status"`
	// appended 記錄這一段是以 `11h PRINT` 起頭（true）還是 `12h PRINTCLEAR`（false）。
	// 只有前者可能是「接在呼叫端那一頁後面」的片段——它自己不開新頁。
	appended bool
}

type report struct {
	Schema      string   `json:"schema"`
	Limitations []string `json:"limitations"`
	Groups      []group  `json:"groups"`
	Summary     summary  `json:"summary"`
}

type summary struct {
	Groups      int            `json:"groups"`
	Matched     int            `json:"matched"`
	Unmatched   int            `json:"unmatched"`
	GosubInsert int            `json:"gosub_insert"`
	Subroutine  int            `json:"subroutine"`
	ByBlock     map[string]int `json:"unmatched_by_block"`
}

func main() {
	archive := flag.String("archive", "curseoftheazurebonds.zip", "original CoAB archive")
	output := flag.String("output", "", "write the Markdown report to this path")
	outputJSON := flag.String("json", "", "write the machine-readable report to this path")
	block := flag.String("block", "", "limit to one block, e.g. ECL3.DAX/0x10")
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
			"只統計靜態可達的文字：TraceGraph 跟 GOTO／GOSUB、循序 fallthrough 與 IF 的兩條路（spec 1106），但**不跟** ON GOTO／ON GOSUB 的動態目的地與選單分支——分母只會再往上。",
			"一段 ＝ 一頁：`12h PRINTCLEAR` 重設文字游標就是開新頁（spec 1104）。原作的翻頁提示放在 `02h GOSUB` 進去的子程式裡，本工具不追進去，所以看不到 run 的邊界——不要用相鄰段落去推論它們是不是同一次 run。",
			"`02h GOSUB` 進去印的文字不會併進這一段，標 `gosub_inserts` 的頁實際文字比這裡多。插入點在中間（`ECL5.DAX/0x33 +091Fh` 實際是「THE STAIRS LEAD ⟨UP｜DOWN⟩ HERE. DO YOU WANT TO TAKE THEM?」）或在開頭當前綴（`ECL4.DAX/0x25` 的八個遭遇頁前面都有「YOU ARE ATTACKED BY」）。⇒ 寫 `all_contains` 時**片段不要跨越插入點**，否則實機接不上。",
			"反過來，被 `GOSUB` 呼叫的純文字子程式會自成一段（如 `UP`／`DOWN`／`YOU ARE ATTACKED BY`）。它們在實機不會單獨出現，**不要**替它們寫規則——只有一兩個字的規則會攔截到別的文字。",
			"『已接上』只表示有一條 text_rule 的 all_contains 全部命中，不代表譯文正確或事件副作用已還原。",
		},
		Summary: summary{ByBlock: map[string]int{}},
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
			groups, subroutines, err := blockGroups(member, blockID, raw.Data)
			if err != nil {
				log.Fatalf("%s block %s: %v", member, blockID, err)
			}
			for _, item := range groups {
				match := pack.MatchText([]string{item.Text}, pack.DefaultLocale)
				switch {
				case match.Matched:
					item.RuleID = match.RuleID
					item.Status = "matched"
					result.Summary.Matched++
				// 只有「以 PRINT 起頭又落在子程式體內」才是片段。用 PRINTCLEAR 起頭的
				// 子程式是**共用的一整頁**（`THE DOOR IS LOCKED` 就是），玩家看得到，
				// 那是待辦不是雜訊。
				case item.appended && subroutines[item.Offset]:
					item.Status = "subroutine"
					result.Summary.Subroutine++
				case len(item.GosubInserts) > 0:
					item.Status = "gosub-insert"
					result.Summary.GosubInsert++
				default:
					item.Status = "unmatched"
					result.Summary.Unmatched++
					result.Summary.ByBlock[member+"/"+blockID]++
				}
				result.Summary.Groups++
				result.Groups = append(result.Groups, item)
			}
		}
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
	fmt.Fprintf(os.Stderr, "text_groups=%d matched=%d unmatched=%d gosub-insert=%d subroutine=%d\n",
		result.Summary.Groups, result.Summary.Matched, result.Summary.Unmatched,
		result.Summary.GosubInsert, result.Summary.Subroutine)
}

// blockGroups collects the reachable PRINTCLEAR/PRINT runs of one block, plus
// the set of group offsets that live inside a GOSUB'd subroutine.
func blockGroups(member, blockID string, data []byte) ([]group, map[string]bool, error) {
	if len(data) < 2 {
		return nil, nil, nil
	}
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil, nil, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil, nil, err
	}
	instructions := append([]ecl.Instruction(nil), graph.Instructions...)
	sort.Slice(instructions, func(i, j int) bool {
		return instructions[i].Offset < instructions[j].Offset
	})

	var groups []group
	var current *group
	var pendingGosub []string
	guarded := false
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
				inserts = nil
			}
			groups = append(groups, group{
				Member: member, Block: blockID,
				Offset: fmt.Sprintf("0x%04X", instruction.Offset), Text: text,
				GosubInserts: inserts,
				appended:     opcode == 0x11 && !cleared,
			})
			current, pendingGosub, cleared = &groups[len(groups)-1], nil, false
			continue
		}
		current.Text += " " + text
		current.GosubInserts = append(current.GosubInserts, pendingGosub...)
		pendingGosub = nil
	}
	return groups, subroutineOffsets(instructions), nil
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
func subroutineOffsets(instructions []ecl.Instruction) map[string]bool {
	targets := map[int]bool{}
	for _, instruction := range instructions {
		if instruction.Command.Opcode != 0x02 || len(instruction.Operands) != 1 {
			continue
		}
		if operand := instruction.Operands[0]; operand.WordSet {
			targets[int(operand.Word)-ecl.CodeAddressBase] = true
		}
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
	out.WriteString("# ECL 文字覆蓋（原作文字段落 → game pack）\n\n")
	out.WriteString("由 `cmd/ecl-text-coverage` 產生，不要手改。\n\n")
	for _, limitation := range result.Limitations {
		fmt.Fprintf(&out, "- %s\n", limitation)
	}
	fmt.Fprintf(&out, "\n## 摘要\n\n| 處置 | 數量 | 意思 |\n|---|---:|---|\n")
	fmt.Fprintf(&out, "| 靜態可達的文字段落 | %d | — |\n", result.Summary.Groups)
	fmt.Fprintf(&out, "| `matched` | %d | 有規則命中 |\n", result.Summary.Matched)
	fmt.Fprintf(&out, "| **`unmatched`** | **%d** | **還沒寫規則——這才是待辦** |\n",
		result.Summary.Unmatched)
	fmt.Fprintf(&out, "| `gosub-insert` | %d | 文字被子程式插過，本工具比對不了；規則可能早就寫好了 |\n",
		result.Summary.GosubInsert)
	fmt.Fprintf(&out, "| `subroutine` | %d | 這一段本身是某支子程式的內容，實機不會單獨出現 |\n",
		result.Summary.Subroutine)

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
	out.WriteString("\n## 未接上的段落，依 block\n\n| Block | 未接上 |\n|---|---:|\n")
	for _, key := range blocks {
		fmt.Fprintf(&out, "| `%s` | %d |\n", key, result.Summary.ByBlock[key])
	}

	out.WriteString("\n## 逐段\n\n| Block | offset | 處置 | 已接上的規則 | 原作文字 |\n|---|---|---|---|---|\n")
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
			text += " ⚠ 另有 `GOSUB " + strings.Join(item.GosubInserts, "`／`") + "` 印出的一段"
		}
		fmt.Fprintf(&out, "| `%s/%s` | `%s` | `%s` | %s | %s |\n",
			item.Member, item.Block, item.Offset, item.Status, rule, text)
	}
	return []byte(out.String())
}
