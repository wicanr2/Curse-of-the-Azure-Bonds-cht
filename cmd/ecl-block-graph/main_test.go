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
	// 每個 block 都要有入邊或出邊——孤立的 block 代表圖算漏了。
	// ⚠ 開場 `0x52` 例外：它由**引擎**選中（`DS:4FBCh ≠ 0` 時主迴圈直接載它，
	// spec 1141），不是被 `NEWECL` 指過去的，所以在這張圖裡本來就是孤立的。
	incoming := map[int]int{}
	for _, g := range graphs {
		for target := range g.newECL {
			incoming[target]++
		}
	}
	for key, g := range graphs {
		var id int
		fmt.Sscanf(key.block, "0x%X", &id)
		if key == (blockKey{"ECL1.DAX", "0x52"}) {
			continue
		}
		if len(g.newECL) == 0 && incoming[id] == 0 {
			t.Errorf("%v 既沒有出邊也沒有入邊——孤立的 block 代表圖算漏了", key)
		}
	}
	for _, hub := range []blockKey{{"ECL1.DAX", "0x50"}, {"ECL1.DAX", "0x51"}} {
		if len(graphs[hub].newECL) < 5 {
			t.Fatalf("%v 只有 %d 條出邊——世界地圖 hub 的出邊掉了，多半是可達性又不跟 ON GOTO 了",
				hub, len(graphs[hub].newECL))
		}
	}
	// `LOAD FILES` 的第一個運算元就是那一段載的地圖區塊；世界地圖 hub 與開場
	// 用 `7F`（不載 3D 地圖）。這條釘住兩者的形狀。
	for _, hub := range []blockKey{{"ECL1.DAX", "0x50"}, {"ECL1.DAX", "0x51"}, {"ECL1.DAX", "0x52"}} {
		if _, ok := graphs[hub].loadFiles["7F/7F/7F"]; !ok {
			t.Errorf("%v 的 LOAD FILES 不是 7F/7F/7F：%v", hub, graphs[hub].loadFiles)
		}
	}
	for _, dungeon := range []struct {
		key  blockKey
		want string
	}{
		{blockKey{"ECL2.DAX", "0x03"}, "03/02/FF"},
		{blockKey{"ECL4.DAX", "0x20"}, "20/02/FF"},
		{blockKey{"ECL6.DAX", "0x40"}, "40/02/FF"},
	} {
		if _, ok := graphs[dungeon.key].loadFiles[dungeon.want]; !ok {
			t.Errorf("%v 的 LOAD FILES 少了 %s：%v", dungeon.key, dungeon.want, graphs[dungeon.key].loadFiles)
		}
	}
}
