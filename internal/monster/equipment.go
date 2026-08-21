package monster

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/origtext"
)

const (
	ItemRecordSize     = 0x3F
	AffectRecordSize   = 9
	BaseItemHeaderSize = 2
	BaseItemRecordSize = 16
)

// BaseItem is one descriptor from the original ITEMS table. The table is
// indexed by the item type byte stored in an inventory item record.
type BaseItem struct {
	Type               uint8
	Slot               uint8
	HandsRequired      uint8
	LargeDamageDice    uint8
	LargeDamageSides   uint8
	LargeDamageBonus   int8
	RateOfFire         uint8 // stored as rate multiplied by two
	ACAdjustment       uint8
	WeaponType         uint8 // 0 slashing, 1 piercing, 0x80 bashing
	Unknown08          uint8
	SmallDamageDice    uint8
	SmallDamageSides   uint8
	SmallDamageBonus   int8
	Range              uint8
	ClassUsabilityMask uint8
	AmmunitionType     uint8
	Unknown0F          uint8
}

// BaseItemCatalog is the decompressed ITEMS table plus its two-byte header.
// Header is preserved because the reference uses it as part of the data file,
// but its meaning is not needed to index the 128 descriptors.
type BaseItemCatalog struct {
	Header uint16
	Items  []BaseItem
}

// EquipmentEffect is the small, renderer-neutral projection needed by the
// current combat adapter. Charges and magical effect stacks remain separate.
type EquipmentEffect struct {
	Slot                  uint8
	AttackBonus           int
	ArmorClassImprovement int
	DamageDiceCount       int
	DamageDiceSides       int
	DamageBonus           int
	AttacksPerTurn        int
	AmmunitionType        uint8
	MovementAllowance     int
	WeaponRange           int
	MissileWeapon         bool
	ThrownWeapon          bool
}

// ArmorClassImprovement decodes the reference's packed AC adjustment.
func (item BaseItem) ArmorClassImprovement() int {
	if item.ACAdjustment >= 178 {
		return int(item.ACAdjustment) - 178
	}
	if item.ACAdjustment >= 128 {
		return int(item.ACAdjustment) - 128
	}
	return 0
}

// MovementAllowance returns the RuleBook maximum squares for the observed
// armor item types. Zero means this descriptor does not impose an armor cap.
func (item BaseItem) MovementAllowance() int {
	switch item.Type {
	case 50: // leather
		return 12
	case 51, 52, 53, 55, 57: // padded, studded, ring, chain, banded
		return 9
	case 54, 56, 58: // scale, splint, plate
		return 6
	default:
		return 0
	}
}

// IsMissileWeapon 用**原作自己的判斷**：類別表 `+0Eh` 的 bit 3（spec 1120）。
//
// ★ 這裡原本寫的是「類別 41..47」——那是照觀察到的弓／弩／投石索列舉出來的。
// 自動換裝那一支讀出來之後才知道原作查的是旗標位元，而設了 bit 3 的類別有 18 個：
// 除了 41..47 還有飛鏢（9）、21、28、85..88、98、100、101、127。
// 列舉法漏掉的那 11 個在戰鬥中會被當成近戰武器。
func (item BaseItem) IsMissileWeapon() bool {
	return item.AmmunitionType&baseItemFlagFired != 0
}

// IsThrownWeapon 同理用 bit 4。原本只認飛鏢（類別 9），實際有 13 個類別設了這個位元。
//
// ⚠ 兩者**不互斥**：飛鏢兩個位元都有（丟出去、而且有射速）。
func (item BaseItem) IsThrownWeapon() bool {
	return item.AmmunitionType&baseItemFlagThrown != 0
}

// UsableByMask reports whether the supplied reference class bit is allowed:
// magic-user=1, cleric=2, thief=4, fighter=8, druid=16, monk=32,
// paladin=64, ranger=128.
func (item BaseItem) UsableByMask(classBit uint8) bool {
	return classBit != 0 && item.ClassUsabilityMask&classBit != 0
}

// Effect projects one inventory item onto combat stats. large selects the
// large-target damage columns; the current party slice uses false.
func (item ItemRecord) Effect(catalog BaseItemCatalog, large bool) (EquipmentEffect, error) {
	base, ok := catalog.Lookup(item.Type)
	if !ok {
		return EquipmentEffect{}, fmt.Errorf("item type 0x%02X is outside base catalog", item.Type)
	}
	effect := EquipmentEffect{
		Slot:                  base.Slot,
		AttackBonus:           item.Plus,
		ArmorClassImprovement: base.ArmorClassImprovement(),
		AmmunitionType:        base.AmmunitionType,
		MovementAllowance:     base.MovementAllowance(),
		WeaponRange:           int(base.Range),
		MissileWeapon:         base.IsMissileWeapon(),
		ThrownWeapon:          base.IsThrownWeapon(),
	}
	if base.RateOfFire > 0 {
		effect.AttacksPerTurn = int(base.RateOfFire) / 2
		if effect.AttacksPerTurn < 1 {
			effect.AttacksPerTurn = 1
		}
	}
	if large {
		effect.DamageDiceCount = int(base.LargeDamageDice)
		effect.DamageDiceSides = int(base.LargeDamageSides)
		effect.DamageBonus = int(base.LargeDamageBonus) + item.Plus
	} else {
		effect.DamageDiceCount = int(base.SmallDamageDice)
		effect.DamageDiceSides = int(base.SmallDamageSides)
		effect.DamageBonus = int(base.SmallDamageBonus) + item.Plus
	}
	return effect, nil
}

// ParseBaseItems decodes the standalone ITEMS member. Its current image is
// 2050 bytes: a two-byte header followed by 128 fixed-size descriptors.
func ParseBaseItems(data []byte) (BaseItemCatalog, error) {
	if len(data) < BaseItemHeaderSize || (len(data)-BaseItemHeaderSize)%BaseItemRecordSize != 0 {
		return BaseItemCatalog{}, fmt.Errorf("base ITEMS data is %d bytes, not a header plus %d-byte records", len(data), BaseItemRecordSize)
	}
	count := (len(data) - BaseItemHeaderSize) / BaseItemRecordSize
	items := make([]BaseItem, count)
	for index := 0; index < count; index++ {
		offset := BaseItemHeaderSize + index*BaseItemRecordSize
		item := data[offset : offset+BaseItemRecordSize]
		items[index] = BaseItem{
			Type: uint8(index), Slot: item[0], HandsRequired: item[1],
			LargeDamageDice: item[2], LargeDamageSides: item[3], LargeDamageBonus: int8(item[4]),
			RateOfFire: item[5], ACAdjustment: item[6], WeaponType: item[7], Unknown08: item[8],
			SmallDamageDice: item[9], SmallDamageSides: item[10], SmallDamageBonus: int8(item[11]),
			Range: item[12], ClassUsabilityMask: item[13], AmmunitionType: item[14], Unknown0F: item[15],
		}
	}
	return BaseItemCatalog{Header: binary.LittleEndian.Uint16(data[:2]), Items: items}, nil
}

// 卷軸的判別（spec 1171）。
//
// ★ 原作判「這是不是卷軸」看的是**類別表的第一格**（裝備槽），不是物品類別
// 本身：`byte[5CF6h ＋ 類別 × 16]` 落在 `0Bh`..`0Dh` 才是卷軸。
//
// ⚠ 用物品類別 `3Ch`..`3Eh` 去判會多抓三件——`3Ch` 的槽是 `0Ah`，名字雖然叫
// `Scroll`，原作把它當**充能物品**（效果 `5Fh`／`60h`）。
const (
	scrollSlotLow  = 0x0B
	scrollSlotHigh = 0x0D
	// ClericalScrollSlot 是牧師卷軸的槽。牧師讀這一種不需要法術辨識
	// （`overlay-22:08A1h` 只對這個槽放行）。
	ClericalScrollSlot = 0x0C
)

// ItemSlot 回傳類別表的第一格（裝備槽）。
func (c BaseItemCatalog) ItemSlot(itemType uint8) (uint8, bool) {
	base, ok := c.Lookup(itemType)
	if !ok {
		return 0, false
	}
	return base.Slot, true
}

// IsScroll 重現原作的卷軸判別。目錄查不到那個類別時回 false——原作的那一支對
// NIL 物品也是回 0。
func (c BaseItemCatalog) IsScroll(itemType uint8) bool {
	slot, ok := c.ItemSlot(itemType)
	return ok && slot > scrollSlotLow-1 && slot < scrollSlotHigh+1
}

// Lookup returns a descriptor by item type without allowing an out-of-range
// type byte to panic a save/import or corrupted-file path.
func (c BaseItemCatalog) Lookup(itemType uint8) (BaseItem, bool) {
	if int(itemType) >= len(c.Items) {
		return BaseItem{}, false
	}
	return c.Items[itemType], true
}

type ItemRecord struct {
	Name            string
	Type            uint8
	NameNumbers     [3]uint8
	Plus            int
	PlusSave        uint8
	Readied         bool
	HiddenNameFlags uint8
	Cursed          bool
	Weight          int16
	Count           uint8
	Value           int16
	Affects         [3]uint8
}

type ConsumableKind uint8

const (
	ConsumableScroll ConsumableKind = iota + 1
	ConsumablePotion
	ConsumableCharged
)

// ConsumableUse is a data signal for the later spell/effect engine. SpellIDs
// are populated for scrolls; charged items expose the effect byte and charge
// transition without applying any AD&D rule here.
type ConsumableUse struct {
	Kind          ConsumableKind
	ItemType      uint8
	EffectID      uint8
	SpellIDs      []uint8
	ChargesBefore int
	ChargesAfter  int
}

// DecodeConsumable interprets the item properties documented by the original
// item format. Types 60-62 are scrolls, 71/84 are one-use potions, and 78/79
// are charged wands. The catalog lookup keeps corrupted item types bounded.
func (item ItemRecord) DecodeConsumable(catalog BaseItemCatalog) (ConsumableUse, error) {
	if _, ok := catalog.Lookup(item.Type); !ok {
		return ConsumableUse{}, fmt.Errorf("item type 0x%02X is outside base catalog", item.Type)
	}
	use := ConsumableUse{ItemType: item.Type}
	switch item.Type {
	case 60, 61, 62:
		use.Kind = ConsumableScroll
		for _, spellID := range item.Affects {
			if spellID != 0 {
				use.SpellIDs = append(use.SpellIDs, spellID)
			}
		}
		use.ChargesBefore, use.ChargesAfter = 1, 0
	case 71, 84:
		use.Kind = ConsumablePotion
		use.ChargesBefore, use.ChargesAfter = 1, 0
	case 78, 79:
		use.Kind = ConsumableCharged
		use.EffectID = item.Affects[1]
		use.ChargesBefore = int(item.Affects[0])
		use.ChargesAfter = use.ChargesBefore
	default:
		return ConsumableUse{}, fmt.Errorf("item type 0x%02X is not a supported consumable", item.Type)
	}
	return use, nil
}

type AffectRecord struct {
	Kind uint8
	// Value is retained as the raw little-endian duration for CLI/API
	// compatibility. Duration names the same documented field explicitly.
	Value    uint16
	Duration uint16
	Strength uint8
	Raw4     uint8
	Active   bool
	// Data preserves the serialized runtime linked-list pointer at bytes 5..8.
	// LOADMONSTER copies each nine-byte record, then clears these four bytes and
	// rebuilds the list. They are not spell parameters and gameplay code must
	// not assign them effect-specific semantics.
	Data [4]byte
}

// TextResolver keeps CoAB translations outside the reusable legacy codecs.
type TextResolver interface {
	Text(key, fallback string) string
}

// itemNameConnectors 是那幾個「X of Y」的連接詞（`of`、`Of Prot.`、各神祇的
// `Of ...`）。原作把它們當成一般成分排在中間，英文因此讀成 `Wand of Fireballs`。
//
// ⚠ 中文的修飾語在中心語**前面**，照原順序會排成「魔杖之火球」。所以三個成分而
// 中間是連接詞時，把前後兩段對調 → 「火球之魔杖」。這是**翻譯層**的決定，不動
// 任何原始欄位；成分本身與原作逐一對應（spec 1178）。
var itemNameConnectors = map[uint8]bool{
	0xA7: true, // of
	0xC8: true, // Of Seeking
	0xE0: true, // Of Prot.
	0xEB: true, // Of Maglubiyet
	0xF9: true, // Of Tyr
	0xFA: true, // Of Tempus
	0xFB: true, // Of Sune
}

// LocalizedItemName 依原作 `overlay-24:0467h` 的組名規則產生顯示名稱。
//
// ★ 名稱**完全由三個名稱編號組成**，物品類別（`+2Eh`）一個字都不出現
// （spec 1178）。類別名只在三個編號都看不見時當保險絲用——原版 253 件物品
// 沒有一件走到那條路。
//
// 規則（原作逐條對應）：
//
//	`+39h` Count > 0     前面加上數量（原作連 `1` 都印，`1 Oil`）
//	成分順序             `+31h` → `+30h` → `+2Fh`（原作的迴圈由 3 數到 1）
//	`+35h` 的 bit 3−N    設起來就藏住第 N 個成分（未鑑定的魔法屬性）
//
// ⚠ **`+32h`（Plus）與 `+36h`（Cursed）不進名字。** 加值是走名稱編號
// `A2h..A6h`（`+1`..`+5`）那幾個**可以被藏起來**的成分；原作 72 件帶 Plus 的
// 物品裡只有 18 件同時帶 `+N` 成分，剩下 54 件在原作永遠不顯示加值。
func LocalizedItemName(item ItemRecord, resolver TextResolver) string {
	parts := make([]string, 0, 3)
	numbers := make([]uint8, 0, 3)
	for slot := 3; slot >= 1; slot-- {
		nameNumber := item.NameNumbers[slot-1]
		if nameNumber == 0 || item.HiddenNameFlags&(1<<(3-slot)) != 0 {
			continue
		}
		translated := resolver.Text(fmt.Sprintf("item_name_%02X", nameNumber), "")
		if translated == "" {
			continue
		}
		parts = append(parts, translated)
		numbers = append(numbers, nameNumber)
	}
	if len(parts) == 3 && itemNameConnectors[numbers[1]] {
		parts[0], parts[2] = parts[2], parts[0]
	}
	base := strings.Join(parts, "")
	if base == "" {
		// 三個成分都看不見或都沒翻譯：退回類別名，再不行才露出原始編號。
		base = resolver.Text(fmt.Sprintf("item_type_%02X", item.Type), "")
		if base == "" {
			return fmt.Sprintf(resolver.Text("item_unknown", "item 0x%02X"), item.Type)
		}
	}
	if item.Count > 0 {
		base = fmt.Sprintf(resolver.Text("item_count", "%d %s"), item.Count, base)
	}
	return base
}

// LocalizedAffectName resolves an observed raw effect kind through game-pack
// locale data. The raw kind remains authoritative and unknown values stay
// visible diagnostics rather than receiving inferred spell semantics.
func LocalizedAffectName(affect AffectRecord, resolver TextResolver) string {
	if name := resolver.Text(fmt.Sprintf("affect_kind_%02X", affect.Kind), ""); name != "" {
		return name
	}
	format := resolver.Text("affect_unknown", "effect 0x%02X")
	return fmt.Sprintf(format, affect.Kind)
}

// ParseOriginalItems 讀**原版**的 `.SWG`／ITEM 位元組：版面與 ParseItems 完全
// 相同，差別只有名字以原版編碼（Big5，ASCII 相容）解讀。
//
// ★ 為什麼要兩支。 這個版面同時被原版資料與 remake 自己的 sidecar 使用，而
// remake 的寫入端寫的是 UTF-8。編碼由**來源**決定不是版面決定——在
// `parseItem` 裡一律當 Big5 會把 remake 自己的存檔讀壞（英文原版看不出來，
// 因為 ASCII 相容；中文版才會炸）。
func ParseOriginalItems(data []byte) ([]ItemRecord, error) {
	items, err := ParseItems(data)
	if err != nil {
		return nil, err
	}
	for index := range items {
		// Go 的 string 保留無效 UTF-8 位元組，所以轉回 []byte 拿得到原始位元組。
		items[index].Name = origtext.Decode([]byte(items[index].Name))
	}
	return items, nil
}

func ParseItems(data []byte) ([]ItemRecord, error) {
	if len(data)%ItemRecordSize != 0 {
		return nil, fmt.Errorf("item data is %d bytes, not a multiple of %d", len(data), ItemRecordSize)
	}
	items := make([]ItemRecord, 0, len(data)/ItemRecordSize)
	for offset := 0; offset < len(data); offset += ItemRecordSize {
		items = append(items, parseItem(data[offset:offset+ItemRecordSize]))
	}
	return items, nil
}

// EncodeItems serializes the documented 0x3F-byte SWG records and leaves no
// undocumented bytes to guess. It is the inverse of ParseItems for fields
// currently understood by the remake.
func EncodeItems(items []ItemRecord) ([]byte, error) {
	data := make([]byte, len(items)*ItemRecordSize)
	for index, item := range items {
		if len([]byte(item.Name)) > 0x2A {
			return nil, fmt.Errorf("item %d name exceeds 0x2A bytes", index)
		}
		if item.Plus < -128 || item.Plus > 127 {
			return nil, fmt.Errorf("item %d plus %d does not fit signed byte", index, item.Plus)
		}
		offset := index * ItemRecordSize
		copy(data[offset:offset+0x2A], []byte(item.Name))
		data[offset+0x2E] = item.Type
		copy(data[offset+0x2F:offset+0x32], item.NameNumbers[:])
		data[offset+0x32] = byte(int8(item.Plus))
		data[offset+0x33] = item.PlusSave
		if item.Readied {
			data[offset+0x34] = 1
		}
		data[offset+0x35] = item.HiddenNameFlags
		if item.Cursed {
			data[offset+0x36] = 1
		}
		binary.LittleEndian.PutUint16(data[offset+0x37:offset+0x39], uint16(item.Weight))
		data[offset+0x39] = item.Count
		binary.LittleEndian.PutUint16(data[offset+0x3A:offset+0x3C], uint16(item.Value))
		copy(data[offset+0x3C:offset+0x3F], item.Affects[:])
	}
	return data, nil
}

func parseItem(data []byte) ItemRecord {
	// ⚠ 這裡刻意不走 origtext：`EncodeItems` 是 remake 自己的寫入端，寫的是
	// UTF-8（`internal/game/creation.go`）。這個版面同時被原版 ITEM DAX 與
	// remake sidecar 使用，編碼由**來源**決定而不是版面決定；讀原版中文資料
	// 要在載入原版檔的那一層解碼，不能在這裡一律當 Big5。
	name := strings.TrimRight(string(data[:0x2A]), "\x00 ")
	return ItemRecord{
		Name: name, Type: data[0x2E],
		NameNumbers: [3]uint8{data[0x2F], data[0x30], data[0x31]},
		Plus:        int(int8(data[0x32])), PlusSave: data[0x33],
		Readied: data[0x34] != 0, HiddenNameFlags: data[0x35], Cursed: data[0x36] != 0,
		Weight: int16(binary.LittleEndian.Uint16(data[0x37:0x39])), Count: data[0x39],
		Value:   int16(binary.LittleEndian.Uint16(data[0x3A:0x3C])),
		Affects: [3]uint8{data[0x3C], data[0x3D], data[0x3E]},
	}
}

func ParseAffects(data []byte) ([]AffectRecord, error) {
	if len(data)%AffectRecordSize != 0 {
		return nil, fmt.Errorf("affect data is %d bytes, not a multiple of %d", len(data), AffectRecordSize)
	}
	affects := make([]AffectRecord, 0, len(data)/AffectRecordSize)
	for offset := 0; offset < len(data); offset += AffectRecordSize {
		duration := binary.LittleEndian.Uint16(data[offset+1 : offset+3])
		record := AffectRecord{
			Kind: data[offset], Value: duration, Duration: duration,
			Strength: data[offset+3], Raw4: data[offset+4], Active: data[offset+4] != 0,
		}
		copy(record.Data[:], data[offset+5:offset+9])
		affects = append(affects, record)
	}
	return affects, nil
}

// EncodeAffects serializes the documented 9-byte FX records. Data is kept
// verbatim as the legacy runtime linked-list pointer for round trips.
func EncodeAffects(affects []AffectRecord) []byte {
	data := make([]byte, len(affects)*AffectRecordSize)
	for index, affect := range affects {
		offset := index * AffectRecordSize
		duration := affect.Duration
		if duration == 0 {
			duration = affect.Value
		}
		data[offset] = affect.Kind
		binary.LittleEndian.PutUint16(data[offset+1:offset+3], duration)
		data[offset+3] = affect.Strength
		data[offset+4] = affect.Raw4
		if data[offset+4] == 0 && affect.Active {
			data[offset+4] = 1
		}
		copy(data[offset+5:offset+9], affect.Data[:])
	}
	return data
}

// AdvanceAffects consumes effect duration in minutes. Strength 255 is the
// documented permanent marker and is never expired. The raw effect payload is
// preserved for active records; gameplay application remains an outer layer.
func AdvanceAffects(affects []AffectRecord, minutes uint16) []AffectRecord {
	if minutes == 0 {
		return append([]AffectRecord(nil), affects...)
	}
	active := make([]AffectRecord, 0, len(affects))
	for _, affect := range affects {
		if affect.Strength != 0xFF {
			if affect.Duration <= minutes {
				continue
			}
			affect.Duration -= minutes
			affect.Value = affect.Duration
		}
		active = append(active, affect)
	}
	return active
}
