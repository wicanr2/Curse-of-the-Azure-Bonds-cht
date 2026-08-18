package combat

import "fmt"

// 效果寫進**記錄**而不是全域的那一類（spec 1123）。
//
// ★ 有些效果不是「把某個修正加進暫存值」，而是直接改對象的記錄：
// 纏繞術把戰鬥狀態的移動率設成 0、妖火把角色的兩格護甲各加 2（封頂）。
// 那些同樣是規則，只是落點不同。
//
// ⚠ remake 只搬**看得懂對應**的那幾格。看不懂的照樣列進 `Unread`——
// 悄悄忽略會讓覆蓋報告把「沒搬」算成「沒有影響」。

// 原作記錄裡已經有對應的那幾格。
const (
	// combatStateMovement 是戰鬥狀態 `+06h`：移動率（spec 806）。
	combatStateMovement = 0x06
	// playerArmourClassPrimary 是角色 `+19Ah`，兩格護甲的第一格。
	playerArmourClassPrimary = 0x19A
	// playerArmourClassSecondary 是 `+19Bh`：背後攻擊用的第二個護甲欄位
	// （spec 1000 §七）。原作兩格各套一次同樣的修正，所以 remake 也要兩格
	// 都套——只套第一格會讓「正面與背後的差」隨效果漂移。
	playerArmourClassSecondary = 0x19B
)

// recordWriteIsMapped 回答「remake 搬得動這一格嗎」。搬不動的記錄寫入不能算
// 進判讀範圍——那會讓覆蓋報告把「抄了個位移但不知道是什麼」算成已完成。
func recordWriteIsMapped(record string, field int) bool {
	switch {
	case record == "combat_state" && field == combatStateMovement:
		return true
	case record == "player" && field == playerArmourClassPrimary:
		return true
	case record == "player" && field == playerArmourClassSecondary:
		return true
	}
	return false
}

// ApplyEffectRecordWrites 跑一次時機查詢，並把 remake 對得上的記錄寫入套上去。
// 回傳「有沒有改到東西」。
func (b *Battle) ApplyEffectRecordWrites(fighterID string, timing uint8) (bool, error) {
	if b == nil {
		return false, errNoPRNG
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", fighterID)
	}
	detail, err := CheckFX(fighter, timing, nil)
	if err != nil {
		return false, err
	}
	changed := false
	for _, write := range detail.Records {
		switch {
		case write.Record == "combat_state" && write.Field == combatStateMovement:
			fighter.MovementAllowance = applyRecordOp(write, fighter.MovementAllowance)
			changed = true
		case write.Record == "player" && write.Field == playerArmourClassPrimary:
			fighter.ArmorClass = applyRecordOp(write, fighter.ArmorClass)
			changed = true
		case write.Record == "player" && write.Field == playerArmourClassSecondary:
			// 沒有第二個 AC 的戰鬥員（隊員、合成記錄）不套——那一格是 0，
			// 套下去會憑空生出一個「比正面好打很多」的背後 AC。
			if fighter.ArmorClassFacingKnown {
				fighter.ArmorClassFacing = applyRecordOp(write, fighter.ArmorClassFacing)
				changed = true
			}
		}
	}
	if changed {
		b.fighters[fighterID] = fighter
	}
	return changed, nil
}

// applyRecordOp 套一個記錄寫入。
func applyRecordOp(write CheckFXRecordWrite, current int) int {
	switch write.Op {
	case "set":
		return write.Value
	case "add":
		return current + write.Value
	case "sub":
		return current - write.Value
	case "add_capped":
		if current >= write.CapThreshold {
			return write.Cap
		}
		return current + write.Value
	}
	return current
}
