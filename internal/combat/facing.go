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
