package dungeon

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestPickLockUsesLivingMarchingOrderAndInclusiveD100(t *testing.T) {
	roster := party.Roster{
		{HitPoints: 0, ThiefSkills: []uint8{0, 99}},
		{HitPoints: 8, ThiefSkills: []uint8{0, 20}},
		{HitPoints: 8, ThiefSkills: []uint8{0, 40}},
	}
	rolls := []uint8{1, 21, 40}
	result := PickLock(roster, func() uint8 {
		value := rolls[0]
		rolls = rolls[1:]
		return value
	})
	if !result.Attempted || !result.Opened || result.CharacterIndex != 2 {
		t.Fatalf("result=%#v", result)
	}
}

func TestPickLockFailureStillConsumesAttempt(t *testing.T) {
	result := PickLock(party.Roster{{HitPoints: 5, ThiefSkills: []uint8{0, 20}}}, func() uint8 { return 21 })
	if !result.Attempted || result.Opened || result.CharacterIndex != -1 {
		t.Fatalf("result=%#v", result)
	}
}

func TestConsumeSpellUsesFirstPartyOccurrence(t *testing.T) {
	roster := party.Roster{{SpellSlots: []uint8{KnockSpellID, 2}}, {SpellSlots: []uint8{KnockSpellID}}}
	updated, ok := ConsumeSpell(roster, KnockSpellID)
	if !ok || len(updated[0].SpellSlots) != 1 || updated[0].SpellSlots[0] != 2 || len(updated[1].SpellSlots) != 1 {
		t.Fatalf("updated=%#v ok=%t", updated, ok)
	}
}
