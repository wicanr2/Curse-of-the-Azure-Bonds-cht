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

// `37h LOAD PIECES` 的兩槽分支**真的會走到**：四個 block 在自己的 `LOAD PIECES`
// 之前就把 `4BE7h`／`4BE8h` 兩個閘門都設成非零（spec 1153）。
//
// ★ 這一條的用途是擋住「把分支簡化掉」：spec 1153 先前推論那一支是死碼，
// 是這支工具掃出 producer 才推翻的。四個現場的第二個運算元是**死資料**——
// 照逐槽迴圈載的話會多寫一組原作沒有的牆面組。
func TestTwoSlotWallSetGatesArePrimedBeforeLoadPieces(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	primed := map[[2]int]bool{}
	for _, gate := range []uint16{0x4BE7, 0x4BE8} {
		refs, collectErr := collect(archive, gate, gate)
		if collectErr != nil {
			t.Fatal(collectErr)
		}
		if len(refs) == 0 {
			t.Fatalf("%04X 一處都沒找到——正對照失敗，這個測試等於沒跑", gate)
		}
		for _, ref := range refs {
			// 設成非零的那幾處都排在該 block 的第一個 LOAD FILES 之前，
			// 而 LOAD FILES 又緊接著 LOAD PIECES。
			if ref.beforeLoadFiles {
				key := [2]int{ref.member, int(ref.block)}
				if gate == 0x4BE7 {
					primed[key] = true
				} else if !primed[key] {
					t.Errorf("ECL%d/0x%02X 只把其中一個閘門設在 LOAD FILES 之前", ref.member, ref.block)
				}
			}
		}
	}
	want := [][2]int{{5, 0x33}, {5, 0x35}, {6, 0x40}, {6, 0x45}}
	for _, key := range want {
		if !primed[key] {
			t.Errorf("ECL%d/0x%02X 應該在 LOAD PIECES 之前開啟兩槽分支", key[0], key[1])
		}
	}
	if len(primed) != len(want) {
		t.Errorf("開啟兩槽分支的 block 有 %d 個，want %d：%v", len(primed), len(want), primed)
	}
}
