package combat

import "fmt"

// 士氣崩潰的四種結果（spec 831）。
//
// ★ 為什麼不是「士氣崩了就逃走」。 原作擲一次 1d100 分四段：**只有 10% 真的
// 逃走**，50% 混亂、20% 狂暴、20% 暴怒。四段加起來剛好 100，沒有落空的骰值。
// 只做「逃走」會讓怪物在士氣崩潰之後的行為少掉九成。

// MoraleOutcome 是四段的其中一段。
type MoraleOutcome int

const (
	// MoraleRunsAway 是 1..10：真的離場。
	MoraleRunsAway MoraleOutcome = iota + 1
	// MoraleConfused 是 11..60：混亂。
	MoraleConfused
	// MoraleBerserk 是 61..80：狂暴。
	MoraleBerserk
	// MoraleEnraged 是 81..100：暴怒。
	MoraleEnraged
)

// MessageID 是這一段對應的訊息鍵。
func (o MoraleOutcome) MessageID() string {
	switch o {
	case MoraleRunsAway:
		return "combat_morale_runs_away"
	case MoraleConfused:
		return "combat_morale_confused"
	case MoraleBerserk:
		return "combat_morale_berserk"
	case MoraleEnraged:
		return "combat_morale_enraged"
	}
	return ""
}

// moraleRestoredByte 是「逃走」那一段補回去的士氣（`0B3h` ＝ 旗標設起、值 102，
// spec 758／798）。**只在最高位沒設時才寫**——原作是 `if +0F7h <= 7Fh`。
const moraleRestoredByte = 0xB3

// moraleFleeEffectKind 是逃走那一段掛上的效果碼 `23h`。語意未解讀。
const moraleFleeEffectKind = 0x23

// MoraleFailureResult 是一次士氣檢定失敗的結果。
type MoraleFailureResult struct {
	FighterID string        `json:"fighter_id"`
	Roll      int           `json:"roll"`
	Outcome   MoraleOutcome `json:"outcome"`
	MessageID string        `json:"message_id"`
}

// ResolveMoraleFailure 擲 1d100 並套上對應那一段。
//
// ⚠ 只有「逃走」那一段會動到戰鬥員的狀態；混亂／狂暴／暴怒在原作是各自呼叫
// 另一支處理（`far 1572h`／`本模組 143Ah`），那幾支還沒解讀。這裡如實回報結果與
// 訊息，**不假裝已經套上效果**——覆蓋報告要看得出這個差別。
func (b *Battle) ResolveMoraleFailure(fighterID string) (MoraleFailureResult, error) {
	if b == nil || b.rng == nil {
		return MoraleFailureResult{}, errNoPRNG
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return MoraleFailureResult{}, fmt.Errorf("unknown fighter %q", fighterID)
	}
	result := MoraleFailureResult{FighterID: fighterID, Roll: b.rng.Intn(100) + 1}
	switch {
	case result.Roll <= 10:
		result.Outcome = MoraleRunsAway
	case result.Roll <= 60:
		result.Outcome = MoraleConfused
	case result.Roll <= 80:
		result.Outcome = MoraleBerserk
	default:
		result.Outcome = MoraleEnraged
	}
	result.MessageID = result.Outcome.MessageID()
	if result.Outcome != MoraleRunsAway {
		return result, nil
	}
	fighter.MonsterAffects = append(fighter.MonsterAffects,
		MonsterAffect{Kind: moraleFleeEffectKind, Active: true})
	// `+18Dh^[10h] := 1`：這一格同時擋掉「被選為目標」（spec 815）、
	// 「印恐慌逃走訊息」（spec 830）與「被障礙格擋住」（spec 1119）。
	fighter.RawCombatState10 = 1
	fighter.Escaped = true
	if fighter.ControlMorale <= 0x7F {
		fighter.ControlMorale = moraleRestoredByte
	}
	// 清掉目前目標（`+18Dh^[0Ah]`，spec 782／790）。動的是
	// `ActionTargetID` 不是 `TargetID`——後者是待處理法術的交接欄位，
	// 清掉它會讓一支還沒解算的法術失去目標。
	fighter.CombatAction.ClearActionTarget()
	b.fighters[fighterID] = fighter
	b.updateStatus()
	return result, nil
}
