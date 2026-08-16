package combat

import "fmt"

// EscapeAttempt 是一次「走出戰場邊界」的判定結果（spec 799／1112）。
//
// ★ 原作沒有 FLEE 指令。 Gold Box 的逃跑就是把角色走出戰鬥地圖的邊界，
// 那一步會觸發 `overlay-13:0D1Ch` 的判定。所以這一支的呼叫時機是**移動**，
// 不是一個獨立的選單指令。
type EscapeAttempt struct {
	FighterID string
	// Contested 是「有敵方戰鬥員貼著」。沒有的話原作連骰都不擲。
	Contested bool
	// OwnRate 是逃跑者的移動率（格）。
	OwnRate int
	// FastestOpponentRate 是**對面隊伍**中還能行動的最快移動率。
	// 原作用它跟 OwnRate 比，不是跟距離或先攻比。
	FastestOpponentRate int
	// TieRoll 只有在兩個速率相等時才有值（1d2，1 才逃得掉）。0 代表沒擲。
	TieRoll int
	Escaped bool
}

// AttemptEscape 重現原作的逃跑判定：
//
//	沒有敵方戰鬥員貼著                    → 逃掉
//	我的移動率 >  對面最快的移動率        → 逃掉
//	我的移動率 == 對面最快的移動率        → 1d2 ＝ 1 才逃掉（平手偏向失敗）
//	我的移動率 <  對面最快的移動率        → 逃不掉
//
// ⚠ 原作用 `overlay-24 entry#32` 取「貼著的戰鬥員」，第一個參數在逃跑這裡是
// `0FFh`、在機會攻擊那裡是 `1`，**兩者的差別沒有讀出來**（spec 799）。
// 這裡取「對面陣營、活著、能行動」的相鄰者——那是機會攻擊那條路已經確定的語意。
// 若日後讀出 `0FFh` 包含同陣營，改的是這個判斷，不是整條規則。
func (b *Battle) AttemptEscape(fighterID string) (EscapeAttempt, error) {
	if b == nil {
		return EscapeAttempt{}, fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return EscapeAttempt{}, fmt.Errorf("unknown fighter %q", fighterID)
	}
	if b.status != StatusActive {
		return EscapeAttempt{}, fmt.Errorf("battle is already over")
	}
	if fighter.HitPoints <= 0 {
		return EscapeAttempt{}, fmt.Errorf("dead fighter cannot escape")
	}
	if fighter.Escaped {
		return EscapeAttempt{}, fmt.Errorf("fighter %q already left the battlefield", fighterID)
	}
	attempt := EscapeAttempt{FighterID: fighterID, OwnRate: fighter.MovementAllowance}
	for _, other := range b.Fighters() {
		if !opposesInBattle(fighter, other) {
			continue
		}
		if adjacent(fighter, other) {
			attempt.Contested = true
		}
		if other.MovementAllowance > attempt.FastestOpponentRate {
			attempt.FastestOpponentRate = other.MovementAllowance
		}
	}
	if !attempt.Contested {
		attempt.Escaped = true
	} else {
		switch {
		case attempt.FastestOpponentRate < attempt.OwnRate:
			attempt.Escaped = true
		case attempt.FastestOpponentRate == attempt.OwnRate:
			if b.rng == nil {
				return EscapeAttempt{}, fmt.Errorf("battle PRNG is unavailable")
			}
			// 原作是 `overlay-23 entry#9(2, 1)` ＝ 1d2（spec 1112），
			// 回 1 才逃得掉。
			attempt.TieRoll = b.rng.Intn(2) + 1
			attempt.Escaped = attempt.TieRoll == 1
		}
	}
	if attempt.Escaped {
		fighter.Escaped = true
		fighter.HasCombatPosition = false
		b.fighters[fighterID] = fighter
		b.updateStatus()
	}
	return attempt, nil
}

// opposesInBattle 是「對面那一隊、活著、還能行動」。原作的條件是
// `overlay-24 entry#30(角色) = p^[197h]`（對面隊號）加上 `p^[196h] <> 0`
// （站著、能行動，spec 1010）。已經離場的當然也不算。
func opposesInBattle(fighter, other Fighter) bool {
	return other.Side != fighter.Side && other.HitPoints > 0 &&
		!other.Escaped && other.HasCombatPosition && !other.MonsterIsHeld()
}
