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

func TestEncodeEquipmentSidecarsRoundTripKnownFields(t *testing.T) {
	items := []ItemRecord{{Name: "長劍", Type: 36, NameNumbers: [3]uint8{1, 2, 3}, Plus: -1, PlusSave: 2, Readied: true, HiddenNameFlags: 4, Cursed: true, Weight: 7, Count: 2, Value: 123, Affects: [3]uint8{8, 9, 10}}}
	data, err := EncodeItems(items)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ParseItems(data)
	if err != nil || len(decoded) != 1 || decoded[0] != items[0] {
		t.Fatalf("item round trip=%#v err=%v", decoded, err)
	}
	affects := []AffectRecord{{Kind: 0x27, Value: 3, Duration: 3, Strength: 2, Active: true, Data: [4]byte{1, 2, 3, 4}}}
	data = EncodeAffects(affects)
	parsed, err := ParseAffects(data)
	if len(data) != AffectRecordSize || err != nil || len(parsed) != 1 || parsed[0] != affects[0] || parsed[0].Value != 3 {
		t.Fatalf("effect round trip=%#v bytes=%x err=%v", parsed, data, err)
	}
}

func TestParseBaseItemsRejectsMalformedData(t *testing.T) {
	if _, err := ParseBaseItems([]byte{0, 1, 2}); err == nil {
		t.Fatal("expected malformed ITEMS error")
	}
}

func TestEquipmentEffectUsesBaseDamageAndPackedAC(t *testing.T) {
	data := make([]byte, BaseItemHeaderSize+2*BaseItemRecordSize)
	data[BaseItemHeaderSize+2] = 1
	data[BaseItemHeaderSize+3] = 6
	data[BaseItemHeaderSize+4] = 1
	data[BaseItemHeaderSize+9] = 2
	data[BaseItemHeaderSize+10] = 4
	data[BaseItemHeaderSize+11] = 2
	data[BaseItemHeaderSize+6] = 183
	data[BaseItemHeaderSize+BaseItemRecordSize+0] = 2
	data[BaseItemHeaderSize+BaseItemRecordSize+1] = 1
	data[BaseItemHeaderSize+BaseItemRecordSize+6] = 129
	catalog, err := ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	weapon, err := (ItemRecord{Type: 0, Plus: 1}).Effect(catalog, false)
	if err != nil {
		t.Fatal(err)
	}
	if weapon.DamageDiceCount != 2 || weapon.DamageDiceSides != 4 || weapon.DamageBonus != 3 || weapon.AttackBonus != 1 {
		t.Fatalf("weapon effect=%#v", weapon)
	}
	armor, err := (ItemRecord{Type: 1}).Effect(catalog, false)
	if err != nil {
		t.Fatal(err)
	}
	if armor.Slot != 2 || armor.ArmorClassImprovement != 1 {
		t.Fatalf("armor effect=%#v", armor)
	}
}

func TestEquipmentEffectProjectsRateOfFireAsAttacksPerTurn(t *testing.T) {
	items := make([]BaseItem, 42)
	items[41] = BaseItem{Type: 41, SmallDamageDice: 1, SmallDamageSides: 6, RateOfFire: 4, Range: 22, AmmunitionType: 11}
	catalog := BaseItemCatalog{Items: items}
	item := ItemRecord{Type: 41, Count: 1}
	effect, err := item.Effect(catalog, false)
	if err != nil {
		t.Fatal(err)
	}
	if effect.AttacksPerTurn != 2 || effect.AmmunitionType != 11 || effect.WeaponRange != 22 || !effect.MissileWeapon || effect.ThrownWeapon {
		t.Fatalf("effect=%+v", effect)
	}
}

func TestDecodeConsumableKindsAndProperties(t *testing.T) {
	catalog, err := ParseBaseItems(make([]byte, BaseItemHeaderSize+128*BaseItemRecordSize))
	if err != nil {
		t.Fatal(err)
	}
	scroll, err := (ItemRecord{Type: 60, Affects: [3]uint8{3, 0x18, 0}}).DecodeConsumable(catalog)
	if err != nil || scroll.Kind != ConsumableScroll || len(scroll.SpellIDs) != 2 || scroll.SpellIDs[1] != 0x18 {
		t.Fatalf("scroll=%#v err=%v", scroll, err)
	}
	potion, err := (ItemRecord{Type: 84}).DecodeConsumable(catalog)
	if err != nil || potion.Kind != ConsumablePotion || potion.ChargesBefore != 1 {
		t.Fatalf("potion=%#v err=%v", potion, err)
	}
	wand, err := (ItemRecord{Type: 78, Affects: [3]uint8{4, 0x5A, 0}}).DecodeConsumable(catalog)
	if err != nil || wand.Kind != ConsumableCharged || wand.EffectID != 0x5A || wand.ChargesBefore != 4 || wand.ChargesAfter != 4 {
		t.Fatalf("wand=%#v err=%v", wand, err)
	}
	if _, err := (ItemRecord{Type: 1}).DecodeConsumable(catalog); err == nil {
		t.Fatal("weapon should not be decoded as consumable")
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
	if len(affects) != 1 || affects[0].Kind != 0x27 || affects[0].Value != 0x1234 || affects[0].Duration != 0x1234 || affects[0].Strength != 7 || !affects[0].Active || affects[0].Data != [4]byte{8, 9, 10, 11} {
		t.Fatalf("affects=%#v", affects)
	}
}

func TestAdvanceAffectsExpiresFiniteAndPreservesPermanent(t *testing.T) {
	input := []AffectRecord{
		{Kind: 1, Duration: 10, Value: 10, Strength: 2},
		{Kind: 2, Duration: 5, Value: 5, Strength: 3},
		{Kind: 3, Duration: 0, Value: 0, Strength: 0xFF},
	}
	output := AdvanceAffects(input, 5)
	if len(output) != 2 || output[0].Duration != 5 || output[0].Value != 5 || output[1].Strength != 0xFF {
		t.Fatalf("advanced affects=%#v", output)
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
	if got := ChineseAffectName(AffectRecord{Kind: 0x27}); got != "加速" {
		t.Fatalf("haste=%q", got)
	}
}
