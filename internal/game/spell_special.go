package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

// 形狀各自不同的那幾支法術（spec 1124）。共通形狀的都走資料驅動那幾條路。

// combatSpellPrologue 收攏「這支法術現在施得出來嗎」共用的前置：職業對、記著。
// 回傳角色索引與法術位索引。
func (s *State) combatSpellPrologue(spellID uint8) (int, int, combat.Fighter, error) {
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return -1, -1, combat.Fighter{},
			fmt.Errorf("spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	required, ok := combatSpellCasterClasses[definition.CasterClass]
	if !ok {
		return -1, -1, combat.Fighter{},
			fmt.Errorf("spell 0x%02X declares unknown caster class %q", spellID, definition.CasterClass)
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return -1, -1, combat.Fighter{}, fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for index, character := range s.partyRoster {
		if character.ID == caster.ID && character.HasClass(required) {
			characterIndex = index
			break
		}
	}
	if characterIndex < 0 {
		return -1, -1, combat.Fighter{},
			fmt.Errorf("caster %q cannot cast spell 0x%02X", caster.ID, spellID)
	}
	spellIndex := -1
	for index, memorized := range s.partyRoster[characterIndex].SpellSlots {
		if memorized == spellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return -1, -1, combat.Fighter{},
			fmt.Errorf("caster %q has no memorized spell 0x%02X", caster.ID, spellID)
	}
	return characterIndex, spellIndex, caster, nil
}

// combatFinishSpell 收攏施法之後的共用收尾。
func (s *State) combatFinishSpell(message string) error {
	s.CancelCombatCast()
	s.combatMessage = message
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// combatCastSlayLiving 是屠殺活物（法術 76）。
func (s *State) combatCastSlayLiving(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	enemies := s.livingBySide(combat.SideEnemy)
	if len(enemies) == 0 {
		return fmt.Errorf("combat has no living enemies")
	}
	index := s.combatTargetIndex
	if index >= len(enemies) {
		index = 0
	}
	target := enemies[index]
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastSlayLiving(caster.ID, target.ID)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	key := "combat_cause_light_wounds"
	amount := result.Damage
	if result.Slain {
		key = "combat_morale_runs_away"
		amount = 0
	}
	message := fmt.Sprintf(s.catalog.Text(key, key), caster.Name, target.Name, amount)
	if result.Slain {
		message = fmt.Sprintf(s.catalog.Text(key, key), target.Name)
	}
	return s.combatFinishSpell(message)
}

// combatCastCureBlindness 是治療失明（法術 37）：把致盲效果從隊員身上拿掉。
func (s *State) combatCastCureBlindness(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	targets := s.livingBySide(combat.SideParty)
	blinded := ""
	for _, target := range targets {
		for _, affect := range target.MonsterAffects {
			if affect.Kind == 0x21 {
				blinded = target.ID
				break
			}
		}
		if blinded != "" {
			break
		}
	}
	if blinded == "" {
		return fmt.Errorf("no blinded party member to cure")
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	if _, err := s.battle.CureBlindness(blinded); err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	target, _ := s.fighter(blinded)
	return s.combatFinishSpell(fmt.Sprintf(
		s.catalog.Text("combat_cure_light_wounds", "combat_cure_light_wounds"),
		caster.Name, target.Name, 0))
}

// combatCanCastSlayLiving／combatCanCastCureBlindness 是兩支的可施放判斷。
func (s *State) combatCanCastSlayLiving(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	if _, _, _, err := s.combatSpellPrologue(spellID); err != nil {
		return false
	}
	return len(s.livingBySide(combat.SideEnemy)) > 0
}

func (s *State) combatCanCastCureBlindness(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	if _, _, _, err := s.combatSpellPrologue(spellID); err != nil {
		return false
	}
	for _, target := range s.livingBySide(combat.SideParty) {
		for _, affect := range target.MonsterAffects {
			if affect.Kind == 0x21 {
				return true
			}
		}
	}
	return false
}

// combatCastApplyEffectCodes 是「效果碼在 handler 裡」的那幾支（毒、疾病）。
func (s *State) combatCastApplyEffectCodes(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	codes, ok := combat.SpellExtraEffectCodes(spellID)
	if !ok || len(codes) == 0 {
		return fmt.Errorf("spell 0x%02X has no handler-supplied effect codes", spellID)
	}
	enemies := s.livingBySide(combat.SideEnemy)
	if len(enemies) == 0 {
		return fmt.Errorf("combat has no living enemies")
	}
	index := s.combatTargetIndex
	if index >= len(enemies) {
		index = 0
	}
	target := enemies[index]
	entry, _ := gamepack.SpellByID(int(spellID))
	duration := entry.PrimaryDuration(casterLevel(s.partyRoster[characterIndex]), false)
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	if _, err := s.battle.ApplyEffectCodes(target.ID, codes, duration); err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	return s.combatFinishSpell(fmt.Sprintf(
		s.catalog.Text("combat_spell_applied", "combat_spell_applied"),
		caster.Name, target.Name, s.combatPlayerSpellLabel(spellID)))
}

// combatCanCastApplyEffectCodes 是上面那一支的可施放判斷。
func (s *State) combatCanCastApplyEffectCodes(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	if _, ok := combat.SpellExtraEffectCodes(spellID); !ok {
		return false
	}
	if _, _, _, err := s.combatSpellPrologue(spellID); err != nil {
		return false
	}
	return len(s.livingBySide(combat.SideEnemy)) > 0
}

// combatCastMessageOnly 是「原作那一支就只印訊息」的法術。
//
// ★ 法師版的降咒術（法術 100）屬性表 `+0Ah` 是 0，而 handler 呼叫施法收尾時
// 傷害與效果全部傳 0——所以它在戰鬥中確實只消耗法術位並印一句話。
// 這不是 remake 少做，是原作就這樣。
func (s *State) combatCastMessageOnly(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	return s.combatFinishSpell(fmt.Sprintf(
		s.catalog.Text("combat_spell_no_effect", "combat_spell_no_effect"),
		caster.Name))
}

// combatCanCastMessageOnly 同上。
func (s *State) combatCanCastMessageOnly(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	_, _, _, err := s.combatSpellPrologue(spellID)
	return err == nil
}

// combatCanCastEffectSpell 是資料驅動效果類法術的可施放判斷：職業對、記著、
// 表裡真的有效果碼、而且有合法目標。
func (s *State) combatCanCastEffectSpell(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return false
	}
	spell, ok := gamepack.SpellByID(int(spellID))
	if !ok || spell.EffectID == 0 {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok {
		return false
	}
	if characterIndex, _ := s.memorizedSpellSlot(caster.ID, spellID); characterIndex < 0 {
		return false
	}
	if definition.TargetMode == "area_point" {
		// 落點還沒選，所以這裡不能拿「那一格有沒有人」當條件——只要場上
		// 還有活著的敵人就施得出來，目標由選點那一步決定。
		return len(s.livingBySide(combat.SideEnemy)) > 0
	}
	return len(s.effectSpellTargets(caster, spellID, definition.TargetMode)) > 0
}
