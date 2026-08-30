package ecl

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
)

// `24h COMBAT` 的商店與神殿旗標**由腳本自己寫**，慣用法一模一樣：
//
//	CLEARMONSTERS          ← 保證第一支（有怪就打）不成立
//	SAVE 01 <旗標>
//	COMBAT
//
// ★ 這條取代了先前那個 `TestCampRequestFlagHasNoProducer`。那一條釘的是
// **一個假零**：spec 1095 說「這兩格在 1,355 條 ECL 指令裡沒有任何一條寫過
// （已全掃確認）」，而全 corpus 掃下去是 `7EE2h` 4 處、`7F6Ch` 9 處。
// 「掃過了、沒有」是最容易自洽而錯的結論，所以這裡改成**正對照**：
// 數字對不上就紅，而且兩個方向都擋——多了要重新確認，少了代表掃描壞了。
const saveOpcode = 0x09 // `09h SAVE`

func TestServiceRequestFlagsHaveScriptProducers(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("找不到遊戲 image，跳過：%v", err)
	}
	defer archive.Close()

	counts := map[uint16]int{}
	for member := 1; member <= 6; member++ {
		name := "ECL" + string(rune('0'+member)) + ".DAX"
		payload := realZipMember(t, archive, name)
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatalf("%s: %v", name, parseErr)
		}
		for _, block := range blocks {
			// 走控制流而不是掃位元組：變長指令用線性掃描會錯位，
			// 而錯位的結果剛好會長成「一處都沒有」（`cmd/ecl-cell-refs` 同一條路）。
			points, _, pointsErr := EntryPoints(block.Data, 5)
			if pointsErr != nil {
				continue
			}
			starts := make([]int, 0, len(points))
			for _, point := range points {
				starts = append(starts, int(point)-CodeAddressBase)
			}
			graph, graphErr := TraceGraph(block.Data, starts, len(block.Data)*8)
			if graphErr != nil {
				continue
			}
			seen := map[int]bool{}
			for _, instruction := range graph.Instructions {
				if seen[instruction.Offset] || instruction.Command.Opcode != saveOpcode {
					continue
				}
				seen[instruction.Offset] = true
				for _, operand := range instruction.Operands {
					if operand.WordSet {
						counts[operand.Word]++
					}
				}
			}
		}
	}
	if counts[addrTempleRequest] == 0 || counts[addrShopRequest] == 0 {
		t.Fatalf("旗標的 producer 掃成 0（神殿 %d、商店 %d）——那是假零，先確認掃描而不是下結論",
			counts[addrTempleRequest], counts[addrShopRequest])
	}
	for _, item := range []struct {
		address uint16
		name    string
		want    int
	}{
		{addrTempleRequest, "神殿 7EE2h", 4},
		{addrShopRequest, "商店 7F6Ch", 9},
	} {
		if counts[item.address] != item.want {
			t.Errorf("%s 的 producer ＝ %d 處，want %d（數字變了就要回頭確認 corpus）",
				item.name, counts[item.address], item.want)
		}
	}
}
