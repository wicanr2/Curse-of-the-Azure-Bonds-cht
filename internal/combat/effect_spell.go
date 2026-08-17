package combat

import "fmt"

// 資料驅動的效果類法術：原作不是一支法術一段程式碼。
//
// ★ 原作的施法收尾是同一個迴圈（spec 731 的 `sub_F06h`）：對每個目標擲豁免、
// 依法術屬性表的 `+0Ah` 套上效果、持續時間由 `+2h`／`+3h` 的係數算出來。
// 換句話說，**法術之間的差別幾乎全在資料裡**——remake 這一側需要的是一條路，
// 不是 67 段程式碼。
//
// ⚠ 這一條只負責「把效果記上去」。效果碼本身的語意（定身、隱形、加速…）
// 由 `Fighter` 那幾支 `MonsterAffect*` 判讀；記得上去不等於解讀得了。
// 覆蓋報告要把兩者分開講。

// EffectSpellImpact 是一個目標的結果。
type EffectSpellImpact struct {
	TargetID string `json:"target_id"`
	// SaveRoll 是 1d20；`Rolled` 為 false 代表這支法術不擲豁免（`+8h = 0`）。
	SaveRoll int  `json:"save_roll,omitempty"`
	Rolled   bool `json:"rolled,omitempty"`
	Saved    bool `json:"saved,omitempty"`
	// Applied 是「效果真的記上去了」。豁免成功而且該法術是「豁免就沒事」那一類
	// 時為 false。
	Applied bool `json:"applied"`
}

// EffectSpellResult 是一次效果類施法的結果。
type EffectSpellResult struct {
	CasterID   string              `json:"caster_id"`
	SpellID    uint8               `json:"spell_id"`
	EffectKind uint8               `json:"effect_kind"`
	Duration   int                 `json:"duration"`
	Impacts    []EffectSpellImpact `json:"impacts"`
}

// EffectSpellRequest 是施法所需的資料，全部來自法術屬性表（spec 1111）。
type EffectSpellRequest struct {
	SpellID uint8
	// EffectKind 是屬性表的 `+0Ah`。0 代表這支法術不套效果——呼叫端不該走這條路。
	EffectKind uint8
	// Duration 由 `SpellEntry.PrimaryDuration` 算好再傳進來：那條算式含兩個
	// 「補成 1」的修正與 byte 回捲，屬於資料層，不該在戰鬥層重寫。
	Duration int
	// SaveKind 是 `+8h`：0 就完全不擲豁免。
	SaveKind int
	// SaveCategory 是 `+9h`，直接當 `SavingThrows` 的索引（0..4）。
	SaveCategory int
}

// saveNegatesKind 是 `+8h = 1` 那一組。spec 1111 只把它記成觀察（那一組全是
// 「豁免就沒事」型的法術），不是讀過 `entry#20` 的定論——所以判斷寫在這裡一處，
// 日後讀出真正的語意時只改這一行。
const saveNegatesKind = 1

// CastEffectSpell 對每個目標擲豁免並套上效果。
func (b *Battle) CastEffectSpell(casterID string, targetIDs []string, request EffectSpellRequest) (EffectSpellResult, error) {
	if b == nil || b.rng == nil {
		return EffectSpellResult{}, fmt.Errorf("battle PRNG is unavailable")
	}
	if b.status != StatusActive {
		return EffectSpellResult{}, fmt.Errorf("battle is already over")
	}
	if _, ok := b.fighters[casterID]; !ok {
		return EffectSpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if request.EffectKind == 0 {
		return EffectSpellResult{}, fmt.Errorf("spell %d has no effect to apply", request.SpellID)
	}
	if request.SaveCategory < 0 || request.SaveCategory > 4 {
		return EffectSpellResult{}, fmt.Errorf("saving throw category %d is outside 0..4", request.SaveCategory)
	}
	result := EffectSpellResult{
		CasterID: casterID, SpellID: request.SpellID,
		EffectKind: request.EffectKind, Duration: request.Duration,
	}
	for _, targetID := range targetIDs {
		target, ok := b.fighters[targetID]
		if !ok {
			return EffectSpellResult{}, fmt.Errorf("unknown target %q", targetID)
		}
		if target.HitPoints <= 0 || target.Escaped {
			continue
		}
		impact := EffectSpellImpact{TargetID: targetID}
		if request.SaveKind != 0 {
			impact.Rolled = true
			impact.SaveRoll = b.rng.Intn(20) + 1
			impact.Saved = b.savingThrowSucceeds(target, request.SaveCategory, impact.SaveRoll)
		}
		if impact.Saved && request.SaveKind == saveNegatesKind {
			result.Impacts = append(result.Impacts, impact)
			continue
		}
		// Duration 與 Value 兩格都寫：到期迴圈先看 Duration，為 0 時退回 Value
		// （`AdvanceMonsterAffects`）。**Strength 不是持續時間**——`0FFh` 是
		// 「永久」的標記（spec 441／758），拿它裝時間會讓效果變成不會過期。
		target.MonsterAffects = append(target.MonsterAffects, MonsterAffect{
			Kind:   request.EffectKind,
			Value:  uint16(request.Duration),
			Duration: uint16(request.Duration),
			Active: true,
		})
		b.fighters[targetID] = target
		impact.Applied = true
		result.Impacts = append(result.Impacts, impact)
	}
	b.updateStatus()
	return result, nil
}

// savingThrowSucceeds 用角色記錄的第 category 個豁免門檻。門檻缺席時**算失敗**
// ——不要當成 0（那會變成「一定成功」），怪物記錄沒有豁免值時該是弱點不是免疫。
func (b *Battle) savingThrowSucceeds(target Fighter, category, roll int) bool {
	if category >= len(target.SavingThrows) {
		return false
	}
	threshold := int(target.SavingThrows[category])
	if threshold == 0 {
		return false
	}
	return roll+target.SavingThrowBonus >= threshold
}
