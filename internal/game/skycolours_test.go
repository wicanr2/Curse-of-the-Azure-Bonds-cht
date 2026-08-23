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

// TestSkyColourTableExcludesGatedBlocks 釘住「被閘門擋掉的段不進表」。
//
// ★ 段的開頭幾乎都有一道 `COMPARE 4BF2h, <自己的段號>` 的閘門，而**兩個方向
// 都有人用**：
//
//	ECL4/0x21  IF =  → GOTO   本來就在這一段 ⇒ 跳過整段前置（含天空）
//	ECL3/0x11  IF =  → EXIT   本來就在這一段 ⇒ 直接結束
//	ECL5/0x33  IF <> → GOTO   不在這一段才跳走 ⇒ 前置照跑
//	ECL3/0x15  （沒有閘門）    一定跑
//
// `4BF2h` 是存檔的 `LastECLBlockID`，而套這張表的時機正是「載入一份記著這一段的
// 存檔」⇒ 那一刻 `4BF2h == 段號`，前兩種的天空寫入**一條都不會跑**。
//
// ⚠ 只掃「有沒有這條 SAVE」會多收八段。第 749 輪就是這樣把四張圖的天空改錯
// （`geo2-b03`／`geo3-b10`／`geo4-b21`／`geo4-b25` 合計 32 萬格），而症狀是
// 整片純色換成另一片純色——**看不出異常**（spec 1185）。
func TestSkyColourTableExcludesGatedBlocks(t *testing.T) {
	// 這幾段有天空 `SAVE`，但載檔時走不到。
	for _, gated := range []eclBlockKey{
		{area: 2, block: 0x02}, {area: 2, block: 0x03}, {area: 2, block: 0x04},
		{area: 3, block: 0x10}, {area: 3, block: 0x11}, {area: 3, block: 0x12},
		{area: 4, block: 0x21}, {area: 4, block: 0x22}, {area: 4, block: 0x25},
	} {
		if colours, ok := eclBlockSkyColours[gated]; ok {
			t.Errorf("ECL%d/0x%02X 的天空寫入被閘門擋著，不該進表（拿到 %+v）",
				gated.area, gated.block, colours)
		}
	}
	// 這幾段走得到，一定要在表裡。
	for _, reachable := range []eclBlockKey{
		{area: 3, block: 0x15}, // 完全沒有閘門
		{area: 5, block: 0x33}, // 閘門是 `IF <>` ⇒ 相等時照跑
		{area: 5, block: 0x35},
		{area: 6, block: 0x45},
	} {
		if _, ok := eclBlockSkyColours[reachable]; !ok {
			t.Errorf("ECL%d/0x%02X 載檔時走得到天空寫入，卻不在表裡",
				reachable.area, reachable.block)
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
