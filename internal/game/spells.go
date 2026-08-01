package game

import (
	"fmt"
	"sort"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

const (
	BlessSpellID              uint8 = 1
	CurseSpellID              uint8 = 2
	CauseLightWoundsSpellID   uint8 = 4
	ProtectionFromEvilSpellID uint8 = 6
	ProtectionFromGoodSpellID uint8 = 7
	MagicMissileSpellID       uint8 = 0x0F
	StinkingCloudSpellID      uint8 = 0x22
	FireballSpellID           uint8 = 0x2F
	LightningBoltSpellID      uint8 = 0x33
	CloudkillSpellID          uint8 = 0x5B
)

// firstLevelSpellKeys contains the bounded spell names whose table order is
// verified in the supplied RuleBook. Player records store global spell-table
// IDs: cleric level-one entries begin at 0x01, while the currently translated
// magic-user entries begin at 0x09. IDs outside this catalog remain visible as
// hex so imported DOS/PC-98 slot data is never silently relabeled.
var firstLevelSpellKeys = map[party.Class]struct {
	base uint8
	keys []string
}{
	party.ClassCleric: {
		base: 0x01,
		keys: []string{
			"spell_cleric_1", "spell_cleric_2", "spell_cleric_3", "spell_cleric_4",
			"spell_cleric_5", "spell_cleric_6", "spell_cleric_7", "spell_cleric_8",
		},
	},
	party.ClassMagicUser: {
		base: 0x09,
		keys: []string{
			"spell_magic_user_1", "spell_magic_user_2", "spell_magic_user_3", "spell_magic_user_4",
			"spell_magic_user_5", "spell_magic_user_6", "spell_magic_user_7", "spell_magic_user_8",
		},
	},
}

func campSpellLabel(catalog locale.Catalog, class party.Class, spellID uint8) string {
	table, ok := firstLevelSpellKeys[class]
	if !ok || spellID < table.base || int(spellID-table.base) >= len(table.keys) {
		return fmt.Sprintf(catalog.Text("spell_unknown", "未知法術 0x%02X"), spellID)
	}
	key := table.keys[spellID-table.base]
	return catalog.Text(key, fmt.Sprintf("法術 0x%02X", spellID))
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
	if character.HasClass(party.ClassCleric) || character.HasClass(party.ClassMagicUser) {
		return 1
	}
	return 0
}

// firstLevelMemorizationHours follows the RuleBook's bounded first-level
// timing: four hours of minimum preparation plus fifteen minutes per spell.
// Rest UI is expressed in whole hours, so the required duration is rounded up
// and the pending selection remains intact when rest is too short.
func firstLevelMemorizationHours(pending map[int][]uint8) int {
	maxMinutes := 0
	for _, spells := range pending {
		if len(spells) == 0 {
			continue
		}
		minutes := 4*60 + len(spells)*15
		if minutes > maxMinutes {
			maxMinutes = minutes
		}
	}
	if maxMinutes == 0 {
		return 0
	}
	return (maxMinutes + 59) / 60
}

// pendingCharacterIndexes provides deterministic ordering for diagnostics and
// future per-character partial memorization rules.
func pendingCharacterIndexes(pending map[int][]uint8) []int {
	indexes := make([]int, 0, len(pending))
	for index, spells := range pending {
		if len(spells) > 0 {
			indexes = append(indexes, index)
		}
	}
	sort.Ints(indexes)
	return indexes
}
