package main

import "testing"

// 台帳與量測必須對得起來，而且這個檢查要在 `go test ./...` 裡——
// 只有人去跑報告才會發現的漂移，等於沒有閘。
//
// 兩個方向都擋：量到有讀卻標成 `documented`／`unknown`（台帳低估），
// 以及標成 `decoded` 卻沒有任何基準記錄量到（台帳高估，或基準記錄不夠）。
func TestDOSRecordLedgerAgreesWithTheMutationProbe(t *testing.T) {
	result, err := analyze()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Mismatches) != 0 {
		for _, line := range result.Mismatches {
			t.Errorf("對帳不符：%s", line)
		}
		t.FailNow()
	}
	if result.ConsumedAll == 0 {
		t.Fatal("一個位元組都沒量到，突變探針失效了")
	}
	t.Logf("decoded=%d documented=%d unknown=%d，量到有讀 %d／%d",
		result.ByStatus["decoded"], result.ByStatus["documented"],
		result.ByStatus["unknown"], result.ConsumedAll, result.RecordSize)
}
