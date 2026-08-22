package game

import "testing"

// `JournalMessageIDAt` 讓呼叫端指定索引。前端把手札再切成顯示頁，
// 一則可能佔好幾頁，所以它手上的頁碼不是 `JournalPage`（spec 1189）。
func TestJournalMessageIDAtIsIndependentOfTheCurrentPage(t *testing.T) {
	state := NewState(testCatalog())
	state.journalMessageIDs = []string{"journal.01", "journal.02", "journal.03"}
	state.JournalPage = 0

	// ⚠ 重點：`JournalPage` 停在 0，指定索引仍然要拿得到別則。
	// 原本前端只有「目前這一則」這一個入口，而它從來沒被推進過。
	for index, want := range []string{"journal.01", "journal.02", "journal.03"} {
		if got := state.JournalMessageIDAt(index); got != want {
			t.Errorf("索引 %d ＝ %q，want %q", index, got, want)
		}
	}
	if got := state.JournalMessageID(); got != "journal.01" {
		t.Errorf("目前這一則 ＝ %q，want journal.01", got)
	}
	// 越界兩邊都要回空字串，不是 panic 也不是夾到邊界。
	for _, index := range []int{-1, 3, 999} {
		if got := state.JournalMessageIDAt(index); got != "" {
			t.Errorf("越界索引 %d ＝ %q，want 空字串", index, got)
		}
	}
}
