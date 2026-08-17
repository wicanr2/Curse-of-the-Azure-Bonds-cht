package combat

import "fmt"

// AI 的移動：模式決定五個候選方向，走不動就換下一個（spec 830／837／838）。
//
// ★ 為什麼不是「朝目標直線走」。 原作的 AI 每回合抽一個**行為模式**（1..6），
// 模式決定「試方向的順序」——有的偏左、有的偏右、有的沿牆繞。直線衝只是模式
// 1／2 的第一個候選。少了這一層，怪物會在障礙前面卡住，而原版不會。

// aiDirectionOffsets 是 `DS:2B6h` 的方向偏移表，索引是 `模式 × 5 + 序號`，
// **序號 1..5**（spec 837）。偏移加在「基準方向」上再 mod 8，所以 8 ＝ 不轉、
// 1 ＝ ＋45°、7 ＝ −45°、4 ＝ 180°。
//
// 模式 1／2 直線衝（卡住先試小角度，一個偏左一個偏右）、3／4 斜著切、
// 5／6 是左手／右手法則繞行。`TestAIDirectionTableMatchesTheResidentDump`
// 逐 byte 對回原始資料段。
var aiDirectionOffsets = map[int][5]int{
	1: {8, 7, 6, 1, 2},
	2: {8, 1, 2, 7, 6},
	3: {7, 1, 8, 6, 2},
	4: {1, 7, 8, 2, 6},
	5: {8, 7, 6, 5, 4},
	6: {8, 1, 2, 3, 4},
}

// aiApproachIterationCap 是原作的 20 次上限（`14h`，spec 838）。
// 那是防呆，不是設計參數——照抄就好。
const aiApproachIterationCap = 20

// ApproachResult 是一次 AI 接近的結果。
type ApproachResult struct {
	FighterID string
	// Mode 是這一回合用的行為模式（1..6）。
	Mode int
	// Steps 是實際走過的格子，依序。
	Steps []TilePoint
	// HalfTilesUsed 是花掉的移動力，單位**半格**：正向 ×2、斜向 ×3（spec 837）。
	HalfTilesUsed int
	// InWeaponRange 是「停下來的時候打得到目標」。
	InWeaponRange bool
	// Blocked 是「五個候選方向都走不了」——沒有移動力、被擋、或地形不可通行。
	Blocked bool
	Iterations int
}

// RollAIMode 重現每回合的模式重骰（spec 830）：
//
//	模式 ∈ 1..4 且 1d4 ≠ 1  → 保持
//	其餘                     → 重骰：1d8 擲到 8 就再擲 1d2 ＋ 4（得 5 或 6），
//	                                 否則擲 1d4（得 1..4）
//
// ⚠ 那次 1d8 **的點數本身被丟掉**，只用來決定走哪一條分支。要對齊亂數序列就
// 不能省這一擲。
func RollAIMode(roll func(sides int) int, current int) int {
	if current >= 1 && current <= 4 && roll(4) != 1 {
		return current
	}
	if roll(8) == 8 {
		return roll(2) + 4
	}
	return roll(4)
}

// BeginMonsterTurn 依 spec 830 重骰這個戰鬥員的行為模式並記進 Fighter。
// 亂數走 Battle 自己那一條，所以模式重骰與其他戰鬥擲骰共用同一個序列。
func (b *Battle) BeginMonsterTurn(fighterID string) (int, error) {
	if b == nil || b.rng == nil {
		return 0, fmt.Errorf("battle PRNG is unavailable")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return 0, fmt.Errorf("unknown fighter %q", fighterID)
	}
	fighter.AIMode = RollAIMode(func(sides int) int { return b.rng.Intn(sides) + 1 }, fighter.AIMode)
	b.fighters[fighterID] = fighter
	return fighter.AIMode, nil
}

// MonsterApproach 讓一個戰鬥員朝目標走，直到打得到、走不動、或撞到 20 次上限。
//
// 這一支只做**移動**：攻擊由呼叫端接手，因為攻擊要動到武器、彈藥與訊息，
// 那些不屬於移動。回傳的 `InWeaponRange` 就是「可以打了」。
//
// 移動力的單位是半格：預算取 `MovementAllowance × 2`，正向一步花 `地形 × 2`、
// 斜向花 `地形 × 3`。所以移動力 6 的怪走得了 6 步正向或 4 步斜向。
func (b *Battle) MonsterApproach(fighterID, targetID string, mode int, terrain MovementTerrain) (ApproachResult, error) {
	if b == nil {
		return ApproachResult{}, fmt.Errorf("battle is nil")
	}
	offsets, ok := aiDirectionOffsets[mode]
	if !ok {
		return ApproachResult{}, fmt.Errorf("AI mode %d is outside 1..6", mode)
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return ApproachResult{}, fmt.Errorf("unknown fighter %q", fighterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return ApproachResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if !fighter.HasCombatPosition || !target.HasCombatPosition {
		return ApproachResult{}, fmt.Errorf("approach needs both combatants placed")
	}
	result := ApproachResult{FighterID: fighterID, Mode: mode}
	// 移動率的效果修正走 `CHECKFX(12h)`（spec 1123）：緩速折半、加速…
	allowance, err := CheckFXValue(fighter, CheckFXMovement, scratchMovement, fighter.MovementAllowance)
	if err != nil {
		return ApproachResult{}, err
	}
	if allowance < 0 {
		allowance = 0
	}
	budget := allowance * 2
	weaponRange := fighter.WeaponRange
	if weaponRange < 1 {
		// 沒有裝備武器時射程 1（spec 838：表填 0 或減 1 溢位都當 1）。
		weaponRange = 1
	}
	for result.Iterations < aiApproachIterationCap {
		if weightedTileDistance(
			TilePoint{X: fighter.CombatX, Y: fighter.CombatY},
			TilePoint{X: target.CombatX, Y: target.CombatY}) <= weaponRange*2+1 {
			result.InWeaponRange = true
			return result, nil
		}
		result.Iterations++
		base := directionToward(fighter, target)
		moved := false
		for _, offset := range offsets {
			direction := (base + offset) % 8
			delta, ok := DirectionDelta(uint8(direction))
			if !ok {
				continue
			}
			next := TilePoint{X: fighter.CombatX + delta.X, Y: fighter.CombatY + delta.Y}
			cost, passable := stepCost(terrain, fighter, next, direction)
			if !passable || cost > budget {
				continue
			}
			if b.occupiedAt(fighter, next) {
				continue
			}
			if b.obstacleBlocks(fighter, next) {
				continue
			}
			fighter.CombatX, fighter.CombatY = next.X, next.Y
			b.fighters[fighterID] = fighter
			budget -= cost
			result.HalfTilesUsed += cost
			result.Steps = append(result.Steps, next)
			moved = true
			break
		}
		if !moved {
			result.Blocked = true
			return result, nil
		}
	}
	return result, nil
}

// directionToward 是「從我看向目標是哪一個方向」（原作 `0CA2h+1`）。
// 方向編號沿用 `DirectionDelta`：0 北、順時針到 7 西北，奇數是斜向。
func directionToward(fighter, target Fighter) int {
	dx := sign(target.CombatX - fighter.CombatX)
	dy := sign(target.CombatY - fighter.CombatY)
	for direction := 0; direction < 8; direction++ {
		delta, _ := DirectionDelta(uint8(direction))
		if delta.X == dx && delta.Y == dy {
			return direction
		}
	}
	return 0
}

// stepCost 依 spec 837：正向 `地形成本 × 2`、斜向 `× 3`。
// 大於一格的腳印每一格都要能走，成本取最大的那一格——與
// `MoveWithTerrainAndFreeAttacks` 同一個約定。
func stepCost(terrain MovementTerrain, fighter Fighter, next TilePoint, direction int) (int, bool) {
	multiplier := 2
	if direction%2 == 1 {
		multiplier = 3
	}
	if terrain == nil {
		return multiplier, true
	}
	footprint := FootprintForSize(fighter.CombatSize)
	worst := 1
	for y := next.Y; y < next.Y+footprint.Height; y++ {
		for x := next.X; x < next.X+footprint.Width; x++ {
			cost, passable := terrain(x, y)
			if !passable || cost < 1 {
				return 0, false
			}
			worst = max(worst, cost)
		}
	}
	return worst * multiplier, true
}

// occupiedAt 是「這一步會不會踩到別人」。倒下與離場的不算佔位。
func (b *Battle) occupiedAt(mover Fighter, next TilePoint) bool {
	for _, other := range b.Fighters() {
		if other.ID == mover.ID || other.HitPoints <= 0 || other.Escaped || !other.HasCombatPosition {
			continue
		}
		if FootprintsOverlapAt(mover, next.X, next.Y, other) {
			return true
		}
	}
	return false
}

// SelectInRangeTarget 是 spec 838 §五的目標挑法：**先看射程內有沒有人**，
// 有就從候選裡**均勻隨機**挑一個（原作是 `1d n`），沒有才輪到移動。
//
// ★ 這不是「挑最近的」也不是「挑最弱的」。原作把射程內的候選編號填進
// `DS:758Eh`，再擲一次 `1d n` 取其中一個——所以同一個場面每回合的目標會變。
// 挑最近的會讓 AI 顯得比原版聰明且可預測。
func (b *Battle) SelectInRangeTarget(fighterID string, side Side) (Fighter, bool, error) {
	if b == nil || b.rng == nil {
		return Fighter{}, false, fmt.Errorf("battle PRNG is unavailable")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return Fighter{}, false, fmt.Errorf("unknown fighter %q", fighterID)
	}
	if !fighter.HasCombatPosition {
		return Fighter{}, false, nil
	}
	weaponRange := fighter.WeaponRange
	if weaponRange < 1 {
		weaponRange = 1
	}
	candidates := make([]Fighter, 0, len(b.fighterOrder))
	for _, id := range b.fighterOrder {
		other := b.fighters[id]
		if other.Side != side || other.HitPoints <= 0 || other.Escaped || !other.HasCombatPosition {
			continue
		}
		distance := weightedTileDistance(
			TilePoint{X: fighter.CombatX, Y: fighter.CombatY},
			TilePoint{X: other.CombatX, Y: other.CombatY})
		if distance <= weaponRange*2+1 {
			candidates = append(candidates, other)
		}
	}
	if len(candidates) == 0 {
		return Fighter{}, false, nil
	}
	return candidates[b.rng.Intn(len(candidates))], true, nil
}

// obstacleBlocks 是 spec 1119 的兩種障礙格。腳印大於一格時**每一格都要能過**
// ——原作查的就是四個腳印格。
//
// ⚠ 豁免只在真的擋到時才擲：把 `RollSavingThrow` 提到迴圈外會多擲好幾次骰，
// 亂數序列就跟原版分家了。
func (b *Battle) obstacleBlocks(fighter Fighter, next TilePoint) bool {
	if b == nil {
		return false
	}
	lookup := b.terrainCode
	if lookup == nil {
		// 沒有外部地形碼來源時，退回**戰鬥中自己鋪出來的**那兩種雲
		// （spec 1128：惡臭之雲寫 `1Eh`、致命毒雲寫 `1Ch`）。
		//
		// ⚠ 這條路一律回報 `onMap = true`：它不知道地圖邊界，而呼叫端把
		// `onMap = false` 當成「擋住」——謊報出界會讓所有移動全部卡死。
		lookup = func(x, y int) (uint8, bool) {
			code, _ := b.PersistentAreaTerrainCode(x, y)
			return code, true
		}
	}
	footprint := FootprintForSize(fighter.CombatSize)
	waiver := fighter.RawCombatState10 != 0
	for y := next.Y; y < next.Y+footprint.Height; y++ {
		for x := next.X; x < next.X+footprint.Width; x++ {
			code, onMap := lookup(x, y)
			if !onMap {
				return true
			}
			if code != ObstacleTerrainSaveable && code != ObstacleTerrainVeteran {
				continue
			}
			save := func() bool {
				// 類別 0 ＝ 五張豁免表的第一張，修正 0（原作 `far 1458h(角色, 0, 0)`）。
				result, err := b.RollSavingThrow(fighter, 0, 0)
				return err == nil && result.Saved
			}
			if ObstacleTerrainBlocks(fighter, code, waiver, save) {
				return true
			}
		}
	}
	return false
}
