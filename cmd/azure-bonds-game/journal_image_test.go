package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// 彈窗的兩個純函式：倍率與平移夾限。兩者都不碰 Ebiten，所以能在沒有顯示的
// 情況下釘住行為。
func TestJournalImageScale(t *testing.T) {
	const viewWidth, viewHeight = 576, 336
	tests := []struct {
		name          string
		width, height int
		zoom          bool
		want          float64
	}{
		{"高圖以高度為準", 338, 1000, false, float64(viewHeight) / 1000},
		{"寬圖以寬度為準", 883, 400, false, float64(viewWidth) / 883},
		{"小圖不放大", 216, 194, false, 1},
		{"原尺寸一律 1:1", 883, 1000, true, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := journalImageScale(test.width, test.height, viewWidth, viewHeight, test.zoom)
			if got != test.want {
				t.Fatalf("scale=%v, want %v", got, test.want)
			}
		})
	}
}

// 平移量必須夾在「圖還看得到」的範圍。圖比視窗小的那一軸要歸零，否則玩家會把
// 圖推出畫面外再也找不回來。
func TestClampPan(t *testing.T) {
	tests := []struct {
		name                  string
		offset, content, view int
		want                  int
	}{
		{"圖比視窗小就歸零", 120, 300, 576, 0},
		{"圖剛好一樣大也歸零", -40, 576, 576, 0},
		{"在範圍內不動", 100, 1000, 576, 100},
		{"超過右界夾住", 900, 1000, 576, 212},
		{"超過左界夾住", -900, 1000, 576, -212},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := clampPan(test.offset, test.content, test.view); got != test.want {
				t.Fatalf("clampPan(%d,%d,%d)=%d, want %d",
					test.offset, test.content, test.view, got, test.want)
			}
		})
	}
}

// 一則手札可能被切成好幾個顯示頁，所以「第幾個顯示頁」**不等於**「第幾則手札」。
// 圖綁在手札上，用顯示頁編號去查會查錯。
//
// ★ 原本的 bug：翻頁走前端自己的 `journalDisplayPage`，按 `I` 看圖卻走
// `state.JournalMessageID()`——那一支讀 `State.JournalPage`，而前端**從來沒有
// 推進過它**。於是不論翻到哪一頁，跳出來的都是第一則的圖（spec 1189）。
func TestJournalDisplayPageMapsBackToItsSourceEntry(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	// 第一則故意用很長的（會被切成多頁），第二則短的。
	long := pack.Locales["zh-TW"]["journal.48"]
	if long == "" {
		t.Skip("game pack 沒有 journal.48")
	}
	face := basicfont.Face7x13
	width := font.MeasureString(face, "AAAAAA").Ceil()
	pages, sources := journalDisplayPagesWithSources([]string{long, "第二則"}, "", face, width, 3)
	if len(pages) != len(sources) {
		t.Fatalf("頁數 %d 與來源數 %d 不一致", len(pages), len(sources))
	}
	// ⚠ 正對照：第一則**必須**真的被切成多頁，否則這條測不到「多對一」。
	first := 0
	for _, source := range sources {
		if source == 0 {
			first++
		}
	}
	if first < 2 {
		t.Fatalf("第一則只佔 %d 個顯示頁，測不到多對一", first)
	}
	if sources[0] != 0 {
		t.Fatalf("第一個顯示頁來自第 %d 則，want 0", sources[0])
	}
	if last := sources[len(sources)-1]; last != 1 {
		t.Fatalf("最後一個顯示頁來自第 %d 則，want 1", last)
	}
	// 來源索引必須是不遞減的：頁是照順序攤平的。
	for index := 1; index < len(sources); index++ {
		if sources[index] < sources[index-1] {
			t.Fatalf("來源索引在第 %d 頁倒退：%v", index, sources)
		}
	}
	// ★ 關鍵：**翻到後面的頁，來源不能還是第 0 則**——那正是原本的行為。
	if sources[len(sources)-1] == sources[0] {
		t.Fatal("最後一頁與第一頁來自同一則，對照表沒有作用")
	}
}

// TestJournalPagingAdvancesTheStateSourcePage 釘住「翻頁會把 `State.JournalPage`
// 一起推」。
//
// ★ 為什麼值得一支測試。 翻頁走的是前端自己的 `journalDisplayPage`（第幾個顯示
// 頁），而 `State` 那一側記的是**第幾則**。兩個編號不同步時畫面完全正常——文字
// 換了、頁碼也換了——只有 `State.JournalPage` 永遠停在 0，於是
// `NextJournalPage` 這一支**玩家按不出來**（`cmd/player-path-audit` 唯一的硬缺口）。
func TestJournalPagingAdvancesTheStateSourcePage(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	long := pack.Locales["zh-TW"]["journal.48"]
	if long == "" {
		t.Skip("game pack 沒有 journal.48")
	}
	face := basicfont.Face7x13
	state := game.NewState(locale.Catalog{Language: "zh-TW", Strings: map[string]string{}})
	state.JournalPages = []string{long, "第二則"}
	state.JournalText = long
	if err := state.OpenJournal(); err != nil {
		t.Fatal(err)
	}
	// ⚠ 走**真的按鍵**，不要直接呼叫 `syncJournalSourcePage`：直接呼叫的話，
	// 把按鍵處理那一段的呼叫拿掉這條照樣會過——而缺口正是在那一段。
	keys := newScriptedKeys()
	application := &app{state: &state, face: face, keys: keys}

	_, sources := journalDisplayPagesWithSources(
		state.JournalPages, state.JournalText, face, 22*faceCellWidth(face), 7)
	if len(sources) < 2 {
		t.Skipf("這一組測資只有 %d 個顯示頁，測不到翻頁", len(sources))
	}
	press := func(key ebiten.Key) {
		t.Helper()
		keys.press(key)
		if err := application.Update(); err != nil {
			t.Fatal(err)
		}
		keys.release()
		if err := application.Update(); err != nil {
			t.Fatal(err)
		}
	}
	// 一路按到最後一個顯示頁，`State.JournalPage` 要跟著到最後一則。
	for index := 0; index+1 < len(sources); index++ {
		press(ebiten.KeyRight)
	}
	if got, want := state.JournalPage, sources[len(sources)-1]; got != want {
		t.Fatalf("翻到最後一頁之後 State.JournalPage=%d，want %d", got, want)
	}
	// 再按回第一頁，也要跟著退回去（不是只會往前）。
	for index := len(sources) - 1; index > 0; index-- {
		press(ebiten.KeyLeft)
	}
	if state.JournalPage != sources[0] {
		t.Fatalf("翻回第一頁之後 State.JournalPage=%d，want %d", state.JournalPage, sources[0])
	}
}
