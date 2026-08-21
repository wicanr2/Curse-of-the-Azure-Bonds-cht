package ecl

import "testing"

// 三句旁白的查看順序，逐個距離釘住（spec 1144）。
//
// ⚠ 這條擋的是「循環」被寫成「往下找到底就停」：距離 1 與 2 的原作分支各有一條
// 繞回第 1 句的 `mov [bp+var_43B], 1`，少了它，中距與遠距那兩支在後面幾句是空的
// 時候會挑不到任何一句。
func TestEncounterPromptSlots(t *testing.T) {
	for _, want := range []struct {
		distance int
		order    [3]int
	}{
		{0, [3]int{0, 1, 2}},
		{1, [3]int{1, 2, 0}},
		{2, [3]int{2, 0, 1}},
		// 原作只比 0／1／2；大於 2 會讀到沒寫過的緩衝區，這裡一律夾到 2。
		{3, [3]int{2, 0, 1}},
		{-1, [3]int{0, 1, 2}},
	} {
		if got := EncounterPromptSlots(want.distance); got != want.order {
			t.Errorf("距離 %d 的順序是 %v，應該是 %v", want.distance, got, want.order)
		}
	}
}
