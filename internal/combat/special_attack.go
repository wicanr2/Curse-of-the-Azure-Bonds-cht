package combat

// 怪物特殊攻擊（spec 1202；handler 語意 spec 720／722／723／725／735）。
//
// 原作把這五支掛在 CALLEFFECT 效果分派表（`INITSPELLS` 尾段），AI 在
// type-14 特殊行動階段以效果碼呼叫。這裡照各 handler 的控制流轉錄：
// 亂數消耗的順序（先擲 1d100 再判距離）、次數欄、豁免類別與傷害參數
// 都按 asm；**選目標與範圍形狀是近似**（模式碼 `41h`／`3Dh`／`24h` 與
// `sub_175Bh` 未讀，見 pack 宣告的 ⚠ 註記），由呼叫端提供目標。
//
// 傷害走與法術同一條路：`CHECKFX(06h)` 帶傷害屬性旗標（抗火因此對噴火
// 生效），豁免成功減半（entry#20 本體未讀，沿用 spec 1124 傷害法術的
// 既有慣例）。

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
)

// SpecialAttackForm 是五支 handler 的形狀。
type SpecialAttackForm string

const (
	// SpecialAttackSpit 是 `5974h` 吐酸單體（效果 56h）。
	SpecialAttackSpit SpecialAttackForm = "spit_single"
	// SpecialAttackBreathTouch 是 `5E11h` 龍息火（效果 83h）：相鄰、五成不發動。
	SpecialAttackBreathTouch SpecialAttackForm = "breath_touch"
	// SpecialAttackBreathAreaSameSide 是 `5AA8h` 區域吐酸（效果 5Ah）：
	// 範圍裡混進同伴就整次取消（同類檢查的近似，`entry#30`／`+197h` 未讀）。
	SpecialAttackBreathAreaSameSide SpecialAttackForm = "breath_area_same_side"
	// SpecialAttackBreathArea 是 `5C86h` 噴火（效果 80h）：沒有同類檢查，
	// 火會燒到自己人。
	SpecialAttackBreathArea SpecialAttackForm = "breath_area"
	// SpecialAttackGaze 是 `6022h` 凝視（效果 7Eh）：豁免成功完全沒事，
	// 失敗掛麻痺。
	SpecialAttackGaze SpecialAttackForm = "gaze"
)

// SpecialAttackRule 是 pack 宣告的一筆特殊攻擊。數字全部出自 asm 立即數。
type SpecialAttackRule struct {
	ID         string
	EffectKind uint8
	Form       SpecialAttackForm
	// MissChance：1d100 大於它就不命中（吐酸 30、龍息火 50；0＝必中）。
	// 龍息火的不命中是靜默返回，吐酸有 miss 訊息——由 Form 區分。
	MissChance int
	// MaxDistance 是 `entry#33`（半格值 div 2）的放棄門檻：距離 >= 它就不打
	//（吐酸 7、龍息火 2；0＝不判距離）。
	MaxDistance int
	// Uses：非 0 時每場上限（區域吐酸／噴火都是 3），最後成功那一步才扣。
	Uses int
	// DamageMask 是 `DS:6F95h` 的旗標（龍息火 1、噴火 21h、吐酸 30h）。
	DamageMask uint8
	// ConstantDamage 非 0 時傷害是常數（龍息火 7）；0＝攻擊者的 HP 上限。
	ConstantDamage int
	// SaveCategory 是 `entry#8` 的類別（傷害類 3、凝視 1）。
	SaveCategory int
	// TargetRange 是呼叫端選目標的近似射程（模式碼未讀）。
	TargetRange int
	// AreaRadius 是區域形狀的近似半徑（`sub_175Bh` 未讀）。
	AreaRadius int
	// ParalysisEffect／ParalysisDuration 是凝視失敗時掛的效果
	//（`entry#11(p, 34h, 3Ch, 0FFh, 0)`）。
	ParalysisEffect   uint8
	ParalysisDuration int
	// Message 家族是 locale 的 stable ID（pack 宣告）。
	Message          string
	MissMessage      string
	ParalyzedMessage string
}

// SpecialAttackImpact 是一名目標的結果。
type SpecialAttackImpact struct {
	TargetID string
	Saved    bool
	Damage   int
	TargetHP int
	// Paralyzed 只有凝視用：豁免失敗掛上了麻痺。
	Paralyzed bool
}

// SpecialAttackResult 是一次發動的結果。
type SpecialAttackResult struct {
	// Refrained：亂數或距離讓這次靜默放棄（龍息火的 1d100 > 50、距離太遠、
	// 次數用完、區域同類檢查取消）。呼叫端不顯示訊息、不消耗行動以外的東西。
	Refrained bool
	// Missed：吐酸擲失（有專屬訊息，回合照樣消耗）。
	Missed  bool
	Impacts []SpecialAttackImpact
}

// MonsterSpecialAttackRules 依身上有效的效果碼取出可用的特殊攻擊宣告。
func (f Fighter) MonsterSpecialAttackRules() []SpecialAttackRule {
	if len(f.SpecialAttackRules) == 0 {
		return nil
	}
	active := map[uint8]bool{}
	for _, affect := range f.MonsterAffects {
		if affect.operational() {
			active[affect.Kind] = true
		}
	}
	resolved := make([]SpecialAttackRule, 0, 1)
	for _, rule := range f.SpecialAttackRules {
		if active[rule.EffectKind] {
			resolved = append(resolved, rule)
		}
	}
	return resolved
}

// SetMonsterSpecialAttacks 把 pack 宣告掛到每一名戰鬥員（同 SetMonsterSpellRules）。
func (b *Battle) SetMonsterSpecialAttacks(rules []SpecialAttackRule) {
	if b == nil {
		return
	}
	for id, fighter := range b.fighters {
		fighter.SpecialAttackRules = append([]SpecialAttackRule(nil), rules...)
		b.fighters[id] = fighter
	}
}

// specialAttackDistance 近似 `overlay-24 entry#33`：SCAN 加權距離（直 2 斜 3）
// 除以 2。原作讀的是共用緩衝裡的半格值 div 2（spec 720）。
func specialAttackDistance(attacker, target Fighter) int {
	return weightedTileDistance(
		TilePoint{X: attacker.CombatX, Y: attacker.CombatY},
		TilePoint{X: target.CombatX, Y: target.CombatY}) / 2
}

// specialAttackDamage 套一名目標：豁免（類別照規則）→ 成功減半 →
// `CHECKFX(06h)` 帶旗標 → 扣血。
func (b *Battle) specialAttackDamage(attacker Fighter, targetID string,
	rule SpecialAttackRule) (SpecialAttackImpact, error) {
	target, ok := b.fighters[targetID]
	if !ok {
		return SpecialAttackImpact{}, fmt.Errorf("unknown target %q", targetID)
	}
	impact := SpecialAttackImpact{TargetID: targetID}
	save, err := b.RollSavingThrow(target, rule.SaveCategory, 0)
	if err != nil {
		return SpecialAttackImpact{}, err
	}
	impact.Saved = save.Saved
	damage := rule.ConstantDamage
	if damage == 0 {
		damage = attacker.MaxHitPoints
	}
	if impact.Saved {
		damage /= 2
	}
	adjusted, err := CheckFX(target, checkFXDamage, map[string]int{
		scratchDamage: damage, scratchDamageElement: int(rule.DamageMask)})
	if err != nil {
		return SpecialAttackImpact{}, err
	}
	applyRecordWritesTo(&target, adjusted)
	damage = adjusted.Applied[scratchDamage]
	if damage < 0 {
		damage = 0
	}
	impact.Damage = b.applyPositiveDamage(&target, damage)
	impact.TargetHP = target.HitPoints
	b.fighters[targetID] = target
	// 特殊攻擊與一般近戰／法術一樣可能放倒最後一名隊員。少了這一步會留下
	// HP=0 但 StatusActive 的不可能狀態，前端已回地城、存檔卻仍被 active battle
	// 擋住（spec 1230）。
	b.updateStatus()
	return impact, nil
}

// specialAttackUses 對應 `arg_2^[3]`：`DS:758Dh = 0`（回合計數還沒動）時
// 重設成宣告值。回傳「還有沒有次數」。
func (b *Battle) specialAttackUses(attacker *Fighter, rule SpecialAttackRule) bool {
	if rule.Uses == 0 {
		return true
	}
	if !attacker.SpecialAttackUsesSet {
		attacker.SpecialAttackUses = rule.Uses
		attacker.SpecialAttackUsesSet = true
	}
	return attacker.SpecialAttackUses > 0
}

// SpecialAttackSingle 轉錄兩支單體 handler（吐酸 `5974h`／龍息火 `5E11h`）。
//
// 順序照 asm：選完目標（呼叫端）→ 擲 1d100 → 距離門檻 → 命中分歧。
// 吐酸擲失回 Missed（有訊息）；龍息火擲失回 Refrained（靜默）。
func (b *Battle) SpecialAttackSingle(attackerID, targetID string,
	rule SpecialAttackRule) (SpecialAttackResult, error) {
	if b == nil || b.rng == nil {
		return SpecialAttackResult{}, errNoPRNG
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return SpecialAttackResult{}, fmt.Errorf("unknown attacker %q", attackerID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpecialAttackResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	roll := b.rng.Intn(100) + 1
	if rule.MaxDistance > 0 && specialAttackDistance(attacker, target) >= rule.MaxDistance {
		return SpecialAttackResult{Refrained: true}, nil
	}
	if rule.MissChance > 0 && roll > rule.MissChance {
		if rule.Form == SpecialAttackSpit {
			return SpecialAttackResult{Missed: true}, nil
		}
		return SpecialAttackResult{Refrained: true}, nil
	}
	impact, err := b.specialAttackDamage(attacker, targetID, rule)
	if err != nil {
		return SpecialAttackResult{}, err
	}
	b.updateStatus()
	return SpecialAttackResult{Impacts: []SpecialAttackImpact{impact}}, nil
}

// SpecialAttackAreaBreath 轉錄兩支區域 handler（區域吐酸 `5AA8h`／噴火 `5C86h`）。
//
// 傷害對範圍內每一個站著的戰鬥員各套一次（噴火沒有陣營檢查——會燒到
// 攻擊者的同伴；區域吐酸範圍裡混進攻擊者同側的單位就整次取消）。
// 次數最後才扣：中途任何放棄都不消耗。
func (b *Battle) SpecialAttackAreaBreath(attackerID string, center TilePoint,
	rule SpecialAttackRule) (SpecialAttackResult, error) {
	if b == nil || b.rng == nil {
		return SpecialAttackResult{}, errNoPRNG
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return SpecialAttackResult{}, fmt.Errorf("unknown attacker %q", attackerID)
	}
	if !b.specialAttackUses(&attacker, rule) {
		b.fighters[attackerID] = attacker
		return SpecialAttackResult{Refrained: true}, nil
	}
	b.fighters[attackerID] = attacker
	targets := make([]string, 0, len(b.fighterOrder))
	for _, id := range b.fighterOrder {
		candidate := b.fighters[id]
		if id == attackerID || candidate.HitPoints <= 0 || candidate.Escaped ||
			!candidate.HasCombatPosition {
			continue
		}
		if !fighterFootprintWithinRadius(candidate, center, rule.AreaRadius) {
			continue
		}
		if rule.Form == SpecialAttackBreathAreaSameSide && candidate.Side == attacker.Side {
			// 同類檢查（近似）：範圍裡有自己人就整次取消，次數不扣。
			return SpecialAttackResult{Refrained: true}, nil
		}
		targets = append(targets, id)
	}
	if len(targets) == 0 {
		return SpecialAttackResult{Refrained: true}, nil
	}
	result := SpecialAttackResult{}
	for _, id := range targets {
		impact, err := b.specialAttackDamage(attacker, id, rule)
		if err != nil {
			return SpecialAttackResult{}, err
		}
		result.Impacts = append(result.Impacts, impact)
	}
	if rule.Uses > 0 {
		attacker = b.fighters[attackerID]
		attacker.SpecialAttackUses--
		b.fighters[attackerID] = attacker
	}
	b.updateStatus()
	return result, nil
}

// SpecialAttackGazeAt 轉錄凝視 `6022h`：豁免（類別 1）成功完全沒事；
// 失敗掛麻痺效果（`34h`，時間 `3Ch`）並回 Paralyzed。沒有傷害。
func (b *Battle) SpecialAttackGazeAt(attackerID, targetID string,
	rule SpecialAttackRule) (SpecialAttackResult, error) {
	if b == nil || b.rng == nil {
		return SpecialAttackResult{}, errNoPRNG
	}
	if _, ok := b.fighters[attackerID]; !ok {
		return SpecialAttackResult{}, fmt.Errorf("unknown attacker %q", attackerID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpecialAttackResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	impact := SpecialAttackImpact{TargetID: targetID, TargetHP: target.HitPoints}
	save, err := b.RollSavingThrow(target, rule.SaveCategory, 0)
	if err != nil {
		return SpecialAttackResult{}, err
	}
	impact.Saved = save.Saved
	if !impact.Saved {
		if _, err := b.ApplyEffectCodes(targetID,
			[]uint8{rule.ParalysisEffect}, rule.ParalysisDuration); err != nil {
			return SpecialAttackResult{}, err
		}
		impact.Paralyzed = true
	}
	b.updateStatus()
	return SpecialAttackResult{Impacts: []SpecialAttackImpact{impact}}, nil
}

// SpecialAttackRulesFromPack 把 gamepack 的宣告轉成戰鬥層規則並驗證形狀。
func SpecialAttackRulesFromPack() ([]SpecialAttackRule, error) {
	table, err := gamepack.SpecialAttacks()
	if err != nil {
		return nil, err
	}
	forms := map[string]SpecialAttackForm{
		string(SpecialAttackSpit):               SpecialAttackSpit,
		string(SpecialAttackBreathTouch):        SpecialAttackBreathTouch,
		string(SpecialAttackBreathAreaSameSide): SpecialAttackBreathAreaSameSide,
		string(SpecialAttackBreathArea):         SpecialAttackBreathArea,
		string(SpecialAttackGaze):               SpecialAttackGaze,
	}
	rules := make([]SpecialAttackRule, 0, len(table.Attacks))
	for _, attack := range table.Attacks {
		form, ok := forms[attack.Form]
		if !ok {
			return nil, fmt.Errorf("special attack %q has unknown form %q", attack.ID, attack.Form)
		}
		if form == SpecialAttackGaze && (attack.ParalysisEffect == 0 || attack.ParalysisDuration <= 0) {
			return nil, fmt.Errorf("special attack %q gaze needs a paralysis effect and duration", attack.ID)
		}
		if form != SpecialAttackGaze && attack.DamageMask == 0 {
			return nil, fmt.Errorf("special attack %q needs a damage mask", attack.ID)
		}
		rules = append(rules, SpecialAttackRule{
			ID: attack.ID, EffectKind: attack.EffectKind, Form: form,
			MissChance: attack.MissChance, MaxDistance: attack.MaxDistance,
			Uses: attack.Uses, DamageMask: attack.DamageMask,
			ConstantDamage: attack.ConstantDamage, SaveCategory: attack.SaveCategory,
			TargetRange: attack.TargetRange, AreaRadius: attack.AreaRadius,
			ParalysisEffect: attack.ParalysisEffect, ParalysisDuration: attack.ParalysisDuration,
			Message: attack.Message, MissMessage: attack.MissMessage,
			ParalyzedMessage: attack.ParalyzedMessage,
		})
	}
	return rules, nil
}

// NearestSpecialAttackTarget 是選目標的近似（模式碼 `41h`／`3Dh`／`24h` 未讀）：
// 射程內最近的對方戰鬥員，距離同用 `entry#33` 的近似，平手取行動序靠前的。
// 不消耗亂數——原作的選目標內部未讀，這裡不猜它擲不擲骰。
func (b *Battle) NearestSpecialAttackTarget(attackerID string, side Side,
	rangeCells int) (Fighter, bool) {
	if b == nil {
		return Fighter{}, false
	}
	attacker, ok := b.fighters[attackerID]
	if !ok || !attacker.HasCombatPosition {
		return Fighter{}, false
	}
	best, bestDistance, found := Fighter{}, 0, false
	for _, id := range b.fighterOrder {
		candidate := b.fighters[id]
		if candidate.Side != side || candidate.HitPoints <= 0 || candidate.Escaped ||
			!candidate.HasCombatPosition {
			continue
		}
		distance := specialAttackDistance(attacker, candidate)
		if rangeCells > 0 && distance > rangeCells {
			continue
		}
		if !found || distance < bestDistance {
			best, bestDistance, found = candidate, distance, true
		}
	}
	return best, found
}
