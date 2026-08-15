package party

import (
	"fmt"
	"math/rand"
)

// AbilityRollLookup 是建角擲點需要的兩組規則資料，實作由 game pack 提供。
type AbilityRollLookup interface {
	// AbilityRollSpec 回傳擲幾顆幾面、每次加多少、擲幾次取最大。
	AbilityRollSpec() (diceCount, diceSize, bonus, attempts int, ok bool)
	// RaceAbilityAdjustments 是六個屬性的種族調整，順序為力、智、睿、敏、體、魅。
	RaceAbilityAdjustments(raceID int) ([6]int, bool)
}

// RollCreationAbilities 重現原作建角的擲點（spec 1103 §二）：
// 每個屬性擲 attempts 次 `nds + bonus`，種族調整加在每一次上，取最大。
//
// ⚠ `bonus` 與 `attempts` 都不是 AD&D 桌上規則，是這個引擎自己加的；
// 底層的骰子確實是 3d6。回傳的是**未夾值**的六個數，
// 上下限與職業最低要求由呼叫端再套（spec 1093 §四）。
func RollCreationAbilities(tables AbilityRollLookup, raceID int, seed int64) ([6]int, error) {
	if tables == nil {
		return [6]int{}, fmt.Errorf("ability roll lookup is required")
	}
	diceCount, diceSize, bonus, attempts, ok := tables.AbilityRollSpec()
	if !ok {
		return [6]int{}, fmt.Errorf("game pack has no ability roll rules")
	}
	adjustments, ok := tables.RaceAbilityAdjustments(raceID)
	if !ok {
		return [6]int{}, fmt.Errorf("race %d has no ability adjustments", raceID)
	}
	rng := rand.New(rand.NewSource(seed))

	var values [6]int
	for index := range values {
		best := 0
		for attempt := 0; attempt < attempts; attempt++ {
			total := bonus
			for die := 0; die < diceCount; die++ {
				total += rng.Intn(diceSize) + 1
			}
			total += adjustments[index]
			if total > best {
				best = total
			}
		}
		values[index] = best
	}
	return values, nil
}
