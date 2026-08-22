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
	// playerCurrentHitPoints 是角色 `+1A4h`：**目前 HP**。
	//
	// ★ 四份規格各自獨立指到同一格：spec 098（記錄版面）、spec 1079（屬性重算
	// 依最大 HP 的增減調它）、spec 1101（建角時 `+1A4h := +78h`）、
	// spec 1005（效果碼 78 拿它當 `STANDUP` 的引數）。
	//
	// ⚠ PC-98 是 `+1A5h`，差一格（spec 1005）。這裡是 DOS 的版面。
	playerCurrentHitPoints = 0x1A4
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
	case record == "player" && field == playerCurrentHitPoints:
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
	changed := applyRecordWritesTo(&fighter, detail)
	if changed {
		b.fighters[fighterID] = fighter
	}
	return changed, nil
}

// applyRecordWritesTo 把一次 `CHECKFX` 算出來的記錄寫入套到戰鬥員身上，
// 回報有沒有改到東西。
//
// ★ 抽出來是因為**傷害時機也要用**：`PUTDAMAGE` 那一路（`CHECKFX(06h)`）原本只讀
// `Applied[damage]`，把 `Records` 丟掉——於是效果 `54h`（被電回血 8 點）算出來了
// 卻沒有人套上去。算得出來但沒套，覆蓋報告還是會把它算成已接。
func applyRecordWritesTo(fighter *Fighter, detail CheckFXDetail) bool {
	changed := false
	for _, write := range detail.Records {
		switch {
		case write.Record == "combat_state" && write.Field == combatStateMovement:
			fighter.MovementAllowance = applyRecordOp(write, fighter.MovementAllowance)
			changed = true
		case write.Record == "player" && write.Field == playerArmourClassPrimary:
			// ⚠ 效果表裡的護甲修正是**儲存刻度**（`add_capped 2 cap 60` 的
			// 上限 60 就是儲存值的 AC 0），而 `Fighter.ArmorClass` 是畫面刻度。
			// 換過去套完再換回來，符號與上下限都不必自己翻。
			fighter.ArmorClass = DisplayArmorClass(applyRecordOp(write, StoredArmorClass(fighter.ArmorClass)))
			changed = true
		case write.Record == "player" && write.Field == playerCurrentHitPoints:
			// 效果 `54h`：傷害時機（`06h`）問到時，若傷害屬性帶電（`damage_element`
			// 的 bit 2），就把目前 HP 加 8——被電反而回血。
			//
			// ⚠ 上限要壓在最大 HP：原作那一支沒有自己夾，因為屬性重算與治療
			// 那幾條路各自會把它拉回來；remake 這一側只跑這一條，不夾就會累積
			// 出超過最大值的 HP。
			updated := applyRecordOp(write, fighter.HitPoints)
			if fighter.MaxHitPoints > 0 && updated > fighter.MaxHitPoints {
				updated = fighter.MaxHitPoints
			}
			if updated < 0 {
				updated = 0
			}
			if updated != fighter.HitPoints {
				fighter.HitPoints = updated
				changed = true
			}
		case write.Record == "player" && write.Field == playerArmourClassSecondary:
			// 沒有第二個 AC 的戰鬥員（隊員、合成記錄）不套——那一格是 0，
			// 套下去會憑空生出一個「比正面好打很多」的背後 AC。
			if fighter.ArmorClassFacingKnown {
				fighter.ArmorClassFacing = DisplayArmorClass(applyRecordOp(write, StoredArmorClass(fighter.ArmorClassFacing)))
				changed = true
			}
		}
	}
	return changed
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
