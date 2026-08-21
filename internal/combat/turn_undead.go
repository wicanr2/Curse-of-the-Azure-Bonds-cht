package combat

import (
	"fmt"
	"sort"
)

// turnUndeadMatrix 是原作 `DS:0367h` 那 100 個 byte，逐 byte 取自
// `workplace/re-sweep/dos/dseg` 的資料段 dump（spec 834）。
// 索引：`[不死種類 − 1][職業級距 − 1]`，值以**有號**解讀。
//
//	2..20   1d20 要 ≥ 這個值才驅離
//	1       1d20 必過 → 自動驅離
//	0／−1   abs 之後是 0 或 1，1d20 必過，而且 v ≤ 0 走摧毀那條 → 自動摧毀
//	99      1d20 到不了 99 → 不可能
//
// ⚠ `0` 與 `−1` 的**驅離／摧毀**行為相同，但**額外配額那一條不同**：原作寫的是
// `if 有號(v) < 0`，所以 `0` 那幾格摧毀之後不會把配額加回來。表要照原始位元組
// 存，不能把兩者統一成同一個值。
var turnUndeadMatrix = [10][10]int8{
	{10, 7, 4, 1, 1, 0, 0, -1, -1, -1},
	{13, 10, 7, 1, 1, 0, 0, 0, -1, -1},
	{16, 13, 10, 4, 1, 1, 0, 0, 0, -1},
	{19, 16, 13, 7, 4, 1, 1, 0, 0, 0},
	{20, 19, 16, 10, 7, 4, 1, 1, 0, 0},
	{99, 20, 19, 13, 10, 7, 4, 1, 1, 0},
	{99, 99, 20, 16, 13, 10, 7, 4, 1, 0},
	{99, 99, 99, 20, 16, 13, 10, 7, 4, 1},
	{99, 99, 99, 99, 20, 16, 13, 10, 7, 1},
	{99, 99, 99, 99, 99, 20, 16, 13, 10, 4},
}

// turnUndeadExtraDestroys 是「摧毀不佔 1d12 配額」的上限（原作 `額外 := 6`）。
const turnUndeadExtraDestroys = 6

// TurnUndeadOutcome 是一個目標的結果。
type TurnUndeadOutcome struct {
	TargetID   string `json:"target_id"`
	UndeadType uint8  `json:"undead_type"`
	// Threshold 是查表得到的門檻（已取絕對值）。
	Threshold int  `json:"threshold"`
	Destroyed bool `json:"destroyed,omitempty"`
}

// TurnUndeadResult 是一次驅散不死（`overlay-13:1227h`，spec 834）。
type TurnUndeadResult struct {
	ClericID string `json:"cleric_id"`
	// Column 是職業級距 1..10：1..8 直接用牧師槽的等級，9..13 是 9，其餘是 10。
	Column int `json:"column"`
	// Roll 是那一顆 1d20，**整批目標共用**。
	Roll int `json:"roll"`
	// Allowance 是 1d12：這一次最多影響幾個。
	Allowance int                 `json:"allowance"`
	Outcomes  []TurnUndeadOutcome `json:"outcomes,omitempty"`
	// Stopped 是「有一個沒過，整批就停」。原作不是逐個判定。
	Stopped bool `json:"stopped,omitempty"`
}

// Nothing 回報這一次一個都沒成（原作印 `Nothing Happens...`）。
func (r TurnUndeadResult) Nothing() bool { return len(r.Outcomes) == 0 }

// turnUndeadColumn 把牧師的等級換成矩陣的欄。
//
// ⚠ 原作在 1..8 這一段用的是**牧師槽的等級**（`+109h`）而不是合計等級 `L`，
// 所以多職角色在這個區間會取到第一職業的等級（spec 834）。
func turnUndeadColumn(clericLevel, totalLevel int) int {
	switch {
	case totalLevel >= 1 && totalLevel <= 8:
		return clericLevel
	case totalLevel >= 9 && totalLevel <= 13:
		return 9
	default:
		return 10
	}
}

// TurnUndead 重現原作的驅散不死。
//
// ★ 四個容易做錯的地方（spec 834）：
//
//   - **1d20 只擲一次**，整批目標共用；1d12 決定這一次最多影響幾個。
//   - **第一個沒過就整批停**，不是逐個判定。
//   - **摧毀不佔那 1d12 的額度**：配額歸零時若還有額外次數（初值 6）而且這一次
//     走的是摧毀（`v < 0`），就把配額加回 1。
//   - **目標從最弱的開始**（`+0E9h` 最小的，spec 815 的挑選器）。
func (b *Battle) TurnUndead(clericID string) (TurnUndeadResult, error) {
	if b == nil {
		return TurnUndeadResult{}, fmt.Errorf("battle is nil")
	}
	if b.rng == nil {
		return TurnUndeadResult{}, fmt.Errorf("battle PRNG is unavailable")
	}
	cleric, ok := b.fighters[clericID]
	if !ok {
		return TurnUndeadResult{}, fmt.Errorf("unknown fighter %q", clericID)
	}
	if b.status != StatusActive {
		return TurnUndeadResult{}, fmt.Errorf("battle is already over")
	}
	if cleric.ClericLevel == 0 {
		return TurnUndeadResult{}, fmt.Errorf("fighter %q is not a cleric", clericID)
	}
	if cleric.TriedToTurn {
		return TurnUndeadResult{}, fmt.Errorf("fighter %q already turned this combat", clericID)
	}
	// 原作一開場就設 `+18Dh^[11h] := 1`，用過這個選項就從選單消失（spec 905）。
	cleric.TriedToTurn = true
	b.fighters[clericID] = cleric

	result := TurnUndeadResult{
		ClericID:  clericID,
		Column:    turnUndeadColumn(int(cleric.ClericLevel), int(cleric.TotalLevel())),
		Allowance: b.rng.Intn(12) + 1,
		Roll:      b.rng.Intn(20) + 1,
	}
	if result.Column < 1 || result.Column > 10 {
		return TurnUndeadResult{}, fmt.Errorf("cleric %q maps to turn column %d", clericID, result.Column)
	}

	remaining := result.Allowance
	extra := turnUndeadExtraDestroys
	for _, target := range b.turnUndeadTargets(cleric) {
		if remaining <= 0 {
			break
		}
		value := turnUndeadMatrix[target.UndeadType-1][result.Column-1]
		threshold := int(value)
		if threshold < 0 {
			threshold = -threshold
		}
		if result.Roll < threshold {
			// ⚠ 原作是 `停 := 1`（整批停），不是跳過這一個。**在目前的挑選
			// 順序下兩種寫法看不出差別**：矩陣每一欄由上往下不遞減，而目標又是
			// 由弱到強，所以第一個沒過的後面一定也不會過。照原作寫 break。
			result.Stopped = true
			break
		}
		outcome := TurnUndeadOutcome{
			TargetID: target.ID, UndeadType: target.UndeadType,
			Threshold: threshold, Destroyed: value <= 0,
		}
		affected := b.fighters[target.ID]
		if outcome.Destroyed {
			// 原作寫 `+195h := 8`（Gone）與 `+196h := 0`：離場而且不留屍體。
			affected.HitPoints = 0
			affected.DownedCorpse = false
		} else {
			// `+18Dh^[10h] := 1` ＝ `TURNED`：掉頭就跑（spec 1165）。
			affected.RawCombatState10 = 1
		}
		b.fighters[target.ID] = affected
		result.Outcomes = append(result.Outcomes, outcome)

		if extra > 0 {
			extra--
		}
		remaining--
		if remaining == 0 && extra > 0 && value < 0 {
			remaining++
		}
	}
	return result, nil
}

// turnUndeadTargets 回傳可以被驅散的目標，**最弱的排前面**。
// 種類相同時照戰鬥員的既定順序，結果才可重現。
func (b *Battle) turnUndeadTargets(cleric Fighter) []Fighter {
	targets := make([]Fighter, 0, 8)
	for index, candidate := range b.Fighters() {
		if !opposesInBattle(cleric, candidate) {
			continue
		}
		if candidate.HitPoints <= 0 || candidate.Escaped {
			continue
		}
		if candidate.UndeadType < 1 || candidate.UndeadType > 10 {
			continue
		}
		if candidate.RawCombatState10 != 0 {
			// 已經在逃的不再算一次。
			continue
		}
		candidate.turnOrderIndex = index
		targets = append(targets, candidate)
	}
	sort.SliceStable(targets, func(left, right int) bool {
		if targets[left].UndeadType != targets[right].UndeadType {
			return targets[left].UndeadType < targets[right].UndeadType
		}
		return targets[left].turnOrderIndex < targets[right].turnOrderIndex
	})
	return targets
}

// TotalLevel 是驅散矩陣換算欄位用的合計等級（原作的 `L`）。
//
// ⚠ 原作是 `+109h + +111h × far 1652h(角色)`——那個呼叫的回傳值沒有讀出來，
// 形狀上是「要不要把第二職業算進來」（spec 834）。這裡直接相加：單職角色的
// `+111h` 是 0，結果與原作相同；多職角色是本專案目前的讀法。
func (f Fighter) TotalLevel() uint8 {
	total := int(f.ClericLevel) + int(f.SecondClassLevel)
	if total > 255 {
		return 255
	}
	return uint8(total)
}
