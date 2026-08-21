package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 戰鬥裡的 `使用`：卷軸那一條（spec 921 的失敗判定、spec 1171 的三支常式）。
//
// ★ 卷軸與充能物品的差別只有三件事：法術是**從三個槽裡選一支**、唸得成唸不成
// 要看職業（非施法職業多半失敗）、消耗的是**槽與名字編號**而不是充能。
// 結算本身走同一條路（`resolveChargedItem`），等級一樣是 6。

// readMagicAffectKind 是「法術辨識」的效果碼：`overlay-22:08A1h` 先查它
// （`overlay-24 entry#27` 找節點），有就把卷軸的隱藏名稱旗標清掉（spec 1171）。
const readMagicAffectKind uint8 = 0x10

// scrollNameNumberFloor 是卷軸用完的門檻。名字編號 `+30h` 從 `0D4h` 起算，
// 每唸掉一支減一；**低於 `0D2h` 就把卷軸丟掉**——`0D4h − 2` 正好是「唸完第三支」。
//
// ★ 這個編號同時是玩家看到的名字：`0D4h` ＝ `With 3 Spells`。所以「還剩幾支」
// 不是另一個欄位，就是名字本身。
const (
	scrollNameNumberBase  uint8 = 0xD4
	scrollNameNumberFloor uint8 = 0xD2
)

// ScrollSpell 是卷軸上的一支。
type ScrollSpell struct {
	// Slot 是三個槽的第幾個（1..3），也就是 `+3Ch`／`+3Dh`／`+3Eh`。
	Slot int `json:"slot"`
	// SpellID 是原作編號（已去掉 bit 7）。
	SpellID uint8  `json:"spell_id"`
	Name    string `json:"name"`
}

// CombatScrollSpells 列出卷軸上讀得出來的法術。
//
// ⚠ 讀不出來時回空清單，不是錯誤：原作在那個狀態下連選單都不會出現
// （`overlay-22:08A1h` 的 `物品^[35h] <> 0` 那一路）。
func (s *State) CombatScrollSpells(index int) []ScrollSpell {
	user, ok := s.combatPartyTurn()
	if !ok {
		return nil
	}
	characterIndex := s.partyIndexOf(user.ID)
	if characterIndex < 0 {
		return nil
	}
	equipment := s.partyRoster[characterIndex].Equipment
	if index < 0 || index >= len(equipment) {
		return nil
	}
	item := equipment[index]
	if !s.itemCatalog.IsScroll(item.Type) {
		return nil
	}
	if !s.scrollIsReadable(characterIndex, item) {
		return nil
	}
	spells := make([]ScrollSpell, 0, 3)
	for slot, raw := range item.Affects {
		if raw == 0 {
			continue
		}
		spellID := raw & 0x7F
		name := ""
		if entry, found := gamepack.SpellByID(int(spellID)); found {
			name = entry.Name
		}
		spells = append(spells, ScrollSpell{Slot: slot + 1, SpellID: spellID, Name: name})
	}
	return spells
}

// scrollIsReadable 重現 `overlay-22:08A1h` 的前半段。
//
// ★ 原作是「符合條件就把 `+35h` 清成 0，然後 `+35h` 必須是 0 才列得出法術」。
// `+35h` 是**隱藏名稱旗標**——所以同一個位元組既決定看不看得到法術，也決定
// 玩家看到的是 `Magic User Scroll` 還是 `Magic User Scroll With 3 Spells`。
//
// ⚠ 條件 B 的第二個析取項（`+111h > +0E6h`）沒有讀出來，這裡只實作
// 「是牧師」那一半。影響範圍：多職角色可能比原作嚴格一格。
func (s *State) scrollIsReadable(characterIndex int, item monster.ItemRecord) bool {
	if item.HiddenNameFlags == 0 {
		return true
	}
	character := s.partyRoster[characterIndex]
	if s.battle != nil && s.battle.FighterHasAffect(character.ID, readMagicAffectKind) {
		return true
	}
	// 原作條件 A 的第一個析取項是 `+109h > 0` ＝ 牧師槽有等級。
	if character.ClassLevel(party.ClassCleric) <= 0 {
		return false
	}
	slot, ok := s.itemCatalog.ItemSlot(item.Type)
	return ok && slot == monster.ClericalScrollSlot
}

// CombatSelectScrollSpell 換一支要唸的法術。
func (s *State) CombatSelectScrollSpell(delta int) error {
	selected, ok := s.CombatSelectedItem()
	if !ok {
		return fmt.Errorf("no usable item is selected")
	}
	spells := s.CombatScrollSpells(selected.Index)
	if len(spells) == 0 {
		return fmt.Errorf("the selected item has no readable spell")
	}
	s.combatScrollSpellIndex = (s.combatScrollSpellIndex + delta) % len(spells)
	if s.combatScrollSpellIndex < 0 {
		s.combatScrollSpellIndex += len(spells)
	}
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_scroll_selected", "combat_scroll_selected"),
		spells[s.combatScrollSpellIndex].Name)
	return nil
}

// useScrollAt 唸掉卷軸上目前選中的那一支。
func (s *State) useScrollAt(user combat.Fighter, characterIndex, index int) error {
	item := s.partyRoster[characterIndex].Equipment[index]
	if !s.scrollIsReadable(characterIndex, item) {
		return fmt.Errorf("scroll %q cannot be read yet", item.Name)
	}
	spells := s.CombatScrollSpells(index)
	if len(spells) == 0 {
		return fmt.Errorf("scroll %q has no spell left", item.Name)
	}
	position := s.combatScrollSpellIndex
	if position < 0 || position >= len(spells) {
		position = 0
	}
	chosen := spells[position]
	entry, found := gamepack.SpellByID(int(chosen.SpellID))
	if !found {
		return fmt.Errorf("scroll spell 0x%02X is outside the original spell table", chosen.SpellID)
	}
	behaviour, known := combat.DeriveSpellItemBehaviour(chosen.SpellID, entry.EffectID,
		entry.TargetModeKind == "area")
	if !known {
		// ⚠ 這一支的 handler 還沒讀。回錯誤而**不消耗卷軸**——安靜地失敗會把
		// 玩家的卷軸吃掉，而且看起來像原作的 `'oops!'`。
		return fmt.Errorf("scroll spell 0x%02X has no handler reading yet", chosen.SpellID)
	}

	if !s.scrollCastSucceeds(characterIndex) {
		// 原作只印 `'oops!'`，**沒有任何說明**，而且卷軸不會被消耗（spec 921）。
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text("combat_scroll_oops", "combat_scroll_oops"), user.Name)
		s.requestSound(SoundCast)
		if s.battle.Status() != combat.StatusActive {
			return s.finishCombat()
		}
		s.combatTurnIndex++
		return s.advanceCombatToParty()
	}

	consumed, err := s.resolveChargedItem(user, item, entry, behaviour)
	if err != nil {
		return err
	}
	if consumed {
		s.consumeScrollSpell(characterIndex, index, chosen)
	}
	s.requestSound(SoundCast)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// scrollCastSucceeds 是 spec 921 的失敗判定。
//
// ★★ 三條路正好是 AD&D 1e 的卷軸規則：
//
//	牧師槽（`+109h`）或法師槽（`+10Eh`）有等級 → **一定成功**
//	盜賊槽（`+10Fh`）等級 > 9              → 1d100 ≤ 75 才成功
//	其餘                                    → **一定失敗**
//
// 1e 的「十級以上盜賊有 75% 機率讀得懂卷軸」就是中間那一條。
func (s *State) scrollCastSucceeds(characterIndex int) bool {
	character := s.partyRoster[characterIndex]
	// `+109h` 是牧師槽、`+10Eh` 是法師槽、`+10Fh` 是盜賊槽——`ClassLevel` 的
	// 索引（0／5／6）與原作的偏移一一對應。
	if character.ClassLevel(party.ClassCleric) > 0 || character.ClassLevel(party.ClassMagicUser) > 0 {
		return true
	}
	if character.ClassLevel(party.ClassThief) <= 9 {
		return false
	}
	return s.battle.RollDie(100) <= 75
}

// consumeScrollSpell 重現 `overlay-22 entry#7`：清掉那個槽、名字編號減一，
// 低於門檻就把卷軸丟掉。
//
// ⚠ 比對用的是 `and 7Fh` 之後的編號，寫回去的是 **0**（整格清掉），
// 所以 bit 7 也一起沒了。
func (s *State) consumeScrollSpell(characterIndex, index int, chosen ScrollSpell) {
	equipment := s.partyRoster[characterIndex].Equipment
	item := &equipment[index]
	slot := -1
	for position, raw := range item.Affects {
		if raw&0x7F == chosen.SpellID {
			slot = position
		}
	}
	if slot < 0 {
		return
	}
	item.Affects[slot] = 0
	item.NameNumbers[1]--
	s.combatScrollSpellIndex = 0
	if item.NameNumbers[1] >= scrollNameNumberFloor {
		return
	}
	s.partyRoster[characterIndex].Equipment = append(equipment[:index], equipment[index+1:]...)
	s.combatItemIndex = 0
}

// partyIndexOf 找出隊伍名冊裡的位置。
func (s *State) partyIndexOf(id string) int {
	for index, character := range s.partyRoster {
		if character.ID == id {
			return index
		}
	}
	return -1
}
