package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// 戰鬥裡的 `使用`：充能物品那一條（spec 921 的分派、spec 1169 的九支 handler）。
//
// ★ 這條路上**沒有一個規則是新寫的**。效果編號 ＝ `物品^[3Dh] and 7Fh`，
// 那個編號就是法術主表的列；目標模式（`+6`）、豁免（`+8`）、效果碼（`+0Ah`）、
// 持續時間係數（`+2`／`+3`）全部查表，骰子查 spec 1169 的 handler 表。
//
// ⚠ 卷軸那一條（物品類別 `3Ch`..`3Eh`）還沒接：原作用 `sub_3474h` 讓玩家從
// 卷軸上的三個法術挑一個，那一支沒有讀出來。按下去會回錯誤，**不會**默默扣掉
// 卷軸。

// itemCasterLevel 是物品施放時的施法者等級。
//
// ★ 原作在 `DS:7563h` 非 0（效果來自物品）時等級一律當 6，不查使用者
// （spec 733／1016）。所以魔杖的火球永遠是 6d6，不會因為誰拿著而變強。
const itemCasterLevel = 6

// scrollItemTypes 是三種卷軸的物品類別（spec 1168）。
var scrollItemTypes = map[uint8]bool{0x3C: true, 0x3D: true, 0x3E: true}

// UsableItem 是「這個角色身上按 USE 有反應的東西」。
type UsableItem struct {
	// Index 是角色 Equipment 裡的位置。
	Index int    `json:"index"`
	Name  string `json:"name"`
	// Effect 是 `+3Dh and 7Fh`。
	Effect uint8 `json:"effect"`
	// Charges 是 `+3Ch`，Count 是 `+39h`（一疊物品的最後一個才會消耗充能）。
	Charges uint8 `json:"charges"`
	Count   uint8 `json:"count"`
}

// chargedItemEffect 依原作的判準取出充能物品的效果編號（spec 921）。
//
// ⚠ 判準是**物品的三個欄位**，不是物品類別：`+3Dh > 0` 且 `+3Eh < 80h`。
// 用類別去判會漏掉粉塵與項鍊那兩件（它們的類別是 `46h`）。
func chargedItemEffect(item monster.ItemRecord) (uint8, bool) {
	if scrollItemTypes[item.Type] {
		return 0, false
	}
	if item.Affects[1] == 0 || item.Affects[2] >= 0x80 {
		return 0, false
	}
	return item.Affects[1] & 0x7F, true
}

// CombatUsableItems 列出目前行動的角色身上有反應的物品。
func (s *State) CombatUsableItems() []UsableItem {
	fighter, ok := s.combatPartyTurn()
	if !ok {
		return nil
	}
	for _, character := range s.partyRoster {
		if character.ID != fighter.ID {
			continue
		}
		items := make([]UsableItem, 0, len(character.Equipment))
		for index, item := range character.Equipment {
			effect, charged := chargedItemEffect(item)
			if !charged || item.Affects[0] == 0 {
				continue
			}
			if _, known := combat.LookupChargedItemBehaviour(effect); !known {
				continue
			}
			items = append(items, UsableItem{
				Index: index, Name: item.Name, Effect: effect,
				Charges: item.Affects[0], Count: item.Count,
			})
		}
		return items
	}
	return nil
}

// CombatSelectedItem 回傳目前選中的可用物品。
func (s *State) CombatSelectedItem() (UsableItem, bool) {
	items := s.CombatUsableItems()
	if len(items) == 0 {
		return UsableItem{}, false
	}
	index := s.combatItemIndex
	if index < 0 || index >= len(items) {
		index = 0
	}
	return items[index], true
}

// CombatSelectItem 換一件要用的物品。
func (s *State) CombatSelectItem(delta int) error {
	items := s.CombatUsableItems()
	if len(items) == 0 {
		return fmt.Errorf("nobody in the party has a usable item")
	}
	s.combatItemIndex = (s.combatItemIndex + delta) % len(items)
	if s.combatItemIndex < 0 {
		s.combatItemIndex += len(items)
	}
	selected := items[s.combatItemIndex]
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_item_selected", "combat_item_selected"), selected.Name)
	return nil
}

// CombatUseItem 用掉目前選中的物品。
func (s *State) CombatUseItem() error {
	selected, ok := s.CombatSelectedItem()
	if !ok {
		return fmt.Errorf("nobody in the party has a usable item")
	}
	return s.CombatUseItemAt(selected.Index)
}

// CombatUseItemAt 用掉背包第 index 件物品。
func (s *State) CombatUseItemAt(index int) error {
	user, ok := s.combatPartyTurn()
	if !ok {
		return fmt.Errorf("it is not a living party turn")
	}
	characterIndex := -1
	for position, character := range s.partyRoster {
		if character.ID == user.ID {
			characterIndex = position
			break
		}
	}
	if characterIndex < 0 {
		return fmt.Errorf("fighter %q is not in the party roster", user.ID)
	}
	equipment := s.partyRoster[characterIndex].Equipment
	if index < 0 || index >= len(equipment) {
		return fmt.Errorf("equipment index %d is out of range", index)
	}
	item := equipment[index]
	if scrollItemTypes[item.Type] {
		// spec 1168：`sub_3474h`（從卷軸上挑一支法術）還沒讀出來。
		return fmt.Errorf("scroll item type 0x%02X has no reading path yet", item.Type)
	}
	effect, charged := chargedItemEffect(item)
	if !charged {
		// 原作對這一類就是「什麼都不會發生」，連訊息都沒有（spec 921）。
		return fmt.Errorf("item type 0x%02X does nothing when used", item.Type)
	}
	if item.Affects[0] == 0 {
		return fmt.Errorf("item %q has no charges left", item.Name)
	}
	behaviour, known := combat.LookupChargedItemBehaviour(effect)
	if !known {
		return fmt.Errorf("item effect 0x%02X has no handler reading yet", effect)
	}
	entry, found := gamepack.SpellByID(int(effect))
	if !found {
		return fmt.Errorf("item effect 0x%02X is outside the original spell table", effect)
	}

	consumed, err := s.resolveChargedItem(user, item, entry, behaviour)
	if err != nil {
		return err
	}
	if consumed {
		// spec 921：數量 > 1 就只扣數量，否則扣充能；充能歸零就把物品處理掉。
		// **一疊物品的最後一個才會消耗充能。**
		if s.partyRoster[characterIndex].Equipment[index].Count > 1 {
			s.partyRoster[characterIndex].Equipment[index].Count--
		} else {
			s.partyRoster[characterIndex].Equipment[index].Affects[0]--
			if s.partyRoster[characterIndex].Equipment[index].Affects[0] == 0 {
				s.partyRoster[characterIndex].Equipment = append(
					s.partyRoster[characterIndex].Equipment[:index],
					s.partyRoster[characterIndex].Equipment[index+1:]...)
				s.combatItemIndex = 0
			}
		}
	}
	s.requestSound(SoundCast)
	if s.battle.Status() != combat.StatusActive {
		return s.finishCombat()
	}
	s.combatTurnIndex++
	return s.advanceCombatToParty()
}

// resolveChargedItem 依 handler 的形狀結算，回傳「這一次算不算真的用掉了」。
//
// ⚠ 回傳 false 只有一個情況：速度藥水對已經被緩速的目標（`39h` 的 `entry#16`
// 查詢）。原作那一條不呼叫 `sub_F06`，`結果^` 保持 0，所以**充能不會被扣**。
func (s *State) resolveChargedItem(user combat.Fighter, item monster.ItemRecord,
	entry gamepack.SpellEntry, behaviour combat.ChargedItemBehaviour) (bool, error) {
	targets, center, err := s.chargedItemTargets(user, entry, behaviour)
	if err != nil {
		return false, err
	}
	switch behaviour.Shape {
	case combat.ChargedItemNamedSpell:
		return s.resolveNamedSpellItem(user, item, entry, center)
	case combat.ChargedItemEffect:
		return s.resolveChargedItemEffect(user, item, entry, behaviour, targets)
	case combat.ChargedItemGiantStrength:
		return s.resolveGiantStrengthItem(user, item, behaviour, targets)
	case combat.ChargedItemHeal:
		if len(targets) == 0 {
			return false, fmt.Errorf("item %q has no target", item.Name)
		}
		result, healErr := s.battle.CastHealingDice(user.ID, targets[0], behaviour.Dice)
		if healErr != nil {
			return false, healErr
		}
		healed, _ := s.fighter(targets[0])
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text("combat_item_heal", "combat_item_heal"),
			user.Name, item.Name, healed.Name, result.Healing)
		s.requestSound(SoundSpellHit)
		return true, nil
	case combat.ChargedItemDamage:
		if len(targets) == 0 {
			return false, fmt.Errorf("item %q has no target", item.Name)
		}
		return s.resolveChargedItemDamage(user, item, entry, behaviour, targets[0])
	case combat.ChargedItemAreaDamage:
		return s.resolveChargedItemArea(user, item, entry, behaviour, center)
	default:
		return false, fmt.Errorf("item effect shape %d is not implemented", behaviour.Shape)
	}
}

// chargedItemTargets 依**原版主表**的目標模式取目標（spec 1009）。
//
// ★ `+6` 給形狀（自己／固定 N 個／半徑 r），`+7` 給「這一支是對誰用的」：
// 1 ＝ 使用者自己、4 ＝ 我方、0 ＝ 敵方。九個物品效果的兩欄組合完全一致
// （`self` 全是 `+7 = 1`、粉塵是 `+7 = 4`、其餘戰鬥用的都是 `+7 = 0`），
// 所以這裡照兩欄一起判。
//
// ⚠ `+7` 在 spec 719 是**戰鬥外**那條選目標路徑的欄位，拿來當戰鬥裡的敵我判別
// 是推論，不是讀出來的。所以遇到沒見過的組合一律回錯誤——猜一個會得到自洽但
// 沒有依據的規則。
func (s *State) chargedItemTargets(user combat.Fighter, entry gamepack.SpellEntry,
	behaviour combat.ChargedItemBehaviour) ([]string, combat.TilePoint, error) {
	center := combat.TilePoint{X: user.CombatX, Y: user.CombatY}
	if !behaviour.CenterOnUser && s.combatSpellTargetsPoint {
		center = s.combatSpellTargetPoint
	}
	switch entry.TargetModeKind {
	case "self":
		return []string{user.ID}, center, nil
	case "area":
		return nil, center, nil
	case "fixed":
		switch entry.SelectMode {
		case 1:
			return []string{user.ID}, center, nil
		case 4:
			living := s.livingBySide(combat.SideParty)
			limit := entry.TargetCount
			if limit <= 0 || limit > len(living) {
				limit = len(living)
			}
			targets := make([]string, 0, limit)
			for _, fighter := range living[:limit] {
				targets = append(targets, fighter.ID)
			}
			return targets, center, nil
		case 0:
			enemies := s.livingBySide(combat.SideEnemy)
			if len(enemies) == 0 {
				return nil, center, fmt.Errorf("no living enemy to target")
			}
			index := s.combatTargetIndex
			if index >= len(enemies) {
				index = 0
			}
			target := enemies[index]
			return []string{target.ID},
				combat.TilePoint{X: target.CombatX, Y: target.CombatY}, nil
		}
	}
	return nil, center, fmt.Errorf(
		"item effect target mode %q with select mode %d has no reading",
		entry.TargetModeKind, entry.SelectMode)
}

// resolveChargedItemEffect 走 `+0Ah` 那條資料驅動的效果路（spec 1117）。
func (s *State) resolveChargedItemEffect(user combat.Fighter, item monster.ItemRecord,
	entry gamepack.SpellEntry, behaviour combat.ChargedItemBehaviour,
	targets []string) (bool, error) {
	if entry.EffectID == 0 {
		return false, fmt.Errorf("item effect 0x%02X has no effect code in the table", entry.SpellID)
	}
	if len(targets) == 0 {
		return false, fmt.Errorf("item %q has no target", item.Name)
	}
	if behaviour.BlockedByAffect != 0 {
		blocked := true
		for _, id := range targets {
			if !s.battle.FighterHasAffect(id, behaviour.BlockedByAffect) {
				blocked = false
				break
			}
		}
		if blocked {
			// 原作連 `sub_F06` 都不呼叫 ⇒ `結果^` ＝ 0 ⇒ 充能不扣（spec 1169）。
			s.combatMessage = fmt.Sprintf(
				s.catalog.Text("combat_item_no_effect", "combat_item_no_effect"),
				user.Name, item.Name)
			return false, nil
		}
	}
	result, err := s.battle.CastEffectSpell(user.ID, targets, combat.EffectSpellRequest{
		SpellID:    uint8(entry.SpellID),
		EffectKind: uint8(entry.EffectID),
		// fromItem ＝ true：等級一律當 6，不查使用者（spec 733）。
		Duration:     entry.PrimaryDuration(itemCasterLevel, true),
		SaveKind:     entry.SaveKind,
		SaveCategory: entry.SaveCategory,
		CasterLevel:  itemCasterLevel,
	})
	if err != nil {
		return false, err
	}
	applied := 0
	for _, impact := range result.Impacts {
		if impact.Applied {
			applied++
		}
	}
	if applied > 0 {
		s.requestSound(SoundSpellHit)
	}
	s.combatMessage = s.chargedItemMessage(user, item, behaviour, applied, targets)
	return true, nil
}

// resolveGiantStrengthItem 是巨人力量藥水（`3Bh`）。
//
// ★ 原作的形狀（spec 1169）：`entry#19(@舊值, 0, 21, 目標)` 只在 **21 比目前的
// 力量高**時才換上去，成功才印 `is stronger`；接著不論成不成功都掛一筆效果
// 記錄，種類 `92h`、時間 `(1d4 ＋ 4) × 10`，並把舊值存在記錄裡。
//
// ⚠ 到期把力量換回去那一段**還沒接**：`92h` 的到期處理沒有讀出來。記錄照掛，
// 所以資料在，之後接得上。
func (s *State) resolveGiantStrengthItem(user combat.Fighter, item monster.ItemRecord,
	behaviour combat.ChargedItemBehaviour, targets []string) (bool, error) {
	if len(targets) == 0 {
		return false, fmt.Errorf("item %q has no target", item.Name)
	}
	characterIndex := -1
	for position, character := range s.partyRoster {
		if character.ID == targets[0] {
			characterIndex = position
			break
		}
	}
	if characterIndex < 0 {
		return false, fmt.Errorf("giant strength target %q is not in the party roster", targets[0])
	}
	abilities := s.partyRoster[characterIndex].Abilities
	previous := abilities.StrengthFull
	if previous == 0 {
		previous = abilities.Strength
	}
	raised := behaviour.Strength > previous
	if raised {
		s.partyRoster[characterIndex].Abilities.StrengthFull = behaviour.Strength
		s.partyRoster[characterIndex].Abilities.StrengthExceptional = 0
	}
	duration := (s.battle.RollDie(4) + 4) * 10
	s.partyRoster[characterIndex].Effects = append(s.partyRoster[characterIndex].Effects,
		monster.AffectRecord{
			Kind: giantStrengthAffectKind, Value: uint16(duration),
			Duration: uint16(duration), Strength: uint8(previous), Active: true,
		})
	target, _ := s.fighter(targets[0])
	if raised {
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text("combat_item_stronger", "combat_item_stronger"),
			user.Name, item.Name, target.Name)
		s.requestSound(SoundSpellHit)
	} else {
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text("combat_item_no_effect", "combat_item_no_effect"),
			user.Name, item.Name)
	}
	return true, nil
}

// giantStrengthAffectKind 是巨人力量藥水掛上去的效果記錄種類（spec 1169：
// `entry#11` 的第一個參數是 `92h`）。
const giantStrengthAffectKind uint8 = 0x92

// resolveChargedItemDamage 是單一目標傷害（魔法飛彈魔杖 `41h`）。
func (s *State) resolveChargedItemDamage(user combat.Fighter, item monster.ItemRecord,
	entry gamepack.SpellEntry, behaviour combat.ChargedItemBehaviour,
	targetID string) (bool, error) {
	dice, err := s.battle.RollChargedItemDice(behaviour)
	if err != nil {
		return false, err
	}
	halved, negated := false, false
	if entry.RequiresSave() {
		targetFighter, _ := s.fighter(targetID)
		save, saveErr := s.battle.RollSavingThrow(targetFighter, entry.SaveCategory, 0)
		if saveErr != nil {
			return false, saveErr
		}
		if save.Saved {
			halved = entry.SaveHalvesDamage()
			negated = !halved
		}
	}
	damage := 0
	if !negated {
		result, castErr := s.battle.CastDamageDice(user.ID, targetID, dice,
			combat.SpellDamageElement(uint8(entry.SpellID)), halved)
		if castErr != nil {
			return false, castErr
		}
		damage = result.Damage
		s.requestSound(SoundSpellHit)
	}
	targetFighter, _ := s.fighter(targetID)
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_item_damage", "combat_item_damage"),
		user.Name, item.Name, targetFighter.Name, damage)
	return true, nil
}

// resolveChargedItemArea 是範圍傷害（飛彈項鍊 `40h`、比克魔杖 `62h`）。
//
// ⚠ 兩件的圓心不同：項鍊用玩家瞄的那一點，比克魔杖的 handler 自己從使用者
// 算圓心（spec 1169）。`CenterOnUser` 就是這件事。
func (s *State) resolveChargedItemArea(user combat.Fighter, item monster.ItemRecord,
	entry gamepack.SpellEntry, behaviour combat.ChargedItemBehaviour,
	center combat.TilePoint) (bool, error) {
	dice, err := s.battle.RollChargedItemDice(behaviour)
	if err != nil {
		return false, err
	}
	if behaviour.RaceTypeOnly != 0 {
		// 只有那個 `RACETYPE` 吃得到；其餘目標在原作是「豁免成功 ＋ `+8 = 1`」
		// ⇒ 傷害 0。這裡直接不列進去，結果相同而且不會浪費豁免擲骰。
		targets := s.battle.ChargedItemDamageTargets(behaviour, center, entry.AreaRadius)
		hit := 0
		total := 0
		for _, id := range targets {
			result, castErr := s.battle.CastDamageDice(user.ID, id, dice,
				combat.SpellDamageElement(uint8(entry.SpellID)), false)
			if castErr != nil {
				return false, castErr
			}
			hit++
			total += result.Damage
		}
		if hit > 0 {
			s.requestSound(SoundSpellHit)
		}
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text("combat_item_area", "combat_item_area"),
			user.Name, item.Name, hit, total)
		return true, nil
	}
	result, castErr := s.battle.CastAreaDamageDice(user.ID, center, dice,
		entry.AreaRadius, combat.SpellDamageElement(uint8(entry.SpellID)),
		entry.RequiresSave(), entry.SaveCategory, entry.SaveHalvesDamage())
	if castErr != nil {
		return false, castErr
	}
	s.requestSound(SoundSpellHit)
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_item_area", "combat_item_area"),
		user.Name, item.Name, len(result.Impacts), result.BaseDamage)
	return true, nil
}

// chargedItemMessage 組出效果類物品的訊息。handler 有自己那一句就用它
// （`is Speedy`／`is paralyzed`／`is invisible`），沒有的就用通用句。
func (s *State) chargedItemMessage(user combat.Fighter, item monster.ItemRecord,
	behaviour combat.ChargedItemBehaviour, applied int, targets []string) string {
	if applied == 0 {
		return fmt.Sprintf(
			s.catalog.Text("combat_item_no_effect", "combat_item_no_effect"),
			user.Name, item.Name)
	}
	if behaviour.MessageKey != "" && len(targets) == 1 {
		target, _ := s.fighter(targets[0])
		return fmt.Sprintf(
			s.catalog.Text(behaviour.MessageKey, behaviour.MessageKey),
			user.Name, item.Name, target.Name)
	}
	return fmt.Sprintf(
		s.catalog.Text("combat_item_effect", "combat_item_effect"),
		user.Name, item.Name, applied)
}

// resolveNamedSpellItem 是效果編號指到**有名字**的法術那三支魔杖：
// 火球（`2Fh`）、閃電（`33h`）、冰風暴（`57h`）。
//
// ★ 三支都走既有的施法路，只有一個差別：**等級一律 6**（`itemCasterLevel`）。
// 所以火球魔杖永遠是 6d6，不會因為誰拿著而變強。
func (s *State) resolveNamedSpellItem(user combat.Fighter, item monster.ItemRecord,
	entry gamepack.SpellEntry, center combat.TilePoint) (bool, error) {
	switch uint8(entry.SpellID) {
	case FireballSpellID:
		result, err := s.battle.CastFireball(user.ID, center, itemCasterLevel)
		if err != nil {
			return false, err
		}
		s.requestSound(SoundSpellHit)
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text("combat_item_area", "combat_item_area"),
			user.Name, item.Name, len(result.Impacts), result.BaseDamage)
		return true, nil
	case LightningBoltSpellID:
		result, err := s.battle.CastReflectingLineSpell(
			user.ID, LightningBoltSpellID, center, itemCasterLevel,
			combat.ReflectingLineOptions{
				WeightedBudget: 14, FirstReflectionOriginThreshold: 8,
				FirstReflectionPenalty: 8,
				DamageFlags:            combat.DamageFlagElectricity | combat.DamageFlagMagic,
			},
			s.combatLineTerrain,
		)
		if err != nil {
			return false, err
		}
		s.requestSound(SoundSpellHit)
		s.combatMessage = fmt.Sprintf(
			s.catalog.Text("combat_item_area", "combat_item_area"),
			user.Name, item.Name, len(result.Impacts), result.BaseDamage)
		return true, nil
	}
	dice, ok := combat.SpellDamageRoll(uint8(entry.SpellID), itemCasterLevel)
	if !ok {
		return false, fmt.Errorf("item spell 0x%02X has no usable dice", entry.SpellID)
	}
	result, err := s.battle.CastAreaDamageDice(user.ID, center, dice,
		entry.AreaRadius, combat.SpellDamageElement(uint8(entry.SpellID)),
		entry.RequiresSave(), entry.SaveCategory, entry.SaveHalvesDamage())
	if err != nil {
		return false, err
	}
	s.requestSound(SoundSpellHit)
	s.combatMessage = fmt.Sprintf(
		s.catalog.Text("combat_item_area", "combat_item_area"),
		user.Name, item.Name, len(result.Impacts), result.BaseDamage)
	return true, nil
}
