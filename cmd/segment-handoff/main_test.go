package main

import (
	"encoding/json"
	"os"
	"testing"
)

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

// ★ 這條守的是**修法有效**這件事本身：把交接狀態鋪回直入那一側之後，
// 「這一段的碼會讀、而直入沒有」的格子必須大幅減少，而且剩下的多數要能歸因
// （引擎自己設的、由位置推出來的視圖暫存器）。
//
// ⚠ 讀 committed 的報表：它是離線產的（要先跑主線錄到達取樣）。檔案不在就跳過。
func TestSeedingHandoffClosesMostOfTheGap(t *testing.T) {
	raw, err := os.ReadFile("../../docs/audit/segment-handoff.json")
	if err != nil {
		t.Skip("還沒產出交接報表")
	}
	var doc report
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Segments == 0 || doc.TotalRisky == 0 {
		t.Fatalf("報表看起來是空的：segments=%d risky=%d", doc.Segments, doc.TotalRisky)
	}
	if doc.SeededRisky >= doc.TotalRisky {
		t.Fatalf("鋪上交接狀態之後沒有變好：%d → %d", doc.TotalRisky, doc.SeededRisky)
	}
	// 至少要關掉八成，否則這個修法不值得留在進段路徑上。
	if doc.SeededRisky*5 > doc.TotalRisky {
		t.Fatalf("只關掉 %d／%d，不到八成", doc.TotalRisky-doc.SeededRisky, doc.TotalRisky)
	}
	// 剩下的要對得起來：每一類加起來就是全部，否則分類漏了一種成因。
	sum := doc.ResidueEngineSet + doc.ResidueViewCell + doc.ResidueLifecycleOwned +
		doc.ResidueRouteLookup + doc.ResidueScratch + doc.ResidueOther
	if sum != doc.SeededRisky {
		t.Fatalf("殘差分類 %d 加起來不等於 %d：漏了一種成因", sum, doc.SeededRisky)
	}
	// ★ 「歸不了因」現在是 0：39 格全部歸得了因。
	//
	// ⚠ 這條**不是**釘死 0：新的段或新的取樣點冒出沒見過的格子是正常的，
	// 那時候該做的是**查清楚它是什麼**再決定要不要分類，不是把門檻放寬。
	// 門檻放在四分之一，是為了讓「分類等於沒解釋什麼」這件事會紅。
	if doc.ResidueOther*4 > doc.SeededRisky {
		t.Fatalf("歸不了因的有 %d／%d，超過四分之一：分類沒有解釋到東西",
			doc.ResidueOther, doc.SeededRisky)
	}
}
