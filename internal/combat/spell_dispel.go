package combat

import "fmt"

// 解除魔法的對抗（spec 1125，handler `overlay-22@317Eh`）。
//
// 原作對半徑內每一個目標走一遍效果串列，每一個效果各擲一次 `1d100`：
//
//	施法者等級 > 效果等級 ⇒ 機率 ＝ (差) × 5 ＋ 50
//	施法者等級 < 效果等級 ⇒ 機率 ＝ 50 − (差) × 2
//	相等                 ⇒ 機率 ＝ 50
//
// ★ 兩個方向的係數**不一樣**（高出去每級 ＋5，低下去每級 −2）。寫成對稱的
// ±5 會讓低階施法者解得太少、高階施法者解得太多，而且兩邊都不會有人發現。

const (
	// dispelBaseChance 是等級相同時的機率。
	dispelBaseChance = 50
	// dispelPerLevelAbove／Below 是兩個方向各自的係數。
	dispelPerLevelAbove = 5
	dispelPerLevelBelow = 2
)

// DispelChance 回傳「解得掉」的百分比機率。
func DispelChance(casterLevel, effectLevel int) int {
	switch {
	case casterLevel > effectLevel:
		return (casterLevel-effectLevel)*dispelPerLevelAbove + dispelBaseChance
	case casterLevel < effectLevel:
		return dispelBaseChance - (effectLevel-casterLevel)*dispelPerLevelBelow
	default:
		return dispelBaseChance
	}
}

// DispelImpact 是一個目標身上被解掉的效果。
type DispelImpact struct {
	TargetID string `json:"target_id"`
	Kind     uint8  `json:"kind"`
	Chance   int    `json:"chance"`
	Roll     int    `json:"roll"`
	Removed  bool   `json:"removed"`
}

// DispelResult 是一次解除魔法的結果。
type DispelResult struct {
	CasterID string         `json:"caster_id"`
	Center   TilePoint      `json:"center"`
	Impacts  []DispelImpact `json:"impacts"`
}

// Removed 是這一次真的解掉幾個效果。
func (r DispelResult) Removed() int {
	count := 0
	for _, impact := range r.Impacts {
		if impact.Removed {
			count++
		}
	}
	return count
}

// CastDispelMagic 對中心半徑內的每一個戰鬥員走一遍效果串列。
//
// ⚠ 原作跳過 `EFFECTREC +3` 為 `0FFh` 的效果——那是「解不掉」的標記。remake
// 這一側用 `Innate`（`MON*SPC` 帶進來的天生能力）當同一條界線：怪物的天生
// 效果不是誰施的法，沒有等級可以對抗。
func (b *Battle) CastDispelMagic(casterID string, center TilePoint,
	radius, casterLevel int) (DispelResult, error) {
	if b == nil || b.rng == nil {
		return DispelResult{}, errNoPRNG
	}
	if b.status != StatusActive {
		return DispelResult{}, fmt.Errorf("battle is already over")
	}
	caster, ok := b.fighters[casterID]
	if !ok {
		return DispelResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if caster.HitPoints <= 0 {
		return DispelResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if radius < 0 {
		return DispelResult{}, fmt.Errorf("dispel radius %d is negative", radius)
	}
	if casterLevel < 1 {
		return DispelResult{}, fmt.Errorf("caster level must be positive")
	}
	result := DispelResult{CasterID: casterID, Center: center}
	for _, id := range b.fighterOrder {
		target := b.fighters[id]
		if target.HitPoints <= 0 || target.Escaped || !target.HasCombatPosition ||
			!fighterFootprintWithinRadius(target, center, radius) {
			continue
		}
		kept := make([]MonsterAffect, 0, len(target.MonsterAffects))
		changed := false
		for _, affect := range target.MonsterAffects {
			if affect.Innate {
				kept = append(kept, affect)
				continue
			}
			chance := DispelChance(casterLevel, int(affect.Raw4))
			roll := b.rng.Intn(100) + 1
			impact := DispelImpact{TargetID: id, Kind: affect.Kind,
				Chance: chance, Roll: roll, Removed: roll <= chance}
			result.Impacts = append(result.Impacts, impact)
			if impact.Removed {
				changed = true
				continue
			}
			kept = append(kept, affect)
		}
		if changed {
			target.MonsterAffects = kept
			b.fighters[id] = target
		}
	}
	b.updateStatus()
	return result, nil
}

// FighterWithinRadius 是「這個戰鬥員的佔格有沒有落在中心半徑內」的對外入口。
// 範圍類法術的目標清單在 `internal/game` 那一側組，需要同一套幾何。
func FighterWithinRadius(fighter Fighter, center TilePoint, radius int) bool {
	return fighterFootprintWithinRadius(fighter, center, radius)
}
