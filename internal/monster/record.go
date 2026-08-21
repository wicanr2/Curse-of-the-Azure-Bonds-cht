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
	// RawItemCount 是角色記錄 `+14Ch`。
	//
	// ⚠ 這一格先前記成「`CHARREC.MONSTERTYPE`」——那是 **PC-98** 的版面。
	// DOS 的記錄少一格（422 vs 423），`MONSTERTYPE` 正是少掉的那一格，所以
	// DOS 的 `+14Ch` 是 `NUMITEMS`、`+14Dh`..`+150h` 才是物品鏈指標（spec 1166）。
	// 遊戲資料裡它一律是 0，這個欄位只是原樣保留，沒有任何規則讀它。
	RawItemCount uint8
	// UndeadType 是 `+0E9h`（`UNDEADLEVEL`）：1..10 是驅散不死矩陣的列，
	// 0 代表不是不死生物（spec 834／1164）。
	UndeadType     uint8
	Dexterity      uint8
	BaseArmorClass int
	ArmorClass     int
	// ArmorClassFacing 是角色記錄的第二個護甲欄位 `+19Bh`：派生數值重算
	// （`overlay-24:0C28h`，spec 1000）算的是
	// `+19Bh := +19Ah − 敏捷防禦調整 − 盾牌那一槽 − 2`，攻擊結算
	// （`overlay-13:14E8h`）在背後攻擊成立時改用它。
	//
	// ⚠ 這裡存的是**原始儲存值**；投影到 `Fighter` 時與第一格走同一條換算
	// （見 `CombatArmorClassFacing`）。
	ArmorClassFacing int
	AttackBonus      int
	DamageDiceCount  int
	DamageDiceSides  int
	DamageBonus      int
	// Raw1A5 preserves an unresolved byte without assigning the disproven
	// initiative-bonus name. Initiative is derived from Dexterity +17.
	Raw1A5         uint8
	AttacksPerTurn int
	CombatTeam     uint8
	CombatSize     uint8
	// RawSize 是 `+0DEh` 的完整位元組（`SIZE`）。**不要只留 `and 7`**：
	// 低 3 位是佔格大小，而 **bit 7 是「傷害算大型目標」**——`81h` 的佔格是 1
	// 但仍算大型（BEHOLDER、BUGBEAR、兩種巨蛛都是這個值，spec 1175）。
	RawSize          uint8
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

// CombatArmorClassFacing 把第二個護甲欄位換到與 `CombatArmorClass` 相同的刻度。
// 換不換由**第一格**決定：第二格比第一格小 2..8，自己判會掉出視窗。
func (r Record) CombatArmorClassFacing() int {
	if CombatArmorClass(r.ArmorClass) == r.ArmorClass {
		return r.ArmorClassFacing
	}
	return 60 - r.ArmorClassFacing
}

// CombatAttackBonus 是命中能力那一側的同一件事：原作 `+199h` 存的是
// `60 − THAC0`，remake 的 `AttackBonus` 用的是 `20 − THAC0`，兩者差 40。
// 真實記錄落在 40..53，合成／已解碼的小數值原樣保留。
func CombatAttackBonus(raw int) int {
	if raw >= 30 && raw <= 70 {
		return raw - 40
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
		Raw:              append([]byte(nil), data[:RecordSize]...),
		Name:             name,
		THAC0:            int(int8(data[0x73])),
		MaxHitPoints:     maxHP,
		HitPoints:        currentHP,
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
		RawSize:          data[0xDE],
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
		RawItemCount:     data[0x14C],
		Dexterity:        data[0x17],
		// `+0E9h` ＝ `UNDEADLEVEL`：1..10 是驅散矩陣的列，0 代表不是不死生物
		// （spec 834／1164）。
		UndeadType: data[0x0E9],
	}, nil
}

// LargeDamageTarget 重現原作攻擊結算的大／小體型判斷
// （`overlay-13:15EFh`，spec 1175）：
//
//	if (目標^[0DEh] > 80h) or ((目標^[0DEh] and 7) > 1) then 用大型傷害三連
//
// ⚠ 只看 `and 7` 會漏掉 `81h`——那一格的佔格是 1，但 bit 7 說它算大型。
func LargeDamageTarget(rawSize uint8) bool {
	return rawSize > 0x80 || rawSize&7 > 1
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
		RawItemCount: r.RawItemCount,
		UndeadType:   r.UndeadType,
		Dexterity:    r.Dexterity,
		CombatTeam:   r.CombatTeam,
		ArmorClass:   CombatArmorClass(r.ArmorClass), AttackBonus: CombatAttackBonus(r.AttackBonus),
		// 背後攻擊用的第二個 AC（`+19Bh`，spec 1000 §七）。儲存值比 `+19Ah` 小，
		// 換成畫面刻度就比正面那一格大——也就是比較好打。
		//
		// ⚠ 兩格要走**同一個**判斷：第二格系統性地小 2..8，可以掉出
		// `CombatArmorClass` 的視窗，各判各的會讓一格換算、一格沒換。
		ArmorClassFacing:      r.CombatArmorClassFacing(),
		ArmorClassFacingKnown: r.ArmorClass != 0 && r.ArmorClassFacing != 0,
		DamageDiceCount:       r.DamageDiceCount, DamageDiceSides: r.DamageDiceSides,
		DamageBonus:    r.DamageBonus,
		AttacksPerTurn: r.AttacksPerTurn,
		CombatSize:     r.CombatSize,
		LargeTarget:    LargeDamageTarget(r.RawSize),
		SavingThrows:   append([]uint8(nil), r.SavingThrows...), SavingThrowBonus: r.SavingThrowBonus,
		MonsterSpellIDs: append([]uint8(nil), r.SpellIDs...), MonsterSpellUses: r.MonsterSpellUses,
	}
}
