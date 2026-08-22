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

// 表本身要有內容，而且鍵是 ECL 段。
func TestECLBlockWallPiecesCoversTheOriginalBlocks(t *testing.T) {
	if len(eclBlockWallPieces) < 19 {
		t.Fatalf("牆磚選圖表只有 %d 筆，原版掃出來是 19 筆", len(eclBlockWallPieces))
	}
	tilverton, ok := eclBlockWallPieces[eclBlockKey{area: 2, block: 0x01}]
	if !ok || tilverton != [3]uint16{1, 2, 3} {
		t.Fatalf("提爾佛頓（ECL2 段 0x01）＝ %v，want [1 2 3]", tilverton)
	}
	sewers, ok := eclBlockWallPieces[eclBlockKey{area: 2, block: 0x03}]
	if !ok || sewers != [3]uint16{1, 2, 4} {
		t.Fatalf("下水道（ECL2 段 0x03）＝ %v，want [1 2 4]", sewers)
	}
}

// ⚠ 這一支是**查表的鍵為什麼不能用地圖**的理由。GEO5 的 `0x31` 與 `0x32` 共用
// 同一張幾何區塊 `0x32`，牆磚選圖卻不同。照地圖查會在這兩塊之一給出自洽但錯的
// 牆——而畫面看起來仍然正常。這兩筆一旦相等，就表示上游資料變了，查表的鍵要
// 重新檢討，不是把這支測試改掉。
func TestGeo5BlocksShareGeometryButNotWallPieces(t *testing.T) {
	first, firstOK := eclBlockWallPieces[eclBlockKey{area: 5, block: 0x31}]
	second, secondOK := eclBlockWallPieces[eclBlockKey{area: 5, block: 0x32}]
	if !firstOK || !secondOK {
		t.Fatalf("GEO5 的 0x31／0x32 應該都在表裡：%v %v", firstOK, secondOK)
	}
	if first == second {
		t.Fatalf("0x31 與 0x32 的牆磚選圖相同（都是 %v）：查表的鍵改用地圖也不會錯了，"+
			"但那表示上游資料變了，要重新確認", first)
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
