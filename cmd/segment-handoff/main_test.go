package main

import "testing"

// ★ 這條守的是一個**已經發生過**的錯：第一版忘了先造隊伍，於是每一段的
// `EnterSegment`／`SavePartyFile` 都失敗，而失敗路徑把「漏掉幾格」留在 0。
// 報表印出來是「16 段比得成、漏掉 0 格」——**和「完全對得上」長得一模一樣**。
//
// 分母與失敗數必須是兩個欄位，而且失敗的段不能算進分母。
func TestFailedSegmentsAreNotCountedAsComparedZeroes(t *testing.T) {
	doc := report{}
	// 模擬三段比得成、兩段比不成。
	for i := 0; i < 3; i++ {
		doc.Segments++
		doc.Rows = append(doc.Rows, handoffRow{ID: "ok", Dropped: 5})
		doc.TotalDropped += 5
	}
	for i := 0; i < 2; i++ {
		doc.Failed++
		doc.Rows = append(doc.Rows, handoffRow{ID: "bad", Note: "直入進不去：…"})
	}
	if doc.Segments != 3 {
		t.Fatalf("比得成的段數 ＝ %d，want 3（比不成的不算）", doc.Segments)
	}
	if doc.Failed != 2 {
		t.Fatalf("比不成的段數 ＝ %d，want 2", doc.Failed)
	}
	text := render(doc)
	if !contains(text, "比不成") {
		t.Fatal("報表沒有把比不成的段講出來——那個零會被讀成「完全對得上」")
	}
	// 全部都比不成時，報表**不可以**只印一句「漏掉 0 格」就收工。
	allFailed := report{Failed: 4, Rows: []handoffRow{{ID: "bad", Note: "直入進不去：…"}}}
	if !contains(render(allFailed), "比不成") {
		t.Fatal("一段都沒比成的時候更要講")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
