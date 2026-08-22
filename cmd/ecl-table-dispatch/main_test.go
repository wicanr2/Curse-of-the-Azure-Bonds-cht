package main

import (
	"archive/zip"
	"fmt"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// 這一支釘住「查表分派的分母不會悄悄變成 0」。
//
// ★ 為什麼需要。 這個分母整個建立在 `ecl.TraceGraph` 走得到那些指令上。哪天
// 進入點推導或指令解碼改了，掃出來的分派點會安靜地變少甚至歸零——而報表照樣
// 產得出來，只是數字變小。**「靜態掃出 0」與「掃錯地方所以 0」看起來一模一樣**，
// 要有東西把兩者分開。
func TestTableDispatchDenominatorStaysPopulated(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("沒有原版 image：%v", err)
	}
	defer archive.Close()

	blocks, sites, distinct := 0, 0, 0
	for member := 1; member <= 6; member++ {
		name := fmt.Sprintf("ECL%d.DAX", member)
		payload := zipMember(&archive.Reader, name)
		if payload == nil {
			continue
		}
		parsed, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatalf("%s：%v", name, parseErr)
		}
		for _, block := range parsed {
			stat, ok := scan(name, block.Entry.ID, block.Data)
			if !ok {
				continue
			}
			blocks++
			sites += stat.onSites
			distinct += stat.distinct
		}
	}
	// 門檻取得比實測值寬鬆：這是「有沒有整批掉光」的警報，不是逐數字的快照。
	if blocks < 15 || sites < 120 || distinct < 500 {
		t.Fatalf("查表分派掃出來只有 段=%d 分派點=%d 相異目標=%d，"+
			"實測基準約 23／185／835——掉這麼多通常表示追蹤或解碼壞了，不是資料變了",
			blocks, sites, distinct)
	}
}
