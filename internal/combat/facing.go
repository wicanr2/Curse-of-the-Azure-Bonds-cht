package combat

import (
	"fmt"

	engineability "github.com/wicanr2/golden-box-remake-engine/combat/ability"
	enginearmorclass "github.com/wicanr2/golden-box-remake-engine/combat/armorclass"
	enginefacing "github.com/wicanr2/golden-box-remake-engine/combat/facing"
)

// 戰鬥面向與轉向記帳（原作戰鬥狀態記錄的 `+09h`／`+0Fh`／`+12h`，spec 1137）。
//
// 幾何與記帳本身在共用 engine 的 `combat/facing`；這一檔只做 `Fighter` 與
// 那個 plain-value 契約之間的轉接，以及 CoAB 專屬的閘（效果碼、旁路條件）。
//
// ★ 原作兩個寫入端對「`+09h` 到底指哪個方向」給出相反的形狀（spec 1019 的
// `(d + 4) mod 8` 與 `sub_194A` 的 `|d − 面向|`），所以「角色到底在看哪裡」
// 還沒有結論。engine 與這裡都只做字面移植。

// facingStateOf 把 `Fighter` 投影成 engine 的面向狀態。
func facingStateOf(fighter Fighter) enginefacing.State {
	return enginefacing.State{
		X: fighter.CombatX, Y: fighter.CombatY,
		Facing:      fighter.CombatFacing,
		ActionCount: fighter.CombatActionCount,
		TurnTotal:   fighter.CombatTurnTotal,
	}
}

// applyFacingState 把 engine 算完的狀態寫回 `Fighter`（座標不動）。
func applyFacingState(fighter *Fighter, state enginefacing.State) {
	fighter.CombatFacing = state.Facing
	fighter.CombatActionCount = state.ActionCount
	fighter.CombatTurnTotal = state.TurnTotal
}

// CombatDirectionFrom 回傳從 `(fromX,fromY)` 看向 `(toX,toY)` 的八方向碼
// （原作 `overlay-13:29A2h`）。
func CombatDirectionFrom(fromX, fromY, toX, toY int) uint8 {
	return enginefacing.DirectionFrom(fromX, fromY, toX, toY)
}

// combatDirection 是原作的參數順序：回傳**從 (x2,y2) 看向 (x1,y1)** 的方向碼。
func combatDirection(x1, y1, x2, y2 int) uint8 {
	return enginefacing.Direction(x1, y1, x2, y2)
}

// turnDistance 是最短轉法（0..4，順逆時針不分）。
func turnDistance(a, b uint8) uint8 { return enginefacing.TurnDistance(a, b) }

// InFacingCone 是原作 `overlay-31:054Ah`（`INARC`，spec 1002）：目標在不在
// 「起點往 direction 前進一格」為頂點、半角 45° 的扇形內。
func InFacingCone(startX, startY, targetX, targetY int, direction uint8) bool {
	return enginefacing.InArc(startX, startY, targetX, targetY, direction, enginefacing.GoldBoxBounds)
}

// AccountTurn 是原作 `overlay-13:194Ah`（`retf 8`，(角色, 對象)）：
// 動作計數加一，並把最短轉法累加進 `+12h`。⚠ 原作這一支**不寫 `+09h`**。
func (b *Battle) AccountTurn(fighterID, towardID string) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	toward, ok := b.fighters[towardID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", towardID)
	}
	state := facingStateOf(fighter)
	state.AccountTurn(toward.CombatX, toward.CombatY)
	applyFacingState(&fighter, state)
	b.fighters[fighterID] = fighter
	return nil
}

// FaceTarget 是原作攻擊動作（`overlay-13:19D8h`，spec 1019）的轉向段：
// **動作計數小於 2** 時才把 `+09h` 設成 `(方向 + 4) mod 8`。
//
// ⚠ DOS 的 `+ 4` 是平台差異，PC-98 把這一段整個拿掉；本專案以 DOS 為 oracle。
func (b *Battle) FaceTarget(fighterID, targetID string) (bool, error) {
	if b == nil {
		return false, fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", fighterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", targetID)
	}
	state := facingStateOf(fighter)
	turned := state.FaceTarget(target.CombatX, target.CombatY)
	if !turned {
		return false, nil
	}
	applyFacingState(&fighter, state)
	b.fighters[fighterID] = fighter
	return true, nil
}

// RearAttackApplies 是原作攻擊結算（`overlay-13:14E8h`）挑第二個 AC 的三道條件。
// ⚠ 三道都是字面移植，**沒有宣稱它叫「背後攻擊」**。
func (b *Battle) RearAttackApplies(attackerID, defenderID string) (bool, error) {
	if b == nil {
		return false, fmt.Errorf("battle is nil")
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", attackerID)
	}
	defender, ok := b.fighters[defenderID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", defenderID)
	}
	return enginefacing.RearAttackApplies(facingStateOf(attacker), facingStateOf(defender)), nil
}

// initialFacingTable 是原作 `DS:2FAh` 的初始面向表（兩平台相同，spec 805）。
// ⚠ 資料段那裡實際上有六個位元組，但索引是**朝向除以 2**，所以只用得到前四筆。
var initialFacingTable = []uint8{7, 2, 3, 6}

// ApplyInitialFacing 是原作戰鬥開場配置戰鬥狀態時設的面向（spec 805）：
// `+09h := 表[有號(隊伍朝向) div 2]`，敵方再轉 180 度。
func (b *Battle) ApplyInitialFacing(partyDirection uint8) {
	if b == nil {
		return
	}
	for _, id := range b.fighterOrder {
		fighter := b.fighters[id]
		facing, ok := enginefacing.InitialFacing(initialFacingTable, partyDirection,
			fighter.Side == SideEnemy)
		if !ok {
			continue
		}
		fighter.CombatFacing = facing
		b.fighters[id] = fighter
	}
}

// opportunityAttackBlockingAffects 是原作 `overlay-24 entry#6`（`sub_BE3`）查的
// 四個效果碼，表在 `DS:27CAh`..`27CDh`（已初始化資料，不是 BSS）。四個都掛在
// `CHECKFX(07h)`（「這一回合動得了嗎」）的同一支 handler `overlay-12:0075h`，
// 其中 `35h` 的訊息是 `falls asleep`。
//
// ⚠ 這四個是 `MonsterIsHeld` 那五個的**子集**——原作這張表沒有 `1Bh`。
// 直接借用 `MonsterIsHeld` 會多擋一個效果。
var opportunityAttackBlockingAffects = []uint8{0x33, 0x34, 0x35, 0x1F}

// opportunityAttackWithdrawnAffects 是原作連問兩次 `overlay-24 entry#27` 的兩個
// 效果碼。士氣崩潰而且跑得掉時掛的就是這兩個（spec 831），所以「已經撤退的人
// 不打離場的那一下」。
var opportunityAttackWithdrawnAffects = []uint8{0x4B, 0x4A}

// opportunityAttackAllowed 是原作「離開接觸就被打」（`overlay-13:095Ah`，
// spec 1010）在動手之前的四道閘，順序照原作：
//
//  1. `overlay-24 entry#6`：它本身就是對 `entry#27` 問四次，所以查法與
//     第 3 道相同（只比對效果碼，不看持續時間）；打手動不了就不打；
//  2. `sub_1144(角色, 打手)`：`CHECKFX(01h, 角色)` 的否決權——閃現與隱形
//     那一組（`25h`／`19h`／`47h`／`45h`）就是這個時機的成員，
//     所以看不見離場的人就打不到；
//  3. `overlay-24 entry#27` 連問兩次 `4Bh`／`4Ah`：已經撤退的人不打；
//  4. 面向（見下）。
//
// ⚠ `CHECKFX(00h, 打手)` 那一次在 CoAB 是空的（時機 `00h` 沒有任何效果碼），
// 所以第 2 道實際上只剩 `01h` 那一問。
func (b *Battle) opportunityAttackAllowed(attackerID, moverID string) (bool, error) {
	if b == nil {
		return false, fmt.Errorf("battle is nil")
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", attackerID)
	}
	mover, ok := b.fighters[moverID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", moverID)
	}
	if fighterHasAnyAffect(attacker, opportunityAttackBlockingAffects) {
		return false, nil
	}
	if !mover.VisibleTo(attacker) {
		return false, nil
	}
	if fighterHasAnyAffect(attacker, opportunityAttackWithdrawnAffects) {
		return false, nil
	}
	return b.opportunityAttackFacingAllows(attackerID, moverID)
}

// opportunityAttackFacingAllows 是那四道閘的最後一道：打手的**朝向 −2 .. ＋2**
// 五個方向裡，只要有一個讓移動者落在扇形內就打得到——正面加左右各 90°，
// 合起來 180°。
//
// ⚠ 打手的先攻 `+3` 還沒歸零（這一輪還沒輪到他），或這一輪一次都還沒動
// （動作計數 `+0Fh = 0`）時，**整個面向檢查被跳過，無條件打**。
func (b *Battle) opportunityAttackFacingAllows(attackerID, moverID string) (bool, error) {
	if b == nil {
		return false, fmt.Errorf("battle is nil")
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", attackerID)
	}
	mover, ok := b.fighters[moverID]
	if !ok {
		return false, fmt.Errorf("unknown fighter %q", moverID)
	}
	if attacker.CombatAction.Delay > 0 || attacker.CombatActionCount == 0 {
		return true, nil
	}
	return enginefacing.ArcSweep(facingStateOf(attacker), mover.CombatX, mover.CombatY,
		enginefacing.GoldBoxBounds), nil
}

// 原作把護甲值存成 `60 − 畫面上的 AC`（spec 767 的顯示式），命中判定是
//
//	d20 ＋ 攻擊者^[199h] ＋ v >= 目標的儲存 AC        （`overlay-23:123Fh`）
//
// 兩邊同時換回畫面刻度之後就是 AD&D 1e 的標準式
// `d20 ＋ 命中加值 ＋ 目標 AC >= 20`。remake 用的是**畫面刻度**。
const armorClassHitTarget = enginearmorclass.HitTarget

// StoredArmorClass 把畫面刻度的 AC 換成原作記錄裡的儲存值。原作衍生的算式
// （`CanHitTarget`、隊伍戰力）吃的是儲存值，跨這條邊界時要換。
func StoredArmorClass(display int) int { return enginearmorclass.Stored(display) }

// DisplayArmorClass 是 StoredArmorClass 的反向。
func DisplayArmorClass(stored int) int { return enginearmorclass.Display(stored) }

// StoredAttackBonus 把畫面刻度的命中加值（`THAC0 ＝ 20 − 它`）換成原作的
// `+199h`（`THAC0 ＝ 60 − 它`）。
func StoredAttackBonus(bonus int) int { return enginearmorclass.StoredAttackBonus(bonus) }

// DisplayAttackBonus 是 StoredAttackBonus 的反向。
func DisplayAttackBonus(stored int) int { return enginearmorclass.DisplayAttackBonus(stored) }

// DexterityDefenceAdjustment 是原作 `overlay-24:117Ah`（spec 694 的敏捷防禦表）：
// 回傳值加進儲存刻度的護甲值，**正值代表防禦變好**（畫面上的 AC 變小）。
//
// ⚠ 與先攻用的敏捷反應調整（`120Ah`）是**不同表**，值域與端點都不一樣。
func DexterityDefenceAdjustment(dexterity int) int {
	return engineability.DexterityDefenceAdjustment(dexterity)
}
