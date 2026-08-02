package sourceaudit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanGoHanLiteralsUsesASTAndIgnoresTestsCommentsAndData(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "internal/game/main.go", `package game
// 中文註解不是 player-visible literal.
const top = "中文"
func story() { _ = "中文"; _ = "English" }
`)
	writeFixture(t, root, "internal/game/main_test.go", `package game
const ignoredTest = "測試中文"
`)
	writeFixture(t, root, "gamepack/data.json", `{"text":"資料中文"}`)
	writeFixture(t, root, "workplace/probe.go", `package probe
const ignoredResearch = "研究中文"
`)
	findings, err := ScanGoHanLiterals(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Fatalf("findings=%+v", findings)
	}
	for _, finding := range findings {
		if finding.Path != "internal/game/main.go" || finding.Occurrences != 1 {
			t.Fatalf("finding=%+v", finding)
		}
	}
}

func TestCompareRequiresExactReducedBaseline(t *testing.T) {
	base := []Finding{{Path: "a.go", Function: "f", SHA256: "one", Occurrences: 2, Category: "runtime_ui_debt"}}
	if added, removed := Compare(base, base); len(added) != 0 || len(removed) != 0 {
		t.Fatalf("unchanged added=%v removed=%v", added, removed)
	}
	reduced := []Finding{{Path: "a.go", Function: "f", SHA256: "one", Occurrences: 1, Category: "runtime_ui_debt"}}
	added, removed := Compare(reduced, base)
	if len(added) != 1 || len(removed) != 1 {
		t.Fatalf("reduced baseline must drift until refreshed: added=%v removed=%v", added, removed)
	}
}

func TestRepositoryGoHanLiteralBaselineIsExact(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	current, err := ScanGoHanLiterals(root)
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := LoadBaseline(filepath.Join(root, "docs", "audit", "go-han-literals-baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	added, removed := Compare(current, baseline.Findings)
	if len(added) != 0 || len(removed) != 0 {
		t.Fatalf("Go Han literal baseline drift: new=%v removed=%v", added, removed)
	}
}

func writeFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
