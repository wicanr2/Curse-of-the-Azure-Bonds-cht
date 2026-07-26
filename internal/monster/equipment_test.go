package monster

import (
	"encoding/binary"
	"testing"
)

func TestParseBaseItemsUsesTypeIndexAndSignedBonuses(t *testing.T) {
	data := make([]byte, BaseItemHeaderSize+2*BaseItemRecordSize)
	binary.LittleEndian.PutUint16(data, 0x76)
	data[BaseItemHeaderSize+0] = 2
	data[BaseItemHeaderSize+1] = 1
	data[BaseItemHeaderSize+2] = 1
	data[BaseItemHeaderSize+3] = 8
	data[BaseItemHeaderSize+4] = 0xFF
	data[BaseItemHeaderSize+5] = 4
	data[BaseItemHeaderSize+6] = 0xB2
	data[BaseItemHeaderSize+7] = 0x80
	data[BaseItemHeaderSize+9] = 1
	data[BaseItemHeaderSize+10] = 6
	data[BaseItemHeaderSize+11] = 2
	data[BaseItemHeaderSize+12] = 6
	data[BaseItemHeaderSize+13] = 0x8F
	data[BaseItemHeaderSize+14] = 0x8A
	data[BaseItemHeaderSize+BaseItemRecordSize+0] = 10

	catalog, err := ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Header != 0x76 || len(catalog.Items) != 2 {
		t.Fatalf("catalog header/items=%#x/%d", catalog.Header, len(catalog.Items))
	}
	item, ok := catalog.Lookup(0)
	if !ok || item.Type != 0 || item.Slot != 2 || item.LargeDamageBonus != -1 || item.WeaponType != 0x80 || item.ClassUsabilityMask != 0x8F {
		t.Fatalf("base item=%#v ok=%t", item, ok)
	}
	if _, ok := catalog.Lookup(2); ok {
		t.Fatal("out-of-range item type should not be found")
	}
}

func TestParseBaseItemsRejectsMalformedData(t *testing.T) {
	if _, err := ParseBaseItems([]byte{0, 1, 2}); err == nil {
		t.Fatal("expected malformed ITEMS error")
	}
}

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

func TestChineseNameUsesObservedItemTypes(t *testing.T) {
	if got := ChineseName(ItemRecord{Type: 28, Count: 9}); got != "弩矢 ×9" {
		t.Fatalf("quarrel name=%q", got)
	}
	if got := ChineseName(ItemRecord{Type: 55}); got != "鏈甲" {
		t.Fatalf("chain name=%q", got)
	}
	if got := ChineseName(ItemRecord{Type: 0xEE}); got != "未翻譯物品(0xEE)" {
		t.Fatalf("unknown name=%q", got)
	}
	if got := ChineseName(ItemRecord{Type: 36}); got != "長劍" {
		t.Fatalf("long sword name=%q", got)
	}
}

func TestChineseAffectNameUsesObservedKinds(t *testing.T) {
	if got := ChineseAffectName(AffectRecord{Kind: 0x18}); got != "偵測隱形" {
		t.Fatalf("detect invisibility=%q", got)
	}
	if got := ChineseAffectName(AffectRecord{Kind: 0x5A}); got != "酸液吐息" {
		t.Fatalf("acid breath=%q", got)
	}
}
