package combat

import "fmt"

// 幾支形狀各自不同、但 handler 已經讀完的法術（spec 1124 §「逐支」）。
//
// ★ 原作本來就是一支法術一段程式碼——這一檔收的是那些**真的沒有共通形狀**的。
// 有共通形狀的（骰子、效果碼、範圍）都已經走資料驅動那幾條路。

// slayLivingFallbackDice 是豁免過關之後改吃的傷害：`2d8 ＋ 1`。
const (
	slayLivingFallbackCount = 2
	slayLivingFallbackSides = 8
	slayLivingFallbackBonus = 1
	// slayLivingElement 是豁免過關那一段改寫的傷害屬性旗標（`8`）。
	slayLivingElement uint8 = 0x08
)

// SlayLivingResult 是一次屠殺活物的結果。
type SlayLivingResult struct {
	CasterID string `json:"caster_id"`
	TargetID string `json:"target_id"`
	SaveRoll int    `json:"save_roll"`
	Saved    bool   `json:"saved"`
	// Slain 是沒過豁免：直接死亡。
	Slain  bool `json:"slain,omitempty"`
	Damage int  `json:"damage,omitempty"`
}

// CastSlayLiving 重現 `overlay-22 entry#89`：豁免（類別 4）沒過就直接死，
// 過了改吃 `2d8 ＋ 1`。
//
// ⚠ 把它接成「單純 2d8 傷害」會少掉主要效果——那個 `2d8` 是**豁免成功**那一支。
func (b *Battle) CastSlayLiving(casterID, targetID string) (SlayLivingResult, error) {
	if b == nil || b.rng == nil {
		return SlayLivingResult{}, errNoPRNG
	}
	if _, ok := b.fighters[casterID]; !ok {
		return SlayLivingResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SlayLivingResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SlayLivingResult{}, fmt.Errorf("battle is already over")
	}
	if target.HitPoints <= 0 {
		return SlayLivingResult{}, fmt.Errorf("target %q is already down", targetID)
	}
	// 豁免類別 4 ＝ 法術（原作 `entry#8(目標, 4, 0)`）。
	save, err := b.RollSavingThrow(target, 4, 0)
	if err != nil {
		return SlayLivingResult{}, err
	}
	result := SlayLivingResult{CasterID: casterID, TargetID: targetID,
		SaveRoll: save.Roll, Saved: save.Saved}
	if !save.Saved {
		result.Slain = true
		b.applyPositiveDamage(&target, target.HitPoints)
		b.fighters[targetID] = target
		b.updateStatus()
		return result, nil
	}
	damage := slayLivingFallbackBonus
	for roll := 0; roll < slayLivingFallbackCount; roll++ {
		damage += b.rng.Intn(slayLivingFallbackSides) + 1
	}
	adjusted, err := CheckFX(target, checkFXDamage, map[string]int{
		scratchDamage: damage, scratchDamageElement: int(slayLivingElement)})
	if err != nil {
		return SlayLivingResult{}, err
	}
	damage = adjusted.Applied[scratchDamage]
	if damage < 0 {
		damage = 0
	}
	result.Damage = b.applyPositiveDamage(&target, damage)
	b.fighters[targetID] = target
	b.updateStatus()
	return result, nil
}

// cureBlindnessEffect 是治療失明移除的效果碼（`21h`，原作 `entry#16` 查詢後移除）。
const cureBlindnessEffect uint8 = 0x21

// CureBlindness 把目標身上的致盲效果拿掉。回傳「有沒有真的拿掉」。
func (b *Battle) CureBlindness(targetID string) (bool, error) {
	if b == nil {
		return false, fmt.Errorf("battle is nil")
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return false, fmt.Errorf("unknown target %q", targetID)
	}
	kept := target.MonsterAffects[:0]
	removed := false
	for _, affect := range target.MonsterAffects {
		if affect.Kind == cureBlindnessEffect {
			removed = true
			continue
		}
		kept = append(kept, affect)
	}
	if !removed {
		return false, nil
	}
	target.MonsterAffects = append([]MonsterAffect(nil), kept...)
	b.fighters[targetID] = target
	return true, nil
}
