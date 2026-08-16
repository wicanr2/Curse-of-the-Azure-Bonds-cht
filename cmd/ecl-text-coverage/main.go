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
}

type report struct {
	Schema      string   `json:"schema"`
	Limitations []string `json:"limitations"`
	Groups      []group  `json:"groups"`
	Summary     summary  `json:"summary"`
}

type summary struct {
	Groups    int            `json:"groups"`
	Matched   int            `json:"matched"`
	Unmatched int            `json:"unmatched"`
	ByBlock   map[string]int `json:"unmatched_by_block"`
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
			"分母是原作的文字段落數，不是「玩家實際會看到幾段」——同一段可能永遠不會觸發。",
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
			groups, err := blockGroups(member, blockID, raw.Data)
			if err != nil {
				log.Fatalf("%s block %s: %v", member, blockID, err)
			}
			for _, item := range groups {
				if match := pack.MatchText([]string{item.Text}, pack.DefaultLocale); match.Matched {
					item.RuleID = match.RuleID
					result.Summary.Matched++
				} else {
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
	fmt.Fprintf(os.Stderr, "text_groups=%d matched=%d unmatched=%d\n",
		result.Summary.Groups, result.Summary.Matched, result.Summary.Unmatched)
}

// blockGroups collects the reachable PRINTCLEAR/PRINT runs of one block.
func blockGroups(member, blockID string, data []byte) ([]group, error) {
	if len(data) < 2 {
		return nil, nil
	}
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
	instructions := append([]ecl.Instruction(nil), graph.Instructions...)
	sort.Slice(instructions, func(i, j int) bool {
		return instructions[i].Offset < instructions[j].Offset
	})

	var groups []group
	var current *group
	for _, instruction := range instructions {
		opcode := instruction.Command.Opcode
		if opcode != 0x11 && opcode != 0x12 {
			// 任何其他指令都結束目前這一段：原作的文字框在下一個 PRINTCLEAR
			// 才會被清掉，但中間夾了別的 opcode 就不是同一次呈現了。
			current = nil
			continue
		}
		text := instructionText(instruction)
		if text == "" {
			continue
		}
		if opcode == 0x12 || current == nil {
			groups = append(groups, group{
				Member: member, Block: blockID,
				Offset: fmt.Sprintf("0x%04X", instruction.Offset), Text: text,
			})
			current = &groups[len(groups)-1]
			continue
		}
		current.Text += " " + text
	}
	return groups, nil
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
	fmt.Fprintf(&out, "\n## 摘要\n\n| 項目 | 數量 |\n|---|---:|\n")
	fmt.Fprintf(&out, "| 靜態可達的文字段落 | %d |\n", result.Summary.Groups)
	fmt.Fprintf(&out, "| 已接上 `text_rule` | %d |\n", result.Summary.Matched)
	fmt.Fprintf(&out, "| **未接上** | **%d** |\n", result.Summary.Unmatched)

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

	out.WriteString("\n## 逐段\n\n| Block | offset | 已接上的規則 | 原作文字 |\n|---|---|---|---|\n")
	for _, item := range result.Groups {
		rule := "—"
		if item.RuleID != "" {
			rule = "`" + item.RuleID + "`"
		}
		text := strings.ReplaceAll(item.Text, "|", "\\|")
		if len(text) > 160 {
			text = text[:160] + "…"
		}
		fmt.Fprintf(&out, "| `%s/%s` | `%s` | %s | %s |\n",
			item.Member, item.Block, item.Offset, rule, text)
	}
	return []byte(out.String())
}
