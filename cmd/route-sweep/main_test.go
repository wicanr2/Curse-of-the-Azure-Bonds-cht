package main

import (
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gamecorpus"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
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
	legs := []leg(nil)
	for _, id := range []string{"ECL1/0x50", "ECL1/0x51"} {
		hub, ok := segment.Lookup(id)
		if !ok {
			t.Fatalf("註冊表沒有 %s", id)
		}
		found, _ := walk(data, hub, 8)
		legs = append(legs, found...)
	}
	if len(legs) < 15 {
		t.Fatalf("只走到 %d 個分支，宣告的下限是 15", len(legs))
	}
	for _, item := range legs {
		if item.language == "原文" {
			t.Errorf("`%s` 地點 %d 選「%s」之後演的是原文：%s",
				item.hub, item.from, item.option, firstLine(item.text))
		}
		if item.optionLanguage == "原文" {
			t.Errorf("`%s` 地點 %d 的選項「%s」沒有中文化", item.hub, item.from, item.option)
		}
	}
}
