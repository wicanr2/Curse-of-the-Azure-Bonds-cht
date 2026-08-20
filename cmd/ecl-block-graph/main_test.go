package main

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"
)

// 轉移圖是分段驗證的骨架（`docs/plan/mainline-segmented-verification.md`），
// 所以它的形狀要被釘住。
//
// ⚠ 這條同時擋一個假零：`ecl.TraceGraph` **會跟 `ON GOTO` 的每一個目的地**，
// 事件目錄那份不跟。用不跟的那份算，兩個世界地圖 hub（`0x50`／`0x51`）會看起來
// 一條出邊都沒有——而那是整條主線的樞紐。
func TestBlockGraphShape(t *testing.T) {
	const imagePath = "../../curseoftheazurebonds.zip"
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	graphs := map[blockKey]*blockGraph{}
	for chapter := 1; chapter <= 6; chapter++ {
		member := fmt.Sprintf("ECL%d.DAX", chapter)
		payload := memberPayload(archive, member)
		if payload == nil {
			t.Fatalf("image 裡沒有 %s", member)
		}
		blocks, err := daxParse(payload)
		if err != nil {
			t.Fatal(err)
		}
		for _, raw := range blocks {
			graphs[blockKey{member, fmt.Sprintf("0x%02X", raw.Entry.ID)}] = traceBlock(raw.Data)
		}
	}
	if len(graphs) != 25 {
		t.Fatalf("block 數 %d，預期 25", len(graphs))
	}
	edges, noExit := 0, []blockKey{}
	for key, g := range graphs {
		edges += len(g.newECL)
		if len(g.newECL) == 0 {
			noExit = append(noExit, key)
		}
	}
	if edges != 47 {
		t.Errorf("NEWECL 出邊 %d 條，預期 47", edges)
	}
	if len(noExit) != 2 {
		t.Errorf("沒有出邊的 block 有 %d 個（%v），預期 2（開場 0x52 與結局 0x43）", len(noExit), noExit)
	}
	for _, hub := range []blockKey{{"ECL1.DAX", "0x50"}, {"ECL1.DAX", "0x51"}} {
		if len(graphs[hub].newECL) < 5 {
			t.Fatalf("%v 只有 %d 條出邊——世界地圖 hub 的出邊掉了，多半是可達性又不跟 ON GOTO 了",
				hub, len(graphs[hub].newECL))
		}
	}
}
