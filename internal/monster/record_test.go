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
	record, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if record.Name != "ORC!!" || record.MaxHitPoints != 12 || record.HitPoints != 9 || record.ArmorClass != 10 || record.AttackBonus != 3 || record.DamageDiceSides != 8 || record.DamageBonus != 2 {
		t.Fatalf("record=%#v", record)
	}
	fighter := record.Fighter("orc-1", combat.SideEnemy)
	if fighter.ID != "orc-1" || fighter.Side != combat.SideEnemy || fighter.HitPoints != 9 {
		t.Fatalf("fighter=%#v", fighter)
	}
}

func TestParseRejectsShortRecord(t *testing.T) {
	if _, err := Parse(make([]byte, RecordSize-1)); err == nil {
		t.Fatal("expected short record error")
	}
}
