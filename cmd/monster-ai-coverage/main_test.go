package main

import (
	"archive/zip"
	"fmt"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// ★ 釘住本作的一個事實：**沒有任何一隻怪物帶記憶法術**。
//
// 這件事決定了施法 AI 的分母是 0——`AIChooseSpell` 那條路（門檻掃描、每輪抽
// 3 個、士氣閘門，spec 835／836／1116）在 CoAB 一次都不會被觸發。不把它寫下來，
// 下一個人看到「施法 AI 的測試全綠」會以為那條路被驗過了，其實一次都沒跑到。
//
// ⚠ 這支測試同時是**解析器的正對照**：哪天 `internal/monster` 的法術槽位移
// （`+33h..+6Ah`）讀錯了，這裡會從「0 隻」變成「有幾隻」，當場紅。
// 「掃出來是 0」與「掃錯地方所以是 0」看起來一模一樣，要有東西把兩者分開。
func TestNoMonsterCarriesMemorisedSpells(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("沒有原版 image：%v", err)
	}
	defer archive.Close()

	total, withSpells := 0, 0
	for chapter := 1; chapter <= 6; chapter++ {
		payload := member(&archive.Reader, fmt.Sprintf("MON%dCHA.DAX", chapter))
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatalf("MON%dCHA.DAX：%v", chapter, parseErr)
		}
		for _, block := range blocks {
			record, recordErr := monster.Parse(block.Data)
			if recordErr != nil {
				continue
			}
			total++
			if len(record.SpellIDs) > 0 {
				withSpells++
			}
		}
	}
	if total < 80 {
		t.Fatalf("只解析到 %d 隻怪物，原版是 81 隻——解析器或資料路徑有問題", total)
	}
	if withSpells != 0 {
		t.Fatalf("有 %d 隻怪物帶記憶法術；本作原本是 0 隻。"+
			"若資料判讀改了，施法 AI 的分母就不再是 0，報表與規格都要跟著改", withSpells)
	}
}
