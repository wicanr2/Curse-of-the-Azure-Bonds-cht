package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 十呎半徑版的防護邪惡／善良（法術 52／53／69）。
//
// ★ 與單目標版（6／7／16／17）的差別只有**套給誰**：原作的效果碼不同
// （`2Dh`／`2Eh` 對 `08h`／`09h`），但兩邊的規則是同一條，所以 remake 這一側
// 沿用同一支 `CastProtectionFrom*`，只是對全隊各跑一次。
//
// ⚠ 「十呎半徑」在 remake 等同「整支隊伍」——與隱形術十呎半徑（法術 50）
// 早就採用的口徑一致。原作是不是真的用距離篩選沒有讀過。

// combatCanCastProtectionRadius 判斷這支範圍防護現在施得出來嗎。
func (s *State) combatCanCastProtectionRadius(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return false
	}
	required, ok := combatSpellCasterClasses[definition.CasterClass]
	if !ok {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(required) {
			continue
		}
		for _, slot := range character.SpellSlots {
			if slot == spellID {
				return true
			}
		}
	}
	return false
}

// combatCastProtectionRadius 對每個活著的隊員套上防護。
func (s *State) combatCastProtectionRadius(spellID uint8, required party.Class, evil bool) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(required) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("caster %q cannot cast spell 0x%02X", caster.ID, spellID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == spellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no memorized spell 0x%02X", caster.ID, spellID)
	}
	level := casterLevel(s.partyRoster[characterIndex])
	targets := s.livingBySide(combat.SideParty)
	if len(targets) == 0 {
		return fmt.Errorf("no living party member can be protected")
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	applied := 0
	for _, target := range targets {
		var err error
		if evil {
			_, err = s.battle.CastProtectionFromEvil(caster.ID, target.ID, level)
		} else {
			_, err = s.battle.CastProtectionFromGood(caster.ID, target.ID, level)
		}
		if err != nil {
			// 已經有防護的人會被拒絕——那不是失敗，是這一位不需要。
			continue
		}
		applied++
	}
	if applied == 0 {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return fmt.Errorf("every party member already has that protection")
	}
	messageKey := "combat_protection_from_good"
	if evil {
		messageKey = "combat_protection_from_evil"
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(s.catalog.Text(messageKey, messageKey),
		caster.Name, caster.Name, 3*level)
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}
