// Package glossary enforces one rule: a proper noun from the original game has
// exactly one Traditional-Chinese rendering across every string the player can
// see.
//
// 譯名先前散在三份 JSON 裡，沒有任何機制擋住漂移，實測結果是九組同名異譯
// （Bane 貝恩／班恩、Zhentil Keep 散提爾堡／散塔林堡、Flamed One 三種寫法…）。
// 內容量還會再成長數倍，事後回頭校對比現在建表貴得多。
//
// 表在 `docs/knowledge/coab-glossary.md`，怪物名不進表——它們的唯一來源是 game
// pack 的 `combatant_name_rules`（spec 479），本套件自動併入，不要求人工同步。
package glossary

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Term is one row of the authored table.
type Term struct {
	Source    string   `json:"source"`
	Chinese   string   `json:"chinese"`
	Category  string   `json:"category"`
	Forbidden []string `json:"forbidden,omitempty"`
	Note      string   `json:"note,omitempty"`
	// Imported marks rows taken from combatant_name_rules rather than the table.
	Imported bool `json:"imported,omitempty"`
	// Uses counts how many scanned strings contain Chinese.
	Uses int `json:"uses"`
}

// Issue is a fail-closed finding. Any issue makes the gate red.
type Issue struct {
	Code   string `json:"code"`
	Term   string `json:"term"`
	Detail string `json:"detail"`
}

type Report struct {
	Schema  string          `json:"schema"`
	Sources []stringCatalog `json:"sources"`
	Terms   []Term          `json:"terms"`
	Issues  []Issue         `json:"issues"`
}

type stringCatalog struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// scanned is one player-visible string with enough identity to name it in a
// failure message.
type scanned struct {
	path  string
	key   string
	value string
}

// TablePath is the authored table; ScanPaths are every file whose values reach
// the player in Traditional Chinese.
const TablePath = "docs/knowledge/coab-glossary.md"

var scanTargets = []struct {
	path string
	// nest walks into the object that actually holds the strings. Empty means
	// the document root is already that object.
	nest []string
}{
	{path: filepath.Join("gamepack", "pack", "20-locale.zh-TW.json"), nest: []string{"locales", "zh-TW"}},
	{path: filepath.Join("assets", "locale", "zh-TW.json"), nest: []string{"strings"}},
	{path: filepath.Join("internal", "tooltext", "messages", "zh-TW.json")},
}

// Run parses the table, scans every Traditional-Chinese catalog and returns a
// fail-closed report.
func Run(root string) (Report, error) {
	terms, err := parseTable(filepath.Join(root, TablePath))
	if err != nil {
		return Report{}, err
	}
	imported, err := importCombatantNames(root)
	if err != nil {
		return Report{}, err
	}
	items, catalogs, err := scan(root)
	if err != nil {
		return Report{}, err
	}
	report := Report{Schema: "coab-glossary/1", Sources: catalogs}
	report.Terms = append(append([]Term(nil), terms...), imported...)
	for index := range report.Terms {
		for _, item := range items {
			if strings.Contains(item.value, report.Terms[index].Chinese) {
				report.Terms[index].Uses++
			}
		}
	}
	report.Issues = audit(report.Terms, items)
	sort.Slice(report.Issues, func(i, j int) bool {
		if report.Issues[i].Code != report.Issues[j].Code {
			return report.Issues[i].Code < report.Issues[j].Code
		}
		return report.Issues[i].Detail < report.Issues[j].Detail
	})
	return report, nil
}

func audit(terms []Term, items []scanned) []Issue {
	var issues []Issue
	// 同一個原文可以同時是人工詞條與 combatant 規則（`MOGION` 就是），那不是重複，
	// 是**跨檔交叉檢查的機會**：兩邊的譯名必須一樣。所以分組用大寫折疊後的原文。
	bySource := map[string][]Term{}
	byChinese := map[string]map[string]bool{}
	for _, term := range terms {
		key := strings.ToUpper(term.Source)
		bySource[key] = append(bySource[key], term)
		if byChinese[term.Chinese] == nil {
			byChinese[term.Chinese] = map[string]bool{}
		}
		byChinese[term.Chinese][key] = true
	}
	for source, group := range bySource {
		renderings := map[string]bool{}
		for _, term := range group {
			renderings[term.Chinese] = true
		}
		if len(renderings) > 1 {
			var list []string
			for rendering := range renderings {
				list = append(list, rendering)
			}
			sort.Strings(list)
			issues = append(issues, Issue{Code: "conflicting_rendering", Term: source,
				Detail: fmt.Sprintf("原文 %s 有 %d 種譯名：%s（譯名表與 combatant_name_rules 要對齊）",
					source, len(list), strings.Join(list, "、"))})
		}
	}
	for chinese, sources := range byChinese {
		if len(sources) < 2 {
			continue
		}
		list := make([]string, 0, len(sources))
		for source := range sources {
			list = append(list, source)
		}
		sort.Strings(list)
		issues = append(issues, Issue{Code: "duplicate_rendering", Term: chinese,
			Detail: fmt.Sprintf("繁中 %q 同時是 %s 的譯名", chinese, strings.Join(list, "、"))})
	}
	for _, term := range terms {
		if term.Uses == 0 {
			issues = append(issues, Issue{Code: "unused_term", Term: term.Source,
				Detail: fmt.Sprintf("%s ＝ %q 在任何繁中字串裡都沒出現", term.Source, term.Chinese)})
		}
		for _, variant := range term.Forbidden {
			// 一個禁用寫法若是某個詞條譯名的一部分，這條規則永遠無法滿足。
			for _, other := range terms {
				if strings.Contains(other.Chinese, variant) {
					issues = append(issues, Issue{Code: "unsatisfiable_ban", Term: term.Source,
						Detail: fmt.Sprintf("禁用寫法 %q 是 %s ＝ %q 的一部分",
							variant, other.Source, other.Chinese)})
				}
			}
			for _, item := range items {
				if strings.Contains(item.value, variant) {
					issues = append(issues, Issue{Code: "forbidden_variant", Term: term.Source,
						Detail: fmt.Sprintf("%s 應寫 %q，但 %s 的 %s 出現 %q",
							term.Source, term.Chinese, item.path, item.key, variant)})
				}
			}
		}
	}
	return issues
}

// parseTable reads the `| 原文 | 繁中 | 禁用寫法 | 備註 |` rows. The category is
// the heading the row sits under, so the table never repeats it per row.
func parseTable(path string) ([]Term, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var terms []Term
	category := ""
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if heading, ok := strings.CutPrefix(line, "## "); ok {
			category = strings.TrimSpace(heading)
			continue
		}
		if !strings.HasPrefix(line, "|") {
			continue
		}
		fields := strings.Split(strings.Trim(line, "|"), "|")
		if len(fields) < 4 {
			continue
		}
		source := strings.TrimSpace(fields[0])
		if !strings.HasPrefix(source, "`") || !strings.HasSuffix(source, "`") {
			// 表頭、分隔列與「通用寫法」那張沒有原文欄的表都走這裡。
			continue
		}
		terms = append(terms, Term{
			Source:    strings.Trim(source, "`"),
			Chinese:   strings.TrimSpace(fields[1]),
			Category:  category,
			Forbidden: splitForbidden(fields[2]),
			Note:      strings.TrimSpace(fields[3]),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(terms) == 0 {
		return nil, fmt.Errorf("%s produced no rows; the table format changed", path)
	}
	return terms, nil
}

func splitForbidden(field string) []string {
	field = strings.TrimSpace(field)
	if field == "" || field == "—" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(field, "、") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// importCombatantNames turns the game pack's combatant rules into glossary rows
// so monster names are audited without being copied into the table.
func importCombatantNames(root string) ([]Term, error) {
	raw, err := os.ReadFile(filepath.Join(root, "gamepack", "pack", "00-core.json"))
	if err != nil {
		return nil, err
	}
	var core struct {
		CombatantNameRules []struct {
			Source    string `json:"source"`
			MessageID string `json:"message_id"`
		} `json:"combatant_name_rules"`
	}
	if err := json.Unmarshal(raw, &core); err != nil {
		return nil, err
	}
	messages, err := readNested(filepath.Join(root, "gamepack", "pack", "20-locale.zh-TW.json"),
		[]string{"locales", "zh-TW"})
	if err != nil {
		return nil, err
	}
	var terms []Term
	for _, rule := range core.CombatantNameRules {
		chinese, ok := messages[rule.MessageID]
		if !ok {
			return nil, fmt.Errorf("combatant rule %q points at missing message %q",
				rule.Source, rule.MessageID)
		}
		terms = append(terms, Term{
			Source: rule.Source, Chinese: chinese,
			Category: "怪物（由 combatant_name_rules 匯入）", Imported: true,
		})
	}
	return terms, nil
}

func scan(root string) ([]scanned, []stringCatalog, error) {
	var items []scanned
	var catalogs []stringCatalog
	for _, target := range scanTargets {
		table, err := readNested(filepath.Join(root, target.path), target.nest)
		if err != nil {
			return nil, nil, err
		}
		for key, value := range table {
			items = append(items, scanned{path: target.path, key: key, value: value})
		}
		catalogs = append(catalogs, stringCatalog{Path: target.path, Count: len(table)})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].path != items[j].path {
			return items[i].path < items[j].path
		}
		return items[i].key < items[j].key
	})
	return items, catalogs, nil
}

func readNested(path string, nest []string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	for _, step := range nest {
		object, ok := document.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: %q is not an object", path, step)
		}
		document, ok = object[step]
		if !ok {
			return nil, fmt.Errorf("%s: missing %q", path, step)
		}
	}
	object, ok := document.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: the nested value is not an object", path)
	}
	table := make(map[string]string, len(object))
	for key, value := range object {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%s: %q is not a string", path, key)
		}
		table[key] = text
	}
	return table, nil
}
