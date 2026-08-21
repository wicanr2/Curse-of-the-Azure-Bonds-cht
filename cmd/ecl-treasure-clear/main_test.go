package main

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

const imagePath = "../../curseoftheazurebonds.zip"

// corpus 裡「發了戰利品之後還走得到 `1Ch`」的地方只有一處：提爾佛頓火刀首領的
// 重打迴圈。這條把那個數字釘住。
//
// ⚠ 同時擋兩種假零：
//   - 前向走訪只用 `TraceGraph.Edges`（只記跳躍）會得到 0 個配對，看起來像「原作
//     這個行為在本作用不到」，而那是走訪壞掉。
//   - `27h` 的處數塌下來代表掃描根本沒跑到，不是缺口消失。
func TestTreasureFollowedByClearMonsters(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	pairs := []pair(nil)
	treasures := 0
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
			found, sites, scanErr := scanBlock(member, block.Entry.ID, block.Data)
			if scanErr != nil {
				t.Fatalf("ECL%d/0x%02X: %v", member, block.Entry.ID, scanErr)
			}
			pairs = append(pairs, found...)
			treasures += sites
		}
	}
	if treasures != 63 {
		t.Errorf("掃到 %d 處 `27h TREASURE`，應該是 63 處", treasures)
	}
	if len(pairs) != 1 {
		t.Fatalf("走得到的配對有 %d 組，應該是 1 組：%+v", len(pairs), pairs)
	}
	got := pairs[0]
	if got.member != 2 || got.block != 0x04 || got.treasure != 0x037B || got.clear != 0x036C {
		t.Errorf("配對是 ECL%d/0x%02X `%#04x`→`%#04x`，應該是 ECL2/0x04 `0x37b`→`0x36c`",
			got.member, got.block, got.treasure, got.clear)
	}
	if !got.viaCombat {
		t.Error("那一組的路上應該經過 `24h COMBAT`（重打迴圈）")
	}
}
