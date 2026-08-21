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
	affects := []AffectRecord{{Kind: 0x27, Value: 3, Duration: 3, Strength: 2, Raw4: 1, Active: true, Data: [4]byte{1, 2, 3, 4}}}
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

type testItemText map[string]string

func (c testItemText) Text(key, fallback string) string {
	if value := c[key]; value != "" {
		return value
	}
	return fallback
}

func TestLocalizedItemNameComposesTypedFields(t *testing.T) {
	text := testItemText{
		"item_type_1C": "bolt", "item_type_24": "sword",
		"item_name_24": "sword", "item_name_A2": "+1",
		"item_unknown": "unknown(0x%02X)", "item_count": "%d %s",
	}
	// 三個名稱編號都是 0：退回類別名，而數量照原作**前綴**（`10 Arrow`）。
	if got := LocalizedItemName(ItemRecord{Type: 28, Count: 9}, text); got != "9 bolt" {
		t.Fatalf("quarrel name=%q", got)
	}
	if got := LocalizedItemName(ItemRecord{Type: 0xEE}, text); got != "unknown(0xEE)" {
		t.Fatalf("unknown name=%q", got)
	}
	// 成分由 `+31h` 排到 `+2Fh`：`{+1, Long Sword, 0}` 讀成 `sword` 再接 `+1`。
	item := ItemRecord{Type: 36, NameNumbers: [3]uint8{0xA2, 0x24, 0}}
	if got := LocalizedItemName(item, text); got != "sword+1" {
		t.Fatalf("visible name numbers=%q", got)
	}
	item.HiddenNameFlags = 4 // 藏住第一個名稱編號（未鑑定的加值）
	if got := LocalizedItemName(item, text); got != "sword" {
		t.Fatalf("hidden name number=%q", got)
	}
}

// 名字**完全**來自三個名稱編號：類別名一個字都不該冒出來（spec 1178）。
// ⚠ 這條與上一個測試不同——上面驗的是「有翻譯時組得對」，這裡驗的是
// 「類別名不會偷偷混進去」。原版 253 件物品每一件都至少有一個名稱編號。
func TestLocalizedItemNameIgnoresTheItemTypeWhenNameNumbersExist(t *testing.T) {
	text := testItemText{
		"item_type_46": "雜項", "item_name_6C": "伊奧恩石", "item_name_74": "‧深紅",
	}
	item := ItemRecord{Type: 0x46, NameNumbers: [3]uint8{0x74, 0x6C, 0}}
	if got := LocalizedItemName(item, text); got != "伊奧恩石‧深紅" {
		t.Fatalf("名稱 ＝ %q，預期 %q", got, "伊奧恩石‧深紅")
	}
}

// `+32h`（Plus）與 `+36h`（Cursed）不進名字：加值是走名稱編號 `A2h..A6h`，
// 原作 72 件帶 Plus 的物品裡有 54 件從頭到尾不顯示加值（spec 1178）。
func TestLocalizedItemNameDoesNotPrintPlusOrCursed(t *testing.T) {
	text := testItemText{"item_type_24": "sword", "item_name_24": "sword"}
	item := ItemRecord{Type: 0x24, NameNumbers: [3]uint8{0, 0, 0x24}, Plus: 3, Cursed: true}
	if got := LocalizedItemName(item, text); got != "sword" {
		t.Fatalf("名稱 ＝ %q，預期 %q", got, "sword")
	}
}

// 三個成分而中間是連接詞時把前後對調：英文 `Wand of Fireballs`，
// 中文要讀成「火球之魔杖」而不是「魔杖之火球」。
func TestLocalizedItemNameSwapsAroundConnectors(t *testing.T) {
	text := testItemText{
		"item_name_45": "魔杖", "item_name_A7": "之", "item_name_F2": "火球",
		"item_name_6C": "伊奧恩石", "item_name_74": "‧深紅", "item_name_40": "藥水",
	}
	item := ItemRecord{Type: 0x4F, NameNumbers: [3]uint8{0xF2, 0xA7, 0x45}}
	if got := LocalizedItemName(item, text); got != "火球之魔杖" {
		t.Fatalf("名稱 ＝ %q，預期 %q", got, "火球之魔杖")
	}
	// 中間不是連接詞就照原順序，不能一律對調。
	plain := ItemRecord{Type: 0x46, NameNumbers: [3]uint8{0x74, 0x6C, 0x40}}
	if got := LocalizedItemName(plain, text); got != "藥水伊奧恩石‧深紅" {
		t.Fatalf("名稱 ＝ %q，預期 %q", got, "藥水伊奧恩石‧深紅")
	}
}

func TestLocalizedAffectNameUsesRawStableKind(t *testing.T) {
	text := testItemText{
		"affect_kind_18": "detect",
		"affect_kind_5A": "acid",
		"affect_unknown": "unknown(0x%02X)",
	}
	if got := LocalizedAffectName(AffectRecord{Kind: 0x18}, text); got != "detect" {
		t.Fatalf("known effect=%q", got)
	}
	if got := LocalizedAffectName(AffectRecord{Kind: 0x5A}, text); got != "acid" {
		t.Fatalf("second known effect=%q", got)
	}
	if got := LocalizedAffectName(AffectRecord{Kind: 0xEE}, text); got != "unknown(0xEE)" {
		t.Fatalf("unknown effect=%q", got)
	}
}
