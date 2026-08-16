package glossary

import (
	"path/filepath"
	"testing"
)

const repoRoot = "../.."

// 譯名一致性是 fail-closed 的：新寫的字串用了禁用寫法、或譯名表與
// combatant_name_rules 對不上，這一支就紅。
func TestGlossaryHasNoDrift(t *testing.T) {
	report, err := Run(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range report.Issues {
		t.Errorf("%s: %s", issue.Code, issue.Detail)
	}
	if len(report.Terms) == 0 {
		t.Fatal("glossary produced no terms")
	}
}

// 表本身也要能被稽核：每個詞條都要有原文、繁中與類別，禁用寫法不可以是另一個
// 詞條譯名的一部分（那種規則永遠無法滿足，而且會在無關的字串上誤報）。
func TestGlossaryTableIsWellFormed(t *testing.T) {
	terms, _, err := parseTable(filepath.Join(repoRoot, TablePath))
	if err != nil {
		t.Fatal(err)
	}
	for _, term := range terms {
		if term.Source == "" || term.Chinese == "" || term.Category == "" {
			t.Errorf("incomplete row: %+v", term)
		}
	}
	for _, term := range terms {
		for _, variant := range term.Forbidden {
			if variant == term.Chinese {
				t.Errorf("%s bans its own rendering %q", term.Source, variant)
			}
		}
	}
}

// 掃描範圍不可以悄悄縮水。少掃一份繁中目錄，稽核仍會綠，但漂移擋不住。
func TestGlossaryScansEveryChineseCatalog(t *testing.T) {
	report, err := Run(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		filepath.Join("gamepack", "pack", "20-locale.zh-TW.json"):       false,
		filepath.Join("assets", "locale", "zh-TW.json"):                 false,
		filepath.Join("internal", "tooltext", "messages", "zh-TW.json"): false,
	}
	for _, source := range report.Sources {
		if _, ok := want[source.Path]; !ok {
			t.Errorf("unexpected scan target %q", source.Path)
			continue
		}
		if source.Count == 0 {
			t.Errorf("%s scanned zero strings", source.Path)
		}
		want[source.Path] = true
	}
	for path, seen := range want {
		if !seen {
			t.Errorf("%s is no longer scanned", path)
		}
	}
}

// 怪物名不寫進表，而是從 combatant_name_rules 匯入。這一支釘住那條路徑還在——
// 匯入斷掉的話跨檔交叉檢查會靜默失效。
func TestCombatantNamesAreImported(t *testing.T) {
	imported, err := importCombatantNames(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(imported) == 0 {
		t.Fatal("no combatant name was imported")
	}
	for _, term := range imported {
		if !term.Imported || term.Chinese == "" {
			t.Errorf("bad imported term: %+v", term)
		}
	}
}
