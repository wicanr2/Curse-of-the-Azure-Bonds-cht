// Command ecl-run-coverage 把「實跑路線執行過的 ECL 指令」對上靜態可達的
// 指令全集，回答「全城市／全房間走訪」那一列先前沒有的分母。
//
// 分母與 `cmd/ecl-effect-coverage` 同一套（EntryPoints ＋ TraceGraph）；
// 分子來自 `COAB_ECL_COVERAGE` 記錄器（internal/ecl/runcoverage.go）。
// 產生分子的指令寫在報表開頭，要與程式碼同版。
package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

var eclMembers = []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"}

type blockCoverage struct {
	Label     string `json:"label"`
	Reachable int    `json:"reachable"`
	Executed  int    `json:"executed"`
}

type cluster struct {
	Label   string `json:"label"`
	Address string `json:"start_address"`
	Size    int    `json:"size"`
}

type report struct {
	Schema    string          `json:"schema"`
	Reachable int             `json:"reachable_instructions"`
	Executed  int             `json:"executed_instructions"`
	OffGraph  int             `json:"off_graph_executed"`
	ByBlock   []blockCoverage `json:"by_block"`
	Clusters  []cluster       `json:"largest_uncovered_clusters"`
}

func readMember(reader *zip.Reader, name string) ([]byte, error) {
	for _, file := range reader.File {
		if strings.EqualFold(file.Name, name) {
			handle, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer handle.Close()
			return io.ReadAll(handle)
		}
	}
	return nil, fmt.Errorf("%s is not in the archive", name)
}

func main() {
	archive := flag.String("archive", "curseoftheazurebonds.zip", tooltext.Text("ecl_run_coverage.usage_archive"))
	coverage := flag.String("coverage", "workplace/ecl-run-coverage.lines", tooltext.Text("ecl_run_coverage.usage_coverage"))
	jsonPath := flag.String("json", "docs/audit/ecl-run-coverage.json", tooltext.Text("ecl_run_coverage.usage_json"))
	mdPath := flag.String("md", "docs/audit/ecl-run-coverage.md", tooltext.Text("ecl_run_coverage.usage_md"))
	top := flag.Int("top", 20, tooltext.Text("ecl_run_coverage.usage_top"))
	flag.Parse()

	reader, err := zip.OpenReader(*archive)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// 每個 block 的可達指令位移（排序過）。記錄器只記 block ID，所以 ID 撞號
	// 會讓兩段分不開——載入時就擋下來。
	labels := map[int]string{}
	reachable := map[int][]int{}
	for _, member := range eclMembers {
		data, err := readMember(&reader.Reader, member)
		if err != nil {
			log.Fatal(err)
		}
		parsed, err := dax.Parse(data)
		if err != nil {
			log.Fatalf("%s: %v", member, err)
		}
		for _, raw := range parsed {
			id := int(raw.Entry.ID)
			label := fmt.Sprintf("%s/0x%02X", member, raw.Entry.ID)
			if existing, dup := labels[id]; dup {
				log.Fatalf(tooltext.Text("ecl_run_coverage.duplicate_block"), id, existing, label)
			}
			labels[id] = label
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
				if !seen[instruction.Offset] {
					seen[instruction.Offset] = true
					reachable[id] = append(reachable[id], instruction.Offset)
				}
			}
			sort.Ints(reachable[id])
		}
	}

	executed := map[int]map[int]bool{}
	total := 0
	for _, path := range strings.Split(*coverage, ",") {
		file, err := os.Open(strings.TrimSpace(path))
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		line := 0
		for scanner.Scan() {
			line++
			var id, offset int
			if _, err := fmt.Sscanf(scanner.Text(), "0x%02X,0x%04X", &id, &offset); err != nil {
				log.Fatalf(tooltext.Text("ecl_run_coverage.bad_line"), path, line, scanner.Text())
			}
			if executed[id] == nil {
				executed[id] = map[int]bool{}
			}
			if !executed[id][offset] {
				executed[id][offset] = true
				total++
			}
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	// 正對照（02-query-returned-empty）：分子整包是空的代表輸入壞了，
	// 不是「覆蓋 0%」這個結論。
	if total == 0 {
		log.Fatal(tooltext.Text("ecl_run_coverage.no_coverage"))
	}

	result := report{Schema: "coab-ecl-run-coverage/1"}
	ids := make([]int, 0, len(labels))
	for id := range labels {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	var clusters []cluster
	offGraphRows := [][2]string{}
	for _, id := range ids {
		row := blockCoverage{Label: labels[id], Reachable: len(reachable[id])}
		reach := map[int]bool{}
		for _, offset := range reachable[id] {
			reach[offset] = true
		}
		for offset := range executed[id] {
			if reach[offset] {
				row.Executed++
			} else {
				result.OffGraph++
				offGraphRows = append(offGraphRows, [2]string{labels[id],
					fmt.Sprintf("0x%04X", ecl.CodeAddressBase+offset)})
			}
		}
		result.Reachable += row.Reachable
		result.Executed += row.Executed
		result.ByBlock = append(result.ByBlock, row)
		// 叢集＝可達指令依位址排序後，連續沒被執行的一段。
		run := 0
		start := 0
		flush := func() {
			if run > 0 {
				clusters = append(clusters, cluster{Label: labels[id],
					Address: fmt.Sprintf("0x%04X", ecl.CodeAddressBase+start), Size: run})
			}
			run = 0
		}
		for _, offset := range reachable[id] {
			if executed[id][offset] {
				flush()
				continue
			}
			if run == 0 {
				start = offset
			}
			run++
		}
		flush()
	}
	sort.Slice(clusters, func(i, j int) bool {
		if clusters[i].Size != clusters[j].Size {
			return clusters[i].Size > clusters[j].Size
		}
		if clusters[i].Label != clusters[j].Label {
			return clusters[i].Label < clusters[j].Label
		}
		return clusters[i].Address < clusters[j].Address
	})
	if len(clusters) > *top {
		clusters = clusters[:*top]
	}
	result.Clusters = clusters

	encoded, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*jsonPath, append(encoded, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}

	var md strings.Builder
	md.WriteString(tooltext.Text("ecl_run_coverage.md_header"))
	md.WriteString(tooltext.Text("ecl_run_coverage.md_inputs"))
	md.WriteString(tooltext.Text("ecl_run_coverage.md_table_header"))
	for _, row := range result.ByBlock {
		percent := 0.0
		if row.Reachable > 0 {
			percent = float64(row.Executed) * 100 / float64(row.Reachable)
		}
		fmt.Fprintf(&md, "| `%s` | %d | %d | %.1f%% |\n", row.Label, row.Reachable, row.Executed, percent)
	}
	fmt.Fprintf(&md, tooltext.Text("ecl_run_coverage.md_total_row"),
		result.Reachable, result.Executed, float64(result.Executed)*100/float64(result.Reachable))
	fmt.Fprintf(&md, tooltext.Text("ecl_run_coverage.md_clusters_header"), *top)
	for _, entry := range result.Clusters {
		fmt.Fprintf(&md, "| `%s` | `%s` | %d |\n", entry.Label, entry.Address, entry.Size)
	}
	fmt.Fprintf(&md, tooltext.Text("ecl_run_coverage.md_offgraph"), result.OffGraph)
	if len(offGraphRows) > 0 {
		md.WriteString(tooltext.Text("ecl_run_coverage.md_offgraph_rows"))
		for _, row := range offGraphRows {
			fmt.Fprintf(&md, "| `%s` | `%s` |\n", row[0], row[1])
		}
	}
	if err := os.WriteFile(*mdPath, []byte(md.String()), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf(tooltext.Text("ecl_run_coverage.summary"),
		result.Reachable, result.Executed,
		float64(result.Executed)*100/float64(result.Reachable), result.OffGraph)
}
