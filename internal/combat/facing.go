package combat

import "fmt"

// 戰鬥面向與轉向記帳（原作戰鬥狀態記錄的 `+09h`／`+0Fh`／`+12h`，spec 1137）。
//
// ★ 這一支刻意做**字面移植**，不做語意詮釋。原作兩個寫入端對「`+09h` 到底指
// 哪個方向」給出相反的形狀（spec 1019 的 `(d + 4) mod 8` 與 `sub_194A` 的
// `|d − 面向|`），所以「角色到底在看哪裡」還沒有結論。把算式原樣搬過來，
// remake 算出來的數字就與原作相同——不論那個數字最後怎麼解釋。

// combatDirection 是原作 `overlay-13:29A2h`：回傳**從 (x2,y2) 看向 (x1,y1)**
// 的八方向碼。判準是兩個定點正切門檻：
//
//	0x26A / 0x100 = 2.4140625（tan 67.5°）
//	0x006A / 0x100 = 0.4140625（tan 22.5°）
//
// ⚠ 乘除都是 16-bit 整數且**先乘後除**，順序不能換：`dx * 618 / 256`
// 與 `dx * (618/256)` 在整數上是兩個不同的函式。
func combatDirection(x1, y1, x2, y2 int) uint8 {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	wide := dx * 0x26A / 0x100  // 需要「很扁」才成立的那一邊
	narrow := dx * 0x6A / 0x100 // 需要「很陡」才成立的那一邊
	for direction := 0; direction < 8; direction++ {
		var ok bool
		switch direction {
		case 0:
			ok = y1 <= y2 && wide <= dy
		case 1:
			ok = y1 <= y2 && x1 >= x2 && wide >= dy && narrow <= dy
		case 2:
			ok = x1 >= x2 && narrow >= dy
		case 3:
			ok = y1 >= y2 && x1 >= x2 && wide >= dy && narrow <= dy
		case 4:
			ok = y1 >= y2 && wide <= dy
		case 5:
			ok = y1 >= y2 && x1 <= x2 && wide >= dy && narrow <= dy
		case 6:
			ok = x1 <= x2 && narrow >= dy
		case 7:
			ok = y1 <= y2 && x1 <= x2 && wide >= dy && narrow <= dy
		}
		if ok {
			return uint8(direction)
		}
	}
	// 八個扇形覆蓋整個平面（同格時 direction 0 成立），走不到這裡。
	return 0
}

// CombatDirectionFrom 是 combatDirection 的具名版本，給遊戲層與測試使用。
func CombatDirectionFrom(fromX, fromY, toX, toY int) uint8 {
	return combatDirection(toX, toY, fromX, fromY)
}

// turnDistance 是原作 `sub_194A` 的最短轉法：`(a − b + 8) mod 8`，大於 4 時
// 取 `8 − 差`。結果落在 0..4，順逆時針不分。
func turnDistance(a, b uint8) uint8 {
	difference := (int(a) - int(b) + 8) % 8
	if difference > 4 {
		difference = 8 - difference
	}
	return uint8(difference)
}

// AccountTurn 是原作 `overlay-13:194Ah`（`retf 8`，(角色, 對象)）：
// 動作計數加一，並把「對象看向角色的方向」與角色目前的 `+09h` 之間的最短轉法
// 累加進 `+12h`（對 8 取餘，所以它是 0..7 的環狀量，不是單調成本）。
//
// ⚠ 原作這一支**不寫 `+09h`**。面向的實際寫入在別處（開場的初始面向、
// 攻擊動作、以及 `overlay-32 entry#14` 的轉向）。
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
	fighter.CombatActionCount++
	direction := combatDirection(fighter.CombatX, fighter.CombatY, toward.CombatX, toward.CombatY)
	fighter.CombatTurnTotal = (fighter.CombatTurnTotal + turnDistance(direction, fighter.CombatFacing)) % 8
	b.fighters[fighterID] = fighter
	return nil
}

// FaceTarget 是原作攻擊動作（`overlay-13:19D8h`，spec 1019）的轉向段：
// **動作計數小於 2** 時才把 `+09h` 設成 `(方向 + 4) mod 8`。
//
// ⚠ DOS 的 `+ 4` 是平台差異，PC-98 把這一段整個拿掉（spec 1019）。
// 這裡照 DOS 走——本專案以 DOS 為行為 oracle。
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
	if fighter.CombatActionCount >= 2 {
		return false, nil
	}
	direction := combatDirection(fighter.CombatX, fighter.CombatY, target.CombatX, target.CombatY)
	fighter.CombatFacing = (direction + 4) % 8
	b.fighters[fighterID] = fighter
	return true, nil
}

// RearAttackApplies 是原作攻擊結算（`overlay-13:14E8h`）挑第二個 AC 欄位的三道
// 條件，全部成立才成立：
//
//  1. 防禦者這一回合的動作計數 `+0Fh` 大於 1；
//  2. `29A2h(攻擊者, 防禦者)`（＝從防禦者看向攻擊者的方向）等於防禦者的 `+09h`；
//  3. 防禦者的累計轉向 `+12h` 大於 4。
//
// ⚠ 三道都是字面移植。**沒有宣稱它叫「背後攻擊」**——那是形狀上的猜測，
// 而 `+09h` 的幾何意義本身還沒有結論（見本檔開頭）。
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
	if defender.CombatActionCount <= 1 || defender.CombatTurnTotal <= 4 {
		return false, nil
	}
	direction := combatDirection(attacker.CombatX, attacker.CombatY, defender.CombatX, defender.CombatY)
	return direction == defender.CombatFacing, nil
}

// initialFacingTable 是原作 `DS:2FAh` 的四筆初始面向（兩平台相同，spec 805）。
var initialFacingTable = [4]uint8{7, 2, 3, 6}

// ApplyInitialFacing 是原作戰鬥開場配置戰鬥狀態時設的面向（`overlay-10 entry#?`，
// spec 805）：`+09h := 表[有號(隊伍朝向) div 2]`，敵方（`+197h = 1`）再轉 180 度。
//
// ⚠ 索引是**隊伍在地城裡的朝向除以 2**，所以八個朝向只對到四筆。
func (b *Battle) ApplyInitialFacing(partyDirection uint8) {
	if b == nil {
		return
	}
	facing := initialFacingTable[(partyDirection%8)/2]
	for _, id := range b.fighterOrder {
		fighter := b.fighters[id]
		fighter.CombatFacing = facing
		if fighter.Side == SideEnemy {
			fighter.CombatFacing = (facing + 4) % 8
		}
		b.fighters[id] = fighter
	}
}

// combatMapMaxX／combatMapMaxY 是原作扇形判斷的座標上下界（`31h` 欄、`18h` 列），
// 也就是 50×25 的戰場（spec 1002）。四個座標任一個出界就直接否，不是夾到邊界。
const (
	combatMapMaxX = 0x31
	combatMapMaxY = 0x18
)

// nineCellOffsetX／nineCellOffsetY 是原作 `DS:2694h`／`DS:269Dh` 的九格方向表
// （spec 999）。索引 0..7 是八方向，索引 8 是「原地」。
var (
	nineCellOffsetX = [9]int{0, 1, 1, 1, 0, -1, -1, -1, 0}
	nineCellOffsetY = [9]int{-1, -1, 0, 1, 1, 1, 0, -1, 0}
)

// InFacingCone 是原作 `overlay-31:054Ah`（`retf 0Ah`，spec 1002）：
// 目標在不在「起點往 direction 前進一格」為頂點、半角 45° 的扇形內。
//
// ⚠ 頂點是 `(px, py) = 起點 + 位移[direction]`，**不是起點本身**；起點自己與
// 正前方那一格另外由前置判斷放行，所以正後方一格不算在內。
//
// direction 為 0FFh 時當成 8（原地），而 8 一律成立。0..8 與 0FFh 以外的值在
// 原作會讀到未初始化的區域變數，這裡回 false——沒有呼叫端會傳那種值。
func InFacingCone(startX, startY, targetX, targetY int, direction uint8) bool {
	if startX < 0 || startX > combatMapMaxX || startY < 0 || startY > combatMapMaxY ||
		targetX < 0 || targetX > combatMapMaxX || targetY < 0 || targetY > combatMapMaxY {
		return false
	}
	if direction == 0xFF {
		direction = 8
	}
	if direction > 8 {
		return false
	}
	px := startX + nineCellOffsetX[direction]
	py := startY + nineCellOffsetY[direction]
	if (startX == targetX && startY == targetY) || (px == targetX && py == targetY) {
		return true
	}
	// 八段各是兩支半平面判斷，字面照抄。正交四向的兩支互為鏡像，合起來是
	// 「沿軸的投影 ≥ 垂直方向的絕對值」；斜向四向的兩支則是分割線兩側的
	// 兩個 45° 楔形，合起來同樣是 90°。兩者的程式形狀完全相同。
	switch direction {
	case 0: // 上
		if targetX >= px && targetY <= py+px-targetX {
			return true
		}
		if targetX > px {
			return false
		}
		return targetY <= py+targetX-px
	case 1: // 右上
		if targetX >= px && targetY <= py+px-targetX {
			return true
		}
		if targetX < px+py-targetY {
			return false
		}
		return targetY <= py
	case 2: // 右
		if targetX >= px+py-targetY && targetY <= py {
			return true
		}
		if targetX < px+targetY-py {
			return false
		}
		return targetY >= py
	case 3: // 右下
		if targetX >= px+targetY-py && targetY >= py {
			return true
		}
		if targetX < px {
			return false
		}
		return targetY >= py+targetX-px
	case 4: // 下
		if targetX >= px && targetY >= py+targetX-px {
			return true
		}
		if targetX > px {
			return false
		}
		return targetY >= py+px-targetX
	case 5: // 左下
		if targetX <= px && targetY >= py+px-targetX {
			return true
		}
		if targetX > px+py-targetY {
			return false
		}
		return targetY >= py
	case 6: // 左
		if targetX <= px+py-targetY && targetY >= py {
			return true
		}
		if targetX > px+targetY-py {
			return false
		}
		return targetY <= py
	case 7: // 左上
		if targetX <= px+targetY-py && targetY <= py {
			return true
		}
		if targetX > px {
			return false
		}
		return targetY <= py+targetX-px
	}
	return true // direction 8：原地，無條件成立
}

// opportunityAttackFacingAllows 是原作「離開接觸就被打」（`overlay-13:095Ah`，
// spec 1010）動手前的最後一道閘：打手的**朝向 −2 .. ＋2** 這五個方向裡，
// 只要有一個讓移動者落在扇形內就打得到——正面加左右各 90°，合起來 180°。
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
	for offset := uint8(6); offset <= 0x0A; offset++ {
		direction := (attacker.CombatFacing + offset) % 8
		if InFacingCone(attacker.CombatX, attacker.CombatY, mover.CombatX, mover.CombatY, direction) {
			return true, nil
		}
	}
	return false, nil
}

// 原作把護甲值存成 `60 − 畫面上的 AC`（spec 767 的顯示式），命中判定是
//
//	d20 ＋ 攻擊者^[199h] ＋ v >= 目標的儲存 AC        （`overlay-23:123Fh`）
//
// 兩邊同時換回畫面刻度（`儲存值 ＝ 60 − 畫面值`、`+199h ＝ 60 − THAC0`）之後
// 就是 AD&D 1e 的標準式：
//
//	d20 ＋ 命中加值 ＋ 目標 AC >= 20
//
// remake 的 `Fighter.ArmorClass`／`AttackBonus` 用的是**畫面刻度**
// （AC 越小越難打、命中加值越大越好），所以判定用後面那一式。
const (
	armorClassStoredBase = 60
	armorClassHitTarget  = 20
)

// StoredArmorClass 把畫面刻度的 AC 換成原作記錄裡的儲存值。原作衍生的算式
// （`CanHitTarget`、隊伍戰力）吃的是儲存值，跨這條邊界時要換。
func StoredArmorClass(display int) int { return armorClassStoredBase - display }

// DisplayArmorClass 是 StoredArmorClass 的反向。
func DisplayArmorClass(stored int) int { return armorClassStoredBase - stored }

// StoredAttackBonus 把畫面刻度的命中加值（`THAC0 ＝ 20 − 它`）換成原作的
// `+199h`（`THAC0 ＝ 60 − 它`）。
func StoredAttackBonus(bonus int) int { return armorClassStoredBase - (armorClassHitTarget - bonus) }

// DisplayAttackBonus 是 StoredAttackBonus 的反向。
func DisplayAttackBonus(stored int) int { return armorClassHitTarget - (armorClassStoredBase - stored) }
