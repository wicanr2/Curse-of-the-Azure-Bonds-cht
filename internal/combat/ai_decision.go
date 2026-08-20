package combat

import (
	"fmt"

	engineaiscan "github.com/wicanr2/golden-box-remake-engine/combat/aiscan"
)

// AI 的「這回合做什麼」：用道具與施法共用一套門檻掃描（spec 835／836）。
//
// ★ 兩支決策的骨架一模一樣：門檻從 7 開始，每輪降 1，掃幾輪由 1d7 決定。
// 差別只有兩點——用道具**全掃物品鏈**、施法**每輪隨機抽 3 個**，以及各自的
// 前置閘門。所以骨架寫一次，差異用參數帶進去。
//
// 實務上的意思是：**AI 每回合隨機決定自己肯將就到什麼程度**——運氣好只掃一輪
// （只肯用分數 7 的），運氣差掃七輪（連分數 1 的都肯用）。

const (
	// aiInitialThreshold 是門檻的起點（spec 835／836）。
	aiInitialThreshold = 7
	// aiThresholdRoundsDie 是「掃幾輪」的骰子面數：1d7 ⇒ 最低要求落在 8 − 1d7。
	aiThresholdRoundsDie = 7
	// aiSpellPicksPerRound 是施法每輪隨機抽幾個候選（spec 836）。
	aiSpellPicksPerRound = 3
	// aiItemEffectRemapAbove／aiItemEffectRemapDelta 是 spec 835 的效果碼重映射：
	// `if e > 38h then e := e − 17h`。
	aiItemEffectRemapAbove = 0x38
	aiItemEffectRemapDelta = 0x17
)

// MoraleValue 解 `+0F7h` 的編碼（spec 758）：**bit 7 是「這個值有效」的旗標**，
// 低 7 位存的是原值的一半，讀出來要乘 2；超過 66h（102）一律當 0。
//
// ⚠ 不要把整個位元組當成士氣值。`0B3h` 看起來像 179，實際是「旗標已設、
// 原值 102」——那是資料裡最常見的一個值。
func MoraleValue(raw uint8) (int, bool) {
	if raw <= 0x7F {
		return 0, false
	}
	value := int(raw&0x7F) * 2
	if value > 0x66 {
		return 0, true
	}
	return value, true
}

// AIThresholdScan 是共用骨架。`perRound` 拿到當輪門檻，回傳選中的候選；
// 選到就停，否則門檻降 1 進下一輪。
//
// ⚠ **1d7 一定要擲**，即使候選清單是空的：原作在檢查清單之前就擲了
// （spec 835／836 都是），省掉它會讓後續的亂數序列整條偏掉。
func AIThresholdScan(roll func(sides int) int, perRound func(threshold int) (uint8, bool)) (uint8, bool) {
	return engineaiscan.Scan(engineaiscan.Rules{
		InitialThreshold: aiInitialThreshold,
		RoundsDie:        aiThresholdRoundsDie,
	}, roll, perRound)
}

// AIPriorityLookup 回傳一個候選的 AI 分數（法術屬性表的 `+0Dh`，spec 802／835）
// 與「查不查得到」。查不到的候選一律略過——**不要當成分數 0**，那會讓沒有資料的
// 法術在門檻降到 0 之後突然變成合法選擇。
type AIPriorityLookup func(id uint8) (int, bool)

// AIChooseSpell 重現 `overlay-09:0605h`：從記憶法術裡挑一個施放。
//
// ★ 會讓 remake「太強」的那個原作行為必須保留：**每輪只隨機抽 3 個**，
// 重複抽到同一格也算一次。記得越多法術，單一法術被看到的機率反而越低
// （每輪 3/n）；n > 21 時一定有法術從頭到尾沒被考慮。
//
// ⚠ **原作的 off-by-one 在這裡重現不了，也不該假裝重現。** spec 836 讀到 AI
// 收集清單時從 `+1Fh` 開始（`for k := 1 to 53h`），所以角色記錄裡 `+1Eh` 那一格
// 對 AI 是隱形的。但 `MonsterSpellIDs` 是**壓縮過**的清單——`internal/monster`
// 從 `+33h..+6Ah` 讀進來時就把 0 濾掉了，索引 0 不等於 `+1Eh`。
// 在壓縮清單上跳過第一個等於隨機丟掉一支真的法術，那是照抄表面、丟掉語意。
// 要重現它得先保留帶洞的原始槽位陣列，那是另一輪的事。
func (b *Battle) AIChooseSpell(fighterID string, priority AIPriorityLookup) (uint8, bool, error) {
	if b == nil || b.rng == nil {
		return 0, false, fmt.Errorf("battle PRNG is unavailable")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return 0, false, fmt.Errorf("unknown fighter %q", fighterID)
	}
	roll := func(sides int) int { return b.rng.Intn(sides) + 1 }
	// 士氣崩了就不施法（spec 836 的第一道閘門）。
	if morale, valid := MoraleValue(fighter.ControlMorale); valid && morale == 0 {
		return 0, false, nil
	}
	candidates := aiSpellCandidates(fighter)
	choice, found := AIThresholdScan(roll, func(threshold int) (uint8, bool) {
		if len(candidates) == 0 {
			return 0, false
		}
		for pick := 0; pick < aiSpellPicksPerRound; pick++ {
			spellID := candidates[roll(len(candidates))-1]
			score, known := priority(spellID)
			if known && score >= threshold {
				return spellID, true
			}
		}
		return 0, false
	})
	return choice, found, nil
}

// aiSpellCandidates 收集可用的法術。清單本身已經是壓縮過的（見 AIChooseSpell
// 的 ⚠），所以這裡不再跳過任何一格。
func aiSpellCandidates(fighter Fighter) []uint8 {
	candidates := make([]uint8, 0, len(fighter.MonsterSpellIDs))
	for _, spellID := range fighter.MonsterSpellIDs {
		if spellID > 0 {
			candidates = append(candidates, spellID)
		}
	}
	return candidates
}

// AIChooseItemEffect 重現 `overlay-09:04AAh`：從**裝備中**的物品裡挑一個發動。
//
// 與施法的差別是每一輪**整條物品鏈全掃**，第一個過門檻的就用。回傳的是效果碼，
// 已經套過原作的重映射（`e > 38h` 時減 17h）。
func (b *Battle) AIChooseItemEffect(fighterID string, priority AIPriorityLookup) (uint8, bool, error) {
	if b == nil || b.rng == nil {
		return 0, false, fmt.Errorf("battle PRNG is unavailable")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return 0, false, fmt.Errorf("unknown fighter %q", fighterID)
	}
	roll := func(sides int) int { return b.rng.Intn(sides) + 1 }
	choice, found := AIThresholdScan(roll, func(threshold int) (uint8, bool) {
		for _, raw := range fighter.ReadiedItemEffects {
			effect := AIRemapItemEffect(raw)
			score, known := priority(effect)
			if known && score >= threshold {
				return effect, true
			}
		}
		return 0, false
	})
	return choice, found, nil
}

// AIRemapItemEffect 是 spec 835 的 `if e > 38h then e := e − 17h`。
// 原作為什麼要這樣映射沒有讀出來，照抄。
func AIRemapItemEffect(effect uint8) uint8 {
	if effect > aiItemEffectRemapAbove {
		return effect - aiItemEffectRemapDelta
	}
	return effect
}
