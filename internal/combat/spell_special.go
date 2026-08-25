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
	// 同上：這一次查詢也可能改記錄（效果 `54h`）。
	applyRecordWritesTo(&target, adjusted)
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

// spellExtraEffectCodes 是「效果碼不在法術屬性表裡，而在 handler 裡」的那幾支。
//
// ★ 屬性表的 `+0Ah` 為 0，但 handler 自己呼叫 `overlay-23 entry#2` 掛上效果。
// 這幾個碼是從 handler 的立即數讀出來的，不是猜的（spec 1124）。
var spellExtraEffectCodes = map[uint8][]uint8{
	// 毒（法術 68）：`overlay-22 entry#82` 掛 `40h`。
	68: {0x40},
	// 疾病（法術 40）：`overlay-12` 的效果 `22h` handler 連掛 `2Bh` 與 `2Ch`。
	40: {0x2B, 0x2C},
}

// SpellExtraEffectCodes 回傳一支法術由 handler 掛上的效果碼。
func SpellExtraEffectCodes(spellID uint8) ([]uint8, bool) {
	codes, ok := spellExtraEffectCodes[spellID]
	return codes, ok
}

// ApplyEffectCodes 把一組效果碼掛到目標身上。
//
// ⚠ 原作在掛之前會跑一次 `CHECKFX(09h)`（`PUTEFFECT` 時機）讓免疫類的效果把它
// 擋掉。remake 這一側照跑：查詢回報有效果介入時就不掛。
func (b *Battle) ApplyEffectCodes(targetID string, codes []uint8, duration int) (int, error) {
	if b == nil {
		return 0, fmt.Errorf("battle is nil")
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return 0, fmt.Errorf("unknown target %q", targetID)
	}
	applied := 0
	for _, code := range codes {
		gate, err := CheckFX(target, checkFXPutEffect, nil)
		if err != nil {
			return 0, err
		}
		if len(gate.Contributed) > 0 {
			// 有效果介入這一次掛載——原作那一段會把待掛的碼清掉。
			continue
		}
		target.MonsterAffects = append(target.MonsterAffects, MonsterAffect{
			Kind: code, Value: uint16(duration), Duration: uint16(duration), Active: true})
		applied++
	}
	if applied > 0 {
		b.fighters[targetID] = target
	}
	return applied, nil
}

// removeCurseAffect 是移除詛咒的第一步：目標身上有效果 `24h` 就拿掉。
const removeCurseAffect uint8 = 0x24

// RemoveAffectKind 拿掉目標身上某一個效果碼，回傳有沒有拿掉。
//
// 原作的 `overlay-23 entry#16(目標, 碼)` 就是這件事：找到就移除並回 1，
// 找不到回 0。移除詛咒、次元門都靠它的回傳值決定下一步。
func (b *Battle) RemoveAffectKind(targetID string, kind uint8) (bool, error) {
	if b == nil {
		return false, fmt.Errorf("battle is nil")
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return false, fmt.Errorf("unknown target %q", targetID)
	}
	kept := make([]MonsterAffect, 0, len(target.MonsterAffects))
	removed := false
	for _, affect := range target.MonsterAffects {
		if !removed && affect.Kind == kind {
			removed = true
			continue
		}
		kept = append(kept, affect)
	}
	if !removed {
		return false, nil
	}
	target.MonsterAffects = kept
	b.fighters[targetID] = target
	return true, nil
}

// RemoveCurseAffect 是移除詛咒在戰鬥員身上的那一半：拿掉效果 `24h`。
//
// ★ 原作先試這一步，**拿掉了就結束**；沒拿掉才去翻裝備。物品那一半在
// `internal/game`，因為裝備掛在角色而不是戰鬥員身上。
func (b *Battle) RemoveCurseAffect(targetID string) (bool, error) {
	return b.RemoveAffectKind(targetID, removeCurseAffect)
}

// 火焰護盾的兩種形態（`overlay-22@4BECh`）。
//
// 原作跳一個選單讓玩家選 `H`（熱）或 `C`（冷），由電腦操作時改成擲 `1d10`
// 大於 5 選熱、否則選冷。兩種形態各掛兩個效果，第二個是共用的。
const (
	fireShieldHotAffect    uint8 = 0x32
	fireShieldColdAffect   uint8 = 0x36
	fireShieldSharedAffect uint8 = 0x8F
	// fireShieldHotThreshold 是那一顆 `1d10` 的門檻：**大於** 5 才是熱的。
	fireShieldHotThreshold = 5
)

// FireShieldAffects 回傳一種形態要掛的效果碼。
func FireShieldAffects(hot bool) []uint8 {
	if hot {
		return []uint8{fireShieldHotAffect, fireShieldSharedAffect}
	}
	return []uint8{fireShieldColdAffect, fireShieldSharedAffect}
}

// RollFireShieldHot 擲原作那一顆 `1d10` 決定形態。
func (b *Battle) RollFireShieldHot() (bool, error) {
	if b == nil || b.rng == nil {
		return false, errNoPRNG
	}
	return b.rng.Intn(10)+1 > fireShieldHotThreshold, nil
}

// CastFireShield 把一種形態的兩個效果掛到施法者身上。
func (b *Battle) CastFireShield(casterID string, hot bool, duration int) ([]uint8, error) {
	if b == nil {
		return nil, fmt.Errorf("battle is nil")
	}
	codes := FireShieldAffects(hot)
	if _, err := b.ApplyEffectCodes(casterID, codes, duration); err != nil {
		return nil, err
	}
	return codes, nil
}

// 次元門（`overlay-22@48E4h`）。
//
// 施法者身上有效果 `3Ah` 時，先把半徑 1 內、被自己「勾住」的戰鬥員身上的
// `90h` 與 `8Bh` 拿掉，再把自己移到指定格。沒有 `3Ah` 就只做移動。
const (
	dimensionDoorGateAffect uint8 = 0x3A
	dimensionDoorLinkRadius       = 1
)

var dimensionDoorLinkAffects = []uint8{0x90, 0x8B}

// CastDimensionDoor 把施法者傳送到指定格，並解開身邊的勾住標記。
func (b *Battle) CastDimensionDoor(casterID string, destination TilePoint) (int, error) {
	if b == nil {
		return 0, fmt.Errorf("battle is nil")
	}
	if b.status != StatusActive {
		return 0, fmt.Errorf("battle is already over")
	}
	caster, ok := b.fighters[casterID]
	if !ok {
		return 0, fmt.Errorf("unknown caster %q", casterID)
	}
	if caster.HitPoints <= 0 {
		return 0, fmt.Errorf("dead fighter cannot cast")
	}
	if !caster.HasCombatPosition {
		return 0, fmt.Errorf("caster %q has no combat position", casterID)
	}
	released := 0
	if fighterHasAnyAffect(caster, []uint8{dimensionDoorGateAffect}) {
		origin := TilePoint{X: caster.CombatX, Y: caster.CombatY}
		for _, id := range b.fighterOrder {
			if id == casterID {
				continue
			}
			neighbour := b.fighters[id]
			if neighbour.HitPoints <= 0 || !neighbour.HasCombatPosition ||
				!fighterFootprintWithinRadius(neighbour, origin, dimensionDoorLinkRadius) {
				continue
			}
			for _, kind := range dimensionDoorLinkAffects {
				removed, err := b.RemoveAffectKind(id, kind)
				if err != nil {
					return 0, err
				}
				if removed {
					released++
				}
			}
		}
	}
	caster = b.fighters[casterID]
	caster.CombatX = destination.X
	caster.CombatY = destination.Y
	b.fighters[casterID] = caster
	return released, nil
}

// 寒冰錐的半徑（`overlay-22@54EBh`）：`(施法者等級 ＋ 1) div 2`，最小 1。
//
// ★ 這支法術的半徑**不在**法術屬性表裡（表裡寫 0），是 handler 自己算的。
// 照表用 0 會讓它只打中心那一格。
func ConeOfColdRadius(casterLevel int) int {
	radius := (casterLevel + 1) / 2
	if radius < 1 {
		radius = 1
	}
	return radius
}
