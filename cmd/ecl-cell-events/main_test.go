package main

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// 對照表是逐格盤點的地圖，所以它的形狀要被釘住：少偵測到一個分派器，
// 那個 block 的內容就會看起來「沒有每格事件」——那是假零。
func TestCellEventTableShape(t *testing.T) {
	const imagePath = "../../curseoftheazurebonds.zip"
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	catalogs := map[uint8]geo.Catalog{}
	for set := 1; set <= 6; set++ {
		payload := memberPayload(archive, fmt.Sprintf("GEO%d.DAX", set))
		if payload == nil {
			continue
		}
		catalog := geo.NewCatalog()
		if err := catalog.AddDAX(uint8(set), payload); err != nil {
			t.Fatal(err)
		}
		catalogs[uint8(set)] = catalog
	}
	reports := map[string]blockReport{}
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			t.Fatalf("image 裡沒有 ECL%d.DAX", member)
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			key := fmt.Sprintf("ECL%d/0x%02X", member, block.Entry.ID)
			reports[key] = describe(catalogs, member, block.Entry.ID, block.Data)
		}
	}
	if len(reports) != 25 {
		t.Fatalf("掃到 %d 個 block，應該是 25 個", len(reports))
	}
	dispatchers := 0
	for _, report := range reports {
		if report.found {
			dispatchers++
		}
	}
	if dispatchers != 14 {
		t.Errorf("有地形分派的 block 是 %d 個，宣告的是 14 個", dispatchers)
	}
	// ⚠ 遮罩不是固定的：`0x7F` 與 `0x3F` 都量到過。寫死一種會讓另一種整個落空。
	masks := map[int]int{}
	for _, report := range reports {
		if report.found {
			masks[report.mask]++
		}
	}
	if masks[0x7F] == 0 || masks[0x3F] == 0 {
		t.Errorf("遮罩分布是 %v，兩種都應該還在", masks)
	}
	// 逐格盤點實際驗過的三個 block 要能對到格子。
	for _, check := range []struct {
		key   string
		index int
		cell  string
	}{
		{"ECL6/0x43", 23, "(10,7)"},  // 內城的上樓樓梯
		{"ECL6/0x43", 26, "(6,1)"},   // 最終戰
		{"ECL6/0x40", 3, "(13,14)"},  // 黛米爾公主
		{"ECL5/0x32", 12, "(7,12)"},  // 克林卓
	} {
		report := reports[check.key]
		if !report.found {
			t.Errorf("%s 沒有偵測到地形分派", check.key)
			continue
		}
		found := false
		for _, event := range report.events {
			if event.index != check.index {
				continue
			}
			for _, cell := range event.cells {
				if cell == check.cell {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("%s 的索引 %d 沒有對到 %s", check.key, check.index, check.cell)
		}
	}
}
