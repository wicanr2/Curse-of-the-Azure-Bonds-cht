package party

import "testing"

// 台帳必須蓋滿整份記錄——少一段會讓「還剩多少不知道」的數字偏低。
func TestDOSPlayerRecordFieldsCoverEveryByte(t *testing.T) {
	if err := ValidateDOSPlayerRecordFields(); err != nil {
		t.Fatal(err)
	}
	counts := map[DOSRecordFieldStatus]int{}
	for _, field := range DOSPlayerRecordFields {
		counts[field.Status] += field.Size
	}
	total := counts[DOSFieldDecoded] + counts[DOSFieldDocumented] + counts[DOSFieldUnknown]
	if total != DOSPlayerRecordSize {
		t.Fatalf("三態合計 %d，記錄是 %d bytes", total, DOSPlayerRecordSize)
	}
	t.Logf("decoded=%d documented=%d unknown=%d（共 %d bytes）",
		counts[DOSFieldDecoded], counts[DOSFieldDocumented], counts[DOSFieldUnknown], total)
}
