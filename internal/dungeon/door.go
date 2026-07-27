// Package dungeon contains platform-neutral dungeon interaction rules.
package dungeon

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"

// KnockSpellID is the original Spells.knock value from the DOS rules.
const KnockSpellID uint8 = 0x1F

// PickLockResult records the deterministic outcome of one original pick-lock
// attempt. Attempted is true whenever the door menu consumed the attempt,
// including a failed roll.
type PickLockResult struct {
	Attempted      bool
	Opened         bool
	CharacterIndex int
}

// PickLock performs the verified part of ovr015.pick_lock. The roll function
// returns one d100 result per eligible living character, in marching order.
// A nil roll function is treated as an unavailable attempt. Character HP is
// the current data-neutral equivalent of the original Status.okey check.
func PickLock(roster party.Roster, roll func() uint8) PickLockResult {
	result := PickLockResult{Attempted: true, CharacterIndex: -1}
	if roll == nil {
		return result
	}
	for index, character := range roster {
		// Keep the reference evaluation order: roll_dice is called before
		// the short-circuit health check in pick_lock().
		rollResult := roll()
		if rollResult <= character.OpenLocksSkill() && character.HitPoints > 0 {
			result.Opened = true
			result.CharacterIndex = index
			return result
		}
	}
	return result
}

// ConsumeSpell removes the first memorized occurrence of spellID, matching
// the original TeamMemberHasSpell/RemoveKnockSpell first-match behavior.
func ConsumeSpell(roster party.Roster, spellID uint8) (party.Roster, bool) {
	for characterIndex := range roster {
		for slotIndex, knownSpell := range roster[characterIndex].SpellSlots {
			if knownSpell != spellID {
				continue
			}
			roster[characterIndex].SpellSlots = append(
				roster[characterIndex].SpellSlots[:slotIndex:slotIndex],
				roster[characterIndex].SpellSlots[slotIndex+1:]...,
			)
			return roster, true
		}
	}
	return roster, false
}
