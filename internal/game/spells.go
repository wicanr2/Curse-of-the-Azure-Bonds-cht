package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

const (
	BlessSpellID              uint8 = 1
	CurseSpellID              uint8 = 2
	CauseLightWoundsSpellID   uint8 = 4
	ProtectionFromEvilSpellID uint8 = 6
	ProtectionFromGoodSpellID uint8 = 7
	BurningHandsSpellID       uint8 = 9
	MagicMissileSpellID       uint8 = 0x0F
	SleepSpellID              uint8 = 0x15
	StinkingCloudSpellID      uint8 = 0x22
	FireballSpellID           uint8 = 0x2F
	LightningBoltSpellID      uint8 = 0x33
	CloudkillSpellID          uint8 = 0x5B
)

// spellMessageID 是法術名的 locale 鍵：`spell_<施法職業>_<原作編號>`。
//
// ★ 職業與編號都來自**原作法術表**（`gamepack.SpellByID`），不是呼叫端傳進來的
// 角色職業——同一個編號只屬於一個施法職業，而雙職角色的法術槽兩邊都有。
// 靠角色職業去選鍵，多職角色就會拿到另一個職業的名字。
//
// ⚠ 編號一律是**全域編號**。這個前綴曾經同時存在兩套編號（法師的 1..13 是
// 「法師第幾支」而 29 以上是全域編號），兩套在 1..13 這一段會互相蓋掉。
// `TestSpellNameKeysFollowTheOriginalTable` 擋住這種漂移。
func spellMessageID(spellID uint8) (string, bool) {
	spell, ok := gamepack.SpellByID(int(spellID))
	if !ok || spell.Placeholder || spell.CasterClass == "" {
		return "", false
	}
	// 新增的 locale ID 必須遵守 spec 1105 的點分命名；既有的
	// spell_<class>_<id> 只能維持相容，不再擴張。
	if key, found := map[uint8]string{
		22: "spell.cleric.find-traps",
		39: "spell.cleric.cure-disease",
	}[spellID]; found {
		return key, true
	}
	return fmt.Sprintf("spell_%s_%d", strings.ReplaceAll(spell.CasterClass, "-", "_"), spell.SpellID), true
}

// campSpellLabel 把法術槽裡的原作編號翻成玩家看得到的名字。
// 沒有譯名時**保留十六進位編號**，不要退回一個看起來像法術名的字串——
// 匯入的 DOS／PC-98 存檔裡可能有 remake 還沒認識的編號，蓋掉就查不出來了。
func campSpellLabel(catalog locale.Catalog, spellID uint8) string {
	unknown := fmt.Sprintf(catalog.Text("spell_unknown", "spell_unknown 0x%02X"), spellID)
	key, ok := spellMessageID(spellID)
	if !ok {
		return unknown
	}
	if translated := catalog.Text(key, ""); translated != "" {
		return translated
	}
	return unknown
}

// firstLevelMemorizedCapacity is kept as the narrow level-one query used by
// existing callers and tests. The CAMP memorization path itself uses the full
// original 3x5 capacity table through memorizedCapacity.
func firstLevelMemorizedCapacity(character party.Character) int {
	capacity := 0
	if character.HasClass(party.ClassCleric) {
		capacity = int(character.SpellCastCount[0][0])
	}
	if character.HasClass(party.ClassMagicUser) && int(character.SpellCastCount[2][0]) > capacity {
		capacity = int(character.SpellCastCount[2][0])
	}
	// 舊 JSON 沒有 SpellCastCount；保留先前的一級 fallback，但已記憶格數
	// 若更大就不能因施放／載入投影而縮小容量。
	if capacity == 0 && character.Level >= 1 &&
		(character.HasClass(party.ClassCleric) || character.HasClass(party.ClassMagicUser)) {
		capacity = 1
	}
	if len(character.SpellSlots) > capacity {
		capacity = len(character.SpellSlots)
	}
	return capacity
}

func memorizedCapacity(character party.Character) int {
	capacity := 0
	for class := range character.SpellCastCount {
		for level := range character.SpellCastCount[class] {
			capacity += int(character.SpellCastCount[class][level])
		}
	}
	if capacity == 0 {
		return firstLevelMemorizedCapacity(character)
	}
	if len(character.SpellSlots) > capacity {
		return len(character.SpellSlots)
	}
	return capacity
}

// memorizationSlot maps a spell to the original class/level capacity cell.
// Spell IDs are global table IDs; deriving the cell from the table prevents a
// mixed-class character from spending (for example) a cleric level-two slot on
// a magic-user level-two spell.
func memorizationSlot(character party.Character, spellID uint8) (class, level, capacity int, ok bool) {
	spell, found := gamepack.SpellByID(int(spellID))
	if !found || spell.Placeholder || spell.CasterClassID < 0 ||
		spell.CasterClassID >= len(character.SpellCastCount) || spell.Level < 1 || spell.Level > 5 {
		return 0, 0, 0, false
	}
	class, level = spell.CasterClassID, spell.Level-1
	capacity = int(character.SpellCastCount[class][level])
	if capacity == 0 && spell.Level == 1 {
		hasRecordedCapacity := false
		for capacityClass := range character.SpellCastCount {
			for capacityLevel := range character.SpellCastCount[capacityClass] {
				hasRecordedCapacity = hasRecordedCapacity || character.SpellCastCount[capacityClass][capacityLevel] > 0
			}
		}
		if !hasRecordedCapacity {
			capacity = firstLevelMemorizedCapacity(character)
		}
	}
	return class, level, capacity, capacity > 0
}

// memorizationCandidates follows the original BuildSpellList grouping rules:
// clerics may prepare their complete class list, while magic-users and druids
// are restricted to known-spell flags. Only levels with a non-zero capacity
// are offered. This preserves the distinction between a cleric prayer list and
// a magic-user grimoire for multiclass characters as well.
func memorizationCandidates(character party.Character) []uint8 {
	known := make(map[uint8]bool, len(character.KnownSpells))
	for _, spellID := range character.KnownSpells {
		known[spellID] = true
	}
	result := make([]uint8, 0, len(character.KnownSpells)+16)
	for spellID := 1; ; spellID++ {
		spell, found := gamepack.SpellByID(spellID)
		if !found {
			break
		}
		_, _, _, available := memorizationSlot(character, uint8(spellID))
		if !available {
			continue
		}
		if (spell.CasterClassID == 0 && character.HasClass(party.ClassCleric)) || known[uint8(spellID)] {
			result = append(result, uint8(spellID))
		}
	}
	return result
}

// memorizationHours follows the RuleBook preparation contract: levels one and
// two have a four-hour minimum, levels three and four six hours, and level five
// eight hours, plus fifteen minutes per spell level. Characters prepare in
// parallel, so the party requirement is the slowest character rather than the
// sum of every character's work.
func memorizationHours(pending map[int][]uint8) int {
	maxMinutes := 0
	for _, spells := range pending {
		if len(spells) == 0 {
			continue
		}
		highestLevel := 1
		spellMinutes := 0
		for _, spellID := range spells {
			level := 1
			if spell, ok := gamepack.SpellByID(int(spellID)); ok && spell.Level >= 1 && spell.Level <= 5 {
				level = spell.Level
			}
			if level > highestLevel {
				highestLevel = level
			}
			spellMinutes += level * 15
		}
		minimumHours := 4
		if highestLevel >= 5 {
			minimumHours = 8
		} else if highestLevel >= 3 {
			minimumHours = 6
		}
		minutes := minimumHours*60 + spellMinutes
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
