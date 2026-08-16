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
	if len(names) != 4 {
		t.Fatalf("pack parts=%v, want the four parts of spec 1105 §二", names)
	}
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(pack.TextRules); got != 464 {
		t.Fatalf("text_rules=%d, want 464", got)
	}
	if got := len(pack.OptionRules); got != 113 {
		t.Fatalf("option_rules=%d, want 113", got)
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
	en, zh := len(pack.Locales["en"]), len(pack.Locales["zh-TW"])
	if en != 724 || zh != 724 {
		t.Fatalf("locales en=%d zh-TW=%d, want 724/724", en, zh)
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
