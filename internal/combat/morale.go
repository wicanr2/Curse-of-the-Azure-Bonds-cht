package combat

import "fmt"

// 士氣崩潰的四種結果（spec 831）。
//
// ★ 為什麼不是「士氣崩了就逃走」。 原作擲一次 1d100 分四段：**只有 10% 真的
// 逃走**，50% 混亂、20% 狂暴、20% 暴怒。四段加起來剛好 100，沒有落空的骰值。
// 只做「逃走」會讓怪物在士氣崩潰之後的行為少掉九成。

// MoraleCheckResult 是一次士氣檢定（`overlay-09:01388h`，spec 1122）。
type MoraleCheckResult struct {
	FighterID string `json:"fighter_id"`
	// AlreadyFled 是「這個人已經在逃了」（`+18Dh^[10h] <> 0`）。這種情況不做
	// 檢定，直接標成本回合撤退。
	//
	// ⚠ `+10h` 在 PC-98 的 `COMBATVARREC` 裡叫 `TURNED`——**被牧師逼退的不死
	// 生物**。兩者的可觀察行為相同（都是掉頭就跑），但這一格的來源是轉化，
	// 不是先前的士氣崩潰；`+14h`（`ROUTING`）才是士氣那條（spec 1165）。
	AlreadyFled bool `json:"already_fled,omitempty"`
	// Checked 是「真的做了檢定」。`+0F7h` 最高位沒設就沒有士氣資料，直接離開。
	Checked bool `json:"checked"`
	Morale  int  `json:"morale,omitempty"`
	// Threshold 是 `100 − 目前HP佔最大HP的百分比`：傷得越重門檻越高。
	Threshold int `json:"threshold,omitempty"`
	Passed    bool `json:"passed"`
	// Withdrew 是 `+18Dh^[14h] := 1`（PC-98 叫 `ROUTING`）：士氣崩了而且跑得掉。
	Withdrew bool `json:"withdrew,omitempty"`
	// MessageID 只在要印訊息時有值。
	MessageID string `json:"message_id,omitempty"`
}

// moraleWithdrawEffectKinds 是撤退成功時掛上的兩個效果碼。語意未解讀。
var moraleWithdrawEffectKinds = []uint8{0x4A, 0x4B}

// CheckMorale 重現 `overlay-09:01388h`：AI 回合開始時做的士氣檢定。
//
// ★ 這不是「士氣歸零就逃」。 門檻與**受傷程度**綁在一起：`士氣 ≥ 100 − HP%`
// 才算過，所以滿血時幾乎一定過、剩一成血時要 90 以上。過不了才看跑不跑得掉。
//
// ⚠ 跑得掉的判準是**移動率**（`敵方最快的 ÷ 2 ≤ 自己的 ÷ 2`），不擲骰——
// 與 spec 799 的脫離戰場是不同的兩支，那一支平手時擲 1d2。
func (b *Battle) CheckMorale(fighterID string) (MoraleCheckResult, error) {
	if b == nil {
		return MoraleCheckResult{}, fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return MoraleCheckResult{}, fmt.Errorf("unknown fighter %q", fighterID)
	}
	result := MoraleCheckResult{FighterID: fighterID}
	// 每回合先清 `+14h`，所以上一回合的撤退不會殘留。
	fighter.RetreatedThisRound = false
	if fighter.RawCombatState10 != 0 {
		fighter.RetreatedThisRound = true
		result.AlreadyFled, result.Withdrew = true, true
		b.fighters[fighterID] = fighter
		return result, nil
	}
	morale, valid := MoraleValue(fighter.ControlMorale)
	if !valid {
		// 沒有士氣資料就不檢定——**不是當成 0**，那會讓沒有資料的怪物一律逃走。
		b.fighters[fighterID] = fighter
		return result, nil
	}
	// 士氣的效果修正走 `CHECKFX(11h)`（spec 1123）：祝福 ＋5、詛咒 −5、魅惑。
	morale, err := CheckFXValue(fighter, CheckFXMorale, scratchMorale, morale)
	if err != nil {
		return MoraleCheckResult{}, err
	}
	if morale < 0 {
		morale = 0
	}
	result.Checked, result.Morale = true, morale
	healthPercent := 0
	if fighter.MaxHitPoints > 0 {
		healthPercent = fighter.HitPoints * 100 / fighter.MaxHitPoints
	}
	result.Threshold = 100 - healthPercent
	if morale >= result.Threshold && morale != 0 {
		result.Passed = true
		b.fighters[fighterID] = fighter
		return result, nil
	}
	if b.canOutrunOpposingSide(fighter) {
		fighter.RetreatedThisRound = true
		result.Withdrew = true
		for _, kind := range moraleWithdrawEffectKinds {
			fighter.MonsterAffects = append(fighter.MonsterAffects,
				MonsterAffect{Kind: kind, Active: true})
		}
	}
	b.fighters[fighterID] = fighter
	return result, nil
}

// canOutrunOpposingSide 是 `overlay-13 entry#26` 對上 `entry#2`：
// 對面最快的移動率折半，與自己的移動率折半比。**整數除法**——原作兩邊都
// `div 2`，所以 5 與 4 折半之後都是 2，算跑得掉。
func (b *Battle) canOutrunOpposingSide(fighter Fighter) bool {
	fastest := 0
	for _, other := range b.fighters {
		if other.Side == fighter.Side || other.HitPoints <= 0 || other.Escaped {
			continue
		}
		if half := other.MovementAllowance / 2; half > fastest {
			fastest = half
		}
	}
	return fastest <= fighter.MovementAllowance/2
}

// PanicMessageID 是 spec 830 的條件：`+14h <> 0` 且 `+10h = 0` 才印。
// 已經逃走的人不會再被通報一次。
func (f Fighter) PanicMessageID() (string, bool) {
	if f.RetreatedThisRound && f.RawCombatState10 == 0 {
		return "combat_flees_in_panic", true
	}
	return "", false
}

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

// AreaMoraleBreakResult 是一次範圍士氣崩潰（混亂術，效果碼 `23h`）。
type AreaMoraleBreakResult struct {
	CasterID string                `json:"caster_id"`
	Center   TilePoint             `json:"center"`
	Impacts  []AreaMoraleBreakHit  `json:"impacts"`
}

// AreaMoraleBreakHit 是一個目標的結果。
type AreaMoraleBreakHit struct {
	TargetID string `json:"target_id"`
	SaveRoll int    `json:"save_roll,omitempty"`
	Saved    bool   `json:"saved,omitempty"`
	// Outcome 只在沒過豁免時有值。
	Outcome   MoraleOutcome `json:"outcome,omitempty"`
	MessageID string        `json:"message_id,omitempty"`
}

// CastAreaMoraleBreak 讓半徑內每個沒過豁免的目標跑一次士氣崩潰的四段表。
//
// ★ 混亂術（法術 82）的效果碼就是 `23h`，而 `23h` 的處理常式正是那張四段表
// （spec 831／1122）。所以這支法術不需要新的規則，只需要接上既有的那一張表。
func (b *Battle) CastAreaMoraleBreak(casterID string, center TilePoint, radius,
	saveCategory int) (AreaMoraleBreakResult, error) {
	if b == nil || b.rng == nil {
		return AreaMoraleBreakResult{}, errNoPRNG
	}
	if _, ok := b.fighters[casterID]; !ok {
		return AreaMoraleBreakResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return AreaMoraleBreakResult{}, fmt.Errorf("battle is already over")
	}
	result := AreaMoraleBreakResult{CasterID: casterID, Center: center}
	for _, id := range b.fighterOrder {
		target := b.fighters[id]
		if target.HitPoints <= 0 || target.Escaped || !target.HasCombatPosition ||
			!fighterFootprintWithinRadius(target, center, radius) {
			continue
		}
		save, err := b.RollSavingThrow(target, saveCategory, 0)
		if err != nil {
			return AreaMoraleBreakResult{}, err
		}
		hit := AreaMoraleBreakHit{TargetID: id, SaveRoll: save.Roll, Saved: save.Saved}
		if !save.Saved {
			broken, err := b.ResolveMoraleFailure(id)
			if err != nil {
				return AreaMoraleBreakResult{}, err
			}
			hit.Outcome, hit.MessageID = broken.Outcome, broken.MessageID
		}
		result.Impacts = append(result.Impacts, hit)
	}
	b.updateStatus()
	return result, nil
}
