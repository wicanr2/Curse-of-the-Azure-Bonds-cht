package combat

// 充能物品在戰鬥裡按 `使用` 會做什麼（spec 921 的分派、spec 1169 的九支 handler）。
//
// ★ 為什麼是一張表而不是九段程式碼。 目標模式、豁免種類、效果碼、持續時間係數
// 全部在原版的法術主表裡（spec 1009／1111），照表走就對了。這裡只留**表裡沒有
// 的那一部分**——骰子與四個 handler 自己多做的事，逐支讀出來的來源寫在 spec 1169。
//
// ⚠ 這張表不是「哪些物品能用」的清單。能不能用由物品自己決定
// （`+3Dh > 0` 且 `+3Eh < 80h`，spec 921）；表裡沒有的效果編號代表**那一支
// handler 還沒讀**，呼叫端要當成錯誤，不可以退回「什麼都不做」——安靜地失敗
// 會讓玩家以為充能被吃掉了。

// ChargedItemShape 是一支物品效果 handler 的形狀。
type ChargedItemShape uint8

const (
	// ChargedItemEffect：只把主表 `+0Ah` 的效果碼套上去（`sub_F06` 傷害傳 0）。
	ChargedItemEffect ChargedItemShape = iota + 1
	// ChargedItemDamage：單一目標傷害。
	ChargedItemDamage
	// ChargedItemAreaDamage：範圍傷害。
	ChargedItemAreaDamage
	// ChargedItemHeal：治療。
	ChargedItemHeal
	// ChargedItemGiantStrength：把力量提到某個值（巨人力量藥水）。
	ChargedItemGiantStrength
	// ChargedItemNamedSpell：效果編號指到一支**有名字**的法術，走既有的施法路。
	ChargedItemNamedSpell
)

// ChargedItemBehaviour 是一支物品效果 handler 讀出來的東西。
type ChargedItemBehaviour struct {
	Shape ChargedItemShape
	// BlockedByAffect 非 0 時：目標身上有這個效果碼就整支不做事。
	// 速度藥水（`39h`）先查 `2Ah`（緩速），有就什麼都不發生。
	BlockedByAffect uint8
	// Dice 是收尾拿到的傷害／治療量。
	Dice SpellDamageFormula
	// RandomDiceCount：骰數本身要現擲。飛彈項鍊（`40h`）是 `(1d3 × 2 ＋ 1)d6`，
	// 由火球的 handler 為它分出來的那一條路算（spec 1169）。
	RandomDiceCount bool
	// CenterOnUser：範圍的圓心取使用者自己的座標，不是玩家瞄的那一點。
	// `62h` 的 handler 自己從施法者算圓心。
	CenterOnUser bool
	// RaceTypeOnly 非 0 時只有這個 `RACETYPE` 吃得到傷害。`62h` 的原作寫法是
	// 「不是這個種類就當作豁免成功」，而它的 `+8` 是 1（豁免完全無效），
	// 等價於「其他人傷害是 0」。
	RaceTypeOnly uint8
	// Strength 是 ChargedItemGiantStrength 要提到的力量值。
	Strength int
	// MessageKey 是 handler 裡那句訊息的 locale 鍵；空字串代表原作沒有訊息。
	MessageKey string
}

// chargedItemBehaviours 的鍵是效果編號 ＝ 物品的 `+3Dh and 7Fh`（spec 921）。
// 十二個編號就是六章 253 件物品用到的全部（`docs/audit/combat-item-use.md`）。
var chargedItemBehaviours = map[uint8]ChargedItemBehaviour{
	// 有名字的法術：魔杖裡裝的是火球／閃電／冰風暴本人。
	0x2F: {Shape: ChargedItemNamedSpell},
	0x33: {Shape: ChargedItemNamedSpell},
	0x57: {Shape: ChargedItemNamedSpell},

	// 主表裡沒有名字的九列（spec 1169）。
	0x39: {Shape: ChargedItemEffect, BlockedByAffect: 0x2A, MessageKey: "combat_item_speedy"},
	0x3B: {Shape: ChargedItemGiantStrength, Strength: 21, MessageKey: "combat_item_stronger"},
	0x3D: {Shape: ChargedItemEffect, MessageKey: "combat_item_paralyzed"},
	0x3F: {Shape: ChargedItemEffect, MessageKey: "combat_item_invisible"},
	0x40: {Shape: ChargedItemAreaDamage, Dice: SpellDamageFormula{Sides: 6}, RandomDiceCount: true},
	0x41: {Shape: ChargedItemDamage, Dice: SpellDamageFormula{Count: 2, Sides: 4, Bonus: 2}},
	0x61: {Shape: ChargedItemEffect},
	0x62: {Shape: ChargedItemAreaDamage, Dice: SpellDamageFormula{Count: 6, Sides: 6},
		CenterOnUser: true, RaceTypeOnly: 0x12},
	// `63h` 的 `is Healed` 由治療那條路自己的訊息帶出來，不走 MessageKey。
	0x63: {Shape: ChargedItemHeal, Dice: SpellDamageFormula{Count: 2, Sides: 4, Bonus: 2}},
}

// LookupChargedItemBehaviour 回傳一個效果編號的 handler 行為。第二個回傳值為
// false 代表**那一支還沒讀**，不是「沒有效果」。
func LookupChargedItemBehaviour(effect uint8) (ChargedItemBehaviour, bool) {
	behaviour, ok := chargedItemBehaviours[effect]
	return behaviour, ok
}

// RollChargedItemDice 把 handler 的骰子算成這一次的實際骰數。
//
// ⚠ 骰數要**在套用之前擲一次**，整片範圍共用——原作的火球那一支就是先算好
// `[bp-1]` 再進 `sub_F06`（spec 1169）。每個目標各擲一次會讓範圍傷害的變異
// 比原版大很多。
func (b *Battle) RollChargedItemDice(behaviour ChargedItemBehaviour) (SpellDamageFormula, error) {
	dice := behaviour.Dice
	if !behaviour.RandomDiceCount {
		return dice, nil
	}
	if b == nil || b.rng == nil {
		return SpellDamageFormula{}, errNoPRNG
	}
	dice.Count = (b.rng.Intn(3)+1)*2 + 1
	return dice, nil
}

// ChargedItemDamageTargets 篩掉吃不到傷害的目標。
//
// `62h` 的 handler 只讓 `RACETYPE = 12h`（Moander 的植物眷屬）真的受傷，
// 其餘目標一律當成豁免成功，而它的 `+8` 是 1 ⇒ 傷害歸 0（spec 1169）。
func (b *Battle) ChargedItemDamageTargets(behaviour ChargedItemBehaviour, center TilePoint, radius int) []string {
	if b == nil {
		return nil
	}
	targets := make([]string, 0, 8)
	for _, id := range b.fighterOrder {
		fighter := b.fighters[id]
		if fighter.HitPoints <= 0 || fighter.Escaped || !fighter.HasCombatPosition {
			continue
		}
		if !fighterFootprintWithinRadius(fighter, center, radius) {
			continue
		}
		if behaviour.RaceTypeOnly != 0 && fighter.raceType() != behaviour.RaceTypeOnly {
			continue
		}
		targets = append(targets, id)
	}
	return targets
}

// FighterHasAffect 回報戰鬥員身上有沒有某個效果碼還活著。
func (b *Battle) FighterHasAffect(fighterID string, kind uint8) bool {
	if b == nil {
		return false
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return false
	}
	for _, affect := range fighter.MonsterAffects {
		if affect.Active && affect.Kind == kind {
			return true
		}
	}
	return false
}

// RollDie 擲一顆 `1..sides`。物品效果的持續時間要它（`(1d4 ＋ 4) × 10`），
// 而那條算式在原作是 handler 自己擲的，不在屬性表裡。
func (b *Battle) RollDie(sides int) int {
	if b == nil || b.rng == nil || sides <= 0 {
		return 0
	}
	return b.rng.Intn(sides) + 1
}
