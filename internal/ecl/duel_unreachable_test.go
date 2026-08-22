package ecl

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// `24h COMBAT` 的第一支有兩個條件：`DS:8B69h`（有怪要打）**或** `DS:8B56h`。
//
// ★ `8B56h` ＝ **決鬥旗標**。整個 DOS 執行檔裡只有一處把它設成 1：
// `overlay-07 entry#25` ＝ `GODUEL`（spec 626／917／1149／1182）。而 `GODUEL`
// 只有 `2Dh CALL` 的 `8000h`／`8001h` 兩個選擇子到得了（spec 1150）——
// **corpus 兩個都沒用到**。
//
// ⇒ 這一支在 CoAB **走不到**，所以 remake 用「怪物鏈非空」代替 `8B69h`
// 已經覆蓋了整個可達條件。
//
// ⚠ 本輪已經被假零咬過兩次，所以這條一定要有**正對照**：先確認掃得到那四個
// 真的有在用的選擇子，數字對不上就先懷疑掃描而不是下結論。
func TestDuelCallSelectorsAreUnusedWhileTheOthersAreNot(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("找不到遊戲 image，跳過：%v", err)
	}
	defer archive.Close()

	counts := map[uint16]int{}
	for member := 1; member <= 6; member++ {
		name := "ECL" + string(rune('0'+member)) + ".DAX"
		blocks, parseErr := dax.Parse(realZipMember(t, archive, name))
		if parseErr != nil {
			t.Fatalf("%s: %v", name, parseErr)
		}
		for _, block := range blocks {
			points, pointsErr := entryStartsForScan(block.Data)
			if pointsErr != nil {
				continue
			}
			graph, graphErr := TraceGraph(block.Data, points, len(block.Data)*8)
			if graphErr != nil {
				continue
			}
			seen := map[int]bool{}
			for _, instruction := range graph.Instructions {
				if seen[instruction.Offset] || instruction.Command.Opcode != callOpcode {
					continue
				}
				seen[instruction.Offset] = true
				if operand := instruction.Operands[0]; operand.WordSet {
					counts[operand.Word]++
				}
			}
		}
	}
	// 正對照：這四個是 spec 1150 量過的，掃不到就是掃描壞了。
	for _, item := range []struct {
		selector uint16
		want     int
		name     string
	}{
		{0x2E10, 125, "重畫"}, {0xB200, 19, "音效"},
		{0xC01E, 13, "前進一格"}, {0x6803, 11, "推圖片格"},
	} {
		if counts[item.selector] != item.want {
			t.Fatalf("`%04Xh`（%s）＝ %d 處，want %d——先確認掃描，不要直接改結論",
				item.selector, item.name, counts[item.selector], item.want)
		}
	}
	// 正對照過了，這兩個的零才有意義。
	for _, selector := range []uint16{0x8000, 0x8001} {
		if counts[selector] != 0 {
			t.Errorf("`%04Xh`（GODUEL）出現 %d 處——決鬥不再是死路，`24h` 的 `8B56h` 那一支要重看",
				selector, counts[selector])
		}
	}
}

const callOpcode = 0x2D // `2Dh CALL`

func entryStartsForScan(block []byte) ([]int, error) {
	points, _, err := EntryPoints(block, 5)
	if err != nil {
		return nil, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-CodeAddressBase)
	}
	return starts, nil
}
