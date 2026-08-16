package monster

import "testing"

// 兩張台帳都要蓋滿整份記錄。
func TestSidecarRecordFieldsCoverEveryByte(t *testing.T) {
	for _, item := range []struct {
		name   string
		fields []RecordField
		size   int
	}{
		{".SWG 物品記錄", ItemRecordFields, ItemRecordSize},
		{".FX 效果記錄", AffectRecordFields, AffectRecordSize},
	} {
		if err := ValidateRecordFields(item.name, item.fields, item.size); err != nil {
			t.Fatal(err)
		}
		counts := map[RecordFieldStatus]int{}
		for _, field := range item.fields {
			counts[field.Status] += field.Size
		}
		t.Logf("%s：decoded=%d documented=%d unknown=%d（共 %d bytes）",
			item.name, counts[FieldDecoded], counts[FieldDocumented],
			counts[FieldUnknown], item.size)
	}
}

// 名字只到 +29h：`+2Ah` 起是物品鏈的指標，讀進名字在英文原版只會看起來像
// 「名字後面有幾個怪字元」，中文資料則會直接壞掉。
func TestItemNameStopsBeforeTheChainPointer(t *testing.T) {
	data := make([]byte, ItemRecordSize)
	copy(data, "LONG SWORD")
	// 指標欄填上非零位元組。
	data[0x2A], data[0x2B], data[0x2C], data[0x2D] = 0xEF, 0xBE, 0xAD, 0xDE
	items, err := ParseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Name != "LONG SWORD" {
		t.Fatalf("名字=%q，鏈結指標被讀進名字了", items[0].Name)
	}
}
