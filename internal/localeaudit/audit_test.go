package localeaudit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunKeepsAssetAndGamepackResponsibilitiesSeparate(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, root, "assets/locale/zh-TW.json", map[string]any{
		"language": "zh-TW",
		"strings":  map[string]string{"ui_title": "標題", "unused_ui": "未使用"},
	})
	writeJSON(t, root, "gamepack/events/pit-of-moander.json", map[string]any{
		"locales": map[string]any{
			"en":    map[string]string{"event.one": "Event", "option.one": "YES", "orphan": "Orphan"},
			"zh-TW": map[string]string{"event.one": "事件", "option.one": "是", "orphan": "孤立"},
		},
		"option_rules": []map[string]string{{"id": "option.one", "source": "YES", "message_id": "option.one"}},
		"text_rules":   []map[string]string{{"id": "event.one", "message_id": "event.one"}},
	})
	writeFile(t, root, "internal/game/state.go", `package game
func show(c interface{ Text(string, string) string }) string { return c.Text("ui_title", "Title") }
`)
	report, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	if ViolationCount(report) != 0 {
		t.Fatalf("unexpected violations: %+v", report.Issues)
	}
	if len(report.AssetsLocale.OrphanKeys) != 1 || report.AssetsLocale.OrphanKeys[0] != "unused_ui" {
		t.Fatalf("asset orphan report=%v", report.AssetsLocale.OrphanKeys)
	}
	if len(report.GamepackLocale.OrphanKeys) != 1 || report.GamepackLocale.OrphanKeys[0] != "orphan" {
		t.Fatalf("gamepack orphan report=%v", report.GamepackLocale.OrphanKeys)
	}
}

func TestRunFailsBrokenReferencesAndAsymmetry(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, root, "assets/locale/zh-TW.json", map[string]any{"strings": map[string]string{"other": "其他"}})
	writeJSON(t, root, "gamepack/events/pit-of-moander.json", map[string]any{
		"locales":      map[string]any{"en": map[string]string{"event": "Event"}, "zh-TW": map[string]string{"zh-only": "中文"}},
		"option_rules": []map[string]string{{"id": "", "source": "YES", "message_id": "missing"}},
	})
	writeFile(t, root, "internal/game/state.go", `package game
func show(c interface{ Text(string, string) string }) string { return c.Text("missing-ui", "Missing") }
`)
	report, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	if ViolationCount(report) < 3 {
		t.Fatalf("expected broken-reference violations: %+v", report.Issues)
	}
}

func writeJSON(t *testing.T, root, name string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, name, string(data))
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
