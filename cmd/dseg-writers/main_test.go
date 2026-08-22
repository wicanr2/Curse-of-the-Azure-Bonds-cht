package main

import "testing"

// wanted 是三個地圖暫存器（spec 1172）。
func wantedCells() map[uint16]bool {
	return map[uint16]bool{0x720F: true, 0x7210: true, 0x7211: true}
}

func kinds(t *testing.T, code []byte) map[Kind][]Access {
	t.Helper()
	// 掃描器每一輪要看得到 5 個位元組，尾巴補白才不會漏掉最後一條。
	return scan(append(code, 0, 0, 0, 0, 0, 0), wantedCells())
}

// ★ 分類錯一格就會憑空生出或抹掉一個寫入者。本工具第一版把 `cmpb $0, 0x7211`
// （`80 3E 11 72 00`，`/7` ＝ CMP）算成寫入，於是 THREED 被列成地圖暫存器的
// 寫入者——那正是「引擎會自己放置隊伍」這個錯誤結論需要的證據。
func TestScanSeparatesCompareFromWrite(t *testing.T) {
	got := kinds(t, []byte{0x80, 0x3E, 0x11, 0x72, 0x00})
	if len(got[KindWrite]) != 0 {
		t.Fatalf("`cmpb` 不寫記憶體：%+v", got[KindWrite])
	}
	if len(got[KindRead]) != 1 || got[KindRead][0].Address != 0x7211 {
		t.Fatalf("`cmpb` 應該算讀取：%+v", got[KindRead])
	}

	// 同一組編碼的 `/5`（SUB）就是真的寫入。
	got = kinds(t, []byte{0x80, 0x2E, 0x11, 0x72, 0x01})
	if len(got[KindWrite]) != 1 || got[KindWrite][0].Mnemoni != "sub [addr],imm" {
		t.Fatalf("`subb` 應該算寫入：%+v", got[KindWrite])
	}
}

// `A0`／`A1` 是載入、`A2`／`A3` 是儲存。這兩對差一個位元，而它們決定
// 「這個模組到底碰不碰這格」——INTERPET 全部是 `A0`，所以它只讀不寫。
func TestScanSeparatesAccumulatorLoadFromStore(t *testing.T) {
	got := kinds(t, []byte{0xA0, 0x0F, 0x72})
	if len(got[KindRead]) != 1 || len(got[KindWrite]) != 0 {
		t.Fatalf("`mov al,[720F]` 應該只算讀取：%+v", got)
	}
	got = kinds(t, []byte{0xA2, 0x0F, 0x72})
	if len(got[KindWrite]) != 1 || len(got[KindRead]) != 0 {
		t.Fatalf("`mov [720F],al` 應該只算寫入：%+v", got)
	}
}

// ⚠ 「取位址」那一欄是掃不到的指標寫入的替代證據：Turbo Pascal 的全域要被
// 指標寫，位址一定得先進暫存器。LOADSAVE 的 5 byte 區塊複製就是這個形狀
// （`bf 0f 72` ＝ `mov di, 720Fh`）。
func TestScanReportsAddressTaking(t *testing.T) {
	got := kinds(t, []byte{0xBF, 0x0F, 0x72})
	if len(got[KindAddress]) != 1 || got[KindAddress][0].Address != 0x720F {
		t.Fatalf("`mov di,720Fh` 應該算取位址：%+v", got[KindAddress])
	}
	if len(got[KindWrite]) != 0 || len(got[KindRead]) != 0 {
		t.Fatalf("取位址不該同時算成讀寫：%+v", got)
	}
}

// `MOVEFORWARD` 那一段的形狀：`inc`／`dec` 之後接一條環繞夾回的 `mov imm`。
// 兩者都要算寫入，少算一種會讓「誰會移動隊伍」的答案不完整。
func TestScanCountsIncrementAndDecrementAsWrites(t *testing.T) {
	got := kinds(t, append(append([]byte{}, 0xFE, 0x06, 0x10, 0x72), 0xFE, 0x0E, 0x0F, 0x72))
	if len(got[KindWrite]) != 2 {
		t.Fatalf("`inc`／`dec` 都要算寫入：%+v", got[KindWrite])
	}
}

// 不在查詢清單裡的位址一律不收——否則報表會被無關的存取灌爆，
// 而「合計 N 處」這個數字就失去意義。
func TestScanIgnoresOtherAddresses(t *testing.T) {
	got := kinds(t, []byte{0xA2, 0x00, 0x50})
	if len(got[KindWrite]) != 0 {
		t.Fatalf("`5000h` 不在清單裡：%+v", got[KindWrite])
	}
}
