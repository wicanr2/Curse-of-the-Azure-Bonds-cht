package monster

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

func TestParseMonsterRecordOffsets(t *testing.T) {
	data := make([]byte, RecordSize)
	data[0] = 5
	copy(data[1:], "ORC!!")
	data[0x73] = 16
	data[0x78] = 12
	data[0x19E] = 1
	data[0x1A0] = 8
	data[0x1A2] = byte(int8(2))
	data[0x19A] = 10
	data[0x199] = byte(int8(3))
	data[0x1A4] = 9
	data[0x1A5] = 2
	data[0xA1] = 3
	data[0x33] = combat.MonsterMagicMissileSpellID
	data[0xB5] = 1
	record, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if record.Name != "ORC!!" || record.MaxHitPoints != 12 || record.HitPoints != 9 || record.ArmorClass != 10 || record.AttackBonus != 3 || record.DamageDiceSides != 8 || record.DamageBonus != 2 || record.AttacksPerTurn != 3 {
		t.Fatalf("record=%#v", record)
	}
	if len(record.SpellIDs) != 1 || record.SpellIDs[0] != combat.MonsterMagicMissileSpellID || record.MonsterSpellUses[0] != 1 {
		t.Fatalf("monster spell fields=%#v uses=%#v", record.SpellIDs, record.MonsterSpellUses)
	}
	fighter := record.Fighter("orc-1", combat.SideEnemy)
	if fighter.ID != "orc-1" || fighter.Side != combat.SideEnemy || fighter.HitPoints != 9 || fighter.AttacksPerTurn != 3 || len(fighter.MonsterSpellIDs) != 1 || fighter.MonsterSpellUses[0] != 1 {
		t.Fatalf("fighter=%#v", fighter)
	}
}

func TestParseRejectsShortRecord(t *testing.T) {
	if _, err := Parse(make([]byte, RecordSize-1)); err == nil {
		t.Fatal("expected short record error")
	}
}

func TestCombatArmorClassNormalizesPackedMonsterAC(t *testing.T) {
	if got := CombatArmorClass(59); got != 1 {
		t.Fatalf("packed AC 59=%d, want 1", got)
	}
	if got := CombatArmorClass(10); got != 10 {
		t.Fatalf("already decoded AC 10=%d, want 10", got)
	}
	if got := (Record{ArmorClass: 59}).Fighter("fire-knife", combat.SideEnemy).ArmorClass; got != 1 {
		t.Fatalf("fighter AC=%d, want 1", got)
	}
}
