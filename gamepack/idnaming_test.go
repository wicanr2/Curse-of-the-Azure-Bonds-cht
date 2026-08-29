package gamepack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

// idPattern is spec 1105 §四: lowercase, dot-separated namespaces, hyphens
// inside a segment, at least one dot. No underscores.
var idPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*(\.[a-z0-9]+(-[a-z0-9]+)*)+$`)

type idBaselineEntry struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type idBaseline struct {
	Schema  string            `json:"schema"`
	Pattern string            `json:"pattern"`
	Entries []idBaselineEntry `json:"entries"`
}

// The committed pack still holds 155 IDs from before the naming rule existed
// (mostly flat snake_case locale keys). They are listed so they do not fail the
// gate, and so "how many are left" is answerable. A new ID that breaks the rule
// is not in the baseline, so it turns this test red.
func TestGamePackIDNamingBaselineIsExact(t *testing.T) {
	baselinePath := filepath.Join("..", "docs", "audit", "gamepack-id-baseline.json")
	raw, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	var baseline idBaseline
	if err := json.Unmarshal(raw, &baseline); err != nil {
		t.Fatal(err)
	}
	if baseline.Pattern != idPattern.String() {
		t.Fatalf("baseline pattern %q does not match the test's %q",
			baseline.Pattern, idPattern.String())
	}
	known := make(map[idBaselineEntry]bool, len(baseline.Entries))
	for _, entry := range baseline.Entries {
		known[entry] = true
	}

	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[idBaselineEntry]bool{}
	record := func(kind, id string) {
		if idPattern.MatchString(id) {
			return
		}
		seen[idBaselineEntry{Kind: kind, ID: id}] = true
	}
	for _, entries := range pack.Locales {
		for key := range entries {
			record("locale", key)
		}
	}
	for _, rule := range pack.TextRules {
		record("text_rules", rule.ID)
	}
	for _, rule := range pack.OptionRules {
		record("option_rules", rule.ID)
	}
	for _, event := range pack.Events {
		record("events", event.ID)
	}

	var added, removed []string
	for entry := range seen {
		if !known[entry] {
			added = append(added, entry.Kind+" "+entry.ID)
		}
	}
	for entry := range known {
		if !seen[entry] {
			removed = append(removed, entry.Kind+" "+entry.ID)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	if len(added) > 0 {
		t.Fatalf("new IDs break the spec 1105 §四 naming rule: %v\n"+
			"命名是 <群組>.<地點或子系統>.<東西>，全小寫、分隔用 `.`、群組內用 `-`、不用底線。", added)
	}
	if len(removed) > 0 {
		t.Fatalf("baseline lists IDs the pack no longer has: %v\n"+
			"收斂了就把它們從 docs/audit/gamepack-id-baseline.json 刪掉。", removed)
	}
}

// The split must not lose or duplicate content. These counts are what the
// single-file pack held before spec 1105 split it.
func TestDefaultPackMergesAllCommittedParts(t *testing.T) {
	names, err := PartNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 6 {
		t.Fatalf("pack parts=%v, want core, content, and four locale parts", names)
	}
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	// ⚠ 這裡**不能**釘死一個數字。內容產出（`ENG-01`）每補一批劇情就會讓
	// text_rules 變多，釘死的快照會在每一批之後變紅，而它擋不住任何真正的錯誤
	// ——「合併有沒有遺漏或重複」跟「總共幾條」是兩件事。
	// 真正的不變量是：合併後的條數等於各分檔加總，而且沒有重複的 id。
	wantTextRules, wantOptionRules := 0, 0
	seen := map[string]string{}
	for _, name := range names {
		raw, err := Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		var part struct {
			TextRules   []struct{ ID string } `json:"text_rules"`
			OptionRules []struct{ ID string } `json:"option_rules"`
		}
		if err := json.Unmarshal(raw, &part); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		wantTextRules += len(part.TextRules)
		wantOptionRules += len(part.OptionRules)
		for _, rule := range part.TextRules {
			if first, ok := seen[rule.ID]; ok {
				t.Fatalf("text rule %q 同時出現在 %s 與 %s", rule.ID, first, name)
			}
			seen[rule.ID] = name
		}
	}
	if got := len(pack.TextRules); got != wantTextRules {
		t.Fatalf("text_rules=%d, 各分檔加總=%d", got, wantTextRules)
	}
	if got := len(pack.OptionRules); got != wantOptionRules {
		t.Fatalf("option_rules=%d, 各分檔加總=%d", got, wantOptionRules)
	}
	if got := len(pack.Events); got != 4 {
		t.Fatalf("events=%d, want 4", got)
	}
	if got := len(pack.Maps); got != 20 {
		t.Fatalf("maps=%d, want 20", got)
	}
	if pack.CharacterCreation == nil || pack.Presentation == nil || pack.Search == nil {
		t.Fatal("core sections did not survive the merge")
	}
	// 同上：條數會隨內容產出成長，釘死快照只會每一批紅一次。
	// 要釘的是「兩語系的鍵完全對齊」，那才是分成兩個檔之後會漂掉的東西。
	en, zh := len(pack.Locales["en"]), len(pack.Locales["zh-TW"])
	if en != zh {
		t.Fatalf("locales en=%d zh-TW=%d，兩語系的條數必須一致", en, zh)
	}
	if en == 0 {
		t.Fatal("merged pack has no locale entries")
	}
	// 兩語系的 key 必須完全對齊——這是分成兩個檔之後最容易漂掉的東西。
	for key := range pack.Locales["en"] {
		if _, ok := pack.Locales["zh-TW"][key]; !ok {
			t.Fatalf("locale key %q exists in en but not zh-TW", key)
		}
	}
	for key := range pack.Locales["zh-TW"] {
		if _, ok := pack.Locales["en"][key]; !ok {
			t.Fatalf("locale key %q exists in zh-TW but not en", key)
		}
	}
}
