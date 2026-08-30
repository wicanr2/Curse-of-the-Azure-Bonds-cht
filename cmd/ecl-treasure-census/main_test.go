package main

import (
	"archive/zip"
	"fmt"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
)

// 門檻測試：寶物分母不會悄悄掉光。
//
// ★ 這個分母建立在 `ecl.TraceGraph` 走得到 `27h` 上。追蹤或解碼一改，掃出來的
// 發放點會安靜地變少，而報表照樣產得出來。「掃出 0」與「掃錯地方所以 0」看起來
// 一模一樣，要有東西把兩者分開。
//
// ⚠ 門檻取得比實測寬鬆：這是「有沒有整批掉光」的警報，不是逐數字快照——
// 逐數字釘住的話，任何解碼改進都會讓它變紅，久了就會被習慣性改掉。
func TestTreasureCensusStaysPopulated(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("沒有原版 image：%v", err)
	}
	defer archive.Close()

	total, classified := 0, 0
	for member := 1; member <= 6; member++ {
		payload := zipMember(&archive.Reader, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatalf("ECL%d.DAX：%v", member, parseErr)
		}
		for _, block := range blocks {
			for _, item := range scan(fmt.Sprintf("ECL%d.DAX", member), block.Entry.ID, block.Data) {
				total++
				if item.kind != "動態" {
					classified++
				}
			}
		}
	}
	if total < 45 {
		t.Fatalf("只掃到 %d 處 `27h TREASURE`，實測基準 63——掉這麼多通常是追蹤或解碼壞了", total)
	}
	if classified*10 < total*8 {
		t.Fatalf("%d／%d 處的物品參數解得出常數，比例低於八成——"+
			"運算元讀法可能改壞了", classified, total)
	}
}
