package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// newWallPieceState 給一個帶 session 的最小 State：`restoreWallPiecesForLoadedBlock`
// 會讀 session 的記憶體去算 wall set 指派，沒有 session 就走不到。
func newWallPieceState(t *testing.T) *State {
	t.Helper()
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x42: {0, 0}}, 0x42)
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(testCatalog())
	state.session = session
	return &state
}

// 這一組測試釘住「載入原版存檔之後，牆磚選圖補得回來」。
//
// ★ 為什麼值得一支測試。 補不回來的症狀是**沒有症狀**：`LoadPieces` 全零時
// 第一人稱把每一面牆都畫成空氣，天空、地板、UI 全都在，畫面看起來完全正常。
// 第 675 輪就是這樣把一張沒有牆的畫面拿去跟原版比，差了 37.6%，差點被判成
// 「第一人稱畫錯了」（spec 1185）。

// 表收的是「**載檔時真的會重發** `LOAD PIECES` 的段」，鍵是 ECL 段。
//
// ⚠ 這不是「所有發過 `37h` 的段」。段的開頭幾乎都有一道
// `COMPARE 4BF2h, <段號>` 的閘門，載檔時 `4BF2h == 段號` ⇒ 有些段整段前置被
// 跳過，`LOAD PIECES` 一次都不會發（`ecl.ReachableOnLoad`，spec 1185）。
// 那些段要**以存檔記的 `OLDWALL` 為準**，不能拿表去蓋。
func TestECLBlockWallPiecesOnlyCoversLoadTimeReloads(t *testing.T) {
	want := map[eclBlockKey][3]uint16{
		{area: 2, block: 0x01}: {1, 2, 3},
		{area: 2, block: 0x02}: {1, 2, 4},
		{area: 2, block: 0x03}: {1, 2, 4},
		{area: 2, block: 0x04}: {1, 2, 4},
		{area: 3, block: 0x15}: {3, 5, 7},
		{area: 5, block: 0x31}: {12, 255, 255},
		{area: 5, block: 0x35}: {14, 15, 8},
		{area: 6, block: 0x45}: {17, 18, 13},
	}
	if len(eclBlockWallPieces) != len(want) {
		t.Fatalf("表有 %d 筆，預期 %d 筆", len(eclBlockWallPieces), len(want))
	}
	for identity, pieces := range want {
		got, ok := eclBlockWallPieces[identity]
		if !ok || got != pieces {
			t.Errorf("ECL%d/0x%02X ＝ %v（有 %v），預期 %v",
				identity.area, identity.block, got, ok, pieces)
		}
	}
}

// ⚠ 這一支是**查表的鍵為什麼不能用地圖**的理由。`ECL2` 的 `0x01`（提爾佛頓）
// 與 `0x03`（下水道）同屬 GEO2，牆磚選圖卻不同（`1,2,3` vs `1,2,4`）——只差
// 槽 3 那一塊。第 749 輪就是這一塊：載檔時以存檔為準的話槽 3 會用成提爾佛頓的
// `3`，44 張畫面每一張差約 2,900 格，而畫面看起來完全正常（spec 1185）。
func TestSameAreaBlocksDifferInWallPieces(t *testing.T) {
	tilverton, firstOK := eclBlockWallPieces[eclBlockKey{area: 2, block: 0x01}]
	sewers, secondOK := eclBlockWallPieces[eclBlockKey{area: 2, block: 0x03}]
	if !firstOK || !secondOK {
		t.Fatalf("ECL2 的 0x01／0x03 應該都在表裡：%v %v", firstOK, secondOK)
	}
	if tilverton == sewers {
		t.Fatalf("0x01 與 0x03 的牆磚選圖相同（都是 %v）：查表的鍵改用地圖也不會錯了，"+
			"但那表示上游資料變了，要重新確認", tilverton)
	}
}

// 載入之後補上；已經有值的不覆蓋。
func TestRestoreWallPiecesOnlyFillsAnEmptySelector(t *testing.T) {
	state := newWallPieceState(t)
	state.Area.GameArea = 2
	state.Area.LastECLBlockID = 0x01
	state.LoadPieces = [3]uint16{}
	state.restoreWallPiecesForLoadedBlock()
	if state.LoadPieces != [3]uint16{1, 2, 3} {
		t.Fatalf("補完之後 ＝ %v，want [1 2 3]", state.LoadPieces)
	}

	// 腳本已經發過的值不能被表蓋掉：表是「沒有值時的後備」，不是真相來源。
	state.LoadPieces = [3]uint16{9, 9, 9}
	state.restoreWallPiecesForLoadedBlock()
	if state.LoadPieces != [3]uint16{9, 9, 9} {
		t.Fatalf("已有的選圖被表蓋掉了：%v", state.LoadPieces)
	}
}

// 表裡沒有的段：安靜不動，不能 panic。
func TestRestoreWallPiecesIgnoresUnknownBlocks(t *testing.T) {
	state := newWallPieceState(t)
	state.Area.GameArea = 1
	state.Area.LastECLBlockID = 0x50
	state.LoadPieces = [3]uint16{}
	state.restoreWallPiecesForLoadedBlock()
	if state.LoadPieces != [3]uint16{} {
		t.Fatalf("世界地圖段不該補到任何值：%v", state.LoadPieces)
	}
}
