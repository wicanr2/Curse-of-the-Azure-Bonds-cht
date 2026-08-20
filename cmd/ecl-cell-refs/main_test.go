package main

import (
	"archive/zip"
	"os"
	"testing"
)

const imagePath = "../../curseoftheazurebonds.zip"

// 這支存在的理由就是「拿事件目錄查會少報」。少報的後果是把「沒有人寫這個旗標」
// 當成結論——那是假零，而且是會讓整條推論反過來的那一種。
//
// 釘住兩件事：走訪找得到目錄漏掉的那一處，以及總數遠多於目錄。
func TestScanFindsSitesTheEventCatalogMisses(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	refs, err := collect(archive, 0x4C06, 0x4C06)
	if err != nil {
		t.Fatal(err)
	}
	// `docs/audit/ecl-event-catalog.json` 對 4C06 只列 18 處。
	if len(refs) < 60 {
		t.Errorf("4C06 只找到 %d 處，事件目錄自己就有 18 處，走訪應該遠多於它", len(refs))
	}
	// 密斯卓諾內城廚房的 once-only 標記——目錄裡沒有這一處。
	found := false
	for _, ref := range refs {
		if ref.member == 6 && ref.block == 0x43 && ref.offset == 0x0dea {
			found = true
		}
	}
	if !found {
		t.Error("找不到 ECL6/0x43 0x0dea 的 `SAVE 01 4C06`")
	}
}

// `LOAD FILES` 不會把 4C00 區清掉——猶拉什在它之前就把自己的地圖參數寫進去，
// 之後整段都在讀。這一條把那個反例釘住：它一消失，「換地圖時整批換掉」的模型
// 就會被錯誤地重新採信。
func TestYulashWritesMapParametersBeforeLoadFiles(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	refs, err := collect(archive, 0x4C05, 0x4C05)
	if err != nil {
		t.Fatal(err)
	}
	before, after := 0, 0
	for _, ref := range refs {
		if ref.member != 3 || ref.block != 0x10 {
			continue
		}
		if ref.beforeLoadFiles {
			before++
			continue
		}
		after++
	}
	if before != 1 {
		t.Errorf("猶拉什在 LOAD FILES 之前碰 4C05 的次數是 %d，宣告的是 1（`SAVE 80 4C05`）", before)
	}
	if after < 10 {
		t.Errorf("猶拉什在 LOAD FILES 之後只碰 4C05 %d 次，那個值應該整段都在用", after)
	}
}
