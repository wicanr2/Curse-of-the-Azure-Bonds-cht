package monster

import "testing"

func TestParseItemsAndSignedFields(t *testing.T) {
	data := make([]byte, ItemRecordSize)
	copy(data, "LONG SWORD")
	data[0x2E], data[0x2F], data[0x32] = 0x23, 0x01, 0xFF
	data[0x34], data[0x36], data[0x39] = 1, 1, 2
	data[0x37], data[0x38] = 0x2C, 0x01
	data[0x3A], data[0x3B] = 0x90, 0x01
	data[0x3C], data[0x3D], data[0x3E] = 1, 2, 3
	items, err := ParseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	item := items[0]
	if item.Name != "LONG SWORD" || item.Type != 0x23 || item.Plus != -1 || !item.Readied || !item.Cursed || item.Count != 2 || item.Weight != 300 || item.Value != 400 || item.Affects != [3]uint8{1, 2, 3} {
		t.Fatalf("item=%#v", item)
	}
}

func TestParseAffects(t *testing.T) {
	data := []byte{0x27, 0x34, 0x12, 7, 1, 8, 9, 10, 11}
	affects, err := ParseAffects(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(affects) != 1 || affects[0].Kind != 0x27 || affects[0].Value != 0x1234 || affects[0].Duration != 7 || !affects[0].Active || affects[0].Data != [4]byte{8, 9, 10, 11} {
		t.Fatalf("affects=%#v", affects)
	}
}

func TestRejectsMisalignedRecords(t *testing.T) {
	if _, err := ParseItems([]byte{1}); err == nil {
		t.Fatal("expected item alignment error")
	}
	if _, err := ParseAffects([]byte{1}); err == nil {
		t.Fatal("expected affect alignment error")
	}
}
