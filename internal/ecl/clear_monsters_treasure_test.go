package ecl

import "testing"

// `1Ch CLEARMONSTERS` 會把**還沒領走的戰利品堆**一起丟掉（spec 1144 的鄰居：
// 原作 `overlay-02:120Eh` 把 `DS:6F70h` 起 28 bytes 歸零、沿 `DS:6F8Ch` 鏈逐節點
// `FreeMem(63)`，那兩塊分別是七種貨幣／寶石／珠寶的池與 `27h` 串進去的物品節點）。
//
// ⚠ 順序是關鍵，所以兩個方向都要驗：
//   - `TREASURE` → `CLEARMONSTERS`：這一堆被丟掉。
//   - `CLEARMONSTERS` → `TREASURE`：這是 corpus 的慣用法（先清乾淨再擺下一場），
//     新擺的那一堆必須留著。只驗前者的話，把 `1Ch` 寫成「無條件清空 result」
//     也會過，而那會讓每一場遭遇的戰利品都消失。
func TestClearMonstersDropsPendingTreasureInOrder(t *testing.T) {
	// 每個運算元兩個位元組：`<code> <low>`，`code = 0` 是立即數。八個運算元。
	treasure := []byte{0x27,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF}

	t.Run("先發後清就丟掉", func(t *testing.T) {
		block := append([]byte{0, 0}, treasure...)
		block = append(block, 0x1C, 0x00)
		result, err := RunSubset(block, 0, 10)
		if err != nil {
			t.Fatal(err)
		}
		if !result.ClearMonstersRequested {
			t.Error("跑過 `1Ch` 卻沒有標記")
		}
		if len(result.TreasureRequests) != 0 {
			t.Errorf("戰利品堆還有 %d 筆，`1Ch` 應該把它丟掉", len(result.TreasureRequests))
		}
	})

	t.Run("先清後發要留著", func(t *testing.T) {
		block := append([]byte{0, 0, 0x1C}, treasure...)
		block = append(block, 0x00)
		result, err := RunSubset(block, 0, 10)
		if err != nil {
			t.Fatal(err)
		}
		if !result.ClearMonstersRequested {
			t.Error("跑過 `1Ch` 卻沒有標記")
		}
		if len(result.TreasureRequests) != 1 {
			t.Errorf("清完之後擺的那一堆剩 %d 筆，應該是 1 筆", len(result.TreasureRequests))
		}
	})
}
