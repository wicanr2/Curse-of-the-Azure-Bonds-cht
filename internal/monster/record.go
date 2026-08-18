// Package monster decodes the fixed-size MON*CHA character record used by the
// original engine. It intentionally keeps the raw record separate from ECL
// spawn descriptors and the combat model.
package monster

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/origtext"
)

const RecordSize = 0x1A6

type Record struct {
	// Raw preserves the shared 0x1A6 Player record used by load_npc. Combat
	// parsing reads only a subset; the party adapter needs race/class/icon and
	// other verified player fields without reparsing the DAX container.
	Raw          []byte
	Name         string
	THAC0        int
	MaxHitPoints int
	HitPoints    int
	HitDice      uint8
	RawPlayer74  uint8
	// RaceType is CHARREC.RACETYPE at +11A. MonsterType remains a
	// source-compatible alias for callers that used the old field name.
	RaceType       uint8
	MonsterType    uint8
	Alignment      uint8
	AlignmentKnown bool
	// RawMonsterType is CHARREC.MONSTERTYPE at +14C; its gameplay meaning is
	// intentionally not inferred by this record decoder.
	RawMonsterType  uint8
	Dexterity       uint8
	BaseArmorClass  int
	ArmorClass      int
	// ArmorClassFacing 是角色記錄的第二個護甲欄位 `+19Bh`：派生數值重算
	// （`overlay-24:0C28h`，spec 1000）算的是
	// `+19Bh := +19Ah − 敏捷防禦調整 − 盾牌那一槽 − 2`，攻擊結算
	// （`overlay-13:14E8h`）在背後攻擊成立時改用它。
	//
	// ⚠ 兩個欄位的**絕對值**刻度與 `Fighter.ArmorClass` 不同（見
	// `CombatArmorClass`），所以投影時搬的是**差值**，不是 `+19Bh` 本身。
	ArmorClassFacing int
	AttackBonus     int
	DamageDiceCount int
	DamageDiceSides int
	DamageBonus     int
	// Raw1A5 preserves an unresolved byte without assigning the disproven
	// initiative-bonus name. Initiative is derived from Dexterity +17.
	Raw1A5           uint8
	AttacksPerTurn   int
	CombatTeam       uint8
	CombatSize       uint8
	ModID            uint8
	SpellIDs         []uint8
	MonsterSpellUses [3]uint8
	SavingThrows     []uint8
	SavingThrowBonus int
}

// CombatArmorClass converts the packed MON*CHA armor value used by the
// original records to the signed AD&D combat AC. Real CoAB records store
// values 50..70 as 60-AC (for example FIRE KNIFE 0x3B means AC 1 and the
// post-Dracandros DRACOLICH uses 0x42 for AC -6). Small
// values are retained for synthetic/decoded records that already use combat
// AC, keeping the intermediate parser representation backwards compatible.
func CombatArmorClass(raw int) int {
	if raw >= 50 && raw <= 70 {
		return 60 - raw
	}
	return raw
}

func Parse(data []byte) (Record, error) {
	if len(data) < RecordSize {
		return Record{}, fmt.Errorf("monster record is %d bytes, need %d", len(data), RecordSize)
	}
	nameLength := int(data[0])
	if nameLength > 15 {
		nameLength = 15
	}
	name := origtext.Decode(data[1 : 1+nameLength])
	if nameLength == 0 {
		name = origtext.DecodeField(data[:15])
	}
	maxHP := int(data[0x78])
	currentHP := int(data[0x1A4])
	if currentHP == 0 {
		currentHP = maxHP
	}
	diceCount := int(data[0x19E])
	if diceCount == 0 {
		diceCount = int(data[0x11E])
	}
	diceSides := int(data[0x1A0])
	if diceSides == 0 {
		diceSides = int(data[0x120])
	}
	damageBonus := int(int8(data[0x1A2]))
	if damageBonus == 0 {
		damageBonus = int(int8(data[0x122]))
	}
	spellIDs := make([]uint8, 0, 0x38)
	for _, spellID := range data[0x33:0x6B] {
		if spellID != 0 {
			spellIDs = append(spellIDs, spellID)
		}
	}
	var spellUses [3]uint8
	copy(spellUses[:], data[0xB5:0xB8])
	return Record{
		Raw:            append([]byte(nil), data[:RecordSize]...),
		Name:           name,
		THAC0:          int(int8(data[0x73])),
		MaxHitPoints:   maxHP,
		HitPoints:      currentHP,
		BaseArmorClass:   int(data[0x124]),
		ArmorClass:       int(data[0x19A]),
		ArmorClassFacing: int(data[0x19B]),
		// The reference marks hitBonus/ac as IByte (unsigned byte), unlike
		// the signed damage and THAC0 fields.
		AttackBonus:      int(data[0x199]),
		DamageDiceCount:  diceCount,
		DamageDiceSides:  diceSides,
		DamageBonus:      damageBonus,
		Raw1A5:           data[0x1A5],
		AttacksPerTurn:   int(data[0xA1]),
		CombatTeam:       data[0x198],
		CombatSize:       data[0xDE] & 7,
		ModID:            data[0x126],
		SpellIDs:         spellIDs,
		MonsterSpellUses: spellUses,
		SavingThrows:     append([]uint8(nil), data[0xDF:0xE4]...),
		SavingThrowBonus: int(int8(data[0x186])),
		HitDice:          data[0xE5],
		RawPlayer74:      data[0x74],
		RaceType:         data[0x11A],
		MonsterType:      data[0x11A],
		Alignment:        data[0x11B],
		AlignmentKnown:   true,
		RawMonsterType:   data[0x14C],
		Dexterity:        data[0x17],
	}, nil
}

func (r Record) Fighter(id string, side combat.Side) combat.Fighter {
	return combat.Fighter{
		ID: id, Name: r.Name, SourceName: r.Name, Side: side,
		HitPoints: r.HitPoints, MaxHitPoints: r.MaxHitPoints,
		HitDice:       r.HitDice,
		RawPlayer74:   r.RawPlayer74,
		RaceType:      r.RaceType,
		RaceTypeKnown: true,
		MonsterType:   r.MonsterType,
		Alignment:     r.Alignment, AlignmentKnown: r.AlignmentKnown,
		RawMonsterType: r.RawMonsterType,
		Dexterity:      r.Dexterity,
		CombatTeam:     r.CombatTeam,
		ArmorClass:     CombatArmorClass(r.ArmorClass), AttackBonus: r.AttackBonus,
		// 背後攻擊用的第二個 AC。原作 `+19Ah` 與 `+19Bh` 同域，差值就是
		// 「不算敏捷、不算盾牌，再 −2」那筆減免（spec 1000 §七）。
		// `ResolveAttack` 的判定是 `attackTotal >= AC`——數字小才好打——
		// 所以背後那一格是**減掉**差值，不是直接搬 `+19Bh` 的絕對值。
		ArmorClassFacing:      CombatArmorClass(r.ArmorClass) - (r.ArmorClass - r.ArmorClassFacing),
		ArmorClassFacingKnown: r.ArmorClass != 0 && r.ArmorClassFacing != 0,
		DamageDiceCount: r.DamageDiceCount, DamageDiceSides: r.DamageDiceSides,
		DamageBonus:    r.DamageBonus,
		AttacksPerTurn: r.AttacksPerTurn,
		CombatSize:     r.CombatSize,
		SavingThrows:   append([]uint8(nil), r.SavingThrows...), SavingThrowBonus: r.SavingThrowBonus,
		MonsterSpellIDs: append([]uint8(nil), r.SpellIDs...), MonsterSpellUses: r.MonsterSpellUses,
	}
}
