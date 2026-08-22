package main

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// 分母（walkPages）與比對（walkRuns）走的是同一條控制流，但狀態不同：
// 前者只帶一個位元、一定走得完；後者帶整份已累積文字，可能碰到上限。
//
// 因此**唯一成立的方向是 walkRuns ⊆ walkPages**。這條測試釘住那個包含關係，
// 並把兩者的差額印出來——差額不是 0 不算失敗（那正是截斷的代價），
// 但它必須是**看得到的數字**，不能只在報告裡寫一句「可能有頁沒進分母」。
func TestPageWalkCoversEveryPageTheRunWalkFinds(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()

	totalPages, missedByRunWalk := 0, 0
	for _, member := range members {
		data, err := readMember(&archive.Reader, member)
		if err != nil {
			t.Fatal(err)
		}
		blocks, err := dax.Parse(data)
		if err != nil {
			t.Fatalf("%s: %v", member, err)
		}
		for _, block := range blocks {
			label := fmt.Sprintf("%s/0x%02X", member, block.Entry.ID)
			pages, err := walkPages(block.Data)
			if err != nil {
				t.Fatalf("%s page walk: %v", label, err)
			}
			paths, runPages, err := walkRuns(label, block.Data)
			if err != nil {
				t.Fatalf("%s run walk: %v", label, err)
			}
			_ = paths
			totalPages += len(pages)
			for offset := range runPages {
				if _, ok := pages[offset]; !ok {
					t.Fatalf("%s %s 只被帶文字的走訪找到，分母走訪沒找到——"+
						"分母是過近似，這個方向不該發生", label, offset)
				}
			}
			for offset := range pages {
				if _, ok := runPages[offset]; !ok {
					missedByRunWalk++
				}
			}
		}
	}
	if totalPages == 0 {
		t.Fatal("分母走訪一頁都沒找到")
	}
	t.Logf("分母 %d 頁，其中帶文字的走訪沒走到 %d 頁（截斷的代價，這些頁一律算 unmatched）",
		totalPages, missedByRunWalk)
}

// ★ 這一組釘住的是**逐頁與逐 run 兩種算法的差別**，不是某個數字。
//
// 缺陷的形狀：一條規則的 `all_contains` 含了**兄弟句**的片段，於是兩句一起印的
// 那條路徑命中，而只印其中一句的那條路徑沒有規則。逐頁的算法看不到——那一頁在
// 第一條路徑上已經被標成 `matched`，第二條路徑就不再有人問。玩家走的是**一條**
// run，走到第二條就看到英文。
//
// 這不是假想的失誤。把 `wizard-tower.efreet-band`、
// `tilverton.sewers.spoils-and-master-alone`、`standing-stone.arrival`、
// `zhentil.arena.manticores-attack` 任一條從 pack 拿掉重跑 `cmd/ecl-text-coverage`，
// 逐頁的 `unmatched` **仍然是 0**，而 `runs_orphan` 會從 0 變成拿掉的條數，
// 每一條都指名是哪一份文字——`runs_orphan` 的 0 是撐得住的 0，逐頁的不是。
func TestOrphanRunAccountingCatchesTheAloneSiblingBug(t *testing.T) {
	// 兩條 run 共用 `0x20` 這一頁：A 把它和兄弟句一起印（有規則），
	// B 單獨印（沒有規則）。
	const shared = "0x20"
	notes := map[string]group{shared: {Offset: shared}, "0x10": {Offset: "0x10"}}
	subroutines := map[string]bool{}
	aloneOnly := map[string]bool{shared: false, "0x10": false}

	// 逐頁：這一頁被命中的 run A 經過，所以是 `matched`。
	covered := map[string]string{shared: "sibling-rule"}
	if _, ok := covered[shared]; !ok {
		t.Fatal("前提就錯了：這一頁應該被 run A 標成已接上")
	}

	// 逐 run：run B 只有這一頁、沒有插入點、不是子程式片段 → `orphan`。
	if got := classifyRun([]string{shared}, notes, subroutines, aloneOnly); got != "orphan" {
		t.Fatalf("只印兄弟句其中一句的那條 run 應該是 orphan，實際 %q", got)
	}
}

// 三種豁免都要真的擋得住，否則 `runs_orphan` 會被雜訊淹掉而失去意義。
func TestClassifyRunExemptions(t *testing.T) {
	for _, item := range []struct {
		name        string
		pages       []string
		notes       map[string]group
		subroutines map[string]bool
		aloneOnly   map[string]bool
		want        string
	}{
		{
			name:  "印了執行期的值就靜態比對不到",
			pages: []string{"0x10", "0x20"},
			notes: map[string]group{"0x20": {VariableInserts: []string{"0x24"}}},
			want:  "variable-insert",
		},
		{
			name:        "整條 run 就是一頁共用子程式片段",
			pages:       []string{"0x30"},
			notes:       map[string]group{"0x30": {}},
			subroutines: map[string]bool{"0x30": true},
			aloneOnly:   map[string]bool{"0x30": true},
			want:        "subroutine",
		},
		{
			// ⚠ 和上面差在 aloneOnly：曾經和別的頁同屬一份 run 的子程式頁
			// 是真的整頁事件，不能豁免。
			name:        "子程式裡但曾經和別頁同屬一份 run",
			pages:       []string{"0x30"},
			notes:       map[string]group{"0x30": {}},
			subroutines: map[string]bool{"0x30": true},
			aloneOnly:   map[string]bool{"0x30": false},
			want:        "orphan",
		},
		{
			name:  "有插入點時手上的文字不完整",
			pages: []string{"0x40"},
			notes: map[string]group{"0x40": {GosubInserts: []string{"0x44"}}},
			want:  "gosub-insert",
		},
	} {
		t.Run(item.name, func(t *testing.T) {
			if item.subroutines == nil {
				item.subroutines = map[string]bool{}
			}
			if item.aloneOnly == nil {
				item.aloneOnly = map[string]bool{}
			}
			got := classifyRun(item.pages, item.notes, item.subroutines, item.aloneOnly)
			if got != item.want {
				t.Fatalf("classifyRun ＝ %q，want %q", got, item.want)
			}
		})
	}
}

// truncatedRun 是 `runs_orphan` 那個 0 的守門員：判太鬆會把真的待辦濾掉，
// 判太嚴會讓半句混進清單。兩個方向都要釘。
func TestTruncatedRunTellsHalfSentencesFromRealPages(t *testing.T) {
	all := map[string]bool{
		"YOU ARE AT THE STANDING STONES.": true,
		"THE PATRONS BECOME SURLY. THE":   true,
		"THE PATRONS BECOME SURLY. THE BARTENDER SUGGESTS YOU LEAVE. DO YOU GO?": true,
		"HE CONTINUES, '":         true,
		"A TRAP.'":                true,
		"YOU ARE BEING ATTACKED!": true,
	}
	for _, item := range []struct {
		text string
		want bool
		why  string
	}{
		{"YOU ARE AT THE STANDING STONES.", false, "完整一句，是真的待辦"},
		{"YOU ARE BEING ATTACKED!", false, "驚嘆號同樣是句子結束"},
		{"A TRAP.'", false, "引號前是句號 → 話講完了"},
		{"THE PATRONS BECOME SURLY. THE", true, "停在句子中間"},
		{"HE CONTINUES, '", true, "引號前是逗號 → 話還沒講完"},
		{"", true, "空的一律當半截"},
	} {
		if got := truncatedRun(item.text, all); got != item.want {
			t.Errorf("truncatedRun(%q) ＝ %v，want %v（%s）", item.text, got, item.want, item.why)
		}
	}
	// 嚴格前綴：完整的那條才是實機會演的。
	if !truncatedRun("THE PATRONS BECOME SURLY. THE", all) {
		t.Error("是別條 run 的前綴，應該判成截斷")
	}
}
