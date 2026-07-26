package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// firstLevelSpellKeys contains only spell names whose class table order is
// verified in the supplied RuleBook. Spell IDs outside this bounded catalog
// remain visible as hex so the DOS slot data is never silently relabeled.
var firstLevelSpellKeys = map[party.Class][]string{
	party.ClassCleric: {
		"spell_cleric_1", "spell_cleric_2", "spell_cleric_3", "spell_cleric_4",
		"spell_cleric_5", "spell_cleric_6", "spell_cleric_7", "spell_cleric_8",
	},
	party.ClassMagicUser: {
		"spell_magic_user_1", "spell_magic_user_2", "spell_magic_user_3", "spell_magic_user_4",
		"spell_magic_user_5", "spell_magic_user_6", "spell_magic_user_7", "spell_magic_user_8",
	},
}

func campSpellLabel(catalog locale.Catalog, class party.Class, spellID uint8) string {
	keys := firstLevelSpellKeys[class]
	if spellID == 0 || int(spellID) > len(keys) {
		return fmt.Sprintf(catalog.Text("spell_unknown", "未知法術 0x%02X"), spellID)
	}
	return catalog.Text(keys[spellID-1], fmt.Sprintf("法術 0x%02X", spellID))
}

// firstLevelMemorizedCapacity is the bounded preparation adapter used by the
// current first-level spell catalog. Imported characters retain the observed
// number of memorized slots; newly created spellcasters use the documented
// first-level capacity. Higher-level and multi-level slot tables remain a
// rules-data task.
func firstLevelMemorizedCapacity(character party.Character) int {
	if len(character.SpellSlots) > 0 {
		return len(character.SpellSlots)
	}
	if character.Level < 1 {
		return 0
	}
	switch character.Class {
	case party.ClassCleric, party.ClassMagicUser:
		return 1
	default:
		return 0
	}
}
