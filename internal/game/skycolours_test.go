package game

import "testing"

// TestECLBlockSkyColoursMatchOriginalWrites 釘住幾個由原版位元組碼直接讀得出來
// 的選色，讓表被重生時漂移會被抓到。
//
// ★ 挑這幾筆的理由：
//
//   - `ECL2/0x01`（提爾佛頓）室內 `9` 與原版底檔 `workplace/orig-savgamb.dat`
//     的 Area1 `01FCh` 完全一致——**兩個獨立來源**，證明位移與語意都對。
//   - `ECL5/0x35` 室內 `10` 是第 749 輪那 66% 差異的正解（`skyPalette[2]` ＝
//     EGA 4）。
//   - `ECL5/0x33` 室外 `11` 而不是 `8`：`8` 那一條被 `IF >` 守著，靜態決定不了，
//     刻意不進表。
//   - `ECL6/0x45` 只寫室外、沒寫室內 ⇒ 室內要維持 `-1`（沿用存檔的值），不是 0。
func TestECLBlockSkyColoursMatchOriginalWrites(t *testing.T) {
	for _, testCase := range []struct {
		area, block     uint8
		outdoor, indoor int
	}{
		{area: 2, block: 0x01, outdoor: -1, indoor: 9},
		{area: 5, block: 0x33, outdoor: 11, indoor: 10},
		{area: 5, block: 0x35, outdoor: -1, indoor: 10},
		{area: 6, block: 0x45, outdoor: 11, indoor: -1},
	} {
		colours, ok := eclBlockSkyColours[eclBlockKey{area: testCase.area, block: testCase.block}]
		if !ok {
			t.Errorf("ECL%d/0x%02X 不在選色表裡", testCase.area, testCase.block)
			continue
		}
		if colours.Outdoor != testCase.outdoor || colours.Indoor != testCase.indoor {
			t.Errorf("ECL%d/0x%02X 選色是 室外 %d／室內 %d，預期 室外 %d／室內 %d",
				testCase.area, testCase.block,
				colours.Outdoor, colours.Indoor, testCase.outdoor, testCase.indoor)
		}
	}
}

// TestSkyColourTableCoversEveryWallPieceBlock 釘住兩張表**涵蓋同一批段**，
// 例外要逐段列名。
//
// ⚠ 兩張表從同一份 ECL corpus、同一條可達性掃出來；其中一張少了某一段，代表
// 掃描漏了，而漏掉的後果是那一段安靜地沿用別章的天空。
//
// ★ 唯一的例外是 `ECL2/0x04`：它的兩處寫入是同一個 `COMPARE 0xC04F 0x95` 的
// 兩條互斥分支（`IF =` → `0Ah`、`IF <>` → `09h`），值要看執行時的狀態，靜態
// 決定不了 ⇒ 刻意不進表。**這裡列名而不是放寬條件**：放寬的話下一個真的漏掉的
// 段就不會被抓到。
func TestSkyColourTableCoversEveryWallPieceBlock(t *testing.T) {
	runtimeDecided := map[eclBlockKey]bool{{area: 2, block: 0x04}: true}
	for identity := range eclBlockWallPieces {
		_, ok := eclBlockSkyColours[identity]
		if ok == runtimeDecided[identity] {
			if ok {
				t.Errorf("ECL%d/0x%02X 標成執行時才決定，卻進了選色表",
					identity.area, identity.block)
			} else {
				t.Errorf("ECL%d/0x%02X 有牆磚選圖卻沒有天空選色",
					identity.area, identity.block)
			}
		}
	}
}

// TestRestoreSkyColoursOverridesTheBaseSaveValue 釘住**還原真的會發生**。
//
// ⚠ 只釘表的內容不夠：第 749 輪表是對的，值卻在啟動流程後段被 game pack 的
// 預設地圖選色蓋回去，量到的差異一格都沒少（spec 1185）。表對 ≠ 有生效。
func TestRestoreSkyColoursOverridesTheBaseSaveValue(t *testing.T) {
	state := newWallPieceState(t)
	// 底檔（提爾佛頓）帶進來的值：室外 11、室內 9。
	state.Area.OutdoorSkyColor, state.Area.IndoorSkyColor = 11, 9
	state.Area.GameArea = 5
	state.Area.LastECLBlockID = 0x35
	state.restoreSkyColoursForLoadedBlock()

	// `ECL5/0x35` 只寫室內 `0Ah`，沒寫室外 ⇒ 室外要維持底檔的 11。
	if state.Area.IndoorSkyColor != 10 {
		t.Errorf("室內選色 %d，預期 10（ECL5/0x35 的 `SAVE 0Ah 4BFEh`）", state.Area.IndoorSkyColor)
	}
	if state.Area.OutdoorSkyColor != 11 {
		t.Errorf("室外選色 %d，預期維持底檔的 11", state.Area.OutdoorSkyColor)
	}
	// ECL 記憶體那兩格也要跟著改：腳本後續讀的是它們，不是 `Area`。
	if value, ok := state.session.MemoryValue(indoorSkyColourCell); !ok || value != 10 {
		t.Errorf("ECL 格 4BFEh ＝ %d（有值 %v），預期 10", value, ok)
	}
}

// TestRestoreSkyColoursLeavesUnknownBlocksAlone 釘住「表裡沒有的段不要亂動」。
func TestRestoreSkyColoursLeavesUnknownBlocksAlone(t *testing.T) {
	state := newWallPieceState(t)
	state.Area.OutdoorSkyColor, state.Area.IndoorSkyColor = 11, 9
	state.Area.GameArea = 2
	state.Area.LastECLBlockID = 0x04 // 執行時才決定，刻意不進表
	state.restoreSkyColoursForLoadedBlock()
	if state.Area.OutdoorSkyColor != 11 || state.Area.IndoorSkyColor != 9 {
		t.Errorf("表裡沒有的段被改成 室外 %d／室內 %d，應該原封不動",
			state.Area.OutdoorSkyColor, state.Area.IndoorSkyColor)
	}
}
