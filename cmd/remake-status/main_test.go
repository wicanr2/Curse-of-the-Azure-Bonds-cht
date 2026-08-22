package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ★ 這張表的價值全在「沒量到就不要印數字」。抓不到摘要時退回填 0，會把
// 「沒量到」印成「沒有缺口」——正好是這張表存在的理由的反面。
func TestGrepNumbersFailsClosed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.md")
	if err := os.WriteFile(path, []byte("| 充能物品 | 18 |\n18 件裡還有 0 個效果沒接。\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fields, ok := grepNumbers(dir, "report.md",
		`\|\s*充能物品\s*\|\s*(\d+)\s*\|`, `(\d+) 件裡還有 (\d+) 個效果沒接`)
	if !ok || len(fields) != 3 || fields[0] != "18" || fields[2] != "0" {
		t.Fatalf("撈出來的欄位 ＝ %#v ok=%v", fields, ok)
	}

	// 報表換了措辭 ⇒ 整列不出現，不是回一組空字串。
	if err := os.WriteFile(path, []byte("| 充能物品 | 18 |\n措辭改過了。\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if fields, ok := grepNumbers(dir, "report.md",
		`\|\s*充能物品\s*\|\s*(\d+)\s*\|`, `(\d+) 件裡還有 (\d+) 個效果沒接`); ok {
		t.Fatalf("對不上就該整列放棄，實際回了 %#v", fields)
	}

	// 報表根本不在也一樣。
	if _, ok := grepNumbers(dir, "missing.md", `(\d+)`); ok {
		t.Fatal("缺檔應該回 false")
	}
}

// ⚠ 撈出來的字串會直接進 Markdown 表格。整行塞進去會被行內的 `|` 切開欄位，
// 所以只收 capture group、不收整行。
func TestGrepNumbersReturnsGroupsNotWholeLines(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "r.md"),
		[]byte("| 卷軸（類別表的槽 `0Bh`..`0Dh`）| 12 |\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fields, ok := grepNumbers(dir, "r.md", `\|\s*卷軸[^|]*\|\s*(\d+)\s*\|`)
	if !ok || len(fields) != 1 {
		t.Fatalf("fields=%#v ok=%v", fields, ok)
	}
	for _, field := range fields {
		for _, r := range field {
			if r == '|' {
				t.Fatalf("撈出來的欄位含有 `|`，會切壞表格：%q", field)
			}
		}
	}
}

// 缺一份報表時那一列不出現，但**要往 stderr 報**——少一列不能看起來像
// 「那一類沒有問題」。這裡只驗行為的一半（回 false），訊息由呼叫端負責。
func TestReadJSONMissingFileIsNotFatal(t *testing.T) {
	var into struct{}
	if readJSON(t.TempDir(), "nope.json", &into) {
		t.Fatal("缺檔應該回 false")
	}
}

// ★ 「沒有數字」的那幾列是這張表最重要的部分——它們是還沒被量的缺口。
// 每一列都必須帶著「為什麼沒有數字」，否則那一列讀起來就只是一個空格，
// 跟省略它沒有差別。
func TestUnmeasuredRowsAlwaysExplainThemselves(t *testing.T) {
	rows := loadRows(t)
	if len(rows.Unmeasured) == 0 {
		t.Fatal("量不到的類別一列都沒有——那才是需要解釋的狀態")
	}
	for _, row := range rows.Unmeasured {
		if row.Category == "" || row.NotProof == "" {
			t.Errorf("量不到的列缺說明：%+v", row)
		}
		if row.Source != "" || row.Format != "" {
			t.Errorf("量不到的列不該有出處或版面：%+v", row)
		}
	}
	if rows.UnmeasuredLabel == "" {
		t.Error("量不到的列需要一個明確的標示，空白會被讀成「沒有問題」")
	}
}

// 量得到的列反過來：一定要有出處與版面，而且**也要**寫清楚證明不了什麼。
// 少了最後那一句，數字會被當成完成度讀。
func TestMeasuredRowsCarryTheirCaveat(t *testing.T) {
	rows := loadRows(t)
	for key, row := range rows.Measured {
		if row.Category == "" || row.Format == "" || row.Source == "" || row.NotProof == "" {
			t.Errorf("%s 欄位不齊：%+v", key, row)
		}
	}
	// 從 Markdown 撈數字的列要三樣齊全，否則會在執行時才炸。
	for key, row := range rows.Measured {
		hasAny := row.Report != "" || len(row.Patterns) > 0 || len(row.Fields) > 0
		hasAll := row.Report != "" && len(row.Patterns) > 0 && len(row.Fields) > 0
		if hasAny && !hasAll {
			t.Errorf("%s 的 report／patterns／fields 只給了一部分：%+v", key, row)
		}
	}
}

func loadRows(t *testing.T) rowsFile {
	t.Helper()
	var rows rowsFile
	if err := json.Unmarshal(rowsJSON, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}
