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
