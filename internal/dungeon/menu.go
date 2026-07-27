package dungeon

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"

// DoorMenuOptions is the capability-filtered action set shown by the
// reference locked_door menu.
type DoorMenuOptions struct {
	Bash  bool
	Pick  bool
	Knock bool
}

// DoorMenuOptionsFor returns the reference menu capabilities for a locked
// door. Detail 2 is pickable; detail 3 is explicitly unpickable. The current
// party model treats a non-zero imported open-lock value as verified thief
// capability, and spell lookup is the first memorized Knock slot.
func DoorMenuOptionsFor(roster party.Roster, detail uint8) DoorMenuOptions {
	if detail != 2 && detail != 3 {
		return DoorMenuOptions{}
	}
	options := DoorMenuOptions{Bash: true}
	if detail == 2 {
		for _, character := range roster {
			if character.OpenLocksSkill() > 0 {
				options.Pick = true
				break
			}
		}
	}
	_, options.Knock = roster.FindSpell(uint16(KnockSpellID))
	return options
}
