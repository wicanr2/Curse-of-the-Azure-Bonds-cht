package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// AI 回合開頭的自動換裝（spec 1120）。
//
// ★ 原作的換裝是 AI 模組（`COMPTACT`）的一部分，所以它管的是「這一回合由電腦
// 操作的人」——包含開了快速戰鬥的隊員。手動操作的隊員不會被換裝。
//
// ⚠ 這裡只換**武器槽**（類別表 `+0` ＝ 0）。盾牌那一槽原作另有一段，
// 而且會動到派生數值重算，還沒接。

// autoEquipAmmunitionSlots 是角色記錄的 `+17Dh`／`+181h`。裝備槽區塊從 `+151h`
// 起、每格 4 bytes，所以那兩個位移就是第 11 與第 12 格。
const (
	autoEquipAmmunitionSlotA uint8 = 11
	autoEquipAmmunitionSlotB uint8 = 12
)

// classUsabilityBit 是類別表 `+0Dh` 那張遮罩用的職業位元
// （`BaseItem.UsableByMask` 的註解已經記著這組對應）。
func classUsabilityBit(class party.Class) uint8 {
	switch class {
	case party.ClassMagicUser:
		return 1
	case party.ClassCleric:
		return 2
	case party.ClassThief:
		return 4
	case party.ClassFighter:
		return 8
	case party.ClassRanger:
		return 128
	case party.ClassPaladin:
		return 64
	}
	return 0
}

// autoEquipContext 把角色與戰場狀況整理成評分需要的輸入。
func (s *State) autoEquipContext(character party.Character, fighter combat.Fighter) monster.AutoEquipCharacter {
	context := monster.AutoEquipCharacter{
		Alignment: character.Alignment,
		// `+185h` 在評分之前已經被主流程把「已裝備物品的手數」全部扣回去，
		// 所以掃描當下它是 0。保留這個欄位是為了對得上原作的算式。
		HandsInUse:    0,
		EnemyAdjacent: s.autoEquipEnemyAdjacent(fighter),
	}
	for _, class := range []party.Class{party.ClassCleric, party.ClassFighter,
		party.ClassRanger, party.ClassPaladin, party.ClassMagicUser, party.ClassThief} {
		if character.HasClass(class) {
			context.ClassUsabilityMask |= classUsabilityBit(class)
		}
	}
	for _, item := range character.Equipment {
		if !item.Readied {
			continue
		}
		base, ok := s.itemCatalog.Lookup(item.Type)
		if !ok {
			continue
		}
		switch base.Slot {
		case autoEquipAmmunitionSlotA:
			context.AmmunitionSlotA = true
		case autoEquipAmmunitionSlotB:
			context.AmmunitionSlotB = true
		}
	}
	return context
}

// autoEquipEnemyAdjacent 是 `overlay-24 entry#32(角色, 1) <> 0`：半徑 1 內有敵人。
func (s *State) autoEquipEnemyAdjacent(fighter combat.Fighter) bool {
	if s.battle == nil || !fighter.HasCombatPosition {
		return false
	}
	for _, other := range s.battle.Fighters() {
		if other.Side == fighter.Side || other.HitPoints <= 0 || other.Escaped ||
			!other.HasCombatPosition {
			continue
		}
		dx, dy := other.CombatX-fighter.CombatX, other.CombatY-fighter.CombatY
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= 1 && dy <= 1 {
			return true
		}
	}
	return false
}

// autoEquipBeforeAITurn 讓電腦操作的戰鬥員在行動前換上評分最高的武器。
//
// 回傳「有沒有真的換」。找不到角色、沒有物品目錄、或選中的就是現在拿的，
// 都回 false 而不是錯誤——換裝是加分項，不該讓整個回合失敗。
func (s *State) autoEquipBeforeAITurn(fighterID string) (bool, error) {
	if !s.itemCatalogReady || s.battle == nil {
		return false, nil
	}
	index := -1
	for position, character := range s.partyRoster {
		if character.ID == fighterID {
			index = position
			break
		}
	}
	if index < 0 {
		// 不在隊伍名冊裡的是怪物（含暫時盟友）：走物品鏈那一側。
		if fighter, ok := s.fighter(fighterID); ok {
			return s.autoEquipMonster(fighter)
		}
		return false, nil
	}
	if len(s.partyRoster[index].Equipment) == 0 {
		return false, nil
	}
	fighter, ok := s.fighter(fighterID)
	if !ok {
		return false, nil
	}
	character := s.partyRoster[index]
	choice := monster.ChooseAutoEquipWeapon(character.Equipment, s.itemCatalog,
		s.autoEquipContext(character, fighter))
	if choice.Chosen == nil {
		return false, nil
	}
	current := -1
	for position, item := range character.Equipment {
		base, found := s.itemCatalog.Lookup(item.Type)
		if !found || base.Slot != 0 || !item.Readied {
			continue
		}
		current = position
		break
	}
	if current >= 0 {
		// 詛咒的裝備換不掉（spec 1004 §四）。
		if character.Equipment[current].Cursed {
			return false, nil
		}
		if character.Equipment[current].Name == choice.Chosen.Name &&
			character.Equipment[current].Type == choice.Chosen.Type {
			return false, nil
		}
		s.partyRoster[index].Equipment[current].Readied = false
	}
	swapped := false
	for position := range s.partyRoster[index].Equipment {
		item := &s.partyRoster[index].Equipment[position]
		if item.Type == choice.Chosen.Type && item.Name == choice.Chosen.Name && !item.Readied {
			item.Readied = true
			swapped = true
			break
		}
	}
	if !swapped {
		if current >= 0 {
			s.partyRoster[index].Equipment[current].Readied = true
		}
		return false, nil
	}
	updated, err := s.partyRoster[index].FighterWithEquipment(s.itemCatalog)
	if err != nil {
		return false, fmt.Errorf("auto-equip re-projection for %q: %w", fighterID, err)
	}
	if err := s.battle.ReplaceFighterEquipment(fighterID, updated); err != nil {
		return false, err
	}
	return true, nil
}

// autoEquipMonster 是同一支換裝的怪物側（spec 1004／1120）：候選來自
// `MonsterItems` 物品鏈、可用性遮罩用記錄的 `+12Bh`。取捨與隊員側同一支
// `ChooseAutoEquipWeapon`——相鄰時遠程不會被選；近戰候選為 nil 時**照樣換**：
// 把槽 0 武器放下，空手用記錄的天生攻擊（spec 1174 的基準骰）。
// 全遊戲帶「發射型」槽 0 武器的怪只有四隻（CENTAUR、GRENDEL、MONKEY、
// RAKSHASA），牠們貼身時的收弓就是走這條。
func (s *State) autoEquipMonster(fighter combat.Fighter) (bool, error) {
	if !s.itemCatalogReady || len(fighter.MonsterItems) == 0 {
		return false, nil
	}
	records := make([]monster.ItemRecord, len(fighter.MonsterItems))
	for position, item := range fighter.MonsterItems {
		records[position] = monster.ItemRecord{
			Name: item.Name, Type: item.Type, Plus: item.Plus,
			Readied: item.Readied, Cursed: item.Cursed, Count: item.Count,
			Weight: item.Weight, Value: item.Value, Affects: item.Affects,
		}
	}
	context := monster.AutoEquipCharacter{
		ClassUsabilityMask: fighter.ClassUsabilityMask,
		Alignment:          fighter.Alignment,
		HandsInUse:         0,
		EnemyAdjacent:      s.autoEquipEnemyAdjacent(fighter),
	}
	for _, item := range fighter.MonsterItems {
		if !item.Readied {
			continue
		}
		if base, ok := s.itemCatalog.Lookup(item.Type); ok {
			switch base.Slot {
			case autoEquipAmmunitionSlotA:
				context.AmmunitionSlotA = true
			case autoEquipAmmunitionSlotB:
				context.AmmunitionSlotB = true
			}
		}
	}
	choice := monster.ChooseAutoEquipWeapon(records, s.itemCatalog, context)
	chosen := -1
	if choice.Chosen != nil {
		for position := range records {
			if choice.Chosen == &records[position] {
				chosen = position
				break
			}
		}
	}
	current := -1
	for position, item := range fighter.MonsterItems {
		base, ok := s.itemCatalog.Lookup(item.Type)
		if !ok || base.Slot != 0 || !item.Readied {
			continue
		}
		current = position
		break
	}
	if current == chosen {
		return false, nil
	}
	if current >= 0 && fighter.MonsterItems[current].Cursed {
		// 詛咒的裝備換不掉（spec 1004 §四）。
		return false, nil
	}
	items := append([]combat.MonsterItem(nil), fighter.MonsterItems...)
	if current >= 0 {
		items[current].Readied = false
	}
	if chosen >= 0 {
		items[chosen].Readied = true
	}
	updated := fighter
	updated.MonsterItems = items
	updated = monster.ReprojectMonsterWeapon(updated, s.itemCatalog)
	if err := s.battle.ReplaceMonsterItems(fighter.ID, items); err != nil {
		return false, err
	}
	if err := s.battle.ReplaceFighterEquipment(fighter.ID, updated); err != nil {
		return false, err
	}
	return true, nil
}
