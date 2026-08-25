// Command re-ledger 產生「全函式覆蓋台帳」：把兩平台 IDA 全掃結果、Borland
// 符號與人工判定合併成可重生的索引。
//
// 設計原則：**狀態只來自明確記錄**。本工具不會因為某份文件提到某個位址就把
// 函式標成已解讀；文件命中只當 `citations` 提示欄。真正的狀態一律讀
// `docs/audit/re-function-ledger.json`，預設 `待解讀`。這樣「還剩多少沒看」
// 才是可信數字，而不是關鍵字比對的副產品。
//
// 用法：
//
//	re-ledger -sweep workplace/re-sweep -docs docs \
//	  -ledger docs/audit/re-function-ledger.json \
//	  -out-json docs/audit/coab-function-index.json \
//	  -out-md docs/audit/coab-function-index.md \
//	  -out-detail docs/audit/function-index
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ---- IDA 匯出（tools/ida/export_module.py 的 schema）----

type idaFunction struct {
	EA           int    `json:"ea"`
	Name         string `json:"name"`
	AutoNamed    bool   `json:"auto_named"`
	Size         int    `json:"size"`
	Instructions int    `json:"instructions"`
	IsLib        bool   `json:"is_lib"`
	IsThunk      bool   `json:"is_thunk"`
	RetBytes     int    `json:"ret_bytes"`
	Calls        []int  `json:"calls"`
	Callers      []int  `json:"callers"`
}

type idaSegment struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type idaOverlay struct {
	Module     string `json:"module"`
	Index      int    `json:"index"`
	CodeSize   int    `json:"code_size"`
	CodeSHA256 string `json:"code_sha256"`
	EntryCount int    `json:"entry_count"`
	Seeded     []struct {
		CodeOffset int    `json:"code_offset"`
		Origin     string `json:"origin"`
	} `json:"seeded_entries"`
}

type idaModule struct {
	Input struct {
		Name string `json:"name"`
		Hash struct {
			Value string `json:"value"`
		} `json:"hash"`
	} `json:"input"`
	Segments  []idaSegment    `json:"segments"`
	Functions []idaFunction   `json:"functions"`
	Overlay   *idaOverlay     `json:"overlay"`
	Undefined []idaSegment    `json:"undefined_ranges"`
	TotalsRaw json.RawMessage `json:"totals"`
}

// ---- Borland 符號 ----

type borlandSymbols struct {
	Symbols []struct {
		Name   string `json:"name"`
		Offset int    `json:"offset"`
		Module string `json:"overlay_module"`
		InCode bool   `json:"overlay_code_offset"`
		Flags  int    `json:"flags"`
	} `json:"symbols"`
}

// ---- 人工判定台帳 ----

type ledgerEntry struct {
	Platform string `json:"platform"`
	Module   string `json:"module"`
	EA       int    `json:"ea"`
	State    string `json:"state"` // 已解讀／不阻塞／待解讀
	Level    string `json:"level,omitempty"`
	Spec     string `json:"spec,omitempty"`
	Note     string `json:"note,omitempty"`
}

type ledgerRule struct {
	Platform string `json:"platform"` // 空字串 = 兩平台
	Module   string `json:"module"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
}

type ledgerFile struct {
	Schema      string            `json:"schema"`
	ModuleUnits map[string]string `json:"module_units"`
	Rules       []ledgerRule      `json:"module_rules"`
	Entries     []ledgerEntry     `json:"functions"`
}

// ---- 輸出 ----

type outFunction struct {
	Platform  string   `json:"platform"`
	Module    string   `json:"module"`
	Unit      string   `json:"unit,omitempty"`
	EA        int      `json:"ea"`
	IDAName   string   `json:"ida_name"`
	Symbol    string   `json:"borland_symbol,omitempty"`
	Size      int      `json:"size"`
	Instr     int      `json:"instructions"`
	Callers   int      `json:"callers"`
	Calls     int      `json:"calls"`
	IsEntry   bool     `json:"is_overlay_entry"`
	State     string   `json:"state"`
	Level     string   `json:"level,omitempty"`
	Spec      string   `json:"spec,omitempty"`
	Note      string   `json:"note,omitempty"`
	Citations []string `json:"citations,omitempty"`
}

type moduleStat struct {
	Platform, Module, Unit                       string
	Functions, Read, NotBlocking, Todo, Fragment int
	CodeBytes, Undefined, SegmentBytes           int
	Instructions                                 int
	SHA256                                       string
}

func die(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "re-ledger: %v\n", err)
		os.Exit(1)
	}
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// docCitations 掃描 docs/ 底下的 Markdown，找出「同時提到某 overlay module 與
// 某個十六進位值」的檔案。這只是提示：命中不代表該函式已被解讀，未命中也不代表
// 沒人寫過（位址可能以 segment:offset 或其他位址空間表示）。
func docCitations(root string) (map[string]map[int][]string, error) {
	moduleRE := regexp.MustCompile(`overlay-(\d{2})`)
	hexRE := regexp.MustCompile(`\b(?:0x([0-9A-Fa-f]{1,5})|([0-9A-Fa-f]{2,5})h)\b`)
	result := map[string]map[int][]string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		modules := map[string]bool{}
		for _, m := range moduleRE.FindAllStringSubmatch(text, -1) {
			modules["overlay-"+m[1]] = true
		}
		if len(modules) == 0 {
			return nil
		}
		values := map[int]bool{}
		for _, m := range hexRE.FindAllStringSubmatch(text, -1) {
			token := m[1]
			if token == "" {
				token = m[2]
			}
			if value, err := strconv.ParseInt(token, 16, 64); err == nil {
				values[int(value)] = true
			}
		}
		rel, _ := filepath.Rel(root, path)
		for module := range modules {
			if result[module] == nil {
				result[module] = map[int][]string{}
			}
			for value := range values {
				result[module][value] = append(result[module][value], rel)
			}
		}
		return nil
	})
	return result, err
}

func main() {
	sweepDir := flag.String("sweep", "workplace/re-sweep", tooltext.Text("h.c8c406146286"))
	docsDir := flag.String("docs", "docs", tooltext.Text("h.deff32f5f98a"))
	ledgerPath := flag.String("ledger", "docs/audit/re-function-ledger.json", tooltext.Text("h.8538b0861458"))
	outJSON := flag.String("out-json", "docs/audit/coab-function-index.json", "")
	outMD := flag.String("out-md", "docs/audit/coab-function-index.md", "")
	outDetail := flag.String("out-detail", "docs/audit/function-index", tooltext.Text("h.8692b60fe460"))
	flag.Parse()

	var ledger ledgerFile
	if err := readJSON(*ledgerPath, &ledger); err != nil {
		if !os.IsNotExist(err) {
			die(err)
		}
		ledger = ledgerFile{Schema: "coab-re-function-ledger/1",
			ModuleUnits: map[string]string{}}
	}
	if ledger.ModuleUnits == nil {
		ledger.ModuleUnits = map[string]string{}
	}

	manualByKey := map[string]ledgerEntry{}
	for _, entry := range ledger.Entries {
		manualByKey[fmt.Sprintf("%s/%s/%d", entry.Platform, entry.Module, entry.EA)] = entry
	}
	// 人工台帳是按 (平台, 模組, 位址) 對上函式的。IDA 重掃後函式邊界會變，
	// 對不上的那幾列就再也不會被讀到——而它們掛的「已解讀」仍留在 JSON 裡，
	// 於是完成度被高估，且完全靜默。逐列記錄有沒有被用到，收尾時報出來。
	manualUsed := map[string]bool{}
	rulesByKey := map[string]ledgerRule{}
	for _, rule := range ledger.Rules {
		rulesByKey[rule.Platform+"/"+rule.Module] = rule
	}

	citations, err := docCitations(*docsDir)
	die(err)

	var functions []outFunction
	var stats []moduleStat

	for _, platform := range []string{"dos", "pc98"} {
		symbolByModule := map[string]map[int]string{}
		var symbols borlandSymbols
		if err := readJSON(filepath.Join(*sweepDir, platform, "borland-symbols.json"), &symbols); err == nil {
			for _, symbol := range symbols.Symbols {
				if !symbol.InCode || symbol.Module == "" {
					continue
				}
				if symbolByModule[symbol.Module] == nil {
					symbolByModule[symbol.Module] = map[int]string{}
				}
				// 同一位址可能有多個名字（別名）；保留第一個並附加其餘。
				if existing, ok := symbolByModule[symbol.Module][symbol.Offset]; ok {
					if !strings.Contains(existing, symbol.Name) {
						symbolByModule[symbol.Module][symbol.Offset] = existing + "／" + symbol.Name
					}
					continue
				}
				symbolByModule[symbol.Module][symbol.Offset] = symbol.Name
			}
		}

		paths, err := filepath.Glob(filepath.Join(*sweepDir, platform, "out", "*.json"))
		die(err)
		sort.Strings(paths)
		for _, path := range paths {
			if strings.HasSuffix(path, ".error.log") {
				continue
			}
			var module idaModule
			if err := readJSON(path, &module); err != nil {
				die(fmt.Errorf("%s: %w", path, err))
			}
			name := module.Input.Name
			if module.Overlay != nil {
				name = module.Overlay.Module
			}
			unit := ledger.ModuleUnits[platform+"/"+name]
			if unit == "" {
				unit = ledger.ModuleUnits[name]
			}

			entrySeeds := map[int]bool{}
			if module.Overlay != nil {
				for _, seed := range module.Overlay.Seeded {
					if seed.Origin == "" {
						entrySeeds[seed.CodeOffset] = true
					}
				}
			}

			stat := moduleStat{Platform: platform, Module: name, Unit: unit,
				SHA256: module.Input.Hash.Value}
			for _, segment := range module.Segments {
				stat.SegmentBytes += segment.End - segment.Start
			}
			for _, undefined := range module.Undefined {
				stat.Undefined += undefined.End - undefined.Start
			}

			for _, function := range module.Functions {
				record := outFunction{
					Platform: platform, Module: name, Unit: unit,
					EA: function.EA, IDAName: function.Name, Size: function.Size,
					Instr: function.Instructions, Callers: len(function.Callers),
					Calls: len(function.Calls), IsEntry: entrySeeds[function.EA],
					Symbol: symbolByModule[name][function.EA],
					State:  tooltext.Text("h.5f65ebe71610"),
				}
				if rule, ok := rulesByKey[platform+"/"+name]; ok {
					record.State, record.Note = rule.State, rule.Reason
				} else if rule, ok := rulesByKey["/"+name]; ok {
					record.State, record.Note = rule.State, rule.Reason
				}
				manualKey := fmt.Sprintf("%s/%s/%d", platform, name, function.EA)
				if manual, ok := manualByKey[manualKey]; ok {
					record.State, record.Level = manual.State, manual.Level
					record.Spec, record.Note = manual.Spec, manual.Note
					manualUsed[manualKey] = true
				}
				if hits := citations[name][function.EA]; len(hits) > 0 {
					sort.Strings(hits)
					record.Citations = uniqueStrings(hits)
				}
				switch record.State {
				case tooltext.Text("h.4e0bb7d0bf17"):
					stat.Read++
				case tooltext.Text("h.0c629c21d625"):
					stat.NotBlocking++
				case tooltext.Text("h.6b1ebedfaf86"):
					// IDA 在 raw overlay 上建錯的函式邊界；不是真的函式，
					// 單獨計數才不會讓分母與待解讀數失真。
					stat.Fragment++
				default:
					stat.Todo++
				}
				stat.Functions++
				stat.CodeBytes += function.Size
				stat.Instructions += function.Instructions
				functions = append(functions, record)
			}
			stats = append(stats, stat)
		}
	}

	die(os.MkdirAll(filepath.Dir(*outJSON), 0o755))
	payload := map[string]any{
		"schema":    "coab-function-index/1",
		"functions": functions,
		"modules":   stats,
	}
	encoded, err := json.MarshalIndent(payload, "", " ")
	die(err)
	die(os.WriteFile(*outJSON, append(encoded, '\n'), 0o644))

	orphans := make([]ledgerEntry, 0)
	for _, entry := range ledger.Entries {
		key := fmt.Sprintf("%s/%s/%d", entry.Platform, entry.Module, entry.EA)
		if !manualUsed[key] {
			orphans = append(orphans, entry)
		}
	}
	sort.Slice(orphans, func(i, j int) bool {
		if orphans[i].Platform != orphans[j].Platform {
			return orphans[i].Platform < orphans[j].Platform
		}
		if orphans[i].Module != orphans[j].Module {
			return orphans[i].Module < orphans[j].Module
		}
		return orphans[i].EA < orphans[j].EA
	})

	die(os.MkdirAll(*outDetail, 0o755))
	writeMarkdown(*outMD, *outDetail, stats, functions, orphans)

	var total, read, notBlocking, todo, fragment int
	for _, stat := range stats {
		total += stat.Functions
		read += stat.Read
		notBlocking += stat.NotBlocking
		todo += stat.Todo
		fragment += stat.Fragment
	}
	orphanRead := 0
	for _, entry := range orphans {
		if entry.State == tooltext.Text("h.4e0bb7d0bf17") {
			orphanRead++
		}
	}
	fmt.Fprint(os.Stderr, tooltext.Format("h.d38280e35b6c", total, read, notBlocking, fragment, todo, *outMD))
	fmt.Fprint(os.Stderr, tooltext.Format("h.8cb28ebf9135", len(orphans), orphanRead))
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) > 6 {
		out = out[:6]
	}
	return out
}

func writeMarkdown(indexPath, detailDir string, stats []moduleStat, functions []outFunction, orphans []ledgerEntry) {
	var builder strings.Builder
	builder.WriteString(tooltext.Text("h.7b50aa6a444f"))
	builder.WriteString(tooltext.Text("h.9a505e811528"))
	builder.WriteString(tooltext.Text("h.404652a093e8"))
	builder.WriteString(tooltext.Text("h.1bb9aed8b116"))
	builder.WriteString(tooltext.Text("h.be8a73fd2a10"))
	builder.WriteString(tooltext.Text("h.d5ada40132b1"))

	byPlatform := map[string][]int{}
	for index, stat := range stats {
		byPlatform[stat.Platform] = append(byPlatform[stat.Platform], index)
	}
	for _, platform := range []string{"dos", "pc98"} {
		indexes := byPlatform[platform]
		if len(indexes) == 0 {
			continue
		}
		var functionCount, read, notBlocking, todo, fragment, code, undefined int
		for _, i := range indexes {
			functionCount += stats[i].Functions
			read += stats[i].Read
			notBlocking += stats[i].NotBlocking
			todo += stats[i].Todo
			fragment += stats[i].Fragment
			code += stats[i].CodeBytes
			undefined += stats[i].Undefined
		}
		fmt.Fprintf(&builder, "## %s\n\n", strings.ToUpper(platform))
		fmt.Fprintf(&builder, tooltext.Text("h.89e98f0623d2")+
			tooltext.Text("h.8c3e88ac01e2"),
			len(indexes), functionCount, read, notBlocking, fragment, todo, code, undefined)
		builder.WriteString(tooltext.Text("h.0e3bf3837130"))
		builder.WriteString("|---|---|---:|---:|---:|---:|---:|---:|---:|---|\n")
		for _, i := range indexes {
			stat := stats[i]
			detail := fmt.Sprintf("%s-%s.md", stat.Platform, stat.Module)
			fmt.Fprint(&builder, tooltext.Format("h.93eb52077cff", stat.Module, orDash(stat.Unit), stat.Functions, stat.Read, stat.NotBlocking, stat.Fragment, stat.Todo, stat.CodeBytes, stat.Undefined, detail))
		}
		builder.WriteString("\n")
	}
	if len(orphans) > 0 {
		orphanRead := 0
		for _, entry := range orphans {
			if entry.State == tooltext.Text("h.4e0bb7d0bf17") {
				orphanRead++
			}
		}
		builder.WriteString(tooltext.Text("h.4b2a4d9681d5"))
		fmt.Fprintf(&builder, tooltext.Text("h.f871566c85fb")+
			tooltext.Text("h.df8e06ebd4ae"), len(orphans), orphanRead)
		builder.WriteString(tooltext.Text("h.d5413c310ba6"))
		builder.WriteString(tooltext.Text("h.5f724c4e4ddf"))
		builder.WriteString(tooltext.Text("h.dbe524438171"))
		builder.WriteString(tooltext.Text("h.ff1c78f63d39"))
		builder.WriteString(tooltext.Text("h.68acfa2b6d4b"))
		builder.WriteString("|---|---|---|---|---|\n")
		for _, entry := range orphans {
			fmt.Fprintf(&builder, "| %s | %s | `%04X` | %s | %s |\n",
				entry.Platform, entry.Module, entry.EA, entry.State, orDash(entry.Spec))
		}
		builder.WriteString("\n")
	}

	// 檔尾固定一個換行，重跑才會得到位元組相同的產物。
	die(os.WriteFile(indexPath,
		[]byte(strings.TrimRight(builder.String(), "\n")+"\n"), 0o644))

	perModule := map[string][]outFunction{}
	for _, function := range functions {
		key := function.Platform + "-" + function.Module
		perModule[key] = append(perModule[key], function)
	}
	for key, list := range perModule {
		sort.Slice(list, func(i, j int) bool { return list[i].EA < list[j].EA })
		var detail strings.Builder
		fmt.Fprint(&detail, tooltext.Format("h.ca8817ca2e95", key))
		detail.WriteString(tooltext.Text("h.4ceffa855ba3"))
		detail.WriteString(tooltext.Text("h.08c576e8d99a"))
		detail.WriteString(tooltext.Text("h.7ebe1e90de31"))
		detail.WriteString("|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|\n")
		for _, function := range list {
			entry := ""
			if function.IsEntry {
				entry = "✓"
			}
			// 規格與理由都要出現：只留其一會讓 opcode binding 這類註記消失。
			parts := []string{}
			if function.Spec != "" {
				parts = append(parts, function.Spec)
			}
			if function.Note != "" {
				parts = append(parts, function.Note)
			}
			note := strings.Join(parts, "<br>")
			fmt.Fprintf(&detail, "| `%04X` | %s | %s | %d | %d | %d | %d | %s | %s | %s | %s | %s |\n",
				function.EA, function.IDAName, orDash(function.Symbol), function.Size,
				function.Instr, function.Callers, function.Calls, entry,
				function.State, orDash(function.Level), orDash(note),
				orDash(strings.Join(function.Citations, "<br>")))
		}
		die(os.WriteFile(filepath.Join(detailDir, key+".md"),
			[]byte(strings.TrimRight(detail.String(), "\n")+"\n"), 0o644))
	}
}

func orDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}
