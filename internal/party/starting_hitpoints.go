package party

import (
	"fmt"
	"math/rand"
)

// RollStartingHitPoints 計算 CoAB NEWCHAR 完成後、已依起始經驗提升到最終
// 職業等級的累計 HP。RollCreationHitPoints 對應「各職業槽剛寫成 1」的單次
// 建角 helper；原版 BOB.GUY 則證實同一名新角在存出來時已是戰士 5 級、5 HD，
// 因此兩個生命週期不能混成同一支函式。
func RollStartingHitPoints(tables HitDiceLookup, classLevels [8]uint8, classCombo, constitution int, seed int64) (CreationHitPoints, error) {
	if tables == nil {
		return CreationHitPoints{}, fmt.Errorf("hit dice lookup is required")
	}
	rng := rand.New(rand.NewSource(seed))
	roll := func(count, size int) int {
		total := 0
		for index := 0; index < count; index++ {
			total += rng.Intn(size) + 1
		}
		return total
	}

	classCount, rolled, bonus := 0, 0, 0
	for slot, raw := range classLevels {
		level := int(raw)
		if level <= 0 {
			continue
		}
		count, size, levelCap, ok := tables.HitDiceFor(slot)
		if !ok {
			return CreationHitPoints{}, fmt.Errorf("class slot %d has no hit dice", slot)
		}
		classCount++
		for currentLevel := 1; currentLevel <= level; currentLevel++ {
			if currentLevel >= levelCap {
				rolled += fixedHitPointsPastCap(slot)
				continue
			}
			dice := 1
			if currentLevel == 1 {
				dice = count
			}
			first, second := roll(dice, size), roll(dice, size)
			if second > first {
				first = second
			}
			rolled += first
			bonus += tables.ConstitutionHPBonus(constitution)
			bonus += tables.FighterConstitutionHPBonus(classCombo, constitution)
			if slot == rangerClassSlot && currentLevel == 1 {
				bonus *= 2
			}
		}
	}
	if classCount == 0 {
		return CreationHitPoints{}, fmt.Errorf("character has no class levels")
	}
	maxHP := rolled
	if bonus < 0 && maxHP <= classCount+(-bonus) {
		maxHP = 1
	} else {
		maxHP = (maxHP + bonus) / classCount
	}
	return CreationHitPoints{MaxHitPoints: maxHP, BaseMaxHitPoints: rolled / classCount}, nil
}
