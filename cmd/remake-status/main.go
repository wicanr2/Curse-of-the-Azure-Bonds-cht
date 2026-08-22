// Command remake-status 把 `docs/audit/` 底下**已經由工具產生**的覆蓋率報表
// 併成一張表：每個大類目前量到什麼數字、數字出自哪一份、以及**那個數字證明不了
// 什麼**。
//
// ★ 存在的理由：`CONTEXT.md` 與 `WORKLIST.md` 的「尚未完成」是手寫的，而**沒有
// 任何東西會在做完那一刻更新它**——完成的 commit 改的是程式碼與那一輪的紀錄，
// 散在別處的現在式句子照樣留著。第 664 輪抽查兩條就抓到兩條假的（手札 59 的
// renderer 早就做完、`+11Ch` 的讀法早就對完）。這一支把「還剩多少」換成可重生的
// 數字，讓那份手寫清單降級成指標。
//
// ⚠ **這張表不是完成度百分比。** 每一列都配一句「證明不了什麼」，因為每個分母
// 的意義都不同：ECL 指令的 `done` 是「這個 opcode 的語意讀完了」，不是「玩家走
// 得到那一條」；漢字 literal 的數字是債務不是進度。把它們平均起來會得到一個
// 沒有意義而且偏樂觀的數字。
//
// ⚠ 量不到的那幾類**要留在表上**，寫成「沒有數字」而不是省略。音訊生命週期、
// 逐張 UI fidelity、開場到結局的同一 session、三平台發行都屬於這一類：省略它們
// 會讓這張表自己變成下一個假的完成宣告。
//
// ★ 表格的中文一律住在 `rows.json`，Go 這一側只留數字的取法與版面。
// `cmd/coab-audit` 的漢字 gate 只准數量下降，而把敘述留在 Go 裡就是新增債務；
// 這也正是那條 gate 想要的方向（文字進資料、Go 留 format contract）。
//
// 用法：
//
//	./tools/go.sh run ./cmd/remake-status -output docs/audit/remake-status.md
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed rows.json
var rowsJSON []byte

// rowSpec 是一列的敘述部分：類別、數字的版面、出處、以及它證明不了什麼。
type rowSpec struct {
	Category string `json:"category"`
	Format   string `json:"format"`
	Source   string `json:"source"`
	NotProof string `json:"not_proof"`
	// Report／Patterns／Fields 給沒有 JSON 版的報表用：從 Markdown 裡撈數字。
	// ⚠ 比對樣式也住在資料裡——它們比對的是中文報表的措辭，措辭改了要跟著改，
	// 而那是資料的事不是程式的事。
	Report   string   `json:"report"`
	Patterns []string `json:"patterns"`
	Fields   []int    `json:"fields"`
}

type rowsFile struct {
	Title           string             `json:"title"`
	Preamble        []string           `json:"preamble"`
	Header          []string           `json:"header"`
	UnmeasuredLabel string             `json:"unmeasured_label"`
	Measured        map[string]rowSpec `json:"measured"`
	Unmeasured      []rowSpec          `json:"unmeasured"`
}

// measure 是輸出的一列。
type measure struct {
	spec    rowSpec
	numbers string
}

func main() {
	auditDir := flag.String("audit", "docs/audit", "directory holding the generated audit reports")
	output := flag.String("output", "", "Markdown output path (empty prints to stdout)")
	flag.Parse()

	var rowsData rowsFile
	if err := json.Unmarshal(rowsJSON, &rowsData); err != nil {
		log.Fatal(err)
	}

	rows := make([]measure, 0, 16)
	add := func(key string, args ...any) {
		spec, ok := rowsData.Measured[key]
		if !ok {
			log.Fatalf("rows.json has no measured row %q", key)
		}
		rows = append(rows, measure{spec: spec, numbers: fmt.Sprintf(spec.Format, args...)})
	}

	var effects struct {
		InstructionsByStatus map[string]int `json:"instructions_by_status"`
		OpcodesByStatus      map[string]int `json:"opcodes_by_status"`
		Reachable            int            `json:"reachable_instructions"`
	}
	if readJSON(*auditDir, "ecl-effect-coverage.json", &effects) {
		add("ecl_effects", effects.Reachable,
			effects.InstructionsByStatus["done"], effects.InstructionsByStatus["partial"],
			effects.OpcodesByStatus["done"], effects.OpcodesByStatus["partial"])
	}

	var text struct {
		Summary struct {
			Groups    int `json:"groups"`
			Matched   int `json:"matched"`
			Unmatched int `json:"unmatched"`
		} `json:"summary"`
	}
	if readJSON(*auditDir, "ecl-text-coverage.json", &text) {
		add("ecl_text", text.Summary.Groups, text.Summary.Matched, text.Summary.Unmatched)
	}

	var save struct {
		RecordSize    int            `json:"record_size"`
		BytesByStatus map[string]int `json:"bytes_by_status"`
		ConsumedBytes int            `json:"consumed_bytes"`
	}
	if readJSON(*auditDir, "dos-save-field-coverage.json", &save) {
		unknown := save.RecordSize - save.BytesByStatus["decoded"] - save.BytesByStatus["documented"]
		add("save_fields", save.RecordSize, save.BytesByStatus["decoded"],
			save.BytesByStatus["documented"], unknown, save.ConsumedBytes)
	}

	var index struct {
		Modules []struct {
			Functions   int
			Read        int
			NotBlocking int
			Todo        int
			Fragment    int
			Undefined   int
		} `json:"modules"`
	}
	if readJSON(*auditDir, "coab-function-index.json", &index) {
		total, read, notBlocking, todo, fragment, undefined := 0, 0, 0, 0, 0, 0
		for _, module := range index.Modules {
			total += module.Functions
			read += module.Read
			notBlocking += module.NotBlocking
			todo += module.Todo
			fragment += module.Fragment
			undefined += module.Undefined
		}
		add("re_ledger", total, read, notBlocking, fragment, todo, undefined)
	}

	var glossary struct {
		Terms  []json.RawMessage `json:"terms"`
		Issues []json.RawMessage `json:"issues"`
	}
	if readJSON(*auditDir, "glossary.json", &glossary) {
		add("glossary", len(glossary.Terms), len(glossary.Issues))
	}

	var han struct {
		Findings []struct {
			Occurrences int    `json:"occurrences"`
			Category    string `json:"category"`
		} `json:"findings"`
	}
	if readJSON(*auditDir, "go-han-literals-baseline.json", &han) {
		byCategory := map[string]int{}
		total := 0
		for _, finding := range han.Findings {
			byCategory[finding.Category] += finding.Occurrences
			total += finding.Occurrences
		}
		// ⚠ 三個分類一律印出來，**0 也要印**。省略一個 0 讀起來像「那一類沒有
		// 問題」，而它其實跟「已經清乾淨了」是同一個字面。
		parts := make([]string, 0, 3)
		for _, key := range []string{"localization_debt", "frontend_ui_debt", "runtime_ui_debt"} {
			parts = append(parts, fmt.Sprintf("`%s` %d", key, byCategory[key]))
		}
		add("han_literals", total, strings.Join(parts, "／"))
	}

	// 這幾份沒有 JSON，摘要在 Markdown 的句子裡；比對樣式在 `rows.json`。
	for _, key := range []string{"charged_items", "cell_sweep", "cell_revisit", "cell_random_branches", "fp_screens", "monster_ai", "table_dispatch", "treasure_sites", "proofread", "audio_triggers", "music_triggers", "sound_triggers", "dos_sound_map", "campaign_frames", "player_input_path"} {
		spec := rowsData.Measured[key]
		fields, ok := grepNumbers(*auditDir, spec.Report, spec.Patterns...)
		if !ok {
			continue
		}
		args := make([]any, 0, len(spec.Fields))
		for _, index := range spec.Fields {
			if index < 0 || index >= len(fields) {
				log.Fatalf("rows.json %s: field %d out of range (%d captured)", key, index, len(fields))
			}
			args = append(args, fields[index])
		}
		add(key, args...)
	}

	for _, spec := range rowsData.Unmeasured {
		rows = append(rows, measure{spec: spec, numbers: rowsData.UnmeasuredLabel})
	}

	var report strings.Builder
	report.WriteString(rowsData.Title)
	for _, line := range rowsData.Preamble {
		fmt.Fprintf(&report, "\n%s\n", line)
	}
	fmt.Fprintf(&report, "\n| %s |\n|---|---|---|---|\n", strings.Join(rowsData.Header, " | "))
	for _, row := range rows {
		source := row.spec.Source
		if source == "" {
			source = "—"
		}
		fmt.Fprintf(&report, "| %s | %s | %s | %s |\n",
			row.spec.Category, row.numbers, source, row.spec.NotProof)
	}

	text2 := report.String()
	if *output == "" {
		fmt.Print(text2)
	} else if err := os.WriteFile(*output, []byte(text2), 0o644); err != nil {
		log.Fatal(err)
	}
	measured := 0
	for _, row := range rows {
		if row.spec.Source != "" {
			measured++
		}
	}
	fmt.Fprintf(os.Stderr, "rows=%d measured=%d unmeasured=%d\n", len(rows), measured, len(rows)-measured)
}

// grepNumbers 從 Markdown 報表裡把數字撈出來（那幾份沒有 JSON 版）。
//
// ⚠ 任何一條 pattern 對不上就**整列不出現**，並往 stderr 報。不要退回填 0——
// 那會把「沒量到」印成「沒有缺口」，正好是這張表存在的理由的反面。
// ⚠ 也不要把整行塞進表格：報表的行裡有 `|`，會把 Markdown 的欄位切開。
func grepNumbers(dir, name string, patterns ...string) ([]string, bool) {
	payload, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "skip %s: %v\n", name, err)
		return nil, false
	}
	out := make([]string, 0, len(patterns)*2)
	for _, pattern := range patterns {
		match := regexp.MustCompile(pattern).FindStringSubmatch(string(payload))
		if match == nil {
			fmt.Fprintf(os.Stderr, "skip %s: pattern %q did not match\n", name, pattern)
			return nil, false
		}
		out = append(out, match[1:]...)
	}
	return out, true
}

// readJSON 讀一份報表。缺檔不是錯誤——報表是手動重生的，缺的那一列直接不出現，
// 但**回報到 stderr**，不要讓「少一列」看起來像「那一類沒有問題」。
func readJSON(dir, name string, into any) bool {
	payload, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "skip %s: %v\n", name, err)
		return false
	}
	if err := json.Unmarshal(payload, into); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	return true
}
