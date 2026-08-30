package main

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
)

const imagePath = "../../curseoftheazurebonds.zip"

// `24h COMBAT` 的 199 處分成兩堆：前面把怪擺上去的（真的要打），以及前面什麼都
// 沒擺的（`24h` 的服務分派那一支——商店／營地／神殿）。
//
// ⚠ 這一條同時擋兩種假零：
//   - 反向走如果只用 `TraceGraph` 的邊（那些邊**只記跳躍**），多數 `24h` 找不到
//     任何前驅，會得到「0 處擺過怪」。第一版就是這樣。
//   - 「沒擺過怪」那一堆如果變成 0，代表分類垮了，不是缺口消失了。
func TestCombatSiteClassification(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	sites := []site(nil)
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
			found, scanErr := scanBlock(member, block.Entry.ID, block.Data)
			if scanErr != nil {
				t.Fatalf("ECL%d/0x%02X: %v", member, block.Entry.ID, scanErr)
			}
			sites = append(sites, found...)
		}
	}
	armed, service := 0, 0
	for _, item := range sites {
		if item.armed {
			armed++
			continue
		}
		service++
		// 沒擺過怪的那一堆，每一處往回走都碰得到 `CLEARMONSTERS`——場是被清掉的，
		// 不是「走訪跟丟了」。
		if !item.cleared {
			t.Errorf("ECL%d/0x%02X %#04x 前面既沒擺怪也沒清場，分類說不出它是哪一支",
				item.member, item.block, item.offset)
		}
	}
	for _, check := range []struct {
		label string
		got   int
		want  int
	}{
		{"`24h COMBAT` 的處數", len(sites), 199},
		{"前面擺過怪的", armed, 153},
		{"服務分派那一支", service, 46},
	} {
		if check.got != check.want {
			t.Errorf("%s是 %d，先前量到的是 %d", check.label, check.got, check.want)
		}
	}
}

// 循序前驅是分類的基礎：拿掉它就會得到「一處都沒擺過怪」的假零。
func TestFallThroughStopsAtUnconditionalTransfers(t *testing.T) {
	for _, check := range []struct {
		opcode byte
		want   bool
	}{
		{opExit, false}, {opGoto, false}, {opReturn, false},
		{opLoadMonster, true}, {opSetupMonster, true}, {opCombat, true},
	} {
		instruction := syntheticInstruction(check.opcode)
		if got := fallsThrough(instruction); got != check.want {
			t.Errorf("opcode %#02X 的 fallsThrough 是 %v，應該是 %v",
				check.opcode, got, check.want)
		}
	}
}
