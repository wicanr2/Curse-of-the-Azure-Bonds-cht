package combat

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"

// 骰子含施法者等級的那幾支法術（spec 1124）。
//
// 傷害表是機械抽出來的，只認得「骰數、面數、加值全是立即數」的那一種。
// 剩下的骰數是算出來的，抽取只能標 `computed`——那不是缺口，是這幾支本來
// 就要人去讀 handler。讀完的算式寫在這裡，每一支附 handler 位址。
//
// ★ 不要用表裡 `computed` 那幾列的殘值。腳本會撿到附近的立即數，
// 拿它當骰數會得到「看起來合理但錯」的傷害——寒冰錐會變成 0d4、
// 電擊之握會少掉整個施法者等級。

// SpellDamageFormula 是一支法術這一次要擲的骰子。
type SpellDamageFormula struct {
	Count int
	Sides int
	Bonus int
}

// Total 是這個算式的最大值，供上層做健全性檢查。
func (f SpellDamageFormula) Total() int { return f.Count*f.Sides + f.Bonus }

// casterLevelDamage 是逐支讀完 handler 的算式表。
//
// | 法術 | handler | 算式 |
// |---|---|---|
// | 9 燃燒之手 | `ov22@1E7Ch` | 不擲骰，傷害 ＝ 施法者等級 |
// | 20 電擊之握 | `ov22@2260h` | `1d8 ＋ 施法者等級` |
// | 92 寒冰錐 | `ov22@54EBh` | `等級d4 ＋ 施法者等級` |
//
// 三支都是「先算出等級，再和骰子相加」，所以只有等級進得了算式，
// 沒有別的變數。等級由 `<overlay-24 entry#44>(法術編號)` 算出來
// （`sub_F06` 自己在等級參數為 0 時也呼叫同一支）。
var casterLevelDamage = map[uint8]func(level int) SpellDamageFormula{
	9: func(level int) SpellDamageFormula {
		return SpellDamageFormula{Bonus: level}
	},
	20: func(level int) SpellDamageFormula {
		return SpellDamageFormula{Count: 1, Sides: 8, Bonus: level}
	},
	92: func(level int) SpellDamageFormula {
		return SpellDamageFormula{Count: level, Sides: 4, Bonus: level}
	},
}

// SpellDamageRoll 回傳一支法術這一次的骰子：先查逐支讀完的算式，再退回
// 機械抽出來的表。兩邊都沒有就回 false——那代表這支的傷害還沒解讀，
// 不可以拿殘值湊一個數字出來。
func SpellDamageRoll(spellID uint8, casterLevel int) (SpellDamageFormula, bool) {
	if formula, found := casterLevelDamage[spellID]; found {
		if casterLevel < 1 {
			return SpellDamageFormula{}, false
		}
		return formula(casterLevel), true
	}
	table, err := gamepack.SpellDamage()
	if err != nil {
		return SpellDamageFormula{}, false
	}
	count, sides, bonus, ok := table.Dice(spellID)
	if !ok {
		return SpellDamageFormula{}, false
	}
	return SpellDamageFormula{Count: count, Sides: sides, Bonus: bonus}, true
}

// SpellDamageElement 回傳推進 `sub_F06` 的傷害屬性旗標（bit 0 火、bit 1 冷、
// bit 2 電），抗性效果就是看它決定要不要介入。
func SpellDamageElement(spellID uint8) uint8 {
	table, err := gamepack.SpellDamage()
	if err != nil {
		return 0
	}
	return table.Element(spellID)
}
