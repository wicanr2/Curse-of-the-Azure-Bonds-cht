package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 解除魔法、移除詛咒、次元門、火焰護盾、寒冰錐（spec 1125）。
//
// 這幾支的共通點是「收尾不是標準那兩支」：傷害骰表只記得到它們擲了什麼，
// 擲出來的東西**不是傷害**。所以每一支都要讀完 handler 才接得上。

// combatCastDispelMagic 是解除魔法（法術 41 牧師版／46 法師版，同一支 handler）。
func (s *State) combatCastDispelMagic(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	entry, ok := gamepack.SpellByID(int(spellID))
	if !ok {
		return fmt.Errorf("spell 0x%02X is not in the spell table", spellID)
	}
	level := casterLevel(s.partyRoster[characterIndex])
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastDispelMagic(caster.ID, s.combatSpellTargetPoint,
		entry.AreaRadius, level)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	return s.combatFinishSpell(fmt.Sprintf(
		s.catalog.Text("combat_dispel_magic", "combat_dispel_magic"),
		caster.Name, result.Removed()))
}

// combatCanCastDispelMagic 判斷解除魔法現在施得出來嗎。
func (s *State) combatCanCastDispelMagic(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	if _, _, _, err := s.combatSpellPrologue(spellID); err != nil {
		return false
	}
	entry, ok := gamepack.SpellByID(int(spellID))
	return ok && entry.AreaRadius >= 0
}

// combatCastRemoveCurse 是移除詛咒（法術 43 牧師版／89 法師版，同一支 handler）。
//
// ★ 原作**先**試著拿掉效果 `24h`，拿掉了就結束；沒拿掉才去翻裝備，
// 而且只處理**第一件**詛咒裝備。兩步順序反過來會讓身上有 `24h` 的人
// 連裝備也一起被解開。
//
// ⚠ 原作把那件裝備的「已裝備」旗標清成 0，**沒有**清詛咒旗標本身——
// 所以它只是讓人脫得下來，不是把詛咒去掉。
func (s *State) combatCastRemoveCurse(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	targets := s.combatHealingTargets()
	if len(targets) == 0 {
		targets = s.livingBySide(combat.SideParty)
	}
	if len(targets) == 0 {
		return fmt.Errorf("no party member can be targeted")
	}
	index := s.combatSpellTargetIndex
	if index < 0 || index >= len(targets) {
		index = 0
	}
	target := targets[index]
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	removed, err := s.battle.RemoveCurseAffect(target.ID)
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	freed := ""
	if !removed {
		freed = s.freeFirstCursedItem(target.ID)
	}
	message := s.catalog.Text("combat_remove_curse_none", "combat_remove_curse_none")
	if removed || freed != "" {
		message = s.catalog.Text("combat_remove_curse", "combat_remove_curse")
	}
	return s.combatFinishSpell(fmt.Sprintf(message, caster.Name, target.Name))
}

// freeFirstCursedItem 把目標身上**第一件**詛咒裝備解除裝備狀態，回傳它的名字。
func (s *State) freeFirstCursedItem(fighterID string) string {
	for index := range s.partyRoster {
		if s.partyRoster[index].ID != fighterID {
			continue
		}
		character := &s.partyRoster[index]
		for slot := range character.Equipment {
			item := character.Equipment[slot]
			if !item.Readied || !item.Cursed {
				continue
			}
			character.Equipment[slot].Readied = false
			// 脫下來會動到護甲等級與傷害骰，所以要重投影一次。
			updated, err := s.partyRoster[index].FighterWithEquipment(s.itemCatalog)
			if err != nil {
				character.Equipment[slot].Readied = true
				return ""
			}
			if err := s.battle.ReplaceFighterEquipment(fighterID, updated); err != nil {
				character.Equipment[slot].Readied = true
				return ""
			}
			return item.Name
		}
		return ""
	}
	return ""
}

// combatCanCastRemoveCurse 判斷移除詛咒現在施得出來嗎。
func (s *State) combatCanCastRemoveCurse(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	if _, _, _, err := s.combatSpellPrologue(spellID); err != nil {
		return false
	}
	return len(s.livingBySide(combat.SideParty)) > 0
}

// combatCastDimensionDoor 是次元門（法術 83）：把施法者傳送到選定的格子。
//
// 施法者身上有效果 `3Ah` 時，先把身邊半徑 1 內被自己勾住的戰鬥員上的
// `90h`／`8Bh` 解開再走——不解開的話那些標記會留在原地指著已經不在的人。
func (s *State) combatCastDimensionDoor(spellID uint8) error {
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
	if _, err := s.battle.CastDimensionDoor(caster.ID, s.combatSpellTargetPoint); err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	return s.combatFinishSpell(fmt.Sprintf(
		s.catalog.Text("combat_dimension_door", "combat_dimension_door"), caster.Name))
}

// combatCanCastDimensionDoor 判斷次元門現在施得出來嗎。
func (s *State) combatCanCastDimensionDoor(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	_, _, _, err := s.combatSpellPrologue(spellID)
	return err == nil
}

// combatCastFireShield 是火焰護盾（法術 85）。
//
// ★ 原作跳選單讓玩家選熱／冷，由電腦操作時擲 `1d10`（大於 5 選熱）。
// remake 目前兩邊都走那顆 `1d10`——選單還沒接，這是**介面的缺口**，
// 兩種形態掛的效果碼與持續時間都照原作。
func (s *State) combatCastFireShield(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	hot, err := s.battle.RollFireShieldHot()
	if err != nil {
		return err
	}
	entry, _ := gamepack.SpellByID(int(spellID))
	duration := entry.PrimaryDuration(casterLevel(s.partyRoster[characterIndex]), false)
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	if _, err := s.battle.CastFireShield(caster.ID, hot, duration); err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	key := "combat_fire_shield_cold"
	if hot {
		key = "combat_fire_shield_hot"
	}
	return s.combatFinishSpell(fmt.Sprintf(s.catalog.Text(key, key), caster.Name))
}

// combatCanCastFireShield 判斷火焰護盾現在施得出來嗎。
func (s *State) combatCanCastFireShield(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	_, _, _, err := s.combatSpellPrologue(spellID)
	return err == nil
}

// combatCastConeOfCold 是寒冰錐（法術 92）：半徑與骰子都由施法者等級算出來。
//
// ★ 半徑**不在**法術屬性表裡（表裡是 0），是 handler 自己算的
// `(等級 ＋ 1) div 2` 最小 1。照表用 0 只會打中心那一格。
func (s *State) combatCastConeOfCold(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	level := casterLevel(s.partyRoster[characterIndex])
	dice, ok := combat.SpellDamageRoll(spellID, level)
	if !ok {
		return fmt.Errorf("spell 0x%02X has no usable dice", spellID)
	}
	entry, ok := gamepack.SpellByID(int(spellID))
	if !ok {
		return fmt.Errorf("spell 0x%02X is not in the spell table", spellID)
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	result, err := s.battle.CastAreaDamageDice(caster.ID, s.combatSpellTargetPoint,
		dice, combat.ConeOfColdRadius(level), combat.SpellDamageElement(spellID),
		entry.RequiresSave(), entry.SaveCategory, entry.SaveHalvesDamage())
	if err != nil {
		s.partyRoster[characterIndex].SpellSlots = append(
			s.partyRoster[characterIndex].SpellSlots, spellID)
		return err
	}
	return s.combatFinishSpell(fmt.Sprintf(
		s.catalog.Text("combat_area_damage_spell", "combat_area_damage_spell"),
		caster.Name, len(result.Impacts), result.BaseDamage))
}

// combatCanCastConeOfCold 判斷寒冰錐現在施得出來嗎。
func (s *State) combatCanCastConeOfCold(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	if _, _, _, err := s.combatSpellPrologue(spellID); err != nil {
		return false
	}
	if _, ok := combat.SpellDamageRoll(spellID, 1); !ok {
		return false
	}
	return len(s.livingBySide(combat.SideEnemy)) > 0
}

// combatCastRestoration 是復原術（法術 56，`overlay-22@3D27h`）：把被吸掉的
// 一級還回去。
//
// 原作的順序：
//
//  1. `+0E7h`（被吸的級數）為 0 就什麼都不做。
//  2. 每級的 HP ＝ `+0E8h ÷ +0E7h`（整數除），三格 HP（最大、基準最大、目前）
//     各加上去，再從 `+0E8h` 扣掉、`+0E7h` 減一。
//  3. 挑一個職業把等級加回去：走八個職業，只看**等級不高於目前最佳**且
//     下一級門檻不為 0 的，取門檻**最小**的那一個。
//  4. 角色經驗值不足那個門檻時補到剛好。
//
// ★ 第 3 步的兩個條件是**同時**成立的（`lvl <= 最佳等級` 而且門檻更小），
// 而且「最佳等級」會隨著挑選過程往下收。只比門檻會挑到高等職業。
func (s *State) combatCastRestoration(spellID uint8) error {
	if s.combatCastingSpell != 0 && s.combatCastingSpell != spellID {
		return fmt.Errorf("a different spell target is being selected")
	}
	characterIndex, spellIndex, caster, err := s.combatSpellPrologue(spellID)
	if err != nil {
		return err
	}
	targets := s.livingBySide(combat.SideParty)
	if len(targets) == 0 {
		return fmt.Errorf("no party member can be targeted")
	}
	index := s.combatSpellTargetIndex
	if index < 0 || index >= len(targets) {
		index = 0
	}
	target := targets[index]
	targetIndex := -1
	for position := range s.partyRoster {
		if s.partyRoster[position].ID == target.ID {
			targetIndex = position
			break
		}
	}
	if targetIndex < 0 {
		return fmt.Errorf("target %q is not in the roster", target.ID)
	}
	s.partyRoster[characterIndex].SpellSlots = append(
		s.partyRoster[characterIndex].SpellSlots[:spellIndex],
		s.partyRoster[characterIndex].SpellSlots[spellIndex+1:]...)
	restored := restoreDrainedLevel(&s.partyRoster[targetIndex])
	message := s.catalog.Text("combat_restoration_none", "combat_restoration_none")
	if restored {
		message = s.catalog.Text("combat_restoration", "combat_restoration")
		if err := s.syncTempleCharacter(targetIndex); err != nil {
			return err
		}
	}
	return s.combatFinishSpell(fmt.Sprintf(message, caster.Name, target.Name))
}

// restoreDrainedLevel 把一級還回去，回傳有沒有還成。
func restoreDrainedLevel(character *party.Character) bool {
	if character == nil || character.DrainedLevels <= 0 {
		return false
	}
	perLevel := character.DrainedHitPoints / character.DrainedLevels
	character.MaxHitPoints += perLevel
	character.BaseMaxHitPoints += perLevel
	character.HitPoints += perLevel
	character.DrainedHitPoints -= perLevel
	character.DrainedLevels--

	bestSlot, bestLevel, bestThreshold := -1, 13, uint32(0)
	for _, info := range trainingClasses {
		level := character.ClassLevel(info.Class)
		if level <= 0 || level > bestLevel || level >= len(info.Threshold) {
			continue
		}
		threshold := info.Threshold[level]
		if threshold == 0 {
			continue
		}
		if bestSlot >= 0 && threshold >= bestThreshold {
			continue
		}
		bestSlot, bestLevel, bestThreshold = info.Slot, level, threshold
	}
	if bestSlot < 0 {
		return true
	}
	character.ClassLevels[bestSlot]++
	if character.Experience < bestThreshold {
		character.Experience = bestThreshold
	}
	return true
}

// combatCanCastRestoration 判斷復原術現在施得出來嗎。
func (s *State) combatCanCastRestoration(spellID uint8) bool {
	if !s.CombatActive() {
		return false
	}
	if _, _, _, err := s.combatSpellPrologue(spellID); err != nil {
		return false
	}
	return len(s.livingBySide(combat.SideParty)) > 0
}
