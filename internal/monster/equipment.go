package monster

import (
	"encoding/binary"
	"fmt"
	"strings"
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

// IsMissileWeapon identifies the observed bow, crossbow and sling item group.
// The group is explicit because Range alone also covers thrown weapons.
func (item BaseItem) IsMissileWeapon() bool {
	return item.Type >= 41 && item.Type <= 47
}

// IsThrownWeapon currently records only the RuleBook-confirmed dart
// exception. Other thrown weapon profiles remain data-layer work.
func (item BaseItem) IsThrownWeapon() bool { return item.Type == 9 }

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

// LocalizedItemName composes a display name from typed legacy fields and a
// game-pack locale. Unknown IDs remain visible diagnostics rather than guessed
// translations; unknown name-number components are intentionally omitted.
func LocalizedItemName(item ItemRecord, resolver TextResolver) string {
	base := resolver.Text(fmt.Sprintf("item_type_%02X", item.Type), "")
	if base == "" {
		format := resolver.Text("item_unknown", "item 0x%02X")
		return fmt.Sprintf(format, item.Type)
	}
	parts := make([]string, 0, 3)
	for slot := 3; slot >= 1; slot-- {
		nameNumber := item.NameNumbers[slot-1]
		if nameNumber == 0 || item.HiddenNameFlags&(1<<(3-slot)) != 0 {
			continue
		}
		if translated := resolver.Text(fmt.Sprintf("item_name_%02X", nameNumber), ""); translated != "" {
			// The type's base name is already rendered below; avoid duplicating
			// the common type token when it is also present in NameNumbers.
			if translated != base {
				parts = append(parts, translated)
			}
		}
	}
	if len(parts) > 0 {
		base = strings.Join(append(parts, base), " ")
	}
	if item.Plus > 0 {
		base = fmt.Sprintf(resolver.Text("item_plus", "+%d %s"), item.Plus, base)
	}
	if item.Cursed {
		base += resolver.Text("item_cursed_suffix", " [cursed]")
	}
	if item.Count > 1 && item.Type == 28 {
		base = fmt.Sprintf(resolver.Text("item_count", "%s x%d"), base, item.Count)
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
