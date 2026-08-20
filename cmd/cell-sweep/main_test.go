package main

import (
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

const (
	imagePath  = "../../curseoftheazurebonds.zip"
	localeFile = "../../assets/locale/zh-TW.json"
)

// 逐格實測是「段內文字有沒有接上」的閘。它擋的是三件事：
//
//  1. 站上去演出來的字落回原文；
//  2. 沒演出來卻說不出原因——那是假零，跟「這一格沒有內容」長得一樣；
//  3. 掃到的 block 變少——少一個 block 就整批內容從盤點裡消失。
func TestCellSweepHasNoUntranslatedOrUnexplainedCells(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	data, err := loadCorpus(imagePath, localeFile, 8)
	if err != nil {
		t.Fatal(err)
	}
	sweeps := make([]blockSweep, 0, 25)
	for _, seg := range segment.All() {
		sweeps = append(sweeps, sweepBlock(data, seg))
	}
	counts := summarise(sweeps)

	// ⚠ 13 而不是 14：`ECL4/0x25` 有分派表，但從註冊表宣告的入口進去是魔法商店，
	// 到不了它那張地圖（GEO4/0x25 的莊園）。remake 的眼魔洞穴走的是 game pack
	// 事件重建，不是這個 block 的每格分派。
	if counts["block"] != 13 {
		t.Errorf("掃到 %d 個 block，宣告的是 13 個", counts["block"])
	}
	if counts["原文"] != 0 {
		t.Errorf("有 %d 格演出來是原文", counts["原文"])
	}
	if counts["中文"] < 185 {
		t.Errorf("演出中文的格子剩 %d 個，宣告的下限是 185", counts["中文"])
	}
	for _, sweep := range sweeps {
		for _, cell := range sweep.cells {
			if cell.played() || cell.note == "地圖上沒有這個地形碼" {
				continue
			}
			if cell.guard == "" {
				t.Errorf("%s 索引 %d（%s）沒演出來，也說不出守衛是什麼",
					sweep.id, cell.index, cell.cell)
			}
		}
	}
}

// 邊界那一圈的處理常式是 `COMPARE C04D <方向> / IF <> / EXIT`，而 `C04D` ＝
// 朝向 ÷ 2。掃 0..3 只蓋得到 `C04D` 的 0 與 1，南、西兩面會整批落空——那正是
// 猶拉什四個出城口曾經看起來「沒內容」的原因。
func TestCellSweepCoversAllFourFacings(t *testing.T) {
	seen := map[string]bool{}
	for _, facing := range []uint8{0, 2, 4, 6} {
		seen[facingName(facing)] = true
	}
	for _, want := range []string{"北", "東", "南", "西"} {
		if !seen[want] {
			t.Errorf("掃描的朝向蓋不到%s面", want)
		}
	}
}
