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
	Name            string
	THAC0           int
	MaxHitPoints    int
	HitPoints       int
	BaseArmorClass  int
	ArmorClass      int
	AttackBonus     int
	DamageDiceCount int
	DamageDiceSides int
	DamageBonus     int
	InitiativeBonus int
	ModID           uint8
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
	return Record{
		Name:           name,
		THAC0:          int(int8(data[0x73])),
		MaxHitPoints:   maxHP,
		HitPoints:      currentHP,
		BaseArmorClass: int(data[0x124]),
		ArmorClass:     int(data[0x19A]),
		// The reference marks hitBonus/ac as IByte (unsigned byte), unlike
		// the signed damage and THAC0 fields.
		AttackBonus:     int(data[0x199]),
		DamageDiceCount: diceCount,
		DamageDiceSides: diceSides,
		DamageBonus:     damageBonus,
		InitiativeBonus: int(data[0x1A5]),
		ModID:           data[0x126],
	}, nil
}

func (r Record) Fighter(id string, side combat.Side) combat.Fighter {
	return combat.Fighter{
		ID: id, Name: r.Name, Side: side,
		HitPoints: r.HitPoints, MaxHitPoints: r.MaxHitPoints,
		ArmorClass: r.ArmorClass, AttackBonus: r.AttackBonus,
		DamageDiceCount: r.DamageDiceCount, DamageDiceSides: r.DamageDiceSides,
		DamageBonus: r.DamageBonus, InitiativeBonus: r.InitiativeBonus,
	}
}
