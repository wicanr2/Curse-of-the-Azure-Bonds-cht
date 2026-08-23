package gfx_test

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

// wallSetSlotBytes 是一個牆面組槽位的位元組數（原作 `WALLSETBYTES = 780`
// ＝ `WALLBYTES 156` × 5）。
const wallSetSlotBytes = 780

// TestWallSetBlockSizeCarriesSlotCount 釘住「一塊 `WALLDEF` 佔幾個槽由它自己的
// 大小決定」。
//
// ★ 為什麼要釘。 `ECL5/0x33`／`0x35`／`ECL6/0x40`／`0x45` 的第二個運算元是
// 15／18，而 15／18 在任何一個 `WALLDEF*.DAX` 裡都查不到。看到這個很容易推出
// 「資料缺了」或「槽 2 保留上一張圖的內容」——兩個都不對。真正的原因是選圖
// 14／17 各 1,560 bytes ＝ **兩個槽**，`LOADWALLSET` 一次呼叫就把連續的兩個槽
// 一起填滿（`overlay-30:0F8Fh`，spec 1185）。
//
// 第二個運算元剛好是第一個 ＋ 1，所以它讀起來很像「另一塊資料」，其實是同一塊
// 的續段編號。兩槽模式那條閘門的用途就是不要重覆載一次。
func TestWallSetBlockSizeCarriesSlotCount(t *testing.T) {
	multiSlot := map[string]map[uint8]int{
		"WALLDEF5.DAX": {14: 2},
		"WALLDEF6.DAX": {17: 2},
	}
	for area := 2; area <= 6; area++ {
		name := fmt.Sprintf("WALLDEF%d.DAX", area)
		blocks := imageBlocks(t, name)
		if len(blocks) == 0 {
			t.Fatalf("%s 沒有任何區塊", name)
		}
		for _, block := range blocks {
			if len(block.Data)%wallSetSlotBytes != 0 {
				t.Fatalf("%s 區塊 %d 是 %d bytes，不是 %d 的倍數",
					name, block.Entry.ID, len(block.Data), wallSetSlotBytes)
			}
			slots := len(block.Data) / wallSetSlotBytes
			want := 1
			if expected, ok := multiSlot[name][block.Entry.ID]; ok {
				want = expected
			}
			if slots != want {
				t.Errorf("%s 區塊 %d 佔 %d 個槽，預期 %d", name, block.Entry.ID, slots, want)
			}
		}
	}
}

// TestMultiSlotWallSetsResolveContinuationSymbols 釘住多槽牆面組的符號區塊編號
// ＝ `選圖 × 10 + 續段序號`（原作 `overlay-30:11BAh`）。
//
// ⚠ 這一條就是「15／18 不存在」的另一面：`8X8D5.DAX` 沒有區塊 14 也沒有 15，
// 它有的是 **141／142**；`8X8D6.DAX` 有 **171／172**。少了這條對應，四張圖的
// 槽 2 會安靜地畫成空氣。
func TestMultiSlotWallSetsResolveContinuationSymbols(t *testing.T) {
	for _, testCase := range []struct {
		area       int
		selector   uint8
		wantSymbol []uint8
	}{
		{area: 5, selector: 14, wantSymbol: []uint8{141, 142}},
		{area: 6, selector: 17, wantSymbol: []uint8{171, 172}},
	} {
		wallBlocks := imageBlocks(t, fmt.Sprintf("WALLDEF%d.DAX", testCase.area))
		symbolBlocks := imageBlocks(t, fmt.Sprintf("8X8D%d.DAX", testCase.area))
		// 槽 1 起載：兩個槽剛好放得下（原作的界限是 `槽 + 段數 <= 4`）。
		set, err := gfx.ParsePieceSet(1, testCase.selector, wallBlocks, symbolBlocks)
		if err != nil {
			t.Fatalf("選圖 %d：%v", testCase.selector, err)
		}
		if len(set.SymbolBlockIDs) != len(testCase.wantSymbol) {
			t.Fatalf("選圖 %d 切出 %d 段，預期 %d",
				testCase.selector, len(set.SymbolBlockIDs), len(testCase.wantSymbol))
		}
		for index, want := range testCase.wantSymbol {
			if set.SymbolBlockIDs[index] != want {
				t.Errorf("選圖 %d 第 %d 段用符號區塊 %d，預期 %d",
					testCase.selector, index+1, set.SymbolBlockIDs[index], want)
			}
			if set.SymbolSetIDs[index] != uint8(index+1) {
				t.Errorf("選圖 %d 第 %d 段落在全域符號組 %d，預期 %d",
					testCase.selector, index+1, set.SymbolSetIDs[index], index+1)
			}
		}
	}
}

func imageBlocks(t *testing.T, member string) []dax.Block {
	t.Helper()
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("讀不到原版 image：%v", err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, member) {
			continue
		}
		handle, openErr := file.Open()
		if openErr != nil {
			t.Fatalf("開 %s：%v", member, openErr)
		}
		defer handle.Close()
		payload, readErr := io.ReadAll(handle)
		if readErr != nil {
			t.Fatalf("讀 %s：%v", member, readErr)
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatalf("解析 %s：%v", member, parseErr)
		}
		return blocks
	}
	t.Fatalf("原版 image 裡沒有 %s", member)
	return nil
}
