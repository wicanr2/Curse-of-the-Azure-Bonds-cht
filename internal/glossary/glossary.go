// Package glossary enforces one rule: a proper noun from the original game has
// exactly one Traditional-Chinese rendering across every string the player can
// see.
//
// 譯名散在三份 JSON 裡，沒有任何機制擋住漂移；建表時實測到十三組同名異譯
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
	"regexp"
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

// TablePath is the authored table.
const TablePath = "docs/knowledge/coab-glossary.md"

// exceptionHeading is the table whose rows waive missingRendering for one
// message. The parser switches on the heading, so renaming it here and in the
// document must happen together.
const exceptionHeading = "例外：譯文可以不出現的名字"

// exception waives one (message, term) pair.
type exception struct {
	MessageID string
	Source    string
}

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
	terms, exceptions, err := parseTable(filepath.Join(root, TablePath))
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
	english, err := readNested(filepath.Join(root, "gamepack", "pack", "20-locale.en.json"),
		[]string{"locales", "en"})
	if err != nil {
		return Report{}, err
	}
	chinese, err := readNested(filepath.Join(root, "gamepack", "pack", "20-locale.zh-TW.json"),
		[]string{"locales", "zh-TW"})
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
	report.Issues = append(audit(report.Terms, items),
		auditRenderings(report.Terms, english, chinese, exceptions)...)
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

// parseTable reads the `| 原文 | 繁中 | 禁用寫法 | 備註 |` rows plus the
// exception table. The category is the heading the row sits under, so the table
// never repeats it per row.
func parseTable(path string) ([]Term, []exception, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var terms []Term
	var exceptions []exception
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
		if len(fields) < 3 {
			continue
		}
		first := strings.TrimSpace(fields[0])
		if !strings.HasPrefix(first, "`") || !strings.HasSuffix(first, "`") {
			// 表頭、分隔列與「通用寫法」那張沒有原文欄的表都走這裡。
			continue
		}
		if category == exceptionHeading {
			exceptions = append(exceptions, exception{
				MessageID: strings.Trim(first, "`"),
				Source:    strings.Trim(strings.TrimSpace(fields[1]), "`"),
			})
			continue
		}
		if len(fields) < 4 {
			continue
		}
		terms = append(terms, Term{
			Source:    strings.Trim(first, "`"),
			Chinese:   strings.TrimSpace(fields[1]),
			Category:  category,
			Forbidden: splitForbidden(fields[2]),
			Note:      strings.TrimSpace(fields[3]),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	if len(terms) == 0 {
		return nil, nil, fmt.Errorf("%s produced no rows; the table format changed", path)
	}
	return terms, exceptions, nil
}

// auditRenderings catches the half that a forbidden-string list cannot: the
// English gloss names a term and the Chinese simply does not carry it.
//
// 這一類是**誤譯**而不是變體，字串比對本身看不出來。實測 45 個詞條只有 6 筆
// 命中，其中 4 筆是真的（Tyranthraxus 寫成泰蘭索斯、Fire Knives 寫成火焰匕首…），
// 訊噪比夠高才把它做成閘；剩下兩筆合理的省略寫進例外表，而不是放寬規則。
func auditRenderings(terms []Term, english, chinese map[string]string, exceptions []exception) []Issue {
	waived := map[exception]bool{}
	used := map[exception]bool{}
	for _, item := range exceptions {
		waived[item] = true
	}
	var issues []Issue
	for _, term := range terms {
		if term.Imported {
			// 怪物名來自 MON record 的原始字串，英文釋義不一定會提到它。
			continue
		}
		// 大小寫敏感是刻意的：英文釋義裡專名一律大寫開頭，普通名詞不是。
		// `Silk` 若不分大小寫，會撞上巫師塔臥房的 `red silk tapestries`——
		// 那是誤報，而把誤報寫進例外表會讓例外表失去意義。代價是漏檢，不是誤報。
		// 尾巴容許一個複數 `s`：英文釋義會寫 `Red Plumes`、`the Harpers`，
		// 而繁中沒有單複數之分，同一個詞條要蓋住兩者。
		pattern, err := regexp.Compile(`(?:^|[^A-Za-z])` + regexp.QuoteMeta(term.Source) + `s?(?:[^A-Za-z]|$)`)
		if err != nil {
			issues = append(issues, Issue{Code: "bad_term", Term: term.Source, Detail: err.Error()})
			continue
		}
		for key, value := range english {
			if !pattern.MatchString(value) {
				continue
			}
			translated, ok := chinese[key]
			if !ok || strings.Contains(translated, term.Chinese) {
				continue
			}
			item := exception{MessageID: key, Source: term.Source}
			if waived[item] {
				used[item] = true
				continue
			}
			issues = append(issues, Issue{Code: "missing_rendering", Term: term.Source,
				Detail: fmt.Sprintf("%s 的英文釋義提到 %s，繁中卻沒有 %q",
					key, term.Source, term.Chinese)})
		}
	}
	for item := range waived {
		if !used[item] {
			issues = append(issues, Issue{Code: "stale_exception", Term: item.Source,
				Detail: fmt.Sprintf("例外 %s／%s 已經不成立，從表裡刪掉", item.MessageID, item.Source)})
		}
	}
	return issues
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
