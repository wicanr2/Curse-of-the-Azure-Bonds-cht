package ecl

import "testing"

// `21h LOAD FILES` 的 handler 自己會寫兩格 ECL 記憶體（spec 1087／1181）：
//
//	if o[1] <> 0FFh and o[1] <> 7Fh and bank0^[1CCh] <> 0 then
//	    bank0^[18Ah] := o[1];  LOAD3DMAP(o[1]);  bank1^[592h] := 0
//
// ★ 那兩格**腳本讀得到**，所以少寫是控制流分歧不是表現層問題——
// `ECL2/0x03:00F6h` 就在 `COMPARE 7EC9 FF` 上分岔。
func TestLoadFilesWritesTheMapCellsOnlyInThreeDMode(t *testing.T) {
	// 前兩個位元組是 block 表頭。`21h` ＝ LOAD FILES，三個運算元各兩個位元組
	// （`<code> <low>`，`code = 0` 是立即數），運算元是 `<地圖> 02 FF`；
	// 最後的 `00h` ＝ EXIT。
	program := []byte{0, 0, 0x21, 0x00, 0x45, 0x00, 0x02, 0x00, 0xFF, 0x00}

	for _, item := range []struct {
		name       string
		gate       uint16
		piece      byte
		wantLoaded bool
	}{
		{name: "閘門開著、片號正常 ⇒ 兩格都寫", gate: 1, piece: 0x45, wantLoaded: true},
		{name: "閘門是 0 ⇒ 什麼都不寫（原作改走載大圖那一路）", gate: 0, piece: 0x45},
		{name: "片號 0FFh ⇒ 不載", gate: 1, piece: 0xFF},
		{name: "片號 7Fh ⇒ 不載", gate: 1, piece: 0x7F},
	} {
		code := append([]byte(nil), program...)
		code[4] = item.piece
		// ⚠ `Started` 要先設起來，否則 `runtime.Memory` 不會被帶進 VM——
		// 那樣測試會在「閘門沒開」的路徑上全部通過，等於沒驗到東西。
		runtime := NewRuntimeState(0)
		runtime.Started = true
		runtime.Memory[loadFilesThreeDGate] = item.gate
		runtime.Memory[loadFilesMapStaleCell] = 0xFF // 腳本先前寫的「地圖要重載」
		result, err := runSubsetWithState(code, 0, 8, nil, false, 1, runtime)
		memory := runtime.Memory
		if err != nil {
			t.Fatalf("%s：%v", item.name, err)
		}
		if !result.LoadFilesRequested {
			t.Fatalf("%s：LOAD FILES 沒有被記下來", item.name)
		}
		if result.LoadFilesLoaded3DMap != item.wantLoaded {
			t.Errorf("%s：走 3D 那一路 ＝ %v，want %v",
				item.name, result.LoadFilesLoaded3DMap, item.wantLoaded)
		}
		if item.wantLoaded {
			if got := memory[loadFilesMapBlockCell]; got != uint16(item.piece) {
				t.Errorf("%s：`4BC5h` ＝ %02X，want %02X", item.name, got, item.piece)
			}
			if got := memory[loadFilesMapStaleCell]; got != 0 {
				t.Errorf("%s：`7EC9h` ＝ %02X，want 0（handler 載完會清掉）", item.name, got)
			}
			continue
		}
		// ⚠ 沒走那一路就**一格都不能動**：`7EC9h` 要維持腳本寫進去的 `FFh`。
		if got := memory[loadFilesMapStaleCell]; got != 0xFF {
			t.Errorf("%s：`7EC9h` 被改成 %02X，應該維持 FF", item.name, got)
		}
		if _, written := memory[loadFilesMapBlockCell]; written {
			t.Errorf("%s：`4BC5h` 不該被寫", item.name)
		}
	}
}

// 兩格都要進 `SaveWrites`——存檔與重播都靠它，只改 map 不記錄的話重載會不一致。
func TestLoadFilesMapCellsReachSaveWrites(t *testing.T) {
	program := []byte{0, 0, 0x21, 0x00, 0x45, 0x00, 0x02, 0x00, 0xFF, 0x00}
	runtime := NewRuntimeState(0)
	runtime.Started = true
	runtime.Memory[loadFilesThreeDGate] = 1
	runtime.Memory[loadFilesMapStaleCell] = 0xFF
	result, err := runSubsetWithState(program, 0, 8, nil, false, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[uint16]uint16{}
	for _, write := range result.SaveWrites {
		seen[write.Address] = write.Value
	}
	if seen[loadFilesMapBlockCell] != 0x45 || seen[loadFilesMapStaleCell] != 0 {
		t.Fatalf("SaveWrites ＝ %+v，want 4BC5h=45h 與 7EC9h=0", result.SaveWrites)
	}
}
