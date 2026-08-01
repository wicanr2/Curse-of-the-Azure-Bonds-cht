// Package monster decodes the fixed-size MON*CHA character record used by the
// original engine. It intentionally keeps the raw record separate from ECL
// spawn descriptors and the combat model.
package monster

import (
	"fmt"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

const RecordSize = 0x1A6

type Record struct {
	// Raw preserves the shared 0x1A6 Player record used by load_npc. Combat
	// parsing reads only a subset; the party adapter needs race/class/icon and
	// other verified player fields without reparsing the DAX container.
	Raw              []byte
	Name             string
	THAC0            int
	MaxHitPoints     int
	HitPoints        int
	HitDice          uint8
	MonsterType      uint8
	BaseArmorClass   int
	ArmorClass       int
	AttackBonus      int
	DamageDiceCount  int
	DamageDiceSides  int
	DamageBonus      int
	InitiativeBonus  int
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
	name := string(data[1 : 1+nameLength])
	if nameLength == 0 {
		name = strings.TrimRight(string(data[:15]), "\x00 ")
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
		BaseArmorClass: int(data[0x124]),
		ArmorClass:     int(data[0x19A]),
		// The reference marks hitBonus/ac as IByte (unsigned byte), unlike
		// the signed damage and THAC0 fields.
		AttackBonus:      int(data[0x199]),
		DamageDiceCount:  diceCount,
		DamageDiceSides:  diceSides,
		DamageBonus:      damageBonus,
		InitiativeBonus:  int(data[0x1A5]),
		AttacksPerTurn:   int(data[0xA1]),
		CombatTeam:       data[0x197],
		CombatSize:       data[0xDE] & 7,
		ModID:            data[0x126],
		SpellIDs:         spellIDs,
		MonsterSpellUses: spellUses,
		SavingThrows:     append([]uint8(nil), data[0xDF:0xE4]...),
		SavingThrowBonus: int(int8(data[0x186])),
		HitDice:          data[0xE5],
		MonsterType:      data[0x11A],
	}, nil
}

func (r Record) Fighter(id string, side combat.Side) combat.Fighter {
	return combat.Fighter{
		ID: id, Name: r.Name, Side: side,
		HitPoints: r.HitPoints, MaxHitPoints: r.MaxHitPoints,
		HitDice:     r.HitDice,
		MonsterType: r.MonsterType,
		ArmorClass:  CombatArmorClass(r.ArmorClass), AttackBonus: r.AttackBonus,
		DamageDiceCount: r.DamageDiceCount, DamageDiceSides: r.DamageDiceSides,
		DamageBonus: r.DamageBonus, InitiativeBonus: r.InitiativeBonus,
		AttacksPerTurn: r.AttacksPerTurn,
		CombatSize:     r.CombatSize,
		SavingThrows:   append([]uint8(nil), r.SavingThrows...), SavingThrowBonus: r.SavingThrowBonus,
		MonsterSpellIDs: append([]uint8(nil), r.SpellIDs...), MonsterSpellUses: r.MonsterSpellUses,
	}
}
