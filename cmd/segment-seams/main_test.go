package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestClassifyPutsFreshStartBeforeNotVisited(t *testing.T) {
	for _, item := range []struct {
		name     string
		declared uint8
		visited  bool
		from     map[uint8]bool
		want     string
	}{
		{name: "全新開局，主線沒有轉移進來", declared: 0x00, visited: false, want: "fresh-start"},
		{name: "全新開局也不會被判成沒走到", declared: 0x00, visited: true,
			from: map[uint8]bool{0x51: true}, want: "fresh-start"},
		{name: "主線沒走到", declared: 0x51, visited: false, want: "not-visited"},
		{name: "宣告的就是實際的", declared: 0x51, visited: true,
			from: map[uint8]bool{0x51: true}, want: "match"},
		{name: "有好幾個入口，宣告的是其中一個", declared: 0x10, visited: true,
			from: map[uint8]bool{0x10: true, 0x12: true}, want: "match"},
		{name: "宣告的不在實際的裡面", declared: 0x51, visited: true,
			from: map[uint8]bool{0x10: true}, want: "mismatch"},
	} {
		t.Run(item.name, func(t *testing.T) {
			if got := classify(item.declared, item.visited, item.from); got != item.want {
				t.Fatalf("classify ＝ %q，want %q", got, item.want)
			}
		})
	}
}

// ★ 這條守的是**已經產出的那份報表**：接縫一旦對不上，代表某一段直入時寫進
// `LastECL` 的值和主線實際走進去的不一樣——那一段的分段驗收就不能代表主線。
//
// ⚠ 報表是離線產的（要先跑主線錄段落轉移），所以這裡讀committed 的那一份。
// 檔案不在就跳過：這條測試守的是內容，不是「有沒有跑過」。
func TestCommittedSeamReportHasNoMismatch(t *testing.T) {
	raw, err := os.ReadFile("../../docs/audit/segment-seams.json")
	if err != nil {
		t.Skip("還沒產出接縫報表")
	}
	var doc report
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Segments == 0 {
		t.Fatal("接縫報表是空的")
	}
	if doc.Mismatched != 0 {
		for _, row := range doc.Rows {
			if row.Verdict == "mismatch" {
				t.Errorf("%s 宣告從 %02Xh 進來，主線實際是 %v", row.ID, row.Declared, row.Observed)
			}
		}
		t.Fatalf("有 %d 段的接縫對不上", doc.Mismatched)
	}
	// 主線走到的段數是這份報表的**有效分母**：掉下去代表主線繞路了或錄製漏了。
	if doc.Visited < 18 {
		t.Fatalf("主線只走到 %d 段，之前是 18 段", doc.Visited)
	}
}
