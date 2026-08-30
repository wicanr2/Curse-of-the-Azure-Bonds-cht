// Command ecl-effect-coverage answers the side-effect half of RE-04: of the
// instructions the player can actually reach, how many have their original
// effect reproduced in the remake?
//
// `cmd/ecl-text-coverage` 回答的是「文字接上沒有」，而文字只是事件的一半。
// 另一半是**副作用**：旗標、寶物、法術、畫面、資產載入。那一半在程式碼裡不容易
// 看出缺口——`runtime.go` 對每個 opcode 都有 `case`，只把運算元讀掉再往下走的
// 那幾個，和完整執行的長得一模一樣，測試也一樣綠。
//
// 這支把 `internal/ecl.OpcodeEffects` 的還原狀態乘上 corpus 的**可達出現次數**，
// 把「還差多少」變成數字。
//
//	go run ./cmd/ecl-effect-coverage -output docs/audit/ecl-effect-coverage.md
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

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

var members = []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"}

type opcodeRow struct {
	Opcode      string   `json:"opcode"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Note        string   `json:"note"`
	Occurrences int      `json:"occurrences"`
	Blocks      []string `json:"blocks"`
}

type report struct {
	Schema       string         `json:"schema"`
	Limitations  []string       `json:"limitations"`
	Rows         []opcodeRow    `json:"opcodes"`
	Summary      map[string]int `json:"instructions_by_status"`
	Opcodes      map[string]int `json:"opcodes_by_status"`
	Reached      int            `json:"reachable_instructions"`
	ByBlock      map[string]int `json:"reachable_by_block"`
	Unreproduced map[string]int `json:"unreproduced_by_block"`
}

func main() {
	archive := flag.String("archive", "curseoftheazurebonds.zip", "original CoAB archive")
	output := flag.String("output", "", "write the Markdown report to this path")
	outputJSON := flag.String("json", "", "write the machine-readable report to this path")
	flag.Parse()

	reader, err := zip.OpenReader(*archive)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	counts := map[byte]int{}
	blocks := map[byte]map[string]bool{}
	result := report{
		Schema: "coab-ecl-effect-coverage/1",
		Limitations: []string{
			tooltext.Text("h.1ad6e7ec8949"),
			tooltext.Text("h.6d3c6d9fd3a5"),
			tooltext.Text("h.6096175df7ec"),
			tooltext.Text("h.be28b8856b31"),
		},
		Summary:      map[string]int{},
		Opcodes:      map[string]int{},
		ByBlock:      map[string]int{},
		Unreproduced: map[string]int{},
	}

	for _, member := range members {
		data, err := readMember(&reader.Reader, member)
		if err != nil {
			log.Fatal(err)
		}
		parsed, err := dax.Parse(data)
		if err != nil {
			log.Fatalf("%s: %v", member, err)
		}
		for _, raw := range parsed {
			label := fmt.Sprintf("%s/0x%02X", member, raw.Entry.ID)
			points, _, err := ecl.EntryPoints(raw.Data, 5)
			if err != nil {
				log.Fatalf("%s entry points: %v", label, err)
			}
			starts := make([]int, 0, len(points))
			for _, point := range points {
				starts = append(starts, int(point)-ecl.CodeAddressBase)
			}
			graph, err := ecl.TraceGraph(raw.Data, starts, len(raw.Data)*8)
			if err != nil {
				log.Fatalf("%s graph: %v", label, err)
			}
			seen := map[int]bool{}
			for _, instruction := range graph.Instructions {
				if seen[instruction.Offset] {
					continue
				}
				seen[instruction.Offset] = true
				opcode := instruction.Command.Opcode
				counts[opcode]++
				if blocks[opcode] == nil {
					blocks[opcode] = map[string]bool{}
				}
				blocks[opcode][label] = true
				result.Reached++
				result.ByBlock[label]++
				effect, ok := ecl.OpcodeEffects[opcode]
				if !ok {
					log.Fatalf(tooltext.Text("h.1318ee02f82d")+
						tooltext.Text("h.ede1980f60af"), label, instruction.Offset, opcode)
				}
				result.Summary[string(effect.Status)]++
				if effect.Status != ecl.EffectDone {
					result.Unreproduced[label]++
				}
			}
		}
	}

	opcodes := make([]int, 0, len(counts))
	for opcode := range counts {
		opcodes = append(opcodes, int(opcode))
	}
	sort.Ints(opcodes)
	for _, value := range opcodes {
		opcode := byte(value)
		effect := ecl.OpcodeEffects[opcode]
		names := make([]string, 0, len(blocks[opcode]))
		for label := range blocks[opcode] {
			names = append(names, label)
		}
		sort.Strings(names)
		result.Rows = append(result.Rows, opcodeRow{
			Opcode:      fmt.Sprintf("0x%02X", opcode),
			Name:        ecl.KnownCommands[opcode].Name,
			Status:      string(effect.Status),
			Note:        effect.Note,
			Occurrences: counts[opcode],
			Blocks:      names,
		})
		result.Opcodes[string(effect.Status)]++
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
	fmt.Fprintf(os.Stderr,
		"reachable=%d done=%d partial=%d consumed=%d（opcode：done=%d partial=%d consumed=%d）\n",
		result.Reached, result.Summary["done"], result.Summary["partial"], result.Summary["consumed"],
		result.Opcodes["done"], result.Opcodes["partial"], result.Opcodes["consumed"])
}

func renderMarkdown(result report) []byte {
	var out strings.Builder
	out.WriteString(tooltext.Text("h.375750fcde7b"))
	out.WriteString(tooltext.Text("h.1524d84e5e25"))
	for _, limitation := range result.Limitations {
		fmt.Fprintf(&out, "- %s\n", limitation)
	}
	fmt.Fprint(&out, tooltext.Format("h.95ec57dada54"))
	fmt.Fprint(&out, tooltext.Format("h.8c0af6ac737f", result.Summary["done"], result.Opcodes["done"]))
	fmt.Fprint(&out, tooltext.Format("h.d8c488e7c0f4", result.Summary["partial"], result.Opcodes["partial"]))
	fmt.Fprint(&out, tooltext.Format("h.f92778f6ba60", result.Summary["consumed"], result.Opcodes["consumed"]))
	fmt.Fprint(&out, tooltext.Format("h.ca832ad3777b", result.Reached, len(result.Rows)))

	out.WriteString(tooltext.Text("h.2ab171c6b700"))
	rows := append([]opcodeRow(nil), result.Rows...)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Status != rows[j].Status {
			return statusOrder(rows[i].Status) < statusOrder(rows[j].Status)
		}
		if rows[i].Occurrences != rows[j].Occurrences {
			return rows[i].Occurrences > rows[j].Occurrences
		}
		return rows[i].Opcode < rows[j].Opcode
	})
	for _, row := range rows {
		fmt.Fprintf(&out, "| `%s` | %s | `%s` | %d | %d | %s |\n",
			row.Opcode, row.Name, row.Status, row.Occurrences, len(row.Blocks), row.Note)
	}

	out.WriteString(tooltext.Text("h.fd65723c7d44"))
	labels := make([]string, 0, len(result.ByBlock))
	for label := range result.ByBlock {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool {
		if result.Unreproduced[labels[i]] != result.Unreproduced[labels[j]] {
			return result.Unreproduced[labels[i]] > result.Unreproduced[labels[j]]
		}
		return labels[i] < labels[j]
	})
	for _, label := range labels {
		fmt.Fprintf(&out, "| `%s` | %d | %d |\n", label, result.ByBlock[label], result.Unreproduced[label])
	}
	return []byte(out.String())
}

func statusOrder(status string) int {
	switch status {
	case string(ecl.EffectConsumed):
		return 0
	case string(ecl.EffectPartial):
		return 1
	default:
		return 2
	}
}

func readMember(reader *zip.Reader, name string) ([]byte, error) {
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer handle.Close()
		return io.ReadAll(handle)
	}
	return nil, fmt.Errorf("%s is not in the archive", name)
}
