package party

import (
	"fmt"
	"math/rand"
)

// HitDiceLookup 是建角 HP 需要的三組規則資料。實作由 game pack 提供，
// 這個套件只留機制（spec 850／869／1101）。
type HitDiceLookup interface {
	// HitDiceFor 回傳職業槽一級的骰數、面數，以及體質加值的等級上限
	// （DS:3EB9h，等級 >= levelCap 就不再擲骰也不再給體質加值）。
	HitDiceFor(slot int) (count, size, levelCap int, ok bool)
	// ConstitutionHPBonus 是一般職業的體質加值，可為負。
	ConstitutionHPBonus(constitution int) int
	// FighterConstitutionHPBonus 是戰士系的額外加值。判斷用的是**職業組合
	// 編號**而不是職業槽，所以多職組合拿不到（spec 869）。
	FighterConstitutionHPBonus(classCombo, constitution int) int
}

// CreationHitPoints 是建角算出來的兩個 HP 欄位。
type CreationHitPoints struct {
	// MaxHitPoints 是 +78h：擲出來的總和加體質加值後除以職業數。
	MaxHitPoints int
	// BaseMaxHitPoints 是 +12Ch：**不含**體質加值的總和除以職業數。
	// 屬性重算時 +78h 會先回到這一格（spec 1079）。
	BaseMaxHitPoints int
}

// RollCreationHitPoints 重現原作建角尾端的 HP 決定（spec 1101）：
// 八個職業槽各擲一次生命骰，加上逐槽的體質加值，再除以職業數。
//
// classCombo 是原作的 +75h，只用在戰士系的額外體質加值判斷。
func RollCreationHitPoints(tables HitDiceLookup, classLevels [8]uint8, classCombo, constitution int, seed int64) (CreationHitPoints, error) {
	if tables == nil {
		return CreationHitPoints{}, fmt.Errorf("hit dice lookup is required")
	}
	rng := rand.New(rand.NewSource(seed))
	roll := func(count, size int) int {
		total := 0
		for i := 0; i < count; i++ {
			total += rng.Intn(size) + 1
		}
		return total
	}

	classCount := 0
	rolled := 0
	bonus := 0
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

		if level >= levelCap {
			// 過了 HD 上限：原作改給固定值，而且用的是覆寫不是累加
			// （spec 850 標記過的原作行為）。建角時等級一律是 1，
			// 走不到這條路；保留是為了讓同一支也能算已升級的角色。
			rolled = fixedHitPointsPastCap(slot)
			continue
		}
		// 只有一級才用多顆骰——遊俠 2d8、武僧 2d4（spec 850）。
		dice := count
		if level > 1 {
			dice = 1
		}
		// ⚠ 擲兩次取大不是 AD&D 規則，是引擎自己加的（spec 850）。
		first, second := roll(dice, size), roll(dice, size)
		if second > first {
			first = second
		}
		rolled += first

		bonus += tables.ConstitutionHPBonus(constitution)
		bonus += tables.FighterConstitutionHPBonus(classCombo, constitution)
		if slot == rangerClassSlot && level == 1 {
			// ⚠ 乘的是累積到當下的整個總和，不只是遊俠這一份（spec 869）。
			bonus *= 2
		}
	}
	if classCount == 0 {
		// 原作在這裡會 idiv 除以零，呼叫端必須保證有職業（spec 850）。
		return CreationHitPoints{}, fmt.Errorf("character has no class levels")
	}

	maxHP := rolled
	if bonus < 0 && maxHP <= classCount+(-bonus) {
		maxHP = 1
	} else {
		maxHP = (maxHP + bonus) / classCount
	}
	return CreationHitPoints{
		MaxHitPoints:     maxHP,
		BaseMaxHitPoints: rolled / classCount,
	}, nil
}

// rangerClassSlot 是遊俠的職業槽（spec 1084 的順序）。
const rangerClassSlot = 4

// fixedHitPointsPastCap 是名望等級之後的固定值（spec 850）：
// 戰士／聖騎士 3、牧師／遊俠／盜賊 2、法師 1，德魯伊與武僧原作沒有分支。
func fixedHitPointsPastCap(slot int) int {
	switch slot {
	case 2, 3:
		return 3
	case 0, 4, 6:
		return 2
	case 5:
		return 1
	}
	return 0
}
