package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

// 資料驅動的傷害／治療類法術（spec 1124）。
//
// ★ 治療輕傷、中傷、重傷在原作是三支各自的 handler，但差別只有骰數
// （`1d8`／`2d8`／`3d8`）。骰子既然抽得出來，remake 這一側就只需要一條路。
//
// ⚠ 只接**單一目標**那一批。範圍傷害（火焰打擊、冰風暴）的形狀不同，
// 而且火球那一支 handler 擲了不只一次骰，表裡標成 `ambiguous`。

// combatCastSpellDice 施放一支骰子來自 spec 1124 表的法術。
func (s *State) combatCastSpellDice(spellID uint8, healing bool) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return fmt.Errorf("spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	required, ok := combatSpellCasterClasses[definition.CasterClass]
	if !ok {
		return fmt.Errorf("spell 0x%02X declares unknown caster class %q", spellID, definition.CasterClass)
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
	dice, ok := combat.SpellDamageRoll(spellID, casterLevel(s.partyRoster[characterIndex]))
	if !ok {
		return fmt.Errorf("spell 0x%02X has no usable dice in the damage table", spellID)
	}

	target, err := s.combatSpellDiceTarget(healing)
	if err != nil {
		return err
	}
	// 豁免只對傷害那一側有意義：`+8h = 2` 是「過了減半」，`= 1` 是「過了沒事」
	// （spec 1111／1117）。`= 0` 連擲都不擲——原作在那一支之前就跳過了。
	halved, negated := false, false
	if !healing {
		entry, ok := gamepack.SpellByID(int(spellID))
		if ok && entry.RequiresSave() {
			targetFighter, _ := s.fighter(target)
			save, saveErr := s.battle.RollSavingThrow(targetFighter, entry.SaveCategory, 0)
			if saveErr != nil {
				return saveErr
			}
			if save.Saved {
				halved = entry.SaveHalvesDamage()
				negated = !halved
			}
		}
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	messageKey := "combat_cause_light_wounds"
	if healing {
		messageKey = "combat_cure_light_wounds"
	} else if spellID != CauseLightWoundsSpellID {
		messageKey = "combat_damage_spell"
	}
	var result combat.SpellResult
	switch {
	case healing:
		result, err = s.battle.CastHealingDice(caster.ID, target, dice)
	case negated:
		// 豁免過了而且這支是「過了沒事」：法術位照樣消耗，傷害不套。
	default:
		result, err = s.battle.CastDamageDice(caster.ID, target, dice,
			combat.SpellDamageElement(spellID), halved)
	}
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	targetFighter, _ := s.fighter(target)
	amount := result.Damage
	if healing {
		amount = result.Healing
	}
	s.CancelCombatCast()
	if messageKey == "combat_damage_spell" {
		s.combatMessage = fmt.Sprintf(s.catalog.Text(messageKey, messageKey),
			caster.Name, targetFighter.Name, s.combatPlayerSpellLabel(spellID), amount)
	} else {
		s.combatMessage = fmt.Sprintf(s.catalog.Text(messageKey, messageKey),
			caster.Name, targetFighter.Name, amount)
	}
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// combatSpellDiceTarget 挑目標：治療找受傷的隊員，傷害找目前選中的敵人。
func (s *State) combatSpellDiceTarget(healing bool) (string, error) {
	if healing {
		targets := s.combatHealingTargets()
		index := s.combatSpellTargetIndex
		if s.combatCastingSpell == 0 {
			index = -1
			for position, target := range targets {
				if target.HitPoints < target.MaxHitPoints {
					index = position
					break
				}
			}
		}
		if index < 0 || index >= len(targets) ||
			targets[index].HitPoints >= targets[index].MaxHitPoints {
			return "", fmt.Errorf("no wounded party member can be healed")
		}
		return targets[index].ID, nil
	}
	targets := s.causeLightWoundsTargets(mustFighter(s))
	if len(targets) == 0 {
		return "", fmt.Errorf("no adjacent enemy can be struck")
	}
	index := s.combatTargetIndex
	if index >= len(targets) {
		index = 0
	}
	return targets[index].ID, nil
}

// mustFighter 取目前輪到的隊員；呼叫端已經確認過是隊伍回合。
func mustFighter(s *State) combat.Fighter {
	fighter, _ := s.combatPartyTurn()
	return fighter
}

// combatCanCastSpellDice 判斷這支法術現在施得出來嗎：職業對、記著、有目標、
// 而且表裡真的有骰子。
func (s *State) combatCanCastSpellDice(spellID uint8, healing bool) bool {
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
	memorized := false
	for _, character := range s.partyRoster {
		if character.ID != caster.ID || !character.HasClass(required) {
			continue
		}
		for _, slot := range character.SpellSlots {
			if slot == spellID {
				memorized = true
				break
			}
		}
	}
	if !memorized {
		return false
	}
	if _, ok := combat.SpellDamageRoll(spellID, 1); !ok {
		return false
	}
	_, err := s.combatSpellDiceTarget(healing)
	return err == nil
}

// CombatCanCastBurningHands exposes the same memorized-slot, class and
// adjacent-target gate used by BeginCombatCast. The frontend and normal-path
// harness may ask this question, but only the cast transaction may consume it.
func (s *State) CombatCanCastBurningHands() bool {
	return s.combatCanCastSpellDice(BurningHandsSpellID, false)
}

// combatCastAreaDamageDice 是範圍版的傷害法術（spec 1124）：骰子與傷害屬性
// 旗標讀表，半徑讀法術主表的宣告半徑。
func (s *State) combatCastAreaDamageDice(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return fmt.Errorf("spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	required, ok := combatSpellCasterClasses[definition.CasterClass]
	if !ok {
		return fmt.Errorf("spell 0x%02X declares unknown caster class %q", spellID, definition.CasterClass)
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
	entry, ok := gamepack.SpellByID(int(spellID))
	if !ok || entry.AreaRadius <= 0 {
		return fmt.Errorf("spell 0x%02X has no declared area radius", spellID)
	}
	dice, ok := combat.SpellDamageRoll(spellID, casterLevel(s.partyRoster[characterIndex]))
	if !ok {
		return fmt.Errorf("spell 0x%02X has no usable dice in the damage table", spellID)
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastAreaDamageDice(caster.ID, s.combatSpellTargetPoint,
		dice, entry.AreaRadius, combat.SpellDamageElement(spellID),
		entry.RequiresSave(), entry.SaveCategory, entry.SaveHalvesDamage())
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	s.CancelCombatCast()
	// ⚠ 參數順序是「目標數、傷害」——`combat_fireball` 那一支也是這個順序，
	// 反過來會印出「波及 12 名目標，共造成 3 點傷害」這種對調的句子。
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_area_damage_spell", "combat_area_damage_spell"),
		caster.Name, len(result.Impacts), result.BaseDamage)
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// combatCanCastAreaDamageDice 判斷範圍傷害法術現在施得出來嗎。
func (s *State) combatCanCastAreaDamageDice(spellID uint8) bool {
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
	entry, ok := gamepack.SpellByID(int(spellID))
	if !ok || entry.AreaRadius <= 0 {
		return false
	}
	if _, ok := combat.SpellDamageRoll(spellID, 1); !ok {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok || len(s.livingBySide(combat.SideEnemy)) == 0 {
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

// combatCastAreaMoraleBreak 是混亂術（法術 82）：效果碼 `23h` 的處理常式就是
// 士氣崩潰的四段表（spec 831／1122），所以這支法術不需要新的規則。
func (s *State) combatCastAreaMoraleBreak(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	definition, found := s.combatPlayerSpellDefinition(spellID)
	if !found {
		return fmt.Errorf("spell 0x%02X is not declared in combat_player_spells", spellID)
	}
	required, ok := combatSpellCasterClasses[definition.CasterClass]
	if !ok {
		return fmt.Errorf("spell 0x%02X declares unknown caster class %q", spellID, definition.CasterClass)
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
	entry, ok := gamepack.SpellByID(int(spellID))
	if !ok || entry.AreaRadius <= 0 {
		return fmt.Errorf("spell 0x%02X has no declared area radius", spellID)
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastAreaMoraleBreak(caster.ID, s.combatSpellTargetPoint,
		entry.AreaRadius, entry.SaveCategory)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	broken := 0
	for _, impact := range result.Impacts {
		if !impact.Saved {
			broken++
		}
	}
	s.CancelCombatCast()
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_morale_confused", "combat_morale_confused"),
		fmt.Sprintf("%s（%d）", caster.Name, broken))
	s.requestSound(SoundCast)
	s.requestSound(SoundSpellHit)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// combatCanCastAreaMoraleBreak 判斷混亂術現在施得出來嗎。
func (s *State) combatCanCastAreaMoraleBreak(spellID uint8) bool {
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
	entry, ok := gamepack.SpellByID(int(spellID))
	if !ok || entry.AreaRadius <= 0 {
		return false
	}
	caster, ok := s.combatPartyTurn()
	if !ok || len(s.livingBySide(combat.SideEnemy)) == 0 {
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
