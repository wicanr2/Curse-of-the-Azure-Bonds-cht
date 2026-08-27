package main

import (
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gamecorpus"
)

const (
	imagePath  = "../../curseoftheazurebonds.zip"
	localeFile = "../../assets/locale/zh-TW.json"
)

// 世界地圖上玩家看得到的字**與選項本身**都不准落回原文。
//
// ⚠ 選項要單獨驗：`languageOf(演出來的字)` 看的是敘述，敘述中文化了、選項還是
// 英文的情況真的發生過（溫拉爾軍隊那一場的 `MEET THEM`／`PASS`／`HIDE`）。
func TestWorldMapBranchesAreLocalized(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	data, err := gamecorpus.Load(imagePath, localeFile)
	if err != nil {
		t.Fatal(err)
	}
	legs, _, err := sweepAllLocations(data, 6)
	if err != nil {
		t.Fatal(err)
	}
	// ⚠ 下限釘的是**走到的分支數**：走訪一旦壞掉（起點進不去、選單認錯），
	// 分支數會塌下來而每一條剩下的都還是中文——語系全綠但其實什麼都沒驗到。
	// 現行 game pack 與深度 6 的決定性結果是 698；連續重跑兩次一致。
	// 這裡仍釘住完整現行下限，少一條就會讓測試失敗。
	if len(legs) < 698 {
		t.Fatalf("只走到 %d 個分支，宣告的下限是 698", len(legs))
	}
	for _, item := range legs {
		if item.language == "原文" {
			t.Errorf("起點 %d 地點 %d 選「%s」之後演的是原文：%s",
				item.origin, item.from, item.option, firstLine(item.text))
		}
		if item.optionLanguage == "原文" {
			t.Errorf("起點 %d 地點 %d 的選項「%s」沒有中文化",
				item.origin, item.from, item.option)
		}
	}
}
